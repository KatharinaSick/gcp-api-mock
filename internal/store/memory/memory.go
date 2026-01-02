package memory

import (
	"sync"

	"github.com/KatharinaSick/gcp-api-mock/internal/store"
)

// MemoryStore is a thread-safe in-memory implementation of store.Store
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// New creates a new MemoryStore
func New() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a resource by ID
func (m *MemoryStore) Get(id string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[id]
	return val, ok
}

// Set stores a resource with the given ID
func (m *MemoryStore) Set(id string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = value
}

// Delete removes a resource by ID
func (m *MemoryStore) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; ok {
		delete(m.data, id)
		return true
	}
	return false
}

// List returns all resources
func (m *MemoryStore) List() []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]interface{}, 0, len(m.data))
	for _, v := range m.data {
		result = append(result, v)
	}
	return result
}

// Exists checks if a resource exists
func (m *MemoryStore) Exists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[id]
	return ok
}

// Ensure MemoryStore implements store.Store
var _ store.Store = (*MemoryStore)(nil)