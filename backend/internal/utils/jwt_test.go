package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		testName string
		userID   int
		email    string
		userName string
		secret   string
	}{
		{
			testName: "Valid user data",
			userID:   123,
			email:    "test@example.com",
			userName: "Test User",
			secret:   "test-secret",
		},
		{
			testName: "Empty name",
			userID:   456,
			email:    "user@example.com",
			userName: "",
			secret:   "test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.email, tt.userName, tt.secret)

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed
			claims, err := ValidateJWT(token, tt.secret)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.email, claims.Email)
			assert.Equal(t, tt.userName, claims.Name)
		})
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	secret := "test-secret"
	userID := 123
	email := "test@example.com"
	userName := "Test User"

	token, err := GenerateJWT(userID, email, userName, secret)
	require.NoError(t, err)

	claims, err := ValidateJWT(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userName, claims.Name)
	assert.Equal(t, "MyAnimeLogs", claims.Issuer)
}

func TestValidateJWT_InvalidSecret(t *testing.T) {
	secret := "test-secret"
	wrongSecret := "wrong-secret"

	token, err := GenerateJWT(123, "test@example.com", "Test User", secret)
	require.NoError(t, err)

	_, err = ValidateJWT(token, wrongSecret)
	assert.Error(t, err)
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	secret := "test-secret"
	malformedToken := "invalid.token.format"

	_, err := ValidateJWT(malformedToken, secret)
	assert.Error(t, err)
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	// Create expired token manually
	claims := JWTClaims{
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

	_, err = ValidateJWT(tokenString, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	secret := "test-secret"

	_, err := ValidateJWT("", secret)
	assert.Error(t, err)
}

func TestJWTClaims_Structure(t *testing.T) {
	secret := "test-secret"
	userID := 123
	email := "test@example.com"
	userName := "Test User"

	token, err := GenerateJWT(userID, email, userName, secret)
	require.NoError(t, err)

	claims, err := ValidateJWT(token, secret)
	require.NoError(t, err)

	// Verify all required fields are present
	assert.NotZero(t, claims.UserID)
	assert.NotEmpty(t, claims.Email)
	assert.NotEmpty(t, claims.Issuer)
	assert.NotEmpty(t, claims.Subject)
	assert.NotEmpty(t, claims.Audience)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}
