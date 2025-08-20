package session

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ActivateCommand implements the \session-activate command for session activation.
// It provides session activation functionality with smart matching and auto-activation.
type ActivateCommand struct{}

// Name returns the command name "session-activate" for registration and lookup.
func (c *ActivateCommand) Name() string {
	return "session-activate"
}

// ParseMode returns ParseModeKeyValue for argument parsing with options.
func (c *ActivateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-activate command does.
func (c *ActivateCommand) Description() string {
	return "Activate session by name or ID with smart matching and auto-activation"
}

// Usage returns the syntax and usage examples for the session-activate command.
func (c *ActivateCommand) Usage() string {
	return `\session-activate[id=false] session_text
\session-activate[id=true] id_prefix
\session-activate

Examples:
  \session-activate                        %% Show active session or auto-activate latest
  \session-activate my-project             %% Activate by name (default) - matches any session name containing "my-project"
  \session-activate[id=true] 1234          %% Activate by ID prefix - matches any session ID starting with "1234"
  \session-activate proj                   %% Activate by partial name match
  \session-activate[id=true] abc123        %% Activate by ID prefix match

Options:
  id - Search by session ID prefix instead of name (default: false)

Notes:
  - Without parameters: shows active session or auto-activates most recent session
  - By default, searches session names for matches (partial matching supported)  
  - With id=true, searches session ID prefixes
  - If multiple sessions match, shows list of matches and asks for more specific input
  - If no sessions match, shows helpful suggestions
  - Sets the matched session as the currently active session for all operations
  - Updates system variables with active session information`
}

// HelpInfo returns structured help information for the session-activate command.
func (c *ActivateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-activate[id=false] session_text",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "id",
				Description: "Search by session ID prefix instead of name",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-activate",
				Description: "Show active session or auto-activate latest",
			},
			{
				Command:     "\\session-activate my-project",
				Description: "Activate by name match (default behavior)",
			},
			{
				Command:     "\\session-activate[id=true] 1234",
				Description: "Activate by ID prefix match",
			},
			{
				Command:     "\\session-activate proj",
				Description: "Activate by partial name match",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_session_id",
				Description: "ID of the currently active session",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4-8716-446655440000",
			},
			{
				Name:        "#active_session_id",
				Description: "ID of the currently active session",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4-8716-446655440000",
			},
			{
				Name:        "#active_session_name",
				Description: "Name of the currently active session",
				Type:        "system_metadata",
				Example:     "my-project",
			},
			{
				Name:        "#active_session_created",
				Description: "Creation timestamp of the active session",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#active_session_message_count",
				Description: "Number of messages in the active session",
				Type:        "system_metadata",
				Example:     "5",
			},
			{
				Name:        "_output",
				Description: "Command result message",
				Type:        "command_output",
				Example:     "Activated session 'my-project' (ID: 550e8400)",
			},
		},
		Notes: []string{
			"Without parameters: shows active session or auto-activates most recent",
			"By default searches session names (partial matching supported)",
			"Use id=true to search by session ID prefix instead",
			"If multiple sessions match, shows list and asks for more specific input",
			"If no sessions match, shows helpful suggestions",
			"Sets the matched session as currently active for all operations",
			"Variables in session text are interpolated before processing",
		},
	}
}

// Execute activates a session using smart matching or handles no-parameter cases.
func (c *ActivateCommand) Execute(args map[string]string, input string) error {
	// Parse arguments
	idStr := args["id"]
	byID := idStr == "true"

	// Get services
	chatSessionService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Validate input
	searchText := input
	if searchText == "" {
		// No parameters: handle smart activation
		return c.handleNoParameter(chatSessionService, variableService)
	}

	// With parameter: find and activate specific session
	return c.findAndActivateSession(searchText, byID, chatSessionService, variableService)
}

// handleNoParameter handles the no-parameter case with smart logic.
func (c *ActivateCommand) handleNoParameter(chatSessionService *services.ChatSessionService, variableService *services.VariableService) error {
	// Case 1: Try to get active session
	activeSession, err := chatSessionService.GetActiveSession()
	if err == nil {
		// Active session exists - show it
		sessionID := activeSession.ID
		sessionName := activeSession.Name
		outputMsg := fmt.Sprintf("Active session: %s (ID: %s)", sessionName, sessionID[:8])
		printer := printing.NewDefaultPrinter()
		printer.Info(outputMsg)

		// Store session ID in system variable
		return variableService.SetSystemVariable("_session_id", sessionID)
	}

	// Case 2 & 3: No active session - check if sessions exist
	sessions := chatSessionService.ListSessions()

	if len(sessions) == 0 {
		// Case 2: No sessions exist
		printer := printing.NewDefaultPrinter()
		printer.Info("No sessions found. Use \\session-new to create a session.")
		return nil
	}

	// Case 3: Sessions exist but no active - auto-activate latest
	latestSession := c.findLatestSessionByTimestamp(sessions)
	if latestSession == nil {
		return fmt.Errorf("failed to find latest session")
	}

	// Activate the latest session
	err = chatSessionService.SetActiveSession(latestSession.ID)
	if err != nil {
		return fmt.Errorf("failed to auto-activate latest session '%s': %w", latestSession.Name, err)
	}

	outputMsg := fmt.Sprintf("No active session found. Activated most recent session: %s (ID: %s)",
		latestSession.Name, latestSession.ID[:8])
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	// Update activation variables
	if err := c.updateActivationVariables(latestSession, variableService); err != nil {
		return fmt.Errorf("failed to update activation variables: %w", err)
	}

	// Trigger auto-save if enabled
	chatSessionService.TriggerAutoSave(latestSession.ID)

	// Store session ID in system variable
	return variableService.SetSystemVariable("_session_id", latestSession.ID)
}

