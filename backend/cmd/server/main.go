// Deprecated: This file is kept for backward compatibility.
// The main application entry point is now at cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myanimelogs/internal/common"
	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/migration"
	"myanimelogs/internal/router"
	"myanimelogs/internal/scheduler"
	"myanimelogs/internal/services"

	"github.com/joho/godotenv"
)

const (
	// ヘッダーを少しずつ送って接続を占有する攻撃(Slowloris)を防ぐ
	serverReadHeaderTimeout = 5 * time.Second
	// ヘッダーと本文を読み終えるまでの上限。本文は最大でも 64KiB なので短くできる。
	// 読み終えた後のハンドラーの処理には影響しない。
	serverReadTimeout = 15 * time.Second
	// 応答を返し終えるまでの上限。カバー画像の取得は AniList のレート制限
	// (90 req/分)を待つため、一覧表示でも1分を超えうる。
	serverWriteTimeout = 120 * time.Second
	// Fly のプロキシが再利用しようとした接続を先に閉じないよう、長めにとる
	serverIdleTimeout = 120 * time.Second
)

// newHTTPServer はタイムアウトを設定した HTTP サーバーを作る。
// 管理者の同期APIは middleware.ClearWriteDeadline で書き込み期限を外している。
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
}

// runMigrations executes database migrations using the migrator
func runMigrations(db *database.DB) error {
	// Check if migrations directory exists
	migrationsPath := "migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Printf("Migrations directory %s not found, skipping migrations", migrationsPath)
		return nil
	}

	log.Printf("Starting migrations from directory: %s", migrationsPath)
	migrator := migration.NewMigrator(db, migrationsPath)
	return migrator.RunMigrations()
}

func main() {
	log.Printf("MyAnimeLogs Backend Starting...")

	// Load environment variables based on environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	log.Printf("Environment: %s", env)

	var envFile string
	switch env {
	case "production":
		envFile = ".env.production"
	case "staging":
		envFile = ".env.staging"
	default:
		envFile = ".env.local"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No %s file found, using system environment variables", envFile)
	} else {
		log.Printf("Loaded configuration from %s", envFile)
	}

	// Initialize configuration
	log.Printf("Initializing configuration...")
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Log database connection info (without sensitive data)
	if cfg.DatabaseURL != "" {
		log.Printf("Database URL configured (Supabase)")
	} else {
		log.Printf("No DATABASE_URL found, using individual postgres settings")
	}

	// Log other important config
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)

	// Initialize database with connection pool configuration optimized for Supabase
	log.Printf("Connecting to Supabase database...")
	dbConfig := database.ConnectionConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	}

	// Validate database URL
	if cfg.DatabaseURL == "" {
		log.Fatalf("DATABASE_URL is required but not set")
	}

	db, err := database.NewWithConfig(cfg.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Supabase database: %v", err)
	}
	defer db.Close()
	log.Printf("Supabase database connection established")

	// Run migrations automatically on startup
	log.Printf("Running database migrations...")
	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed, stopping application: %v", err)
	}
	log.Printf("Database migrations completed")

	// Initialize logger (shared across all components)
	logger := common.NewLogger(common.LogLevel(cfg.LogLevel), cfg.LogFormat)
	log.Printf("Logger initialized with level: %s", cfg.LogLevel)

	// まずAnnictClientのポインタを宣言し、後から初期化と設定を行う
	var annictClient *services.AnnictClient

	// Initialize anime service. AnnictClientは後からSetする。
	animeService := services.NewAnimeService(db, logger)
	log.Printf("Anime service initialized")

	// Initialize Annict client (shared across all components)
	annictClient = services.NewAnnictClientWithDB(cfg.AnnictAPIKey, db, logger)
	log.Printf("Annict client initialized")

	// AnnictClientにAnimeServiceを設定（循環参照回避のため後から設定）
	annictClient.SetAnimeService(animeService)

	// AnimeServiceにAnnictClientを設定（エピソード作成でAnnict APIを使用）
	animeService.SetAnnictClient(annictClient)

	// Initialize and start anime sync scheduler
	// Pass animeService.GetAniListClient() to share the same rate-limited instance
	animeSyncScheduler, err := scheduler.NewAnimeSyncScheduler(
		annictClient,
		animeService,
		animeService.GetAniListClient(),
		logger,
		cfg.AnimeSyncEnabled,
		cfg.SchedulerRunHour,
		cfg.SchedulerRunMinute,
	)
	if err != nil {
		log.Fatalf("Failed to create anime sync scheduler: %v", err)
	}
	animeSyncScheduler.Start()

	// Setup router (pass shared dependencies)
	log.Printf("Setting up routes...")
	r := router.Setup(db, cfg, animeSyncScheduler, logger, annictClient)
	log.Printf("Routes configured")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Health check available at /health on port %s", port)

	// Create HTTP server for graceful shutdown
	srv := newHTTPServer("0.0.0.0:"+port, r)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Printf("Shutting down gracefully...")

	// Stop scheduler
	animeSyncScheduler.Stop()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Printf("Server stopped")
}
