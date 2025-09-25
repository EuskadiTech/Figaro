// Package database provides database connection and migration functionality for Figaro.
package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB holds the database connection
var DB *sql.DB

// Initialize initializes the database connection and runs migrations
func Initialize(dataDir string) error {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database file path
	dbPath := filepath.Join(dataDir, "figaro.db")
	
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=ON")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	// Run migrations
	if err := runMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Test that we can query the database after migrations
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to test database connection after migrations: %w", err)
	}
	log.Printf("Database initialized successfully with %d users", count)

	return nil
}

// runMigrations executes database migrations using simple SQL execution
func runMigrations(db *sql.DB) error {
	log.Printf("Running database migrations...")

	// Read migration file
	migrationSQL, err := migrationFS.ReadFile("migrations/001_initial_schema.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration SQL
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Printf("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}