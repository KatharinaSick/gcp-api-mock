package service

import (
	"testing"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

func newTestBucketService() *BucketService {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})
	return NewBucketService(factory)
}

func TestBucketService_Create(t *testing.T) {
	svc := newTestBucketService()

	tests := []struct {
		name          string
		bucketName    string
		projectNumber string
		location      string
		storageClass  string
		wantErr       bool
		errMessage    string
	}{
		{
			name:          "valid bucket creation",
			bucketName:    "my-test-bucket",
			projectNumber: "123456789",
			location:      "US",
			storageClass:  "STANDARD",
			wantErr:       false,
		},
		{
			name:          "invalid bucket name - too short",
			bucketName:    "ab",
			projectNumber: "123456789",
			location:      "US",
			storageClass:  "STANDARD",
			wantErr:       true,
			errMessage:    "bucket name must be between 3 and 63 characters",
		},
		{
			name:          "invalid bucket name - starts with goog",
			bucketName:    "google-bucket",
			projectNumber: "123456789",
			location:      "US",
			storageClass:  "STANDARD",
			wantErr:       true,
			errMessage:    "bucket name cannot start with 'goog' prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset service for each test
			svc = newTestBucketService()

			bucket, err := svc.Create(tt.bucketName, tt.projectNumber, tt.location, tt.storageClass)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("Create() error = %v, want %v", err.Error(), tt.errMessage)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if bucket.Name != tt.bucketName {
				t.Errorf("Create() bucket.Name = %v, want %v", bucket.Name, tt.bucketName)
			}
			if bucket.ProjectNumber != tt.projectNumber {
				t.Errorf("Create() bucket.ProjectNumber = %v, want %v", bucket.ProjectNumber, tt.projectNumber)
			}
			if bucket.Location != tt.location {
				t.Errorf("Create() bucket.Location = %v, want %v", bucket.Location, tt.location)
			}
			if bucket.StorageClass != tt.storageClass {
				t.Errorf("Create() bucket.StorageClass = %v, want %v", bucket.StorageClass, tt.storageClass)
			}
			if bucket.Kind != "storage#bucket" {
				t.Errorf("Create() bucket.Kind = %v, want storage#bucket", bucket.Kind)
			}
		})
	}
}

func TestBucketService_Create_Duplicate(t *testing.T) {
	svc := newTestBucketService()

	// Create first bucket
	_, err := svc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create() first bucket error = %v", err)
	}

	// Try to create duplicate
	_, err = svc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err == nil {
		t.Error("Create() expected error for duplicate bucket but got none")
	}
	if err != nil && err != ErrBucketAlreadyExists {
		t.Errorf("Create() error = %v, want %v", err, ErrBucketAlreadyExists)
	}
}

func TestBucketService_Get(t *testing.T) {
	svc := newTestBucketService()

	// Create a bucket first
	created, err := svc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get the bucket
	bucket, err := svc.Get("my-bucket")
	if err != nil {
		t.Errorf("Get() error = %v", err)
		return
	}

	if bucket.Name != created.Name {
		t.Errorf("Get() bucket.Name = %v, want %v", bucket.Name, created.Name)
	}
}

func TestBucketService_Get_NotFound(t *testing.T) {
	svc := newTestBucketService()

	_, err := svc.Get("non-existent-bucket")
	if err == nil {
		t.Error("Get() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("Get() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestBucketService_Delete(t *testing.T) {
	svc := newTestBucketService()

	// Create a bucket first
	_, err := svc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the bucket
	err = svc.Delete("my-bucket")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's deleted
	_, err = svc.Get("my-bucket")
	if err != ErrBucketNotFound {
		t.Error("Get() expected ErrBucketNotFound after deletion")
	}
}

func TestBucketService_Delete_NotFound(t *testing.T) {
	svc := newTestBucketService()

	err := svc.Delete("non-existent-bucket")
	if err == nil {
		t.Error("Delete() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("Delete() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestBucketService_List(t *testing.T) {
	svc := newTestBucketService()

	// List empty
	buckets, err := svc.List("123456789")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(buckets) != 0 {
		t.Errorf("List() expected 0 buckets, got %d", len(buckets))
	}

	// Create some buckets
	_, _ = svc.Create("bucket-1", "123456789", "US", "STANDARD")
	_, _ = svc.Create("bucket-2", "123456789", "EU", "STANDARD")
	_, _ = svc.Create("bucket-3", "999999999", "US", "STANDARD") // Different project

	// List buckets for project
	buckets, err = svc.List("123456789")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(buckets) != 2 {
		t.Errorf("List() expected 2 buckets for project 123456789, got %d", len(buckets))
	}

	// List buckets for other project
	buckets, err = svc.List("999999999")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(buckets) != 1 {
		t.Errorf("List() expected 1 bucket for project 999999999, got %d", len(buckets))
	}
}

func TestBucketService_Update(t *testing.T) {
	svc := newTestBucketService()

	// Create a bucket first
	_, err := svc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the bucket
	updates := models.Bucket{
		StorageClass: "NEARLINE",
	}
	bucket, err := svc.Update("my-bucket", updates)
	if err != nil {
		t.Errorf("Update() error = %v", err)
		return
	}

	if bucket.StorageClass != "NEARLINE" {
		t.Errorf("Update() bucket.StorageClass = %v, want NEARLINE", bucket.StorageClass)
	}
	// Verify other fields are preserved
	if bucket.Location != "US" {
		t.Errorf("Update() bucket.Location = %v, want US", bucket.Location)
	}
}

func TestBucketService_Update_NotFound(t *testing.T) {
	svc := newTestBucketService()

	updates := models.Bucket{
		StorageClass: "NEARLINE",
	}
	_, err := svc.Update("non-existent-bucket", updates)
	if err == nil {
		t.Error("Update() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("Update() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestBucketService_Delete_NonEmptyBucket(t *testing.T) {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})
	bucketSvc := NewBucketService(factory)
	objectSvc := NewObjectService(factory)
	objectSvc.SetBucketService(bucketSvc)
	bucketSvc.SetObjectService(objectSvc)

	// Create a bucket
	_, err := bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create bucket error = %v", err)
	}

	// Add an object to the bucket
	_, err = objectSvc.Create("my-bucket", "test-object.txt", []byte("content"), "text/plain")
	if err != nil {
		t.Fatalf("Create object error = %v", err)
	}

	// Try to delete the non-empty bucket
	err = bucketSvc.Delete("my-bucket")
	if err == nil {
		t.Error("Delete() expected error for non-empty bucket but got none")
	}
	if err != nil && err != ErrBucketNotEmpty {
		t.Errorf("Delete() error = %v, want %v", err, ErrBucketNotEmpty)
	}

	// Verify bucket still exists
	_, err = bucketSvc.Get("my-bucket")
	if err != nil {
		t.Error("Get() expected bucket to still exist after failed delete")
	}
}