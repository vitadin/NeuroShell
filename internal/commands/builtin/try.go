package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

// TryCommand implements the \try command for error handling and execution control.
// It executes another command and captures any errors without terminating the shell session.
// Similar to Stata's capture command, it allows graceful error handling in scripts.
type TryCommand struct{}

// Name returns the command name "try" for registration and lookup.
func (c *TryCommand) Name() string {
	return "try"
}

// ParseMode returns ParseModeRaw to pass the entire command line to the try handler.
func (c *TryCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the try command does.
func (c *TryCommand) Description() string {
	return "Execute a command and capture errors without terminating the session"
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
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\try \\bash ls /nonexistent",
				Description: "Try to list a non-existent directory, capture error",
			},
			{
				Command:     "\\try \\set[invalid_var] value",
				Description: "Try to set an invalid variable, capture validation error",
			},
			{
				Command:     "\\try \\send test message",
				Description: "Try to send a message, capture any API errors",
			},
			{
				Command:     "\\try \\bash echo \"success\"",
				Description: "Execute a successful command, set status to 0",
			},
		},
		Notes: []string{
			"Executes the specified command and captures any errors",
			"Sets ${_status} to 0 for success, 1 for failure",
			"Sets ${_error} to the error message if command fails",
			"Sets ${_output} to command output if available",
			"Prevents command errors from terminating the shell session",
			"Essential for error handling in scripts and conditional execution",
		},
	}
}

// Execute runs the specified command and captures any errors without terminating the session.
func (c *TryCommand) Execute(_ map[string]string, input string) error {
	// Parse the command to execute
	commandToExecute := strings.TrimSpace(input)
	if commandToExecute == "" {
		// If no command provided, set success status and return
		// This aligns with the principle that \try never causes errors
		globalCtx := neuroshellcontext.GetGlobalContext()
		if neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext); ok {
			_ = neuroCtx.SetSystemVariable("_status", "0")
			_ = neuroCtx.SetSystemVariable("_error", "")
		}
		return nil
	}

	// Skip comment lines (same logic as shell handler and script service)
	if strings.HasPrefix(commandToExecute, "%%") {
		// Comments are treated as successful no-ops
		globalCtx := neuroshellcontext.GetGlobalContext()
		if neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext); ok {
			_ = neuroCtx.SetSystemVariable("_status", "0")
			_ = neuroCtx.SetSystemVariable("_error", "")
			_ = neuroCtx.SetSystemVariable("_output", "")
		}
		return nil
	}

	// Initialize system variables for try command results
	globalCtx := neuroshellcontext.GetGlobalContext()
	var neuroCtx *neuroshellcontext.NeuroContext
	var ok bool

	if neuroCtx, ok = globalCtx.(*neuroshellcontext.NeuroContext); !ok {
		return fmt.Errorf("unable to access global context for variable setting")
	}

	// Parse the command to execute
	parsedCmd := parser.ParseInput(commandToExecute)
	if parsedCmd == nil {
		// Set error variables
		_ = neuroCtx.SetSystemVariable("_status", "1")
		_ = neuroCtx.SetSystemVariable("_error", "failed to parse command")
		_ = neuroCtx.SetSystemVariable("_output", "")
		return nil // Don't return error - that's the point of try
	}

	// Get the global command registry
	registry := commands.GetGlobalRegistry()

	// Check if the command exists
	if !registry.IsValidCommand(parsedCmd.Name) {
		// Set error variables
		_ = neuroCtx.SetSystemVariable("_status", "1")
		_ = neuroCtx.SetSystemVariable("_error", fmt.Sprintf("unknown command: %s", parsedCmd.Name))
		_ = neuroCtx.SetSystemVariable("_output", "")
		return nil // Don't return error - capture it instead
	}

	// Get the command's parse mode to properly handle its arguments
	cmd, _ := registry.Get(parsedCmd.Name)
	parseMode := cmd.ParseMode()

	var args map[string]string
	var message string

	// Handle arguments based on the target command's parse mode
	switch parseMode {
	case neurotypes.ParseModeRaw:
		// For raw mode commands, pass the entire message
		args = make(map[string]string)
		message = parsedCmd.Message
	case neurotypes.ParseModeKeyValue:
		// For key-value mode commands, parse the bracket content as options
		args = parseBracketContent(parsedCmd.BracketContent)
		message = parsedCmd.Message
	default:
		// Default to key-value parsing
		args = parseBracketContent(parsedCmd.BracketContent)
		message = parsedCmd.Message
	}

	// Execute the command and capture any errors
	err := registry.Execute(parsedCmd.Name, args, message)

	if err != nil {
		// Command execution failed at the Go level (e.g., command not found, service error)
		_ = neuroCtx.SetSystemVariable("_status", "1")
		_ = neuroCtx.SetSystemVariable("_error", err.Error())
		// Don't modify _output here - let the failed command's output stand
	} else {
		// Command executed successfully at the Go level
		// But we need to check if the command itself succeeded
		// by examining the _status variable set by the command
		status, statusErr := neuroCtx.GetVariable("_status")
		switch {
		case statusErr != nil || status == "":
			// If no status was set, assume success
			_ = neuroCtx.SetSystemVariable("_status", "0")
			_ = neuroCtx.SetSystemVariable("_error", "")
		case status != "0":
			// Command set a non-zero status, which indicates failure
			// The error message should already be set by the command in _error
			// No action needed - preserve the error state
		default:
			// Command set status to 0, which indicates success
			_ = neuroCtx.SetSystemVariable("_error", "")
		}
	}

	// Always return nil - the whole point of try is to not propagate errors
	return nil
}

// parseBracketContent parses the content within brackets into key-value pairs.
// This is a simplified version that handles basic key=value syntax.
func parseBracketContent(content string) map[string]string {
	args := make(map[string]string)

	if content == "" {
		return args
	}

	// Simple parsing for key=value pairs separated by commas
	// This could be enhanced to handle more complex syntax if needed
	pairs := strings.Split(content, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Check if it's a key=value pair
		if parts := strings.SplitN(pair, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			args[key] = value
		} else {
			// Treat as a flag (key with empty value)
			args[pair] = ""
		}
	}

	return args
}

func init() {
	if err := commands.GlobalRegistry.Register(&TryCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register try command: %v", err))
	}
}
