package shell

import (
	"fmt"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// PromptPreviewCommand previews the current prompt with current context
type PromptPreviewCommand struct{}

// Name returns the command name "shell-prompt-preview" for registration and lookup.
func (c *PromptPreviewCommand) Name() string {
	return "shell-prompt-preview"
}

// Description returns a brief description of what the shell-prompt-preview command does.
func (c *PromptPreviewCommand) Description() string {
	return "Preview the current shell prompt with current context"
}

// Usage returns the syntax and usage examples for the shell-prompt-preview command.
func (c *PromptPreviewCommand) Usage() string {
	return `\shell-prompt-preview

Show how the current shell prompt would appear with interpolated variables.

Example:
  \shell-prompt-preview`
}

// ParseMode returns ParseModeRaw for no argument parsing.
func (c *PromptPreviewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// HelpInfo returns structured help information for the shell-prompt-preview command.
func (c *PromptPreviewCommand) HelpInfo() neurotypes.HelpInfo {
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
func (c *PromptPreviewCommand) Execute(_ map[string]string, _ string) error {
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

	// Show each line with interpolation and color processing
	for i, template := range lines {
		// Interpolate variables using context
		interpolated := ctx.InterpolateVariables(template)

		// Process color markup
		colored := c.processPromptLine(interpolated)
		fmt.Printf("%s\n", colored)

		// Don't add extra newline after the last line
		if i == len(lines)-1 {
			// Add cursor to show where input would go
			fmt.Print("█")
		}
	}
	fmt.Println() // Final newline

	return nil
}

// processPromptLine processes a prompt line with color markup, similar to main.go
func (c *PromptPreviewCommand) processPromptLine(template string) string {
	colorService, err := services.GetGlobalRegistry().GetService("prompt_color")
	if err != nil {
		// If color service is not available, return as-is
		return template
	}

	promptColor, ok := colorService.(*services.PromptColorService)
	if !ok {
		return template
	}

	return promptColor.ProcessColorMarkup(template)
}

// IsReadOnly returns true as this command only displays a preview.
func (c *PromptPreviewCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GlobalRegistry.Register(&PromptPreviewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register shell-prompt-preview command: %v", err))
	}
}
