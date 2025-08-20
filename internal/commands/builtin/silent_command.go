package builtin

import (
	"fmt"
	"strings"
	"sync/atomic"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// Global counter for generating unique silent IDs
var silentCounter int64

// SilentCommand implements the \silent command for output suppression.
// It executes a command while suppressing all stdout output, preserving stderr for errors.
type SilentCommand struct{}

// Name returns the command name "silent" for registration and lookup.
func (c *SilentCommand) Name() string {
	return "silent"
}

// ParseMode returns ParseModeRaw since silent commands need to preserve the entire message.
func (c *SilentCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the silent command does.
func (c *SilentCommand) Description() string {
	return "Execute commands with stdout output suppressed"
}

// Usage returns the syntax and usage examples for the silent command.
func (c *SilentCommand) Usage() string {
	return "\\silent command_to_execute"
}

// HelpInfo returns structured help information for the silent command.
func (c *SilentCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for silent command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\silent \\echo Hello World",
				Description: "Execute echo command without printing output to console",
			},
			{
				Command:     "\\silent \\bash ls -la",
				Description: "Execute bash command silently (no stdout, stderr preserved)",
			},
			{
				Command:     "\\silent \\set[var=value]",
				Description: "Set variable silently without confirmation message",
			},
			{
				Command:     "\\silent \\model-activate my-model",
				Description: "Activate model without printing activation message",
			},
			{
				Command:     "\\silent",
				Description: "Empty silent command (does nothing silently)",
			},
		},
		Notes: []string{
			"Suppresses all stdout output (fmt.Print*, fmt.Println, etc.)",
			"Preserves stderr for error messages - errors remain visible",
			"Variables are still set normally - only console output is suppressed",
			"Command execution behavior is unchanged, only output visibility",
			"Nested silent commands are supported",
			"Can be used to reduce console noise in scripts and automated workflows",
		},
	}
}

// Execute executes the silent command with silent boundary markers.
// The silent command suppresses stdout output from the target command.
func (c *SilentCommand) Execute(_ map[string]string, input string) error {
	// Extract target command from the input message
	targetCommand := strings.TrimSpace(input)

	// Get required services
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	if targetCommand == "" {
		// Empty silent command - do nothing silently
		return nil
	}

	// Generate unique silent ID using atomic counter
	silentID := fmt.Sprintf("silent_id_%d", atomic.AddInt64(&silentCounter, 1))

	// Push silent boundary markers around target command (reverse order for LIFO)
	stackService.PushCommand("SILENT_BOUNDARY_END:" + silentID)
	stackService.PushCommand(targetCommand)
	stackService.PushCommand("SILENT_BOUNDARY_START:" + silentID)

	return nil
}

// IsReadOnly returns false as the silent command modifies system state.
func (c *SilentCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&SilentCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register silent command: %v", err))
	}
}
