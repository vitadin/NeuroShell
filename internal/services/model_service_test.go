package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
)

func TestModelService_Name(t *testing.T) {
	service := NewModelService()
	assert.Equal(t, "model", service.Name())
}

func TestModelService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  *testutils.MockContext
		want error
	}{
		{
			name: "successful initialization",
			ctx:  testutils.NewMockContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewModelService()
			err := service.Initialize(tt.ctx)

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

func TestModelService_GetModelCatalog(t *testing.T) {
	service := NewModelService()
	ctx := testutils.NewMockContext()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize(ctx))

	t.Run("successful catalog retrieval", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		require.NoError(t, err)
		require.NotNil(t, models)

		// Verify we get models from both providers
		assert.Greater(t, len(models), 0, "Should have at least one model")

		// Check that we have models from both providers
		hasAnthropic := false
		hasOpenAI := false

		for _, model := range models {
			// Verify model structure
			assert.NotEmpty(t, model.Name, "Model name should not be empty")
			assert.NotEmpty(t, model.DisplayName, "Display name should not be empty")
			assert.NotEmpty(t, model.Description, "Description should not be empty")
			assert.Greater(t, model.ContextWindow, 0, "Context window should be positive")
			assert.NotEmpty(t, model.Capabilities, "Capabilities should not be empty")

			// Check provider types based on model names
			if model.Name == "claude-3-opus-20240229" || model.Name == "claude-3-5-sonnet-20241022" {
				hasAnthropic = true
			}
			if model.Name == "gpt-4" || model.Name == "gpt-4o" {
				hasOpenAI = true
			}
		}

		assert.True(t, hasAnthropic, "Should have Anthropic models")
		assert.True(t, hasOpenAI, "Should have OpenAI models")
	})
}

func TestModelService_GetModelCatalogByProvider(t *testing.T) {
	service := NewModelService()
	ctx := testutils.NewMockContext()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize(ctx))

	t.Run("get anthropic models", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("anthropic")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should have Anthropic models")

		// Verify all models are Anthropic models (Claude family)
		for _, model := range models {
			assert.Contains(t, model.Name, "claude", "Anthropic models should contain 'claude'")
			assert.NotEmpty(t, model.DisplayName)
			assert.Greater(t, model.ContextWindow, 0)
		}
	})

	t.Run("get openai models", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("openai")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should have OpenAI models")

		// Verify all models are OpenAI models (GPT family or embeddings)
		for _, model := range models {
			assert.True(t,
				model.Name == "gpt-4" ||
					model.Name == "gpt-4o" ||
					model.Name == "gpt-4-turbo" ||
					model.Name == "gpt-4o-mini" ||
					model.Name == "gpt-3.5-turbo" ||
					model.Name == "gpt-3.5-turbo-instruct" ||
					model.Name == "text-embedding-3-large" ||
					model.Name == "text-embedding-3-small" ||
					model.Name == "text-embedding-ada-002",
				"Should be a known OpenAI model: %s", model.Name)
			assert.NotEmpty(t, model.DisplayName)
			assert.Greater(t, model.ContextWindow, 0)
		}
	})

	t.Run("case insensitive provider names", func(t *testing.T) {
		modelsLower, err1 := service.GetModelCatalogByProvider("anthropic")
		modelsUpper, err2 := service.GetModelCatalogByProvider("ANTHROPIC")
		modelsMixed, err3 := service.GetModelCatalogByProvider("Anthropic")

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		assert.Equal(t, len(modelsLower), len(modelsUpper))
		assert.Equal(t, len(modelsLower), len(modelsMixed))
	})

	t.Run("unsupported provider", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider")
		assert.Nil(t, models)
	})
}

