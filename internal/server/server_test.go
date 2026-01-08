package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksick/gcp-api-mock/internal/config"
	"github.com/ksick/gcp-api-mock/internal/storage"
)

// changeToProjectRoot changes to the project root directory for tests.
// This is needed because templates are loaded relative to the working directory.
func changeToProjectRoot(t *testing.T) func() {
	t.Helper()

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Find project root by looking for go.mod
	dir := originalDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root")
		}
		dir = parent
	}

	// Change to project root
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}

	// Return cleanup function
	return func() {
		os.Chdir(originalDir)
	}
}

func TestServer_StorageRoutes(t *testing.T) {
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	cfg := &config.Config{}
	srv := New(cfg)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "list buckets",
			method:     http.MethodGet,
			path:       "/storage/v1/b",
			wantStatus: http.StatusOK,
		},
		{
			name:       "create bucket",
			method:     http.MethodPost,
			path:       "/storage/v1/b",
			body:       `{"name": "test-bucket"}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			rr := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestServer_BucketCRUD(t *testing.T) {
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	cfg := &config.Config{}
	srv := New(cfg)

	// Create bucket
	createReq := httptest.NewRequest(http.MethodPost, "/storage/v1/b",
		strings.NewReader(`{"name": "integration-bucket", "location": "US"}`))
	createReq.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, createReq)

	if rr.Code != http.StatusOK {
		t.Fatalf("create bucket failed: %d - %s", rr.Code, rr.Body.String())
	}

	var bucket storage.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&bucket); err != nil {
		t.Fatalf("failed to decode bucket: %v", err)
	}

	if bucket.Name != "integration-bucket" {
		t.Errorf("bucket name = %s, want integration-bucket", bucket.Name)
	}

	// Get bucket
	getReq := httptest.NewRequest(http.MethodGet, "/storage/v1/b/integration-bucket", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, getReq)

	if rr.Code != http.StatusOK {
		t.Errorf("get bucket failed: %d", rr.Code)
	}

	// List buckets
	listReq := httptest.NewRequest(http.MethodGet, "/storage/v1/b", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, listReq)

	if rr.Code != http.StatusOK {
		t.Errorf("list buckets failed: %d", rr.Code)
	}

	var bucketList storage.BucketList
	if err := json.NewDecoder(rr.Body).Decode(&bucketList); err != nil {
		t.Fatalf("failed to decode bucket list: %v", err)
	}

	if len(bucketList.Items) != 1 {
		t.Errorf("expected 1 bucket, got %d", len(bucketList.Items))
	}

	// Delete bucket
	deleteReq := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/integration-bucket", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, deleteReq)

	if rr.Code != http.StatusNoContent {
		t.Errorf("delete bucket failed: %d - %s", rr.Code, rr.Body.String())
	}

	// Verify bucket is gone
	getReq = httptest.NewRequest(http.MethodGet, "/storage/v1/b/integration-bucket", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, getReq)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected not found after delete, got %d", rr.Code)
	}
}

func TestServer_ObjectCRUD(t *testing.T) {
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	cfg := &config.Config{}
	srv := New(cfg)

	// Create bucket first
	createBucketReq := httptest.NewRequest(http.MethodPost, "/storage/v1/b",
		strings.NewReader(`{"name": "object-test-bucket"}`))
	createBucketReq.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, createBucketReq)

	if rr.Code != http.StatusOK {
		t.Fatalf("create bucket failed: %d", rr.Code)
	}

	// Upload object
	content := "Hello, World!"
	uploadReq := httptest.NewRequest(http.MethodPost,
		"/upload/storage/v1/b/object-test-bucket/o?name=test.txt",
		strings.NewReader(content))
	uploadReq.Header.Set("Content-Type", "text/plain")
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, uploadReq)

	if rr.Code != http.StatusOK {
		t.Fatalf("upload object failed: %d - %s", rr.Code, rr.Body.String())
	}

	var obj storage.Object
	if err := json.NewDecoder(rr.Body).Decode(&obj); err != nil {
		t.Fatalf("failed to decode object: %v", err)
	}

	if obj.Name != "test.txt" {
		t.Errorf("object name = %s, want test.txt", obj.Name)
	}

	// Get object metadata
	getReq := httptest.NewRequest(http.MethodGet, "/storage/v1/b/object-test-bucket/o/test.txt", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, getReq)

	if rr.Code != http.StatusOK {
		t.Errorf("get object failed: %d", rr.Code)
	}

	// Download object content
	downloadReq := httptest.NewRequest(http.MethodGet,
		"/storage/v1/b/object-test-bucket/o/test.txt?alt=media", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, downloadReq)

	if rr.Code != http.StatusOK {
		t.Errorf("download object failed: %d", rr.Code)
	}

	if rr.Body.String() != content {
		t.Errorf("object content = %s, want %s", rr.Body.String(), content)
	}

	// List objects
	listReq := httptest.NewRequest(http.MethodGet, "/storage/v1/b/object-test-bucket/o", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, listReq)

	if rr.Code != http.StatusOK {
		t.Errorf("list objects failed: %d", rr.Code)
	}

	// Delete object
	deleteReq := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/object-test-bucket/o/test.txt", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, deleteReq)

	if rr.Code != http.StatusNoContent {
		t.Errorf("delete object failed: %d", rr.Code)
	}

	// Verify object is gone
	getReq = httptest.NewRequest(http.MethodGet, "/storage/v1/b/object-test-bucket/o/test.txt", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, getReq)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected not found after delete, got %d", rr.Code)
	}
}

func TestServer_HealthEndpoints(t *testing.T) {
	cleanup := changeToProjectRoot(t)
	defer cleanup()

	cfg := &config.Config{}
	srv := New(cfg)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "health check",
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ready check",
			path:       "/ready",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
