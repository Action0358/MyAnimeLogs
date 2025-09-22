package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type ConnectionConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
}

func New(databaseURL string) (*DB, error) {
	return NewWithConfig(databaseURL, ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5,
		ConnMaxIdleTime: 1,
	})
}

func NewWithConfig(databaseURL string, config ConnectionConfig) (*DB, error) {
	var db *sql.DB
	var err error

	// Retry connection for a few times
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		// Configure connection pool
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Minute)
		db.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Minute)

		err = db.Ping()
		if err == nil {
			// Connection successful
			log.Println("Successfully connected to the database with connection pool configured.")
			log.Printf("Connection pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
				config.MaxOpenConns, config.MaxIdleConns, time.Duration(config.ConnMaxLifetime)*time.Minute, time.Duration(config.ConnMaxIdleTime)*time.Minute)
			return &DB{db}, nil
		}

		log.Printf("Failed to ping database (attempt %d/%d): %v. Retrying in 5 seconds...", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// Note: Database migrations are now handled by scripts/migrate.go
// This provides better version control and rollback capabilities
