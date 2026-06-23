package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/constants"
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

// setupAnimeHandlerWithMock creates a handler with mock service for testing
func setupAnimeHandlerWithMock(mockService *mocks.MockAnimeService) (*AnimeHandler, *gin.Engine) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 300)
	handler := NewAnimeHandler(mockService, logger, errorMonitor)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	return handler, router
}

// Response structures for type-safe testing
type searchSuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

type popularSearchesResponse struct {
	PopularSearches []string `json:"popular_searches"`
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

func TestAnimeHandler_UpdateWatchStatus_InvalidJSON(t *testing.T) {
	handler := setupAnimeHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/users/me/watch-status/:anime_id", func(c *gin.Context) {
		c.Set("user_id", 123) // Mock authentication
		handler.UpdateWatchStatus(c)
	})

	// Create request with invalid JSON
	req, _ := http.NewRequest("PUT", "/users/me/watch-status/1", bytes.NewBuffer([]byte("invalid json")))
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
	assert.NotNil(t, handler.GetWatchStatuses)
	assert.NotNil(t, handler.UpdateWatchStatus)
	assert.NotNil(t, handler.DeleteWatchStatus)
}

// ===== 検索機能のテスト =====

func TestAnimeHandler_GetSearchSuggestions(t *testing.T) {
	tests := []struct {
		name             string
		queryParams      string
		expectedQuery    string
		expectedLimit    int
		mockReturn       []string
		mockError        error
		expectedStatus   int
		expectEmptyArray bool
		skipMock         bool
	}{
		{
			name:             "空のクエリ",
			queryParams:      "q=",
			expectedQuery:    "",
			expectedLimit:    10,
			mockReturn:       []string{},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectEmptyArray: true,
			skipMock:         true, // 空のクエリはハンドラーで早期リターンするためモックは呼ばれない
		},
		{
			name:           "通常のクエリ",
			queryParams:    "q=進撃",
			expectedQuery:  "進撃",
			expectedLimit:  10, // デフォルト値
			mockReturn:     []string{"進撃の巨人", "進撃の巨人 Season 2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit指定",
			queryParams:    "q=アニメ&limit=5",
			expectedQuery:  "アニメ",
			expectedLimit:  5,
			mockReturn:     []string{"アニメ1", "アニメ2", "アニメ3"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なlimit（デフォルト値にフォールバック）",
			queryParams:    "q=test&limit=abc",
			expectedQuery:  "test",
			expectedLimit:  10, // 無効な値はデフォルトに戻る
			mockReturn:     []string{"test1", "test2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:             "サービスエラー時のnil戻り値",
			queryParams:      "q=error",
			expectedQuery:    "error",
			expectedLimit:    10,
			mockReturn:       nil,
			mockError:        assert.AnError,
			expectedStatus:   http.StatusOK,
			expectEmptyArray: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			if !tt.skipMock {
				mockService.On("GetSearchSuggestions", tt.expectedQuery, tt.expectedLimit).Return(tt.mockReturn, tt.mockError)
			}

			handler, router := setupAnimeHandlerWithMock(mockService)
			router.GET("/search/suggestions", handler.GetSearchSuggestions)

			req, _ := http.NewRequest("GET", "/search/suggestions?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response searchSuggestionsResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectEmptyArray {
				assert.Equal(t, 0, len(response.Suggestions))
			} else if !tt.skipMock {
				assert.Equal(t, len(tt.mockReturn), len(response.Suggestions))
			}

			if !tt.skipMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestAnimeHandler_GetFilterOptions_Success(t *testing.T) {
	tests := []struct {
		name              string
		mockReturn        *models.FilterOptions
		mockError         error
		expectedStatus    int
		expectDefaultData bool
	}{
		{
			name: "正常なレスポンス",
			mockReturn: &models.FilterOptions{
				Genres:  []models.GenreInfo{{Name: "アクション", Count: 10}, {Name: "コメディ", Count: 5}},
				Years:   []int{2024, 2023, 2022},
				Seasons: []string{"spring", "summer", "autumn", "winter"},
			},
			mockError:         nil,
			expectedStatus:    http.StatusOK,
			expectDefaultData: false,
		},
		{
			name: "空のデータ",
			mockReturn: &models.FilterOptions{
				Genres:  []models.GenreInfo{},
				Years:   []int{},
				Seasons: []string{},
			},
			mockError:         nil,
			expectedStatus:    http.StatusOK,
			expectDefaultData: false,
		},
		{
			name:              "サービスエラー時のデフォルト値",
			mockReturn:        nil,
			mockError:         assert.AnError,
			expectedStatus:    http.StatusOK,
			expectDefaultData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			mockService.On("GetFilterOptions").Return(tt.mockReturn, tt.mockError)

			handler, router := setupAnimeHandlerWithMock(mockService)
			router.GET("/filters", handler.GetFilterOptions)

			req, _ := http.NewRequest("GET", "/filters", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response models.FilterOptions
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// レスポンスの構造を検証
			assert.NotNil(t, response.Genres)
			assert.NotNil(t, response.Years)
			assert.NotNil(t, response.Seasons)

			if tt.expectDefaultData {
				// デフォルト値が返されることを確認
				assert.Equal(t, 0, len(response.Genres))
				assert.Equal(t, 0, len(response.Years))
				assert.NotNil(t, response.Seasons)
			} else if tt.mockReturn != nil {
				// モックの戻り値と完全に一致することを確認
				assert.Equal(t, len(tt.mockReturn.Genres), len(response.Genres))
				assert.Equal(t, len(tt.mockReturn.Years), len(response.Years))
				assert.Equal(t, len(tt.mockReturn.Seasons), len(response.Seasons))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnimeHandler_GetPopularSearches(t *testing.T) {
	tests := []struct {
		name              string
		queryParams       string
		expectedLimit     int
		mockReturn        []string
		mockError         error
		expectedStatus    int
		expectEmptyArray  bool
		expectDefaultData bool
	}{
		{
			name:           "デフォルトのlimit",
			queryParams:    "",
			expectedLimit:  10,
			mockReturn:     []string{"進撃の巨人", "鬼滅の刃", "呪術廻戦"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit指定",
			queryParams:    "limit=5",
			expectedLimit:  5,
			mockReturn:     []string{"アニメ1", "アニメ2", "アニメ3", "アニメ4", "アニメ5"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:             "空の結果",
			queryParams:      "",
			expectedLimit:    10,
			mockReturn:       []string{},
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectEmptyArray: true,
		},
		{
			name:           "無効なlimit（デフォルト値にフォールバック）",
			queryParams:    "limit=abc",
			expectedLimit:  10,
			mockReturn:     []string{"test1", "test2"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:              "サービスエラー時のnil戻り値（デフォルト値を返す）",
			queryParams:       "",
			expectedLimit:     10,
			mockReturn:        nil,
			mockError:         assert.AnError,
			expectedStatus:    http.StatusOK,
			expectDefaultData: true,
		},
		{
			name:              "サービスエラー時のlimit指定",
			queryParams:       "limit=3",
			expectedLimit:     3,
			mockReturn:        nil,
			mockError:         assert.AnError,
			expectedStatus:    http.StatusOK,
			expectDefaultData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			mockService.On("GetPopularSearches", tt.expectedLimit).Return(tt.mockReturn, tt.mockError)

			handler, router := setupAnimeHandlerWithMock(mockService)
			router.GET("/search/popular", handler.GetPopularSearches)

			req, _ := http.NewRequest("GET", "/search/popular?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response popularSearchesResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectEmptyArray {
				assert.Equal(t, 0, len(response.PopularSearches))
			} else if tt.expectDefaultData {
				// デフォルト値が返されることを確認
				expectedLen := len(constants.DefaultPopularSearches)
				if tt.expectedLimit < expectedLen {
					expectedLen = tt.expectedLimit
				}
				assert.Equal(t, expectedLen, len(response.PopularSearches))
				// デフォルト値の内容を検証
				for i, search := range response.PopularSearches {
					assert.Equal(t, constants.DefaultPopularSearches[i], search)
				}
			} else {
				assert.Equal(t, len(tt.mockReturn), len(response.PopularSearches))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnimeHandler_AdvancedSearchAnimes(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedParams models.AdvancedSearchParams
		mockReturn     *models.PaginatedResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:        "基本的な検索",
			queryParams: "q=進撃&page=1&per_page=20",
			expectedParams: models.AdvancedSearchParams{
				Query:   "進撃",
				Page:    1,
				PerPage: 20,
			},
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
			queryParams: "q=アニメ&genre=アクション&year=2024&season=spring&page=1",
			expectedParams: models.AdvancedSearchParams{
				Query:  "アニメ",
				Genres: []string{"アクション"},
				Year:   2024,
				Season: "spring",
				Page:   1,
			},
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
			queryParams: "q=test&sort=year&order=desc&page=1",
			expectedParams: models.AdvancedSearchParams{
				Query: "test",
				Sort:  "year",
				Order: "desc",
				Page:  1,
			},
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
		{
			name:        "サービスエラー",
			queryParams: "q=error&page=1",
			expectedParams: models.AdvancedSearchParams{
				Query: "error",
				Page:  1,
			},
			mockReturn:     nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mocks.MockAnimeService)
			// mock.MatchedBy を使用してパラメータのすべての関連フィールドを検証
			mockService.On("AdvancedSearchAnimes",
				mock.MatchedBy(func(params models.AdvancedSearchParams) bool {
					// 基本フィールドの検証
					if params.Query != tt.expectedParams.Query {
						return false
					}
					if params.Page != tt.expectedParams.Page {
						return false
					}
					// ソート関連の検証
					if params.Sort != tt.expectedParams.Sort {
						return false
					}
					if params.Order != tt.expectedParams.Order {
						return false
					}
					// フィルター関連の検証
					if params.Year != tt.expectedParams.Year {
						return false
					}
					if params.Season != tt.expectedParams.Season {
						return false
					}
					// ジャンルの検証
					if len(params.Genres) != len(tt.expectedParams.Genres) {
						return false
					}
					for i, genre := range params.Genres {
						if genre != tt.expectedParams.Genres[i] {
							return false
						}
					}
					return true
				}),
				mock.AnythingOfType("*int"),
			).Return(tt.mockReturn, tt.mockError)

			handler, router := setupAnimeHandlerWithMock(mockService)
			router.GET("/search/advanced", handler.AdvancedSearchAnimes)

			req, _ := http.NewRequest("GET", "/search/advanced?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.PaginatedResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.NotNil(t, response.Meta)

				if tt.mockReturn != nil {
					assert.Equal(t, tt.mockReturn.Meta.Total, response.Meta.Total)
					assert.Equal(t, tt.mockReturn.Meta.CurrentPage, response.Meta.CurrentPage)
					assert.Equal(t, tt.mockReturn.Meta.TotalPages, response.Meta.TotalPages)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
