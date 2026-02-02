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

// GetAnimeEpisodes mocks the GetAnimeEpisodes method
func (m *MockAnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	args := m.Called(animeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.Episode)
	if !ok {
		panic("mock: GetAnimeEpisodes return value is not []models.Episode")
	}
	return result, args.Error(1)
}

// GetUserFavorites mocks the GetUserFavorites method
func (m *MockAnimeService) GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error) {
	args := m.Called(userID, status)
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
func (m *MockAnimeService) UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error) {
	args := m.Called(userID, animeID, status)
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

// UpdateWatchStatus mocks the UpdateWatchStatus method
func (m *MockAnimeService) UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error) {
	args := m.Called(userID, episodeID, viewed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.WatchStatus)
	if !ok {
		panic("mock: UpdateWatchStatus return value is not *models.WatchStatus")
	}
	return result, args.Error(1)
}

// UpdateRating mocks the UpdateRating method
func (m *MockAnimeService) UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error) {
	args := m.Called(userID, animeID, rating, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).(*models.UserRating)
	if !ok {
		panic("mock: UpdateRating return value is not *models.UserRating")
	}
	return result, args.Error(1)
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
	result, ok := args.Get(0).([]models.WatchHistory)
	if !ok {
		panic("mock: GetRecentlyWatched return value is not []models.WatchHistory")
	}
	return result, args.Error(1)
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

// GetEpisodesByType mocks the GetEpisodesByType method
func (m *MockAnimeService) GetEpisodesByType(animeID int, isSpecial bool, userID *int) ([]models.Episode, error) {
	args := m.Called(animeID, isSpecial, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, ok := args.Get(0).([]models.Episode)
	if !ok {
		panic("mock: GetEpisodesByType return value is not []models.Episode")
	}
	return result, args.Error(1)
}

// GetEpisodesSeparated mocks the GetEpisodesSeparated method
func (m *MockAnimeService) GetEpisodesSeparated(animeID int, userID *int) (regular, specials []models.Episode, err error) {
	args := m.Called(animeID, userID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	regular, ok := args.Get(0).([]models.Episode)
	if !ok {
		panic("mock: GetEpisodesSeparated regular return value is not []models.Episode")
	}
	specials, ok = args.Get(1).([]models.Episode)
	if !ok {
		panic("mock: GetEpisodesSeparated specials return value is not []models.Episode")
	}
	return regular, specials, args.Error(2)
}
