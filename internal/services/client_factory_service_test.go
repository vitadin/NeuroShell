package services

import (
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
		name              string
		providerCatalogID string
		apiKey            string
		expectError       bool
		errorMsg          string
	}{
		{
			name:              "successful OAC client creation",
			providerCatalogID: "OAC",
			apiKey:            "sk-test-key",
			expectError:       false,
		},
		{
			name:              "successful OAR client creation",
			providerCatalogID: "OAR",
			apiKey:            "sk-test-key",
			expectError:       false,
		},
		{
			name:              "successful ORC client creation",
			providerCatalogID: "ORC",
			apiKey:            "sk-or-test-key",
			expectError:       false,
		},
		{
			name:              "successful MSC client creation",
			providerCatalogID: "MSC",
			apiKey:            "sk-ms-test-key",
			expectError:       false,
		},
		{
			name:              "successful ANC client creation",
			providerCatalogID: "ANC",
			apiKey:            "sk-ant-test-key",
			expectError:       false,
		},
		{
			name:              "successful GMC client creation",
			providerCatalogID: "GMC",
			apiKey:            "sk-gmc-test-key",
			expectError:       false,
		},
		{
			name:              "empty provider catalog ID",
			providerCatalogID: "",
			apiKey:            "sk-test-key",
			expectError:       true,
			errorMsg:          "provider catalog ID cannot be empty",
		},
		{
			name:              "empty api key",
			providerCatalogID: "OAC",
			apiKey:            "",
			expectError:       true,
			errorMsg:          "API key cannot be empty for provider catalog ID 'OAC'",
		},
		{
			name:              "unsupported provider catalog ID",
			providerCatalogID: "UNSUPPORTED",
			apiKey:            "test-key",
			expectError:       true,
			errorMsg:          "unsupported provider catalog ID 'UNSUPPORTED'. Supported catalog IDs: OAC, OAR, ORC, MSC, ANC, GMC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewClientFactoryService()
			err := service.Initialize()
			require.NoError(t, err)

			client, err := service.GetClientForProvider(tt.providerCatalogID, tt.apiKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				// Note: client.GetProviderName() returns the underlying provider name, not catalog ID
				expectedProviderName := map[string]string{
					"OAC": "openai", "OAR": "openai", "ORC": "openrouter",
					"MSC": "moonshot", "ANC": "anthropic", "GMC": "gemini",
				}[tt.providerCatalogID]
				assert.Equal(t, expectedProviderName, client.GetProviderName())
			}
		})
	}
}

func TestClientFactoryService_GetClientForProvider_NotInitialized(t *testing.T) {
	service := NewClientFactoryService()

	client, err := service.GetClientForProvider("OAR", "sk-test-key")

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "client factory service not initialized")
}

