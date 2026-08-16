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
	// Supabase Auth設定
	SupabaseJWTSecret string
	// SupabaseURL / SupabaseServiceRoleKey はアカウント削除で Admin API を
	// 呼ぶために使う。service_role キーは RLS を無視する強い権限を持つため、
	// ログや API レスポンスに絶対に含めないこと。
	SupabaseURL            string
	SupabaseServiceRoleKey string
	AnnictAPIKey           string
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
	UploadDir        string
	UploadPublicPath string // Public URL path for uploaded files (e.g., "/uploads")
	MaxUploadSize    int64  // in bytes
	// Scheduler configuration
	AnimeSyncEnabled   bool
	SchedulerRunHour   int
	SchedulerRunMinute int
}

// New creates a new Config struct from environment variables.
// Returns an error if required secrets are missing.
func New() (*Config, error) {
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

	// Supabase Auth
	supabaseJWTSecret := getEnv("SUPABASE_JWT_SECRET", "")
	if supabaseJWTSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET is required but not set")
	}

	// Supabase Admin API（アカウント削除で使用）
	// 未設定でも起動は妨げない。削除エンドポイントのみが 503 を返す。
	// これは既存環境が service_role キー未設定のまま動いているため、
	// 必須化すると設定が行き渡るまでサービス全体が起動しなくなるのを避ける判断。
	supabaseURL := getEnv("SUPABASE_URL", "")
	supabaseServiceRoleKey := getEnv("SUPABASE_SERVICE_ROLE_KEY", "")

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
	uploadPublicPath := getEnv("UPLOAD_PUBLIC_PATH", "/uploads")
	maxUploadSize := int64(getEnvAsInt("MAX_UPLOAD_SIZE", 5*1024*1024)) // 5MB default

	// Scheduler configuration
	animeSyncEnabled := getEnv("ANIME_SYNC_ENABLED", "false") == "true"
	schedulerRunHour := getEnvAsInt("SCHEDULER_RUN_HOUR", 3)     // Default: 3 AM
	schedulerRunMinute := getEnvAsInt("SCHEDULER_RUN_MINUTE", 0) // Default: 0 minutes

	return &Config{
		DatabaseURL:         databaseURL,
		Port:                port,
		Environment:         string(env),
		RateLimit:           rateLimit,
		SupabaseJWTSecret:      supabaseJWTSecret,
		SupabaseURL:            supabaseURL,
		SupabaseServiceRoleKey: supabaseServiceRoleKey,
		AnnictAPIKey:           annictAPIKey,
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
		UploadPublicPath:    uploadPublicPath,
		MaxUploadSize:       maxUploadSize,
		AnimeSyncEnabled:    animeSyncEnabled,
		SchedulerRunHour:    schedulerRunHour,
		SchedulerRunMinute:  schedulerRunMinute,
	}, nil
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
