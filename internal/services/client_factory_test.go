package services

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestClientFactoryService_Name(t *testing.T) {
	service := NewClientFactoryService()
	assert.Equal(t, "client_factory", service.Name())
}

func TestClientFactoryService_Initialize(t *testing.T) {
	service := NewClientFactoryService()

	// Test successful initialization
	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Test duplicate initialization is idempotent
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

func TestClientFactoryService_GetClientForProvider(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		apiKey      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful openai client creation",
			provider:    "openai",
			apiKey:      "sk-test-key",
			expectError: false,
		},
		{
			name:        "empty provider",
			provider:    "",
			apiKey:      "sk-test-key",
			expectError: true,
			errorMsg:    "provider cannot be empty",
		},
		{
			name:        "empty api key",
			provider:    "openai",
			apiKey:      "",
			expectError: true,
			errorMsg:    "API key cannot be empty for provider 'openai'",
		},
		{
			name:        "unsupported provider",
			provider:    "unsupported",
			apiKey:      "test-key",
			expectError: true,
			errorMsg:    "unsupported provider 'unsupported'. Supported providers: openai, anthropic",
		},
		{
			name:        "anthropic provider not yet implemented",
			provider:    "anthropic",
			apiKey:      "test-key",
			expectError: true,
			errorMsg:    "anthropic provider is not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewClientFactoryService()
			err := service.Initialize()
			require.NoError(t, err)

			client, err := service.GetClientForProvider(tt.provider, tt.apiKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.provider, client.GetProviderName())
			}
		})
	}
}

func TestClientFactoryService_GetClientForProvider_NotInitialized(t *testing.T) {
	service := NewClientFactoryService()

	client, err := service.GetClientForProvider("openai", "sk-test-key")

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "client factory service not initialized")
}

func TestClientFactoryService_ClientCaching(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	provider := "openai"
	apiKey := "sk-test-key"

	// First call should create a new client
	client1, err := service.GetClientForProvider(provider, apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Second call with same provider and key should return cached client
	client2, err := service.GetClientForProvider(provider, apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	assert.Equal(t, 1, service.GetCachedClientCount())
	assert.Same(t, client1, client2) // Should be the same instance

	// Different API key should create new client
	client3, err := service.GetClientForProvider(provider, "sk-different-key")
	assert.NoError(t, err)
	assert.NotNil(t, client3)
	assert.Equal(t, 2, service.GetCachedClientCount())
	assert.NotSame(t, client1, client3)
}

func TestClientFactoryService_CacheKeyFormat(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test that cache keys are provider-specific
	openaiClient, err := service.GetClientForProvider("openai", "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Different provider with same key should be treated as different
	// (Note: anthropic will fail because it's not implemented yet)
	_, err = service.GetClientForProvider("anthropic", "test-key")
	assert.Error(t, err)                               // Expected because anthropic is not implemented
	assert.Equal(t, 1, service.GetCachedClientCount()) // Should remain 1 due to error
}

func TestClientFactoryService_DetermineAPIKeyForProvider(t *testing.T) {
	// Save original environment variables
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalAnthropic := os.Getenv("ANTHROPIC_API_KEY")

	// Clean up after test
	defer func() {
		if originalOpenAI != "" {
			_ = os.Setenv("OPENAI_API_KEY", originalOpenAI)
		} else {
			_ = os.Unsetenv("OPENAI_API_KEY")
		}
		if originalAnthropic != "" {
			_ = os.Setenv("ANTHROPIC_API_KEY", originalAnthropic)
		} else {
			_ = os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name        string
		provider    string
		envVar      string
		envValue    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "openai key found",
			provider:    "openai",
			envVar:      "OPENAI_API_KEY",
			envValue:    "sk-test-openai-key",
			expectError: false,
		},
		{
			name:        "anthropic key found",
			provider:    "anthropic",
			envVar:      "ANTHROPIC_API_KEY",
			envValue:    "sk-test-anthropic-key",
			expectError: false,
		},
		{
			name:        "openai key not found",
			provider:    "openai",
			envVar:      "OPENAI_API_KEY",
			envValue:    "",
			expectError: true,
			errorMsg:    "openai API key not found. Please set the OPENAI_API_KEY environment variable",
		},
		{
			name:        "anthropic key not found",
			provider:    "anthropic",
			envVar:      "ANTHROPIC_API_KEY",
			envValue:    "",
			expectError: true,
			errorMsg:    "anthropic API key not found. Please set the ANTHROPIC_API_KEY environment variable",
		},
		{
			name:        "empty provider",
			provider:    "",
			expectError: true,
			errorMsg:    "provider cannot be empty",
		},
		{
			name:        "unsupported provider",
			provider:    "unsupported",
			expectError: true,
			errorMsg:    "unsupported provider 'unsupported'. Supported providers: openai, anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable for this test
			if tt.envVar != "" {
				if tt.envValue != "" {
					_ = os.Setenv(tt.envVar, tt.envValue)
				} else {
					_ = os.Unsetenv(tt.envVar)
				}
			}

			apiKey, err := service.DetermineAPIKeyForProvider(tt.provider)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, apiKey)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.envValue, apiKey)
			}
		})
	}
}

