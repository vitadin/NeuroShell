package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ShellPromptCommand configures the shell prompt display
type ShellPromptCommand struct{}

// Name returns the command name "shell-prompt" for registration and lookup.
func (c *ShellPromptCommand) Name() string {
	return "shell-prompt"
}

// Description returns a brief description of what the shell-prompt command does.
func (c *ShellPromptCommand) Description() string {
	return "Configure the shell prompt display"
}

// Usage returns the syntax and usage examples for the shell-prompt command.
func (c *ShellPromptCommand) Usage() string {
	return `\shell-prompt[lines=N, line1="template", line2="template", ...]

Configure the shell prompt appearance with 1-5 lines.

Options:
  lines=N          Set number of prompt lines (1-5)
  line1="template" Set template for first line
  line2="template" Set template for second line
  line3="template" Set template for third line
  line4="template" Set template for fourth line
  line5="template" Set template for fifth line

Examples:
  \shell-prompt[lines=1, line1="neuro> "]
  \shell-prompt[lines=2, line1="${@pwd} [${#session_name}]", line2="❯ "]
  \shell-prompt[lines=3, line1="┌─[${@time}]", line2="├─[${@pwd}]", line3="└➤ "]

Available variables:
  ${@pwd}           Current working directory
  ${@user}          Current username
  ${@date}          Current date
  ${@time}          Current time
  ${#session_name}  Active session name
  ${#message_count} Number of messages in session
  ${#active_model}  Currently active model`
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ShellPromptCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// HelpInfo returns structured help information for the shell-prompt command.
func (c *ShellPromptCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "lines",
				Description: "Number of prompt lines (1-5)",
				Required:    false,
				Type:        "number",
			},
			{
				Name:        "line1",
				Description: "Template for first prompt line",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "line2",
				Description: "Template for second prompt line",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\shell-prompt[lines=2, line1=\"${@pwd}\", line2=\"❯ \"]",
				Description: "Set two-line prompt with path and arrow",
			},
			{
				Command:     "\\shell-prompt",
				Description: "Show current prompt configuration",
			},
		},
	}
}

// Execute configures the shell prompt display with specified options.
func (c *ShellPromptCommand) Execute(options map[string]string, _ string) error {
	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Handle setting number of lines
	if lines, ok := options["lines"]; ok {
		if n, err := strconv.Atoi(lines); err == nil && n >= 1 && n <= 5 {
			if err := ctx.SetVariable("_prompt_lines_count", lines); err != nil {
				return fmt.Errorf("failed to set prompt lines count: %w", err)
			}
			fmt.Printf("Prompt lines set to %s\n", lines)
		} else {
			return fmt.Errorf("invalid lines value: must be 1-5, got: %s", lines)
		}
	}

	// Handle setting specific lines
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("line%d", i)
		if template, ok := options[key]; ok {
			varName := fmt.Sprintf("_prompt_line%d", i)
			if err := ctx.SetVariable(varName, template); err != nil {
				return fmt.Errorf("failed to set prompt line %d: %w", i, err)
			}
			fmt.Printf("Prompt line %d set to: %s\n", i, template)
		}
	}

	// If no options provided, show current configuration
	if len(options) == 0 {
		return c.showCurrentConfiguration(ctx)
	}

	return nil
}

func (c *ShellPromptCommand) showCurrentConfiguration(ctx *neuroshellcontext.NeuroContext) error {
	fmt.Println("Current shell prompt configuration:")

	// Get lines count
	linesCount, _ := ctx.GetVariable("_prompt_lines_count")
	if linesCount == "" {
		linesCount = "1" // default
	}
	fmt.Printf("  Lines: %s\n", linesCount)

	// Show configured lines
	count, _ := strconv.Atoi(linesCount)
	if count < 1 || count > 5 {
		count = 1
	}

	for i := 1; i <= count; i++ {
		varName := fmt.Sprintf("_prompt_line%d", i)
		template, _ := ctx.GetVariable(varName)
		if template == "" && i == 1 {
			template = "neuro> " // default
		}
		if template != "" {
			fmt.Printf("  Line %d: %s\n", i, template)
		}
	}

	return nil
}

// IsReadOnly returns false as this command modifies prompt configuration.
func (c *ShellPromptCommand) IsReadOnly() bool {
	return false
}

// ShellPromptShowCommand displays the current prompt configuration
type ShellPromptShowCommand struct{}

// Name returns the command name "shell-prompt-show" for registration and lookup.
func (c *ShellPromptShowCommand) Name() string {
	return "shell-prompt-show"
}

