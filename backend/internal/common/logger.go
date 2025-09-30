package common

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger provides structured logging functionality
type Logger struct {
	*slog.Logger
}

// LogLevel represents log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// NewLogger creates a new structured logger
func NewLogger(level LogLevel, format string) *Logger {
	var slogLevel slog.Level
	switch level {
	case LogLevelDebug:
		slogLevel = slog.LevelDebug
	case LogLevelInfo:
		slogLevel = slog.LevelInfo
	case LogLevelWarn:
		slogLevel = slog.LevelWarn
	case LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}

// WithUserID adds user ID to logger context
func (l *Logger) WithUserID(userID int) *Logger {
	return &Logger{
		Logger: l.Logger.With("user_id", userID),
	}
}

// WithError adds error to logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(c *gin.Context, duration time.Duration) {
	l.Logger.Info("HTTP Request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"status", c.Writer.Status(),
		"duration_ms", duration.Milliseconds(),
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"request_id", c.GetString("request_id"),
	)
}

// LogError logs error with context
func (l *Logger) LogError(ctx context.Context, message string, err error, fields ...interface{}) {
	args := []interface{}{"error", err.Error()}
	args = append(args, fields...)
	l.Logger.ErrorContext(ctx, message, args...)
}

// LogInfo logs info message with context
func (l *Logger) LogInfo(ctx context.Context, message string, fields ...interface{}) {
	l.Logger.InfoContext(ctx, message, fields...)
}

// LogWarn logs warning message with context
func (l *Logger) LogWarn(ctx context.Context, message string, fields ...interface{}) {
	l.Logger.WarnContext(ctx, message, fields...)
}

// LogDebug logs debug message with context
func (l *Logger) LogDebug(ctx context.Context, message string, fields ...interface{}) {
	l.Logger.DebugContext(ctx, message, fields...)
}

// Performance metrics logging
func (l *Logger) LogPerformance(operation string, duration time.Duration, fields ...interface{}) {
	args := []interface{}{
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	}
	args = append(args, fields...)
	l.Logger.Info("Performance Metric", args...)
}

// Database operation logging
func (l *Logger) LogDBOperation(operation string, query string, duration time.Duration, err error) {
	args := []interface{}{
		"operation", operation,
		"query", query,
		"duration_ms", duration.Milliseconds(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		l.Logger.Error("Database Operation Failed", args...)
	} else {
		l.Logger.Debug("Database Operation", args...)
	}
}

// External API call logging
func (l *Logger) LogExternalAPI(service string, endpoint string, method string, statusCode int, duration time.Duration, err error) {
	args := []interface{}{
		"service", service,
		"endpoint", endpoint,
		"method", method,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
	}

	if err != nil {
		args = append(args, "error", err.Error())
		l.Logger.Error("External API Call Failed", args...)
	} else {
		l.Logger.Info("External API Call", args...)
	}
}
