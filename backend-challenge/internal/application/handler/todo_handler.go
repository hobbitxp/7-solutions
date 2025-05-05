package handler

import (
	"encoding/json"
	"net/http"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/service"
	"backend-challenge/internal/infrastructure/auth"
	"github.com/gorilla/mux"
)

// TodoHandler handles todo-related requests
type TodoHandler struct {
	todoService service.TodoService
}

// RegisterTodoHandler registers todo routes
func RegisterTodoHandler(r *mux.Router, todoService service.TodoService, authService auth.AuthService) {
	handler := &TodoHandler{
		todoService: todoService,
	}

	// Define protected routes
	protected := r.PathPrefix("/api/todos").Subrouter()
	protected.Use(createAuthMiddleware(authService))

	// Register routes
	protected.HandleFunc("", handler.ListTodos).Methods("GET")
	protected.HandleFunc("", handler.CreateTodo).Methods("POST")
	protected.HandleFunc("/{id}", handler.GetTodo).Methods("GET")
	protected.HandleFunc("/{id}", handler.UpdateTodo).Methods("PUT")
	protected.HandleFunc("/{id}", handler.DeleteTodo).Methods("DELETE")
	protected.HandleFunc("/{id}/click", handler.ClickTodo).Methods("POST")
}

// ListTodos handles the request to list all todos grouped by status and type
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	// Get todos
	todos, err := h.todoService.List(r.Context())
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, todos, http.StatusOK)
}

// CreateTodo handles the request to create a todo
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var input model.CreateTodoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Create todo
	todo, err := h.todoService.Create(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, todo, http.StatusCreated)
}

// GetTodo handles the request to get a todo by ID
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get todo
	todo, err := h.todoService.GetByID(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, todo, http.StatusOK)
}

// UpdateTodo handles the request to update a todo
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse request body
	var input model.UpdateTodoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Update todo
	todo, err := h.todoService.Update(r.Context(), id, &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, todo, http.StatusOK)
}

// DeleteTodo handles the request to delete a todo
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete todo
	err := h.todoService.Delete(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Create success response
	response := SuccessResponse{
		Message: "Todo deleted successfully",
	}

	respondWithJSON(w, response, http.StatusOK)
}

// ClickTodo handles the request to click a todo
func (h *TodoHandler) ClickTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Process click action
	todo, err := h.todoService.Click(r.Context(), id)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, todo, http.StatusOK)
}

// Helper function to create an auth middleware
func createAuthMiddleware(authService auth.AuthService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from request
			tokenString, err := authService.ExtractTokenFromRequest(r)
			if err != nil {
				respondWithError(w, err, http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
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
}