package repository

import (
	"context"

	"backend-challenge/internal/domain/model"
)

// ExternalUserRepository defines the interface for external user data access
type ExternalUserRepository interface {
	// Create creates a new external user in the database
	Create(ctx context.Context, user *model.ExternalUser) error

	// GetByID fetches an external user by ID
	GetByID(ctx context.Context, id string) (*model.ExternalUser, error)

	// Update updates an external user in the database
	Update(ctx context.Context, user *model.ExternalUser) error

	// Delete removes an external user from the database
	Delete(ctx context.Context, id string) error

	// List returns all external users
	List(ctx context.Context) ([]*model.ExternalUser, error)

	// ListByDepartment returns all external users from a specific department
	ListByDepartment(ctx context.Context, department string) ([]*model.ExternalUser, error)

	// ImportFromAPI imports external users from an API and saves them to database
	ImportFromAPI(ctx context.Context, apiURL string) (int, error)

	// Disconnect closes the database connection
	Disconnect(ctx context.Context) error
}