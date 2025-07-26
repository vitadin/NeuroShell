package services

import (
	"testing"

	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected struct {
			apiKey string
			client interface{}
		}
	}{
		{
			name:   "with API key",
			apiKey: "test-api-key",
			expected: struct {
				apiKey string
				client interface{}
			}{
				apiKey: "test-api-key",
				client: (*openai.Client)(nil), // Should be nil due to lazy initialization
			},
		},
		{
			name:   "with empty API key",
			apiKey: "",
			expected: struct {
				apiKey string
				client interface{}
			}{
				apiKey: "",
				client: (*openai.Client)(nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey)

			assert.Equal(t, tt.expected.apiKey, client.apiKey)
			assert.Equal(t, tt.expected.client, client.client)
		})
	}
}

func TestOpenAIClient_GetProviderName(t *testing.T) {
	client := NewOpenAIClient("test-api-key")
	assert.Equal(t, "openai", client.GetProviderName())
}

func TestOpenAIClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "configured with API key",
			apiKey:   "test-api-key",
			expected: true,
		},
		{
			name:     "not configured - empty API key",
			apiKey:   "",
			expected: false,
		},
		{
			name:     "not configured - whitespace API key",
			apiKey:   "   ",
			expected: true, // Non-empty string, even if just whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey)
			assert.Equal(t, tt.expected, client.IsConfigured())
		})
	}
}

func TestOpenAIClient_SendChatCompletion_NotConfigured(t *testing.T) {
	client := NewOpenAIClient("")

	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Test message"},
		},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
	}

	_, err := client.SendChatCompletion(session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAI API key not configured")
}

func TestOpenAIClient_StreamChatCompletion_NotConfigured(t *testing.T) {
	client := NewOpenAIClient("")

	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Test message"},
		},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
	}

	_, err := client.StreamChatCompletion(session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAI API key not configured")
}

func TestOpenAIClient_ConvertMessagesToOpenAI(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	tests := []struct {
		name                 string
		session              *neurotypes.ChatSession
		expectedMessageCount int
		expectedRoles        []string
		expectedContents     []string
	}{
		{
			name: "mixed message types",
			session: &neurotypes.ChatSession{
				Messages: []neurotypes.Message{
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
					{Role: "system", Content: "Be helpful"},
					{Role: "user", Content: "How are you?"},
					{Role: "unknown", Content: "This should be skipped"},
				},
			},
			expectedMessageCount: 4, // user, assistant, system, user (unknown skipped)
			expectedRoles:        []string{"user", "assistant", "system", "user"},
			expectedContents:     []string{"Hello", "Hi there!", "Be helpful", "How are you?"},
		},
		{
			name: "no system messages",
			session: &neurotypes.ChatSession{
				Messages: []neurotypes.Message{
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
					{Role: "user", Content: "How are you?"},
				},
			},
			expectedMessageCount: 3,
			expectedRoles:        []string{"user", "assistant", "user"},
			expectedContents:     []string{"Hello", "Hi there!", "How are you?"},
		},
		{
			name: "only system messages",
			session: &neurotypes.ChatSession{
				Messages: []neurotypes.Message{
					{Role: "system", Content: "First instruction"},
					{Role: "system", Content: "Second instruction"},
				},
			},
			expectedMessageCount: 2,
			expectedRoles:        []string{"system", "system"},
			expectedContents:     []string{"First instruction", "Second instruction"},
		},
		{
			name: "empty session",
			session: &neurotypes.ChatSession{
				Messages: []neurotypes.Message{},
			},
			expectedMessageCount: 0,
			expectedRoles:        []string{},
			expectedContents:     []string{},
		},
		{
			name: "unknown roles filtered out",
			session: &neurotypes.ChatSession{
				Messages: []neurotypes.Message{
					{Role: "user", Content: "Valid message"},
					{Role: "invalid", Content: "Invalid role"},
					{Role: "", Content: "Empty role"},
					{Role: "function", Content: "Function role"},
					{Role: "assistant", Content: "Valid assistant"},
				},
			},
			expectedMessageCount: 2, // Only user and assistant should remain
			expectedRoles:        []string{"user", "assistant"},
			expectedContents:     []string{"Valid message", "Valid assistant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := client.convertMessagesToOpenAI(tt.session)

			assert.Len(t, messages, tt.expectedMessageCount)

			// We can't easily test the actual content of openai.ChatCompletionMessageParamUnion
			// without complex type assertions, but we can verify the length and structure
			assert.Len(t, messages, len(tt.expectedRoles))
		})
	}
}

