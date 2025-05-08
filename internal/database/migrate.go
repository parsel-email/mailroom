package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// MigrateUp runs all available migrations
func MigrateUp(db *sql.DB, migrationsPath string) error { // Changed db type
	m, err := getMigrator(db, migrationsPath)
	if err != nil {
		return fmt.Errorf("error creating migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back all migrations
func MigrateDown(db *sql.DB, migrationsPath string) error { // Changed db type
	m, err := getMigrator(db, migrationsPath)
	if err != nil {
		return fmt.Errorf("error creating migrator: %w", err)
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error running down migrations: %w", err)
	}

	slog.Info("Database rollback completed successfully")
	return nil
}

// MigrateTo migrates to a specific version
func MigrateTo(db *sql.DB, migrationsPath string, version uint) error { // Changed db type
	m, err := getMigrator(db, migrationsPath)
	if err != nil {
		return fmt.Errorf("error creating migrator: %w", err)
	}

	if err := m.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error migrating to version %d: %w", version, err)
	}

	slog.Info("Database migration to version completed successfully", "version", version)
	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(db *sql.DB, migrationsPath string) (uint, bool, error) { // Changed db type
	m, err := getMigrator(db, migrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("error creating migrator: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("error getting migration version: %w", err)
	}

	return version, dirty, nil
}

// getMigrator creates a new migrator instance
func getMigrator(db *sql.DB, migrationsPath string) (*migrate.Migrate, error) { // Changed db type
	// Verify migrations directory exists
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Test the provided database connection
	if err = db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create driver for sqlite3
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite3 driver: %w", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"sqlite3", // Changed from "postgres"
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}
