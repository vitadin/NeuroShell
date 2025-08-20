// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"os"
	"path/filepath"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// SaveCommand implements the \session-save command for saving sessions to the auto-save directory.
// It provides automatic session backup functionality with fixed path and naming conventions.
type SaveCommand struct{}

// Name returns the command name "session-save" for registration and lookup.
func (c *SaveCommand) Name() string {
	return "session-save"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *SaveCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-save command does.
func (c *SaveCommand) Description() string {
	return "Save chat session to auto-save directory"
}

// Usage returns the syntax and usage examples for the session-save command.
func (c *SaveCommand) Usage() string {
	return `\session-save session_identifier

Examples:
  \session-save work                    %% Save "work" session to auto-save directory
  \session-save project                 %% Save "project" session
  \session-save 550e8400                %% Save session by ID
  \session-save proj                    %% Save using prefix matching

Note: Session identifier can be exact name, exact ID, or unique prefix.
      Sessions are automatically saved to ~/.config/neuroshell/sessions/{session-id}.json
      This command overwrites any existing auto-save file for the session.
      For custom export paths, use \session-export instead.`
}

// HelpInfo returns structured help information for the session-save command.
func (c *SaveCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-save session_identifier",
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options - fixed behavior
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-save work",
				Description: "Save session named 'work' to auto-save directory",
			},
			{
				Command:     "\\session-save 550e8400",
				Description: "Save session by ID to auto-save directory",
			},
			{
				Command:     "\\session-save proj",
				Description: "Save session using prefix matching",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Save result message with session details",
				Type:        "command_output",
				Example:     "Session saved to sessions/550e8400-e29b-41d4-a716-446655440000.json",
			},
		},
		Notes: []string{
			"Sessions are saved to ~/.config/neuroshell/sessions/ directory",
			"Filename format: {session-id}.json",
			"Always overwrites existing auto-save file for the session",
			"Creates the sessions directory if it doesn't exist",
			"For custom export paths, use \\session-export command instead",
		},
	}
}

// Execute saves a chat session to the auto-save directory.
// The input parameter specifies the session identifier (name, ID, or prefix).
// No options are accepted - the path and filename are automatically determined.
func (c *SaveCommand) Execute(_ map[string]string, input string) error {

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

	// Validate session identifier
	if input == "" {
		return fmt.Errorf("session identifier is required")
	}

	// Find session using prefix matching (follows existing pattern)
	session, err := chatService.FindSessionByPrefix(input)
	if err != nil {
		return fmt.Errorf("session lookup failed: %w", err)
	}

	// Get user config directory for sessions path
	ctx := neuroshellcontext.GetGlobalContext()
	configDir, err := ctx.GetUserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Construct sessions directory path
	sessionsDir := filepath.Join(configDir, "sessions")

	// Create sessions directory and any missing parent directories (like mkdir -p)
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Construct filename using session ID
	filename := fmt.Sprintf("%s.json", session.ID)
	filepath := filepath.Join(sessionsDir, filename)

	// Export session to the auto-save location
	if err := chatService.ExportSessionToJSON(session.ID, filepath); err != nil {
		return fmt.Errorf("auto-save failed: %w", err)
	}

	// Prepare output message (simplified for auto-save)
	outputMsg := fmt.Sprintf("Session saved to sessions/%s", filename)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	return nil
}

// IsReadOnly returns false as the session-save command modifies system state.
func (c *SaveCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&SaveCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-save command: %v", err))
	}
}
