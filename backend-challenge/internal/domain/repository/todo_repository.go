package repository

import (
	"context"

	"backend-challenge/internal/domain/model"
)

// TodoRepository defines the interface for todo data access
type TodoRepository interface {
	// Create creates a new todo item in the database
	Create(ctx context.Context, todo *model.TodoItem) error

	// GetByID fetches a todo item by ID
	GetByID(ctx context.Context, id string) (*model.TodoItem, error)

	// Update updates a todo item in the database
	Update(ctx context.Context, todo *model.TodoItem) error

	// Delete removes a todo item from the database
	Delete(ctx context.Context, id string) error

	// List returns all todo items
	List(ctx context.Context) ([]*model.TodoItem, error)
	
	// FindByStatus returns all todo items with a specific status
	FindByStatus(ctx context.Context, status model.ItemStatus) ([]*model.TodoItem, error)
	
	// FindByTypeAndStatus returns all todo items with a specific type and status
	FindByTypeAndStatus(ctx context.Context, itemType model.ItemType, status model.ItemStatus) ([]*model.TodoItem, error)
	
	// UpdateStatus updates the status of a todo item
	UpdateStatus(ctx context.Context, id string, status model.ItemStatus) error
	
	// FindToReturn finds all todo items that should be returned to the main list
	FindToReturn(ctx context.Context, currentTime string) ([]*model.TodoItem, error)
}