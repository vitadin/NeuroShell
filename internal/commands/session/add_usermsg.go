package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// AddUserMessageCommand implements the \session-add-usermsg command for adding user messages to sessions.
// It provides the ability to add user messages to specified sessions for LLM conversation workflows.
type AddUserMessageCommand struct{}

// Name returns the command name "session-add-usermsg" for registration and lookup.
func (c *AddUserMessageCommand) Name() string {
	return "session-add-usermsg"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *AddUserMessageCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-add-usermsg command does.
func (c *AddUserMessageCommand) Description() string {
	return "Add user message to specified session"
}

// Usage returns the syntax and usage examples for the session-add-usermsg command.
func (c *AddUserMessageCommand) Usage() string {
	return `\session-add-usermsg[session=session_id] message_content
\session-add-usermsg message_content

Examples:
  \session-add-usermsg Hello, how are you?                             %% Use active session
  \session-add-usermsg[session=${session_id}] Hello, how are you?
  \session-add-usermsg[session=work-session] Can you help me debug this code?
  \session-add-usermsg[session=${active_session}] ${user_input}

Note: Message content is required. Session parameter is optional and defaults to active session.
      This command is typically used before making LLM calls to add user context.`
}

// HelpInfo returns structured help information for the session-add-usermsg command.
func (c *AddUserMessageCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       `\session-add-usermsg[session=session_id] message_content or \session-add-usermsg message_content`,
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "session",
				Description: "Session ID or name to add the user message to (optional, defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\session-add-usermsg Hello, how are you?`,
				Description: "Add user message to active session",
			},
			{
				Command:     `\session-add-usermsg[session=${session_id}] Hello, how are you?`,
				Description: "Add user message to session specified by variable",
			},
			{
				Command:     `\session-add-usermsg[session=work-session] Can you help me debug this code?`,
				Description: "Add user message to named session",
			},
		},
		Notes: []string{
			"Message content is required",
			"Session parameter is optional and defaults to active session",
			"Session can be specified by ID or name",
			"Message content is taken from the input parameter",
			"Typically used before \\llm-call to add user context to conversation",
			"Messages are timestamped and added to session history",
			"Adding a message to a session makes it the active session",
		},
	}
}

// Execute adds a user message to the specified session or active session if none specified.
// The session is specified via the 'session' parameter or defaults to active session.
func (c *AddUserMessageCommand) Execute(args map[string]string, input string) error {
	// Validate message content
	if input == "" {
		return fmt.Errorf("message content is required. Usage: %s", c.Usage())
	}

	// Get chat session service
	chatService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Determine session - use provided session or default to active session
	sessionID := args["session"]
	if sessionID == "" {
		// No session specified, use active session
		activeSession, err := chatService.GetActiveSession()
		if err != nil {
			return fmt.Errorf("no session specified and no active session found: %w. Usage: %s", err, c.Usage())
		}
		sessionID = activeSession.ID
	}

	// Add user message to session (this will auto-activate the session)
	err = chatService.AddMessage(sessionID, "user", input)
	if err != nil {
		return fmt.Errorf("failed to add user message to session '%s': %w", sessionID, err)
	}

	// Output confirmation
	fmt.Printf("Added user message to session '%s'\n", sessionID)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&AddUserMessageCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-add-usermsg command: %v", err))
	}
}
