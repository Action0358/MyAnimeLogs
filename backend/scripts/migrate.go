package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/migration"
)

func main() {
	log.Println("=== Migration process starting ===")
	log.Printf("PID: %d", os.Getpid())

	// 全体タイムアウト: 3分でマイグレーションが完了しなければ強制終了
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case sig := <-sigCh:
			log.Printf("Received signal %v, shutting down...", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	// 強制終了タイマー: 4分後にos.Exit（コンテキストキャンセルが効かない場合の保険）
	go func() {
		time.Sleep(4 * time.Minute)
		log.Fatal("FATAL: Hard timeout reached (4 min). Force exiting.")
	}()

	// 設定を読み込む
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	log.Printf("Environment: %s", cfg.Environment)

	// DATABASE_URLのホスト情報をログ出力（パスワードは除く）
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		if parsed, err := url.Parse(dbURL); err == nil {
			log.Printf("Database host: %s, port: %s", parsed.Hostname(), parsed.Port())
		}
	} else {
		log.Println("DATABASE_URL not set, using fallback config")
	}

	// データベースに接続
	log.Println("Connecting to database...")
	connectStart := time.Now()
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database (took %v): %v", time.Since(connectStart), err)
	}
	defer db.Close()
	log.Printf("Database connected successfully (took %v)", time.Since(connectStart))

	// コンテキストがキャンセルされていないか確認
	if ctx.Err() != nil {
		log.Fatalf("Context cancelled before migration: %v", ctx.Err())
	}

	// Migratorを作成
	migrator := migration.NewMigrator(db, "migrations")

	// 実行済みマイグレーションを確認
	executed, err := migrator.GetExecutedMigrations()
	if err != nil {
		log.Printf("Warning: could not get executed migrations: %v", err)
	} else {
		log.Printf("Already executed migrations: %d", len(executed))
	}

	// マイグレーションを実行
	log.Println("Running migrations...")
	migrateStart := time.Now()

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- migrator.RunMigrations()
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			log.Fatalf("Failed to run migrations (took %v): %v", time.Since(migrateStart), err)
		}
	case <-ctx.Done():
		log.Fatalf("Migration timed out after %v (context: %v)", time.Since(migrateStart), ctx.Err())
	}

	log.Printf("=== All migrations completed successfully! (took %v) ===", time.Since(migrateStart))
	fmt.Println("Migration done.") // stdoutにも出力（Fly.ioのログ取得用）
}
