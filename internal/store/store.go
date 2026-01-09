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

	"github.com/katharinasick/gcp-api-mock/internal/sqladmin"
	"github.com/katharinasick/gcp-api-mock/internal/storage"
)

// Store is the main in-memory data store for all GCP resources.
// It is safe for concurrent access.
type Store struct {
	mu sync.RWMutex

	// Cloud Storage data
	buckets map[string]*storage.Bucket
	// objects is a map of bucket name to a map of object name to object
	objects map[string]map[string]*ObjectData

	// Cloud SQL data
	// sqlInstances is a map of instance name to database instance
	sqlInstances map[string]*sqladmin.DatabaseInstance
	// sqlDatabases is a map of instance name to a map of database name to database
	sqlDatabases map[string]map[string]*sqladmin.Database
	// sqlUsers is a map of instance name to a map of user key (name@host) to user
	sqlUsers map[string]map[string]*sqladmin.User
	// sqlOperations is a map of operation name to operation
	sqlOperations map[string]*sqladmin.Operation

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
		sqlInstances:  make(map[string]*sqladmin.DatabaseInstance),
		sqlDatabases:  make(map[string]map[string]*sqladmin.Database),
		sqlUsers:      make(map[string]map[string]*sqladmin.User),
		sqlOperations: make(map[string]*sqladmin.Operation),
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
	s.sqlInstances = make(map[string]*sqladmin.DatabaseInstance)
	s.sqlDatabases = make(map[string]map[string]*sqladmin.Database)
	s.sqlUsers = make(map[string]map[string]*sqladmin.User)
	s.sqlOperations = make(map[string]*sqladmin.Operation)
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
		Kind:             "storage#bucket",
		ID:               req.Name,
		SelfLink:         fmt.Sprintf("%s/storage/v1/b/%s", s.baseURL, req.Name),
		ProjectNumber:    s.projectNumber,
		Name:             req.Name,
		TimeCreated:      now,
		Updated:          now,
		Metageneration:   1,
		Location:         location,
		LocationType:     "region",
		StorageClass:     storageClass,
		Etag:             generateEtag(),
		Labels:           req.Labels,
		IamConfiguration: req.IamConfiguration,
		Versioning:       req.Versioning,
		Lifecycle:        req.Lifecycle,
		SoftDeletePolicy: req.SoftDeletePolicy,
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

	if req.IamConfiguration != nil {
		bucket.IamConfiguration = req.IamConfiguration
	}

	if req.Versioning != nil {
		bucket.Versioning = req.Versioning
	}

	if req.Lifecycle != nil {
		bucket.Lifecycle = req.Lifecycle
	}

	if req.SoftDeletePolicy != nil {
		bucket.SoftDeletePolicy = req.SoftDeletePolicy
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
// If an object with the same name and content already exists, returns the existing object.
func (s *Store) CreateObject(bucketName, objectName, contentType string, content []byte, metadata map[string]string) (*storage.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[bucketName]
	if !exists {
		return nil, fmt.Errorf("bucket %s not found", bucketName)
	}

	// Check if object already exists with the same content
	if existingObjData, exists := s.objects[bucketName][objectName]; exists {
		existingMD5 := existingObjData.Metadata.Md5Hash
		newMD5 := computeMD5Hash(content)

		// If content is the same, check if metadata is also the same
		if existingMD5 == newMD5 && metadataEqual(existingObjData.Metadata.Metadata, metadata) {
			// Content and metadata unchanged, return existing object
			return existingObjData.Metadata, nil
		}
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

// metadataEqual compares two metadata maps for equality.
func metadataEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// =============================================================================
// Cloud SQL Instance Operations
// =============================================================================

// CreateSQLInstance creates a new Cloud SQL instance in the store.
// Returns an error if an instance with the same name already exists.
func (s *Store) CreateSQLInstance(req *sqladmin.InstanceInsertRequest) (*sqladmin.DatabaseInstance, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sqlInstances[req.Name]; exists {
		return nil, nil, fmt.Errorf("instance %s already exists", req.Name)
	}

	now := time.Now().UTC()

	// Set defaults
	region := req.Region
	if region == "" {
		region = "us-central1"
	}

	databaseVersion := req.DatabaseVersion
	if databaseVersion == "" {
		databaseVersion = "MYSQL_8_0"
	}

	// Create default settings if not provided
	settings := req.Settings
	if settings == nil {
		settings = &sqladmin.Settings{}
	}
	settings.Kind = "sql#settings"
	settings.SettingsVersion = 1

	if settings.Tier == "" {
		settings.Tier = "db-n1-standard-1"
	}
	if settings.AvailabilityType == "" {
		settings.AvailabilityType = "ZONAL"
	}
	if settings.PricingPlan == "" {
		settings.PricingPlan = "PER_USE"
	}
	if settings.ActivationPolicy == "" {
		settings.ActivationPolicy = "ALWAYS"
	}
	if settings.DataDiskType == "" {
		settings.DataDiskType = "PD_SSD"
	}
	if settings.DataDiskSizeGb == 0 {
		settings.DataDiskSizeGb = 10
	}

	// Generate a mock IP address
	mockIP := fmt.Sprintf("10.%d.%d.%d", time.Now().UnixNano()%256, time.Now().UnixNano()%256, time.Now().UnixNano()%256)

	instance := &sqladmin.DatabaseInstance{
		Kind:            "sql#instance",
		Name:            req.Name,
		State:           "RUNNABLE",
		DatabaseVersion: databaseVersion,
		Region:          region,
		Project:         s.projectID,
		BackendType:     "SECOND_GEN",
		InstanceType:    "CLOUD_SQL_INSTANCE",
		SelfLink:        fmt.Sprintf("%s/sql/v1beta4/projects/%s/instances/%s", s.baseURL, s.projectID, req.Name),
		ConnectionName:  fmt.Sprintf("%s:%s:%s", s.projectID, region, req.Name),
		CreateTime:      now,
		Settings:        settings,
		Etag:            generateEtag(),
		GceZone:         fmt.Sprintf("%s-a", region),
		IPAddresses: []*sqladmin.IPMapping{
			{
				Type:      "PRIMARY",
				IPAddress: mockIP,
			},
		},
		ServiceAccountEmailAddress: fmt.Sprintf("p%d-abc123@gcp-sa-cloud-sql.iam.gserviceaccount.com", s.projectNumber),
	}

	if req.MasterInstanceName != "" {
		instance.MasterInstanceName = req.MasterInstanceName
		instance.InstanceType = "READ_REPLICA_INSTANCE"
	}

	s.sqlInstances[req.Name] = instance
	s.sqlDatabases[req.Name] = make(map[string]*sqladmin.Database)
	s.sqlUsers[req.Name] = make(map[string]*sqladmin.User)

	// Create a default database
	defaultDB := &sqladmin.Database{
		Kind:      "sql#database",
		Name:      "mysql",
		Charset:   "utf8",
		Collation: "utf8_general_ci",
		Instance:  req.Name,
		Project:   s.projectID,
		SelfLink:  fmt.Sprintf("%s/sql/v1beta4/projects/%s/instances/%s/databases/mysql", s.baseURL, s.projectID, req.Name),
		Etag:      generateEtag(),
	}
	s.sqlDatabases[req.Name]["mysql"] = defaultDB

	// Create a default root user
	rootUser := &sqladmin.User{
		Kind:     "sql#user",
		Name:     "root",
		Host:     "%",
		Instance: req.Name,
		Project:  s.projectID,
		Type:     "BUILT_IN",
		Etag:     generateEtag(),
	}
	s.sqlUsers[req.Name]["root@%"] = rootUser

	// Create operation
	op := s.createOperation("CREATE", req.Name, now)

	return instance, op, nil
}

// GetSQLInstance retrieves a Cloud SQL instance by name.
// Returns nil if the instance doesn't exist.
func (s *Store) GetSQLInstance(name string) *sqladmin.DatabaseInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sqlInstances[name]
}

// ListSQLInstances returns all Cloud SQL instances in the store.
func (s *Store) ListSQLInstances() []*sqladmin.DatabaseInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*sqladmin.DatabaseInstance, 0, len(s.sqlInstances))
	for _, instance := range s.sqlInstances {
		instances = append(instances, instance)
	}

	// Sort by name for consistent ordering
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})

	return instances
}

