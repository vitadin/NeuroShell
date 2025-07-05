package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EditorCommand allows users to open an external editor for input composition.
type EditorCommand struct{}

// NewEditorCommand creates a new EditorCommand instance.
func NewEditorCommand() *EditorCommand {
	return &EditorCommand{}
}

// Name returns the command name "editor".
func (e *EditorCommand) Name() string {
	return "editor"
}

// ParseMode returns the parsing mode for the editor command.
func (e *EditorCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of the editor command.
func (e *EditorCommand) Description() string {
	return "Open external editor for composing input"
}

// Usage returns the usage information for the editor command.
func (e *EditorCommand) Usage() string {
	return `\editor - Open external editor for input composition

Opens your configured external editor (from $EDITOR or auto-detected) to compose
a multi-line input. When you save and exit the editor, the content will be
stored in the ${_output} variable.

The editor preference can be configured with:
  \set[@editor="your-preferred-editor"]

Examples:
  \editor                    - Open editor and store content in ${_output}
  \send ${_output}          - Send the editor content
  \set[@editor="code --wait"] - Configure VS Code as editor
  
After editing, you can use the content with:
  \send ${_output}          - Send the editor content to LLM
  \bash[echo "${_output}"]  - Use content in bash command
  \set[myvar="${_output}"]  - Store in another variable`
}

// HelpInfo returns structured help information for the editor command.
func (e *EditorCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     e.Name(),
		Description: e.Description(),
		Usage:       "\\editor",
		ParseMode:   e.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\editor",
				Description: "Open external editor for composing multi-line input",
			},
			{
				Command:     "\\set[@editor=\"code --wait\"]",
				Description: "Configure VS Code as your preferred editor",
			},
			{
				Command:     "\\editor",
				Description: "Edit content, then use with \\send ${_output}",
			},
			{
				Command:     "\\set[myvar=\"${_output}\"]",
				Description: "Store editor content in a custom variable",
			},
		},
		Notes: []string{
			"Editor opens with ${_output} content, or a default template if empty",
			"Content is stored in ${_output} variable when editor is saved and closed",
			"Editor preference: 1) ${@editor} variable, 2) $EDITOR env var, 3) auto-detect",
			"Supports vim, nano, code, subl, atom, and other common editors",
			"Use 'editor --wait' flag for GUI editors to wait for file closure",
		},
	}
}

// Execute opens the external editor and stores the resulting content in ${_output}.
func (e *EditorCommand) Execute(args map[string]string, _ string, ctx neurotypes.Context) error {
	logger.Debug("Executing editor command", "args", args)

	// Set global context for service access
	context.SetGlobalContext(ctx)

	// Get the editor service
	editorService, err := services.GetGlobalRegistry().GetService("editor")
	if err != nil {
		return fmt.Errorf("editor service not available: %w", err)
	}

	es := editorService.(*services.EditorService)

	// Open the editor and get content
	content, err := es.OpenEditor()
	if err != nil {
		return fmt.Errorf("editor operation failed: %w", err)
	}

	// Get the variable service to set system variable
	variableService, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	vs := variableService.(*services.VariableService)

	// Store the content in _output system variable
	if err := vs.SetSystemVariable("_output", content); err != nil {
		return fmt.Errorf("failed to store editor content: %w", err)
	}

	if content == "" {
		fmt.Println("Editor returned empty content - stored empty string in ${_output}")
	} else {
		fmt.Printf("Editor content stored in ${_output} (%d characters)\n", len(content))
		fmt.Printf("Use \\send ${_output} to send the content\n")
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&EditorCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register editor command: %v", err))
	}
}
