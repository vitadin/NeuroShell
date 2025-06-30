package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/parser"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/types"
)

func TestInterpolationService_Name(t *testing.T) {
	service := NewInterpolationService()
	assert.Equal(t, "interpolation", service.Name())
}

func TestInterpolationService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  *testutils.MockContext
		want error
	}{
		{
			name: "successful initialization",
			ctx:  testutils.NewMockContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewInterpolationService()
			err := service.Initialize(tt.ctx)

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
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test InterpolateString - will fail since MockContext is not NeuroContext
	result, err := service.InterpolateString("Hello ${name}", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Empty(t, result)
}

func TestInterpolationService_InterpolateString_NotInitialized(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	result, err := service.InterpolateString("Hello ${name}", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpolation service not initialized")
	assert.Empty(t, result)
}

func TestInterpolationService_InterpolateCommand(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
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
		ParseMode: types.ParseModeKeyValue,
	}

	// Test InterpolateCommand - will fail since MockContext is not NeuroContext
	result, err := service.InterpolateCommand(testCmd, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Nil(t, result)
}

func TestInterpolationService_InterpolateCommand_NotInitialized(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	testCmd := &parser.Command{
		Name:    "set",
		Message: "Hello ${name}",
	}

	result, err := service.InterpolateCommand(testCmd, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpolation service not initialized")
	assert.Nil(t, result)
}

// Test command structure preservation
func TestInterpolationService_CommandStructurePreservation(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
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
				ParseMode: types.ParseModeKeyValue,
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
				ParseMode: types.ParseModeKeyValue,
			},
		},
		{
			name: "empty command",
			cmd: &parser.Command{
				Name:           "test",
				Message:        "",
				BracketContent: "",
				Options:        make(map[string]string),
				ParseMode:      types.ParseModeRaw,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This will fail due to MockContext, but we test the structure
			result, err := service.InterpolateCommand(tc.cmd, ctx)

			// Expect error due to MockContext
			assert.Error(t, err)
			assert.Nil(t, result)

			// Verify original command is unchanged
			expectedName := "test"
			if tc.name == "basic command" {
				expectedName = "set"
			} else if tc.name == "command with interpolation patterns" {
				expectedName = "send"
			}
			assert.Equal(t, expectedName, tc.cmd.Name)
		})
	}
}

// Test string interpolation patterns
func TestInterpolationService_StringInterpolationPatterns(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		input string
	}{
		{"no interpolation", "plain text"},
		{"single variable", "Hello ${name}"},
		{"multiple variables", "${greeting}, ${name}!"},
		{"system variables", "User: ${@user}, PWD: ${@pwd}"},
		{"nested patterns", "${prefix}${middle}${suffix}"},
		{"empty variable", "Value: ${empty}"},
		{"mixed content", "Start ${var1} middle ${var2} end"},
		{"special characters", "${var} with symbols !@#$%"},
		{"unicode content", "测试 ${unicode_var} 中文"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Will fail due to MockContext, but tests service behavior
			result, err := service.InterpolateString(tc.input, ctx)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "context is not a NeuroContext")
			assert.Empty(t, result)
		})
	}
}

// Benchmark tests
func BenchmarkInterpolationService_InterpolateString_Simple(b *testing.B) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	input := "Hello ${name}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateString(input, ctx)
	}
}

func BenchmarkInterpolationService_InterpolateString_Complex(b *testing.B) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	input := "Complex interpolation: ${var1} ${var2} ${@user} ${@pwd} ${#session_id} ${nested_${var3}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateString(input, ctx)
	}
}

func BenchmarkInterpolationService_InterpolateCommand(b *testing.B) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
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
		ParseMode: types.ParseModeKeyValue,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Will error due to MockContext, but measures service overhead
		_, _ = service.InterpolateCommand(cmd, ctx)
	}
}

// Test edge cases
func TestInterpolationService_EdgeCases(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	edgeCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only variables", "${var1}${var2}${var3}"},
		{"malformed variables", "${unclosed ${nested} ${}"},
		{"dollar without braces", "Price: $100"},
		{"multiple dollars", "$$${var}$$"},
		{"very long string", "This is a very long string with ${var1} and ${var2} and many other words to test performance with large strings"},
		{"special characters", "${var} !@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"newlines and tabs", "Line1\n${var}\tTab"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should handle edge cases without panicking
			result, err := service.InterpolateString(tc.input, ctx)

			// Expect error due to MockContext
			assert.Error(t, err)
			assert.Empty(t, result)
		})
	}
}

// Test concurrent access
func TestInterpolationService_ConcurrentAccess(t *testing.T) {
	// Test concurrent initialization and interpolation with separate service instances
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(_ int) {
			// Each goroutine gets its own service instance to avoid race conditions
			service := NewInterpolationService()
			ctx := testutils.NewMockContext()
			err := service.Initialize(ctx)
			assert.NoError(t, err)

			// Try various interpolations concurrently
			testStrings := []string{
				"Hello ${name}",
				"${greeting}, ${name}!",
				"User: ${@user}",
				"Complex: ${var1} ${var2} ${@pwd}",
			}

			for _, str := range testStrings {
				_, err := service.InterpolateString(str, ctx)
				// Expect error due to MockContext, but shouldn't panic
				assert.Error(t, err)
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

	// Test operations before initialization
	_, err := service.InterpolateString("Hello ${name}", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = service.InterpolateCommand(&parser.Command{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize
	ctx := testutils.NewMockContext()
	err = service.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Test operations after initialization (will still error due to MockContext)
	_, err = service.InterpolateString("Hello ${name}", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")

	// Re-initialization should work
	err = service.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

// Test nil handling
func TestInterpolationService_NilHandling(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test nil command
	result, err := service.InterpolateCommand(nil, ctx)
	// Should handle nil gracefully but will error due to MockContext first
	assert.Error(t, err)
	assert.Nil(t, result)

	// Test nil context
	_, err = service.InterpolateString("test", nil)
	// Should error due to nil context type assertion
	assert.Error(t, err)
}

// Test command option handling
func TestInterpolationService_CommandOptionHandling(t *testing.T) {
	service := NewInterpolationService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

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
			result, err := service.InterpolateCommand(tc.cmd, ctx)

			// Expect error due to MockContext
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
