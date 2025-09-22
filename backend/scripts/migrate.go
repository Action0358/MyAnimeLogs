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
	// Load configuration
	cfg := config.New()

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migrations table if not exists
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

	// Get migration files
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Execute migrations
	for _, filename := range sqlFiles {
		err := executeMigrationWithTransaction(db, filename, migrationsDir)
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}
	}

	fmt.Println("All migrations completed successfully!")
}

// executeMigrationWithTransaction executes a single migration within a transaction
func executeMigrationWithTransaction(db *database.DB, filename, migrationsDir string) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for %s: %v", filename, rollbackErr)
			}
		}
	}()

	// Check if migration already executed
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		fmt.Printf("Migration %s already executed, skipping...\n", filename)
		// Commit the transaction (even though no changes were made)
		return tx.Commit()
	}

	// Read migration file
	filePath := filepath.Join(migrationsDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration SQL
	fmt.Printf("Executing migration %s...\n", filename)
	_, err = tx.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in the same transaction
	_, err = tx.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	fmt.Printf("Migration %s executed successfully\n", filename)
	return nil
}
