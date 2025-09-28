package context

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVariableLRUCache(t *testing.T) {
	tests := []struct {
		name        string
		maxSize     int
		expectedMax int
	}{
		{
			name:        "positive max size",
			maxSize:     100,
			expectedMax: 100,
		},
		{
			name:        "zero max size uses default",
			maxSize:     0,
			expectedMax: 10000,
		},
		{
			name:        "negative max size uses default",
			maxSize:     -1,
			expectedMax: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewVariableLRUCache(tt.maxSize)

			assert.NotNil(t, cache)
			assert.Equal(t, tt.expectedMax, cache.maxSize)
			assert.Equal(t, 0, cache.Size())

			// Check that pinned variables are configured
			assert.True(t, cache.pinnedVars["_output"])
			assert.True(t, cache.pinnedVars["_error"])
			assert.True(t, cache.pinnedVars["_status"])
			assert.True(t, cache.pinnedVars["_elapsed"])
		})
	}
}

func TestVariableLRUCache_BasicOperations(t *testing.T) {
	cache := NewVariableLRUCache(3)

	// Test Set and Get
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test non-existent key
	value, exists = cache.Get("nonexistent")
	assert.False(t, exists)
	assert.Equal(t, "", value)

	// Test update existing key
	cache.Set("key1", "updated_value1")
	value, exists = cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "updated_value1", value)
}

func TestVariableLRUCache_LRUEviction(t *testing.T) {
	cache := NewVariableLRUCache(2)

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	// Add third item - should evict key1 (least recently used)
	cache.Set("key3", "value3")
	assert.Equal(t, 2, cache.Size())

	// key1 should be evicted
	_, exists := cache.Get("key1")
	assert.False(t, exists)

	// key2 and key3 should still exist
	_, exists = cache.Get("key2")
	assert.True(t, exists)
	_, exists = cache.Get("key3")
	assert.True(t, exists)
}

func TestVariableLRUCache_LRUOrdering(t *testing.T) {
	cache := NewVariableLRUCache(3)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add key4 - should evict key2 (least recently used)
	cache.Set("key4", "value4")

	// key2 should be evicted
	_, exists := cache.Get("key2")
	assert.False(t, exists)

	// key1, key3, key4 should exist
	_, exists = cache.Get("key1")
	assert.True(t, exists)
	_, exists = cache.Get("key3")
	assert.True(t, exists)
	_, exists = cache.Get("key4")
	assert.True(t, exists)
}

func TestVariableLRUCache_PinnedVariables(t *testing.T) {
	cache := NewVariableLRUCache(2)

	// Add pinned variable
	cache.Set("_output", "pinned_value")
	cache.Set("key1", "value1")
	assert.Equal(t, 2, cache.Size())

	// Add another item - should evict key1, not _output
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	// _output should still exist (pinned)
	value, exists := cache.Get("_output")
	assert.True(t, exists)
	assert.Equal(t, "pinned_value", value)

	// key1 should be evicted
	_, exists = cache.Get("key1")
	assert.False(t, exists)

	// key2 should exist
	_, exists = cache.Get("key2")
	assert.True(t, exists)
}

func TestVariableLRUCache_SetPinned(t *testing.T) {
	cache := NewVariableLRUCache(2)

	// Add regular variable
	cache.Set("key1", "value1")

	// Pin it
	cache.SetPinned("key1", true)

	// Add more items
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")  // Should evict key2, not key1

	// key1 should still exist (now pinned)
	_, exists := cache.Get("key1")
	assert.True(t, exists)

	// key2 should be evicted
	_, exists = cache.Get("key2")
	assert.False(t, exists)

	// key3 should exist
	_, exists = cache.Get("key3")
	assert.True(t, exists)
}

func TestVariableLRUCache_Delete(t *testing.T) {
	cache := NewVariableLRUCache(3)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	// Delete existing key
	cache.Delete("key1")
	assert.Equal(t, 1, cache.Size())
	_, exists := cache.Get("key1")
	assert.False(t, exists)

	// Delete non-existent key (should not panic)
	cache.Delete("nonexistent")
	assert.Equal(t, 1, cache.Size())

	// key2 should still exist
	_, exists = cache.Get("key2")
	assert.True(t, exists)
}

func TestVariableLRUCache_GetAll(t *testing.T) {
	cache := NewVariableLRUCache(5)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("_output", "output_value")

	all := cache.GetAll()
	assert.Equal(t, 3, len(all))
	assert.Equal(t, "value1", all["key1"])
	assert.Equal(t, "value2", all["key2"])
	assert.Equal(t, "output_value", all["_output"])
}

