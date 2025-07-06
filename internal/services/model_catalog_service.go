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
func (m *ModelCatalogService) Initialize(_ neurotypes.Context) error {
	m.initialized = true
	return nil
}

// GetModelCatalog returns the complete model catalog from all providers.
func (m *ModelCatalogService) GetModelCatalog() ([]neurotypes.ModelCatalogEntry, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model catalog service not initialized")
	}

	var allModels []neurotypes.ModelCatalogEntry

	// Load Anthropic models
	anthropicModels, err := m.loadProviderCatalog("anthropic", embedded.AnthropicCatalogData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Anthropic catalog: %w", err)
	}
	allModels = append(allModels, anthropicModels...)

	// Load OpenAI models
	openaiModels, err := m.loadProviderCatalog("openai", embedded.OpenaiCatalogData)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAI catalog: %w", err)
	}
	allModels = append(allModels, openaiModels...)

	return allModels, nil
}

// GetModelCatalogByProvider returns models from a specific provider.
func (m *ModelCatalogService) GetModelCatalogByProvider(provider string) ([]neurotypes.ModelCatalogEntry, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model catalog service not initialized")
	}

	switch strings.ToLower(provider) {
	case "anthropic":
		return m.loadProviderCatalog("anthropic", embedded.AnthropicCatalogData)
	case "openai":
		return m.loadProviderCatalog("openai", embedded.OpenaiCatalogData)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
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

// loadProviderCatalog loads and parses a provider's model catalog from embedded YAML data.
func (m *ModelCatalogService) loadProviderCatalog(providerName string, data []byte) ([]neurotypes.ModelCatalogEntry, error) {
	var catalog neurotypes.ModelCatalogProvider

	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse %s catalog: %w", providerName, err)
	}

	return catalog.Models, nil
}
