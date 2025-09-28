package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
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
	assert.Contains(t, strings.ToLower(desc), "output")
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
			expected: "", // Should succeed and output empty
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
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(map[string]string{}, tt.input)
				assert.NoError(t, err) // Echo never returns errors
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
			input:      "Hello Bob",
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
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
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
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
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

	// Capture stdout
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(args, input)
		assert.NoError(t, err) // Echo never returns errors, uses default value
	})

	// Should use default silent=false and output the message
	output = strings.TrimSuffix(output, "\n")
	assert.Equal(t, "Test message", output)
}

func TestEchoCommand_Execute_CombinedOptions(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	// Test silent=true with custom to variable (no interpolation - that's handled by state machine)
	args := map[string]string{
		"silent": "true",
		"to":     "greeting_message",
	}
	input := "Welcome, user! Today is a great day."
	expected := "Welcome, user! Today is a great day."

	// Capture stdout
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(args, input)
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
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(map[string]string{}, tt.input)
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

func TestEchoCommand_Execute_RawOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		expectedStored string
	}{
		{
			name:           "raw true - newline escape sequence",
			args:           map[string]string{"raw": "true"},
			input:          "Line 1\\nLine 2",
			expectedOutput: "Line 1\\nLine 2\n", // Literal backslash-n, plus added newline
			expectedStored: "Line 1\\nLine 2",   // Stored without added newline
		},
		{
			name:           "raw false - newline escape sequence",
			args:           map[string]string{"raw": "false"},
			input:          "Line 1\\nLine 2",
			expectedOutput: "Line 1\nLine 2\n", // Interpreted newline, plus added newline
			expectedStored: "Line 1\nLine 2",   // Stored without added newline
		},
		{
			name:           "raw true - tab escape sequence",
			args:           map[string]string{"raw": "true"},
			input:          "A\\tB\\tC",
			expectedOutput: "A\\tB\\tC\n", // Literal backslash-t, plus added newline
			expectedStored: "A\\tB\\tC",   // Stored without added newline
		},
		{
			name:           "raw false - tab escape sequence",
			args:           map[string]string{"raw": "false"},
			input:          "A\\tB\\tC",
			expectedOutput: "A\tB\tC\n", // Interpreted tabs, plus added newline
			expectedStored: "A\tB\tC",   // Stored without added newline
		},
		{
			name:           "raw true - multiple escape sequences",
			args:           map[string]string{"raw": "true"},
			input:          "Hello\\nWorld\\t!\\r\\\\",
			expectedOutput: "Hello\\nWorld\\t!\\r\\\\\n", // All literal, plus added newline
			expectedStored: "Hello\\nWorld\\t!\\r\\\\",   // Stored without added newline
		},
		{
			name:           "raw false - multiple escape sequences",
			args:           map[string]string{"raw": "false"},
			input:          "Hello\\nWorld\\t!\\r\\\\",
			expectedOutput: "Hello\nWorld\t!\r\\\n", // All interpreted, plus added newline
			expectedStored: "Hello\nWorld\t!\r\\",   // Stored without added newline
		},
		{
			name:           "raw true - quote escape sequences",
			args:           map[string]string{"raw": "true"},
			input:          "Say \\\"Hello\\\" to me",
			expectedOutput: "Say \\\"Hello\\\" to me\n", // Literal backslash-quote, plus added newline
			expectedStored: "Say \\\"Hello\\\" to me",   // Stored without added newline
		},
		{
			name:           "raw false - quote escape sequences",
			args:           map[string]string{"raw": "false"},
			input:          "Say \\\"Hello\\\" to me",
			expectedOutput: "Say \"Hello\" to me\n", // Interpreted quotes, plus added newline
			expectedStored: "Say \"Hello\" to me",   // Stored without added newline
		},
		{
			name:           "raw true - no escape sequences",
			args:           map[string]string{"raw": "true"},
			input:          "Normal text with no escapes",
			expectedOutput: "Normal text with no escapes\n", // Same as normal, plus added newline
			expectedStored: "Normal text with no escapes",   // Stored without added newline
		},
		{
			name:           "raw false - no escape sequences",
			args:           map[string]string{"raw": "false"},
			input:          "Normal text with no escapes",
			expectedOutput: "Normal text with no escapes\n", // Same as normal, plus added newline
			expectedStored: "Normal text with no escapes",   // Stored without added newline
		},
		{
			name:           "default behavior (no raw option)",
			args:           map[string]string{},
			input:          "Text\\nwith\\tescapes",
			expectedOutput: "Text\nwith\tescapes\n", // Default is raw=false, plus added newline
			expectedStored: "Text\nwith\tescapes",   // Stored without added newline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err)
			})

			assert.Equal(t, tt.expectedOutput, output)

			// Check that result is stored correctly in ${_output}
			value, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStored, value)
		})
	}
}

