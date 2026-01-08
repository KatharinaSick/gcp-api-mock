// Package handler provides HTTP handlers for the GCP API Mock.
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ksick/gcp-api-mock/internal/storage"
	"github.com/ksick/gcp-api-mock/internal/store"
)

// Storage handles Cloud Storage API endpoints.
type Storage struct {
	store *store.Store
}

// NewStorage creates a new Storage handler.
func NewStorage(s *store.Store) *Storage {
	return &Storage{store: s}
}

// ListBuckets handles GET /storage/v1/b - List buckets in a project.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/list
func (h *Storage) ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets := h.store.ListBuckets()

	response := &storage.BucketList{
		Kind:  "storage#buckets",
		Items: buckets,
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateBucket handles POST /storage/v1/b - Create a new bucket.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/insert
func (h *Storage) CreateBucket(w http.ResponseWriter, r *http.Request) {
	var req storage.BucketInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body", "invalid")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	bucket, err := h.store.CreateBucket(&req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			respondError(w, http.StatusConflict, err.Error(), "conflict")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	respondJSON(w, http.StatusOK, bucket)
}

// GetBucket handles GET /storage/v1/b/{bucket} - Get bucket metadata.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/get
func (h *Storage) GetBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := extractBucketName(r.URL.Path, "/storage/v1/b/")

	if bucketName == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	bucket := h.store.GetBucket(bucketName)
	if bucket == nil {
		respondError(w, http.StatusNotFound, "Bucket not found", "notFound")
		return
	}

	respondJSON(w, http.StatusOK, bucket)
}

// UpdateBucket handles PUT /storage/v1/b/{bucket} - Update bucket metadata.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/update
func (h *Storage) UpdateBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := extractBucketName(r.URL.Path, "/storage/v1/b/")

	if bucketName == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	var req storage.BucketUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body", "invalid")
		return
	}

	bucket, err := h.store.UpdateBucket(bucketName, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error(), "notFound")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	respondJSON(w, http.StatusOK, bucket)
}

// DeleteBucket handles DELETE /storage/v1/b/{bucket} - Delete a bucket.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/delete
func (h *Storage) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := extractBucketName(r.URL.Path, "/storage/v1/b/")

	if bucketName == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	err := h.store.DeleteBucket(bucketName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error(), "notFound")
			return
		}
		if strings.Contains(err.Error(), "not empty") {
			respondError(w, http.StatusConflict, err.Error(), "conflict")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListObjects handles GET /storage/v1/b/{bucket}/o - List objects in a bucket.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/list
func (h *Storage) ListObjects(w http.ResponseWriter, r *http.Request) {
	// Extract bucket name from path like /storage/v1/b/{bucket}/o
	path := r.URL.Path
	bucketName := extractBucketFromObjectsPath(path)

	if bucketName == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Check if bucket exists
	if h.store.GetBucket(bucketName) == nil {
		respondError(w, http.StatusNotFound, "Bucket not found", "notFound")
		return
	}

	// Get query parameters
	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")

	objects, prefixes := h.store.ListObjects(bucketName, prefix, delimiter)

	response := &storage.ObjectList{
		Kind:     "storage#objects",
		Items:    objects,
		Prefixes: prefixes,
	}

	respondJSON(w, http.StatusOK, response)
}

// InsertObject handles POST /upload/storage/v1/b/{bucket}/o - Upload an object.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/insert
func (h *Storage) InsertObject(w http.ResponseWriter, r *http.Request) {
	// Extract bucket name from path like /upload/storage/v1/b/{bucket}/o
	path := r.URL.Path
	bucketName := extractBucketFromUploadPath(path)

	if bucketName == "" {
		respondError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Check if bucket exists
	if h.store.GetBucket(bucketName) == nil {
		respondError(w, http.StatusNotFound, "Bucket not found", "notFound")
		return
	}

	// Get object name from query parameter
	objectName := r.URL.Query().Get("name")
	if objectName == "" {
		respondError(w, http.StatusBadRequest, "Object name is required", "required")
		return
	}

	// URL decode the object name
	decodedName, err := url.QueryUnescape(objectName)
	if err == nil {
		objectName = decodedName
	}

	// Read content
	content, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read request body", "invalid")
		return
	}

	// Get content type from header
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Get metadata from query parameters (x-goog-meta-*)
	var metadata map[string]string
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "x-goog-meta-") && len(values) > 0 {
			if metadata == nil {
				metadata = make(map[string]string)
			}
			metaKey := strings.TrimPrefix(key, "x-goog-meta-")
			metadata[metaKey] = values[0]
		}
	}

	obj, err := h.store.CreateObject(bucketName, objectName, contentType, content, metadata)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	respondJSON(w, http.StatusOK, obj)
}

