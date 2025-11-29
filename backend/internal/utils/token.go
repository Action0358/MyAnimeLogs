package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenerateRefreshTokenPair generates a refresh token and its hash
// Returns: (token, hash, error)
// The token should be sent to the client, and the hash should be stored in the database
func GenerateRefreshTokenPair() (string, string, error) {
	// Generate 32 bytes of random data
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 for the client
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token for database storage
	hash := HashRefreshToken(token)

	return token, hash, nil
}

// HashRefreshToken hashes a refresh token using SHA-256
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
