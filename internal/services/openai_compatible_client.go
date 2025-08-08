package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// OpenAICompatibleClient implements the LLMClient interface for OpenAI-compatible APIs.
// This client works with any provider that implements the OpenAI Chat Completions API,
// such as OpenRouter, Together AI, and other OpenAI-compatible services.
type OpenAICompatibleClient struct {
	providerName string
	apiKey       string
	baseURL      string
	headers      map[string]string
	endpoint     string
	httpClient   *http.Client
}

// OpenAICompatibleConfig holds configuration for the OpenAI-compatible client.
type OpenAICompatibleConfig struct {
	ProviderName string
	APIKey       string
	BaseURL      string
	Headers      map[string]string
	Endpoint     string // Custom endpoint path (defaults to "/chat/completions")
}

// ChatCompletionRequest represents the request payload for OpenAI-compatible chat completions.
type ChatCompletionRequest struct {
	Model            string                  `json:"model"`
	Messages         []ChatCompletionMessage `json:"messages"`
	Temperature      *float64                `json:"temperature,omitempty"`
	MaxTokens        *int                    `json:"max_tokens,omitempty"`
	TopP             *float64                `json:"top_p,omitempty"`
	FrequencyPenalty *float64                `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64                `json:"presence_penalty,omitempty"`
	Stream           bool                    `json:"stream,omitempty"`
}

// ChatCompletionMessage represents a message in the chat completion request.
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from OpenAI-compatible chat completions.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *ChatCompletionUsage   `json:"usage,omitempty"`
	Error   *ChatCompletionError   `json:"error,omitempty"`
}

// ChatCompletionChoice represents a choice in the chat completion response.
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	Delta        *ChatCompletionMessage `json:"delta,omitempty"`
	FinishReason *string                `json:"finish_reason"`
}

// ChatCompletionUsage represents token usage information.
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionError represents an error response.
type ChatCompletionError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAICompatibleClient creates a new OpenAI-compatible client.
// If no baseURL is provided, it defaults to OpenRouter's API endpoint.
func NewOpenAICompatibleClient(config OpenAICompatibleConfig) *OpenAICompatibleClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	// Ensure base URL doesn't end with slash for consistent URL building
	baseURL = strings.TrimSuffix(baseURL, "/")

	headers := make(map[string]string)
	if config.Headers != nil {
		for k, v := range config.Headers {
			headers[k] = v
		}
	}

	providerName := config.ProviderName
	if providerName == "" {
		providerName = "openai-compatible"
	}

	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "/chat/completions"
	}

	return &OpenAICompatibleClient{
		providerName: providerName,
		apiKey:       config.APIKey,
		baseURL:      baseURL,
		headers:      headers,
		endpoint:     endpoint,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetProviderName returns the provider name for this client.
func (c *OpenAICompatibleClient) GetProviderName() string {
	return c.providerName
}

// IsConfigured returns true if the client has a valid API key and base URL.
func (c *OpenAICompatibleClient) IsConfigured() bool {
	return c.apiKey != "" && c.baseURL != ""
}

// SendChatCompletion sends a chat completion request to the OpenAI-compatible API.
func (c *OpenAICompatibleClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("OpenAI-compatible SendChatCompletion starting", "model", modelConfig.BaseModel, "baseURL", c.baseURL)

	if !c.IsConfigured() {
		return "", fmt.Errorf("OpenAI-compatible client not configured: missing API key or base URL")
	}

	// Convert session messages to OpenAI format
	messages := c.convertMessagesToOpenAI(session)
	logger.Debug("Messages converted", "message_count", len(messages))

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := ChatCompletionMessage{
			Role:    "system",
			Content: session.SystemPrompt,
		}
		messages = append([]ChatCompletionMessage{systemMsg}, messages...)
		logger.Debug("System prompt added", "system_prompt", session.SystemPrompt)
	}

	// Build completion request
	request := ChatCompletionRequest{
		Model:    modelConfig.BaseModel,
		Messages: messages,
		Stream:   false,
	}
	logger.Debug("Completion request built", "model", modelConfig.BaseModel, "message_count", len(messages))

	// Apply model parameters if available
	c.applyModelParameters(&request, modelConfig)

	// Send HTTP request
	response, err := c.sendHTTPRequest(c.endpoint, request)
	if err != nil {
		logger.Error("OpenAI-compatible request failed", "error", err)
		return "", fmt.Errorf("openai-compatible request failed: %w", err)
	}

	// Parse response
	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(response, &chatResponse); err != nil {
		logger.Error("Failed to parse response", "error", err)
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if chatResponse.Error != nil {
		logger.Error("API returned error", "error", chatResponse.Error.Message)
		return "", fmt.Errorf("API error: %s", chatResponse.Error.Message)
	}

	// Extract response content
	if len(chatResponse.Choices) == 0 {
		logger.Error("No response choices returned")
		return "", fmt.Errorf("no response choices returned")
	}

	if chatResponse.Choices[0].Message == nil {
		logger.Error("No message in response choice")
		return "", fmt.Errorf("no message in response choice")
	}

	content := chatResponse.Choices[0].Message.Content
	if content == "" {
		logger.Error("Empty response content")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("OpenAI-compatible response received", "content_length", len(content))
	return content, nil
}

// SendStructuredCompletion sends a chat completion request to the OpenAI-compatible API and returns structured response.
// Since OpenAI-compatible models don't have native thinking content, this returns regular text with no thinking blocks.
func (c *OpenAICompatibleClient) SendStructuredCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (*neurotypes.StructuredLLMResponse, error) {
	logger.Debug("OpenAI-compatible SendStructuredCompletion starting", "model", modelConfig.BaseModel, "baseURL", c.baseURL)

	// Use regular completion since OpenAI-compatible models don't have native thinking content
	textContent, err := c.SendChatCompletion(session, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("openai-compatible structured completion failed: %w", err)
	}

	// Create structured response with no thinking blocks (compatible models don't provide thinking blocks)
	structuredResponse := &neurotypes.StructuredLLMResponse{
		TextContent:    textContent,
		ThinkingBlocks: []neurotypes.ThinkingBlock{}, // Empty - compatible models don't provide thinking blocks
	}

	logger.Debug("OpenAI-compatible structured response created", "content_length", len(textContent), "thinking_blocks", 0)
	return structuredResponse, nil
}

// SetDebugTransport sets the HTTP transport for network debugging.
// Currently, debug transport is not implemented for OpenAI-compatible client - this is a placeholder.
func (c *OpenAICompatibleClient) SetDebugTransport(_ http.RoundTripper) {
	// Dummy implementation - will be implemented later
	// The OpenAI-compatible client will eventually use this transport for HTTP debugging
}

// StreamChatCompletion sends a streaming chat completion request to the OpenAI-compatible API.
func (c *OpenAICompatibleClient) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.Debug("OpenAI-compatible StreamChatCompletion starting", "model", modelConfig.BaseModel, "baseURL", c.baseURL)

	if !c.IsConfigured() {
		return nil, fmt.Errorf("OpenAI-compatible client not configured: missing API key or base URL")
	}

	// Convert session messages to OpenAI format
	messages := c.convertMessagesToOpenAI(session)

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := ChatCompletionMessage{
			Role:    "system",
			Content: session.SystemPrompt,
		}
		messages = append([]ChatCompletionMessage{systemMsg}, messages...)
	}

	// Build completion request with streaming enabled
	request := ChatCompletionRequest{
		Model:    modelConfig.BaseModel,
		Messages: messages,
		Stream:   true,
	}

	// Apply model parameters if available
	c.applyModelParameters(&request, modelConfig)

	// Create streaming HTTP request
	responseChan := make(chan neurotypes.StreamChunk, 10)

	go func() {
		defer close(responseChan)

		if err := c.sendStreamingHTTPRequest(c.endpoint, request, responseChan); err != nil {
			responseChan <- neurotypes.StreamChunk{
				Content: "",
				Done:    true,
				Error:   err,
			}
		}
	}()

	return responseChan, nil
}

// convertMessagesToOpenAI converts NeuroShell messages to OpenAI-compatible format.
func (c *OpenAICompatibleClient) convertMessagesToOpenAI(session *neurotypes.ChatSession) []ChatCompletionMessage {
	messages := make([]ChatCompletionMessage, 0, len(session.Messages))

	for _, msg := range session.Messages {
		switch msg.Role {
		case "user", "assistant", "system":
			messages = append(messages, ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		default:
			// Skip unknown roles
			continue
		}
	}

	return messages
}

// applyModelParameters applies model configuration parameters to the request.
func (c *OpenAICompatibleClient) applyModelParameters(request *ChatCompletionRequest, modelConfig *neurotypes.ModelConfig) {
	if modelConfig.Parameters == nil {
		return
	}

	// Apply temperature
	if temp, ok := modelConfig.Parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			request.Temperature = &tempFloat
		}
	}

	// Apply max_tokens
	if maxTokens, ok := modelConfig.Parameters["max_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			request.MaxTokens = &maxTokensInt
		}
	}

	// Apply top_p
	if topP, ok := modelConfig.Parameters["top_p"]; ok {
		if topPFloat, ok := topP.(float64); ok {
			request.TopP = &topPFloat
		}
	}

	// Apply frequency_penalty
	if freqPenalty, ok := modelConfig.Parameters["frequency_penalty"]; ok {
		if freqPenaltyFloat, ok := freqPenalty.(float64); ok {
			request.FrequencyPenalty = &freqPenaltyFloat
		}
	}

	// Apply presence_penalty
	if presPenalty, ok := modelConfig.Parameters["presence_penalty"]; ok {
		if presPenaltyFloat, ok := presPenalty.(float64); ok {
			request.PresencePenalty = &presPenaltyFloat
		}
	}
}

// sendHTTPRequest sends a non-streaming HTTP request to the API.
func (c *OpenAICompatibleClient) sendHTTPRequest(endpoint string, payload interface{}) ([]byte, error) {
	// Marshall request payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Set authentication header based on provider
	if c.providerName == "anthropic" {
		req.Header.Set("x-api-key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// sendStreamingHTTPRequest sends a streaming HTTP request to the API.
func (c *OpenAICompatibleClient) sendStreamingHTTPRequest(endpoint string, payload interface{}, responseChan chan<- neurotypes.StreamChunk) error {
	// Marshall request payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Set authentication header based on provider
	if c.providerName == "anthropic" {
		req.Header.Set("x-api-key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Process streaming response
	return c.processStreamingResponse(resp.Body, responseChan)
}

// processStreamingResponse processes Server-Sent Events from the streaming response.
func (c *OpenAICompatibleClient) processStreamingResponse(body io.Reader, responseChan chan<- neurotypes.StreamChunk) error {
	// Read response line by line
	buffer := make([]byte, 4096)
	leftover := ""

	for {
		n, err := body.Read(buffer)
		if n > 0 {
			data := leftover + string(buffer[:n])
			lines := strings.Split(data, "\n")

			// Keep the last incomplete line for next iteration
			if !strings.HasSuffix(data, "\n") {
				leftover = lines[len(lines)-1]
				lines = lines[:len(lines)-1]
			} else {
				leftover = ""
			}

			// Process each complete line
			for _, line := range lines {
				if err := c.processStreamLine(line, responseChan); err != nil {
					return err
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}
	}

	// Send final chunk
	responseChan <- neurotypes.StreamChunk{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	return nil
}

// processStreamLine processes a single line from the Server-Sent Events stream.
func (c *OpenAICompatibleClient) processStreamLine(line string, responseChan chan<- neurotypes.StreamChunk) error {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, ":") {
		return nil
	}

	// Parse Server-Sent Events format
	if strings.HasPrefix(line, "data: ") {
		data := strings.TrimPrefix(line, "data: ")

		// Check for end of stream
		if data == "[DONE]" {
			return nil
		}

		// Parse JSON data
		var response ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			logger.Debug("Failed to parse streaming chunk", "data", data, "error", err)
			return nil // Skip malformed chunks
		}

		// Check for API error
		if response.Error != nil {
			return fmt.Errorf("API error: %s", response.Error.Message)
		}

		// Extract content from delta
		if len(response.Choices) > 0 && response.Choices[0].Delta != nil {
			content := response.Choices[0].Delta.Content
			if content != "" {
				responseChan <- neurotypes.StreamChunk{
					Content: content,
					Done:    false,
					Error:   nil,
				}
			}
		}
	}

	return nil
}
