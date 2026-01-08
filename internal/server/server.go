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
func newRouter(cfg *config.Config, dataStore *store.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Create handlers
	healthHandler := handler.NewHealth()
	storageHandler := handler.NewStorage(dataStore)
	sqlAdminHandler := handler.NewSQLAdmin(dataStore)

	// Health check routes
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// UI routes (HTMX templates)
	uiHandler := handler.NewUI(cfg)
	mux.HandleFunc("GET /", uiHandler.Index)

	// Cloud Storage API routes
	// Bucket operations
	mux.HandleFunc("GET /storage/v1/b", storageHandler.ListBuckets)
	mux.HandleFunc("POST /storage/v1/b", storageHandler.CreateBucket)
	mux.HandleFunc("GET /storage/v1/b/{bucket}", storageHandler.GetBucket)
	mux.HandleFunc("PUT /storage/v1/b/{bucket}", storageHandler.UpdateBucket)
	mux.HandleFunc("DELETE /storage/v1/b/{bucket}", storageHandler.DeleteBucket)

	// Object operations
	mux.HandleFunc("GET /storage/v1/b/{bucket}/o", storageHandler.ListObjects)
	mux.HandleFunc("GET /storage/v1/b/{bucket}/o/{object...}", storageHandler.GetObject)
	mux.HandleFunc("PUT /storage/v1/b/{bucket}/o/{object...}", storageHandler.UpdateObject)
	mux.HandleFunc("DELETE /storage/v1/b/{bucket}/o/{object...}", storageHandler.DeleteObject)

	// Object upload (uses different path prefix)
	mux.HandleFunc("POST /upload/storage/v1/b/{bucket}/o", storageHandler.InsertObject)

	// Object download (alternative download endpoint)
	mux.HandleFunc("GET /download/storage/v1/b/{bucket}/o/{object...}", storageHandler.DownloadObject)

	// Cloud SQL Admin API routes
	// Note: Cloud SQL uses v1beta4 API (unlike Storage which uses v1)
	// Instance operations
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/instances", sqlAdminHandler.ListInstances)
	mux.HandleFunc("POST /sql/v1beta4/projects/{project}/instances", sqlAdminHandler.CreateInstance)
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/instances/{instance}", sqlAdminHandler.GetInstance)
	mux.HandleFunc("PATCH /sql/v1beta4/projects/{project}/instances/{instance}", sqlAdminHandler.UpdateInstance)
	mux.HandleFunc("DELETE /sql/v1beta4/projects/{project}/instances/{instance}", sqlAdminHandler.DeleteInstance)

	// Database operations
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/instances/{instance}/databases", sqlAdminHandler.ListDatabases)
	mux.HandleFunc("POST /sql/v1beta4/projects/{project}/instances/{instance}/databases", sqlAdminHandler.CreateDatabase)
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database}", sqlAdminHandler.GetDatabase)
	mux.HandleFunc("PATCH /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database}", sqlAdminHandler.UpdateDatabase)
	mux.HandleFunc("DELETE /sql/v1beta4/projects/{project}/instances/{instance}/databases/{database}", sqlAdminHandler.DeleteDatabase)

	// User operations
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/instances/{instance}/users", sqlAdminHandler.ListUsers)
	mux.HandleFunc("POST /sql/v1beta4/projects/{project}/instances/{instance}/users", sqlAdminHandler.CreateUser)
	mux.HandleFunc("PUT /sql/v1beta4/projects/{project}/instances/{instance}/users", sqlAdminHandler.UpdateUser)
	mux.HandleFunc("DELETE /sql/v1beta4/projects/{project}/instances/{instance}/users", sqlAdminHandler.DeleteUser)

	// Operation operations
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/operations", sqlAdminHandler.ListOperations)
	mux.HandleFunc("GET /sql/v1beta4/projects/{project}/operations/{operation}", sqlAdminHandler.GetOperation)

	return mux
}
