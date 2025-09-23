package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrator_RunMigrations(t *testing.T) {
	// Create temporary directory for test migrations
	tempDir := t.TempDir()

	// Create test migration files
	migration1 := `CREATE TABLE test_table1 (id SERIAL PRIMARY KEY, name VARCHAR(100));`
	migration2 := `CREATE TABLE test_table2 (id SERIAL PRIMARY KEY, value INTEGER);`

	err := os.WriteFile(filepath.Join(tempDir, "001_create_test_table1.sql"), []byte(migration1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "002_create_test_table2.sql"), []byte(migration2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test migration file: %v", err)
	}

	// Note: This test would require a test database connection
	// For now, we'll test the file reading functionality
	migrator := &Migrator{migrationsPath: tempDir}

	files, err := migrator.getMigrationFiles()
	if err != nil {
		t.Fatalf("Failed to get migration files: %v", err)
	}

	expectedFiles := []string{"001_create_test_table1.sql", "002_create_test_table2.sql"}
	if len(files) != len(expectedFiles) {
		t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(files))
	}

	for i, expected := range expectedFiles {
		if files[i] != expected {
			t.Errorf("Expected file %s, got %s", expected, files[i])
		}
	}
}

func TestMigrator_GetMigrationFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	migrator := &Migrator{migrationsPath: tempDir}
	files, err := migrator.getMigrationFiles()
	if err != nil {
		t.Fatalf("Failed to get migration files: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(files))
	}
}

func TestMigrator_GetMigrationFiles_NonSQLFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create non-SQL files that should be ignored
	err := os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "001_migration.sql"), []byte("CREATE TABLE test;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test migration: %v", err)
	}

	migrator := &Migrator{migrationsPath: tempDir}
	files, err := migrator.getMigrationFiles()
	if err != nil {
		t.Fatalf("Failed to get migration files: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0] != "001_migration.sql" {
		t.Errorf("Expected 001_migration.sql, got %s", files[0])
	}
}
