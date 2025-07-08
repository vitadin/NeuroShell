package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SendSyncCommand implements the send-sync command for synchronous LLM responses.
type SendSyncCommand struct{}

// Name returns the command name.
func (c *SendSyncCommand) Name() string {
	return "send-sync"
}

// ParseMode returns the parse mode for this command.
func (c *SendSyncCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns the command description.
func (c *SendSyncCommand) Description() string {
	return "Send message to LLM agent with synchronous response"
}

// Usage returns the command usage string.
func (c *SendSyncCommand) Usage() string {
	return "\\send-sync message"
}

// HelpInfo returns detailed help information for the command.
func (c *SendSyncCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     "send-sync",
		Description: "Send message to LLM agent with synchronous response",
		Usage:       "\\send-sync message",
		ParseMode:   neurotypes.ParseModeKeyValue,
		Options:     []neurotypes.HelpOption{},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\send-sync Hello, how are you?",
				Description: "Send a message with synchronous response",
			},
			{
				Command:     "\\send-sync What is the weather like?",
				Description: "Ask a question with synchronous response",
			},
		},
		Notes: []string{
			"Messages are sent to the active chat session",
			"Response is received as a complete message",
			"Message history variables (${1}, ${2}, etc.) are updated after completion",
			"Requires OPENAI_API_KEY environment variable to be set",
		},
	}
}

// Execute runs the send-sync command.
func (c *SendSyncCommand) Execute(_ map[string]string, input string) error {
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

	// 6. Send synchronous request to LLM
	response, err := llmService.SendChatCompletionWithGlobalContext(activeSession, modelConfig)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}

	// 7. Display response to user
	fmt.Println(response)

	// 8. Add LLM response to session
	err = chatSessionService.AddMessage(activeSession.ID, "assistant", response)
	if err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// 9. Update message history variables (${1}, ${2}, etc.)
	err = variableService.UpdateMessageHistoryVariables(activeSession)
	if err != nil {
		return fmt.Errorf("failed to update variables: %w", err)
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendSyncCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-sync command: %v", err))
	}
}
