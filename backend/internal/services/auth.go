package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthService(db *database.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(req models.UserRegisterRequest) (*models.UserResponse, error) {
	// Check if email already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	var user models.User
	err = s.db.QueryRow(`
		INSERT INTO users (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, email, description, icon, created_at, updated_at
	`, req.Name, req.Email, string(hashedPassword)).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
	}, nil
}

func (s *AuthService) Login(req models.UserLoginRequest) (*models.LoginResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, name, email, password, description, icon
		FROM users WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Description, &user.Icon)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Name, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User: models.UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Description: user.Description,
			Icon:        user.Icon,
		},
	}, nil
}

func (s *AuthService) GetUserByID(userID int) (*models.UserResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, name, email, description, icon
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
	}, nil
}

func (s *AuthService) UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Icon != nil {
		setParts = append(setParts, fmt.Sprintf("icon = $%d", argIndex))
		args = append(args, *req.Icon)
		argIndex++
	}

	if len(setParts) == 0 {
		// No fields to update, just return current user
		return s.GetUserByID(userID)
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, name, email, description, icon
	`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
	}, nil
}

func (s *AuthService) GetUserStats(userID int) (*models.UserStats, error) {
	stats := &models.UserStats{}

	// Get total episodes watched
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_statuses 
		WHERE user_id = $1 AND viewed = true
	`, userID).Scan(&stats.TotalEpisodesWatched)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes watched: %w", err)
	}

	// Get total animes completed
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM user_favorites 
		WHERE user_id = $1 AND status = 'completed'
	`, userID).Scan(&stats.TotalAnimesCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed animes: %w", err)
	}

	// Estimate total watch time (assuming 24 minutes per episode)
	stats.TotalWatchTimeHours = float64(stats.TotalEpisodesWatched) * 0.4

	// Get favorite genres (top 3)
	rows, err := s.db.Query(`
		SELECT ag.genre, COUNT(*) as count
		FROM anime_genres ag
		JOIN user_favorites uf ON ag.anime_id = uf.anime_id
		WHERE uf.user_id = $1
		GROUP BY ag.genre
		ORDER BY count DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre string
		var count int
		if err := rows.Scan(&genre, &count); err != nil {
			continue
		}
		stats.FavoriteGenres = append(stats.FavoriteGenres, genre)
	}

	// Get this month's episodes
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM watch_history 
		WHERE user_id = $1 AND watched_at >= date_trunc('month', CURRENT_DATE)
	`, userID).Scan(&stats.ThisMonthEpisodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month episodes: %w", err)
	}

	return stats, nil
}

func (s *AuthService) ExportUserData(userID int, format string) (interface{}, error) {
	// Get user info
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Get favorites
	favorites := []models.UserFavorite{}
	rows, err := s.db.Query(`
		SELECT id, user_id, anime_id, status, created_at, updated_at
		FROM user_favorites WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorites: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fav models.UserFavorite
		if err := rows.Scan(&fav.ID, &fav.UserID, &fav.AnimeID, &fav.Status, &fav.CreatedAt, &fav.UpdatedAt); err != nil {
			continue
		}
		favorites = append(favorites, fav)
	}

	// Get watch history
	history := []models.WatchHistory{}
	rows, err = s.db.Query(`
		SELECT wh.id, wh.user_id, wh.episode_id, wh.watched_at
		FROM watch_history wh WHERE wh.user_id = $1
		ORDER BY wh.watched_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watch history: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var h models.WatchHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.EpisodeID, &h.WatchedAt); err != nil {
			continue
		}
		history = append(history, h)
	}

	// Get ratings
	ratings := []models.UserRating{}
	rows, err = s.db.Query(`
		SELECT id, user_id, anime_id, rating, comment, created_at, updated_at
		FROM user_ratings WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rating models.UserRating
		if err := rows.Scan(&rating.ID, &rating.UserID, &rating.AnimeID, &rating.Rating, &rating.Comment, &rating.CreatedAt, &rating.UpdatedAt); err != nil {
			continue
		}
		ratings = append(ratings, rating)
	}

	data := map[string]interface{}{
		"user_info":     user,
		"favorites":     favorites,
		"watch_history": history,
		"ratings":       ratings,
		"exported_at":   time.Now(),
	}

	if format == "csv" {
		// Convert to CSV format (simplified)
		csv := "type,id,data\n"
		csv += fmt.Sprintf("user,%d,\"%s\"\n", user.ID, user.Name)
		for _, fav := range favorites {
			csv += fmt.Sprintf("favorite,%d,\"%s\"\n", fav.AnimeID, fav.Status)
		}
		return csv, nil
	}

	return data, nil
}