// GetObject handles GET /storage/v1/b/{bucket}/o/{object} - Get object metadata.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/get
func (h *Storage) GetObject(w http.ResponseWriter, r *http.Request) {
	bucketName, objectName := extractBucketAndObjectNames(r.URL.Path)

	if bucketName == "" || objectName == "" {
		respondError(w, http.StatusBadRequest, "Bucket and object names are required", "required")
		return
	}

	// URL decode the object name
	decodedName, err := url.QueryUnescape(objectName)
	if err == nil {
		objectName = decodedName
	}

	// Check if this is a media download request
	if r.URL.Query().Get("alt") == "media" {
		h.downloadObject(w, r, bucketName, objectName)
		return
	}

	obj := h.store.GetObject(bucketName, objectName)
	if obj == nil {
		respondError(w, http.StatusNotFound, "Object not found", "notFound")
		return
	}

	respondJSON(w, http.StatusOK, obj)
}

// downloadObject handles media downloads for objects.
func (h *Storage) downloadObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {
	obj := h.store.GetObject(bucketName, objectName)
	if obj == nil {
		respondError(w, http.StatusNotFound, "Object not found", "notFound")
		return
	}

	content := h.store.GetObjectContent(bucketName, objectName)
	if content == nil {
		respondError(w, http.StatusNotFound, "Object content not found", "notFound")
		return
	}

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", string(rune(len(content))))
	w.Header().Set("ETag", obj.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// DownloadObject handles GET /download/storage/v1/b/{bucket}/o/{object} - Download object content.
// This is an alternative download endpoint.
func (h *Storage) DownloadObject(w http.ResponseWriter, r *http.Request) {
	// Extract bucket and object names from path like /download/storage/v1/b/{bucket}/o/{object}
	path := strings.TrimPrefix(r.URL.Path, "/download/storage/v1/b/")
	parts := strings.SplitN(path, "/o/", 2)

	if len(parts) != 2 {
		respondError(w, http.StatusBadRequest, "Invalid path", "invalid")
		return
	}

	bucketName := parts[0]
	objectName := parts[1]

	// URL decode the object name
	decodedName, err := url.QueryUnescape(objectName)
	if err == nil {
		objectName = decodedName
	}

	h.downloadObject(w, r, bucketName, objectName)
}

// UpdateObject handles PUT /storage/v1/b/{bucket}/o/{object} - Update object metadata.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/update
func (h *Storage) UpdateObject(w http.ResponseWriter, r *http.Request) {
	bucketName, objectName := extractBucketAndObjectNames(r.URL.Path)

	if bucketName == "" || objectName == "" {
		respondError(w, http.StatusBadRequest, "Bucket and object names are required", "required")
		return
	}

	// URL decode the object name
	decodedName, err := url.QueryUnescape(objectName)
	if err == nil {
		objectName = decodedName
	}

	var reqBody struct {
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body", "invalid")
		return
	}

	obj, err := h.store.UpdateObject(bucketName, objectName, reqBody.Metadata)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error(), "notFound")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	respondJSON(w, http.StatusOK, obj)
}

// DeleteObject handles DELETE /storage/v1/b/{bucket}/o/{object} - Delete an object.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/delete
func (h *Storage) DeleteObject(w http.ResponseWriter, r *http.Request) {
	bucketName, objectName := extractBucketAndObjectNames(r.URL.Path)

	if bucketName == "" || objectName == "" {
		respondError(w, http.StatusBadRequest, "Bucket and object names are required", "required")
		return
	}

	// URL decode the object name
	decodedName, err := url.QueryUnescape(objectName)
	if err == nil {
		objectName = decodedName
	}

	err = h.store.DeleteObject(bucketName, objectName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error(), "notFound")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error(), "internalError")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondError writes a JSON error response matching the GCS API format.
func respondError(w http.ResponseWriter, statusCode int, message, reason string) {
	errResp := storage.APIError{
		Error: storage.ErrorDetails{
			Code:    statusCode,
			Message: message,
			Errors: []storage.ErrorReason{
				{
					Domain:  "global",
					Reason:  reason,
					Message: message,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)
}

// extractBucketName extracts the bucket name from a path like /storage/v1/b/{bucket}.
func extractBucketName(path, prefix string) string {
	path = strings.TrimPrefix(path, prefix)
	// Remove any trailing path segments
	if idx := strings.Index(path, "/"); idx >= 0 {
		path = path[:idx]
	}
	return path
}

// extractBucketFromObjectsPath extracts the bucket name from a path like /storage/v1/b/{bucket}/o.
func extractBucketFromObjectsPath(path string) string {
	path = strings.TrimPrefix(path, "/storage/v1/b/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractBucketFromUploadPath extracts the bucket name from a path like /upload/storage/v1/b/{bucket}/o.
func extractBucketFromUploadPath(path string) string {
	path = strings.TrimPrefix(path, "/upload/storage/v1/b/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractBucketAndObjectNames extracts bucket and object names from a path like /storage/v1/b/{bucket}/o/{object}.
func extractBucketAndObjectNames(path string) (string, string) {
	path = strings.TrimPrefix(path, "/storage/v1/b/")
	parts := strings.SplitN(path, "/o/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
