package service

import (
	"errors"
	"time"

	"github.com/KatharinaSick/gcp-api-mock/internal/models"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
)

// Object-level errors
var (
	ErrObjectNotFound = errors.New("object not found")
)

// ObjectService provides business logic for object operations
type ObjectService struct {
	factory       *store.StoreFactory
	bucketService *BucketService
}

// NewObjectService creates a new ObjectService with the given store factory
func NewObjectService(factory *store.StoreFactory) *ObjectService {
	return &ObjectService{
		factory: factory,
	}
}

// SetBucketService sets the bucket service for validating bucket existence
func (s *ObjectService) SetBucketService(bucketSvc *BucketService) {
	s.bucketService = bucketSvc
}

// objectStoreKey returns the store key for objects in a specific bucket
func objectStoreKey(bucketName string) string {
	return "objects:" + bucketName
}

// objectStore returns the store for objects in a specific bucket
func (s *ObjectService) objectStore(bucketName string) store.Store {
	return s.factory.GetStore(objectStoreKey(bucketName))
}

// Create creates a new object in the specified bucket
func (s *ObjectService) Create(bucketName, objectName string, content []byte, contentType string) (*models.Object, error) {
	// Check if bucket exists
	if s.bucketService != nil {
		_, err := s.bucketService.Get(bucketName)
		if err != nil {
			return nil, ErrBucketNotFound
		}
	}

	// Create the object
	object := models.NewObject(bucketName, objectName, content, contentType)

	// Store the object
	s.objectStore(bucketName).Set(objectName, object)

	return &object, nil
}

// Get retrieves an object by bucket and name
func (s *ObjectService) Get(bucketName, objectName string) (*models.Object, error) {
	// Check if bucket exists
	if s.bucketService != nil {
		_, err := s.bucketService.Get(bucketName)
		if err != nil {
			return nil, ErrBucketNotFound
		}
	}

	data, ok := s.objectStore(bucketName).Get(objectName)
	if !ok {
		return nil, ErrObjectNotFound
	}

	object, ok := data.(models.Object)
	if !ok {
		return nil, ErrObjectNotFound
	}

	return &object, nil
}

// Delete removes an object by bucket and name
func (s *ObjectService) Delete(bucketName, objectName string) error {
	// Check if bucket exists
	if s.bucketService != nil {
		_, err := s.bucketService.Get(bucketName)
		if err != nil {
			return ErrBucketNotFound
		}
	}

	// Check if object exists
	if !s.objectStore(bucketName).Exists(objectName) {
		return ErrObjectNotFound
	}

	// Delete the object
	s.objectStore(bucketName).Delete(objectName)

	return nil
}

// List returns all objects in a bucket
func (s *ObjectService) List(bucketName string) ([]models.Object, error) {
	// Check if bucket exists
	if s.bucketService != nil {
		_, err := s.bucketService.Get(bucketName)
		if err != nil {
			return nil, ErrBucketNotFound
		}
	}

	items := s.objectStore(bucketName).List()
	objects := make([]models.Object, 0, len(items))

	for _, item := range items {
		object, ok := item.(models.Object)
		if !ok {
			continue
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// Update updates an object (re-uploads content)
func (s *ObjectService) Update(bucketName, objectName string, content []byte, contentType string) (*models.Object, error) {
	// Check if bucket exists
	if s.bucketService != nil {
		_, err := s.bucketService.Get(bucketName)
		if err != nil {
			return nil, ErrBucketNotFound
		}
	}

	// Check if object exists
	existing, err := s.Get(bucketName, objectName)
	if err != nil {
		return nil, err
	}

	// Create updated object (preserves some metadata, updates content)
	object := models.NewObject(bucketName, objectName, content, contentType)
	object.TimeCreated = existing.TimeCreated
	object.Updated = time.Now().UTC()

	// Store the updated object
	s.objectStore(bucketName).Set(objectName, object)

	return &object, nil
}

// HasObjects checks if a bucket has any objects
func (s *ObjectService) HasObjects(bucketName string) bool {
	items := s.objectStore(bucketName).List()
	return len(items) > 0
}

// DeleteAllInBucket deletes all objects in a bucket
func (s *ObjectService) DeleteAllInBucket(bucketName string) error {
	items := s.objectStore(bucketName).List()
	for _, item := range items {
		object, ok := item.(models.Object)
		if ok {
			s.objectStore(bucketName).Delete(object.Name)
		}
	}
	return nil
}