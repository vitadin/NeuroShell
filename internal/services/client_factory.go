package services

import (
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

// GetClient returns an LLM client for the given API key.
// If a client for this API key already exists, it returns the cached client.
// If not, it creates a new client and caches it for future use.
func (f *ClientFactoryService) GetClient(apiKey string) (neurotypes.LLMClient, error) {
	if !f.initialized {
		return nil, fmt.Errorf("client factory service not initialized")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	f.mutex.RLock()
	if client, exists := f.clients[apiKey]; exists {
		f.mutex.RUnlock()
		logger.Debug("Returning cached client", "provider", client.GetProviderName())
		return client, nil
	}
	f.mutex.RUnlock()

	// Create new client with write lock
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Double-check pattern - another goroutine might have created it
	if client, exists := f.clients[apiKey]; exists {
		logger.Debug("Returning cached client (double-check)", "provider", client.GetProviderName())
		return client, nil
	}

	// Create new OpenAI client (lazy initialization)
	client := NewOpenAIClient(apiKey)
	f.clients[apiKey] = client

	logger.Debug("Created new client", "provider", client.GetProviderName())
	return client, nil
}

// GetClientForProvider returns an LLM client for the specified provider and API key.
// This allows for explicit provider selection when multiple providers are supported.
func (f *ClientFactoryService) GetClientForProvider(provider, apiKey string) (neurotypes.LLMClient, error) {
	if !f.initialized {
		return nil, fmt.Errorf("client factory service not initialized")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
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
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	f.clients[cacheKey] = client

	logger.Debug("Created new provider client", "provider", provider)
	return client, nil
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
