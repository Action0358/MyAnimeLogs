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
	result, ok := args.Get(0).(*models.PaginatedResponse)
	if !ok {
		panic("mock: SearchAnimes return value is not *models.PaginatedResponse")
	}
	return result, args.Error(1)
}

// AdvancedSearchAnimes mocks the AdvancedSearchAnimes method
func (m *MockAnimeService) AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error) {
	args := m.Called(params, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.PaginatedResponse)
	if !ok {
		panic("mock: AdvancedSearchAnimes return value is not *models.PaginatedResponse")
	}
	return result, args.Error(1)
}

// GetSearchSuggestions mocks the GetSearchSuggestions method
func (m *MockAnimeService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	args := m.Called(query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]string)
	if !ok {
		panic("mock: GetSearchSuggestions return value is not []string")
	}
	return result, args.Error(1)
}

// GetPopularSearches mocks the GetPopularSearches method
func (m *MockAnimeService) GetPopularSearches(limit int) ([]string, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]string)
	if !ok {
		panic("mock: GetPopularSearches return value is not []string")
	}
	return result, args.Error(1)
}

// GetFilterOptions mocks the GetFilterOptions method
func (m *MockAnimeService) GetFilterOptions() (*models.FilterOptions, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.FilterOptions)
	if !ok {
		panic("mock: GetFilterOptions return value is not *models.FilterOptions")
	}
	return result, args.Error(1)
}

// GetAnimeByID mocks the GetAnimeByID method
func (m *MockAnimeService) GetAnimeByID(id int, userID *int) (*models.AnimeWithUserData, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.AnimeWithUserData)
	if !ok {
		panic("mock: GetAnimeByID return value is not *models.AnimeWithUserData")
	}
	return result, args.Error(1)
}

// GetUserFavorites mocks the GetUserFavorites method
func (m *MockAnimeService) GetUserFavorites(userID int) ([]models.AnimeWithUserData, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.AnimeWithUserData)
	if !ok {
		panic("mock: GetUserFavorites return value is not []models.AnimeWithUserData")
	}
	return result, args.Error(1)
}

// UpdateFavorite mocks the UpdateFavorite method
func (m *MockAnimeService) UpdateFavorite(userID, animeID int) (*models.UserFavorite, error) {
	args := m.Called(userID, animeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.UserFavorite)
	if !ok {
		panic("mock: UpdateFavorite return value is not *models.UserFavorite")
	}
	return result, args.Error(1)
}

// DeleteFavorite mocks the DeleteFavorite method
func (m *MockAnimeService) DeleteFavorite(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

// GetUserWatchStatuses mocks the GetUserWatchStatuses method
func (m *MockAnimeService) GetUserWatchStatuses(userID int, status string) ([]models.AnimeWithUserData, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.AnimeWithUserData)
	if !ok {
		panic("mock: GetUserWatchStatuses return value is not []models.AnimeWithUserData")
	}
	return result, args.Error(1)
}

// UpdateWatchStatus mocks the UpdateWatchStatus method
func (m *MockAnimeService) UpdateWatchStatus(userID, animeID int, status string) (*models.UserWatchStatus, error) {
	args := m.Called(userID, animeID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.UserWatchStatus)
	if !ok {
		panic("mock: UpdateWatchStatus return value is not *models.UserWatchStatus")
	}
	return result, args.Error(1)
}

// DeleteWatchStatus mocks the DeleteWatchStatus method
func (m *MockAnimeService) DeleteWatchStatus(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

// GetCurrentSeasonAnimes mocks the GetCurrentSeasonAnimes method
func (m *MockAnimeService) GetCurrentSeasonAnimes(limit int) ([]models.Anime, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.Anime)
	if !ok {
		panic("mock: GetCurrentSeasonAnimes return value is not []models.Anime")
	}
	return result, args.Error(1)
}

// GetPopularAnimes mocks the GetPopularAnimes method
func (m *MockAnimeService) GetPopularAnimes(limit int) ([]models.Anime, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.Anime)
	if !ok {
		panic("mock: GetPopularAnimes return value is not []models.Anime")
	}
	return result, args.Error(1)
}

// GetSearchHistory mocks the GetSearchHistory method
func (m *MockAnimeService) GetSearchHistory(userID int, limit int) ([]models.SearchLog, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.SearchLog)
	if !ok {
		panic("mock: GetSearchHistory return value is not []models.SearchLog")
	}
	return result, args.Error(1)
}

// DeleteSearchHistory mocks the DeleteSearchHistory method
func (m *MockAnimeService) DeleteSearchHistory(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

// DeleteSearchHistoryItem mocks the DeleteSearchHistoryItem method
func (m *MockAnimeService) DeleteSearchHistoryItem(userID int, logID int) error {
	args := m.Called(userID, logID)
	return args.Error(0)
}

