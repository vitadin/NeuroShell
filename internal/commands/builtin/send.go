package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
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
			"Use without explicit \\send for convenience - plain text is auto-sent",
		},
	}
}

// Execute sends the provided message to an LLM agent.
// Currently returns a placeholder message as actual LLM integration is not yet implemented.
func (c *SendCommand) Execute(_ map[string]string, input string, _ neurotypes.Context) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// TODO: Implement actual LLM agent communication
	// For now, just echo the message
	fmt.Printf("Sending: %s\n", input)

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SendCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register send command: %v", err))
	}
}
