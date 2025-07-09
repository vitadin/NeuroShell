// Package neurotypes defines LLM-related types and interfaces for NeuroShell.
// This file contains types for LLM client abstraction, streaming, and service interfaces.
package neurotypes

// StreamChunk represents a single chunk of streaming response.
type StreamChunk struct {
	Content string // The text content of this chunk
	Done    bool   // Whether this is the final chunk
	Error   error  // Any error that occurred during streaming
}

// LLMClient defines the interface for LLM provider implementations.
// This interface abstracts different LLM providers (OpenAI, Anthropic, etc.)
// and provides a common way to interact with them.
type LLMClient interface {
	// SendChatCompletion sends a chat completion request and returns the full response.
	SendChatCompletion(session *ChatSession, model *ModelConfig) (string, error)

	// StreamChatCompletion sends a streaming chat completion request.
	// It returns a channel that receives response chunks as they arrive.
	StreamChatCompletion(session *ChatSession, model *ModelConfig) (<-chan StreamChunk, error)

	// GetProviderName returns the name of the LLM provider (e.g., "openai", "anthropic").
	GetProviderName() string

	// IsConfigured returns true if the client has valid configuration and can make requests.
	IsConfigured() bool
}

// ClientFactory manages the creation and caching of LLM clients.
// It provides a centralized way to get clients based on API keys and supports
// lazy initialization to avoid creating clients until they're actually needed.
type ClientFactory interface {
	Service

	// GetClient returns an LLM client for the given API key.
	// If a client for this API key already exists, it returns the cached client.
	// If not, it creates a new client and caches it for future use.
	GetClient(apiKey string) (LLMClient, error)

	// GetClientForProvider returns an LLM client for the specified provider and API key.
	// This allows for explicit provider selection when multiple providers are supported.
	GetClientForProvider(provider, apiKey string) (LLMClient, error)
}

// LLMService defines the refined interface for LLM operations.
// This interface focuses on pure business logic without external service dependencies.
// All required data is passed as explicit parameters rather than being fetched internally.
type LLMService interface {
	Service

	// SendCompletion sends a chat completion request using the provided client.
	// All dependencies are explicitly provided as parameters.
	SendCompletion(client LLMClient, session *ChatSession, model *ModelConfig, message string) (string, error)

	// StreamCompletion sends a streaming chat completion request using the provided client.
	// All dependencies are explicitly provided as parameters.
	StreamCompletion(client LLMClient, session *ChatSession, model *ModelConfig, message string) (<-chan StreamChunk, error)
}
