// Package neurotypes defines command system types for NeuroShell.
// This file contains the core types for command parsing, execution, and help system,
// including argument parsing modes and structured help information.
package neurotypes

// ParseMode defines how command arguments are parsed from user input.
type ParseMode int

const (
	// ParseModeKeyValue parses arguments as key=value pairs within brackets
	ParseModeKeyValue ParseMode = iota
	// ParseModeRaw treats the entire input as raw text without parsing
	ParseModeRaw
	// ParseModeWithOptions parses arguments with options support
	ParseModeWithOptions
)

// CommandArgs contains the parsed arguments and message content for command execution.
// It provides a structured way to pass user input to command implementations.
type CommandArgs struct {
	Options map[string]string
	Message string
}

// HelpInfo represents structured help information for a command.
// It provides rich help data that can be rendered in both plain text and styled formats.
type HelpInfo struct {
	Command         string               `json:"command"`                    // Command name
	Description     string               `json:"description"`                // Brief description of what the command does
	Usage           string               `json:"usage"`                      // Usage syntax
	ParseMode       ParseMode            `json:"parse_mode"`                 // How the command parses arguments
	Options         []HelpOption         `json:"options,omitempty"`          // Command options/parameters
	Examples        []HelpExample        `json:"examples,omitempty"`         // Usage examples
	StoredVariables []HelpStoredVariable `json:"stored_variables,omitempty"` // Variables automatically stored by this command
	Notes           []string             `json:"notes,omitempty"`            // Additional notes or warnings
}

// HelpOption represents a command option/parameter with detailed information.
type HelpOption struct {
	Name        string `json:"name"`              // Option name
	Description string `json:"description"`       // What this option does
	Required    bool   `json:"required"`          // Whether this option is required
	Type        string `json:"type"`              // Data type (string, bool, int, etc.)
	Default     string `json:"default,omitempty"` // Default value if not specified
}

// HelpExample represents a usage example with explanation.
type HelpExample struct {
	Command     string `json:"command"`     // Example command
	Description string `json:"description"` // What this example demonstrates
}

// HelpStoredVariable represents a variable that a command automatically stores.
type HelpStoredVariable struct {
	Name        string `json:"name"`              // Variable name (e.g., "_output", "#session_id")
	Description string `json:"description"`       // What this variable contains
	Type        string `json:"type"`              // Variable type category (e.g., "command_output", "system_metadata")
	Example     string `json:"example,omitempty"` // Optional example value
}
