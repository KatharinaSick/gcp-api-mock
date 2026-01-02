package service

import (
	"testing"

	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

func newTestObjectService() (*ObjectService, *BucketService) {
	factory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})
	bucketSvc := NewBucketService(factory)
	objectSvc := NewObjectService(factory)
	objectSvc.SetBucketService(bucketSvc)
	return objectSvc, bucketSvc
}

func TestObjectService_Create(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket first
	_, err := bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	if err != nil {
		t.Fatalf("Create bucket error = %v", err)
	}

	tests := []struct {
		name        string
		bucketName  string
		objectName  string
		content     []byte
		contentType string
		wantErr     bool
		errType     error
	}{
		{
			name:        "valid object creation",
			bucketName:  "my-bucket",
			objectName:  "test-object.txt",
			content:     []byte("Hello, World!"),
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "object with empty content type defaults",
			bucketName:  "my-bucket",
			objectName:  "binary-file",
			content:     []byte{0x00, 0x01, 0x02},
			contentType: "",
			wantErr:     false,
		},
		{
			name:        "object in non-existent bucket",
			bucketName:  "non-existent-bucket",
			objectName:  "test.txt",
			content:     []byte("test"),
			contentType: "text/plain",
			wantErr:     true,
			errType:     ErrBucketNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := objSvc.Create(tt.bucketName, tt.objectName, tt.content, tt.contentType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("Create() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if obj.Name != tt.objectName {
				t.Errorf("Create() object.Name = %v, want %v", obj.Name, tt.objectName)
			}
			if obj.Bucket != tt.bucketName {
				t.Errorf("Create() object.Bucket = %v, want %v", obj.Bucket, tt.bucketName)
			}
			if string(obj.Content) != string(tt.content) {
				t.Errorf("Create() object.Content = %v, want %v", obj.Content, tt.content)
			}
			if obj.Kind != "storage#object" {
				t.Errorf("Create() object.Kind = %v, want storage#object", obj.Kind)
			}
			if tt.contentType == "" && obj.ContentType != "application/octet-stream" {
				t.Errorf("Create() object.ContentType = %v, want application/octet-stream", obj.ContentType)
			}
		})
	}
}

func TestObjectService_Get(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket and object first
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	created, err := objSvc.Create("my-bucket", "test-object.txt", []byte("Hello"), "text/plain")
	if err != nil {
		t.Fatalf("Create object error = %v", err)
	}

	// Get the object
	obj, err := objSvc.Get("my-bucket", "test-object.txt")
	if err != nil {
		t.Errorf("Get() error = %v", err)
		return
	}

	if obj.Name != created.Name {
		t.Errorf("Get() object.Name = %v, want %v", obj.Name, created.Name)
	}
	if string(obj.Content) != string(created.Content) {
		t.Errorf("Get() object.Content = %v, want %v", obj.Content, created.Content)
	}
}

func TestObjectService_Get_NotFound(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket but no object
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")

	_, err := objSvc.Get("my-bucket", "non-existent-object")
	if err == nil {
		t.Error("Get() expected error for non-existent object but got none")
	}
	if err != nil && err != ErrObjectNotFound {
		t.Errorf("Get() error = %v, want %v", err, ErrObjectNotFound)
	}
}

