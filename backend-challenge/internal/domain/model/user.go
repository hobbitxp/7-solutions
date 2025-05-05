package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"` // Never return password in JSON responses
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// RegisterUserInput represents the input for user registration
type RegisterUserInput struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}

// LoginUserInput represents the input for user login
type LoginUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// NewUser creates a new user from registration input
func NewUser(input *RegisterUserInput) (*User, error) {
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword verifies the provided password against the user's hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Update applies the provided updates to the user
func (u *User) Update(input *UpdateUserInput) {
	if input.Name != "" {
		u.Name = input.Name
	}
	if input.Email != "" {
		u.Email = input.Email
	}
}