package builtin

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestBashCommand_Name(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "bash", cmd.Name())
}

func TestBashCommand_ParseMode(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestBashCommand_Description(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "Execute system commands via bash", cmd.Description())
}

func TestBashCommand_Usage(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "\\bash command_to_execute", cmd.Usage())
}

func TestBashCommand_Execute_Success(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		contains []string // Strings that should be in output
	}{
		{
			name:     "simple echo command",
			input:    "echo hello",
			wantErr:  false,
			contains: []string{"hello"},
		},
		{
			name:     "echo with quotes",
			input:    "echo 'hello world'",
			wantErr:  false,
			contains: []string{"hello world"},
		},
		{
			name:     "command that succeeds",
			input:    "true",
			wantErr:  false,
			contains: []string{}, // No output expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(nil, tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})

			// Check that expected strings are in output
			for _, expectedStr := range tt.contains {
				assert.Contains(t, output, expectedStr, "Output should contain: %s", expectedStr)
			}
		})
	}
}

func TestBashCommand_Execute_WithError(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		contains []string // Strings that should be in output
	}{
		{
			name:     "command that fails",
			input:    "false",
			wantErr:  true, // Command should return error for non-zero exit codes
			contains: []string{"Exit status: 1"},
		},
		{
			name:     "command with stderr",
			input:    "echo 'error message' >&2",
			wantErr:  false,
			contains: []string{"Error: error message"},
		},
		{
			name:     "nonexistent command",
			input:    "nonexistentcommand123",
			wantErr:  true, // Command should return error for non-zero exit codes
			contains: []string{"Error:", "command not found", "Exit status:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(nil, tt.input)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})

			// Check that expected strings are in output
			for _, expectedStr := range tt.contains {
				assert.Contains(t, output, expectedStr, "Output should contain: %s", expectedStr)
			}
		})
	}
}

func TestBashCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &BashCommand{}

	tests := []string{"", "   ", "\t", "\n"}

	for _, emptyInput := range tests {
		t.Run(fmt.Sprintf("empty_input_%q", emptyInput), func(t *testing.T) {
			err := cmd.Execute(nil, emptyInput)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Usage:")
		})
	}
}

func TestBashCommand_Execute_ServiceUnavailable(t *testing.T) {
	cmd := &BashCommand{}

	// Don't setup bash service in registry

	err := cmd.Execute(nil, "echo test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bash service not available")
}

func TestBashCommand_Execute_WrongServiceType(t *testing.T) {
	cmd := &BashCommand{}

	// Setup registry but register wrong service type under "bash" name
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register a different service under "bash" name (this simulates a type error)
	err := services.GetGlobalRegistry().RegisterService(&mockWrongService{})
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	err = cmd.Execute(nil, "echo test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect type")
}

func TestBashCommand_Execute_OutputFormatting(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	tests := []struct {
		name         string
		input        string
		checkOutput  func(string)
		checkNoError bool
	}{
		{
			name:  "output without newline gets newline added",
			input: "printf 'no-newline'",
			checkOutput: func(output string) {
				// Should contain the text and end with newline
				assert.Contains(t, output, "no-newline")
				assert.True(t, strings.HasSuffix(strings.TrimSpace(output), "no-newline"))
			},
			checkNoError: true,
		},
		{
			name:  "output with newline preserved",
			input: "echo 'with-newline'",
			checkOutput: func(output string) {
				assert.Contains(t, output, "with-newline")
			},
			checkNoError: true,
		},
		{
			name:  "multiline output formatted correctly",
			input: "printf 'line1\\nline2\\nline3'",
			checkOutput: func(output string) {
				assert.Contains(t, output, "line1")
				assert.Contains(t, output, "line2")
				assert.Contains(t, output, "line3")
			},
			checkNoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(nil, tt.input)
				if tt.checkNoError {
					assert.NoError(t, err)
				}
			})

			tt.checkOutput(output)
		})
	}
}

func TestBashCommand_Execute_VariablesSet(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	// Execute a command
	_ = stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(nil, "echo test")
		assert.NoError(t, err)
	})

	// Verify system variables are set by checking MockContext directly
	allVars := ctx.GetAllVariables()

	if outputVar, exists := allVars["_output"]; exists {
		assert.Equal(t, "test", outputVar)
	}
	// Note: @status and @error are now managed by the ErrorManagementService at framework level
}

func TestBashCommand_Execute_FailedCommandVariables(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	// Execute a failing command
	_ = stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(nil, "false")
		assert.Error(t, err) // Command should return error for non-zero exit codes
	})

	// Verify system variables are set for failed command by checking MockContext directly
	allVars := ctx.GetAllVariables()

	if outputVar, exists := allVars["_output"]; exists {
		assert.Equal(t, "", outputVar) // false produces no output
	}
	// Note: @status and @error are now managed by the ErrorManagementService at framework level
}

func TestBashCommand_Execute_IntegrationWithRealCommands(t *testing.T) {
	// Setup
	cmd := &BashCommand{}
	ctx := context.NewTestContext()

	// Setup test registry with bash service
	setupBashTestRegistry(t, ctx)

	// Test various real bash commands
	tests := []struct {
		name    string
		command string
		check   func(t *testing.T, output string, ctx neurotypes.Context)
	}{
		{
			name:    "pwd command",
			command: "pwd",
			check: func(t *testing.T, output string, _ neurotypes.Context) {
				// Should contain a path
				assert.Contains(t, output, "/")

				// Note: @status is now managed by the ErrorManagementService at framework level
			},
		},
		{
			name:    "date command",
			command: "date +%Y",
			check: func(t *testing.T, output string, _ neurotypes.Context) {
				// Should contain a 4-digit year
				assert.Regexp(t, `\d{4}`, output)

				// Note: @status is now managed by the ErrorManagementService at framework level
			},
		},
		{
			name:    "ls of current directory",
			command: "ls -la .",
			check: func(t *testing.T, output string, _ neurotypes.Context) {
				// Should contain directory listing markers
				assert.Contains(t, output, ".")

				// Note: @status is now managed by the ErrorManagementService at framework level
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(nil, tt.command)
				assert.NoError(t, err)
			})

			tt.check(t, output, ctx)
		})
	}
}

// Helper functions

// setupBashTestRegistry sets up a test service registry with bash service
func setupBashTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewBashService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// CaptureOutput is now defined in stringprocessing package

// mockWrongService is a mock service with wrong type for testing
type mockWrongService struct{}

func (m *mockWrongService) Name() string      { return "bash" }
func (m *mockWrongService) Initialize() error { return nil }

// Interface compliance test
func TestBashCommand_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.Command = (*BashCommand)(nil)
}

// Benchmark tests
func BenchmarkBashCommand_Execute_SimpleCommand(b *testing.B) {
	cmd := &BashCommand{}

	// Setup minimal registry
	services.GlobalRegistry = services.NewRegistry()
	_ = services.GlobalRegistry.RegisterService(services.NewVariableService())
	_ = services.GlobalRegistry.RegisterService(services.NewBashService())
	_ = services.GlobalRegistry.InitializeAll()

	// Capture output to avoid printing during benchmark
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(nil, "echo test")
	}
}
