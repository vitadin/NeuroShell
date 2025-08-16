// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EditMessageCommand implements the \session-edit-msg command for editing message content by index.
// It provides message editing functionality with dual indexing system (reverse and normal order).
type EditMessageCommand struct{}

// Name returns the command name "session-edit-msg" for registration and lookup.
func (c *EditMessageCommand) Name() string {
	return "session-edit-msg"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *EditMessageCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-edit-msg command does.
func (c *EditMessageCommand) Description() string {
	return "Edit message content by index using dual indexing system"
}

// Usage returns the syntax and usage examples for the session-edit-msg command.
func (c *EditMessageCommand) Usage() string {
	return `\session-edit-msg[idx=N, session=session_id] new_message_content
\session-edit-msg[idx=.N] new_message_content

Examples:
  \session-edit-msg[idx=1] Edit the last message content                         %% Reverse order: 1=last, 2=second-to-last
  \session-edit-msg[idx=2] Edit second-to-last message                          %% Reverse order
  \session-edit-msg[idx=.1] Edit the first message content                      %% Normal order: .1=first, .2=second
  \session-edit-msg[idx=.3] Edit the third message                              %% Normal order
  \session-edit-msg[session=work, idx=1] Edit last message in work session     %% Specific session
  \session-edit-msg[idx=1] ${new_content}                                       %% Using variable

Options:
  idx     - Message index (required): N for reverse order (1=last), .N for normal order (.1=first)
  session - Session name or ID (optional, defaults to active session)

Input: New message content to replace the existing content

Note: Indexing systems:
      - Reverse order (idx=N): 1=last message, 2=second-to-last, 3=third-to-last, etc.
      - Normal order (idx=.N): .1=first message, .2=second message, .3=third message, etc.
      Message metadata (ID, role, timestamp) is preserved, only content is changed.`
}

// HelpInfo returns structured help information for the session-edit-msg command.
func (c *EditMessageCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-edit-msg[idx=N OR idx=.N, session=session_id] new_message_content",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "idx",
				Description: "Message index: N for reverse order (1=last), .N for normal order (.1=first)",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "session",
				Description: "Session name or ID (optional, defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-edit-msg[idx=1] Corrected last message",
				Description: "Edit the last (most recent) message content",
			},
			{
				Command:     "\\session-edit-msg[idx=2] Updated second-to-last message",
				Description: "Edit the second-to-last message using reverse indexing",
			},
			{
				Command:     "\\session-edit-msg[idx=.1] Updated first message",
				Description: "Edit the first (oldest) message using normal indexing",
			},
			{
				Command:     "\\session-edit-msg[session=work, idx=1] Fixed last message in work session",
				Description: "Edit last message in a specific session",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#message_id",
				Description: "ID of the edited message",
				Type:        "system_metadata",
				Example:     "msg-550e8400-e29b-41d4",
			},
			{
				Name:        "#message_role",
				Description: "Role of the edited message (user or assistant)",
				Type:        "system_metadata",
				Example:     "user",
			},
			{
				Name:        "#message_index",
				Description: "The index used to identify the message",
				Type:        "system_metadata",
				Example:     "1",
			},
			{
				Name:        "#message_position",
				Description: "Human-readable position description",
				Type:        "system_metadata",
				Example:     "last message",
			},
			{
				Name:        "#old_content",
				Description: "Original message content before editing",
				Type:        "system_metadata",
				Example:     "Original message text",
			},
			{
				Name:        "#new_content",
				Description: "New message content after editing",
				Type:        "system_metadata",
				Example:     "Updated message text",
			},
			{
				Name:        "#message_timestamp",
				Description: "Original timestamp of the message (preserved)",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#session_id",
				Description: "ID of the session containing the message",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "Name of the session containing the message",
				Type:        "system_metadata",
				Example:     "work_session",
			},
			{
				Name:        "_output",
				Description: "Edit operation result message",
				Type:        "command_output",
				Example:     "Edited message 1 (last message) in session 'work_session'",
			},
		},
		Notes: []string{
			"Dual indexing system: idx=N (reverse: 1=last), idx=.N (normal: .1=first)",
			"Message metadata (ID, role, timestamp) is preserved, only content changes",
			"Session parameter is optional and defaults to active session",
			"Reverse order indexing is most common for editing recent messages",
			"Normal order indexing useful for editing historical conversation starts",
			"Index bounds are validated before editing to prevent errors",
		},
	}
}

