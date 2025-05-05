package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/7-solutions/backend-challenge/internal/domain/model"
	"github.com/7-solutions/backend-challenge/internal/domain/repository"
)

// Mock UserRepository for testing
type mockUserRepository struct {
	users map[string]*model.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	// Check if email already exists
	for _, existingUser := range m.users {
		if existingUser.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	// Set ID if not set
	if user.ID == "" {
		user.ID = "user-" + time.Now().Format(time.RFC3339Nano)
	}

	// Set creation time if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// Store user
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) Update(ctx context.Context, user *model.User) error {
	_, ok := m.users[user.ID]
	if !ok {
		return errors.New("user not found")
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	_, ok := m.users[id]
	if !ok {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	var users []*model.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepository) CountUsers(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *mockUserRepository) Disconnect(ctx context.Context) error {
	return nil
}

// Test RegisterUser
func TestRegisterUser(t *testing.T) {
	// Create mock repository
	repo := newMockUserRepository()

	// Create service
	service := NewUserService(repo)

	// Test case: successful register
	input := &model.RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := service.Register(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, user.Email)
	}

	// Test case: email already exists
	_, err = service.Register(context.Background(), input)
	if err != ErrEmailExists {
		t.Errorf("Expected error %v, got %v", ErrEmailExists, err)
	}
}

// Test Login
func TestLogin(t *testing.T) {
	// Create mock repository
	repo := newMockUserRepository()

	// Create service
	service := NewUserService(repo)

	// Create a user
	input := &model.RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, _ := service.Register(context.Background(), input)

	// Test case: successful login
	loginInput := &model.LoginUserInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	loggedInUser, err := service.Login(context.Background(), loginInput)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if loggedInUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, loggedInUser.ID)
	}

	// Test case: user not found
	loginInput.Email = "nonexistent@example.com"
	_, err = service.Login(context.Background(), loginInput)
	if err != ErrUserNotFound {
		t.Errorf("Expected error %v, got %v", ErrUserNotFound, err)
	}

	// Test case: invalid password
	loginInput.Email = "test@example.com"
	loginInput.Password = "wrongpassword"
	_, err = service.Login(context.Background(), loginInput)
	if err != ErrInvalidPassword {
		t.Errorf("Expected error %v, got %v", ErrInvalidPassword, err)
	}
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
	// Create mock repository
	repo := newMockUserRepository()

	// Create service
	service := NewUserService(repo)

	// Create a user
	input := &model.RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, _ := service.Register(context.Background(), input)

	// Test case: successful update
	updateInput := &model.UpdateUserInput{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}
	updatedUser, err := service.UpdateUser(context.Background(), user.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if updatedUser.Name != updateInput.Name {
		t.Errorf("Expected name %s, got %s", updateInput.Name, updatedUser.Name)
	}
	if updatedUser.Email != updateInput.Email {
		t.Errorf("Expected email %s, got %s", updateInput.Email, updatedUser.Email)
	}

	// Test case: user not found
	_, err = service.UpdateUser(context.Background(), "nonexistent-id", updateInput)
	if err != ErrUserNotFound {
		t.Errorf("Expected error %v, got %v", ErrUserNotFound, err)
	}

	// Test case: email already exists
	// Create another user
	anotherInput := &model.RegisterUserInput{
		Name:     "Another User",
		Email:    "another@example.com",
		Password: "password123",
	}
	anotherUser, _ := service.Register(context.Background(), anotherInput)

	// Try to update to the email of the first user
	updateInput.Email = "updated@example.com"
	_, err = service.UpdateUser(context.Background(), anotherUser.ID, updateInput)
	if err != ErrEmailExists {
		t.Errorf("Expected error %v, got %v", ErrEmailExists, err)
	}
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	// Create mock repository
	repo := newMockUserRepository()

	// Create service
	service := NewUserService(repo)

	// Create a user
	input := &model.RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, _ := service.Register(context.Background(), input)

	// Test case: successful delete
	err := service.DeleteUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify user was deleted
	_, err = service.GetByID(context.Background(), user.ID)
	if err != ErrUserNotFound {
		t.Errorf("Expected error %v, got %v", ErrUserNotFound, err)
	}

	// Test case: user not found
	err = service.DeleteUser(context.Background(), "nonexistent-id")
	if err != ErrUserNotFound {
		t.Errorf("Expected error %v, got %v", ErrUserNotFound, err)
	}
}