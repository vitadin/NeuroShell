package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SetCommand implements the \set command for setting variable values.
// It supports both bracket syntax (\set[var=value]) and space syntax (\set var value).
type SetCommand struct{}

// Name returns the command name "set" for registration and lookup.
func (c *SetCommand) Name() string {
	return "set"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the set command does.
func (c *SetCommand) Description() string {
	return "Set a variable"
}

// Usage returns the syntax and usage examples for the set command.
func (c *SetCommand) Usage() string {
	return "\\set[var=value] or \\set var value"
}

// HelpInfo returns structured help information for the set command.
func (c *SetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\set[var=value] or \\set var value",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "var",
				Description: "Variable name to set",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "value",
				Description: "Value to assign to the variable",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\set[name=John]",
				Description: "Set variable 'name' to 'John' using bracket syntax",
			},
			{
				Command:     "\\set name John",
				Description: "Set variable 'name' to 'John' using space syntax",
			},
			{
				Command:     "\\set[greeting=\"Hello World\"]",
				Description: "Set variable with quoted value containing spaces",
			},
		},
		Notes: []string{
			"Supports both bracket syntax (\\set[var=value]) and space syntax (\\set var value)",
			"Variable values can contain spaces when properly quoted",
			"Variables are stored in session context and persist until session ends",
		},
	}
}

// Execute sets variable values using either bracket or space syntax.
// It handles multiple variable assignments and provides confirmation output.
func (c *SetCommand) Execute(args map[string]string, input string) error {
	if len(args) == 0 && input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get variable service
	variableService, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}
	varService := variableService.(*services.VariableService)

	// Handle bracket syntax: \set[var=value]
	if len(args) > 0 {
		for key, value := range args {
			if err := varService.Set(key, value); err != nil {
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

			if err := varService.Set(key, value); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", key, err)
			}
			fmt.Printf("Setting %s = %s\n", key, value)
		}
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register set command: %v", err))
	}
}
