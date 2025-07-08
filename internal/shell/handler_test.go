package shell

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/abiosoft/ishell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/builtin"
	"neuroshell/internal/context"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// MockIShellContext provides a mock implementation of ishell.Context
type MockIShellContext struct {
	RawArgs       []string
	output        strings.Builder
	printfCalled  bool
	printlnCalled bool
}

func NewMockIShellContext(args []string) *MockIShellContext {
	return &MockIShellContext{
		RawArgs: args,
	}
}

func (m *MockIShellContext) Printf(format string, args ...interface{}) {
	m.printfCalled = true
	m.output.WriteString(fmt.Sprintf(format, args...))
}

func (m *MockIShellContext) Println(args ...interface{}) {
	m.printlnCalled = true
	m.output.WriteString(fmt.Sprintln(args...))
}

func (m *MockIShellContext) Print(args ...interface{}) {
	m.output.WriteString(fmt.Sprint(args...))
}

func (m *MockIShellContext) GetOutput() string {
	return m.output.String()
}

func (m *MockIShellContext) WasPrintfCalled() bool {
	return m.printfCalled
}

func (m *MockIShellContext) WasPrintlnCalled() bool {
	return m.printlnCalled
}

func (m *MockIShellContext) ClearOutput() {
	m.output.Reset()
	m.printfCalled = false
	m.printlnCalled = false
}

// Add methods to satisfy ishell.Context interface
func (m *MockIShellContext) Cmd() *ishell.Cmd                              { return nil }
func (m *MockIShellContext) Args() []string                                { return m.RawArgs }
func (m *MockIShellContext) Get(_ string) interface{}                      { return nil }
func (m *MockIShellContext) Set(_ string, _ interface{})                   {}
func (m *MockIShellContext) Del(_ string) interface{}                      { return nil }
func (m *MockIShellContext) Keys() []string                                { return nil }
func (m *MockIShellContext) Values() []interface{}                         { return nil }
func (m *MockIShellContext) Err(_ error)                                   {}
func (m *MockIShellContext) ReadLine() (string, error)                     { return "", nil }
func (m *MockIShellContext) ReadPassword(_ string) (string, error)         { return "", nil }
func (m *MockIShellContext) ReadMultiLines(_ string) (string, error)       { return "", nil }
func (m *MockIShellContext) ShowPrompt(_ bool)                             {}
func (m *MockIShellContext) ProgressBar() *ishell.ProgressBar              { return nil }
func (m *MockIShellContext) SetPrompt(_ string)                            {}
func (m *MockIShellContext) MultiChoice(_ []string, _ string) int          { return 0 }
func (m *MockIShellContext) Checklist(_ []string, _ string, _ []int) []int { return nil }
func (m *MockIShellContext) Stop()                                         {}

// Test setup and teardown helpers
func setupTestEnvironment(t *testing.T) func() {
	// Create a fresh test context
	testCtx := context.New()
	testCtx.SetTestMode(true)

	// Set as global context
	context.SetGlobalContext(testCtx)

	// Clear and reinitialize registries using thread-safe functions
	services.SetGlobalRegistry(services.NewRegistry())
	commands.SetGlobalRegistry(commands.NewRegistry())

	// Register builtin commands manually since we cleared the registry
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SetCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.GetCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.HelpCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.ExitCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendSyncCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendStreamCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.RunCommand{}))
	// Note: Bash command has been removed

	// Initialize services
	err := InitializeServices(true)
	require.NoError(t, err, "Failed to initialize test services")

	return func() {
		// Reset global context
		context.ResetGlobalContext()
	}
}

// testProcessInput is a wrapper for ProcessInput that accepts our mock
func testProcessInput(mockCtx *MockIShellContext) {
	if len(mockCtx.RawArgs) == 0 {
		return
	}

	rawInput := strings.Join(mockCtx.RawArgs, " ")
	rawInput = strings.TrimSpace(rawInput)

	// Skip comment lines (same logic as real ProcessInput)
	if strings.HasPrefix(rawInput, "%%") {
		return
	}

	// Parse input
	cmd := parser.ParseInput(rawInput)

	// Execute the command using our test executeCommand
	testExecuteCommand(mockCtx, cmd)
}

