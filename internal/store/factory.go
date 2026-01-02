package store

import (
	"sync"
)

// StoreConstructor is a function that creates a new Store
type StoreConstructor func() Store

// StoreFactory manages the creation and retrieval of service-specific stores
type StoreFactory struct {
	mu          sync.RWMutex
	stores      map[string]Store
	constructor StoreConstructor
}

// NewStoreFactory creates a new StoreFactory with the given constructor
func NewStoreFactory(constructor StoreConstructor) *StoreFactory {
	return &StoreFactory{
		stores:      make(map[string]Store),
		constructor: constructor,
	}
}

// GetStore returns a store for the given service name, creating one if it doesn't exist
func (f *StoreFactory) GetStore(serviceName string) Store {
	f.mu.RLock()
	if store, ok := f.stores[serviceName]; ok {
		f.mu.RUnlock()
		return store
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if store, ok := f.stores[serviceName]; ok {
		return store
	}

	store := f.constructor()
	f.stores[serviceName] = store
	return store
}

// GetOrCreateStore is an alias for GetStore (for clarity in usage)
func (f *StoreFactory) GetOrCreateStore(serviceName string) Store {
	return f.GetStore(serviceName)
}

// HasStore checks if a store exists for the given service name
func (f *StoreFactory) HasStore(serviceName string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.stores[serviceName]
	return ok
}

// ListServices returns all service names that have stores
func (f *StoreFactory) ListServices() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	services := make([]string, 0, len(f.stores))
	for name := range f.stores {
		services = append(services, name)
	}
	return services
}

// Reset clears all stores (useful for testing)
func (f *StoreFactory) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stores = make(map[string]Store)
}