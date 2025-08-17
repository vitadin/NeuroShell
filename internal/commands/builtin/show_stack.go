package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ShowStackCommand implements the \show-stack command for displaying the execution stack.
// It shows all stack elements vertically with clear marking of top and bottom positions.
type ShowStackCommand struct{}

// Name returns the command name "show-stack" for registration and lookup.
func (c *ShowStackCommand) Name() string {
	return "show-stack"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ShowStackCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the show-stack command does.
func (c *ShowStackCommand) Description() string {
	return "Display the execution stack for development and debugging"
}

// Usage returns the syntax and usage examples for the show-stack command.
func (c *ShowStackCommand) Usage() string {
	return "\\show-stack[detailed=true]"
}

// HelpInfo returns structured help information for the show-stack command.
func (c *ShowStackCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\show-stack[detailed=true]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "detailed",
				Description: "Show additional stack information including indices and context",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\show-stack",
				Description: "Display current execution stack",
			},
			{
				Command:     "\\show-stack[detailed=true]",
				Description: "Show stack with indices and try/silent block context",
			},
		},
		Notes: []string{
			"This is primarily a development and debugging tool",
			"Stack is typically empty in interactive mode since commands execute immediately",
			"Most useful scenarios:",
			"• Inside stdlib scripts that delegate to other commands",
			"• During .neuro batch file execution when commands are queued",
			"• Debugging try/silent block interactions",
			"• Understanding command delegation chains",
			"• Commands like \\model-new can push \\show-stack to show execution flow",
			"Stack uses LIFO (Last In, First Out) order - top item executes next",
		},
	}
}

// Execute runs the show-stack command to display the current execution stack.
func (c *ShowStackCommand) Execute(options map[string]string, _ string) error {
	// Get the stack service
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("failed to get stack service: %v", err)
	}

	// Get current stack
	stack := stackService.PeekStack()
	stackSize := stackService.GetStackSize()

	// Check for detailed output option
	detailed := false
	if detailedStr, exists := options["detailed"]; exists && (detailedStr == "true" || detailedStr == "1") {
		detailed = true
	}

	// Create printer for proper output formatting
	printer := printing.NewDefaultPrinter()

	// Display stack information
	if stackSize == 0 {
		printer.Info("Execution stack is empty")
		return nil
	}

	// Header
	if detailed {
		printer.Info(fmt.Sprintf("Execution Stack (Size: %d)", stackSize))
	} else {
		printer.Info("Execution Stack:")
	}

	// Display stack elements vertically
	for i, command := range stack {
		var marker string
		var formattedCmd string

		// Determine position marker and format command for display
		switch {
		case i == 0:
			marker = "TOP"
			if detailed {
				formattedCmd = fmt.Sprintf("[%s] %s", marker, command)
			} else {
				formattedCmd = fmt.Sprintf("🔝 %s", command)
			}
		case i == len(stack)-1:
			marker = "BOTTOM"
			if detailed {
				formattedCmd = fmt.Sprintf("[%s] %s", marker, command)
			} else {
				formattedCmd = fmt.Sprintf("🔻 %s", command)
			}
		default:
			marker = fmt.Sprintf("   %d", i)
			if detailed {
				formattedCmd = fmt.Sprintf("[%s] %s", marker, command)
			} else {
				formattedCmd = fmt.Sprintf("   %s", command)
			}
		}

		// Truncate very long commands for readability
		if len(formattedCmd) > 80 {
			formattedCmd = formattedCmd[:77] + "..."
		}

		// Color coding based on position
		switch {
		case i == 0:
			printer.Success(formattedCmd) // Top in green
		case i == len(stack)-1:
			printer.Warning(formattedCmd) // Bottom in yellow
		default:
			printer.Info(formattedCmd) // Middle in default
		}
	}

	// Additional information if detailed
	if detailed {
		printer.Info("\nStack operations: LIFO (Last In, First Out)")
		printer.Info(fmt.Sprintf("Next command to execute: %s", stack[0]))

		// Show context information
		if stackService.IsInTryBlock() {
			printer.Info(fmt.Sprintf("Currently in try block: %s (depth: %d)", stackService.GetCurrentTryID(), stackService.GetCurrentTryDepth()))
		}
		if stackService.IsInSilentBlock() {
			printer.Info(fmt.Sprintf("Currently in silent block: %s (depth: %d)", stackService.GetCurrentSilentID(), stackService.GetCurrentSilentDepth()))
		}
	}

	return nil
}

// init registers the ShowStackCommand with the global command registry.
func init() {
	if err := commands.GetGlobalRegistry().Register(&ShowStackCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register show-stack command: %v", err))
	}
}
