// Package services provides LLM client implementations and core services for the NeuroShell CLI.
package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ThinkingInfo contains information about thinking blocks found in the response.
type ThinkingInfo struct {
	ThinkingBlocks int // Number of thinking blocks processed
	RedactedBlocks int // Number of redacted thinking blocks processed
	ThinkingTokens int // Estimated thinking tokens used (if available)
}

// AnthropicClient implements the LLMClient interface for Anthropic's API.
// It provides lazy initialization of the Anthropic client and handles
// all Anthropic-specific communication logic.
type AnthropicClient struct {
	apiKey         string
	client         *anthropic.Client
	debugTransport http.RoundTripper
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

	// Create Anthropic client with API key and optional debug transport
	var options []option.RequestOption
	options = append(options, option.WithAPIKey(c.apiKey))

	if c.debugTransport != nil {
		httpClient := &http.Client{Transport: c.debugTransport}
		options = append(options, option.WithHTTPClient(httpClient))
		logger.Debug("Anthropic client initialized with debug transport", "provider", "anthropic")
	} else {
		logger.Debug("Anthropic client initialized", "provider", "anthropic")
	}

	client := anthropic.NewClient(options...)
	c.client = &client
	return nil
}

// SendChatCompletion sends a chat completion request to Anthropic.
func (c *AnthropicClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("Anthropic SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Use shared request logic to get raw response
	message, err := c.sendChatCompletionRequest(session, modelConfig)
	if err != nil {
		return "", err
	}

	// Extract response content with formatted thinking (traditional processing)
	if len(message.Content) == 0 {
		logger.Error("No response content returned")
		return "", fmt.Errorf("no response content returned")
	}

	// Process all content blocks (text, thinking, redacted_thinking) with formatting
	content, thinkingInfo := c.processResponseBlocks(message.Content)

	if content == "" {
		logger.Error("Empty response content")
		return "", fmt.Errorf("empty response content")
	}

	logger.Debug("Anthropic response received", "content_length", len(content), "thinking_blocks", thinkingInfo.ThinkingBlocks, "redacted_blocks", thinkingInfo.RedactedBlocks)
	return content, nil
}

// sendChatCompletionRequest handles the core request logic shared by both SendChatCompletion and SendStructuredCompletion.
// Returns the raw Anthropic message response for processing.
func (c *AnthropicClient) sendChatCompletionRequest(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (*anthropic.BetaMessage, error) {
	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic client: %w", err)
	}

	// Convert session messages to Anthropic format
	messages, additionalSystemInstructions := c.convertMessagesToAnthropic(session)
	logger.Debug("Messages converted", "message_count", len(messages))

	// Build message parameters - use BetaMessageNewParams for thinking support
	params := anthropic.BetaMessageNewParams{
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
		params.System = []anthropic.BetaTextBlockParam{
			{Text: systemPrompt},
		}
		logger.Debug("System prompt added", "system_prompt", systemPrompt)
	}

	// Apply other model parameters
	c.applyModelParameters(&params, modelConfig)

	// Send request using beta API for thinking support
	logger.Debug("Sending Anthropic beta request", "model", modelConfig.BaseModel)
	message, err := c.client.Beta.Messages.New(context.Background(), params)
	if err != nil {
		logger.Error("Anthropic request failed", "error", err)
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}

	return message, nil
}

// SendStructuredCompletion sends a chat completion request to Anthropic and returns structured response.
// This method reuses SendChatCompletion logic and post-processes the response to separate thinking blocks.
func (c *AnthropicClient) SendStructuredCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (*neurotypes.StructuredLLMResponse, error) {
	logger.Debug("Anthropic SendStructuredCompletion starting", "model", modelConfig.BaseModel)

	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize Anthropic client: %w", err)
	}

	// Reuse the core logic from SendChatCompletion but return raw response blocks for structured processing
	message, err := c.sendChatCompletionRequest(session, modelConfig)
	if err != nil {
		return nil, err
	}

	// Extract response content and thinking blocks separately (structured processing)
	if len(message.Content) == 0 {
		logger.Error("No response content returned")
		return nil, fmt.Errorf("no response content returned")
	}

	// Process all content blocks and extract thinking blocks separately
	textContent, thinkingBlocks := c.processResponseBlocksStructured(message.Content)

	if textContent == "" {
		logger.Error("Empty response content")
		return nil, fmt.Errorf("empty response content")
	}

	// Create structured response
	structuredResponse := &neurotypes.StructuredLLMResponse{
		TextContent:    textContent,
		ThinkingBlocks: thinkingBlocks,
	}

	logger.Debug("Anthropic structured response received", "content_length", len(textContent), "thinking_blocks", len(thinkingBlocks))
	return structuredResponse, nil
}

// SetDebugTransport sets the HTTP transport for network debugging.
func (c *AnthropicClient) SetDebugTransport(transport http.RoundTripper) {
	c.debugTransport = transport
	// Clear the existing client to force re-initialization with debug transport
	c.client = nil
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
func (c *AnthropicClient) convertMessagesToAnthropic(session *neurotypes.ChatSession) ([]anthropic.BetaMessageParam, string) {
	messages := make([]anthropic.BetaMessageParam, 0, len(session.Messages))
	var additionalSystemInstructions []string

	for _, msg := range session.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(msg.Content)))
		case "assistant":
			messages = append(messages, anthropic.BetaMessageParam{
				Role: anthropic.BetaMessageParamRoleAssistant,
				Content: []anthropic.BetaContentBlockParamUnion{
					anthropic.NewBetaTextBlock(msg.Content),
				},
			})
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
func (c *AnthropicClient) applyModelParameters(params *anthropic.BetaMessageNewParams, modelConfig *neurotypes.ModelConfig) {
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

	// Apply extended thinking parameters
	c.applyThinkingParameters(params, modelConfig)
}

