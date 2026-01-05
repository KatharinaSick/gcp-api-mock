package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KatharinaSick/gcp-api-mock/internal/service"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

func setupTestDashboard() (*DashboardHandler, *Router) {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})

	bucketSvc := service.NewBucketService(factory)
	objectSvc := service.NewObjectService(factory)
	bucketSvc.SetObjectService(objectSvc)
	objectSvc.SetBucketService(bucketSvc)

	requestLogger := NewRequestLogger(50)
	handler := NewDashboardHandler(bucketSvc, objectSvc, requestLogger, "test-project")

	router := NewRouter()
	handler.RegisterRoutes(router)

	return handler, router
}

func TestDashboardHandler_GetResources_Empty(t *testing.T) {
	_, router := setupTestDashboard()

	req := httptest.NewRequest("GET", "/api/dashboard/resources", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(data.Buckets) != 0 {
		t.Errorf("Expected 0 buckets, got %d", len(data.Buckets))
	}

	if data.Stats.BucketCount != 0 {
		t.Errorf("Expected bucket count 0, got %d", data.Stats.BucketCount)
	}
}

func TestDashboardHandler_CreateBucket_JSON(t *testing.T) {
	_, router := setupTestDashboard()

	body := `{"name": "my-bucket", "location": "US-CENTRAL1", "storageClass": "NEARLINE"}`
	req := httptest.NewRequest("POST", "/api/dashboard/buckets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify bucket was created
	req = httptest.NewRequest("GET", "/api/dashboard/resources", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var data DashboardData
	json.NewDecoder(w.Body).Decode(&data)

	if len(data.Buckets) != 1 {
		t.Errorf("Expected 1 bucket, got %d", len(data.Buckets))
	}

	if data.Buckets[0].Bucket.Name != "my-bucket" {
		t.Errorf("Expected bucket name 'my-bucket', got '%s'", data.Buckets[0].Bucket.Name)
	}
}

func TestDashboardHandler_CreateBucket_FormData(t *testing.T) {
	_, router := setupTestDashboard()

	form := strings.NewReader("name=form-bucket&location=EU&storageClass=STANDARD")
	req := httptest.NewRequest("POST", "/api/dashboard/buckets", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDashboardHandler_DeleteBucket(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create a bucket first
	handler.bucketService.Create("to-delete", "test-project", "US", "STANDARD")

	// Delete the bucket
	req := httptest.NewRequest("DELETE", "/api/dashboard/buckets/to-delete", nil)
	// Set path params in context
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "to-delete"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify bucket is deleted
	req = httptest.NewRequest("GET", "/api/dashboard/resources", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var data DashboardData
	json.NewDecoder(w.Body).Decode(&data)

	if len(data.Buckets) != 0 {
		t.Errorf("Expected 0 buckets after delete, got %d", len(data.Buckets))
	}
}

func TestDashboardHandler_DeleteBucket_NotFound(t *testing.T) {
	_, router := setupTestDashboard()

	req := httptest.NewRequest("DELETE", "/api/dashboard/buckets/nonexistent", nil)
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "nonexistent"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDashboardHandler_UpdateBucket(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create a bucket first
	handler.bucketService.Create("update-test", "test-project", "US", "STANDARD")

	// Update the bucket
	body := `{"storageClass": "NEARLINE"}`
	req := httptest.NewRequest("PATCH", "/api/dashboard/buckets/update-test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "update-test"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify update
	bucket, _ := handler.bucketService.Get("update-test")
	if bucket.StorageClass != "NEARLINE" {
		t.Errorf("Expected storage class 'NEARLINE', got '%s'", bucket.StorageClass)
	}
}

func TestDashboardHandler_UploadObject(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create a bucket first
	handler.bucketService.Create("upload-test", "test-project", "US", "STANDARD")

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.WriteString(part, "Hello, World!")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/dashboard/buckets/upload-test/objects", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "upload-test"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify object was created
	objects, _ := handler.objectService.List("upload-test")
	if len(objects) != 1 {
		t.Fatalf("Expected 1 object, got %d", len(objects))
	}
	if objects[0].Name != "test.txt" {
		t.Errorf("Expected object name 'test.txt', got '%s'", objects[0].Name)
	}
}

func TestDashboardHandler_DownloadObject(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create a bucket and object
	handler.bucketService.Create("download-test", "test-project", "US", "STANDARD")
	handler.objectService.Create("download-test", "myfile.txt", []byte("Hello!"), "text/plain")

	req := httptest.NewRequest("GET", "/api/dashboard/buckets/download-test/objects/myfile.txt", nil)
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "download-test", Object: "myfile.txt"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	if w.Body.String() != "Hello!" {
		t.Errorf("Expected body 'Hello!', got '%s'", w.Body.String())
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected content-type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestDashboardHandler_DeleteObject(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create a bucket and object
	handler.bucketService.Create("delete-obj-test", "test-project", "US", "STANDARD")
	handler.objectService.Create("delete-obj-test", "to-delete.txt", []byte("Delete me"), "text/plain")

	req := httptest.NewRequest("DELETE", "/api/dashboard/buckets/delete-obj-test/objects/to-delete.txt", nil)
	ctx := SetPathParams(req.Context(), &PathParams{Bucket: "delete-obj-test", Object: "to-delete.txt"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify object is deleted
	objects, _ := handler.objectService.List("delete-obj-test")
	if len(objects) != 0 {
		t.Errorf("Expected 0 objects after delete, got %d", len(objects))
	}
}

func TestDashboardHandler_GetStats(t *testing.T) {
	handler, router := setupTestDashboard()

	// Create some data
	handler.bucketService.Create("stats-test-1", "test-project", "US", "STANDARD")
	handler.bucketService.Create("stats-test-2", "test-project", "US", "STANDARD")
	handler.objectService.Create("stats-test-1", "file1.txt", []byte("Hello"), "text/plain")
	handler.objectService.Create("stats-test-1", "file2.txt", []byte("World!"), "text/plain")

	req := httptest.NewRequest("GET", "/api/dashboard/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats DashboardStats
	json.NewDecoder(w.Body).Decode(&stats)

	if stats.BucketCount != 2 {
		t.Errorf("Expected 2 buckets, got %d", stats.BucketCount)
	}
	if stats.ObjectCount != 2 {
		t.Errorf("Expected 2 objects, got %d", stats.ObjectCount)
	}
	if stats.TotalSize != 11 { // "Hello" + "World!" = 5 + 6 = 11
		t.Errorf("Expected total size 11, got %d", stats.TotalSize)
	}
}

func TestDashboardHandler_GetRequests(t *testing.T) {
	handler, router := setupTestDashboard()

	// Log some requests
	handler.requestLogger.Log("GET", "/storage/v1/b", 200, 10)
	handler.requestLogger.Log("POST", "/storage/v1/b", 201, 15)

	req := httptest.NewRequest("GET", "/api/dashboard/requests", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var requests []RequestLogEntry
	json.NewDecoder(w.Body).Decode(&requests)

	if len(requests) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(requests))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}
