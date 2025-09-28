package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
)

// TestEchoCommand_Integration_DisplayOnlyBehavior performs integration testing
// of the display_only option to ensure it works correctly in isolation
func TestEchoCommand_Integration_DisplayOnlyBehavior(t *testing.T) {
	tests := []struct {
		name          string
		args          map[string]string
		input         string
		shouldStore   bool
		shouldDisplay bool
		targetVar     string
	}{
		{
			name:          "display_only=true should not store in _output",
			args:          map[string]string{"display_only": "true"},
			input:         "Test message",
			shouldStore:   false,
			shouldDisplay: true,
			targetVar:     "_output",
		},
		{
			name:          "display_only=true with to=custom should store in custom var",
			args:          map[string]string{"display_only": "true", "to": "custom"},
			input:         "Test message",
			shouldStore:   true,
			shouldDisplay: true,
			targetVar:     "custom",
		},
		{
			name:          "display_only=true with silent=true should do nothing",
			args:          map[string]string{"display_only": "true", "silent": "true"},
			input:         "Test message",
			shouldStore:   false,
			shouldDisplay: false,
			targetVar:     "_output",
		},
		{
			name:          "display_only=false should store normally",
			args:          map[string]string{"display_only": "false"},
			input:         "Test message",
			shouldStore:   true,
			shouldDisplay: true,
			targetVar:     "_output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			cmd := &EchoCommand{}
			ctx := context.New()

			// Set up isolated test registry
			oldRegistry := services.GetGlobalRegistry()
			services.SetGlobalRegistry(services.NewRegistry())
			context.SetGlobalContext(ctx)

			err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
			require.NoError(t, err)
			err = services.GetGlobalRegistry().InitializeAll()
			require.NoError(t, err)

			// Clean up after test
			defer func() {
				services.SetGlobalRegistry(oldRegistry)
				context.ResetGlobalContext()
			}()

			// Clear target variable before test
			if tt.targetVar[0] == '_' || tt.targetVar[0] == '#' || tt.targetVar[0] == '@' {
				_ = ctx.SetSystemVariable(tt.targetVar, "")
			} else {
				_ = ctx.SetVariable(tt.targetVar, "")
			}

			// Capture stdout
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				assert.NoError(t, err)
			})

			// Check display behavior
			if tt.shouldDisplay {
				expectedOutput := tt.input
				if !strings.HasSuffix(expectedOutput, "\n") {
					expectedOutput += "\n" // Echo adds newline
				}
				assert.Equal(t, expectedOutput, output)
			} else {
				assert.Empty(t, output, "Should not display when silent")
			}

			// Check storage behavior
			value, err := ctx.GetVariable(tt.targetVar)
			if tt.shouldStore {
				assert.NoError(t, err)
				assert.Equal(t, tt.input, value)
			} else if err == nil {
				// Variable should either not exist or be empty
				assert.Empty(t, value, "Variable should not be set when not storing")
			}
		})
	}
}

// TestEchoCommand_Integration_OptionCombinations tests complex option combinations
func TestEchoCommand_Integration_OptionCombinations(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]string
		input        string
		expectStored map[string]string // map of variable -> expected value
		expectEmpty  []string          // variables that should be empty
	}{
		{
			name:  "display_only with raw and custom variable",
			args:  map[string]string{"display_only": "true", "raw": "true", "to": "result"},
			input: "Raw\\ntext",
			expectStored: map[string]string{
				"result": "Raw\\ntext",
			},
			expectEmpty: []string{"_output"},
		},
		{
			name:         "display_only with silent and no to option",
			args:         map[string]string{"display_only": "true", "silent": "true"},
			input:        "Silent test",
			expectStored: map[string]string{
				// Nothing should be stored
			},
			expectEmpty: []string{"_output"},
		},
		{
			name:  "all options combined for maximum coverage",
			args:  map[string]string{"display_only": "true", "silent": "false", "raw": "false", "to": "final"},
			input: "Complex\\ntest",
			expectStored: map[string]string{
				"final": "Complex\ntest", // raw=false interprets escapes
			},
			expectEmpty: []string{"_output"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			cmd := &EchoCommand{}
			ctx := context.New()

			// Set up isolated test registry
			oldRegistry := services.GetGlobalRegistry()
			services.SetGlobalRegistry(services.NewRegistry())
			context.SetGlobalContext(ctx)

			err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
			require.NoError(t, err)
			err = services.GetGlobalRegistry().InitializeAll()
			require.NoError(t, err)

			// Clean up after test
			defer func() {
				services.SetGlobalRegistry(oldRegistry)
				context.ResetGlobalContext()
			}()

			// Execute command
			err = cmd.Execute(tt.args, tt.input)
			assert.NoError(t, err)

			// Check expected stored variables
			for varName, expectedValue := range tt.expectStored {
				value, err := ctx.GetVariable(varName)
				assert.NoError(t, err, "Variable %s should exist", varName)
				assert.Equal(t, expectedValue, value, "Variable %s should have correct value", varName)
			}

			// Check variables that should be empty
			for _, varName := range tt.expectEmpty {
				value, err := ctx.GetVariable(varName)
				if err == nil {
					assert.Empty(t, value, "Variable %s should be empty", varName)
				}
				// If err != nil, the variable doesn't exist, which is also acceptable
			}
		})
	}
}
