package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
)

// ObjectHandler handles HTTP requests for object operations
type ObjectHandler struct {
	objectService *service.ObjectService
}

// NewObjectHandler creates a new ObjectHandler with the given object service
func NewObjectHandler(objectService *service.ObjectService) *ObjectHandler {
	return &ObjectHandler{
		objectService: objectService,
	}
}

// ListObjects handles GET /storage/v1/b/{bucket}/o - List objects in a bucket
func (h *ObjectHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	// Get bucket name from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// List objects in the bucket
	objects, err := h.objectService.List(params.Bucket)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	// Return GCP-compatible response
	response := models.ObjectListResponse{
		Kind:  "storage#objects",
		Items: objects,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UploadObject handles POST /upload/storage/v1/b/{bucket}/o - Upload object (multipart or simple)
func (h *ObjectHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	// Get bucket name from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}

	// Get object name from query parameter
	objectName := r.URL.Query().Get("name")

	var content []byte
	var contentType string
	var err error

	// Check if this is a multipart upload
	mediaType, params2, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if strings.HasPrefix(mediaType, "multipart/") {
		content, contentType, objectName, err = h.handleMultipartUpload(r, params2, objectName)
	} else {
		content, contentType, err = h.handleSimpleUpload(r)
	}

	if err != nil {
		WriteGCPError(w, http.StatusBadRequest, err.Error(), "parseError")
		return
	}

	if objectName == "" {
		WriteGCPError(w, http.StatusBadRequest, "Required parameter 'name' is missing", "required")
		return
	}

	// Create the object
	object, err := h.objectService.Create(params.Bucket, objectName, content, contentType)
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
	json.NewEncoder(w).Encode(object)
}

// handleSimpleUpload handles a simple (non-multipart) upload
func (h *ObjectHandler) handleSimpleUpload(r *http.Request) ([]byte, string, error) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return content, contentType, nil
}

// handleMultipartUpload handles a multipart/related upload
func (h *ObjectHandler) handleMultipartUpload(r *http.Request, params map[string]string, objectName string) ([]byte, string, string, error) {
	boundary := params["boundary"]
	if boundary == "" {
		return nil, "", "", errors.New("missing boundary in Content-Type")
	}

	mr := multipart.NewReader(r.Body, boundary)

	var content []byte
	var contentType string
	var metadata struct {
		Name        string `json:"name"`
		ContentType string `json:"contentType"`
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", "", err
		}

		partContentType := part.Header.Get("Content-Type")
		if strings.HasPrefix(partContentType, "application/json") {
			// This is the metadata part
			if err := json.NewDecoder(part).Decode(&metadata); err != nil {
				return nil, "", "", err
			}
			if objectName == "" && metadata.Name != "" {
				objectName = metadata.Name
			}
			if metadata.ContentType != "" {
				contentType = metadata.ContentType
			}
		} else {
			// This is the content part
			content, err = io.ReadAll(part)
			if err != nil {
				return nil, "", "", err
			}
			if contentType == "" {
				contentType = partContentType
			}
		}
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, objectName, nil
}

// GetObject handles GET /storage/v1/b/{bucket}/o/{object} - Get object metadata
func (h *ObjectHandler) GetObject(w http.ResponseWriter, r *http.Request) {
	// Get bucket and object names from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}
	if params.Object == "" {
		WriteGCPError(w, http.StatusBadRequest, "Object name is required", "required")
		return
	}

	// Get the object
	object, err := h.objectService.Get(params.Bucket, params.Object)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		if errors.Is(err, service.ErrObjectNotFound) {
			WriteGCPError(w, http.StatusNotFound, "No such object: "+params.Bucket+"/"+params.Object, "notFound")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	// Check if alt=media query param is present for content download
	if r.URL.Query().Get("alt") == "media" {
		h.downloadObjectContent(w, object)
		return
	}

	// Return object metadata
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(object)
}

// downloadObjectContent writes the object content as the response
func (h *ObjectHandler) downloadObjectContent(w http.ResponseWriter, object *models.Object) {
	w.Header().Set("Content-Type", object.ContentType)
	w.Header().Set("Content-Length", object.Size)
	w.Header().Set("ETag", object.Etag)
	w.Header().Set("X-Goog-Generation", object.Generation)
	w.WriteHeader(http.StatusOK)
	w.Write(object.Content)
}

// DeleteObject handles DELETE /storage/v1/b/{bucket}/o/{object} - Delete object
func (h *ObjectHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	// Get bucket and object names from path params
	params := GetPathParams(r.Context())
	if params == nil || params.Bucket == "" {
		WriteGCPError(w, http.StatusBadRequest, "Bucket name is required", "required")
		return
	}
	if params.Object == "" {
		WriteGCPError(w, http.StatusBadRequest, "Object name is required", "required")
		return
	}

	// Delete the object
	err := h.objectService.Delete(params.Bucket, params.Object)
	if err != nil {
		if errors.Is(err, service.ErrBucketNotFound) {
			WriteGCPError(w, http.StatusNotFound, "The specified bucket does not exist.", "notFound")
			return
		}
		if errors.Is(err, service.ErrObjectNotFound) {
			WriteGCPError(w, http.StatusNotFound, "No such object: "+params.Bucket+"/"+params.Object, "notFound")
			return
		}
		WriteGCPError(w, http.StatusInternalServerError, err.Error(), "backendError")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers all object routes on the given router
func (h *ObjectHandler) RegisterRoutes(router *Router) {
	// List objects - GET /storage/v1/b/{bucket}/o
	router.HandleGCPRoute("/storage/v1/b/{bucket}/o", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListObjects(w, r)
		default:
			WriteGCPError(w, http.StatusMethodNotAllowed, "Method not allowed", "methodNotAllowed")
		}
	})

	// Upload object - POST /upload/storage/v1/b/{bucket}/o
	router.HandleGCPRoute("/upload/storage/v1/b/{bucket}/o", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.UploadObject(w, r)
		default:
			WriteGCPError(w, http.StatusMethodNotAllowed, "Method not allowed", "methodNotAllowed")
		}
	})

	// Single object operations - GET/DELETE /storage/v1/b/{bucket}/o/{object...}
	router.HandleGCPRoute("/storage/v1/b/{bucket}/o/{object...}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetObject(w, r)
		case http.MethodDelete:
			h.DeleteObject(w, r)
		default:
			WriteGCPError(w, http.StatusMethodNotAllowed, "Method not allowed", "methodNotAllowed")
		}
	})
}

