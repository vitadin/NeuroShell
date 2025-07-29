package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestProviderCatalogService_Name(t *testing.T) {
	service := NewProviderCatalogService()
	assert.Equal(t, "provider_catalog", service.Name())
}

func TestProviderCatalogService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		{
			name: "successful initialization",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewProviderCatalogService()
			err := service.Initialize()

			if tt.want != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.want.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestProviderCatalogService_GetProviderCatalog(t *testing.T) {
	service := NewProviderCatalogService()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		providers, err := service.GetProviderCatalog()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, providers)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("successful catalog retrieval", func(t *testing.T) {
		providers, err := service.GetProviderCatalog()
		require.NoError(t, err)
		require.NotNil(t, providers)

		// Verify we get providers from all expected providers
		assert.Equal(t, 6, len(providers), "Should have exactly 6 providers")

		// Check that we have providers from all expected providers
		providerNames := make(map[string]bool)
		for _, provider := range providers {
			// Verify provider structure
			assert.NotEmpty(t, provider.ID, "Provider ID should not be empty")
			assert.NotEmpty(t, provider.Provider, "Provider name should not be empty")
			assert.NotEmpty(t, provider.DisplayName, "Display name should not be empty")
			assert.NotEmpty(t, provider.BaseURL, "Base URL should not be empty")
			assert.NotEmpty(t, provider.ClientType, "Client type should not be empty")
			assert.NotEmpty(t, provider.Description, "Description should not be empty")

			// Track provider names
			providerNames[provider.Provider] = true
		}

		// Verify all expected providers are present
		expectedProviders := []string{"openai", "openai-reasoning", "anthropic", "moonshot", "openrouter", "gemini"}
		for _, expected := range expectedProviders {
			assert.True(t, providerNames[expected], "Should have provider: %s", expected)
		}
	})
}

func TestProviderCatalogService_GetProvidersByProvider(t *testing.T) {
	service := NewProviderCatalogService()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, providers)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("get openai providers", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("openai")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should have exactly 1 OpenAI provider")

		provider := providers[0]
		assert.Equal(t, "OAC", provider.ID)
		assert.Equal(t, "openai", provider.Provider)
		assert.Equal(t, "openai", provider.ClientType)
		assert.Contains(t, provider.BaseURL, "api.openai.com")
		assert.NotEmpty(t, provider.DisplayName)
	})

	t.Run("get anthropic providers", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("anthropic")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should have exactly 1 Anthropic provider")

		provider := providers[0]
		assert.Equal(t, "ANC", provider.ID)
		assert.Equal(t, "anthropic", provider.Provider)
		assert.Equal(t, "openai-compatible", provider.ClientType)
		assert.Contains(t, provider.BaseURL, "api.anthropic.com")
		assert.NotEmpty(t, provider.Headers, "Anthropic should have headers")
		assert.Equal(t, "2023-06-01", provider.Headers["anthropic-version"])
	})

	t.Run("get moonshot providers", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("moonshot")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should have exactly 1 Moonshot provider")

		provider := providers[0]
		assert.Equal(t, "MSC", provider.ID)
		assert.Equal(t, "moonshot", provider.Provider)
		assert.Equal(t, "openai-compatible", provider.ClientType)
		assert.Contains(t, provider.BaseURL, "api.moonshot.ai")
	})

	t.Run("get openrouter providers", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("openrouter")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should have exactly 1 OpenRouter provider")

		provider := providers[0]
		assert.Equal(t, "ORC", provider.ID)
		assert.Equal(t, "openrouter", provider.Provider)
		assert.Equal(t, "openai-compatible", provider.ClientType)
		assert.Contains(t, provider.BaseURL, "openrouter.ai")
		assert.NotEmpty(t, provider.Headers, "OpenRouter should have headers")
		assert.Equal(t, "https://github.com/vitadin/NeuroShell", provider.Headers["HTTP-Referer"])
		assert.Equal(t, "NeuroShell", provider.Headers["X-Title"])
	})

	t.Run("get gemini providers", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("gemini")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should have exactly 1 Gemini provider")

		provider := providers[0]
		assert.Equal(t, "GMC", provider.ID)
		assert.Equal(t, "gemini", provider.Provider)
		assert.Equal(t, "gemini", provider.ClientType)
		assert.Contains(t, provider.BaseURL, "generativelanguage.googleapis.com")
		assert.Contains(t, provider.Endpoint, "generateContent")
		assert.NotEmpty(t, provider.Headers, "Gemini should have headers")
		assert.Equal(t, "{API_KEY}", provider.Headers["x-goog-api-key"])
		assert.Equal(t, "application/json", provider.Headers["Content-Type"])
		assert.Equal(t, "Natively supported by NeuroShell", provider.ImplementationNotes)
	})

	t.Run("case insensitive provider names", func(t *testing.T) {
		providersLower, err1 := service.GetProvidersByProvider("openai")
		providersUpper, err2 := service.GetProvidersByProvider("OPENAI")
		providersMixed, err3 := service.GetProvidersByProvider("OpenAI")

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		assert.Equal(t, len(providersLower), len(providersUpper))
		assert.Equal(t, len(providersLower), len(providersMixed))
		assert.Equal(t, 1, len(providersLower))
	})

	t.Run("unsupported provider", func(t *testing.T) {
		providers, err := service.GetProvidersByProvider("unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider")
		assert.Nil(t, providers)
	})
}

