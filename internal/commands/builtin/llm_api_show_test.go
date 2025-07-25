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

func TestLLMAPIShowCommand_Name(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	assert.Equal(t, "llm-api-show", cmd.Name())
}

func TestLLMAPIShowCommand_ParseMode(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestLLMAPIShowCommand_Description(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "api")
	assert.Contains(t, strings.ToLower(desc), "key")
}

func TestLLMAPIShowCommand_Usage(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\llm-api-show")
}

func TestLLMAPIShowCommand_HelpInfo(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "llm-api-show", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that provider option exists
	found := false
	for _, option := range helpInfo.Options {
		if option.Name == "provider" {
			found = true
			assert.False(t, option.Required)
			assert.Equal(t, "all", option.Default)
			break
		}
	}
	assert.True(t, found, "provider option should exist")
}

func TestLLMAPIShowCommand_Execute_NoKeys(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	ctx := setupLLMAPIShowTestRegistry(t)

	// Override all possible LLM environment variables to empty to simulate no keys
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "")
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", "")
	ctx.SetTestEnvOverride("MOONSHOT_API_KEY", "")
	ctx.SetTestEnvOverride("OPENROUTER_API_KEY", "")
	ctx.SetTestEnvOverride("NEURO_OPENAI_API_KEY", "")
	ctx.SetTestEnvOverride("NEURO_ANTHROPIC_API_KEY", "")
	defer ctx.ClearAllTestEnvOverrides()

	// Register actual configuration service
	err := services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     map[string]string
		expected string
	}{
		{
			name:     "no keys found - all providers",
			args:     map[string]string{},
			expected: "No API keys found in any source.",
		},
		{
			name:     "no keys found - specific provider",
			args:     map[string]string{"provider": "openai"},
			expected: "No API keys found for provider 'openai'.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				assert.NoError(t, err)
			})

			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestLLMAPIShowCommand_Execute_WithKeys(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	ctx := setupLLMAPIShowTestRegistry(t)

	// Clear any existing environment variables and set controlled test data
	ctx.ClearAllTestEnvOverrides()
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "sk-1234567890abcdef1234567890abcdef")
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", "ant-1234567890abcdef1234567890abcdef")
	defer ctx.ClearAllTestEnvOverrides()

	// Register actual configuration service
	err := services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	tests := []struct {
		name             string
		args             map[string]string
		expectedContains []string
		expectedStatus   string
	}{
		{
			name: "show all keys",
			args: map[string]string{},
			expectedContains: []string{
				"LLM API Keys Found",
				"os.OPENAI_API_KEY",
				"os.ANTHROPIC_API_KEY",
				"sk-...def",
				"ant...def",
				"INACTIVE",
			},
		},
		{
			name: "show openai keys only",
			args: map[string]string{"provider": "openai"},
			expectedContains: []string{
				"LLM API Keys Found (Openai only)",
				"os.OPENAI_API_KEY",
				"sk-...def",
				"INACTIVE",
			},
		},
		{
			name: "show anthropic keys only",
			args: map[string]string{"provider": "anthropic"},
			expectedContains: []string{
				"LLM API Keys Found (Anthropic only)",
				"os.ANTHROPIC_API_KEY",
				"ant...def",
				"INACTIVE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				assert.NoError(t, err)
			})

			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// Verify that keys are stored as variables
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			if tt.args["provider"] == "" || tt.args["provider"] == "all" || tt.args["provider"] == "openai" {
				value, err := variableService.Get("os.OPENAI_API_KEY")
				assert.NoError(t, err)
				assert.Equal(t, "sk-1234567890abcdef1234567890abcdef", value)
			}

			if tt.args["provider"] == "" || tt.args["provider"] == "all" || tt.args["provider"] == "anthropic" {
				value, err := variableService.Get("os.ANTHROPIC_API_KEY")
				assert.NoError(t, err)
				assert.Equal(t, "ant-1234567890abcdef1234567890abcdef", value)
			}
		})
	}
}

