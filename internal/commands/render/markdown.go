// Package render provides markdown and general rendering commands for NeuroShell.
// It includes commands for rendering markdown content to ANSI terminal output.
package render

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// MarkdownCommand implements the \render-markdown command for rendering markdown content.
// It uses the MarkdownService to render markdown to ANSI terminal output.
type MarkdownCommand struct{}

// Name returns the command name "render-markdown" for registration and lookup.
func (c *MarkdownCommand) Name() string {
	return "render-markdown"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *MarkdownCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the render-markdown command does.
func (c *MarkdownCommand) Description() string {
	return "Render markdown content to ANSI terminal output using Glamour"
}

// Usage returns the syntax and usage examples for the render-markdown command.
func (c *MarkdownCommand) Usage() string {
	return "\\render-markdown markdown content to render"
}

// HelpInfo returns structured help information for the render-markdown command.
func (c *MarkdownCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\render-markdown # Hello World\\nThis is **bold** text",
				Description: "Render a simple markdown heading and bold text",
			},
			{
				Command:     "\\render-markdown ## Features\\n- Item 1\\n- Item 2\\n- Item 3",
				Description: "Render a heading with a bulleted list",
			},
			{
				Command:     "\\render-markdown ```go\\nfunc main() {\\n    fmt.Println(\"Hello\")\\n}\\n```",
				Description: "Render a code block with syntax highlighting",
			},
			{
				Command:     "\\render-markdown [Link](https://example.com) and `inline code`",
				Description: "Render links and inline code",
			},
			{
				Command:     "\\render-markdown # Title\\n\\nThis is a paragraph\\nwith line breaks.",
				Description: "Use escape sequences for formatting (\\n creates newlines)",
			},
		},
		Notes: []string{
			"Uses Glamour library for high-quality markdown rendering",
			"Supports full markdown syntax including tables, code blocks, and links",
			"Automatically detects terminal theme for optimal styling",
			"Rendered output is stored in _output variable",
			"Integrates with NeuroShell's theme system",
			"Supports escape sequences: \\n (newline), \\t (tab), \\r (carriage return), \\\\ (backslash)",
		},
	}
}

// Execute renders markdown content using the MarkdownService.
// The rendered output is displayed to the console and stored in the _output variable.
func (c *MarkdownCommand) Execute(_ map[string]string, input string) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get markdown service
	markdownService, err := services.GetGlobalRegistry().GetService("markdown")
	if err != nil {
		return fmt.Errorf("markdown service not available: %w", err)
	}

	// Cast to MarkdownService
	mdService := markdownService.(*services.MarkdownService)

	// Render the markdown content using theme integration
	rendered, err := mdService.RenderWithTheme(input)
	if err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}

	// Get variable service to store the result
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Store the rendered output in the _output variable
	err = variableService.SetSystemVariable("_output", rendered)
	if err != nil {
		return fmt.Errorf("failed to store rendered output: %w", err)
	}

	// Display the rendered markdown to the console
	fmt.Print(rendered)

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&MarkdownCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register render-markdown command: %v", err))
	}
}
