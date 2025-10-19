package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"myanimelogs/internal/common"
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
	userRepo     UserRepositoryInterface
	db           *database.DB // Keep for export methods temporarily
	statsService *StatsService
	logger       *common.Logger
	jwtSecret    string
}

func NewAuthService(db *database.DB, jwtSecret string) *AuthService {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	return &AuthService{
		userRepo:     nil, // Will be set by NewAuthServiceWithRepo
		db:           db,
		statsService: NewStatsService(db, logger),
		logger:       logger,
		jwtSecret:    jwtSecret,
	}
}

// NewAuthServiceWithRepo creates a new AuthService with dependency injection
func NewAuthServiceWithRepo(userRepo UserRepositoryInterface, db *database.DB, jwtSecret string) *AuthService {
	logger := common.NewLogger(common.LogLevelInfo, "json")
	return &AuthService{
		userRepo:     userRepo,
		db:           db,
		statsService: NewStatsService(db, logger),
		logger:       logger,
		jwtSecret:    jwtSecret,
	}
}

func (s *AuthService) Register(req models.UserRegisterRequest) (*models.UserResponse, error) {
	// Use repository if available, otherwise fallback to direct DB access
	if s.userRepo != nil {
		return s.registerWithRepo(req)
	}
	return s.registerWithDB(req)
}

func (s *AuthService) registerWithRepo(req models.UserRegisterRequest) (*models.UserResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(req.Email)
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

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) registerWithDB(req models.UserRegisterRequest) (*models.UserResponse, error) {
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
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) Login(req models.UserLoginRequest) (*models.LoginResponse, error) {
	// Use repository if available, otherwise fallback to direct DB access
	if s.userRepo != nil {
		return s.loginWithRepo(req)
	}
	return s.loginWithDB(req)
}

func (s *AuthService) loginWithRepo(req models.UserLoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
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
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

func (s *AuthService) loginWithDB(req models.UserLoginRequest) (*models.LoginResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, name, email, password, description, icon, created_at
		FROM users WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Description, &user.Icon, &user.CreatedAt)
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
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

func (s *AuthService) GetUserByID(userID int) (*models.UserResponse, error) {
	// Use repository if available, otherwise fallback to direct DB access
	if s.userRepo != nil {
		return s.getUserByIDWithRepo(userID)
	}
	return s.getUserByIDWithDB(userID)
}

func (s *AuthService) getUserByIDWithRepo(userID int) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) getUserByIDWithDB(userID int) (*models.UserResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, name, email, description, icon, created_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon, &user.CreatedAt)
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
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	// Use repository if available, otherwise fallback to direct DB access
	if s.userRepo != nil {
		return s.updateUserWithRepo(userID, req)
	}
	return s.updateUserWithDB(userID, req)
}

func (s *AuthService) updateUserWithRepo(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.UpdatePartial(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &models.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Description: user.Description,
		Icon:        user.Icon,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) updateUserWithDB(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
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
		RETURNING id, name, email, description, icon, created_at
	`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.Description, &user.Icon, &user.CreatedAt,
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
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *AuthService) GetUserStats(userID int) (*models.UserStats, error) {
	// Delegate to StatsService
	return s.statsService.GetUserStats(userID)
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
