package builtin

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SetEnvCommand implements the \set-env command for setting environment variable values.
// It supports both bracket syntax (\set-env[VAR=value]) and space syntax (\set-env VAR value).
type SetEnvCommand struct{}

// Name returns the command name "set-env" for registration and lookup.
func (c *SetEnvCommand) Name() string {
	return "set-env"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SetEnvCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the set-env command does.
func (c *SetEnvCommand) Description() string {
	return "Set an environment variable"
}

// Usage returns the syntax and usage examples for the set-env command.
func (c *SetEnvCommand) Usage() string {
	return "\\set-env[VAR=value] or \\set-env VAR value"
}

// HelpInfo returns structured help information for the set-env command.
func (c *SetEnvCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\set-env[VAR=value] or \\set-env VAR value",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "VAR",
				Description: "Environment variable name to set",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "value",
				Description: "Value to assign to the environment variable",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\set-env[OPENAI_API_KEY=sk-1234567890abcdef]",
				Description: "Set OPENAI_API_KEY environment variable using bracket syntax",
			},
			{
				Command:     "\\set-env PYTHONPATH /usr/local/lib/python3.9",
				Description: "Set PYTHONPATH environment variable using space syntax",
			},
			{
				Command:     "\\set-env[MY_VAR=\"Hello World\"]",
				Description: "Set environment variable with quoted value containing spaces",
			},
		},
		Notes: []string{
			"Supports both bracket syntax (\\set-env[VAR=value]) and space syntax (\\set-env VAR value)",
			"Environment variable values can contain spaces when properly quoted",
			"In test mode, sets test environment overrides for clean testing",
			"In production mode, sets actual OS environment variables",
			"Environment variables persist for the duration of the shell session",
		},
	}
}

// Execute sets environment variable values using either bracket or space syntax.
// It handles multiple variable assignments and provides confirmation output.
func (c *SetEnvCommand) Execute(args map[string]string, input string) error {
	if len(args) == 0 && input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get variable service
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Handle bracket syntax: \set-env[VAR=value]
	if len(args) > 0 {
		// Sort keys to ensure deterministic output order
		var keys []string
		for key := range args {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := args[key]
			err := variableService.SetEnvVariable(key, value)
			if err != nil {
				return fmt.Errorf("failed to set environment variable %s: %w", key, err)
			}
			fmt.Printf("Setting %s = %s\n", key, value)
		}
		return nil
	}

	// Handle space syntax: \set-env VAR value
	if input != "" {
		// Trim only leading whitespace, then split
		trimmedInput := strings.TrimLeft(input, " \t")
		var varName, value string
		if trimmedInput != "" {
			parts := strings.SplitN(trimmedInput, " ", 2)
			if len(parts) == 2 {
				varName, value = parts[0], strings.TrimLeft(parts[1], " \t")
			} else {
				varName, value = parts[0], ""
			}

			err := variableService.SetEnvVariable(varName, value)
			if err != nil {
				return fmt.Errorf("failed to set environment variable %s: %w", varName, err)
			}
			fmt.Printf("Setting %s = %s\n", varName, value)
		}
		return nil
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&SetEnvCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register set-env command: %v", err))
	}
}