// Description returns a brief description of what the shell-prompt-show command does.
func (c *ShellPromptShowCommand) Description() string {
	return "Display current shell prompt configuration"
}

// Usage returns the syntax and usage examples for the shell-prompt-show command.
func (c *ShellPromptShowCommand) Usage() string {
	return `\shell-prompt-show

Display the current shell prompt configuration including templates and preview.

Example:
  \shell-prompt-show`
}

// ParseMode returns ParseModeRaw for no argument parsing.
func (c *ShellPromptShowCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// HelpInfo returns structured help information for the shell-prompt-show command.
func (c *ShellPromptShowCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\shell-prompt-show",
				Description: "Display current shell prompt configuration",
			},
		},
	}
}

// Execute displays the current shell prompt configuration.
func (c *ShellPromptShowCommand) Execute(_ map[string]string, _ string) error {
	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	fmt.Println("Shell Prompt Configuration:")
	fmt.Println("==========================")

	// Get and display configuration
	linesCount, _ := ctx.GetVariable("_prompt_lines_count")
	if linesCount == "" {
		linesCount = "1"
	}
	fmt.Printf("Number of lines: %s\n\n", linesCount)

	// Show each line template and preview
	count, _ := strconv.Atoi(linesCount)
	if count < 1 || count > 5 {
		count = 1
	}

	fmt.Println("Templates:")
	for i := 1; i <= count; i++ {
		varName := fmt.Sprintf("_prompt_line%d", i)
		template, _ := ctx.GetVariable(varName)
		if template == "" && i == 1 {
			template = "neuro> "
		}
		if template != "" {
			fmt.Printf("  Line %d: %s\n", i, template)
		}
	}

	// Show preview
	fmt.Println("\nPreview:")
	for i := 1; i <= count; i++ {
		varName := fmt.Sprintf("_prompt_line%d", i)
		template, _ := ctx.GetVariable(varName)
		if template == "" && i == 1 {
			template = "neuro> "
		}
		if template != "" {
			// Use context's interpolation to show preview
			interpolated := ctx.InterpolateVariables(template)
			fmt.Printf("  %s\n", interpolated)
		}
	}

	return nil
}

// IsReadOnly returns true as this command only displays information.
func (c *ShellPromptShowCommand) IsReadOnly() bool {
	return true
}

// ShellPromptPresetCommand applies preset prompt configurations
type ShellPromptPresetCommand struct{}

// Name returns the command name "shell-prompt-preset" for registration and lookup.
func (c *ShellPromptPresetCommand) Name() string {
	return "shell-prompt-preset"
}

// Description returns a brief description of what the shell-prompt-preset command does.
func (c *ShellPromptPresetCommand) Description() string {
	return "Apply preset shell prompt configurations"
}

// Usage returns the syntax and usage examples for the shell-prompt-preset command.
func (c *ShellPromptPresetCommand) Usage() string {
	return `\shell-prompt-preset[style=NAME]

Apply a preset prompt configuration.

Available presets:
  minimal     - Single line minimal prompt ("> ")
  default     - Two-line default prompt with path and session
  developer   - Two-line developer prompt with git and status
  powerline   - Three-line powerline-style prompt

Examples:
  \shell-prompt-preset[style=minimal]
  \shell-prompt-preset[style=default]
  \shell-prompt-preset[style=developer]
  \shell-prompt-preset[style=powerline]`
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ShellPromptPresetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// HelpInfo returns structured help information for the shell-prompt-preset command.
func (c *ShellPromptPresetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "style",
				Description: "Preset style name (minimal, default, developer, powerline)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\shell-prompt-preset[style=default]",
				Description: "Apply default two-line prompt preset",
			},
			{
				Command:     "\\shell-prompt-preset[style=powerline]",
				Description: "Apply powerline three-line prompt preset",
			},
		},
	}
}

