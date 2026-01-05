package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

func setupTestObjectHandler() (*ObjectHandler, *service.BucketService, *service.ObjectService, *store.StoreFactory) {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})
	bucketService := service.NewBucketService(factory)
	objectService := service.NewObjectService(factory)
	objectService.SetBucketService(bucketService)
	bucketService.SetObjectService(objectService)
	handler := NewObjectHandler(objectService)
	return handler, bucketService, objectService, factory
}

func TestObjectHandler_ListObjects(t *testing.T) {
	t.Run("returns empty list when no objects exist", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		// Create a bucket first
		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o", nil)
		w := httptest.NewRecorder()

		// Add bucket name to path params
		params := &PathParams{Bucket: "test-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.ListObjects(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.ObjectListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Kind != "storage#objects" {
			t.Errorf("expected kind 'storage#objects', got '%s'", response.Kind)
		}

		if len(response.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(response.Items))
		}
	})

	t.Run("returns objects in the bucket", func(t *testing.T) {
		handler, bucketService, objectService, _ := setupTestObjectHandler()

		// Create bucket and objects
		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")
		objectService.Create("test-bucket", "object1.txt", []byte("content1"), "text/plain")
		objectService.Create("test-bucket", "object2.txt", []byte("content2"), "text/plain")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.ListObjects(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.ObjectListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(response.Items))
		}
	})

	t.Run("returns error when bucket does not exist", func(t *testing.T) {
		handler, _, _, _ := setupTestObjectHandler()

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/non-existent-bucket/o", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "non-existent-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.ListObjects(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error when bucket name is missing", func(t *testing.T) {
		handler, _, _, _ := setupTestObjectHandler()

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b//o", nil)
		w := httptest.NewRecorder()

		handler.ListObjects(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestObjectHandler_UploadObject(t *testing.T) {
	t.Run("uploads object with simple upload", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		content := []byte("Hello, World!")
		req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o?name=test-file.txt", bytes.NewReader(content))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.UploadObject(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var object models.Object
		if err := json.NewDecoder(w.Body).Decode(&object); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if object.Name != "test-file.txt" {
			t.Errorf("expected name 'test-file.txt', got '%s'", object.Name)
		}
		if object.Kind != "storage#object" {
			t.Errorf("expected kind 'storage#object', got '%s'", object.Kind)
		}
		if object.ContentType != "text/plain" {
			t.Errorf("expected contentType 'text/plain', got '%s'", object.ContentType)
		}
		if object.Bucket != "test-bucket" {
			t.Errorf("expected bucket 'test-bucket', got '%s'", object.Bucket)
		}
	})

	t.Run("uploads object with multipart upload", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		// Create multipart request
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		// Add metadata part
		metadataHeader := make(map[string][]string)
		metadataHeader["Content-Type"] = []string{"application/json"}
		metadataPart, _ := writer.CreatePart(metadataHeader)
		metadataPart.Write([]byte(`{"name": "multipart-file.txt", "contentType": "text/plain"}`))

		// Add content part
		contentHeader := make(map[string][]string)
		contentHeader["Content-Type"] = []string{"text/plain"}
		contentPart, _ := writer.CreatePart(contentHeader)
		contentPart.Write([]byte("Multipart content"))

		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.UploadObject(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var object models.Object
		if err := json.NewDecoder(w.Body).Decode(&object); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if object.Name != "multipart-file.txt" {
			t.Errorf("expected name 'multipart-file.txt', got '%s'", object.Name)
		}
	})

	t.Run("returns error when object name is missing", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		content := []byte("Hello")
		req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o", bytes.NewReader(content))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.UploadObject(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error when bucket does not exist", func(t *testing.T) {
		handler, _, _, _ := setupTestObjectHandler()

		content := []byte("Hello")
		req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/non-existent-bucket/o?name=test.txt", bytes.NewReader(content))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "non-existent-bucket"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.UploadObject(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestObjectHandler_GetObject(t *testing.T) {
	t.Run("returns object metadata successfully", func(t *testing.T) {
		handler, bucketService, objectService, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")
		objectService.Create("test-bucket", "test-file.txt", []byte("Hello, World!"), "text/plain")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/test-file.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket", Object: "test-file.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.GetObject(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var object models.Object
		if err := json.NewDecoder(w.Body).Decode(&object); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if object.Name != "test-file.txt" {
			t.Errorf("expected name 'test-file.txt', got '%s'", object.Name)
		}
		if object.Kind != "storage#object" {
			t.Errorf("expected kind 'storage#object', got '%s'", object.Kind)
		}
	})

	t.Run("downloads object content with alt=media", func(t *testing.T) {
		handler, bucketService, objectService, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")
		objectService.Create("test-bucket", "test-file.txt", []byte("Hello, World!"), "text/plain")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/test-file.txt?alt=media", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket", Object: "test-file.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.GetObject(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Type") != "text/plain" {
			t.Errorf("expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
		}

		body, _ := io.ReadAll(w.Body)
		if string(body) != "Hello, World!" {
			t.Errorf("expected body 'Hello, World!', got '%s'", string(body))
		}
	})

	t.Run("returns error when object does not exist", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/non-existent.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket", Object: "non-existent.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.GetObject(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error when bucket does not exist", func(t *testing.T) {
		handler, _, _, _ := setupTestObjectHandler()

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/non-existent-bucket/o/test.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "non-existent-bucket", Object: "test.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.GetObject(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestObjectHandler_DeleteObject(t *testing.T) {
	t.Run("deletes object successfully", func(t *testing.T) {
		handler, bucketService, objectService, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")
		objectService.Create("test-bucket", "test-file.txt", []byte("Hello"), "text/plain")

		req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket/o/test-file.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket", Object: "test-file.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.DeleteObject(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		// Verify object is deleted
		_, err := objectService.Get("test-bucket", "test-file.txt")
		if err == nil {
			t.Error("expected object to be deleted")
		}
	})

	t.Run("returns error when object does not exist", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket/o/non-existent.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket", Object: "non-existent.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.DeleteObject(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error when bucket does not exist", func(t *testing.T) {
		handler, _, _, _ := setupTestObjectHandler()

		req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/non-existent-bucket/o/test.txt", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "non-existent-bucket", Object: "test.txt"}
		ctx := SetPathParams(req.Context(), params)
		req = req.WithContext(ctx)

		handler.DeleteObject(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestObjectHandler_RegisterRoutes(t *testing.T) {
	t.Run("routes are registered correctly", func(t *testing.T) {
		handler, bucketService, _, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		router := NewRouter()
		handler.RegisterRoutes(router)

		// Test list objects route
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for list objects, got %d", http.StatusOK, w.Code)
		}

		// Test upload route
		req = httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o?name=test.txt", bytes.NewReader([]byte("content")))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for upload, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		// Test get object route
		req = httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/test.txt", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for get object, got %d", http.StatusOK, w.Code)
		}

		// Test delete object route
		req = httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket/o/test.txt", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d for delete object, got %d", http.StatusNoContent, w.Code)
		}
	})

	t.Run("handles objects with path separators", func(t *testing.T) {
		handler, bucketService, objectService, _ := setupTestObjectHandler()

		bucketService.Create("test-bucket", "123456789", "US", "STANDARD")
		objectService.Create("test-bucket", "folder/subfolder/file.txt", []byte("content"), "text/plain")

		router := NewRouter()
		handler.RegisterRoutes(router)

		// Test get object with path separators
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/folder/subfolder/file.txt", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for object with path, got %d", http.StatusOK, w.Code)
		}

		var object models.Object
		if err := json.NewDecoder(w.Body).Decode(&object); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if object.Name != "folder/subfolder/file.txt" {
			t.Errorf("expected name 'folder/subfolder/file.txt', got '%s'", object.Name)
		}
	})
}

