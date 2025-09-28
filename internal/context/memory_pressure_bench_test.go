package context

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// BenchmarkMemoryPressure_UnboundedMap demonstrates memory pressure with simple map
func BenchmarkMemoryPressure_UnboundedMap(b *testing.B) {
	scenarios := []struct {
		name          string
		numVariables  int
		valueSize     int
	}{
		{"10K_vars_small_values", 10000, 100},
		{"100K_vars_small_values", 100000, 100},
		{"500K_vars_small_values", 500000, 100},
		{"10K_vars_large_values", 10000, 10000},    // 10KB per variable
		{"100K_vars_large_values", 100000, 10000},  // 1GB total data
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			var m runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m)
			startMem := m.Alloc

			// Simple map (unbounded growth)
			variables := make(map[string]string)
			var mutex sync.RWMutex

			// Simulate variable accumulation over time
			largeValue := strings.Repeat("x", scenario.valueSize)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Add variables (simulating message history growth)
				for j := 0; j < scenario.numVariables/b.N; j++ {
					varName := fmt.Sprintf("var_%d_%d", i, j)
					mutex.Lock()
					variables[varName] = largeValue
					mutex.Unlock()
				}

				// Access some variables (simulating normal usage)
				if len(variables) > 0 {
					accessKey := fmt.Sprintf("var_%d_0", i)
					mutex.RLock()
					_ = variables[accessKey]
					mutex.RUnlock()
				}
			}
			b.StopTimer()

			runtime.GC()
			runtime.ReadMemStats(&m)
			endMem := m.Alloc
			memUsed := endMem - startMem

			b.ReportMetric(float64(memUsed)/1024/1024, "MB_used")
			b.ReportMetric(float64(len(variables)), "total_vars")
		})
	}
}

// BenchmarkMemoryPressure_LRUCache demonstrates controlled memory with LRU cache
func BenchmarkMemoryPressure_LRUCache(b *testing.B) {
	scenarios := []struct {
		name          string
		numVariables  int
		valueSize     int
		cacheSize     int
	}{
		{"10K_vars_small_values", 10000, 100, 1000},
		{"100K_vars_small_values", 100000, 100, 1000},
		{"500K_vars_small_values", 500000, 100, 1000},
		{"10K_vars_large_values", 10000, 10000, 1000},
		{"100K_vars_large_values", 100000, 10000, 1000},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			var m runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m)
			startMem := m.Alloc

			// LRU cache (bounded growth)
			cache := NewVariableLRUCache(scenario.cacheSize)

			// Simulate variable accumulation over time
			largeValue := strings.Repeat("x", scenario.valueSize)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Add variables (simulating message history growth)
				for j := 0; j < scenario.numVariables/b.N; j++ {
					varName := fmt.Sprintf("var_%d_%d", i, j)
					cache.Set(varName, largeValue)
				}

				// Access some variables (simulating normal usage)
				if cache.Size() > 0 {
					accessKey := fmt.Sprintf("var_%d_0", i)
					cache.Get(accessKey)
				}
			}
			b.StopTimer()

			runtime.GC()
			runtime.ReadMemStats(&m)
			endMem := m.Alloc
			memUsed := endMem - startMem

			b.ReportMetric(float64(memUsed)/1024/1024, "MB_used")
			b.ReportMetric(float64(cache.Size()), "total_vars")
		})
	}
}

// BenchmarkMemoryPressure_Comparison directly compares both approaches side by side
func BenchmarkMemoryPressure_Comparison(b *testing.B) {
	const numVariables = 50000
	const valueSize = 1000 // 1KB per variable
	const cacheSize = 2000

	b.Run("unbounded_map", func(b *testing.B) {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		startMem := m.Alloc

		variables := make(map[string]string)
		var mutex sync.RWMutex
		largeValue := strings.Repeat("data", valueSize/4)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate continuous variable creation (like message history)
			varName := fmt.Sprintf("message_%d", i)
			mutex.Lock()
			variables[varName] = largeValue
			mutex.Unlock()

			// Occasionally access recent variables
			if i > 0 && i%10 == 0 {
				accessKey := fmt.Sprintf("message_%d", i-1)
				mutex.RLock()
				_ = variables[accessKey]
				mutex.RUnlock()
			}
		}
		b.StopTimer()

		runtime.GC()
		runtime.ReadMemStats(&m)
		endMem := m.Alloc
		memUsed := endMem - startMem

		b.ReportMetric(float64(memUsed)/1024/1024, "MB_used")
		b.ReportMetric(float64(len(variables)), "total_vars")
		b.ReportMetric(float64(memUsed)/float64(len(variables)), "bytes_per_var")
	})

	b.Run("lru_cache", func(b *testing.B) {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		startMem := m.Alloc

		cache := NewVariableLRUCache(cacheSize)
		largeValue := strings.Repeat("data", valueSize/4)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate continuous variable creation (like message history)
			varName := fmt.Sprintf("message_%d", i)
			cache.Set(varName, largeValue)

			// Occasionally access recent variables
			if i > 0 && i%10 == 0 {
				accessKey := fmt.Sprintf("message_%d", i-1)
				cache.Get(accessKey)
			}
		}
		b.StopTimer()

		runtime.GC()
		runtime.ReadMemStats(&m)
		endMem := m.Alloc
		memUsed := endMem - startMem

		b.ReportMetric(float64(memUsed)/1024/1024, "MB_used")
		b.ReportMetric(float64(cache.Size()), "total_vars")
		if cache.Size() > 0 {
			b.ReportMetric(float64(memUsed)/float64(cache.Size()), "bytes_per_var")
		}
	})
}

