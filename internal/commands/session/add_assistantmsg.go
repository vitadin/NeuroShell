package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// AddAssistantMessageCommand implements the \session-add-assistantmsg command for adding assistant messages to sessions.
// It provides the ability to add assistant responses to specified sessions for LLM conversation workflows.
type AddAssistantMessageCommand struct{}

// Name returns the command name "session-add-assistantmsg" for registration and lookup.
func (c *AddAssistantMessageCommand) Name() string {
	return "session-add-assistantmsg"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *AddAssistantMessageCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-add-assistantmsg command does.
func (c *AddAssistantMessageCommand) Description() string {
	return "Add assistant message to specified session"
}

// Usage returns the syntax and usage examples for the session-add-assistantmsg command.
func (c *AddAssistantMessageCommand) Usage() string {
	return `\session-add-assistantmsg[session=session_id] response_content
\session-add-assistantmsg response_content

Examples:
  \session-add-assistantmsg I'm doing well, thank you!                 %% Use active session
  \session-add-assistantmsg[session=${session_id}] I'm doing well, thank you!
  \session-add-assistantmsg[session=work-session] ${llm_response}
  \session-add-assistantmsg ${_output}

Note: Response content is required. Session parameter is optional and defaults to active session.
      This command is typically used after LLM calls to store assistant responses.`
}

// HelpInfo returns structured help information for the session-add-assistantmsg command.
func (c *AddAssistantMessageCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       `\session-add-assistantmsg[session=session_id] response_content or \session-add-assistantmsg response_content`,
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "session",
				Description: "Session ID or name to add the assistant response to (optional, defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\session-add-assistantmsg I'm doing well, thank you!`,
				Description: "Add assistant response to active session",
			},
			{
				Command:     `\session-add-assistantmsg[session=${session_id}] I'm doing well, thank you!`,
				Description: "Add assistant response to session specified by variable",
			},
			{
				Command:     `\session-add-assistantmsg ${_output}`,
				Description: "Add LLM output to active session",
			},
		},
		Notes: []string{
			"Response content is required",
			"Session parameter is optional and defaults to active session",
			"Session can be specified by ID or name",
			"Response content is taken from the input parameter",
			"Typically used after \\llm-call to store assistant responses",
			"Messages are timestamped and added to session history",
			"Updates message history variables (${1}, ${2}, etc.)",
			"Adding a message to a session makes it the active session",
		},
	}
}

// Execute adds an assistant message to the specified session or active session if none specified.
// The session is specified via the 'session' parameter or defaults to active session.
func (c *AddAssistantMessageCommand) Execute(args map[string]string, input string) error {
	// Validate response content
	if input == "" {
		return fmt.Errorf("response content is required. Usage: %s", c.Usage())
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

	// Add assistant message to session (this will auto-activate the session)
	err = chatService.AddMessage(sessionID, "assistant", input)
	if err != nil {
		return fmt.Errorf("failed to add assistant message to session '%s': %w", sessionID, err)
	}

	// Get variable service for updating message history variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Update message history variables - latest assistant response becomes ${1}
	err = variableService.Set("1", input)
	if err != nil {
		return fmt.Errorf("failed to update message history variable: %w", err)
	}

	// Output confirmation
	fmt.Printf("Added assistant message to session '%s'\n", sessionID)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&AddAssistantMessageCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-add-assistantmsg command: %v", err))
	}
}
