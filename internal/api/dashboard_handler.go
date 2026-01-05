package api

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
)

//go:embed templates/*.html templates/partials/*.html
var templatesFS embed.FS

//go:embed static/css/* static/js/*
var staticFS embed.FS

// DashboardStats contains summary statistics for the dashboard
type DashboardStats struct {
	BucketCount int    `json:"bucketCount"`
	ObjectCount int    `json:"objectCount"`
	TotalSize   int64  `json:"totalSize"`
	SizeDisplay string `json:"sizeDisplay"`
}

// BucketWithObjects represents a bucket with its objects for the dashboard
type BucketWithObjects struct {
	Bucket      models.Bucket   `json:"bucket"`
	Objects     []models.Object `json:"objects"`
	ObjectCount int             `json:"objectCount"`
	TotalSize   int64           `json:"totalSize"`
	SizeDisplay string          `json:"sizeDisplay"`
	Expanded    bool            `json:"expanded"`
}

// DashboardData contains all data needed to render the dashboard
type DashboardData struct {
	ProjectID   string              `json:"projectId"`
	Buckets     []BucketWithObjects `json:"buckets"`
	Requests    []RequestLogEntry   `json:"requests"`
	Stats       DashboardStats      `json:"stats"`
	AutoRefresh bool                `json:"autoRefresh"`
	RefreshRate int                 `json:"refreshRate"` // in seconds
}

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	bucketService *service.BucketService
	objectService *service.ObjectService
	requestLogger *RequestLogger
	templates     *template.Template
	projectID     string
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	bucketService *service.BucketService,
	objectService *service.ObjectService,
	requestLogger *RequestLogger,
	projectID string,
) *DashboardHandler {
	// Parse templates from embedded filesystem
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"statusClass": func(status int) string {
			switch {
			case status >= 500:
				return "status-error"
			case status >= 400:
				return "status-warn"
			case status >= 300:
				return "status-redirect"
			case status >= 200:
				return "status-success"
			default:
				return "status-info"
			}
		},
		"truncate": func(s string, max int) string {
			if len(s) <= max {
				return s
			}
			return s[:max-3] + "..."
		},
		"ago": func(t interface{}) string {
			// Returns relative time (not implemented for simplicity)
			return ""
		},
	}).ParseFS(templatesFS, "templates/*.html", "templates/partials/*.html"))

	return &DashboardHandler{
		bucketService: bucketService,
		objectService: objectService,
		requestLogger: requestLogger,
		templates:     tmpl,
		projectID:     projectID,
	}
}

// RegisterRoutes registers all dashboard routes
func (h *DashboardHandler) RegisterRoutes(router *Router) {
	// Dashboard page routes - use specific path, root handled via fallback
	router.HandleFunc("GET /dashboard", h.ServeDashboard)

	// Static files - serve from embedded filesystem
	staticSubFS, _ := fs.Sub(staticFS, "static")
	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticSubFS)))

	// API endpoints for dashboard
	router.HandleFunc("GET /api/dashboard/resources", h.GetResources)
	router.HandleFunc("GET /api/dashboard/requests", h.GetRequests)
	router.HandleFunc("GET /api/dashboard/stats", h.GetStats)

	// HTMX partial endpoints
	router.HandleFunc("GET /api/dashboard/partials/buckets", h.GetBucketsPartial)
	router.HandleFunc("GET /api/dashboard/partials/requests", h.GetRequestsPartial)
	router.HandleFunc("GET /api/dashboard/partials/stats", h.GetStatsPartial)
	router.HandleFunc("GET /api/dashboard/partials/all", h.GetAllPartials)

	// Dashboard CRUD endpoints
	router.HandleFunc("POST /api/dashboard/buckets", h.CreateBucket)
	router.HandleFunc("PATCH /api/dashboard/buckets/{bucket}", h.UpdateBucket)
	router.HandleFunc("DELETE /api/dashboard/buckets/{bucket}", h.DeleteBucket)
	router.HandleFunc("POST /api/dashboard/buckets/{bucket}/objects", h.UploadObject)
	router.HandleFunc("GET /api/dashboard/buckets/{bucket}/objects/{object...}", h.DownloadObject)
	router.HandleFunc("DELETE /api/dashboard/buckets/{bucket}/objects/{object...}", h.DeleteObject)
}