func TestObjectService_Get_BucketNotFound(t *testing.T) {
	objSvc, _ := newTestObjectService()

	_, err := objSvc.Get("non-existent-bucket", "test-object")
	if err == nil {
		t.Error("Get() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("Get() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestObjectService_Delete(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket and object first
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	_, err := objSvc.Create("my-bucket", "test-object.txt", []byte("Hello"), "text/plain")
	if err != nil {
		t.Fatalf("Create object error = %v", err)
	}

	// Delete the object
	err = objSvc.Delete("my-bucket", "test-object.txt")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify it's deleted
	_, err = objSvc.Get("my-bucket", "test-object.txt")
	if err != ErrObjectNotFound {
		t.Error("Get() expected ErrObjectNotFound after deletion")
	}
}

func TestObjectService_Delete_NotFound(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create bucket but no object
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")

	err := objSvc.Delete("my-bucket", "non-existent-object")
	if err == nil {
		t.Error("Delete() expected error for non-existent object but got none")
	}
	if err != nil && err != ErrObjectNotFound {
		t.Errorf("Delete() error = %v, want %v", err, ErrObjectNotFound)
	}
}

func TestObjectService_Delete_BucketNotFound(t *testing.T) {
	objSvc, _ := newTestObjectService()

	err := objSvc.Delete("non-existent-bucket", "test-object")
	if err == nil {
		t.Error("Delete() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("Delete() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestObjectService_List(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")

	// List empty
	objects, err := objSvc.List("my-bucket")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("List() expected 0 objects, got %d", len(objects))
	}

	// Create some objects
	_, _ = objSvc.Create("my-bucket", "object-1.txt", []byte("content1"), "text/plain")
	_, _ = objSvc.Create("my-bucket", "object-2.txt", []byte("content2"), "text/plain")
	_, _ = objSvc.Create("my-bucket", "folder/object-3.txt", []byte("content3"), "text/plain")

	// List objects
	objects, err = objSvc.List("my-bucket")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(objects) != 3 {
		t.Errorf("List() expected 3 objects, got %d", len(objects))
	}
}

func TestObjectService_List_BucketNotFound(t *testing.T) {
	objSvc, _ := newTestObjectService()

	_, err := objSvc.List("non-existent-bucket")
	if err == nil {
		t.Error("List() expected error for non-existent bucket but got none")
	}
	if err != nil && err != ErrBucketNotFound {
		t.Errorf("List() error = %v, want %v", err, ErrBucketNotFound)
	}
}

func TestObjectService_Update(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket and object first
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	_, err := objSvc.Create("my-bucket", "test-object.txt", []byte("Original"), "text/plain")
	if err != nil {
		t.Fatalf("Create object error = %v", err)
	}

	// Update the object
	newContent := []byte("Updated content")
	obj, err := objSvc.Update("my-bucket", "test-object.txt", newContent, "text/html")
	if err != nil {
		t.Errorf("Update() error = %v", err)
		return
	}

	if string(obj.Content) != string(newContent) {
		t.Errorf("Update() object.Content = %v, want %v", string(obj.Content), string(newContent))
	}
	if obj.ContentType != "text/html" {
		t.Errorf("Update() object.ContentType = %v, want text/html", obj.ContentType)
	}
}

func TestObjectService_Update_NotFound(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create bucket but no object
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")

	_, err := objSvc.Update("my-bucket", "non-existent-object", []byte("content"), "text/plain")
	if err == nil {
		t.Error("Update() expected error for non-existent object but got none")
	}
	if err != nil && err != ErrObjectNotFound {
		t.Errorf("Update() error = %v, want %v", err, ErrObjectNotFound)
	}
}

func TestObjectService_HasObjects(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")

	// Initially no objects
	if objSvc.HasObjects("my-bucket") {
		t.Error("HasObjects() expected false for empty bucket")
	}

	// Add an object
	_, _ = objSvc.Create("my-bucket", "test.txt", []byte("content"), "text/plain")

	// Now should have objects
	if !objSvc.HasObjects("my-bucket") {
		t.Error("HasObjects() expected true after adding object")
	}
}

func TestObjectService_DeleteAllInBucket(t *testing.T) {
	objSvc, bucketSvc := newTestObjectService()

	// Create a bucket with objects
	_, _ = bucketSvc.Create("my-bucket", "123456789", "US", "STANDARD")
	_, _ = objSvc.Create("my-bucket", "object-1.txt", []byte("content1"), "text/plain")
	_, _ = objSvc.Create("my-bucket", "object-2.txt", []byte("content2"), "text/plain")

	// Verify objects exist
	if !objSvc.HasObjects("my-bucket") {
		t.Error("HasObjects() expected true before deletion")
	}

	// Delete all objects
	err := objSvc.DeleteAllInBucket("my-bucket")
	if err != nil {
		t.Errorf("DeleteAllInBucket() error = %v", err)
	}

	// Verify all objects are deleted
	if objSvc.HasObjects("my-bucket") {
		t.Error("HasObjects() expected false after DeleteAllInBucket")
	}
}