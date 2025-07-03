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
	return `\session-delete[name=session_name] [session_name_or_id]

Examples:
  \session-delete work project        # Delete by exact name
  \session-delete work                # Delete by prefix (if unique match)
  \session-delete abc123-uuid         # Delete by session ID
  \session-delete[name=work]          # Delete using name option
  \session-delete ${#session_id}      # Delete current session by ID variable
  
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

// Execute deletes the specified chat session using smart search.
// Options:
//   - name: session name to delete (optional if provided as input)
//
// Input: session name or ID to delete (optional if provided as name option)
// Error if both name option and input are provided.
func (c *DeleteCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get chat session service
	chatService, err := c.getChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

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

	// Interpolate variables in session identifier
	sessionIdentifier, err = variableService.InterpolateString(sessionIdentifier, ctx)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in session identifier: %w", err)
	}

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

	// Update session-related variables after deletion
	if err := c.updateSessionVariablesAfterDeletion(variableService, ctx); err != nil {
		return fmt.Errorf("failed to update session variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Deleted session '%s' (ID: %s)", sessionName, sessionID)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg, ctx); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateSessionVariablesAfterDeletion clears session-related system variables if the deleted session was active
func (c *DeleteCommand) updateSessionVariablesAfterDeletion(variableService *services.VariableService, ctx neurotypes.Context) error {
	// Get chat session service to check current session
	chatService, err := c.getChatSessionService()
	if err != nil {
		return err
	}

	// Get current session (if any)
	currentSession, err := chatService.GetActiveSession()
	if err != nil {
		// Handle the case where there's no active session
		if err.Error() == "no active session" {
			currentSession = nil
		} else {
			return err
		}
	}
	if currentSession != nil {
		// Session is still active, update variables with current session info
		variables := map[string]string{
			"#session_id":      currentSession.ID,
			"#session_name":    currentSession.Name,
			"#message_count":   fmt.Sprintf("%d", len(currentSession.Messages)),
			"#system_prompt":   currentSession.SystemPrompt,
			"#session_created": currentSession.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		for name, value := range variables {
			if err := variableService.SetSystemVariable(name, value, ctx); err != nil {
				return fmt.Errorf("failed to set variable %s: %w", name, err)
			}
		}
	} else {
		// No active session, clear session variables
		sessionVariables := []string{
			"#session_id",
			"#session_name",
			"#message_count",
			"#system_prompt",
			"#session_created",
		}

		for _, name := range sessionVariables {
			if err := variableService.SetSystemVariable(name, "", ctx); err != nil {
				return fmt.Errorf("failed to clear variable %s: %w", name, err)
			}
		}
	}

	return nil
}

// getChatSessionService retrieves the chat session service from the global registry
func (c *DeleteCommand) getChatSessionService() (*services.ChatSessionService, error) {
	service, err := services.GetGlobalRegistry().GetService("chat_session")
	if err != nil {
		return nil, err
	}

	chatService, ok := service.(*services.ChatSessionService)
	if !ok {
		return nil, fmt.Errorf("chat session service has incorrect type")
	}

	return chatService, nil
}

// getVariableService retrieves the variable service from the global registry
func (c *DeleteCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&DeleteCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-delete command: %v", err))
	}
}
