package builtin

import (
	"fmt"
	
	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type BashCommand struct{}

func (c *BashCommand) Name() string {
	return "bash"
}

func (c *BashCommand) ParseMode() types.ParseMode {
	return types.ParseModeRaw
}

func (c *BashCommand) Description() string {
	return "Execute system command"
}

func (c *BashCommand) Usage() string {
	return "\\bash[command] or \\bash command"
}

func (c *BashCommand) Execute(args map[string]string, input string, ctx types.Context) error {
	// For bash command, we need to check the raw bracket content
	// This will require coordination with the parser to pass the raw content
	var command string
	
	// Try to get command from input (both bracket and space syntax)
	if input != "" {
		command = input
	}
	
	if command == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}
	
	// TODO: Implement actual bash execution with proper security and sandboxing
	// For now, just echo what would be executed
	fmt.Printf("Executing: %s (not implemented yet)\n", command)
	
	return nil
}

func init() {
	commands.GlobalRegistry.Register(&BashCommand{})
}