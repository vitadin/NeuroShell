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
	assert.Contains(t, usage, "instruction=")
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
	expectedOptions := []string{"source", "target", "strategy", "instruction", "glossary"}
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
	assert.True(t, len(helpInfo.StoredVariables) >= 6, "Should have at least 6 stored variables")

	// Check that the new language list variables are present
	storedVarNames := make([]string, len(helpInfo.StoredVariables))
	for i, variable := range helpInfo.StoredVariables {
		storedVarNames[i] = variable.Name
	}
	assert.Contains(t, storedVarNames, "_zai_source_languages")
	assert.Contains(t, storedVarNames, "_zai_target_languages")

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

func TestZaiTranslateCommand_Execute_NoAPIKey_SkippedIfKeysPresent(t *testing.T) {
	cmd := &ZaiTranslateCommand{}

	// Save current registry and restore after test
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	// Set up registry with required services using a clean context
	testRegistry := services.NewRegistry()

	// Create a variable service that won't have access to the real environment
	variableService := services.NewVariableService()
	err := testRegistry.RegisterService(variableService)
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	// The variable service in test mode should not have access to OS environment variables
	// unless explicitly set, so the API key lookup should fail

	options := map[string]string{}
	input := "Hello world"

	err = cmd.Execute(options, input)

	// Check if we have a real API key in the environment that's leaking into the test
	variableServiceInstance, err2 := testRegistry.GetService("variable")
	require.NoError(t, err2)
	vs := variableServiceInstance.(*services.VariableService)
	apiKey1, _ := vs.Get("os.Z_DOT_AI_API_KEY")
	apiKey2, _ := vs.Get("os.ZAI_API_KEY")

	if apiKey1 != "" || apiKey2 != "" {
		t.Skip("Skipping test: API keys found in test environment - cannot test no-key scenario reliably")
		return
	}

	// If no API keys were found, we should expect an error
	assert.Error(t, err, "Expected error when no API key is available")
	if err != nil {
		assert.Contains(t, err.Error(), "ZAI API key not found")
	}
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

			// Check if we have real API keys in environment
			variableService, err := testRegistry.GetService("variable")
			require.NoError(t, err)
			vs := variableService.(*services.VariableService)

			// Check for existing API keys
			apiKey1, _ := vs.Get("os.Z_DOT_AI_API_KEY")
			apiKey2, _ := vs.Get("os.ZAI_API_KEY")

			if apiKey1 == "" && apiKey2 == "" {
				// Set a mock API key for testing (will fail at HTTP stage)
				err = vs.Set("os.ZAI_API_KEY", "test-key")
				require.NoError(t, err)
			}

			options := map[string]string{
				"strategy": strategy,
			}
			input := "Hello world"

			// Execute the command - strategy validation should pass
			err = cmd.Execute(options, input)

			// If we have real API keys, command might succeed with actual translation
			// If we have mock keys, it should fail at HTTP stage
			// Either way, strategy should be valid (no "unsupported strategy" error)
			if err != nil {
				assert.NotContains(t, err.Error(), "unsupported strategy", "Strategy %s should be valid", strategy)
			}
			// We don't assert.Error here because with real API keys, the command might succeed
		})
	}
}

func TestZaiTranslateCommand_IsReadOnly(t *testing.T) {
	cmd := &ZaiTranslateCommand{}
	assert.False(t, cmd.IsReadOnly(), "ZAI translate command should not be read-only as it sets variables")
}

func TestZaiTranslateCommand_Execute_LanguageVariablesSetOnEmptyInput(t *testing.T) {
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

	// Get variable service
	variableService, err := testRegistry.GetService("variable")
	require.NoError(t, err)

	options := map[string]string{}
	input := "" // Empty input

	// Execute command with empty input (should show help but set language variables)
	err = cmd.Execute(options, input)
	assert.NoError(t, err)

	// Check that language variables were set
	sourceLanguages, err := variableService.(*services.VariableService).Get("_zai_source_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, sourceLanguages)
	assert.Contains(t, sourceLanguages, "auto")
	assert.Contains(t, sourceLanguages, "en")
	assert.Contains(t, sourceLanguages, "zh-CN")

	targetLanguages, err := variableService.(*services.VariableService).Get("_zai_target_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, targetLanguages)
	assert.Contains(t, targetLanguages, "en")
	assert.Contains(t, targetLanguages, "zh-CN")
	assert.NotContains(t, targetLanguages, "auto") // Target shouldn't contain auto
}

func TestZaiTranslateCommand_Execute_LanguageVariablesSetWithInput(t *testing.T) {
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

	// Get variable service
	variableService, err := testRegistry.GetService("variable")
	require.NoError(t, err)

	// Set a mock API key (this will still fail at HTTP stage, but that's expected)
	err = variableService.(*services.VariableService).Set("os.ZAI_API_KEY", "test-key")
	require.NoError(t, err)

	options := map[string]string{}
	input := "Hello world"

	// Execute command - language variables should be set regardless of success/failure
	_ = cmd.Execute(options, input)

	// Command might succeed with real API keys or fail with mock keys
	// We don't assert error here since behavior depends on environment

	// Language variables should be set regardless of translation success/failure
	sourceLanguages, err := variableService.(*services.VariableService).Get("_zai_source_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, sourceLanguages)
	assert.Contains(t, sourceLanguages, "auto")
	assert.Contains(t, sourceLanguages, "en")
	assert.Contains(t, sourceLanguages, "zh-CN")

	targetLanguages, err := variableService.(*services.VariableService).Get("_zai_target_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, targetLanguages)
	assert.Contains(t, targetLanguages, "en")
	assert.Contains(t, targetLanguages, "zh-CN")
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
			defaultValue: "en",
			expected:     "en",
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
