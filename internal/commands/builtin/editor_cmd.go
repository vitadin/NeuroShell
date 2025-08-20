package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
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
	return `\editor [initial_text] - Open external editor for input composition

Opens your configured external editor (from ${_editor} variable, $EDITOR env var, or auto-detected) 
to compose a multi-line input. When you save and exit the editor, the content will be
stored in the ${_output} variable.

If initial_text is provided, the editor will open with that text as starting content.

The editor preference can be configured with:
  \set[_editor="your-preferred-editor"]

Examples:
  \editor                              - Open empty editor
  \editor Write a blog post about AI  - Open editor with initial text
  \editor Draft email to customer     - Start with prompt text
  \send ${_output}                    - Send the editor content
  \set[_editor="code --wait"]         - Configure VS Code as editor
  
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
		Usage:       "\\editor [initial_text]",
		ParseMode:   e.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\editor",
				Description: "Open empty editor for composing multi-line input",
			},
			{
				Command:     "\\editor Write a detailed analysis",
				Description: "Open editor with initial text as starting content",
			},
			{
				Command:     "\\editor Draft email to customer",
				Description: "Start editing with prompt text pre-filled",
			},
			{
				Command:     "\\set[_editor=\"code --wait\"]",
				Description: "Configure VS Code as your preferred editor",
			},
			{
				Command:     "\\set[myvar=\"${_output}\"]",
				Description: "Store editor content in a custom variable",
			},
		},
		Notes: []string{
			"If initial_text is provided, editor opens with that text as starting content",
			"Without initial text, editor opens empty - use \\editor ${_output} to edit existing content",
			"Content is stored in ${_output} variable when editor is saved and closed",
			"Editor preference: 1) ${_editor} variable, 2) $EDITOR env var, 3) auto-detect",
			"Supports vim, nano, code, subl, atom, and other common editors",
			"Use 'editor --wait' flag for GUI editors to wait for file closure",
		},
	}
}

// Execute opens the external editor and stores the resulting content in ${_output}.
// If input text is provided, it will be used as initial content in the editor.
func (e *EditorCommand) Execute(args map[string]string, input string) error {
	logger.Debug("Executing editor command", "args", args, "input", input)

	// Get the editor service
	es, err := services.GetGlobalEditorService()
	if err != nil {
		return fmt.Errorf("editor service not available: %w", err)
	}

	// Open the editor with or without initial content
	var content string
	if input != "" {
		content, err = es.OpenEditorWithContent(input)
	} else {
		content, err = es.OpenEditor()
	}
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

	printer := printing.NewDefaultPrinter()
	if content == "" {
		printer.Info("Editor returned empty content - stored empty string in ${_output}")
	} else {
		printer.Success(fmt.Sprintf("Editor content stored in ${_output} (%d characters)", len(content)))
		printer.Info("Use \\send ${_output} to send the content")
	}

	return nil
}

// IsReadOnly returns false as the editor command modifies system state.
func (e *EditorCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EditorCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register editor command: %v", err))
	}
}
