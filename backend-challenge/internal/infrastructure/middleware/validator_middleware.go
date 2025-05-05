package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/7-solutions/backend-challenge/pkg/validator"
)

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

// Validate is a middleware that validates a request body against a model
func Validate(model interface{}) func(next http.Handler) http.Handler {
	// Get the type of the model
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new instance of the model
			modelValue := reflect.New(modelType).Interface()

			// Parse the request body
			if err := json.NewDecoder(r.Body).Decode(modelValue); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid request body: " + err.Error(),
				})
				return
			}

			// Validate the model
			v := validator.New()
			if err := v.ValidateStruct(modelValue); err != nil {
				var validationErr *validator.ValidationError
				if errors.As(err, &validationErr) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(ValidationErrorResponse{
						Error:   "Validation error",
						Details: v.GetErrorMessages(),
					})
					return
				}

				if err == validator.ErrValidation {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(ValidationErrorResponse{
						Error:   "Validation error",
						Details: v.GetErrorMessages(),
					})
					return
				}

				// Unknown validation error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Internal server error during validation",
				})
				return
			}

			// Store the validated model in the request context
			ctx := r.Context()
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}