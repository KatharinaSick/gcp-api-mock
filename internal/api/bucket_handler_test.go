package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

func setupTestBucketHandler() (*BucketHandler, *store.StoreFactory) {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})
	bucketService := service.NewBucketService(factory)
	handler := NewBucketHandler(bucketService)
	return handler, factory
}

func TestBucketHandler_ListBuckets(t *testing.T) {
	handler, _ := setupTestBucketHandler()

	t.Run("returns empty list when no buckets exist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b?project=test-project", nil)
		w := httptest.NewRecorder()

		handler.ListBuckets(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.BucketListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Kind != "storage#buckets" {
			t.Errorf("expected kind 'storage#buckets', got '%s'", response.Kind)
		}

		if len(response.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(response.Items))
		}
	})

	t.Run("returns error when project parameter is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b", nil)
		w := httptest.NewRecorder()

		handler.ListBuckets(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns buckets for the specified project", func(t *testing.T) {
		handler, factory := setupTestBucketHandler()
		bucketService := service.NewBucketService(factory)
		
		// Create test buckets
		_, _ = bucketService.Create("test-bucket-1", "123456789", "US", "STANDARD")
		_, _ = bucketService.Create("test-bucket-2", "123456789", "EU", "STANDARD")
		_, _ = bucketService.Create("other-project-bucket", "987654321", "US", "STANDARD")

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b?project=123456789", nil)
		w := httptest.NewRecorder()

		handler.ListBuckets(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response models.BucketListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(response.Items))
		}
	})
}

func TestBucketHandler_CreateBucket(t *testing.T) {
	t.Run("creates a new bucket successfully", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		body := map[string]interface{}{
			"name":         "test-bucket",
			"location":     "US",
			"storageClass": "STANDARD",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/storage/v1/b?project=123456789", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBucket(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var bucket models.Bucket
		if err := json.NewDecoder(w.Body).Decode(&bucket); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if bucket.Name != "test-bucket" {
			t.Errorf("expected name 'test-bucket', got '%s'", bucket.Name)
		}
		if bucket.Kind != "storage#bucket" {
			t.Errorf("expected kind 'storage#bucket', got '%s'", bucket.Kind)
		}
		if bucket.ProjectNumber != "123456789" {
			t.Errorf("expected projectNumber '123456789', got '%s'", bucket.ProjectNumber)
		}
	})

	t.Run("returns error when project parameter is missing", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		body := map[string]interface{}{
			"name": "test-bucket",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBucket(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for invalid bucket name", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		body := map[string]interface{}{
			"name": "ab", // Too short
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/storage/v1/b?project=123456789", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBucket(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error when bucket already exists", func(t *testing.T) {
		handler, factory := setupTestBucketHandler()
		bucketService := service.NewBucketService(factory)
		_, _ = bucketService.Create("existing-bucket", "123456789", "US", "STANDARD")

		body := map[string]interface{}{
			"name": "existing-bucket",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/storage/v1/b?project=123456789", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateBucket(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})
}

func TestBucketHandler_GetBucket(t *testing.T) {
	t.Run("returns bucket metadata successfully", func(t *testing.T) {
		handler, factory := setupTestBucketHandler()
		bucketService := service.NewBucketService(factory)
		_, _ = bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		// Create a request with bucket parameter in context
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket", nil)
		w := httptest.NewRecorder()

		// Add bucket name to path params
		params := &PathParams{Bucket: "test-bucket"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.GetBucket(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var bucket models.Bucket
		if err := json.NewDecoder(w.Body).Decode(&bucket); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if bucket.Name != "test-bucket" {
			t.Errorf("expected name 'test-bucket', got '%s'", bucket.Name)
		}
	})

	t.Run("returns 404 for non-existent bucket", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/nonexistent", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "nonexistent"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.GetBucket(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestBucketHandler_DeleteBucket(t *testing.T) {
	t.Run("deletes bucket successfully", func(t *testing.T) {
		handler, factory := setupTestBucketHandler()
		bucketService := service.NewBucketService(factory)
		_, _ = bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.DeleteBucket(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}

		// Verify bucket is deleted
		_, err := bucketService.Get("test-bucket")
		if err != service.ErrBucketNotFound {
			t.Errorf("expected bucket to be deleted")
		}
	})

	t.Run("returns 404 for non-existent bucket", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/nonexistent", nil)
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "nonexistent"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.DeleteBucket(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestBucketHandler_UpdateBucket(t *testing.T) {
	t.Run("updates bucket metadata successfully", func(t *testing.T) {
		handler, factory := setupTestBucketHandler()
		bucketService := service.NewBucketService(factory)
		_, _ = bucketService.Create("test-bucket", "123456789", "US", "STANDARD")

		body := map[string]interface{}{
			"storageClass": "NEARLINE",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/storage/v1/b/test-bucket", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "test-bucket"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.UpdateBucket(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var bucket models.Bucket
		if err := json.NewDecoder(w.Body).Decode(&bucket); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if bucket.StorageClass != "NEARLINE" {
			t.Errorf("expected storageClass 'NEARLINE', got '%s'", bucket.StorageClass)
		}
	})

	t.Run("returns 404 for non-existent bucket", func(t *testing.T) {
		handler, _ := setupTestBucketHandler()

		body := map[string]interface{}{
			"storageClass": "NEARLINE",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/storage/v1/b/nonexistent", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		params := &PathParams{Bucket: "nonexistent"}
		ctx := req.Context()
		ctx = SetPathParams(ctx, params)
		req = req.WithContext(ctx)

		handler.UpdateBucket(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}