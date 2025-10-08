//go:build ignore
// +build ignore

package main

// Reset Migrations Tool
//
// ⚠️ WARNING: This tool resets the entire migration history in Production
//
// Purpose:
// - Apply improved migration files after modification
// - Recover from corrupted migration history
//
// Usage:
// 1. Build: GOOS=linux GOARCH=amd64 go build -o /tmp/reset-migrations scripts/reset-migrations.go
// 2. Upload: flyctl ssh sftp shell -a myanimelogs-production
// 3. Execute: flyctl ssh console -a myanimelogs-production -C "/tmp/reset-migrations"
//
// Note:
// - DO NOT use in normal operations
// - Data is preserved, only migration history is deleted
// - Always backup before running
//
// Last used: 2025-10-07 (Production DB sync issue resolution)

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Current Migration History ===")
	rows, err := db.Query("SELECT id, filename, executed_at FROM migrations ORDER BY id")
	if err != nil {
		log.Fatalf("Failed to query migrations: %v", err)
	}

	count := 0
	for rows.Next() {
		var id int
		var filename, executedAt string
		if err := rows.Scan(&id, &filename, &executedAt); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("%d. %s (executed: %s)\n", id, filename, executedAt)
		count++
	}
	rows.Close()

	fmt.Printf("\nTotal migrations: %d\n\n", count)

	fmt.Println("=== Resetting Migration History ===")
	_, err = db.Exec("TRUNCATE TABLE migrations")
	if err != nil {
		log.Fatalf("Failed to truncate migrations table: %v", err)
	}

	fmt.Println("✅ Migration history reset complete")
	fmt.Println("\nNext steps:")
	fmt.Println("1. Restart the application: flyctl apps restart myanimelogs-production")
	fmt.Println("2. Check migration logs: flyctl logs -a myanimelogs-production -n | grep -i migration")
}
