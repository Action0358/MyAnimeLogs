package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"myanimelogs/internal/common"
	"myanimelogs/internal/models"
	"myanimelogs/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService implements services.AuthServiceInterface for testing
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

func (m *MockAuthService) UpdateUserIcon(userID int, iconURL string) (*models.UserResponse, error) {
	args := m.Called(userID, iconURL)
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

// setupAuthHandler creates a test handler with mock service
func setupAuthHandler() (*AuthHandler, *MockAuthService) {
	mockService := &MockAuthService{}
	logger := common.NewLogger(common.LogLevelDebug, "text")
	errorMonitor := common.NewErrorMonitor(logger, 10, 300)
	handler := NewAuthHandler(mockService, logger, errorMonitor)
	return handler, mockService
}

func TestAuthHandler_Register_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	expectedUser := &models.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	req := models.UserRegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockService.On("Register", req).Return(expectedUser, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
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

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "このメールアドレスは既に登録されています", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	handler, _ := setupAuthHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request with invalid JSON
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &models.LoginResponse{
		AccessToken: "mock-jwt-token",
		TokenType:   "Bearer",
		User: models.UserResponse{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	mockService.On("Login", req).Return(expectedResponse, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/login", handler.Login)

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "mock-jwt-token", response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 1, response.User.ID)

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

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	expectedUser := &models.UserResponse{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
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
	assert.Contains(t, response, "message")
}

func TestAuthHandler_UpdateMe_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	newName := "Updated Name"
	req := models.UserUpdateRequest{
		Name: &newName,
	}

	expectedUser := &models.UserResponse{
		ID:   userID,
		Name: newName,
	}

	mockService.On("UpdateUser", userID, req).Return(expectedUser, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/auth/me", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateMe(c)
	})

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("PUT", "/auth/me", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetStats_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	expectedStats := &models.UserStats{
		TotalEpisodesWatched: 100,
		TotalAnimesCompleted: 10,
		TotalWatchTimeHours:  40.0,
		FavoriteGenres:       []string{"Action", "Comedy"},
		ThisMonthEpisodes:    15,
	}

	mockService.On("GetUserStats", userID).Return(expectedStats, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/stats", func(c *gin.Context) {
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
	assert.Equal(t, 100, response.TotalEpisodesWatched)
	assert.Equal(t, 10, response.TotalAnimesCompleted)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ExportData_JSON_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userID := 1
	format := "json"
	expectedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"id":   1,
			"name": "Test User",
		},
		"favorites": []interface{}{},
	}

	mockService.On("ExportUserData", userID, format).Return(expectedData, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/export", func(c *gin.Context) {
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

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ExportData_InvalidFormat(t *testing.T) {
	handler, _ := setupAuthHandler()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/export", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.ExportData(c)
	})

	// Create request with invalid format
	httpReq, _ := http.NewRequest("GET", "/auth/export?format=invalid", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
}

func TestAuthHandler_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	req := models.UserRegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock internal server error
	mockService.On("Register", req).Return(nil, errors.New("database error"))

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", handler.Register)

	// Create request body
	jsonBody, _ := json.Marshal(req)

	// Create request
	httpReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, httpReq)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Equal(t, "ユーザー登録に失敗しました", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteIcon_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	// Create a real upload service with test directory
	logger := common.NewLogger(common.LogLevelInfo, "text")
	uploadDir := "./test_uploads_delete_icon"
	uploadService := services.NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)
	handler.uploadService = uploadService

	// Clean up test directory
	defer func() {
		_ = os.RemoveAll(uploadDir)
	}()

	iconURL := "/uploads/icons/user_1_icon.jpg"
	mockUser := &models.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Icon:  &iconURL,
	}

	updatedUser := &models.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Icon:  nil,
	}

	mockService.On("GetUserByID", 1).Return(mockUser, nil)
	mockService.On("UpdateUserIcon", 1, "").Return(updatedUser, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/api/users/me/icon", nil)

	handler.DeleteIcon(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "プロフィール画像を削除しました", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteIcon_NoIconSet(t *testing.T) {
	handler, mockService := setupAuthHandler()

	// Create a real upload service with test directory
	logger := common.NewLogger(common.LogLevelInfo, "text")
	uploadDir := "./test_uploads_delete_icon_no_icon"
	uploadService := services.NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)
	handler.uploadService = uploadService

	// Clean up test directory
	defer func() {
		_ = os.RemoveAll(uploadDir)
	}()

	mockUser := &models.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Icon:  nil,
	}

	mockService.On("GetUserByID", 1).Return(mockUser, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/api/users/me/icon", nil)

	handler.DeleteIcon(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	// ValidationErrorはデフォルトメッセージを返すので、ステータスコードで確認
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteIcon_Unauthorized(t *testing.T) {
	handler, _ := setupAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// user_idを設定しない（認証なし）
	c.Request = httptest.NewRequest("DELETE", "/api/users/me/icon", nil)

	handler.DeleteIcon(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "認証が必要です")
}

func TestAuthHandler_DeleteIcon_GetUserError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	// Create a real upload service with test directory
	logger := common.NewLogger(common.LogLevelInfo, "text")
	uploadDir := "./test_uploads_delete_icon_error"
	uploadService := services.NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)
	handler.uploadService = uploadService

	// Clean up test directory
	defer func() {
		_ = os.RemoveAll(uploadDir)
	}()

	mockService.On("GetUserByID", 1).Return(nil, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/api/users/me/icon", nil)

	handler.DeleteIcon(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteIcon_UpdateError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	// Create a real upload service with test directory
	logger := common.NewLogger(common.LogLevelInfo, "text")
	uploadDir := "./test_uploads_delete_icon_update_error"
	uploadService := services.NewUploadService(uploadDir, "/uploads", 5*1024*1024, logger)
	handler.uploadService = uploadService

	// Clean up test directory
	defer func() {
		_ = os.RemoveAll(uploadDir)
	}()

	iconURL := "/uploads/icons/user_1_icon.jpg"
	mockUser := &models.UserResponse{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Icon:  &iconURL,
	}

	mockService.On("GetUserByID", 1).Return(mockUser, nil)
	mockService.On("UpdateUserIcon", 1, "").Return(nil, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest("DELETE", "/api/users/me/icon", nil)

	handler.DeleteIcon(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}
