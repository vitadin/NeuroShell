// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"encoding/json"
	"fmt"
	"os"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// JSONImportCommand implements the \session-json-import command for importing sessions from JSON.
// It provides JSON import functionality with automatic session reconstruction and activation.
type JSONImportCommand struct{}

// Name returns the command name "session-json-import" for registration and lookup.
func (c *JSONImportCommand) Name() string {
	return "session-json-import"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *JSONImportCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-json-import command does.
func (c *JSONImportCommand) Description() string {
	return "Import chat session from JSON file"
}

// Usage returns the syntax and usage examples for the session-json-import command.
func (c *JSONImportCommand) Usage() string {
	return `\session-json-import[file=path]

Examples:
  \session-json-import[file=backup.json]                        %% Import session from backup.json
  \session-json-import[file=/tmp/session.json]                 %% Import from absolute path
  \session-json-import[file=sessions/export.json]              %% Import from subdirectory

Options:
  file - Path to JSON import file (required)

Note: Imported session gets a new ID, auto-generated name, and current timestamps.
      All messages and system prompt are preserved from the original session.
      The imported session automatically becomes the active session.`
}

// HelpInfo returns structured help information for the session-json-import command.
func (c *JSONImportCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-json-import[file=path]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "file",
				Description: "Path to JSON import file",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-json-import[file=backup.json]",
				Description: "Import session from backup.json with auto-generated name",
			},
			{
				Command:     "\\session-json-import[file=/tmp/session.json]",
				Description: "Import session from absolute file path",
			},
			{
				Command:     "\\session-json-import[file=exports/work.json]",
				Description: "Import session from subdirectory",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#session_id",
				Description: "New unique identifier of the imported session (in memory)",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "New auto-generated name of the imported session (in memory)",
				Type:        "system_metadata",
				Example:     "Session 1",
			},
			{
				Name:        "#session_original_id",
				Description: "Original session ID from the imported file",
				Type:        "system_metadata",
				Example:     "abc12345-678d-90ef",
			},
			{
				Name:        "#session_original_name",
				Description: "Original session name from the imported file",
				Type:        "system_metadata",
				Example:     "my-project",
			},
			{
				Name:        "#message_count",
				Description: "Number of messages preserved from the original session",
				Type:        "system_metadata",
				Example:     "5",
			},
			{
				Name:        "#system_prompt",
				Description: "System prompt preserved from the original session",
				Type:        "system_metadata",
				Example:     "You are a helpful assistant",
			},
			{
				Name:        "#session_created",
				Description: "New creation timestamp (when session was imported)",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#session_original_created",
				Description: "Original creation timestamp from the imported file",
				Type:        "system_metadata",
				Example:     "2024-01-10 09:15:42",
			},
			{
				Name:        "_output",
				Description: "Import result message with new and original session details",
				Type:        "command_output",
				Example:     "Imported session as 'Session 1' (ID: 550e8400) from backup.json (original: 'my-project')",
			},
		},
		Notes: []string{
			"Imported session gets completely new identity (ID, name, timestamps)",
			"All conversation messages are preserved with original timestamps",
			"System prompt is preserved from the original session",
			"Imported session automatically becomes active session",
			"Session name is auto-generated using default naming convention",
		},
	}
}

// Execute imports a chat session from JSON file.
// Options:
//   - file: path to JSON import file (required)
func (c *JSONImportCommand) Execute(args map[string]string, _ string) error {

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

	// Read and parse the original session data first (before import reconstructs it)
	jsonData, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var originalSession neurotypes.ChatSession
	if err := json.Unmarshal(jsonData, &originalSession); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Import session from JSON file (this reconstructs with new identity)
	newSession, err := chatService.ImportSessionFromJSON(filepath)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Auto-push session activation command to stack service for seamless UX
	// The import process already activates the session, but this ensures consistency
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := fmt.Sprintf("\\silent \\session-activate[id=true] %s", newSession.ID)
		stackService.PushCommand(activateCommand)
	}

	// Update session-related variables (both new and original)
	if err := c.updateSessionVariables(newSession, &originalSession, variableService); err != nil {
		return fmt.Errorf("failed to update session variables: %w", err)
	}

	// Prepare output message
	shortID := newSession.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	outputMsg := fmt.Sprintf("Imported session as '%s' (ID: %s) from %s (original: '%s')", newSession.Name, shortID, filepath, originalSession.Name)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	return nil
}

// updateSessionVariables sets session-related system variables including original session data
func (c *JSONImportCommand) updateSessionVariables(newSession *neurotypes.ChatSession, originalSession *neurotypes.ChatSession, variableService *services.VariableService) error {
	// Set session variables (both new and original)
	variables := map[string]string{
		// New session variables (in memory)
		"#session_id":      newSession.ID,
		"#session_name":    newSession.Name,
		"#session_created": newSession.CreatedAt.Format("2006-01-02 15:04:05"),

		// Original session variables (from file)
		"#session_original_id":      originalSession.ID,
		"#session_original_name":    originalSession.Name,
		"#session_original_created": originalSession.CreatedAt.Format("2006-01-02 15:04:05"),

		// Preserved content variables
		"#message_count": fmt.Sprintf("%d", len(newSession.Messages)),
		"#system_prompt": newSession.SystemPrompt,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&JSONImportCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-json-import command: %v", err))
	}
}
