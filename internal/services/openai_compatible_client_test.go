package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestNewOpenAICompatibleClient(t *testing.T) {
	tests := []struct {
		name     string
		config   OpenAICompatibleConfig
		expected struct {
			apiKey  string
			baseURL string
			headers map[string]string
		}
	}{
		{
			name: "with custom base URL and headers",
			config: OpenAICompatibleConfig{
				ProviderName: "custom-provider",
				APIKey:       "test-key",
				BaseURL:      "https://api.example.com/v1",
				Headers: map[string]string{
					"HTTP-Referer": "https://example.com",
					"X-Title":      "Test App",
				},
			},
			expected: struct {
				apiKey  string
				baseURL string
				headers map[string]string
			}{
				apiKey:  "test-key",
				baseURL: "https://api.example.com/v1",
				headers: map[string]string{
					"HTTP-Referer": "https://example.com",
					"X-Title":      "Test App",
				},
			},
		},
		{
			name: "with empty base URL defaults to OpenRouter",
			config: OpenAICompatibleConfig{
				ProviderName: "",
				APIKey:       "test-key",
				BaseURL:      "",
				Headers:      nil,
			},
			expected: struct {
				apiKey  string
				baseURL string
				headers map[string]string
			}{
				apiKey:  "test-key",
				baseURL: "https://openrouter.ai/api/v1",
				headers: map[string]string{},
			},
		},
		{
			name: "base URL with trailing slash is trimmed",
			config: OpenAICompatibleConfig{
				ProviderName: "",
				APIKey:       "test-key",
				BaseURL:      "https://api.example.com/v1/",
				Headers:      nil,
			},
			expected: struct {
				apiKey  string
				baseURL string
				headers map[string]string
			}{
				apiKey:  "test-key",
				baseURL: "https://api.example.com/v1",
				headers: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAICompatibleClient(tt.config)

			assert.Equal(t, tt.expected.apiKey, client.apiKey)
			assert.Equal(t, tt.expected.baseURL, client.baseURL)

			if tt.expected.headers == nil {
				assert.Empty(t, client.headers)
			} else {
				assert.Equal(t, tt.expected.headers, client.headers)
			}
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestOpenAICompatibleClient_GetProviderName(t *testing.T) {
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "test-key",
		BaseURL:      "https://api.example.com/v1",
	})

	assert.Equal(t, "test-provider", client.GetProviderName())
}

func TestOpenAICompatibleClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   OpenAICompatibleConfig
		expected bool
	}{
		{
			name: "fully configured",
			config: OpenAICompatibleConfig{
				ProviderName: "test",
				APIKey:       "test-key",
				BaseURL:      "https://api.example.com/v1",
			},
			expected: true,
		},
		{
			name: "missing API key",
			config: OpenAICompatibleConfig{
				ProviderName: "test",
				APIKey:       "",
				BaseURL:      "https://api.example.com/v1",
			},
			expected: false,
		},
		{
			name: "missing base URL",
			config: OpenAICompatibleConfig{
				ProviderName: "test",
				APIKey:       "test-key",
				BaseURL:      "",
			},
			expected: false, // Will be set to OpenRouter default, so should be true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAICompatibleClient(tt.config)
			// For empty baseURL case, the constructor sets it to OpenRouter default
			if tt.name == "missing base URL" {
				tt.expected = true
			}
			assert.Equal(t, tt.expected, client.IsConfigured())
		})
	}
}

func TestOpenAICompatibleClient_SendChatCompletion(t *testing.T) {
	// Create a mock server that mimics OpenAI-compatible API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "https://example.com", r.Header.Get("HTTP-Referer"))
		assert.Equal(t, "Test App", r.Header.Get("X-Title"))

		// Parse and verify request body
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "gpt-4", req.Model)
		assert.False(t, req.Stream)
		assert.Len(t, req.Messages, 2) // system + user message
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "You are a helpful assistant.", req.Messages[0].Content)
		assert.Equal(t, "user", req.Messages[1].Role)
		assert.Equal(t, "Hello, world!", req.Messages[1].Content)

		// Send mock response
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: &ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: stringPtr("stop"),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "test-api-key",
		BaseURL:      mockServer.URL,
		Headers: map[string]string{
			"HTTP-Referer": "https://example.com",
			"X-Title":      "Test App",
		},
	})

	// Create test session
	session := &neurotypes.ChatSession{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []neurotypes.Message{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
	}

	// Create test model config
	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Parameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		},
	}

	// Test successful completion
	result, err := client.SendChatCompletion(session, modelConfig)
	require.NoError(t, err)
	assert.Equal(t, "Hello! How can I help you today?", result)
}

