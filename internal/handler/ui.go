// Package handler provides HTTP handlers for the GCP API Mock.
package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ksick/gcp-api-mock/internal/config"
	"github.com/ksick/gcp-api-mock/internal/sqladmin"
	"github.com/ksick/gcp-api-mock/internal/storage"
	"github.com/ksick/gcp-api-mock/internal/store"
)

// RequestLogEntry represents a single API request log entry.
type RequestLogEntry struct {
	Timestamp   string
	Method      string
	MethodLower string
	Path        string
	Status      int
	Success     bool
}

// RequestLogger stores API request logs for the UI.
type RequestLogger struct {
	mu      sync.RWMutex
	entries []RequestLogEntry
	maxSize int
}

// NewRequestLogger creates a new request logger.
func NewRequestLogger(maxSize int) *RequestLogger {
	return &RequestLogger{
		entries: make([]RequestLogEntry, 0),
		maxSize: maxSize,
	}
}

// Add adds a new log entry.
func (rl *RequestLogger) Add(method, path string, status int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry := RequestLogEntry{
		Timestamp:   time.Now().Format("15:04:05"),
		Method:      method,
		MethodLower: strings.ToLower(method),
		Path:        path,
		Status:      status,
		Success:     status >= 200 && status < 400,
	}

	// Prepend new entry (newest first)
	rl.entries = append([]RequestLogEntry{entry}, rl.entries...)

	// Trim to max size
	if len(rl.entries) > rl.maxSize {
		rl.entries = rl.entries[:rl.maxSize]
	}
}

// GetAll returns all log entries.
func (rl *RequestLogger) GetAll() []RequestLogEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make([]RequestLogEntry, len(rl.entries))
	copy(result, rl.entries)
	return result
}

// Clear removes all log entries.
func (rl *RequestLogger) Clear() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.entries = make([]RequestLogEntry, 0)
}

// UI handles web UI endpoints with HTMX templates.
type UI struct {
	cfg       *config.Config
	templates *template.Template
	store     *store.Store
	logger    *RequestLogger
}

// NewUI creates a new UI handler.
func NewUI(cfg *config.Config, dataStore *store.Store, logger *RequestLogger) *UI {
	// Parse all templates from the templates directory
	tmpl := template.Must(template.ParseGlob(filepath.Join("web", "templates", "*.html")))

	return &UI{
		cfg:       cfg,
		templates: tmpl,
		store:     dataStore,
		logger:    logger,
	}
}

// GetLogger returns the request logger for use in middleware.
func (u *UI) GetLogger() *RequestLogger {
	return u.logger
}

// PageData holds common data passed to templates.
type PageData struct {
	Title       string
	Environment string
}

// Index renders the main dashboard page.
func (u *UI) Index(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "GCP API Mock",
		Environment: u.cfg.Environment,
	}

	if err := u.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

// ListBucketsUI renders the bucket list partial for HTMX.
func (u *UI) ListBucketsUI(w http.ResponseWriter, r *http.Request) {
	buckets := u.store.ListBuckets()

	if err := u.templates.ExecuteTemplate(w, "buckets.html", buckets); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

// CreateBucketUI handles bucket creation from the UI form.
func (u *UI) CreateBucketUI(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	location := r.FormValue("location")
	storageClass := r.FormValue("storageClass")

	if name == "" {
		http.Error(w, "bucket name is required", http.StatusBadRequest)
		return
	}

	req := &storage.BucketInsertRequest{
		Name:         name,
		Location:     location,
		StorageClass: storageClass,
	}

	_, err := u.store.CreateBucket(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Log the request
	u.logger.Add("POST", "/storage/v1/b", http.StatusOK)

	// Return updated bucket list
	u.ListBucketsUI(w, r)
}

// DeleteBucketUI handles bucket deletion from the UI.
func (u *UI) DeleteBucketUI(w http.ResponseWriter, r *http.Request) {
	// Extract bucket name from path: /ui/buckets/{bucket}
	path := r.URL.Path
	bucketName := strings.TrimPrefix(path, "/ui/buckets/")

	if bucketName == "" {
		http.Error(w, "bucket name is required", http.StatusBadRequest)
		return
	}

	err := u.store.DeleteBucket(bucketName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Log the request
	u.logger.Add("DELETE", "/storage/v1/b/"+bucketName, http.StatusNoContent)

	// Return updated bucket list
	u.ListBucketsUI(w, r)
}

// ListSQLInstancesUI renders the SQL instance list partial for HTMX.
func (u *UI) ListSQLInstancesUI(w http.ResponseWriter, r *http.Request) {
	instances := u.store.ListSQLInstances()

	if err := u.templates.ExecuteTemplate(w, "sql_instances.html", instances); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

// CreateSQLInstanceUI handles SQL instance creation from the UI form.
func (u *UI) CreateSQLInstanceUI(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	databaseVersion := r.FormValue("databaseVersion")
	region := r.FormValue("region")
	tier := r.FormValue("tier")

	if name == "" {
		http.Error(w, "instance name is required", http.StatusBadRequest)
		return
	}

	req := &sqladmin.InstanceInsertRequest{
		Name:            name,
		DatabaseVersion: databaseVersion,
		Region:          region,
		Settings: &sqladmin.Settings{
			Tier: tier,
		},
	}

	_, _, err := u.store.CreateSQLInstance(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Log the request
	u.logger.Add("POST", "/sql/v1beta4/projects/mock-project/instances", http.StatusOK)

	// Return updated instance list
	u.ListSQLInstancesUI(w, r)
}

// DeleteSQLInstanceUI handles SQL instance deletion from the UI.
func (u *UI) DeleteSQLInstanceUI(w http.ResponseWriter, r *http.Request) {
	// Extract instance name from path: /ui/sql/instances/{instance}
	path := r.URL.Path
	instanceName := strings.TrimPrefix(path, "/ui/sql/instances/")

	if instanceName == "" {
		http.Error(w, "instance name is required", http.StatusBadRequest)
		return
	}

	_, err := u.store.DeleteSQLInstance(instanceName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the request
	u.logger.Add("DELETE", "/sql/v1beta4/projects/mock-project/instances/"+instanceName, http.StatusOK)

	// Return updated instance list
	u.ListSQLInstancesUI(w, r)
}

// GetLogsUI renders the request log partial for HTMX.
func (u *UI) GetLogsUI(w http.ResponseWriter, r *http.Request) {
	entries := u.logger.GetAll()

	if err := u.templates.ExecuteTemplate(w, "logs.html", entries); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}

// ClearLogsUI clears all request logs.
func (u *UI) ClearLogsUI(w http.ResponseWriter, r *http.Request) {
	u.logger.Clear()
	u.GetLogsUI(w, r)
}
