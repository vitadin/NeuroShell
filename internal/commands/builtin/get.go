package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
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

// HelpInfo returns structured help information for the get command.
func (c *GetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\get[var] or \\get var",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "var",
				Description: "Variable name to retrieve",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\get[name]",
				Description: "Get the value of variable 'name' using bracket syntax",
			},
			{
				Command:     "\\get name",
				Description: "Get the value of variable 'name' using space syntax",
			},
			{
				Command:     "\\get @pwd",
				Description: "Get system variable for current directory",
			},
		},
		Notes: []string{
			"Supports both bracket syntax (\\get[var]) and space syntax (\\get var)",
			"System variables start with @ (e.g., @pwd, @user, @date)",
			"Message history variables use numbers (e.g., ${1}, ${2})",
		},
	}
}

// Execute retrieves and displays the value of the specified variable.
// It handles both bracket and space syntax for variable specification.
func (c *GetCommand) Execute(args map[string]string, input string, _ neurotypes.Context) error {
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

	// Get variable service
	variableService, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}
	varService := variableService.(*services.VariableService)

	value, err := varService.Get(variable)
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
