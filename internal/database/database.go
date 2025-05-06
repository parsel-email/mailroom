/*
Copyright © 2025 Parsel Email
*/
package database

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parsel-email/lib-go/database"
)

// DB is the global database instance
var DB *database.DB

// Config represents database configuration
type Config struct {
	Type string
	Path string
}

// Initialize sets up the database connection
func Initialize() error {
	dbType := getEnvWithDefault("DB_TYPE", "libsql")
	dbFile := getEnvWithDefault("DB_FILE", "./db.libsql")

	var err error

	cfg := database.Config{
		Path: dbFile,
	}

	DB, err = database.Open(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Printf("Connected to %s database at %s\n", dbType, dbFile)
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