// applyThinkingParameters applies extended thinking configuration to the Anthropic request.
func (c *AnthropicClient) applyThinkingParameters(params *anthropic.BetaMessageNewParams, modelConfig *neurotypes.ModelConfig) {
	logger.Debug("Applying thinking parameters", "has_parameters", modelConfig.Parameters != nil)
	if modelConfig.Parameters == nil {
		logger.Debug("No parameters found, skipping thinking configuration")
		return
	}

	// Check for thinking_budget parameter
	if budgetRaw, ok := modelConfig.Parameters["thinking_budget"]; ok {
		logger.Debug("Found thinking_budget parameter", "raw_value", budgetRaw, "type", fmt.Sprintf("%T", budgetRaw))
		var budget int64
		switch v := budgetRaw.(type) {
		case int:
			budget = int64(v)
			logger.Debug("Converted int to int64", "value", budget)
		case int64:
			budget = v
			logger.Debug("Using int64 value", "value", budget)
		case float64:
			budget = int64(v)
			logger.Debug("Converted float64 to int64", "value", budget)
		case string:
			// Handle string conversion for cases like thinking_budget="8192"
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				budget = parsed
				logger.Debug("Converted string to int64", "value", budget)
			} else {
				logger.Debug("Failed to parse string thinking_budget", "value", v, "error", err)
				return
			}
		default:
			logger.Debug("Invalid thinking_budget type, ignoring", "type", fmt.Sprintf("%T", v), "value", v)
			return
		}

		// Enable thinking if budget > 0
		if budget > 0 {
			params.Thinking = anthropic.BetaThinkingConfigParamUnion{
				OfEnabled: &anthropic.BetaThinkingConfigEnabledParam{
					BudgetTokens: budget,
				},
			}
			logger.Debug("Extended thinking enabled successfully", "budget_tokens", budget, "model", modelConfig.BaseModel)
		} else {
			logger.Debug("Thinking budget is 0 or negative, not enabling thinking", "budget", budget)
		}
	} else {
		logger.Debug("No thinking_budget parameter found in model config")
		if modelConfig.Parameters != nil {
			logger.Debug("Available parameters", "params", modelConfig.Parameters)
		}
	}
}

// processResponseBlocks processes all content blocks from Anthropic beta response.
// Handles text, thinking, and redacted_thinking blocks appropriately.
func (c *AnthropicClient) processResponseBlocks(blocks []anthropic.BetaContentBlockUnion) (string, ThinkingInfo) {
	var content string
	info := ThinkingInfo{}

	for _, block := range blocks {
		// Handle text blocks
		textBlock := block.AsText()
		if textBlock.Type == "text" {
			content += textBlock.Text
			logger.Debug("Text block processed", "text_length", len(textBlock.Text))
			continue
		}

		// Handle thinking blocks
		thinkingBlock := block.AsThinking()
		if thinkingBlock.Type == "thinking" {
			info.ThinkingBlocks++
			logger.Debug("Thinking block processed", "thinking_length", len(thinkingBlock.Thinking))
			// Note: Thinking content is Claude's internal reasoning - not included in response
			continue
		}

		// Handle redacted thinking blocks
		redactedBlock := block.AsRedactedThinking()
		if redactedBlock.Type == "redacted_thinking" {
			info.RedactedBlocks++
			logger.Debug("Redacted thinking block processed", "data_length", len(redactedBlock.Data))
			// Note: Redacted blocks contain encrypted content - not included in response
			continue
		}

		// Unknown block type - log warning
		logger.Debug("Unknown content block type encountered")
	}

	return content, info
}

// processResponseBlocksStructured processes all content blocks from Anthropic beta response for structured output.
// Separates text content from thinking blocks instead of discarding thinking content.
func (c *AnthropicClient) processResponseBlocksStructured(blocks []anthropic.BetaContentBlockUnion) (string, []neurotypes.ThinkingBlock) {
	var textContent string
	var thinkingBlocks []neurotypes.ThinkingBlock

	for _, block := range blocks {
		// Handle text blocks
		textBlock := block.AsText()
		if textBlock.Type == "text" {
			textContent += textBlock.Text
			logger.Debug("Text block processed for structured response", "text_length", len(textBlock.Text))
			continue
		}

		// Handle thinking blocks - extract instead of discarding
		thinkingBlock := block.AsThinking()
		if thinkingBlock.Type == "thinking" {
			thinkingBlocks = append(thinkingBlocks, neurotypes.ThinkingBlock{
				Content:  thinkingBlock.Thinking,
				Provider: "anthropic",
				Type:     "thinking",
			})
			logger.Debug("Thinking block extracted for structured response", "thinking_length", len(thinkingBlock.Thinking))
			continue
		}

		// Handle redacted thinking blocks - extract metadata
		redactedBlock := block.AsRedactedThinking()
		if redactedBlock.Type == "redacted_thinking" {
			// For redacted thinking, we can't show the actual content but we can indicate it exists
			thinkingBlocks = append(thinkingBlocks, neurotypes.ThinkingBlock{
				Content:  "[Thinking content redacted by Anthropic]",
				Provider: "anthropic",
				Type:     "redacted_thinking",
			})
			logger.Debug("Redacted thinking block extracted for structured response", "data_length", len(redactedBlock.Data))
			continue
		}

		// Unknown block type - log warning
		logger.Debug("Unknown content block type encountered in structured response")
	}

	return textContent, thinkingBlocks
}
