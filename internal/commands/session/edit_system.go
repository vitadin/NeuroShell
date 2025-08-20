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

// EditSystemCommand implements the \session-edit-system command for editing session system prompts.
// It provides system prompt editing functionality with session lookup and variable management.
type EditSystemCommand struct{}

// Name returns the command name "session-edit-system" for registration and lookup.
func (c *EditSystemCommand) Name() string {
	return "session-edit-system"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *EditSystemCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-edit-system command does.
func (c *EditSystemCommand) Description() string {
	return "Edit session system prompt"
}

// Usage returns the syntax and usage examples for the session-edit-system command.
func (c *EditSystemCommand) Usage() string {
	return `\session-edit-system[session=session_id] new_system_prompt
\session-edit-system new_system_prompt

Examples:
  \session-edit-system You are a helpful programming assistant                     %% Edit active session system prompt
  \session-edit-system[session=work] You are a code review assistant              %% Edit specific session
  \session-edit-system ""                                                         %% Clear system prompt (empty)
  \session-edit-system[session=main] ${custom_prompt}                             %% Using variable

Options:
  session - Session name or ID (optional, defaults to active session)

Input: New system prompt content (can be empty to clear the system prompt)

Note: System prompts provide context and instructions to LLM agents.
      Empty system prompts remove all system-level instructions.
      Changes are saved immediately and affect future LLM interactions.`
}

// HelpInfo returns structured help information for the session-edit-system command.
func (c *EditSystemCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-edit-system[session=session_id] new_system_prompt",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "session",
				Description: "Session name or ID (optional, defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-edit-system You are a helpful assistant",
				Description: "Edit the active session's system prompt",
			},
			{
				Command:     "\\session-edit-system[session=work] You are a code reviewer",
				Description: "Edit system prompt for a specific session",
			},
			{
				Command:     "\\session-edit-system \"\"",
				Description: "Clear the system prompt (set to empty)",
			},
			{
				Command:     "\\session-edit-system[session=chat] ${my_prompt}",
				Description: "Set system prompt using a variable",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#session_id",
				Description: "ID of the session that was modified",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#session_name",
				Description: "Name of the session that was modified",
				Type:        "system_metadata",
				Example:     "work_session",
			},
			{
				Name:        "#old_system_prompt",
				Description: "Previous system prompt content",
				Type:        "system_metadata",
				Example:     "You are a helpful assistant",
			},
			{
				Name:        "#new_system_prompt",
				Description: "New system prompt content",
				Type:        "system_metadata",
				Example:     "You are a code reviewer",
			},
			{
				Name:        "_output",
				Description: "Edit operation result message",
				Type:        "command_output",
				Example:     "Updated system prompt for session 'work_session'",
			},
		},
		Notes: []string{
			"Session parameter is optional and defaults to active session",
			"System prompts can be empty (cleared) or contain detailed instructions",
			"Changes take effect immediately for future LLM interactions",
			"Existing conversation history is preserved",
			"Use smart prefix matching for session lookup",
		},
	}
}

// Execute edits the system prompt of the specified session.
// Options:
//   - session: Session name or ID (optional, defaults to active session)
func (c *EditSystemCommand) Execute(args map[string]string, input string) error {
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
	sessionID := args["session"]

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

	// Store the original system prompt for reference
	originalSystemPrompt := targetSession.SystemPrompt

	// Update the system prompt using the service
	err = chatService.UpdateSystemPrompt(targetSession.ID, input)
	if err != nil {
		return fmt.Errorf("failed to update system prompt: %w", err)
	}

	// Get the updated session to confirm changes
	updatedSession, err := chatService.GetSessionByNameOrID(targetSession.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated session: %w", err)
	}

	// Update result variables
	if err := c.updateSystemVariables(updatedSession, originalSystemPrompt, input, variableService); err != nil {
		return fmt.Errorf("failed to update system variables: %w", err)
	}

	// Prepare output message
	var outputMsg string
	if input == "" {
		outputMsg = fmt.Sprintf("Cleared system prompt for session '%s'", updatedSession.Name)
	} else {
		outputMsg = fmt.Sprintf("Updated system prompt for session '%s'", updatedSession.Name)
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	// Trigger auto-save if enabled
	chatService.TriggerAutoSave(targetSession.ID)

	return nil
}

// updateSystemVariables sets system prompt related system variables
func (c *EditSystemCommand) updateSystemVariables(session *neurotypes.ChatSession, oldSystemPrompt, newSystemPrompt string, variableService *services.VariableService) error {
	// Set session variables
	variables := map[string]string{
		"#session_id":        session.ID,
		"#session_name":      session.Name,
		"#old_system_prompt": oldSystemPrompt,
		"#new_system_prompt": newSystemPrompt,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// IsReadOnly returns false as the session-edit-system command modifies system state.
func (c *EditSystemCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&EditSystemCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-edit-system command: %v", err))
	}
}