func TestProviderCatalogService_SearchProviderCatalog(t *testing.T) {
	service := NewProviderCatalogService()

	// Test not initialized (indirectly through GetProviderCatalog)
	t.Run("not initialized", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, providers)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("search by provider ID", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("OAC")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should find exactly 1 provider")

		provider := providers[0]
		assert.Equal(t, "OAC", provider.ID)
		assert.Equal(t, "openai", provider.Provider)
	})

	t.Run("search by provider name", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("anthropic")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Equal(t, 1, len(providers), "Should find exactly 1 Anthropic provider")

		provider := providers[0]
		assert.Equal(t, "anthropic", provider.Provider)
		assert.Equal(t, "ANC", provider.ID)
	})

	t.Run("search by display name", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("Chat Completions")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Greater(t, len(providers), 1, "Should find multiple providers with 'Chat Completions'")

		// Verify all results contain the search term in name, display name, or description
		for _, provider := range providers {
			searchTerm := "chat completions"
			assert.True(t,
				strings.Contains(strings.ToLower(provider.ID), searchTerm) ||
					strings.Contains(strings.ToLower(provider.DisplayName), searchTerm) ||
					strings.Contains(strings.ToLower(provider.Provider), searchTerm) ||
					strings.Contains(strings.ToLower(provider.Description), searchTerm),
				"Provider should contain search term in ID, display name, provider, or description")
		}
	})

	t.Run("search by description", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("API")
		require.NoError(t, err)
		require.NotNil(t, providers)
		assert.Greater(t, len(providers), 0, "Should find providers with 'API' in description")

		// Verify results contain providers with API in description
		for _, provider := range providers {
			searchTerm := "api"
			assert.True(t,
				strings.Contains(strings.ToLower(provider.ID), searchTerm) ||
					strings.Contains(strings.ToLower(provider.DisplayName), searchTerm) ||
					strings.Contains(strings.ToLower(provider.Provider), searchTerm) ||
					strings.Contains(strings.ToLower(provider.Description), searchTerm),
				"Provider should contain 'API' in ID, display name, provider, or description")
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		providersLower, err1 := service.SearchProviderCatalog("openai")
		providersUpper, err2 := service.SearchProviderCatalog("OPENAI")
		providersMixed, err3 := service.SearchProviderCatalog("OpenAI")

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		assert.Equal(t, len(providersLower), len(providersUpper))
		assert.Equal(t, len(providersLower), len(providersMixed))
	})

	t.Run("no results found", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("nonexistentprovider123")
		require.NoError(t, err)
		assert.Equal(t, 0, len(providers), "Should return empty slice for no matches")
	})

	t.Run("empty search query", func(t *testing.T) {
		providers, err := service.SearchProviderCatalog("")
		require.NoError(t, err)
		require.NotNil(t, providers)
		// Should return all providers since empty query matches everything
		assert.Equal(t, 6, len(providers), "Empty search should return all providers")
	})
}

func TestProviderCatalogService_GetSupportedProviders(t *testing.T) {
	service := NewProviderCatalogService()

	// This method doesn't require initialization
	providers := service.GetSupportedProviders()
	require.NotNil(t, providers)
	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
	assert.Contains(t, providers, "moonshot")
	assert.Contains(t, providers, "openrouter")
	assert.Contains(t, providers, "gemini")
	assert.Equal(t, 5, len(providers), "Should have exactly 5 supported providers")
}