// UpdateSQLInstance updates an existing Cloud SQL instance.
// Returns an error if the instance doesn't exist.
func (s *Store) UpdateSQLInstance(name string, req *sqladmin.InstancePatchRequest) (*sqladmin.DatabaseInstance, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, exists := s.sqlInstances[name]
	if !exists {
		return nil, nil, fmt.Errorf("instance %s not found", name)
	}

	now := time.Now().UTC()

	if req.Settings != nil {
		if req.Settings.Tier != "" {
			instance.Settings.Tier = req.Settings.Tier
		}
		if req.Settings.AvailabilityType != "" {
			instance.Settings.AvailabilityType = req.Settings.AvailabilityType
		}
		if req.Settings.DataDiskSizeGb > 0 {
			instance.Settings.DataDiskSizeGb = req.Settings.DataDiskSizeGb
		}
		if req.Settings.UserLabels != nil {
			instance.Settings.UserLabels = req.Settings.UserLabels
		}
		if req.Settings.IPConfiguration != nil {
			instance.Settings.IPConfiguration = req.Settings.IPConfiguration
		}
		if req.Settings.BackupConfiguration != nil {
			instance.Settings.BackupConfiguration = req.Settings.BackupConfiguration
		}
		if req.Settings.MaintenanceWindow != nil {
			instance.Settings.MaintenanceWindow = req.Settings.MaintenanceWindow
		}
		if req.Settings.DatabaseFlags != nil {
			instance.Settings.DatabaseFlags = req.Settings.DatabaseFlags
		}
		instance.Settings.DeletionProtectionEnabled = req.Settings.DeletionProtectionEnabled
		instance.Settings.SettingsVersion++
	}

	instance.Etag = generateEtag()

	// Create operation
	op := s.createOperation("UPDATE", name, now)

	return instance, op, nil
}

