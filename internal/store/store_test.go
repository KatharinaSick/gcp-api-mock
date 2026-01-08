package store

import (
	"testing"

	"github.com/ksick/gcp-api-mock/internal/sqladmin"
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

// =============================================================================
// Cloud SQL Instance Tests
// =============================================================================

func TestStore_CreateSQLInstance(t *testing.T) {
	tests := []struct {
		name    string
		req     *sqladmin.InstanceInsertRequest
		wantErr bool
	}{
		{
			name:    "create simple instance",
			req:     &sqladmin.InstanceInsertRequest{Name: "test-instance"},
			wantErr: false,
		},
		{
			name: "create instance with all options",
			req: &sqladmin.InstanceInsertRequest{
				Name:            "test-instance-2",
				DatabaseVersion: "POSTGRES_15",
				Region:          "europe-west1",
				Settings: &sqladmin.Settings{
					Tier:             "db-custom-2-4096",
					AvailabilityType: "REGIONAL",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			instance, op, err := s.CreateSQLInstance(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSQLInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if instance.Name != tt.req.Name {
					t.Errorf("instance name = %s, want %s", instance.Name, tt.req.Name)
				}
				if instance.Kind != "sql#instance" {
					t.Errorf("instance kind = %s, want sql#instance", instance.Kind)
				}
				if op.Kind != "sql#operation" {
					t.Errorf("operation kind = %s, want sql#operation", op.Kind)
				}
				if op.OperationType != "CREATE" {
					t.Errorf("operation type = %s, want CREATE", op.OperationType)
				}
			}
		})
	}
}

func TestStore_CreateSQLInstance_DuplicateError(t *testing.T) {
	s := New()
	req := &sqladmin.InstanceInsertRequest{Name: "test-instance"}

	_, _, err := s.CreateSQLInstance(req)
	if err != nil {
		t.Fatalf("first CreateSQLInstance() failed: %v", err)
	}

	_, _, err = s.CreateSQLInstance(req)
	if err == nil {
		t.Error("expected error for duplicate instance, got nil")
	}
}

func TestStore_GetSQLInstance(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	instance := s.GetSQLInstance("test-instance")
	if instance == nil {
		t.Error("GetSQLInstance() returned nil for existing instance")
	}

	instance = s.GetSQLInstance("non-existent")
	if instance != nil {
		t.Error("GetSQLInstance() returned non-nil for non-existent instance")
	}
}

func TestStore_ListSQLInstances(t *testing.T) {
	s := New()

	// Empty list
	instances := s.ListSQLInstances()
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}

	// Create instances
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "instance-b"})
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "instance-a"})
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "instance-c"})

	instances = s.ListSQLInstances()
	if len(instances) != 3 {
		t.Errorf("expected 3 instances, got %d", len(instances))
	}

	// Should be sorted by name
	if instances[0].Name != "instance-a" || instances[1].Name != "instance-b" || instances[2].Name != "instance-c" {
		t.Error("instances are not sorted by name")
	}
}

func TestStore_UpdateSQLInstance(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	updated, op, err := s.UpdateSQLInstance("test-instance", &sqladmin.InstancePatchRequest{
		Settings: &sqladmin.Settings{
			Tier:       "db-n1-standard-2",
			UserLabels: map[string]string{"env": "test"},
		},
	})

	if err != nil {
		t.Fatalf("UpdateSQLInstance() error: %v", err)
	}

	if updated.Settings.Tier != "db-n1-standard-2" {
		t.Errorf("tier = %s, want db-n1-standard-2", updated.Settings.Tier)
	}

	if updated.Settings.UserLabels["env"] != "test" {
		t.Error("labels not updated correctly")
	}

	if op.OperationType != "UPDATE" {
		t.Errorf("operation type = %s, want UPDATE", op.OperationType)
	}

	// Update non-existent instance
	_, _, err = s.UpdateSQLInstance("non-existent", &sqladmin.InstancePatchRequest{})
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_DeleteSQLInstance(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	op, err := s.DeleteSQLInstance("test-instance")
	if err != nil {
		t.Fatalf("DeleteSQLInstance() error: %v", err)
	}

	if op.OperationType != "DELETE" {
		t.Errorf("operation type = %s, want DELETE", op.OperationType)
	}

	if s.GetSQLInstance("test-instance") != nil {
		t.Error("instance still exists after delete")
	}

	// Delete non-existent instance
	_, err = s.DeleteSQLInstance("non-existent")
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_DeleteSQLInstance_DeletionProtection(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{
		Name: "protected-instance",
		Settings: &sqladmin.Settings{
			DeletionProtectionEnabled: true,
		},
	})

	_, err := s.DeleteSQLInstance("protected-instance")
	if err == nil {
		t.Error("expected error when deleting protected instance")
	}
}

// =============================================================================
// Cloud SQL Database Tests
// =============================================================================

func TestStore_CreateSQLDatabase(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	db, op, err := s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{
		Name:      "mydb",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_general_ci",
	})

	if err != nil {
		t.Fatalf("CreateSQLDatabase() error: %v", err)
	}

	if db.Name != "mydb" {
		t.Errorf("database name = %s, want mydb", db.Name)
	}

	if db.Charset != "utf8mb4" {
		t.Errorf("charset = %s, want utf8mb4", db.Charset)
	}

	if op.OperationType != "CREATE_DATABASE" {
		t.Errorf("operation type = %s, want CREATE_DATABASE", op.OperationType)
	}
}

