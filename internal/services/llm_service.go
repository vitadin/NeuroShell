package services

import (
	"fmt"
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// LLMService implements the neurotypes.LLMService interface.
// It provides pure business logic for LLM operations without external dependencies.
type LLMService struct {
	initialized bool
}

// NewLLMService creates a new LLMService instance.
func NewLLMService() *LLMService {
	return &LLMService{
		initialized: false,
	}
}

// Name returns the service name "llm" for registration.
func (s *LLMService) Name() string {
	return "llm"
}

// Initialize sets up the LLMService for operation.
// No longer depends on API keys or external configuration.
func (s *LLMService) Initialize() error {
	logger.ServiceOperation("llm", "initialize", "starting")
	s.initialized = true
	logger.ServiceOperation("llm", "initialize", "completed")
	return nil
}

// SendCompletion sends a chat completion request using the provided client.
// The session is sent as-is with no message manipulation - this is the caller's responsibility.
func (s *LLMService) SendCompletion(client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig) (string, error) {
	logger.ServiceOperation("llm", "send_completion", "starting")

	if !s.initialized {
		logger.Error("LLM service not initialized")
		return "", fmt.Errorf("llm service not initialized")
	}

	if client == nil {
		logger.Error("LLM client is nil")
		return "", fmt.Errorf("llm client cannot be nil")
	}

	if !client.IsConfigured() {
		logger.Error("LLM client is not configured")
		return "", fmt.Errorf("llm client is not configured")
	}

	logger.Debug("Sending completion request", "provider", client.GetProviderName(), "model", model.BaseModel, "messages", len(session.Messages))

	// Send the completion request using the client with session as-is
	response, err := client.SendChatCompletion(session, model)
	if err != nil {
		logger.Error("Completion request failed", "error", err)
		return "", fmt.Errorf("completion request failed: %w", err)
	}

	logger.Debug("Completion request completed", "response_length", len(response))
	logger.ServiceOperation("llm", "send_completion", "completed")
	return response, nil
}

// SendStructuredCompletion sends a chat completion request using the provided client and returns structured response.
// This separates thinking/reasoning content from regular text for proper rendering control.
// All errors are encoded in the StructuredLLMResponse.Error field - no Go errors are returned.
func (s *LLMService) SendStructuredCompletion(client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig) *neurotypes.StructuredLLMResponse {
	logger.ServiceOperation("llm", "send_structured_completion", "starting")

	if !s.initialized {
		logger.Error("LLM service not initialized")
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "service_not_initialized",
				Message: "llm service not initialized",
				Type:    "service_error",
			},
			Metadata: map[string]interface{}{"service": "llm"},
		}
	}

	if client == nil {
		logger.Error("LLM client is nil")
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "client_nil",
				Message: "llm client cannot be nil",
				Type:    "client_error",
			},
			Metadata: map[string]interface{}{"service": "llm"},
		}
	}

	if !client.IsConfigured() {
		logger.Error("LLM client is not configured")
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "client_not_configured",
				Message: "llm client is not configured",
				Type:    "client_error",
			},
			Metadata: map[string]interface{}{"service": "llm", "provider": client.GetProviderName()},
		}
	}

	logger.Debug("Sending structured completion request", "provider", client.GetProviderName(), "model", model.BaseModel, "messages", len(session.Messages))

	// Send the structured completion request using the client with session as-is
	response := client.SendStructuredCompletion(session, model)

	logger.Debug("Structured completion request completed", "text_length", len(response.TextContent), "thinking_blocks", len(response.ThinkingBlocks))
	logger.ServiceOperation("llm", "send_structured_completion", "completed")
	return response
}

// MockLLMService provides a mock implementation of LLMService for testing
type MockLLMService struct {
	initialized bool
	responses   map[string]string // model -> response mapping
}

// NewMockLLMService creates a new MockLLMService instance
func NewMockLLMService() *MockLLMService {
	return &MockLLMService{
		initialized: false,
		responses: map[string]string{
			"gpt-4":         "Hello! This is a mock GPT-4 response.",
			"gpt-3.5-turbo": "Hi! This is a mock GPT-3.5 response.",
			"default":       "This is a mock LLM response.",
		},
	}
}

// Name returns the service name "llm" for registration
func (m *MockLLMService) Name() string {
	return "llm"
}

// Initialize sets up the MockLLMService for operation
func (m *MockLLMService) Initialize() error {
	m.initialized = true
	return nil
}

// SendCompletion mocks sending a completion request
func (m *MockLLMService) SendCompletion(_ neurotypes.LLMClient, session *neurotypes.ChatSession, _ *neurotypes.ModelConfig) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("mock llm service not initialized")
	}

	// Create a mock response with message count and last message info for debugging
	messageCount := len(session.Messages)
	lastMessage := "no messages"
	if messageCount > 0 {
		lastMessage = session.Messages[messageCount-1].Content
	}

	return fmt.Sprintf("This is a mocking reply (received %d messages, last: %s)", messageCount, lastMessage), nil
}