// DeleteSQLInstance deletes a Cloud SQL instance by name.
// Returns an error if the instance doesn't exist.
func (s *Store) DeleteSQLInstance(name string) (*sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, exists := s.sqlInstances[name]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	// Check for deletion protection
	if instance.Settings != nil && instance.Settings.DeletionProtectionEnabled {
		return nil, fmt.Errorf("instance %s has deletion protection enabled", name)
	}

	now := time.Now().UTC()

	delete(s.sqlInstances, name)
	delete(s.sqlDatabases, name)
	delete(s.sqlUsers, name)

	// Create operation
	op := s.createOperation("DELETE", name, now)

	return op, nil
}

// =============================================================================
// Cloud SQL Database Operations
// =============================================================================

// CreateSQLDatabase creates a new database in a Cloud SQL instance.
// Returns an error if the instance doesn't exist or database already exists.
func (s *Store) CreateSQLDatabase(instanceName string, req *sqladmin.DatabaseInsertRequest) (*sqladmin.Database, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceDBs, exists := s.sqlDatabases[instanceName]
	if !exists {
		return nil, nil, fmt.Errorf("instance %s not found", instanceName)
	}

	if _, exists := instanceDBs[req.Name]; exists {
		return nil, nil, fmt.Errorf("database %s already exists in instance %s", req.Name, instanceName)
	}

	now := time.Now().UTC()

	// Set defaults
	charset := req.Charset
	if charset == "" {
		charset = "utf8"
	}

	collation := req.Collation
	if collation == "" {
		collation = "utf8_general_ci"
	}

	db := &sqladmin.Database{
		Kind:      "sql#database",
		Name:      req.Name,
		Charset:   charset,
		Collation: collation,
		Instance:  instanceName,
		Project:   s.projectID,
		SelfLink:  fmt.Sprintf("%s/sql/v1beta4/projects/%s/instances/%s/databases/%s", s.baseURL, s.projectID, instanceName, req.Name),
		Etag:      generateEtag(),
	}

	instanceDBs[req.Name] = db

	// Create operation
	op := s.createOperation("CREATE_DATABASE", instanceName, now)

	return db, op, nil
}