// testExecuteCommand is a wrapper for executeCommand that accepts our mock
func testExecuteCommand(mockCtx *MockIShellContext, cmd *parser.Command) {
	// Get interpolation service
	interpolationService, err := services.GlobalRegistry.GetService("interpolation")
	if err != nil {
		mockCtx.Printf("Error: interpolation service not available: %s\n", err.Error())
		return
	}

	is := interpolationService.(*services.InterpolationService)

	// Interpolate command components (new API without context)
	interpolatedCmd, err := is.InterpolateCommand(cmd)
	if err != nil {
		mockCtx.Printf("Error: interpolation failed: %s\n", err.Error())
		return
	}

	// Prepare input for execution
	input := interpolatedCmd.Message

	// Execute command with interpolated values
	err = commands.GlobalRegistry.Execute(interpolatedCmd.Name, interpolatedCmd.Options, input)
	if err != nil {
		mockCtx.Printf("Error: %s\n", err.Error())
		if cmd.Name != "help" {
			mockCtx.Println("Type \\help for available commands")
		}
	}
}

func TestProcessInput_EmptyArgs(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name    string
		rawArgs []string
	}{
		{
			name:    "nil args",
			rawArgs: nil,
		},
		{
			name:    "empty args",
			rawArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext(tt.rawArgs)

			// Should return early without any output
			testProcessInput(mockCtx)

			assert.Empty(t, mockCtx.GetOutput())
			assert.False(t, mockCtx.WasPrintfCalled())
			assert.False(t, mockCtx.WasPrintlnCalled())
		})
	}
}

func TestProcessInput_ValidCommands(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		rawArgs     []string
		expectError bool
		setup       func(*testing.T)
	}{
		{
			name:        "simple set command",
			rawArgs:     []string{"\\set[var=value]"},
			expectError: false,
			setup: func(_ *testing.T) {
				// No additional setup needed
			},
		},
		{
			name:        "get command",
			rawArgs:     []string{"\\get[var]"},
			expectError: false,
			setup: func(t *testing.T) {
				// Pre-set a variable
				err := GetGlobalContext().SetVariable("var", "test_value")
				require.NoError(t, err)
			},
		},
		{
			name:        "help command",
			rawArgs:     []string{"\\help"},
			expectError: false,
			setup:       func(_ *testing.T) {},
		},
		{
			name:        "command with message",
			rawArgs:     []string{"\\set[name=test]", "some", "message"},
			expectError: false,
			setup:       func(_ *testing.T) {},
		},
		{
			name:        "plain text auto-send",
			rawArgs:     []string{"hello", "world"},
			expectError: false, // send command is implemented and working
			setup:       func(_ *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			mockCtx := NewMockIShellContext(tt.rawArgs)
			testProcessInput(mockCtx)

			if tt.expectError {
				// Should have error output
				output := mockCtx.GetOutput()
				assert.Contains(t, output, "Error:", "Expected error output")
			} else {
				// Should execute without error
				output := mockCtx.GetOutput()
				assert.NotContains(t, output, "Error:", "Unexpected error output: %s", output)
			}
		})
	}
}

func TestProcessInput_CommandParsing(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name            string
		rawArgs         []string
		expectedCmd     string
		expectedMessage string
	}{
		{
			name:            "simple command",
			rawArgs:         []string{"\\help"},
			expectedCmd:     "help",
			expectedMessage: "",
		},
		{
			name:            "command with message",
			rawArgs:         []string{"\\set[var=value]", "hello", "world"},
			expectedCmd:     "set",
			expectedMessage: "hello world",
		},
		{
			name:            "command with spaces",
			rawArgs:         []string{"\\get", "var"},
			expectedCmd:     "get",
			expectedMessage: "var",
		},
		{
			name:            "auto-send plain text",
			rawArgs:         []string{"just", "plain", "text"},
			expectedCmd:     "send",
			expectedMessage: "just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Join raw args like ProcessInput does
			rawInput := strings.Join(tt.rawArgs, " ")
			rawInput = strings.TrimSpace(rawInput)

			cmd := parser.ParseInput(rawInput)

			assert.Equal(t, tt.expectedCmd, cmd.Name)
			assert.Equal(t, tt.expectedMessage, cmd.Message)
		})
	}
}

