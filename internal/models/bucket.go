package models

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// Bucket represents a GCP Cloud Storage bucket resource
type Bucket struct {
	Kind          string    `json:"kind"`
	ID            string    `json:"id"`
	SelfLink      string    `json:"selfLink"`
	Name          string    `json:"name"`
	ProjectNumber string    `json:"projectNumber"`
	TimeCreated   time.Time `json:"timeCreated"`
	Updated       time.Time `json:"updated"`
	Location      string    `json:"location"`
	StorageClass  string    `json:"storageClass"`
	Etag          string    `json:"etag"`
}

// BucketListResponse represents the GCP API response for listing buckets
type BucketListResponse struct {
	Kind  string   `json:"kind"`
	Items []Bucket `json:"items"`
}

// NewBucket creates a new Bucket with the given parameters and sets default values
func NewBucket(name, projectNumber, location, storageClass string) Bucket {
	now := time.Now().UTC()
	return Bucket{
		Kind:          "storage#bucket",
		ID:            name,
		SelfLink:      fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s", name),
		Name:          name,
		ProjectNumber: projectNumber,
		TimeCreated:   now,
		Updated:       now,
		Location:      location,
		StorageClass:  storageClass,
		Etag:          generateEtag(),
	}
}

// generateEtag generates a simple etag for the bucket
func generateEtag() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
}

// validBucketNameChars matches lowercase letters, numbers, hyphens, underscores, and periods
var validBucketNameChars = regexp.MustCompile(`^[a-z0-9\-_.]+$`)

// ValidateBucketName validates a bucket name according to GCP naming rules
// Rules:
// - 3-63 characters long
// - Lowercase letters, numbers, hyphens, underscores, and periods
// - Must start and end with letter or number
// - Cannot be IP address format
// - Cannot start with "goog" prefix
func ValidateBucketName(name string) error {
	// Check length (3-63 characters)
	if len(name) < 3 || len(name) > 63 {
		return errors.New("bucket name must be between 3 and 63 characters")
	}

	// Check for valid characters
	if !validBucketNameChars.MatchString(name) {
		return errors.New("bucket name can only contain lowercase letters, numbers, hyphens, underscores, and periods")
	}

	// Check start and end characters (must be letter or number)
	firstChar := name[0]
	lastChar := name[len(name)-1]
	if !isAlphanumeric(firstChar) || !isAlphanumeric(lastChar) {
		return errors.New("bucket name must start and end with a letter or number")
	}

	// Check for IP address format
	if isIPAddress(name) {
		return errors.New("bucket name cannot be an IP address")
	}

	// Check for "goog" prefix
	if strings.HasPrefix(name, "goog") {
		return errors.New("bucket name cannot start with 'goog' prefix")
	}

	return nil
}

// isAlphanumeric checks if a byte is a lowercase letter or number
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// isIPAddress checks if the string is a valid IP address
func isIPAddress(s string) bool {
	return net.ParseIP(s) != nil
}