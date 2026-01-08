// Package store provides an in-memory data store for the GCP API Mock.
package store

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
	"time"

	"github.com/ksick/gcp-api-mock/internal/storage"
)

// Store is the main in-memory data store for all GCP resources.
// It is safe for concurrent access.
type Store struct {
	mu sync.RWMutex

	// Cloud Storage data
	buckets map[string]*storage.Bucket
	// objects is a map of bucket name to a map of object name to object
	objects map[string]map[string]*ObjectData

	// baseURL is the base URL for generating self links
	baseURL string
	// projectID is the default project ID for the mock
	projectID string
	// projectNumber is the default project number for the mock
	projectNumber uint64
}

// ObjectData stores the object metadata and its binary content.
type ObjectData struct {
	Metadata *storage.Object
	Content  []byte
}

// New creates a new empty Store.
func New() *Store {
	return &Store{
		buckets:       make(map[string]*storage.Bucket),
		objects:       make(map[string]map[string]*ObjectData),
		baseURL:       "http://localhost:8080",
		projectID:     "mock-project",
		projectNumber: 123456789012,
	}
}

// Reset clears all data from the store.
// Useful for testing and resetting state.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets = make(map[string]*storage.Bucket)
	s.objects = make(map[string]map[string]*ObjectData)
}

// SetBaseURL sets the base URL for generating self links.
func (s *Store) SetBaseURL(baseURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseURL = baseURL
}

// SetProject sets the project ID and number for the mock.
func (s *Store) SetProject(projectID string, projectNumber uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projectID = projectID
	s.projectNumber = projectNumber
}

// CreateBucket creates a new bucket in the store.
// Returns an error if a bucket with the same name already exists.
func (s *Store) CreateBucket(req *storage.BucketInsertRequest) (*storage.Bucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.buckets[req.Name]; exists {
		return nil, fmt.Errorf("bucket %s already exists", req.Name)
	}

	now := time.Now().UTC()

	// Set defaults if not provided
	location := req.Location
	if location == "" {
		location = "US"
	}

	storageClass := req.StorageClass
	if storageClass == "" {
		storageClass = "STANDARD"
	}

	bucket := &storage.Bucket{
		Kind:           "storage#bucket",
		ID:             req.Name,
		SelfLink:       fmt.Sprintf("%s/storage/v1/b/%s", s.baseURL, req.Name),
		ProjectNumber:  s.projectNumber,
		Name:           req.Name,
		TimeCreated:    now,
		Updated:        now,
		Metageneration: 1,
		Location:       location,
		LocationType:   "region",
		StorageClass:   storageClass,
		Etag:           generateEtag(),
		Labels:         req.Labels,
	}

	s.buckets[req.Name] = bucket
	s.objects[req.Name] = make(map[string]*ObjectData)

	return bucket, nil
}

// GetBucket retrieves a bucket by name.
// Returns nil if the bucket doesn't exist.
func (s *Store) GetBucket(name string) *storage.Bucket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.buckets[name]
}

// ListBuckets returns all buckets in the store.
func (s *Store) ListBuckets() []*storage.Bucket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets := make([]*storage.Bucket, 0, len(s.buckets))
	for _, bucket := range s.buckets {
		buckets = append(buckets, bucket)
	}

	// Sort by name for consistent ordering
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})

	return buckets
}

// UpdateBucket updates an existing bucket.
// Returns an error if the bucket doesn't exist.
func (s *Store) UpdateBucket(name string, req *storage.BucketUpdateRequest) (*storage.Bucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[name]
	if !exists {
		return nil, fmt.Errorf("bucket %s not found", name)
	}

	if req.StorageClass != "" {
		bucket.StorageClass = req.StorageClass
	}

	if req.Labels != nil {
		bucket.Labels = req.Labels
	}

	bucket.Updated = time.Now().UTC()
	bucket.Metageneration++
	bucket.Etag = generateEtag()

	return bucket, nil
}

// DeleteBucket deletes a bucket by name.
// Returns an error if the bucket doesn't exist or contains objects.
func (s *Store) DeleteBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.buckets[name]; !exists {
		return fmt.Errorf("bucket %s not found", name)
	}

	// Check if bucket has objects
	if len(s.objects[name]) > 0 {
		return fmt.Errorf("bucket %s is not empty", name)
	}

	delete(s.buckets, name)
	delete(s.objects, name)

	return nil
}

