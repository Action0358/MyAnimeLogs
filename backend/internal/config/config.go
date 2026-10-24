package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL  string
	Port         string
	Environment  string
	// RateLimitPerMinute は1つのIPアドレスから1分間に受け付けるリクエスト数の上限
	RateLimitPerMinute int
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
	// Scheduler configuration
	AnimeSyncEnabled   bool
	SchedulerRunHour   int
	SchedulerRunMinute int
	// MaintenanceMode が true の間、/health と管理系を除く全リクエストに 503（MAINTENANCE）を返す。
	// 環境変数 MAINTENANCE_MODE=true で有効化する（fly secrets set で切り替え）
	MaintenanceMode bool
	// MaintenanceEndsAt は終了予定（MAINTENANCE_ENDS_AT, RFC3339）。未設定・不正な値なら nil。
	// 切り替え中に起動が止まらないよう、不正な値でも起動は妨げない
	MaintenanceEndsAt *time.Time
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

	// Rate Limiter（IPアドレスごと・1分間あたり）
	//
	// 旧 RATE_LIMIT は単位が曖昧だった。コードは「1秒あたり」として扱っていたが、
	// コメントとREADMEは「1分あたり」と書いており、60倍ずれていた。
	// 同じ名前のまま意味を変えると、設定済みの値（60など）が突然厳しくなるため、
	// 単位を名前に含めた新しい環境変数に切り替えた。旧 RATE_LIMIT は読まない。
	//
	// 外部API（Annict・AniList・Jikan・TMDB）の上限は、この制限ではなく各クライアント側で守る。
	rateLimitPerMinute := getEnvAsInt("RATE_LIMIT_PER_MINUTE", 600)

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

	// Scheduler configuration
	animeSyncEnabled := getEnv("ANIME_SYNC_ENABLED", "false") == "true"
	schedulerRunHour := getEnvAsInt("SCHEDULER_RUN_HOUR", 3)     // Default: 3 AM
	schedulerRunMinute := getEnvAsInt("SCHEDULER_RUN_MINUTE", 0) // Default: 0 minutes

	maintenanceMode := getEnv("MAINTENANCE_MODE", "false") == "true"
	maintenanceEndsAt := parseMaintenanceEndsAt(getEnv("MAINTENANCE_ENDS_AT", ""))

	return &Config{
		DatabaseURL:         databaseURL,
		Port:                port,
		Environment:         string(env),
		RateLimitPerMinute:  rateLimitPerMinute,
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
		AnimeSyncEnabled:    animeSyncEnabled,
		SchedulerRunHour:    schedulerRunHour,
		SchedulerRunMinute:  schedulerRunMinute,
		MaintenanceMode:     maintenanceMode,
		MaintenanceEndsAt:   maintenanceEndsAt,
	}, nil
}

// parseMaintenanceEndsAt は RFC3339 の終了予定を解釈する。空・不正な値は nil。
func parseMaintenanceEndsAt(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
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
