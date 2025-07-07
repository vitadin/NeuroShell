package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// APIResponse represents a standardized response from LLM providers.
type APIResponse struct {
	Content      string                 `json:"content"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	Usage        *UsageInfo             `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
}

// UsageInfo contains token usage information from API responses.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo represents unified model information across providers.
type ModelInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	Owned       string   `json:"owned_by,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	CreatedAt   int64    `json:"created,omitempty"`
}

// APIService provides API connectivity checking and management for LLM providers.
type APIService struct {
	initialized  bool
	httpClient   *http.Client
	timeout      time.Duration
	endpoints    map[string]string
	openaiClient *openai.Client
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

	// Initialize OpenAI client
	if err := a.initializeOpenAIClient(); err != nil {
		logger.Debug("Failed to initialize OpenAI client", "error", err)
		// Don't fail initialization if OpenAI client fails - just log it
	}

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

// initializeOpenAIClient sets up the OpenAI client with API key from environment or context.
func (a *APIService) initializeOpenAIClient() error {
	// Get OpenAI API key
	apiKey, err := a.getAPIKey("openai")
	if err != nil {
		return fmt.Errorf("failed to get OpenAI API key: %w", err)
	}

	// Create OpenAI client with options
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	// Check for custom base URL
	if baseURL := a.getOpenAIBaseURL(); baseURL != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(options...)
	a.openaiClient = &client
	logger.Debug("OpenAI client initialized successfully")
	return nil
}

// getOpenAIBaseURL retrieves custom OpenAI base URL from environment or context.
func (a *APIService) getOpenAIBaseURL() string {
	ctx := neuroshellcontext.GetGlobalContext()

	// Check context variable first
	if baseURL, err := ctx.GetVariable("@openai_base_url"); err == nil && baseURL != "" {
		logger.Debug("Using context OpenAI base URL", "baseURL", baseURL)
		return baseURL
	}

	// Check environment variable
	if baseURL := os.Getenv("OPENAI_API_BASE_URL"); baseURL != "" {
		logger.Debug("Using environment OpenAI base URL", "baseURL", baseURL)
		return baseURL
	}

	return ""
}

// SendMessage sends a message to the specified provider and model, returning a standardized response.
func (a *APIService) SendMessage(provider, model, message string, options map[string]any) (*APIResponse, error) {
	if !a.initialized {
		return nil, fmt.Errorf("api service not initialized")
	}

	switch strings.ToLower(provider) {
	case "openai":
		return a.sendOpenAIMessage(model, message, options)
	case "anthropic":
		return nil, fmt.Errorf("anthropic provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendOpenAIMessage handles sending messages to OpenAI using the official client.
func (a *APIService) sendOpenAIMessage(model, message string, options map[string]any) (*APIResponse, error) {
	if a.openaiClient == nil {
		return nil, fmt.Errorf("openai client not initialized")
	}

	// Build chat completion parameters
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(message),
		},
	}

	// Apply optional parameters
	if options != nil {
		if temp, ok := options["temperature"]; ok {
			if tempFloat, ok := temp.(float64); ok {
				params.Temperature = openai.Float(tempFloat)
			}
		}
		if maxTokens, ok := options["max_tokens"]; ok {
			if maxTokensInt, ok := maxTokens.(int); ok {
				params.MaxTokens = openai.Int(int64(maxTokensInt))
			}
		}
		if topP, ok := options["top_p"]; ok {
			if topPFloat, ok := topP.(float64); ok {
				params.TopP = openai.Float(topPFloat)
			}
		}
	}

	// Make the API call
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	completion, err := a.openaiClient.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai api call failed: %w", err)
	}

	// Convert to standardized response
	response := &APIResponse{
		Provider: "openai",
		Model:    model,
	}

	if len(completion.Choices) > 0 {
		response.Content = completion.Choices[0].Message.Content
		response.FinishReason = string(completion.Choices[0].FinishReason)
	}

	if completion.Usage.PromptTokens > 0 || completion.Usage.CompletionTokens > 0 || completion.Usage.TotalTokens > 0 {
		response.Usage = &UsageInfo{
			PromptTokens:     int(completion.Usage.PromptTokens),
			CompletionTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:      int(completion.Usage.TotalTokens),
		}
	}

	// Add metadata
	response.Metadata = map[string]interface{}{
		"id":      completion.ID,
		"object":  completion.Object,
		"created": completion.Created,
	}

	if completion.SystemFingerprint != "" {
		response.Metadata["system_fingerprint"] = completion.SystemFingerprint
	}

	logger.Debug("OpenAI message sent successfully", "model", model, "usage", response.Usage)
	return response, nil
}

// SendMessageStreaming sends a message with streaming responses via callback.
func (a *APIService) SendMessageStreaming(provider, model, message string, options map[string]any, callback func(chunk string)) error {
	if !a.initialized {
		return fmt.Errorf("api service not initialized")
	}

	switch strings.ToLower(provider) {
	case "openai":
		return a.sendOpenAIMessageStreaming(model, message, options, callback)
	case "anthropic":
		return fmt.Errorf("anthropic provider not yet implemented")
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendOpenAIMessageStreaming handles streaming responses from OpenAI.
func (a *APIService) sendOpenAIMessageStreaming(model, message string, options map[string]any, callback func(chunk string)) error {
	if a.openaiClient == nil {
		return fmt.Errorf("openai client not initialized")
	}

	// Build chat completion parameters for streaming
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(message),
		},
	}

	// Apply optional parameters
	if options != nil {
		if temp, ok := options["temperature"]; ok {
			if tempFloat, ok := temp.(float64); ok {
				params.Temperature = openai.Float(tempFloat)
			}
		}
		if maxTokens, ok := options["max_tokens"]; ok {
			if maxTokensInt, ok := maxTokens.(int); ok {
				params.MaxTokens = openai.Int(int64(maxTokensInt))
			}
		}
		if topP, ok := options["top_p"]; ok {
			if topPFloat, ok := topP.(float64); ok {
				params.TopP = openai.Float(topPFloat)
			}
		}
	}

	// Create streaming context
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	// For now, use non-streaming API and call callback with full response
	// Streaming implementation would require additional setup with SSE
	completion, err := a.openaiClient.Chat.Completions.New(ctx, params)
	if err != nil {
		return fmt.Errorf("openai api call failed: %w", err)
	}

	// Call callback with the complete response
	if len(completion.Choices) > 0 {
		callback(completion.Choices[0].Message.Content)
	}

	logger.Debug("OpenAI streaming completed successfully", "model", model)
	return nil
}

// ListProviderModels lists available models from the specified provider.
func (a *APIService) ListProviderModels(provider string) ([]ModelInfo, error) {
	if !a.initialized {
		return nil, fmt.Errorf("api service not initialized")
	}

	switch strings.ToLower(provider) {
	case "openai":
		return a.listOpenAIModels()
	case "anthropic":
		return nil, fmt.Errorf("anthropic provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// listOpenAIModels retrieves available models from OpenAI.
func (a *APIService) listOpenAIModels() ([]ModelInfo, error) {
	if a.openaiClient == nil {
		return nil, fmt.Errorf("openai client not initialized")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	// Call OpenAI models API
	models, err := a.openaiClient.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list openai models: %w", err)
	}

	// Convert to standardized format
	var modelInfos []ModelInfo
	for _, model := range models.Data {
		modelInfo := ModelInfo{
			ID:        model.ID,
			Name:      model.ID, // OpenAI uses ID as name
			Provider:  "openai",
			Owned:     model.OwnedBy,
			CreatedAt: model.Created,
		}

		// Extract permissions if available
		// Note: OpenAI API doesn't always provide detailed permissions in the list endpoint
		// This is a placeholder for future extension
		modelInfo.Permissions = []string{}

		modelInfos = append(modelInfos, modelInfo)
	}

	logger.Debug("OpenAI models listed successfully", "count", len(modelInfos))
	return modelInfos, nil
}
