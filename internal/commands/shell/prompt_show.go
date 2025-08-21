package shell

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// PromptShowCommand displays the current prompt configuration
type PromptShowCommand struct{}

// Name returns the command name "shell-prompt-show" for registration and lookup.
func (c *PromptShowCommand) Name() string {
	return "shell-prompt-show"
}

// Description returns a brief description of what the shell-prompt-show command does.
func (c *PromptShowCommand) Description() string {
	return "Display current shell prompt configuration"
}

// Usage returns the syntax and usage examples for the shell-prompt-show command.
func (c *PromptShowCommand) Usage() string {
	return `\shell-prompt-show

Display the current shell prompt configuration including templates and preview.

Example:
  \shell-prompt-show`
}

// ParseMode returns ParseModeRaw for no argument parsing.
func (c *PromptShowCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// HelpInfo returns structured help information for the shell-prompt-show command.
func (c *PromptShowCommand) HelpInfo() neurotypes.HelpInfo {
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
func (c *PromptShowCommand) Execute(_ map[string]string, _ string) error {
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
func (c *PromptShowCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GlobalRegistry.Register(&PromptShowCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-show command: %v", err))
	}
}
