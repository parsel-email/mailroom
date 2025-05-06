/*
Copyright © 2025 Parsel Email
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/parsel-email/lib-go/database"
	"github.com/spf13/cobra"
)

var (
	db *database.DB
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mailroom",
	Short: "Processor for email messages",
	Long:  `Store and take action on email messages`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initDatabase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Mailroom!")
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		closeDatabase()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initDatabase initializes the database connection
func initDatabase() {
	dbType := getEnvWithDefault("DB_TYPE", "sqlite")
	dbFile := getEnvWithDefault("DB_FILE", "./db.sqlite")

	var err error

	cfg := database.Config{
		Path: dbFile,
	}

	db, err = database.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Printf("Connected to %s database at %s\n", dbType, dbFile)
}

// closeDatabase closes the database connection
func closeDatabase() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	// Add flags here if needed
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
