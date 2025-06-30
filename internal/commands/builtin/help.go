package builtin

import (
	"fmt"
	"sort"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

// HelpCommand implements the \help command for displaying available commands and usage information.
// It lists all registered commands with their descriptions and provides examples.
type HelpCommand struct{}

// Name returns the command name "help" for registration and lookup.
func (c *HelpCommand) Name() string {
	return "help"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *HelpCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

// Description returns a brief description of what the help command does.
func (c *HelpCommand) Description() string {
	return "Show command help"
}

// Usage returns the syntax and usage examples for the help command.
func (c *HelpCommand) Usage() string {
	return "\\help [command]"
}

// Execute displays a list of all available commands with their descriptions and usage examples.
// It provides an overview of the NeuroShell command system.
func (c *HelpCommand) Execute(_ map[string]string, _ string, _ types.Context) error {
	// Get all commands from the registry
	allCommands := commands.GlobalRegistry.GetAll()

	// Sort commands by name for consistent output
	sort.Slice(allCommands, func(i, j int) bool {
		return allCommands[i].Name() < allCommands[j].Name()
	})

	fmt.Println("Neuro Shell Commands:")
	for _, cmd := range allCommands {
		fmt.Printf("  %-20s - %s\n", cmd.Usage(), cmd.Description())
	}

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  \\send Hello world")
	fmt.Println("  \\set[name=\"John\"]")
	fmt.Println("  \\get[name]")
	fmt.Println("  \\bash[ls -la]")
	fmt.Println()
	fmt.Println("Note: Text without \\ prefix is sent to LLM automatically")

	return nil
}

func init() {
	commands.GlobalRegistry.Register(&HelpCommand{})
}
