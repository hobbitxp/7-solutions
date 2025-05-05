package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/7-solutions/backend-challenge/internal/domain/model"
	"github.com/7-solutions/backend-challenge/internal/domain/service"
	"github.com/7-solutions/backend-challenge/internal/infrastructure/auth"
	"github.com/gorilla/mux"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService service.UserService
	authService auth.AuthService
}

// RegisterUserHandler registers user routes
func RegisterUserHandler(r *mux.Router, userService service.UserService, authService auth.AuthService) {
	handler := &UserHandler{
		userService: userService,
		authService: authService,
	}

	// Define protected routes
	protected := r.PathPrefix("/api/users").Subrouter()
	protected.Use(handler.AuthMiddleware)

	// Register routes
	protected.HandleFunc("", handler.ListUsers).Methods("GET")
	protected.HandleFunc("/{id}", handler.GetUser).Methods("GET")
	protected.HandleFunc("/{id}", handler.UpdateUser).Methods("PUT")
	protected.HandleFunc("/{id}", handler.DeleteUser).Methods("DELETE")
}

// AuthMiddleware verifies JWT token
func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request
		tokenString, err := h.authService.ExtractTokenFromRequest(r)
		if err != nil {
			respondWithError(w, err, http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			respondWithError(w, err, http.StatusUnauthorized)
			return
		}

		// Store user ID in request context
		ctx := r.Context()
		r = r.WithContext(ctx)

		// Set user ID in header for use in handlers
		r.Header.Set("X-User-ID", claims.UserID)

		next.ServeHTTP(w, r)
	})
}

// GetUser handles get user by ID request
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user from service
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, user, http.StatusOK)
}

// ListUsers handles list users request
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	// Default values
	page, pageSize := 1, 10

	// Parse page
	if pageStr != "" {
		if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	// Parse page size
	if pageSizeStr != "" {
		if pageSizeVal, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	// Get users
	users, err := h.userService.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Get total count
	totalItems, err := h.userService.CountUsers(r.Context())
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Create response
	response := ResponseWithPagination{
		Data:       users,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
	}

	respondWithJSON(w, response, http.StatusOK)
}

// UpdateUser handles update user request
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if user is updating their own profile
	userID := r.Header.Get("X-User-ID")
	if userID != id {
		respondWithError(w, service.ErrUserNotFound, http.StatusNotFound)
		return
	}

	// Parse request body
	var input model.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Update user
	user, err := h.userService.UpdateUser(r.Context(), id, &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, user, http.StatusOK)
}

// DeleteUser handles delete user request
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if user is deleting their own profile
	userID := r.Header.Get("X-User-ID")
	if userID != id {
		respondWithError(w, service.ErrUserNotFound, http.StatusNotFound)
		return
	}

	// Delete user
	err := h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Create success response
	response := SuccessResponse{
		Message: "User deleted successfully",
	}

	respondWithJSON(w, response, http.StatusOK)
}