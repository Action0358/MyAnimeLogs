package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
)

// MockAnimeService for testing
type MockAnimeService struct {
	mock.Mock
}

func (m *MockAnimeService) SearchAnimes(params models.AnimeSearchParams, userID *int) (*models.PaginatedResponse, error) {
	args := m.Called(params, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedResponse), args.Error(1)
}

func (m *MockAnimeService) AdvancedSearchAnimes(params models.AdvancedSearchParams, userID *int) (*models.PaginatedResponse, error) {
	args := m.Called(params, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaginatedResponse), args.Error(1)
}

func (m *MockAnimeService) GetAnimeByID(animeID int, userID *int) (*models.AnimeWithUserData, error) {
	args := m.Called(animeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnimeWithUserData), args.Error(1)
}

func (m *MockAnimeService) GetAnimeEpisodes(animeID int, userID *int) ([]models.Episode, error) {
	args := m.Called(animeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Episode), args.Error(1)
}

func (m *MockAnimeService) GetUserFavorites(userID int, status string) ([]models.AnimeWithUserData, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AnimeWithUserData), args.Error(1)
}

func (m *MockAnimeService) UpdateFavorite(userID, animeID int, status string) (*models.UserFavorite, error) {
	args := m.Called(userID, animeID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserFavorite), args.Error(1)
}

func (m *MockAnimeService) DeleteFavorite(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

func (m *MockAnimeService) UpdateWatchStatus(userID, episodeID int, viewed bool) (*models.WatchStatus, error) {
	args := m.Called(userID, episodeID, viewed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WatchStatus), args.Error(1)
}

func (m *MockAnimeService) UpdateRating(userID, animeID int, rating float64, comment *string) (*models.UserRating, error) {
	args := m.Called(userID, animeID, rating, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserRating), args.Error(1)
}

func (m *MockAnimeService) DeleteRating(userID, animeID int) error {
	args := m.Called(userID, animeID)
	return args.Error(0)
}

func (m *MockAnimeService) GetRecentlyWatched(userID, limit int) ([]models.WatchHistory, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.WatchHistory), args.Error(1)
}

func (m *MockAnimeService) GetFilterOptions() (*models.FilterOptions, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FilterOptions), args.Error(1)
}

func (m *MockAnimeService) GetPopularSearches(limit int) ([]string, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAnimeService) GetSearchSuggestions(query string, limit int) ([]string, error) {
	args := m.Called(query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAnimeService) SyncEpisodeDurations(animeID *int) (int, int, int, int, []string, error) {
	args := m.Called(animeID)
	return args.Int(0), args.Int(1), args.Int(2), args.Int(3), args.Get(4).([]string), args.Error(5)
}

func TestAdminHandler_SyncEpisodeDurations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)

	t.Run("Success - Sync All Animes", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		mockService.On("SyncEpisodeDurations", (*int)(nil)).
			Return(100, 90, 5, 5, []string{}, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor)

		// Create request
		reqBody := SyncEpisodeDurationsRequest{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.SyncEpisodeDurations(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(100), response["total_animes"])
		assert.Equal(t, float64(90), response["success_count"])
		assert.Equal(t, float64(5), response["failure_count"])
		assert.Equal(t, float64(5), response["skipped_count"])

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Sync Specific Anime", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		animeID := 123
		mockService.On("SyncEpisodeDurations", &animeID).
			Return(1, 1, 0, 0, []string{}, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor)

		// Create request
		reqBody := SyncEpisodeDurationsRequest{AnimeID: &animeID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.SyncEpisodeDurations(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["total_animes"])
		assert.Equal(t, float64(1), response["success_count"])
		assert.Equal(t, float64(0), response["failure_count"])
		assert.Equal(t, float64(0), response["skipped_count"])

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Sync Specific Anime Without MAL ID (Skipped)", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		animeID := 456
		mockService.On("SyncEpisodeDurations", &animeID).
			Return(1, 0, 0, 1, []string{}, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor)

		// Create request
		reqBody := SyncEpisodeDurationsRequest{AnimeID: &animeID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.SyncEpisodeDurations(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["total_animes"])
		assert.Equal(t, float64(0), response["success_count"])
		assert.Equal(t, float64(0), response["failure_count"])
		assert.Equal(t, float64(1), response["skipped_count"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid Request Body", func(t *testing.T) {
		mockService := new(MockAnimeService)
		handler := NewAdminHandler(mockService, logger, errorMonitor)

		// Invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SyncEpisodeDurations(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminHandler_NewAdminHandler(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)
	mockService := new(MockAnimeService)

	handler := NewAdminHandler(mockService, logger, errorMonitor)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.BaseHandler)
	assert.NotNil(t, handler.animeService)
	assert.NotNil(t, handler.logger)
}
