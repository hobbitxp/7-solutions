package service

import (
	"context"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/repository"
)

// User service errors are now defined in errors.go

// UserService defines the user business logic service
type UserService interface {
	// Register creates a new user
	Register(ctx context.Context, input *model.RegisterUserInput) (*model.User, error)

	// Login authenticates a user
	Login(ctx context.Context, input *model.LoginUserInput) (*model.User, error)

	// GetByID fetches a user by ID
	GetByID(ctx context.Context, id string) (*model.User, error)

	// UpdateUser updates a user
	UpdateUser(ctx context.Context, id string, input *model.UpdateUserInput) (*model.User, error)

	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id string) error

	// ListUsers returns all users with pagination
	ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error)

	// CountUsers returns the total number of users
	CountUsers(ctx context.Context) (int64, error)
}

// userService implements UserService
type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// Register creates a new user
func (s *userService) Register(ctx context.Context, input *model.RegisterUserInput) (*model.User, error) {
	// Check if email already exists
	existingUser, _ := s.repo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Create new user
	user, err := model.NewUser(input)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user
func (s *userService) Login(ctx context.Context, input *model.LoginUserInput) (*model.User, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.CheckPassword(input.Password) {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// GetByID fetches a user by ID
func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id string, input *model.UpdateUserInput) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// If email is being updated, check if it already exists
	if input.Email != "" && input.Email != user.Email {
		existingUser, _ := s.repo.GetByEmail(ctx, input.Email)
		if existingUser != nil {
			return nil, ErrEmailExists
		}
	}

	// Update user fields
	user.Update(input)

	// Save to repository
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser removes a user
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidID
	}

	// Check if user exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.repo.Delete(ctx, id)
}

// ListUsers returns all users with pagination
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.repo.List(ctx, page, pageSize)
}

// CountUsers returns the total number of users
func (s *userService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.CountUsers(ctx)
}