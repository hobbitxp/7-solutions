package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend-challenge/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
)

// Common errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrMissingToken     = errors.New("missing token")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthService defines the authentication service
type AuthService interface {
	// GenerateToken generates a JWT token for a user
	GenerateToken(user *model.User) (string, error)

	// ValidateToken validates a JWT token
	ValidateToken(tokenString string) (*JWTClaims, error)

	// ExtractTokenFromRequest extracts a token from an HTTP request
	ExtractTokenFromRequest(r *http.Request) (string, error)
}

// jwtAuthService implements the AuthService interface
type jwtAuthService struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTAuthService creates a new JWT auth service
func NewJWTAuthService(secretKey string, tokenDuration time.Duration) AuthService {
	return &jwtAuthService{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a JWT token for a user
func (s *jwtAuthService) GenerateToken(user *model.User) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(s.tokenDuration)

	// Create claims
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *jwtAuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.secretKey), nil
		},
	)

	if err != nil {
		// Check specific error types
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractTokenFromRequest extracts a token from an HTTP request
func (s *jwtAuthService) ExtractTokenFromRequest(r *http.Request) (string, error) {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
		return "", ErrInvalidToken
	}

	// Check query parameter
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, nil
	}

	return "", ErrMissingToken
}