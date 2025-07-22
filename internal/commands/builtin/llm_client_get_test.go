package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestLLMClientGetCommand_Name(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	assert.Equal(t, "llm-client-get", cmd.Name())
}

func TestLLMClientGetCommand_ParseMode(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestLLMClientGetCommand_Description(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	assert.Equal(t, "Get or create LLM client for provider", cmd.Description())
}

func TestLLMClientGetCommand_Usage(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	assert.Equal(t, "\\llm-client-get[key=api_key, provider=openai] or \\llm-client-get (uses env vars)", cmd.Usage())
}

func TestLLMClientGetCommand_HelpInfo(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "llm-client-get", help.Command)
	assert.Equal(t, "Get or create LLM client for provider", help.Description)
	assert.Equal(t, neurotypes.ParseModeKeyValue, help.ParseMode)

	// Check options
	require.Len(t, help.Options, 2)

	// Check provider option
	providerOption := help.Options[0]
	assert.Equal(t, "provider", providerOption.Name)
	assert.Equal(t, "LLM provider name (openai, anthropic)", providerOption.Description)
	assert.False(t, providerOption.Required)
	assert.Equal(t, "openai", providerOption.Default)

	// Check key option
	keyOption := help.Options[1]
	assert.Equal(t, "key", keyOption.Name)
	assert.Equal(t, "API key for the provider (optional if environment variable is set)", keyOption.Description)
	assert.False(t, keyOption.Required)

	// Check examples
	assert.Len(t, help.Examples, 3)
	assert.Contains(t, help.Examples[0].Command, "llm-client-get[provider=openai, key=sk-...]")
	assert.Contains(t, help.Examples[1].Command, "llm-client-get[key=${OPENAI_API_KEY}]")
	assert.Contains(t, help.Examples[2].Command, "llm-client-get")

	// Check notes
	assert.Greater(t, len(help.Notes), 0)
	assert.Contains(t, help.Notes[0], "Creates and caches LLM clients")
}

func TestLLMClientGetCommand_Execute_Success(t *testing.T) {
	cmd := &LLMClientGetCommand{}

	tests := []struct {
		name             string
		args             map[string]string
		expectedProvider string
		expectedOutput   string
	}{
		{
			name: "openai provider with explicit key",
			args: map[string]string{
				"provider": "openai",
				"key":      "sk-test-fake-key-123456789",
			},
			expectedProvider: "openai",
			expectedOutput:   "LLM client ready: openai:sk-test-**** (configured: true)",
		},
		{
			name: "default openai provider",
			args: map[string]string{
				"key": "sk-another-test-key-987654321",
			},
			expectedProvider: "openai",
			expectedOutput:   "LLM client ready: openai:sk-anoth**** (configured: true)",
		},
		{
			name: "short api key",
			args: map[string]string{
				"provider": "openai",
				"key":      "sk-123",
			},
			expectedProvider: "openai",
			expectedOutput:   "LLM client ready: openai:**** (configured: true)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupLLMClientGetTestRegistry(t)

			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(tt.args, "")
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput+"\n", outputStr)

			// Verify system variables were set
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			// Check _client_id
			clientID, err := variableService.Get("_client_id")
			assert.NoError(t, err)
			assert.Contains(t, clientID, tt.expectedProvider)

			// Check _output
			output, err := variableService.Get("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "LLM client ready")

			// Check metadata variables
			provider, err := variableService.Get("#client_provider")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedProvider, provider)

			configured, err := variableService.Get("#client_configured")
			assert.NoError(t, err)
			assert.Equal(t, "true", configured)

			cacheCount, err := variableService.Get("#client_cache_count")
			assert.NoError(t, err)
			assert.NotEmpty(t, cacheCount)
		})
	}
}

func TestLLMClientGetCommand_Execute_EnvironmentVariableSuccess(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	setupLLMClientGetTestRegistry(t)

	// Set the context to test mode to get predictable environment variables
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	tests := []struct {
		name             string
		args             map[string]string
		expectedProvider string
		expectedOutput   string
	}{
		{
			name:             "openai with environment variable",
			args:             map[string]string{"provider": "openai"},
			expectedProvider: "openai",
			expectedOutput:   "LLM client ready: openai:test-ope**** (configured: true)",
		},
		{
			name:             "default provider with environment variable",
			args:             map[string]string{},
			expectedProvider: "openai",
			expectedOutput:   "LLM client ready: openai:test-ope**** (configured: true)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(tt.args, "")
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput+"\n", outputStr)

			// Verify system variables were set
			variableService, err := services.GetGlobalVariableService()
			require.NoError(t, err)

			// Check _client_id
			clientID, err := variableService.Get("_client_id")
			assert.NoError(t, err)
			assert.Contains(t, clientID, tt.expectedProvider)

			// Check _output
			output, err := variableService.Get("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "LLM client ready")

			// Check metadata variables
			provider, err := variableService.Get("#client_provider")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedProvider, provider)

			configured, err := variableService.Get("#client_configured")
			assert.NoError(t, err)
			assert.Equal(t, "true", configured)
		})
	}
}

