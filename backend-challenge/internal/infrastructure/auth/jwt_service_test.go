package auth

import (
	"net/http"
	"testing"
	"time"

	"backend-challenge/internal/domain/model"
)

func TestJWTAuthService(t *testing.T) {
	// Create auth service
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	authService := NewJWTAuthService(secretKey, tokenDuration)

	// Create user for testing
	user := &model.User{
		ID:        "user-123",
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	// Test GenerateToken
	token, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Errorf("Expected non-empty token")
	}

	// Test ValidateToken
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	// Test invalid token
	_, err = authService.ValidateToken("invalid-token")
	if err == nil {
		t.Errorf("Expected error for invalid token")
	}
}

func TestExtractTokenFromRequest(t *testing.T) {
	// Create auth service
	secretKey := "test-secret-key"
	tokenDuration := 1 * time.Hour
	authService := NewJWTAuthService(secretKey, tokenDuration)

	// Test extraction from Authorization header
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	token, err := authService.ExtractTokenFromRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token != "test-token" {
		t.Errorf("Expected token %s, got %s", "test-token", token)
	}

	// Test extraction from URL query parameter
	req, _ = http.NewRequest("GET", "/api/users?token=test-token-from-query", nil)
	token, err = authService.ExtractTokenFromRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token != "test-token-from-query" {
		t.Errorf("Expected token %s, got %s", "test-token-from-query", token)
	}

	// Test missing token
	req, _ = http.NewRequest("GET", "/api/users", nil)
	_, err = authService.ExtractTokenFromRequest(req)
	if err != ErrMissingToken {
		t.Errorf("Expected error %v, got %v", ErrMissingToken, err)
	}

	// Test invalid Bearer format
	req, _ = http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Invalid test-token")
	_, err = authService.ExtractTokenFromRequest(req)
	if err != ErrInvalidToken {
		t.Errorf("Expected error %v, got %v", ErrInvalidToken, err)
	}
}