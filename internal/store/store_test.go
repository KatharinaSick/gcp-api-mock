package store

import (
	"testing"

	"github.com/ksick/gcp-api-mock/internal/storage"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("New() returned nil")
	}
	if s.buckets == nil {
		t.Error("buckets map not initialized")
	}
	if s.objects == nil {
		t.Error("objects map not initialized")
	}
}

func TestStore_Reset(t *testing.T) {
	s := New()
	// Create a bucket first
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	s.Reset()

	// Bucket should be gone
	if s.GetBucket("test-bucket") != nil {
		t.Error("Reset() did not clear buckets")
	}
}

func TestStore_CreateBucket(t *testing.T) {
	tests := []struct {
		name    string
		req     *storage.BucketInsertRequest
		wantErr bool
	}{
		{
			name:    "create simple bucket",
			req:     &storage.BucketInsertRequest{Name: "test-bucket"},
			wantErr: false,
		},
		{
			name: "create bucket with location",
			req: &storage.BucketInsertRequest{
				Name:     "eu-bucket",
				Location: "EU",
			},
			wantErr: false,
		},
		{
			name: "create bucket with storage class",
			req: &storage.BucketInsertRequest{
				Name:         "nearline-bucket",
				StorageClass: "NEARLINE",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			bucket, err := s.CreateBucket(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if bucket.Name != tt.req.Name {
					t.Errorf("bucket name = %s, want %s", bucket.Name, tt.req.Name)
				}
				if bucket.Kind != "storage#bucket" {
					t.Errorf("bucket kind = %s, want storage#bucket", bucket.Kind)
				}
			}
		})
	}
}

func TestStore_CreateBucket_DuplicateError(t *testing.T) {
	s := New()
	req := &storage.BucketInsertRequest{Name: "test-bucket"}

	_, err := s.CreateBucket(req)
	if err != nil {
		t.Fatalf("first CreateBucket() failed: %v", err)
	}

	_, err = s.CreateBucket(req)
	if err == nil {
		t.Error("expected error for duplicate bucket, got nil")
	}
}

func TestStore_GetBucket(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	bucket := s.GetBucket("test-bucket")
	if bucket == nil {
		t.Error("GetBucket() returned nil for existing bucket")
	}

	bucket = s.GetBucket("non-existent")
	if bucket != nil {
		t.Error("GetBucket() returned non-nil for non-existent bucket")
	}
}

func TestStore_ListBuckets(t *testing.T) {
	s := New()

	// Empty list
	buckets := s.ListBuckets()
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets, got %d", len(buckets))
	}

	// Create buckets
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "bucket-b"})
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "bucket-a"})
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "bucket-c"})

	buckets = s.ListBuckets()
	if len(buckets) != 3 {
		t.Errorf("expected 3 buckets, got %d", len(buckets))
	}

	// Should be sorted by name
	if buckets[0].Name != "bucket-a" || buckets[1].Name != "bucket-b" || buckets[2].Name != "bucket-c" {
		t.Error("buckets are not sorted by name")
	}
}

func TestStore_UpdateBucket(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	updated, err := s.UpdateBucket("test-bucket", &storage.BucketUpdateRequest{
		StorageClass: "NEARLINE",
		Labels:       map[string]string{"env": "test"},
	})

	if err != nil {
		t.Fatalf("UpdateBucket() error: %v", err)
	}

	if updated.StorageClass != "NEARLINE" {
		t.Errorf("storage class = %s, want NEARLINE", updated.StorageClass)
	}

	if updated.Labels["env"] != "test" {
		t.Error("labels not updated correctly")
	}

	// Update non-existent bucket
	_, err = s.UpdateBucket("non-existent", &storage.BucketUpdateRequest{})
	if err == nil {
		t.Error("expected error for non-existent bucket")
	}
}

func TestStore_DeleteBucket(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	err := s.DeleteBucket("test-bucket")
	if err != nil {
		t.Fatalf("DeleteBucket() error: %v", err)
	}

	if s.GetBucket("test-bucket") != nil {
		t.Error("bucket still exists after delete")
	}

	// Delete non-existent bucket
	err = s.DeleteBucket("non-existent")
	if err == nil {
		t.Error("expected error for non-existent bucket")
	}
}

func TestStore_DeleteBucket_NotEmpty(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test-object", "text/plain", []byte("hello"), nil)

	err := s.DeleteBucket("test-bucket")
	if err == nil {
		t.Error("expected error when deleting non-empty bucket")
	}
}

