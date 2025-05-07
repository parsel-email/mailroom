package database

import (
	"context"

	"github.com/parsel-email/mailroom/db/lib/schema" // sqlc generated types
)

// UserService defines the interface for user-related database operations.
// Implementations will typically use sqlc-generated query methods.
type UserService interface {
	GetUserByID(ctx context.Context, id string) (schema.User, error)
	GetUserByProviderID(ctx context.Context, arg schema.GetUserByProviderIDParams) (schema.User, error)
	CreateUser(ctx context.Context, arg schema.CreateUserParams) (schema.User, error)
}
