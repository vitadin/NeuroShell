// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// RenameCommand implements the \session-rename command for changing session names.
// It provides session renaming functionality with validation and conflict checking.
type RenameCommand struct{}

// Name returns the command name "session-rename" for registration and lookup.
func (c *RenameCommand) Name() string {
	return "session-rename"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *RenameCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-rename command does.
func (c *RenameCommand) Description() string {
	return "Change session name"
}

// Usage returns the syntax and usage examples for the session-rename command.
func (c *RenameCommand) Usage() string {
	return `\session-rename[session=session_id] new_session_name
\session-rename new_session_name

Examples:
  \session-rename My New Session                                                  %% Rename active session
  \session-rename[session=work] Development Session                              %% Rename specific session
  \session-rename[session=old_name] "Project Alpha"                              %% Name with spaces (quoted)
  \session-rename[session=temp] ${project_name}                                  %% Using variable

Options:
  session - Session name or ID (optional, defaults to active session)

Input: New session name (must be valid and unique)

Note: Session names must be unique and follow naming conventions.
      Names are automatically validated and processed for consistency.
      The session retains all its content and history after renaming.`
}

// HelpInfo returns structured help information for the session-rename command.
func (c *RenameCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-rename[session=session_id] new_session_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "session",
				Description: "Session name or ID (optional, defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-rename My Project Session",
				Description: "Rename the active session",
			},
			{
				Command:     "\\session-rename[session=work] Development Work",
				Description: "Rename a specific session by name",
			},
			{
				Command:     "\\session-rename[session=550e8400] \"New Project Name\"",
				Description: "Rename a session by ID with quoted name",
			},
			{
				Command:     "\\session-rename[session=old] ${new_name}",
				Description: "Rename using a variable",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#session_id",
				Description: "ID of the session that was renamed",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#old_session_name",
				Description: "Previous session name",
				Type:        "system_metadata",
				Example:     "work_session",
			},
			{
				Name:        "#new_session_name",
				Description: "New session name",
				Type:        "system_metadata",
				Example:     "My Project Session",
			},
			{
				Name:        "_output",
				Description: "Rename operation result message",
				Type:        "command_output",
				Example:     "Renamed session from 'work_session' to 'My Project Session'",
			},
		},
		Notes: []string{
			"Session parameter is optional and defaults to active session",
			"New name must be unique across all existing sessions",
			"Names are validated and processed for consistency",
			"Session content and history are preserved during rename",
			"Use smart prefix matching for session lookup",
			"Empty or invalid names are rejected with helpful error messages",
		},
	}
}

// Execute renames the specified session to the new name.
// Options:
//   - session: Session name or ID (optional, defaults to active session)
func (c *RenameCommand) Execute(args map[string]string, input string) error {
	// Validate input name
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("new session name is required. Usage: %s", c.Usage())
	}

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

	// Parse arguments
	sessionID := args["session"]

	// Determine session - use provided session or default to active session
	var targetSession *neurotypes.ChatSession
	if sessionID == "" {
		// No session specified, use active session
		targetSession, err = chatService.GetActiveSession()
		if err != nil {
			return fmt.Errorf("no session specified and no active session found: %w. Usage: %s", err, c.Usage())
		}
	} else {
		// Get specified session
		targetSession, err = chatService.GetSessionByNameOrID(sessionID)
		if err != nil {
			return fmt.Errorf("failed to find session '%s': %w", sessionID, err)
		}
	}

	// Store the original name for reference
	originalName := targetSession.Name

	// Rename the session using the service
	err = chatService.RenameSession(targetSession.ID, input)
	if err != nil {
		return fmt.Errorf("failed to rename session: %w", err)
	}

	// Get the updated session to confirm changes
	updatedSession, err := chatService.GetSessionByNameOrID(targetSession.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated session: %w", err)
	}

	// Update result variables
	if err := c.updateRenameVariables(updatedSession, originalName, updatedSession.Name, variableService); err != nil {
		return fmt.Errorf("failed to update rename variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Renamed session from '%s' to '%s'", originalName, updatedSession.Name)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateRenameVariables sets rename-related system variables
func (c *RenameCommand) updateRenameVariables(session *neurotypes.ChatSession, oldName, newName string, variableService *services.VariableService) error {
	// Set rename variables
	variables := map[string]string{
		"#session_id":       session.ID,
		"#old_session_name": oldName,
		"#new_session_name": newName,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&RenameCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-rename command: %v", err))
	}
}
