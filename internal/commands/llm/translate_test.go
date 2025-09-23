package llm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestTranslateCommand_Name(t *testing.T) {
	cmd := &TranslateCommand{}
	assert.Equal(t, "translate", cmd.Name())
}

func TestTranslateCommand_ParseMode(t *testing.T) {
	cmd := &TranslateCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestTranslateCommand_Description(t *testing.T) {
	cmd := &TranslateCommand{}
	description := cmd.Description()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "translate")
	assert.Contains(t, description, "AI translation")
}

func TestTranslateCommand_Usage(t *testing.T) {
	cmd := &TranslateCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\translate")
	assert.Contains(t, usage, "translator=")
	assert.Contains(t, usage, "source=")
	assert.Contains(t, usage, "target=")
	assert.Contains(t, usage, "instruction=")
}

func TestTranslateCommand_HelpInfo(t *testing.T) {
	cmd := &TranslateCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "translate", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check that all expected options are present
	expectedOptions := []string{"translator", "source", "target", "instruction"}
	assert.Len(t, helpInfo.Options, len(expectedOptions))

	optionNames := make([]string, len(helpInfo.Options))
	for i, option := range helpInfo.Options {
		optionNames[i] = option.Name
	}

	for _, expected := range expectedOptions {
		assert.Contains(t, optionNames, expected, "Option %s should be present", expected)
	}

	// Check translator option details
	var translatorOption *neurotypes.HelpOption
	for i := range helpInfo.Options {
		if helpInfo.Options[i].Name == "translator" {
			translatorOption = &helpInfo.Options[i]
			break
		}
	}
	require.NotNil(t, translatorOption)
	assert.Equal(t, "zai", translatorOption.Default)

	// Check examples are present
	assert.NotEmpty(t, helpInfo.Examples)
	assert.True(t, len(helpInfo.Examples) >= 3, "Should have at least 3 examples")
}

func TestTranslateCommand_Execute_NoVariableService(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set empty registry for test
	services.SetGlobalRegistry(services.NewRegistry())

	options := map[string]string{}
	input := "Hello world"

	err := cmd.Execute(options, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestTranslateCommand_Execute_NoText(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{}
	input := ""

	err = cmd.Execute(options, input)
	assert.NoError(t, err) // Should show help and return successfully
}

func TestTranslateCommand_Execute_EmptyText(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{}
	input := "   " // Only whitespace

	err = cmd.Execute(options, input)
	assert.NoError(t, err) // Should show help and return successfully for whitespace too
}

func TestTranslateCommand_Execute_InvalidTranslator(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{
		"translator": "invalid-provider",
	}
	input := "Hello world"

	err = cmd.Execute(options, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported translator 'invalid-provider'")
	assert.Contains(t, err.Error(), "zai, deepl, google")
}

func TestTranslateCommand_Execute_DefaultOptions(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize the variable service
	variableService, err := testRegistry.GetService("variable")
	require.NoError(t, err)
	err = variableService.Initialize()
	require.NoError(t, err)

	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{}
	input := "Hello world"

	err = cmd.Execute(options, input)
	assert.NoError(t, err)

	// Command should execute successfully with default options
	// Variable setting and actual translation will be implemented in future phases
}

func TestTranslateCommand_Execute_CustomOptions(t *testing.T) {
	cmd := &TranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize the variable service
	variableService, err := testRegistry.GetService("variable")
	require.NoError(t, err)
	err = variableService.Initialize()
	require.NoError(t, err)

	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{
		"translator":  "deepl",
		"source":      "french",
		"target":      "spanish",
		"instruction": "make it business style and formal",
	}
	input := "Bonjour le monde"

	err = cmd.Execute(options, input)
	assert.NoError(t, err)

	// Command should execute successfully with custom options
	// Variable setting and actual translation will be implemented in future phases
}

func TestTranslateCommand_Execute_AllSupportedTranslators(t *testing.T) {
	cmd := &TranslateCommand{}

	supportedTranslators := []string{"zai", "deepl", "google"}

	for _, translator := range supportedTranslators {
		t.Run("translator_"+translator, func(t *testing.T) {
			// Save current registry and restore after test
			originalRegistry := services.GetGlobalRegistry()
			defer services.SetGlobalRegistry(originalRegistry)

			// Set up registry with variable service
			testRegistry := services.NewRegistry()
			err := testRegistry.RegisterService(services.NewVariableService())
			require.NoError(t, err)

			// Initialize the variable service
			variableService, err := testRegistry.GetService("variable")
			require.NoError(t, err)
			err = variableService.Initialize()
			require.NoError(t, err)

			services.SetGlobalRegistry(testRegistry)

			options := map[string]string{
				"translator": translator,
			}
			input := "Hello world"

			err = cmd.Execute(options, input)
			assert.NoError(t, err, "Should succeed with translator: %s", translator)

			// Command should execute successfully for all supported translators
			// Specific translator validation will be implemented in future phases
		})
	}
}

func TestTranslateCommand_IsReadOnly(t *testing.T) {
	cmd := &TranslateCommand{}
	assert.False(t, cmd.IsReadOnly(), "Translate command should not be read-only as it sets variables")
}

func TestGetOption(t *testing.T) {
	tests := []struct {
		name         string
		options      map[string]string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "existing option",
			options:      map[string]string{"key": "value"},
			key:          "key",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "missing option",
			options:      map[string]string{},
			key:          "key",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty option value",
			options:      map[string]string{"key": ""},
			key:          "key",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "nil options map",
			options:      nil,
			key:          "key",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOption(tt.options, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
