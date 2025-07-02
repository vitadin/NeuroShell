package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestEchoCommand_Name(t *testing.T) {
	cmd := &EchoCommand{}
	assert.Equal(t, "echo", cmd.Name())
}

func TestEchoCommand_ParseMode(t *testing.T) {
	cmd := &EchoCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEchoCommand_Description(t *testing.T) {
	cmd := &EchoCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "expand")
}

func TestEchoCommand_Usage(t *testing.T) {
	cmd := &EchoCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\echo")
}

func TestEchoCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "", // Should return usage error
		},
		{
			name:     "text with spaces",
			input:    "   Hello   World   ",
			expected: "   Hello   World   ",
		},
		{
			name:     "special characters",
			input:    "Hello @#$%^&*() World!",
			expected: "Hello @#$%^&*() World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == "" {
				// Test empty input should return error
				err := cmd.Execute(map[string]string{}, tt.input, ctx)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "Usage:")
				return
			}

			// Capture stdout
			output := captureOutput(func() {
				err := cmd.Execute(map[string]string{}, tt.input, ctx)
				assert.NoError(t, err)
			})

			// Remove trailing newline that echo adds
			output = strings.TrimSuffix(output, "\n")
			assert.Equal(t, tt.expected, output)

			// Check that result is stored in ${_output}
			value, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestEchoCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	// Set up test variables
	require.NoError(t, ctx.SetVariable("name", "Alice"))
	require.NoError(t, ctx.SetVariable("greeting", "Hello"))
	require.NoError(t, ctx.SetVariable("x", "1"))
	require.NoError(t, ctx.SetSystemVariable("_testuser", "testuser"))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single variable",
			input:    "Hello ${name}",
			expected: "Hello Alice",
		},
		{
			name:     "multiple variables",
			input:    "${greeting}, ${name}!",
			expected: "Hello, Alice!",
		},
		{
			name:     "system variable",
			input:    "User: ${_testuser}",
			expected: "User: testuser",
		},
		{
			name:     "mixed variables and text",
			input:    "${greeting} there, ${name}. Welcome to the system, ${_testuser}!",
			expected: "Hello there, Alice. Welcome to the system, testuser!",
		},
		{
			name:     "undefined variable",
			input:    "Hello ${undefined}",
			expected: "Hello ", // Undefined variables become empty strings
		},
		{
			name:     "nested variables",
			input:    "${${greeting}}",
			expected: "", // ${greeting} -> "Hello", then ${Hello} -> empty (doesn't exist)
		},
		{
			name:     "nested numeric variable",
			input:    "this is ${${name}}",
			expected: "this is ", // ${name} -> "Alice", then ${Alice} -> empty (doesn't exist)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := captureOutput(func() {
				err := cmd.Execute(map[string]string{}, tt.input, ctx)
				assert.NoError(t, err)
			})

			// Remove trailing newline that echo adds
			output = strings.TrimSuffix(output, "\n")
			assert.Equal(t, tt.expected, output)

			// Check that result is stored in ${_output}
			value, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestEchoCommand_Execute_ToOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	// Set up test variable
	require.NoError(t, ctx.SetVariable("name", "Bob"))

	tests := []struct {
		name       string
		args       map[string]string
		input      string
		expectedTo string
		expected   string
	}{
		{
			name:       "store in custom variable",
			args:       map[string]string{"to": "result"},
			input:      "Hello ${name}",
			expectedTo: "result",
			expected:   "Hello Bob",
		},
		{
			name:       "store in regular variable name",
			args:       map[string]string{"to": "custom"},
			input:      "Test message",
			expectedTo: "custom",
			expected:   "Test message",
		},
		{
			name:       "default to _output",
			args:       map[string]string{},
			input:      "Default storage",
			expectedTo: "_output",
			expected:   "Default storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := captureOutput(func() {
				err := cmd.Execute(tt.args, tt.input, ctx)
				assert.NoError(t, err)
			})

			// Remove trailing newline that echo adds
			output = strings.TrimSuffix(output, "\n")
			assert.Equal(t, tt.expected, output)

			// Check that result is stored in expected variable
			value, err := ctx.GetVariable(tt.expectedTo)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestEchoCommand_Execute_SilentOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expected    string
		shouldPrint bool
	}{
		{
			name:        "silent true - no output",
			args:        map[string]string{"silent": "true"},
			input:       "Hidden message",
			expected:    "Hidden message",
			shouldPrint: false,
		},
		{
			name:        "silent false - with output",
			args:        map[string]string{"silent": "false"},
			input:       "Visible message",
			expected:    "Visible message",
			shouldPrint: true,
		},
		{
			name:        "silent true with custom to variable",
			args:        map[string]string{"silent": "true", "to": "hidden"},
			input:       "Secret message",
			expected:    "Secret message",
			shouldPrint: false,
		},
		{
			name:        "default behavior (no silent) - with output",
			args:        map[string]string{},
			input:       "Normal message",
			expected:    "Normal message",
			shouldPrint: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := captureOutput(func() {
				err := cmd.Execute(tt.args, tt.input, ctx)
				assert.NoError(t, err)
			})

			if tt.shouldPrint {
				// Remove trailing newline that echo adds
				output = strings.TrimSuffix(output, "\n")
				assert.Equal(t, tt.expected, output)
			} else {
				assert.Empty(t, output, "Silent mode should produce no output")
			}

			// Check that result is always stored in variable
			targetVar := tt.args["to"]
			if targetVar == "" {
				targetVar = "_output"
			}
			value, err := ctx.GetVariable(targetVar)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestEchoCommand_Execute_InvalidSilentOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	args := map[string]string{"silent": "invalid"}
	input := "Test message"

	err := cmd.Execute(args, input, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for silent option")
	assert.Contains(t, err.Error(), "must be true or false")
}

func TestEchoCommand_Execute_CombinedOptions(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	// Set up test variable
	require.NoError(t, ctx.SetVariable("user", "Charlie"))

	// Test silent=true with custom to variable and variable interpolation
	args := map[string]string{
		"silent": "true",
		"to":     "greeting_message",
	}
	input := "Welcome, ${user}! Today is a great day."
	expected := "Welcome, Charlie! Today is a great day."

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(args, input, ctx)
		assert.NoError(t, err)
	})

	// Should have no console output due to silent=true
	assert.Empty(t, output)

	// Check that result is stored in custom variable
	value, err := ctx.GetVariable("greeting_message")
	assert.NoError(t, err)
	assert.Equal(t, expected, value)

	// Verify that ${_output} is NOT set (since we used custom variable)
	outputValue, err := ctx.GetVariable("_output")
	if err == nil {
		// If it exists, it should not be our message
		assert.NotEqual(t, expected, outputValue)
	}
}

