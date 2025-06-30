package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/pkg/neurotypes"
)

// GetCommand implements the \get command for retrieving variable values.
// It supports both bracket syntax (\get[var]) and space syntax (\get var).
type GetCommand struct{}

// Name returns the command name "get" for registration and lookup.
func (c *GetCommand) Name() string {
	return "get"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *GetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the get command does.
func (c *GetCommand) Description() string {
	return "Get a variable"
}

// Usage returns the syntax and usage examples for the get command.
func (c *GetCommand) Usage() string {
	return "\\get[var] or \\get var"
}

// Execute retrieves and displays the value of the specified variable.
// It handles both bracket and space syntax for variable specification.
func (c *GetCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
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
	if err := commands.GlobalRegistry.Register(&GetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register get command: %v", err))
	}
}
