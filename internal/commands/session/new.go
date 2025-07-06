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
	return `\session-new[system=system_prompt] session_name

Examples:
  \session-new work project                       %% Create session named "work project"
  \session-new debug                              %% Create session named "debug"
  \session-new[system=You are a code reviewer] code review  %% Named "code review" with custom system prompt
  \session-new "my project"                       %% Session name with quotes (auto-processed)
  
Options:
  system - System prompt for LLM context (optional, defaults to helpful assistant)
  
Note: Session name is required and taken from the input parameter.
      Use quotes if the name contains special characters.
      Initial messages can be added later with \send command.`
}

// HelpInfo returns structured help information for the session-new command.
func (c *NewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-new[system=system_prompt] session_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "system",
				Description: "System prompt for LLM context",
				Required:    false,
				Type:        "string",
				Default:     "helpful assistant",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-new work",
				Description: "Create new session named 'work'",
			},
			{
				Command:     "\\session-new[system=You are a code reviewer] code-review",
				Description: "Create session with custom system prompt",
			},
			{
				Command:     "\\session-new \"my project\"",
				Description: "Create session with quoted name containing spaces",
			},
			{
				Command:     "\\session-new debug-${@date}",
				Description: "Create session with interpolated variables in name",
			},
		},
		Notes: []string{
			"Session name is required and taken from the input parameter",
			"Use quotes if the name contains special characters or spaces",
			"Variables in session name and system prompt are interpolated",
			"Session becomes active immediately after creation",
			"Session ID and metadata are stored in system variables (${#session_id}, etc.)",
			"Initial messages can be added later with \\send command",
		},
	}
}

// Execute creates a new chat session with the specified parameters.
// The input parameter is used as the session name (required).
// Options:
//   - system: system prompt for LLM context (optional, default helpful assistant)
func (c *NewCommand) Execute(args map[string]string, input string) error {

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

	// Get interpolation service for variable interpolation
	interpolationService, err := c.getInterpolationService()
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	// Parse arguments - session name comes from input, not from options
	sessionName := input
	systemPrompt := args["system"]

	// Session name is required
	if sessionName == "" {
		return fmt.Errorf("session name is required\n\nUsage: %s", c.Usage())
	}

	// Interpolate variables in session name and system prompt
	sessionName, err = interpolationService.InterpolateString(sessionName)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in session name: %w", err)
	}

	if systemPrompt != "" {
		systemPrompt, err = interpolationService.InterpolateString(systemPrompt)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in system prompt: %w", err)
		}
	}

	// Create new session (no initial message support)
	session, err := chatService.CreateSession(sessionName, systemPrompt, "")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Update session-related variables
	if err := c.updateSessionVariables(session, variableService); err != nil {
		return fmt.Errorf("failed to update session variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Created session '%s' (ID: %s)", session.Name, session.ID[:8])

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateSessionVariables sets session-related system variables
func (c *NewCommand) updateSessionVariables(session *neurotypes.ChatSession, variableService *services.VariableService) error {
	// Set session variables
	variables := map[string]string{
		"#session_id":      session.ID,
		"#session_name":    session.Name,
		"#message_count":   fmt.Sprintf("%d", len(session.Messages)),
		"#system_prompt":   session.SystemPrompt,
		"#session_created": session.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
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

// getInterpolationService retrieves the interpolation service from the global registry.
func (c *NewCommand) getInterpolationService() (*services.InterpolationService, error) {
	service, err := services.GetGlobalRegistry().GetService("interpolation")
	if err != nil {
		return nil, err
	}

	interpolationService, ok := service.(*services.InterpolationService)
	if !ok {
		return nil, fmt.Errorf("interpolation service has incorrect type")
	}

	return interpolationService, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-new command: %v", err))
	}
}
