// Package storage provides data models for the Google Cloud Storage API mock.
package storage

import "time"

// Bucket represents a Cloud Storage bucket.
// Based on the official GCS JSON API v1 specification.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets
type Bucket struct {
	// Kind is the kind of item this is. For buckets, this is always "storage#bucket".
	Kind string `json:"kind"`
	// ID is the ID of the bucket.
	ID string `json:"id"`
	// SelfLink is the URI of this bucket.
	SelfLink string `json:"selfLink"`
	// ProjectNumber is the project number of the project the bucket belongs to.
	ProjectNumber uint64 `json:"projectNumber,string"`
	// Name is the name of the bucket.
	Name string `json:"name"`
	// TimeCreated is the creation time of the bucket in RFC 3339 format.
	TimeCreated time.Time `json:"timeCreated"`
	// Updated is the modification time of the bucket in RFC 3339 format.
	Updated time.Time `json:"updated"`
	// Metageneration is the metadata generation of this bucket.
	Metageneration int64 `json:"metageneration,string"`
	// Location is the location of the bucket.
	Location string `json:"location"`
	// LocationType describes the type of location (e.g., "region", "dual-region", "multi-region").
	LocationType string `json:"locationType"`
	// StorageClass is the default storage class of the bucket.
	StorageClass string `json:"storageClass"`
	// Etag is the HTTP 1.1 Entity tag for the bucket.
	Etag string `json:"etag"`
	// Labels are user-provided labels, in key/value pairs.
	Labels map[string]string `json:"labels,omitempty"`
}

// BucketList represents a list of buckets.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/buckets/list
type BucketList struct {
	// Kind is the kind of item this is. For bucket lists, this is always "storage#buckets".
	Kind string `json:"kind"`
	// Items is the list of buckets.
	Items []*Bucket `json:"items"`
	// NextPageToken is the continuation token for paginated results.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// Object represents a Cloud Storage object.
// Based on the official GCS JSON API v1 specification.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects
type Object struct {
	// Kind is the kind of item this is. For objects, this is always "storage#object".
	Kind string `json:"kind"`
	// ID is the ID of the object, including the bucket name, object name, and generation number.
	ID string `json:"id"`
	// SelfLink is the link to this object.
	SelfLink string `json:"selfLink"`
	// MediaLink is the link to access the object's data.
	MediaLink string `json:"mediaLink"`
	// Name is the name of the object.
	Name string `json:"name"`
	// Bucket is the name of the bucket containing this object.
	Bucket string `json:"bucket"`
	// Generation is the content generation of this object.
	Generation int64 `json:"generation,string"`
	// Metageneration is the version of the metadata for this object.
	Metageneration int64 `json:"metageneration,string"`
	// ContentType is the Content-Type of the object data.
	ContentType string `json:"contentType"`
	// TimeCreated is the creation time of the object in RFC 3339 format.
	TimeCreated time.Time `json:"timeCreated"`
	// Updated is the modification time of the object's metadata in RFC 3339 format.
	Updated time.Time `json:"updated"`
	// StorageClass is the storage class of the object.
	StorageClass string `json:"storageClass"`
	// Size is the Content-Length of the data in bytes.
	Size uint64 `json:"size,string"`
	// Md5Hash is the MD5 hash of the data; encoded using base64.
	Md5Hash string `json:"md5Hash"`
	// Crc32c is the CRC32c checksum; encoded using base64.
	Crc32c string `json:"crc32c"`
	// Etag is the HTTP 1.1 Entity tag for the object.
	Etag string `json:"etag"`
	// Metadata are user-provided metadata, in key/value pairs.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ObjectList represents a list of objects.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/objects/list
type ObjectList struct {
	// Kind is the kind of item this is. For object lists, this is always "storage#objects".
	Kind string `json:"kind"`
	// Items is the list of objects.
	Items []*Object `json:"items"`
	// Prefixes are object name prefixes for objects that matched the listing request.
	Prefixes []string `json:"prefixes,omitempty"`
	// NextPageToken is the continuation token for paginated results.
	NextPageToken string `json:"nextPageToken,omitempty"`
}

// BucketInsertRequest represents the request body for creating a bucket.
type BucketInsertRequest struct {
	Name         string            `json:"name"`
	Location     string            `json:"location,omitempty"`
	StorageClass string            `json:"storageClass,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// BucketUpdateRequest represents the request body for updating a bucket.
type BucketUpdateRequest struct {
	StorageClass string            `json:"storageClass,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// APIError represents an error response from the GCS API.
// Reference: https://cloud.google.com/storage/docs/json_api/v1/status-codes
type APIError struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails contains the details of an API error.
type ErrorDetails struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Errors  []ErrorReason `json:"errors,omitempty"`
}

// ErrorReason contains the reason for an error.
type ErrorReason struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
