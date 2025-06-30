package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/pkg/types"
)

// SetCommand implements the \set command for setting variable values.
// It supports both bracket syntax (\set[var=value]) and space syntax (\set var value).
type SetCommand struct{}

// Name returns the command name "set" for registration and lookup.
func (c *SetCommand) Name() string {
	return "set"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SetCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

// Description returns a brief description of what the set command does.
func (c *SetCommand) Description() string {
	return "Set a variable"
}

// Usage returns the syntax and usage examples for the set command.
func (c *SetCommand) Usage() string {
	return "\\set[var=value] or \\set var value"
}

// Execute sets variable values using either bracket or space syntax.
// It handles multiple variable assignments and provides confirmation output.
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
		// Trim only leading whitespace, then split
		trimmedInput := strings.TrimLeft(input, " \t")
		var key, value string
		if trimmedInput != "" {
			parts := strings.SplitN(trimmedInput, " ", 2)
			if len(parts) == 2 {
				key, value = parts[0], strings.TrimLeft(parts[1], " \t")
			} else {
				key, value = parts[0], ""
			}

			if err := ctx.SetVariable(key, value); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", key, err)
			}
			fmt.Printf("Setting %s = %s\n", key, value)
		}
	}

	return nil
}

func init() {
	commands.GlobalRegistry.Register(&SetCommand{})
}
