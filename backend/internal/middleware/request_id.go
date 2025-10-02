package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware generates and adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")

		// Generate new request ID if not provided
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Failure to read from crypto/rand indicates a serious system problem
		// with the entropy source. The application should not continue.
		// Wrap the error to preserve structure before converting to string for panic
		wrappedErr := fmt.Errorf("failed to generate random bytes for request ID: %w", err)
		panic(wrappedErr.Error())
	}
	return hex.EncodeToString(bytes)
}
