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
}

// New creates a new Config struct from environment variables
func New() *Config {
	// Database
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "user")
	dbPassword := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("POSTGRES_DB", "myanimelogs")
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

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

	return &Config{
		DatabaseURL:  databaseURL,
		Port:         port,
		Environment:  environment,
		RateLimit:    rateLimit,
		JWTSecret:    jwtSecret,
		AnnictAPIKey: annictAPIKey,
		RedisURL:     redisURL,
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
