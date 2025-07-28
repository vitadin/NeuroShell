// Package services provides LLM client implementations and core services for the NeuroShell CLI.
package services

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// AnthropicClient implements the LLMClient interface for Anthropic's API.
// It provides lazy initialization of the Anthropic client and handles
// all Anthropic-specific communication logic.
type AnthropicClient struct {
	apiKey string
	client *anthropic.Client
}

// NewAnthropicClient creates a new Anthropic client with lazy initialization.
// The actual Anthropic client is created only when the first request is made.
func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		client: nil, // Will be initialized lazily
	}
}

// GetProviderName returns the provider name for this client.
func (c *AnthropicClient) GetProviderName() string {
	return "anthropic"
}

// IsConfigured returns true if the client has a valid API key.
func (c *AnthropicClient) IsConfigured() bool {
	return c.apiKey != ""
}

// initializeClientIfNeeded initializes the Anthropic client if it hasn't been initialized yet.
func (c *AnthropicClient) initializeClientIfNeeded() error {
	if c.client != nil {
		return nil // Already initialized
	}

	if c.apiKey == "" {
		return fmt.Errorf("anthropic API key not configured")
	}

	// Create Anthropic client with API key
	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))
	c.client = &client

	logger.Debug("Anthropic client initialized", "provider", "anthropic")
	return nil
}

// SendChatCompletion sends a chat completion request to Anthropic.
func (c *AnthropicClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("Anthropic SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return "", fmt.Errorf("failed to initialize Anthropic client: %w", err)
	}

	// Convert session messages to Anthropic format
	messages, additionalSystemInstructions := c.convertMessagesToAnthropic(session)
	logger.Debug("Messages converted", "message_count", len(messages))

	// Build message parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(modelConfig.BaseModel),
		MaxTokens: 1024, // Default, will be overridden by parameters if set
		Messages:  messages,
	}

	// Add system prompt if present
	var systemPrompt string
	if session.SystemPrompt != "" {
		systemPrompt = session.SystemPrompt
	}

	// Combine with any additional system instructions from conversation
	if additionalSystemInstructions != "" {
		if systemPrompt != "" {
			systemPrompt += "\n\n" + additionalSystemInstructions
		} else {
			systemPrompt = additionalSystemInstructions
		}
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
		logger.Debug("System prompt added", "system_prompt", systemPrompt)
	}

	// Apply other model parameters
	c.applyModelParameters(&params, modelConfig)

	// Send request
	logger.Debug("Sending Anthropic request", "model", modelConfig.BaseModel)
	message, err := c.client.Messages.New(context.Background(), params)
	if err != nil {
		logger.Error("Anthropic request failed", "error", err)
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}

	// Extract response content
	if len(message.Content) == 0 {
		logger.Error("No response content returned")
		return "", fmt.Errorf("no response content returned")
	}

	// Concatenate all text blocks
	var content string
	for _, block := range message.Content {
		// Get text content from block
		content += block.Text
	}

	if content == "" {
		logger.Error("Empty response content")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("Anthropic response received", "content_length", len(content))
	return content, nil
}

// StreamChatCompletion sends a streaming chat completion request to Anthropic.
func (c *AnthropicClient) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.Debug("Anthropic StreamChatCompletion starting", "model", modelConfig.BaseModel)

	// For now, use regular completion and return as single chunk
	content, err := c.SendChatCompletion(session, modelConfig)
	if err != nil {
		return nil, err
	}

	// Create response channel and send the complete response
	responseChan := make(chan neurotypes.StreamChunk, 2)
	go func() {
		defer close(responseChan)

		// Send content as single chunk
		responseChan <- neurotypes.StreamChunk{
			Content: content,
			Done:    false,
			Error:   nil,
		}

		// Send completion signal
		responseChan <- neurotypes.StreamChunk{
			Content: "",
			Done:    true,
			Error:   nil,
		}
	}()

	return responseChan, nil
}

// convertMessagesToAnthropic converts NeuroShell messages to Anthropic format.
// Returns the conversation messages and any additional system instructions found in the conversation.
func (c *AnthropicClient) convertMessagesToAnthropic(session *neurotypes.ChatSession) ([]anthropic.MessageParam, string) {
	messages := make([]anthropic.MessageParam, 0, len(session.Messages))
	var additionalSystemInstructions []string

	for _, msg := range session.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		case "system":
			// Collect system messages to combine with system prompt
			additionalSystemInstructions = append(additionalSystemInstructions, msg.Content)
		default:
			// Skip unknown roles
			continue
		}
	}

	// Combine additional system instructions into a single string
	var combinedSystemInstructions string
	if len(additionalSystemInstructions) > 0 {
		for i, instruction := range additionalSystemInstructions {
			if i > 0 {
				combinedSystemInstructions += "\n\n"
			}
			combinedSystemInstructions += instruction
		}
	}

	return messages, combinedSystemInstructions
}

// applyModelParameters applies model configuration parameters to the Anthropic request.
func (c *AnthropicClient) applyModelParameters(params *anthropic.MessageNewParams, modelConfig *neurotypes.ModelConfig) {
	if modelConfig.Parameters == nil {
		return
	}

	// Apply temperature
	if temp, ok := modelConfig.Parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			params.Temperature = anthropic.Float(tempFloat)
		}
	}

	// Apply max_tokens
	if maxTokens, ok := modelConfig.Parameters["max_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			params.MaxTokens = int64(maxTokensInt)
		}
	}

	// Apply top_p
	if topP, ok := modelConfig.Parameters["top_p"]; ok {
		if topPFloat, ok := topP.(float64); ok {
			params.TopP = anthropic.Float(topPFloat)
		}
	}

	// Apply top_k
	if topK, ok := modelConfig.Parameters["top_k"]; ok {
		if topKInt, ok := topK.(int); ok {
			params.TopK = anthropic.Int(int64(topKInt))
		}
	}
}