func TestOpenAIClient_ApplyModelParameters(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	tests := []struct {
		name           string
		parameters     map[string]interface{}
		expectedParams map[string]interface{}
	}{
		{
			name: "all parameters with valid types",
			parameters: map[string]interface{}{
				"temperature":       0.7,
				"max_tokens":        100,
				"top_p":             0.9,
				"frequency_penalty": 0.1,
				"presence_penalty":  0.2,
			},
			expectedParams: map[string]interface{}{
				"temperature":       0.7,
				"max_tokens":        100,
				"top_p":             0.9,
				"frequency_penalty": 0.1,
				"presence_penalty":  0.2,
			},
		},
		{
			name:           "no parameters",
			parameters:     nil,
			expectedParams: map[string]interface{}{},
		},
		{
			name:           "empty parameters",
			parameters:     map[string]interface{}{},
			expectedParams: map[string]interface{}{},
		},
		{
			name: "invalid parameter types should be ignored",
			parameters: map[string]interface{}{
				"temperature":       "invalid", // Should be float64
				"max_tokens":        "invalid", // Should be int
				"top_p":             "invalid", // Should be float64
				"frequency_penalty": "invalid", // Should be float64
				"presence_penalty":  "invalid", // Should be float64
				"valid_temperature": 0.8,       // This should be ignored (not a standard param)
			},
			expectedParams: map[string]interface{}{}, // All invalid, so nothing applied
		},
		{
			name: "mixed valid and invalid parameters",
			parameters: map[string]interface{}{
				"temperature":       0.7,       // Valid
				"max_tokens":        "invalid", // Invalid
				"top_p":             0.9,       // Valid
				"frequency_penalty": "invalid", // Invalid
				"presence_penalty":  0.2,       // Valid
			},
			expectedParams: map[string]interface{}{
				"temperature":      0.7,
				"top_p":            0.9,
				"presence_penalty": 0.2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock ChatCompletionNewParams to test parameter application
			params := openai.ChatCompletionNewParams{
				Model:    openai.ChatModel("gpt-4"),
				Messages: []openai.ChatCompletionMessageParamUnion{},
			}

			modelConfig := &neurotypes.ModelConfig{
				BaseModel:  "gpt-4",
				Parameters: tt.parameters,
			}

			// This should not panic regardless of parameter types
			assert.NotPanics(t, func() {
				client.applyModelParameters(&params, modelConfig)
			})

			// Note: We can't easily test the actual parameter values without
			// exposing internal implementation details of the OpenAI SDK.
			// The main goal is to ensure the method doesn't panic and handles
			// type conversions gracefully.
		})
	}
}

func TestOpenAIClient_SystemPromptHandling(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	tests := []struct {
		name                string
		sessionSystemPrompt string
		messages            []neurotypes.Message
		expectSystemInStart bool // Whether we expect system message at the start
	}{
		{
			name:                "session with system prompt",
			sessionSystemPrompt: "You are a helpful assistant.",
			messages: []neurotypes.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			expectSystemInStart: true,
		},
		{
			name:                "session without system prompt",
			sessionSystemPrompt: "",
			messages: []neurotypes.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			expectSystemInStart: false,
		},
		{
			name:                "system prompt with system messages in conversation",
			sessionSystemPrompt: "You are a helpful assistant.",
			messages: []neurotypes.Message{
				{Role: "user", Content: "Hello"},
				{Role: "system", Content: "Be concise"},
				{Role: "assistant", Content: "Hi there!"},
			},
			expectSystemInStart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &neurotypes.ChatSession{
				SystemPrompt: tt.sessionSystemPrompt,
				Messages:     tt.messages,
			}

			messages := client.convertMessagesToOpenAI(session)

			// The convertMessagesToOpenAI method processes messages as-is
			// System prompt handling is done in SendChatCompletion
			expectedCount := 0
			for _, msg := range tt.messages {
				if msg.Role == "user" || msg.Role == "assistant" || msg.Role == "system" {
					expectedCount++
				}
			}

			assert.Len(t, messages, expectedCount)
		})
	}
}

func TestOpenAIClient_EmptyResponses(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	// Test with empty session
	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{},
	}

	messages := client.convertMessagesToOpenAI(session)
	assert.Empty(t, messages)

	// Test with nil messages
	session = &neurotypes.ChatSession{
		Messages: nil,
	}

	messages = client.convertMessagesToOpenAI(session)
	assert.Empty(t, messages)
}

func TestOpenAIClient_MessageRoleConversion(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "User message"},
			{Role: "assistant", Content: "Assistant message"},
			{Role: "system", Content: "System message"},
			{Role: "invalid", Content: "Invalid role message"},
			{Role: "", Content: "Empty role message"},
		},
	}

	messages := client.convertMessagesToOpenAI(session)

	// Should have 3 messages (user + assistant + system), invalid roles filtered out
	assert.Len(t, messages, 3)

	// We can't easily verify the content without complex type assertions on the OpenAI types
	// But we can verify the count, which indicates proper filtering
}

