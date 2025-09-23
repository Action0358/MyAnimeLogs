package migration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"myanimelogs/internal/database"
)

// Migrator handles database migrations
type Migrator struct {
	db             *database.DB
	migrationsPath string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *database.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// RunMigrations executes all pending migrations
func (m *Migrator) RunMigrations() error {
	log.Printf("Starting database migrations from path: %s", m.migrationsPath)

	// Create migrations table if not exists
	log.Printf("Creating migrations tracking table...")
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	log.Printf("Migrations tracking table ready")

	// Get migration files
	log.Printf("Scanning for migration files...")
	migrationFiles, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	log.Printf("Found %d migration files: %v", len(migrationFiles), migrationFiles)

	if len(migrationFiles) == 0 {
		log.Println("No migration files found")
		return nil
	}

	// Execute migrations
	executedCount := 0
	for _, filename := range migrationFiles {
		log.Printf("Processing migration: %s", filename)
		executed, err := m.executeMigration(filename)
		if err != nil {
			log.Printf("Migration %s failed: %v", filename, err)
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
		if executed {
			executedCount++
			log.Printf("Migration %s completed successfully", filename)
		}
	}

	if executedCount > 0 {
		log.Printf("Successfully executed %d new migrations", executedCount)
	} else {
		log.Println("All migrations were already executed")
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getMigrationFiles returns sorted list of migration files
func (m *Migrator) getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var sqlFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") && !file.IsDir() {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	sort.Strings(sqlFiles)
	return sqlFiles, nil
}

// executeMigration executes a single migration file
func (m *Migrator) executeMigration(filename string) (bool, error) {
	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
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
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		log.Printf("Migration %s already executed, skipping...", filename)
		// Commit the transaction (even though no changes were made)
		if err = tx.Commit(); err != nil {
			return false, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return false, nil
	}

	// Read migration file
	filePath := filepath.Join(m.migrationsPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read migration file %s: %w", filename, err)
	}

	// Validate migration content
	if len(strings.TrimSpace(string(content))) == 0 {
		return false, fmt.Errorf("migration file %s is empty", filename)
	}

	// Execute migration SQL
	log.Printf("Executing migration %s...", filename)
	_, err = tx.Exec(string(content))
	if err != nil {
		return false, fmt.Errorf("failed to execute migration SQL in %s: %w", filename, err)
	}

	// Record migration in the same transaction
	_, err = tx.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return false, fmt.Errorf("failed to record migration %s: %w", filename, err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit transaction for %s: %w", filename, err)
	}

	log.Printf("Migration %s executed successfully", filename)
	return true, nil
}

// GetExecutedMigrations returns list of executed migrations
func (m *Migrator) GetExecutedMigrations() ([]string, error) {
	rows, err := m.db.Query("SELECT filename FROM migrations ORDER BY executed_at")
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, fmt.Errorf("failed to scan migration filename: %w", err)
		}
		migrations = append(migrations, filename)
	}

	return migrations, nil
}
