package services

import (
	"myanimelogs/internal/models"
	"myanimelogs/internal/repository"
)

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetOrCreateByAuthUserID(authUserID, email, name string) (*models.User, repository.AuthLinkResult, error)
	Create(user *models.User) error
}

// AuthServiceInterface defines the interface for user management service
type AuthServiceInterface interface {
	// DeleteAccount はアカウントを完全に削除する（復旧不可）
	DeleteAccount(userID int, authUserID string) error
	// avatarURL はSupabase JWTから取得したGoogleアカウントのプロフィール画像URL（#779）
	GetUserByID(userID int, avatarURL string) (*models.UserResponse, error)
	ExportUserData(userID int, format string) (interface{}, error)
}

// Ensure AuthService implements AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// AnimeServiceInterface defines the interface for anime service
type AnimeServiceInterface interface {
	SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error)
	AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error)
	GetAnimeByID(id int, userID *int) (*models.AnimeWithUserData, error)
	GetUserFavorites(userID int) ([]models.AnimeWithUserData, error)
	UpdateFavorite(userID, animeID int) (*models.UserFavorite, error)
	DeleteFavorite(userID, animeID int) error
	GetUserWatchStatuses(userID int, status string) ([]models.AnimeWithUserData, error)
	UpdateWatchStatus(userID, animeID int, status string) (*models.UserWatchStatus, error)
	DeleteWatchStatus(userID, animeID int) error
	GetFilterOptions() (*models.FilterOptions, error)
	GetPopularSearches(limit int) ([]string, error)
	GetSearchSuggestions(query string, limit int) ([]string, error)
	GetCurrentSeasonAnimes(limit int) ([]models.Anime, error)
	GetPopularAnimes(limit int, random bool) ([]models.Anime, error)
	// 検索履歴
	GetSearchHistory(userID int, limit int) ([]models.SearchLog, error)
	DeleteSearchHistory(userID int) error
	DeleteSearchHistoryItem(userID int, logID int) error
}

// Ensure AnimeService implements AnimeServiceInterface
var _ AnimeServiceInterface = (*AnimeService)(nil)
