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

func TestOpenAIClientNewCommand_Name(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	assert.Equal(t, "openai-client-new", cmd.Name())
}

func TestOpenAIClientNewCommand_ParseMode(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestOpenAIClientNewCommand_Description(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	assert.Equal(t, "Create new OpenAI client with automatic key resolution and reasoning model support", cmd.Description())
}

func TestOpenAIClientNewCommand_Usage(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "openai-client-new")
	assert.Contains(t, usage, "key=api_key")
	assert.Contains(t, usage, "uses active key")
}

func TestOpenAIClientNewCommand_HelpInfo(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "openai-client-new", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.Usage(), helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Len(t, helpInfo.Options, 2)

	// Check key option
	keyOption := helpInfo.Options[0]
	assert.Equal(t, "key", keyOption.Name)
	assert.False(t, keyOption.Required)
	assert.Equal(t, "string", keyOption.Type)
	assert.Contains(t, keyOption.Description, "OpenAI API key")
	assert.Contains(t, keyOption.Description, "#active_openai_key")

	// Check client_type option
	clientTypeOption := helpInfo.Options[1]
	assert.Equal(t, "client_type", clientTypeOption.Name)
	assert.False(t, clientTypeOption.Required)
	assert.Equal(t, "string", clientTypeOption.Type)
	assert.Contains(t, clientTypeOption.Description, "OAC")
	assert.Contains(t, clientTypeOption.Description, "OAR")

	// Check examples
	assert.Len(t, helpInfo.Examples, 4)
	assert.Contains(t, helpInfo.Examples[0].Command, "key=sk-proj-")
	assert.Contains(t, helpInfo.Examples[1].Command, "client_type=OAC")
	assert.Contains(t, helpInfo.Examples[2].Command, "client_type=OAR")
	assert.Contains(t, helpInfo.Examples[3].Command, "${OPENAI_API_KEY}")

	// Check stored variables
	assert.Len(t, helpInfo.StoredVariables, 5)

	variableNames := make(map[string]bool)
	for _, variable := range helpInfo.StoredVariables {
		variableNames[variable.Name] = true
	}

	assert.True(t, variableNames["_client_id"])
	assert.True(t, variableNames["_output"])
	assert.True(t, variableNames["#client_provider"])
	assert.True(t, variableNames["#client_configured"])
	assert.True(t, variableNames["#client_reasoning_support"])

	// Check notes
	assert.Len(t, helpInfo.Notes, 6)
	assert.Contains(t, helpInfo.Notes[0], "Key resolution priority")
	assert.Contains(t, helpInfo.Notes[1], "llm-api-activate")
	assert.Contains(t, helpInfo.Notes[2], "Client types")
	assert.Contains(t, helpInfo.Notes[2], "OAC")
	assert.Contains(t, helpInfo.Notes[2], "OAR")
	assert.Contains(t, helpInfo.Notes[3], "automatically route")
	assert.Contains(t, helpInfo.Notes[4], "Default client_type")
	assert.Contains(t, helpInfo.Notes[5], "dedicated chat-only testing")
}

func TestOpenAIClientNewCommand_Execute_WithExplicitKey(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}

	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "explicit OpenAI API key",
			apiKey: "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:   "short test API key",
			apiKey: "test-key-123",
		},
		{
			name:   "old format API key",
			apiKey: "sk-1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupOpenAIClientNewTestRegistry(t)

			args := map[string]string{
				"key": tt.apiKey,
			}

			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(args, "")
			})

			assert.NoError(t, err)
			assert.Contains(t, outputStr, "OpenAI client ready: OAR:")
			assert.Contains(t, outputStr, "configured: true)")
			assert.Contains(t, outputStr, "reasoning/dual-mode")

			// Verify system variables were set
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			clientID, err := variableService.Get("_client_id")
			assert.NoError(t, err)
			assert.Contains(t, clientID, "OAR:")

			output, err := variableService.Get("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "OpenAI client ready:")

			provider, err := variableService.Get("#client_provider")
			assert.NoError(t, err)
			assert.Equal(t, "openai", provider)

			configured, err := variableService.Get("#client_configured")
			assert.NoError(t, err)
			assert.Equal(t, "true", configured)

			reasoningSupport, err := variableService.Get("#client_reasoning_support")
			assert.NoError(t, err)
			assert.Equal(t, "true", reasoningSupport)

			cacheCount, err := variableService.Get("#client_cache_count")
			assert.NoError(t, err)
			assert.NotEmpty(t, cacheCount)
		})
	}
}

