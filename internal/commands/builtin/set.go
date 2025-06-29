package builtin

import (
	"fmt"
	"strings"
	
	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type SetCommand struct{}

func (c *SetCommand) Name() string {
	return "set"
}

func (c *SetCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *SetCommand) Description() string {
	return "Set a variable"
}

func (c *SetCommand) Usage() string {
	return "\\set[var=value] or \\set var value"
}

func (c *SetCommand) Execute(args map[string]string, input string, ctx types.Context) error {
	if len(args) == 0 && input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}
	
	// Handle bracket syntax: \set[var=value]
	if len(args) > 0 {
		for key, value := range args {
			if err := ctx.SetVariable(key, value); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", key, err)
			}
			fmt.Printf("Setting %s = %s\n", key, value)
		}
		return nil
	}
	
	// Handle space syntax: \set var value
	if input != "" {
		parts := strings.SplitN(input, " ", 2)
		var key, value string
		if len(parts) == 2 {
			key, value = parts[0], parts[1]
		} else {
			key, value = parts[0], ""
		}
		
		if err := ctx.SetVariable(key, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", key, err)
		}
		fmt.Printf("Setting %s = %s\n", key, value)
	}
	
	return nil
}

func init() {
	commands.GlobalRegistry.Register(&SetCommand{})
}