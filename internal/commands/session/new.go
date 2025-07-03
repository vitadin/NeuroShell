// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// NewCommand implements the \session-new command for creating new chat sessions.
// It provides session creation functionality with configurable names and system prompts.
type NewCommand struct{}

// Name returns the command name "session-new" for registration and lookup.
func (c *NewCommand) Name() string {
	return "session-new"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *NewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-new command does.
func (c *NewCommand) Description() string {
	return "Create new chat session for LLM interactions"
}

// Usage returns the syntax and usage examples for the session-new command.
func (c *NewCommand) Usage() string {
	return `\session-new[name=session_name, system=system_prompt] [initial_message]

Examples:
  \session-new                                    # Create anonymous session
  \session-new[name=work]                         # Create named session
  \session-new[system=You are a code reviewer]   # Create session with system prompt
  \session-new[name=debug, system=You are helpful] Debug this code
  
Options:
  name   - Session name (3-64 chars, alphanumeric plus _.- )
  system - System prompt for LLM context`
}

// Execute creates a new chat session with the specified parameters.
// Options:
//   - name: user-friendly session name (optional, auto-generated if omitted)
//   - system: system prompt for LLM context (optional, default helpful assistant)
//   - initial_message: first user message to start conversation (optional)
func (c *NewCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get chat session service
	chatService, err := c.getChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing session variables
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse options
	sessionName := args["name"]
	systemPrompt := args["system"]
	initialMessage := input

	// Interpolate variables in system prompt and initial message
	if systemPrompt != "" {
		systemPrompt, err = variableService.InterpolateString(systemPrompt, ctx)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in system prompt: %w", err)
		}
	}

	if initialMessage != "" {
		initialMessage, err = variableService.InterpolateString(initialMessage, ctx)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in initial message: %w", err)
		}
	}

	// Validate session name if provided
	if sessionName != "" {
		if err := chatService.ValidateSessionName(sessionName); err != nil {
			return fmt.Errorf("invalid session name: %w", err)
		}

		if !chatService.IsSessionNameAvailable(sessionName) {
			return fmt.Errorf("session name '%s' is already in use", sessionName)
		}
	}

	// Create new session
	session, err := chatService.CreateSession(sessionName, systemPrompt, initialMessage)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Update session-related variables
	if err := c.updateSessionVariables(session, variableService, ctx); err != nil {
		return fmt.Errorf("failed to update session variables: %w", err)
	}

	// Prepare output message
	var outputMsg string
	if initialMessage != "" {
		outputMsg = fmt.Sprintf("Created session '%s' (ID: %s) with initial message", session.Name, session.ID[:8])
	} else {
		outputMsg = fmt.Sprintf("Created session '%s' (ID: %s)", session.Name, session.ID[:8])
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg, ctx); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateSessionVariables sets session-related system variables
func (c *NewCommand) updateSessionVariables(session *neurotypes.ChatSession, variableService *services.VariableService, ctx neurotypes.Context) error {
	// Set session variables
	variables := map[string]string{
		"#session_id":      session.ID,
		"#session_name":    session.Name,
		"#message_count":   fmt.Sprintf("%d", len(session.Messages)),
		"#system_prompt":   session.SystemPrompt,
		"#session_created": session.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value, ctx); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// getChatSessionService retrieves the chat session service from the global registry
func (c *NewCommand) getChatSessionService() (*services.ChatSessionService, error) {
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
func (c *NewCommand) getVariableService() (*services.VariableService, error) {
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
	if err := commands.GlobalRegistry.Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-new command: %v", err))
	}
}
