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
		// Check if migration already executed
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations WHERE filename = $1", filename).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}

		if count > 0 {
			fmt.Printf("Migration %s already executed, skipping...\n", filename)
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", filename, err)
		}

		// Execute migration
		fmt.Printf("Executing migration %s...\n", filename)
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", filename, err)
		}

		fmt.Printf("Migration %s executed successfully\n", filename)
	}

	fmt.Println("All migrations completed successfully!")
}
