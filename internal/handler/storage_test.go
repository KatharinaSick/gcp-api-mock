package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/katharinasick/gcp-api-mock/internal/storage"
	"github.com/katharinasick/gcp-api-mock/internal/store"
)

func setupTestStorage() (*Storage, *store.Store) {
	s := store.New()
	return NewStorage(s), s
}

func TestStorage_ListBuckets_Empty(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b", nil)
	rr := httptest.NewRecorder()

	h.ListBuckets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp storage.BucketList
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "storage#buckets" {
		t.Errorf("expected kind 'storage#buckets', got '%s'", resp.Kind)
	}

	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

func TestStorage_CreateBucket(t *testing.T) {
	h, _ := setupTestStorage()

	body := `{"name": "test-bucket", "location": "US", "storageClass": "STANDARD"}`
	req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateBucket(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var bucket storage.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&bucket); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if bucket.Name != "test-bucket" {
		t.Errorf("expected name 'test-bucket', got '%s'", bucket.Name)
	}

	if bucket.Kind != "storage#bucket" {
		t.Errorf("expected kind 'storage#bucket', got '%s'", bucket.Kind)
	}
}

func TestStorage_CreateBucket_InvalidJSON(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateBucket(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestStorage_CreateBucket_MissingName(t *testing.T) {
	h, _ := setupTestStorage()

	body := `{"location": "US"}`
	req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateBucket(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestStorage_CreateBucket_Duplicate(t *testing.T) {
	h, s := setupTestStorage()

	// Create first bucket
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	// Try to create duplicate
	body := `{"name": "test-bucket"}`
	req := httptest.NewRequest(http.MethodPost, "/storage/v1/b", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateBucket(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestStorage_GetBucket(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket", nil)
	rr := httptest.NewRecorder()

	h.GetBucket(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var bucket storage.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&bucket); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if bucket.Name != "test-bucket" {
		t.Errorf("expected name 'test-bucket', got '%s'", bucket.Name)
	}
}

func TestStorage_GetBucket_NotFound(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/non-existent", nil)
	rr := httptest.NewRecorder()

	h.GetBucket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_UpdateBucket(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	body := `{"storageClass": "NEARLINE", "labels": {"env": "test"}}`
	req := httptest.NewRequest(http.MethodPut, "/storage/v1/b/test-bucket", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateBucket(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var bucket storage.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&bucket); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if bucket.StorageClass != "NEARLINE" {
		t.Errorf("expected storage class 'NEARLINE', got '%s'", bucket.StorageClass)
	}
}

func TestStorage_UpdateBucket_NotFound(t *testing.T) {
	h, _ := setupTestStorage()

	body := `{"storageClass": "NEARLINE"}`
	req := httptest.NewRequest(http.MethodPut, "/storage/v1/b/non-existent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateBucket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_DeleteBucket(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket", nil)
	rr := httptest.NewRecorder()

	h.DeleteBucket(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	// Verify bucket is gone
	if s.GetBucket("test-bucket") != nil {
		t.Error("bucket should be deleted")
	}
}

func TestStorage_DeleteBucket_NotFound(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/non-existent", nil)
	rr := httptest.NewRecorder()

	h.DeleteBucket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_DeleteBucket_NotEmpty(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test-object", "text/plain", []byte("hello"), nil)

	req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket", nil)
	rr := httptest.NewRecorder()

	h.DeleteBucket(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestStorage_ListBuckets_WithBuckets(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "bucket-a"})
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "bucket-b"})

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b", nil)
	rr := httptest.NewRecorder()

	h.ListBuckets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp storage.BucketList
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestStorage_ListObjects_Empty(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o", nil)
	rr := httptest.NewRecorder()

	h.ListObjects(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp storage.ObjectList
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Kind != "storage#objects" {
		t.Errorf("expected kind 'storage#objects', got '%s'", resp.Kind)
	}
}

func TestStorage_ListObjects_BucketNotFound(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/non-existent/o", nil)
	rr := httptest.NewRecorder()

	h.ListObjects(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_InsertObject(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	content := []byte("Hello, World!")
	req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o?name=test.txt", bytes.NewReader(content))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	h.InsertObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var obj storage.Object
	if err := json.NewDecoder(rr.Body).Decode(&obj); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if obj.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got '%s'", obj.Name)
	}

	if obj.Kind != "storage#object" {
		t.Errorf("expected kind 'storage#object', got '%s'", obj.Kind)
	}

	if obj.Size != uint64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), obj.Size)
	}
}

func TestStorage_InsertObject_BucketNotFound(t *testing.T) {
	h, _ := setupTestStorage()

	req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/non-existent/o?name=test.txt", bytes.NewReader([]byte("data")))
	rr := httptest.NewRecorder()

	h.InsertObject(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_InsertObject_MissingName(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o", bytes.NewReader([]byte("data")))
	rr := httptest.NewRecorder()

	h.InsertObject(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestStorage_InsertObject_MultipartRelated(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	// Create a multipart/related body like Terraform sends
	boundary := "boundary123"
	body := "--" + boundary + "\r\n" +
		"Content-Type: application/json\r\n\r\n" +
		`{"contentType":"application/json","metadata":{"key":"value"}}` + "\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: application/json\r\n\r\n" +
		`{"version":4,"terraform_version":"1.11.0"}` + "\r\n" +
		"--" + boundary + "--\r\n"

	req := httptest.NewRequest(http.MethodPost, "/upload/storage/v1/b/test-bucket/o?name=state.tfstate", strings.NewReader(body))
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)
	rr := httptest.NewRecorder()

	h.InsertObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var obj storage.Object
	if err := json.NewDecoder(rr.Body).Decode(&obj); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if obj.Name != "state.tfstate" {
		t.Errorf("expected name 'state.tfstate', got '%s'", obj.Name)
	}

	// Content type should be from the metadata JSON, not the outer multipart header
	if obj.ContentType != "application/json" {
		t.Errorf("expected content type 'application/json', got '%s'", obj.ContentType)
	}

	// Verify metadata was parsed
	if obj.Metadata["key"] != "value" {
		t.Errorf("expected metadata key='value', got '%s'", obj.Metadata["key"])
	}

	// Verify content was stored correctly
	content := s.GetObjectContent("test-bucket", "state.tfstate")
	expectedContent := `{"version":4,"terraform_version":"1.11.0"}`
	if string(content) != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, string(content))
	}
}

func TestStorage_GetObject(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test.txt", "text/plain", []byte("Hello"), nil)

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/test.txt", nil)
	rr := httptest.NewRecorder()

	h.GetObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var obj storage.Object
	if err := json.NewDecoder(rr.Body).Decode(&obj); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if obj.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got '%s'", obj.Name)
	}
}

func TestStorage_GetObject_NotFound(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/non-existent", nil)
	rr := httptest.NewRecorder()

	h.GetObject(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_GetObject_MediaDownload(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	content := []byte("Hello, World!")
	_, _ = s.CreateObject("test-bucket", "test.txt", "text/plain", content, nil)

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o/test.txt?alt=media", nil)
	rr := httptest.NewRecorder()

	h.GetObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got '%s'", rr.Header().Get("Content-Type"))
	}

	if rr.Body.String() != string(content) {
		t.Errorf("expected body '%s', got '%s'", string(content), rr.Body.String())
	}
}

func TestStorage_DownloadObject(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	content := []byte("Hello, World!")
	_, _ = s.CreateObject("test-bucket", "test.txt", "text/plain", content, nil)

	req := httptest.NewRequest(http.MethodGet, "/download/storage/v1/b/test-bucket/o/test.txt", nil)
	rr := httptest.NewRecorder()

	h.DownloadObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != string(content) {
		t.Errorf("expected body '%s', got '%s'", string(content), rr.Body.String())
	}
}

func TestStorage_UpdateObject(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test.txt", "text/plain", []byte("Hello"), nil)

	body := `{"metadata": {"key": "value"}}`
	req := httptest.NewRequest(http.MethodPut, "/storage/v1/b/test-bucket/o/test.txt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateObject(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var obj storage.Object
	if err := json.NewDecoder(rr.Body).Decode(&obj); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if obj.Metadata["key"] != "value" {
		t.Error("metadata not updated correctly")
	}
}

func TestStorage_UpdateObject_NotFound(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	body := `{"metadata": {}}`
	req := httptest.NewRequest(http.MethodPut, "/storage/v1/b/test-bucket/o/non-existent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.UpdateObject(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_DeleteObject(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test.txt", "text/plain", []byte("Hello"), nil)

	req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket/o/test.txt", nil)
	rr := httptest.NewRecorder()

	h.DeleteObject(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	// Verify object is gone
	if s.GetObject("test-bucket", "test.txt") != nil {
		t.Error("object should be deleted")
	}
}

func TestStorage_DeleteObject_NotFound(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	req := httptest.NewRequest(http.MethodDelete, "/storage/v1/b/test-bucket/o/non-existent", nil)
	rr := httptest.NewRecorder()

	h.DeleteObject(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestStorage_ListObjects_WithPrefix(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "folder/file1.txt", "text/plain", []byte("1"), nil)
	_, _ = s.CreateObject("test-bucket", "folder/file2.txt", "text/plain", []byte("2"), nil)
	_, _ = s.CreateObject("test-bucket", "other/file3.txt", "text/plain", []byte("3"), nil)

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o?prefix=folder/", nil)
	rr := httptest.NewRecorder()

	h.ListObjects(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp storage.ObjectList
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestStorage_ListObjects_WithDelimiter(t *testing.T) {
	h, s := setupTestStorage()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "file.txt", "text/plain", []byte("0"), nil)
	_, _ = s.CreateObject("test-bucket", "folder/file1.txt", "text/plain", []byte("1"), nil)
	_, _ = s.CreateObject("test-bucket", "folder/file2.txt", "text/plain", []byte("2"), nil)

	req := httptest.NewRequest(http.MethodGet, "/storage/v1/b/test-bucket/o?delimiter=/", nil)
	rr := httptest.NewRecorder()

	h.ListObjects(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp storage.ObjectList
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should only return the root level file
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item at root level, got %d", len(resp.Items))
	}

	// Should have folder prefix
	if len(resp.Prefixes) != 1 || resp.Prefixes[0] != "folder/" {
		t.Errorf("expected prefix 'folder/', got %v", resp.Prefixes)
	}
}

func TestExtractBucketName(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   string
	}{
		{"/storage/v1/b/test-bucket", "/storage/v1/b/", "test-bucket"},
		{"/storage/v1/b/my-bucket/o", "/storage/v1/b/", "my-bucket"},
		{"/storage/v1/b/", "/storage/v1/b/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractBucketName(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("extractBucketName(%s, %s) = %s, want %s", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestExtractBucketAndObjectNames(t *testing.T) {
	tests := []struct {
		path       string
		wantBucket string
		wantObject string
	}{
		{"/storage/v1/b/test-bucket/o/test.txt", "test-bucket", "test.txt"},
		{"/storage/v1/b/my-bucket/o/folder/file.txt", "my-bucket", "folder/file.txt"},
		{"/storage/v1/b/bucket/o/", "bucket", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			bucket, object := extractBucketAndObjectNames(tt.path)
			if bucket != tt.wantBucket || object != tt.wantObject {
				t.Errorf("extractBucketAndObjectNames(%s) = (%s, %s), want (%s, %s)",
					tt.path, bucket, object, tt.wantBucket, tt.wantObject)
			}
		})
	}
}
