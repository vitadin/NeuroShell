package session

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// Display constants for smart content truncation
const (
	MaxSystemPromptDisplay = 100 // chars before truncation
	MaxMessageDisplay      = 80  // chars before truncation
	MaxMessagesShown       = 10  // show first 5 + last 5 when > 10
	TruncationIndicator    = "..."
)

// ShowCommand implements the \session-show command for displaying session information.
// It provides session information display functionality with smart matching and content rendering.
type ShowCommand struct{}

// Name returns the command name "session-show" for registration and lookup.
func (c *ShowCommand) Name() string {
	return "session-show"
}

// ParseMode returns ParseModeKeyValue for argument parsing with options.
func (c *ShowCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-show command does.
func (c *ShowCommand) Description() string {
	return "Display detailed session information with smart content rendering"
}

// Usage returns the syntax and usage examples for the session-show command.
func (c *ShowCommand) Usage() string {
	return `\session-show[id=false] session_text
\session-show[id=true] id_prefix
\session-show

Examples:
  \session-show                        %% Show active session details or list if none active
  \session-show my-project             %% Show session by name (default) - matches any session name containing "my-project"
  \session-show[id=true] 1234          %% Show session by ID prefix - matches any session ID starting with "1234"
  \session-show proj                   %% Show session by partial name match
  \session-show[id=true] abc123        %% Show session by ID prefix match

Options:
  id - Search by session ID prefix instead of name (default: false)

Notes:
  - Without parameters: shows active session details or helpful guidance if none active
  - By default, searches session names for matches (partial matching supported)  
  - With id=true, searches session ID prefixes
  - If multiple sessions match, shows list of matches and asks for more specific input
  - If no sessions match, shows helpful suggestions
  - Displays session metadata, system prompt, and message history with smart truncation
  - Long content is truncated with ellipsis and character counts shown`
}

// HelpInfo returns structured help information for the session-show command.
func (c *ShowCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-show[id=false] session_text",
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
				Command:     "\\session-show",
				Description: "Show active session details or guidance if none active",
			},
			{
				Command:     "\\session-show my-project",
				Description: "Show session by name match (default behavior)",
			},
			{
				Command:     "\\session-show[id=true] 1234",
				Description: "Show session by ID prefix match",
			},
			{
				Command:     "\\session-show proj",
				Description: "Show session by partial name match",
			},
		},
		Notes: []string{
			"Without parameters: shows active session or helpful guidance",
			"By default searches session names (partial matching supported)",
			"Use id=true to search by session ID prefix instead",
			"If multiple sessions match, shows list and asks for more specific input",
			"If no sessions match, shows helpful suggestions",
			"Displays rich session information with smart content truncation",
			"Variables in session text are interpolated before processing",
		},
	}
}

// Execute displays session information using smart matching or handles no-parameter cases.
func (c *ShowCommand) Execute(args map[string]string, input string) error {
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
		// No parameters: handle smart display
		return c.handleNoParameter(chatSessionService, variableService)
	}

	// With parameter: find and show specific session
	return c.findAndShowSession(searchText, byID, chatSessionService, variableService)
}

// handleNoParameter handles the no-parameter case with smart logic.
func (c *ShowCommand) handleNoParameter(chatSessionService *services.ChatSessionService, variableService *services.VariableService) error {
	// Case 1: Try to get active session
	activeSession, err := chatSessionService.GetActiveSession()
	if err == nil {
		// Active session exists - show it
		return c.renderSessionInfo(activeSession, variableService)
	}

	// Case 2 & 3: No active session - check if sessions exist
	sessions := chatSessionService.ListSessions()

	if len(sessions) == 0 {
		// Case 2: No sessions exist
		printer := c.createPrinter()
		printer.Info("No sessions found. Use \\session-new to create a session.")
		return nil
	}

	// Case 3: Sessions exist but no active - show guidance
	printer := c.createPrinter()
	printer.Info(fmt.Sprintf("No active session found. Use \\session-activate to activate a session.\n\nAvailable sessions (%d):", len(sessions)))
	for _, session := range sessions {
		messageCount := len(session.Messages)
		printer.Info(fmt.Sprintf("  %s (ID: %s, Messages: %d)", session.Name, session.ID[:8], messageCount))
	}

	return nil
}

// findAndShowSession finds and displays a session by search text.
func (c *ShowCommand) findAndShowSession(searchText string, byID bool, chatSessionService *services.ChatSessionService, variableService *services.VariableService) error {
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
		// Unique match - proceed with display
		return c.renderSessionInfo(matches[0], variableService)
	default:
		// Multiple matches - ask for more specific input
		return c.handleMultipleMatches(matches, searchText, byID)
	}
}