func TestEchoCommand_Execute_NewlineHandling(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name              string
		input             string
		expectedOutput    string
		expectedStored    string
		shouldHaveNewline bool
	}{
		{
			name:              "text without newline",
			input:             "Hello World",
			expectedOutput:    "Hello World\n", // Echo adds newline
			expectedStored:    "Hello World",   // Stored without added newline
			shouldHaveNewline: true,
		},
		{
			name:              "text with trailing newline",
			input:             "Hello World\n",
			expectedOutput:    "Hello World\n", // No additional newline added
			expectedStored:    "Hello World\n", // Stored as-is
			shouldHaveNewline: false,           // No additional newline added
		},
		{
			name:              "text with multiple newlines",
			input:             "Line 1\nLine 2\n",
			expectedOutput:    "Line 1\nLine 2\n", // No additional newline
			expectedStored:    "Line 1\nLine 2\n", // Stored as-is
			shouldHaveNewline: false,              // No additional newline
		},
		{
			name:              "empty input with newline",
			input:             "\n",
			expectedOutput:    "\n",  // Preserved as-is
			expectedStored:    "\n",  // Stored as-is
			shouldHaveNewline: false, // No additional newline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty string test as it triggers usage error
			if tt.input == "" {
				return
			}

			// Capture stdout
			output := captureOutput(func() {
				err := cmd.Execute(map[string]string{}, tt.input, ctx)
				assert.NoError(t, err)
			})

			assert.Equal(t, tt.expectedOutput, output)

			// Check stored value
			value, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStored, value)
		})
	}
}

// setupEchoTestRegistry sets up a test environment with variable service
func setupEchoTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})
}

// captureOutput is defined in bash_test.go - using that implementation

// Interface compliance check
var _ neurotypes.Command = (*EchoCommand)(nil)
