package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// GetEnvCommand implements the \get-env command for retrieving environment variable values.
// It supports both bracket syntax (\get-env[VAR]) and space syntax (\get-env VAR).
// Automatically creates #os.VAR neuro variable with the retrieved value.
type GetEnvCommand struct{}

// Name returns the command name "get-env" for registration and lookup.
func (c *GetEnvCommand) Name() string {
	return "get-env"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *GetEnvCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the get-env command does.
func (c *GetEnvCommand) Description() string {
	return "Get an environment variable and create #os.VAR neuro variable"
}

// Usage returns the syntax and usage examples for the get-env command.
func (c *GetEnvCommand) Usage() string {
	return "\\get-env[VAR] or \\get-env VAR"
}

// HelpInfo returns structured help information for the get-env command.
func (c *GetEnvCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\get-env[VAR] or \\get-env VAR",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "VAR",
				Description: "Environment variable name to retrieve",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\get-env[HOME]",
				Description: "Get HOME environment variable using bracket syntax and create #os.HOME",
			},
			{
				Command:     "\\get-env PATH",
				Description: "Get PATH environment variable using space syntax and create #os.PATH",
			},
			{
				Command:     "\\get-env[OPENAI_API_KEY]",
				Description: "Get OPENAI_API_KEY environment variable and create #os.OPENAI_API_KEY",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#os.VAR",
				Description: "Contains the retrieved environment variable value (VAR = actual variable name)",
				Type:        "environment_mirror",
				Example:     "#os.HOME = \"/Users/username\"",
			},
		},
		Notes: []string{
			"Supports both bracket syntax (\\get-env[VAR]) and space syntax (\\get-env VAR)",
			"In test mode, retrieves test environment overrides for clean testing",
			"In production mode, retrieves actual OS environment variables",
			"Returns empty string if environment variable doesn't exist",
		},
	}
}

// Execute retrieves and displays the value of the specified environment variable.
// It handles both bracket and space syntax and automatically creates #os.VAR neuro variable.
func (c *GetEnvCommand) Execute(args map[string]string, input string) error {
	var variable string

	// Handle bracket syntax: \get-env[VAR]
	if len(args) > 0 {
		for key := range args {
			variable = key
			break
		}
	} else if input != "" {
		// Handle space syntax: \get-env VAR
		fields := strings.Fields(input)
		if len(fields) > 0 {
			variable = fields[0]
		}
	}

	if variable == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get variable service from global registry
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get environment variable value
	value, err := variableService.GetEnvVariable(variable)
	if err != nil {
		return fmt.Errorf("failed to get environment variable %s: %w", variable, err)
	}

	// Automatically create #os.VAR neuro variable with the retrieved value
	neuroVarName := "#os." + variable
	err = variableService.SetSystemVariable(neuroVarName, value)
	if err != nil {
		return fmt.Errorf("failed to set neuro variable %s: %w", neuroVarName, err)
	}

	// Display the result (matching \get command format)
	printer := c.createPrinter()
	printer.Pair(variable, value)
	return nil
}

// createPrinter creates a printer with theme service as style provider
func (c *GetEnvCommand) createPrinter() *output.Printer {
	// Try to get theme service as style provider
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// Fall back to plain style provider
		return output.NewPrinter(output.WithStyles(output.NewPlainStyleProvider()))
	}

	return output.NewPrinter(output.WithStyles(themeService))
}

func init() {
	if err := commands.GlobalRegistry.Register(&GetEnvCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register get-env command: %v", err))
	}
}
