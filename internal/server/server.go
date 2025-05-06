package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parsel-email/mailroom/internal/database"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer(dbService database.Service) *http.Server { // Added dbService parameter
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Use the provided dbService instead of initializing a new one
	NewServer := &Server{
		port: port,
		db:   dbService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
