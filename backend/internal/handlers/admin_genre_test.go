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

	"myanimelogs/internal/common"
)

// MockAdminAnimeServiceForGenres is a mock implementation of AdminAnimeService for genre sync tests
type MockAdminAnimeServiceForGenres struct {
	mock.Mock
}

func (m *MockAdminAnimeServiceForGenres) SyncGenresForAllAnimes() (int, int, []string, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Get(2).([]string), args.Error(3)
}

func (m *MockAdminAnimeServiceForGenres) ClearAnimeCache() {
	m.Called()
}

func (m *MockAdminAnimeServiceForGenres) SyncAllAnimeStatuses(ctx context.Context) (int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockAdminAnimeServiceForGenres) SyncReleasedOnFromJikan() (int, int, int, error) {
	args := m.Called()
	return args.Int(0), args.Int(1), args.Int(2), args.Error(3)
}

func TestSyncGenresForAllAnimes_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAdminAnimeServiceForGenres)
	logger := common.NewLogger(common.LogLevelInfo, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)
	handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

	// Mock service response
	mockService.On("SyncGenresForAllAnimes").Return(95, 5, []string{"Error 1", "Error 2"}, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/admin/sync/genres", nil)

	// Execute
	handler.SyncGenresForAllAnimes(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response SyncGenresResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 95, response.SuccessCount)
	assert.Equal(t, 5, response.FailureCount)
	assert.Equal(t, 2, len(response.Errors))
	assert.Contains(t, response.Message, "95 succeeded")
	assert.Contains(t, response.Message, "5 failed")

	mockService.AssertExpectations(t)
}

func TestSyncGenresForAllAnimes_ServiceError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAdminAnimeServiceForGenres)
	logger := common.NewLogger(common.LogLevelInfo, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)
	handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

	// Mock service error
	mockService.On("SyncGenresForAllAnimes").Return(0, 0, []string{}, assert.AnError)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/admin/sync/genres", nil)

	// Execute
	handler.SyncGenresForAllAnimes(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestSyncGenresForAllAnimes_NoErrors(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAdminAnimeServiceForGenres)
	logger := common.NewLogger(common.LogLevelInfo, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)
	handler := NewAdminHandler(mockService, logger, errorMonitor, nil)

	// Mock service response with no errors
	mockService.On("SyncGenresForAllAnimes").Return(100, 0, []string{}, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/admin/sync/genres", nil)

	// Execute
	handler.SyncGenresForAllAnimes(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response SyncGenresResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 100, response.SuccessCount)
	assert.Equal(t, 0, response.FailureCount)
	assert.Empty(t, response.Errors)
	assert.Contains(t, response.Message, "100 succeeded")
	assert.Contains(t, response.Message, "0 failed")

	mockService.AssertExpectations(t)
}
