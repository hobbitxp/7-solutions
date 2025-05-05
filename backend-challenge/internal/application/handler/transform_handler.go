package handler

import (
	"encoding/json"
	"net/http"
	"io"

	"backend-challenge/internal/domain/model"
	"backend-challenge/internal/domain/service"
	"github.com/gorilla/mux"
)

// TransformHandler handles data transformation requests
type TransformHandler struct {
	transformService service.TransformService
}

// RegisterTransformHandler registers transform routes
func RegisterTransformHandler(r *mux.Router, transformService service.TransformService) {
	handler := &TransformHandler{
		transformService: transformService,
	}

	// Define routes
	r.HandleFunc("/api/transform/group-by-department", handler.GroupUsersByDepartment).Methods("POST")
	r.HandleFunc("/api/transform/fetch-and-transform", handler.FetchAndTransform).Methods("POST", "GET")
}

// GroupUsersByDepartment handles grouping users by department
func (h *TransformHandler) GroupUsersByDepartment(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var rawData struct {
		Users []model.ExternalUserData `json:"users"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		respondWithError(w, err, http.StatusBadRequest)
		return
	}

	// Create input
	input := &model.GroupUsersByDepartmentInput{
		Users: rawData.Users,
	}

	// Transform data
	result, err := h.transformService.GroupUsersByDepartment(r.Context(), input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, result, http.StatusOK)
}

// FetchAndTransform handles fetching and transforming user data
func (h *TransformHandler) FetchAndTransform(w http.ResponseWriter, r *http.Request) {
	var input model.FetchAndTransformInput

	// GET request: use query parameter
	if r.Method == http.MethodGet {
		input.APIURL = r.URL.Query().Get("apiUrl")
	} else {
		// POST request: parse from body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, err, http.StatusBadRequest)
			return
		}
		
		if len(body) > 0 {
			if err := json.Unmarshal(body, &input); err != nil {
				respondWithError(w, err, http.StatusBadRequest)
				return
			}
		}
	}

	// Fetch and transform data from database/API
	result, err := h.transformService.FetchAndTransform(r.Context(), &input)
	if err != nil {
		respondWithDomainError(w, err)
		return
	}

	respondWithJSON(w, result, http.StatusOK)
}