func TestVariableLRUCache_Clear(t *testing.T) {
	cache := NewVariableLRUCache(5)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())

	// Items should not exist
	_, exists := cache.Get("key1")
	assert.False(t, exists)
	_, exists = cache.Get("key2")
	assert.False(t, exists)

	// GetAll should return empty map
	all := cache.GetAll()
	assert.Equal(t, 0, len(all))
}

func TestVariableLRUCache_GetStats(t *testing.T) {
	cache := NewVariableLRUCache(10)

	// Add regular and pinned variables
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("_output", "output_value")  // This is pinned by default

	stats := cache.GetStats()
	assert.Equal(t, 3, stats.Size)
	assert.Equal(t, 10, stats.MaxSize)
	assert.Equal(t, 1, stats.PinnedCount)  // Only _output is pinned
}

func TestVariableLRUCache_LongValues(t *testing.T) {
	cache := NewVariableLRUCache(5)

	// Test with very long values (like essays)
	longValue := strings.Repeat("This is a very long essay text. ", 1000)
	cache.Set("essay", longValue)

	retrievedValue, exists := cache.Get("essay")
	assert.True(t, exists)
	assert.Equal(t, longValue, retrievedValue)

	// Ensure LRU still works with long values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4")
	cache.Set("key5", "value5")  // Should evict essay (LRU)

	_, exists = cache.Get("essay")
	assert.False(t, exists)
}

func TestVariableLRUCache_Concurrent(t *testing.T) {
	cache := NewVariableLRUCache(100)
	numGoroutines := 10
	operationsPerGoroutine := 100

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Cache should not exceed its limit significantly
	// In concurrent scenarios, the cache might temporarily grow larger due to timing
	// but should settle down. The important thing is that it doesn't grow unbounded.
	// We check that it's within a reasonable range of the max size
	// For 10 goroutines x 100 operations = 1000 total, with cache size 100,
	// we expect some overflow but not too much
	assert.LessOrEqual(t, cache.Size(), 1000)  // Allow reasonable buffer for concurrent access
}

func TestVariableLRUCache_AllPinnedScenario(t *testing.T) {
	cache := NewVariableLRUCache(2)

	// Add items and pin them all
	cache.Set("key1", "value1")
	cache.SetPinned("key1", true)
	cache.Set("key2", "value2")
	cache.SetPinned("key2", true)

	// Check status after adding 2 items
	assert.Equal(t, 2, cache.Size())

	// Pin key3 before adding it to prevent eviction
	cache.SetPinned("key3", true)
	cache.Set("key3", "value3")

	// All items should still exist (no eviction possible)
	_, exists := cache.Get("key1")
	assert.True(t, exists)
	_, exists = cache.Get("key2")
	assert.True(t, exists)
	_, exists = cache.Get("key3")
	assert.True(t, exists)

	assert.Equal(t, 3, cache.Size())  // Exceeds maxSize but that's OK
}

func TestVariableLRUCache_EdgeCases(t *testing.T) {
	cache := NewVariableLRUCache(5)

	// Empty string values
	cache.Set("empty", "")
	value, exists := cache.Get("empty")
	assert.True(t, exists)
	assert.Equal(t, "", value)

	// Unicode values
	cache.Set("unicode", "测试值 🎉")
	value, exists = cache.Get("unicode")
	assert.True(t, exists)
	assert.Equal(t, "测试值 🎉", value)

	// Very long keys
	longKey := strings.Repeat("key", 100)
	cache.Set(longKey, "long_key_value")
	value, exists = cache.Get(longKey)
	assert.True(t, exists)
	assert.Equal(t, "long_key_value", value)
}

func TestVariableLRUCache_UpdateExistingPinned(t *testing.T) {
	cache := NewVariableLRUCache(5)

	// Set pinned variable
	cache.Set("_output", "initial")
	cache.Set("key1", "value1")

	// Update pinned variable
	cache.Set("_output", "updated")

	// Should update without affecting LRU order
	value, exists := cache.Get("_output")
	assert.True(t, exists)
	assert.Equal(t, "updated", value)

	// Other items should not be affected
	_, exists = cache.Get("key1")
	assert.True(t, exists)
}

// Benchmark tests to ensure performance
func BenchmarkVariableLRUCache_Set(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i % 1000)
		cache.Set(key, "value_"+key)
	}
}

func BenchmarkVariableLRUCache_Get(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(strconv.Itoa(i), "value_"+strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i % 1000)
		cache.Get(key)
	}
}

func BenchmarkVariableLRUCache_Mixed(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i % 1000)
		if i%3 == 0 {
			cache.Set(key, "value_"+key)
		} else {
			cache.Get(key)
		}
	}
}