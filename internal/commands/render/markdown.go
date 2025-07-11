// Package render provides markdown and general rendering commands for NeuroShell.
// It includes commands for rendering markdown content to ANSI terminal output.
// TODO: Integrate into state machine - temporarily commented out for build compatibility
package render

/*

import (
	"fmt"
	"strconv"

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
	return "\\render-markdown[raw=true] markdown content to render"
}

// HelpInfo returns structured help information for the render-markdown command.
func (c *MarkdownCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "raw",
				Description: "Treat escape sequences as literal characters without interpreting them",
				Required:    false,
				Type:        "bool",
				Default:     "true",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\render-markdown # Hello World\\nThis is **bold** text",
				Description: "Render markdown with literal \\n (default raw=true)",
			},
			{
				Command:     "\\render-markdown[raw=false] # Hello World\\nThis is **bold** text",
				Description: "Render markdown with \\n converted to actual newlines",
			},
			{
				Command:     "\\render-markdown ## Features\\n- Item 1\\n- Item 2\\n- Item 3",
				Description: "Render a heading with a bulleted list (raw mode)",
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
				Command:     "\\render-markdown # Multiline input ...\\n... with continuation markers",
				Description: "Handle multiline input with continuation markers",
			},
		},
		Notes: []string{
			"Uses Glamour library for high-quality markdown rendering",
			"Supports full markdown syntax including tables, code blocks, and links",
			"Automatically detects terminal theme for optimal styling",
			"Rendered output is stored in _output variable",
			"Integrates with NeuroShell's theme system",
			"Default raw=true treats \\n as literal characters",
			"Set raw=false to interpret escape sequences: \\n (newline), \\t (tab), \\r (carriage return), \\\\ (backslash)",
			"Continuation markers (...) are always processed regardless of raw setting",
		},
	}
}

// Execute renders markdown content using the MarkdownService.
// The rendered output is displayed to the console and stored in the _output variable.
// Options:
//   - raw: treat escape sequences as literal characters (default: true)
func (c *MarkdownCommand) Execute(args map[string]string, input string) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Parse raw option (default to true for backward compatibility with the user's requirement)
	rawStr := args["raw"]
	raw := true // Default to raw=true (treat \n as literal)
	if rawStr != "" {
		var err error
		raw, err = strconv.ParseBool(rawStr)
		if err != nil {
			return fmt.Errorf("invalid value for raw option: %s (must be true or false)", rawStr)
		}
	}

	// Get markdown service
	markdownService, err := services.GetGlobalRegistry().GetService("markdown")
	if err != nil {
		return fmt.Errorf("markdown service not available: %w", err)
	}

	// Cast to MarkdownService
	mdService := markdownService.(*services.MarkdownService)

	// Render the markdown content using theme integration with configurable escape processing
	// Note: raw=true means DON'T interpret escapes, so we pass !raw to interpretEscapes
	rendered, err := mdService.RenderWithThemeAndOptions(input, !raw)
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
*/
