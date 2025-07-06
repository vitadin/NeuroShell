package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestCatalogCommand_Name(t *testing.T) {
	cmd := &CatalogCommand{}
	assert.Equal(t, "model-catalog", cmd.Name())
}

func TestCatalogCommand_ParseMode(t *testing.T) {
	cmd := &CatalogCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestCatalogCommand_Description(t *testing.T) {
	cmd := &CatalogCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "llm")
	assert.Contains(t, strings.ToLower(desc), "model")
	assert.Contains(t, strings.ToLower(desc), "catalog")
}

func TestCatalogCommand_Usage(t *testing.T) {
	cmd := &CatalogCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\model-catalog")
	assert.Contains(t, usage, "provider=")
	assert.Contains(t, usage, "sort=")
	assert.Contains(t, usage, "search=")
	assert.Contains(t, usage, "openai")
	assert.Contains(t, usage, "anthropic")
}

func TestCatalogCommand_HelpInfo(t *testing.T) {
	cmd := &CatalogCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "model-catalog", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	require.Greater(t, len(helpInfo.Options), 0)
	optionNames := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionNames[option.Name] = true
	}
	assert.True(t, optionNames["provider"], "Should have provider option")
	assert.True(t, optionNames["sort"], "Should have sort option")
	assert.True(t, optionNames["search"], "Should have search option")

	// Check examples
	assert.Greater(t, len(helpInfo.Examples), 0, "Should have usage examples")

	// Check notes
	assert.Greater(t, len(helpInfo.Notes), 0, "Should have helpful notes")
}

func TestCatalogCommand_Execute(t *testing.T) {
	// Setup services and context
	ctx := context.New()
	registry := services.NewRegistry()

	// Register required services
	require.NoError(t, registry.RegisterService(services.NewModelService()))
	require.NoError(t, registry.RegisterService(services.NewVariableService()))
	require.NoError(t, registry.InitializeAll(ctx))
	services.SetGlobalRegistry(registry)

	cmd := &CatalogCommand{}

	t.Run("basic catalog listing", func(t *testing.T) {
		args := map[string]string{}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		// Check that _output variable was set
		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Model Catalog")
		assert.Contains(t, output, "models)")
	})

	t.Run("filter by provider - openai", func(t *testing.T) {
		args := map[string]string{"provider": "openai"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Openai")
		assert.Contains(t, output, "O3")
	})

	t.Run("filter by provider - anthropic", func(t *testing.T) {
		args := map[string]string{"provider": "anthropic"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Anthropic")
		assert.Contains(t, output, "Claude")
	})

	t.Run("search functionality", func(t *testing.T) {
		args := map[string]string{"search": "o3"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Search: 'o3'")
		assert.Contains(t, output, "O3")
	})

	t.Run("sort by name", func(t *testing.T) {
		args := map[string]string{"sort": "name"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Model Catalog")
		// With name sorting, models should be alphabetically ordered
		assert.NotEmpty(t, output)
	})

	t.Run("combined options", func(t *testing.T) {
		args := map[string]string{
			"provider": "anthropic",
			"sort":     "name",
			"search":   "claude",
		}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Anthropic")
		assert.Contains(t, output, "Search: 'claude'")
		assert.Contains(t, output, "Claude")
	})
}

func TestCatalogCommand_validateArguments(t *testing.T) {
	cmd := &CatalogCommand{}

	t.Run("valid providers", func(t *testing.T) {
		validProviders := []string{"all", "openai", "anthropic"}
		for _, provider := range validProviders {
			err := cmd.validateArguments(provider, "name")
			assert.NoError(t, err, "Provider %s should be valid", provider)
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		err := cmd.validateArguments("invalid", "name")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid provider option")
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("valid sort options", func(t *testing.T) {
		validSorts := []string{"name", "provider"}
		for _, sort := range validSorts {
			err := cmd.validateArguments("all", sort)
			assert.NoError(t, err, "Sort option %s should be valid", sort)
		}
	})

	t.Run("invalid sort option", func(t *testing.T) {
		err := cmd.validateArguments("all", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sort option")
		assert.Contains(t, err.Error(), "invalid")
	})
}

func TestCatalogCommand_filterModelsBySearch(t *testing.T) {
	cmd := &CatalogCommand{}

	// Create test models
	models := []neurotypes.ModelCatalogEntry{
		{
			Name:         "gpt-4",
			DisplayName:  "GPT-4",
			Description:  "Advanced language model",
			Capabilities: []string{"text", "reasoning"},
		},
		{
			Name:         "claude-3-opus",
			DisplayName:  "Claude 3 Opus",
			Description:  "Powerful AI assistant",
			Capabilities: []string{"text", "analysis"},
		},
		{
			Name:         "text-embedding-ada-002",
			DisplayName:  "Ada Embeddings",
			Description:  "Text embedding model",
			Capabilities: []string{"embeddings"},
		},
	}

	t.Run("search by name", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "gpt-4")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "gpt-4", results[0].Name)
	})

	t.Run("search by display name", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "Claude")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "claude-3-opus", results[0].Name)
	})

	t.Run("search by description", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "embedding")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "text-embedding-ada-002", results[0].Name)
	})

	t.Run("case insensitive search", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "GPT")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "gpt-4", results[0].Name)
	})

	t.Run("no matches", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, 0, len(results))
	})

	t.Run("empty query returns all", func(t *testing.T) {
		results, err := cmd.filterModelsBySearch(models, "")
		require.NoError(t, err)
		assert.Equal(t, 3, len(results))
	})
}

