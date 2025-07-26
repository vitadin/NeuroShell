package services

import (
	"context"
	"fmt"

	"google.golang.org/genai"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// GeminiClient implements the LLMClient interface for Google Gemini API.
// It provides lazy initialization of the Gemini client and handles
// all Gemini-specific communication logic.
type GeminiClient struct {
	apiKey string
	client *genai.Client
}

// NewGeminiClient creates a new Gemini client with lazy initialization.
// The actual Gemini client is created only when the first request is made.
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		client: nil, // Will be initialized lazily
	}
}

// GetProviderName returns the provider name for this client.
func (c *GeminiClient) GetProviderName() string {
	return "gemini"
}

// IsConfigured returns true if the client has a valid API key.
func (c *GeminiClient) IsConfigured() bool {
	return c.apiKey != ""
}

// initializeClientIfNeeded initializes the Gemini client if it hasn't been initialized yet.
func (c *GeminiClient) initializeClientIfNeeded() error {
	if c.client != nil {
		return nil // Already initialized
	}

	if c.apiKey == "" {
		return fmt.Errorf("google API key not configured")
	}

	// Create Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	c.client = client
	logger.Debug("Gemini client initialized", "provider", "gemini")
	return nil
}

// SendChatCompletion sends a chat completion request to Google Gemini.
func (c *GeminiClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("Gemini SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return "", fmt.Errorf("failed to initialize Gemini client: %w", err)
	}

	// Convert session messages to Gemini format
	contents := c.convertMessagesToGemini(session)
	logger.Debug("Messages converted", "content_count", len(contents))

	// Build generation config from model parameters
	config := c.buildGenerationConfig(modelConfig)

	// Send request to Gemini
	ctx := context.Background()
	result, err := c.client.Models.GenerateContent(
		ctx,
		modelConfig.BaseModel,
		contents,
		config,
	)
	if err != nil {
		logger.Error("Gemini request failed", "error", err)
		return "", fmt.Errorf("gemini request failed: %w", err)
	}

	// Extract response text
	responseText := result.Text()
	if responseText == "" {
		logger.Error("Empty response content")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("Gemini response received", "content_length", len(responseText))
	return responseText, nil
}

// StreamChatCompletion sends a streaming chat completion request to Google Gemini.
func (c *GeminiClient) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.Debug("Gemini StreamChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize Gemini client: %w", err)
	}

	// Convert session messages to Gemini format
	contents := c.convertMessagesToGemini(session)

	// Build generation config from model parameters
	config := c.buildGenerationConfig(modelConfig)

	// Create streaming request
	ctx := context.Background()
	stream := c.client.Models.GenerateContentStream(
		ctx,
		modelConfig.BaseModel,
		contents,
		config,
	)

	// Create response channel
	responseChan := make(chan neurotypes.StreamChunk, 10)

	// Start goroutine to handle streaming response
	go func() {
		defer close(responseChan)

		for response, err := range stream {
			if err != nil {
				// Stream error
				responseChan <- neurotypes.StreamChunk{
					Content: "",
					Done:    true,
					Error:   fmt.Errorf("stream error: %w", err),
				}
				return
			}

			// Extract content from response
			responseText := response.Text()
			if responseText != "" {
				responseChan <- neurotypes.StreamChunk{
					Content: responseText,
					Done:    false,
					Error:   nil,
				}
			}
		}

		// Stream completed successfully
		responseChan <- neurotypes.StreamChunk{
			Content: "",
			Done:    true,
			Error:   nil,
		}
	}()

	return responseChan, nil
}

// convertMessagesToGemini converts NeuroShell messages to Gemini format.
// Returns the conversation as a slice of genai.Content.
func (c *GeminiClient) convertMessagesToGemini(session *neurotypes.ChatSession) []*genai.Content {
	contents := make([]*genai.Content, 0)

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemContents := genai.Text("System: " + session.SystemPrompt)
		contents = append(contents, systemContents...)
	}

	// Convert conversation messages
	for _, msg := range session.Messages {
		var prefix string
		switch msg.Role {
		case "user":
			prefix = "User: "
		case "assistant":
			prefix = "Assistant: "
		case "system":
			prefix = "System: "
		default:
			// Skip unknown roles
			continue
		}

		msgContents := genai.Text(prefix + msg.Content)
		contents = append(contents, msgContents...)
	}

	// If no contents were added, add a default empty text content
	if len(contents) == 0 {
		emptyContents := genai.Text("")
		contents = append(contents, emptyContents...)
	}

	return contents
}

// buildGenerationConfig creates a Gemini generation config from NeuroShell model parameters.
func (c *GeminiClient) buildGenerationConfig(modelConfig *neurotypes.ModelConfig) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	if modelConfig.Parameters == nil {
		return config
	}

	// Apply temperature
	if temp, ok := modelConfig.Parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			tempFloat32 := float32(tempFloat)
			config.Temperature = &tempFloat32
		}
	}

	// Apply max_tokens (mapped to MaxOutputTokens)
	if maxTokens, ok := modelConfig.Parameters["max_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			maxTokensInt32 := int32(maxTokensInt)
			config.MaxOutputTokens = maxTokensInt32
		}
	}

	return config
}
