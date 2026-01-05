package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucket_JSONSerialization(t *testing.T) {
	// Test that Bucket serializes to GCP-compatible JSON format
	bucket := Bucket{
		Kind:          "storage#bucket",
		ID:            "my-bucket",
		SelfLink:      "https://storage.googleapis.com/storage/v1/b/my-bucket",
		Name:          "my-bucket",
		ProjectNumber: "123456789",
		TimeCreated:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Location:      "US",
		StorageClass:  "STANDARD",
		Etag:          "CAE=",
	}

	jsonBytes, err := json.Marshal(bucket)
	require.NoError(t, err)

	// Parse the JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "storage#bucket", result["kind"])
	assert.Equal(t, "my-bucket", result["id"])
	assert.Equal(t, "https://storage.googleapis.com/storage/v1/b/my-bucket", result["selfLink"])
	assert.Equal(t, "my-bucket", result["name"])
	assert.Equal(t, "123456789", result["projectNumber"])
	assert.Equal(t, "US", result["location"])
	assert.Equal(t, "STANDARD", result["storageClass"])
	assert.Equal(t, "CAE=", result["etag"])
	// Time should be in RFC3339 format
	assert.Equal(t, "2024-01-01T00:00:00Z", result["timeCreated"])
	assert.Equal(t, "2024-01-01T00:00:00Z", result["updated"])
}

func TestBucket_JSONDeserialization(t *testing.T) {
	jsonStr := `{
		"kind": "storage#bucket",
		"id": "test-bucket",
		"selfLink": "https://storage.googleapis.com/storage/v1/b/test-bucket",
		"name": "test-bucket",
		"projectNumber": "987654321",
		"timeCreated": "2024-06-15T10:30:00Z",
		"updated": "2024-06-15T10:30:00Z",
		"location": "EU",
		"storageClass": "NEARLINE",
		"etag": "ABC="
	}`

	var bucket Bucket
	err := json.Unmarshal([]byte(jsonStr), &bucket)
	require.NoError(t, err)

	assert.Equal(t, "storage#bucket", bucket.Kind)
	assert.Equal(t, "test-bucket", bucket.ID)
	assert.Equal(t, "https://storage.googleapis.com/storage/v1/b/test-bucket", bucket.SelfLink)
	assert.Equal(t, "test-bucket", bucket.Name)
	assert.Equal(t, "987654321", bucket.ProjectNumber)
	assert.Equal(t, "EU", bucket.Location)
	assert.Equal(t, "NEARLINE", bucket.StorageClass)
	assert.Equal(t, "ABC=", bucket.Etag)
	assert.Equal(t, 2024, bucket.TimeCreated.Year())
	assert.Equal(t, time.June, bucket.TimeCreated.Month())
	assert.Equal(t, 15, bucket.TimeCreated.Day())
}

func TestValidateBucketName_ValidNames(t *testing.T) {
	validNames := []string{
		"my-bucket",
		"my_bucket",
		"mybucket",
		"my.bucket.name",
		"bucket123",
		"123bucket",
		"a1b",                                                             // minimum length 3
		"abcdefghijklmnopqrstuvwxyz0123456789-_.abcdefghijklmnopqrstuvwx", // 63 chars (max)
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := ValidateBucketName(name)
			assert.NoError(t, err, "Expected bucket name '%s' to be valid", name)
		})
	}
}

func TestValidateBucketName_InvalidNames(t *testing.T) {
	testCases := []struct {
		name        string
		bucketName  string
		expectedErr string
	}{
		{
			name:        "too short",
			bucketName:  "ab",
			expectedErr: "bucket name must be between 3 and 63 characters",
		},
		{
			name:        "too long",
			bucketName:  "abcdefghijklmnopqrstuvwxyz0123456789-_.abcdefghijklmnopqrstuvwxyz",
			expectedErr: "bucket name must be between 3 and 63 characters",
		},
		{
			name:        "uppercase letters",
			bucketName:  "MyBucket",
			expectedErr: "bucket name can only contain lowercase letters, numbers, hyphens, underscores, and periods",
		},
		{
			name:        "starts with hyphen",
			bucketName:  "-mybucket",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "ends with hyphen",
			bucketName:  "mybucket-",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "starts with period",
			bucketName:  ".mybucket",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "ends with period",
			bucketName:  "mybucket.",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "starts with underscore",
			bucketName:  "_mybucket",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "ends with underscore",
			bucketName:  "mybucket_",
			expectedErr: "bucket name must start and end with a letter or number",
		},
		{
			name:        "IP address format",
			bucketName:  "192.168.1.1",
			expectedErr: "bucket name cannot be an IP address",
		},
		{
			name:        "starts with goog prefix",
			bucketName:  "googlebucket",
			expectedErr: "bucket name cannot start with 'goog' prefix",
		},
		{
			name:        "starts with goog prefix variation",
			bucketName:  "goog-bucket",
			expectedErr: "bucket name cannot start with 'goog' prefix",
		},
		{
			name:        "invalid characters",
			bucketName:  "my@bucket",
			expectedErr: "bucket name can only contain lowercase letters, numbers, hyphens, underscores, and periods",
		},
		{
			name:        "space in name",
			bucketName:  "my bucket",
			expectedErr: "bucket name can only contain lowercase letters, numbers, hyphens, underscores, and periods",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBucketName(tc.bucketName)
			require.Error(t, err, "Expected bucket name '%s' to be invalid", tc.bucketName)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestBucketListResponse_JSONSerialization(t *testing.T) {
	response := BucketListResponse{
		Kind: "storage#buckets",
		Items: []Bucket{
			{
				Kind:          "storage#bucket",
				ID:            "bucket-1",
				Name:          "bucket-1",
				Location:      "US",
				StorageClass:  "STANDARD",
				ProjectNumber: "123",
				TimeCreated:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Updated:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Kind:          "storage#bucket",
				ID:            "bucket-2",
				Name:          "bucket-2",
				Location:      "EU",
				StorageClass:  "NEARLINE",
				ProjectNumber: "123",
				TimeCreated:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Updated:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	jsonBytes, err := json.Marshal(response)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "storage#buckets", result["kind"])
	items, ok := result["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 2)
}

func TestNewBucket(t *testing.T) {
	projectNumber := "123456789"
	name := "test-bucket"
	location := "US"
	storageClass := "STANDARD"

	bucket := NewBucket(name, projectNumber, location, storageClass)

	assert.Equal(t, "storage#bucket", bucket.Kind)
	assert.Equal(t, name, bucket.ID)
	assert.Equal(t, name, bucket.Name)
	assert.Equal(t, projectNumber, bucket.ProjectNumber)
	assert.Equal(t, location, bucket.Location)
	assert.Equal(t, storageClass, bucket.StorageClass)
	assert.Contains(t, bucket.SelfLink, name)
	assert.NotEmpty(t, bucket.Etag)
	assert.False(t, bucket.TimeCreated.IsZero())
	assert.False(t, bucket.Updated.IsZero())
}