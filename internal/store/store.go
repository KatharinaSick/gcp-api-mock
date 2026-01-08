// Package store provides an in-memory data store for the GCP API Mock.
package store

import (
	"sync"
)

// Store is the main in-memory data store for all GCP resources.
// It is safe for concurrent access.
type Store struct {
	mu sync.RWMutex

	// Future storage maps will be added here as services are implemented
	// Example:
	// buckets map[string]*storage.Bucket
	// objects map[string]map[string]*storage.Object
}

// New creates a new empty Store.
func New() *Store {
	return &Store{}
}

// Reset clears all data from the store.
// Useful for testing and resetting state.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset all maps when implemented
	// s.buckets = make(map[string]*storage.Bucket)
}
