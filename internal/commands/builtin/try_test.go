package builtin

import (
	"testing"

	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// setupTryTestRegistry initializes the registry with necessary services for try command tests.
func setupTryTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	oldCommandRegistry := commands.GetGlobalRegistry()
	commands.SetGlobalRegistry(commands.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewInterpolationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewBashService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	// Register commands needed for testing
	err = commands.GlobalRegistry.Register(&SetCommand{})
	require.NoError(t, err)

	err = commands.GlobalRegistry.Register(&GetCommand{})
	require.NoError(t, err)

	err = commands.GlobalRegistry.Register(&BashCommand{})
	require.NoError(t, err)

	err = commands.GlobalRegistry.Register(&TryCommand{})
	require.NoError(t, err)

	// Cleanup function to restore original registries
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		commands.SetGlobalRegistry(oldCommandRegistry)
		context.ResetGlobalContext()
	})
}

func TestTryCommand_Name(t *testing.T) {
	cmd := &TryCommand{}
	if cmd.Name() != "try" {
		t.Errorf("Expected command name 'try', got '%s'", cmd.Name())
	}
}

func TestTryCommand_ParseMode(t *testing.T) {
	cmd := &TryCommand{}
	if cmd.ParseMode() != neurotypes.ParseModeRaw {
		t.Errorf("Expected ParseModeRaw, got %v", cmd.ParseMode())
	}
}