func TestLLMClientGetCommand_Execute_MissingAPIKey(t *testing.T) {
	t.Skip("Skipping test that requires empty environment - will address in future")
	cmd := &LLMClientGetCommand{}

	// Set up a custom test context that returns empty environment variables
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a custom context that returns empty strings for environment variables
	ctx := context.New()
	ctx.SetTestMode(false) // We want to control env vars ourselves
	context.SetGlobalContext(ctx)

	// Register services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "no key parameter and no environment variable",
			args: map[string]string{"provider": "openai"},
		},
		{
			name: "empty key parameter and no environment variable",
			args: map[string]string{
				"provider": "openai",
				"key":      "",
			},
		},
		{
			name: "only empty key and no environment variable",
			args: map[string]string{"key": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, "")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "API key not found")
			assert.Contains(t, err.Error(), "Usage:")
			assert.Contains(t, err.Error(), "OPENAI_API_KEY")
		})
	}
}

func TestLLMClientGetCommand_Execute_UnsupportedProvider(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	setupLLMClientGetTestRegistry(t)

	args := map[string]string{
		"provider": "unsupported-provider",
		"key":      "test-key-123",
	}

	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get client for provider unsupported-provider")
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestLLMClientGetCommand_Execute_ClientCaching(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	setupLLMClientGetTestRegistry(t)

	args := map[string]string{
		"provider": "openai",
		"key":      "sk-test-caching-key-123",
	}

	// First call - should create client
	var err1 error
	_ = stringprocessing.CaptureOutput(func() {
		err1 = cmd.Execute(args, "")
	})
	assert.NoError(t, err1)

	// Get initial cache count
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	initialCount, err := variableService.Get("#client_cache_count")
	assert.NoError(t, err)

	// Second call with same parameters - should use cached client
	var err2 error
	_ = stringprocessing.CaptureOutput(func() {
		err2 = cmd.Execute(args, "")
	})
	assert.NoError(t, err2)

	// Cache count should be the same (client was reused)
	finalCount, err := variableService.Get("#client_cache_count")
	assert.NoError(t, err)
	assert.Equal(t, initialCount, finalCount)
}

func TestLLMClientGetCommand_Execute_DifferentAPIKeys(t *testing.T) {
	cmd := &LLMClientGetCommand{}
	setupLLMClientGetTestRegistry(t)

	// First client with one API key
	args1 := map[string]string{
		"provider": "openai",
		"key":      "sk-first-key-123",
	}

	var err1 error
	output1 := stringprocessing.CaptureOutput(func() {
		err1 = cmd.Execute(args1, "")
	})
	assert.NoError(t, err1)
	assert.Contains(t, output1, "sk-first")

	// Second client with different API key
	args2 := map[string]string{
		"provider": "openai",
		"key":      "sk-second-key-456",
	}

	var err2 error
	output2 := stringprocessing.CaptureOutput(func() {
		err2 = cmd.Execute(args2, "")
	})
	assert.NoError(t, err2)
	assert.Contains(t, output2, "sk-secon")

	// Verify different client IDs were generated
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	clientID, err := variableService.Get("_client_id")
	assert.NoError(t, err)
	assert.Contains(t, clientID, "sk-secon") // Should have the latest client ID
}

func TestLLMClientGetCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &LLMClientGetCommand{}

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

func TestLLMClientGetCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &LLMClientGetCommand{}

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

func TestLLMClientGetCommand_TruncateAPIKey(t *testing.T) {
	cmd := &LLMClientGetCommand{}

	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "normal length key",
			apiKey:   "sk-test-fake-key-123456789",
			expected: "sk-test-****",
		},
		{
			name:     "short key",
			apiKey:   "sk-123",
			expected: "****",
		},
		{
			name:     "exactly 8 characters",
			apiKey:   "sk-12345",
			expected: "****",
		},
		{
			name:     "9 characters",
			apiKey:   "sk-123456",
			expected: "sk-12345****",
		},
		{
			name:     "empty key",
			apiKey:   "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.truncateAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
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

	err = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
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
func BenchmarkLLMClientGetCommand_Execute(b *testing.B) {
	cmd := &LLMClientGetCommand{}

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

func BenchmarkLLMClientGetCommand_TruncateAPIKey(b *testing.B) {
	cmd := &LLMClientGetCommand{}
	apiKey := "sk-very-long-test-api-key-for-benchmarking-purposes-123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.truncateAPIKey(apiKey)
	}
}
