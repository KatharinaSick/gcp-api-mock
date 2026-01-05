package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject_JSONSerialization(t *testing.T) {
	// Test that Object serializes to GCP-compatible JSON format
	object := Object{
		Kind:        "storage#object",
		ID:          "my-bucket/my-object/1234567890",
		SelfLink:    "https://storage.googleapis.com/storage/v1/b/my-bucket/o/my-object",
		Name:        "my-object",
		Bucket:      "my-bucket",
		Generation:  "1234567890",
		ContentType: "application/octet-stream",
		Size:        "1024",
		MD5Hash:     "abc123",
		TimeCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Etag:        "CAE=",
	}

	jsonBytes, err := json.Marshal(object)
	require.NoError(t, err)

	// Parse the JSON to verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "storage#object", result["kind"])
	assert.Equal(t, "my-bucket/my-object/1234567890", result["id"])
	assert.Equal(t, "https://storage.googleapis.com/storage/v1/b/my-bucket/o/my-object", result["selfLink"])
	assert.Equal(t, "my-object", result["name"])
	assert.Equal(t, "my-bucket", result["bucket"])
	assert.Equal(t, "1234567890", result["generation"])
	assert.Equal(t, "application/octet-stream", result["contentType"])
	assert.Equal(t, "1024", result["size"])
	assert.Equal(t, "abc123", result["md5Hash"])
	assert.Equal(t, "CAE=", result["etag"])
	// Time should be in RFC3339 format
	assert.Equal(t, "2024-01-01T00:00:00Z", result["timeCreated"])
	assert.Equal(t, "2024-01-01T00:00:00Z", result["updated"])
}

func TestObject_JSONDeserialization(t *testing.T) {
	jsonStr := `{
		"kind": "storage#object",
		"id": "test-bucket/test-object/9876543210",
		"selfLink": "https://storage.googleapis.com/storage/v1/b/test-bucket/o/test-object",
		"name": "test-object",
		"bucket": "test-bucket",
		"generation": "9876543210",
		"contentType": "text/plain",
		"size": "2048",
		"md5Hash": "xyz789",
		"timeCreated": "2024-06-15T10:30:00Z",
		"updated": "2024-06-15T10:30:00Z",
		"etag": "XYZ="
	}`

	var object Object
	err := json.Unmarshal([]byte(jsonStr), &object)
	require.NoError(t, err)

	assert.Equal(t, "storage#object", object.Kind)
	assert.Equal(t, "test-bucket/test-object/9876543210", object.ID)
	assert.Equal(t, "https://storage.googleapis.com/storage/v1/b/test-bucket/o/test-object", object.SelfLink)
	assert.Equal(t, "test-object", object.Name)
	assert.Equal(t, "test-bucket", object.Bucket)
	assert.Equal(t, "9876543210", object.Generation)
	assert.Equal(t, "text/plain", object.ContentType)
	assert.Equal(t, "2048", object.Size)
	assert.Equal(t, "xyz789", object.MD5Hash)
	assert.Equal(t, "XYZ=", object.Etag)
	assert.Equal(t, 2024, object.TimeCreated.Year())
	assert.Equal(t, time.June, object.TimeCreated.Month())
	assert.Equal(t, 15, object.TimeCreated.Day())
}

func TestObjectListResponse_JSONSerialization(t *testing.T) {
	response := ObjectListResponse{
		Kind: "storage#objects",
		Items: []Object{
			{
				Kind:        "storage#object",
				ID:          "bucket-1/object-1/123",
				Name:        "object-1",
				Bucket:      "bucket-1",
				Generation:  "123",
				ContentType: "text/plain",
				Size:        "100",
				TimeCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Updated:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Kind:        "storage#object",
				ID:          "bucket-1/object-2/456",
				Name:        "object-2",
				Bucket:      "bucket-1",
				Generation:  "456",
				ContentType: "application/json",
				Size:        "200",
				TimeCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Updated:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	jsonBytes, err := json.Marshal(response)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "storage#objects", result["kind"])
	items, ok := result["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 2)
}

func TestNewObject(t *testing.T) {
	bucketName := "test-bucket"
	objectName := "test-object.txt"
	content := []byte("Hello, World!")
	contentType := "text/plain"

	object := NewObject(bucketName, objectName, content, contentType)

	assert.Equal(t, "storage#object", object.Kind)
	assert.Contains(t, object.ID, bucketName)
	assert.Contains(t, object.ID, objectName)
	assert.Equal(t, objectName, object.Name)
	assert.Equal(t, bucketName, object.Bucket)
	assert.Equal(t, contentType, object.ContentType)
	assert.Equal(t, "13", object.Size) // "Hello, World!" is 13 bytes
	assert.NotEmpty(t, object.Generation)
	assert.NotEmpty(t, object.MD5Hash)
	assert.NotEmpty(t, object.Etag)
	assert.Contains(t, object.SelfLink, bucketName)
	assert.Contains(t, object.SelfLink, objectName)
	assert.False(t, object.TimeCreated.IsZero())
	assert.False(t, object.Updated.IsZero())
}

func TestNewObject_DefaultContentType(t *testing.T) {
	object := NewObject("bucket", "object", []byte("data"), "")

	assert.Equal(t, "application/octet-stream", object.ContentType)
}

func TestObject_ContentStorage(t *testing.T) {
	content := []byte("Test content for storage")
	object := NewObject("bucket", "object", content, "text/plain")

	// Content should be stored and retrievable
	assert.Equal(t, content, object.Content)
}