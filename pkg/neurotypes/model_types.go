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

	// ID is a short, meaningful identifier for easy reference (e.g., "O3", "CS4", "CO37")
	// Used for quick model selection in commands like \model-new[catalog_id=CS4]
	// Must be unique across all models (case-insensitive)
	ID string `yaml:"id" json:"id"`

	// Provider is the LLM provider name (e.g., "openai", "anthropic")
	Provider string `yaml:"provider" json:"provider"`

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

	// Deprecated: indicates if this model is deprecated and should not be used for new projects
	Deprecated bool `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`

	// MaxOutputTokens is the maximum number of tokens the model can generate (if different from context window)
	MaxOutputTokens *int `yaml:"max_output_tokens,omitempty" json:"max_output_tokens,omitempty"`

	// KnowledgeCutoff is the training data cutoff date for the model
	KnowledgeCutoff *string `yaml:"knowledge_cutoff,omitempty" json:"knowledge_cutoff,omitempty"`

	// ReasoningTokens indicates whether the model supports reasoning tokens
	ReasoningTokens *bool `yaml:"reasoning_tokens,omitempty" json:"reasoning_tokens,omitempty"`

	// Modalities lists the supported input/output types (e.g., "text-input", "image-input", "audio-output")
	Modalities []string `yaml:"modalities,omitempty" json:"modalities,omitempty"`

	// Pricing contains cost information for the model
	Pricing *ModelPricing `yaml:"pricing,omitempty" json:"pricing,omitempty"`

	// Features contains feature support information
	Features *ModelFeatures `yaml:"features,omitempty" json:"features,omitempty"`

	// Tools lists the tools/capabilities supported by the model
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`

	// Snapshots lists available model snapshots or aliases
	Snapshots []string `yaml:"snapshots,omitempty" json:"snapshots,omitempty"`
}

// ModelPricing contains pricing information for a model.
type ModelPricing struct {
	// InputPerMToken is the cost per 1 million input tokens in USD
	InputPerMToken float64 `yaml:"input_per_m_token,omitempty" json:"input_per_m_token,omitempty"`

	// OutputPerMToken is the cost per 1 million output tokens in USD
	OutputPerMToken float64 `yaml:"output_per_m_token,omitempty" json:"output_per_m_token,omitempty"`
}

// ModelFeatures contains feature support information for a model.
type ModelFeatures struct {
	// Streaming indicates whether the model supports streaming responses
	Streaming *bool `yaml:"streaming,omitempty" json:"streaming,omitempty"`

	// FunctionCalling indicates whether the model supports function/tool calling
	FunctionCalling *bool `yaml:"function_calling,omitempty" json:"function_calling,omitempty"`

	// StructuredOutputs indicates whether the model supports structured output formats
	StructuredOutputs *bool `yaml:"structured_outputs,omitempty" json:"structured_outputs,omitempty"`

	// FineTuning indicates whether the model supports fine-tuning
	FineTuning *bool `yaml:"fine_tuning,omitempty" json:"fine_tuning,omitempty"`

	// Vision indicates whether the model supports vision/image processing
	Vision *bool `yaml:"vision,omitempty" json:"vision,omitempty"`

	// Distillation indicates whether the model supports distillation
	Distillation *bool `yaml:"distillation,omitempty" json:"distillation,omitempty"`

	// PredictedOutputs indicates whether the model supports predicted outputs
	PredictedOutputs *bool `yaml:"predicted_outputs,omitempty" json:"predicted_outputs,omitempty"`
}

// ModelCatalogFile represents an individual model file loaded from YAML.
// Each model file contains a complete ModelCatalogEntry with provider information.
type ModelCatalogFile struct {
	ModelCatalogEntry `yaml:",inline" json:",inline"`
}

// ModelCatalogProvider represents a provider's model catalog loaded from YAML.
// Deprecated: Use individual model files (ModelCatalogFile) instead of provider-based files.
type ModelCatalogProvider struct {
	// Provider is the provider name (e.g., "openai", "anthropic")
	Provider string `yaml:"provider" json:"provider"`

	// Models is the list of models available from this provider
	Models []ModelCatalogEntry `yaml:"models" json:"models"`
}
