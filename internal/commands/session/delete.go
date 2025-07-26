// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// DeleteCommand implements the \session-delete command for deleting chat sessions.
// It provides session deletion functionality with support for both session names and IDs.
type DeleteCommand struct{}

// Name returns the command name "session-delete" for registration and lookup.
func (c *DeleteCommand) Name() string {
	return "session-delete"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *DeleteCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-delete command does.
func (c *DeleteCommand) Description() string {
	return "Delete an existing chat session"
}

// Usage returns the syntax and usage examples for the session-delete command.
func (c *DeleteCommand) Usage() string {
	return `\session-delete[name=session_name] or \session-delete session_name_or_id

Examples:
  \session-delete work project        %% Delete by exact name
  \session-delete work                %% Delete by prefix (if unique match)
  \session-delete abc123-uuid         %% Delete by session ID
  \session-delete[name=work]          %% Delete using name option
  \session-delete ${#session_id}      %% Delete current session by ID variable

Smart Search Priority:
  1. Exact session name match
  2. Exact session ID match
  3. Prefix matching (must be unique)

Options:
  name - Session name to delete (cannot combine with input parameter)

Error Handling:
  - Multiple prefix matches: Shows all matching session names
  - No matches: Indicates tried exact name, ID, and prefix search
  - Both arguments: Error - cannot specify both name option and input

Note: Session names are user-friendly. Use prefix matching for efficiency.
      Create sessions with: \session-new session_name_here`
}

// HelpInfo returns structured help information for the session-delete command.
func (c *DeleteCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-delete[name=session_name] or \\session-delete session_name_or_id",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "name",
				Description: "Session name to delete (cannot combine with input parameter)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-delete work",
				Description: "Delete session by name prefix (smart search)",
			},
			{
				Command:     "\\session-delete work project",
				Description: "Delete session by exact name match",
			},
			{
				Command:     "\\session-delete abc123-uuid",
				Description: "Delete session by exact ID match",
			},
			{
				Command:     "\\session-delete[name=work]",
				Description: "Delete using name option syntax",
			},
			{
				Command:     "\\session-delete ${#session_id}",
				Description: "Delete current session using session ID variable",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Deletion result message",
				Type:        "command_output",
				Example:     "Deleted session 'work project' (ID: 550e8400)",
			},
		},
		Notes: []string{
			"Smart search: tries exact name → exact ID → prefix matching",
			"Prefix matching must result in a unique match or shows all candidates",
			"Cannot specify both name option and input parameter",
			"Variables in session identifier are interpolated before search",
			"Session variables are updated or cleared after deletion",
		},
	}
}

// Execute deletes the specified chat session using smart search.
// Options:
//   - name: session name to delete (optional if provided as input)
//
// Input: session name or ID to delete (optional if provided as name option)
// Error if both name option and input are provided.
func (c *DeleteCommand) Execute(args map[string]string, input string) error {
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

	// Note: Variable interpolation is now handled by the state machine before commands execute

	// Validate arguments - cannot specify both name option and input
	nameOption := args["name"]
	if nameOption != "" && input != "" {
		return fmt.Errorf("cannot specify both name option and input parameter\n\nUsage: %s", c.Usage())
	}

	// Determine session identifier
	sessionIdentifier := input
	if sessionIdentifier == "" {
		sessionIdentifier = nameOption
	}

	// Validate that we have a session identifier
	if sessionIdentifier == "" {
		return fmt.Errorf("session name or ID is required\n\nUsage: %s", c.Usage())
	}

	// Note: Variable interpolation for session identifier is handled by state machine

	// Use smart search to find the session
	session, err := chatService.FindSessionByPrefix(sessionIdentifier)
	if err != nil {
		return err // Error message already includes context from FindSessionByPrefix
	}
	sessionName := session.Name
	sessionID := session.ID[:8] // Short ID for display

	// Delete the session using the exact session ID
	err = chatService.DeleteSession(session.ID)
	if err != nil {
		return fmt.Errorf("failed to delete session '%s': %w", sessionName, err)
	}

	// Auto-push session activation command to stack service to handle active session state
	// This will either show current active session, activate latest session, or show "no sessions" message
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := "\\silent \\session-activate"
		stackService.PushCommand(activateCommand)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Deleted session '%s' (ID: %s)", sessionName, sessionID)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&DeleteCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-delete command: %v", err))
	}
}