func TestProcessInput_WithVariableInterpolation(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up test variables
	err := GetGlobalContext().SetVariable("name", "Alice")
	require.NoError(t, err)
	err = GetGlobalContext().SetVariable("greeting", "Hello")
	require.NoError(t, err)

	tests := []struct {
		name    string
		rawArgs []string
		setup   func(*testing.T)
	}{
		{
			name:    "interpolated message",
			rawArgs: []string{"\\set[msg=${greeting} ${name}]"},
			setup:   func(_ *testing.T) {},
		},
		{
			name:    "system variable interpolation",
			rawArgs: []string{"\\set[user=${@user}]"},
			setup:   func(_ *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			mockCtx := NewMockIShellContext(tt.rawArgs)
			testProcessInput(mockCtx)

			// Should not have error output
			output := mockCtx.GetOutput()
			assert.NotContains(t, output, "Error:", "Unexpected error output: %s", output)
		})
	}
}

func TestExecuteCommand_BasicFlow(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		cmd         *parser.Command
		expectError bool
		setup       func(*testing.T)
	}{
		{
			name: "valid set command",
			cmd: &parser.Command{
				Name:    "set",
				Message: "",
				Options: map[string]string{"var": "value"},
			},
			expectError: false,
			setup:       func(_ *testing.T) {},
		},
		{
			name: "valid get command",
			cmd: &parser.Command{
				Name:    "get",
				Message: "",
				Options: map[string]string{"var": ""},
			},
			expectError: false,
			setup: func(t *testing.T) {
				err := GetGlobalContext().SetVariable("var", "test_value")
				require.NoError(t, err)
			},
		},
		{
			name: "invalid command",
			cmd: &parser.Command{
				Name:    "nonexistent",
				Message: "",
				Options: map[string]string{},
			},
			expectError: true,
			setup:       func(_ *testing.T) {},
		},
		{
			name: "help command",
			cmd: &parser.Command{
				Name:    "help",
				Message: "",
				Options: map[string]string{},
			},
			expectError: false,
			setup:       func(_ *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			mockCtx := NewMockIShellContext([]string{})
			testExecuteCommand(mockCtx, tt.cmd)

			output := mockCtx.GetOutput()
			if tt.expectError {
				assert.Contains(t, output, "Error:", "Expected error output")
				if tt.cmd.Name != "help" {
					assert.Contains(t, output, "\\help", "Should suggest help command")
				}
			} else {
				assert.NotContains(t, output, "Error:", "Unexpected error output: %s", output)
			}
		})
	}
}

func TestExecuteCommand_InterpolationFlow(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Set up test variables
	err := GetGlobalContext().SetVariable("name", "Alice")
	require.NoError(t, err)

	tests := []struct {
		name string
		cmd  *parser.Command
	}{
		{
			name: "command with variable interpolation",
			cmd: &parser.Command{
				Name:    "set",
				Message: "Setting variable",
				Options: map[string]string{"greeting": "Hello ${name}"},
			},
		},
		{
			name: "command with system variable",
			cmd: &parser.Command{
				Name:    "set",
				Message: "",
				Options: map[string]string{"current_user": "${@user}"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext([]string{})
			testExecuteCommand(mockCtx, tt.cmd)

			output := mockCtx.GetOutput()
			assert.NotContains(t, output, "Error:", "Unexpected error output: %s", output)
		})
	}
}

func TestExecuteCommand_BashCommandSpecialHandling(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name string
		cmd  *parser.Command
	}{
		{
			name: "non-bash command",
			cmd: &parser.Command{
				Name:           "set",
				Message:        "setting value",
				BracketContent: "var=value",
				Options:        map[string]string{"var": "value"},
				ParseMode:      neurotypes.ParseModeRaw,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext([]string{})
			testExecuteCommand(mockCtx, tt.cmd)

			output := mockCtx.GetOutput()

			// Commands should execute normally
			assert.NotContains(t, output, "Error:", "Unexpected error output: %s", output)
		})
	}
}

