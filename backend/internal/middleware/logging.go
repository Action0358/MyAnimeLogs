package middleware

import (
	"time"

	"myanimelogs/internal/common"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides structured logging for requests
func LoggingMiddleware(logger *common.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request after processing
		duration := time.Since(start)

		// Get request ID from context
		requestID := c.GetString("request_id")

		// Create logger with request context
		contextLogger := logger.WithRequestID(requestID)

		// Add user ID if available
		if userID, exists := c.Get("user_id"); exists {
			contextLogger = contextLogger.WithUserID(userID.(int))
		}

		// Log the request
		contextLogger.Logger.Info("HTTP Request Completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"content_length", c.Request.ContentLength,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				contextLogger.WithError(err.Err).Logger.Error("Request Error",
					"error_type", err.Type,
					"error_meta", err.Meta,
				)
			}
		}
	}
}
