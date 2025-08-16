// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
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
	return `\session-new[system=system_prompt] [session_name]

Examples:
  \session-new                                    %% Auto-generate name (e.g., "Session 1")
  \session-new work project                       %% Create session named "work project"
  \session-new debug                              %% Create session named "debug"
  \session-new[system=You are a code reviewer] code review  %% Named "code review" with custom system prompt
  \session-new "my project"                       %% Session name with quotes (auto-processed)

Options:
  system - System prompt for LLM context (optional, defaults to helpful assistant)

Note: Session name is optional. If not provided, an auto-generated name is used (e.g., "Session 1").
      Use quotes if the name contains special characters.
      Initial messages can be added later with \send command.`
}

// HelpInfo returns structured help information for the session-new command.
func (c *NewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-new[system=system_prompt] [session_name]",
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
				Command:     "\\session-new",
				Description: "Create session with auto-generated name (e.g., 'Session 1')",
			},
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
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#session_id",
				Description: "Unique identifier of the created session",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "Name of the created session",
				Type:        "system_metadata",
				Example:     "work",
			},
			{
				Name:        "#message_count",
				Description: "Number of messages in the session",
				Type:        "system_metadata",
				Example:     "0",
			},
			{
				Name:        "#system_prompt",
				Description: "System prompt used for the session",
				Type:        "system_metadata",
				Example:     "You are a helpful assistant",
			},
			{
				Name:        "#session_created",
				Description: "Session creation timestamp",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "_output",
				Description: "Command result message",
				Type:        "command_output",
				Example:     "Created session 'work' (ID: 550e8400)",
			},
		},
		Notes: []string{
			"Session name is optional. If not provided, an auto-generated name is used",
			"Use quotes if the name contains special characters or spaces",
			"Variables in session name and system prompt are interpolated",
			"Session automatically becomes active after creation via stack service",
		},
	}
}

// Execute creates a new chat session with the specified parameters.
// The input parameter is used as the session name (required).
// Options:
//   - system: system prompt for LLM context (optional, default helpful assistant)
func (c *NewCommand) Execute(args map[string]string, input string) error {

	// Get chat session service
	chatService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing session variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Note: Variable interpolation is now handled by the state machine before commands execute

	// Parse arguments - session name comes from input, not from options
	sessionName := input
	systemPrompt := args["system"]

	// Auto-generate session name if not provided (improved UX)
	if sessionName == "" {
		sessionName = chatService.GenerateDefaultSessionName()
		printer := printing.NewDefaultPrinter()
		printer.Info(fmt.Sprintf("Auto-generated session name: '%s'", sessionName))
	}

	// Note: Variable interpolation for session name and system prompt is handled by state machine

	// Create new session (no initial message support)
	session, err := chatService.CreateSession(sessionName, systemPrompt, "")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Auto-push session activation command to stack service for seamless UX
	// Use precise ID-based activation to avoid any ambiguity
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := fmt.Sprintf("\\silent \\session-activate[id=true] %s", session.ID)
		stackService.PushCommand(activateCommand)
	}

	// Update session-related variables (not active session variables - that's done by session-activate)
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
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

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

func init() {
	if err := commands.GetGlobalRegistry().Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-new command: %v", err))
	}
}
