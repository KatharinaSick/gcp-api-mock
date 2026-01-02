package service

import (
	"errors"
	"time"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
)

// Service-level errors
var (
	ErrBucketNotFound      = errors.New("bucket not found")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrBucketNotEmpty      = errors.New("bucket is not empty")
)

// BucketService provides business logic for bucket operations
type BucketService struct {
	factory       *store.StoreFactory
	objectService *ObjectService
}

// NewBucketService creates a new BucketService with the given store factory
func NewBucketService(factory *store.StoreFactory) *BucketService {
	return &BucketService{
		factory: factory,
	}
}

// bucketStore returns the store for buckets
func (s *BucketService) bucketStore() store.Store {
	return s.factory.GetStore("buckets")
}

// Create creates a new bucket with the given parameters
func (s *BucketService) Create(name, projectNumber, location, storageClass string) (*models.Bucket, error) {
	// Validate bucket name
	if err := models.ValidateBucketName(name); err != nil {
		return nil, err
	}

	// Check if bucket already exists
	if s.bucketStore().Exists(name) {
		return nil, ErrBucketAlreadyExists
	}

	// Create the bucket
	bucket := models.NewBucket(name, projectNumber, location, storageClass)

	// Store the bucket
	s.bucketStore().Set(name, bucket)

	return &bucket, nil
}

// Get retrieves a bucket by name
func (s *BucketService) Get(name string) (*models.Bucket, error) {
	data, ok := s.bucketStore().Get(name)
	if !ok {
		return nil, ErrBucketNotFound
	}

	bucket, ok := data.(models.Bucket)
	if !ok {
		return nil, ErrBucketNotFound
	}

	return &bucket, nil
}

// Delete removes a bucket by name
func (s *BucketService) Delete(name string) error {
	// Check if bucket exists
	if !s.bucketStore().Exists(name) {
		return ErrBucketNotFound
	}

	// Check if bucket has objects (only if objectService is set)
	if s.objectService != nil && s.objectService.HasObjects(name) {
		return ErrBucketNotEmpty
	}

	// Delete the bucket
	s.bucketStore().Delete(name)

	return nil
}

// List returns all buckets for a given project
func (s *BucketService) List(projectNumber string) ([]models.Bucket, error) {
	items := s.bucketStore().List()
	buckets := make([]models.Bucket, 0)

	for _, item := range items {
		bucket, ok := item.(models.Bucket)
		if !ok {
			continue
		}
		if bucket.ProjectNumber == projectNumber {
			buckets = append(buckets, bucket)
		}
	}

	return buckets, nil
}

// Update updates a bucket's metadata
func (s *BucketService) Update(name string, updates models.Bucket) (*models.Bucket, error) {
	// Get existing bucket
	bucket, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	// Apply updates (only non-empty fields)
	if updates.StorageClass != "" {
		bucket.StorageClass = updates.StorageClass
	}
	if updates.Location != "" {
		bucket.Location = updates.Location
	}

	// Update timestamp
	bucket.Updated = time.Now().UTC()
	bucket.Etag = models.GenerateEtag()

	// Save the updated bucket
	s.bucketStore().Set(name, *bucket)

	return bucket, nil
}

// SetObjectService sets the object service for checking if bucket is empty
// This is used for the Delete operation to check if bucket has objects
func (s *BucketService) SetObjectService(objSvc *ObjectService) {
	s.objectService = objSvc
}