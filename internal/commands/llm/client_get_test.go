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

func TestClientGetCommand_Name(t *testing.T) {
	cmd := &ClientGetCommand{}
	assert.Equal(t, "llm-client-get", cmd.Name())
}

func TestClientGetCommand_ParseMode(t *testing.T) {
	cmd := &ClientGetCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestClientGetCommand_Description(t *testing.T) {
	cmd := &ClientGetCommand{}
	assert.Equal(t, "Get or create LLM client for provider catalog ID", cmd.Description())
}

func TestClientGetCommand_Usage(t *testing.T) {
	cmd := &ClientGetCommand{}
	assert.Equal(t, "\\llm-client-get[key=api_key, provider_catalog_id=OAC|OAR|ORC|MSC|ANC|GMC] or \\llm-client-get (uses env vars)", cmd.Usage())
}

func TestClientGetCommand_HelpInfo(t *testing.T) {
	cmd := &ClientGetCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "llm-client-get", help.Command)
	assert.Equal(t, "Get or create LLM client for provider catalog ID", help.Description)
	assert.Equal(t, neurotypes.ParseModeKeyValue, help.ParseMode)

	// Check options
	require.Len(t, help.Options, 2)

	// Check provider_catalog_id option
	providerOption := help.Options[0]
	assert.Equal(t, "provider_catalog_id", providerOption.Name)
	assert.Equal(t, "LLM provider catalog ID (OAC, OAR, ORC, MSC, ANC, GMC)", providerOption.Description)
	assert.False(t, providerOption.Required)
	assert.Equal(t, "OAR", providerOption.Default)

	// Check key option
	keyOption := help.Options[1]
	assert.Equal(t, "key", keyOption.Name)
	assert.Equal(t, "API key for the provider (optional if environment variable is set)", keyOption.Description)
	assert.False(t, keyOption.Required)

	// Check examples
	assert.Len(t, help.Examples, 7)
	assert.Contains(t, help.Examples[0].Command, "llm-client-get[provider_catalog_id=OAR, key=sk-...]")
	assert.Contains(t, help.Examples[1].Command, "llm-client-get[provider_catalog_id=OAC, key=sk-...]")
	assert.Contains(t, help.Examples[2].Command, "llm-client-get[provider_catalog_id=ORC, key=sk-or-...]")
	assert.Contains(t, help.Examples[3].Command, "llm-client-get[provider_catalog_id=ANC, key=sk-ant-...]")
	assert.Contains(t, help.Examples[4].Command, "llm-client-get[provider_catalog_id=GMC, key=AIzaSy...]")
	assert.Contains(t, help.Examples[5].Command, "llm-client-get[key=${OPENAI_API_KEY}]")
	assert.Contains(t, help.Examples[6].Command, "llm-client-get")

	// Check notes
	assert.Greater(t, len(help.Notes), 0)
	assert.Contains(t, help.Notes[0], "Creates and caches LLM clients")
}

// TestClientGetCommand_Execute_Success tests removed - OpenAI functionality now comprehensively tested in openai_client_new_test.go
// This command delegates to specialized providers, so provider-specific testing belongs in their respective test files

// TestClientGetCommand_Execute_EnvironmentVariableSuccess tests removed - OpenAI functionality now comprehensively tested in openai_client_new_test.go
// Environment variable resolution for OpenAI is tested in the specialized command

// TestClientGetCommand_Execute_MissingAPIKey tests removed - OpenAI key resolution error handling now tested in openai_client_new_test.go
// Missing key scenarios are comprehensively covered in the specialized command tests

// TestClientGetCommand_Execute_TestEnvOverride test removed - OpenAI environment override behavior now tested in openai_client_new_test.go
// Custom environment variable testing is covered in the specialized command

func TestClientGetCommand_Execute_UnsupportedProvider(t *testing.T) {
	cmd := &ClientGetCommand{}
	setupLLMClientGetTestRegistry(t)

	args := map[string]string{
		"provider_catalog_id": "INVALID",
		"key":                 "test-key-123",
	}

	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider catalog ID 'INVALID'")
	assert.Contains(t, err.Error(), "Supported IDs:")
}

// TestClientGetCommand_Execute_ClientCaching test removed - OpenAI client caching behavior is inherent to the service layer
// and doesn't need specific testing at the delegation level since the specialized command handles all OpenAI scenarios

// TestClientGetCommand_Execute_DifferentAPIKeys test removed - OpenAI key differentiation is tested in openai_client_new_test.go
// Client ID generation and key-based differentiation is covered in the specialized command

func TestClientGetCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &ClientGetCommand{}

	// Don't set up test registry - this will cause service not available error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	args := map[string]string{
		"provider": "openai",
		"key":      "sk-test-key-123",
	}

	err := cmd.Execute(args, "")

	assert.Error(t, err)
	// Since the command checks variable service first, that's the error we get
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestClientGetCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &ClientGetCommand{}

	// Set up registry with no services to trigger variable service error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	ctx := context.New()
	context.SetGlobalContext(ctx)

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	args := map[string]string{
		"provider": "openai",
		"key":      "sk-test-key-123",
	}

	// Should fail due to missing variable service
	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

// setupLLMClientGetTestRegistry creates a clean test registry for llm-client-get command tests
func setupLLMClientGetTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}

// Benchmark tests
func BenchmarkClientGetCommand_Execute(b *testing.B) {
	cmd := &ClientGetCommand{}

	// Set up test registry for benchmarking
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	ctx := context.New()
	context.SetGlobalContext(ctx)

	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	if err != nil {
		b.Fatal(err)
	}
	err = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	if err != nil {
		b.Fatal(err)
	}
	err = services.GetGlobalRegistry().InitializeAll()
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	}()

	args := map[string]string{
		"provider": "openai",
		"key":      "sk-benchmark-key-123456789",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Suppress output for benchmark
		_ = stringprocessing.CaptureOutput(func() {
			_ = cmd.Execute(args, "")
		})
	}
}
