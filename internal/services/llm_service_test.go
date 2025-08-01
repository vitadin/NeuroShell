package services

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

// Test the real LLMService (basic functionality without OpenAI calls)
func TestLLMService_Name(t *testing.T) {
	service := NewLLMService()
	assert.Equal(t, "llm", service.Name())
}

func TestLLMService_Initialize(t *testing.T) {
	service := NewLLMService()

	// The new LLMService no longer depends on API keys for initialization
	err := service.Initialize()
	require.NoError(t, err)
}

// MockLLMClient for testing
type MockLLMClient struct {
	configured bool
	response   string
}

func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		configured: true,
		response:   "This is a mock LLM response.",
	}
}

func (m *MockLLMClient) SendChatCompletion(_ *neurotypes.ChatSession, _ *neurotypes.ModelConfig) (string, error) {
	return m.response, nil
}

func (m *MockLLMClient) SetDebugTransport(_ http.RoundTripper) {
	// Dummy implementation for mock client
}

func (m *MockLLMClient) StreamChatCompletion(_ *neurotypes.ChatSession, _ *neurotypes.ModelConfig) (<-chan neurotypes.StreamChunk, error) {
	ch := make(chan neurotypes.StreamChunk, 1)
	ch <- neurotypes.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func (m *MockLLMClient) GetProviderName() string {
	return "mock"
}

func (m *MockLLMClient) IsConfigured() bool {
	return m.configured
}

// Test LLMService with mock client
func TestLLMService_SendCompletion(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	client := NewMockLLMClient()

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	response, err := service.SendCompletion(client, session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "This is a mock LLM response.", response)
}

func TestLLMService_SendCompletion_DifferentModels(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	client := NewMockLLMClient()

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	tests := []struct {
		model    string
		expected string
	}{
		{"gpt-4", "This is a mock LLM response."},
		{"gpt-3.5-turbo", "This is a mock LLM response."},
		{"unknown-model", "This is a mock LLM response."},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			modelConfig := &neurotypes.ModelConfig{
				BaseModel: tt.model,
				Provider:  "openai",
			}

			response, err := service.SendCompletion(client, session, modelConfig)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestLLMService_SendCompletion_CustomResponse(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create mock client with custom response
	client := &MockLLMClient{
		configured: true,
		response:   "Custom response for testing",
	}

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "custom-model",
		Provider:  "openai",
	}

	response, err := service.SendCompletion(client, session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "Custom response for testing", response)
}

func TestLLMService_SendCompletion_WithActualMessages(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	client := NewMockLLMClient()

	// Test with existing messages in session
	session := &neurotypes.ChatSession{
		ID:   "test-session",
		Name: "test",
		Messages: []neurotypes.Message{
			{
				ID:      "msg1",
				Role:    "user",
				Content: "Previous message",
			},
		},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	response, err := service.SendCompletion(client, session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "This is a mock LLM response.", response)
}

func TestLLMService_SendCompletion_NotInitialized(t *testing.T) {
	service := NewLLMService()
	// Don't initialize the service

	client := NewMockLLMClient()
	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	_, err := service.SendCompletion(client, session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm service not initialized")
}

func TestLLMService_SendCompletion_NilClient(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	_, err = service.SendCompletion(nil, session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm client cannot be nil")
}

func TestLLMService_SendCompletion_UnconfiguredClient(t *testing.T) {
	service := NewLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create unconfigured client
	client := &MockLLMClient{
		configured: false,
		response:   "Should not see this",
	}

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	_, err = service.SendCompletion(client, session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "llm client is not configured")
}
