package model

import (
	"time"

	"github.com/google/uuid"
)

// ItemType represents the type of todo item
type ItemType string

// ItemStatus represents the status of todo item
type ItemStatus string

const (
	// TypeFruit represents a fruit item
	TypeFruit ItemType = "Fruit"
	// TypeVegetable represents a vegetable item
	TypeVegetable ItemType = "Vegetable"

	// StatusMain represents an item in the main list
	StatusMain ItemStatus = "MAIN"
	// StatusColumn represents an item moved to its type column
	StatusColumn ItemStatus = "COLUMN"
)

// TodoItem represents a todo item in the system
type TodoItem struct {
	ID        string     `json:"id" bson:"_id,omitempty"`
	Type      ItemType   `json:"type" bson:"type"`
	Name      string     `json:"name" bson:"name"`
	Status    ItemStatus `json:"status" bson:"status"`
	ClickedAt time.Time  `json:"clicked_at,omitempty" bson:"clicked_at,omitempty"`
	ReturnAt  time.Time  `json:"return_at,omitempty" bson:"return_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
}

// CreateTodoInput represents the input for creating a todo item
type CreateTodoInput struct {
	Type ItemType `json:"type" validate:"required"`
	Name string   `json:"name" validate:"required,min=1,max=100"`
}

// UpdateTodoInput represents the input for updating a todo item
type UpdateTodoInput struct {
	Type ItemType `json:"type" validate:"omitempty"`
	Name string   `json:"name" validate:"omitempty,min=1,max=100"`
}

// TodosGrouped represents todos grouped by status and type
type TodosGrouped struct {
	Main   []*TodoItem           `json:"main"`
	Column map[ItemType][]*TodoItem `json:"column"`
}

// NewTodoItem creates a new todo item
func NewTodoItem(input *CreateTodoInput) *TodoItem {
	now := time.Now()
	return &TodoItem{
		ID:        uuid.New().String(),
		Type:      input.Type,
		Name:      input.Name,
		Status:    StatusMain,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates a todo item with the provided input
func (t *TodoItem) Update(input *UpdateTodoInput) {
	if input.Type != "" {
		t.Type = input.Type
	}
	if input.Name != "" {
		t.Name = input.Name
	}
	t.UpdatedAt = time.Now()
}

// Click marks a todo item as clicked and sets it to return after a specific duration
func (t *TodoItem) Click() {
	now := time.Now()
	t.ClickedAt = now
	t.ReturnAt = now.Add(5 * time.Second)
	t.Status = StatusColumn
	t.UpdatedAt = now
}

// Return marks a todo item as returned to the main list
func (t *TodoItem) Return() {
	t.Status = StatusMain
	t.UpdatedAt = time.Now()
}