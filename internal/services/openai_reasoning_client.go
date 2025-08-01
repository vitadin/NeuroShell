package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// OpenAIReasoningClient implements the LLMClient interface with dual API support.
// It automatically detects reasoning models and uses the appropriate OpenAI API endpoint:
// - /chat/completions for regular GPT models
// - /responses for reasoning models (o3, o4-mini, o1, etc.)
type OpenAIReasoningClient struct {
	apiKey         string
	client         *openai.Client
	debugTransport http.RoundTripper
}

// NewOpenAIReasoningClient creates a new OpenAI reasoning client with lazy initialization.
// The actual OpenAI client is created only when the first request is made.
func NewOpenAIReasoningClient(apiKey string) *OpenAIReasoningClient {
	return &OpenAIReasoningClient{
		apiKey: apiKey,
		client: nil, // Will be initialized lazily
	}
}

// GetProviderName returns the provider name for this client.
func (c *OpenAIReasoningClient) GetProviderName() string {
	return "openai"
}

// IsConfigured returns true if the client has a valid API key.
func (c *OpenAIReasoningClient) IsConfigured() bool {
	return c.apiKey != ""
}

// SetDebugTransport sets the HTTP transport for network debugging.
func (c *OpenAIReasoningClient) SetDebugTransport(transport http.RoundTripper) {
	c.debugTransport = transport
	// Clear the existing client to force re-initialization with debug transport
	c.client = nil
}

// initializeClientIfNeeded initializes the OpenAI client if it hasn't been initialized yet.
func (c *OpenAIReasoningClient) initializeClientIfNeeded() error {
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
		logger.Debug("OpenAI reasoning client initialized with debug transport", "provider", "openai")
	} else {
		logger.Debug("OpenAI reasoning client initialized", "provider", "openai")
	}

	client := openai.NewClient(options...)
	c.client = &client

	return nil
}

// SendChatCompletion sends a chat completion request to OpenAI.
// Automatically routes to /responses endpoint for reasoning models or /chat/completions for regular models.
func (c *OpenAIReasoningClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("OpenAI SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return "", fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	// Check if this is a reasoning model
	isReasoningModel := c.isReasoningModel(modelConfig)
	logger.Debug("Model type detected", "is_reasoning", isReasoningModel, "model", modelConfig.BaseModel)

	if isReasoningModel {
		return c.sendReasoningCompletion(session, modelConfig)
	}
	return c.sendChatCompletion(session, modelConfig)
}

// StreamChatCompletion sends a streaming chat completion request to OpenAI.
// Automatically routes to appropriate endpoint based on model type.
func (c *OpenAIReasoningClient) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.Debug("OpenAI StreamChatCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	// Check if this is a reasoning model
	isReasoningModel := c.isReasoningModel(modelConfig)
	logger.Debug("Model type detected for streaming", "is_reasoning", isReasoningModel, "model", modelConfig.BaseModel)

	if isReasoningModel {
		return c.streamReasoningCompletion(session, modelConfig)
	}
	return c.streamChatCompletion(session, modelConfig)
}

// isReasoningModel determines if a model should use reasoning mode based on explicit parameters.
// Only uses reasoning mode when reasoning_effort parameter is explicitly provided.
// This allows O-series models to use both chat and reasoning modes based on user intent.
func (c *OpenAIReasoningClient) isReasoningModel(modelConfig *neurotypes.ModelConfig) bool {
	// Only use reasoning mode if reasoning_effort parameter is explicitly provided
	if _, hasReasoningEffort := modelConfig.Parameters["reasoning_effort"]; hasReasoningEffort {
		return true
	}

	// Default to chat mode for all models (including O-series)
	// This allows O-series models to work in both modes based on user parameters
	return false
}

// sendChatCompletion handles regular chat completions via /chat/completions endpoint.
func (c *OpenAIReasoningClient) sendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	// Convert session messages to OpenAI format
	messages := c.convertMessagesToOpenAI(session)
	logger.Debug("Messages converted for chat completion", "message_count", len(messages))

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := openai.SystemMessage(session.SystemPrompt)
		messages = append([]openai.ChatCompletionMessageParamUnion{systemMsg}, messages...)
		logger.Debug("System prompt added to chat completion", "system_prompt", session.SystemPrompt)
	}

	// Build completion parameters
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(modelConfig.BaseModel),
		Messages: messages,
	}
	logger.Debug("Chat completion parameters built", "model", modelConfig.BaseModel, "message_count", len(messages))

	// Apply model parameters if available
	c.applyChatParameters(&params, modelConfig)

	// Send request
	logger.Debug("Sending OpenAI chat completion request", "model", modelConfig.BaseModel)
	completion, err := c.client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		logger.Error("OpenAI chat completion request failed", "error", err)
		return "", fmt.Errorf("openai chat completion request failed: %w", err)
	}

	// Extract response content
	if len(completion.Choices) == 0 {
		logger.Error("No response choices returned from chat completion")
		return "", fmt.Errorf("no response choices returned")
	}

	content := completion.Choices[0].Message.Content
	if content == "" {
		logger.Error("Empty response content from chat completion")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("OpenAI chat completion response received", "content_length", len(content))
	return content, nil
}

