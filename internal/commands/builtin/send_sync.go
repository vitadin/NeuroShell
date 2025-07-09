package builtin

import (
	"fmt"

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
			"Requires provider-specific API key: OPENAI_API_KEY for OpenAI, ANTHROPIC_API_KEY for Anthropic",
		},
	}
}

// Execute runs the send-sync command by delegating to the main send command with sync mode.
func (c *SendSyncCommand) Execute(args map[string]string, input string) error {
	logger.CommandExecution("send-sync", args)
	logger.Debug("Send-sync delegating to main send command", "input", input)

	// Set _reply_way to sync mode temporarily
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		logger.Error("Failed to get variable service", "error", err)
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// Save original _reply_way value
	originalReplyWay, _ := variableService.Get("_reply_way")

	// Set sync mode
	err = variableService.SetSystemVariable("_reply_way", "sync")
	if err != nil {
		logger.Error("Failed to set _reply_way", "error", err)
		return fmt.Errorf("failed to set _reply_way: %w", err)
	}

	// Restore original value after execution
	defer func() {
		_ = variableService.SetSystemVariable("_reply_way", originalReplyWay)
	}()

	// Delegate to main send command
	sendCommand := &SendCommand{}
	return sendCommand.Execute(args, input)
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendSyncCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-sync command: %v", err))
	}
}
