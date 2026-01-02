package store

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStore is a simple in-memory store for testing
type mockStore struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string]interface{}),
	}
}

func (m *mockStore) Get(id string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[id]
	return val, ok
}

func (m *mockStore) Set(id string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[id] = value
}

func (m *mockStore) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; ok {
		delete(m.data, id)
		return true
	}
	return false
}

func (m *mockStore) List() []interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]interface{}, 0, len(m.data))
	for _, v := range m.data {
		result = append(result, v)
	}
	return result
}

func (m *mockStore) Exists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[id]
	return ok
}

// Helper to create a factory with mock store constructor
func newTestFactory() *StoreFactory {
	return NewStoreFactory(func() Store {
		return newMockStore()
	})
}

func TestNewStoreFactory(t *testing.T) {
	factory := newTestFactory()
	require.NotNil(t, factory)
	assert.NotNil(t, factory.stores)
	assert.Empty(t, factory.stores)
}

func TestStoreFactory_GetStore(t *testing.T) {
	t.Run("creates new store if not exists", func(t *testing.T) {
		factory := newTestFactory()

		store := factory.GetStore("storage")
		require.NotNil(t, store)
	})

	t.Run("returns existing store", func(t *testing.T) {
		factory := newTestFactory()

		store1 := factory.GetStore("storage")
		store2 := factory.GetStore("storage")

		assert.Same(t, store1, store2)
	})

	t.Run("creates separate stores for different services", func(t *testing.T) {
		factory := newTestFactory()

		storageStore := factory.GetStore("storage")
		pubsubStore := factory.GetStore("pubsub")

		assert.NotSame(t, storageStore, pubsubStore)
	})

	t.Run("stores are independent", func(t *testing.T) {
		factory := newTestFactory()

		storageStore := factory.GetStore("storage")
		pubsubStore := factory.GetStore("pubsub")

		storageStore.Set("bucket-1", "data-1")
		pubsubStore.Set("topic-1", "data-2")

		// Check storage store
		val, ok := storageStore.Get("bucket-1")
		assert.True(t, ok)
		assert.Equal(t, "data-1", val)

		_, ok = storageStore.Get("topic-1")
		assert.False(t, ok)

		// Check pubsub store
		val, ok = pubsubStore.Get("topic-1")
		assert.True(t, ok)
		assert.Equal(t, "data-2", val)

		_, ok = pubsubStore.Get("bucket-1")
		assert.False(t, ok)
	})
}

func TestStoreFactory_GetOrCreateStore(t *testing.T) {
	t.Run("is alias for GetStore", func(t *testing.T) {
		factory := newTestFactory()

		store1 := factory.GetStore("storage")
		store2 := factory.GetOrCreateStore("storage")

		assert.Same(t, store1, store2)
	})
}

func TestStoreFactory_HasStore(t *testing.T) {
	t.Run("returns false for non-existing service", func(t *testing.T) {
		factory := newTestFactory()

		assert.False(t, factory.HasStore("storage"))
	})

	t.Run("returns true for existing service", func(t *testing.T) {
		factory := newTestFactory()
		factory.GetStore("storage")

		assert.True(t, factory.HasStore("storage"))
	})
}

func TestStoreFactory_ListServices(t *testing.T) {
	t.Run("returns empty list for new factory", func(t *testing.T) {
		factory := newTestFactory()

		services := factory.ListServices()
		assert.Empty(t, services)
	})

	t.Run("returns all service names", func(t *testing.T) {
		factory := newTestFactory()
		factory.GetStore("storage")
		factory.GetStore("pubsub")
		factory.GetStore("bigquery")

		services := factory.ListServices()
		assert.Len(t, services, 3)
		assert.Contains(t, services, "storage")
		assert.Contains(t, services, "pubsub")
		assert.Contains(t, services, "bigquery")
	})
}

func TestStoreFactory_Reset(t *testing.T) {
	t.Run("clears all stores", func(t *testing.T) {
		factory := newTestFactory()
		factory.GetStore("storage")
		factory.GetStore("pubsub")

		assert.True(t, factory.HasStore("storage"))
		assert.True(t, factory.HasStore("pubsub"))

		factory.Reset()

		assert.False(t, factory.HasStore("storage"))
		assert.False(t, factory.HasStore("pubsub"))
		assert.Empty(t, factory.ListServices())
	})
}

func TestStoreFactory_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent GetStore calls", func(t *testing.T) {
		factory := newTestFactory()
		var wg sync.WaitGroup
		iterations := 100
		stores := make([]Store, iterations)

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				stores[i] = factory.GetStore("storage")
			}(i)
		}

		wg.Wait()

		// All should be the same store
		for i := 1; i < iterations; i++ {
			assert.Same(t, stores[0], stores[i])
		}
	})

	t.Run("concurrent GetStore for different services", func(t *testing.T) {
		factory := newTestFactory()
		var wg sync.WaitGroup
		services := []string{"storage", "pubsub", "bigquery", "firestore", "spanner"}

		for _, svc := range services {
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(svc string) {
					defer wg.Done()
					store := factory.GetStore(svc)
					store.Set("key", "value")
					store.Get("key")
				}(svc)
			}
		}

		wg.Wait()

		// All services should have stores
		for _, svc := range services {
			assert.True(t, factory.HasStore(svc))
		}
	})

	t.Run("concurrent operations", func(t *testing.T) {
		factory := newTestFactory()
		var wg sync.WaitGroup

		// GetStore operations
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				factory.GetStore("service-" + string(rune(i%5)))
			}(i)
		}

		// HasStore operations
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				factory.HasStore("service-" + string(rune(i%5)))
			}(i)
		}

		// ListServices operations
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				factory.ListServices()
			}()
		}

		wg.Wait()
	})
}