func TestExecuteCommand_ServiceErrors(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test case where interpolation service is not available
	t.Run("interpolation service unavailable", func(t *testing.T) {
		// Clear the services registry to simulate missing service
		services.GlobalRegistry = services.NewRegistry()

		cmd := &parser.Command{
			Name:    "set",
			Message: "",
			Options: map[string]string{"var": "value"},
		}

		mockCtx := NewMockIShellContext([]string{})
		testExecuteCommand(mockCtx, cmd)

		output := mockCtx.GetOutput()
		assert.Contains(t, output, "Error: interpolation service not available")
	})
}

func TestInitializeServices_Success(t *testing.T) {
	// Clear registry for clean test
	services.GlobalRegistry = services.NewRegistry()

	tests := []struct {
		name     string
		testMode bool
	}{
		{
			name:     "initialize in test mode",
			testMode: true,
		},
		{
			name:     "initialize in production mode",
			testMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear registry for each test
			services.GlobalRegistry = services.NewRegistry()

			// Reset global context for each test
			defer context.ResetGlobalContext()
			testCtx := context.New()
			context.SetGlobalContext(testCtx)

			// Skip production mode test if OPENAI_API_KEY is not set (CI/CD environment)
			if !tt.testMode {
				if os.Getenv("OPENAI_API_KEY") == "" {
					t.Skip("Skipping production mode test: OPENAI_API_KEY not set (likely CI/CD environment)")
				}
			}

			err := InitializeServices(tt.testMode)
			assert.NoError(t, err)

			// Verify test mode was set correctly
			assert.Equal(t, tt.testMode, GetGlobalContext().IsTestMode())

			// Verify all services were registered
			expectedServices := []string{
				"script",
				"variable",
				"executor",
				"interpolation",
			}

			for _, serviceName := range expectedServices {
				service, err := services.GlobalRegistry.GetService(serviceName)
				assert.NoError(t, err, "Service %s should be registered", serviceName)
				assert.NotNil(t, service, "Service %s should not be nil", serviceName)
			}
		})
	}
}

