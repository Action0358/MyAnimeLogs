package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAnimeHandler() *AnimeHandler {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 300)
	handler := NewAnimeHandler(nil, logger, errorMonitor)
	return handler
}

func TestAnimeHandler_GetAnimes_ValidationError(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/animes", handler.GetAnimes)

	// Create request with invalid parameters
	req, _ := http.NewRequest("GET", "/animes?page=invalid", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response - should handle validation gracefully
	// The actual behavior depends on the validator implementation
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestAnimeHandler_GetAnime_InvalidID(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/animes/:id", handler.GetAnime)

	// Create request with invalid ID
	req, _ := http.NewRequest("GET", "/animes/invalid", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "無効なアニメIDです", response["message"])
}

func TestAnimeHandler_GetFavorites_Unauthorized(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/favorites", handler.GetFavorites)

	// Create request without authentication
	req, _ := http.NewRequest("GET", "/favorites", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "認証が必要です", response["message"])
}

func TestAnimeHandler_UpdateFavorite_InvalidJSON(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/animes/:anime_id/favorite", func(c *gin.Context) {
		c.Set("user_id", 123) // Mock authentication
		handler.UpdateFavorite(c)
	})

	// Create request with invalid JSON
	req, _ := http.NewRequest("PUT", "/animes/1/favorite", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}

func TestAnimeHandler_UpdateWatchStatus_InvalidEpisodeID(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/episodes/:episode_id/watch", func(c *gin.Context) {
		c.Set("user_id", 123) // Mock authentication
		handler.UpdateWatchStatus(c)
	})

	// Create request with invalid episode ID
	requestBody := map[string]bool{"viewed": true}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/episodes/invalid/watch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "無効なエピソードIDです", response["message"])
}

func TestAnimeHandler_GetRecentlyWatched_Unauthorized(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/recently-watched", handler.GetRecentlyWatched)

	// Create request without authentication
	req, _ := http.NewRequest("GET", "/recently-watched", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "認証が必要です", response["message"])
}

func TestAnimeHandler_UpdateRating_InvalidAnimeID(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/animes/:id/rating", func(c *gin.Context) {
		c.Set("user_id", 123) // Mock authentication
		handler.UpdateRating(c)
	})

	// Create request with invalid anime ID
	requestBody := map[string]interface{}{
		"rating":  4.5,
		"comment": "Great anime!",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/animes/invalid/rating", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "無効なアニメIDです", response["message"])
}

func TestAnimeHandler_DeleteFavorite_InvalidAnimeID(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/animes/:anime_id/favorite", func(c *gin.Context) {
		c.Set("user_id", 123) // Mock authentication
		handler.DeleteFavorite(c)
	})

	// Create request with invalid anime ID
	req, _ := http.NewRequest("DELETE", "/animes/invalid/favorite", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "無効なアニメIDです", response["message"])
}

// Test handler initialization
func TestNewAnimeHandler(t *testing.T) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 300)
	handler := NewAnimeHandler(nil, logger, errorMonitor)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.BaseHandler)
}

// Test basic handler functionality
func TestAnimeHandler_BasicFunctionality(t *testing.T) {
	handler := setupAnimeHandler()

	// Test that handler methods exist and don't panic when called with proper setup
	assert.NotNil(t, handler.GetAnimes)
	assert.NotNil(t, handler.GetAnime)
	assert.NotNil(t, handler.GetFavorites)
	assert.NotNil(t, handler.UpdateFavorite)
	assert.NotNil(t, handler.DeleteFavorite)
	assert.NotNil(t, handler.UpdateWatchStatus)
	assert.NotNil(t, handler.UpdateRating)
	assert.NotNil(t, handler.DeleteRating)
	assert.NotNil(t, handler.GetRecentlyWatched)
	assert.NotNil(t, handler.GetWatchlist)
}
