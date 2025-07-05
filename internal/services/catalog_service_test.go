package services

import (
	"strings"
	"testing"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestCatalogService_Name(t *testing.T) {
	service := NewCatalogService()
	expected := "catalog"
	actual := service.Name()

	if actual != expected {
		t.Errorf("Expected service name %q, got %q", expected, actual)
	}
}

func TestCatalogService_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		setupService  func() *CatalogService
		expectedError bool
	}{
		{
			name:          "successful initialization",
			setupService:  NewCatalogService,
			expectedError: false,
		},
		{
			name: "multiple initializations should not fail",
			setupService: func() *CatalogService {
				service := NewCatalogService()
				ctx := testutils.NewMockContext()
				_ = service.Initialize(ctx) // First initialization
				return service
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			ctx := testutils.NewMockContext()

			err := service.Initialize(ctx)

			if tt.expectedError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			// Verify service is initialized
			if !service.initialized {
				t.Error("Service should be marked as initialized")
			}

			// Verify catalog data is loaded
			if service.catalog == nil {
				t.Error("Catalog should be loaded after initialization")
			}

			// Verify models are flattened
			if len(service.allModels) == 0 {
				t.Error("All models should be flattened after initialization")
			}
		})
	}
}

func TestCatalogService_SearchModels(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	tests := []struct {
		name           string
		options        neurotypes.CatalogSearchOptions
		expectedCount  int
		expectedErrors bool
		validateResult func(t *testing.T, result *neurotypes.CatalogSearchResult)
	}{
		{
			name:          "search all models",
			options:       neurotypes.CatalogSearchOptions{},
			expectedCount: 8, // Based on our test catalog data
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				if result.Count != len(result.Models) {
					t.Errorf("Count mismatch: count=%d, models=%d", result.Count, len(result.Models))
				}
				if result.QueryTime <= 0 {
					t.Error("Query time should be positive")
				}
			},
		},
		{
			name: "filter by openai provider",
			options: neurotypes.CatalogSearchOptions{
				Provider: "openai",
			},
			expectedCount: 3, // gpt-3.5-turbo, gpt-4, gpt-4-turbo
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				if result.Provider != "openai" {
					t.Errorf("Expected provider 'openai', got '%s'", result.Provider)
				}
				for _, model := range result.Models {
					if model.Provider != "openai" {
						t.Errorf("Expected all models to be from openai, got %s", model.Provider)
					}
				}
			},
		},
		{
			name: "filter by anthropic provider",
			options: neurotypes.CatalogSearchOptions{
				Provider: "anthropic",
			},
			expectedCount: 3, // claude-3-haiku, claude-3-sonnet, claude-3-opus
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				if result.Provider != "anthropic" {
					t.Errorf("Expected provider 'anthropic', got '%s'", result.Provider)
				}
			},
		},
		{
			name: "search by pattern gpt",
			options: neurotypes.CatalogSearchOptions{
				Pattern: "gpt",
			},
			expectedCount: 3, // All OpenAI models contain "gpt"
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				for _, model := range result.Models {
					if !strings.Contains(strings.ToLower(model.ID+" "+model.Name+" "+model.Description), "gpt") {
						t.Errorf("Model %s should match pattern 'gpt'", model.ID)
					}
				}
			},
		},
		{
			name: "search by specific pattern",
			options: neurotypes.CatalogSearchOptions{
				Pattern: "gpt-4-turbo",
			},
			expectedCount: 4, // Fuzzy search finds multiple matches containing "turbo"
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				// Should find models with "turbo" in name/description
				if result.Count == 0 {
					t.Error("Should find models matching turbo pattern")
				}
				// With multiple results, ModelID should be empty
				if result.ModelID != "" {
					t.Errorf("Expected empty ModelID for multiple results, got '%s'", result.ModelID)
				}
			},
		},
		{
			name: "combined provider and pattern filter",
			options: neurotypes.CatalogSearchOptions{
				Provider: "anthropic",
				Pattern:  "sonnet",
			},
			expectedCount: 2, // Fuzzy search finds multiple matches with shared characters
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				// Should find sonnet with higher score than opus
				foundSonnet := false
				for _, model := range result.Models {
					if model.ID == "claude-3-sonnet" {
						foundSonnet = true
						break
					}
				}
				if !foundSonnet {
					t.Error("Should find claude-3-sonnet in results")
				}
			},
		},
		{
			name: "no matches",
			options: neurotypes.CatalogSearchOptions{
				Pattern: "qwertyuiopasdfghjklzxcvbnm",
			},
			expectedCount: 0,
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				if result.ModelID != "" {
					t.Error("ModelID should be empty for no results")
				}
				if result.Provider != "" {
					t.Error("Provider should be empty for no results")
				}
			},
		},
		{
			name: "sort by context length",
			options: neurotypes.CatalogSearchOptions{
				Sort: "context_length",
			},
			expectedCount: 8,
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				// Should be sorted by context length descending
				for i := 1; i < len(result.Models); i++ {
					if result.Models[i].ContextLength > result.Models[i-1].ContextLength {
						t.Errorf("Models not sorted by context length: %d > %d",
							result.Models[i].ContextLength, result.Models[i-1].ContextLength)
					}
				}
			},
		},
		{
			name: "sort by provider",
			options: neurotypes.CatalogSearchOptions{
				Sort: "provider",
			},
			expectedCount: 8,
			validateResult: func(t *testing.T, result *neurotypes.CatalogSearchResult) {
				// Should be sorted by provider alphabetically
				for i := 1; i < len(result.Models); i++ {
					if result.Models[i].Provider < result.Models[i-1].Provider {
						t.Errorf("Models not sorted by provider: %s < %s",
							result.Models[i].Provider, result.Models[i-1].Provider)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.SearchModels(tt.options)

			if tt.expectedErrors && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedErrors && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if result.Count != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, result.Count)
			}

			if len(result.Models) != tt.expectedCount {
				t.Errorf("Expected %d models, got %d", tt.expectedCount, len(result.Models))
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestCatalogService_SearchModels_NotInitialized(t *testing.T) {
	service := NewCatalogService()
	options := neurotypes.CatalogSearchOptions{}

	result, err := service.SearchModels(options)

	if err == nil {
		t.Error("Expected error for uninitialized service")
	}
	if result != nil {
		t.Error("Result should be nil for uninitialized service")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Error should mention not initialized, got: %v", err)
	}
}

func TestCatalogService_GetModel(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	tests := []struct {
		name          string
		provider      string
		modelID       string
		expectedError bool
		expectedModel string
	}{
		{
			name:          "get existing openai model",
			provider:      "openai",
			modelID:       "gpt-4",
			expectedError: false,
			expectedModel: "gpt-4",
		},
		{
			name:          "get existing anthropic model",
			provider:      "anthropic",
			modelID:       "claude-3-sonnet",
			expectedError: false,
			expectedModel: "claude-3-sonnet",
		},
		{
			name:          "get nonexistent model",
			provider:      "openai",
			modelID:       "nonexistent",
			expectedError: true,
		},
		{
			name:          "get model from wrong provider",
			provider:      "anthropic",
			modelID:       "gpt-4",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := service.GetModel(tt.provider, tt.modelID)

			if tt.expectedError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tt.expectedError {
				if model == nil {
					t.Fatal("Model should not be nil")
				}
				if model.ID != tt.expectedModel {
					t.Errorf("Expected model ID %s, got %s", tt.expectedModel, model.ID)
				}
				if model.Provider != tt.provider {
					t.Errorf("Expected provider %s, got %s", tt.provider, model.Provider)
				}
			}
		})
	}
}

func TestCatalogService_GetProviders(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	providers, err := service.GetProviders()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedProviders := []string{"anthropic", "local", "openai"}
	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	// Check if all expected providers are present (order doesn't matter due to sorting)
	providerMap := make(map[string]bool)
	for _, provider := range providers {
		providerMap[provider] = true
	}

	for _, expected := range expectedProviders {
		if !providerMap[expected] {
			t.Errorf("Expected provider %s not found", expected)
		}
	}
}

func TestCatalogService_IsValidModel(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	tests := []struct {
		name     string
		provider string
		modelID  string
		expected bool
	}{
		{"valid openai model", "openai", "gpt-4", true},
		{"valid anthropic model", "anthropic", "claude-3-sonnet", true},
		{"invalid model", "openai", "nonexistent", false},
		{"wrong provider", "anthropic", "gpt-4", false},
		{"empty provider", "", "gpt-4", false},
		{"empty model", "openai", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsValidModel(tt.provider, tt.modelID)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCatalogService_SuggestModels(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	tests := []struct {
		name           string
		provider       string
		pattern        string
		maxSuggestions int
		expectedMin    int
		expectedMax    int
		shouldContain  []string
	}{
		{
			name:           "suggest gpt models",
			provider:       "openai",
			pattern:        "gpt",
			maxSuggestions: 5,
			expectedMin:    1,
			expectedMax:    3,
			shouldContain:  []string{"gpt-4", "gpt-3.5-turbo"},
		},
		{
			name:           "suggest claude models",
			provider:       "anthropic",
			pattern:        "claude",
			maxSuggestions: 5,
			expectedMin:    1,
			expectedMax:    3,
			shouldContain:  []string{"claude-3-sonnet"},
		},
		{
			name:           "suggest with typo",
			provider:       "openai",
			pattern:        "gp4",
			maxSuggestions: 5,
			expectedMin:    0,
			expectedMax:    3,
		},
		{
			name:           "suggest all providers",
			provider:       "",
			pattern:        "turbo",
			maxSuggestions: 5,
			expectedMin:    1,
			expectedMax:    4, // Fuzzy search may find more models with "turbo" in descriptions
		},
		{
			name:           "no matches",
			provider:       "openai",
			pattern:        "xyz123",
			maxSuggestions: 5,
			expectedMin:    0,
			expectedMax:    0,
		},
		{
			name:           "limit suggestions",
			provider:       "",
			pattern:        "model",
			maxSuggestions: 2,
			expectedMin:    0,
			expectedMax:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := service.SuggestModels(tt.provider, tt.pattern, tt.maxSuggestions)

			if len(suggestions) < tt.expectedMin {
				t.Errorf("Expected at least %d suggestions, got %d", tt.expectedMin, len(suggestions))
			}
			if len(suggestions) > tt.expectedMax {
				t.Errorf("Expected at most %d suggestions, got %d", tt.expectedMax, len(suggestions))
			}

			// Check that expected models are contained in suggestions
			for _, expected := range tt.shouldContain {
				found := false
				for _, suggestion := range suggestions {
					if strings.Contains(suggestion, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected suggestions to contain %s, got %v", expected, suggestions)
				}
			}
		})
	}
}

func TestCatalogService_AutoCreateModelFromCatalog(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Get a specific test model from the catalog
	catalogModel, err := service.GetModel("openai", "gpt-4-turbo")
	if err != nil {
		t.Fatalf("Failed to get test model: %v", err)
	}

	tests := []struct {
		name      string
		modelName string
		validate  func(t *testing.T, config *neurotypes.ModelConfig)
	}{
		{
			name:      "create with simple name",
			modelName: "my-gpt4-turbo",
			validate: func(t *testing.T, config *neurotypes.ModelConfig) {
				if config.Name != "my-gpt4-turbo" {
					t.Errorf("Expected name 'my-gpt4-turbo', got %s", config.Name)
				}
				if config.Provider != "openai" {
					t.Errorf("Expected provider 'openai', got %s", config.Provider)
				}
				if config.BaseModel != "gpt-4-turbo" {
					t.Errorf("Expected base model 'gpt-4-turbo', got %s", config.BaseModel)
				}
				if !strings.Contains(config.Description, "Auto-created from catalog") {
					t.Errorf("Description should mention auto-creation: %s", config.Description)
				}
				if config.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if config.ID == "" {
					t.Error("ID should be generated")
				}
			},
		},
		{
			name:      "create with complex name",
			modelName: "custom-model-for-testing",
			validate: func(t *testing.T, config *neurotypes.ModelConfig) {
				if config.Name != "custom-model-for-testing" {
					t.Errorf("Expected name preserved")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := service.AutoCreateModelFromCatalog(*catalogModel, tt.modelName)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if config == nil {
				t.Fatal("Config should not be nil")
			}

			tt.validate(t, config)
		})
	}
}

func TestCatalogService_GenerateAutoModelName(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Get a test model
	catalogModel := neurotypes.CatalogModel{
		ID:       "gpt-4",
		Name:     "GPT-4",
		Provider: "openai",
	}

	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "simple pattern",
			pattern:  "gpt4",
			expected: "gpt4-auto",
		},
		{
			name:     "pattern with spaces",
			pattern:  "my gpt model",
			expected: "my-gpt-model-auto",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: "openai-gpt-4-auto",
		},
		{
			name:     "very long pattern",
			pattern:  strings.Repeat("a", 100),
			expected: "openai-gpt-4-auto", // Should fall back to descriptive name
		},
		{
			name:     "pattern with newlines",
			pattern:  "test\nmodel",
			expected: "openai-gpt-4-auto", // Should fall back to descriptive name
		},
		{
			name:     "reasonable pattern",
			pattern:  "my-custom-gpt",
			expected: "my-custom-gpt-auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateAutoModelName(tt.pattern, catalogModel)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCatalogService_CalculateSimilarityScore(t *testing.T) {
	service := NewCatalogService()

	tests := []struct {
		name     string
		text     string
		pattern  string
		expected int
	}{
		{"exact match", "gpt-4", "gpt-4", 100},
		{"starts with", "gpt-4-turbo", "gpt", 90},
		{"contains", "my-gpt-model", "gpt", 80},
		{"ends with", "model-gpt", "gpt", 70},
		{"no similarity", "claude", "gpt", 0},
		{"partial chars", "gpt4", "gpt", 90}, // "gpt4" starts with "gpt"
		{"empty pattern", "gpt-4", "", 0},
		{"empty text", "", "gpt", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateSimilarityScore(tt.text, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCatalogService_GetCatalogInfo(t *testing.T) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	catalog, err := service.GetCatalogInfo()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		return
	}
	if catalog == nil {
		t.Error("Catalog should not be nil")
		return
	}
	if catalog.Version == "" {
		t.Error("Catalog version should be set")
	}
	if len(catalog.Providers) == 0 {
		t.Error("Catalog should have providers")
	}
}

func TestCatalogService_GetCatalogInfo_NotInitialized(t *testing.T) {
	service := NewCatalogService()

	catalog, err := service.GetCatalogInfo()
	if err == nil {
		t.Error("Expected error for uninitialized service")
	}
	if catalog != nil {
		t.Error("Catalog should be nil for uninitialized service")
	}
}

// Benchmark tests for performance
func BenchmarkCatalogService_SearchModels(b *testing.B) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		b.Fatalf("Failed to initialize service: %v", err)
	}

	options := neurotypes.CatalogSearchOptions{
		Provider: "openai",
		Pattern:  "gpt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.SearchModels(options)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

func BenchmarkCatalogService_SuggestModels(b *testing.B) {
	service := NewCatalogService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	if err != nil {
		b.Fatalf("Failed to initialize service: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.SuggestModels("openai", "gp", 5)
	}
}