func TestClientFactoryService_ConcurrentAccess(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	results := make(chan neurotypes.LLMClient, numGoroutines*numIterations)

	// Launch multiple goroutines that create clients concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				client, err := service.GetClientForProvider("openai", "sk-test-key")
				if err != nil {
					t.Errorf("Goroutine %d, iteration %d: unexpected error: %v", goroutineID, j, err)
					return
				}
				results <- client
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Verify all clients are the same (cached)
	var firstClient neurotypes.LLMClient
	count := 0
	for client := range results {
		if firstClient == nil {
			firstClient = client
		}
		assert.Same(t, firstClient, client, "All clients should be the same cached instance")
		count++
	}

	assert.Equal(t, numGoroutines*numIterations, count)
	assert.Equal(t, 1, service.GetCachedClientCount()) // Only one client should be cached
}

func TestClientFactoryService_ClearCache(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create some clients
	_, err = service.GetClientForProvider("openai", "sk-key1")
	assert.NoError(t, err)
	_, err = service.GetClientForProvider("openai", "sk-key2")
	assert.NoError(t, err)

	assert.Equal(t, 2, service.GetCachedClientCount())

	// Clear cache
	service.ClearCache()
	assert.Equal(t, 0, service.GetCachedClientCount())

	// Verify new clients are created after clearing
	client, err := service.GetClientForProvider("openai", "sk-key1")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 1, service.GetCachedClientCount())
}

func TestClientFactoryService_GetCachedClientCount(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Initially no clients
	assert.Equal(t, 0, service.GetCachedClientCount())

	// Create clients and verify count
	_, err = service.GetClientForProvider("openai", "sk-key1")
	assert.NoError(t, err)
	assert.Equal(t, 1, service.GetCachedClientCount())

	_, err = service.GetClientForProvider("openai", "sk-key2")
	assert.NoError(t, err)
	assert.Equal(t, 2, service.GetCachedClientCount())

	// Same key should not increase count
	_, err = service.GetClientForProvider("openai", "sk-key1")
	assert.NoError(t, err)
	assert.Equal(t, 2, service.GetCachedClientCount())
}

func TestClientFactoryService_ErrorMessages(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test error message formatting
	tests := []struct {
		name     string
		provider string
		apiKey   string
		expected string
	}{
		{
			name:     "empty provider",
			provider: "",
			apiKey:   "test-key",
			expected: "provider cannot be empty",
		},
		{
			name:     "empty api key",
			provider: "openai",
			apiKey:   "",
			expected: "API key cannot be empty for provider 'openai'",
		},
		{
			name:     "unsupported provider",
			provider: "gpt",
			apiKey:   "test-key",
			expected: "unsupported provider 'gpt'. Supported providers: openai, anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetClientForProvider(tt.provider, tt.apiKey)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestClientFactoryService_DetermineAPIKeyForProvider_ErrorMessages(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Clear environment variables
	_ = os.Unsetenv("OPENAI_API_KEY")
	_ = os.Unsetenv("ANTHROPIC_API_KEY")

	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "openai key missing",
			provider: "openai",
			expected: "openai API key not found. Please set the OPENAI_API_KEY environment variable",
		},
		{
			name:     "anthropic key missing",
			provider: "anthropic",
			expected: "anthropic API key not found. Please set the ANTHROPIC_API_KEY environment variable",
		},
		{
			name:     "empty provider",
			provider: "",
			expected: "provider cannot be empty",
		},
		{
			name:     "unsupported provider",
			provider: "claude",
			expected: "unsupported provider 'claude'. Supported providers: openai, anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.DetermineAPIKeyForProvider(tt.provider)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

// Integration test that verifies the full flow
func TestClientFactoryService_Integration(t *testing.T) {
	// Set up test environment
	_ = os.Setenv("OPENAI_API_KEY", "sk-test-integration-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()

	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test the full flow: determine API key -> create client
	apiKey, err := service.DetermineAPIKeyForProvider("openai")
	assert.NoError(t, err)
	assert.Equal(t, "sk-test-integration-key", apiKey)

	client, err := service.GetClientForProvider("openai", apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "openai", client.GetProviderName())
	assert.True(t, client.IsConfigured())
}

func TestClientFactoryService_ProviderSpecificCaching(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test that different providers with same API key are cached separately
	// This tests the cache key format: "provider:apikey"

	// Create OpenAI client
	openaiClient, err := service.GetClientForProvider("openai", "same-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Attempting to create Anthropic client with same key should fail
	// (because anthropic is not implemented yet)
	_, err = service.GetClientForProvider("anthropic", "same-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "anthropic provider is not yet implemented")

	// Cache count should remain 1 (no new client created due to error)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// But different OpenAI keys should create separate entries
	openaiClient2, err := service.GetClientForProvider("openai", "different-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient2)
	assert.Equal(t, 2, service.GetCachedClientCount())
	assert.NotSame(t, openaiClient, openaiClient2)
}
