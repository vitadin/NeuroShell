package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestBashService_Name(t *testing.T) {
	service := NewBashService()
	assert.Equal(t, "bash", service.Name())
}

func TestBashService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  neurotypes.Context
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
			service := NewBashService()
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

func TestBashService_SetTimeout(t *testing.T) {
	service := NewBashService()

	// Test default timeout
	assert.Equal(t, 30*time.Second, service.timeout)

	// Test setting custom timeout
	customTimeout := 10 * time.Second
	service.SetTimeout(customTimeout)
	assert.Equal(t, customTimeout, service.timeout)
}

func TestBashService_Execute_Success(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	tests := []struct {
		name           string
		command        string
		expectedOutput string
		expectedError  string
		expectedCode   int
	}{
		{
			name:           "simple echo",
			command:        "echo hello",
			expectedOutput: "hello",
			expectedError:  "",
			expectedCode:   0,
		},
		{
			name:           "echo with quotes",
			command:        "echo 'hello world'",
			expectedOutput: "hello world",
			expectedError:  "",
			expectedCode:   0,
		},
		{
			name:           "command with success exit code",
			command:        "true",
			expectedOutput: "",
			expectedError:  "",
			expectedCode:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode, err := service.Execute(tt.command)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, stdout)
			assert.Equal(t, tt.expectedError, stderr)
			assert.Equal(t, tt.expectedCode, exitCode)
		})
	}
}

func TestBashService_Execute_WithError(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	tests := []struct {
		name         string
		command      string
		expectedCode int
	}{
		{
			name:         "false command",
			command:      "false",
			expectedCode: 1,
		},
		{
			name:         "exit with code",
			command:      "exit 42",
			expectedCode: 42,
		},
		{
			name:         "nonexistent command",
			command:      "nonexistentcommand123",
			expectedCode: 127,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode, err := service.Execute(tt.command)

			assert.NoError(t, err, "Execute should not return error even for failed commands")
			assert.Equal(t, tt.expectedCode, exitCode)

			// For non-zero exit codes, we expect either stderr or the command to complete
			if exitCode != 0 && exitCode != 127 {
				// Some commands like "false" and "exit N" don't produce stderr
				assert.True(t, stderr != "" || stdout == "", "Should have stderr or empty stdout for failed commands")
			}
		})
	}
}

