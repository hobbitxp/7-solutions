package repository

import (
	"context"

	"github.com/7-solutions/backend-challenge/internal/domain/model"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *model.User) error

	// GetByID fetches a user by ID
	GetByID(ctx context.Context, id string) (*model.User, error)

	// GetByEmail fetches a user by email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// Update updates a user in the database
	Update(ctx context.Context, user *model.User) error

	// Delete removes a user from the database
	Delete(ctx context.Context, id string) error

	// List returns all users with pagination
	List(ctx context.Context, page, pageSize int) ([]*model.User, error)

	// CountUsers returns the total number of users in the database
	CountUsers(ctx context.Context) (int64, error)

	// Disconnect closes the database connection
	Disconnect(ctx context.Context) error
}