func TestProviderCatalogService_GetProviderByID(t *testing.T) {
	service := NewProviderCatalogService()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		provider, err := service.GetProviderByID("openai_chat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Equal(t, neurotypes.ProviderCatalogEntry{}, provider)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("get provider by exact ID", func(t *testing.T) {
		provider, err := service.GetProviderByID("OAC")
		require.NoError(t, err)
		assert.Equal(t, "OAC", provider.ID)
		assert.Equal(t, "openai", provider.Provider)
		assert.Equal(t, "openai", provider.ClientType)
	})

	t.Run("get provider by case-insensitive ID", func(t *testing.T) {
		testCases := []string{"oac", "OAC", "Oac", "oAc"}
		for _, testID := range testCases {
			provider, err := service.GetProviderByID(testID)
			require.NoError(t, err, "Should find provider with ID: %s", testID)
			assert.Equal(t, "OAC", provider.ID)
			assert.Equal(t, "openai", provider.Provider)
		}
	})

	t.Run("provider not found", func(t *testing.T) {
		provider, err := service.GetProviderByID("NONEXISTENT")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in catalog")
		assert.Equal(t, neurotypes.ProviderCatalogEntry{}, provider)
	})

	t.Run("empty ID", func(t *testing.T) {
		provider, err := service.GetProviderByID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in catalog")
		assert.Equal(t, neurotypes.ProviderCatalogEntry{}, provider)
	})
}

func TestProviderCatalogService_loadProviderFile(t *testing.T) {
	service := NewProviderCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("load individual provider file", func(t *testing.T) {
		// Test loading individual provider file with embedded data
		provider, err := service.loadProviderFile([]byte(`
id: test_provider
provider: test
display_name: Test Provider
description: Test provider description
base_url: https://api.test.com/v1
endpoint: /test
client_type: openai-compatible
headers:
  test-header: test-value
`))
		require.NoError(t, err)

		assert.Equal(t, "test_provider", provider.ID)
		assert.Equal(t, "test", provider.Provider)
		assert.Equal(t, "Test Provider", provider.DisplayName)
		assert.Equal(t, "Test provider description", provider.Description)
		assert.Equal(t, "https://api.test.com/v1", provider.BaseURL)
		assert.Equal(t, "/test", provider.Endpoint)
		assert.Equal(t, "openai-compatible", provider.ClientType)
		assert.Equal(t, "test-value", provider.Headers["test-header"])
	})

	t.Run("invalid yaml", func(t *testing.T) {
		provider, err := service.loadProviderFile([]byte("invalid: yaml: ["))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse provider file")
		assert.Equal(t, neurotypes.ProviderCatalogEntry{}, provider)
	})

	t.Run("empty yaml", func(t *testing.T) {
		provider, err := service.loadProviderFile([]byte(""))
		require.NoError(t, err)
		assert.Equal(t, "", provider.ID, "Empty YAML should return empty provider")
	})
}

func TestProviderCatalogService_validateUniqueIDs(t *testing.T) {
	service := NewProviderCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("unique IDs pass validation", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{
			{ID: "OAC", Provider: "openai", DisplayName: "OpenAI Chat"},
			{ID: "ANC", Provider: "anthropic", DisplayName: "Anthropic Chat"},
			{ID: "MSC", Provider: "moonshot", DisplayName: "Moonshot Chat"},
		}
		err := service.validateUniqueIDs(providers)
		assert.NoError(t, err)
	})

	t.Run("duplicate IDs fail validation", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{
			{ID: "OAC", Provider: "openai", DisplayName: "OpenAI Chat"},
			{ID: "OAC", Provider: "openai", DisplayName: "OpenAI Chat Duplicate"},
		}
		err := service.validateUniqueIDs(providers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate provider ID found")
		assert.Contains(t, err.Error(), "OAC")
	})

	t.Run("case-insensitive duplicate IDs fail validation", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{
			{ID: "OAC", Provider: "openai", DisplayName: "OpenAI Chat"},
			{ID: "oac", Provider: "openai", DisplayName: "OpenAI Chat Lowercase"},
		}
		err := service.validateUniqueIDs(providers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate provider ID found")
		assert.Contains(t, err.Error(), "case insensitive")
	})

	t.Run("empty ID fails validation", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{
			{ID: "", Provider: "test", DisplayName: "Empty ID Provider"},
		}
		err := service.validateUniqueIDs(providers)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty ID field")
		assert.Contains(t, err.Error(), "Empty ID Provider")
	})

	t.Run("empty providers list passes validation", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{}
		err := service.validateUniqueIDs(providers)
		assert.NoError(t, err)
	})
}

