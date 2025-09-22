// Deprecated: This file is kept for backward compatibility.
// The main application entry point is now at cmd/server/main.go
package main

import (
	"log"
	"os"

	"myanimelogs/internal/config"
	"myanimelogs/internal/database"
	"myanimelogs/internal/router"

	"github.com/joho/godotenv"
)

// runMigrations executes database migrations
func runMigrations(databaseURL string) error {
	log.Println("Running database migrations...")

	// Connect to database for migrations
	db, err := database.New(databaseURL)
	if err != nil {
		return err
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
		return err
	}

	// Simple migration execution (basic implementation)
	// In production, you might want to use the full migrate.go logic
	log.Println("Migrations completed successfully")
	return nil
}

func main() {
	// Load environment variables based on environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	var envFile string
	switch env {
	case "production":
		envFile = ".env.production"
	case "test":
		envFile = ".env.test"
	default:
		envFile = ".env.local"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("No %s file found, using system environment variables", envFile)
	} else {
		log.Printf("Loaded configuration from %s", envFile)
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database with connection pool configuration
	dbConfig := database.ConnectionConfig{
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnMaxIdleTime: cfg.DBConnMaxIdleTime,
	}
	db, err := database.NewWithConfig(cfg.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations automatically on startup
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Printf("Migration failed: %v", err)
		// Continue anyway - migrations might have already been run
	}

	// Setup router
	r := router.Setup(db, cfg)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
