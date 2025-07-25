package builtin

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestSetEnvCommand_Name(t *testing.T) {
	cmd := &SetEnvCommand{}
	assert.Equal(t, "set-env", cmd.Name())
}

func TestSetEnvCommand_ParseMode(t *testing.T) {
	cmd := &SetEnvCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestSetEnvCommand_Description(t *testing.T) {
	cmd := &SetEnvCommand{}
	assert.Equal(t, "Set an environment variable", cmd.Description())
}

func TestSetEnvCommand_Usage(t *testing.T) {
	cmd := &SetEnvCommand{}
	assert.Equal(t, "\\set-env[VAR=value] or \\set-env VAR value", cmd.Usage())
}

func TestSetEnvCommand_HelpInfo(t *testing.T) {
	cmd := &SetEnvCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "set-env", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that VAR option exists
	found := false
	for _, option := range helpInfo.Options {
		if option.Name == "VAR" {
			found = true
			assert.True(t, option.Required)
			assert.Equal(t, "string", option.Type)
			break
		}
	}
	assert.True(t, found, "VAR option should exist")
}

func TestSetEnvCommand_Execute_BracketSyntax_TestMode(t *testing.T) {
	cmd := &SetEnvCommand{}

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		expectedEnvVar string
		expectedValue  string
	}{
		{
			name:           "set single environment variable",
			args:           map[string]string{"TEST_VAR": "testvalue"},
			input:          "",
			expectedOutput: "Setting TEST_VAR = testvalue\n",
			expectedEnvVar: "TEST_VAR",
			expectedValue:  "testvalue",
		},
		{
			name:           "set variable with empty value",
			args:           map[string]string{"EMPTY_VAR": ""},
			input:          "",
			expectedOutput: "Setting EMPTY_VAR = \n",
			expectedEnvVar: "EMPTY_VAR",
			expectedValue:  "",
		},
		{
			name:           "set variable with special characters",
			args:           map[string]string{"SPECIAL_VAR": "value with spaces & symbols!"},
			input:          "",
			expectedOutput: "Setting SPECIAL_VAR = value with spaces & symbols!\n",
			expectedEnvVar: "SPECIAL_VAR",
			expectedValue:  "value with spaces & symbols!",
		},
		{
			name:           "set API key variable",
			args:           map[string]string{"OPENAI_API_KEY": "sk-1234567890abcdef1234567890abcdef"},
			input:          "",
			expectedOutput: "Setting OPENAI_API_KEY = sk-1234567890abcdef1234567890abcdef\n",
			expectedEnvVar: "OPENAI_API_KEY",
			expectedValue:  "sk-1234567890abcdef1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupSetEnvCommandTestRegistry(t)

			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(tt.args, tt.input)
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, outputStr)

			// In test mode, verify the environment variable was set via test override
			actualValue := ctx.GetEnv(tt.expectedEnvVar)
			assert.Equal(t, tt.expectedValue, actualValue)

			// Also verify via variable service
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)
			serviceValue := variableService.GetEnv(tt.expectedEnvVar)
			assert.Equal(t, tt.expectedValue, serviceValue)
		})
	}
}

func TestSetEnvCommand_Execute_SpaceSyntax_TestMode(t *testing.T) {
	cmd := &SetEnvCommand{}

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		expectedOutput string
		expectedEnvVar string
		expectedValue  string
	}{
		{
			name:           "set variable with space syntax",
			args:           map[string]string{},
			input:          "TEST_VAR testvalue",
			expectedOutput: "Setting TEST_VAR = testvalue\n",
			expectedEnvVar: "TEST_VAR",
			expectedValue:  "testvalue",
		},
		{
			name:           "set variable with multiple words in value",
			args:           map[string]string{},
			input:          "PATH_VAR /usr/local/bin:/usr/bin:/bin",
			expectedOutput: "Setting PATH_VAR = /usr/local/bin:/usr/bin:/bin\n",
			expectedEnvVar: "PATH_VAR",
			expectedValue:  "/usr/local/bin:/usr/bin:/bin",
		},
		{
			name:           "set variable with single word (empty value)",
			args:           map[string]string{},
			input:          "EMPTY_VAR",
			expectedOutput: "Setting EMPTY_VAR = \n",
			expectedEnvVar: "EMPTY_VAR",
			expectedValue:  "",
		},
		{
			name:           "set variable with spaces around",
			args:           map[string]string{},
			input:          "  TRIMMED_VAR   testvalue with spaces  ",
			expectedOutput: "Setting TRIMMED_VAR = testvalue with spaces  \n",
			expectedEnvVar: "TRIMMED_VAR",
			expectedValue:  "testvalue with spaces  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupSetEnvCommandTestRegistry(t)

			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(tt.args, tt.input)
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, outputStr)

			// Verify the environment variable was set
			actualValue := ctx.GetEnv(tt.expectedEnvVar)
			assert.Equal(t, tt.expectedValue, actualValue)
		})
	}
}

