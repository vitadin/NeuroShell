package builtin

import (
	"fmt"

	"os"

	"neuroshell/internal/commands"
	"neuroshell/internal/logger"
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
	logger.CommandExecution("send-sync", nil)
	logger.Debug("Send-sync starting", "input", input)

	// 1. Input validation
	if input == "" {
		logger.Debug("Input validation failed - empty input")
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

	// 4. Note: User message will be added by the LLM service during request

	// 5. Get model configuration for active session
	logger.Debug("Getting model configuration")
	modelConfig, err := modelService.GetActiveModelConfigWithGlobalContext()
	if err != nil {
		logger.Error("Failed to get model config", "error", err)
		return fmt.Errorf("failed to get model config: %w", err)
	}
	logger.Debug("Model config acquired", "model", modelConfig.BaseModel, "provider", modelConfig.Provider)

	// 6. Determine API key source (model config, user config, or env var)
	apiKey := c.determineAPIKey(modelConfig)
	if apiKey == "" {
		logger.Error("No API key found")
		return fmt.Errorf("no API key found. Set OPENAI_API_KEY environment variable or configure in model")
	}

	// 7. Get appropriate client
	logger.Debug("Getting LLM client", "provider", modelConfig.Provider)
	client, err := clientFactory.GetClient(apiKey)
	if err != nil {
		logger.Error("Failed to get LLM client", "error", err)
		return fmt.Errorf("failed to get LLM client: %w", err)
	}

	// 8. Send synchronous request to LLM using new orchestration pattern
	logger.Debug("Sending LLM request", "model", modelConfig.BaseModel)
	response, err := llmService.SendCompletion(client, activeSession, modelConfig, input)
	if err != nil {
		logger.Error("LLM request failed", "error", err)
		return fmt.Errorf("LLM request failed: %w", err)
	}
	logger.Debug("LLM response received", "response_length", len(response))

	// 9. Display response to user
	logger.Debug("Displaying response to user")
	fmt.Println(response)

	// 10. Add user message to session
	logger.Debug("Adding user message to session")
	err = chatSessionService.AddMessage(activeSession.ID, "user", input)
	if err != nil {
		logger.Error("Failed to add user message", "error", err)
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// 11. Add LLM response to session
	logger.Debug("Adding LLM response to session")
	err = chatSessionService.AddMessage(activeSession.ID, "assistant", response)
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

	logger.Debug("Send-sync completed successfully")
	return nil
}

// determineAPIKey determines the API key from multiple sources in order of preference:
// 1. Model configuration
// 2. Environment variable
func (c *SendSyncCommand) determineAPIKey(_ *neurotypes.ModelConfig) string {
	// Check model configuration first (future enhancement)
	// if modelConfig.APIKey != "" {
	//     return modelConfig.APIKey
	// }

	// Check environment variable
	return os.Getenv("OPENAI_API_KEY")
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendSyncCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-sync command: %v", err))
	}
}