func TestInitializeServices_RegistrationFailure(t *testing.T) {
	// Test registration failure by registering the same service twice
	originalRegistry := services.GetGlobalRegistry()
	defer func() { services.SetGlobalRegistry(originalRegistry) }()

	// Create fresh registry
	services.SetGlobalRegistry(services.NewRegistry())

	// Register a service first
	scriptService := services.NewScriptService()
	err := services.GetGlobalRegistry().RegisterService(scriptService)
	require.NoError(t, err)

	// Try to register the same service again - should fail
	err = services.GetGlobalRegistry().RegisterService(scriptService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestInitializeServices_InitializationFailure(t *testing.T) {
	// Create a registry with a service that fails initialization
	services.GlobalRegistry = services.NewRegistry()

	// Register a failing service
	failingService := &FailingService{
		name:       "failing",
		shouldFail: true,
	}

	err := services.GlobalRegistry.RegisterService(failingService)
	require.NoError(t, err)

	// Reset global context for this test
	defer context.ResetGlobalContext()
	testCtx := context.New()
	context.SetGlobalContext(testCtx)

	// This should fail during InitializeAll
	err = services.GlobalRegistry.InitializeAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initialization failed")
}

func TestProcessInput_Integration(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name     string
		commands [][]string
		verify   func(*testing.T)
	}{
		{
			name: "set and get variable sequence",
			commands: [][]string{
				{"\\set[name=Alice]"},
				{"\\get[name]"},
			},
			verify: func(t *testing.T) {
				value, err := GetGlobalContext().GetVariable("name")
				assert.NoError(t, err)
				assert.Equal(t, "Alice", value)
			},
		},
		{
			name: "variable interpolation sequence",
			commands: [][]string{
				{"\\set[greeting=Hello]"},
				{"\\set[name=World]"},
				{"\\set[message=${greeting} ${name}!]"},
				{"\\get[message]"},
			},
			verify: func(t *testing.T) {
				value, err := GetGlobalContext().GetVariable("message")
				assert.NoError(t, err)
				assert.Equal(t, "Hello World!", value)
			},
		},
		{
			name: "system variable access",
			commands: [][]string{
				{"\\get[@user]"},
				{"\\get[#test_mode]"},
			},
			verify: func(_ *testing.T) {
				// Just verify no errors occurred
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, cmdArgs := range tt.commands {
				mockCtx := NewMockIShellContext(cmdArgs)
				testProcessInput(mockCtx)

				output := mockCtx.GetOutput()
				assert.NotContains(t, output, "Error:", "Command %d failed: %s", i+1, output)
			}

			tt.verify(t)
		})
	}
}

func TestProcessInput_ErrorScenarios(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name         string
		rawArgs      []string
		expectOutput string
	}{
		{
			name:         "unknown command",
			rawArgs:      []string{"\\unknown", "command"},
			expectOutput: "Error:",
		},
		{
			name:         "get non-existent variable",
			rawArgs:      []string{"\\get[nonexistent]"},
			expectOutput: "Error:",
		},
		{
			name:         "malformed command",
			rawArgs:      []string{"\\malformed["},
			expectOutput: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext(tt.rawArgs)
			testProcessInput(mockCtx)

			output := mockCtx.GetOutput()
			assert.Contains(t, output, tt.expectOutput)
		})
	}
}

func TestProcessInput_EdgeCases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name    string
		rawArgs []string
	}{
		{
			name:    "whitespace only args",
			rawArgs: []string{"   ", "\t", "\n"},
		},
		{
			name:    "empty strings in args",
			rawArgs: []string{"", "\\help", ""},
		},
		{
			name:    "very long command",
			rawArgs: []string{"\\set[var=" + strings.Repeat("a", 1000) + "]"},
		},
		{
			name:    "unicode characters",
			rawArgs: []string{"\\set[unicode=héllo wørld 🌍]"},
		},
		{
			name:    "special characters",
			rawArgs: []string{"\\set[special=!@#$%^&*()_+-=[]{}|;:,.<>?]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext(tt.rawArgs)

			// Should not panic
			assert.NotPanics(t, func() {
				testProcessInput(mockCtx)
			})
		})
	}
}

func TestProcessInput_CommentHandling(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name            string
		rawArgs         []string
		expectOutput    bool
		expectExecution bool
	}{
		{
			name:            "simple comment",
			rawArgs:         []string{"%%", "This", "is", "a", "comment"},
			expectOutput:    false,
			expectExecution: false,
		},
		{
			name:            "comment without space",
			rawArgs:         []string{"%%comment"},
			expectOutput:    false,
			expectExecution: false,
		},
		{
			name:            "comment with leading spaces",
			rawArgs:         []string{"  %%", "comment", "with", "spaces"},
			expectOutput:    false,
			expectExecution: false,
		},
		{
			name:            "comment with unicode",
			rawArgs:         []string{"%%", "unicode", "comment", "世界", "🌍"},
			expectOutput:    false,
			expectExecution: false,
		},
		{
			name:            "not a comment - %% in middle",
			rawArgs:         []string{"text", "with", "%%", "in", "middle"},
			expectOutput:    true,
			expectExecution: true,
		},
		{
			name:            "not a comment - starts with text",
			rawArgs:         []string{"text", "%%", "not", "comment"},
			expectOutput:    true,
			expectExecution: true,
		},
		{
			name:            "valid command after comment test",
			rawArgs:         []string{"\\help"},
			expectOutput:    true,
			expectExecution: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := NewMockIShellContext(tt.rawArgs)
			testProcessInput(mockCtx)

			output := mockCtx.GetOutput()

			if tt.expectOutput {
				// For non-comment inputs, we expect some output or execution
				if tt.expectExecution {
					// Should either have output or at least no error
					if len(output) > 0 {
						// If there's output, it shouldn't be an error for valid commands
						if strings.Contains(strings.Join(tt.rawArgs, " "), "\\help") {
							assert.NotContains(t, output, "Error:", "Valid command should not error: %s", output)
						}
					}
				}
			} else {
				// For comment inputs, we expect no output and no execution
				assert.Empty(t, output, "Comment should produce no output, got: %s", output)
				assert.False(t, mockCtx.WasPrintfCalled(), "Comment should not call Printf")
				assert.False(t, mockCtx.WasPrintlnCalled(), "Comment should not call Println")
			}
		})
	}
}

