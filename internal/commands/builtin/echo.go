package builtin

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
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
	return "\\echo[to=var_name, silent=true, raw=true, display_only=false] message"
}

// HelpInfo returns structured help information for the echo command.
func (c *EchoCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\echo[to=var_name, silent=true, raw=true, display_only=false] message",
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
			{
				Name:        "display_only",
				Description: "Only display output without storing in default variable (see Notes for interaction with 'to' and 'silent')",
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
			{
				Command:     "\\echo[display_only=true] Debug: current value is ${value}",
				Description: "Display interpolated text without storing in any variable",
			},
			{
				Command:     "\\echo[display_only=true, raw=true] Literal:\\ntext",
				Description: "Display raw text without storage or escape interpretation",
			},
			{
				Command:     "\\echo[display_only=true, to=temp] Store but display only",
				Description: "Display text and store in 'temp' variable (display_only with 'to' allows storage)",
			},
			{
				Command:     "\\echo[display_only=true, silent=true] Nothing happens",
				Description: "Neither display nor store (display_only + silent = no action)",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Command output (default storage location, not used when display_only=true)",
				Type:        "command_output",
				Example:     "Hello, World!",
			},
			{
				Name:        "{variable_name}",
				Description: "Custom variable when using to= option (not used when display_only=true)",
				Type:        "user_variable",
				Example:     "greeting = \"Hello ${name}!\"",
			},
		},
		Notes: []string{
			"When raw=true, escape sequences like \\n are shown literally",
			"When raw=false, escape sequences are interpreted (\\n becomes newline)",
			"Result is stored in the specified variable (default: _output) based on option combinations:",
			"  • Normal behavior: display + store in variable",
			"  • silent=true: no display + store in variable",
			"  • display_only=true: display + no storage (unless 'to' is specified)",
			"  • display_only=true + to=var: display + store in specified variable",
			"  • display_only=true + silent=true: no display + no storage",
			"Variable interpolation is handled by the state machine before echo executes",
			"Invalid option values are ignored gracefully (echo never fails)",
		},
	}
}

// Execute outputs the input message and stores the result.
// Variable interpolation is handled by the state machine before this function is called.
// Options:
//   - to: store result in specified variable (default: ${_output})
//   - silent: suppress console output (default: false)
//   - raw: output string literals without interpreting escape sequences (default: false)
//   - display_only: only display without storing in any variable (default: false)
func (c *EchoCommand) Execute(args map[string]string, input string) error {
	// Parse display_only option with tolerant default
	displayOnly := false
	if displayOnlyStr := args["display_only"]; displayOnlyStr != "" {
		if parsedDisplayOnly, err := strconv.ParseBool(displayOnlyStr); err == nil {
			displayOnly = parsedDisplayOnly
		}
		// If parsing fails, displayOnly remains false (tolerant default)
	}

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

	// Determine storage behavior based on option combinations
	shouldStore := true
	if displayOnly && silent {
		// display_only=true AND silent=true: no console output and no storage
		shouldStore = false
	} else if displayOnly && args["to"] == "" {
		// display_only=true with no 'to' option: don't store anywhere
		shouldStore = false
	}
	// If display_only=true AND 'to' is specified (but not silent): still store to the specified variable

	// Store result in variable based on computed behavior
	if shouldStore {
		// Get variable service - if not available, continue without storing (graceful degradation)
		if variableService, err := services.GetGlobalVariableService(); err == nil {
			// Store result in target variable
			if targetVar == "_output" {
				// Store in system variable (only for specific system variables)
				_ = variableService.SetSystemVariable(targetVar, storeMessage)
			} else {
				// Store in user variable (including custom variables with _ prefix)
				_ = variableService.Set(targetVar, storeMessage)
			}
			// Ignore storage errors to ensure echo never fails
		}
	}

	// Determine display behavior based on option combinations
	shouldDisplay := !silent
	if displayOnly && silent {
		// display_only=true AND silent=true: no console output and no storage (above)
		shouldDisplay = false
	}

	// Output to console based on computed behavior
	if shouldDisplay {
		// Create output printer with optional style injection
		var styleProvider output.StyleProvider
		if themeService, err := services.GetGlobalThemeService(); err == nil {
			styleProvider = themeService // ThemeService implements StyleProvider
		}

		printer := output.NewPrinter(output.WithStyles(styleProvider))

		// Use Print method which handles newlines appropriately
		if displayMessage != "" {
			if len(displayMessage) > 0 && displayMessage[len(displayMessage)-1] == '\n' {
				// Message already has newline, use Print to avoid double newline
				printer.Print(displayMessage)
			} else {
				// No newline, use Println to add one
				printer.Println(displayMessage)
			}
		}
		// For empty strings, print nothing (maintain original behavior)
	}

	// Echo command never returns errors - it always succeeds with graceful degradation
	return nil
}

// interpretEscapeSequences converts escape sequences in a string to their actual characters
// using Go's built-in strconv.Unquote for robust handling.
func interpretEscapeSequences(s string) string {
	// Use Go's built-in strconv.Unquote which properly handles all escape sequences
	// We need to wrap the string in quotes for strconv.Unquote to work
	quoted := `"` + s + `"`

	// strconv.Unquote handles all standard Go escape sequences correctly
	if unquoted, err := strconv.Unquote(quoted); err == nil {
		return unquoted
	}

	// If unquoting fails (e.g., malformed escape sequences), return original string
	return s
}

// IsReadOnly returns true as the echo command doesn't modify system state.
func (c *EchoCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EchoCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register echo command: %v", err))
	}
}
