package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

func TestInterpolationService_Name(t *testing.T) {
	service := NewInterpolationService()
	assert.Equal(t, "interpolation", service.Name())
}

func TestInterpolationService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		{
			name: "successful initialization",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInterpolationService()
			err := service.Initialize()

			if tt.want != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.want.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestInterpolationService_InterpolateString(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Set up test variable
	_ = ctx.SetVariable("name", "world")

	// Test InterpolateString - should work with real context
	result, err := service.InterpolateString("Hello ${name}")

	assert.NoError(t, err)
	assert.Equal(t, "Hello world", result)
}

func TestInterpolationService_InterpolateString_NotInitialized(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.InterpolateString("Hello ${name}")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpolation service not initialized")
	assert.Empty(t, result)
}

func TestInterpolationService_InterpolateCommand(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Create a test command
	testCmd := &parser.Command{
		Name:           "set",
		Message:        "Hello ${name}",
		BracketContent: "var=${value}",
		Options: map[string]string{
			"var":   "${variable}",
			"value": "${data}",
		},
		ParseMode: neurotypes.ParseModeKeyValue,
	}

	// Set up test variables
	_ = ctx.SetVariable("name", "world")
	_ = ctx.SetVariable("variable", "test_var")
	_ = ctx.SetVariable("data", "test_data")

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Test InterpolateCommand - should work with real context
	result, err := service.InterpolateCommand(testCmd)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hello world", result.Message)
	assert.Equal(t, "test_var", result.Options["var"])
	assert.Equal(t, "test_data", result.Options["value"])
}

func TestInterpolationService_InterpolateCommand_NotInitialized(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	testCmd := &parser.Command{
		Name:    "set",
		Message: "Hello ${name}",
	}

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.InterpolateCommand(testCmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpolation service not initialized")
	assert.Nil(t, result)
}

// Test command structure preservation
func TestInterpolationService_CommandStructurePreservation(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	testCases := []struct {
		name string
		cmd  *parser.Command
	}{
		{
			name: "basic command",
			cmd: &parser.Command{
				Name:           "set",
				Message:        "test message",
				BracketContent: "var=value",
				Options: map[string]string{
					"var": "value",
				},
				ParseMode: neurotypes.ParseModeKeyValue,
			},
		},
		{
			name: "command with interpolation patterns",
			cmd: &parser.Command{
				Name:           "send",
				Message:        "Hello ${name}, today is ${@date}",
				BracketContent: "model=${model}, temp=${temperature}",
				Options: map[string]string{
					"model":       "${ai_model}",
					"temperature": "${temp_setting}",
					"system":      "${@user}",
				},
				ParseMode: neurotypes.ParseModeKeyValue,
			},
		},
		{
			name: "empty command",
			cmd: &parser.Command{
				Name:           "test",
				Message:        "",
				BracketContent: "",
				Options:        make(map[string]string),
				ParseMode:      neurotypes.ParseModeRaw,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			// Set up test variables for interpolation
			_ = ctx.SetVariable("name", "world")
			_ = ctx.SetVariable("model", "gpt-4")
			_ = ctx.SetVariable("temperature", "0.7")
			_ = ctx.SetVariable("ai_model", "claude-3")
			_ = ctx.SetVariable("temp_setting", "0.5")

			// Test command interpolation - should work with real context
			result, err := service.InterpolateCommand(tc.cmd)

			// Expect success with real context
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify original command is unchanged
			expectedName := "test"
			switch tc.name {
			case "basic command":
				expectedName = "set"
			case "command with interpolation patterns":
				expectedName = "send"
			}
			assert.Equal(t, expectedName, tc.cmd.Name)

			// Verify interpolated result has correct structure
			assert.Equal(t, expectedName, result.Name)
			assert.Equal(t, tc.cmd.ParseMode, result.ParseMode)
		})
	}
}

// Test string interpolation patterns
func TestInterpolationService_StringInterpolationPatterns(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"no interpolation", "plain text", "plain text"},
		{"single variable", "Hello ${name}", "Hello world"},
		{"multiple variables", "${greeting}, ${name}!", "Hello, world!"},
		{"system variables", "User: ${@user}, PWD: ${@pwd}", ""}, // Will be set dynamically
		{"nested patterns", "${prefix}${middle}${suffix}", "startmiddleend"},
		{"empty variable", "Value: ${empty}", "Value: "},
		{"mixed content", "Start ${var1} middle ${var2} end", "Start value1 middle value2 end"},
		{"special characters", "${var} with symbols !@#$%", "test with symbols !@#$%"},
		{"unicode content", "测试 ${unicode_var} 中文", "测试 unicode_value 中文"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			// Set up test variables for interpolation
			_ = ctx.SetVariable("name", "world")
			_ = ctx.SetVariable("greeting", "Hello")
			_ = ctx.SetVariable("prefix", "start")
			_ = ctx.SetVariable("middle", "middle")
			_ = ctx.SetVariable("suffix", "end")
			_ = ctx.SetVariable("empty", "")
			_ = ctx.SetVariable("var1", "value1")
			_ = ctx.SetVariable("var2", "value2")
			_ = ctx.SetVariable("var", "test")
			_ = ctx.SetVariable("unicode_var", "unicode_value")

			// Test string interpolation - should work with real context
			result, err := service.InterpolateString(tc.input)

			assert.NoError(t, err)

			// Handle dynamic cases
			if tc.name == "system variables" {
				// Just verify it contains the expected patterns
				assert.Contains(t, result, "User: ")
				assert.Contains(t, result, "PWD: ")
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkInterpolationService_InterpolateString_Simple(b *testing.B) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(b, err)

	input := "Hello ${name}"

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateString(input)
	}
}

func BenchmarkInterpolationService_InterpolateString_Complex(b *testing.B) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(b, err)

	input := "Complex interpolation: ${var1} ${var2} ${@user} ${@pwd} ${#session_id} ${nested_${var3}}"

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateString(input)
	}
}

