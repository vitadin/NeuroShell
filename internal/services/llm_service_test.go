package services

import (
	"os"
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

func TestLLMService_Initialize_WithoutAPIKey(t *testing.T) {
	service := NewLLMService()

	// Temporarily unset OPENAI_API_KEY
	originalKey := os.Getenv("OPENAI_API_KEY")
	_ = os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			_ = os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	// This should fail because OPENAI_API_KEY is not set
	err := service.Initialize()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY environment variable not set")
}

// Test MockLLMService
func TestMockLLMService_Basic(t *testing.T) {
	service := NewMockLLMService()

	// Test name
	assert.Equal(t, "llm", service.Name())

	// Test initialization
	err := service.Initialize()
	require.NoError(t, err)

	// Test chat completion
	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	response, err := service.SendChatCompletion(session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "Hello! This is a mock GPT-4 response.", response)
}

func TestMockLLMService_DifferentModels(t *testing.T) {
	service := NewMockLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	tests := []struct {
		model    string
		expected string
	}{
		{"gpt-4", "Hello! This is a mock GPT-4 response."},
		{"gpt-3.5-turbo", "Hi! This is a mock GPT-3.5 response."},
		{"unknown-model", "This is a mock LLM response."},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			modelConfig := &neurotypes.ModelConfig{
				BaseModel: tt.model,
				Provider:  "openai",
			}

			response, err := service.SendChatCompletion(session, modelConfig)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestMockLLMService_CustomResponse(t *testing.T) {
	service := NewMockLLMService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set custom response
	service.SetMockResponse("custom-model", "Custom response for testing")

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "custom-model",
		Provider:  "openai",
	}

	response, err := service.SendChatCompletion(session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "Custom response for testing", response)
}

func TestMockLLMService_NotInitialized(t *testing.T) {
	service := NewMockLLMService()

	session := &neurotypes.ChatSession{
		ID:       "test-session",
		Name:     "test",
		Messages: []neurotypes.Message{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Provider:  "openai",
	}

	_, err := service.SendChatCompletion(session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock llm service not initialized")
}
