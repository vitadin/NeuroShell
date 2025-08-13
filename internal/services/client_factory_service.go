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

// GetClientForProvider returns an LLM client for the specified provider catalog ID and API key.
// This allows for explicit provider catalog selection when multiple endpoints are supported.
func (f *ClientFactoryService) GetClientForProvider(providerCatalogID, apiKey string) (neurotypes.LLMClient, error) {
	if !f.initialized {
		return nil, fmt.Errorf("client factory service not initialized")
	}

	if providerCatalogID == "" {
		return nil, fmt.Errorf("provider catalog ID cannot be empty")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty for provider catalog ID '%s'", providerCatalogID)
	}

	// Generate client ID for caching
	clientID := f.generateClientID(providerCatalogID, apiKey)
	ctx := neuroshellcontext.GetGlobalContext()

	// Check if client exists in context cache
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Returning cached provider client", "provider_catalog_id", providerCatalogID, "clientID", clientID)
		return client, nil
	}

	// Create client based on provider catalog ID
	var client neurotypes.LLMClient
	switch providerCatalogID {
	case "OAC": // OpenAI Chat Completions
		client = NewOpenAIClient(apiKey)
	case "OAR": // OpenAI Reasoning/Responses
		client = NewOpenAIReasoningClient(apiKey)
	case "ANC": // Anthropic Chat
		client = NewAnthropicClient(apiKey)
	case "GMC": // Gemini Chat
		client = NewGeminiClient(apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider catalog ID '%s'. Supported catalog IDs: OAC, OAR, ANC, GMC", providerCatalogID)
	}

	// Store client in context cache
	ctx.SetLLMClient(clientID, client)

	logger.Debug("Created new provider client", "provider_catalog_id", providerCatalogID, "clientID", clientID)
	return client, nil
}

// generateClientID creates a unique, secure client ID for the given provider catalog ID and API key.
// Uses SHA-256 hash with first 8 hex characters for uniqueness while maintaining usability.
// Format: "catalog_id:hashed-api-key" (e.g., "OAC:a1b2c3d4")
func (f *ClientFactoryService) generateClientID(providerCatalogID, apiKey string) string {
	if apiKey == "" {
		return fmt.Sprintf("%s:empty***", providerCatalogID)
	}

	// Generate SHA-256 hash of the API key
	hash := sha256.Sum256([]byte(apiKey))

	// Convert to hex and take first 8 characters for usability
	hexHash := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s:%s", providerCatalogID, hexHash[:8])
}

// GetClientWithID returns both an LLM client and its unique client ID for the specified provider catalog ID and API key.
// This method provides both client management and consistent ID generation for external use.
func (f *ClientFactoryService) GetClientWithID(providerCatalogID, apiKey string) (neurotypes.LLMClient, string, error) {
	if !f.initialized {
		return nil, "", fmt.Errorf("client factory service not initialized")
	}

	if providerCatalogID == "" {
		return nil, "", fmt.Errorf("provider catalog ID cannot be empty")
	}

	if apiKey == "" {
		return nil, "", fmt.Errorf("API key cannot be empty for provider catalog ID '%s'", providerCatalogID)
	}

	// Generate client ID for external use (also serves as cache key)
	clientID := f.generateClientID(providerCatalogID, apiKey)
	ctx := neuroshellcontext.GetGlobalContext()

	// Check if client exists in context cache
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Returning cached provider client with ID", "provider_catalog_id", providerCatalogID, "clientID", clientID)
		return client, clientID, nil
	}

	// Create client based on provider catalog ID
	var client neurotypes.LLMClient
	switch providerCatalogID {
	case "OAC": // OpenAI Chat Completions
		client = NewOpenAIClient(apiKey)
	case "OAR": // OpenAI Reasoning/Responses
		client = NewOpenAIReasoningClient(apiKey)
	case "ANC": // Anthropic Chat
		client = NewAnthropicClient(apiKey)
	case "GMC": // Gemini Chat
		client = NewGeminiClient(apiKey)
	default:
		return nil, "", fmt.Errorf("unsupported provider catalog ID '%s'. Supported catalog IDs: OAC, OAR, ANC, GMC", providerCatalogID)
	}

	// Store client in context cache
	ctx.SetLLMClient(clientID, client)

	logger.Debug("Created new provider client with ID", "provider_catalog_id", providerCatalogID, "clientID", clientID)
	return client, clientID, nil
}

// GetClientByID retrieves a cached LLM client by its client ID.
// Client ID format: "catalog_id:hashed-api-key" (e.g., "OAC:a1b2c3d4...")
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
		return nil, fmt.Errorf("invalid client ID format: %s (expected 'catalog_id:hash')", clientID)
	}

	// Ensure both catalog ID and hash parts are non-empty
	parts := strings.Split(clientID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid client ID format: %s (expected 'catalog_id:hash')", clientID)
	}

	// Direct lookup using client ID from context
	ctx := neuroshellcontext.GetGlobalContext()
	if client, exists := ctx.GetLLMClient(clientID); exists {
		logger.Debug("Retrieved cached client by ID", "clientID", clientID)
		return client, nil
	}

	return nil, fmt.Errorf("client with ID '%s' not found in cache", clientID)
}

// FindClientByProviderCatalogID finds any existing client with the specified provider catalog ID.
// This method scans all cached clients and returns the first match found.
// Used by llm-client-activate to activate clients by provider catalog ID.
func (f *ClientFactoryService) FindClientByProviderCatalogID(providerCatalogID string) (neurotypes.LLMClient, string, error) {
	if !f.initialized {
		return nil, "", fmt.Errorf("client factory service not initialized")
	}

	if providerCatalogID == "" {
		return nil, "", fmt.Errorf("provider catalog ID cannot be empty")
	}

	// Get all cached clients from context
	ctx := neuroshellcontext.GetGlobalContext()
	allClients := ctx.GetAllLLMClients()

	// Search for any client with matching provider catalog ID prefix
	expectedPrefix := providerCatalogID + ":"
	for clientID, client := range allClients {
		if strings.HasPrefix(clientID, expectedPrefix) {
			logger.Debug("Found client by provider catalog ID", "provider_catalog_id", providerCatalogID, "clientID", clientID)
			return client, clientID, nil
		}
	}

	return nil, "", fmt.Errorf("no client found with provider catalog ID '%s'", providerCatalogID)
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
