package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation that embeds the real AuthService methods
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req models.UserRegisterRequest) (*models.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) Login(req models.UserLoginRequest) (*models.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) GetUserByID(userID int) (*models.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) GetUserStats(userID int) (*models.UserStats, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserStats), args.Error(1)
}

func (m *MockAuthService) ExportUserData(userID int, format string) (interface{}, error) {
	args := m.Called(userID, format)
	return args.Get(0), args.Error(1)
}

// TestableAuthHandler is a testable version of AuthHandler that accepts an interface
type TestableAuthHandler struct {
	*BaseHandler
	authService AuthServiceInterface
}

// AuthServiceInterface defines the interface that AuthService should implement
type AuthServiceInterface interface {
	Register(req models.UserRegisterRequest) (*models.UserResponse, error)
	Login(req models.UserLoginRequest) (*models.LoginResponse, error)
	GetUserByID(userID int) (*models.UserResponse, error)
	UpdateUser(userID int, req models.UserUpdateRequest) (*models.UserResponse, error)
	GetUserStats(userID int) (*models.UserStats, error)
	ExportUserData(userID int, format string) (interface{}, error)
}

// Implement all the handler methods for TestableAuthHandler
func (h *TestableAuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if err == services.ErrEmailAlreadyExists {
			h.response.ConflictError(c, "このメールアドレスは既に登録されています")
			return
		}
		h.response.InternalServerError(c, "ユーザー登録に失敗しました")
		return
	}

	h.response.Created(c, gin.H{
		"message": "ユーザー登録が完了しました。",
		"user":    user,
	})
}

func (h *TestableAuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	loginResp, err := h.authService.Login(req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			h.response.UnauthorizedError(c, "メールアドレスまたはパスワードが正しくありません")
			return
		}
		h.response.InternalServerError(c, "ログインに失敗しました")
		return
	}

	h.response.Success(c, loginResp)
}

func (h *TestableAuthHandler) GetMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.response.NotFoundError(c, "ユーザーが見つかりません")
		return
	}

	h.response.Success(c, user)
}

func (h *TestableAuthHandler) UpdateMe(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.response.ValidationError(c, err.Error())
		return
	}

	user, err := h.authService.UpdateUser(userID, req)
	if err != nil {
		h.response.InternalServerError(c, "プロフィール更新に失敗しました")
		return
	}

	h.response.Success(c, gin.H{
		"message": "プロフィールを更新しました。",
		"user":    user,
	})
}

func (h *TestableAuthHandler) GetStats(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	stats, err := h.authService.GetUserStats(userID)
	if err != nil {
		h.response.InternalServerError(c, "統計情報の取得に失敗しました")
		return
	}

	h.response.Success(c, stats)
}

func (h *TestableAuthHandler) ExportData(c *gin.Context) {
	userID, err := h.validator.GetUserIDFromContext(c)
	if err != nil {
		h.response.UnauthorizedError(c, "認証が必要です")
		return
	}

	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "csv" {
		h.response.ValidationError(c, "フォーマットはjsonまたはcsvを指定してください")
		return
	}

	data, err := h.authService.ExportUserData(userID, format)
	if err != nil {
		h.response.InternalServerError(c, "データエクスポートに失敗しました")
		return
	}

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=myanimelog_data.csv")

		if csvData, ok := data.(string); ok {
			c.String(http.StatusOK, csvData)
		} else {
			h.response.InternalServerError(c, "CSVデータの形式が正しくありません")
		}
	} else {
		h.response.Success(c, data)
	}
}

