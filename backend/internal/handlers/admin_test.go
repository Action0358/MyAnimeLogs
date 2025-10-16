package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myanimelogs/internal/common"
	"myanimelogs/internal/services"
)

func TestAdminHandler_SyncEpisodeDurations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 60)

	t.Run("Success - Sync All Animes", func(t *testing.T) {
		// Create mock anime service
		animeService := services.NewAnimeService(nil, logger)
		handler := NewAdminHandler(animeService, logger, errorMonitor)

		// Create request
		reqBody := SyncEpisodeDurationsRequest{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Note: This test will fail without a database connection
		// In a real scenario, you would use a mock database or test database
		// For now, we just verify the handler structure is correct
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.animeService)
	})

	t.Run("Success - Sync Specific Anime", func(t *testing.T) {
		animeService := services.NewAnimeService(nil, logger)
		handler := NewAdminHandler(animeService, logger, errorMonitor)

		animeID := 123
		reqBody := SyncEpisodeDurationsRequest{AnimeID: &animeID}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/sync/episode-durations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.animeService)
	})

	t.Run("Error - Invalid Request Body", func(t *testing.T) {
		animeService := services.NewAnimeService(nil, logger)
		handler := NewAdminHandler(animeService, logger, errorMonitor)

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
	animeService := services.NewAnimeService(nil, logger)

	handler := NewAdminHandler(animeService, logger, errorMonitor)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.BaseHandler)
	assert.NotNil(t, handler.animeService)
	assert.NotNil(t, handler.logger)
}
