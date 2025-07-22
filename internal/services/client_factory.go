package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// ClientFactoryService implements the ClientFactory interface.
// It manages the creation and caching of LLM clients based on API keys.
type ClientFactoryService struct {
	initialized bool
	clients     map[string]neurotypes.LLMClient
	mutex       sync.RWMutex
}

// NewClientFactoryService creates a new ClientFactoryService instance.
func NewClientFactoryService() *ClientFactoryService {
	return &ClientFactoryService{
		initialized: false,
		clients:     make(map[string]neurotypes.LLMClient),
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

	// Create a composite key for provider-specific caching
	cacheKey := fmt.Sprintf("%s:%s", provider, apiKey)

	f.mutex.RLock()
	if client, exists := f.clients[cacheKey]; exists {
		f.mutex.RUnlock()
		logger.Debug("Returning cached provider client", "provider", provider)
		return client, nil
	}
	f.mutex.RUnlock()

	// Create new client with write lock
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Double-check pattern
	if client, exists := f.clients[cacheKey]; exists {
		logger.Debug("Returning cached provider client (double-check)", "provider", provider)
		return client, nil
	}

	// Create client based on provider
	var client neurotypes.LLMClient
	switch provider {
	case "openai":
		client = NewOpenAIClient(apiKey)
	case "anthropic":
		// TODO: Implement AnthropicClient when available
		return nil, fmt.Errorf("anthropic provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider '%s'. Supported providers: openai, anthropic", provider)
	}

	f.clients[cacheKey] = client

	logger.Debug("Created new provider client", "provider", provider)
	return client, nil
}

// generateClientID creates a unique, secure client ID for the given provider and API key.
// Uses SHA-256 hash with first 8 hex characters for uniqueness while maintaining security.
func (f *ClientFactoryService) generateClientID(provider, apiKey string) string {
	if apiKey == "" {
		return fmt.Sprintf("%s:empty***", provider)
	}

	// Generate SHA-256 hash of the API key
	hash := sha256.Sum256([]byte(apiKey))

	// Convert to hex and take first 8 characters
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

	// Generate client ID for external use
	clientID := f.generateClientID(provider, apiKey)

	// Use full API key for internal caching (more secure than using hash for cache key)
	cacheKey := fmt.Sprintf("%s:%s", provider, apiKey)

	f.mutex.RLock()
	if client, exists := f.clients[cacheKey]; exists {
		f.mutex.RUnlock()
		logger.Debug("Returning cached provider client with ID", "provider", provider, "clientID", clientID)
		return client, clientID, nil
	}
	f.mutex.RUnlock()

	// Create new client with write lock
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Double-check pattern
	if client, exists := f.clients[cacheKey]; exists {
		logger.Debug("Returning cached provider client with ID (double-check)", "provider", provider, "clientID", clientID)
		return client, clientID, nil
	}

	// Create client based on provider
	var client neurotypes.LLMClient
	switch provider {
	case "openai":
		client = NewOpenAIClient(apiKey)
	case "anthropic":
		// TODO: Implement AnthropicClient when available
		return nil, "", fmt.Errorf("anthropic provider is not yet implemented")
	default:
		return nil, "", fmt.Errorf("unsupported provider '%s'. Supported providers: openai, anthropic", provider)
	}

	f.clients[cacheKey] = client

	logger.Debug("Created new provider client with ID", "provider", provider, "clientID", clientID)
	return client, clientID, nil
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
	case "anthropic":
		envVarName = "ANTHROPIC_API_KEY"
		apiKey = ctx.GetEnv(envVarName)
	default:
		return "", fmt.Errorf("unsupported provider '%s'. Supported providers: openai, anthropic", provider)
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
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.clients)
}

// ClearCache removes all cached clients (for testing/debugging).
func (f *ClientFactoryService) ClearCache() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.clients = make(map[string]neurotypes.LLMClient)
	logger.Debug("Client cache cleared")
}