func setupAuthHandler() (*TestableAuthHandler, *MockAuthService) {
	logger := common.NewLogger(common.LogLevelDebug, "text")
	monitor := common.NewErrorMonitor(logger, 10, 5*time.Minute)
	mockService := &MockAuthService{}

	handler := &TestableAuthHandler{
		BaseHandler: NewBaseHandlerWithLogger(logger, monitor),
		authService: mockService,
	}

	return handler, mockService
}

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	// Setup mock expectations
	req := models.UserRegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUser := &models.UserResponse{
		ID:          1,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: "",
		Icon:        nil,
	}

	mockService.On("Register", req).Return(expectedUser, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Contains(t, response, "user")
	assert.Equal(t, "ユーザー登録が完了しました。", response["message"])

	user := response["user"].(map[string]interface{})
	assert.Equal(t, float64(1), user["id"])
	assert.Equal(t, "Test User", user["name"])
	assert.Equal(t, "test@example.com", user["email"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserRegisterRequest{
		Name:     "Test User",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockService.On("Register", req).Return(nil, services.ErrEmailAlreadyExists)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
	assert.Equal(t, "このメールアドレスは既に登録されています", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	handler, _ := setupAuthHandler()

	// Invalid request (missing required fields)
	req := map[string]interface{}{
		"name":     "",              // Empty name should fail validation
		"email":    "invalid-email", // Invalid email format
		"password": "123",           // Too short password
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &models.LoginResponse{
		AccessToken: "jwt-token-here",
		TokenType:   "Bearer",
		User: models.UserResponse{
			ID:          1,
			Name:        "Test User",
			Email:       "test@example.com",
			Description: "",
			Icon:        nil,
		},
	}

	mockService.On("Login", req).Return(expectedResponse, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/login", handler.Login)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "jwt-token-here", response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 1, response.User.ID)
	assert.Equal(t, "Test User", response.User.Name)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", req).Return(nil, services.ErrInvalidCredentials)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/login", handler.Login)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
	assert.Equal(t, "メールアドレスまたはパスワードが正しくありません", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	expectedUser := &models.UserResponse{
		ID:          userID,
		Name:        "Test User",
		Email:       "test@example.com",
		Description: "Test description",
		Icon:        nil,
	}

	mockService.On("GetUserByID", userID).Return(expectedUser, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/me", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", userID)
		handler.GetMe(c)
	})

	// Create request
	httpReq, _ := http.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, userID, response.ID)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "test@example.com", response.Email)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Unauthorized(t *testing.T) {
	handler, _ := setupAuthHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/me", handler.GetMe)

	// Create request without user context
	httpReq, _ := http.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
	assert.Equal(t, "認証が必要です", response["message"])
}

func TestAuthHandler_UpdateMe_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	name := "Updated Name"
	description := "Updated description"

	req := models.UserUpdateRequest{
		Name:        &name,
		Description: &description,
	}

	expectedUser := &models.UserResponse{
		ID:          userID,
		Name:        name,
		Email:       "test@example.com",
		Description: description,
		Icon:        nil,
	}

	mockService.On("UpdateUser", userID, req).Return(expectedUser, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/auth/me", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", userID)
		handler.UpdateMe(c)
	})

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("PUT", "/auth/me", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Contains(t, response, "user")
	assert.Equal(t, "プロフィールを更新しました。", response["message"])

	user := response["user"].(map[string]interface{})
	assert.Equal(t, float64(userID), user["id"])
	assert.Equal(t, name, user["name"])
	assert.Equal(t, description, user["description"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetStats_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	expectedStats := &models.UserStats{
		TotalEpisodesWatched: 150,
		TotalAnimesCompleted: 25,
		TotalWatchTimeHours:  60.0,
		FavoriteGenres:       []string{"Action", "Drama", "Comedy"},
		ThisMonthEpisodes:    12,
	}

	mockService.On("GetUserStats", userID).Return(expectedStats, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/stats", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", userID)
		handler.GetStats(c)
	})

	// Create request
	httpReq, _ := http.NewRequest("GET", "/auth/stats", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserStats
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 150, response.TotalEpisodesWatched)
	assert.Equal(t, 25, response.TotalAnimesCompleted)
	assert.Equal(t, 60.0, response.TotalWatchTimeHours)
	assert.Equal(t, []string{"Action", "Drama", "Comedy"}, response.FavoriteGenres)
	assert.Equal(t, 12, response.ThisMonthEpisodes)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ExportData_JSON_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	expectedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"id":   1,
			"name": "Test User",
		},
		"favorites":   []interface{}{},
		"exported_at": time.Now(),
	}

	mockService.On("ExportUserData", userID, "json").Return(expectedData, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/export", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", userID)
		handler.ExportData(c)
	})

	// Create request
	httpReq, _ := http.NewRequest("GET", "/auth/export?format=json", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "user_info")
	assert.Contains(t, response, "favorites")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ExportData_CSV_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	csvData := "type,id,data\nuser,1,\"Test User\"\n"

	mockService.On("ExportUserData", userID, "csv").Return(csvData, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/export", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", userID)
		handler.ExportData(c)
	})

	// Create request
	httpReq, _ := http.NewRequest("GET", "/auth/export?format=csv", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Equal(t, csvData, w.Body.String())

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ExportData_InvalidFormat(t *testing.T) {
	handler, _ := setupAuthHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/export", func(c *gin.Context) {
		// Simulate authenticated user context
		c.Set("user_id", 1)
		handler.ExportData(c)
	})

	// Create request with invalid format
	httpReq, _ := http.NewRequest("GET", "/auth/export?format=xml", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
	assert.Equal(t, "フォーマットはjsonまたはcsvを指定してください", response["message"])
}

func TestAuthHandler_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserRegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock service returns internal error
	mockService.On("Register", req).Return(nil, errors.New("database connection failed"))

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
	assert.Equal(t, "ユーザー登録に失敗しました", response["message"])

	mockService.AssertExpectations(t)
}
