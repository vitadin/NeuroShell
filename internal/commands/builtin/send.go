package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
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

// Execute acts as a router to delegate to either send-stream or send-sync based on the _reply_way variable.
// It follows the established router pattern similar to the \try command.
func (c *SendCommand) Execute(args map[string]string, input string) error {
	logger.CommandExecution("send", args)
	logger.Debug("Send router starting", "input", input)

	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// 1. Get variable service to check routing configuration
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		logger.Error("Failed to get variable service", "error", err)
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// 2. Check _reply_way variable (default to "sync" if not set)
	replyWay, _ := variableService.Get("_reply_way")
	if replyWay == "" {
		replyWay = "sync" // Default to synchronous mode
	}
	logger.Debug("Reply way determined", "reply_way", replyWay)

	// 3. Determine target command based on reply way
	var targetCommand string
	if replyWay == "stream" {
		targetCommand = "send-stream"
	} else {
		targetCommand = "send-sync"
	}
	logger.Debug("Target command determined", "target_command", targetCommand)

	// 4. Get command registry and execute target command
	registry := commands.GetGlobalRegistry()
	logger.Debug("Executing target command", "command", targetCommand, "args", args, "input", input)
	err = registry.Execute(targetCommand, args, input)

	// 5. Handle errors (following \try pattern - never fail, capture in variables)
	if err != nil {
		logger.Error("Target command failed", "command", targetCommand, "error", err)
		_ = variableService.SetSystemVariable("_status", "1")
		_ = variableService.SetSystemVariable("_error", err.Error())
	} else {
		logger.Debug("Target command succeeded", "command", targetCommand)
		_ = variableService.SetSystemVariable("_status", "0")
		_ = variableService.SetSystemVariable("_error", "")
	}

	return nil // Router never fails, always captures errors in variables
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send command: %v", err))
	}
}
