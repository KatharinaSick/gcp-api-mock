// Package handler provides HTTP handlers for the GCP API Mock.
package handler

import (
	"encoding/json"
	"net/http"
)

// Health handles health check endpoints.
type Health struct{}

// NewHealth creates a new Health handler.
func NewHealth() *Health {
	return &Health{}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// Check handles the /health endpoint for liveness probes.
func (h *Health) Check(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Ready handles the /ready endpoint for readiness probes.
func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, HealthResponse{Status: "ready"})
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
