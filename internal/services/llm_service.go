package services

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"neuroshell/pkg/neurotypes"
)

// LLMService provides LLM communication capabilities for NeuroShell.
// It handles both streaming and synchronous communication with OpenAI models.
type LLMService struct {
	initialized bool
	client      *openai.Client
}

// NewLLMService creates a new LLMService instance.
func NewLLMService() *LLMService {
	return &LLMService{
		initialized: false,
	}
}

// Name returns the service name "llm" for registration.
func (l *LLMService) Name() string {
	return "llm"
}

// Initialize sets up the LLMService for operation.
// It creates the OpenAI client with API key from environment.
func (l *LLMService) Initialize() error {
	// Create OpenAI client (will get API key from OPENAI_API_KEY env var)
	client := openai.NewClient()
	l.client = &client
	l.initialized = true
	return nil
}

// SendChatCompletion sends a chat completion request synchronously.
// It takes the full conversation history from the session and returns the complete response.
func (l *LLMService) SendChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	if !l.initialized {
		return "", fmt.Errorf("llm service not initialized")
	}

	// Convert session messages to OpenAI format
	messages := l.convertMessagesToOpenAI(session)

	// Add system prompt if present
	if session.SystemPrompt != "" {
		systemMsg := openai.SystemMessage(session.SystemPrompt)
		messages = append([]openai.ChatCompletionMessageParamUnion{systemMsg}, messages...)
	}

	// Build completion parameters
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(modelConfig.BaseModel),
		Messages: messages,
	}

	// Apply model parameters if available
	l.applyModelParameters(&params, modelConfig)

	// Send request
	completion, err := l.client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}

	// Extract response content
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	content := completion.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty response content")
	}

	return content, nil
}

// SendChatCompletionWithGlobalContext sends a chat completion request using the global context singleton.
func (l *LLMService) SendChatCompletionWithGlobalContext(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (string, error) {
	return l.SendChatCompletion(session, modelConfig)
}

// StreamChatCompletion sends a streaming chat completion request.
// It returns a channel that receives response chunks as they arrive.
func (l *LLMService) StreamChatCompletion(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan StreamChunk, error) {
	if !l.initialized {
		return nil, fmt.Errorf("llm service not initialized")
	}

	// Convert session messages to OpenAI format
	messages := l.convertMessagesToOpenAI(session)

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
	l.applyModelParameters(&params, modelConfig)

	// Create streaming request
	stream := l.client.Chat.Completions.NewStreaming(context.Background(), params)

	// Create response channel
	responseChan := make(chan StreamChunk, 10)

	// Start goroutine to handle streaming response
	go func() {
		defer close(responseChan)
		defer func() { _ = stream.Close() }()

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				responseChan <- StreamChunk{
					Content: chunk.Choices[0].Delta.Content,
					Done:    false,
				}
			}
		}

		if err := stream.Err(); err != nil {
			responseChan <- StreamChunk{
				Content: "",
				Done:    true,
				Error:   err,
			}
		} else {
			responseChan <- StreamChunk{
				Content: "",
				Done:    true,
				Error:   nil,
			}
		}
	}()

	return responseChan, nil
}

// StreamChatCompletionWithGlobalContext sends a streaming chat completion request using the global context singleton.
func (l *LLMService) StreamChatCompletionWithGlobalContext(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig) (<-chan StreamChunk, error) {
	return l.StreamChatCompletion(session, modelConfig)
}

// StreamChunk represents a single chunk of streaming response.
type StreamChunk struct {
	Content string // The text content of this chunk
	Done    bool   // Whether this is the final chunk
	Error   error  // Any error that occurred during streaming
}

// convertMessagesToOpenAI converts NeuroShell messages to OpenAI format.
func (l *LLMService) convertMessagesToOpenAI(session *neurotypes.ChatSession) []openai.ChatCompletionMessageParamUnion {
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
func (l *LLMService) applyModelParameters(params *openai.ChatCompletionNewParams, modelConfig *neurotypes.ModelConfig) {
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

// SendChatCompletionWithCallback sends a streaming chat completion request with a callback function.
// The callback is called for each chunk of the response as it arrives.
func (l *LLMService) SendChatCompletionWithCallback(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig, onChunk func(string)) error {
	stream, err := l.StreamChatCompletion(session, modelConfig)
	if err != nil {
		return err
	}

	for chunk := range stream {
		if chunk.Error != nil {
			return chunk.Error
		}
		if chunk.Content != "" {
			onChunk(chunk.Content)
		}
		if chunk.Done {
			break
		}
	}

	return nil
}

// SendChatCompletionWithCallbackAndGlobalContext sends a streaming chat completion request with a callback using the global context singleton.
func (l *LLMService) SendChatCompletionWithCallbackAndGlobalContext(session *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig, onChunk func(string)) error {
	return l.SendChatCompletionWithCallback(session, modelConfig, onChunk)
}
