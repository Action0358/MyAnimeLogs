package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"
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
	if args.Get(4) == nil {
		return args.Int(0), args.Int(1), args.Int(2), args.Int(3), nil, args.Error(5)
	}
	return args.Int(0), args.Int(1), args.Int(2), args.Int(3), args.Get(4).([]string), args.Error(5)
}

func (m *MockAnimeService) CreateEpisodesForAllAnimes() (int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockAnimeService) EnrichEpisodesWithTitles(animeID *int) (int, int, []string, error) {
	args := m.Called(animeID)
	if args.Get(2) == nil {
		return args.Int(0), args.Int(1), nil, args.Error(3)
	}
	return args.Int(0), args.Int(1), args.Get(2).([]string), args.Error(3)
}

func (m *MockAnimeService) GetEpisodeSyncStatus(animeID int) (*models.EpisodeSyncStatus, error) {
	args := m.Called(animeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EpisodeSyncStatus), args.Error(1)
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

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

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

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

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

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

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
		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SyncEpisodeDurations(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - Anime Not Found (404)", func(t *testing.T) {
		mockService := new(MockAnimeService)
		animeID := 999
		mockService.On("SyncEpisodeDurations", &animeID).
			Return(0, 0, 0, 0, nil, services.ErrAnimeNotFound)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		reqBody := SyncEpisodeDurationsRequest{AnimeID: &animeID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.SyncEpisodeDurations(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAdminHandler_CreateEpisodesForAllAnimes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)

	t.Run("Success - Create Episodes", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		mockService.On("CreateEpisodesForAllAnimes").
			Return(95, 0, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/admin/create-episodes", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.CreateEpisodesForAllAnimes(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response CreateEpisodesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 95, response.SuccessCount)
		assert.Equal(t, 0, response.FailureCount)
		assert.Contains(t, response.Message, "95 succeeded, 0 failed")

		mockService.AssertExpectations(t)
	})

	t.Run("Success - With Some Failures", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		mockService.On("CreateEpisodesForAllAnimes").
			Return(90, 5, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/admin/create-episodes", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.CreateEpisodesForAllAnimes(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response CreateEpisodesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 90, response.SuccessCount)
		assert.Equal(t, 5, response.FailureCount)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Error", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		mockService.On("CreateEpisodesForAllAnimes").
			Return(0, 0, fmt.Errorf("database connection failed"))

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/admin/create-episodes", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.CreateEpisodesForAllAnimes(c)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockService.AssertExpectations(t)
	})
}

func TestAdminHandler_EnrichEpisodeTitles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)

	t.Run("Success - Enrich All Animes", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		mockService.On("EnrichEpisodesWithTitles", (*int)(nil)).
			Return(10, 0, []string{}, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		reqBody := EnrichEpisodeTitlesRequest{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/enrich/episode-titles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.EnrichEpisodeTitles(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response EnrichEpisodeTitlesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 10, response.SuccessCount)
		assert.Equal(t, 0, response.FailureCount)
		assert.Empty(t, response.Errors)
		assert.Contains(t, response.Message, "10 succeeded, 0 failed")

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Enrich Specific Anime", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		animeID := 1
		mockService.On("EnrichEpisodesWithTitles", &animeID).
			Return(1, 0, []string{}, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		reqBody := EnrichEpisodeTitlesRequest{AnimeID: &animeID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/enrich/episode-titles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.EnrichEpisodeTitles(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response EnrichEpisodeTitlesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.SuccessCount)
		assert.Equal(t, 0, response.FailureCount)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - With Some Failures", func(t *testing.T) {
		// Setup mock
		mockService := new(MockAnimeService)
		errors := []string{"Anime ID 5: no episodes found"}
		mockService.On("EnrichEpisodesWithTitles", (*int)(nil)).
			Return(9, 1, errors, nil)

		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		reqBody := EnrichEpisodeTitlesRequest{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/enrich/episode-titles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.EnrichEpisodeTitles(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response EnrichEpisodeTitlesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 9, response.SuccessCount)
		assert.Equal(t, 1, response.FailureCount)
		assert.Len(t, response.Errors, 1)
		assert.Contains(t, response.Errors[0], "Anime ID 5")

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid Request Body", func(t *testing.T) {
		mockService := new(MockAnimeService)
		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/admin/enrich/episode-titles", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.EnrichEpisodeTitles(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminHandler_GetSyncStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)

	t.Run("Success - Get Sync Status", func(t *testing.T) {
		// Setup mock scheduler
		mockScheduler := new(MockSyncScheduler)
		mockStatus := models.SyncStatus{
			LastSyncSuccess: true,
			LastSyncCount:   95,
			TotalSyncsToday: 1,
			IsRunning:       false,
			Enabled:         true,
			CurrentSeason:   "2025-autumn",
			PreviousSeason:  "2025-summer",
		}
		mockScheduler.On("GetStatus").Return(mockStatus)

		mockService := new(MockAnimeService)
		handler := NewAdminHandler(mockService, logger, errorMonitor, mockScheduler)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/admin/sync/status", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.GetSyncStatus(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.SyncStatus
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.LastSyncSuccess)
		assert.Equal(t, 95, response.LastSyncCount)
		assert.Equal(t, "2025-autumn", response.CurrentSeason)

		mockScheduler.AssertExpectations(t)
	})

	t.Run("Error - Scheduler Not Available", func(t *testing.T) {
		mockService := new(MockAnimeService)
		handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/admin/sync/status", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Execute handler
		handler.GetSyncStatus(c)

		// Assert
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// MockSyncScheduler is a mock implementation of SyncScheduler
type MockSyncScheduler struct {
	mock.Mock
}

func (m *MockSyncScheduler) GetStatus() models.SyncStatus {
	args := m.Called()
	return args.Get(0).(models.SyncStatus)
}

func TestAdminHandler_NewAdminHandler(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)
	mockService := new(MockAnimeService)

	handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.BaseHandler)
	assert.NotNil(t, handler.animeService)
	assert.NotNil(t, handler.logger)
}
