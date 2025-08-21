package shell

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// PromptPresetCommand applies preset prompt configurations
type PromptPresetCommand struct{}

// Name returns the command name "shell-prompt-preset" for registration and lookup.
func (c *PromptPresetCommand) Name() string {
	return "shell-prompt-preset"
}

// Description returns a brief description of what the shell-prompt-preset command does.
func (c *PromptPresetCommand) Description() string {
	return "Apply preset shell prompt configurations"
}

// Usage returns the syntax and usage examples for the shell-prompt-preset command.
func (c *PromptPresetCommand) Usage() string {
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
func (c *PromptPresetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// HelpInfo returns structured help information for the shell-prompt-preset command.
func (c *PromptPresetCommand) HelpInfo() neurotypes.HelpInfo {
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
func (c *PromptPresetCommand) Execute(options map[string]string, _ string) error {
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
func (c *PromptPresetCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GlobalRegistry.Register(&PromptPresetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-preset command: %v", err))
	}
}
