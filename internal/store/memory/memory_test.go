package memory

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	store := New()
	require.NotNil(t, store)
	assert.NotNil(t, store.data)
	assert.Empty(t, store.data)
}

func TestMemoryStore_Get(t *testing.T) {
	t.Run("returns value and true for existing key", func(t *testing.T) {
		store := New()
		store.data["test-key"] = "test-value"

		val, ok := store.Get("test-key")
		assert.True(t, ok)
		assert.Equal(t, "test-value", val)
	})

	t.Run("returns nil and false for non-existing key", func(t *testing.T) {
		store := New()

		val, ok := store.Get("non-existing")
		assert.False(t, ok)
		assert.Nil(t, val)
	})
}

func TestMemoryStore_Set(t *testing.T) {
	t.Run("stores a new value", func(t *testing.T) {
		store := New()
		store.Set("key1", "value1")

		assert.Equal(t, "value1", store.data["key1"])
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		store := New()
		store.Set("key1", "value1")
		store.Set("key1", "value2")

		assert.Equal(t, "value2", store.data["key1"])
	})

	t.Run("stores different types", func(t *testing.T) {
		store := New()

		// Store string
		store.Set("string-key", "string-value")
		assert.Equal(t, "string-value", store.data["string-key"])

		// Store int
		store.Set("int-key", 42)
		assert.Equal(t, 42, store.data["int-key"])

		// Store struct
		type TestStruct struct {
			Name string
			Age  int
		}
		testStruct := TestStruct{Name: "John", Age: 30}
		store.Set("struct-key", testStruct)
		assert.Equal(t, testStruct, store.data["struct-key"])

		// Store map
		testMap := map[string]string{"foo": "bar"}
		store.Set("map-key", testMap)
		assert.Equal(t, testMap, store.data["map-key"])
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	t.Run("deletes existing key and returns true", func(t *testing.T) {
		store := New()
		store.data["key1"] = "value1"

		ok := store.Delete("key1")
		assert.True(t, ok)
		_, exists := store.data["key1"]
		assert.False(t, exists)
	})

	t.Run("returns false for non-existing key", func(t *testing.T) {
		store := New()

		ok := store.Delete("non-existing")
		assert.False(t, ok)
	})
}

func TestMemoryStore_List(t *testing.T) {
	t.Run("returns empty slice for empty store", func(t *testing.T) {
		store := New()

		result := store.List()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns all values", func(t *testing.T) {
		store := New()
		store.data["key1"] = "value1"
		store.data["key2"] = "value2"
		store.data["key3"] = "value3"

		result := store.List()
		assert.Len(t, result, 3)
		assert.Contains(t, result, "value1")
		assert.Contains(t, result, "value2")
		assert.Contains(t, result, "value3")
	})
}

func TestMemoryStore_Exists(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		store := New()
		store.data["key1"] = "value1"

		assert.True(t, store.Exists("key1"))
	})

	t.Run("returns false for non-existing key", func(t *testing.T) {
		store := New()

		assert.False(t, store.Exists("non-existing"))
	})
}

// Concurrent access tests
func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent writes", func(t *testing.T) {
		store := New()
		var wg sync.WaitGroup
		iterations := 1000

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				store.Set("key", i)
			}(i)
		}

		wg.Wait()

		// Should have exactly one value for the key
		val, ok := store.Get("key")
		assert.True(t, ok)
		assert.NotNil(t, val)
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		store := New()
		store.Set("key", "initial")
		var wg sync.WaitGroup
		iterations := 1000

		// Start readers
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.Get("key")
			}()
		}

		// Start writers
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				store.Set("key", i)
			}(i)
		}

		wg.Wait()

		// Store should still be functional
		val, ok := store.Get("key")
		assert.True(t, ok)
		assert.NotNil(t, val)
	})

	t.Run("concurrent operations on multiple keys", func(t *testing.T) {
		store := New()
		var wg sync.WaitGroup
		iterations := 100
		keysPerIteration := 10

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for j := 0; j < keysPerIteration; j++ {
					key := "key" + string(rune(j))
					store.Set(key, i*keysPerIteration+j)
					store.Get(key)
					store.Exists(key)
				}
			}(i)
		}

		wg.Wait()

		// All keys should exist
		for j := 0; j < keysPerIteration; j++ {
			key := "key" + string(rune(j))
			assert.True(t, store.Exists(key))
		}
	})

	t.Run("concurrent delete and list", func(t *testing.T) {
		store := New()
		// Pre-populate
		for i := 0; i < 100; i++ {
			store.Set("key"+string(rune(i)), i)
		}

		var wg sync.WaitGroup

		// Start deleters
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				store.Delete("key" + string(rune(i)))
			}(i)
		}

		// Start listers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.List()
			}()
		}

		wg.Wait()

		// Should complete without panics or deadlocks
	})
}

// Integration test - typical usage patterns
func TestMemoryStore_Integration(t *testing.T) {
	t.Run("CRUD operations workflow", func(t *testing.T) {
		store := New()

		// Create
		store.Set("bucket-1", map[string]interface{}{
			"name":    "my-bucket",
			"region":  "us-central1",
			"created": "2024-01-01T00:00:00Z",
		})

		// Read
		val, ok := store.Get("bucket-1")
		require.True(t, ok)
		bucket, ok := val.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "my-bucket", bucket["name"])

		// Update
		bucket["region"] = "eu-west1"
		store.Set("bucket-1", bucket)
		
		val, ok = store.Get("bucket-1")
		require.True(t, ok)
		updatedBucket, ok := val.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "eu-west1", updatedBucket["region"])

		// List
		items := store.List()
		assert.Len(t, items, 1)

		// Delete
		deleted := store.Delete("bucket-1")
		assert.True(t, deleted)

		// Verify deleted
		_, ok = store.Get("bucket-1")
		assert.False(t, ok)
		assert.False(t, store.Exists("bucket-1"))
	})

	t.Run("multiple resources", func(t *testing.T) {
		store := New()

		// Add multiple resources
		store.Set("bucket-1", map[string]string{"name": "bucket-1"})
		store.Set("bucket-2", map[string]string{"name": "bucket-2"})
		store.Set("bucket-3", map[string]string{"name": "bucket-3"})

		// Verify all exist
		assert.True(t, store.Exists("bucket-1"))
		assert.True(t, store.Exists("bucket-2"))
		assert.True(t, store.Exists("bucket-3"))

		// List all
		items := store.List()
		assert.Len(t, items, 3)

		// Delete one
		store.Delete("bucket-2")

		// Verify state
		assert.True(t, store.Exists("bucket-1"))
		assert.False(t, store.Exists("bucket-2"))
		assert.True(t, store.Exists("bucket-3"))

		items = store.List()
		assert.Len(t, items, 2)
	})
}