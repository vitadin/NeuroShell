package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type SendCommand struct{}

func (c *SendCommand) Name() string {
	return "send"
}

func (c *SendCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *SendCommand) Description() string {
	return "Send message to LLM agent"
}

func (c *SendCommand) Usage() string {
	return "\\send message"
}

func (c *SendCommand) Execute(args map[string]string, input string, ctx types.Context) error {
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
