// Package database provides database connection and migration functionality for Figaro.
package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB holds the database connection
var DB *sql.DB

// MigrationState represents the current state of applied migrations
type MigrationState struct {
	AppliedMigrations []string `json:"applied_migrations"`
}

// loadMigrationState loads the current migration state from migrations.json
func loadMigrationState(dataDir string) (*MigrationState, error) {
	migrationsFile := filepath.Join(dataDir, "migrations.json")
	
	// If file doesn't exist, check if this is a fresh install or existing database
	if _, err := os.Stat(migrationsFile); os.IsNotExist(err) {
		// Check if database exists and has tables (existing installation)
		dbPath := filepath.Join(dataDir, "figaro.db")
		if dbExists, hasData := isDatabaseWithData(dbPath); dbExists && hasData {
			// Database exists with data, try to detect which migrations are already applied
			log.Printf("migrations.json not found but database exists with data, detecting applied migrations")
			return detectAppliedMigrations(dataDir)
		}
		
		// Fresh installation (no database or empty database)
		log.Printf("migrations.json not found, assuming new installation")
		return &MigrationState{AppliedMigrations: []string{}}, nil
	}
	
	data, err := os.ReadFile(migrationsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations.json: %w", err)
	}
	
	var state MigrationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse migrations.json: %w", err)
	}
	
	log.Printf("Loaded migration state: %d applied migrations", len(state.AppliedMigrations))
	return &state, nil
}

// isDatabaseWithData checks if a database file exists and contains data
func isDatabaseWithData(dbPath string) (bool, bool) {
	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false, false
	}
	
	// Try to open and check for data
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=ON")
	if err != nil {
		return true, false // File exists but can't open
	}
	defer db.Close()
	
	// Check if any tables exist (indicating this is not a fresh database)
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return true, false // File exists but can't query
	}
	defer rows.Close()
	
	// If we get any table names, the database has data
	return true, rows.Next()
}

// detectAppliedMigrations tries to detect which migrations have been applied to an existing database
func detectAppliedMigrations(dataDir string) (*MigrationState, error) {
	dbPath := filepath.Join(dataDir, "figaro.db")
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database for detection: %w", err)
	}
	defer db.Close()

	state := &MigrationState{AppliedMigrations: []string{}}
	
	// Get all migration files to check against
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	
	var migrationFiles []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)
	
	// Check each migration to see if its changes are already present
	for _, filename := range migrationFiles {
		applied, err := isMigrationAlreadyInDB(db, filename)
		if err != nil {
			log.Printf("Warning: Could not detect state of migration %s: %v", filename, err)
			continue
		}
		
		if applied {
			log.Printf("Detected migration %s as already applied", filename)
			state.AppliedMigrations = append(state.AppliedMigrations, filename)
		} else {
			log.Printf("Detected migration %s as not applied", filename)
		}
	}
	
	return state, nil
}

// isMigrationAlreadyInDB checks if a migration's changes are already present in the database
func isMigrationAlreadyInDB(db *sql.DB, migrationName string) (bool, error) {
	switch migrationName {
	case "001_initial_schema.up.sql":
		// Check if users table exists
		_, err := db.Exec("SELECT 1 FROM users LIMIT 1")
		return err == nil, nil
		
	case "002_user_sessions.up.sql":
		// Check if user_sessions table exists
		_, err := db.Exec("SELECT 1 FROM user_sessions LIMIT 1")
		return err == nil, nil
		
	case "003_add_activity_urls.up.sql":
		// Check if meeting_url column exists in activities table
		rows, err := db.Query("PRAGMA table_info(activities)")
		if err != nil {
			return false, nil // Table doesn't exist
		}
		defer rows.Close()
		
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, dfltValue, pk interface{}
			if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
				continue
			}
			if name == "meeting_url" {
				return true, nil
			}
		}
		return false, nil
		
	case "004_activity_sharing_and_links.up.sql":
		// Check if activity_shares table exists
		_, err := db.Exec("SELECT 1 FROM activity_shares LIMIT 1")
		return err == nil, nil
		
	case "005_add_activity_status_material_category.up.sql":
		// Check if status column exists in activities table
		rows, err := db.Query("PRAGMA table_info(activities)")
		if err != nil {
			return false, nil // Table doesn't exist
		}
		defer rows.Close()
		
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, dfltValue, pk interface{}
			if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
				continue
			}
			if name == "status" {
				return true, nil
			}
		}
		return false, nil
		
	case "006_test_migration.up.sql":
		// Check if test_migration_table exists
		_, err := db.Exec("SELECT 1 FROM test_migration_table LIMIT 1")
		return err == nil, nil
	}
	
	// Unknown migration, assume not applied
	return false, nil
}

// saveMigrationState saves the current migration state to migrations.json
func saveMigrationState(dataDir string, state *MigrationState) error {
	migrationsFile := filepath.Join(dataDir, "migrations.json")
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal migration state: %w", err)
	}
	
	if err := os.WriteFile(migrationsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write migrations.json: %w", err)
	}
	
	log.Printf("Saved migration state: %d applied migrations", len(state.AppliedMigrations))
	return nil
}

// isMigrationApplied checks if a migration has already been applied
func isMigrationApplied(state *MigrationState, migrationName string) bool {
	for _, applied := range state.AppliedMigrations {
		if applied == migrationName {
			return true
		}
	}
	return false
}

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
	if err := runMigrations(db, dataDir); err != nil {
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

// runMigrations executes database migrations using migration tracking
func runMigrations(db *sql.DB, dataDir string) error {
	log.Printf("Running database migrations...")

	// Load current migration state
	state, err := loadMigrationState(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load migration state: %w", err)
	}

	// Get all migration files
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .up.sql files
	var migrationFiles []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort to ensure migrations run in order
	sort.Strings(migrationFiles)

	// Track whether any migrations were applied
	migrationsApplied := false

	// Execute each migration that hasn't been applied yet
	for _, filename := range migrationFiles {
		if isMigrationApplied(state, filename) {
			log.Printf("Migration %s already applied, skipping", filename)
			continue
		}

		log.Printf("Applying migration: %s", filename)

		migrationSQL, err := migrationFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration SQL
		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Add to applied migrations
		state.AppliedMigrations = append(state.AppliedMigrations, filename)
		migrationsApplied = true

		log.Printf("Applied migration: %s", filename)
	}

	// Save updated migration state if any migrations were applied
	if migrationsApplied {
		if err := saveMigrationState(dataDir, state); err != nil {
			return fmt.Errorf("failed to save migration state: %w", err)
		}
	} else {
		// Save state even when no migrations were applied (to create migrations.json for existing databases)
		migrationsFile := filepath.Join(dataDir, "migrations.json")
		if _, err := os.Stat(migrationsFile); os.IsNotExist(err) {
			if err := saveMigrationState(dataDir, state); err != nil {
				return fmt.Errorf("failed to save migration state: %w", err)
			}
		}
	}

	if migrationsApplied {
		log.Printf("Database migrations completed successfully")
	} else {
		log.Printf("No new migrations to apply")
	}
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
