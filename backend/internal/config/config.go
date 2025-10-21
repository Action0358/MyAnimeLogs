package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL  string
	Port         string
	Environment  string
	RateLimit    int
	JWTSecret    string
	AnnictAPIKey string
	RedisURL     string
	// Database connection pool settings
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime int // in minutes
	DBConnMaxIdleTime int // in minutes
	// Logging configuration
	LogLevel  string
	LogFormat string
	LogOutput string
	// Error monitoring configuration
	ErrorAlertThreshold int
	ErrorTimeWindow     int // in minutes
	// Upload configuration
	UploadDir     string
	MaxUploadSize int64 // in bytes
}

// New creates a new Config struct from environment variables
func New() *Config {
	env := GetEnvironment()

	// Database
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		databaseURL = buildDatabaseURL()
	}

	// Server
	port := getEnv("PORT", "8080")

	// Rate Limiter
	rateLimit := getEnvAsInt("RATE_LIMIT", 60) // 60 requests per minute

	// JWT
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key-change-me")

	// External APIs
	annictAPIKey := getEnv("ANNICT_API_KEY", "")

	// Cache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

	// Database connection pool settings (environment-specific defaults)
	dbConfig := getDBConfigForEnvironment(env)

	// Logging configuration
	logLevel := getEnv("LOG_LEVEL", "INFO")
	logFormat := getEnv("LOG_FORMAT", "json")
	logOutput := getEnv("LOG_OUTPUT", "stdout")

	// Error monitoring configuration
	errorAlertThreshold := getEnvAsInt("ERROR_ALERT_THRESHOLD", 10)
	errorTimeWindow := getEnvAsInt("ERROR_TIME_WINDOW", 5) // 5 minutes

	// Upload configuration
	uploadDir := getEnv("UPLOAD_DIR", "./uploads/icons")
	maxUploadSize := int64(getEnvAsInt("MAX_UPLOAD_SIZE", 5*1024*1024)) // 5MB default

	return &Config{
		DatabaseURL:         databaseURL,
		Port:                port,
		Environment:         string(env),
		RateLimit:           rateLimit,
		JWTSecret:           jwtSecret,
		AnnictAPIKey:        annictAPIKey,
		RedisURL:            redisURL,
		DBMaxOpenConns:      dbConfig.MaxOpenConns,
		DBMaxIdleConns:      dbConfig.MaxIdleConns,
		DBConnMaxLifetime:   dbConfig.ConnMaxLifetime,
		DBConnMaxIdleTime:   dbConfig.ConnMaxIdleTime,
		LogLevel:            logLevel,
		LogFormat:           logFormat,
		LogOutput:           logOutput,
		ErrorAlertThreshold: errorAlertThreshold,
		ErrorTimeWindow:     errorTimeWindow,
		UploadDir:           uploadDir,
		MaxUploadSize:       maxUploadSize,
	}
}

// buildDatabaseURL constructs database URL from individual components
func buildDatabaseURL() string {
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "user")
	dbPassword := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("POSTGRES_DB", "myanimelogs")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}

// getDBConfigForEnvironment returns database configuration based on environment
// Optimized for Supabase connection pooling
func getDBConfigForEnvironment(env Environment) struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
} {
	switch env {
	case Production:
		// Production settings optimized for Supabase
		return struct {
			MaxOpenConns    int
			MaxIdleConns    int
			ConnMaxLifetime int
			ConnMaxIdleTime int
		}{
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25), // Conservative for Supabase
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 15), // Longer for managed service
			ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 5),
		}
	case Staging:
		// Staging settings optimized for Supabase
		return struct {
			MaxOpenConns    int
			MaxIdleConns    int
			ConnMaxLifetime int
			ConnMaxIdleTime int
		}{
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 15),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 3),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 15),
			ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 5),
		}
	default: // Development, Test
		// Development settings for Supabase
		return struct {
			MaxOpenConns    int
			MaxIdleConns    int
			ConnMaxLifetime int
			ConnMaxIdleTime int
		}{
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 2),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 10),
			ConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 3),
		}
	}
}

// getEnv retrieves an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvAsInt retrieves an environment variable as an integer or returns a fallback value
func getEnvAsInt(key string, fallback int) int {
	if valueStr, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}
