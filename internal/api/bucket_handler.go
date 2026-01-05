package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
)

// BucketHandler handles HTTP requests for bucket operations
type BucketHandler struct {
	bucketService *service.BucketService
}

// NewBucketHandler creates a new BucketHandler with the given bucket service
func NewBucketHandler(bucketService *service.BucketService) *BucketHandler {
	return &BucketHandler{
		bucketService: bucketService,
	}
}

// CreateBucketRequest represents the request body for creating a bucket
type CreateBucketRequest struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	StorageClass string `json:"storageClass"`
}

// ListBuckets handles GET /storage/v1/b - List buckets in a project
func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	// Get project from query parameter
	project := r.URL.Query().Get("project")
	if project == "" {
		WriteGCPError(w, http.StatusBadRequest, "Required parameter 'project' is missing", "required")
		return
	}

	// List buckets for the project
	buckets, err := h.bucketService.List(project)
	if err != nil {
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	// Return GCP-compatible response
	response := models.BucketListResponse{
		Kind:  "storage#buckets",
		Items: buckets,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateBucket handles POST /storage/v1/b - Create a new bucket
func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	// Get project from query parameter
	project := r.URL.Query().Get("project")
	if project == "" {
		WriteGCPError(w, http.StatusBadRequest, "Required parameter 'project' is missing", "required")
		return
	}

	// Parse request body
	var req CreateBucketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteGCPError(w, http.StatusBadRequest, "Invalid JSON in request body", "parseError")
		return
	}

	// Set defaults
	if req.Location == "" {
		req.Location = "US"
	}
	if req.StorageClass == "" {
		req.StorageClass = "STANDARD"
	}

	// Create the bucket
	bucket, err := h.bucketService.Create(req.Name, project, req.Location, req.StorageClass)
	if err != nil {
		if errors.Is(err, service.ErrBucketAlreadyExists) {
			WriteGCPError(w, http.StatusConflict, "You already own this bucket. Please select another name.", "conflict")
			return
		}
		// Validation errors
		WriteGCPError(w, http.StatusBadRequest, err.Error(), "invalid")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bucket)
}

// GetBucket handles GET /storage/v1/b/{bucket} - Get bucket metadata
func (h *BucketHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	// Get bucket name from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Get the bucket
	bucket, err := h.bucketService.Get(params.Bucket)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bucket)
}

// DeleteBucket handles DELETE /storage/v1/b/{bucket} - Delete a bucket
func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	// Get bucket name from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Delete the bucket
	err := h.bucketService.Delete(params.Bucket)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		if errors.Is(err, service.ErrBucketNotEmpty) {
			WriteGCPError(w, http.StatusConflict, "The bucket you tried to delete is not empty.", "conflict")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateBucket handles PATCH /storage/v1/b/{bucket} - Update bucket metadata
func (h *BucketHandler) UpdateBucket(w http.ResponseWriter, r *http.Request) {
	// Get bucket name from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Parse request body
	var updates models.Bucket
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		WriteGCPError(w, http.StatusBadRequest, "Invalid JSON in request body", "parseError")
		return
	}

	// Update the bucket
	bucket, err := h.bucketService.Update(params.Bucket, updates)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bucket)
}

// RegisterRoutes registers all bucket routes on the given router
func (h *BucketHandler) RegisterRoutes(router *Router) {
	// List buckets - GET /storage/v1/b
	router.HandleGCPRoute("/storage/v1/b", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListBuckets(w, r)
		case http.MethodPost:
			h.CreateBucket(w, r)
		default:
			WriteGCPError(w, http.StatusMethodNotAllowed, "Method not allowed", "methodNotAllowed")
		}
	})

	// Single bucket operations - GET/DELETE/PATCH /storage/v1/b/{bucket}
	router.HandleGCPRoute("/storage/v1/b/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetBucket(w, r)
		case http.MethodDelete:
			h.DeleteBucket(w, r)
		case http.MethodPatch:
			h.UpdateBucket(w, r)
		default:
			WriteGCPError(w, http.StatusMethodNotAllowed, "Method not allowed", "methodNotAllowed")
		}
	})
}