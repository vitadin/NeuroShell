// Package services provides the CatalogService for managing and searching the LLM model catalog.
// This service handles model discovery, filtering, and integration with the model management system.
package services

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"neuroshell/pkg/neurotypes"
)

//go:embed model_catalog.json
var catalogData []byte

// CatalogService provides model catalog management and search capabilities.
// It loads the embedded catalog data and provides search, filter, and discovery functionality.
type CatalogService struct {
	initialized bool
	catalog     *neurotypes.ModelCatalog
	allModels   []neurotypes.CatalogModel // Flattened list for efficient searching
}

// NewCatalogService creates a new catalog service instance.
func NewCatalogService() *CatalogService {
	return &CatalogService{
		initialized: false,
	}
}

// Name returns the service name for registration and identification.
func (c *CatalogService) Name() string {
	return "catalog"
}

// Initialize loads and parses the embedded catalog data.
func (c *CatalogService) Initialize(_ neurotypes.Context) error {
	if c.initialized {
		return nil // Already initialized
	}

	// Parse embedded catalog JSON
	catalog := &neurotypes.ModelCatalog{}
	if err := json.Unmarshal(catalogData, catalog); err != nil {
		return fmt.Errorf("failed to parse model catalog: %w", err)
	}

	// Flatten models for efficient searching and set provider references
	var allModels []neurotypes.CatalogModel
	for providerID, provider := range catalog.Providers {
		for _, model := range provider.Models {
			// Set provider reference for convenience
			model.Provider = providerID
			allModels = append(allModels, model)
		}
	}

	c.catalog = catalog
	c.allModels = allModels
	c.initialized = true

	return nil
}

// SearchModels performs a search across the model catalog with the given options.
// Returns a structured result with matched models and metadata.
func (c *CatalogService) SearchModels(options neurotypes.CatalogSearchOptions) (*neurotypes.CatalogSearchResult, error) {
	if !c.initialized {
		return nil, fmt.Errorf("catalog service not initialized")
	}

	startTime := time.Now()
	var matchedModels []neurotypes.CatalogModel

	// Apply filters
	for _, model := range c.allModels {
		if c.matchesFilter(model, options) {
			matchedModels = append(matchedModels, model)
		}
	}

	// Sort results
	c.sortModels(matchedModels, options.Sort)

	// Prepare result metadata
	result := &neurotypes.CatalogSearchResult{
		Models:    matchedModels,
		Count:     len(matchedModels),
		QueryTime: time.Since(startTime),
	}

	// Set provider if all results are from the same provider
	if len(matchedModels) > 0 {
		provider := matchedModels[0].Provider
		allSameProvider := true
		for _, model := range matchedModels {
			if model.Provider != provider {
				allSameProvider = false
				break
			}
		}
		if allSameProvider {
			result.Provider = provider
		}
	}

	// Set model ID if exactly one result
	if len(matchedModels) == 1 {
		result.ModelID = matchedModels[0].ID
	}

	return result, nil
}

// GetModel retrieves a specific model by provider and ID.
func (c *CatalogService) GetModel(provider, modelID string) (*neurotypes.CatalogModel, error) {
	if !c.initialized {
		return nil, fmt.Errorf("catalog service not initialized")
	}

	for _, model := range c.allModels {
		if model.Provider == provider && model.ID == modelID {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model '%s' not found for provider '%s'", modelID, provider)
}

// GetProviders returns a list of all available providers in the catalog.
func (c *CatalogService) GetProviders() ([]string, error) {
	if !c.initialized {
		return nil, fmt.Errorf("catalog service not initialized")
	}

	providers := make([]string, 0, len(c.catalog.Providers))
	for providerID := range c.catalog.Providers {
		providers = append(providers, providerID)
	}

	sort.Strings(providers)
	return providers, nil
}

// IsValidModel checks if a model exists in the catalog for the given provider.
func (c *CatalogService) IsValidModel(provider, modelID string) bool {
	if !c.initialized {
		return false
	}

	_, err := c.GetModel(provider, modelID)
	return err == nil
}

// SuggestModels provides model suggestions based on fuzzy matching for the given provider and pattern.
// Useful for error recovery when users make typos.
func (c *CatalogService) SuggestModels(provider, pattern string, maxSuggestions int) []string {
	if !c.initialized {
		return nil
	}

	var suggestions []string
	pattern = strings.ToLower(pattern)

	// Score models by similarity to the pattern
	type suggestion struct {
		model string
		score int
	}
	var scored []suggestion

	for _, model := range c.allModels {
		if provider != "" && model.Provider != provider {
			continue
		}

		score := c.calculateSimilarityScore(strings.ToLower(model.ID), pattern)
		if score > 0 {
			modelName := model.ID
			if provider == "" {
				modelName = fmt.Sprintf("%s/%s", model.Provider, model.ID)
			}
			scored = append(scored, suggestion{model: modelName, score: score})
		}
	}

	// Sort by score (higher is better)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top suggestions
	limit := maxSuggestions
	if limit <= 0 || limit > len(scored) {
		limit = len(scored)
	}

	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, scored[i].model)
	}

	return suggestions
}