// GetSQLDatabase retrieves a database by instance and database name.
// Returns nil if the database doesn't exist.
func (s *Store) GetSQLDatabase(instanceName, dbName string) *sqladmin.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instanceDBs, exists := s.sqlDatabases[instanceName]
	if !exists {
		return nil
	}

	return instanceDBs[dbName]
}

// ListSQLDatabases returns all databases in a Cloud SQL instance.
func (s *Store) ListSQLDatabases(instanceName string) ([]*sqladmin.Database, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instanceDBs, exists := s.sqlDatabases[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", instanceName)
	}

	databases := make([]*sqladmin.Database, 0, len(instanceDBs))
	for _, db := range instanceDBs {
		databases = append(databases, db)
	}

	// Sort by name for consistent ordering
	sort.Slice(databases, func(i, j int) bool {
		return databases[i].Name < databases[j].Name
	})

	return databases, nil
}

// UpdateSQLDatabase updates an existing database in a Cloud SQL instance.
// Returns an error if the instance or database doesn't exist.
func (s *Store) UpdateSQLDatabase(instanceName, dbName string, req *sqladmin.DatabasePatchRequest) (*sqladmin.Database, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceDBs, exists := s.sqlDatabases[instanceName]
	if !exists {
		return nil, nil, fmt.Errorf("instance %s not found", instanceName)
	}

	db, exists := instanceDBs[dbName]
	if !exists {
		return nil, nil, fmt.Errorf("database %s not found in instance %s", dbName, instanceName)
	}

	now := time.Now().UTC()

	if req.Charset != "" {
		db.Charset = req.Charset
	}
	if req.Collation != "" {
		db.Collation = req.Collation
	}
	db.Etag = generateEtag()

	// Create operation
	op := s.createOperation("UPDATE_DATABASE", instanceName, now)

	return db, op, nil
}

// DeleteSQLDatabase deletes a database from a Cloud SQL instance.
// Returns an error if the instance or database doesn't exist.
func (s *Store) DeleteSQLDatabase(instanceName, dbName string) (*sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceDBs, exists := s.sqlDatabases[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", instanceName)
	}

	if _, exists := instanceDBs[dbName]; !exists {
		return nil, fmt.Errorf("database %s not found in instance %s", dbName, instanceName)
	}

	now := time.Now().UTC()

	delete(instanceDBs, dbName)

	// Create operation
	op := s.createOperation("DELETE_DATABASE", instanceName, now)

	return op, nil
}

// =============================================================================
// Cloud SQL User Operations
// =============================================================================

// userKey generates a unique key for a user based on name and host.
func userKey(name, host string) string {
	if host == "" {
		host = "%"
	}
	return fmt.Sprintf("%s@%s", name, host)
}

// CreateSQLUser creates a new user in a Cloud SQL instance.
// Returns an error if the instance doesn't exist or user already exists.
func (s *Store) CreateSQLUser(instanceName string, req *sqladmin.UserInsertRequest) (*sqladmin.User, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceUsers, exists := s.sqlUsers[instanceName]
	if !exists {
		return nil, nil, fmt.Errorf("instance %s not found", instanceName)
	}

	host := req.Host
	if host == "" {
		host = "%"
	}

	key := userKey(req.Name, host)
	if _, exists := instanceUsers[key]; exists {
		return nil, nil, fmt.Errorf("user %s already exists in instance %s", key, instanceName)
	}

	now := time.Now().UTC()

	userType := req.Type
	if userType == "" {
		userType = "BUILT_IN"
	}

	user := &sqladmin.User{
		Kind:     "sql#user",
		Name:     req.Name,
		Host:     host,
		Instance: instanceName,
		Project:  s.projectID,
		Type:     userType,
		Etag:     generateEtag(),
	}

	instanceUsers[key] = user

	// Create operation
	op := s.createOperation("CREATE_USER", instanceName, now)

	return user, op, nil
}

