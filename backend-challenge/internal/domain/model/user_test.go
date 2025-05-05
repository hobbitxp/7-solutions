package model

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	// Create input
	input := &RegisterUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Create user
	user, err := NewUser(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check user fields
	if user.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, user.Email)
	}
	if user.Password == input.Password {
		t.Errorf("Password should be hashed")
	}
	if user.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should not be zero")
	}
}

func TestHashPassword(t *testing.T) {
	// Hash password
	password := "password123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that password is hashed
	if hashedPassword == password {
		t.Errorf("Password should be hashed")
	}
}

func TestCheckPassword(t *testing.T) {
	// Create user with hashed password
	password := "password123"
	hashedPassword, _ := HashPassword(password)
	user := &User{
		ID:        "1",
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	// Test valid password
	if !user.CheckPassword(password) {
		t.Errorf("Expected valid password check")
	}

	// Test invalid password
	if user.CheckPassword("wrongpassword") {
		t.Errorf("Expected invalid password check")
	}
}

func TestUpdate(t *testing.T) {
	// Create user
	user := &User{
		ID:        "1",
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	// Create update input
	input := &UpdateUserInput{
		Name:  "New Name",
		Email: "new@example.com",
	}

	// Update user
	user.Update(input)

	// Check updated fields
	if user.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, user.Email)
	}

	// Test partial update
	partialInput := &UpdateUserInput{
		Name: "Another Name",
	}
	user.Update(partialInput)

	// Check that only name was updated
	if user.Name != partialInput.Name {
		t.Errorf("Expected name %s, got %s", partialInput.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, user.Email)
	}
}