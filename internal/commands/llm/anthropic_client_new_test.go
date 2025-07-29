// Package llm contains tests for LLM-related commands.
package llm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestAnthropicClientNewCommand_Name(t *testing.T) {
	cmd := &AnthropicClientNewCommand{}
	assert.Equal(t, "anthropic-client-new", cmd.Name())
}

func TestAnthropicClientNewCommand_ParseMode(t *testing.T) {
	cmd := &AnthropicClientNewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestAnthropicClientNewCommand_Description(t *testing.T) {
	cmd := &AnthropicClientNewCommand{}
	assert.Equal(t, "Create new Anthropic client with automatic key resolution and extended thinking support", cmd.Description())
}

func TestAnthropicClientNewCommand_Usage(t *testing.T) {
	cmd := &AnthropicClientNewCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "anthropic-client-new")
	assert.Contains(t, usage, "key=api_key")
	assert.Contains(t, usage, "uses active key")
}

func TestAnthropicClientNewCommand_HelpInfo(t *testing.T) {
	cmd := &AnthropicClientNewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "anthropic-client-new", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.Usage(), helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Len(t, helpInfo.Options, 1)
	keyOption := helpInfo.Options[0]
	assert.Equal(t, "key", keyOption.Name)
	assert.Equal(t, "string", keyOption.Type)
	assert.False(t, keyOption.Required)
	assert.Contains(t, keyOption.Description, "Anthropic API key")

	// Check examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.Len(t, helpInfo.Examples, 3)

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	expectedVars := []string{"_client_id", "_output", "#client_provider", "#client_configured", "#client_thinking_support"}
	actualVars := make(map[string]bool)
	for _, storedVar := range helpInfo.StoredVariables {
		actualVars[storedVar.Name] = true
	}
	for _, expectedVar := range expectedVars {
		assert.True(t, actualVars[expectedVar], "Expected stored variable %s not found", expectedVar)
	}

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
	assert.Contains(t, helpInfo.Notes[0], "Key resolution priority")
	// Look for extended thinking mention in any of the notes
	foundExtendedThinking := false
	for _, note := range helpInfo.Notes {
		if strings.Contains(note, "extended thinking") {
			foundExtendedThinking = true
			break
		}
	}
	assert.True(t, foundExtendedThinking, "Should mention extended thinking in notes")
}

func TestAnthropicClientNewCommand_Execute_WithKeyParameter(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	cmd := &AnthropicClientNewCommand{}
	testKey := "sk-ant-api03-test123"

	// Execute with explicit key
	err = cmd.Execute(map[string]string{"key": testKey}, "")
	assert.NoError(t, err)

	// Verify system variables were set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "anthropic:")

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Anthropic client ready")

	provider, err := variableService.Get("#client_provider")
	assert.NoError(t, err)
	assert.Equal(t, "anthropic", provider)

	configured, err := variableService.Get("#client_configured")
	assert.NoError(t, err)
	assert.Equal(t, "true", configured)

	thinkingSupport, err := variableService.Get("#client_thinking_support")
	assert.NoError(t, err)
	assert.Equal(t, "true", thinkingSupport)
}

func TestAnthropicClientNewCommand_Execute_WithActiveKey(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Set active anthropic key
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	testKey := "sk-ant-api03-active123"
	err = variableService.SetSystemVariable("#active_anthropic_key", testKey)
	require.NoError(t, err)

	cmd := &AnthropicClientNewCommand{}

	// Execute without explicit key (should use active key)
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Verify client was created
	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "anthropic:")
}

func TestAnthropicClientNewCommand_Execute_WithEnvironmentKey(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Set environment variable using test context override
	testKey := "sk-ant-api03-env123"
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", testKey)

	cmd := &AnthropicClientNewCommand{}

	// Execute without explicit key or active key (should use env var)
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Verify client was created
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "anthropic:")
}

func TestAnthropicClientNewCommand_Execute_NoKeyAvailable(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Override the default test env value to be empty to simulate no key available
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", "")

	cmd := &AnthropicClientNewCommand{}

	// Execute without any key available
	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key found")
	assert.Contains(t, err.Error(), "anthropic-client-new[key=")
	assert.Contains(t, err.Error(), "llm-api-activate[provider=anthropic")
	assert.Contains(t, err.Error(), "ANTHROPIC_API_KEY")
}

func TestAnthropicClientNewCommand_Execute_VariableServiceError(t *testing.T) {
	// Create empty registry without services
	originalRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	t.Cleanup(func() {
		services.SetGlobalRegistry(originalRegistry)
		context.ResetGlobalContext()
	})

	cmd := &AnthropicClientNewCommand{}

	// Execute should fail when variable service is not available
	err := cmd.Execute(map[string]string{"key": "test"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestAnthropicClientNewCommand_Execute_ClientFactoryServiceError(t *testing.T) {
	// Initialize test environment with only variable service
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register only variable service (no client factory)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	cmd := &AnthropicClientNewCommand{}

	// Execute should fail when client factory service is not available
	err = cmd.Execute(map[string]string{"key": "test"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client factory service not available")
}

func TestAnthropicClientNewCommand_ResolveAPIKey_Priority(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Set up all possible key sources
	paramKey := "sk-ant-param"
	activeKey := "sk-ant-active"
	envKey := "sk-ant-env"

	err = variableService.SetSystemVariable("#active_anthropic_key", activeKey)
	require.NoError(t, err)
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", envKey)

	cmd := &AnthropicClientNewCommand{}

	// Test priority 1: Parameter key should win
	resolvedKey, err := cmd.resolveAPIKey(map[string]string{"key": paramKey}, variableService)
	assert.NoError(t, err)
	assert.Equal(t, paramKey, resolvedKey)

	// Test priority 2: Active key should win when no parameter
	resolvedKey, err = cmd.resolveAPIKey(map[string]string{}, variableService)
	assert.NoError(t, err)
	assert.Equal(t, activeKey, resolvedKey)

	// Test priority 3: Environment key should win when no parameter or active key
	err = variableService.SetSystemVariable("#active_anthropic_key", "")
	require.NoError(t, err)
	resolvedKey, err = cmd.resolveAPIKey(map[string]string{}, variableService)
	assert.NoError(t, err)
	assert.Equal(t, envKey, resolvedKey)

	// Test no key available - clear active key and env override
	err = variableService.SetSystemVariable("#active_anthropic_key", "")
	require.NoError(t, err)
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", "")
	_, err = cmd.resolveAPIKey(map[string]string{}, variableService)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key found")
}