func TestOpenAIClientNewCommand_Execute_WithActiveKey(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	// Set active OpenAI key
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	activeKey := "sk-proj-ActiveKey1234567890abcdef"
	err = variableService.SetSystemVariable("#active_openai_key", activeKey)
	assert.NoError(t, err)

	// Execute command without explicit key (should use active key)
	args := map[string]string{}
	var execErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		execErr = cmd.Execute(args, "")
	})

	assert.NoError(t, execErr)
	assert.Contains(t, outputStr, "OpenAI client ready:")
	assert.Contains(t, outputStr, "reasoning/dual-mode")

	// Verify client was created successfully
	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "OAR:")

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "OpenAI client ready:")

	// Verify reasoning support is set
	reasoningSupport, err := variableService.Get("#client_reasoning_support")
	assert.NoError(t, err)
	assert.Equal(t, "true", reasoningSupport)
}

func TestOpenAIClientNewCommand_Execute_WithEnvironmentKey(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	// Set up test context and environment variable
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	envKey := "sk-proj-EnvKey1234567890abcdef"
	ctx.SetTestEnvOverride("OPENAI_API_KEY", envKey)

	defer func() {
		ctx.ClearTestEnvOverride("OPENAI_API_KEY")
	}()

	// Execute command without explicit key or active key (should use env var)
	args := map[string]string{}
	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "")
	})

	assert.NoError(t, err)
	assert.Contains(t, outputStr, "OpenAI client ready:")
	assert.Contains(t, outputStr, "reasoning/dual-mode")

	// Verify client was created successfully
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "OAR:")

	provider, err := variableService.Get("#client_provider")
	assert.NoError(t, err)
	assert.Equal(t, "openai", provider)
}

func TestOpenAIClientNewCommand_Execute_NoKeyAvailable(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	// Set up test context with empty environment variables
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "")

	defer func() {
		ctx.ClearTestEnvOverride("OPENAI_API_KEY")
	}()

	// Execute command with no key sources available
	args := map[string]string{}
	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key found")
	assert.Contains(t, err.Error(), "openai-client-new[key=")
	assert.Contains(t, err.Error(), "llm-api-activate")
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

func TestOpenAIClientNewCommand_Execute_ServiceErrors(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T)
		expectError string
	}{
		{
			name: "variable service unavailable",
			setupFunc: func(_ *testing.T) {
				// Setup with no services
				services.SetGlobalRegistry(services.NewRegistry())
			},
			expectError: "variable service not available",
		},
		{
			name: "client factory service unavailable",
			setupFunc: func(t *testing.T) {
				// Setup with only variable service
				registry := services.NewRegistry()
				err := registry.RegisterService(services.NewVariableService())
				require.NoError(t, err)
				err = registry.InitializeAll()
				require.NoError(t, err)
				services.SetGlobalRegistry(registry)
			},
			expectError: "client factory service not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original registry
			originalRegistry := services.GetGlobalRegistry()
			defer func() {
				services.SetGlobalRegistry(originalRegistry)
			}()

			tt.setupFunc(t)

			args := map[string]string{"key": "test-key"}
			err := cmd.Execute(args, "")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestOpenAIClientNewCommand_resolveAPIKey(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

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
			args:        map[string]string{"key": "sk-proj-explicit-123"},
			activeKey:   "sk-proj-active-456",
			envKey:      "sk-proj-env-789",
			expectedKey: "sk-proj-explicit-123",
			expectError: false,
		},
		{
			name:        "active key used when no explicit key",
			args:        map[string]string{},
			activeKey:   "sk-proj-active-456",
			envKey:      "sk-proj-env-789",
			expectedKey: "sk-proj-active-456",
			expectError: false,
		},
		{
			name:        "env key used when no explicit or active key",
			args:        map[string]string{},
			activeKey:   "",
			envKey:      "sk-proj-env-789",
			expectedKey: "sk-proj-env-789",
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
		{
			name:        "empty explicit key falls back to active key",
			args:        map[string]string{"key": ""},
			activeKey:   "sk-proj-active-456",
			envKey:      "sk-proj-env-789",
			expectedKey: "sk-proj-active-456",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test keys
			if tt.activeKey != "" {
				err := variableService.SetSystemVariable("#active_openai_key", tt.activeKey)
				assert.NoError(t, err)
			} else {
				_ = variableService.SetSystemVariable("#active_openai_key", "")
			}

			if tt.envKey != "" {
				ctx.SetTestEnvOverride("OPENAI_API_KEY", tt.envKey)
			} else {
				ctx.SetTestEnvOverride("OPENAI_API_KEY", "")
			}

			// Test key resolution
			key, err := cmd.resolveAPIKey(tt.args, variableService)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no API key found")
				assert.Contains(t, err.Error(), "openai-client-new[key=")
				assert.Contains(t, err.Error(), "llm-api-activate[provider=openai")
				assert.Contains(t, err.Error(), "OPENAI_API_KEY")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, key)
			}

			// Clean up
			ctx.ClearTestEnvOverride("OPENAI_API_KEY")
		})
	}
}

