// Package neurotypes provides type definitions for the NeuroShell model catalog system.
// This package defines data structures for managing and discovering LLM models from various providers.
package neurotypes

import (
	"fmt"
	"time"
)

// ModelCatalog represents the complete catalog of available LLM models from all providers.
// It serves as the root structure for the embedded JSON catalog data.
type ModelCatalog struct {
	Version     string                     `json:"version"`      // Catalog schema version
	LastUpdated string                     `json:"last_updated"` // ISO date of last catalog update
	Providers   map[string]CatalogProvider `json:"providers"`    // Map of provider ID to provider info
}

// CatalogProvider represents a single LLM provider (e.g., OpenAI, Anthropic) in the catalog.
// Contains metadata about the provider and all their available models.
type CatalogProvider struct {
	Name        string         `json:"name"`        // Human-readable provider name
	Description string         `json:"description"` // Brief description of the provider
	Website     string         `json:"website"`     // Provider's website URL
	Models      []CatalogModel `json:"models"`      // List of available models from this provider
}

// CatalogModel represents a single LLM model available from a provider.
// Contains comprehensive metadata about the model's capabilities and parameters.
type CatalogModel struct {
	ID              string                 `json:"id"`                // Unique model identifier (e.g., "gpt-4", "claude-3-sonnet")
	Name            string                 `json:"name"`              // Human-readable model name
	Description     string                 `json:"description"`       // Detailed description of model capabilities
	ContextLength   int                    `json:"context_length"`    // Maximum context window in tokens
	MaxOutputTokens int                    `json:"max_output_tokens"` // Maximum output tokens per request
	PricingTier     string                 `json:"pricing_tier"`      // Pricing category (e.g., "free", "premium", "enterprise")
	Capabilities    []string               `json:"capabilities"`      // List of capabilities (e.g., "text", "function_calling", "json_mode")
	ReleaseDate     string                 `json:"release_date"`      // ISO date of model release
	Status          string                 `json:"status"`            // Model status (e.g., "stable", "beta", "deprecated")
	Parameters      CatalogModelParameters `json:"parameters"`        // Supported parameter ranges and defaults
	Provider        string                 `json:"-"`                 // Provider ID (set during catalog loading)
}

// CatalogModelParameters defines the supported parameter ranges and defaults for a model.
// Used for validation and providing sensible defaults when creating model configurations.
type CatalogModelParameters struct {
	Temperature      *ParameterRange `json:"temperature,omitempty"`       // Temperature parameter constraints
	MaxTokens        *ParameterRange `json:"max_tokens,omitempty"`        // Max tokens parameter constraints
	TopP             *ParameterRange `json:"top_p,omitempty"`             // Top-p parameter constraints
	TopK             *ParameterRange `json:"top_k,omitempty"`             // Top-k parameter constraints
	PresencePenalty  *ParameterRange `json:"presence_penalty,omitempty"`  // Presence penalty constraints
	FrequencyPenalty *ParameterRange `json:"frequency_penalty,omitempty"` // Frequency penalty constraints
}

// ParameterRange defines the valid range and default value for a model parameter.
// Supports both integer and floating-point parameters with min/max validation.
type ParameterRange struct {
	Min     *float64 `json:"min,omitempty"`     // Minimum allowed value
	Max     *float64 `json:"max,omitempty"`     // Maximum allowed value
	Default *float64 `json:"default,omitempty"` // Default value if not specified
	Type    string   `json:"type"`              // Parameter type ("int", "float", "bool", "string")
}

// CatalogSearchOptions defines search and filter criteria for model catalog queries.
// Used by the catalog service to filter and sort model results.
type CatalogSearchOptions struct {
	Provider string // Filter by provider (e.g., "openai", "anthropic")
	Pattern  string // Search pattern for model names/descriptions (supports fuzzy matching)
	Sort     string // Sort field ("name", "context_length", "pricing_tier", "provider", "release_date")
}

// CatalogSearchResult represents the result of a catalog search operation.
// Contains matched models and metadata for use in commands and variable setting.
type CatalogSearchResult struct {
	Models    []CatalogModel // List of models matching the search criteria
	Count     int            // Number of matching models
	Provider  string         // Single provider if all results from same provider (empty if mixed)
	ModelID   string         // Single model ID if exactly one result (empty if multiple)
	QueryTime time.Duration  // Time taken to execute the search query
}

