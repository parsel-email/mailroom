/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/parsel-email/mailroom/internal/database"
	"github.com/spf13/cobra"
)

var (
	db database.DB
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mailroom",
	Short: "Processor for email messages",
	Long:  `Store and take action on email messages`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize the database
		dbType := os.Getenv("DB_TYPE")
		dbFile := os.Getenv("DB_FILE")

		if dbType == "" {
			dbType = "sqlite" // Default to SQLite
		}

		if dbFile == "" {
			dbFile = "./db.sqlite" // Default database file
		}

		var err error
		db, err = database.New(dbType, dbFile)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		// Run migrations
		if err := db.RunMigrations(); err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}

		fmt.Printf("Connected to %s database at %s\n", dbType, dbFile)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close database connection when the command completes
		if db != nil {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.watch.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
