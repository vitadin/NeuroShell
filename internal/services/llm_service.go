package services

import (
	"fmt"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// StreamChunk is an alias for neurotypes.StreamChunk for backward compatibility.
type StreamChunk = neurotypes.StreamChunk

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

// StreamCompletion sends a streaming chat completion request using the provided client.
// The session is sent as-is with no message manipulation - this is the caller's responsibility.
func (s *LLMService) StreamCompletion(client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	logger.ServiceOperation("llm", "stream_completion", "starting")

	if !s.initialized {
		logger.Error("LLM service not initialized")
		return nil, fmt.Errorf("llm service not initialized")
	}

	if client == nil {
		logger.Error("LLM client is nil")
		return nil, fmt.Errorf("llm client cannot be nil")
	}

	if !client.IsConfigured() {
		logger.Error("LLM client is not configured")
		return nil, fmt.Errorf("llm client is not configured")
	}

	logger.Debug("Sending streaming completion request", "provider", client.GetProviderName(), "model", model.BaseModel, "messages", len(session.Messages))

	// Send the streaming completion request using the client with session as-is
	responseChan, err := client.StreamChatCompletion(session, model)
	if err != nil {
		logger.Error("Streaming completion request failed", "error", err)
		return nil, fmt.Errorf("streaming completion request failed: %w", err)
	}

	logger.Debug("Streaming completion request started")
	logger.ServiceOperation("llm", "stream_completion", "completed")
	return responseChan, nil
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

// StreamCompletion mocks streaming completion (returns a channel with mock response)
func (m *MockLLMService) StreamCompletion(_ neurotypes.LLMClient, session *neurotypes.ChatSession, _ *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	if !m.initialized {
		return nil, fmt.Errorf("mock llm service not initialized")
	}

	// Create a channel for streaming response
	responseChan := make(chan neurotypes.StreamChunk, 1)

	// Create a mock response with message count and last message info for debugging
	messageCount := len(session.Messages)
	lastMessage := "no messages"
	if messageCount > 0 {
		lastMessage = session.Messages[messageCount-1].Content
	}

	response := fmt.Sprintf("This is a mocking reply (received %d messages, last: %s)", messageCount, lastMessage)

	// Send the response through the channel and close it
	go func() {
		defer close(responseChan)
		responseChan <- neurotypes.StreamChunk{
			Content: response,
			Done:    true,
			Error:   nil,
		}
	}()

	return responseChan, nil
}

// SetMockResponse sets a mock response for a specific model
func (m *MockLLMService) SetMockResponse(model, response string) {
	m.responses[model] = response
}