func TestLLMAPIShowCommand_Execute_WithActiveKeys(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	ctx := setupLLMAPIShowTestRegistry(t)

	// Clear any existing environment variables and set controlled test data
	ctx.ClearAllTestEnvOverrides()
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "sk-1234567890abcdef1234567890abcdef")
	defer ctx.ClearAllTestEnvOverrides()

	// Register actual configuration service
	err := services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Set an active key
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	err = variableService.SetSystemVariable("#active_openai_key", "sk-1234567890abcdef1234567890abcdef")
	require.NoError(t, err)

	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "🟢 ACTIVE")
	assert.Contains(t, output, "Active keys:")
}

func TestLLMAPIShowCommand_maskAPIKey(t *testing.T) {
	cmd := &LLMAPIShowCommand{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal API key",
			input:    "sk-1234567890abcdef",
			expected: "sk-...def",
		},
		{
			name:     "short key (masked entirely)",
			input:    "short",
			expected: "*****",
		},
		{
			name:     "very short key",
			input:    "abc",
			expected: "***",
		},
		{
			name:     "exactly 6 chars",
			input:    "123456",
			expected: "******",
		},
		{
			name:     "longer key",
			input:    "sk-proj-1234567890abcdef1234567890abcdef1234567890",
			expected: "sk-...890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.maskAPIKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMAPIShowCommand_getKeyStatus(t *testing.T) {
	cmd := &LLMAPIShowCommand{}
	setupLLMAPIShowTestRegistry(t)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Store a test key
	err = variableService.Set("os.OPENAI_API_KEY", "test-key-value")
	require.NoError(t, err)

	tests := []struct {
		name           string
		provider       string
		varName        string
		activeKeyValue string
		setActiveKey   bool
		expectedStatus string
	}{
		{
			name:           "no active key set",
			provider:       "openai",
			varName:        "os.OPENAI_API_KEY",
			setActiveKey:   false,
			expectedStatus: "INACTIVE",
		},
		{
			name:           "active key matches",
			provider:       "openai",
			varName:        "os.OPENAI_API_KEY",
			activeKeyValue: "test-key-value",
			setActiveKey:   true,
			expectedStatus: "ACTIVE",
		},
		{
			name:           "active key doesn't match",
			provider:       "openai",
			varName:        "os.OPENAI_API_KEY",
			activeKeyValue: "different-key-value",
			setActiveKey:   true,
			expectedStatus: "INACTIVE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing active key
			_ = variableService.SetSystemVariable("#active_"+tt.provider+"_key", "")

			if tt.setActiveKey {
				err := variableService.SetSystemVariable("#active_"+tt.provider+"_key", tt.activeKeyValue)
				require.NoError(t, err)
			}

			status := cmd.getKeyStatus(tt.provider, tt.varName, variableService)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestLLMAPIShowCommand_Execute_ConfigurationServiceError(t *testing.T) {
	cmd := &LLMAPIShowCommand{}

	// Don't set up configuration service to simulate error
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	services.SetGlobalRegistry(services.NewRegistry())

	// Register only variable service, not configuration service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration service not available")
}

func TestLLMAPIShowCommand_Execute_VariableServiceError(t *testing.T) {
	cmd := &LLMAPIShowCommand{}

	// Set up context but not variable service to simulate error
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	ctx.ClearAllTestEnvOverrides()
	defer ctx.ClearAllTestEnvOverrides()

	services.SetGlobalRegistry(services.NewRegistry())

	// Register configuration service but not variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Don't register variable service to simulate error
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

// setupLLMAPIShowTestRegistry sets up a test environment with required services
func setupLLMAPIShowTestRegistry(t *testing.T) neurotypes.Context {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewMarkdownService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewThemeService())
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
var _ neurotypes.Command = (*LLMAPIShowCommand)(nil)
