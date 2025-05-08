package auth

import "errors"

var (
	ErrProviderNotInitialized = errors.New("auth provider not initialized")
	ErrUnknownProvider        = errors.New("unknown auth provider")
	ErrInvalidToken           = errors.New("invalid token")
	ErrEmptyToken             = errors.New("empty token")
	ErrInvalidUser            = errors.New("invalid user: required fields are empty")
	ErrDatabaseNotInitialized = errors.New("database service not initialized")
	ErrMissingAuthSecret      = errors.New("auth secret is not set")
)
