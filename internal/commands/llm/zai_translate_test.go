package llm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestZaiTranslateCommand_Name(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	assert.Equal(t, "zai-translate", cmd.Name())
}

func TestZaiTranslateCommand_ParseMode(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestZaiTranslateCommand_Description(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	description := cmd.Description()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "zai")
	assert.Contains(t, description, "translation")
}

func TestZaiTranslateCommand_Usage(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\zai-translate")
	assert.Contains(t, usage, "source=")
	assert.Contains(t, usage, "target=")
	assert.Contains(t, usage, "strategy=")
	assert.Contains(t, usage, "suggestion=")
	assert.Contains(t, usage, "glossary=")
}

func TestZaiTranslateCommand_HelpInfo(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "zai-translate", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check that all expected options are present
	expectedOptions := []string{"source", "target", "strategy", "suggestion", "glossary"}
	assert.Len(t, helpInfo.Options, len(expectedOptions))

	optionNames := make([]string, len(helpInfo.Options))
	for i, option := range helpInfo.Options {
		optionNames[i] = option.Name
	}

	for _, expected := range expectedOptions {
		assert.Contains(t, optionNames, expected, "Option %s should be present", expected)
	}

	// Check examples are present
	assert.NotEmpty(t, helpInfo.Examples)
	assert.True(t, len(helpInfo.Examples) >= 5, "Should have at least 5 examples")

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	assert.True(t, len(helpInfo.StoredVariables) >= 3, "Should have at least 3 stored variables")

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestZaiTranslateCommand_Execute_NoVariableService(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

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

func TestZaiTranslateCommand_Execute_NoHTTPService(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with only variable service
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	options := map[string]string{}
	input := "Hello world"

	err = cmd.Execute(options, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http request service not available")
}

func TestZaiTranslateCommand_Execute_NoText(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with required services
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	options := map[string]string{}
	input := ""

	err = cmd.Execute(options, input)
	assert.NoError(t, err) // Should show help and return successfully
}

func TestZaiTranslateCommand_Execute_EmptyText(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with required services
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	options := map[string]string{}
	input := "   " // Only whitespace

	err = cmd.Execute(options, input)
	assert.NoError(t, err) // Should show help and return successfully for whitespace too
}

func TestZaiTranslateCommand_Execute_InvalidStrategy(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with required services
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	options := map[string]string{
		"strategy": "invalid-strategy",
	}
	input := "Hello world"

	err = cmd.Execute(options, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported strategy 'invalid-strategy'")
	assert.Contains(t, err.Error(), "general, paraphrase, two_step, three_step, reflection")
}

func TestZaiTranslateCommand_Execute_NoAPIKey(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with required services
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	options := map[string]string{}
	input := "Hello world"

	err = cmd.Execute(options, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ZAI API key not found")
}

func TestZaiTranslateCommand_Execute_ValidStrategies(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	validStrategies := []string{"general", "paraphrase", "two_step", "three_step", "reflection"}

	for _, strategy := range validStrategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			// Save current registry and restore after test
			originalRegistry := services.GetGlobalRegistry()
			defer services.SetGlobalRegistry(originalRegistry)

			// Set up registry with required services
			testRegistry := services.NewRegistry()
			err := testRegistry.RegisterService(services.NewVariableService())
			require.NoError(t, err)
			err = testRegistry.RegisterService(services.NewHTTPRequestService())
			require.NoError(t, err)
			services.SetGlobalRegistry(testRegistry)

			// Initialize services
			err = testRegistry.InitializeAll()
			require.NoError(t, err)

			// Set a mock API key
			variableService, err := testRegistry.GetService("variable")
			require.NoError(t, err)
			err = variableService.(*services.VariableService).Set("ZAI_API_KEY", "test-key")
			require.NoError(t, err)

			options := map[string]string{
				"strategy": strategy,
			}
			input := "Hello world"

			// This will fail at HTTP request stage, but should pass strategy validation
			err = cmd.Execute(options, input)
			assert.Error(t, err) // Expected to fail at HTTP stage with mock key
			assert.NotContains(t, err.Error(), "unsupported strategy", "Strategy %s should be valid", strategy)
		})
	}
}

func TestZaiTranslateCommand_IsReadOnly(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	assert.False(t, cmd.IsReadOnly(), "ZAI translate command should not be read-only as it sets variables")
}

func TestZaiTranslateCommand_DefaultOptions(t *testing.T) {
	// Test that default values are used when options are not provided
	tests := []struct {
		name         string
		options      map[string]string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "default source",
			options:      map[string]string{},
			key:          "source",
			defaultValue: "auto",
			expected:     "auto",
		},
		{
			name:         "default target",
			options:      map[string]string{},
			key:          "target",
			defaultValue: "zh-CN",
			expected:     "zh-CN",
		},
		{
			name:         "default strategy",
			options:      map[string]string{},
			key:          "strategy",
			defaultValue: "general",
			expected:     "general",
		},
		{
			name:         "override default",
			options:      map[string]string{"source": "en"},
			key:          "source",
			defaultValue: "auto",
			expected:     "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOption(tt.options, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