func TestEchoCommand_Execute_RawOptionCombinations(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		expectedVar    string
		expectedValue  string
		shouldPrint    bool
	}{
		{
			name:           "raw=true with silent=true",
			args:           map[string]string{"raw": "true", "silent": "true"},
			input:          "Hidden\\nmessage",
			expectedOutput: "",                 // No output due to silent
			expectedVar:    "_output",          // Default variable
			expectedValue:  "Hidden\\nmessage", // Stored as literal
			shouldPrint:    false,
		},
		{
			name:           "raw=true with custom to variable",
			args:           map[string]string{"raw": "true", "to": "custom_var"},
			input:          "Custom\\tvalue",
			expectedOutput: "Custom\\tvalue\n", // Literal output
			expectedVar:    "custom_var",       // Custom variable
			expectedValue:  "Custom\\tvalue",   // Stored as literal
			shouldPrint:    true,
		},
		{
			name:           "raw=true with silent=true and custom to variable",
			args:           map[string]string{"raw": "true", "silent": "true", "to": "secret"},
			input:          "Secret\\r\\nmessage",
			expectedOutput: "",                    // No output due to silent
			expectedVar:    "secret",              // Custom variable
			expectedValue:  "Secret\\r\\nmessage", // Stored as literal
			shouldPrint:    false,
		},
		{
			name:           "raw=false with silent=true and custom to variable",
			args:           map[string]string{"raw": "false", "silent": "true", "to": "interpreted"},
			input:          "Normal\\tmode",
			expectedOutput: "",             // No output due to silent
			expectedVar:    "interpreted",  // Custom variable
			expectedValue:  "Normal\tmode", // Stored as interpreted
			shouldPrint:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err)
			})

			if tt.shouldPrint {
				assert.Equal(t, tt.expectedOutput, output)
			} else {
				assert.Empty(t, output, "Silent mode should produce no output")
			}

			// Check that result is stored in correct variable
			value, err := ctx.GetVariable(tt.expectedVar)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestEchoCommand_Execute_InvalidRawOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	args := map[string]string{"raw": "invalid"}
	input := "Test message"

	// Capture stdout
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(args, input)
		assert.NoError(t, err) // Echo never returns errors, uses default value
	})

	// Should use default raw=false and output the message normally
	output = strings.TrimSuffix(output, "\n")
	assert.Equal(t, "Test message", output)
}

