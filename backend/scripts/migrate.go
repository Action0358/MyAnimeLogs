package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/migration"
)

func main() {
	log.Println("Migration process starting...")

	// 全体タイムアウト: 5分でマイグレーションが完了しなければ強制終了
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// 設定を読み込む
	cfg := config.New()
	log.Printf("Environment: %s", cfg.Environment)

	// データベースに接続
	log.Println("Connecting to database...")
	connectStart := time.Now()
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database (took %v): %v", time.Since(connectStart), err)
	}
	defer db.Close()
	log.Printf("Database connected (took %v)", time.Since(connectStart))

	// コンテキストがキャンセルされていないか確認
	if ctx.Err() != nil {
		log.Fatalf("Context cancelled before migration: %v", ctx.Err())
	}

	// Migratorを作成
	migrator := migration.NewMigrator(db, "migrations")

	// マイグレーションを実行
	log.Println("Running migrations...")
	migrateStart := time.Now()

	// マイグレーションをgoroutineで実行し、タイムアウトを監視
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
		log.Fatalf("Migration timed out after %v", time.Since(migrateStart))
	}

	log.Printf("All migrations completed successfully! (took %v)", time.Since(migrateStart))
}
