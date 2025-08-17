package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"

	"google.golang.org/genai"
)

// GeminiThinkingInfo contains information about thinking blocks found in Gemini responses.
type GeminiThinkingInfo struct {
	ThinkingBlocks int // Number of thinking blocks processed
	TextBlocks     int // Number of text blocks processed
	ThinkingTokens int // Estimated thinking tokens used (if available)
}

// GeminiClient implements the LLMClient interface for Google Gemini API.
// It provides lazy initialization of the Gemini client and handles
// all Gemini-specific communication logic.
type GeminiClient struct {
	apiKey         string
	client         *genai.Client
	debugTransport http.RoundTripper
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

	// Create Gemini client with API key configuration
	ctx := context.Background()
	clientConfig := &genai.ClientConfig{
		APIKey: c.apiKey,
	}

	// Add debug transport if available
	if c.debugTransport != nil {
		httpClient := &http.Client{Transport: c.debugTransport}
		clientConfig.HTTPClient = httpClient
		logger.Debug("Gemini client initialized with debug transport", "provider", "gemini")
	} else {
		logger.Debug("Gemini client initialized", "provider", "gemini")
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	c.client = client
	return nil
}

// SendChatCompletion sends a chat completion request to Google Gemini.
func (c *GeminiClient) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	logger.Debug("Gemini SendChatCompletion starting", "model", modelConfig.BaseModel)

	// Use shared request logic to get raw response
	result, err := c.sendChatCompletionRequest(session, modelConfig)
	if err != nil {
		return "", err
	}

	// Process response with formatted thinking (traditional processing)
	content, thinkingInfo := c.processGeminiResponse(result)
	// Content can be empty if response contains only thinking blocks (which are now skipped)
	if content == "" {
		logger.Debug("Gemini response contains no text content (may have thinking blocks only)")
	}

	logger.Debug("Gemini response received", "content_length", len(content), "thinking_blocks", thinkingInfo.ThinkingBlocks, "text_blocks", thinkingInfo.TextBlocks)
	return content, nil
}

// sendChatCompletionRequest handles the core request logic shared by both SendChatCompletion and SendStructuredCompletion.
// Returns the raw Gemini response for processing.
func (c *GeminiClient) sendChatCompletionRequest(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (*genai.GenerateContentResponse, error) {
	// Initialize client if needed
	if err := c.initializeClientIfNeeded(); err != nil {
		return nil, fmt.Errorf("failed to initialize Gemini client: %w", err)
	}

	// Convert session messages to Gemini format
	contents := c.convertMessagesToGemini(session)
	logger.Debug("Messages converted", "content_count", len(contents))

	// Build generation config from model parameters and session
	config := c.buildGenerationConfig(modelConfig, session)

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
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}

	return result, nil
}

// SendStructuredCompletion sends a chat completion request to Google Gemini and returns structured response.
// This method reuses SendChatCompletion logic and post-processes the response to separate thinking blocks.
// All errors are encoded in the StructuredLLMResponse.Error field - no Go errors are returned.
func (c *GeminiClient) SendStructuredCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) *neurotypes.StructuredLLMResponse {
	logger.Debug("Gemini SendStructuredCompletion starting", "model", modelConfig.BaseModel)

	// Use shared request logic to get raw response
	result, err := c.sendChatCompletionRequest(session, modelConfig)
	if err != nil {
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "api_request_failed",
				Message: err.Error(),
				Type:    "api_error",
			},
			Metadata: map[string]interface{}{"provider": "gemini", "model": modelConfig.BaseModel},
		}
	}

	// Process response with structured thinking block extraction
	textContent, thinkingBlocks := c.processGeminiResponseStructured(result)
	if textContent == "" && len(thinkingBlocks) == 0 {
		logger.Error("No content in Gemini structured response")
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "empty_response",
				Message: "no content in response",
				Type:    "response_error",
			},
			Metadata: map[string]interface{}{"provider": "gemini", "model": modelConfig.BaseModel},
		}
	}

	// Create structured response
	structuredResponse := &neurotypes.StructuredLLMResponse{
		TextContent:    textContent,
		ThinkingBlocks: thinkingBlocks,
		Error:          nil, // No error in successful case
		Metadata:       map[string]interface{}{"provider": "gemini", "model": modelConfig.BaseModel},
	}

	logger.Debug("Gemini structured response received", "content_length", len(textContent), "thinking_blocks", len(thinkingBlocks))
	return structuredResponse
}

// SetDebugTransport sets the HTTP transport for network debugging.
func (c *GeminiClient) SetDebugTransport(transport http.RoundTripper) {
	c.debugTransport = transport
	// Clear the existing client to force re-initialization with debug transport
	c.client = nil
}

