package context

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// BenchmarkVariableLRUCache_LargeDataset benchmarks performance with a large number of variables
func BenchmarkVariableLRUCache_LargeDataset_Set(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewVariableLRUCache(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("var_%d", i%size)
				value := fmt.Sprintf("value_%d", i%size)
				cache.Set(key, value)
			}
		})
	}
}

func BenchmarkVariableLRUCache_LargeDataset_Get(b *testing.B) {
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewVariableLRUCache(size)

			// Pre-populate cache
			for i := 0; i < size; i++ {
				cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("value_%d", i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("var_%d", i%size)
				cache.Get(key)
			}
		})
	}
}

// BenchmarkVariableLRUCache_FrequentlyAccessedVariables benchmarks the performance
// of frequently accessed variables like _output, _error that are pinned
func BenchmarkVariableLRUCache_FrequentlyAccessedVariables(b *testing.B) {
	cache := NewVariableLRUCache(10000)

	// Pre-populate with many variables
	for i := 0; i < 5000; i++ {
		cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Set up frequently accessed variables
	cache.Set("_output", "command output")
	cache.Set("_error", "error message")
	cache.Set("_status", "0")
	cache.Set("_elapsed", "100ms")

	frequentVars := []string{"_output", "_error", "_status", "_elapsed"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		varName := frequentVars[i%len(frequentVars)]
		cache.Get(varName)
	}
}

// BenchmarkVariableLRUCache_LongValues tests performance with essay-length values
func BenchmarkVariableLRUCache_LongValues(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	// Create different sized "essays"
	valueSizes := []int{1000, 10000, 100000} // characters

	for _, size := range valueSizes {
		b.Run(fmt.Sprintf("value_size_%d", size), func(b *testing.B) {
			longValue := strings.Repeat("This is a very long essay text with meaningful content. ", size/56)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("essay_%d", i%100)
				cache.Set(key, longValue)
			}
		})
	}
}

// BenchmarkVariableLRUCache_GetLongValues tests retrieval of long values
func BenchmarkVariableLRUCache_GetLongValues(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	// Pre-populate with long values
	for i := 0; i < 100; i++ {
		longValue := strings.Repeat("Essay content paragraph. ", 1000)
		cache.Set(fmt.Sprintf("essay_%d", i), longValue)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("essay_%d", i%100)
		cache.Get(key)
	}
}

// BenchmarkVariableLRUCache_InterpolationPattern simulates variable interpolation workload
func BenchmarkVariableLRUCache_InterpolationPattern(b *testing.B) {
	cache := NewVariableLRUCache(10000)

	// Set up typical variables used in interpolation
	cache.Set("user", "developer")
	cache.Set("project", "NeuroShell")
	cache.Set("task", "benchmarking")
	cache.Set("_output", "success")
	cache.Set("1", "Hello world")  // Message history
	cache.Set("2", "How are you?") // Message history

	// Add many other variables to create realistic cache pressure
	for i := 0; i < 5000; i++ {
		cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Simulate interpolation access pattern (some variables accessed more frequently)
	accessPattern := []string{"user", "project", "_output", "1", "2", "task", "_output", "user"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		varName := accessPattern[i%len(accessPattern)]
		cache.Get(varName)
	}
}

// BenchmarkVariableLRUCache_EvictionPressure tests performance under high eviction pressure
func BenchmarkVariableLRUCache_EvictionPressure(b *testing.B) {
	cache := NewVariableLRUCache(1000) // Small cache to force evictions

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will cause constant evictions as we exceed cache size
		key := fmt.Sprintf("var_%d", i)
		value := fmt.Sprintf("value_%d", i)
		cache.Set(key, value)
	}
}

// BenchmarkVariableLRUCache_RandomAccess tests performance with random access patterns
func BenchmarkVariableLRUCache_RandomAccess(b *testing.B) {
	cache := NewVariableLRUCache(10000)

	// Pre-populate
	for i := 0; i < 5000; i++ {
		cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Create deterministic random sequence for reproducible benchmarks
	rng := rand.New(rand.NewSource(42))
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("var_%d", rng.Intn(5000))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		cache.Get(key)
	}
}

