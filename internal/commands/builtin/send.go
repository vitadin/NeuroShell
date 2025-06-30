package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

// SendCommand implements the \send command for sending messages to LLM agents.
// It provides explicit message sending functionality within the NeuroShell environment.
type SendCommand struct{}

// Name returns the command name "send" for registration and lookup.
func (c *SendCommand) Name() string {
	return "send"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SendCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

// Description returns a brief description of what the send command does.
func (c *SendCommand) Description() string {
	return "Send message to LLM agent"
}

// Usage returns the syntax and usage examples for the send command.
func (c *SendCommand) Usage() string {
	return "\\send message"
}

// Execute sends the provided message to an LLM agent.
// Currently returns a placeholder message as actual LLM integration is not yet implemented.
func (c *SendCommand) Execute(_ map[string]string, input string, _ types.Context) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// TODO: Implement actual LLM agent communication
	// For now, just echo the message
	fmt.Printf("Sending: %s\n", input)

	return nil
}

func init() {
	commands.GlobalRegistry.Register(&SendCommand{})
}
