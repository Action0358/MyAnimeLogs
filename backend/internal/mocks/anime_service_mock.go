package mocks

import (
	"myanimelogs/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockAnimeService is a mock implementation of AnimeServiceInterface
type MockAnimeService struct {
	mock.Mock
}

// SearchAnimes mocks the SearchAnimes method
func (m *MockAnimeService) SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error) {
	args := m.Called(params, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedResponse), args.Error(1)
}

// AdvancedSearchAnimes mocks the AdvancedSearchAnimes method
func (m *MockAnimeService) AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error) {
	args := m.Called(params, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedResponse), args.Error(1)
}

// GetSearchSuggestions mocks the GetSearchSuggestions method
func (m *MockAnimeService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	args := m.Called(query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetPopularSearches mocks the GetPopularSearches method
func (m *MockAnimeService) GetPopularSearches(limit int) ([]string, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetFilterOptions mocks the GetFilterOptions method
func (m *MockAnimeService) GetFilterOptions() (*models.FilterOptions, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FilterOptions), args.Error(1)
}

// GetAnimeByID mocks the GetAnimeByID method
func (m *MockAnimeService) GetAnimeByID(id int, userID *int) (*models.AnimeWithUserData, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnimeWithUserData), args.Error(1)
}

// GetAnimeEpisodes mocks the GetAnimeEpisodes method
func (m *MockAnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	args := m.Called(animeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Episode), args.Error(1)
}

// GetUserFavorites mocks the GetUserFavorites method
func (m *MockAnimeService) GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AnimeWithUserData), args.Error(1)
}

// UpdateFavorite mocks the UpdateFavorite method
func (m *MockAnimeService) UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error) {
	args := m.Called(userID, animeID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFavorite), args.Error(1)
}

// DeleteFavorite mocks the DeleteFavorite method
func (m *MockAnimeService) DeleteFavorite(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

// UpdateWatchStatus mocks the UpdateWatchStatus method
func (m *MockAnimeService) UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error) {
	args := m.Called(userID, episodeID, viewed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WatchStatus), args.Error(1)
}

// UpdateRating mocks the UpdateRating method
func (m *MockAnimeService) UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error) {
	args := m.Called(userID, animeID, rating, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserRating), args.Error(1)
}

// DeleteRating mocks the DeleteRating method
func (m *MockAnimeService) DeleteRating(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

// GetRecentlyWatched mocks the GetRecentlyWatched method
func (m *MockAnimeService) GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.WatchHistory), args.Error(1)
}
