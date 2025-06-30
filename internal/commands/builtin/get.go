package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

type GetCommand struct{}

func (c *GetCommand) Name() string {
	return "get"
}

func (c *GetCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *GetCommand) Description() string {
	return "Get a variable"
}

func (c *GetCommand) Usage() string {
	return "\\get[var] or \\get var"
}

func (c *GetCommand) Execute(args map[string]string, input string, ctx types.Context) error {
	var variable string

	// Handle bracket syntax: \get[var]
	if len(args) > 0 {
		for key := range args {
			variable = key
			break
		}
	} else if input != "" {
		// Handle space syntax: \get var
		fields := strings.Fields(input)
		if len(fields) > 0 {
			variable = fields[0]
		}
	}

	if variable == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	value, err := ctx.GetVariable(variable)
	if err != nil {
		return fmt.Errorf("failed to get variable %s: %w", variable, err)
	}

	fmt.Printf("%s = %s\n", variable, value)
	return nil
}

func init() {
	commands.GlobalRegistry.Register(&GetCommand{})
}
