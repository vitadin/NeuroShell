package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// ClientFactoryService implements the ClientFactory interface.
// It manages the creation of LLM clients, using the context layer for stateless caching.
type ClientFactoryService struct {
	initialized bool
}

// NewClientFactoryService creates a new ClientFactoryService instance.
func NewClientFactoryService() *ClientFactoryService {
	return &ClientFactoryService{
		initialized: false,
	}
}

// Name returns the service name "client_factory" for registration.
func (f *ClientFactoryService) Name() string {
	return "client_factory"
}

// Initialize sets up the ClientFactoryService for operation.
func (f *ClientFactoryService) Initialize() error {
	logger.ServiceOperation("client_factory", "initialize", "starting")
	f.initialized = true
	logger.ServiceOperation("client_factory", "initialize", "completed")
	return nil
}

// GetClientForProvider returns an LLM client for the specified provider and API key.
// This allows for explicit provider selection when multiple providers are supported.
func (f *ClientFactoryService) GetClientForProvider(provider, apiKey string) (neurotypes.LLMClient, error) {
	if !f.initialized {
		return nil, fmt.Errorf("client factory service not initialized")
	}

	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty for provider '%s'", provider)
	}

	// Generate client ID for caching
	clientID := f.generateClientID(provider, apiKey)
	ctx := neuroshellcontext.GetGlobalContext()

	// Check if client exists in context cache
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Returning cached provider client", "provider", provider, "clientID", clientID)
		return client, nil
	}

	// Create client based on provider
	var client neurotypes.LLMClient
	switch provider {
	case "openai":
		client = NewOpenAIClient(apiKey)
	case "openrouter":
		client = f.createOpenAICompatibleClient(apiKey, "openrouter", "https://openrouter.ai/api/v1")
	case "moonshot":
		client = f.createOpenAICompatibleClient(apiKey, "moonshot", "https://api.moonshot.ai/v1")
	case "anthropic":
		client = f.createAnthropicCompatibleClient(apiKey)
	case "gemini":
		client = NewGeminiClient(apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider '%s'. Supported providers: openai, openrouter, moonshot, anthropic, gemini", provider)
	}

	// Store client in context cache
	ctx.SetLLMClient(clientID, client)

	logger.Debug("Created new provider client", "provider", provider, "clientID", clientID)
	return client, nil
}

// generateClientID creates a unique, secure client ID for the given provider and API key.
// Uses SHA-256 hash with first 8 hex characters for uniqueness while maintaining usability.
// Format: "provider:hashed-api-key" (e.g., "openai:a1b2c3d4")
func (f *ClientFactoryService) generateClientID(provider, apiKey string) string {
	if apiKey == "" {
		return fmt.Sprintf("%s:empty***", provider)
	}

	// Generate SHA-256 hash of the API key
	hash := sha256.Sum256([]byte(apiKey))

	// Convert to hex and take first 8 characters for usability
	hexHash := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:%s", provider, hexHash[:8])
}

// GetClientWithID returns both an LLM client and its unique client ID for the specified provider and API key.
// This method provides both client management and consistent ID generation for external use.
func (f *ClientFactoryService) GetClientWithID(provider, apiKey string) (neurotypes.LLMClient, string, error) {
	if !f.initialized {
		return nil, "", fmt.Errorf("client factory service not initialized")
	}

	if provider == "" {
		return nil, "", fmt.Errorf("provider cannot be empty")
	}

	if apiKey == "" {
		return nil, "", fmt.Errorf("API key cannot be empty for provider '%s'", provider)
	}

	// Generate client ID for external use (also serves as cache key)
	clientID := f.generateClientID(provider, apiKey)
	ctx := neuroshellcontext.GetGlobalContext()

	// Check if client exists in context cache
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Returning cached provider client with ID", "provider", provider, "clientID", clientID)
		return client, clientID, nil
	}

	// Create client based on provider
	var client neurotypes.LLMClient
	switch provider {
	case "openai":
		client = NewOpenAIClient(apiKey)
	case "openrouter":
		client = f.createOpenAICompatibleClient(apiKey, "openrouter", "https://openrouter.ai/api/v1")
	case "moonshot":
		client = f.createOpenAICompatibleClient(apiKey, "moonshot", "https://api.moonshot.ai/v1")
	case "anthropic":
		client = f.createAnthropicCompatibleClient(apiKey)
	case "gemini":
		client = NewGeminiClient(apiKey)
	default:
		return nil, "", fmt.Errorf("unsupported provider '%s'. Supported providers: openai, openrouter, moonshot, anthropic, gemini", provider)
	}

	// Store client in context cache
	ctx.SetLLMClient(clientID, client)

	logger.Debug("Created new provider client with ID", "provider", provider, "clientID", clientID)
	return client, clientID, nil
}