func TestStore_CreateSQLDatabase_InstanceNotFound(t *testing.T) {
	s := New()

	_, _, err := s.CreateSQLDatabase("non-existent", &sqladmin.DatabaseInsertRequest{Name: "mydb"})
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_CreateSQLDatabase_Duplicate(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	_, _, err := s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})
	if err == nil {
		t.Error("expected error for duplicate database")
	}
}

func TestStore_GetSQLDatabase(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	db := s.GetSQLDatabase("test-instance", "mydb")
	if db == nil {
		t.Error("GetSQLDatabase() returned nil for existing database")
	}

	db = s.GetSQLDatabase("test-instance", "non-existent")
	if db != nil {
		t.Error("GetSQLDatabase() returned non-nil for non-existent database")
	}

	db = s.GetSQLDatabase("non-existent", "mydb")
	if db != nil {
		t.Error("GetSQLDatabase() returned non-nil for non-existent instance")
	}
}

func TestStore_ListSQLDatabases(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "db1"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "db2"})

	databases, err := s.ListSQLDatabases("test-instance")
	if err != nil {
		t.Fatalf("ListSQLDatabases() error: %v", err)
	}

	// Should include default 'mysql' database plus our two
	if len(databases) != 3 {
		t.Errorf("expected 3 databases, got %d", len(databases))
	}

	// List from non-existent instance
	_, err = s.ListSQLDatabases("non-existent")
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_UpdateSQLDatabase(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	updated, op, err := s.UpdateSQLDatabase("test-instance", "mydb", &sqladmin.DatabasePatchRequest{
		Charset:   "utf8mb4",
		Collation: "utf8mb4_unicode_ci",
	})

	if err != nil {
		t.Fatalf("UpdateSQLDatabase() error: %v", err)
	}

	if updated.Charset != "utf8mb4" {
		t.Errorf("charset = %s, want utf8mb4", updated.Charset)
	}

	if updated.Collation != "utf8mb4_unicode_ci" {
		t.Errorf("collation = %s, want utf8mb4_unicode_ci", updated.Collation)
	}

	if op.OperationType != "UPDATE_DATABASE" {
		t.Errorf("operation type = %s, want UPDATE_DATABASE", op.OperationType)
	}
}

func TestStore_DeleteSQLDatabase(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLDatabase("test-instance", &sqladmin.DatabaseInsertRequest{Name: "mydb"})

	op, err := s.DeleteSQLDatabase("test-instance", "mydb")
	if err != nil {
		t.Fatalf("DeleteSQLDatabase() error: %v", err)
	}

	if op.OperationType != "DELETE_DATABASE" {
		t.Errorf("operation type = %s, want DELETE_DATABASE", op.OperationType)
	}

	if s.GetSQLDatabase("test-instance", "mydb") != nil {
		t.Error("database still exists after delete")
	}
}

// =============================================================================
// Cloud SQL User Tests
// =============================================================================

func TestStore_CreateSQLUser(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	user, op, err := s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{
		Name: "testuser",
		Host: "%",
	})

	if err != nil {
		t.Fatalf("CreateSQLUser() error: %v", err)
	}

	if user.Name != "testuser" {
		t.Errorf("user name = %s, want testuser", user.Name)
	}

	if user.Host != "%" {
		t.Errorf("host = %s, want %%", user.Host)
	}

	if op.OperationType != "CREATE_USER" {
		t.Errorf("operation type = %s, want CREATE_USER", op.OperationType)
	}
}

