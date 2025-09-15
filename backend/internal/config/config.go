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
}

// New creates a new Config struct from environment variables
func New() *Config {
	// Database
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		dbHost := getEnv("POSTGRES_HOST", "localhost")
		dbPort := getEnv("POSTGRES_PORT", "5432")
		dbUser := getEnv("POSTGRES_USER", "user")
		dbPassword := getEnv("POSTGRES_PASSWORD", "password")
		dbName := getEnv("POSTGRES_DB", "myanimelogs")
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// Server
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")

	// Rate Limiter
	rateLimit := getEnvAsInt("RATE_LIMIT", 60) // 60 requests per minute

	// JWT
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key-change-me")

	// External APIs
	annictAPIKey := getEnv("ANNICT_API_KEY", "")

	// Cache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

	// Database connection pool settings (environment-specific defaults)
	var dbMaxOpenConns, dbMaxIdleConns, dbConnMaxLifetime, dbConnMaxIdleTime int
	switch environment {
	case "production":
		dbMaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 50)
		dbMaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
		dbConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 10) // 10 minutes
		dbConnMaxIdleTime = getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 2) // 2 minutes
	case "staging":
		dbMaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 30)
		dbMaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 8)
		dbConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 8)
		dbConnMaxIdleTime = getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 2)
	default: // development, test
		dbMaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
		dbMaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 5)
		dbConnMaxLifetime = getEnvAsInt("DB_CONN_MAX_LIFETIME", 5)
		dbConnMaxIdleTime = getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 1)
	}

	return &Config{
		DatabaseURL:       databaseURL,
		Port:              port,
		Environment:       environment,
		RateLimit:         rateLimit,
		JWTSecret:         jwtSecret,
		AnnictAPIKey:      annictAPIKey,
		RedisURL:          redisURL,
		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: dbConnMaxLifetime,
		DBConnMaxIdleTime: dbConnMaxIdleTime,
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