// BenchmarkMemoryPressure_LongRunningSession simulates a very long NeuroShell session
func BenchmarkMemoryPressure_LongRunningSession(b *testing.B) {
	b.Run("unbounded_growth", func(b *testing.B) {
		variables := make(map[string]string)
		var mutex sync.RWMutex

		var m runtime.MemStats
		var memUsagePoints []float64

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate 10 variables per iteration (messages, outputs, etc.)
			for j := 0; j < 10; j++ {
				varName := fmt.Sprintf("msg_%d_%d", i, j)
				value := fmt.Sprintf("message_content_%d_%d_with_some_longer_text", i, j)
				mutex.Lock()
				variables[varName] = value
				mutex.Unlock()
			}

			// Sample memory usage every 1000 iterations
			if i%1000 == 0 {
				runtime.GC()
				runtime.ReadMemStats(&m)
				memUsagePoints = append(memUsagePoints, float64(m.Alloc)/1024/1024)
			}
		}
		b.StopTimer()

		// Report final memory usage
		if len(memUsagePoints) > 0 {
			b.ReportMetric(memUsagePoints[len(memUsagePoints)-1], "final_MB")
			if len(memUsagePoints) > 1 {
				growth := memUsagePoints[len(memUsagePoints)-1] - memUsagePoints[0]
				b.ReportMetric(growth, "memory_growth_MB")
			}
		}
		b.ReportMetric(float64(len(variables)), "total_vars")
	})

	b.Run("lru_bounded", func(b *testing.B) {
		cache := NewVariableLRUCache(5000) // Reasonable cache size

		var m runtime.MemStats
		var memUsagePoints []float64

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate 10 variables per iteration (messages, outputs, etc.)
			for j := 0; j < 10; j++ {
				varName := fmt.Sprintf("msg_%d_%d", i, j)
				value := fmt.Sprintf("message_content_%d_%d_with_some_longer_text", i, j)
				cache.Set(varName, value)
			}

			// Sample memory usage every 1000 iterations
			if i%1000 == 0 {
				runtime.GC()
				runtime.ReadMemStats(&m)
				memUsagePoints = append(memUsagePoints, float64(m.Alloc)/1024/1024)
			}
		}
		b.StopTimer()

		// Report final memory usage
		if len(memUsagePoints) > 0 {
			b.ReportMetric(memUsagePoints[len(memUsagePoints)-1], "final_MB")
			if len(memUsagePoints) > 1 {
				growth := memUsagePoints[len(memUsagePoints)-1] - memUsagePoints[0]
				b.ReportMetric(growth, "memory_growth_MB")
			}
		}
		b.ReportMetric(float64(cache.Size()), "total_vars")
	})
}

// BenchmarkMemoryPressure_AccessPatterns tests different access patterns
func BenchmarkMemoryPressure_AccessPatterns(b *testing.B) {
	const totalVariables = 20000
	const cacheSize = 2000

	// Create test data
	testValue := strings.Repeat("test_data_", 100) // ~1KB per variable

	b.Run("random_access_unbounded", func(b *testing.B) {
		variables := make(map[string]string)
		var mutex sync.RWMutex

		// Pre-populate
		for i := 0; i < totalVariables; i++ {
			variables[strconv.Itoa(i)] = testValue
		}

		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		startMem := m.Alloc

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := strconv.Itoa(i % totalVariables)
			mutex.RLock()
			_ = variables[key]
			mutex.RUnlock()
		}
		b.StopTimer()

		runtime.ReadMemStats(&m)
		b.ReportMetric(float64(m.Alloc-startMem)/1024/1024, "MB_used")
	})

	b.Run("random_access_lru", func(b *testing.B) {
		cache := NewVariableLRUCache(cacheSize)

		// Pre-populate (will cause evictions)
		for i := 0; i < totalVariables; i++ {
			cache.Set(strconv.Itoa(i), testValue)
		}

		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		startMem := m.Alloc

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := strconv.Itoa(i % totalVariables)
			cache.Get(key)
		}
		b.StopTimer()

		runtime.ReadMemStats(&m)
		b.ReportMetric(float64(m.Alloc-startMem)/1024/1024, "MB_used")
		b.ReportMetric(float64(cache.Size()), "cache_size")
	})
}