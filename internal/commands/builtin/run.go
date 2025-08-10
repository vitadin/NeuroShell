// Package builtin provides core built-in commands for NeuroShell.
package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// RunCommand implements the \run command for executing NeuroShell script files.
// It pushes the script path to the stack service for execution by the state machine.
type RunCommand struct{}

// Name returns the command name "run" for registration and lookup.
func (c *RunCommand) Name() string {
	return "run"
}

// ParseMode returns ParseModeRaw to pass the script path directly.
func (c *RunCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the run command does.
func (c *RunCommand) Description() string {
	return "Execute a NeuroShell script file"
}

// Usage returns the syntax and usage examples for the run command.
func (c *RunCommand) Usage() string {
	return `\run script_path

Examples:
  \run setup.neuro                    %% Execute script in current directory
  \run /path/to/script.neuro         %% Execute script with absolute path  
  \run ../config/init.neuro          %% Execute script with relative path
  \try \run potentially-failing.neuro %% Execute with error handling

Notes:
  - Script path is required and must point to a .neuro file
  - Both absolute and relative paths are supported
  - Script is executed using the same state machine as batch mode
  - Use \try \run for non-blocking execution that handles errors gracefully`
}

// HelpInfo returns structured help information for the run command.
func (c *RunCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\run script_path",
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\run setup.neuro",
				Description: "Execute script in current directory",
			},
			{
				Command:     "\\run /path/to/script.neuro",
				Description: "Execute script with absolute path",
			},
			{
				Command:     "\\run ../config/init.neuro",
				Description: "Execute script with relative path",
			},
			{
				Command:     "\\try \\run potentially-failing.neuro",
				Description: "Execute script with error handling",
			},
		},
		Notes: []string{
			"Script path is required and must point to a .neuro file",
			"Both absolute and relative paths are supported",
			"Script is executed using the same state machine as batch mode",
			"All path resolution and security checks handled by state machine",
			"Use \\try \\run for non-blocking execution with error handling",
			"Variables are interpolated in the script path parameter",
		},
	}
}

// Execute pushes the script path to the stack service for execution.
func (c *RunCommand) Execute(_ map[string]string, input string) error {
	// Validate script path parameter
	scriptPath := strings.TrimSpace(input)
	if scriptPath == "" {
		return fmt.Errorf("script path is required\n\nUsage: %s", c.Usage())
	}

	// Get stack service
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Push script path with backslash prefix to avoid parser treating it as echo message
	// The state machine will recognize the .neuro suffix and handle execution
	stackService.PushCommand("\\" + scriptPath)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&RunCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register run command: %v", err))
	}
}
