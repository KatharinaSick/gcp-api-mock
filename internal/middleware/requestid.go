package middleware

import (
	"context"
	"net/http"

	"github.com/katharinasick/gcp-api-mock/internal/requestid"
)

// RequestID adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID in header
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = requestid.Generate()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", id)

		// Add to request context
		ctx := context.WithValue(r.Context(), requestid.ContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