// CreateObject creates a new object in the specified bucket.
// Returns an error if the bucket doesn't exist.
func (s *Store) CreateObject(bucketName, objectName, contentType string, content []byte, metadata map[string]string) (*storage.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[bucketName]
	if !exists {
		return nil, fmt.Errorf("bucket %s not found", bucketName)
	}

	now := time.Now().UTC()
	generation := now.UnixNano()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	obj := &storage.Object{
		Kind:           "storage#object",
		ID:             fmt.Sprintf("%s/%s/%d", bucketName, objectName, generation),
		SelfLink:       fmt.Sprintf("%s/storage/v1/b/%s/o/%s", s.baseURL, bucketName, objectName),
		MediaLink:      fmt.Sprintf("%s/download/storage/v1/b/%s/o/%s?alt=media", s.baseURL, bucketName, objectName),
		Name:           objectName,
		Bucket:         bucketName,
		Generation:     generation,
		Metageneration: 1,
		ContentType:    contentType,
		TimeCreated:    now,
		Updated:        now,
		StorageClass:   bucket.StorageClass,
		Size:           uint64(len(content)),
		Md5Hash:        computeMD5Hash(content),
		Crc32c:         computeCRC32C(content),
		Etag:           generateEtag(),
		Metadata:       metadata,
	}

	s.objects[bucketName][objectName] = &ObjectData{
		Metadata: obj,
		Content:  content,
	}

	return obj, nil
}

// GetObject retrieves an object's metadata by bucket and object name.
// Returns nil if the object doesn't exist.
func (s *Store) GetObject(bucketName, objectName string) *storage.Object {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketObjects, exists := s.objects[bucketName]
	if !exists {
		return nil
	}

	objData, exists := bucketObjects[objectName]
	if !exists {
		return nil
	}

	return objData.Metadata
}

// GetObjectContent retrieves an object's content by bucket and object name.
// Returns nil if the object doesn't exist.
func (s *Store) GetObjectContent(bucketName, objectName string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketObjects, exists := s.objects[bucketName]
	if !exists {
		return nil
	}

	objData, exists := bucketObjects[objectName]
	if !exists {
		return nil
	}

	return objData.Content
}

// ListObjects returns all objects in a bucket, optionally filtered by prefix.
func (s *Store) ListObjects(bucketName, prefix, delimiter string) ([]*storage.Object, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketObjects, exists := s.objects[bucketName]
	if !exists {
		return nil, nil
	}

	var objects []*storage.Object
	prefixSet := make(map[string]struct{})

	for name, objData := range bucketObjects {
		// Check prefix filter
		if prefix != "" && !hasPrefix(name, prefix) {
			continue
		}

		// Handle delimiter (for hierarchical listing)
		if delimiter != "" {
			remainingPath := name
			if prefix != "" {
				remainingPath = name[len(prefix):]
			}

			delimIndex := indexOf(remainingPath, delimiter)
			if delimIndex >= 0 {
				// This is a "folder" - add to prefixes
				folderPrefix := prefix + remainingPath[:delimIndex+len(delimiter)]
				prefixSet[folderPrefix] = struct{}{}
				continue
			}
		}

		objects = append(objects, objData.Metadata)
	}

	// Sort objects by name for consistent ordering
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Name < objects[j].Name
	})

	// Convert prefixes set to sorted slice
	var prefixes []string
	for p := range prefixSet {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	return objects, prefixes
}

// UpdateObject updates an object's metadata.
// Returns an error if the object doesn't exist.
func (s *Store) UpdateObject(bucketName, objectName string, metadata map[string]string) (*storage.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucketObjects, exists := s.objects[bucketName]
	if !exists {
		return nil, fmt.Errorf("bucket %s not found", bucketName)
	}

	objData, exists := bucketObjects[objectName]
	if !exists {
		return nil, fmt.Errorf("object %s not found in bucket %s", objectName, bucketName)
	}

	objData.Metadata.Metadata = metadata
	objData.Metadata.Updated = time.Now().UTC()
	objData.Metadata.Metageneration++
	objData.Metadata.Etag = generateEtag()

	return objData.Metadata, nil
}

// DeleteObject deletes an object by bucket and object name.
// Returns an error if the object doesn't exist.
func (s *Store) DeleteObject(bucketName, objectName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucketObjects, exists := s.objects[bucketName]
	if !exists {
		return fmt.Errorf("bucket %s not found", bucketName)
	}

	if _, exists := bucketObjects[objectName]; !exists {
		return fmt.Errorf("object %s not found in bucket %s", objectName, bucketName)
	}

	delete(bucketObjects, objectName)

	return nil
}

// generateEtag generates a simple etag for a resource.
func generateEtag() string {
	return fmt.Sprintf("CAE%d=", time.Now().UnixNano())
}

// computeMD5Hash computes the base64-encoded MD5 hash of data.
func computeMD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// computeCRC32C computes the base64-encoded CRC32C checksum of data.
func computeCRC32C(data []byte) string {
	// Use Castagnoli polynomial for CRC32C
	table := crc32.MakeTable(crc32.Castagnoli)
	checksum := crc32.Checksum(data, table)
	// Convert to 4 bytes big-endian and base64 encode
	bytes := []byte{
		byte(checksum >> 24),
		byte(checksum >> 16),
		byte(checksum >> 8),
		byte(checksum),
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// hasPrefix checks if a string has the given prefix.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// indexOf returns the index of the first occurrence of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
