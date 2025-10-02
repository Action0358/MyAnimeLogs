package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	userID := 123
	email := "test@example.com"
	name := "Test User"

	// Generate valid token
	token, err := utils.GenerateJWT(userID, email, name, secret)
	require.NoError(t, err)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		// Verify user context is set
		contextUserID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)

		contextEmail, exists := c.Get("user_email")
		assert.True(t, exists)
		assert.Equal(t, email, contextEmail)

		contextName, exists := c.Get("user_name")
		assert.True(t, exists)
		assert.Equal(t, name, contextName)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	secret := "test-secret"

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Create request without Authorization header
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, common.ErrMsgUnauthorized, response.Error)
	assert.Equal(t, "認証トークンが必要です", response.Message)
	assert.Equal(t, common.ErrCodeUnauthorized, response.Code)
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name            string
		authHeader      string
		expectedMessage string
	}{
		{
			name:            "Missing Bearer prefix",
			authHeader:      "invalid-token",
			expectedMessage: "無効な認証トークン形式です",
		},
		{
			name:            "Wrong prefix",
			authHeader:      "Basic invalid-token",
			expectedMessage: "無効な認証トークン形式です",
		},
		{
			name:            "Empty token",
			authHeader:      "Bearer ",
			expectedMessage: "無効または期限切れのトークンです",
		},
		{
			name:            "Multiple spaces",
			authHeader:      "Bearer  token with spaces",
			expectedMessage: "無効な認証トークン形式です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(AuthMiddleware(secret))
			router.GET("/protected", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
			})

			// Create request with invalid token format
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response common.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, common.ErrMsgUnauthorized, response.Error)
			assert.Equal(t, tt.expectedMessage, response.Message)
			assert.Equal(t, common.ErrCodeUnauthorized, response.Code)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	secret := "test-secret"
	wrongSecret := "wrong-secret"

	// Generate token with wrong secret
	token, err := utils.GenerateJWT(123, "test@example.com", "Test User", wrongSecret)
	require.NoError(t, err)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, common.ErrMsgUnauthorized, response.Error)
	assert.Equal(t, "無効または期限切れのトークンです", response.Message)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	// Create expired token manually
	claims := utils.JWTClaims{
		UserID: 123,
		Email:  "test@example.com",
		Name:   "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MyAnimeLogs",
			Subject:   "MyAnimeLogs-users",
			Audience:  []string{"MyAnimeLogs-users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Create request with expired token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response common.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "無効または期限切れのトークンです", response.Message)
}

func TestOptionalAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	userID := 123
	email := "test@example.com"
	name := "Test User"

	// Generate valid token
	token, err := utils.GenerateJWT(userID, email, name, secret)
	require.NoError(t, err)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(OptionalAuthMiddleware(secret))
	router.GET("/optional", func(c *gin.Context) {
		// Verify user context is set
		contextUserID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthMiddleware_NoToken(t *testing.T) {
	secret := "test-secret"

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(OptionalAuthMiddleware(secret))
	router.GET("/optional", func(c *gin.Context) {
		// Verify user context is not set
		_, exists := c.Get("user_id")
		assert.False(t, exists)

		c.JSON(http.StatusOK, gin.H{"message": "success without auth"})
	})

	// Create request without token
	req, _ := http.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	secret := "test-secret"

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(OptionalAuthMiddleware(secret))
	router.GET("/optional", func(c *gin.Context) {
		// Verify user context is not set when token is invalid
		_, exists := c.Get("user_id")
		assert.False(t, exists)

		c.JSON(http.StatusOK, gin.H{"message": "success without auth"})
	})

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response - should continue without error
	assert.Equal(t, http.StatusOK, w.Code)
}