func TestProviderCatalogService_normalizeID(t *testing.T) {
	service := NewProviderCatalogService()

	testCases := []struct {
		input    string
		expected string
	}{
		{"OAC", "OAC"},
		{"oac", "OAC"},
		{"Oac", "OAC"},
		{"ANC", "ANC"},
		{"anc", "ANC"},
		{"MSC", "MSC"},
		{"msc", "MSC"},
		{"ORC", "ORC"},
		{"orc", "ORC"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := service.normalizeID(tc.input)
		assert.Equal(t, tc.expected, result, "normalizeID(%s) should return %s", tc.input, tc.expected)
	}
}

func TestProviderCatalogService_IDValidationIntegration(t *testing.T) {
	// This test ensures that the actual embedded YAML files have unique IDs
	service := NewProviderCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("real catalog has unique IDs", func(t *testing.T) {
		providers, err := service.GetProviderCatalog()
		require.NoError(t, err)
		assert.Equal(t, 6, len(providers), "Should have 6 providers in catalog")

		// Verify all providers have IDs
		for _, provider := range providers {
			assert.NotEmpty(t, provider.ID, "Provider %s should have non-empty ID", provider.Provider)
		}

		// Verify IDs are unique (this is tested by GetProviderCatalog calling validateUniqueIDs)
		seenIDs := make(map[string]bool)
		for _, provider := range providers {
			normalizedID := strings.ToUpper(provider.ID)
			assert.False(t, seenIDs[normalizedID], "ID %s should be unique (case-insensitive)", provider.ID)
			seenIDs[normalizedID] = true
		}
	})

	t.Run("expected provider IDs exist", func(t *testing.T) {
		expectedIDs := []string{"OAC", "OAR", "ANC", "MSC", "ORC", "GMC"}

		for _, expectedID := range expectedIDs {
			provider, err := service.GetProviderByID(expectedID)
			require.NoError(t, err, "Should find provider with ID: %s", expectedID)
			assert.Equal(t, expectedID, provider.ID)
			assert.NotEmpty(t, provider.Provider, "Provider should have provider name")
			assert.NotEmpty(t, provider.BaseURL, "Provider should have base URL")
			assert.NotEmpty(t, provider.ClientType, "Provider should have client type")
		}
	})

	t.Run("case-insensitive lookup works for all providers", func(t *testing.T) {
		providers, err := service.GetProviderCatalog()
		require.NoError(t, err)

		for _, provider := range providers {
			// Test lowercase version
			lowerProvider, err := service.GetProviderByID(strings.ToLower(provider.ID))
			require.NoError(t, err, "Should find provider with lowercase ID: %s", strings.ToLower(provider.ID))
			assert.Equal(t, provider.ID, lowerProvider.ID)
			assert.Equal(t, provider.Provider, lowerProvider.Provider)
			assert.Equal(t, provider.BaseURL, lowerProvider.BaseURL)
		}
	})

	t.Run("all providers have required fields", func(t *testing.T) {
		providers, err := service.GetProviderCatalog()
		require.NoError(t, err)

		for _, provider := range providers {
			assert.NotEmpty(t, provider.ID, "Provider should have ID")
			assert.NotEmpty(t, provider.Provider, "Provider should have provider name")
			assert.NotEmpty(t, provider.DisplayName, "Provider should have display name")
			assert.NotEmpty(t, provider.BaseURL, "Provider should have base URL")
			assert.NotEmpty(t, provider.ClientType, "Provider should have client type")
			assert.NotEmpty(t, provider.Description, "Provider should have description")
			assert.NotEmpty(t, provider.ImplementationNotes, "Provider should have implementation notes")

			// Validate client types
			validClientTypes := []string{"openai", "openai_reasoning", "openai-compatible", "gemini"}
			assert.Contains(t, validClientTypes, provider.ClientType, "Provider should have valid client type")

			// Validate base URLs start with https
			assert.True(t, strings.HasPrefix(provider.BaseURL, "https://"), "Provider base URL should use HTTPS")

			// Validate implementation notes values
			validImplementationNotes := []string{"Natively supported by NeuroShell", "Uses OpenAI-compatible API", "Used for models with reasoning_tokens: true"}
			assert.Contains(t, validImplementationNotes, provider.ImplementationNotes, "Provider should have valid implementation notes")
		}
	})
}
