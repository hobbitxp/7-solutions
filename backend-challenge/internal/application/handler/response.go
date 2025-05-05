package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/7-solutions/backend-challenge/internal/domain/service"
	"github.com/7-solutions/backend-challenge/internal/infrastructure/auth"
	repo "github.com/7-solutions/backend-challenge/internal/infrastructure/repository"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseWithPagination represents a response with pagination
type ResponseWithPagination struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalItems int64       `json:"totalItems"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	Token string `json:"token"`
	User  interface{} `json:"user"`
}

// mapErrorToHTTPStatus maps domain errors to HTTP status codes
func mapErrorToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrUserNotFound), errors.Is(err, repo.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrEmailExists), errors.Is(err, repo.ErrDuplicateEmail):
		return http.StatusConflict
	case errors.Is(err, service.ErrInvalidID):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrInvalidPassword):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrMissingToken), errors.Is(err, auth.ErrInvalidToken), 
		errors.Is(err, auth.ErrTokenExpired), errors.Is(err, auth.ErrInvalidSignature):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// respondWithError writes an error response
func respondWithError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: err.Error(),
	}

	json.NewEncoder(w).Encode(response)
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondWithError handles domain errors and maps them to HTTP responses
func respondWithDomainError(w http.ResponseWriter, err error) {
	status := mapErrorToHTTPStatus(err)
	respondWithError(w, err, status)
}