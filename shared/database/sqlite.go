package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a SQLite database connection
type DB struct {
	*sql.DB
	Path string
}

// Open opens a SQLite database at the given path
// Creates the directory if it doesn't exist
func Open(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db, Path: path}, nil
}

// RunMigrations executes SQL migration files
func (db *DB) RunMigrations(migrations []string) error {
	// Create migrations table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run each migration
	for i, migration := range migrations {
		name := fmt.Sprintf("migration_%03d", i+1)

		// Check if already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", name).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			continue // Already applied
		}

		// Apply migration
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", name, err)
		}

		// Record migration
		if _, err := db.Exec("INSERT INTO migrations (name) VALUES (?)", name); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