// sendReasoningCompletion handles reasoning completions via /responses endpoint.
func (c *OpenAIReasoningClient) sendReasoningCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	// Convert session messages to responses API format
	input := c.convertSessionToReasoningInput(session)
	logger.Debug("Messages converted for reasoning completion")

	// Build reasoning parameters
	params := responses.ResponseNewParams{
		Model: shared.ResponsesModel(modelConfig.BaseModel),
		Input: input,
	}

	// Apply reasoning-specific parameters
	c.applyReasoningParameters(&params, modelConfig)

	// Send request to /responses endpoint
	logger.Debug("Sending OpenAI reasoning completion request", "model", modelConfig.BaseModel)
	response, err := c.client.Responses.New(context.Background(), params)
	if err != nil {
		logger.Error("OpenAI reasoning completion request failed", "error", err)
		return "", fmt.Errorf("openai reasoning completion request failed: %w", err)
	}

	// Extract response content from output items
	if len(response.Output) == 0 {
		logger.Error("No response output items returned from reasoning completion")
		return "", fmt.Errorf("no response output items returned")
	}

	// Process output items: extract reasoning summaries and message content
	var responseContent string
	var reasoningSummaries []string

	logger.Debug("Reasoning response received", "output_count", len(response.Output))

	for i, item := range response.Output {
		logger.Debug("Processing output item", "index", i, "item_type", fmt.Sprintf("%T", item))

		// Extract reasoning summaries (thinking process) if present
		if reasoning := item.AsReasoning(); reasoning.Type == "reasoning" {
			logger.Debug("Found reasoning output", "summary_count", len(reasoning.Summary), "status", reasoning.Status)
			for _, summaryItem := range reasoning.Summary {
				if summaryItem.Type == "summary_text" {
					reasoningSummaries = append(reasoningSummaries, summaryItem.Text)
					preview := summaryItem.Text
					if len(preview) > 100 {
						preview = preview[:100] + "..."
					}
					logger.Debug("Extracted reasoning summary", "text_length", len(summaryItem.Text), "summary_preview", preview)
				}
			}
		}

		// Check if this is a message output item (existing working code)
		if message := item.AsMessage(); message.Type == "message" && message.Role == "assistant" {
			logger.Debug("Found message output", "role", message.Role)
			// Process the content array
			for _, contentItem := range message.Content {
				// Check if this is text content
				if text := contentItem.AsOutputText(); text.Type == "output_text" {
					responseContent += text.Text
				}
			}
		}
	}

	// Prepend reasoning summaries to the response content so they get rendered together
	if len(reasoningSummaries) > 0 {
		var reasoningText string
		reasoningText += "\n🧠 **Reasoning Summary:**\n\n"
		for i, summary := range reasoningSummaries {
			if len(reasoningSummaries) > 1 {
				reasoningText += fmt.Sprintf("**Summary %d:**\n", i+1)
			}
			reasoningText += summary + "\n\n"
		}
		reasoningText += "📝 **Final Response:**\n\n"

		// Combine reasoning summaries with the final response
		responseContent = reasoningText + responseContent
		logger.Debug("Combined reasoning summaries with response", "total_summaries", len(reasoningSummaries))
	}

	if responseContent == "" {
		logger.Error("Empty response content from reasoning completion")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("OpenAI reasoning completion response received", "content_length", len(responseContent))
	return responseContent, nil
}

// streamChatCompletion handles streaming chat completions via /chat/completions endpoint.
func (c *OpenAIReasoningClient) streamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
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
	c.applyChatParameters(&params, modelConfig)

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

