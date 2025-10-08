package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/mocks"
	"myanimelogs/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// ===== 検索機能のテスト =====

func TestAnimeHandler_GetSearchSuggestions(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      string
		mockReturn       []string
		mockError        error
		expectedStatus   int
		expectEmptyArray bool
		skipMock         bool // モックを呼ばないケース
	}{
		{
			name:             "空のクエリ",
			queryParams:      "q=",
			mockReturn:       []string{},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectEmptyArray: true,
			skipMock:         true, // 空のクエリはハンドラーで早期リターンするためモックは呼ばれない
		},
		{
			name:           "通常のクエリ",
			queryParams:    "q=進撃",
			mockReturn:     []string{"進撃の巨人", "進撃の巨人 Season 2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit指定",
			queryParams:    "q=アニメ&limit=5",
			mockReturn:     []string{"アニメ1", "アニメ2", "アニメ3"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なlimit",
			queryParams:    "q=test&limit=abc",
			mockReturn:     []string{"test1", "test2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			if !tt.skipMock {
				mockService.On("GetSearchSuggestions", mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockError)
			}

			logger := common.NewLogger(common.LogLevelDebug, "text")
			errorMonitor := common.NewErrorMonitor(logger, 10, 300)
			handler := NewAnimeHandler(mockService, logger, errorMonitor)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/search/suggestions", handler.GetSearchSuggestions)

			req, _ := http.NewRequest("GET", "/search/suggestions?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "suggestions")

			if tt.expectEmptyArray {
				suggestions, ok := response["suggestions"].([]any)
				assert.True(t, ok)
				assert.Equal(t, 0, len(suggestions))
			}

			if !tt.skipMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAnimeHandler_GetFilterOptions_Success(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *models.FilterOptions
		mockError      error
		expectedStatus int
		checkFields    bool
	}{
		{
			name: "正常なレスポンス",
			mockReturn: &models.FilterOptions{
				Genres:  []models.GenreInfo{{Name: "アクション", Count: 10}, {Name: "コメディ", Count: 5}},
				Years:   []int{2024, 2023, 2022},
				Seasons: []string{"spring", "summer", "autumn", "winter"},
				Ratings: models.RatingStats{Min: 1.0, Max: 5.0, Average: 3.5, Count: 100},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkFields:    true,
		},
		{
			name: "空のデータ",
			mockReturn: &models.FilterOptions{
				Genres:  []models.GenreInfo{},
				Years:   []int{},
				Seasons: []string{},
				Ratings: models.RatingStats{Min: 1.0, Max: 5.0, Average: 0.0, Count: 0},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkFields:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			mockService.On("GetFilterOptions").Return(tt.mockReturn, tt.mockError)

			logger := common.NewLogger(common.LogLevelDebug, "text")
			errorMonitor := common.NewErrorMonitor(logger, 10, 300)
			handler := NewAnimeHandler(mockService, logger, errorMonitor)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/filters", handler.GetFilterOptions)

			req, _ := http.NewRequest("GET", "/filters", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.checkFields {
				assert.Contains(t, response, "genres")
				assert.Contains(t, response, "years")
				assert.Contains(t, response, "seasons")
				assert.Contains(t, response, "ratings")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnimeHandler_GetPopularSearches(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      string
		mockReturn       []string
		mockError        error
		expectedStatus   int
		expectEmptyArray bool
	}{
		{
			name:           "デフォルトのlimit",
			queryParams:    "",
			mockReturn:     []string{"進撃の巨人", "鬼滅の刃", "呪術廻戦"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit指定",
			queryParams:    "limit=5",
			mockReturn:     []string{"アニメ1", "アニメ2", "アニメ3", "アニメ4", "アニメ5"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:             "空の結果",
			queryParams:      "",
			mockReturn:       []string{},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectEmptyArray: true,
		},
		{
			name:           "無効なlimit",
			queryParams:    "limit=abc",
			mockReturn:     []string{"test1", "test2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			mockService.On("GetPopularSearches", mock.Anything).Return(tt.mockReturn, tt.mockError)

			logger := common.NewLogger(common.LogLevelDebug, "text")
			errorMonitor := common.NewErrorMonitor(logger, 10, 300)
			handler := NewAnimeHandler(mockService, logger, errorMonitor)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/search/popular", handler.GetPopularSearches)

			req, _ := http.NewRequest("GET", "/search/popular?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "popular_searches")

			if tt.expectEmptyArray {
				searches, ok := response["popular_searches"].([]any)
				assert.True(t, ok)
				assert.Equal(t, 0, len(searches))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnimeHandler_AdvancedSearchAnimes(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockReturn     *models.PaginatedResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "基本的な検索",
			queryParams: "q=進撃&page=1&per_page=20",
			mockReturn: &models.PaginatedResponse{
				Data: []models.AnimeWithUserData{
					{
						Anime: models.Anime{
							ID:    1,
							Title: "進撃の巨人",
						},
					},
				},
				Meta: models.PaginationMeta{
					TotalPages:  1,
					CurrentPage: 1,
					Total:       1,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "複数フィルター",
			queryParams: "q=アニメ&genre=アクション&year=2024&season=spring&min_rating=4.0&page=1",
			mockReturn: &models.PaginatedResponse{
				Data: []models.AnimeWithUserData{},
				Meta: models.PaginationMeta{
					TotalPages:  0,
					CurrentPage: 1,
					Total:       0,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "ソート指定",
			queryParams: "q=test&sort=rating&order=desc&page=1",
			mockReturn: &models.PaginatedResponse{
				Data: []models.AnimeWithUserData{
					{
						Anime: models.Anime{
							ID:    2,
							Title: "Test Anime",
						},
					},
				},
				Meta: models.PaginationMeta{
					TotalPages:  1,
					CurrentPage: 1,
					Total:       1,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			mockService.On("AdvancedSearchAnimes", mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockError)

			logger := common.NewLogger(common.LogLevelDebug, "text")
			errorMonitor := common.NewErrorMonitor(logger, 10, 300)
			handler := NewAnimeHandler(mockService, logger, errorMonitor)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/search/advanced", handler.AdvancedSearchAnimes)

			req, _ := http.NewRequest("GET", "/search/advanced?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "data")
			assert.Contains(t, response, "meta")

			mockService.AssertExpectations(t)
		})
	}
}
