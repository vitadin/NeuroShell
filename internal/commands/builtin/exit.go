package builtin

import (
	"fmt"
	"os"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

// ExitCommand implements the \exit command for terminating the NeuroShell session.
// It provides a clean way to exit the shell environment.
type ExitCommand struct{}

// Name returns the command name "exit" for registration and lookup.
func (c *ExitCommand) Name() string {
	return "exit"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ExitCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

// Description returns a brief description of what the exit command does.
func (c *ExitCommand) Description() string {
	return "Exit the shell"
}

// Usage returns the syntax and usage examples for the exit command.
func (c *ExitCommand) Usage() string {
	return "\\exit"
}

// Execute terminates the NeuroShell session by calling os.Exit(0).
// This provides an immediate exit from the shell environment.
func (c *ExitCommand) Execute(_ map[string]string, _ string, _ types.Context) error {
	fmt.Println("Goodbye!")
	// For now, we'll use os.Exit. In the future, we might want to use a more graceful shutdown
	// that could be coordinated through the context or a shutdown signal
	os.Exit(0)
	return nil
}

func init() {
	commands.GlobalRegistry.Register(&ExitCommand{})
}
