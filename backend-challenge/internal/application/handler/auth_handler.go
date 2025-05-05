package handler

import (
	"encoding/json"
	"net/http"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/service"
	"backend-challenge/internal/infrastructure/auth"
	"github.com/gorilla/mux"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService auth.AuthService
	userService service.UserService
}

// RegisterAuthHandler registers auth routes
func RegisterAuthHandler(r *mux.Router, authService auth.AuthService, userService service.UserService) {
	handler := &AuthHandler{
		authService: authService,
		userService: userService,
	}

	// Create a subrouter for auth endpoints
	authRouter := r.PathPrefix("/api/auth").Subrouter()
	
	// Register routes
	authRouter.HandleFunc("/register", handler.Register).Methods("POST")
	authRouter.HandleFunc("/login", handler.Login).Methods("POST")
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var input model.RegisterUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Register user
	user, err := h.userService.Register(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	// Create response
	responseData := TokenResponse{
		Token: token,
		User:  user,
	}

	respondWithJSON(w, responseData, http.StatusCreated)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var input model.LoginUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.userService.Login(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	// Create response
	responseData := TokenResponse{
		Token: token,
		User:  user,
	}

	respondWithJSON(w, responseData, http.StatusOK)
}