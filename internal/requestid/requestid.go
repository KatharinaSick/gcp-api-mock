// Package requestid provides utilities for generating and retrieving request IDs.
package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// contextKeyType is a custom type for context keys to avoid collisions.
type contextKeyType string

// ContextKey is the context key for request IDs.
const ContextKey contextKeyType = "request_id"

// Generate creates a new unique request ID.
func Generate() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a simple ID if crypto/rand fails
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// FromContext retrieves the request ID from context.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKey).(string); ok {
		return id
	}
	return ""
}
