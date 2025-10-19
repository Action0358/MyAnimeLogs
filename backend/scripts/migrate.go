package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
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

	// マイグレーションテーブルが存在しない場合は作成
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// マイグレーションファイルを取得
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// SQLファイルをフィルタリングしてソート
	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// マイグレーションを実行
	for _, filename := range sqlFiles {
		err := executeMigrationWithTransaction(db, filename, migrationsDir)
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}
	}

	fmt.Println("All migrations completed successfully!")
}

// executeMigrationWithTransaction トランザクション内で単一のマイグレーションを実行
func executeMigrationWithTransaction(db *database.DB, filename, migrationsDir string) (err error) {
	// トランザクションを開始
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// エラー時にトランザクションをロールバック
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for %s: %v", filename, rollbackErr)
			}
		}
	}()

	// マイグレーションが既に実行済みかチェック
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		fmt.Printf("Migration %s already executed, skipping...\n", filename)
		// トランザクションをコミット（変更がなくても）
		return tx.Commit()
	}

	// マイグレーションファイルを読み込む
	filePath := filepath.Join(migrationsDir, filename)
	var content []byte
	content, err = os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to read migration file: %w", err)
		return err
	}

	// マイグレーションSQLを実行
	fmt.Printf("Executing migration %s...\n", filename)
	_, err = tx.Exec(string(content))
	if err != nil {
		err = fmt.Errorf("failed to execute migration SQL: %w", err)
		return err
	}

	// 同じトランザクション内でマイグレーションを記録
	_, err = tx.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	if err != nil {
		err = fmt.Errorf("failed to record migration: %w", err)
		return err
	}

	// トランザクションをコミット
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		return err
	}

	fmt.Printf("Migration %s executed successfully\n", filename)
	return nil
}
