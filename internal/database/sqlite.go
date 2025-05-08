/*
Copyright © 2025 Parsel Email
*/
package database

import (
	"context"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parsel-email/lib-go/database/sqlite3"
	"github.com/parsel-email/lib-go/logger"
)

// Config represents database configuration
type Config struct {
	Type string
	Path string
}

// Initialize sets up a libsql database connection on disk
func Initialize() (Service, error) {
	dbFile := getEnvWithDefault("DB_FILE", "./db.sqlite")

	var err error

	cfg := sqlite3.DefaultConfig()
	cfg.Path = dbFile

	db, err := sqlite3.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	service := &service{
		db: db,
	}

	logger.Info(context.Background(), "Connected to database", "path", dbFile)
	return service, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
