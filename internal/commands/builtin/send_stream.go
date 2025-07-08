package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SendStreamCommand implements the send-stream command for streaming LLM responses.
type SendStreamCommand struct{}

// Name returns the command name.
func (c *SendStreamCommand) Name() string {
	return "send-stream"
}

// ParseMode returns the parse mode for this command.
func (c *SendStreamCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns the command description.
func (c *SendStreamCommand) Description() string {
	return "Send message to LLM agent with streaming response"
}

// Usage returns the command usage string.
func (c *SendStreamCommand) Usage() string {
	return "\\send-stream message"
}

// HelpInfo returns detailed help information for the command.
func (c *SendStreamCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     "send-stream",
		Description: "Send message to LLM agent with streaming response",
		Usage:       "\\send-stream message",
		ParseMode:   neurotypes.ParseModeKeyValue,
		Options:     []neurotypes.HelpOption{},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\send-stream Hello, how are you?",
				Description: "Send a message with streaming response",
			},
			{
				Command:     "\\send-stream What is the weather like?",
				Description: "Ask a question with streaming response",
			},
		},
		Notes: []string{
			"Messages are sent to the active chat session",
			"Responses are streamed in real-time",
			"Message history variables (${1}, ${2}, etc.) are updated after completion",
			"Requires OPENAI_API_KEY environment variable to be set",
		},
	}
}

// Execute runs the send-stream command.
func (c *SendStreamCommand) Execute(_ map[string]string, input string) error {
	// 1. Input validation
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// 2. Get required services
	chatSessionService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("failed to get chat session service: %w", err)
	}

	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("failed to get model service: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	llmService, err := services.GetGlobalLLMService()
	if err != nil {
		return fmt.Errorf("failed to get LLM service: %w", err)
	}

	// 3. Get active chat session (or create default if none exists)
	activeSession, err := chatSessionService.GetActiveSession()
	if err != nil {
		// Try to create a default session
		session, createErr := chatSessionService.CreateSession("default", "", "")
		if createErr != nil {
			return fmt.Errorf("no active chat session and failed to create default: %w", createErr)
		}
		activeSession = session
	}

	// 4. Add user message to session
	err = chatSessionService.AddMessage(activeSession.ID, "user", input)
	if err != nil {
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// 5. Get model configuration for active session
	modelConfig, err := modelService.GetActiveModelConfigWithGlobalContext()
	if err != nil {
		return fmt.Errorf("failed to get model config: %w", err)
	}

	// 6. Send streaming request to LLM
	stream, err := llmService.StreamChatCompletionWithGlobalContext(activeSession, modelConfig)
	if err != nil {
		return fmt.Errorf("LLM stream request failed: %w", err)
	}

	// 7. Process streaming response
	responseBuilder := ""
	for chunk := range stream {
		if chunk.Error != nil {
			return fmt.Errorf("stream error: %w", chunk.Error)
		}

		// Display chunk immediately to user
		fmt.Print(chunk.Content)

		// Build complete response
		responseBuilder += chunk.Content

		if chunk.Done {
			break
		}
	}

	// 8. Print newline after streaming completes
	fmt.Println()

	// 9. Add LLM response to session
	err = chatSessionService.AddMessage(activeSession.ID, "assistant", responseBuilder)
	if err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// 10. Update message history variables (${1}, ${2}, etc.)
	err = variableService.UpdateMessageHistoryVariables(activeSession)
	if err != nil {
		return fmt.Errorf("failed to update variables: %w", err)
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendStreamCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-stream command: %v", err))
	}
}
