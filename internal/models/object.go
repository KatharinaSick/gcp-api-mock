package models

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Object represents a GCP Cloud Storage object resource
type Object struct {
	Kind        string    `json:"kind"`
	ID          string    `json:"id"`
	SelfLink    string    `json:"selfLink"`
	Name        string    `json:"name"`
	Bucket      string    `json:"bucket"`
	Generation  string    `json:"generation"`
	ContentType string    `json:"contentType"`
	Size        string    `json:"size"`
	MD5Hash     string    `json:"md5Hash"`
	TimeCreated time.Time `json:"timeCreated"`
	Updated     time.Time `json:"updated"`
	Etag        string    `json:"etag"`
	// Content holds the actual object data (not serialized to JSON in metadata responses)
	Content []byte `json:"-"`
}

// ObjectListResponse represents the GCP API response for listing objects
type ObjectListResponse struct {
	Kind  string   `json:"kind"`
	Items []Object `json:"items"`
}

// NewObject creates a new Object with the given parameters and sets default values
func NewObject(bucketName, objectName string, content []byte, contentType string) Object {
	now := time.Now().UTC()
	generation := strconv.FormatInt(now.UnixNano(), 10)

	// Default content type if not specified
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Calculate MD5 hash of content
	md5Sum := md5.Sum(content)
	md5Hash := base64.StdEncoding.EncodeToString(md5Sum[:])

	return Object{
		Kind:        "storage#object",
		ID:          fmt.Sprintf("%s/%s/%s", bucketName, objectName, generation),
		SelfLink:    fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", bucketName, url.PathEscape(objectName)),
		Name:        objectName,
		Bucket:      bucketName,
		Generation:  generation,
		ContentType: contentType,
		Size:        strconv.Itoa(len(content)),
		MD5Hash:     md5Hash,
		TimeCreated: now,
		Updated:     now,
		Etag:        generateEtag(),
		Content:     content,
	}
}