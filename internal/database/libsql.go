/*
Copyright © 2025 Parsel Email
*/
package database

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parsel-email/lib-go/database/libsql"
)

// Config represents database configuration
type Config struct {
	Type string
	Path string
}

// InitializeLibSQl sets up a libsql database connection on disk
func InitializeLibSQl() (Service, error) {
	dbFile := getEnvWithDefault("DB_FILE", "./db.libsql")

	var err error

	cfg := libsql.DefaultConfig()
	cfg.Path = dbFile

	db, err := libsql.Open(cfg)
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

	fmt.Printf("Connected to libsql database at %s\n", dbFile)
	return service, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