// AutoCreatedModelInfo contains metadata about models automatically created from catalog discoveries.
// Used to track the relationship between catalog entries and created model configurations.
type AutoCreatedModelInfo struct {
	CatalogModelID  string    // Original catalog model ID
	CatalogProvider string    // Original catalog provider
	CreatedModelID  string    // ID of the created ModelConfig
	CreatedAt       time.Time // When the auto-creation occurred
	SearchPattern   string    // Original search pattern that led to auto-creation
}

// GetDefaultParameters returns a map of default parameter values for this catalog model.
// Used when auto-creating models from catalog entries.
func (m *CatalogModel) GetDefaultParameters() map[string]any {
	params := make(map[string]any)

	if m.Parameters.Temperature != nil && m.Parameters.Temperature.Default != nil {
		params["temperature"] = *m.Parameters.Temperature.Default
	}
	if m.Parameters.MaxTokens != nil && m.Parameters.MaxTokens.Default != nil {
		params["max_tokens"] = int(*m.Parameters.MaxTokens.Default)
	}
	if m.Parameters.TopP != nil && m.Parameters.TopP.Default != nil {
		params["top_p"] = *m.Parameters.TopP.Default
	}
	if m.Parameters.TopK != nil && m.Parameters.TopK.Default != nil {
		params["top_k"] = int(*m.Parameters.TopK.Default)
	}
	if m.Parameters.PresencePenalty != nil && m.Parameters.PresencePenalty.Default != nil {
		params["presence_penalty"] = *m.Parameters.PresencePenalty.Default
	}
	if m.Parameters.FrequencyPenalty != nil && m.Parameters.FrequencyPenalty.Default != nil {
		params["frequency_penalty"] = *m.Parameters.FrequencyPenalty.Default
	}

	return params
}

// ValidateParameter checks if a parameter value is within the allowed range for this model.
// Returns an error if the value is outside the valid range.
func (m *CatalogModel) ValidateParameter(name string, value any) error {
	switch name {
	case "temperature":
		return m.validateFloatParameter(m.Parameters.Temperature, value, "temperature")
	case "max_tokens":
		return m.validateIntParameter(m.Parameters.MaxTokens, value, "max_tokens")
	case "top_p":
		return m.validateFloatParameter(m.Parameters.TopP, value, "top_p")
	case "top_k":
		return m.validateIntParameter(m.Parameters.TopK, value, "top_k")
	case "presence_penalty":
		return m.validateFloatParameter(m.Parameters.PresencePenalty, value, "presence_penalty")
	case "frequency_penalty":
		return m.validateFloatParameter(m.Parameters.FrequencyPenalty, value, "frequency_penalty")
	}
	return nil // Unknown parameters are allowed (provider-specific)
}

// validateFloatParameter validates a floating-point parameter against its range constraints.
func (m *CatalogModel) validateFloatParameter(paramRange *ParameterRange, value any, name string) error {
	if paramRange == nil {
		return nil // No constraints defined
	}

	var floatVal float64
	switch v := value.(type) {
	case float64:
		floatVal = v
	case float32:
		floatVal = float64(v)
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		return nil // Skip validation for non-numeric types
	}

	if paramRange.Min != nil && floatVal < *paramRange.Min {
		return fmt.Errorf("%s must be at least %g", name, *paramRange.Min)
	}
	if paramRange.Max != nil && floatVal > *paramRange.Max {
		return fmt.Errorf("%s must be at most %g", name, *paramRange.Max)
	}

	return nil
}

// validateIntParameter validates an integer parameter against its range constraints.
func (m *CatalogModel) validateIntParameter(paramRange *ParameterRange, value any, name string) error {
	if paramRange == nil {
		return nil // No constraints defined
	}

	var intVal int
	switch v := value.(type) {
	case int:
		intVal = v
	case int64:
		intVal = int(v)
	case float64:
		intVal = int(v)
	case float32:
		intVal = int(v)
	default:
		return nil // Skip validation for non-numeric types
	}

	if paramRange.Min != nil && float64(intVal) < *paramRange.Min {
		return fmt.Errorf("%s must be at least %g", name, *paramRange.Min)
	}
	if paramRange.Max != nil && float64(intVal) > *paramRange.Max {
		return fmt.Errorf("%s must be at most %g", name, *paramRange.Max)
	}

	return nil
}
