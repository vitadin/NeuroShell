package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EchoCommand implements the \echo command for outputting text.
// It provides flexible output options within the NeuroShell environment.
type EchoCommand struct{}

// Name returns the command name "echo" for registration and lookup.
func (c *EchoCommand) Name() string {
	return "echo"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *EchoCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the echo command does.
func (c *EchoCommand) Description() string {
	return "Output text with optional raw mode and variable storage"
}

// Usage returns the syntax and usage examples for the echo command.
func (c *EchoCommand) Usage() string {
	return "\\echo[to=var_name, silent=true, raw=true] message"
}

// HelpInfo returns structured help information for the echo command.
func (c *EchoCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\echo[to=var_name, silent=true, raw=true] message",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "to",
				Description: "Variable name to store the result",
				Required:    false,
				Type:        "string",
				Default:     "_output",
			},
			{
				Name:        "silent",
				Description: "Suppress console output",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
			{
				Name:        "raw",
				Description: "Output string literals without interpreting escape sequences",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\echo Hello, World!",
				Description: "Simple text output",
			},
			{
				Command:     "\\echo[raw=true] Line 1\\nLine 2",
				Description: "Raw output showing literal escape sequences",
			},
			{
				Command:     "\\echo[to=greeting] Hello ${name}!",
				Description: "Store interpolated result in 'greeting' variable",
			},
			{
				Command:     "\\echo[silent=true] Processing...",
				Description: "Store result without console output",
			},
			{
				Command:     "\\echo[to=formatted, raw=false] Tab:\\tNew line:\\n",
				Description: "Store formatted text with interpreted escape sequences",
			},
		},
		Notes: []string{
			"When raw=true, escape sequences like \\n are shown literally",
			"When raw=false, escape sequences are interpreted (\\n becomes newline)",
			"Result is always stored in the specified variable (default: _output)",
			"Use silent=true to suppress console output while still storing result",
			"Variable interpolation is handled by the state machine before echo executes",
		},
	}
}

// Execute outputs the input message and stores the result.
// Variable interpolation is handled by the state machine before this function is called.
// Options:
//   - to: store result in specified variable (default: ${_output})
//   - silent: suppress console output (default: false)
//   - raw: output string literals without interpreting escape sequences (default: false)
func (c *EchoCommand) Execute(args map[string]string, input string) error {
	// Parse options with tolerant defaults (never error)
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	// Parse silent option with tolerant default
	silent := false
	if silentStr := args["silent"]; silentStr != "" {
		if parsedSilent, err := strconv.ParseBool(silentStr); err == nil {
			silent = parsedSilent
		}
		// If parsing fails, silent remains false (tolerant default)
	}

	// Parse raw option with tolerant default
	raw := false
	if rawStr := args["raw"]; rawStr != "" {
		if parsedRaw, err := strconv.ParseBool(rawStr); err == nil {
			raw = parsedRaw
		}
		// If parsing fails, raw remains false (tolerant default)
	}

	// Determine what to store and what to display
	// Note: input comes pre-interpolated from state machine
	var displayMessage string
	var storeMessage string

	if raw {
		// Raw mode: display and store without interpreting escape sequences
		displayMessage = input
		storeMessage = input
	} else {
		// Normal mode: interpret escape sequences for display, store interpreted version
		displayMessage = interpretEscapeSequences(input)
		storeMessage = displayMessage
	}

	// Get variable service - if not available, continue without storing (graceful degradation)
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		// Store result in target variable
		if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
			// Store in system variable (only for specific system variables)
			_ = variableService.SetSystemVariable(targetVar, storeMessage)
		} else {
			// Store in user variable (including custom variables with _ prefix)
			_ = variableService.Set(targetVar, storeMessage)
		}
		// Ignore storage errors to ensure echo never fails
	}

	// Output to console unless silent mode is enabled
	if !silent {
		fmt.Print(displayMessage)
		// Only add newline if the message doesn't already end with one
		if len(displayMessage) > 0 && displayMessage[len(displayMessage)-1] != '\n' {
			fmt.Println()
		}
	}

	// Echo command never returns errors - it always succeeds
	return nil
}

// interpretEscapeSequences converts escape sequences in a string to their actual characters
func interpretEscapeSequences(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\'", "'")
	return s
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EchoCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register echo command: %v", err))
	}
}
