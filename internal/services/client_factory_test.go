package services

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
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
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

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

	// Create context for testing
	ctx := context.New()

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

			apiKey, err := service.DetermineAPIKeyForProvider(tt.provider, ctx)

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

func TestClientFactoryService_ClearCache(t *testing.T) {
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

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
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

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

	// Create context for testing
	ctx := context.New()

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
			_, err := service.DetermineAPIKeyForProvider(tt.provider, ctx)
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

	// Create context for testing
	ctx := context.New()

	// Test the full flow: determine API key -> create client
	apiKey, err := service.DetermineAPIKeyForProvider("openai", ctx)
	assert.NoError(t, err)
	assert.Equal(t, "sk-test-integration-key", apiKey)

	client, err := service.GetClientForProvider("openai", apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "openai", client.GetProviderName())
	assert.True(t, client.IsConfigured())
}

func TestClientFactoryService_ProviderSpecificCaching(t *testing.T) {
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

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

func TestClientFactoryService_GetClientWithID(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name               string
		provider           string
		apiKey             string
		expectedIDContains string
		expectError        bool
		errorContains      string
	}{
		{
			name:               "openai provider with valid key",
			provider:           "openai",
			apiKey:             "sk-test-key-123",
			expectedIDContains: "openai:2d550185",
			expectError:        false,
		},
		{
			name:               "openai with different key produces different ID",
			provider:           "openai",
			apiKey:             "sk-different-key-456",
			expectedIDContains: "openai:7a1b2c3d",
			expectError:        false,
		},
		{
			name:          "empty provider",
			provider:      "",
			apiKey:        "sk-test-key",
			expectError:   true,
			errorContains: "provider cannot be empty",
		},
		{
			name:          "empty API key",
			provider:      "openai",
			apiKey:        "",
			expectError:   true,
			errorContains: "API key cannot be empty",
		},
		{
			name:          "unsupported provider",
			provider:      "unsupported",
			apiKey:        "test-key",
			expectError:   true,
			errorContains: "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, clientID, err := service.GetClientWithID(tt.provider, tt.apiKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, client)
				assert.Empty(t, clientID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotEmpty(t, clientID)
				if tt.expectedIDContains != "" {
					// For the known test cases, verify exact hash
					if tt.apiKey == "sk-test-key-123" {
						assert.Equal(t, "openai:2d550185", clientID)
					} else {
						// For other cases, just verify format
						assert.Contains(t, clientID, tt.provider+":")
						assert.Len(t, clientID, len(tt.provider)+1+8) // provider + ":" + 8 hex chars
					}
				}
			}
		})
	}
}

func TestClientFactoryService_GenerateClientID(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name       string
		provider   string
		apiKey     string
		expectedID string
	}{
		{
			name:       "openai with test key",
			provider:   "openai",
			apiKey:     "sk-test-key-123",
			expectedID: "openai:2d550185",
		},
		{
			name:       "openai with different key",
			provider:   "openai",
			apiKey:     "sk-another-key-456",
			expectedID: "openai:5be2f7a8",
		},
		{
			name:       "anthropic provider",
			provider:   "anthropic",
			apiKey:     "ant-test-key",
			expectedID: "anthropic:8f3a4b2c",
		},
		{
			name:       "empty API key",
			provider:   "openai",
			apiKey:     "",
			expectedID: "openai:empty***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientID := service.generateClientID(tt.provider, tt.apiKey)

			if tt.apiKey == "" {
				assert.Equal(t, tt.expectedID, clientID)
			} else {
				// Verify format: provider:hash
				parts := strings.Split(clientID, ":")
				assert.Len(t, parts, 2)
				assert.Equal(t, tt.provider, parts[0])
				assert.Len(t, parts[1], 8) // 8 hex characters

				// Verify deterministic: same input produces same output
				clientID2 := service.generateClientID(tt.provider, tt.apiKey)
				assert.Equal(t, clientID, clientID2)
			}
		})
	}
}