func BenchmarkInterpolationService_InterpolateCommand(b *testing.B) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(b, err)

	cmd := &parser.Command{
		Name:           "send",
		Message:        "Hello ${name}, today is ${@date}",
		BracketContent: "model=${model}, temp=${temperature}",
		Options: map[string]string{
			"model":       "${ai_model}",
			"temperature": "${temp_setting}",
			"system":      "${@user}",
		},
		ParseMode: neurotypes.ParseModeKeyValue,
	}

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateCommand(cmd)
	}
}

// Test edge cases
func TestInterpolationService_EdgeCases(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	edgeCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"only variables", "${var1}${var2}${var3}", "value1value2value3"},
		{"malformed variables", "${unclosed ${nested} ${}", ""}, // Will be set dynamically
		{"dollar without braces", "Price: $100", "Price: $100"},
		{"multiple dollars", "$$${var}$$", "$$test$$"},
		{"very long string", "This is a very long string with ${var1} and ${var2} and many other words to test performance with large strings", "This is a very long string with value1 and value2 and many other words to test performance with large strings"},
		{"special characters", "${var} !@#$%^&*()_+-=[]{}|;':\",./<>?", "test !@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"newlines and tabs", "Line1\n${var}\tTab", "Line1\ntest\tTab"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			// Set up test variables for interpolation
			_ = ctx.SetVariable("var1", "value1")
			_ = ctx.SetVariable("var2", "value2")
			_ = ctx.SetVariable("var3", "value3")
			_ = ctx.SetVariable("var", "test")

			// Should handle edge cases without panicking
			result, err := service.InterpolateString(tc.input)

			// Expect success with real context
			assert.NoError(t, err)

			// Handle dynamic cases
			if tc.name == "malformed variables" {
				// Just verify it doesn't panic and produces some output
				assert.NotEmpty(t, result)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// Test concurrent access
func TestInterpolationService_ConcurrentAccess(t *testing.T) {
	// Test concurrent initialization and interpolation with separate service instances
	// Set up shared global context to avoid race conditions
	sharedCtx := context.NewTestContext()
	context.SetGlobalContext(sharedCtx)
	defer context.ResetGlobalContext()

	// Set up test variables for interpolation
	_ = sharedCtx.SetVariable("name", "world")
	_ = sharedCtx.SetVariable("greeting", "Hello")
	_ = sharedCtx.SetVariable("var1", "value1")
	_ = sharedCtx.SetVariable("var2", "value2")

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(_ int) {
			// Each goroutine gets its own service instance to avoid race conditions
			service := NewInterpolationService()
			err := service.Initialize()
			assert.NoError(t, err)

			// Try various interpolations concurrently
			testStrings := []string{
				"Hello ${name}",
				"${greeting}, ${name}!",
				"User: ${@user}",
				"Complex: ${var1} ${var2} ${@pwd}",
			}

			for _, str := range testStrings {
				_, err := service.InterpolateString(str)
				// Expect success with real context, shouldn't panic
				assert.NoError(t, err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test initialization state
func TestInterpolationService_InitializationState(t *testing.T) {
	service := NewInterpolationService()

	// Should not be initialized initially
	assert.False(t, service.initialized)

	// Setup global context for testing (even though service is not initialized)
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Set up test variables for interpolation
	_ = ctx.SetVariable("name", "world")

	// Test operations before initialization
	_, err := service.InterpolateString("Hello ${name}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = service.InterpolateCommand(&parser.Command{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Test operations after initialization (should succeed with real context)
	result, err := service.InterpolateString("Hello ${name}")
	assert.NoError(t, err)
	assert.Equal(t, "Hello world", result)

	// Re-initialization should work
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

// Test nil handling
func TestInterpolationService_NilHandling(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	// Test nil command
	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.InterpolateCommand(nil)
	// Should handle nil gracefully - expect error for nil command
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command cannot be nil")
	assert.Nil(t, result)

	// Test with reset global context (singleton creates new context automatically)
	context.ResetGlobalContext()
	result2, err := service.InterpolateString("test")
	// With singleton pattern, GetGlobalContext() creates a new context automatically
	// So the operation should succeed
	assert.NoError(t, err)
	assert.Equal(t, "test", result2)
}

// Test command option handling
func TestInterpolationService_CommandOptionHandling(t *testing.T) {
	service := NewInterpolationService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	// Set the context as global context for the service to use
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	testCases := []struct {
		name string
		cmd  *parser.Command
	}{
		{
			name: "nil options map",
			cmd: &parser.Command{
				Name:    "test",
				Options: nil,
			},
		},
		{
			name: "empty options map",
			cmd: &parser.Command{
				Name:    "test",
				Options: make(map[string]string),
			},
		},
		{
			name: "single option",
			cmd: &parser.Command{
				Name: "test",
				Options: map[string]string{
					"key": "value",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should handle different option scenarios
			result, err := service.InterpolateCommand(tc.cmd)

			// Expect success with real context
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.cmd.Name, result.Name)
		})
	}
}
