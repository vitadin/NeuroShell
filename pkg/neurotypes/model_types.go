// Package neurotypes defines model-related data structures for NeuroShell's LLM model management.
// This file contains the core types for representing and managing LLM model configurations.
package neurotypes

import "time"

// ModelConfig represents a configured LLM model with specific parameters.
// It combines an API provider, base model, and user-specified parameters into a named configuration
// that can be used across sessions and scripts for reproducible LLM interactions.
type ModelConfig struct {
	// ID is a unique identifier for the model configuration (auto-generated UUID)
	ID string `json:"id"`

	// Name is the user-friendly name for the model (unique, no spaces, user-provided)
	Name string `json:"name"`

	// Provider is the API provider name (e.g., "openai", "anthropic", "local")
	Provider string `json:"provider"`

	// BaseModel is the provider's model identifier (e.g., "gpt-4", "claude-3-sonnet", "llama-2")
	BaseModel string `json:"base_model"`

	// Parameters contains model-specific configuration parameters
	// Common parameters: temperature, max_tokens, top_p, top_k, presence_penalty, frequency_penalty
	// Provider-specific parameters can also be included
	Parameters map[string]any `json:"parameters"`

	// Description is an optional human-readable description of the model configuration
	Description string `json:"description"`

	// IsDefault indicates whether this model is the default for new sessions
	// Only one model should have IsDefault=true at any time
	IsDefault bool `json:"is_default"`

	// CreatedAt is the timestamp when the model configuration was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the model configuration was last modified
	UpdatedAt time.Time `json:"updated_at"`
}

// StandardModelParameters defines common parameters used across different LLM providers.
// These parameters have consistent meanings across providers, though not all providers
// support all parameters.
type StandardModelParameters struct {
	// Temperature controls randomness in the model's output (0.0-1.0)
	// 0.0 = deterministic, 1.0 = maximum randomness
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens is the maximum number of tokens to generate in the response
	MaxTokens *int `json:"max_tokens,omitempty"`

	// TopP implements nucleus sampling (0.0-1.0)
	// Only tokens with cumulative probability up to TopP are considered
	TopP *float64 `json:"top_p,omitempty"`

	// TopK limits the number of highest probability tokens to consider
	// Used by some providers (e.g., local models)
	TopK *int `json:"top_k,omitempty"`

	// PresencePenalty penalizes new tokens based on whether they appear in the text so far (-2.0 to 2.0)
	// Positive values encourage the model to talk about new topics
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty penalizes new tokens based on their frequency in the text so far (-2.0 to 2.0)
	// Positive values reduce the likelihood of repeated phrases
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
}

// ModelValidationError represents validation errors that occur during model configuration.
type ModelValidationError struct {
	Field   string `json:"field"`   // The field that failed validation
	Value   string `json:"value"`   // The invalid value
	Message string `json:"message"` // Human-readable error message
}

// Error implements the error interface for ModelValidationError.
func (e ModelValidationError) Error() string {
	return e.Message
}

// ModelProviderInfo contains metadata about a specific LLM provider.
// This is used for validation and capability discovery.
type ModelProviderInfo struct {
	// Name is the provider identifier (e.g., "openai", "anthropic")
	Name string `json:"name"`

	// SupportedModels lists the base models available from this provider
	SupportedModels []string `json:"supported_models"`

	// SupportedParameters lists the parameters supported by this provider
	SupportedParameters []string `json:"supported_parameters"`

	// RequiredParameters lists parameters that must be specified for this provider
	RequiredParameters []string `json:"required_parameters"`

	// DefaultParameters provides default values for optional parameters
	DefaultParameters map[string]any `json:"default_parameters"`

	// ParameterConstraints defines validation rules for parameters
	// Key is parameter name, value describes the constraint (e.g., "0.0-1.0", "1-4000")
	ParameterConstraints map[string]string `json:"parameter_constraints"`
}

// ModelCatalogEntry represents a model entry in the embedded model catalog.
// This contains basic information about available LLM models from various providers.
type ModelCatalogEntry struct {
	// Name is the provider's model identifier (e.g., "gpt-4", "claude-3-sonnet-20240229")
	Name string `yaml:"name" json:"name"`

	// DisplayName is a human-readable name for the model (e.g., "GPT-4", "Claude 3 Sonnet")
	DisplayName string `yaml:"display_name" json:"display_name"`

	// Description provides a brief description of the model's capabilities and use cases
	Description string `yaml:"description" json:"description"`

	// Capabilities lists the types of tasks this model is designed for
	Capabilities []string `yaml:"capabilities" json:"capabilities"`

	// ContextWindow is the maximum number of tokens the model can process
	ContextWindow int `yaml:"context_window" json:"context_window"`

	// Version indicates the model version or release date if applicable
	Version string `yaml:"version,omitempty" json:"version,omitempty"`

	// Deprecated indicates if this model is deprecated and should not be used for new projects
	Deprecated bool `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`
}

// ModelCatalogProvider represents a provider's model catalog loaded from YAML.
type ModelCatalogProvider struct {
	// Provider is the provider name (e.g., "openai", "anthropic")
	Provider string `yaml:"provider" json:"provider"`

	// Models is the list of models available from this provider
	Models []ModelCatalogEntry `yaml:"models" json:"models"`
}
