package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	JTI    string `json:"jti"` // JWT ID for session tracking
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token with 24 hour expiry (backward compatibility)
// Deprecated: Use GenerateAccessToken for new implementations
func GenerateJWT(userID int, email, name, secret string) (string, error) {
	jti := GenerateJTI()
	return GenerateAccessToken(userID, email, name, jti, secret, 24*time.Hour)
}

// GenerateAccessToken generates a short-lived access token (default: 15 minutes)
func GenerateAccessToken(userID int, email, name, jti, secret string, expiry time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MyAnimeLogs",
			Subject:   "MyAnimeLogs-users",
			Audience:  []string{"MyAnimeLogs-users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generates a long-lived refresh token (default: 7 days)
func GenerateRefreshToken(userID int, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "MyAnimeLogs",
		Subject:   "MyAnimeLogs-refresh",
		Audience:  []string{"MyAnimeLogs-users"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        GenerateJTI(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateJTI generates a unique JWT ID
func GenerateJTI() string {
	return uuid.New().String()
}

// ValidateJWT validates a JWT token (backward compatibility)
func ValidateJWT(tokenString, secret string) (*JWTClaims, error) {
	return ValidateAccessToken(tokenString, secret)
}

// ValidateAccessToken validates an access token
func ValidateAccessToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString, secret string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// Verify this is a refresh token
		if claims.Subject != "MyAnimeLogs-refresh" {
			return nil, errors.New("not a refresh token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