// ServeDashboard renders the main dashboard page
func (h *DashboardHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {

	data := h.buildDashboardData()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

// GetResources returns all dashboard resources as JSON
func (h *DashboardHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	data := h.buildDashboardData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// GetRequests returns the request log as JSON
func (h *DashboardHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	requests := h.requestLogger.GetAll()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(requests); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// GetStats returns dashboard statistics as JSON
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.calculateStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// GetBucketsPartial returns the buckets list as an HTML partial
func (h *DashboardHandler) GetBucketsPartial(w http.ResponseWriter, r *http.Request) {
	data := h.buildDashboardData()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "buckets", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// GetRequestsPartial returns the request log as an HTML partial
func (h *DashboardHandler) GetRequestsPartial(w http.ResponseWriter, r *http.Request) {
	requests := h.requestLogger.GetRecent(20)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "requests", map[string]interface{}{"Requests": requests}); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// GetStatsPartial returns the stats bar as an HTML partial
func (h *DashboardHandler) GetStatsPartial(w http.ResponseWriter, r *http.Request) {
	stats := h.calculateStats()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "stats", stats); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// GetAllPartials returns all dashboard partials with OOB swaps for atomic updates
func (h *DashboardHandler) GetAllPartials(w http.ResponseWriter, r *http.Request) {
	data := h.buildDashboardData()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "all", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// CreateBucket creates a new bucket from the dashboard
func (h *DashboardHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		Location     string `json:"location"`
		StorageClass string `json:"storageClass"`
	}

	// Support both JSON and form data
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
		req.Name = r.FormValue("name")
		req.Location = r.FormValue("location")
		req.StorageClass = r.FormValue("storageClass")
	}

	// Set defaults
	if req.Location == "" {
		req.Location = "US"
	}
	if req.StorageClass == "" {
		req.StorageClass = "STANDARD"
	}

	bucket, err := h.bucketService.Create(req.Name, h.projectID, req.Location, req.StorageClass)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If htmx request, return partial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshBuckets")
		h.GetBucketsPartial(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(bucket); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// UpdateBucket updates a bucket's metadata
func (h *DashboardHandler) UpdateBucket(w http.ResponseWriter, r *http.Request) {
	params := GetPathParamsFromRequest(r)
	if params == nil || params.Bucket == "" {
		h.sendError(w, http.StatusBadRequest, "Bucket name required")
		return
	}

	var req struct {
		Location     string `json:"location"`
		StorageClass string `json:"storageClass"`
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			h.sendError(w, http.StatusBadRequest, "Invalid form data")
			return
		}
		req.Location = r.FormValue("location")
		req.StorageClass = r.FormValue("storageClass")
	}

	updates := models.Bucket{
		Location:     req.Location,
		StorageClass: req.StorageClass,
	}

	bucket, err := h.bucketService.Update(params.Bucket, updates)
	if err != nil {
		if err == service.ErrBucketNotFound {
			h.sendError(w, http.StatusNotFound, "Bucket not found")
			return
		}
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If htmx request, return partial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshBuckets")
		h.GetBucketsPartial(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bucket); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// DeleteBucket deletes a bucket
func (h *DashboardHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	params := GetPathParamsFromRequest(r)
	if params == nil || params.Bucket == "" {
		h.sendError(w, http.StatusBadRequest, "Bucket name required")
		return
	}

	if err := h.bucketService.Delete(params.Bucket); err != nil {
		if err == service.ErrBucketNotFound {
			h.sendError(w, http.StatusNotFound, "Bucket not found")
			return
		}
		if err == service.ErrBucketNotEmpty {
			h.sendError(w, http.StatusConflict, "Bucket is not empty")
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If htmx request, return partial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshBuckets")
		h.GetBucketsPartial(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadObject uploads an object to a bucket
func (h *DashboardHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	params := GetPathParamsFromRequest(r)
	if params == nil || params.Bucket == "" {
		h.sendError(w, http.StatusBadRequest, "Bucket name required")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		h.sendError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Error reading file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	obj, err := h.objectService.Create(params.Bucket, header.Filename, content, contentType)
	if err != nil {
		if err == service.ErrBucketNotFound {
			h.sendError(w, http.StatusNotFound, "Bucket not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If htmx request, return partial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshBuckets")
		h.GetBucketsPartial(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// DownloadObject downloads an object from a bucket
func (h *DashboardHandler) DownloadObject(w http.ResponseWriter, r *http.Request) {
	params := GetPathParamsFromRequest(r)
	if params == nil || params.Bucket == "" || params.Object == "" {
		h.sendError(w, http.StatusBadRequest, "Bucket and object name required")
		return
	}

	obj, err := h.objectService.Get(params.Bucket, params.Object)
	if err != nil {
		if err == service.ErrBucketNotFound {
			h.sendError(w, http.StatusNotFound, "Bucket not found")
			return
		}
		if err == service.ErrObjectNotFound {
			h.sendError(w, http.StatusNotFound, "Object not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+obj.Name+"\"")
	w.Header().Set("Content-Length", obj.Size)
	if _, err := w.Write(obj.Content); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}

// DeleteObject deletes an object from a bucket
func (h *DashboardHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	params := GetPathParamsFromRequest(r)
	if params == nil || params.Bucket == "" || params.Object == "" {
		h.sendError(w, http.StatusBadRequest, "Bucket and object name required")
		return
	}

	if err := h.objectService.Delete(params.Bucket, params.Object); err != nil {
		if err == service.ErrBucketNotFound {
			h.sendError(w, http.StatusNotFound, "Bucket not found")
			return
		}
		if err == service.ErrObjectNotFound {
			h.sendError(w, http.StatusNotFound, "Object not found")
			return
		}
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If htmx request, return partial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "refreshBuckets")
		h.GetBucketsPartial(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// buildDashboardData builds the complete dashboard data structure
func (h *DashboardHandler) buildDashboardData() DashboardData {
	bucketsWithObjects := make([]BucketWithObjects, 0)

	// Get all buckets
	buckets, _ := h.bucketService.List(h.projectID)

	for _, bucket := range buckets {
		objects, _ := h.objectService.List(bucket.Name)

		// Sort objects by creation time (newest first), with name as secondary key for stability
		sort.SliceStable(objects, func(i, j int) bool {
			if objects[i].TimeCreated.Equal(objects[j].TimeCreated) {
				return objects[i].Name < objects[j].Name
			}
			return objects[i].TimeCreated.After(objects[j].TimeCreated)
		})

		var totalSize int64
		for _, obj := range objects {
			size, _ := strconv.ParseInt(obj.Size, 10, 64)
			totalSize += size
		}

		bucketsWithObjects = append(bucketsWithObjects, BucketWithObjects{
			Bucket:      bucket,
			Objects:     objects,
			ObjectCount: len(objects),
			TotalSize:   totalSize,
			SizeDisplay: formatBytes(totalSize),
			Expanded:    true, // Default to expanded
		})
	}

	// Sort buckets by creation time (newest first), with name as secondary key for stability
	sort.SliceStable(bucketsWithObjects, func(i, j int) bool {
		if bucketsWithObjects[i].Bucket.TimeCreated.Equal(bucketsWithObjects[j].Bucket.TimeCreated) {
			return bucketsWithObjects[i].Bucket.Name < bucketsWithObjects[j].Bucket.Name
		}
		return bucketsWithObjects[i].Bucket.TimeCreated.After(bucketsWithObjects[j].Bucket.TimeCreated)
	})

	return DashboardData{
		ProjectID:   h.projectID,
		Buckets:     bucketsWithObjects,
		Requests:    h.requestLogger.GetRecent(20),
		Stats:       h.calculateStats(),
		AutoRefresh: true,
		RefreshRate: 2,
	}
}

// calculateStats calculates dashboard statistics
func (h *DashboardHandler) calculateStats() DashboardStats {
	buckets, _ := h.bucketService.List(h.projectID)

	var objectCount int
	var totalSize int64

	for _, bucket := range buckets {
		objects, _ := h.objectService.List(bucket.Name)
		objectCount += len(objects)
		for _, obj := range objects {
			size, _ := strconv.ParseInt(obj.Size, 10, 64)
			totalSize += size
		}
	}

	return DashboardStats{
		BucketCount: len(buckets),
		ObjectCount: objectCount,
		TotalSize:   totalSize,
		SizeDisplay: formatBytes(totalSize),
	}
}

// sendError sends an error response
func (h *DashboardHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "B"
}
