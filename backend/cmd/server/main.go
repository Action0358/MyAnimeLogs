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

	// Note: Migrations are now handled by scripts/migrate.go
	// Run 'make migrate' or 'go run scripts/migrate.go' before starting the server

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
