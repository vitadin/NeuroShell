package services

import (
	"fmt"
	"os"
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

// DetermineAPIKeyForProvider determines the API key for a specific provider.
// It checks provider-specific environment variables in order of preference.
func (f *ClientFactoryService) DetermineAPIKeyForProvider(provider string) (string, error) {
	if provider == "" {
		return "", fmt.Errorf("provider cannot be empty")
	}

	var apiKey string
	var envVarName string

	// Check provider-specific environment variables
	switch provider {
	case "openai":
		envVarName = "OPENAI_API_KEY"
		apiKey = os.Getenv(envVarName)
	case "anthropic":
		envVarName = "ANTHROPIC_API_KEY"
		apiKey = os.Getenv(envVarName)
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
