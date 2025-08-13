package provider

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestProviderCatalogCommand_Name(t *testing.T) {
	cmd := &CatalogCommand{}
	assert.Equal(t, "provider-catalog", cmd.Name())
}

func TestProviderCatalogCommand_ParseMode(t *testing.T) {
	cmd := &CatalogCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestProviderCatalogCommand_Description(t *testing.T) {
	cmd := &CatalogCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "llm")
	assert.Contains(t, strings.ToLower(desc), "provider")
	assert.Contains(t, strings.ToLower(desc), "catalog")
}

func TestProviderCatalogCommand_Usage(t *testing.T) {
	cmd := &CatalogCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\provider-catalog")
	assert.Contains(t, usage, "provider=")
	assert.Contains(t, usage, "sort=")
	assert.Contains(t, usage, "search=")
	assert.Contains(t, usage, "openai")
	assert.Contains(t, usage, "anthropic")
}

func TestProviderCatalogCommand_HelpInfo(t *testing.T) {
	cmd := &CatalogCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "provider-catalog", helpInfo.Command)
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

func TestProviderCatalogCommand_Execute(t *testing.T) {
	// Setup services and context
	registry := services.NewRegistry()

	// Register required services
	require.NoError(t, registry.RegisterService(services.NewProviderCatalogService()))
	require.NoError(t, registry.RegisterService(services.NewVariableService()))
	require.NoError(t, registry.RegisterService(services.NewThemeService()))
	require.NoError(t, registry.InitializeAll())
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
		assert.Contains(t, output, "Provider Catalog")
		assert.Contains(t, output, "providers)")
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
		assert.Contains(t, output, "OAC")
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
		assert.Contains(t, output, "ANC")
	})

	t.Run("search functionality", func(t *testing.T) {
		args := map[string]string{"search": "chat"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Search: 'chat'")
		assert.Contains(t, output, "Chat")
	})

	t.Run("sort by name", func(t *testing.T) {
		args := map[string]string{"sort": "name"}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Provider Catalog")
		// With name sorting, providers should be alphabetically ordered
		assert.NotEmpty(t, output)
	})

	t.Run("combined options", func(t *testing.T) {
		args := map[string]string{
			"provider": "openai",
			"sort":     "name",
			"search":   "openai",
		}
		err := cmd.Execute(args, "")
		require.NoError(t, err)

		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		output, err := variableService.Get("_output")
		require.NoError(t, err)
		assert.Contains(t, output, "Openai")
		assert.Contains(t, output, "Search: 'openai'")
		assert.Contains(t, output, "OpenAI")
	})
}

func TestProviderCatalogCommand_validateArguments(t *testing.T) {
	cmd := &CatalogCommand{}

	t.Run("valid providers", func(t *testing.T) {
		validProviders := []string{"all", "openai", "anthropic", "gemini"}
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

func TestProviderCatalogCommand_filterProvidersBySearch(t *testing.T) {
	cmd := &CatalogCommand{}

	// Create test providers
	providers := []neurotypes.ProviderCatalogEntry{
		{
			ID:          "OAC",
			Provider:    "openai",
			DisplayName: "OpenAI Chat Completions",
			Description: "OpenAI's chat completion API for GPT models",
		},
		{
			ID:          "ANC",
			Provider:    "anthropic",
			DisplayName: "Anthropic Claude Chat",
			Description: "Anthropic Claude chat completions API",
		},
		{
			ID:          "GMC",
			Provider:    "gemini",
			DisplayName: "Google Gemini Chat",
			Description: "Google Gemini generative AI API",
		},
	}

	t.Run("search by ID", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "OAC")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "OAC", results[0].ID)
	})

	t.Run("search by provider", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "openai")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "openai", results[0].Provider)
	})

	t.Run("search by display name", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "Claude")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "ANC", results[0].ID)
	})

	t.Run("search by description", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "API")
		require.NoError(t, err)
		assert.Equal(t, 3, len(results)) // All have "API" in description
	})

	t.Run("case insensitive search", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "OPENAI")
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "openai", results[0].Provider)
	})

	t.Run("no matches", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, 0, len(results))
	})

	t.Run("empty query returns all", func(t *testing.T) {
		results, err := cmd.filterProvidersBySearch(providers, "")
		require.NoError(t, err)
		assert.Equal(t, 3, len(results))
	})
}

func TestProviderCatalogCommand_sortProviders(t *testing.T) {
	cmd := &CatalogCommand{}

	// Create test providers
	providers := []neurotypes.ProviderCatalogEntry{
		{ID: "OAC", Provider: "openai", DisplayName: "OpenAI Chat Completions"},
		{ID: "ANC", Provider: "anthropic", DisplayName: "Anthropic Claude Chat"},
	}

	t.Run("sort by name", func(t *testing.T) {
		testProviders := make([]neurotypes.ProviderCatalogEntry, len(providers))
		copy(testProviders, providers)

		cmd.sortProviders(testProviders, "name", "all")

		// Should be sorted alphabetically by display name
		assert.Equal(t, "Anthropic Claude Chat", testProviders[0].DisplayName)
		assert.Equal(t, "OpenAI Chat Completions", testProviders[1].DisplayName)
	})

	t.Run("sort by provider", func(t *testing.T) {
		testProviders := make([]neurotypes.ProviderCatalogEntry, len(providers))
		copy(testProviders, providers)

		cmd.sortProviders(testProviders, "provider", "all")

		// Should be sorted by provider then by name
		// Anthropic providers first, then openai
		assert.Contains(t, testProviders[0].Provider, "anthropic")
		assert.Contains(t, testProviders[1].Provider, "openai")
	})
}

func TestProviderCatalogCommand_formatProviderCatalog(t *testing.T) {
	cmd := &CatalogCommand{}

	// Create a theme service for testing
	themeService := services.NewThemeService()
	themeObj := themeService.GetThemeByName("plain") // Use plain theme for predictable test output

	t.Run("empty providers list", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{}
		result := cmd.formatProviderCatalog(providers, "all", "provider", "", themeObj)
		assert.Contains(t, result, "No providers found")
	})

	t.Run("with search query", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{}
		result := cmd.formatProviderCatalog(providers, "all", "provider", "openai", themeObj)
		assert.Contains(t, result, "No providers found")
		assert.Contains(t, result, "matching 'openai'")
	})

	t.Run("with specific provider", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{}
		result := cmd.formatProviderCatalog(providers, "openai", "provider", "", themeObj)
		assert.Contains(t, result, "No providers found")
		assert.Contains(t, result, "from openai")
	})

	t.Run("with providers", func(t *testing.T) {
		providers := []neurotypes.ProviderCatalogEntry{
			{
				ID:          "OAC",
				Provider:    "openai",
				DisplayName: "OpenAI Chat Completions",
				Description: "OpenAI's chat completion API",
				BaseURL:     "https://api.openai.com/v1",
				ClientType:  "openai",
			},
		}
		result := cmd.formatProviderCatalog(providers, "all", "provider", "", themeObj)
		assert.Contains(t, result, "Provider Catalog")
		assert.Contains(t, result, "(1 providers)")
		assert.Contains(t, result, "[OAC] OpenAI Chat Completions (openai)")
		assert.Contains(t, result, "https://api.openai.com/v1")
		assert.Contains(t, result, "OpenAI's chat completion API")
	})
}

func TestProviderCatalogCommand_toTitle(t *testing.T) {
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
