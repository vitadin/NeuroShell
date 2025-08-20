package session

import (
	"fmt"
	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EditWithEditorCommand implements the \session-edit-with-editor command.
// It provides an interactive editor-based interface for editing session messages
// by delegating the actual work to a stdlib script.
type EditWithEditorCommand struct{}

// NewEditWithEditorCommand creates a new EditWithEditorCommand instance.
func NewEditWithEditorCommand() *EditWithEditorCommand {
	return &EditWithEditorCommand{}
}

// Name returns the command name "session-edit-with-editor" for registration and lookup.
func (c *EditWithEditorCommand) Name() string {
	return "session-edit-with-editor"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing with options.
func (c *EditWithEditorCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-edit-with-editor command does.
func (c *EditWithEditorCommand) Description() string {
	return "Edit session messages using external editor"
}

// Usage returns the syntax and usage examples for the session-edit-with-editor command.
func (c *EditWithEditorCommand) Usage() string {
	return `\session-edit-with-editor[idx=N, session=session_name]

Opens your external editor with existing message content, allowing you to edit it
interactively. After saving and closing the editor, the session message will be
updated with the new content.

Examples:
  \session-edit-with-editor[idx=1]                    %% Edit last message with editor
  \session-edit-with-editor[idx=2]                    %% Edit second-to-last message  
  \session-edit-with-editor[idx=.1]                   %% Edit first message with editor
  \session-edit-with-editor[idx=.3]                   %% Edit third message with editor
  \session-edit-with-editor[session=work, idx=1]      %% Edit last message in 'work' session

Options:
  idx     - Message index (required): N for reverse order (1=last), .N for normal order (.1=first)
  session - Session name or ID (optional, defaults to active session)

Note: This command uses your configured editor (${@editor} variable or $EDITOR environment variable).
For GUI editors, use the --wait flag (e.g., "code --wait") to ensure proper operation.`
}

// HelpInfo returns structured help information for the session-edit-with-editor command.
func (c *EditWithEditorCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-edit-with-editor[idx=N, session=session_name]",
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
				Description: "Session name or ID (defaults to active session)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-edit-with-editor[idx=1]",
				Description: "Edit last message using external editor",
			},
			{
				Command:     "\\session-edit-with-editor[idx=.1]",
				Description: "Edit first message using external editor",
			},
			{
				Command:     "\\session-edit-with-editor[session=work, idx=2]",
				Description: "Edit second-to-last message in 'work' session",
			},
		},
		Notes: []string{
			"Opens external editor with existing message content as initial text",
			"Supports both reverse order (idx=N) and normal order (idx=.N) indexing",
			"Message metadata (ID, role, timestamp) is preserved, only content changes",
			"Editor preference: 1) ${@editor} variable, 2) $EDITOR env var, 3) auto-detect",
			"For GUI editors, use --wait flag (e.g., 'code --wait') for proper operation",
		},
	}
}

// Execute delegates to the _session-edit-with-editor stdlib script via stack service.
func (c *EditWithEditorCommand) Execute(args map[string]string, _ string) error {
	// Get stack service for delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Build the command string to execute the stdlib script
	commandStr := "\\_session-edit-with-editor"

	// Add parameters if they exist
	if len(args) > 0 {
		commandStr += "["
		first := true
		for key, value := range args {
			if !first {
				commandStr += ", "
			}
			commandStr += key + "=" + value
			first = false
		}
		commandStr += "]"
	}

	// Delegate to _session-edit-with-editor stdlib script
	stackService.PushCommand(commandStr)

	return nil
}

// IsReadOnly returns false as the session-edit-with-editor command modifies system state.
func (c *EditWithEditorCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&EditWithEditorCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-edit-with-editor command: %v", err))
	}
}
