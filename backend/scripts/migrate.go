package main

import (
	"log"

	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/migration"
)

func main() {
	// 設定を読み込む
	cfg := config.New()

	// データベースに接続
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migratorを作成
	migrator := migration.NewMigrator(db, "migrations")

	// マイグレーションを実行
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("All migrations completed successfully!")
}
