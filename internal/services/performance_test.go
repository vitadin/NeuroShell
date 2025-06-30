package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"neuroshell/internal/parser"
	"neuroshell/internal/testutils"
)

// Comprehensive performance tests for service operations

// BenchmarkServiceInitialization tests service initialization performance
func BenchmarkServiceInitialization(b *testing.B) {
	ctx := testutils.NewMockContext()
	
	b.Run("VariableService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewVariableService()
			_ = service.Initialize(ctx)
		}
	})
	
	b.Run("ScriptService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewScriptService()
			_ = service.Initialize(ctx)
		}
	})
	
	b.Run("ExecutorService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewExecutorService()
			_ = service.Initialize(ctx)
		}
	})
	
	b.Run("InterpolationService", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service := NewInterpolationService()
			_ = service.Initialize(ctx)
		}
	})
}

// BenchmarkServiceRegistry_HighLoad tests registry under high load
func BenchmarkServiceRegistry_HighLoad(b *testing.B) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()
	
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
			_ = registry.InitializeAll(ctx)
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
	ctx := testutils.NewMockContextWithVars(vars)
	
	err := service.Initialize(ctx)
	require.NoError(b, err)
	
	b.Run("Get_Existing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			varName := fmt.Sprintf("var_%d", i%10000)
			_, _ = service.Get(varName, ctx)
		}
	})
	
	b.Run("Get_NonExisting", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = service.Get("nonexistent_var", ctx)
		}
	})
	
	b.Run("Set_NewVariables", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			varName := fmt.Sprintf("new_var_%d", i)
			varValue := fmt.Sprintf("new_value_%d", i)
			_ = service.Set(varName, varValue, ctx)
		}
	})
}

// BenchmarkExecutorService_CommandParsing tests command parsing performance
func BenchmarkExecutorService_CommandParsing(b *testing.B) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()
	
	err := service.Initialize(ctx)
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

// BenchmarkInterpolationService_ComplexStrings tests interpolation with complex patterns
func BenchmarkInterpolationService_ComplexStrings(b *testing.B) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()
	
	err := service.Initialize(ctx)
	require.NoError(b, err)
	
	testStrings := []string{
		"Simple: ${var}",
		"Multiple: ${var1} ${var2} ${var3}",
		"System: ${@user} ${@pwd} ${#session_id}",
		"Complex: ${prefix}${middle}${suffix}",
		"Very long string with ${many} ${different} ${variables} ${scattered} ${throughout} ${the} ${text} ${to} ${test} ${performance}",
		"Nested-like: ${var_${type}_${id}}",
		"Mixed: Regular text ${var1} more text ${var2} final text",
		"Unicode: 测试 ${chinese_var} 中文 ${unicode} 字符",
		"Special chars: ${var} !@#$%^&*()_+-=[]{}|;':\",./<>?",
		"Empty vars: ${} ${empty} ${blank}",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		str := testStrings[i%len(testStrings)]
		_, _ = service.InterpolateString(str, ctx)
	}
}

// BenchmarkInterpolationService_CommandStructures tests command interpolation
func BenchmarkInterpolationService_CommandStructures(b *testing.B) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()
	
	err := service.Initialize(ctx)
	require.NoError(b, err)
	
	commands := []*parser.Command{
		{
			Name:    "set",
			Message: "Simple ${var}",
			Options: map[string]string{"key": "${value}"},
		},
		{
			Name:    "send",
			Message: "Complex message with ${multiple} ${variables} ${here}",
			Options: map[string]string{
				"model":       "${ai_model}",
				"temperature": "${temp}",
				"system":      "${system_prompt}",
				"max_tokens":  "${tokens}",
			},
		},
		{
			Name:           "complex",
			Message:        "${greeting}, ${name}! Today is ${@date} and you're in ${@pwd}",
			BracketContent: "arg1=${val1}, arg2=${val2}, arg3=${val3}",
			Options: map[string]string{
				"option1": "${nested_${type}_value}",
				"option2": "${prefix}${middle}${suffix}",
				"option3": "Static value",
				"option4": "${@user}_${#session_id}",
			},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		_, _ = service.InterpolateCommand(cmd, ctx)
	}
}

// BenchmarkConcurrentServiceUsage tests concurrent service operations
func BenchmarkConcurrentServiceUsage(b *testing.B) {
	b.Run("VariableService_Concurrent", func(b *testing.B) {
		service := NewVariableService()
		ctx := testutils.NewMockContext()
		err := service.Initialize(ctx)
		require.NoError(b, err)
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				varName := fmt.Sprintf("concurrent_var_%d", i)
				varValue := fmt.Sprintf("concurrent_value_%d", i)
				_ = service.Set(varName, varValue, ctx)
				_, _ = service.Get(varName, ctx)
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
				ctx := testutils.NewMockContext()
				_ = service.Initialize(ctx)
				
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
			ss := NewScriptService()
			es := NewExecutorService()
			is := NewInterpolationService()
			
			// Prevent optimization
			_ = vs.Name()
			_ = ss.Name()
			_ = es.Name()
			_ = is.Name()
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
		ctx := testutils.NewMockContext()
		_ = service.Initialize(ctx)
		
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd := fmt.Sprintf(`\set[var_%d="value_%d"]`, i, i)
			_, _ = service.ParseCommand(cmd)
		}
	})
}

// Note: MockService is defined in registry_test.go