func TestOpenAICompatibleClient_SendChatCompletion_Error(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    string
	}{
		{
			name: "API error response",
			serverResponse: func(w http.ResponseWriter, _ *http.Request) {
				response := ChatCompletionResponse{
					Error: &ChatCompletionError{
						Message: "Invalid API key",
						Type:    "invalid_request_error",
						Code:    "invalid_api_key",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			},
			expectError: "API error: Invalid API key",
		},
		{
			name: "HTTP error response",
			serverResponse: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorized"))
			},
			expectError: "HTTP error 401: Unauthorized",
		},
		{
			name: "empty choices response",
			serverResponse: func(w http.ResponseWriter, _ *http.Request) {
				response := ChatCompletionResponse{
					ID:      "chatcmpl-test",
					Object:  "chat.completion",
					Created: 1234567890,
					Model:   "gpt-4",
					Choices: []ChatCompletionChoice{},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			},
			expectError: "no response choices returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer mockServer.Close()

			client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
				ProviderName: "test-provider",
				APIKey:       "test-api-key",
				BaseURL:      mockServer.URL,
			})

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
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestOpenAICompatibleClient_StreamChatCompletion(t *testing.T) {
	// Create a mock server that mimics OpenAI-compatible streaming API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))

		// Parse request to verify streaming is enabled
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Stream)

		// Send streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send chunks
		chunks := []string{
			"Hello",
			" there",
			"! How",
			" can I",
			" help?",
		}

		for _, chunk := range chunks {
			response := ChatCompletionResponse{
				ID:      "chatcmpl-test",
				Object:  "chat.completion.chunk",
				Created: 1234567890,
				Model:   "gpt-4",
				Choices: []ChatCompletionChoice{
					{
						Index: 0,
						Delta: &ChatCompletionMessage{
							Content: chunk,
						},
						FinishReason: nil,
					},
				},
			}

			responseJSON, _ := json.Marshal(response)
			_, _ = w.Write([]byte("data: " + string(responseJSON) + "\n\n"))

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		// Send final chunk
		finalResponse := ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion.chunk",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []ChatCompletionChoice{
				{
					Index:        0,
					Delta:        &ChatCompletionMessage{},
					FinishReason: stringPtr("stop"),
				},
			},
		}
		finalJSON, _ := json.Marshal(finalResponse)
		_, _ = w.Write([]byte("data: " + string(finalJSON) + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer mockServer.Close()

	// Create client
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "test-api-key",
		BaseURL:      mockServer.URL,
	})

	// Create test session
	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
	}

	// Test streaming
	stream, err := client.StreamChatCompletion(session, modelConfig)
	require.NoError(t, err)

	// Collect chunks
	var chunks []string
	var finalChunk neurotypes.StreamChunk
	for chunk := range stream {
		if chunk.Done {
			finalChunk = chunk
			break
		}
		if chunk.Content != "" {
			chunks = append(chunks, chunk.Content)
		}
	}

	// Verify results
	expected := []string{"Hello", " there", "! How", " can I", " help?"}
	assert.Equal(t, expected, chunks)
	assert.True(t, finalChunk.Done)
	assert.NoError(t, finalChunk.Error)
}

func TestOpenAICompatibleClient_ConvertMessagesToOpenAI(t *testing.T) {
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "test-key",
		BaseURL:      "https://api.example.com/v1",
	})

	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
			{Role: "unknown", Content: "This should be skipped"},
		},
	}

	messages := client.convertMessagesToOpenAI(session)

	expected := []ChatCompletionMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	assert.Equal(t, expected, messages)
}

func TestOpenAICompatibleClient_ApplyModelParameters(t *testing.T) {
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "test-key",
		BaseURL:      "https://api.example.com/v1",
	})

	request := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []ChatCompletionMessage{},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
		Parameters: map[string]interface{}{
			"temperature":       0.7,
			"max_tokens":        100,
			"top_p":             0.9,
			"frequency_penalty": 0.1,
			"presence_penalty":  0.2,
		},
	}

	client.applyModelParameters(request, modelConfig)

	assert.Equal(t, 0.7, *request.Temperature)
	assert.Equal(t, 100, *request.MaxTokens)
	assert.Equal(t, 0.9, *request.TopP)
	assert.Equal(t, 0.1, *request.FrequencyPenalty)
	assert.Equal(t, 0.2, *request.PresencePenalty)
}

func TestOpenAICompatibleClient_NotConfigured(t *testing.T) {
	client := NewOpenAICompatibleClient(OpenAICompatibleConfig{
		ProviderName: "test-provider",
		APIKey:       "", // Missing API key
		BaseURL:      "https://api.example.com/v1",
	})

	session := &neurotypes.ChatSession{
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Test"},
		},
	}

	modelConfig := &neurotypes.ModelConfig{
		BaseModel: "gpt-4",
	}

	// Test SendChatCompletion
	_, err := client.SendChatCompletion(session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	// Test StreamChatCompletion
	_, err = client.StreamChatCompletion(session, modelConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
