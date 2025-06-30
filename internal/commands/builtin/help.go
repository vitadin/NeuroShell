package builtin

import (
	"fmt"
	"sort"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type HelpCommand struct{}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *HelpCommand) Description() string {
	return "Show command help"
}

func (c *HelpCommand) Usage() string {
	return "\\help [command]"
}

func (c *HelpCommand) Execute(args map[string]string, input string, ctx types.Context) error {
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
