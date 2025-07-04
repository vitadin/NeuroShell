package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/orchestration"
	"neuroshell/pkg/neurotypes"
)

// RunCommand implements the \run command for executing .neuro script files.
// It loads script files and executes their commands in sequence with variable interpolation.
type RunCommand struct{}

// Name returns the command name "run" for registration and lookup.
func (c *RunCommand) Name() string {
	return "run"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *RunCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the run command does.
func (c *RunCommand) Description() string {
	return "Execute a .neuro script file"
}

// Usage returns the syntax and usage examples for the run command.
func (c *RunCommand) Usage() string {
	return "\\run[file=\"script.neuro\"] or \\run script.neuro"
}

// HelpInfo returns structured help information for the run command.
func (c *RunCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\run script.neuro",
				Description: "Execute a .neuro script file",
			},
		},
	}
}

// Execute loads and runs a .neuro script file using the centralized script execution logic.
// It validates arguments and delegates to orchestration.ExecuteScript for the actual execution.
func (c *RunCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get filename from args or input
	filename := ""
	if fileArg, exists := args["file"]; exists && fileArg != "" {
		filename = fileArg
	} else if input != "" {
		filename = input
	} else {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Execute the script using centralized execution logic
	return orchestration.ExecuteScript(filename, ctx)
}

func init() {
	if err := commands.GlobalRegistry.Register(&RunCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register run command: %v", err))
	}
}
