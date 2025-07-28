// Package llm contains tests for LLM-related commands.
package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestGeminiClientNewCommand_Name(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	assert.Equal(t, "gemini-client-new", cmd.Name())
}

func TestGeminiClientNewCommand_ParseMode(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestGeminiClientNewCommand_Description(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	assert.Equal(t, "Create new Gemini client with automatic key resolution", cmd.Description())
}

func TestGeminiClientNewCommand_Usage(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "gemini-client-new")
	assert.Contains(t, usage, "key=api_key")
	assert.Contains(t, usage, "uses active key")
}

func TestGeminiClientNewCommand_HelpInfo(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "gemini-client-new", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.Usage(), helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Len(t, helpInfo.Options, 1)
	keyOption := helpInfo.Options[0]
	assert.Equal(t, "key", keyOption.Name)
	assert.False(t, keyOption.Required)
	assert.Equal(t, "string", keyOption.Type)

	// Check examples
	assert.Len(t, helpInfo.Examples, 3)
	assert.Contains(t, helpInfo.Examples[0].Command, "key=AIzaSy")
	assert.Contains(t, helpInfo.Examples[1].Command, "gemini-client-new")
	assert.Contains(t, helpInfo.Examples[2].Command, "${GOOGLE_API_KEY}")

	// Check stored variables
	assert.Len(t, helpInfo.StoredVariables, 4)

	// Check notes
	assert.Len(t, helpInfo.Notes, 4)
	assert.Contains(t, helpInfo.Notes[0], "Key resolution priority")
	assert.Contains(t, helpInfo.Notes[1], "llm-api-activate")
}

func TestGeminiClientNewCommand_Execute_WithExplicitKey(t *testing.T) {
	cmd := &GeminiClientNewCommand{}

	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "explicit Gemini API key",
			apiKey: "AIzaSyDmZGkEjF1D5kXvQrj3f9mH0pL2nCvW4xY",
		},
		{
			name:   "short API key",
			apiKey: "test-key-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGeminiClientNewTestRegistry(t)

			args := map[string]string{
				"key": tt.apiKey,
			}

			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(args, "")
			})

			assert.NoError(t, err)
			assert.Contains(t, outputStr, "Gemini client ready: gemini:")
			assert.Contains(t, outputStr, "(configured: true)")

			// Verify system variables were set
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			clientID, err := variableService.Get("_client_id")
			assert.NoError(t, err)
			assert.Contains(t, clientID, "gemini:")

			output, err := variableService.Get("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Gemini client ready:")

			provider, err := variableService.Get("#client_provider")
			assert.NoError(t, err)
			assert.Equal(t, "gemini", provider)

			configured, err := variableService.Get("#client_configured")
			assert.NoError(t, err)
			assert.Equal(t, "true", configured)

			cacheCount, err := variableService.Get("#client_cache_count")
			assert.NoError(t, err)
			assert.NotEmpty(t, cacheCount)
		})
	}
}

func TestGeminiClientNewCommand_Execute_WithActiveKey(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	setupGeminiClientNewTestRegistry(t)

	// Set active Gemini key
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	activeKey := "AIzaSyActiveKey1234567890abcdef"
	err = variableService.SetSystemVariable("#active_gemini_key", activeKey)
	assert.NoError(t, err)

	// Execute command without explicit key (should use active key)
	args := map[string]string{}
	var execErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		execErr = cmd.Execute(args, "")
	})

	assert.NoError(t, execErr)
	assert.Contains(t, outputStr, "Gemini client ready:")

	// Verify client was created successfully
	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "gemini:")

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Gemini client ready:")
}

func TestGeminiClientNewCommand_Execute_WithEnvironmentKey(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	setupGeminiClientNewTestRegistry(t)

	// Set up test context and environment variable
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	envKey := "AIzaSyEnvKey1234567890abcdef"
	ctx.SetTestEnvOverride("GOOGLE_API_KEY", envKey)

	defer func() {
		ctx.ClearTestEnvOverride("GOOGLE_API_KEY")
	}()

	// Execute command without explicit key or active key (should use env var)
	args := map[string]string{}
	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "")
	})

	assert.NoError(t, err)
	assert.Contains(t, outputStr, "Gemini client ready:")

	// Verify client was created successfully
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "gemini:")
}

func TestGeminiClientNewCommand_Execute_NoKeyAvailable(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	setupGeminiClientNewTestRegistry(t)

	// Set up test context with empty environment variables
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)
	ctx.SetTestEnvOverride("GOOGLE_API_KEY", "")

	defer func() {
		ctx.ClearTestEnvOverride("GOOGLE_API_KEY")
	}()

	// Execute command with no key sources available
	args := map[string]string{}
	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key found")
	assert.Contains(t, err.Error(), "gemini-client-new[key=")
	assert.Contains(t, err.Error(), "llm-api-activate")
	assert.Contains(t, err.Error(), "GOOGLE_API_KEY")
}

func TestGeminiClientNewCommand_resolveAPIKey(t *testing.T) {
	cmd := &GeminiClientNewCommand{}
	setupGeminiClientNewTestRegistry(t)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Set up test context
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	tests := []struct {
		name        string
		args        map[string]string
		activeKey   string
		envKey      string
		expectedKey string
		expectError bool
	}{
		{
			name:        "explicit key takes priority",
			args:        map[string]string{"key": "explicit-123"},
			activeKey:   "active-456",
			envKey:      "env-789",
			expectedKey: "explicit-123",
			expectError: false,
		},
		{
			name:        "active key used when no explicit key",
			args:        map[string]string{},
			activeKey:   "active-456",
			envKey:      "env-789",
			expectedKey: "active-456",
			expectError: false,
		},
		{
			name:        "env key used when no explicit or active key",
			args:        map[string]string{},
			activeKey:   "",
			envKey:      "env-789",
			expectedKey: "env-789",
			expectError: false,
		},
		{
			name:        "error when no keys available",
			args:        map[string]string{},
			activeKey:   "",
			envKey:      "",
			expectedKey: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test keys
			if tt.activeKey != "" {
				err := variableService.SetSystemVariable("#active_gemini_key", tt.activeKey)
				assert.NoError(t, err)
			} else {
				_ = variableService.SetSystemVariable("#active_gemini_key", "")
			}

			if tt.envKey != "" {
				ctx.SetTestEnvOverride("GOOGLE_API_KEY", tt.envKey)
			} else {
				ctx.SetTestEnvOverride("GOOGLE_API_KEY", "")
			}

			// Test key resolution
			key, err := cmd.resolveAPIKey(tt.args, variableService)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no API key found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, key)
			}

			// Clean up
			ctx.ClearTestEnvOverride("GOOGLE_API_KEY")
		})
	}
}

// setupGeminiClientNewTestRegistry creates a clean test registry for gemini-client-new command tests
func setupGeminiClientNewTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
	})
}

// Interface compliance check
var _ neurotypes.Command = (*GeminiClientNewCommand)(nil)
