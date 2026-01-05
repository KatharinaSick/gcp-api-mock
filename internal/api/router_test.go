package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	assert.NotNil(t, router)
	assert.NotNil(t, router.mux)
}

func TestRouter_HandleFunc(t *testing.T) {
	router := NewRouter()

	// Register a simple handler
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Test that the route is registered and works
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestRouter_Handle(t *testing.T) {
	router := NewRouter()

	// Register a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Created"))
	})
	router.Handle("/resource", handler)

	// Test that the route is registered and works
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "Created", rec.Body.String())
}

func TestRouter_ServeHTTP_NotFound(t *testing.T) {
	router := NewRouter()

	// Request a route that doesn't exist
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// Tests for path parameter extraction
// These tests define the expected behavior for extracting bucket/object names from GCP API paths

func TestExtractPathParams_BucketOnly(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantBucket string
		wantObject string
		wantOK     bool
	}{
		{
			name:       "valid bucket path",
			path:       "/storage/v1/b/my-bucket",
			wantBucket: "my-bucket",
			wantObject: "",
			wantOK:     true,
		},
		{
			name:       "bucket with hyphens and numbers",
			path:       "/storage/v1/b/my-bucket-123",
			wantBucket: "my-bucket-123",
			wantObject: "",
			wantOK:     true,
		},
		{
			name:       "bucket path with trailing slash",
			path:       "/storage/v1/b/my-bucket/",
			wantBucket: "my-bucket",
			wantObject: "",
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, ok := ExtractPathParams(tt.path)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantBucket, params.Bucket)
				assert.Equal(t, tt.wantObject, params.Object)
			}
		})
	}
}

func TestExtractPathParams_BucketAndObject(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantBucket string
		wantObject string
		wantOK     bool
	}{
		{
			name:       "simple object path",
			path:       "/storage/v1/b/my-bucket/o/my-object.txt",
			wantBucket: "my-bucket",
			wantObject: "my-object.txt",
			wantOK:     true,
		},
		{
			name:       "object with nested path",
			path:       "/storage/v1/b/my-bucket/o/folder/subfolder/file.json",
			wantBucket: "my-bucket",
			wantObject: "folder/subfolder/file.json",
			wantOK:     true,
		},
		{
			name:       "object with special characters (url encoded)",
			path:       "/storage/v1/b/my-bucket/o/file%20with%20spaces.txt",
			wantBucket: "my-bucket",
			wantObject: "file with spaces.txt",
			wantOK:     true,
		},
		{
			name:       "object path with trailing slash",
			path:       "/storage/v1/b/my-bucket/o/folder/",
			wantBucket: "my-bucket",
			wantObject: "folder/",
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, ok := ExtractPathParams(tt.path)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantBucket, params.Bucket)
				assert.Equal(t, tt.wantObject, params.Object)
			}
		})
	}
}

func TestExtractPathParams_InvalidPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty path",
			path: "",
		},
		{
			name: "root path",
			path: "/",
		},
		{
			name: "non-storage path",
			path: "/compute/v1/projects/my-project",
		},
		{
			name: "incomplete bucket path",
			path: "/storage/v1/b/",
		},
		{
			name: "incomplete bucket path without trailing slash",
			path: "/storage/v1/b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := ExtractPathParams(tt.path)
			assert.False(t, ok)
		})
	}
}

// Tests for GCP-style route registration pattern
func TestRouter_GCPRoutePattern(t *testing.T) {
	router := NewRouter()

	// Register GCP-style routes
	var capturedParams *PathParams

	router.HandleGCPRoute("/storage/v1/b/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		capturedParams = GetPathParams(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	router.HandleGCPRoute("/storage/v1/b/{bucket}/o/{object...}", func(w http.ResponseWriter, r *http.Request) {
		capturedParams = GetPathParams(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	t.Run("bucket route", func(t *testing.T) {
		capturedParams = nil
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedParams)
		assert.Equal(t, "test-bucket", capturedParams.Bucket)
	})

	t.Run("object route", func(t *testing.T) {
		capturedParams = nil
		req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/path/to/file.txt", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedParams)
		assert.Equal(t, "test-bucket", capturedParams.Bucket)
		assert.Equal(t, "path/to/file.txt", capturedParams.Object)
	})
}

func TestGetPathParams_FromContext(t *testing.T) {
	params := &PathParams{
		Bucket: "my-bucket",
		Object: "my-object.txt",
	}

	ctx := context.WithValue(context.Background(), pathParamsKey, params)
	
	retrieved := GetPathParams(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, "my-bucket", retrieved.Bucket)
	assert.Equal(t, "my-object.txt", retrieved.Object)
}

func TestGetPathParams_NilContext(t *testing.T) {
	ctx := context.Background()
	
	retrieved := GetPathParams(ctx)
	assert.Nil(t, retrieved)
}

// Test for multiple HTTP methods on same path
func TestRouter_MultipleMethodsOnSamePath(t *testing.T) {
	router := NewRouter()

	router.HandleFunc("GET /resource", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("GET"))
	})
	router.HandleFunc("POST /resource", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("POST"))
	})
	router.HandleFunc("DELETE /resource", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("DELETE"))
	})

	tests := []struct {
		method   string
		expected string
	}{
		{http.MethodGet, "GET"},
		{http.MethodPost, "POST"},
		{http.MethodDelete, "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/resource", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.expected, rec.Body.String())
		})
	}
}