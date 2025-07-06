package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// APIService provides API connectivity checking and management for LLM providers.
type APIService struct {
	initialized bool
	httpClient  *http.Client
	timeout     time.Duration
	endpoints   map[string]string
}

// NewAPIService creates a new APIService instance with default configuration.
func NewAPIService() *APIService {
	return &APIService{
		initialized: false,
		timeout:     30 * time.Second,
		endpoints:   make(map[string]string),
	}
}

// Name returns the service name "api" for registration.
func (a *APIService) Name() string {
	return "api"
}

// Initialize sets up the APIService with configuration from context and environment.
func (a *APIService) Initialize(ctx neurotypes.Context) error {
	if a.initialized {
		return nil
	}

	// Configure timeout
	a.timeout = a.getTimeout(ctx)

	// Set up HTTP client
	a.httpClient = &http.Client{
		Timeout: a.timeout,
	}

	// Configure provider endpoints
	a.endpoints = a.getEndpoints(ctx)

	a.initialized = true
	logger.Debug("APIService initialized", "timeout", a.timeout, "endpoints", len(a.endpoints))
	return nil
}

// getTimeout retrieves the API timeout configuration from context or environment.
func (a *APIService) getTimeout(ctx neurotypes.Context) time.Duration {
	// Check context variable first
	if timeoutVar, err := ctx.GetVariable("@api_timeout"); err == nil && timeoutVar != "" {
		if timeout, err := strconv.Atoi(timeoutVar); err == nil && timeout > 0 {
			logger.Debug("Using context timeout", "timeout", timeout)
			return time.Duration(timeout) * time.Second
		}
	}

	// Check environment variable
	if timeoutEnv := os.Getenv("API_TIMEOUT"); timeoutEnv != "" {
		if timeout, err := strconv.Atoi(timeoutEnv); err == nil && timeout > 0 {
			logger.Debug("Using environment timeout", "timeout", timeout)
			return time.Duration(timeout) * time.Second
		}
	}

	// Default timeout
	return 30 * time.Second
}

// getEndpoints retrieves API endpoints configuration from context or environment.
func (a *APIService) getEndpoints(ctx neurotypes.Context) map[string]string {
	endpoints := make(map[string]string)

	// Default endpoints
	endpoints["openai"] = "https://api.openai.com/v1/models"
	endpoints["anthropic"] = "https://api.anthropic.com/v1/models"

	// Check for custom endpoints in context
	if openaiEndpoint, err := ctx.GetVariable("@openai_endpoint"); err == nil && openaiEndpoint != "" {
		endpoints["openai"] = openaiEndpoint
		logger.Debug("Using context OpenAI endpoint", "endpoint", openaiEndpoint)
	}

	if anthropicEndpoint, err := ctx.GetVariable("@anthropic_endpoint"); err == nil && anthropicEndpoint != "" {
		endpoints["anthropic"] = anthropicEndpoint
		logger.Debug("Using context Anthropic endpoint", "endpoint", anthropicEndpoint)
	}

	// Check environment variables
	if openaiEnv := os.Getenv("OPENAI_API_BASE_URL"); openaiEnv != "" {
		endpoints["openai"] = openaiEnv + "/models"
		logger.Debug("Using environment OpenAI endpoint", "endpoint", endpoints["openai"])
	}

	if anthropicEnv := os.Getenv("ANTHROPIC_API_BASE_URL"); anthropicEnv != "" {
		endpoints["anthropic"] = anthropicEnv + "/models"
		logger.Debug("Using environment Anthropic endpoint", "endpoint", endpoints["anthropic"])
	}

	return endpoints
}

// getAPIKey retrieves API key for a provider from context or environment.
func (a *APIService) getAPIKey(provider string) (string, error) {
	ctx := neuroshellcontext.GetGlobalContext()

	// Check context variable first
	contextVar := fmt.Sprintf("@%s_api_key", provider)
	if apiKey, err := ctx.GetVariable(contextVar); err == nil && apiKey != "" {
		logger.Debug("Using context API key", "provider", provider)
		return apiKey, nil
	}

	// Check environment variable
	envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
	if apiKey := os.Getenv(envVar); apiKey != "" {
		logger.Debug("Using environment API key", "provider", provider)
		return apiKey, nil
	}

	return "", fmt.Errorf("no API key found for provider %s", provider)
}

// CheckConnectivity tests connectivity to a specific provider.
func (a *APIService) CheckConnectivity(provider string) error {
	if !a.initialized {
		return fmt.Errorf("api service not initialized")
	}

	endpoint, exists := a.endpoints[provider]
	if !exists {
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Get API key
	apiKey, err := a.getAPIKey(provider)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers based on provider
	switch provider {
	case "openai":
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case "anthropic":
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	// Make request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()
	req = req.WithContext(ctx)

	logger.Debug("Testing connectivity", "provider", provider, "endpoint", endpoint)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("connection timeout after %v", a.timeout)
		}
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Debug("Failed to close response body", "error", closeErr)
		}
	}()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid API key")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	logger.Debug("Connectivity check successful", "provider", provider, "status", resp.StatusCode)
	return nil
}

// GetProviderStatus checks if a provider is available and returns status.
func (a *APIService) GetProviderStatus(provider string) (bool, error) {
	err := a.CheckConnectivity(provider)
	if err != nil {
		return false, err
	}
	return true, nil
}

// TestAllProviders tests connectivity to all configured providers.
func (a *APIService) TestAllProviders() map[string]error {
	results := make(map[string]error)

	for provider := range a.endpoints {
		results[provider] = a.CheckConnectivity(provider)
	}

	return results
}

// GetSupportedProviders returns a list of supported providers.
func (a *APIService) GetSupportedProviders() []string {
	providers := make([]string, 0, len(a.endpoints))
	for provider := range a.endpoints {
		providers = append(providers, provider)
	}
	return providers
}

// SetTimeout configures the HTTP client timeout.
func (a *APIService) SetTimeout(timeout time.Duration) {
	a.timeout = timeout
	if a.httpClient != nil {
		a.httpClient.Timeout = timeout
	}
}

// GetTimeout returns the current timeout configuration.
func (a *APIService) GetTimeout() time.Duration {
	return a.timeout
}