func TestEchoCommand_Execute_RawOptionEdgeCases(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		expectedStored string
	}{
		{
			name:           "raw true - only escape sequences",
			args:           map[string]string{"raw": "true"},
			input:          "\\n\\t\\r",
			expectedOutput: "\\n\\t\\r\n", // All literal, plus added newline
			expectedStored: "\\n\\t\\r",   // Stored without added newline
		},
		{
			name:           "raw false - only escape sequences",
			args:           map[string]string{"raw": "false"},
			input:          "\\n\\t\\r",
			expectedOutput: "\n\t\r\n", // All interpreted, plus added newline
			expectedStored: "\n\t\r",   // Stored without added newline
		},
		{
			name:           "raw true - text already ends with literal newline",
			args:           map[string]string{"raw": "true"},
			input:          "Text\\n",
			expectedOutput: "Text\\n\n", // Literal backslash-n, plus added newline
			expectedStored: "Text\\n",   // Stored without added newline
		},
		{
			name:           "raw false - text already ends with interpreted newline",
			args:           map[string]string{"raw": "false"},
			input:          "Text\\n",
			expectedOutput: "Text\n", // Interpreted newline, no additional newline
			expectedStored: "Text\n", // Stored as-is
		},
		{
			name:           "raw true - empty string",
			args:           map[string]string{"raw": "true"},
			input:          "",
			expectedOutput: "", // Empty input produces no output
			expectedStored: "",
		},
		{
			name:           "raw true - unicode with escapes",
			args:           map[string]string{"raw": "true"},
			input:          "Hello 世界\\nWorld 🌍",
			expectedOutput: "Hello 世界\\nWorld 🌍\n", // Unicode preserved, escapes literal
			expectedStored: "Hello 世界\\nWorld 🌍",   // Stored without added newline
		},
		{
			name:           "raw false - unicode with escapes",
			args:           map[string]string{"raw": "false"},
			input:          "Hello 世界\\nWorld 🌍",
			expectedOutput: "Hello 世界\nWorld 🌍\n", // Unicode preserved, escapes interpreted
			expectedStored: "Hello 世界\nWorld 🌍",   // Stored without added newline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err) // Echo never returns errors
			})

			assert.Equal(t, tt.expectedOutput, output)

			// Check that result is stored correctly in ${_output}
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

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
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

func TestEchoCommand_Execute_DisplayOnlyOption(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		shouldStore    bool
		expectedVar    string
		expectedValue  string
		shouldDisplay  bool
	}{
		{
			name:           "display_only=true without to option",
			args:           map[string]string{"display_only": "true"},
			input:          "Debug message",
			expectedOutput: "Debug message\n",
			shouldStore:    false,
			expectedVar:    "_output",
			expectedValue:  "", // Should not be stored
			shouldDisplay:  true,
		},
		{
			name:           "display_only=true with to option",
			args:           map[string]string{"display_only": "true", "to": "debug_var"},
			input:          "Store and display",
			expectedOutput: "Store and display\n",
			shouldStore:    true,
			expectedVar:    "debug_var",
			expectedValue:  "Store and display",
			shouldDisplay:  true,
		},
		{
			name:           "display_only=true with silent=true",
			args:           map[string]string{"display_only": "true", "silent": "true"},
			input:          "Nothing happens",
			expectedOutput: "",
			shouldStore:    false,
			expectedVar:    "_output",
			expectedValue:  "", // Should not be stored
			shouldDisplay:  false,
		},
		{
			name:           "display_only=false (default behavior)",
			args:           map[string]string{"display_only": "false"},
			input:          "Normal behavior",
			expectedOutput: "Normal behavior\n",
			shouldStore:    true,
			expectedVar:    "_output",
			expectedValue:  "Normal behavior",
			shouldDisplay:  true,
		},
		{
			name:           "display_only=true with raw=true",
			args:           map[string]string{"display_only": "true", "raw": "true"},
			input:          "Raw\\ntext",
			expectedOutput: "Raw\\ntext\n",
			shouldStore:    false,
			expectedVar:    "_output",
			expectedValue:  "", // Should not be stored
			shouldDisplay:  true,
		},
		{
			name:           "display_only=true with to and silent=false",
			args:           map[string]string{"display_only": "true", "to": "custom", "silent": "false"},
			input:          "Display and store",
			expectedOutput: "Display and store\n",
			shouldStore:    true,
			expectedVar:    "custom",
			expectedValue:  "Display and store",
			shouldDisplay:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous values
			err := ctx.SetSystemVariable("_output", "")
			require.NoError(t, err, "Should be able to clear _output variable")
			if tt.expectedVar != "_output" {
				if tt.expectedVar[0] == '_' || tt.expectedVar[0] == '#' || tt.expectedVar[0] == '@' {
					err = ctx.SetSystemVariable(tt.expectedVar, "")
					require.NoError(t, err, "Should be able to clear %s variable", tt.expectedVar)
				} else {
					err = ctx.SetVariable(tt.expectedVar, "")
					require.NoError(t, err, "Should be able to clear %s variable", tt.expectedVar)
				}
			}

			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err) // Echo never returns errors
			})

			// Check console output
			if tt.shouldDisplay {
				assert.Equal(t, tt.expectedOutput, output)
			} else {
				assert.Empty(t, output, "Should not display when silent")
			}

			// Check variable storage
			value, err := ctx.GetVariable(tt.expectedVar)
			if tt.shouldStore {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			} else if err == nil {
				// Variable should either not exist or be empty
				assert.Equal(t, "", value, "Variable should not be set when not storing")
			}
		})
	}
}