// Execute applies the specified preset shell prompt configuration.
func (c *ShellPromptPresetCommand) Execute(options map[string]string, _ string) error {
	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	style, ok := options["style"]
	if !ok {
		return fmt.Errorf("style option is required")
	}

	style = strings.ToLower(style)

	switch style {
	case "minimal":
		if err := ctx.SetVariable("_prompt_lines_count", "1"); err != nil {
			return fmt.Errorf("failed to set lines count: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line1", "> "); err != nil {
			return fmt.Errorf("failed to set prompt line: %w", err)
		}
		fmt.Println("Applied minimal prompt preset")

	case "default":
		if err := ctx.SetVariable("_prompt_lines_count", "2"); err != nil {
			return fmt.Errorf("failed to set lines count: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line1", "${@pwd} [${#session_name:-no-session}]"); err != nil {
			return fmt.Errorf("failed to set prompt line 1: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line2", "neuro> "); err != nil {
			return fmt.Errorf("failed to set prompt line 2: %w", err)
		}
		fmt.Println("Applied default prompt preset")

	case "developer":
		if err := ctx.SetVariable("_prompt_lines_count", "2"); err != nil {
			return fmt.Errorf("failed to set lines count: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line1", "${@pwd} ${@status}"); err != nil {
			return fmt.Errorf("failed to set prompt line 1: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line2", "❯ "); err != nil {
			return fmt.Errorf("failed to set prompt line 2: %w", err)
		}
		fmt.Println("Applied developer prompt preset")

	case "powerline":
		if err := ctx.SetVariable("_prompt_lines_count", "3"); err != nil {
			return fmt.Errorf("failed to set lines count: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line1", "┌─[${@user}@${@hostname:-local}]-[${@time}]"); err != nil {
			return fmt.Errorf("failed to set prompt line 1: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line2", "├─[${#session_name:-no-session}:${#message_count:-0}]-[${#active_model:-none}]"); err != nil {
			return fmt.Errorf("failed to set prompt line 2: %w", err)
		}
		if err := ctx.SetVariable("_prompt_line3", "└─➤ "); err != nil {
			return fmt.Errorf("failed to set prompt line 3: %w", err)
		}
		fmt.Println("Applied powerline prompt preset")

	default:
		return fmt.Errorf("unknown preset style: %s. Available: minimal, default, developer, powerline", style)
	}

	return nil
}

// IsReadOnly returns false as this command modifies prompt configuration.
func (c *ShellPromptPresetCommand) IsReadOnly() bool {
	return false
}

// ShellPromptPreviewCommand previews the current prompt with current context
type ShellPromptPreviewCommand struct{}

// Name returns the command name "shell-prompt-preview" for registration and lookup.
func (c *ShellPromptPreviewCommand) Name() string {
	return "shell-prompt-preview"
}

// Description returns a brief description of what the shell-prompt-preview command does.
func (c *ShellPromptPreviewCommand) Description() string {
	return "Preview the current shell prompt with current context"
}

// Usage returns the syntax and usage examples for the shell-prompt-preview command.
func (c *ShellPromptPreviewCommand) Usage() string {
	return `\shell-prompt-preview

Show how the current shell prompt would appear with interpolated variables.

Example:
  \shell-prompt-preview`
}

// ParseMode returns ParseModeRaw for no argument parsing.
func (c *ShellPromptPreviewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// HelpInfo returns structured help information for the shell-prompt-preview command.
func (c *ShellPromptPreviewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\shell-prompt-preview",
				Description: "Preview current prompt with interpolated variables",
			},
		},
	}
}

// Execute previews the current shell prompt with interpolated variables.
func (c *ShellPromptPreviewCommand) Execute(_ map[string]string, _ string) error {
	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Get the shell prompt service to retrieve templates
	shellPromptService, err := services.GetGlobalRegistry().GetService("shell_prompt")
	if err != nil {
		return fmt.Errorf("shell prompt service not available: %w", err)
	}

	promptService := shellPromptService.(*services.ShellPromptService)
	lines, err := promptService.GetPromptLines()
	if err != nil {
		return fmt.Errorf("failed to get prompt lines: %w", err)
	}

	fmt.Println("Prompt Preview:")
	fmt.Println("===============")

	// Show each line with interpolation
	for i, template := range lines {
		// Interpolate variables using context
		interpolated := ctx.InterpolateVariables(template)
		fmt.Printf("%s\n", interpolated)

		// Don't add extra newline after the last line
		if i == len(lines)-1 {
			// Add cursor to show where input would go
			fmt.Print("█")
		}
	}
	fmt.Println() // Final newline

	return nil
}

// IsReadOnly returns true as this command only displays a preview.
func (c *ShellPromptPreviewCommand) IsReadOnly() bool {
	return true
}

// Register all shell prompt commands
func init() {
	if err := commands.GlobalRegistry.Register(&ShellPromptCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt command: %v", err))
	}
	if err := commands.GlobalRegistry.Register(&ShellPromptShowCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-show command: %v", err))
	}
	if err := commands.GlobalRegistry.Register(&ShellPromptPresetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-preset command: %v", err))
	}
	if err := commands.GlobalRegistry.Register(&ShellPromptPreviewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-preview command: %v", err))
	}
}
