package auth

import (
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// JWT related constants
const (
	TokenExpiryUser    = 24 * time.Hour // 24 hour expiry for user tokens
	TokenExpiryService = 24 * time.Hour // 24 hour expiry for service tokens
	IssuerName         = "parsel-auth-service"
	ServiceAudience    = "parsel-services"
)

// ValidateJWT validates a JWT token
func ValidateJWT(token string) bool {
	// Remove the "Bearer " prefix if it exists
	token = strings.TrimPrefix(token, "Bearer ")

	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("AUTH_SECRET")), nil
	})
	if err != nil {
		// Skip logging for tests to avoid nil pointer
		return false
	}

	// Validate the token
	if !parsedToken.Valid {
		// Skip logging for tests to avoid nil pointer
		return false
	}

	return true
}

// ParseToken parses a JWT token and returns the parsed token
func ParseToken(token string) (*jwt.Token, error) {
	// Remove the "Bearer " prefix if it exists
	token = strings.TrimPrefix(token, "Bearer ")

	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("AUTH_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}

// GetIDFromJWT extracts the user ID from a JWT token
func GetIDFromJWT(token string) (string, error) {
	// Parse the token
	parsedToken, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Get the ID from the claims
	id, ok := claims["ID"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return id, nil
}

// GetSessionIDFromJWT extracts the session ID from a JWT token
func GetSessionIDFromJWT(token string) (string, error) {
	// Parse the token
	parsedToken, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Get the session ID from the claims
	sessionID, ok := claims["sessionID"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return sessionID, nil
}

// GetProviderFromJWT extracts the provider from a JWT token
func GetProviderFromJWT(token string) (string, error) {
	// Parse the token
	parsedToken, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Get the provider from the claims
	provider, ok := claims["provider"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return provider, nil
}

// GetProviderIDFromJWT extracts the provider ID from a JWT token
func GetProviderIDFromJWT(token string) (string, error) {
	// Parse the token
	parsedToken, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Get the provider ID from the claims
	providerID, ok := claims["providerID"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return providerID, nil
}