func TestStore_CreateObject(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})

	content := []byte("Hello, World!")
	obj, err := s.CreateObject("test-bucket", "test-object.txt", "text/plain", content, nil)

	if err != nil {
		t.Fatalf("CreateObject() error: %v", err)
	}

	if obj.Name != "test-object.txt" {
		t.Errorf("object name = %s, want test-object.txt", obj.Name)
	}

	if obj.Bucket != "test-bucket" {
		t.Errorf("object bucket = %s, want test-bucket", obj.Bucket)
	}

	if obj.ContentType != "text/plain" {
		t.Errorf("content type = %s, want text/plain", obj.ContentType)
	}

	if obj.Size != uint64(len(content)) {
		t.Errorf("size = %d, want %d", obj.Size, len(content))
	}

	if obj.Kind != "storage#object" {
		t.Errorf("kind = %s, want storage#object", obj.Kind)
	}
}

func TestStore_CreateObject_BucketNotFound(t *testing.T) {
	s := New()

	_, err := s.CreateObject("non-existent", "object.txt", "text/plain", []byte("data"), nil)
	if err == nil {
		t.Error("expected error for non-existent bucket")
	}
}

func TestStore_GetObject(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test-object.txt", "text/plain", []byte("data"), nil)

	obj := s.GetObject("test-bucket", "test-object.txt")
	if obj == nil {
		t.Error("GetObject() returned nil for existing object")
	}

	obj = s.GetObject("test-bucket", "non-existent")
	if obj != nil {
		t.Error("GetObject() returned non-nil for non-existent object")
	}

	obj = s.GetObject("non-existent", "test-object.txt")
	if obj != nil {
		t.Error("GetObject() returned non-nil for non-existent bucket")
	}
}

func TestStore_GetObjectContent(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	content := []byte("Hello, World!")
	_, _ = s.CreateObject("test-bucket", "test-object.txt", "text/plain", content, nil)

	retrieved := s.GetObjectContent("test-bucket", "test-object.txt")
	if string(retrieved) != string(content) {
		t.Errorf("content = %s, want %s", string(retrieved), string(content))
	}

	retrieved = s.GetObjectContent("test-bucket", "non-existent")
	if retrieved != nil {
		t.Error("GetObjectContent() returned non-nil for non-existent object")
	}
}

func TestStore_ListObjects(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "file1.txt", "text/plain", []byte("1"), nil)
	_, _ = s.CreateObject("test-bucket", "file2.txt", "text/plain", []byte("2"), nil)
	_, _ = s.CreateObject("test-bucket", "folder/file3.txt", "text/plain", []byte("3"), nil)

	// List all objects
	objects, prefixes := s.ListObjects("test-bucket", "", "")
	if len(objects) != 3 {
		t.Errorf("expected 3 objects, got %d", len(objects))
	}
	if len(prefixes) != 0 {
		t.Errorf("expected 0 prefixes, got %d", len(prefixes))
	}

	// List with prefix
	objects, prefixes = s.ListObjects("test-bucket", "folder/", "")
	if len(objects) != 1 {
		t.Errorf("expected 1 object with prefix, got %d", len(objects))
	}

	// List with delimiter (hierarchical)
	objects, prefixes = s.ListObjects("test-bucket", "", "/")
	if len(objects) != 2 {
		t.Errorf("expected 2 objects at root level, got %d", len(objects))
	}
	if len(prefixes) != 1 || prefixes[0] != "folder/" {
		t.Errorf("expected prefix 'folder/', got %v", prefixes)
	}
}

func TestStore_UpdateObject(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test-object.txt", "text/plain", []byte("data"), nil)

	updated, err := s.UpdateObject("test-bucket", "test-object.txt", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("UpdateObject() error: %v", err)
	}

	if updated.Metadata["key"] != "value" {
		t.Error("metadata not updated correctly")
	}

	// Update non-existent object
	_, err = s.UpdateObject("test-bucket", "non-existent", nil)
	if err == nil {
		t.Error("expected error for non-existent object")
	}
}

func TestStore_DeleteObject(t *testing.T) {
	s := New()
	_, _ = s.CreateBucket(&storage.BucketInsertRequest{Name: "test-bucket"})
	_, _ = s.CreateObject("test-bucket", "test-object.txt", "text/plain", []byte("data"), nil)

	err := s.DeleteObject("test-bucket", "test-object.txt")
	if err != nil {
		t.Fatalf("DeleteObject() error: %v", err)
	}

	if s.GetObject("test-bucket", "test-object.txt") != nil {
		t.Error("object still exists after delete")
	}

	// Delete non-existent object
	err = s.DeleteObject("test-bucket", "non-existent")
	if err == nil {
		t.Error("expected error for non-existent object")
	}
}

func TestComputeMD5Hash(t *testing.T) {
	data := []byte("Hello, World!")
	hash := computeMD5Hash(data)

	// MD5 of "Hello, World!" base64 encoded
	expected := "ZajifYh5KDgxtmS9i38K1A=="
	if hash != expected {
		t.Errorf("MD5 hash = %s, want %s", hash, expected)
	}
}

func TestComputeCRC32C(t *testing.T) {
	data := []byte("Hello, World!")
	checksum := computeCRC32C(data)

	// CRC32C should be non-empty base64 string
	if checksum == "" {
		t.Error("CRC32C checksum is empty")
	}
}
