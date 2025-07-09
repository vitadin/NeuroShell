package builtin

import (
	"fmt"
	"os"

	"neuroshell/internal/commands"
	"neuroshell/internal/logger"
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
	logger.Debug("Getting required services")
	chatSessionService, err := services.GetGlobalChatSessionService()
	if err != nil {
		logger.Error("Failed to get chat session service", "error", err)
		return fmt.Errorf("failed to get chat session service: %w", err)
	}

	modelService, err := services.GetGlobalModelService()
	if err != nil {
		logger.Error("Failed to get model service", "error", err)
		return fmt.Errorf("failed to get model service: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		logger.Error("Failed to get variable service", "error", err)
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// Get client factory service
	clientFactoryService, err := services.GetGlobalRegistry().GetService("client_factory")
	if err != nil {
		logger.Error("Failed to get client factory service", "error", err)
		return fmt.Errorf("failed to get client factory service: %w", err)
	}
	clientFactory := clientFactoryService.(neurotypes.ClientFactory)

	// Get new LLM service
	llmServiceRaw, err := services.GetGlobalRegistry().GetService("llm")
	if err != nil {
		logger.Error("Failed to get LLM service", "error", err)
		return fmt.Errorf("failed to get LLM service: %w", err)
	}
	llmService, ok := llmServiceRaw.(neurotypes.LLMService)
	if !ok {
		logger.Error("LLM service does not implement neurotypes.LLMService interface")
		return fmt.Errorf("LLM service does not implement neurotypes.LLMService interface")
	}
	logger.Debug("All services acquired successfully")

	// 3. Get active chat session (or create auto session if none exists)
	logger.Debug("Getting active chat session")
	activeSession, err := chatSessionService.GetActiveSession()
	if err != nil {
		logger.Debug("No active session found, creating auto session", "error", err)
		// Try to create an auto session (default is reserved)
		session, createErr := chatSessionService.CreateSession("auto", "", "")
		if createErr != nil {
			logger.Error("Failed to create auto session", "error", createErr)
			return fmt.Errorf("no active chat session and failed to create auto: %w", createErr)
		}
		activeSession = session
		logger.Debug("Auto session created", "session_id", activeSession.ID)
	} else {
		logger.Debug("Active session found", "session_id", activeSession.ID)
	}

	// 4. Get model configuration for active session
	logger.Debug("Getting model configuration")
	modelConfig, err := modelService.GetActiveModelConfigWithGlobalContext()
	if err != nil {
		logger.Error("Failed to get model config", "error", err)
		return fmt.Errorf("failed to get model config: %w", err)
	}
	logger.Debug("Model config acquired", "model", modelConfig.BaseModel, "provider", modelConfig.Provider)

	// 5. Determine API key source (model config, user config, or env var)
	apiKey := c.determineAPIKey(modelConfig)
	if apiKey == "" {
		logger.Error("No API key found")
		return fmt.Errorf("no API key found. Set OPENAI_API_KEY environment variable or configure in model")
	}

	// 6. Get appropriate client
	logger.Debug("Getting LLM client", "provider", modelConfig.Provider)
	client, err := clientFactory.GetClient(apiKey)
	if err != nil {
		logger.Error("Failed to get LLM client", "error", err)
		return fmt.Errorf("failed to get LLM client: %w", err)
	}

	// 7. Send streaming request to LLM using new orchestration pattern
	logger.Debug("Sending streaming LLM request", "model", modelConfig.BaseModel)
	stream, err := llmService.StreamCompletion(client, activeSession, modelConfig, input)
	if err != nil {
		logger.Error("LLM stream request failed", "error", err)
		return fmt.Errorf("LLM stream request failed: %w", err)
	}

	// 8. Process streaming response
	logger.Debug("Processing streaming response")
	responseBuilder := ""
	for chunk := range stream {
		if chunk.Error != nil {
			logger.Error("Stream error", "error", chunk.Error)
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

	// 9. Print newline after streaming completes
	fmt.Println()

	// 10. Add user message to session
	logger.Debug("Adding user message to session")
	err = chatSessionService.AddMessage(activeSession.ID, "user", input)
	if err != nil {
		logger.Error("Failed to add user message", "error", err)
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// 11. Add LLM response to session
	logger.Debug("Adding LLM response to session")
	err = chatSessionService.AddMessage(activeSession.ID, "assistant", responseBuilder)
	if err != nil {
		logger.Error("Failed to add assistant message", "error", err)
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// 12. Update message history variables (${1}, ${2}, etc.)
	logger.Debug("Updating message history variables")
	err = variableService.UpdateMessageHistoryVariables(activeSession)
	if err != nil {
		logger.Error("Failed to update variables", "error", err)
		return fmt.Errorf("failed to update variables: %w", err)
	}

	logger.Debug("Send-stream completed successfully")
	return nil
}

// determineAPIKey determines the API key from multiple sources in order of preference:
// 1. Model configuration
// 2. Environment variable
func (c *SendStreamCommand) determineAPIKey(_ *neurotypes.ModelConfig) string {
	// Check model configuration first (future enhancement)
	// if modelConfig.APIKey != "" {
	//     return modelConfig.APIKey
	// }

	// Check environment variable
	return os.Getenv("OPENAI_API_KEY")
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendStreamCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-stream command: %v", err))
	}
}
