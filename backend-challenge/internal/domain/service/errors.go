package service

import "errors"

// Common errors shared across services
var (
	// User related errors
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidPassword = errors.New("invalid password")
	
	// Todo related errors
	ErrTodoNotFound    = errors.New("todo item not found")
	
	// Shared errors
	ErrInvalidID       = errors.New("invalid ID")
)