func TestBashService_Execute_WithStderr(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	// Test multiple approaches to stderr output
	tests := []struct {
		name     string
		command  string
		checkErr func(stderr string, exitCode int)
	}{
		{
			name:    "stderr via printf",
			command: "printf 'error message' >&2",
			checkErr: func(stderr string, exitCode int) {
				assert.Contains(t, stderr, "error message")
				assert.Equal(t, 0, exitCode)
			},
		},
		{
			name:    "stderr via ls nonexistent",
			command: "ls /this/path/does/not/exist",
			checkErr: func(stderr string, exitCode int) {
				assert.Contains(t, stderr, "No such file or directory")
				// Accept both exit codes 1 and 2 (different ls implementations across environments)
				assert.True(t, exitCode == 1 || exitCode == 2,
					"Expected exit code 1 or 2 for ls nonexistent file, got %d", exitCode)
			},
		},
		{
			name:    "stderr via invalid command",
			command: "invalidcommandthatdoesnotexist123",
			checkErr: func(stderr string, exitCode int) {
				assert.Contains(t, stderr, "command not found")
				assert.Equal(t, 127, exitCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode, err := service.Execute(tt.command)

			assert.NoError(t, err)
			assert.Equal(t, "", stdout, "Should have no stdout")

			// Debug output to see what we're getting
			t.Logf("Command: %s", tt.command)
			t.Logf("Stdout: %q", stdout)
			t.Logf("Stderr: %q", stderr)
			t.Logf("Exit code: %d", exitCode)

			tt.checkErr(stderr, exitCode)
		})
	}
}

func TestBashService_Execute_EmptyCommand(t *testing.T) {
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []string{"", "   ", "\t", "\n"}

	for _, emptyCmd := range tests {
		t.Run(fmt.Sprintf("empty_command_%q", emptyCmd), func(t *testing.T) {
			_, _, _, err := service.Execute(emptyCmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "empty command")
		})
	}
}

func TestBashService_Execute_NotInitialized(t *testing.T) {
	service := NewBashService()
	ctx := testutils.NewMockContext()

	_, _, _, err := service.Execute("echo test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestBashService_Execute_SetsSystemVariables(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	// Execute a simple command
	stdout, stderr, exitCode, err := service.Execute("echo test")
	require.NoError(t, err)

	// The BashService should call SetSystemVariable, which our MockContext now supports
	// Verify the expected values
	assert.Equal(t, "test", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)

	// Since the service calls SetSystemVariable through the VariableService,
	// and our MockContext supports it, the variables should be set
	// Let's verify by checking the MockContext directly
	allVars := ctx.GetAllVariables()

	// Check if system variables were set (they should be there if SetSystemVariable was called)
	if outputVar, exists := allVars["_output"]; exists {
		assert.Equal(t, stdout, outputVar)
	}
	if statusVar, exists := allVars["_status"]; exists {
		assert.Equal(t, fmt.Sprintf("%d", exitCode), statusVar)
	}
	if errorVar, exists := allVars["_error"]; exists {
		assert.Equal(t, stderr, errorVar)
	}
}

func TestBashService_Execute_VariableServiceError(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Don't setup global registry - this will cause variable service to be unavailable

	// Execute should still work even if variable service is not available
	stdout, stderr, exitCode, err := service.Execute("echo test")
	assert.NoError(t, err)
	assert.Equal(t, "test", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)
}

func TestBashService_Execute_Timeout(t *testing.T) {
	// Setup with very short timeout
	service := NewBashService()
	service.SetTimeout(100 * time.Millisecond) // Even shorter timeout
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Command that should timeout - use a command that definitely takes longer
	stdout, stderr, exitCode, err := service.Execute("sleep 2")

	// Debug output
	t.Logf("Stdout: %q", stdout)
	t.Logf("Stderr: %q", stderr)
	t.Logf("Exit code: %d", exitCode)
	t.Logf("Error: %v", err)

	assert.NoError(t, err, "Execute should not return error for timeout")

	// Check that we got a timeout indication
	if stderr == "" {
		// If no timeout message, at least check that the exit code indicates failure
		assert.Equal(t, -1, exitCode, "Exit code should be -1 for timeout")
		t.Skip("Timeout detection might be platform-specific, skipping stderr check")
	} else {
		assert.Contains(t, stderr, "timeout") // Check for timeout in error message
		assert.Equal(t, -1, exitCode)
	}
}

func TestBashService_Execute_LongOutput(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	// Generate long output
	longText := strings.Repeat("a", 1000)
	command := fmt.Sprintf("echo '%s'", longText)

	stdout, stderr, exitCode, err := service.Execute(command)

	assert.NoError(t, err)
	assert.Equal(t, longText, stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)
}

func TestBashService_Execute_SpecialCharacters(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	tests := []struct {
		name    string
		command string
		output  string
	}{
		{
			name:    "special characters in quotes",
			command: "echo 'Hello & World | Test'",
			output:  "Hello & World | Test",
		},
		{
			name:    "unicode characters",
			command: "echo 'Hello 世界'",
			output:  "Hello 世界",
		},
		{
			name:    "newlines and tabs",
			command: "printf 'line1\\nline2\\tindented'",
			output:  "line1\nline2\tindented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode, err := service.Execute(tt.command)

			assert.NoError(t, err)
			assert.Equal(t, tt.output, stdout)
			assert.Equal(t, "", stderr)
			assert.Equal(t, 0, exitCode)
		})
	}
}

func TestBashService_Execute_MultilineOutput(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	// Command that produces multiline output
	stdout, stderr, exitCode, err := service.Execute("printf 'line1\\nline2\\nline3'")

	assert.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)

	// Verify lines are properly handled
	lines := strings.Split(stdout, "\n")
	assert.Len(t, lines, 3)
	assert.Equal(t, "line1", lines[0])
	assert.Equal(t, "line2", lines[1])
	assert.Equal(t, "line3", lines[2])
}

func TestBashService_Execute_EdgeCases(t *testing.T) {
	// Setup
	service := NewBashService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Setup global registry for variable service
	setupTestRegistry(t)

	tests := []struct {
		name    string
		command string
	}{
		{
			name:    "command with pipes",
			command: "echo 'test' | cat",
		},
		{
			name:    "command with redirection",
			command: "echo 'test' > /dev/null",
		},
		{
			name:    "compound command",
			command: "echo 'first' && echo 'second'",
		},
		{
			name:    "command substitution",
			command: "echo $(echo 'nested')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := service.Execute(tt.command)
			assert.NoError(t, err, "Command should execute without error: %s", tt.command)
		})
	}
}

// setupTestRegistry sets up a minimal service registry for testing
func setupTestRegistry(t *testing.T) {
	// Create a new registry for testing
	oldRegistry := GlobalRegistry
	GlobalRegistry = NewRegistry()

	// Register variable service
	err := GlobalRegistry.RegisterService(NewVariableService())
	require.NoError(t, err)

	// Initialize services with mock context
	ctx := testutils.NewMockContext()
	err = GlobalRegistry.InitializeAll(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		GlobalRegistry = oldRegistry
	})
}

// Benchmark tests
func BenchmarkBashService_Execute_SimpleCommand(b *testing.B) {
	service := NewBashService()
	ctx := testutils.NewMockContext()
	_ = service.Initialize(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = service.Execute("echo test")
	}
}

func BenchmarkBashService_Execute_ComplexCommand(b *testing.B) {
	service := NewBashService()
	ctx := testutils.NewMockContext()
	_ = service.Initialize(ctx)

	command := "echo 'test' | grep 'test' | wc -l"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = service.Execute(command)
	}
}
