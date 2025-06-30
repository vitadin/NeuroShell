// Package builtin provides built-in NeuroShell commands that are available by default.
// These commands implement core functionality such as system execution, variable management,
// and shell control operations.
package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

// BashCommand implements the \bash command for executing system commands.
// It provides a way to run shell commands from within the NeuroShell environment.
type BashCommand struct{}

// Name returns the command name "bash" for registration and lookup.
func (c *BashCommand) Name() string {
	return "bash"
}

// ParseMode returns ParseModeRaw to treat the entire input as a raw command.
func (c *BashCommand) ParseMode() types.ParseMode {
	return types.ParseModeRaw
}

// Description returns a brief description of what the bash command does.
func (c *BashCommand) Description() string {
	return "Execute system command"
}

// Usage returns the syntax and usage examples for the bash command.
func (c *BashCommand) Usage() string {
	return "\\bash[command] or \\bash command"
}

// Execute runs the bash command with the provided input as a system command.
// Currently returns a placeholder message as actual execution is not yet implemented.
func (c *BashCommand) Execute(_ map[string]string, input string, _ types.Context) error {
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
