// Package neurotypes defines LLM-related types and interfaces for NeuroShell.
// This file contains types for LLM client abstraction, streaming, and service interfaces.
package neurotypes

import "net/http"

// StructuredLLMResponse represents a structured response from an LLM provider.
// It separates clean text content from thinking/reasoning blocks for proper rendering control.
type StructuredLLMResponse struct {
	TextContent    string                 // Clean main response content (user-facing)
	ThinkingBlocks []ThinkingBlock        // Extracted thinking/reasoning content
	Error          *LLMError              // Captured error if any
	Metadata       map[string]interface{} // Additional metadata from provider
}

// LLMError represents an error that occurred during LLM processing.
// This captures provider-specific error information for proper handling.
type LLMError struct {
	Code    string `json:"code"`    // Error code from provider
	Message string `json:"message"` // Human-readable error message
	Type    string `json:"type"`    // Error type (rate_limit, invalid_request, server_error, etc.)
}

// ThinkingBlock represents a block of thinking/reasoning content from an LLM.
// This unified structure handles thinking content from all providers (Anthropic, Gemini, OpenAI).
type ThinkingBlock struct {
	Content  string `json:"content"`  // The actual thinking/reasoning text content
	Provider string `json:"provider"` // Source provider: "anthropic", "gemini", "openai"
	Type     string `json:"type"`     // Block type: "thinking", "redacted_thinking", "reasoning"
}

// LLMClient defines the interface for LLM provider implementations.
// This interface abstracts different LLM providers (OpenAI, Anthropic, etc.)
// and provides a common way to interact with them.
type LLMClient interface {
	// SendChatCompletion sends a chat completion request and returns the full response.
	// This is the core method that handles the actual LLM API communication.
	// Response includes both thinking content and regular text formatted together.
	SendChatCompletion(session *ChatSession, model *ModelConfig) (string, error)

	// SendStructuredCompletion sends a chat completion request and returns structured response.
	// This separates thinking/reasoning content from regular text for proper rendering control.
	// Internally uses SendChatCompletion and processes the response to extract thinking blocks.
	// All errors are encoded in the StructuredLLMResponse.Error field - no Go errors are returned.
	SendStructuredCompletion(session *ChatSession, model *ModelConfig) *StructuredLLMResponse

	// GetProviderName returns the name of the LLM provider (e.g., "openai", "anthropic").
	GetProviderName() string

	// IsConfigured returns true if the client has valid configuration and can make requests.
	IsConfigured() bool

	// SetDebugTransport sets the HTTP transport for network debugging.
	// All clients must accept debug transport for consistent debugging infrastructure.
	SetDebugTransport(transport http.RoundTripper)
}

// ClientFactory manages the creation and caching of LLM clients.
// It provides a centralized way to get clients based on API keys and supports
// lazy initialization to avoid creating clients until they're actually needed.
type ClientFactory interface {
	Service

	// GetClientForProvider returns an LLM client for the specified provider catalog ID and API key.
	// This allows for explicit provider selection when multiple endpoints are supported.
	GetClientForProvider(providerCatalogID, apiKey string) (LLMClient, error)
}

// LLMService defines the refined interface for LLM operations.
// This interface focuses on pure business logic without external service dependencies.
// All required data is passed as explicit parameters rather than being fetched internally.
type LLMService interface {
	Service

	// SendCompletion sends a chat completion request using the provided client.
	// The session is sent as-is - message manipulation is the caller's responsibility.
	// Debug transport capture happens transparently via the client's debug transport.
	SendCompletion(client LLMClient, session *ChatSession, model *ModelConfig) (string, error)

	// SendStructuredCompletion sends a chat completion request using the provided client and returns structured response.
	// This separates thinking/reasoning content from regular text for proper rendering control.
	// Debug transport capture happens transparently via the client's debug transport.
	// All errors are encoded in the StructuredLLMResponse.Error field - no Go errors are returned.
	SendStructuredCompletion(client LLMClient, session *ChatSession, model *ModelConfig) *StructuredLLMResponse
}
