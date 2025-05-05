package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// RegisterHealthHandler registers health check routes
func RegisterHealthHandler(r *mux.Router) {
	// Define open routes for health check
	r.HandleFunc("/api/health", HealthCheck).Methods("GET")
}

// HealthCheck handles the health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Message:   "Service is up and running",
	}

	respondWithJSON(w, response, http.StatusOK)
}