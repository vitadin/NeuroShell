package builtin

// COMMENTED OUT: Send command functionality disabled during state machine transition
// This file contains the \send command implementation which is temporarily disabled
// while transitioning to the new state machine execution model.
//
// To re-enable: uncomment all code below (except this comment block)

/*
import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SendCommand implements the \send command for sending messages to LLM agents.
// It provides explicit message sending functionality within the NeuroShell environment.
type SendCommand struct{}

// Name returns the command name "send" for registration and lookup.
func (c *SendCommand) Name() string {
	return "send"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SendCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the send command does.
func (c *SendCommand) Description() string {
	return "Send message to LLM agent"
}

// Usage returns the syntax and usage examples for the send command.
func (c *SendCommand) Usage() string {
	return "\\send message"
}

// HelpInfo returns structured help information for the send command.
func (c *SendCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\send Hello, how are you?",
				Description: "Send a simple message to the LLM agent",
			},
			{
				Command:     "\\send Analyze this data: ${data_variable}",
				Description: "Send message with variable interpolation",
			},
			{
				Command:     "\\send ${_output}",
				Description: "Send content from editor or previous command output",
			},
			{
				Command:     "\\send Please review this code: \\n${code_content}",
				Description: "Send multi-line message with embedded content",
			},
		},
		Notes: []string{
			"Messages are sent to the active LLM session",
			"Variables are interpolated before sending",
			"Supports multi-line messages and embedded content",
			"Response will be stored in message history variables (${1}, ${2}, etc.)",
			"Use \\set[_reply_way=stream] to enable streaming mode",
			"Use \\set[_reply_way=sync] to enable synchronous mode",
			"Routes to send-stream or send-sync based on _reply_way variable",
		},
	}
}

// Execute orchestrates the full send pipeline, handling setup, LLM interaction, and cleanup.
// It determines the reply mode and delegates only the LLM interaction to sync/stream handlers.
func (c *SendCommand) Execute(args map[string]string, input string) error {
	logger.CommandExecution("send", args)
	logger.Debug("Send pipeline starting", "input", input)

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

	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		logger.Error("Failed to get client factory service", "error", err)
		return fmt.Errorf("failed to get client factory service: %w", err)
	}

	llmService, err := services.GetGlobalLLMService()
	if err != nil {
		logger.Error("Failed to get LLM service", "error", err)
		return fmt.Errorf("failed to get LLM service: %w", err)
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

	// 5. Get global context for API key determination
	globalContext := context.GetGlobalContext()

	// 6. Determine API key for the specific provider
	logger.Debug("Determining API key for provider", "provider", modelConfig.Provider)
	apiKey, err := clientFactory.DetermineAPIKeyForProvider(modelConfig.Provider, globalContext)
	if err != nil {
		logger.Error("Failed to determine API key", "provider", modelConfig.Provider, "error", err)
		return fmt.Errorf("failed to determine API key: %w", err)
	}

	// 7. Get appropriate client for the provider
	logger.Debug("Getting LLM client", "provider", modelConfig.Provider)
	client, err := clientFactory.GetClientForProvider(modelConfig.Provider, apiKey)
	if err != nil {
		logger.Error("Failed to get LLM client", "provider", modelConfig.Provider, "error", err)
		return fmt.Errorf("failed to get LLM client: %w", err)
	}

	// 8. Determine reply mode and execute appropriate LLM interaction
	replyWay, _ := variableService.Get("_reply_way")
	if replyWay == "" {
		replyWay = "sync" // Default to synchronous mode
	}
	logger.Debug("Reply way determined", "reply_way", replyWay)

	var response string
	if replyWay == "stream" {
		response, err = c.handleStreamingLLM(llmService, client, activeSession, modelConfig, input)
	} else {
		response, err = c.handleSyncLLM(llmService, client, activeSession, modelConfig, input)
	}
	if err != nil {
		logger.Error("LLM interaction failed", "reply_way", replyWay, "error", err)
		return fmt.Errorf("LLM interaction failed: %w", err)
	}

	// 9. Add user message to session
	logger.Debug("Adding user message to session")
	err = chatSessionService.AddMessage(activeSession.ID, "user", input)
	if err != nil {
		logger.Error("Failed to add user message", "error", err)
		return fmt.Errorf("failed to add user message: %w", err)
	}

	// 10. Add LLM response to session
	logger.Debug("Adding LLM response to session")
	err = chatSessionService.AddMessage(activeSession.ID, "assistant", response)
	if err != nil {
		logger.Error("Failed to add assistant message", "error", err)
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	// 11. Update message history variables (${1}, ${2}, etc.)
	logger.Debug("Updating message history variables")
	err = variableService.UpdateMessageHistoryVariables(activeSession)
	if err != nil {
		logger.Error("Failed to update variables", "error", err)
		return fmt.Errorf("failed to update variables: %w", err)
	}

	logger.Debug("Send pipeline completed successfully")
	return nil
}

// handleSyncLLM handles synchronous LLM completion and returns the complete response.
func (c *SendCommand) handleSyncLLM(llmService neurotypes.LLMService, client neurotypes.LLMClient, activeSession *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig, input string) (string, error) {
	logger.Debug("Sending synchronous LLM request", "model", modelConfig.BaseModel)
	response, err := llmService.SendCompletion(client, activeSession, modelConfig, input)
	if err != nil {
		logger.Error("LLM request failed", "error", err)
		return "", fmt.Errorf("LLM request failed: %w", err)
	}
	logger.Debug("LLM response received", "response_length", len(response))

	// Display response to user
	logger.Debug("Displaying response to user")
	fmt.Println(response)

	return response, nil
}

// handleStreamingLLM handles streaming LLM completion and returns the complete response.
func (c *SendCommand) handleStreamingLLM(llmService neurotypes.LLMService, client neurotypes.LLMClient, activeSession *neurotypes.ChatSession, modelConfig *neurotypes.ModelConfig, input string) (string, error) {
	logger.Debug("Sending streaming LLM request", "model", modelConfig.BaseModel)
	stream, err := llmService.StreamCompletion(client, activeSession, modelConfig, input)
	if err != nil {
		logger.Error("LLM stream request failed", "error", err)
		return "", fmt.Errorf("LLM stream request failed: %w", err)
	}

	// Process streaming response
	logger.Debug("Processing streaming response")
	responseBuilder := ""
	for chunk := range stream {
		if chunk.Error != nil {
			logger.Error("Stream error", "error", chunk.Error)
			return "", fmt.Errorf("stream error: %w", chunk.Error)
		}

		// Display chunk immediately to user
		fmt.Print(chunk.Content)

		// Build complete response
		responseBuilder += chunk.Content

		if chunk.Done {
			break
		}
	}

	// Print newline after streaming completes
	fmt.Println()

	return responseBuilder, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send command: %v", err))
	}
}
*/
