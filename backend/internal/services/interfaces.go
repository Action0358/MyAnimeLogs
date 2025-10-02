package services

import "myanimelogs/internal/models"

// UserRepositoryInterface defines the interface for user repository
type UserRepositoryInterface interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	UpdatePartial(userID int, req models.UserUpdateRequest) (*models.User, error)
	EmailExists(email string) (bool, error)
}

// AuthServiceInterface defines the interface for authentication service
type AuthServiceInterface interface {
	Register(req models.UserRegisterRequest) (*models.UserResponse, error)
	Login(req models.UserLoginRequest) (*models.LoginResponse, error)
	GetUserByID(userID int) (*models.UserResponse, error)
	UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error)
	GetUserStats(userID int) (*models.UserStats, error)
	ExportUserData(userID int, format string) (interface{}, error)
}

// Ensure AuthService implements AuthServiceInterface
var _ AuthServiceInterface = (*AuthService)(nil)

// AnimeServiceInterface defines the interface for anime service
type AnimeServiceInterface interface {
	SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error)
	GetAnimeByID(id int, userID *int) (*models.AnimeWithUserData, error)
	GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error)
	GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error)
	UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error)
	DeleteFavorite(userID, animeID int) error
	UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error)
	UpdateRating(userID, animeID int, rating float64, comment string) (*models.UserRating, error)
	DeleteRating(userID, animeID int) error
	GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error)
}
