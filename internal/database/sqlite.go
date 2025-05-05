/*
Copyright © 2025 Parsel Email <contact@parsel.email>
*/
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite represents a SQLite database connection
type SQLite struct {
	DB     *sql.DB
	DBPath string
}

// NewSQLite creates a new SQLite database connection
func NewSQLite(dbPath string) (*SQLite, error) {
	// Create the directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return &SQLite{
		DB:     db,
		DBPath: dbPath,
	}, nil
}

// Close closes the database connection
func (s *SQLite) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// RunMigrations runs all migrations for the SQLite database
func (s *SQLite) RunMigrations() error {
	// Create a basic schema for the database
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS emails (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id TEXT UNIQUE,
			from_address TEXT NOT NULL,
			to_addresses TEXT NOT NULL,
			subject TEXT,
			body_text TEXT,
			body_html TEXT,
			received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			processed BOOLEAN DEFAULT FALSE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create emails table: %w", err)
	}

	return nil
}
