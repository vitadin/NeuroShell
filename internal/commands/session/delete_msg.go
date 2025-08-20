// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

// DeleteMessageCommand implements the \session-delete-msg command for deleting message by index.
// It provides message deletion functionality with dual indexing system (reverse and normal order).
type DeleteMessageCommand struct{}

// Name returns the command name "session-delete-msg" for registration and lookup.
func (c *DeleteMessageCommand) Name() string {
	return "session-delete-msg"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *DeleteMessageCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-delete-msg command does.
func (c *DeleteMessageCommand) Description() string {
	return "Delete message by index using dual indexing system"
}

// Usage returns the syntax and usage examples for the session-delete-msg command.
func (c *DeleteMessageCommand) Usage() string {
	return `\session-delete-msg[idx=N, session=session_id]
\session-delete-msg[idx=.N]

Examples:
  \session-delete-msg[idx=1]                                                    %% Delete the last message (reverse order)
  \session-delete-msg[idx=2]                                                    %% Delete second-to-last message
  \session-delete-msg[idx=.1]                                                   %% Delete the first message (normal order)
  \session-delete-msg[idx=.3]                                                   %% Delete the third message
  \session-delete-msg[session=work, idx=1]                                     %% Delete last message in work session
  \session-delete-msg[idx=3, confirm=true]                                     %% Delete with explicit confirmation

Options:
  idx     - Message index (required): N for reverse order (1=last), .N for normal order (.1=first)
  session - Session name or ID (optional, defaults to active session)
  confirm - Explicit confirmation (optional, defaults to true for safety)

Note: Indexing systems:
      - Reverse order (idx=N): 1=last message, 2=second-to-last, 3=third-to-last, etc.
      - Normal order (idx=.N): .1=first message, .2=second message, .3=third message, etc.
      Message is permanently removed from the session.`
}

// HelpInfo returns structured help information for the session-delete-msg command.
func (c *DeleteMessageCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-delete-msg[idx=N OR idx=.N, session=session_id, confirm=true]",
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
			{
				Name:        "confirm",
				Description: "Explicit confirmation (optional, defaults to true for safety)",
				Required:    false,
				Type:        "boolean",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-delete-msg[idx=1]",
				Description: "Delete the last (most recent) message",
			},
			{
				Command:     "\\session-delete-msg[idx=2]",
				Description: "Delete the second-to-last message using reverse indexing",
			},
			{
				Command:     "\\session-delete-msg[idx=.1]",
				Description: "Delete the first (oldest) message using normal indexing",
			},
			{
				Command:     "\\session-delete-msg[session=work, idx=1]",
				Description: "Delete last message in a specific session",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#deleted_message_id",
				Description: "ID of the deleted message",
				Type:        "system_metadata",
				Example:     "msg-550e8400-e29b-41d4",
			},
			{
				Name:        "#deleted_message_role",
				Description: "Role of the deleted message (user or assistant)",
				Type:        "system_metadata",
				Example:     "user",
			},
			{
				Name:        "#deleted_message_index",
				Description: "The index used to identify the message",
				Type:        "system_metadata",
				Example:     "1",
			},
			{
				Name:        "#deleted_message_position",
				Description: "Human-readable position description",
				Type:        "system_metadata",
				Example:     "last message",
			},
			{
				Name:        "#deleted_content",
				Description: "Content of the deleted message",
				Type:        "system_metadata",
				Example:     "Deleted message text",
			},
			{
				Name:        "#deleted_message_timestamp",
				Description: "Timestamp of the deleted message",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#session_id",
				Description: "ID of the session containing the deleted message",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "Name of the session containing the deleted message",
				Type:        "system_metadata",
				Example:     "work_session",
			},
			{
				Name:        "#remaining_messages",
				Description: "Number of messages remaining in session after deletion",
				Type:        "system_metadata",
				Example:     "4",
			},
			{
				Name:        "_output",
				Description: "Delete operation result message",
				Type:        "command_output",
				Example:     "Deleted message 1 (last message) from session 'work_session'",
			},
		},
		Notes: []string{
			"Dual indexing system: idx=N (reverse: 1=last), idx=.N (normal: .1=first)",
			"Message deletion is permanent and cannot be undone",
			"Session parameter is optional and defaults to active session",
			"Reverse order indexing is most common for deleting recent messages",
			"Normal order indexing useful for deleting historical conversation starts",
			"Index bounds are validated before deletion to prevent errors",
			"Confirmation is enabled by default for safety",
		},
	}
}

