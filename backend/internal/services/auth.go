package services

import (
	"errors"
	"fmt"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/database"
	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"
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

// getRepository returns the repository instance, creating one if not injected
func (s *AuthService) getRepository() UserRepositoryInterface {
	if s.userRepo != nil {
		return s.userRepo
	}
	// Fallback: create repository instance if not injected
	return repository.NewUserRepository(s.db)
}

func (s *AuthService) Register(req models.UserRegisterRequest) (*models.UserResponse, error) {
	return s.registerWithRepo(req, s.getRepository())
}

// registerWithRepo handles user registration through repository
func (s *AuthService) registerWithRepo(req models.UserRegisterRequest, repo UserRepositoryInterface) (*models.UserResponse, error) {
	// Check if email already exists
	exists, err := repo.EmailExists(req.Email)
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

	err = repo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &models.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Description:   user.Description,
		Icon:          user.Icon,
		FavoriteAnime: user.FavoriteAnime,
		SnsLink:       user.SnsLink,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// registerWithDB is kept for backward compatibility but delegates to repository
// Deprecated: Use Register instead
func (s *AuthService) registerWithDB(req models.UserRegisterRequest) (*models.UserResponse, error) {
	repo := repository.NewUserRepository(s.db)
	return s.registerWithRepo(req, repo)
}

func (s *AuthService) Login(req models.UserLoginRequest) (*models.LoginResponse, error) {
	return s.loginWithRepo(req, s.getRepository())
}

// loginWithRepo handles user login through repository
func (s *AuthService) loginWithRepo(req models.UserLoginRequest, repo UserRepositoryInterface) (*models.LoginResponse, error) {
	user, err := repo.GetByEmail(req.Email)
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
			ID:            user.ID,
			Name:          user.Name,
			Email:         user.Email,
			Description:   user.Description,
			Icon:          user.Icon,
			FavoriteAnime: user.FavoriteAnime,
			SnsLink:       user.SnsLink,
			CreatedAt:     user.CreatedAt,
		},
	}, nil
}

// loginWithDB is kept for backward compatibility but delegates to repository
// Deprecated: Use Login instead
func (s *AuthService) loginWithDB(req models.UserLoginRequest) (*models.LoginResponse, error) {
	repo := repository.NewUserRepository(s.db)
	return s.loginWithRepo(req, repo)
}

func (s *AuthService) GetUserByID(userID int) (*models.UserResponse, error) {
	return s.getUserByIDWithRepo(userID, s.getRepository())
}

// getUserByIDWithDB is kept for backward compatibility but delegates to repository
// Deprecated: Use GetUserByID instead
func (s *AuthService) getUserByIDWithDB(userID int) (*models.UserResponse, error) {
	repo := repository.NewUserRepository(s.db)
	return s.getUserByIDWithRepo(userID, repo)
}

// getUserByIDWithRepo handles user retrieval through repository
func (s *AuthService) getUserByIDWithRepo(userID int, repo UserRepositoryInterface) (*models.UserResponse, error) {
	user, err := repo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &models.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Description:   user.Description,
		Icon:          user.Icon,
		FavoriteAnime: user.FavoriteAnime,
		SnsLink:       user.SnsLink,
		CreatedAt:     user.CreatedAt,
	}, nil
}

func (s *AuthService) UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	return s.updateUserWithRepo(userID, req, s.getRepository())
}

// updateUserWithRepo handles user update through repository
func (s *AuthService) updateUserWithRepo(userID int, req models.UserUpdateRequest, repo UserRepositoryInterface) (*models.UserResponse, error) {
	user, err := repo.UpdatePartial(userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &models.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Description:   user.Description,
		Icon:          user.Icon,
		FavoriteAnime: user.FavoriteAnime,
		SnsLink:       user.SnsLink,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// updateUserWithDB is kept for backward compatibility but delegates to repository
// Deprecated: Use UpdateUser instead
func (s *AuthService) updateUserWithDB(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	repo := repository.NewUserRepository(s.db)
	return s.updateUserWithRepo(userID, req, repo)
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
