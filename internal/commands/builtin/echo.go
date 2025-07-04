package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EchoCommand implements the \echo command for expanding variables and outputting text.
// It provides variable interpolation and flexible output options within the NeuroShell environment.
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
	return "Expand variables and output text with optional raw mode"
}

// Usage returns the syntax and usage examples for the echo command.
func (c *EchoCommand) Usage() string {
	return "\\echo [to=var_name] [silent=true] [raw=true] message"
}

// Execute expands variables in the input message and outputs or stores the result.
// Options:
//   - to: store result in specified variable (default: ${_output})
//   - silent: suppress console output (default: false)
//   - raw: output string literals without interpreting escape sequences (default: false)
func (c *EchoCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get variable service
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Interpolate variables in the input message
	expandedMessage, err := variableService.InterpolateString(input, ctx)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables: %w", err)
	}

	// Parse options
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	silentStr := args["silent"]
	silent := false
	if silentStr != "" {
		silent, err = strconv.ParseBool(silentStr)
		if err != nil {
			return fmt.Errorf("invalid value for silent option: %s (must be true or false)", silentStr)
		}
	}

	rawStr := args["raw"]
	raw := false
	if rawStr != "" {
		raw, err = strconv.ParseBool(rawStr)
		if err != nil {
			return fmt.Errorf("invalid value for raw option: %s (must be true or false)", rawStr)
		}
	}

	// Determine what to store and what to display
	var displayMessage string
	var storeMessage string

	if raw {
		// Raw mode: display and store without interpreting escape sequences
		displayMessage = expandedMessage
		storeMessage = expandedMessage
	} else {
		// Normal mode: interpret escape sequences for display, store interpreted version
		displayMessage = interpretEscapeSequences(expandedMessage)
		storeMessage = displayMessage
	}

	// Store result in target variable
	if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
		// Store in system variable (only for specific system variables)
		err = variableService.SetSystemVariable(targetVar, storeMessage, ctx)
	} else {
		// Store in user variable (including custom variables with _ prefix)
		err = variableService.Set(targetVar, storeMessage, ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to store result in variable '%s': %w", targetVar, err)
	}

	// Output to console unless silent mode is enabled
	if !silent {
		fmt.Print(displayMessage)
		// Only add newline if the message doesn't already end with one
		if len(displayMessage) > 0 && displayMessage[len(displayMessage)-1] != '\n' {
			fmt.Println()
		}
	}

	return nil
}

// getVariableService retrieves the variable service from the global registry
func (c *EchoCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
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
	if err := commands.GlobalRegistry.Register(&EchoCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register echo command: %v", err))
	}
}