// SendCompletionWithDebug mocks sending a completion request with debug info
func (m *MockLLMService) SendCompletionWithDebug(_ neurotypes.LLMClient, session *neurotypes.ChatSession, _ *neurotypes.ModelConfig, debugNetwork bool) (string, string, error) {
	if !m.initialized {
		return "", "", fmt.Errorf("mock llm service not initialized")
	}

	// Create a mock response with message count and last message info for debugging
	messageCount := len(session.Messages)
	lastMessage := "no messages"
	if messageCount > 0 {
		lastMessage = session.Messages[messageCount-1].Content
	}

	response := fmt.Sprintf("This is a mocking reply (received %d messages, last: %s)", messageCount, lastMessage)
	debugInfo := ""
	if debugNetwork {
		debugInfo = "{\"info\": \"mock debug data\", \"request\": {\"messages\": " + fmt.Sprintf("%d", messageCount) + "}, \"response\": {\"content\": \"mock response\"}}"
	}

	return response, debugInfo, nil
}

// SendStructuredCompletion mocks sending a structured completion request
// All errors are encoded in the StructuredLLMResponse.Error field - no Go errors are returned.
func (m *MockLLMService) SendStructuredCompletion(_ neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig) *neurotypes.StructuredLLMResponse {
	if !m.initialized {
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "service_not_initialized",
				Message: "mock llm service not initialized",
				Type:    "service_error",
			},
			Metadata: map[string]interface{}{"service": "mock_llm"},
		}
	}

	// Create a mock response with message count and last message info for debugging
	messageCount := len(session.Messages)
	lastMessage := "no messages"
	if messageCount > 0 {
		lastMessage = session.Messages[messageCount-1].Content
	}

	// Check for error trigger phrases in the last message for testing error scenarios
	lastMessageLower := strings.ToLower(lastMessage)

	// Trigger API error
	if strings.Contains(lastMessageLower, "trigger api error") {
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "api_request_failed",
				Message: "Mock API request failed for testing purposes",
				Type:    "api_error",
			},
			Metadata: map[string]interface{}{"service": "mock_llm", "error_triggered": true},
		}
	}

	// Trigger rate limit error with partial content
	if strings.Contains(lastMessageLower, "trigger rate limit") {
		thinkingContent := "I was thinking about this request, but then encountered a rate limit..."
		return &neurotypes.StructuredLLMResponse{
			TextContent: "This is a partial response before hitting rate limit",
			ThinkingBlocks: []neurotypes.ThinkingBlock{
				{
					Content:  thinkingContent,
					Provider: "mock",
					Type:     "thinking",
				},
			},
			Error: &neurotypes.LLMError{
				Code:    "rate_limit_exceeded",
				Message: "Mock rate limit exceeded - please try again later",
				Type:    "api_error",
			},
			Metadata: map[string]interface{}{"service": "mock_llm", "error_triggered": true},
		}
	}

	// Trigger client error
	if strings.Contains(lastMessageLower, "trigger client error") {
		return &neurotypes.StructuredLLMResponse{
			TextContent:    "",
			ThinkingBlocks: []neurotypes.ThinkingBlock{},
			Error: &neurotypes.LLMError{
				Code:    "client_configuration_error",
				Message: "Mock client configuration error for testing",
				Type:    "client_error",
			},
			Metadata: map[string]interface{}{"service": "mock_llm", "error_triggered": true},
		}
	}

	textContent := fmt.Sprintf("This is a mocking reply (received %d messages, last: %s)", messageCount, lastMessage)

	// Determine provider from model configuration for provider-specific testing
	provider := "mock"
	if model != nil {
		// Extract provider from model description or catalog ID for testing
		desc := strings.ToLower(model.Description)
		catalogID := strings.ToLower(model.CatalogID)

		switch {
		case strings.Contains(desc, "anthropic") || strings.Contains(catalogID, "claude"):
			provider = "anthropic"
		case strings.Contains(desc, "gemini") || strings.Contains(catalogID, "gemini"):
			provider = "gemini"
		case strings.Contains(desc, "openai") || strings.Contains(catalogID, "gpt"):
			provider = "openai"
		}
	}

	// Create content-aware thinking blocks that reflect the actual message content
	thinkingContent := fmt.Sprintf("Thinking about the user's message: \"%s\". This helps verify the message flow in tests. The user sent %d messages total, and I need to provide a helpful response.", lastMessage, messageCount)

	// Create mock thinking blocks for testing
	thinkingBlocks := []neurotypes.ThinkingBlock{
		{
			Content:  thinkingContent,
			Provider: provider,
			Type:     "thinking",
		},
	}

	structuredResponse := &neurotypes.StructuredLLMResponse{
		TextContent:    textContent,
		ThinkingBlocks: thinkingBlocks,
		Error:          nil, // No error in successful case
		Metadata:       map[string]interface{}{"service": "mock_llm", "provider": provider, "model": model.BaseModel},
	}

	return structuredResponse
}

// SetMockResponse sets a mock response for a specific model
func (m *MockLLMService) SetMockResponse(model, response string) {
	m.responses[model] = response
}
