package builtin

import (
	"fmt"
	"os"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type ExitCommand struct{}

func (c *ExitCommand) Name() string {
	return "exit"
}

func (c *ExitCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *ExitCommand) Description() string {
	return "Exit the shell"
}

func (c *ExitCommand) Usage() string {
	return "\\exit"
}

func (c *ExitCommand) Execute(args map[string]string, input string, ctx types.Context) error {
	fmt.Println("Goodbye!")
	// For now, we'll use os.Exit. In the future, we might want to use a more graceful shutdown
	// that could be coordinated through the context or a shutdown signal
	os.Exit(0)
	return nil
}

func init() {
	commands.GlobalRegistry.Register(&ExitCommand{})
}
