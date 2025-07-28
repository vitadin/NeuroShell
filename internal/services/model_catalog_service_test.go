package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestModelCatalogService_Name(t *testing.T) {
	service := NewModelCatalogService()
	assert.Equal(t, "model_catalog", service.Name())
}

func TestModelCatalogService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  neurotypes.Context
		want error
	}{
		{
			name: "successful initialization",
			ctx:  context.NewTestContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewModelCatalogService()
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

func TestModelCatalogService_GetModelCatalog(t *testing.T) {
	service := NewModelCatalogService()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("successful catalog retrieval", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		require.NoError(t, err)
		require.NotNil(t, models)

		// Verify we get models from both providers
		assert.Greater(t, len(models), 0, "Should have at least one model")

		// Check that we have models from all providers
		hasAnthropic := false
		hasOpenAI := false
		hasGemini := false

		for _, model := range models {
			// Verify model structure
			assert.NotEmpty(t, model.Name, "Model name should not be empty")
			assert.NotEmpty(t, model.DisplayName, "Display name should not be empty")
			assert.NotEmpty(t, model.Description, "Description should not be empty")
			assert.Greater(t, model.ContextWindow, 0, "Context window should be positive")
			assert.NotEmpty(t, model.Capabilities, "Capabilities should not be empty")

			// Check provider types based on model names
			if model.Name == "claude-3-7-sonnet-20250219" || model.Name == "claude-sonnet-4-20250514" ||
				model.Name == "claude-3-7-opus-20240229" || model.Name == "claude-opus-4-20250514" {
				hasAnthropic = true
			}
			if model.Name == "o3" {
				hasOpenAI = true
			}
			if model.Name == "gemini-2.5-pro" || model.Name == "gemini-2.5-flash" || model.Name == "gemini-2.5-flash-lite" {
				hasGemini = true
			}
		}

		assert.True(t, hasAnthropic, "Should have Anthropic models")
		assert.True(t, hasOpenAI, "Should have OpenAI models")
		assert.True(t, hasGemini, "Should have Gemini models")
	})
}

func TestModelCatalogService_GetModelCatalogByProvider(t *testing.T) {
	service := NewModelCatalogService()

	// Test not initialized
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

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

		// Verify all models are OpenAI models
		expectedOpenAIModels := map[string]bool{"o3": true, "o4-mini": true, "gpt-4.1-2025-04-14": true, "o3-pro-2025-06-10": true, "o1-2024-12-17": true, "gpt-4o-2024-11-20": true, "o1-pro-2025-03-19": true}
		for _, model := range models {
			assert.True(t,
				expectedOpenAIModels[model.Name],
				"Should be a known OpenAI model: %s", model.Name)
			assert.NotEmpty(t, model.DisplayName)
			assert.Greater(t, model.ContextWindow, 0)
		}
	})

	t.Run("get gemini models", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("gemini")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should have Gemini models")

		// Verify all models are Gemini models
		expectedGeminiModels := map[string]bool{"gemini-2.5-pro": true, "gemini-2.5-flash": true, "gemini-2.5-flash-lite": true}
		for _, model := range models {
			assert.True(t,
				expectedGeminiModels[model.Name],
				"Should be a known Gemini model: %s", model.Name)
			assert.Equal(t, "gemini", model.Provider, "Provider should be gemini")
			assert.NotEmpty(t, model.DisplayName)
			assert.Greater(t, model.ContextWindow, 0)
		}
	})

	t.Run("get openrouter models", func(t *testing.T) {
		models, err := service.GetModelCatalogByProvider("openrouter")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should have OpenRouter models")

		// Verify all models are OpenRouter models
		expectedOpenRouterModels := map[string]bool{
			"moonshotai/kimi-k2":                 true,
			"moonshotai/kimi-k2:free":            true,
			"qwen/qwen3-235b-a22b-07-25":         true,
			"x-ai/grok-4":                        true,
			"qwen/qwen3-235b-a22b-thinking-2507": true,
			"z-ai/glm-4.5":                       true,
			"google/gemini-2.5-flash-lite":       true,
		}
		for _, model := range models {
			assert.True(t,
				expectedOpenRouterModels[model.Name],
				"Should be a known OpenRouter model: %s", model.Name)
			assert.Equal(t, "openrouter", model.Provider, "Provider should be openrouter")
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

func TestModelCatalogService_SearchModelCatalog(t *testing.T) {
	service := NewModelCatalogService()

	// Test not initialized (indirectly through GetModelCatalog)
	t.Run("not initialized", func(t *testing.T) {
		models, err := service.SearchModelCatalog("gpt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Nil(t, models)
	})

	// Initialize service
	require.NoError(t, service.Initialize())

	t.Run("search by model name", func(t *testing.T) {
		models, err := service.SearchModelCatalog("o3")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should find O3 models")

		// Verify all results contain the search term
		for _, model := range models {
			assert.Contains(t, model.Name, "o3", "Model name should contain search term")
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
		models, err := service.SearchModelCatalog("coding")
		require.NoError(t, err)
		require.NotNil(t, models)
		assert.Greater(t, len(models), 0, "Should find models with coding capabilities")

		// Verify results contain models with coding in description or capabilities
		for _, model := range models {
			searchTerm := "coding"
			assert.True(t,
				strings.Contains(strings.ToLower(model.Name), searchTerm) ||
					strings.Contains(strings.ToLower(model.DisplayName), searchTerm) ||
					strings.Contains(strings.ToLower(model.Description), searchTerm),
				"Model should contain 'coding' in name, display name, or description")
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

func TestModelCatalogService_GetSupportedProviders(t *testing.T) {
	service := NewModelCatalogService()

	// This method doesn't require initialization
	providers := service.GetSupportedProviders()
	require.NotNil(t, providers)
	assert.Contains(t, providers, "anthropic")
	assert.Contains(t, providers, "openai")
	assert.Equal(t, 2, len(providers), "Should have exactly 2 supported providers")
}

func TestModelCatalogService_loadModelFile(t *testing.T) {
	service := NewModelCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("load individual model file", func(t *testing.T) {
		// Test loading individual model file with embedded data
		model, err := service.loadModelFile([]byte(`
name: claude-test
provider: anthropic
display_name: Claude Test
description: Test model
capabilities: [text, test]
context_window: 1000
`))
		require.NoError(t, err)

		assert.Equal(t, "claude-test", model.Name)
		assert.Equal(t, "anthropic", model.Provider)
		assert.Equal(t, "Claude Test", model.DisplayName)
		assert.Equal(t, "Test model", model.Description)
		assert.Equal(t, 1000, model.ContextWindow)
		assert.Contains(t, model.Capabilities, "text")
		assert.Contains(t, model.Capabilities, "test")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		model, err := service.loadModelFile([]byte("invalid: yaml: ["))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse model file")
		assert.Equal(t, neurotypes.ModelCatalogEntry{}, model)
	})

	t.Run("empty yaml", func(t *testing.T) {
		model, err := service.loadModelFile([]byte(""))
		require.NoError(t, err)
		assert.Equal(t, "", model.Name, "Empty YAML should return empty model")
	})
}

func TestModelCatalogService_GetModelByID(t *testing.T) {
	service := NewModelCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("get model by exact ID", func(t *testing.T) {
		model, err := service.GetModelByID("O3")
		require.NoError(t, err)
		assert.Equal(t, "O3", model.ID)
		assert.Equal(t, "o3", model.Name)
		assert.Equal(t, "openai", model.Provider)
	})

	t.Run("get model by case-insensitive ID", func(t *testing.T) {
		testCases := []string{"cs4", "CS4", "Cs4", "cS4"}
		for _, testID := range testCases {
			model, err := service.GetModelByID(testID)
			require.NoError(t, err, "Should find model with ID: %s", testID)
			assert.Equal(t, "CS4", model.ID)
			assert.Equal(t, "claude-sonnet-4-20250514", model.Name)
			assert.Equal(t, "anthropic", model.Provider)
		}
	})

	t.Run("model not found", func(t *testing.T) {
		model, err := service.GetModelByID("NONEXISTENT")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in catalog")
		assert.Equal(t, neurotypes.ModelCatalogEntry{}, model)
	})

	t.Run("empty ID", func(t *testing.T) {
		model, err := service.GetModelByID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in catalog")
		assert.Equal(t, neurotypes.ModelCatalogEntry{}, model)
	})

	t.Run("service not initialized", func(t *testing.T) {
		uninitializedService := NewModelCatalogService()
		model, err := uninitializedService.GetModelByID("O3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
		assert.Equal(t, neurotypes.ModelCatalogEntry{}, model)
	})
}

func TestModelCatalogService_validateUniqueIDs(t *testing.T) {
	service := NewModelCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("unique IDs pass validation", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{
			{ID: "O3", Name: "o3", Provider: "openai"},
			{ID: "CS4", Name: "claude-sonnet-4", Provider: "anthropic"},
			{ID: "CO37", Name: "claude-opus-37", Provider: "anthropic"},
		}
		err := service.validateUniqueIDs(models)
		assert.NoError(t, err)
	})

	t.Run("duplicate IDs fail validation", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{
			{ID: "O3", Name: "o3", Provider: "openai"},
			{ID: "O3", Name: "another-o3", Provider: "openai"},
		}
		err := service.validateUniqueIDs(models)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate model ID found")
		assert.Contains(t, err.Error(), "O3")
	})

	t.Run("case-insensitive duplicate IDs fail validation", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{
			{ID: "O3", Name: "o3", Provider: "openai"},
			{ID: "o3", Name: "another-o3", Provider: "openai"},
		}
		err := service.validateUniqueIDs(models)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate model ID found")
		assert.Contains(t, err.Error(), "case insensitive")
	})

	t.Run("empty ID fails validation", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{
			{ID: "", Name: "empty-id-model", Provider: "test"},
		}
		err := service.validateUniqueIDs(models)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty ID field")
		assert.Contains(t, err.Error(), "empty-id-model")
	})

	t.Run("empty models list passes validation", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{}
		err := service.validateUniqueIDs(models)
		assert.NoError(t, err)
	})
}

func TestModelCatalogService_normalizeID(t *testing.T) {
	service := NewModelCatalogService()

	testCases := []struct {
		input    string
		expected string
	}{
		{"O3", "O3"},
		{"o3", "O3"},
		{"Cs4", "CS4"},
		{"cs4", "CS4"},
		{"CS4", "CS4"},
		{"cS4", "CS4"},
		{"CO37", "CO37"},
		{"co37", "CO37"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := service.normalizeID(tc.input)
		assert.Equal(t, tc.expected, result, "normalizeID(%s) should return %s", tc.input, tc.expected)
	}
}

func TestModelCatalogService_IDValidationIntegration(t *testing.T) {
	// This test ensures that the actual embedded YAML files have unique IDs
	service := NewModelCatalogService()
	require.NoError(t, service.Initialize())

	t.Run("real catalog has unique IDs", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		require.NoError(t, err)
		assert.Greater(t, len(models), 0, "Should have models in catalog")

		// Verify all models have IDs
		for _, model := range models {
			assert.NotEmpty(t, model.ID, "Model %s should have non-empty ID", model.Name)
		}

		// Verify IDs are unique (this is tested by GetModelCatalog calling validateUniqueIDs)
		seenIDs := make(map[string]bool)
		for _, model := range models {
			normalizedID := strings.ToUpper(model.ID)
			assert.False(t, seenIDs[normalizedID], "ID %s should be unique (case-insensitive)", model.ID)
			seenIDs[normalizedID] = true
		}
	})

	t.Run("expected model IDs exist", func(t *testing.T) {
		expectedIDs := []string{"O3", "O4M", "CS37", "CS4", "CO37", "CO4"}

		for _, expectedID := range expectedIDs {
			model, err := service.GetModelByID(expectedID)
			require.NoError(t, err, "Should find model with ID: %s", expectedID)
			assert.Equal(t, expectedID, model.ID)
			assert.NotEmpty(t, model.Name, "Model should have name")
			assert.NotEmpty(t, model.Provider, "Model should have provider")
		}
	})

	t.Run("case-insensitive lookup works for all models", func(t *testing.T) {
		models, err := service.GetModelCatalog()
		require.NoError(t, err)

		for _, model := range models {
			// Test lowercase version
			lowerModel, err := service.GetModelByID(strings.ToLower(model.ID))
			require.NoError(t, err, "Should find model with lowercase ID: %s", strings.ToLower(model.ID))
			assert.Equal(t, model.ID, lowerModel.ID)
			assert.Equal(t, model.Name, lowerModel.Name)
		}
	})
}