// TestOpenAIClient_InterfaceCompliance verifies that OpenAIClient implements LLMClient interface
func TestOpenAIClient_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.LLMClient = &OpenAIClient{}
}

func TestOpenAIClient_LazyInitialization(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	// Client should be nil initially (lazy initialization)
	assert.Nil(t, client.client)

	// After creation, it should still be nil until first use
	assert.True(t, client.IsConfigured())
	assert.Nil(t, client.client)
}

func TestOpenAIClient_InitializeClientIfNeeded(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "successful initialization with valid API key",
			apiKey:    "test-api-key",
			expectErr: false,
		},
		{
			name:      "failed initialization with empty API key",
			apiKey:    "",
			expectErr: true,
			errMsg:    "OpenAI API key not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey)

			// Access the method through a public method that calls it
			session := &neurotypes.ChatSession{
				Messages: []neurotypes.Message{
					{Role: "user", Content: "Test"},
				},
			}

			modelConfig := &neurotypes.ModelConfig{
				BaseModel: "gpt-4",
			}

			// Try to send a completion (this will trigger initializeClientIfNeeded)
			_, err := client.SendChatCompletion(session, modelConfig)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else if err != nil {
				// Even with a valid API key, we expect an error because we're not making real API calls
				// But the error should not be about missing API key
				assert.NotContains(t, err.Error(), "OpenAI API key not configured")
			}
		})
	}
}

func TestOpenAIClient_ModelParameterTypeHandling(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	// Test with various parameter types to ensure proper type checking
	testCases := []struct {
		name  string
		value interface{}
		valid bool
	}{
		// Temperature tests
		{"temperature float64", 0.7, true},
		{"temperature int", 1, false}, // Should be float64
		{"temperature string", "0.7", false},

		// Max tokens tests
		{"max_tokens int", 100, true},
		{"max_tokens float64", 100.0, false}, // Should be int
		{"max_tokens string", "100", false},

		// Top P tests
		{"top_p float64", 0.9, true},
		{"top_p int", 1, false}, // Should be float64
		{"top_p string", "0.9", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := openai.ChatCompletionNewParams{
				Model:    openai.ChatModel("gpt-4"),
				Messages: []openai.ChatCompletionMessageParamUnion{},
			}

			var paramMap map[string]interface{}
			switch tc.name {
			case "temperature float64", "temperature int", "temperature string":
				paramMap = map[string]interface{}{"temperature": tc.value}
			case "max_tokens int", "max_tokens float64", "max_tokens string":
				paramMap = map[string]interface{}{"max_tokens": tc.value}
			case "top_p float64", "top_p int", "top_p string":
				paramMap = map[string]interface{}{"top_p": tc.value}
			}

			modelConfig := &neurotypes.ModelConfig{
				BaseModel:  "gpt-4",
				Parameters: paramMap,
			}

			// Should not panic regardless of parameter type
			assert.NotPanics(t, func() {
				client.applyModelParameters(&params, modelConfig)
			})
		})
	}
}

func TestOpenAIClient_ParameterApplicationEdgeCases(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	// Test edge cases for parameter application
	tests := []struct {
		name       string
		parameters map[string]interface{}
	}{
		{
			name:       "nil parameters map",
			parameters: nil,
		},
		{
			name:       "empty parameters map",
			parameters: map[string]interface{}{},
		},
		{
			name: "parameters with nil values",
			parameters: map[string]interface{}{
				"temperature": nil,
				"max_tokens":  nil,
			},
		},
		{
			name: "parameters with zero values",
			parameters: map[string]interface{}{
				"temperature": 0.0,
				"max_tokens":  0,
				"top_p":       0.0,
			},
		},
		{
			name: "parameters with boundary values",
			parameters: map[string]interface{}{
				"temperature": 2.0,  // Max for OpenAI
				"max_tokens":  4096, // Large value
				"top_p":       1.0,  // Max value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := openai.ChatCompletionNewParams{
				Model:    openai.ChatModel("gpt-4"),
				Messages: []openai.ChatCompletionMessageParamUnion{},
			}

			modelConfig := &neurotypes.ModelConfig{
				BaseModel:  "gpt-4",
				Parameters: tt.parameters,
			}

			// Should not panic with any parameter configuration
			assert.NotPanics(t, func() {
				client.applyModelParameters(&params, modelConfig)
			})
		})
	}
}