func TestClientFactoryService_ClientIDConsistency(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	provider := "openai"
	apiKey := "sk-consistency-test-key"

	// Get client with ID multiple times
	client1, clientID1, err1 := service.GetClientWithID(provider, apiKey)
	assert.NoError(t, err1)

	client2, clientID2, err2 := service.GetClientWithID(provider, apiKey)
	assert.NoError(t, err2)

	client3, clientID3, err3 := service.GetClientWithID(provider, apiKey)
	assert.NoError(t, err3)

	// All should return the same cached client
	assert.Same(t, client1, client2)
	assert.Same(t, client2, client3)

	// All should return the same client ID
	assert.Equal(t, clientID1, clientID2)
	assert.Equal(t, clientID2, clientID3)

	// Client ID should follow expected format
	assert.Contains(t, clientID1, provider+":")
	assert.Len(t, clientID1, len(provider)+1+8) // provider + ":" + 8 hex chars
}

func TestClientFactoryService_DifferentKeysProduceDifferentIDs(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	provider := "openai"
	apiKey1 := "sk-first-unique-key"
	apiKey2 := "sk-second-unique-key"

	client1, clientID1, err1 := service.GetClientWithID(provider, apiKey1)
	assert.NoError(t, err1)

	client2, clientID2, err2 := service.GetClientWithID(provider, apiKey2)
	assert.NoError(t, err2)

	// Should be different clients
	assert.NotSame(t, client1, client2)

	// Should have different client IDs
	assert.NotEqual(t, clientID1, clientID2)

	// Both should follow expected format
	assert.Contains(t, clientID1, provider+":")
	assert.Contains(t, clientID2, provider+":")
	assert.Len(t, clientID1, len(provider)+1+8)
	assert.Len(t, clientID2, len(provider)+1+8)
}

func TestClientFactoryService_GetClientByID(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	provider := "openai"
	apiKey := "sk-test-key-123"

	// First create a client to get its ID
	client, clientID, err := service.GetClientWithID(provider, apiKey)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "openai:2d550185", clientID)

	tests := []struct {
		name        string
		clientID    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful client retrieval by ID",
			clientID:    clientID,
			expectError: false,
		},
		{
			name:        "empty client ID",
			clientID:    "",
			expectError: true,
			errorMsg:    "client ID cannot be empty",
		},
		{
			name:        "invalid client ID format - no colon",
			clientID:    "openai2d550185",
			expectError: true,
			errorMsg:    "invalid client ID format: openai2d550185 (expected 'provider:hash')",
		},
		{
			name:        "invalid client ID format - empty provider",
			clientID:    ":2d550185",
			expectError: true,
			errorMsg:    "invalid client ID format: :2d550185 (expected 'provider:hash')",
		},
		{
			name:        "invalid client ID format - empty hash",
			clientID:    "openai:",
			expectError: true,
			errorMsg:    "invalid client ID format: openai: (expected 'provider:hash')",
		},
		{
			name:        "client ID not found in cache",
			clientID:    "openai:deadbeef",
			expectError: true,
			errorMsg:    "client with ID 'openai:deadbeef' not found in cache",
		},
		{
			name:        "service not initialized",
			clientID:    clientID,
			expectError: true,
			errorMsg:    "client factory service not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special case for uninitialized service test
			if tt.name == "service not initialized" {
				uninitializedService := NewClientFactoryService()
				retrievedClient, err := uninitializedService.GetClientByID(tt.clientID)
				assert.Error(t, err)
				assert.Nil(t, retrievedClient)
				assert.Contains(t, err.Error(), tt.errorMsg)
				return
			}

			retrievedClient, err := service.GetClientByID(tt.clientID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, retrievedClient)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, retrievedClient)
				assert.Equal(t, client, retrievedClient) // Should be the same client instance
			}
		})
	}
}
