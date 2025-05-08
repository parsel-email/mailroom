package database

import "database/sql"

type BaseService interface {
	Health() map[string]string
	Close() error
}

type Service interface {
	BaseService
	UserService
	DB() *sql.DB
}

type service struct {
	db *sql.DB
	UserService
}

func (s *service) Health() map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (s *service) Close() error {
	return s.db.Close()
}

func (s *service) DB() *sql.DB {
	return s.db
}