func TestEchoCommand_Execute_DisplayOnlyInvalidValue(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	args := map[string]string{"display_only": "invalid"}
	input := "Test message"

	// Capture stdout
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(args, input)
		assert.NoError(t, err) // Echo never returns errors, uses default value
	})

	// Should use default display_only=false and behave normally
	output = strings.TrimSuffix(output, "\n")
	assert.Equal(t, "Test message", output)

	// Should store in _output as normal
	value, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Equal(t, "Test message", value)
}

func TestEchoCommand_Execute_DisplayOnlyEdgeCases(t *testing.T) {
	cmd := &EchoCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupEchoTestRegistry(t, ctx)

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		shouldStore    bool
		expectedVar    string
		expectedValue  string
	}{
		{
			name:           "display_only=true with empty input",
			args:           map[string]string{"display_only": "true"},
			input:          "",
			expectedOutput: "",
			shouldStore:    false,
			expectedVar:    "_output",
			expectedValue:  "",
		},
		{
			name:           "display_only=true with unicode",
			args:           map[string]string{"display_only": "true"},
			input:          "Hello 世界 🌍",
			expectedOutput: "Hello 世界 🌍\n",
			shouldStore:    false,
			expectedVar:    "_output",
			expectedValue:  "",
		},
		{
			name:           "display_only with all options combined",
			args:           map[string]string{"display_only": "true", "to": "all_opts", "raw": "true", "silent": "false"},
			input:          "Complex\\ntest",
			expectedOutput: "Complex\\ntest\n", // raw=true, silent=false
			shouldStore:    true,               // to= specified
			expectedVar:    "all_opts",
			expectedValue:  "Complex\\ntest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous values
			err := ctx.SetSystemVariable("_output", "")
			require.NoError(t, err, "Should be able to clear _output variable")
			if tt.expectedVar != "_output" {
				if tt.expectedVar[0] == '_' || tt.expectedVar[0] == '#' || tt.expectedVar[0] == '@' {
					err = ctx.SetSystemVariable(tt.expectedVar, "")
					require.NoError(t, err, "Should be able to clear %s variable", tt.expectedVar)
				} else {
					err = ctx.SetVariable(tt.expectedVar, "")
					require.NoError(t, err, "Should be able to clear %s variable", tt.expectedVar)
				}
			}

			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err)
			})

			assert.Equal(t, tt.expectedOutput, output)

			// Check variable storage
			value, err := ctx.GetVariable(tt.expectedVar)
			if tt.shouldStore {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			} else if err == nil {
				// Variable should either not exist or be empty
				assert.Equal(t, "", value, "Variable should not be set when not storing")
			}
		})
	}
}

// Interface compliance check
var _ neurotypes.Command = (*EchoCommand)(nil)
