package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
)

// Comprehensive performance tests for service operations

// BenchmarkServiceInitialization tests service initialization performance
func BenchmarkServiceInitialization(b *testing.B) {

	b.Run("VariableService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewVariableService()
			_ = service.Initialize()
		}
	})

	b.Run("ExecutorService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewExecutorService()
			_ = service.Initialize()
		}
	})
}

// BenchmarkServiceRegistry_HighLoad tests registry under high load
func BenchmarkServiceRegistry_HighLoad(b *testing.B) {
	registry := NewRegistry()

	// Pre-register many services
	for i := 0; i < 1000; i++ {
		service := NewMockService(fmt.Sprintf("service_%d", i))
		_ = registry.RegisterService(service)
	}

	b.Run("GetService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			serviceName := fmt.Sprintf("service_%d", i%1000)
			_, _ = registry.GetService(serviceName)
		}
	})

	b.Run("GetAllServices", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.GetAllServices()
		}
	})

	b.Run("InitializeAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.InitializeAll()
		}
	})
}

// BenchmarkVariableService_LargeDataset tests variable operations with many variables
func BenchmarkVariableService_LargeDataset(b *testing.B) {
	service := NewVariableService()

	// Create context with many variables
	vars := make(map[string]string)
	for i := 0; i < 10000; i++ {
		vars[fmt.Sprintf("var_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	err := service.Initialize()
	require.NoError(b, err)

	b.Run("Get_Existing", func(b *testing.B) {
		// Setup global context for testing
		ctx := context.NewTestContext()
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			varName := fmt.Sprintf("var_%d", i%10000)
			_, _ = service.Get(varName)
		}
	})

	b.Run("Get_NonExisting", func(b *testing.B) {
		// Setup global context for testing
		ctx := context.NewTestContext()
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = service.Get("nonexistent_var")
		}
	})

	b.Run("Set_NewVariables", func(b *testing.B) {
		// Setup global context for testing
		ctx := context.NewTestContext()
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			varName := fmt.Sprintf("new_var_%d", i)
			varValue := fmt.Sprintf("new_value_%d", i)
			_ = service.Set(varName, varValue)
		}
	})
}

// BenchmarkExecutorService_CommandParsing tests command parsing performance
func BenchmarkExecutorService_CommandParsing(b *testing.B) {
	service := NewExecutorService()

	err := service.Initialize()
	require.NoError(b, err)

	commands := []string{
		`\set[name="value"]`,
		`\get[name]`,
		`\send[model="gpt-4", temperature=0.7] Hello world`,
		`\run file.neuro`,
		`\bash ls -la`,
		`\send[system="You are helpful", max_tokens=1000] Write a detailed analysis`,
		`\set[complex_var="${nested_${var}_value}_with_${multiple}_variables"]`,
		`\get[@user]`,
		`\get[#session_id]`,
		`\command[arg1="value1", arg2="value2", arg3="value3", arg4="value4"]`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		_, _ = service.ParseCommand(cmd)
	}
}

// BenchmarkConcurrentServiceUsage tests concurrent service operations
func BenchmarkConcurrentServiceUsage(b *testing.B) {
	b.Run("VariableService_Concurrent", func(b *testing.B) {
		service := NewVariableService()
		err := service.Initialize()
		require.NoError(b, err)

		// Setup global context for testing
		ctx := context.NewTestContext()
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				varName := fmt.Sprintf("concurrent_var_%d", i)
				varValue := fmt.Sprintf("concurrent_value_%d", i)
				_ = service.Set(varName, varValue)
				_, _ = service.Get(varName)
				i++
			}
		})
	})

	b.Run("ExecutorService_Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Each goroutine gets its own service to avoid race conditions
				service := NewExecutorService()
				_ = service.Initialize()

				cmd := fmt.Sprintf(`\set[var_%d="value_%d"]`, i, i)
				_, _ = service.ParseCommand(cmd)
				i++
			}
		})
	})

	b.Run("Registry_Concurrent", func(b *testing.B) {
		registry := NewRegistry()

		// Pre-register some services
		for i := 0; i < 100; i++ {
			service := NewMockService(fmt.Sprintf("service_%d", i))
			_ = registry.RegisterService(service)
		}

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				serviceName := fmt.Sprintf("service_%d", i%100)
				_, _ = registry.GetService(serviceName)
				i++
			}
		})
	})
}

// BenchmarkMemoryUsage tests memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("ServiceCreation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vs := NewVariableService()
			es := NewExecutorService()

			// Prevent optimization
			_ = vs.Name()
			_ = es.Name()
		}
	})

	b.Run("RegistryOperations", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			registry := NewRegistry()
			service := NewMockService(fmt.Sprintf("service_%d", i))
			_ = registry.RegisterService(service)
			_, _ = registry.GetService(service.Name())
		}
	})

	b.Run("CommandParsing", func(b *testing.B) {
		service := NewExecutorService()
		_ = service.Initialize()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd := fmt.Sprintf(`\set[var_%d="value_%d"]`, i, i)
			_, _ = service.ParseCommand(cmd)
		}
	})
}

// Note: MockService is defined in registry_test.go
