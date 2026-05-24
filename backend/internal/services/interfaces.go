package services

import (
	"myanimelogs/internal/models"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetOrCreateByAuthUserID(authUserID, email, name string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error)
	UpdateIcon(userID int, iconURL string) (*models.User, error)
}

// AuthServiceInterface defines the interface for user management service
type AuthServiceInterface interface {
	GetUserByID(userID int) (*models.UserResponse, error)
	UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error)
	UpdateUserIcon(userID int, iconURL string) (*models.UserResponse, error)
	ExportUserData(userID int, format string) (interface{}, error)
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
