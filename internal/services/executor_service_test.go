package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
)

func TestExecutorService_Name(t *testing.T) {
	service := NewExecutorService()
	assert.Equal(t, "executor", service.Name())
}

func TestExecutorService_Initialize(t *testing.T) {
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
			service := NewExecutorService()
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

func TestExecutorService_ParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "parse simple command",
			input:     `\set[name="value"]`,
			wantError: false,
		},
		{
			name:      "parse get command",
			input:     `\get[name]`,
			wantError: false,
		},
		{
			name:      "parse empty command",
			input:     "",
			wantError: false, // Parser should handle empty input
		},
		{
			name:      "parse invalid command",
			input:     "invalid command",
			wantError: false, // Parser should handle invalid input gracefully
		},
		{
			name:      "parse command with complex args",
			input:     `\send[model="gpt-4", temperature=0.7] Hello world`,
			wantError: false,
		},
	}

	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := service.ParseCommand(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, cmd)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cmd)
			}
		})
	}
}

func TestExecutorService_ParseCommand_NotInitialized(t *testing.T) {
	service := NewExecutorService()

	cmd, err := service.ParseCommand(`\set[name="value"]`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
	assert.Nil(t, cmd)
}

func TestExecutorService_GetNextCommand(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test GetNextCommand - will fail since MockContext is not NeuroContext
	cmd, err := service.GetNextCommand(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Nil(t, cmd)
}

func TestExecutorService_GetNextCommand_NotInitialized(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	cmd, err := service.GetNextCommand(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
	assert.Nil(t, cmd)
}

func TestExecutorService_GetQueueStatus(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test GetQueueStatus - will fail since MockContext is not NeuroContext
	status, err := service.GetQueueStatus(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Nil(t, status)
}

func TestExecutorService_GetQueueStatus_NotInitialized(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	status, err := service.GetQueueStatus(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
	assert.Nil(t, status)
}

func TestExecutorService_MarkCommandExecuted(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test MarkCommandExecuted - will fail since MockContext is not NeuroContext
	err = service.MarkCommandExecuted(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
}

func TestExecutorService_MarkCommandExecuted_NotInitialized(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.MarkCommandExecuted(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
}

func TestExecutorService_MarkExecutionError(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	testError := errors.New("test execution error")
	testCommand := `\set[name="value"]`

	// Test MarkExecutionError - will fail since MockContext is not NeuroContext
	err = service.MarkExecutionError(ctx, testError, testCommand)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
}

func TestExecutorService_MarkExecutionError_NotInitialized(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	testError := errors.New("test execution error")
	testCommand := `\set[name="value"]`

	err := service.MarkExecutionError(ctx, testError, testCommand)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
}

func TestExecutorService_MarkExecutionComplete(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test MarkExecutionComplete - will fail since MockContext is not NeuroContext
	err = service.MarkExecutionComplete(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
}

func TestExecutorService_MarkExecutionComplete_NotInitialized(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.MarkExecutionComplete(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor service not initialized")
}

// Test command parsing with various inputs
func TestExecutorService_ParseCommand_Comprehensive(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "basic set command",
			input:       `\set[name="value"]`,
			description: "Simple variable assignment",
		},
		{
			name:        "get command",
			input:       `\get[name]`,
			description: "Variable retrieval",
		},
		{
			name:        "command with message",
			input:       `\send Hello world`,
			description: "Command with message content",
		},
		{
			name:        "command with multiple args",
			input:       `\send[model="gpt-4", temp=0.7] Hello world`,
			description: "Command with multiple parameters",
		},
		{
			name:        "system variable",
			input:       `\get[@user]`,
			description: "System variable access",
		},
		{
			name:        "empty command",
			input:       "",
			description: "Empty input handling",
		},
		{
			name:        "plain text",
			input:       "Just plain text",
			description: "Non-command text",
		},
		{
			name:        "command with special characters",
			input:       `\set[special="!@#$%^&*()"]`,
			description: "Special characters in values",
		},
	}

	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := service.ParseCommand(tc.input)

			// Parser should handle all inputs gracefully
			assert.NoError(t, err, "ParseCommand should not error for: %s", tc.description)
			assert.NotNil(t, cmd, "ParseCommand should return command object for: %s", tc.description)
		})
	}
}

// Benchmark tests
func BenchmarkExecutorService_ParseCommand_Simple(b *testing.B) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	input := `\set[name="value"]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ParseCommand(input)
	}
}

func BenchmarkExecutorService_ParseCommand_Complex(b *testing.B) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	input := `\send[model="gpt-4", temperature=0.7, max_tokens=1000] This is a complex message with multiple parameters and a longer text content that needs to be parsed efficiently`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ParseCommand(input)
	}
}

// Test parsing edge cases
func TestExecutorService_ParseCommand_EdgeCases(t *testing.T) {
	service := NewExecutorService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	edgeCases := []struct {
		name  string
		input string
	}{
		{"very long command", `\set[very_long_variable_name_that_exceeds_normal_length="very long value that contains multiple words and special characters !@#$%^&*()_+-={}[]|\\:";'<>?,./"]`},
		{"unicode characters", `\set[unicode="测试中文字符"]`},
		{"nested quotes", `\set[nested="value with \"nested\" quotes"]`},
		{"multiline-like", `\set[multiline="line1\nline2\nline3"]`},
		{"empty brackets", `\command[]`},
		{"malformed brackets", `\command[name=`},
		{"only command", `\command`},
		{"whitespace heavy", `   \set  [  name  =  "value"  ]   `},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parser should handle edge cases gracefully without panicking
			cmd, err := service.ParseCommand(tc.input)
			assert.NoError(t, err, "Should handle edge case: %s", tc.name)
			assert.NotNil(t, cmd, "Should return command object for edge case: %s", tc.name)
		})
	}
}

// Test concurrent access
func TestExecutorService_ConcurrentAccess(t *testing.T) {
	// Test concurrent initialization and parsing with separate service instances
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(_ int) {
			// Each goroutine gets its own service instance to avoid race conditions
			service := NewExecutorService()
			ctx := testutils.NewMockContext()
			err := service.Initialize(ctx)
			assert.NoError(t, err)

			// Parse different commands concurrently
			commands := []string{
				`\set[var1="value1"]`,
				`\get[var1]`,
				`\send Hello world`,
				`\command[arg="value"]`,
			}

			for _, cmd := range commands {
				parsed, err := service.ParseCommand(cmd)
				assert.NoError(t, err)
				assert.NotNil(t, parsed)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Test initialization state consistency
func TestExecutorService_InitializationState(t *testing.T) {
	service := NewExecutorService()

	// Should not be initialized initially
	assert.False(t, service.initialized)

	// Test operations before initialization
	_, err := service.ParseCommand(`\set[name="value"]`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize
	ctx := testutils.NewMockContext()
	err = service.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Test operations after initialization
	_, err = service.ParseCommand(`\set[name="value"]`)
	assert.NoError(t, err)

	// Re-initialization should work
	err = service.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}