func TestCatalogCommand_sortModels(t *testing.T) {
	cmd := &CatalogCommand{}

	// Create test models
	models := []neurotypes.ModelCatalogEntry{
		{Name: "gpt-4", DisplayName: "GPT-4"},
		{Name: "claude-3-opus", DisplayName: "Claude 3 Opus"},
		{Name: "gpt-3.5-turbo", DisplayName: "GPT-3.5 Turbo"},
	}

	t.Run("sort by name", func(t *testing.T) {
		testModels := make([]neurotypes.ModelCatalogEntry, len(models))
		copy(testModels, models)

		cmd.sortModels(testModels, "name", "all")

		// Should be sorted alphabetically by display name
		assert.Equal(t, "Claude 3 Opus", testModels[0].DisplayName)
		assert.Equal(t, "GPT-3.5 Turbo", testModels[1].DisplayName)
		assert.Equal(t, "GPT-4", testModels[2].DisplayName)
	})

	t.Run("sort by provider", func(t *testing.T) {
		testModels := make([]neurotypes.ModelCatalogEntry, len(models))
		copy(testModels, models)

		cmd.sortModels(testModels, "provider", "all")

		// Should be sorted by provider then by name
		// Anthropic models first (claude), then OpenAI models (gpt)
		assert.Contains(t, testModels[0].Name, "claude")
		assert.Contains(t, testModels[1].Name, "gpt")
		assert.Contains(t, testModels[2].Name, "gpt")
	})
}

func TestCatalogCommand_getProviderFromModel(t *testing.T) {
	cmd := &CatalogCommand{}

	testCases := []struct {
		modelName        string
		expectedProvider string
	}{
		{"gpt-4", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"text-embedding-ada-002", "openai"},
		{"claude-3-opus", "anthropic"},
		{"claude-3-5-sonnet", "anthropic"},
		{"unknown-model", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			model := neurotypes.ModelCatalogEntry{Name: tc.modelName}
			provider := cmd.getProviderFromModel(model)
			assert.Equal(t, tc.expectedProvider, provider)
		})
	}
}

func TestCatalogCommand_formatNumber(t *testing.T) {
	cmd := &CatalogCommand{}

	testCases := []struct {
		input    int
		expected string
	}{
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1000000, "1,000,000"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := cmd.formatNumber(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCatalogCommand_formatModelCatalog(t *testing.T) {
	cmd := &CatalogCommand{}

	t.Run("empty models list", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{}
		result := cmd.formatModelCatalog(models, "all", "provider", "")
		assert.Contains(t, result, "No models found")
	})

	t.Run("with search query", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{}
		result := cmd.formatModelCatalog(models, "all", "provider", "gpt-4")
		assert.Contains(t, result, "No models found")
		assert.Contains(t, result, "matching 'gpt-4'")
	})

	t.Run("with specific provider", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{}
		result := cmd.formatModelCatalog(models, "openai", "provider", "")
		assert.Contains(t, result, "No models found")
		assert.Contains(t, result, "from openai")
	})

	t.Run("with models", func(t *testing.T) {
		models := []neurotypes.ModelCatalogEntry{
			{
				Name:          "gpt-4",
				DisplayName:   "GPT-4",
				Description:   "Advanced model",
				Capabilities:  []string{"text", "reasoning"},
				ContextWindow: 8192,
			},
		}
		result := cmd.formatModelCatalog(models, "all", "provider", "")
		assert.Contains(t, result, "Model Catalog")
		assert.Contains(t, result, "(1 models)")
		assert.Contains(t, result, "GPT-4 (gpt-4)")
		assert.Contains(t, result, "8,192 tokens")
		assert.Contains(t, result, "Advanced model")
	})
}

func TestCatalogCommand_toTitle(t *testing.T) {
	cmd := &CatalogCommand{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"openai", "Openai"},
		{"anthropic", "Anthropic"},
		{"", ""},
		{"a", "A"},
		{"test", "Test"},
		{"UPPER", "UPPER"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := cmd.toTitle(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
