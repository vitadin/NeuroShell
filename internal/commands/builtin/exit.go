package builtin

import (
	"fmt"
	"os"
	"strconv"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ExitCommand implements the \exit command for terminating the NeuroShell session.
// It provides a clean way to exit the shell environment.
type ExitCommand struct{}

// Name returns the command name "exit" for registration and lookup.
func (c *ExitCommand) Name() string {
	return "exit"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ExitCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the exit command does.
func (c *ExitCommand) Description() string {
	return "Exit the shell with optional exit code and message"
}

// Usage returns the syntax and usage examples for the exit command.
func (c *ExitCommand) Usage() string {
	return "\\exit[code=N, message=text]"
}

// HelpInfo returns structured help information for the exit command.
func (c *ExitCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "code",
				Description: "Exit code to return to the operating system (0-255)",
				Required:    false,
				Type:        "int",
				Default:     "0",
			},
			{
				Name:        "message",
				Description: "Message to display before exiting",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\exit",
				Description: "Exit with code 0 (success), no message",
			},
			{
				Command:     "\\exit[code=1]",
				Description: "Exit with code 1 (error), no message",
			},
			{
				Command:     "\\exit[message=\"Goodbye!\"]",
				Description: "Exit with code 0, display farewell message",
			},
			{
				Command:     "\\exit[code=2, message=\"Configuration error\"]",
				Description: "Exit with code 2, display error message",
			},
			{
				Command:     "\\exit[message=\"Task completed successfully\"]",
				Description: "Exit with success message",
			},
		},
		Notes: []string{
			"Exits NeuroShell immediately after displaying message (if provided)",
			"Exit code defaults to 0 if not specified or invalid",
			"Exit codes should be 0-255; invalid codes default to 0",
			"Messages support variable interpolation (e.g., \"Task ${task_name} completed\")",
			"All unsaved session data will be lost",
			"Use Ctrl+C as an alternative exit method",
		},
	}
}

// Execute terminates the NeuroShell session with optional exit code and message.
// This provides an immediate exit from the shell environment.
func (c *ExitCommand) Execute(args map[string]string, _ string) error {
	// Parse exit code parameter with default 0
	exitCode := 0
	if codeStr := args["code"]; codeStr != "" {
		if parsedCode, err := strconv.Atoi(codeStr); err == nil {
			// Validate exit code range (0-255 is typical for most systems)
			if parsedCode >= 0 && parsedCode <= 255 {
				exitCode = parsedCode
			}
			// If code is out of range, exitCode remains 0 (tolerant default)
		}
		// If parsing fails, exitCode remains 0 (tolerant default)
	}

	// Parse message parameter and display if provided
	if message := args["message"]; message != "" {
		// Create output printer with optional style injection
		var styleProvider output.StyleProvider
		if themeService, err := services.GetGlobalThemeService(); err == nil {
			styleProvider = themeService
		}
		printer := output.NewPrinter(output.WithStyles(styleProvider))

		// Use error styling for non-zero exit codes, info styling for success
		if exitCode != 0 {
			printer.Error(message)
		} else {
			printer.Info(message)
		}
	}

	// Exit with the specified code
	os.Exit(exitCode)
	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&ExitCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register exit command: %v", err))
	}
}