// BenchmarkVariableLRUCache_UpdateExisting tests performance of updating existing variables
func BenchmarkVariableLRUCache_UpdateExisting(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	// Pre-populate
	for i := 0; i < 500; i++ {
		cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("initial_value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("var_%d", i%500)
		newValue := fmt.Sprintf("updated_value_%d", i)
		cache.Set(key, newValue)
	}
}

// BenchmarkVariableLRUCache_GetAll tests performance of retrieving all variables
func BenchmarkVariableLRUCache_GetAll(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			cache := NewVariableLRUCache(size + 1000) // Ensure no evictions

			// Pre-populate
			for i := 0; i < size; i++ {
				cache.Set(fmt.Sprintf("var_%d", i), fmt.Sprintf("value_%d", i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = cache.GetAll()
			}
		})
	}
}

// BenchmarkVariableLRUCache_PinnedVsUnpinned compares access performance
func BenchmarkVariableLRUCache_PinnedVsUnpinned(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	// Set up pinned and unpinned variables
	cache.Set("_output", "pinned_value")       // Automatically pinned
	cache.Set("regular_var", "unpinned_value") // Not pinned
	cache.SetPinned("manually_pinned", true)
	cache.Set("manually_pinned", "manual_pinned_value")

	// Add pressure to force LRU operations
	for i := 0; i < 500; i++ {
		cache.Set(fmt.Sprintf("pressure_%d", i), fmt.Sprintf("value_%d", i))
	}

	b.Run("pinned_access", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Get("_output")
		}
	})

	b.Run("unpinned_access", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Get("regular_var")
		}
	})
}

// BenchmarkVariableLRUCache_Memory tests memory allocation patterns
func BenchmarkVariableLRUCache_Memory(b *testing.B) {
	cache := NewVariableLRUCache(1000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("var_%d", i%1000)
		value := fmt.Sprintf("value_%d", i%1000)
		cache.Set(key, value)
		cache.Get(key)
	}
}

// BenchmarkVariableLRUCache_RealWorldScenario simulates a realistic NeuroShell workload
func BenchmarkVariableLRUCache_RealWorldScenario(b *testing.B) {
	cache := NewVariableLRUCache(10000)

	// Initialize with system and frequently used variables
	systemVars := map[string]string{
		"_output":          "last command output",
		"_error":           "",
		"_status":          "0",
		"_elapsed":         "25ms",
		"_style":           "modern",
		"_default_command": "echo",
		"_completion_mode": "tab",
	}

	for k, v := range systemVars {
		cache.Set(k, v)
	}

	// Simulate user variables accumulating over time
	for i := 0; i < 5000; i++ {
		cache.Set(fmt.Sprintf("user_var_%d", i), fmt.Sprintf("user_value_%d", i))
	}

	// Message history variables (recent messages are accessed more frequently)
	for i := 1; i <= 100; i++ {
		cache.Set(strconv.Itoa(i), fmt.Sprintf("message_%d_content", i))
	}

	// Simulate real workload: 70% reads, 30% writes
	// 50% access to frequent vars, 30% to recent vars, 20% to random vars
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		operation := i % 10
		if operation < 7 { // 70% reads
			accessType := i % 10
			switch {
			case accessType < 5: // 50% frequent vars
				frequentVars := []string{"_output", "_error", "_status", "1", "2", "3"}
				cache.Get(frequentVars[i%len(frequentVars)])
			case accessType < 8: // 30% recent vars
				recentVar := strconv.Itoa((i % 20) + 1)
				cache.Get(recentVar)
			default: // 20% random vars
				randomVar := fmt.Sprintf("user_var_%d", i%5000)
				cache.Get(randomVar)
			}
		} else { // 30% writes
			key := fmt.Sprintf("new_var_%d", i)
			value := fmt.Sprintf("new_value_%d", i)
			cache.Set(key, value)
		}
	}
}

// BenchmarkVariableLRUCache_CompareWithMap compares LRU cache vs simple map performance
func BenchmarkVariableLRUCache_CompareWithMap(b *testing.B) {
	// LRU Cache benchmark
	b.Run("lru_cache", func(b *testing.B) {
		cache := NewVariableLRUCache(1000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("var_%d", i%1000)
			value := fmt.Sprintf("value_%d", i%1000)

			if i%3 == 0 {
				cache.Set(key, value)
			} else {
				cache.Get(key)
			}
		}
	})

	// Simple map with mutex benchmark (for comparison)
	b.Run("simple_map", func(b *testing.B) {
		m := make(map[string]string)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("var_%d", i%1000)
			value := fmt.Sprintf("value_%d", i%1000)

			if i%3 == 0 {
				m[key] = value
			} else {
				_ = m[key]
			}
		}
	})
}
