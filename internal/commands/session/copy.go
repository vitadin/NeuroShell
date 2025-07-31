// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CopyCommand implements the \session-copy command for creating deep copies of existing sessions.
// It provides session copying functionality with explicit source identification and optional target naming.
type CopyCommand struct{}

// Name returns the command name "session-copy" for registration and lookup.
func (c *CopyCommand) Name() string {
	return "session-copy"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *CopyCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-copy command does.
func (c *CopyCommand) Description() string {
	return "Create deep copy of existing session with new identity"
}

// Usage returns the syntax and usage examples for the session-copy command.
func (c *CopyCommand) Usage() string {
	return `\session-copy[source_session_id=id, target_session_name="name"]
\session-copy[source_session_name=name, target_session_name="name"]

Examples:
  \session-copy[source_session_id=abc12345-678d-90ef]                           %% Copy by ID with auto-generated name
  \session-copy[source_session_name=my_project]                                %% Copy by name with auto-generated name
  \session-copy[source_session_id=abc123, target_session_name="experiment"]     %% Copy by ID with custom name
  \session-copy[source_session_name=work_session, target_session_name=backup]  %% Copy by name with custom name

Options:
  source_session_id   - Source session UUID (mutually exclusive with source_session_name)
  source_session_name - Source session name (mutually exclusive with source_session_id) 
  target_session_name - Optional custom name for copied session (auto-generated if not provided)

Note: Exactly one source parameter (source_session_id OR source_session_name) must be provided.
      The copied session gets a new UUID, fresh timestamps, and becomes the active session.
      All messages, system prompt, and content are preserved in the deep copy.`
}

// HelpInfo returns structured help information for the session-copy command.
func (c *CopyCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-copy[source_session_id=id OR source_session_name=name, target_session_name=\"name\"]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "source_session_id",
				Description: "Source session UUID (mutually exclusive with source_session_name)",
				Required:    false, // One of the two source options is required
				Type:        "string",
			},
			{
				Name:        "source_session_name",
				Description: "Source session name (mutually exclusive with source_session_id)",
				Required:    false, // One of the two source options is required
				Type:        "string",
			},
			{
				Name:        "target_session_name",
				Description: "Optional custom name for copied session",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-copy[source_session_id=abc12345-678d-90ef]",
				Description: "Copy session by ID with auto-generated target name",
			},
			{
				Command:     "\\session-copy[source_session_name=my_project]",
				Description: "Copy session by name with auto-generated target name",
			},
			{
				Command:     "\\session-copy[source_session_id=abc123, target_session_name=\"experiment\"]",
				Description: "Copy session by ID with custom target name",
			},
			{
				Command:     "\\session-copy[source_session_name=work_session, target_session_name=backup]",
				Description: "Copy session by name with custom target name",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#session_id",
				Description: "New unique identifier of the copied session",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "Name of the copied session (custom or auto-generated)",
				Type:        "system_metadata",
				Example:     "Session 2",
			},
			{
				Name:        "#session_created",
				Description: "Creation timestamp of the copied session",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#source_session_id",
				Description: "Original session UUID that was copied",
				Type:        "system_metadata",
				Example:     "abc12345-678d-90ef",
			},
			{
				Name:        "#source_session_name",
				Description: "Original session name that was copied",
				Type:        "system_metadata",
				Example:     "my_project",
			},
			{
				Name:        "#source_session_created",
				Description: "Original session creation timestamp",
				Type:        "system_metadata",
				Example:     "2024-01-10 09:15:42",
			},
			{
				Name:        "#message_count",
				Description: "Number of messages copied from original session",
				Type:        "system_metadata",
				Example:     "5",
			},
			{
				Name:        "#system_prompt",
				Description: "System prompt copied from original session",
				Type:        "system_metadata",
				Example:     "You are a helpful assistant",
			},
			{
				Name:        "_output",
				Description: "Copy result message with session details",
				Type:        "command_output",
				Example:     "Copied session 'my_project' to 'Session 2' (ID: 550e8400) with 5 messages",
			},
		},
		Notes: []string{
			"Exactly one source parameter (source_session_id OR source_session_name) must be provided",
			"The copied session gets a completely new identity (UUID, timestamps) but preserves all content",
			"All conversation messages are deep copied with new UUIDs but original timestamps",
			"System prompt and all metadata are preserved from the original session",
			"Copied session automatically becomes the active session",
			"Target session name is auto-generated if not provided using default naming convention",
		},
	}
}

// Execute creates a deep copy of an existing session.
// Options:
//   - source_session_id: source session UUID (mutually exclusive with source_session_name)
//   - source_session_name: source session name (mutually exclusive with source_session_id)
//   - target_session_name: optional custom name for copied session
func (c *CopyCommand) Execute(args map[string]string, _ string) error {
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

	// Parse and validate source parameters
	sourceSessionID := args["source_session_id"]
	sourceSessionName := args["source_session_name"]
	targetSessionName := args["target_session_name"]

	// Validate mutually exclusive source parameters
	if sourceSessionID != "" && sourceSessionName != "" {
		return fmt.Errorf("cannot specify both source_session_id and source_session_name - use exactly one")
	}

	if sourceSessionID == "" && sourceSessionName == "" {
		return fmt.Errorf("must specify either source_session_id or source_session_name")
	}

	// Determine source identifier
	var sourceIdentifier string
	if sourceSessionID != "" {
		sourceIdentifier = sourceSessionID
	} else {
		sourceIdentifier = sourceSessionName
	}

	// Copy the session using the service
	copiedSession, err := chatService.CopySession(sourceIdentifier, targetSessionName)
	if err != nil {
		return fmt.Errorf("session copy failed: %w", err)
	}

	// Get the original session for variable storage
	sourceSession, err := chatService.FindSessionByPrefix(sourceIdentifier)
	if err != nil {
		return fmt.Errorf("failed to get source session details: %w", err)
	}

	// Update session-related variables
	if err := c.updateSessionVariables(copiedSession, sourceSession, variableService); err != nil {
		return fmt.Errorf("failed to update session variables: %w", err)
	}

	// Prepare output message
	shortID := copiedSession.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	outputMsg := fmt.Sprintf("Copied session '%s' to '%s' (ID: %s) with %d messages",
		sourceSession.Name, copiedSession.Name, shortID, len(copiedSession.Messages))

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateSessionVariables sets session-related system variables including both copied and source session data
func (c *CopyCommand) updateSessionVariables(copiedSession *neurotypes.ChatSession, sourceSession *neurotypes.ChatSession, variableService *services.VariableService) error {
	// Set session variables (both copied and source)
	variables := map[string]string{
		// Copied session variables (new)
		"#session_id":      copiedSession.ID,
		"#session_name":    copiedSession.Name,
		"#session_created": copiedSession.CreatedAt.Format("2006-01-02 15:04:05"),

		// Source session variables (original)
		"#source_session_id":      sourceSession.ID,
		"#source_session_name":    sourceSession.Name,
		"#source_session_created": sourceSession.CreatedAt.Format("2006-01-02 15:04:05"),

		// Preserved content variables
		"#message_count": fmt.Sprintf("%d", len(copiedSession.Messages)),
		"#system_prompt": copiedSession.SystemPrompt,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CopyCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-copy command: %v", err))
	}
}