// convertMessagesToGemini converts NeuroShell messages to Gemini format.
// Returns the conversation as a slice of genai.Content.
// System prompt is handled separately via SystemInstruction in GenerateContentConfig.
func (c *GeminiClient) convertMessagesToGemini(session *neurotypes.ChatSession) []*genai.Content {
	contents := make([]*genai.Content, 0)

	// Convert conversation messages with proper role mapping
	for _, msg := range session.Messages {
		var role string
		var content string

		switch msg.Role {
		case "user":
			role = "user"
			content = msg.Content
		case "assistant":
			role = "model" // Gemini uses "model" instead of "assistant"
			content = msg.Content
		case "system":
			// System messages are treated as user messages in Gemini
			role = "user"
			content = "System: " + msg.Content
		default:
			// Skip unknown roles
			continue
		}

		msgContent := &genai.Content{
			Parts: []*genai.Part{{Text: content}},
			Role:  role,
		}
		contents = append(contents, msgContent)
	}

	// If no contents were added, add a default empty user content
	if len(contents) == 0 {
		emptyContent := &genai.Content{
			Parts: []*genai.Part{{Text: ""}},
			Role:  "user",
		}
		contents = append(contents, emptyContent)
	}

	return contents
}

// buildGenerationConfig creates a Gemini generation config from NeuroShell model parameters and session.
func (c *GeminiClient) buildGenerationConfig(modelConfig *neurotypes.ModelConfig, session *neurotypes.ChatSession) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	// Add system instruction if present
	if session.SystemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(session.SystemPrompt, genai.RoleUser)
	}

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

	// Apply thinking_budget for Gemini models
	if thinkingBudget, ok := modelConfig.Parameters["thinking_budget"]; ok {
		if thinkingBudgetInt, ok := thinkingBudget.(int); ok {
			// Create ThinkingConfig based on thinking_budget value
			switch {
			case thinkingBudgetInt == -1:
				// Dynamic thinking: let the model decide
				config.ThinkingConfig = &genai.ThinkingConfig{
					IncludeThoughts: true, // Enable thought summaries
				}
			case thinkingBudgetInt == 0:
				// Thinking disabled: set budget to 0
				thinkingBudgetInt32 := int32(0)
				config.ThinkingConfig = &genai.ThinkingConfig{
					ThinkingBudget:  &thinkingBudgetInt32,
					IncludeThoughts: false, // No thoughts when disabled
				}
			case thinkingBudgetInt > 0:
				// Fixed thinking budget
				thinkingBudgetInt32 := int32(thinkingBudgetInt)
				config.ThinkingConfig = &genai.ThinkingConfig{
					ThinkingBudget:  &thinkingBudgetInt32,
					IncludeThoughts: true, // Enable thought summaries
				}
			}
		}
	}

	return config
}

// processGeminiResponse processes all content from Gemini response including thinking blocks.
// Handles both thinking and text parts appropriately, displaying thinking content visibly.
func (c *GeminiClient) processGeminiResponse(result *genai.GenerateContentResponse) (string, GeminiThinkingInfo) {
	var contentBuilder strings.Builder
	info := GeminiThinkingInfo{}

	// Process all candidates (usually just one)
	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}

		// Process all parts in the content
		for _, part := range candidate.Content.Parts {
			if part.Text == "" {
				continue // Skip empty parts
			}

			if part.Thought {
				// This is a thinking block - skip in regular response (will be handled by structured response)
				info.ThinkingBlocks++
				logger.Debug("Gemini thinking block skipped in regular response", "thinking_length", len(part.Text))
			} else {
				// This is regular text content - no formatting
				info.TextBlocks++
				contentBuilder.WriteString(part.Text)
				logger.Debug("Gemini text block processed", "text_length", len(part.Text))
			}
		}
	}

	return contentBuilder.String(), info
}

// processGeminiResponseStructured processes all content from Gemini response for structured output.
// Separates thinking blocks from text content instead of mixing them with formatting.
func (c *GeminiClient) processGeminiResponseStructured(result *genai.GenerateContentResponse) (string, []neurotypes.ThinkingBlock) {
	var textContent strings.Builder
	var thinkingBlocks []neurotypes.ThinkingBlock

	// Process all candidates (usually just one)
	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}

		// Process all parts in the content
		for _, part := range candidate.Content.Parts {
			if part.Text == "" {
				continue // Skip empty parts
			}

			if part.Thought {
				// This is a thinking block - extract separately
				thinkingBlocks = append(thinkingBlocks, neurotypes.ThinkingBlock{
					Content:  part.Text,
					Provider: "gemini",
					Type:     "thinking",
				})
				logger.Debug("Gemini thinking block extracted for structured response", "thinking_length", len(part.Text))
			} else {
				// This is regular text content - no formatting
				textContent.WriteString(part.Text)
				logger.Debug("Gemini text block processed for structured response", "text_length", len(part.Text))
			}
		}
	}

	return textContent.String(), thinkingBlocks
}
