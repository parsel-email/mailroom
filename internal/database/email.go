package database

import (
	"context"

	"github.com/parsel-email/mailroom/db/lib/schema"
)

type EmailService interface {
	// CreateEmail creates a new email in the database.
	CreateEmail(ctx context.Context, arg schema.CreateEmailParams) (schema.Email, error)
	// UpdateEmail updates an existing email in the database.
	UpdateEmail(ctx context.Context, arg schema.UpdateEmailParams) (schema.Email, error)
	// DeleteEmail deletes an email from the database.
	DeleteEmail(ctx context.Context, id string) error
	// ListEmails retrieves a list of emails from the database.
	ListEmails(ctx context.Context, arg schema.ListEmailsParams) ([]schema.Email, error)
	// GetEmailByID retrieves an email by its ID.
	GetEmailByID(ctx context.Context, id string) (schema.Email, error)
}
