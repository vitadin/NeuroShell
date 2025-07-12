package builtin

// COMMENTED OUT: Send-stream command functionality disabled during state machine transition
// This file contains the \send-stream command implementation which is temporarily disabled
// while transitioning to the new state machine execution model.
//
// To re-enable: uncomment all code below (except this comment block)

/*
import (
	"fmt"

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
			"Requires provider-specific API key: OPENAI_API_KEY for OpenAI, ANTHROPIC_API_KEY for Anthropic",
		},
	}
}

// Execute runs the send-stream command by delegating to the main send command with stream mode.
func (c *SendStreamCommand) Execute(args map[string]string, input string) error {
	logger.CommandExecution("send-stream", args)
	logger.Debug("Send-stream delegating to main send command", "input", input)

	// Set _reply_way to stream mode temporarily
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		logger.Error("Failed to get variable service", "error", err)
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// Save original _reply_way value
	originalReplyWay, _ := variableService.Get("_reply_way")

	// Set stream mode
	err = variableService.SetSystemVariable("_reply_way", "stream")
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
	if err := commands.GlobalRegistry.Register(&SendStreamCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send-stream command: %v", err))
	}
}
*/
