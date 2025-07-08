package services

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"neuroshell/internal/data/embedded"
	"neuroshell/pkg/neurotypes"
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

	o4MiniModel, err := m.loadModelFile(embedded.O4MiniModelData)
	if err != nil {
		return nil, fmt.Errorf("failed to load O4-mini model: %w", err)
	}
	allModels = append(allModels, o4MiniModel)

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
