package services

import (
	"fmt"
	"strings"

	"neuroshell/internal/data/embedded"
	"neuroshell/pkg/neurotypes"

	"gopkg.in/yaml.v3"
)

// ProviderCatalogService provides provider catalog operations for NeuroShell.
// It handles loading and searching through embedded YAML provider catalogs
// from different LLM providers (OpenAI, Anthropic, Moonshot, OpenRouter).
type ProviderCatalogService struct {
	initialized bool
}

// NewProviderCatalogService creates a new ProviderCatalogService instance.
func NewProviderCatalogService() *ProviderCatalogService {
	return &ProviderCatalogService{
		initialized: false,
	}
}

// Name returns the service name "provider_catalog" for registration.
func (p *ProviderCatalogService) Name() string {
	return "provider_catalog"
}

// Initialize sets up the ProviderCatalogService for operation.
func (p *ProviderCatalogService) Initialize() error {
	p.initialized = true
	return nil
}

// GetProviderCatalog returns the complete provider catalog from all providers.
func (p *ProviderCatalogService) GetProviderCatalog() ([]neurotypes.ProviderCatalogEntry, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider catalog service not initialized")
	}

	var allProviders []neurotypes.ProviderCatalogEntry

	// Load OpenAI providers
	openaiChat, err := p.loadProviderFile(embedded.OpenAIChatProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAI chat provider: %w", err)
	}
	allProviders = append(allProviders, openaiChat)

	// Load OpenAI Responses provider
	openaiResponses, err := p.loadProviderFile(embedded.OpenAIResponsesProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAI responses provider: %w", err)
	}
	allProviders = append(allProviders, openaiResponses)

	// Load Anthropic providers
	anthropicChat, err := p.loadProviderFile(embedded.AnthropicChatProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Anthropic chat provider: %w", err)
	}
	allProviders = append(allProviders, anthropicChat)

	// Load Moonshot providers
	moonshotChat, err := p.loadProviderFile(embedded.MoonshotChatProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Moonshot chat provider: %w", err)
	}
	allProviders = append(allProviders, moonshotChat)

	// Load OpenRouter providers
	openrouterChat, err := p.loadProviderFile(embedded.OpenRouterChatProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenRouter chat provider: %w", err)
	}
	allProviders = append(allProviders, openrouterChat)

	// Load Gemini providers
	geminiChat, err := p.loadProviderFile(embedded.GeminiChatProviderData)
	if err != nil {
		return nil, fmt.Errorf("failed to load Gemini chat provider: %w", err)
	}
	allProviders = append(allProviders, geminiChat)

	// Validate that all provider IDs are unique (case-insensitive)
	if err := p.validateUniqueIDs(allProviders); err != nil {
		return nil, fmt.Errorf("provider catalog validation failed: %w", err)
	}

	return allProviders, nil
}

// GetProvidersByProvider returns providers from a specific provider name.
func (p *ProviderCatalogService) GetProvidersByProvider(provider string) ([]neurotypes.ProviderCatalogEntry, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider catalog service not initialized")
	}

	// Get all providers first
	allProviders, err := p.GetProviderCatalog()
	if err != nil {
		return nil, err
	}

	// Filter by provider
	var filteredProviders []neurotypes.ProviderCatalogEntry
	providerLower := strings.ToLower(provider)
	for _, providerEntry := range allProviders {
		if strings.ToLower(providerEntry.Provider) == providerLower {
			filteredProviders = append(filteredProviders, providerEntry)
		}
	}

	if len(filteredProviders) == 0 {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return filteredProviders, nil
}

// SearchProviderCatalog searches for providers matching the query string across all providers.
func (p *ProviderCatalogService) SearchProviderCatalog(query string) ([]neurotypes.ProviderCatalogEntry, error) {
	allProviders, err := p.GetProviderCatalog()
	if err != nil {
		return nil, err
	}

	var matches []neurotypes.ProviderCatalogEntry
	queryLower := strings.ToLower(query)

	for _, provider := range allProviders {
		// Search in provider ID, display name, provider name, and description
		if strings.Contains(strings.ToLower(provider.ID), queryLower) ||
			strings.Contains(strings.ToLower(provider.DisplayName), queryLower) ||
			strings.Contains(strings.ToLower(provider.Provider), queryLower) ||
			strings.Contains(strings.ToLower(provider.Description), queryLower) {
			matches = append(matches, provider)
		}
	}

	return matches, nil
}

// GetSupportedProviders returns a list of supported provider names.
func (p *ProviderCatalogService) GetSupportedProviders() []string {
	return []string{"openai", "anthropic", "moonshot", "openrouter", "gemini"}
}

// GetValidCatalogIDs returns a list of all valid provider catalog IDs dynamically.
func (p *ProviderCatalogService) GetValidCatalogIDs() ([]string, error) {
	if !p.initialized {
		return nil, fmt.Errorf("provider catalog service not initialized")
	}

	allProviders, err := p.GetProviderCatalog()
	if err != nil {
		return nil, err
	}

	var catalogIDs []string
	for _, provider := range allProviders {
		catalogIDs = append(catalogIDs, provider.ID)
	}

	return catalogIDs, nil
}

// GetProviderByID returns a provider by its ID (case-insensitive lookup).
func (p *ProviderCatalogService) GetProviderByID(id string) (neurotypes.ProviderCatalogEntry, error) {
	if !p.initialized {
		return neurotypes.ProviderCatalogEntry{}, fmt.Errorf("provider catalog service not initialized")
	}

	allProviders, err := p.GetProviderCatalog()
	if err != nil {
		return neurotypes.ProviderCatalogEntry{}, err
	}

	normalizedID := p.normalizeID(id)
	for _, provider := range allProviders {
		if p.normalizeID(provider.ID) == normalizedID {
			return provider, nil
		}
	}

	return neurotypes.ProviderCatalogEntry{}, fmt.Errorf("provider with ID '%s' not found in catalog", id)
}

// validateUniqueIDs checks for duplicate provider IDs (case-insensitive).
func (p *ProviderCatalogService) validateUniqueIDs(providers []neurotypes.ProviderCatalogEntry) error {
	seenIDs := make(map[string]string) // normalized_id -> original_id

	for _, provider := range providers {
		if provider.ID == "" {
			return fmt.Errorf("provider '%s' has empty ID field", provider.DisplayName)
		}

		normalizedID := p.normalizeID(provider.ID)
		if existingID, exists := seenIDs[normalizedID]; exists {
			return fmt.Errorf("duplicate provider ID found: '%s' and '%s' (case insensitive)", existingID, provider.ID)
		}
		seenIDs[normalizedID] = provider.ID
	}

	return nil
}

// normalizeID converts an ID to uppercase for case-insensitive comparison.
func (p *ProviderCatalogService) normalizeID(id string) string {
	return strings.ToUpper(id)
}

// loadProviderFile loads and parses an individual provider file from embedded YAML data.
func (p *ProviderCatalogService) loadProviderFile(data []byte) (neurotypes.ProviderCatalogEntry, error) {
	var providerFile neurotypes.ProviderCatalogFile

	if err := yaml.Unmarshal(data, &providerFile); err != nil {
		return neurotypes.ProviderCatalogEntry{}, fmt.Errorf("failed to parse provider file: %w", err)
	}

	return providerFile.ProviderCatalogEntry, nil
}
