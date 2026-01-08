// Package server provides the HTTP server setup and configuration.
package server

import (
	"net/http"
	"time"

	"github.com/ksick/gcp-api-mock/internal/config"
	"github.com/ksick/gcp-api-mock/internal/handler"
	"github.com/ksick/gcp-api-mock/internal/middleware"
	"github.com/ksick/gcp-api-mock/internal/store"
)

// New creates and configures a new HTTP server with all routes and middleware.
func New(cfg *config.Config) *http.Server {
	// Initialize in-memory store
	dataStore := store.New()

	// Create router with all routes
	mux := newRouter(cfg, dataStore)

	// Apply middleware stack
	var h http.Handler = mux
	h = middleware.Logger(h)
	h = middleware.Recovery(h)
	h = middleware.RequestID(h)

	return &http.Server{
		Addr:         cfg.Address(),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// newRouter creates and configures the HTTP router with all application routes.
func newRouter(cfg *config.Config, _ *store.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Create handlers
	healthHandler := handler.NewHealth()

	// Health check routes
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// UI routes (HTMX templates)
	uiHandler := handler.NewUI(cfg)
	mux.HandleFunc("GET /", uiHandler.Index)

	// Future API routes will be registered here
	// Example pattern:
	// storageHandler := handler.NewStorage(dataStore)
	// mux.HandleFunc("GET /v1/b", storageHandler.ListBuckets)
	// mux.HandleFunc("POST /v1/b", storageHandler.CreateBucket)

	return mux
}
