// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// JSONExportCommand implements the \session-json-export command for exporting sessions to JSON.
// It provides JSON export functionality for chat sessions with automatic session resolution.
type JSONExportCommand struct{}

// Name returns the command name "session-json-export" for registration and lookup.
func (c *JSONExportCommand) Name() string {
	return "session-json-export"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *JSONExportCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-json-export command does.
func (c *JSONExportCommand) Description() string {
	return "Export chat session to JSON file"
}

// Usage returns the syntax and usage examples for the session-json-export command.
func (c *JSONExportCommand) Usage() string {
	return `\session-json-export[file=path] session_identifier

Examples:
  \session-json-export[file=backup.json] work                    %% Export "work" session to backup.json
  \session-json-export[file=/tmp/session.json] project          %% Export with absolute path
  \session-json-export[file=sessions/backup.json] 550e8400      %% Export by session ID
  \session-json-export[file=export.json] proj                   %% Export using prefix matching

Options:
  file - Path to JSON export file (required)

Note: Session identifier can be exact name, exact ID, or unique prefix.
      Uses the same session resolution as other session commands.
      Exported JSON contains complete session data including messages.`
}

// HelpInfo returns structured help information for the session-json-export command.
func (c *JSONExportCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-json-export[file=path] session_identifier",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "file",
				Description: "Path to JSON export file",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-json-export[file=backup.json] work",
				Description: "Export session named 'work' to backup.json",
			},
			{
				Command:     "\\session-json-export[file=/tmp/session.json] 550e8400",
				Description: "Export session by ID to absolute path",
			},
			{
				Command:     "\\session-json-export[file=export.json] proj",
				Description: "Export session using prefix matching",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Export result message with file path and session details",
				Type:        "command_output",
				Example:     "Exported session 'my-project' (ID: 550e8400) to backup.json",
			},
		},
		Notes: []string{
			"Session identifier can be exact name, exact ID, or unique prefix",
			"Uses same session resolution logic as other session commands",
			"Exported JSON includes complete session data with messages",
			"File path can be relative or absolute",
			"Creates parent directories if they don't exist",
		},
	}
}

// Execute exports a chat session to JSON file.
// The input parameter specifies the session identifier (name, ID, or prefix).
// Options:
//   - file: path to JSON export file (required)
func (c *JSONExportCommand) Execute(args map[string]string, input string) error {

	// Get chat session service
	chatService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Validate required arguments
	filepath := args["file"]
	if filepath == "" {
		return fmt.Errorf("file path is required (use file=path)")
	}

	// Validate session identifier
	if input == "" {
		return fmt.Errorf("session identifier is required")
	}

	// Find session using prefix matching (follows existing pattern)
	session, err := chatService.FindSessionByPrefix(input)
	if err != nil {
		return fmt.Errorf("session lookup failed: %w", err)
	}

	// Export session to JSON file
	if err := chatService.ExportSessionToJSON(session.ID, filepath); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Prepare output message
	shortID := session.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	outputMsg := fmt.Sprintf("Exported session '%s' (ID: %s) to %s", session.Name, shortID, filepath)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&JSONExportCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-json-export command: %v", err))
	}
}
