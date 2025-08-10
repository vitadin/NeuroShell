package services

import (
	"fmt"
	"strings"

	"neuroshell/internal/data/embedded"
	"neuroshell/pkg/neurotypes"

	"gopkg.in/yaml.v3"
)

// ModelCatalogService provides model catalog operations for NeuroShell.
// It handles loading and searching through embedded YAML model catalogs
// from different LLM providers (OpenAI, Anthropic).
type ModelCatalogService struct {
	initialized bool
}

// NewModelCatalogService creates a new ModelCatalogService instance.
func NewModelCatalogService() *ModelCatalogService {
	return &ModelCatalogService{
		initialized: false,
	}
}

// Name returns the service name "model_catalog" for registration.
func (m *ModelCatalogService) Name() string {
	return "model_catalog"
}

// Initialize sets up the ModelCatalogService for operation.
func (m *ModelCatalogService) Initialize() error {
	m.initialized = true
	return nil
}

// GetModelCatalog returns the complete model catalog from all providers.
func (m *ModelCatalogService) GetModelCatalog() ([]neurotypes.ModelCatalogEntry, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model catalog service not initialized")
	}

	var allModels []neurotypes.ModelCatalogEntry

	// Load OpenAI models
	o3Model, err := m.loadModelFile(embedded.O3ModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load O3 model: %w", err)
	}
	allModels = append(allModels, o3Model)

	o4MiniChatModel, err := m.loadModelFile(embedded.O4MiniChatModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load O4-mini Chat model: %w", err)
	}
	allModels = append(allModels, o4MiniChatModel)

	o4MiniReasoningModel, err := m.loadModelFile(embedded.O4MiniReasoningModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load O4-mini Reasoning model: %w", err)
	}
	allModels = append(allModels, o4MiniReasoningModel)

	// Load Anthropic models
	claude37SonnetModel, err := m.loadModelFile(embedded.Claude37SonnetModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Claude 3.7 Sonnet model: %w", err)
	}
	allModels = append(allModels, claude37SonnetModel)

	claudeSonnet4Model, err := m.loadModelFile(embedded.ClaudeSonnet4ModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Claude Sonnet 4 model: %w", err)
	}
	allModels = append(allModels, claudeSonnet4Model)

	claude37OpusModel, err := m.loadModelFile(embedded.Claude37OpusModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Claude 3.7 Opus model: %w", err)
	}
	allModels = append(allModels, claude37OpusModel)

	claudeOpus4Model, err := m.loadModelFile(embedded.ClaudeOpus4ModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Claude Opus 4 model: %w", err)
	}
	allModels = append(allModels, claudeOpus4Model)

	// Load Kimi K2 models
	kimiK2FreeModel, err := m.loadModelFile(embedded.KimiK2FreeOpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kimi K2 Free (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, kimiK2FreeModel)

	kimiK2Model, err := m.loadModelFile(embedded.KimiK2OpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kimi K2 (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, kimiK2Model)

	kimiK2MoonshotModel, err := m.loadModelFile(embedded.KimiK2MoonshotModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kimi K2 (Moonshot) model: %w", err)
	}
	allModels = append(allModels, kimiK2MoonshotModel)

	qwen3235BModel, err := m.loadModelFile(embedded.Qwen3235BOpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Qwen3-235B (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, qwen3235BModel)

	grok4Model, err := m.loadModelFile(embedded.Grok4OpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Grok-4 (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, grok4Model)

	qwen3235BA22BThinkingModel, err := m.loadModelFile(embedded.Qwen3235BA22BThinkingOpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Qwen3 235B A22B Thinking (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, qwen3235BA22BThinkingModel)

	glm45Model, err := m.loadModelFile(embedded.GLM45OpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load GLM 4.5 (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, glm45Model)

	gemini25FlashLiteOpenRouterModel, err := m.loadModelFile(embedded.Gemini25FlashLiteOpenRouterModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gemini 2.5 Flash Lite (OpenRouter) model: %w", err)
	}
	allModels = append(allModels, gemini25FlashLiteOpenRouterModel)

	gpt41Model, err := m.loadModelFile(embedded.GPT41ModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load GPT-4.1 model: %w", err)
	}
	allModels = append(allModels, gpt41Model)

	o3ProModel, err := m.loadModelFile(embedded.O3ProModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load o3-pro model: %w", err)
	}
	allModels = append(allModels, o3ProModel)

	o1Model, err := m.loadModelFile(embedded.O1ModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load o1 model: %w", err)
	}
	allModels = append(allModels, o1Model)

	gpt4oModel, err := m.loadModelFile(embedded.GPT4oModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load GPT-4o model: %w", err)
	}
	allModels = append(allModels, gpt4oModel)

	o1ProModel, err := m.loadModelFile(embedded.O1ProModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load o1-pro model: %w", err)
	}
	allModels = append(allModels, o1ProModel)

	gpt5ChatModel, err := m.loadModelFile(embedded.GPT5ChatModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load GPT-5 Chat model: %w", err)
	}
	allModels = append(allModels, gpt5ChatModel)

	gpt5ResponsesModel, err := m.loadModelFile(embedded.GPT5ResponsesModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load GPT-5 Responses model: %w", err)
	}
	allModels = append(allModels, gpt5ResponsesModel)

	// Load Gemini models
	gemini25ProModel, err := m.loadModelFile(embedded.Gemini25ProModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gemini 2.5 Pro model: %w", err)
	}
	allModels = append(allModels, gemini25ProModel)

	gemini25FlashModel, err := m.loadModelFile(embedded.Gemini25FlashModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gemini 2.5 Flash model: %w", err)
	}
	allModels = append(allModels, gemini25FlashModel)

	gemini25FlashLiteModel, err := m.loadModelFile(embedded.Gemini25FlashLiteModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gemini 2.5 Flash Lite model: %w", err)
	}
	allModels = append(allModels, gemini25FlashLiteModel)

	// Validate that all model IDs are unique (case-insensitive)
	if err := m.validateUniqueIDs(allModels); err != nil {
		return nil, fmt.Errorf("model catalog validation failed: %w", err)
	}

	return allModels, nil
}

// GetModelCatalogByProvider returns models from a specific provider.
func (m *ModelCatalogService) GetModelCatalogByProvider(provider string) ([]neurotypes.ModelCatalogEntry, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model catalog service not initialized")
	}

	// Get all models first
	allModels, err := m.GetModelCatalog()
	if err != nil {
		return nil, err
	}

	// Filter by provider
	var filteredModels []neurotypes.ModelCatalogEntry
	providerLower := strings.ToLower(provider)
	for _, model := range allModels {
		if strings.ToLower(model.Provider) == providerLower {
			filteredModels = append(filteredModels, model)
		}
	}

	if len(filteredModels) == 0 {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return filteredModels, nil
}

// SearchModelCatalog searches for models matching the query string across all providers.
func (m *ModelCatalogService) SearchModelCatalog(query string) ([]neurotypes.ModelCatalogEntry, error) {
	allModels, err := m.GetModelCatalog()
	if err != nil {
		return nil, err
	}

	var matches []neurotypes.ModelCatalogEntry
	queryLower := strings.ToLower(query)

	for _, model := range allModels {
		// Search in model name, display name, and description
		if strings.Contains(strings.ToLower(model.Name), queryLower) ||
			strings.Contains(strings.ToLower(model.DisplayName), queryLower) ||
			strings.Contains(strings.ToLower(model.Description), queryLower) {
			matches = append(matches, model)
		}
	}

	return matches, nil
}

// GetSupportedProviders returns a list of supported providers.
func (m *ModelCatalogService) GetSupportedProviders() []string {
	return []string{"anthropic", "openai"}
}

// GetModelByID returns a model by its ID (case-insensitive lookup).
func (m *ModelCatalogService) GetModelByID(id string) (neurotypes.ModelCatalogEntry, error) {
	if !m.initialized {
		return neurotypes.ModelCatalogEntry{}, fmt.Errorf("model catalog service not initialized")
	}

	allModels, err := m.GetModelCatalog()
	if err != nil {
		return neurotypes.ModelCatalogEntry{}, err
	}

	normalizedID := m.normalizeID(id)
	for _, model := range allModels {
		if m.normalizeID(model.ID) == normalizedID {
			return model, nil
		}
	}

	return neurotypes.ModelCatalogEntry{}, fmt.Errorf("model with ID '%s' not found in catalog", id)
}

// GetProviderCatalogIDsByModelID returns the provider catalog IDs for a given model catalog ID.
// Note: Returns a slice for backward compatibility, but now each model has exactly one catalog ID.
func (m *ModelCatalogService) GetProviderCatalogIDsByModelID(id string) ([]string, error) {
	model, err := m.GetModelByID(id)
	if err != nil {
		return nil, err
	}

	if model.ProviderCatalogID == "" {
		return nil, fmt.Errorf("model with ID '%s' has no provider catalog ID defined", id)
	}

	return []string{model.ProviderCatalogID}, nil
}

// GetProviderCatalogIDByModelID returns the single provider catalog ID for a given model catalog ID.
func (m *ModelCatalogService) GetProviderCatalogIDByModelID(id string) (string, error) {
	model, err := m.GetModelByID(id)
	if err != nil {
		return "", err
	}

	if model.ProviderCatalogID == "" {
		return "", fmt.Errorf("model with ID '%s' has no provider catalog ID defined", id)
	}

	return model.ProviderCatalogID, nil
}

// validateUniqueIDs checks for duplicate model IDs (case-insensitive).
func (m *ModelCatalogService) validateUniqueIDs(models []neurotypes.ModelCatalogEntry) error {
	seenIDs := make(map[string]string) // normalized_id -> original_id

	for _, model := range models {
		if model.ID == "" {
			return fmt.Errorf("model '%s' has empty ID field", model.Name)
		}

		normalizedID := m.normalizeID(model.ID)
		if existingID, exists := seenIDs[normalizedID]; exists {
			return fmt.Errorf("duplicate model ID found: '%s' and '%s' (case insensitive)", existingID, model.ID)
		}
		seenIDs[normalizedID] = model.ID
	}

	return nil
}

// normalizeID converts an ID to uppercase for case-insensitive comparison.
func (m *ModelCatalogService) normalizeID(id string) string {
	return strings.ToUpper(id)
}

// loadModelFile loads and parses an individual model file from embedded YAML data.
func (m *ModelCatalogService) loadModelFile(data []byte) (neurotypes.ModelCatalogEntry, error) {
	var modelFile neurotypes.ModelCatalogFile

	if err := yaml.Unmarshal(data, &modelFile); err != nil {
		return neurotypes.ModelCatalogEntry{}, fmt.Errorf("failed to parse model file: %w", err)
	}

	return modelFile.ModelCatalogEntry, nil
}
