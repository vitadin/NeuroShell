package session

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// GetCommand implements the \session-get command for session retrieval and activation.
// It supports dual functionality: get active session ID or find/activate session by name/prefix.
type GetCommand struct{}

// Name returns the command name "session-get" for registration and lookup.
func (c *GetCommand) Name() string {
	return "session-get"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *GetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-get command does.
func (c *GetCommand) Description() string {
	return "Get active session ID or find/activate session by name/prefix"
}

// Usage returns the syntax and usage examples for the session-get command.
func (c *GetCommand) Usage() string {
	return "\\session-get or \\session-get[name_or_prefix] or \\session-get name_or_prefix"
}

// HelpInfo returns structured help information for the session-get command.
func (c *GetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "name_or_prefix",
				Description: "Session name, ID, or unique prefix to find and activate",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-get",
				Description: "Get current active session ID",
			},
			{
				Command:     "\\session-get[project1]",
				Description: "Find and activate session 'project1', store ID",
			},
			{
				Command:     "\\session-get proj",
				Description: "Find and activate session with prefix 'proj'",
			},
		},
		Notes: []string{
			"Without parameters: gets current active session ID and stores in _session_id",
			"With parameter: finds session by exact name, exact ID, or unique prefix",
			"When session found via parameter, it becomes the active session",
			"Supports both bracket syntax (\\session-get[name]) and space syntax (\\session-get name)",
		},
	}
}

// Execute retrieves active session ID or finds/activates session by name/prefix.
// It handles both bracket and space syntax for session specification.
func (c *GetCommand) Execute(args map[string]string, input string) error {
	var sessionIdentifier string

	// Handle bracket syntax: \session-get[name_or_prefix]
	if len(args) > 0 {
		for key := range args {
			sessionIdentifier = key
			break
		}
	} else if input != "" {
		// Handle space syntax: \session-get name_or_prefix
		fields := strings.Fields(input)
		if len(fields) > 0 {
			sessionIdentifier = fields[0]
		}
	}

	// Get chat session service from global registry
	chatSessionService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing results
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	var sessionID string
	var sessionName string

	if sessionIdentifier == "" {
		// No parameters: get current active session ID
		activeSession, err := chatSessionService.GetActiveSession()
		if err != nil {
			return fmt.Errorf("no active session found: %w. Usage: %s", err, c.Usage())
		}
		sessionID = activeSession.ID
		sessionName = activeSession.Name
		fmt.Printf("Active session: %s (ID: %s)\n", sessionName, sessionID)
	} else {
		// With parameter: find session by name/prefix and activate
		session, err := chatSessionService.FindSessionByPrefix(sessionIdentifier)
		if err != nil {
			return fmt.Errorf("session lookup failed: %w. Usage: %s", err, c.Usage())
		}

		// Set the found session as active
		err = chatSessionService.SetActiveSession(session.ID)
		if err != nil {
			return fmt.Errorf("failed to activate session '%s': %w", session.Name, err)
		}

		sessionID = session.ID
		sessionName = session.Name
		fmt.Printf("Session activated: %s (ID: %s)\n", sessionName, sessionID)
	}

	// Store session ID in system variable
	err = variableService.SetSystemVariable("_session_id", sessionID)
	if err != nil {
		return fmt.Errorf("failed to set _session_id variable: %w", err)
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&GetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-get command: %v", err))
	}
}