func TestSetEnvCommand_Execute_ProductionMode(t *testing.T) {
	// This test verifies production mode behavior (sets actual OS env vars)
	cmd := &SetEnvCommand{}
	ctx := setupSetEnvCommandProductionRegistry(t)

	// Use a unique test environment variable name to avoid conflicts
	testEnvVar := "NEUROSHELL_TEST_ENV_VAR_UNIQUE"
	testValue := "test_production_value"

	// Clean up any existing value
	originalValue := os.Getenv(testEnvVar)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(testEnvVar, originalValue)
		} else {
			_ = os.Unsetenv(testEnvVar)
		}
	}()

	args := map[string]string{testEnvVar: testValue}

	// Capture stdout
	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "")
	})

	assert.NoError(t, err)
	assert.Contains(t, outputStr, fmt.Sprintf("Setting %s = %s", testEnvVar, testValue))

	// In production mode, verify the actual OS environment variable was set
	actualValue := os.Getenv(testEnvVar)
	assert.Equal(t, testValue, actualValue)

	// Also verify via context (should get the same value)
	contextValue := ctx.GetEnv(testEnvVar)
	assert.Equal(t, testValue, contextValue)
}

func TestSetEnvCommand_Execute_MultipleVariables(t *testing.T) {
	cmd := &SetEnvCommand{}
	setupSetEnvCommandTestRegistry(t)

	args := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
		"VAR3": "value3",
	}

	// Capture stdout
	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "")
	})

	assert.NoError(t, err)

	// Check that all variables are mentioned in output (order may vary due to sorting)
	for varName, varValue := range args {
		assert.Contains(t, outputStr, fmt.Sprintf("Setting %s = %s", varName, varValue))
	}

	// Verify all variables were set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	for varName, expectedValue := range args {
		actualValue := variableService.GetEnv(varName)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestSetEnvCommand_Execute_EmptyInputAndArgs(t *testing.T) {
	cmd := &SetEnvCommand{}
	setupSetEnvCommandTestRegistry(t)

	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestSetEnvCommand_Execute_PrioritizeBracketSyntax(t *testing.T) {
	cmd := &SetEnvCommand{}
	ctx := setupSetEnvCommandTestRegistry(t)

	// When both args and input are provided, args (bracket syntax) should take priority
	args := map[string]string{"BRACKET_VAR": "bracketvalue"}
	input := "SPACE_VAR spacevalue"

	// Capture stdout
	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, input)
	})

	assert.NoError(t, err)

	// Should only set bracket syntax variable
	bracketValue := ctx.GetEnv("BRACKET_VAR")
	assert.Equal(t, "bracketvalue", bracketValue)

	// Space syntax variable should not be set
	spaceValue := ctx.GetEnv("SPACE_VAR")
	assert.Equal(t, "", spaceValue)

	// Output should only mention bracket variable
	assert.Contains(t, outputStr, "Setting BRACKET_VAR = bracketvalue")
	assert.NotContains(t, outputStr, "SPACE_VAR")
}

func TestSetEnvCommand_Execute_VariableServiceError(t *testing.T) {
	cmd := &SetEnvCommand{}

	// Don't set up variable service to simulate error
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	services.SetGlobalRegistry(services.NewRegistry())

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	args := map[string]string{"TEST_VAR": "testvalue"}

	err := cmd.Execute(args, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

// setupSetEnvCommandTestRegistry creates a clean test registry with test mode enabled
func setupSetEnvCommandTestRegistry(t *testing.T) neurotypes.Context {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	return ctx
}

// setupSetEnvCommandProductionRegistry creates a test registry with production mode (test mode disabled)
func setupSetEnvCommandProductionRegistry(t *testing.T) neurotypes.Context {
	ctx := context.New() // Use regular context (not test context)
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	return ctx
}

// Interface compliance check
var _ neurotypes.Command = (*SetEnvCommand)(nil)
