package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// GCPError represents a GCP-compatible error response
type GCPError struct {
	Error GCPErrorDetail `json:"error"`
}

// GCPErrorDetail contains the error details
type GCPErrorDetail struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Errors  []GCPErrorItem   `json:"errors"`
}

// GCPErrorItem represents a single error item
type GCPErrorItem struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
}

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// CORSMiddleware adds CORS headers for dashboard requests
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// WriteGCPError writes a GCP-compatible error response
func WriteGCPError(w http.ResponseWriter, code int, message, reason string) {
	gcpErr := GCPError{
		Error: GCPErrorDetail{
			Code:    code,
			Message: message,
			Errors: []GCPErrorItem{
				{
					Message: message,
					Domain:  "global",
					Reason:  reason,
				},
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(gcpErr)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}