package builtin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EchoJSONCommand implements the \echo-json command for pretty-printing JSON data.
// It provides professional JSON formatting for debugging and data visualization.
type EchoJSONCommand struct{}

// Name returns the command name "echo-json" for registration and lookup.
func (c *EchoJSONCommand) Name() string {
	return "echo-json"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *EchoJSONCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the echo-json command does.
func (c *EchoJSONCommand) Description() string {
	return "Pretty-print JSON data in readable format"
}

// Usage returns the syntax and usage examples for the echo-json command.
func (c *EchoJSONCommand) Usage() string {
	return "\\echo-json[to=var_name, indent=2] json_string"
}

// HelpInfo returns structured help information for the echo-json command.
func (c *EchoJSONCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "to",
				Description: "Variable name to store the formatted result",
				Required:    false,
				Type:        "string",
				Default:     "_output",
			},
			{
				Name:        "indent",
				Description: "Number of spaces for indentation (0 for compact)",
				Required:    false,
				Type:        "int",
				Default:     "2",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\echo-json ${_debug_network}",
				Description: "Pretty-print debug network data with default 2-space indentation",
			},
			{
				Command:     "\\echo-json[to=formatted] {\"key\": \"value\", \"nested\": {\"data\": 123}}",
				Description: "Format JSON and store in 'formatted' variable",
			},
			{
				Command:     "\\echo-json[indent=4] ${api_response}",
				Description: "Display API response with 4-space indentation",
			},
			{
				Command:     "\\echo-json[indent=0] {\"compact\": true}",
				Description: "Display compact JSON without indentation",
			},
			{
				Command:     "\\echo-json \"invalid json\"",
				Description: "Gracefully handle invalid JSON input",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Formatted JSON output (default storage location)",
				Type:        "command_output",
				Example:     "{\n  \"key\": \"value\"\n}",
			},
			{
				Name:        "{variable_name}",
				Description: "Custom variable when using to= option",
				Type:        "user_variable",
				Example:     "formatted_json = \"{\n  \"data\": 123\n}\"",
			},
		},
		Notes: []string{
			"Formats valid JSON with configurable indentation (default: 2 spaces)",
			"Use indent=0 for compact JSON output without formatting",
			"Gracefully handles invalid JSON by displaying error message",
			"Always displays formatted output to console",
			"Result is stored in the specified variable (default: _output)",
			"Useful for debugging API responses and network data",
			"Variable interpolation is handled before command execution",
		},
	}
}

// Execute formats JSON input and displays it in a pretty-printed format.
// Variable interpolation is handled by the state machine before this function is called.
// Options:
//   - to: store result in specified variable (default: ${_output})
//   - indent: number of spaces for indentation (default: 2, 0 for compact)
func (c *EchoJSONCommand) Execute(args map[string]string, input string) error {
	// Parse options with tolerant defaults
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	// Parse indent option with tolerant default
	indent := 2 // Default to 2-space indentation
	if indentStr := args["indent"]; indentStr != "" {
		if parsedIndent, err := strconv.Atoi(indentStr); err == nil && parsedIndent >= 0 {
			indent = parsedIndent
		}
		// If parsing fails or negative, indent remains at default (2)
	}

	// Create output printer with optional style injection
	var styleProvider output.StyleProvider
	if themeService, err := services.GetGlobalThemeService(); err == nil {
		styleProvider = themeService
	}
	printer := output.NewPrinter(output.WithStyles(styleProvider))

	// Handle empty input
	if strings.TrimSpace(input) == "" {
		errorMsg := "Error: No JSON data provided"
		printer.Error(errorMsg)
		c.storeResult(targetVar, errorMsg)
		return nil
	}

	// Try to parse and format the JSON
	formattedJSON, err := c.formatJSON(input, indent)
	if err != nil {
		// Handle invalid JSON gracefully
		errorMsg := fmt.Sprintf("Error: Invalid JSON - %s\nOriginal input: %s", err.Error(), input)
		printer.Error(errorMsg)
		c.storeResult(targetVar, errorMsg)
		return nil
	}

	// Display formatted JSON to console
	printer.Println(formattedJSON)

	// Store formatted result in variable
	c.storeResult(targetVar, formattedJSON)

	// Command never returns errors - it always succeeds
	return nil
}

// formatJSON attempts to parse and pretty-print JSON input with configurable indentation.
func (c *EchoJSONCommand) formatJSON(input string, indent int) (string, error) {
	// Trim whitespace from input
	input = strings.TrimSpace(input)

	// First, try to unmarshal to validate JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(input), &jsonData); err != nil {
		return "", err
	}

	// Handle compact format
	if indent == 0 {
		// Compact format - no indentation
		formattedBytes, err := json.Marshal(jsonData)
		if err != nil {
			return "", err
		}
		return string(formattedBytes), nil
	}

	// Create indentation string with specified number of spaces
	indentStr := strings.Repeat(" ", indent)

	// Marshal back with indentation for pretty printing
	formattedBytes, err := json.MarshalIndent(jsonData, "", indentStr)
	if err != nil {
		return "", err
	}

	return string(formattedBytes), nil
}

// storeResult stores the result in the specified variable.
func (c *EchoJSONCommand) storeResult(targetVar, result string) {
	// Get variable service - if not available, continue without storing (graceful degradation)
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		// Silently continue if variable service is not available
		return
	}

	// Store result in target variable
	if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
		// Store in system variable (only for specific system variables)
		_ = variableService.SetSystemVariable(targetVar, result)
	} else {
		// Store in user variable (including custom variables with _ prefix)
		_ = variableService.Set(targetVar, result)
	}
	// Ignore storage errors to ensure echo-json never fails
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EchoJSONCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register echo-json command: %v", err))
	}
}