func TestTryCommand_Description(t *testing.T) {
	cmd := &TryCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestTryCommand_Usage(t *testing.T) {
	cmd := &TryCommand{}
	usage := cmd.Usage()
	if usage == "" {
		t.Error("Usage should not be empty")
	}
}

func TestTryCommand_HelpInfo(t *testing.T) {
	cmd := &TryCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != "try" {
		t.Errorf("Expected command 'try', got '%s'", helpInfo.Command)
	}

	if len(helpInfo.Examples) == 0 {
		t.Error("Expected examples in help info")
	}

	if len(helpInfo.Notes) == 0 {
		t.Error("Expected notes in help info")
	}
}

func TestTryCommand_ExecuteEmptyInput(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	err := cmd.Execute(map[string]string{}, "")

	if err != nil {
		t.Errorf("Expected no error for empty input (try should never fail), got: %v", err)
	}
}

func TestTryCommand_ExecuteSuccessfulBashCommand(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	err := cmd.Execute(map[string]string{}, "\\bash echo \"test output\"")

	if err != nil {
		t.Errorf("Expected no error for successful command, got: %v", err)
	}

	// Check that status is set to success
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status != "0" {
		t.Errorf("Expected _status to be '0', got '%s'", status)
	}

	// Check that error is empty
	errorVar, err := ctx.GetVariable("_error")
	if err != nil {
		t.Errorf("Expected _error variable to be set: %v", err)
	}
	if errorVar != "" {
		t.Errorf("Expected _error to be empty, got '%s'", errorVar)
	}

	// Check that output contains the command output
	output, err := ctx.GetVariable("_output")
	if err != nil {
		t.Errorf("Expected _output variable to be set: %v", err)
	}
	if output != "test output" {
		t.Errorf("Expected _output to be 'test output', got '%s'", output)
	}
}

func TestTryCommand_ExecuteFailingBashCommand(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	err := cmd.Execute(map[string]string{}, "\\bash ls /nonexistent_directory_12345")

	// The try command should NOT return an error even if the inner command fails
	if err != nil {
		t.Errorf("Expected no error from try command even with failing inner command, got: %v", err)
	}

	// Check that status is set to failure (any non-zero value)
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status == "0" {
		t.Errorf("Expected _status to be non-zero (failure), got '%s'", status)
	}

	// Check that error contains error message
	errorVar, err := ctx.GetVariable("_error")
	if err != nil {
		t.Errorf("Expected _error variable to be set: %v", err)
	}
	if errorVar == "" {
		t.Error("Expected _error to contain error message, got empty string")
	}
}

func TestTryCommand_ExecuteSuccessfulSetCommand(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	err := cmd.Execute(map[string]string{}, "\\set[test_var=test_value]")

	if err != nil {
		t.Errorf("Expected no error for successful set command, got: %v", err)
	}

	// Check that status is set to success
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status != "0" {
		t.Errorf("Expected _status to be '0', got '%s'", status)
	}

	// Check that the variable was actually set
	testVar, err := ctx.GetVariable("test_var")
	if err != nil {
		t.Errorf("Expected test_var to be set: %v", err)
	}
	if testVar != "test_value" {
		t.Errorf("Expected test_var to be 'test_value', got '%s'", testVar)
	}
}

func TestTryCommand_ExecuteFailingSetCommand(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	// Try to set an invalid global variable that should be blocked by whitelist
	err := cmd.Execute(map[string]string{}, "\\set[_invalid_global=value]")

	// The try command should NOT return an error even if the inner command fails
	if err != nil {
		t.Errorf("Expected no error from try command even with failing inner command, got: %v", err)
	}

	// Check that status is set to failure (any non-zero value)
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status == "0" {
		t.Errorf("Expected _status to be non-zero (failure), got '%s'", status)
	}

	// Check that error contains error message
	errorVar, err := ctx.GetVariable("_error")
	if err != nil {
		t.Errorf("Expected _error variable to be set: %v", err)
	}
	if errorVar == "" {
		t.Error("Expected _error to contain error message, got empty string")
	}
}

func TestTryCommand_ExecuteUnknownCommand(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	err := cmd.Execute(map[string]string{}, "\\nonexistent_command_12345")

	// The try command should NOT return an error even if the inner command doesn't exist
	if err != nil {
		t.Errorf("Expected no error from try command even with unknown inner command, got: %v", err)
	}

	// Check that status is set to failure (any non-zero value)
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status == "0" {
		t.Errorf("Expected _status to be non-zero (failure), got '%s'", status)
	}

	// Check that error contains "unknown command" message
	errorVar, err := ctx.GetVariable("_error")
	if err != nil {
		t.Errorf("Expected _error variable to be set: %v", err)
	}
	if errorVar == "" {
		t.Error("Expected _error to contain 'unknown command' message, got empty string")
	}
}

func TestTryCommand_ExecuteInvalidSyntax(t *testing.T) {
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	cmd := &TryCommand{}
	// Test with invalid command syntax that should fail parsing
	err := cmd.Execute(map[string]string{}, "invalid command syntax \\\\")

	// The try command should NOT return an error even if parsing fails
	if err != nil {
		t.Errorf("Expected no error from try command even with invalid syntax, got: %v", err)
	}

	// Check that status is set to failure (any non-zero value)
	status, err := ctx.GetVariable("_status")
	if err != nil {
		t.Errorf("Expected _status variable to be set: %v", err)
	}
	if status == "0" {
		t.Errorf("Expected _status to be non-zero (failure), got '%s'", status)
	}

	// Check that error contains error message
	errorVar, err := ctx.GetVariable("_error")
	if err != nil {
		t.Errorf("Expected _error variable to be set: %v", err)
	}
	if errorVar == "" {
		t.Error("Expected _error to contain error message, got empty string")
	}
}

func TestTryCommand_ParseBracketContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single key-value pair",
			input: "key=value",
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "multiple key-value pairs",
			input: "key1=value1,key2=value2",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "quoted values",
			input: "key1=\"quoted value\",key2='single quoted'",
			expected: map[string]string{
				"key1": "quoted value",
				"key2": "single quoted",
			},
		},
		{
			name:  "flag without value",
			input: "flag1,key=value,flag2",
			expected: map[string]string{
				"flag1": "",
				"key":   "value",
				"flag2": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBracketContent(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key '%s', expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestTryCommand_Integration tests the try command in more realistic scenarios
func TestTryCommand_Integration(t *testing.T) {
	// Use real context for this integration test
	ctx := context.New()
	setupTryTestRegistry(t, ctx)

	// Test successful command sequence
	cmd := &TryCommand{}

	// Try a successful set command
	err := cmd.Execute(map[string]string{}, "\\set[var1=hello]")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the variable was set
	if val, err := ctx.GetVariable("var1"); err != nil || val != "hello" {
		t.Errorf("Expected var1='hello', got '%s' with error: %v", val, err)
	}

	// Verify success status
	if status, _ := ctx.GetVariable("_status"); status != "0" {
		t.Errorf("Expected _status='0', got '%s'", status)
	}

	// Try a failing command
	err = cmd.Execute(map[string]string{}, "\\unknown_command test")
	if err != nil {
		t.Errorf("Expected no error from try even with unknown command, got: %v", err)
	}

	// Verify failure status (any non-zero value)
	if status, _ := ctx.GetVariable("_status"); status == "0" {
		t.Errorf("Expected _status to be non-zero (failure), got '%s'", status)
	}

	// Verify error message is set
	if errorVar, _ := ctx.GetVariable("_error"); errorVar == "" {
		t.Error("Expected _error to be set with error message")
	}
}