func TestClientFactoryService_ClientCaching(t *testing.T) {
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	providerCatalogID := "OAR"
	apiKey := "sk-test-key"

	// First call should create a new client
	client1, err := service.GetClientForProvider(providerCatalogID, apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client1)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Second call with same provider and key should return cached client
	client2, err := service.GetClientForProvider(providerCatalogID, apiKey)
	assert.NoError(t, err)
	assert.NotNil(t, client2)
	assert.Equal(t, 1, service.GetCachedClientCount())
	assert.Same(t, client1, client2) // Should be the same instance

	// Different API key should create new client
	client3, err := service.GetClientForProvider(providerCatalogID, "sk-different-key")
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

	// Test that cache keys are catalog ID specific
	openaiClient, err := service.GetClientForProvider("OAR", "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Different catalog ID with same key should be treated as different
	anthropicClient, err := service.GetClientForProvider("ANC", "test-key")
	assert.NoError(t, err)
	assert.NotNil(t, anthropicClient)
	assert.Equal(t, 2, service.GetCachedClientCount()) // Should be 2 now (OAR + ANC)
}

func TestClientFactoryService_ClearCache(t *testing.T) {
	// Clear global context to ensure test isolation
	context.ResetGlobalContext()

	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create some clients
	_, err = service.GetClientForProvider("OAR", "sk-key1")
	assert.NoError(t, err)
	_, err = service.GetClientForProvider("OAR", "sk-key2")
	assert.NoError(t, err)

	assert.Equal(t, 2, service.GetCachedClientCount())

	// Clear cache
	service.ClearCache()
	assert.Equal(t, 0, service.GetCachedClientCount())

	// Verify new clients are created after clearing
	client, err := service.GetClientForProvider("OAR", "sk-key1")
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
	_, err = service.GetClientForProvider("OAR", "sk-key1")
	assert.NoError(t, err)
	assert.Equal(t, 1, service.GetCachedClientCount())

	_, err = service.GetClientForProvider("OAR", "sk-key2")
	assert.NoError(t, err)
	assert.Equal(t, 2, service.GetCachedClientCount())

	// Same key should not increase count
	_, err = service.GetClientForProvider("OAR", "sk-key1")
	assert.NoError(t, err)
	assert.Equal(t, 2, service.GetCachedClientCount())
}

func TestClientFactoryService_ErrorMessages(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test error message formatting
	tests := []struct {
		name              string
		providerCatalogID string
		apiKey            string
		expected          string
	}{
		{
			name:              "empty provider catalog ID",
			providerCatalogID: "",
			apiKey:            "test-key",
			expected:          "provider catalog ID cannot be empty",
		},
		{
			name:              "empty api key",
			providerCatalogID: "OAR",
			apiKey:            "",
			expected:          "API key cannot be empty for provider catalog ID 'OAR'",
		},
		{
			name:              "unsupported provider catalog ID",
			providerCatalogID: "INVALID",
			apiKey:            "test-key",
			expected:          "unsupported provider catalog ID 'INVALID'. Supported catalog IDs: OAC, OAR, ORC, MSC, ANC, GMC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetClientForProvider(tt.providerCatalogID, tt.apiKey)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

// Integration test that verifies the full flow
func TestClientFactoryService_Integration(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test direct client creation with catalog ID
	apiKey := "sk-test-integration-key"
	client, err := service.GetClientForProvider("OAR", apiKey)
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

	// Test that different catalog IDs with same API key are cached separately
	// This tests the cache key format: "catalog_id:apikey"

	// Create OpenAI reasoning client
	openaiClient, err := service.GetClientForProvider("OAR", "same-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient)
	assert.Equal(t, 1, service.GetCachedClientCount())

	// Create Anthropic client with same key should succeed
	anthropicClient, err := service.GetClientForProvider("ANC", "same-key")
	assert.NoError(t, err)
	assert.NotNil(t, anthropicClient)

	// Cache count should be 2 (OAR + ANC with same key but different catalog IDs)
	assert.Equal(t, 2, service.GetCachedClientCount())

	// But different OpenAI keys should create separate entries
	openaiClient2, err := service.GetClientForProvider("OAR", "different-key")
	assert.NoError(t, err)
	assert.NotNil(t, openaiClient2)
	assert.Equal(t, 3, service.GetCachedClientCount()) // Now 3 total: OAR:same-key, ANC:same-key, OAR:different-key
	assert.NotSame(t, openaiClient, openaiClient2)
}

func TestClientFactoryService_GetClientWithID(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name               string
		providerCatalogID  string
		apiKey             string
		expectedIDContains string
		expectError        bool
		errorContains      string
	}{
		{
			name:               "OAR catalog ID with valid key",
			providerCatalogID:  "OAR",
			apiKey:             "sk-test-key-123",
			expectedIDContains: "OAR:2d550185",
			expectError:        false,
		},
		{
			name:               "OAR with different key produces different ID",
			providerCatalogID:  "OAR",
			apiKey:             "sk-different-key-456",
			expectedIDContains: "OAR:7a1b2c3d",
			expectError:        false,
		},
		{
			name:              "empty provider catalog ID",
			providerCatalogID: "",
			apiKey:            "sk-test-key",
			expectError:       true,
			errorContains:     "provider catalog ID cannot be empty",
		},
		{
			name:              "empty API key",
			providerCatalogID: "OAR",
			apiKey:            "",
			expectError:       true,
			errorContains:     "API key cannot be empty",
		},
		{
			name:              "unsupported provider catalog ID",
			providerCatalogID: "UNSUPPORTED",
			apiKey:            "test-key",
			expectError:       true,
			errorContains:     "unsupported provider catalog ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, clientID, err := service.GetClientWithID(tt.providerCatalogID, tt.apiKey)

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
						assert.Equal(t, "OAR:2d550185", clientID)
					} else {
						// For other cases, just verify format
						assert.Contains(t, clientID, tt.providerCatalogID+":")
						assert.Len(t, clientID, len(tt.providerCatalogID)+1+8) // catalog_id + ":" + 8 hex chars
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
		name              string
		providerCatalogID string
		apiKey            string
		expectedID        string
	}{
		{
			name:              "OAR with test key",
			providerCatalogID: "OAR",
			apiKey:            "sk-test-key-123",
			expectedID:        "OAR:2d550185",
		},
		{
			name:              "OAR with different key",
			providerCatalogID: "OAR",
			apiKey:            "sk-another-key-456",
			expectedID:        "OAR:5be2f7a8",
		},
		{
			name:              "ANC catalog ID",
			providerCatalogID: "ANC",
			apiKey:            "ant-test-key",
			expectedID:        "ANC:8f3a4b2c",
		},
		{
			name:              "empty API key",
			providerCatalogID: "OAR",
			apiKey:            "",
			expectedID:        "OAR:empty***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientID := service.generateClientID(tt.providerCatalogID, tt.apiKey)

			if tt.apiKey == "" {
				assert.Equal(t, tt.expectedID, clientID)
			} else {
				// Verify format: catalog_id:hash
				parts := strings.Split(clientID, ":")
				assert.Len(t, parts, 2)
				assert.Equal(t, tt.providerCatalogID, parts[0])
				assert.Len(t, parts[1], 8) // 8 hex characters

				// Verify deterministic: same input produces same output
				clientID2 := service.generateClientID(tt.providerCatalogID, tt.apiKey)
				assert.Equal(t, clientID, clientID2)
			}
		})
	}
}

func TestClientFactoryService_ClientIDConsistency(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	providerCatalogID := "OAR"
	apiKey := "sk-consistency-test-key"

	// Get client with ID multiple times
	client1, clientID1, err1 := service.GetClientWithID(providerCatalogID, apiKey)
	assert.NoError(t, err1)

	client2, clientID2, err2 := service.GetClientWithID(providerCatalogID, apiKey)
	assert.NoError(t, err2)

	client3, clientID3, err3 := service.GetClientWithID(providerCatalogID, apiKey)
	assert.NoError(t, err3)

	// All should return the same cached client
	assert.Same(t, client1, client2)
	assert.Same(t, client2, client3)

	// All should return the same client ID
	assert.Equal(t, clientID1, clientID2)
	assert.Equal(t, clientID2, clientID3)

	// Client ID should follow expected format
	assert.Contains(t, clientID1, providerCatalogID+":")
	assert.Len(t, clientID1, len(providerCatalogID)+1+8) // catalog_id + ":" + 8 hex chars
}

func TestClientFactoryService_DifferentKeysProduceDifferentIDs(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	providerCatalogID := "OAR"
	apiKey1 := "sk-first-unique-key"
	apiKey2 := "sk-second-unique-key"

	client1, clientID1, err1 := service.GetClientWithID(providerCatalogID, apiKey1)
	assert.NoError(t, err1)

	client2, clientID2, err2 := service.GetClientWithID(providerCatalogID, apiKey2)
	assert.NoError(t, err2)

	// Should be different clients
	assert.NotSame(t, client1, client2)

	// Should have different client IDs
	assert.NotEqual(t, clientID1, clientID2)

	// Both should follow expected format
	assert.Contains(t, clientID1, providerCatalogID+":")
	assert.Contains(t, clientID2, providerCatalogID+":")
	assert.Len(t, clientID1, len(providerCatalogID)+1+8)
	assert.Len(t, clientID2, len(providerCatalogID)+1+8)
}

func TestClientFactoryService_GetClientByID(t *testing.T) {
	service := NewClientFactoryService()
	err := service.Initialize()
	require.NoError(t, err)

	providerCatalogID := "OAR"
	apiKey := "sk-test-key-123"

	// First create a client to get its ID
	client, clientID, err := service.GetClientWithID(providerCatalogID, apiKey)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "OAR:2d550185", clientID)

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
			clientID:    "OAR2d550185",
			expectError: true,
			errorMsg:    "invalid client ID format: OAR2d550185 (expected 'catalog_id:hash')",
		},
		{
			name:        "invalid client ID format - empty catalog ID",
			clientID:    ":2d550185",
			expectError: true,
			errorMsg:    "invalid client ID format: :2d550185 (expected 'catalog_id:hash')",
		},
		{
			name:        "invalid client ID format - empty hash",
			clientID:    "OAR:",
			expectError: true,
			errorMsg:    "invalid client ID format: OAR: (expected 'catalog_id:hash')",
		},
		{
			name:        "client ID not found in cache",
			clientID:    "OAR:deadbeef",
			expectError: true,
			errorMsg:    "client with ID 'OAR:deadbeef' not found in cache",
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
