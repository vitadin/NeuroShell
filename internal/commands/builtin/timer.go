package builtin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// TimerCommand implements the \timer command for starting countdown timers.
// It uses the TemporalDisplayService to show a visual countdown timer.
type TimerCommand struct{}

// Name returns the command name "timer" for registration and lookup.
func (c *TimerCommand) Name() string {
	return "timer"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *TimerCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the timer command does.
func (c *TimerCommand) Description() string {
	return "Start a visual countdown timer for the specified number of seconds"
}

// Usage returns the syntax and usage examples for the timer command.
func (c *TimerCommand) Usage() string {
	return "\\timer seconds"
}

// HelpInfo returns structured help information for the timer command.
func (c *TimerCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for this command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\timer 5",
				Description: "Start a 5-second countdown timer",
			},
			{
				Command:     "\\timer 30",
				Description: "Start a 30-second countdown timer",
			},
			{
				Command:     "\\timer 100",
				Description: "Start a 100-second countdown timer (maximum allowed)",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Timer completion message",
				Type:        "command_output",
				Example:     "Timer completed!",
			},
		},
		Notes: []string{
			"Timer duration must be a positive number between 0 (exclusive) and 100 (inclusive) seconds",
			"The timer displays a visual countdown using the temporal display service",
			"Timer automatically disappears when completed and shows a completion message",
			"Only one timer per command execution - starting a new timer will replace any existing one",
		},
	}
}

// Execute starts a countdown timer for the specified duration.
// The timer shows a visual countdown and automatically disappears when completed.
func (c *TimerCommand) Execute(_ map[string]string, input string) error {
	// Parse the input as a number of seconds
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("timer duration is required")
	}

	// Convert input to float64 to handle both integers and decimals
	seconds, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid number", input)
	}

	// Validate the number is positive and within allowed range
	if seconds <= 0 {
		return fmt.Errorf("timer duration must be a positive number greater than 0")
	}

	if seconds > 100 {
		return fmt.Errorf("timer duration is too big (maximum: 100 seconds)")
	}

	// Get the temporal display service
	serviceInterface, err := services.GetGlobalRegistry().GetService("temporal-display")
	if err != nil {
		return fmt.Errorf("temporal display service not available: %w", err)
	}

	temporalService, ok := serviceInterface.(*services.TemporalDisplayService)
	if !ok {
		return fmt.Errorf("temporal display service not available")
	}

	// Convert seconds to duration
	duration := time.Duration(seconds * float64(time.Second))

	// Generate a unique timer ID based on current time to avoid conflicts
	timerID := fmt.Sprintf("timer_%d", time.Now().UnixNano())

	// Start the timer
	err = temporalService.StartTimer(timerID, duration)
	if err != nil {
		return fmt.Errorf("failed to start timer: %w", err)
	}

	// Start a goroutine to wait for timer completion and show completion message
	go c.waitForTimerCompletion(temporalService, timerID, duration, seconds)

	// Store success message in _output variable
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		message := fmt.Sprintf("Timer started for %.1f seconds", seconds)
		_ = variableService.SetSystemVariable("_output", message)
	}

	return nil
}

// waitForTimerCompletion waits for the timer to complete and shows a completion message.
func (c *TimerCommand) waitForTimerCompletion(temporalService *services.TemporalDisplayService, timerID string, duration time.Duration, originalSeconds float64) {
	// Wait for the timer duration plus a small buffer for cleanup
	time.Sleep(duration + 100*time.Millisecond)

	// Verify the timer is no longer active (should have auto-stopped)
	if !temporalService.IsActive(timerID) {
		// Create output printer with optional style injection
		var styleProvider output.StyleProvider
		if themeService, err := services.GetGlobalThemeService(); err == nil {
			styleProvider = themeService
		}
		printer := output.NewPrinter(output.WithStyles(styleProvider))

		// Show completion message
		printer.Info(fmt.Sprintf("Timer completed! (%.1f seconds)", originalSeconds))

		// Store completion message in _output variable
		if variableService, err := services.GetGlobalVariableService(); err == nil {
			message := "Timer completed!"
			_ = variableService.SetSystemVariable("_output", message)
		}
	}
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&TimerCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register timer command: %v", err))
	}
}
