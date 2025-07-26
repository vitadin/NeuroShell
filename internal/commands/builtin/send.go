package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SendCommand implements the \send command as a delegation wrapper to the _send neuro script.
// This provides clean help system integration while keeping the complex LLM logic in the neuro script.
type SendCommand struct{}

// Name returns the command name "send" for registration and lookup.
func (c *SendCommand) Name() string {
	return "send"
}

// ParseMode returns ParseModeRaw to preserve the entire input message for the neuro script.
func (c *SendCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the send command does.
func (c *SendCommand) Description() string {
	return "Send message to LLM agent"
}

// Usage returns the syntax and usage examples for the send command.
func (c *SendCommand) Usage() string {
	return "\\send message"
}

// HelpInfo returns comprehensive help information for the send command.
func (c *SendCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for send command - configured via variables
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
				Command:     "\\send Please review this code:\\n${code_content}",
				Description: "Send multi-line message with embedded content",
			},
			{
				Command:     "\\set[_reply_way=stream] && \\send Tell me a story",
				Description: "Send message with streaming response mode",
			},
			{
				Command:     "\\set[_reply_way=sync] && \\send What is 2+2?",
				Description: "Send message with synchronous response mode",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "1",
				Description: "Latest agent response message",
				Type:        "message_history",
				Example:     "Hello! I'm Claude, an AI assistant...",
			},
			{
				Name:        "2",
				Description: "Previous agent response (2nd most recent)",
				Type:        "message_history",
				Example:     "I can help you with that...",
			},
			{
				Name:        "#message_count",
				Description: "Total number of messages in current session",
				Type:        "system_metadata",
				Example:     "5",
			},
		},
		Notes: []string{
			"Automatically creates chat session if none exists",
			"Uses active model configuration (use \\model-activate to set)",
			"Variables are interpolated before sending (${var} syntax)",
			"Set _reply_way variable to control response mode:",
			"  • _reply_way=sync: Complete response at once (default)",
			"  • _reply_way=stream: Real-time streaming response",
			"Requires API key: OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.",
			"Multi-line messages supported with \\n escape sequences",
			"Error messages preserved on stderr for debugging",
		},
	}
}

// Execute delegates to the _send neuro script via stack service.
func (c *SendCommand) Execute(_ map[string]string, input string) error {
	// Input validation
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get stack service for delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Delegate to _send neuro script
	stackService.PushCommand("\\_send " + input)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&SendCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send command: %v", err))
	}
}
