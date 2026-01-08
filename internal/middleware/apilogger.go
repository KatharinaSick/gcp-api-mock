// Package middleware provides HTTP middleware for the GCP API Mock.
package middleware

import (
	"net/http"
	"strings"
)

// RequestLoggerFunc is a function type for logging requests to the UI.
type RequestLoggerFunc func(method, path string, status int)

// APILogger creates middleware that logs API requests (non-UI, non-static) to the request logger.
func APILogger(logFn RequestLoggerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Only log API requests (storage, sql), not UI or static files
			path := r.URL.Path
			if shouldLogRequest(path) {
				logFn(r.Method, path, wrapped.statusCode)
			}
		})
	}
}

// shouldLogRequest determines if a request should be logged to the UI.
// It logs storage and SQL API requests, but not UI or static file requests.
func shouldLogRequest(path string) bool {
	// Log Cloud Storage API requests
	if strings.HasPrefix(path, "/storage/") || strings.HasPrefix(path, "/upload/storage/") || strings.HasPrefix(path, "/download/storage/") {
		return true
	}
	// Log Cloud SQL API requests
	if strings.HasPrefix(path, "/sql/") {
		return true
	}
	return false
}
