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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Errors  []GCPErrorItem `json:"errors"`
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

	setGCPHeaders(w)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(gcpErr)
}

// WriteGCPJSON writes a GCP-compatible JSON response
func WriteGCPJSON(w http.ResponseWriter, code int, data interface{}) {
	setGCPHeaders(w)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteGCPNoContent writes a GCP-compatible no content response (204)
func WriteGCPNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// setGCPHeaders sets common GCP API response headers
func setGCPHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	w.Header().Set("Vary", "Origin, X-Origin")
	w.Header().Set("X-GUploader-UploadID", generateUploadID())
}

// generateUploadID generates a mock upload ID for GCP compatibility
func generateUploadID() string {
	return time.Now().Format("20060102150405") + "-mock"
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