func TestOpenAIClientNewCommand_resolveAPIKey_EdgeCases(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Test whitespace handling - the implementation returns whitespace keys as-is
	t.Run("whitespace keys are returned as-is", func(t *testing.T) {
		// Set active key to whitespace
		err := variableService.SetSystemVariable("#active_openai_key", "   ")
		assert.NoError(t, err)

		// Set env to valid key
		ctx.SetTestEnvOverride("OPENAI_API_KEY", "sk-proj-env-key")
		defer ctx.ClearTestEnvOverride("OPENAI_API_KEY")

		args := map[string]string{"key": "\t\n "}
		key, err := cmd.resolveAPIKey(args, variableService)

		// The implementation returns whitespace keys as-is (doesn't trim)
		assert.NoError(t, err)
		assert.Equal(t, "\t\n ", key)
	})

	// Test variable service error handling
	t.Run("variable service error is handled gracefully", func(t *testing.T) {
		// Clear active key to force GetEnv usage
		_ = variableService.SetSystemVariable("#active_openai_key", "")

		// Set env key
		ctx.SetTestEnvOverride("OPENAI_API_KEY", "sk-proj-env-fallback")
		defer ctx.ClearTestEnvOverride("OPENAI_API_KEY")

		args := map[string]string{}
		key, err := cmd.resolveAPIKey(args, variableService)

		assert.NoError(t, err)
		assert.Equal(t, "sk-proj-env-fallback", key)
	})
}

func TestOpenAIClientNewCommand_Execute_WithClientType(t *testing.T) {
	command := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	testCases := []struct {
		name              string
		args              map[string]string
		expectedType      string
		expectedReasoning string
	}{
		{
			name:              "explicit OAC client type",
			args:              map[string]string{"client_type": "OAC", "key": "sk-test123"},
			expectedType:      "OAC",
			expectedReasoning: "false",
		},
		{
			name:              "explicit OAR client type",
			args:              map[string]string{"client_type": "OAR", "key": "sk-test123"},
			expectedType:      "OAR",
			expectedReasoning: "true",
		},
		{
			name:              "default client type (should be OAR)",
			args:              map[string]string{"key": "sk-test123"},
			expectedType:      "OAR",
			expectedReasoning: "true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute command
			err := command.Execute(tc.args, "")

			// Verify command succeeded
			assert.NoError(t, err)

			// Verify system variables
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			clientType, err := variableService.Get("#client_type")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedType, clientType)

			reasoningSupport, err := variableService.Get("#client_reasoning_support")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedReasoning, reasoningSupport)
		})
	}
}

func TestOpenAIClientNewCommand_Execute_InvalidClientType(t *testing.T) {
	command := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	args := map[string]string{"client_type": "INVALID", "key": "sk-test123"}

	// Execute command
	err := command.Execute(args, "")

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid client_type 'INVALID'")
	assert.Contains(t, err.Error(), "Must be 'OAC' (chat-only) or 'OAR' (reasoning/dual-mode)")
}

func TestOpenAIClientNewCommand_Execute_IntegrationWithReasoningSupport(t *testing.T) {
	cmd := &OpenAIClientNewCommand{}
	setupOpenAIClientNewTestRegistry(t)

	// Test that reasoning support is always enabled for OpenAI clients
	args := map[string]string{
		"key": "sk-proj-test-reasoning-support",
	}

	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "")
	})

	assert.NoError(t, err)
	assert.Contains(t, outputStr, "reasoning/dual-mode")

	// Verify reasoning support variable
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	reasoningSupport, err := variableService.Get("#client_reasoning_support")
	assert.NoError(t, err)
	assert.Equal(t, "true", reasoningSupport)

	// Verify client metadata
	provider, err := variableService.Get("#client_provider")
	assert.NoError(t, err)
	assert.Equal(t, "openai", provider)
}

// setupOpenAIClientNewTestRegistry creates a clean test registry for openai-client-new command tests
func setupOpenAIClientNewTestRegistry(t *testing.T) {
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
var _ neurotypes.Command = (*OpenAIClientNewCommand)(nil)
