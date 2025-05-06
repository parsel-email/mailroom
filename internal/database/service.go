package database

import "database/sql"

// BaseService represents a service that interacts with a database.
type BaseService interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type Service interface {
	BaseService
	DB() *sql.DB // Added method to get the underlying *sql.DB instance
}

type service struct {
	db *sql.DB
}

func (s *service) Health() map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (s *service) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB instance.
func (s *service) DB() *sql.DB { // Implemented method
	return s.db
}
