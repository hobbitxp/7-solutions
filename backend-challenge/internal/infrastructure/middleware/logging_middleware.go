package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests with method, path, and execution time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Create a custom response writer to capture status code
		lrw := newLoggingResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		log.Printf(
			"[%s] %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			lrw.statusCode,
			duration,
		)
	})
}

// loggingResponseWriter is a custom response writer that captures the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new logging response writer
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
	}
}

// WriteHeader captures the status code and forwards to the underlying ResponseWriter
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// PanicRecoveryMiddleware recovers from panics and logs the error
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				log.Printf("Panic recovered: %v", err)

				// Return a 500 Internal Server Error
				http.Error(
					w,
					fmt.Sprintf("Internal Server Error: %v", err),
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}