// GetClientByID retrieves a cached LLM client by its client ID.
// Client ID format: "provider:hashed-api-key" (e.g., "openai:a1b2c3d4...")
// This method enables direct O(1) lookup using the client ID as the cache key.
func (f *ClientFactoryService) GetClientByID(clientID string) (neurotypes.LLMClient, error) {
	if !f.initialized {
		return nil, fmt.Errorf("client factory service not initialized")
	}

	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	// Validate client ID format (must contain colon separator)
	if !strings.Contains(clientID, ":") {
		return nil, fmt.Errorf("invalid client ID format: %s (expected 'provider:hash')", clientID)
	}

	// Ensure both provider and hash parts are non-empty
	parts := strings.Split(clientID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid client ID format: %s (expected 'provider:hash')", clientID)
	}

	// Direct lookup using client ID from context
	ctx := neuroshellcontext.GetGlobalContext()
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Retrieved cached client by ID", "clientID", clientID)
		return client, nil
	}

	return nil, fmt.Errorf("client with ID '%s' not found in cache", clientID)
}

// DetermineAPIKeyForProvider determines the API key for a specific provider.
// It checks provider-specific environment variables through the context layer.
func (f *ClientFactoryService) DetermineAPIKeyForProvider(provider string, ctx neurotypes.Context) (string, error) {
	if provider == "" {
		return "", fmt.Errorf("provider cannot be empty")
	}

	var apiKey string
	var envVarName string

	// Check provider-specific environment variables through context
	switch provider {
	case "openai":
		envVarName = "OPENAI_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	case "openrouter":
		envVarName = "OPENROUTER_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	case "moonshot":
		envVarName = "MOONSHOT_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	case "anthropic":
		envVarName = "ANTHROPIC_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	case "gemini":
		envVarName = "GOOGLE_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	default:
		return "", fmt.Errorf("unsupported provider '%s'. Supported providers: openai, openrouter, moonshot, anthropic, gemini", provider)
	}

	if apiKey == "" {
		return "", fmt.Errorf("%s API key not found. Please set the %s environment variable",
			provider, envVarName)
	}

	logger.Debug("API key found for provider", "provider", provider, "env_var", envVarName)
	return apiKey, nil
}

// GetCachedClientCount returns the number of cached clients (for testing/debugging).
func (f *ClientFactoryService) GetCachedClientCount() int {
	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.GetLLMClientCount()
}

// ClearCache removes all cached clients (for testing/debugging).
func (f *ClientFactoryService) ClearCache() {
	ctx := neuroshellcontext.GetGlobalContext()
	ctx.ClearLLMClients()
	logger.Debug("Client cache cleared")
}

// createOpenAICompatibleClient creates a new OpenAI-compatible client with the specified provider name and base URL.
func (f *ClientFactoryService) createOpenAICompatibleClient(apiKey, providerName, baseURL string) neurotypes.LLMClient {
	// Build headers map with provider-specific defaults
	headers := make(map[string]string)

	// Set default headers for OpenRouter
	if providerName == "openrouter" {
		headers["HTTP-Referer"] = "https://github.com/vitadin/NeuroShell"
		headers["X-Title"] = "NeuroShell"
	}

	// Create client configuration
	config := OpenAICompatibleConfig{
		ProviderName: providerName,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Headers:      headers,
	}

	logger.Debug("Creating OpenAI-compatible client", "provider", providerName, "baseURL", baseURL, "headerCount", len(headers))
	return NewOpenAICompatibleClient(config)
}

// createAnthropicCompatibleClient creates a new Anthropic-compatible client using the OpenAI-compatible client infrastructure.
func (f *ClientFactoryService) createAnthropicCompatibleClient(apiKey string) neurotypes.LLMClient {
	// Create Anthropic-specific headers
	headers := map[string]string{
		"anthropic-version": "2023-06-01",
	}

	// Create client configuration for Anthropic
	config := OpenAICompatibleConfig{
		ProviderName: "anthropic",
		APIKey:       apiKey,
		BaseURL:      "https://api.anthropic.com/v1",
		Headers:      headers,
		Endpoint:     "/messages",
	}

	logger.Debug("Creating Anthropic-compatible client", "baseURL", config.BaseURL, "endpoint", config.Endpoint)
	return NewOpenAICompatibleClient(config)
}