// findSessionsByName finds sessions whose names contain the search text (case-insensitive).
func (c *ShowCommand) findSessionsByName(sessions []*neurotypes.ChatSession, searchText string) []*neurotypes.ChatSession {
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
func (c *ShowCommand) findSessionsByIDPrefix(sessions []*neurotypes.ChatSession, searchText string) []*neurotypes.ChatSession {
	var matches []*neurotypes.ChatSession
	searchLower := strings.ToLower(searchText)

	for _, session := range sessions {
		if strings.HasPrefix(strings.ToLower(session.ID), searchLower) {
			matches = append(matches, session)
		}
	}

	return matches
}

// handleNoMatches provides helpful error message when no sessions match.
func (c *ShowCommand) handleNoMatches(sessions []*neurotypes.ChatSession, searchText string, byID bool) error {
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
func (c *ShowCommand) handleMultipleMatches(matches []*neurotypes.ChatSession, searchText string, byID bool) error {
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

// renderSessionInfo displays comprehensive session information with smart formatting.
func (c *ShowCommand) renderSessionInfo(session *neurotypes.ChatSession, variableService *services.VariableService) error {
	// Get printer for output
	printer := c.createPrinter()

	// Display session header
	sessionHeader := fmt.Sprintf("Session: %s (ID: %s)", session.Name, session.ID[:8])
	printer.Success(sessionHeader)

	// Display system prompt with truncation
	systemPrompt := c.truncateContent(session.SystemPrompt, MaxSystemPromptDisplay)
	printer.Info(fmt.Sprintf("System: %s", systemPrompt))

	// Display timestamps
	printer.Info(fmt.Sprintf("Created: %s", session.CreatedAt.Format("2006-01-02 15:04:05")))
	printer.Info(fmt.Sprintf("Updated: %s", session.UpdatedAt.Format("2006-01-02 15:04:05")))

	// Display message count
	messageCount := len(session.Messages)
	printer.Info(fmt.Sprintf("Messages: %d total", messageCount))
	printer.Println("") // Add blank line after message count

	// Display messages with smart truncation
	if messageCount > 0 {
		c.renderMessages(session.Messages, printer)
	}

	// Auto-push session activation command to stack service to handle active session state
	// This will activate the shown session for seamless UX
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := fmt.Sprintf("\\silent \\session-activate[id=true] %s", session.ID)
		stackService.PushCommand(activateCommand)
	}

	// Store session information in variables for potential use
	outputMsg := fmt.Sprintf("Displayed session '%s' (ID: %s, Messages: %d)",
		session.Name, session.ID[:8], messageCount)

	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	return nil
}

// renderMessages displays session messages with smart truncation and role information.
func (c *ShowCommand) renderMessages(messages []neurotypes.Message, printer *output.Printer) {
	messageCount := len(messages)

	if messageCount <= MaxMessagesShown {
		// Show all messages
		for i, msg := range messages {
			c.renderSingleMessage(i+1, msg, printer)
		}
	} else {
		// Show first 5 messages
		for i := 0; i < 5; i++ {
			c.renderSingleMessage(i+1, messages[i], printer)
		}

		// Show separator with count
		hiddenCount := messageCount - 10
		separator := fmt.Sprintf("... (%d more messages) ...", hiddenCount)
		printer.Warning(separator)
		printer.Print("")

		// Show last 5 messages
		for i := messageCount - 5; i < messageCount; i++ {
			c.renderSingleMessage(i+1, messages[i], printer)
		}
	}
}

// renderSingleMessage displays a single message with role and truncated content.
func (c *ShowCommand) renderSingleMessage(index int, msg neurotypes.Message, printer *output.Printer) {
	// Truncate message content
	content := c.truncateContent(msg.Content, MaxMessageDisplay)

	// Format timestamp
	timestamp := msg.Timestamp.Format("15:04:05")

	// Display message with index, role, timestamp, and content in original format
	messageOutput := fmt.Sprintf("[%d] %s (%s): %s", index, msg.Role, timestamp, content)
	printer.Println(messageOutput)
}

// truncateContent truncates text content with ellipsis and character count if needed.
// Multi-line content is compressed to single line using Go's %q formatting.
func (c *ShowCommand) truncateContent(content string, maxLength int) string {
	// Use %q to escape newlines and special characters for single-line display
	quoted := fmt.Sprintf("%q", content)
	// Remove the surrounding quotes added by %q
	if len(quoted) >= 2 && quoted[0] == '"' && quoted[len(quoted)-1] == '"' {
		quoted = quoted[1 : len(quoted)-1]
	}

	if len(quoted) <= maxLength {
		return quoted
	}

	// Calculate split points for showing beginning and end
	prefixLength := maxLength / 2
	suffixLength := maxLength - prefixLength - len(TruncationIndicator)

	// Handle very short maxLength
	if maxLength <= len(TruncationIndicator) {
		return TruncationIndicator
	}

	// Handle edge case where suffix would be negative
	if suffixLength < 1 {
		// Just truncate to maxLength - ellipsis length
		truncateAt := maxLength - len(TruncationIndicator)
		if truncateAt < 0 {
			return TruncationIndicator
		}
		return quoted[:truncateAt] + TruncationIndicator
	}

	prefix := quoted[:prefixLength]
	suffix := quoted[len(quoted)-suffixLength:]

	return fmt.Sprintf("%s%s%s (%d chars)", prefix, TruncationIndicator, suffix, len(content))
}

// createPrinter creates a printer with theme service as style provider
func (c *ShowCommand) createPrinter() *output.Printer {
	// Try to get theme service as style provider
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// Fall back to plain style provider
		return output.NewPrinter(output.WithStyles(output.NewPlainStyleProvider()))
	}

	return output.NewPrinter(output.WithStyles(themeService))
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ShowCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-show command: %v", err))
	}
}