func TestStore_CreateSQLUser_InstanceNotFound(t *testing.T) {
	s := New()

	_, _, err := s.CreateSQLUser("non-existent", &sqladmin.UserInsertRequest{Name: "testuser"})
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_CreateSQLUser_Duplicate(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	_, _, err := s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})
	if err == nil {
		t.Error("expected error for duplicate user")
	}
}

func TestStore_GetSQLUser(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	user := s.GetSQLUser("test-instance", "testuser", "%")
	if user == nil {
		t.Error("GetSQLUser() returned nil for existing user")
	}

	user = s.GetSQLUser("test-instance", "testuser", "localhost")
	if user != nil {
		t.Error("GetSQLUser() returned non-nil for non-existent user with different host")
	}

	user = s.GetSQLUser("non-existent", "testuser", "%")
	if user != nil {
		t.Error("GetSQLUser() returned non-nil for non-existent instance")
	}
}

func TestStore_ListSQLUsers(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "user1", Host: "%"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "user2", Host: "%"})

	users, err := s.ListSQLUsers("test-instance")
	if err != nil {
		t.Fatalf("ListSQLUsers() error: %v", err)
	}

	// Should include default 'root' user plus our two
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// List from non-existent instance
	_, err = s.ListSQLUsers("non-existent")
	if err == nil {
		t.Error("expected error for non-existent instance")
	}
}

func TestStore_UpdateSQLUser(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	updated, op, err := s.UpdateSQLUser("test-instance", "testuser", "%", &sqladmin.UserUpdateRequest{
		Host: "localhost",
	})

	if err != nil {
		t.Fatalf("UpdateSQLUser() error: %v", err)
	}

	if updated.Host != "localhost" {
		t.Errorf("host = %s, want localhost", updated.Host)
	}

	if op.OperationType != "UPDATE_USER" {
		t.Errorf("operation type = %s, want UPDATE_USER", op.OperationType)
	}

	// Old key should not exist
	if s.GetSQLUser("test-instance", "testuser", "%") != nil {
		t.Error("old user key still exists after host change")
	}

	// New key should exist
	if s.GetSQLUser("test-instance", "testuser", "localhost") == nil {
		t.Error("new user key does not exist after host change")
	}
}

func TestStore_DeleteSQLUser(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})
	_, _, _ = s.CreateSQLUser("test-instance", &sqladmin.UserInsertRequest{Name: "testuser", Host: "%"})

	op, err := s.DeleteSQLUser("test-instance", "testuser", "%")
	if err != nil {
		t.Fatalf("DeleteSQLUser() error: %v", err)
	}

	if op.OperationType != "DELETE_USER" {
		t.Errorf("operation type = %s, want DELETE_USER", op.OperationType)
	}

	if s.GetSQLUser("test-instance", "testuser", "%") != nil {
		t.Error("user still exists after delete")
	}
}

// =============================================================================
// Cloud SQL Operation Tests
// =============================================================================

func TestStore_GetSQLOperation(t *testing.T) {
	s := New()
	_, op, _ := s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	retrieved := s.GetSQLOperation(op.Name)
	if retrieved == nil {
		t.Error("GetSQLOperation() returned nil for existing operation")
	}

	if retrieved.Name != op.Name {
		t.Errorf("operation name = %s, want %s", retrieved.Name, op.Name)
	}

	retrieved = s.GetSQLOperation("non-existent")
	if retrieved != nil {
		t.Error("GetSQLOperation() returned non-nil for non-existent operation")
	}
}

func TestStore_ListSQLOperations(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "instance-1"})
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "instance-2"})

	// List all operations
	operations := s.ListSQLOperations("")
	if len(operations) != 2 {
		t.Errorf("expected 2 operations, got %d", len(operations))
	}

	// List operations for specific instance
	operations = s.ListSQLOperations("instance-1")
	if len(operations) != 1 {
		t.Errorf("expected 1 operation for instance-1, got %d", len(operations))
	}

	if operations[0].TargetId != "instance-1" {
		t.Errorf("target id = %s, want instance-1", operations[0].TargetId)
	}
}

func TestStore_SQLInstance_CreatesDefaultDatabaseAndUser(t *testing.T) {
	s := New()
	_, _, _ = s.CreateSQLInstance(&sqladmin.InstanceInsertRequest{Name: "test-instance"})

	// Check default database
	db := s.GetSQLDatabase("test-instance", "mysql")
	if db == nil {
		t.Error("default mysql database was not created")
	}

	// Check default root user
	user := s.GetSQLUser("test-instance", "root", "%")
	if user == nil {
		t.Error("default root user was not created")
	}
}
