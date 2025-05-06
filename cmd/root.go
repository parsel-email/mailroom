/*
Copyright © 2025 Parsel Email
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/parsel-email/mailroom/internal/database"
	"github.com/spf13/cobra"
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
	if err := database.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// closeDatabase closes the database connection
func closeDatabase() {
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

func init() {
	// Add flags here if needed
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
