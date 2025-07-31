// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ExportCommand implements the \session-export command for exporting sessions in various formats.
// It provides a general export interface that delegates to format-specific implementations.
type ExportCommand struct{}

// Name returns the command name "session-export" for registration and lookup.
func (c *ExportCommand) Name() string {
	return "session-export"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ExportCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-export command does.
func (c *ExportCommand) Description() string {
	return "Export chat session in specified format"
}

// Usage returns the syntax and usage examples for the session-export command.
func (c *ExportCommand) Usage() string {
	return `\session-export[format=json, file=path] session_identifier

Examples:
  \session-export[file=backup.json] work                        %% Export "work" session (defaults to JSON)
  \session-export[format=json, file=backup.json] project       %% Explicit JSON format
  \session-export[file=/tmp/session.json] 550e8400             %% Export by session ID
  \session-export[file=export.json] proj                       %% Export using prefix matching

Options:
  format - Export format: json (default: json)
  file   - Path to export file (required)

Note: Session identifier can be exact name, exact ID, or unique prefix.
      Currently only JSON format is supported, but other formats may be added.
      Command delegates to format-specific implementations.`
}

// HelpInfo returns structured help information for the session-export command.
func (c *ExportCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-export[format=json, file=path] session_identifier",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "format",
				Description: "Export format: json",
				Required:    false,
				Type:        "string",
				Default:     "json",
			},
			{
				Name:        "file",
				Description: "Path to export file",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-export[file=backup.json] work",
				Description: "Export session 'work' to JSON (default format)",
			},
			{
				Command:     "\\session-export[format=json, file=backup.json] project",
				Description: "Export session with explicit JSON format",
			},
			{
				Command:     "\\session-export[file=export.json] proj",
				Description: "Export session using prefix matching",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Export result message (delegated from format-specific command)",
				Type:        "command_output",
				Example:     "Exported session 'my-project' (ID: 550e8400) to backup.json",
			},
		},
		Notes: []string{
			"Currently only JSON format is supported",
			"Command delegates to format-specific implementations",
			"Session identifier can be exact name, exact ID, or unique prefix",
			"Future formats may include XML, YAML, or custom formats",
		},
	}
}

// Execute exports a chat session in the specified format.
// The input parameter specifies the session identifier (name, ID, or prefix).
// Options:
//   - format: export format (default: json)
//   - file: path to export file (required)
func (c *ExportCommand) Execute(args map[string]string, input string) error {

	// Get stack service for command delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Parse arguments
	format := args["format"]
	if format == "" {
		format = "json" // default format
	}
	filepath := args["file"]
	if filepath == "" {
		return fmt.Errorf("file path is required (use file=path)")
	}

	// Validate session identifier
	if input == "" {
		return fmt.Errorf("session identifier is required")
	}

	// Delegate to format-specific command
	var delegatedCommand string
	switch format {
	case "json":
		delegatedCommand = fmt.Sprintf("\\session-json-export[file=%s] %s", filepath, input)
	default:
		return fmt.Errorf("unsupported export format '%s'. Supported formats: json", format)
	}

	// Push the delegated command to the stack for execution
	stackService.PushCommand(delegatedCommand)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ExportCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-export command: %v", err))
	}
}
