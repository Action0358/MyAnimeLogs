package common

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		level  LogLevel
		format string
	}{
		{
			name:   "Info level with text format",
			level:  LogLevelInfo,
			format: "text",
		},
		{
			name:   "Debug level with json format",
			level:  LogLevelDebug,
			format: "json",
		},
		{
			name:   "Error level with text format",
			level:  LogLevelError,
			format: "text",
		},
		{
			name:   "Invalid level defaults to info",
			level:  LogLevel("INVALID"),
			format: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level, tt.format)
			require.NotNil(t, logger)
			require.NotNil(t, logger.Logger)
		})
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	logger := NewLogger(LogLevelInfo, "text")
	requestID := "test-request-123"

	loggerWithRequestID := logger.WithRequestID(requestID)

	require.NotNil(t, loggerWithRequestID)
	assert.NotEqual(t, logger, loggerWithRequestID)
}

func TestLogger_WithUserID(t *testing.T) {
	logger := NewLogger(LogLevelInfo, "text")
	userID := 12345

	loggerWithUserID := logger.WithUserID(userID)

	require.NotNil(t, loggerWithUserID)
	assert.NotEqual(t, logger, loggerWithUserID)
}

func TestLogger_WithError(t *testing.T) {
	logger := NewLogger(LogLevelInfo, "text")
	testError := errors.New("test error")

	loggerWithError := logger.WithError(testError)

	require.NotNil(t, loggerWithError)
	assert.NotEqual(t, logger, loggerWithError)
}

func TestLogger_LogError(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		err           error
		fields        []interface{}
		expectedInLog []string
	}{
		{
			name:    "Basic error logging",
			message: "Test error occurred",
			err:     errors.New("test error"),
			fields:  []interface{}{"key", "value"},
			expectedInLog: []string{
				"Test error occurred",
				"ERROR",
				"key=value",
				"test error",
			},
		},
		{
			name:    "Error logging without additional fields",
			message: "Simple error",
			err:     errors.New("simple test error"),
			fields:  nil,
			expectedInLog: []string{
				"Simple error",
				"ERROR",
				"simple test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := setupTestLogger(&buf)
			ctx := context.Background()

			logger.LogError(ctx, tt.message, tt.err, tt.fields...)

			output := buf.String()
			for _, expected := range tt.expectedInLog {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestLogger_LogInfo(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		fields        []interface{}
		expectedInLog []string
	}{
		{
			name:    "Info logging with fields",
			message: "Info message",
			fields:  []interface{}{"user_id", 123, "action", "login"},
			expectedInLog: []string{
				"Info message",
				"INFO",
				"user_id=123",
				"action=login",
			},
		},
		{
			name:    "Info logging without fields",
			message: "Simple info message",
			fields:  nil,
			expectedInLog: []string{
				"Simple info message",
				"INFO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := setupTestLogger(&buf)
			ctx := context.Background()

			logger.LogInfo(ctx, tt.message, tt.fields...)

			output := buf.String()
			for _, expected := range tt.expectedInLog {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestLogger_LogWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)
	ctx := context.Background()

	logger.LogWarn(ctx, "Warning message", "severity", "medium")

	output := buf.String()
	assert.Contains(t, output, "Warning message")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "severity=medium")
}

func TestLogger_LogDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)
	ctx := context.Background()

	logger.LogDebug(ctx, "Debug message", "debug_info", "test")

	output := buf.String()
	assert.Contains(t, output, "Debug message")
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "debug_info=test")
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      LogLevel
		logFunction   func(*Logger, context.Context, string)
		expectedLevel string
		shouldLog     bool
	}{
		{
			name:          "Debug level logs debug messages",
			logLevel:      LogLevelDebug,
			logFunction:   func(l *Logger, ctx context.Context, msg string) { l.LogDebug(ctx, msg) },
			expectedLevel: "DEBUG",
			shouldLog:     true,
		},
		{
			name:          "Info level does not log debug messages",
			logLevel:      LogLevelInfo,
			logFunction:   func(l *Logger, ctx context.Context, msg string) { l.LogDebug(ctx, msg) },
			expectedLevel: "DEBUG",
			shouldLog:     false,
		},
		{
			name:          "Error level logs error messages",
			logLevel:      LogLevelError,
			logFunction:   func(l *Logger, ctx context.Context, msg string) { l.LogInfo(ctx, msg) },
			expectedLevel: "INFO",
			shouldLog:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := &slog.HandlerOptions{
				Level:     getSlogLevel(tt.logLevel),
				AddSource: false,
			}
			handler := slog.NewTextHandler(&buf, opts)
			logger := &Logger{Logger: slog.New(handler)}
			ctx := context.Background()

			tt.logFunction(logger, ctx, "Test message")

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, "Test message")
				assert.Contains(t, output, tt.expectedLevel)
			} else {
				assert.Empty(t, strings.TrimSpace(output))
			}
		})
	}
}

// setupTestLogger creates a logger for testing with output captured in buffer
func setupTestLogger(buf *bytes.Buffer) *Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	}
	handler := slog.NewTextHandler(buf, opts)
	return &Logger{Logger: slog.New(handler)}
}

// getSlogLevel converts LogLevel to slog.Level
func getSlogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