func TestModelService_SearchModelCatalog(t *testing.T) {
	service := NewModelService()
	ctx := testutils.NewMockContext()

	// Test not initialized (indirectly through GetModelCatalog)
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.SearchModelCatalog("gpt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize(ctx))

	t.Run("search by model name", func(t *testing.T) {
		models, err := service.SearchModelCatalog("gpt-4")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should find GPT-4 models")

		// Verify all results contain the search term
		for _, model := range models {
			assert.Contains(t, model.Name, "gpt-4", "Model name should contain search term")
		}
	})

	t.Run("search by display name", func(t *testing.T) {
		models, err := service.SearchModelCatalog("Claude")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should find Claude models")

		// Verify all results contain the search term in name or display name
		for _, model := range models {
			searchTerm := "claude"
			assert.True(t,
				strings.Contains(strings.ToLower(model.Name), searchTerm) ||
					strings.Contains(strings.ToLower(model.DisplayName), searchTerm) ||
					strings.Contains(strings.ToLower(model.Description), searchTerm),
				"Model should contain search term in name, display name, or description")
		}
	})

	t.Run("search by description", func(t *testing.T) {
		models, err := service.SearchModelCatalog("embedding")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should find embedding models")

		// Verify results contain embedding models
		for _, model := range models {
			searchTerm := "embedding"
			assert.True(t,
				strings.Contains(strings.ToLower(model.Name), searchTerm) ||
					strings.Contains(strings.ToLower(model.DisplayName), searchTerm) ||
					strings.Contains(strings.ToLower(model.Description), searchTerm),
				"Model should contain 'embedding' in name, display name, or description")
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		modelsLower, err1 := service.SearchModelCatalog("claude")
		modelsUpper, err2 := service.SearchModelCatalog("CLAUDE")
		modelsMixed, err3 := service.SearchModelCatalog("Claude")

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)

		assert.Equal(t, len(modelsLower), len(modelsUpper))
		assert.Equal(t, len(modelsLower), len(modelsMixed))
	})

	t.Run("no results found", func(t *testing.T) {
		models, err := service.SearchModelCatalog("nonexistentmodel123")
		require.NoError(t, err)
		assert.Equal(t, 0, len(models), "Should return empty slice for no matches")
	})

	t.Run("empty search query", func(t *testing.T) {
		models, err := service.SearchModelCatalog("")
		require.NoError(t, err)
		require.NotNil(t, models)
		// Should return all models since empty query matches everything
		assert.Greater(t, len(models), 0, "Empty search should return all models")
	})
}

func TestModelService_GetSupportedProviders(t *testing.T) {
	service := NewModelService()

	// This method doesn't require initialization
	providers := service.GetSupportedProviders()
	require.NotNil(t, providers)
	assert.Contains(t, providers, "anthropic")
	assert.Contains(t, providers, "openai")
	assert.Equal(t, 2, len(providers), "Should have exactly 2 supported providers")
}

func TestModelService_loadProviderCatalog(t *testing.T) {
	service := NewModelService()
	ctx := testutils.NewMockContext()
	require.NoError(t, service.Initialize(ctx))

	t.Run("load anthropic catalog", func(t *testing.T) {
		// Test loading anthropic catalog with embedded data
		models, err := service.loadProviderCatalog("anthropic", []byte(`
provider: anthropic
models:
  - name: claude-test
    display_name: Claude Test
    description: Test model
    capabilities: [text, test]
    context_window: 1000
`))
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Equal(t, 1, len(models))

		model := models[0]
		assert.Equal(t, "claude-test", model.Name)
		assert.Equal(t, "Claude Test", model.DisplayName)
		assert.Equal(t, "Test model", model.Description)
		assert.Equal(t, 1000, model.ContextWindow)
		assert.Contains(t, model.Capabilities, "text")
		assert.Contains(t, model.Capabilities, "test")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		models, err := service.loadProviderCatalog("test", []byte("invalid: yaml: ["))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse test catalog")
		assert.Nil(t, models)
	})

	t.Run("empty yaml", func(t *testing.T) {
		models, err := service.loadProviderCatalog("test", []byte(""))
		require.NoError(t, err)
		assert.Equal(t, 0, len(models), "Empty YAML should return empty models slice")
	})
}
