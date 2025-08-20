package builtin

import (
	"fmt"
	"strings"
	"sync/atomic"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// Global counter for generating unique try IDs
var tryCounter int64

// TryCommand implements the \try command for error handling and error capture.
// It executes a command and captures any errors that occur, setting appropriate variables.
type TryCommand struct{}

// Name returns the command name "try" for registration and lookup.
func (c *TryCommand) Name() string {
	return "try"
}

// ParseMode returns ParseModeRaw since try commands need to preserve the entire message.
func (c *TryCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the try command does.
func (c *TryCommand) Description() string {
	return "Execute commands with error capture and handling"
}

// Usage returns the syntax and usage examples for the try command.
func (c *TryCommand) Usage() string {
	return "\\try command_to_execute"
}

// HelpInfo returns structured help information for the try command.
func (c *TryCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for try command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\try \\bash exit 1",
				Description: "Execute bash command and capture error if it fails",
			},
			{
				Command:     "\\try \\set[nonexistent=${undefined_var}]",
				Description: "Try to set a variable and capture any errors",
			},
			{
				Command:     "\\try \\echo This will succeed",
				Description: "Execute echo command (will succeed)",
			},
			{
				Command:     "\\try",
				Description: "Empty try command (sets success variables)",
			},
		},
		Notes: []string{
			"Captures errors and updates @status, @error system variables",
			"@status: '0' for success, '1' for failure",
			"@error: Error message if command failed, empty if succeeded",
			"@last_status/@last_error: Previous error state preserved",
			"_output: Command output (preserved from before failure)",
			"Try command itself never fails - it always captures errors",
			"Can be used to implement error handling and conditional logic",
		},
	}
}

// Execute executes the try command with error boundary markers.
// The try command captures errors from the target command and sets appropriate variables.
func (c *TryCommand) Execute(_ map[string]string, input string) error {
	// Extract target command from the input message
	targetCommand := strings.TrimSpace(input)

	// Get required services
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	if targetCommand == "" {
		// Empty try command - use error management service to set success state
		errorService, err := services.GetGlobalErrorManagementService()
		if err == nil {
			_ = errorService.SetErrorState("0", "")
		}
		// Note: @status/@error are computed system variables, no direct setting needed
		_ = variableService.SetSystemVariable("_output", "")
		return nil
	}

	// Generate unique try ID using atomic counter
	tryID := fmt.Sprintf("try_id_%d", atomic.AddInt64(&tryCounter, 1))

	// Push error boundary markers around target command (reverse order for LIFO)
	stackService.PushCommand("ERROR_BOUNDARY_END:" + tryID)
	stackService.PushCommand(targetCommand)
	stackService.PushCommand("ERROR_BOUNDARY_START:" + tryID)

	return nil
}

// IsReadOnly returns false as the try command modifies system state.
func (c *TryCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&TryCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register try command: %v", err))
	}
}