// Execute deletes a message from the specified session using the dual indexing system.
// Options:
//   - idx: Message index (required) - N for reverse order, .N for normal order
//   - session: Session name or ID (optional, defaults to active session)
//   - confirm: Explicit confirmation (optional, defaults to true for safety)
func (c *DeleteMessageCommand) Execute(args map[string]string, input string) error {
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
	confirmStr := args["confirm"]

	// Validate idx parameter
	if idxStr == "" {
		return fmt.Errorf("idx parameter is required. Usage: %s", c.Usage())
	}

	// Handle confirmation (default to true for safety)
	confirm := true // Default to requiring confirmation
	if confirmStr != "" {
		confirm = strings.ToLower(confirmStr) == "true"
	}

	// Warn if input is provided (not needed for deletion)
	if strings.TrimSpace(input) != "" {
		printer := printing.NewDefaultPrinter()
		printer.Warning("Input text ignored for message deletion command")
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
		return fmt.Errorf("session '%s' has no messages to delete", targetSession.Name)
	}

	// Parse message index using shared logic
	indexResult, err := stringprocessing.ParseMessageIndex(idxStr, len(targetSession.Messages))
	if err != nil {
		return fmt.Errorf("invalid index '%s': %w. Usage: %s", idxStr, err, c.Usage())
	}

	// Get the message that will be deleted for reference
	messageToDelete := targetSession.Messages[indexResult.ZeroBasedIndex]

	// Apply confirmation if enabled
	if confirm {
		printer := printing.NewDefaultPrinter()
		printer.Warning(fmt.Sprintf("About to delete message %s (%s) from session '%s'",
			idxStr, indexResult.PositionDescription, targetSession.Name))
		printer.Info(fmt.Sprintf("Message content: %s", truncateContent(messageToDelete.Content, 100)))
		printer.Info("Deletion is permanent and cannot be undone")
		printer.Info("To proceed without confirmation, use: confirm=false")
		return fmt.Errorf("deletion cancelled for safety - use confirm=false to bypass confirmation")
	}

	// Delete the message using the service
	err = chatService.DeleteMessage(targetSession.ID, indexResult.ZeroBasedIndex)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Get the updated session to check remaining message count
	updatedSession, err := chatService.GetSessionByNameOrID(targetSession.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated session: %w", err)
	}

	// Update result variables
	if err := c.updateDeletedMessageVariables(updatedSession, &messageToDelete, idxStr, indexResult.PositionDescription, variableService); err != nil {
		return fmt.Errorf("failed to update message variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Deleted message %s (%s) from session '%s'", idxStr, indexResult.PositionDescription, updatedSession.Name)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)
	printer.Info(fmt.Sprintf("Session now has %d messages remaining", len(updatedSession.Messages)))

	return nil
}

// updateDeletedMessageVariables sets message-related system variables for the deleted message
func (c *DeleteMessageCommand) updateDeletedMessageVariables(session *neurotypes.ChatSession, deletedMessage *neurotypes.Message, idxStr, positionDesc string, variableService *services.VariableService) error {
	// Set deleted message variables
	variables := map[string]string{
		"#deleted_message_id":        deletedMessage.ID,
		"#deleted_message_role":      deletedMessage.Role,
		"#deleted_message_index":     idxStr,
		"#deleted_message_position":  positionDesc,
		"#deleted_content":           deletedMessage.Content,
		"#deleted_message_timestamp": deletedMessage.Timestamp.Format("2006-01-02 15:04:05"),
		"#session_id":                session.ID,
		"#session_name":              session.Name,
		"#remaining_messages":        fmt.Sprintf("%d", len(session.Messages)),
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// truncateContent truncates content to a specified length for display purposes
func truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "..."
}

// IsReadOnly returns false as the session-delete-msg command modifies system state.
func (c *DeleteMessageCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&DeleteMessageCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-delete-msg command: %v", err))
	}
}