func TestProcessInput_CommentVsCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test that comments are truly skipped while commands execute
	tests := []struct {
		name     string
		inputs   [][]string
		expected map[string]string
	}{
		{
			name: "comment then command",
			inputs: [][]string{
				{"%%", "This", "is", "a", "comment"},
				{"\\set[var=value]"},
			},
			expected: map[string]string{"var": "value"},
		},
		{
			name: "command then comment",
			inputs: [][]string{
				{"\\set[test=hello]"},
				{"%%", "Comment", "after", "command"},
			},
			expected: map[string]string{"test": "hello"},
		},
		{
			name: "mixed comments and commands",
			inputs: [][]string{
				{"%%", "First", "comment"},
				{"\\set[first=one]"},
				{"%%", "Second", "comment"},
				{"\\set[second=two]"},
				{"%%", "Final", "comment"},
			},
			expected: map[string]string{"first": "one", "second": "two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute all inputs in sequence
			for _, input := range tt.inputs {
				mockCtx := NewMockIShellContext(input)
				testProcessInput(mockCtx)
			}

			// Verify expected variables were set
			for varName, expectedValue := range tt.expected {
				actualValue, err := GetGlobalContext().GetVariable(varName)
				assert.NoError(t, err, "Variable %s should exist", varName)
				assert.Equal(t, expectedValue, actualValue, "Variable %s should have correct value", varName)
			}
		})
	}
}

// Helper neurotypes for testing service failures

type FailingService struct {
	name       string
	shouldFail bool
}

func (f *FailingService) Name() string {
	return f.name
}

func (f *FailingService) Initialize() error {
	if f.shouldFail {
		return errors.New("initialization failed")
	}
	return nil
}

// Benchmark tests
func BenchmarkProcessInput_SimpleCommand(b *testing.B) {
	cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	args := []string{"\\help"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockCtx := NewMockIShellContext(args)
		testProcessInput(mockCtx)
	}
}

func BenchmarkProcessInput_CommandWithInterpolation(b *testing.B) {
	cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	// Set up variables
	if err := GetGlobalContext().SetVariable("name", "test"); err != nil {
		b.Fatal(err)
	}
	if err := GetGlobalContext().SetVariable("value", "benchmark"); err != nil {
		b.Fatal(err)
	}

	args := []string{"\\set[result=${name}_${value}]"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockCtx := NewMockIShellContext(args)
		testProcessInput(mockCtx)
	}
}

func BenchmarkExecuteCommand_DirectCall(b *testing.B) {
	cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	cmd := &parser.Command{
		Name:    "set",
		Options: map[string]string{"var": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockCtx := NewMockIShellContext([]string{})
		testExecuteCommand(mockCtx, cmd)
	}
}

func BenchmarkInitializeServices(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		services.GlobalRegistry = services.NewRegistry()
		testCtx := context.New()
		context.SetGlobalContext(testCtx)
		b.StartTimer()

		if err := InitializeServices(true); err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		context.ResetGlobalContext()
		b.StartTimer()
	}
}
