// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"path/filepath"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ImportCommand implements the \session-import command for importing sessions in various formats.
// It provides a general import interface that delegates to format-specific implementations.
type ImportCommand struct{}

// Name returns the command name "session-import" for registration and lookup.
func (c *ImportCommand) Name() string {
	return "session-import"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ImportCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-import command does.
func (c *ImportCommand) Description() string {
	return "Import chat session from file with format auto-detection"
}

// Usage returns the syntax and usage examples for the session-import command.
func (c *ImportCommand) Usage() string {
	return `\session-import[format=json, file=path]

Examples:
  \session-import[file=backup.json]                             %% Import from JSON (auto-detected by extension)
  \session-import[format=json, file=backup.json]               %% Explicit JSON format
  \session-import[file=/tmp/session.json]                      %% Import from absolute path
  \session-import[file=exports/session.json]                   %% Import from subdirectory

Options:
  format - Import format: json (auto-detected from file extension if not specified)
  file   - Path to import file (required)

Note: Format is auto-detected from file extension if not specified.
      Currently only JSON format (.json) is supported.
      Imported session gets new identity but preserves all content.`
}

// HelpInfo returns structured help information for the session-import command.
func (c *ImportCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-import[format=json, file=path]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "format",
				Description: "Import format: json (auto-detected from file extension)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "file",
				Description: "Path to import file",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-import[file=backup.json]",
				Description: "Import session from JSON with auto-detection",
			},
			{
				Command:     "\\session-import[format=json, file=backup.json]",
				Description: "Import session with explicit JSON format",
			},
			{
				Command:     "\\session-import[file=/tmp/session.json]",
				Description: "Import session from absolute file path",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Import result message (delegated from format-specific command)",
				Type:        "command_output",
				Example:     "Imported session as 'Session 1' (ID: 550e8400) from backup.json",
			},
		},
		Notes: []string{
			"Format is auto-detected from file extension if not specified",
			"Currently only JSON format (.json extension) is supported",
			"Command delegates to format-specific implementations",
			"Imported session gets completely new identity but preserves content",
		},
	}
}

// Execute imports a chat session from file with format auto-detection.
// Options:
//   - format: import format (auto-detected from file extension if not specified)
//   - file: path to import file (required)
func (c *ImportCommand) Execute(args map[string]string, _ string) error {

	// Get stack service for command delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Parse arguments
	format := args["format"]
	filepath := args["file"]
	if filepath == "" {
		return fmt.Errorf("file path is required (use file=path)")
	}

	// Auto-detect format from file extension if not specified
	if format == "" {
		format = c.detectFormatFromExtension(filepath)
		if format == "" {
			return fmt.Errorf("unable to auto-detect format from file extension. Supported extensions: .json")
		}
	}

	// Delegate to format-specific command
	var delegatedCommand string
	switch format {
	case "json":
		delegatedCommand = fmt.Sprintf("\\session-json-import[file=%s]", filepath)
	default:
		return fmt.Errorf("unsupported import format '%s'. Supported formats: json", format)
	}

	// Push the delegated command to the stack for execution
	stackService.PushCommand(delegatedCommand)

	return nil
}

// detectFormatFromExtension auto-detects import format from file extension.
// Returns empty string if format cannot be determined.
func (c *ImportCommand) detectFormatFromExtension(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	default:
		return ""
	}
}

// IsReadOnly returns false as the session-import command modifies system state.
func (c *ImportCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&ImportCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-import command: %v", err))
	}
}
