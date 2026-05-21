package services

import (
	"time"

	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error)
	UpdateIcon(userID int, iconURL string) (*models.User, error)
	EmailExists(email string) (bool, error)
	// Account lockout methods
	IncrementFailedAttempts(userID int, maxAttempts int, lockoutDuration time.Duration) (*repository.LockoutResult, error)
	ResetFailedAttempts(userID int) error
	IsAccountLocked(userID int) (bool, *time.Time, error)
}

// AuthServiceInterface defines the interface for authentication service
type AuthServiceInterface interface {
	Register(req models.UserRegisterRequest) (*models.UserResponse, error)
	Login(req models.UserLoginRequest, userAgent, ipAddress string) (*models.LoginResponse, error)
	GetUserByID(userID int) (*models.UserResponse, error)
	UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error)
	UpdateUserIcon(userID int, iconURL string) (*models.UserResponse, error)
	ExportUserData(userID int, format string) (interface{}, error)
	RefreshToken(refreshToken string) (*models.TokenRefreshResponse, error)
	Logout(refreshToken string) error
	LogoutAllDevices(userID int) error
	GetUserSessions(userID int, currentJTI string) ([]models.SessionResponse, error)
	DeleteSession(sessionID, userID int) error
	GetSessionService() SessionServiceInterface
}

// Ensure AuthService implements AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// AnimeServiceInterface defines the interface for anime service
type AnimeServiceInterface interface {
	SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error)
	AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error)
	GetAnimeByID(id int, userID *int) (*models.AnimeWithUserData, error)
	GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error)
	UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error)
	DeleteFavorite(userID, animeID int) error
	UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error)
	DeleteRating(userID, animeID int) error
	GetFilterOptions() (*models.FilterOptions, error)
	GetPopularSearches(limit int) ([]string, error)
	GetSearchSuggestions(query string, limit int) ([]string, error)
	GetCurrentSeasonAnimes(limit int) ([]models.Anime, error)
	GetPopularAnimes(limit int) ([]models.Anime, error)
	// 検索履歴
	GetSearchHistory(userID int, limit int) ([]models.SearchLog, error)
	DeleteSearchHistory(userID int) error
	DeleteSearchHistoryItem(userID int, logID int) error
}

// Ensure AnimeService implements AnimeServiceInterface
var _ AnimeServiceInterface = (*AnimeService)(nil)

// SessionServiceInterface defines the interface for session service
type SessionServiceInterface interface {
	CreateSession(userID int, jti, userAgent, ipAddress string) (*models.Session, string, error)
	ValidateSession(refreshToken string) (*models.Session, error)
	UpdateActivity(sessionID int) error
	UpdateActivityByJTI(jti string) error
	UpdateSessionJTI(sessionID int, newJTI string) error
	DeleteSession(refreshToken string) error
	DeleteSessionByID(sessionID, userID int) error
	DeleteUserSessions(userID int) error
	GetUserSessions(userID int) ([]models.Session, error)
	CleanupExpiredSessions() (int, error)
	GetSessionByJTI(jti string) (*models.Session, error)
}

// Ensure SessionService implements SessionServiceInterface
var _ SessionServiceInterface = (*SessionService)(nil)