// GetSQLUser retrieves a user by instance, user name, and host.
// Returns nil if the user doesn't exist.
func (s *Store) GetSQLUser(instanceName, userName, host string) *sqladmin.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instanceUsers, exists := s.sqlUsers[instanceName]
	if !exists {
		return nil
	}

	return instanceUsers[userKey(userName, host)]
}

// ListSQLUsers returns all users in a Cloud SQL instance.
func (s *Store) ListSQLUsers(instanceName string) ([]*sqladmin.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instanceUsers, exists := s.sqlUsers[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", instanceName)
	}

	users := make([]*sqladmin.User, 0, len(instanceUsers))
	for _, user := range instanceUsers {
		users = append(users, user)
	}

	// Sort by name for consistent ordering
	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	return users, nil
}

// UpdateSQLUser updates an existing user in a Cloud SQL instance.
// Returns an error if the instance or user doesn't exist.
func (s *Store) UpdateSQLUser(instanceName, userName, host string, req *sqladmin.UserUpdateRequest) (*sqladmin.User, *sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceUsers, exists := s.sqlUsers[instanceName]
	if !exists {
		return nil, nil, fmt.Errorf("instance %s not found", instanceName)
	}

	key := userKey(userName, host)
	user, exists := instanceUsers[key]
	if !exists {
		return nil, nil, fmt.Errorf("user %s not found in instance %s", key, instanceName)
	}

	now := time.Now().UTC()

	// Note: password is not stored in the response
	if req.Host != "" && req.Host != host {
		// Host changed - need to re-key
		delete(instanceUsers, key)
		user.Host = req.Host
		instanceUsers[userKey(userName, req.Host)] = user
	}
	user.Etag = generateEtag()

	// Create operation
	op := s.createOperation("UPDATE_USER", instanceName, now)

	return user, op, nil
}

// DeleteSQLUser deletes a user from a Cloud SQL instance.
// Returns an error if the instance or user doesn't exist.
func (s *Store) DeleteSQLUser(instanceName, userName, host string) (*sqladmin.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceUsers, exists := s.sqlUsers[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", instanceName)
	}

	key := userKey(userName, host)
	if _, exists := instanceUsers[key]; !exists {
		return nil, fmt.Errorf("user %s not found in instance %s", key, instanceName)
	}

	now := time.Now().UTC()

	delete(instanceUsers, key)

	// Create operation
	op := s.createOperation("DELETE_USER", instanceName, now)

	return op, nil
}

// =============================================================================
// Cloud SQL Operation Operations
// =============================================================================

// createOperation creates and stores a new operation.
func (s *Store) createOperation(opType, targetID string, now time.Time) *sqladmin.Operation {
	opName := fmt.Sprintf("operation-%d", now.UnixNano())

	op := &sqladmin.Operation{
		Kind:          "sql#operation",
		Name:          opName,
		Status:        "DONE",
		OperationType: opType,
		InsertTime:    now,
		StartTime:     now,
		EndTime:       now,
		TargetProject: s.projectID,
		TargetId:      targetID,
		SelfLink:      fmt.Sprintf("%s/sql/v1beta4/projects/%s/operations/%s", s.baseURL, s.projectID, opName),
		TargetLink:    fmt.Sprintf("%s/sql/v1beta4/projects/%s/instances/%s", s.baseURL, s.projectID, targetID),
	}

	s.sqlOperations[opName] = op

	return op
}

// GetSQLOperation retrieves an operation by name.
// Returns nil if the operation doesn't exist.
func (s *Store) GetSQLOperation(name string) *sqladmin.Operation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sqlOperations[name]
}

// ListSQLOperations returns all operations in the store, optionally filtered by instance.
func (s *Store) ListSQLOperations(instanceName string) []*sqladmin.Operation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	operations := make([]*sqladmin.Operation, 0, len(s.sqlOperations))
	for _, op := range s.sqlOperations {
		if instanceName == "" || op.TargetId == instanceName {
			operations = append(operations, op)
		}
	}

	// Sort by insert time (newest first)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].InsertTime.After(operations[j].InsertTime)
	})

	return operations
}