// streamReasoningCompletion handles streaming reasoning completions via /responses endpoint.
func (c *OpenAIReasoningClient) streamReasoningCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	// Note: OpenAI responses API may not support streaming in the same way
	// For now, we'll fall back to non-streaming and simulate streaming
	logger.Debug("Reasoning streaming not fully supported, using simulated streaming")

	// Create response channel
	responseChan := make(chan neurotypes.StreamChunk, 10)

	// Start goroutine to handle "streaming" response
	go func() {
		defer close(responseChan)

		// Get full response first
		response, err := c.sendReasoningCompletion(session, modelConfig)
		if err != nil {
			responseChan <- neurotypes.StreamChunk{
				Content: "",
				Done:    true,
				Error:   err,
			}
			return
		}

		// Send response as single chunk
		responseChan <- neurotypes.StreamChunk{
			Content: response,
			Done:    false,
			Error:   nil,
		}

		// Send completion marker
		responseChan <- neurotypes.StreamChunk{
			Content: "",
			Done:    true,
			Error:   nil,
		}
	}()

	return responseChan, nil
}

// convertMessagesToOpenAI converts NeuroShell messages to OpenAI chat completion format.
func (c *OpenAIReasoningClient) convertMessagesToOpenAI(session *neurotypes.ChatSession) []openai.ChatCompletionMessageParamUnion {
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

// convertSessionToReasoningInput converts NeuroShell session to OpenAI Responses API input format.
func (c *OpenAIReasoningClient) convertSessionToReasoningInput(session *neurotypes.ChatSession) responses.ResponseNewParamsInputUnion {
	input := make(responses.ResponseInputParam, 0, len(session.Messages)+1)

	// Add system prompt as instruction if present
	if session.SystemPrompt != "" {
		msg := responses.ResponseInputItemParamOfMessage(
			session.SystemPrompt,
			responses.EasyInputMessageRoleSystem,
		)
		input = append(input, msg)
	}

	// Convert session messages
	for _, msg := range session.Messages {
		var role responses.EasyInputMessageRole
		switch msg.Role {
		case "user":
			role = responses.EasyInputMessageRoleUser
		case "assistant":
			role = responses.EasyInputMessageRoleAssistant
		case "system":
			role = responses.EasyInputMessageRoleSystem
		default:
			// Skip unknown roles
			continue
		}

		convertedMsg := responses.ResponseInputItemParamOfMessage(
			msg.Content,
			role,
		)
		input = append(input, convertedMsg)
	}

	return responses.ResponseNewParamsInputUnion{
		OfInputItemList: input,
	}
}

// applyChatParameters applies model configuration parameters to chat completion requests.
func (c *OpenAIReasoningClient) applyChatParameters(params *openai.ChatCompletionNewParams, modelConfig *neurotypes.ModelConfig) {
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

// applyReasoningParameters applies reasoning-specific parameters to responses API requests.
func (c *OpenAIReasoningClient) applyReasoningParameters(params *responses.ResponseNewParams, modelConfig *neurotypes.ModelConfig) {
	if modelConfig.Parameters == nil {
		return
	}

	// Create reasoning parameter if reasoning-specific params are present
	var reasoningParam shared.ReasoningParam

	// Apply reasoning_effort
	if effort, ok := modelConfig.Parameters["reasoning_effort"]; ok {
		if effortStr, ok := effort.(string); ok {
			reasoningParam.Effort = shared.ReasoningEffort(effortStr)
		}
	}

	// Apply reasoning_summary
	if summary, ok := modelConfig.Parameters["reasoning_summary"]; ok {
		if summaryStr, ok := summary.(string); ok {
			reasoningParam.Summary = shared.ReasoningSummary(summaryStr)
		}
	}

	// Set reasoning parameter if configured
	params.Reasoning = reasoningParam

	// Apply max_output_tokens (for reasoning models)
	if maxTokens, ok := modelConfig.Parameters["max_output_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			params.MaxOutputTokens = param.NewOpt(int64(maxTokensInt))
		}
	} else if maxTokens, ok := modelConfig.Parameters["max_tokens"]; ok {
		// Fallback to max_tokens for compatibility
		if maxTokensInt, ok := maxTokens.(int); ok {
			params.MaxOutputTokens = param.NewOpt(int64(maxTokensInt))
		}
	}

	// Apply temperature
	if temp, ok := modelConfig.Parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			params.Temperature = param.NewOpt(tempFloat)
		}
	}

	// Apply top_p
	if topP, ok := modelConfig.Parameters["top_p"]; ok {
		if topPFloat, ok := topP.(float64); ok {
			params.TopP = param.NewOpt(topPFloat)
		}
	}
}
