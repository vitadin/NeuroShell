package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// OpenAIClient implements the LLMClient interface for OpenAI's API.
// It provides lazy initialization of the OpenAI client and handles
// all OpenAI-specific communication logic.
type OpenAIClient struct {
	apiKey         string
	client         *openai.Client
	debugTransport http.RoundTripper
}

// NewOpenAIClient creates a new OpenAI client with lazy initialization.
// The actual OpenAI client is created only when the first request is made.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		client: nil, // Will be initialized lazily
	}
}

// GetProviderName returns the provider name for this client.
func (c *OpenAIClient) GetProviderName() string {
	return "openai"
}

// IsConfigured returns true if the client has a valid API key.
func (c *OpenAIClient) IsConfigured() bool {
	return c.apiKey != ""
}

// SetDebugTransport sets the HTTP transport for network debugging.
func (c *OpenAIClient) SetDebugTransport(transport http.RoundTripper) {
	c.debugTransport = transport
	// Clear the existing client to force re-initialization with debug transport
	c.client = nil
}

// initializeClientIfNeeded initializes the OpenAI client if it hasn't been initialized yet.
func (c *OpenAIClient) initializeClientIfNeeded() error {
	if c.client != nil {
		return nil // Already initialized
	}

	if c.apiKey == "" {
		return fmt.Errorf("OpenAI API key not configured")
	}

	// Create OpenAI client with API key and optional debug transport
	var options []option.RequestOption
	options = append(options, option.WithAPIKey(c.apiKey))

	if c.debugTransport != nil {
		httpClient := &http.Client{Transport: c.debugTransport}
		options = append(options, option.WithHTTPClient(httpClient))
		logger.Debug("OpenAI client initialized with debug transport", "provider", "openai")
	} else {
		logger.Debug("OpenAI client initialized", "provider", "openai")
	}

	client := openai.NewClient(options...)
	c.client = &client

	return nil
}

// SendChatCompletion sends a chat completion request to OpenAI.
func (c *OpenAIClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("OpenAI SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return "", fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	// Convert session messages to OpenAI format
	messages := c.convertMessagesToOpenAI(session)
	logger.Debug("Messages converted", "message_count", len(messages))

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := openai.SystemMessage(session.SystemPrompt)
		messages = append([]openai.ChatCompletionMessageParamUnion{systemMsg}, messages...)
		logger.Debug("System prompt added", "system_prompt", session.SystemPrompt)
	}

	// Build completion parameters
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(modelConfig.BaseModel),
		Messages: messages,
	}
	logger.Debug("Completion parameters built", "model", modelConfig.BaseModel, "message_count", len(messages))

	// Apply model parameters if available
	c.applyModelParameters(&params, modelConfig)

	// Send request
	logger.Debug("Sending OpenAI request", "model", modelConfig.BaseModel)
	completion, err := c.client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		logger.Error("OpenAI request failed", "error", err)
		return "", fmt.Errorf("openai request failed: %w", err)
	}

	// Extract response content
	if len(completion.Choices) == 0 {
		logger.Error("No response choices returned")
		return "", fmt.Errorf("no response choices returned")
	}

	content := completion.Choices[0].Message.Content
	if content == "" {
		logger.Error("Empty response content")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("OpenAI response received", "content_length", len(content))
	return content, nil
}

// StreamChatCompletion sends a streaming chat completion request to OpenAI.
func (c *OpenAIClient) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.Debug("OpenAI StreamChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	// Convert session messages to OpenAI format
	messages := c.convertMessagesToOpenAI(session)

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := openai.SystemMessage(session.SystemPrompt)
		messages = append([]openai.ChatCompletionMessageParamUnion{systemMsg}, messages...)
	}

	// Build completion parameters with streaming enabled
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(modelConfig.BaseModel),
		Messages: messages,
	}

	// Apply model parameters if available
	c.applyModelParameters(&params, modelConfig)

	// Create streaming request
	stream := c.client.Chat.Completions.NewStreaming(context.Background(), params)

	// Create response channel
	responseChan := make(chan neurotypes.StreamChunk, 10)

	// Start goroutine to handle streaming response
	go func() {
		defer close(responseChan)
		defer func() { _ = stream.Close() }()

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				responseChan <- neurotypes.StreamChunk{
					Content: chunk.Choices[0].Delta.Content,
					Done:    false,
				}
			}
		}

		if err := stream.Err(); err != nil {
			responseChan <- neurotypes.StreamChunk{
				Content: "",
				Done:    true,
				Error:   err,
			}
		} else {
			responseChan <- neurotypes.StreamChunk{
				Content: "",
				Done:    true,
				Error:   nil,
			}
		}
	}()

	return responseChan, nil
}

// convertMessagesToOpenAI converts NeuroShell messages to OpenAI format.
func (c *OpenAIClient) convertMessagesToOpenAI(session *neurotypes.ChatSession) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(session.Messages))

	for _, msg := range session.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		default:
			// Skip unknown roles
			continue
		}
	}

	return messages
}

// applyModelParameters applies model configuration parameters to the OpenAI request.
func (c *OpenAIClient) applyModelParameters(params *openai.ChatCompletionNewParams, modelConfig *neurotypes.ModelConfig) {
	if modelConfig.Parameters == nil {
		return
	}

	// Apply temperature
	if temp, ok := modelConfig.Parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			params.Temperature = openai.Float(tempFloat)
		}
	}

	// Apply max_tokens
	if maxTokens, ok := modelConfig.Parameters["max_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			params.MaxTokens = openai.Int(int64(maxTokensInt))
		}
	}

	// Apply top_p
	if topP, ok := modelConfig.Parameters["top_p"]; ok {
		if topPFloat, ok := topP.(float64); ok {
			params.TopP = openai.Float(topPFloat)
		}
	}

	// Apply frequency_penalty
	if freqPenalty, ok := modelConfig.Parameters["frequency_penalty"]; ok {
		if freqPenaltyFloat, ok := freqPenalty.(float64); ok {
			params.FrequencyPenalty = openai.Float(freqPenaltyFloat)
		}
	}

	// Apply presence_penalty
	if presPenalty, ok := modelConfig.Parameters["presence_penalty"]; ok {
		if presPenaltyFloat, ok := presPenalty.(float64); ok {
			params.PresencePenalty = openai.Float(presPenaltyFloat)
		}
	}
}