// Execute edits a message in the specified session using the dual indexing system.
// Options:
//   - idx: Message index (required) - N for reverse order, .N for normal order
//   - session: Session name or ID (optional, defaults to active session)
func (c *EditMessageCommand) Execute(args map[string]string, input string) error {
	// Validate input content
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("new message content is required. Usage: %s", c.Usage())
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
	idxStr := args["idx"]
	sessionID := args["session"]

	// Validate idx parameter
	if idxStr == "" {
		return fmt.Errorf("idx parameter is required. Usage: %s", c.Usage())
	}

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

	// Check if session has messages
	if len(targetSession.Messages) == 0 {
		return fmt.Errorf("session '%s' has no messages to edit", targetSession.Name)
	}

	// Parse message index using dual indexing system
	messageIndex, positionDesc, err := c.parseMessageIndex(idxStr, len(targetSession.Messages))
	if err != nil {
		return fmt.Errorf("invalid index '%s': %w. Usage: %s", idxStr, err, c.Usage())
	}

	// Store the original content for reference
	originalContent := targetSession.Messages[messageIndex].Content

	// Edit the message using the service
	err = chatService.EditMessage(targetSession.ID, messageIndex, input)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	// Get the updated session to access the edited message
	updatedSession, err := chatService.GetSessionByNameOrID(targetSession.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated session: %w", err)
	}
	editedMessage := updatedSession.Messages[messageIndex]

	// Update result variables
	if err := c.updateMessageVariables(updatedSession, &editedMessage, idxStr, positionDesc, originalContent, input, variableService); err != nil {
		return fmt.Errorf("failed to update message variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Edited message %s (%s) in session '%s'", idxStr, positionDesc, updatedSession.Name)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	return nil
}

// parseMessageIndex parses the idx parameter and converts it to 0-based index
// Returns: (0-based index, position description, error)
func (c *EditMessageCommand) parseMessageIndex(idxStr string, messageCount int) (int, string, error) {
	if strings.HasPrefix(idxStr, ".") {
		// Normal order: .1, .2, .3 -> 0-based: 0, 1, 2
		numStr := idxStr[1:]
		if numStr == "" {
			return -1, "", fmt.Errorf("invalid normal order index format (use .1, .2, .3, etc.)")
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return -1, "", fmt.Errorf("invalid normal order index number: %w", err)
		}

		if num < 1 || num > messageCount {
			return -1, "", fmt.Errorf("normal order index %d is out of bounds (session has %d messages)", num, messageCount)
		}

		zeroBasedIndex := num - 1
		positionDesc := c.getOrdinalPosition(num, false)
		return zeroBasedIndex, positionDesc, nil
	}

	// Reverse order: 1, 2, 3 -> 0-based: last, second-to-last, third-to-last
	num, err := strconv.Atoi(idxStr)
	if err != nil {
		return -1, "", fmt.Errorf("invalid reverse order index number: %w", err)
	}

	if num < 1 || num > messageCount {
		return -1, "", fmt.Errorf("reverse order index %d is out of bounds (session has %d messages)", num, messageCount)
	}

	zeroBasedIndex := messageCount - num
	positionDesc := c.getOrdinalPosition(num, true)
	return zeroBasedIndex, positionDesc, nil
}

// getOrdinalPosition returns a human-readable position description
func (c *EditMessageCommand) getOrdinalPosition(num int, reverse bool) string {
	if reverse {
		switch num {
		case 1:
			return "last message"
		case 2:
			return "second-to-last message"
		case 3:
			return "third-to-last message"
		default:
			return fmt.Sprintf("%d%s from last message", num, getOrdinalSuffix(num))
		}
	} else {
		switch num {
		case 1:
			return "first message"
		case 2:
			return "second message"
		case 3:
			return "third message"
		default:
			return fmt.Sprintf("%d%s message", num, getOrdinalSuffix(num))
		}
	}
}

// getOrdinalSuffix returns the appropriate ordinal suffix (st, nd, rd, th)
func getOrdinalSuffix(num int) string {
	// Special case for numbers ending in 11, 12, 13 (like 11, 12, 13, 111, 112, 113, etc.)
	if num%100 >= 11 && num%100 <= 13 {
		return "th"
	}
	switch num % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// updateMessageVariables sets message-related system variables
func (c *EditMessageCommand) updateMessageVariables(session *neurotypes.ChatSession, message *neurotypes.Message, idxStr, positionDesc, oldContent, newContent string, variableService *services.VariableService) error {
	// Set message variables
	variables := map[string]string{
		"#message_id":        message.ID,
		"#message_role":      message.Role,
		"#message_index":     idxStr,
		"#message_position":  positionDesc,
		"#old_content":       oldContent,
		"#new_content":       newContent,
		"#message_timestamp": message.Timestamp.Format("2006-01-02 15:04:05"),
		"#session_id":        session.ID,
		"#session_name":      session.Name,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EditMessageCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-edit-msg command: %v", err))
	}
}
