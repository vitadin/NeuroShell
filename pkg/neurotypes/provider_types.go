// Package neurotypes defines provider-related data structures for NeuroShell's LLM provider management.
// This file contains the core types for representing and managing LLM provider configurations.
package neurotypes

// ProviderCatalogEntry represents a provider endpoint configuration.
// It combines provider information, API configuration, and client implementation details
// into a named configuration that can be used for LLM client creation.
type ProviderCatalogEntry struct {
	// ID is a unique identifier for the provider configuration (e.g., "openai_chat", "anthropic_chat")
	// Must be unique across all provider entries (case-insensitive)
	ID string `yaml:"id" json:"id"`

	// Provider is the provider name (e.g., "openai", "anthropic", "moonshot")
	Provider string `yaml:"provider" json:"provider"`

	// DisplayName is a human-readable name for the provider endpoint
	// (e.g., "OpenAI Chat Completions", "Anthropic Claude Chat")
	DisplayName string `yaml:"display_name" json:"display_name"`

	// BaseURL is the API base URL for this provider endpoint
	// (e.g., "https://api.openai.com/v1", "https://api.anthropic.com/v1")
	BaseURL string `yaml:"base_url" json:"base_url"`

	// Endpoint is the specific endpoint path for this provider service
	// (e.g., "/chat/completions", "/messages", "/embeddings")
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// Headers contains provider-specific HTTP headers required for requests
	// Common examples: authentication headers, API version headers, custom headers
	Headers map[string]string `yaml:"headers" json:"headers"`

	// ClientType indicates which client implementation to use for this provider
	// Supported types: "openai", "openai-compatible", "anthropic"
	ClientType string `yaml:"client_type" json:"client_type"`

	// Description provides a brief description of the provider endpoint's capabilities
	Description string `yaml:"description" json:"description"`

	// ImplementationNotes provides information about how the provider connection is handled
	// (e.g., "Natively supported by NeuroShell", "Uses OpenAI-compatible API")
	ImplementationNotes string `yaml:"implementation_notes" json:"implementation_notes"`
}

// ProviderCatalogFile represents the structure of a provider catalog YAML file.
// This wraps ProviderCatalogEntry for YAML unmarshaling, similar to ModelCatalogFile.
type ProviderCatalogFile struct {
	ProviderCatalogEntry `yaml:",inline"`
}