// findAndActivateSession finds and activates a session by search text.
func (c *ActivateCommand) findAndActivateSession(searchText string, byID bool, chatSessionService *services.ChatSessionService, variableService *services.VariableService) error {
	// Get all sessions for searching
	sessions := chatSessionService.ListSessions()

	if len(sessions) == 0 {
		return fmt.Errorf("no sessions found. Use \\session-new to create session configurations")
	}

	// Find matching sessions
	var matches []*neurotypes.ChatSession
	if byID {
		matches = c.findSessionsByIDPrefix(sessions, searchText)
	} else {
		matches = c.findSessionsByName(sessions, searchText)
	}

	// Handle different match scenarios
	switch len(matches) {
	case 0:
		// No matches - provide helpful suggestions
		return c.handleNoMatches(sessions, searchText, byID)
	case 1:
		// Unique match - proceed with activation
		return c.activateSession(matches[0], chatSessionService, variableService)
	default:
		// Multiple matches - ask for more specific input
		return c.handleMultipleMatches(matches, searchText, byID)
	}
}

// findSessionsByName finds sessions whose names contain the search text (case-insensitive).
func (c *ActivateCommand) findSessionsByName(sessions []*neurotypes.ChatSession, searchText string) []*neurotypes.ChatSession {
	var matches []*neurotypes.ChatSession
	searchLower := strings.ToLower(searchText)

	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Name), searchLower) {
			matches = append(matches, session)
		}
	}

	return matches
}

// findSessionsByIDPrefix finds sessions whose IDs start with the search text (case-insensitive).
func (c *ActivateCommand) findSessionsByIDPrefix(sessions []*neurotypes.ChatSession, searchText string) []*neurotypes.ChatSession {
	var matches []*neurotypes.ChatSession
	searchLower := strings.ToLower(searchText)

	for _, session := range sessions {
		if strings.HasPrefix(strings.ToLower(session.ID), searchLower) {
			matches = append(matches, session)
		}
	}

	return matches
}

// findLatestSessionByTimestamp finds the most recently updated session for auto-activation.
func (c *ActivateCommand) findLatestSessionByTimestamp(sessions []*neurotypes.ChatSession) *neurotypes.ChatSession {
	var latest *neurotypes.ChatSession
	for _, session := range sessions {
		if latest == nil || session.UpdatedAt.After(latest.UpdatedAt) {
			latest = session
		}
	}
	return latest
}

// handleNoMatches provides helpful error message when no sessions match.
func (c *ActivateCommand) handleNoMatches(sessions []*neurotypes.ChatSession, searchText string, byID bool) error {
	searchType := "name"
	if byID {
		searchType = "ID prefix"
	}

	errorMsg := fmt.Sprintf("No sessions found matching %s '%s'.\n\nAvailable sessions:", searchType, searchText)

	// Show available sessions
	for _, session := range sessions {
		if byID {
			errorMsg += fmt.Sprintf("\n  ID: %s (name: %s)", session.ID[:8], session.Name)
		} else {
			errorMsg += fmt.Sprintf("\n  %s (ID: %s)", session.Name, session.ID[:8])
		}
	}

	return fmt.Errorf("%s", errorMsg)
}

// handleMultipleMatches provides helpful error message when multiple sessions match.
func (c *ActivateCommand) handleMultipleMatches(matches []*neurotypes.ChatSession, searchText string, byID bool) error {
	searchType := "name"
	if byID {
		searchType = "ID prefix"
	}

	errorMsg := fmt.Sprintf("Multiple sessions match %s '%s'. Please be more specific:\n", searchType, searchText)

	for _, session := range matches {
		messageCount := len(session.Messages)
		if byID {
			errorMsg += fmt.Sprintf("  ID: %s (name: %s, messages: %d)\n", session.ID[:8], session.Name, messageCount)
		} else {
			errorMsg += fmt.Sprintf("  %s (ID: %s, messages: %d)\n", session.Name, session.ID[:8], messageCount)
		}
	}

	errorMsg += "\nTip: Use the full name or a longer ID prefix to uniquely identify the session."

	return fmt.Errorf("%s", errorMsg)
}

// activateSession performs the actual session activation.
func (c *ActivateCommand) activateSession(session *neurotypes.ChatSession, chatSessionService *services.ChatSessionService, variableService *services.VariableService) error {
	// Set the session as active
	err := chatSessionService.SetActiveSession(session.ID)
	if err != nil {
		return fmt.Errorf("failed to activate session: %w", err)
	}

	// Prepare success message
	messageCount := len(session.Messages)
	outputMsg := fmt.Sprintf("Activated session '%s' (ID: %s, Messages: %d)",
		session.Name, session.ID[:8], messageCount)

	// Update activation-related variables
	if err := c.updateActivationVariables(session, variableService); err != nil {
		return fmt.Errorf("failed to update activation variables: %w", err)
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Store session ID in system variable
	if err := variableService.SetSystemVariable("_session_id", session.ID); err != nil {
		return fmt.Errorf("failed to set _session_id variable: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	// Trigger auto-save if enabled
	chatSessionService.TriggerAutoSave(session.ID)

	return nil
}

// updateActivationVariables sets session activation-related system variables.
func (c *ActivateCommand) updateActivationVariables(session *neurotypes.ChatSession, variableService *services.VariableService) error {
	// Set activation result variables
	variables := map[string]string{
		"#active_session_id":            session.ID,
		"#active_session_name":          session.Name,
		"#active_session_created":       session.CreatedAt.Format("2006-01-02 15:04:05"),
		"#active_session_message_count": fmt.Sprintf("%d", len(session.Messages)),
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// IsReadOnly returns false as the session-activate command modifies system state.
func (c *ActivateCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&ActivateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-activate command: %v", err))
	}
}