// GetCatalogInfo returns basic information about the loaded catalog.
func (c *CatalogService) GetCatalogInfo() (*neurotypes.ModelCatalog, error) {
	if !c.initialized {
		return nil, fmt.Errorf("catalog service not initialized")
	}

	return c.catalog, nil
}

// matchesFilter checks if a model matches the given search options.
func (c *CatalogService) matchesFilter(model neurotypes.CatalogModel, options neurotypes.CatalogSearchOptions) bool {
	// Provider filter
	if options.Provider != "" && model.Provider != options.Provider {
		return false
	}

	// Pattern filter (fuzzy search across ID, name, and description)
	if options.Pattern != "" {
		pattern := strings.ToLower(options.Pattern)
		searchText := strings.ToLower(model.ID + " " + model.Name + " " + model.Description)
		if !strings.Contains(searchText, pattern) {
			// Try fuzzy matching on ID
			if c.calculateSimilarityScore(strings.ToLower(model.ID), pattern) == 0 {
				return false
			}
		}
	}

	return true
}

// sortModels sorts a slice of models according to the specified sort field.
func (c *CatalogService) sortModels(models []neurotypes.CatalogModel, sortField string) {
	switch sortField {
	case "name":
		sort.Slice(models, func(i, j int) bool {
			return models[i].Name < models[j].Name
		})
	case "context_length":
		sort.Slice(models, func(i, j int) bool {
			return models[i].ContextLength > models[j].ContextLength // Descending
		})
	case "pricing_tier":
		sort.Slice(models, func(i, j int) bool {
			return c.getPricingTierOrder(models[i].PricingTier) < c.getPricingTierOrder(models[j].PricingTier)
		})
	case "provider":
		sort.Slice(models, func(i, j int) bool {
			if models[i].Provider == models[j].Provider {
				return models[i].ID < models[j].ID
			}
			return models[i].Provider < models[j].Provider
		})
	case "release_date":
		sort.Slice(models, func(i, j int) bool {
			return models[i].ReleaseDate > models[j].ReleaseDate // Newest first
		})
	default:
		// Default sort by provider, then by ID
		sort.Slice(models, func(i, j int) bool {
			if models[i].Provider == models[j].Provider {
				return models[i].ID < models[j].ID
			}
			return models[i].Provider < models[j].Provider
		})
	}
}

// getPricingTierOrder returns a numeric order for pricing tiers.
func (c *CatalogService) getPricingTierOrder(tier string) int {
	switch tier {
	case "free":
		return 0
	case "economy":
		return 1
	case "standard":
		return 2
	case "premium":
		return 3
	case "enterprise":
		return 4
	default:
		return 5
	}
}

// calculateSimilarityScore computes a simple similarity score between two strings.
// Higher scores indicate greater similarity. Returns 0 for no similarity.
func (c *CatalogService) calculateSimilarityScore(text, pattern string) int {
	if pattern == "" {
		return 0 // Empty pattern matches nothing
	}

	if text == pattern {
		return 100 // Exact match
	}

	if strings.HasPrefix(text, pattern) {
		return 90 // Starts with pattern
	}

	if strings.HasSuffix(text, pattern) {
		return 70 // Ends with pattern
	}

	if strings.Contains(text, pattern) {
		return 80 // Contains pattern
	}

	// Simple character-based similarity
	common := 0
	for _, char := range pattern {
		if strings.ContainsRune(text, char) {
			common++
		}
	}

	if len(pattern) > 0 {
		score := (common * 50) / len(pattern)
		if score >= 25 { // Only return meaningful scores
			return score
		}
	}

	return 0 // No significant similarity
}

// AutoCreateModelFromCatalog creates a model configuration from a catalog entry.
// Used when catalog search returns exactly one result for auto-creation.
func (c *CatalogService) AutoCreateModelFromCatalog(catalogModel neurotypes.CatalogModel, modelName string) (*neurotypes.ModelConfig, error) {
	if !c.initialized {
		return nil, fmt.Errorf("catalog service not initialized")
	}

	// Generate unique model ID
	modelID := generateUUID()

	// Create model configuration with catalog defaults
	modelConfig := &neurotypes.ModelConfig{
		ID:          modelID,
		Name:        modelName,
		Provider:    catalogModel.Provider,
		BaseModel:   catalogModel.ID,
		Parameters:  catalogModel.GetDefaultParameters(),
		Description: fmt.Sprintf("Auto-created from catalog: %s", catalogModel.Description),
		CreatedAt:   time.Now(),
	}

	return modelConfig, nil
}

// GenerateAutoModelName generates a suitable model name for auto-creation.
func (c *CatalogService) GenerateAutoModelName(pattern string, catalogModel neurotypes.CatalogModel) string {
	// Clean pattern for use as model name
	cleanPattern := strings.TrimSpace(pattern)
	cleanPattern = strings.ReplaceAll(cleanPattern, " ", "-")

	// Use pattern if it's reasonable, otherwise generate descriptive name
	if len(cleanPattern) > 0 && len(cleanPattern) <= 50 && !strings.ContainsAny(cleanPattern, "\n\t\r") {
		return cleanPattern + "-auto"
	}

	// Generate descriptive name
	return fmt.Sprintf("%s-%s-auto", catalogModel.Provider, catalogModel.ID)
}

// generateUUID generates a simple UUID for model IDs.
// This is a placeholder implementation - in production, use a proper UUID library.
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
