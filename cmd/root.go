/*
Copyright © 2025 Parsel Email
*/
package cmd

import (
	"context"
	"database/sql" // Added import for sql.DB type assertion
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/parsel-email/lib-go/logger"
	"github.com/parsel-email/lib-go/tracing"
	"github.com/parsel-email/mailroom/internal/database"
	"github.com/parsel-email/mailroom/internal/server"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mailroom",
	Short: "Processor for email messages",
	Long:  `Store and take action on email messages`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the logger
		logger.Initialize(logger.LevelInfo)

		// Create a root context
		ctx := context.Background()

		// Initialize OpenTelemetry
		tracerShutdown, err := tracing.Initialize(ctx)
		if err != nil {
			logger.Error(ctx, "Failed to initialize tracing", "error", err)
			// Continue without tracing rather than failing
		} else {
			logger.Info(ctx, "OpenTelemetry tracing initialized")
		}

		// Initialize the database service
		dbService, err := database.InitializeLibSQl()
		if err != nil {
			logger.Error(ctx, "Failed to initialize database", "error", err)
			os.Exit(1)
		}
		logger.Info(ctx, "Database initialized successfully")

		// Get the *sql.DB instance from the service for migrations
		sqlDBProvider, ok := dbService.(interface{ DB() *sql.DB })
		if !ok {
			logger.Error(ctx, "Database service does not provide access to *sql.DB instance for migrations")
			os.Exit(1)
		}

		// Run database migrations
		migrationsPath := filepath.Join("db", "migrations")
		if err := database.MigrateUp(sqlDBProvider.DB(), migrationsPath); err != nil {
			logger.Error(ctx, "Failed to run database migrations", "error", err)
			os.Exit(1)
		}
		logger.Info(ctx, "Database migrations completed successfully")

		// The auth package does not require explicit initialization with dbService here.
		// Server handlers will use the dbService passed to server.NewServer().

		server := server.NewServer(dbService) // Pass dbService to NewServer

		// Create a done channel to signal when the shutdown is complete
		done := make(chan bool, 1)

		// Run graceful shutdown in a separate goroutine
		go gracefulShutdown(server, tracerShutdown, dbService, done) // Pass dbService to gracefulShutdown

		logger.Info(ctx, "Starting server", "port", os.Getenv("PORT"))
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("http server error: %s", err))
		}

		// Wait for the graceful shutdown to complete
		<-done
		logger.Info(ctx, "Graceful shutdown complete.")
	},
}

func gracefulShutdown(apiServer *http.Server, tracerShutdown func(context.Context) error, dbService database.Service, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger.Info(context.Background(), "shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the HTTP server
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(context.Background(), "Server forced to shutdown with error", "error", err)
	}

	// Shutdown the tracer provider
	if tracerShutdown != nil {
		if err := tracerShutdown(shutdownCtx); err != nil {
			logger.Error(context.Background(), "Failed to shutdown tracer provider", "error", err)
		}
	}

	// Close the database connection
	if dbService != nil {
		if err := dbService.Close(); err != nil {
			logger.Error(context.Background(), "Failed to close database connection", "error", err)
		}
	}

	logger.Info(context.Background(), "Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func init() {
	// Add flags here if needed
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
