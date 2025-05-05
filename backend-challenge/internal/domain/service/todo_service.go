package service

import (
	"context"
	"time"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"
)

// Todo service errors are now defined in errors.go

// TodoService defines the todo business logic service
type TodoService interface {
	// Create creates a new todo item
	Create(ctx context.Context, input *model.CreateTodoInput) (*model.TodoItem, error)

	// GetByID fetches a todo item by ID
	GetByID(ctx context.Context, id string) (*model.TodoItem, error)

	// Update updates a todo item
	Update(ctx context.Context, id string, input *model.UpdateTodoInput) (*model.TodoItem, error)

	// Delete removes a todo item
	Delete(ctx context.Context, id string) error

	// List returns all todo items grouped by status and type
	List(ctx context.Context) (*model.TodosGrouped, error)
	
	// Click handles the click action on a todo item
	Click(ctx context.Context, id string) (*model.TodoItem, error)
	
	// TimeoutReturn handles the automatic return of a todo item to the main list
	TimeoutReturn(ctx context.Context, id string) error
	
	// ReturnTimedOutItems returns all todo items that should be returned to the main list
	ReturnTimedOutItems(ctx context.Context, currentTime string) (int, error)
}

// todoService implements TodoService
type todoService struct {
	repo repository.TodoRepository
}

// NewTodoService creates a new TodoService
func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{
		repo: repo,
	}
}

// Create creates a new todo item
func (s *todoService) Create(ctx context.Context, input *model.CreateTodoInput) (*model.TodoItem, error) {
	// Create new todo item
	todo := model.NewTodoItem(input)

	// Save to repository
	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// GetByID fetches a todo item by ID
func (s *todoService) GetByID(ctx context.Context, id string) (*model.TodoItem, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// Update updates a todo item
func (s *todoService) Update(ctx context.Context, id string, input *model.UpdateTodoInput) (*model.TodoItem, error) {
	// Get todo item
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTodoNotFound
	}

	// Update todo item
	todo.Update(input)

	// Save to repository
	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// Delete removes a todo item
func (s *todoService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidID
	}

	// Check if todo exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTodoNotFound
	}

	return s.repo.Delete(ctx, id)
}

// List returns all todo items grouped by status and type
func (s *todoService) List(ctx context.Context) (*model.TodosGrouped, error) {
	// Get all todo items
	todos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Group by status and type
	result := &model.TodosGrouped{
		Main:   make([]*model.TodoItem, 0),
		Column: make(map[model.ItemType][]*model.TodoItem),
	}

	for _, todo := range todos {
		if todo.Status == model.StatusMain {
			result.Main = append(result.Main, todo)
		} else if todo.Status == model.StatusColumn {
			if _, ok := result.Column[todo.Type]; !ok {
				result.Column[todo.Type] = make([]*model.TodoItem, 0)
			}
			result.Column[todo.Type] = append(result.Column[todo.Type], todo)
		}
	}

	return result, nil
}

// Click handles the click action on a todo item
func (s *todoService) Click(ctx context.Context, id string) (*model.TodoItem, error) {
	// Get todo item
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTodoNotFound
	}

	// Mark as clicked and update status
	todo.Click()

	// Save to repository
	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, err
	}

	// Schedule auto-return
	time.AfterFunc(5*time.Second, func() {
		// Create a background context since the HTTP context will be gone
		bgCtx := context.Background()
		// Call TimeoutReturn
		_ = s.TimeoutReturn(bgCtx, id)
	})

	return todo, nil
}

// TimeoutReturn handles the automatic return of a todo item to the main list
func (s *todoService) TimeoutReturn(ctx context.Context, id string) error {
	// Get todo item
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTodoNotFound
	}

	// Only return if still in COLUMN status
	if todo.Status == model.StatusColumn {
		// Return to main list
		todo.Return()

		// Save to repository
		if err := s.repo.Update(ctx, todo); err != nil {
			return err
		}
	}

	return nil
}

// ReturnTimedOutItems returns all todo items that should be returned to the main list
func (s *todoService) ReturnTimedOutItems(ctx context.Context, currentTime string) (int, error) {
	// Find items that need to be returned
	items, err := s.repo.FindToReturn(ctx, currentTime)
	if err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	// Return each item
	returnedCount := 0
	for _, item := range items {
		item.Return()
		if err := s.repo.Update(ctx, item); err != nil {
			// Log error but continue with other items
			continue
		}
		returnedCount++
	}

	return returnedCount, nil
}