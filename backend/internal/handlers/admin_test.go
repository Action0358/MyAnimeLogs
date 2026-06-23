package handlers

import (
	"context"
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

func (m *MockAnimeService) SyncGenresForAllAnimes() (int, int, []string, error) {
	args := m.Called()
	if args.Get(2) == nil {
		return args.Int(0), args.Int(1), nil, args.Error(3)
	}
	return args.Int(0), args.Int(1), args.Get(2).([]string), args.Error(3)
}

func (m *MockAnimeService) ClearAnimeCache() {
	m.Called()
}

func (m *MockAnimeService) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockAnimeService) SyncReleasedOnFromJikan() (int, int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Int(2), args.Error(3)
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

