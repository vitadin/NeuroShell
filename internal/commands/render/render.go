// Package render provides rendering commands for NeuroShell.
package render

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"neuroshell/internal/commands"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// Command implements the \render command for styling and highlighting text.
// It provides lipgloss-based text rendering with keyword highlighting and theme support.
type Command struct{}

// Name returns the command name "render" for registration and lookup.
func (c *Command) Name() string {
	return "render"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *Command) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the render command does.
func (c *Command) Description() string {
	return "Style and highlight text using lipgloss with keyword support"
}

// Usage returns the syntax and usage examples for the render command.
func (c *Command) Usage() string {
	return "\\render[keywords=[\\get,\\set], style=bold, theme=dark, to=var] text to render"
}

// HelpInfo returns structured help information for the render command.
func (c *Command) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "keywords",
				Description: "Array of keywords to highlight (e.g., [\\get,\\set])",
				Required:    false,
				Type:        "array",
			},
			{
				Name:        "style",
				Description: "Named style: bold, italic, underline, success, error, warning, info, highlight",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "theme",
				Description: "Color theme: default, dark, light",
				Required:    false,
				Type:        "string",
				Default:     "default",
			},
			{
				Name:        "to",
				Description: "Variable to store result",
				Required:    false,
				Type:        "string",
				Default:     "_output",
			},
			{
				Name:        "silent",
				Description: "Suppress console output",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\render[style=bold] Hello World",
				Description: "Render text with bold styling",
			},
			{
				Command:     "\\render[keywords=[\\get,\\set]] Use \\get and \\set commands",
				Description: "Highlight specific NeuroShell commands in text",
			},
			{
				Command:     "\\render[style=success, theme=dark] Operation completed!",
				Description: "Success message with dark theme styling",
			},
			{
				Command:     "\\render[to=styled_output, silent=true] Formatted text",
				Description: "Store styled text in variable without console output",
			},
			{
				Command:     "\\render[color=#FF5733, background=#000000] Colorful text",
				Description: "Custom foreground and background colors",
			},
		},
		Notes: []string{
			"Uses lipgloss library for professional terminal styling",
			"Keywords are highlighted with special colors when specified",
			"Supports both hex colors (#FF5733) and named colors (red, blue, etc.)",
			"Style options can be combined (e.g., bold + italic + color)",
			"Rendered output is stored in specified variable for reuse",
			"Themes provide coordinated color schemes for consistency",
		},
	}
}

// Execute applies styling and highlighting to text based on the provided options.
// Options:
//   - keywords: array of keywords to highlight (e.g., [\\get,\\set])
//   - style: named style (bold, italic, underline, success, error, warning, info, highlight)
//   - theme: color theme (default, dark, light)
//   - color: foreground color (hex code or color name)
//   - background: background color (hex code or color name)
//   - bold: make text bold (true/false)
//   - italic: make text italic (true/false)
//   - underline: make text underlined (true/false)
//   - to: variable to store result (default: ${_output})
//   - silent: suppress console output (true/false, default: false)
func (c *Command) Execute(args map[string]string, input string) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get theme service
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		return fmt.Errorf("theme service not available: %w", err)
	}

	// Get variable service for storing result
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get theme object
	themeName := args["theme"]
	if themeName == "" {
		themeName = "default"
	}
	theme := themeService.GetThemeByName(themeName)

	// Apply styling to the input text
	styledText := c.renderText(input, args, theme)

	// Store result in target variable
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
		// Store in system variable
		err = variableService.SetSystemVariable(targetVar, styledText)
	} else {
		// Store in user variable
		err = variableService.Set(targetVar, styledText)
	}
	if err != nil {
		return fmt.Errorf("failed to store result in variable '%s': %w", targetVar, err)
	}

	// Parse silent option
	silentStr := args["silent"]
	silent := false
	if silentStr != "" {
		silent, err = strconv.ParseBool(silentStr)
		if err != nil {
			return fmt.Errorf("invalid value for silent option: %s (must be true or false)", silentStr)
		}
	}

	// Output to console unless silent mode is enabled
	if !silent {
		fmt.Print(styledText)
		// Only add newline if the styled text doesn't already end with one
		if len(styledText) > 0 && styledText[len(styledText)-1] != '\n' {
			fmt.Println()
		}
	}

	return nil
}

// renderText applies styling to text using theme objects and command arguments
func (c *Command) renderText(text string, args map[string]string, theme *services.Theme) string {
	result := text

	// Apply keyword highlighting first if specified
	if keywordsStr, exists := args["keywords"]; exists {
		keywords := parser.ParseArrayValue(keywordsStr)
		// Note: Variable interpolation is now handled by the state machine before commands execute
		result = c.highlightKeywords(result, keywords, theme)
	}

	// Apply NeuroShell-specific highlighting
	result = c.highlightNeuroShellSyntax(result, theme)

	// Apply global styling if specified
	if c.hasGlobalStyling(args) {
		result = c.applyGlobalStyling(result, args, theme)
	}

	return result
}

// hasGlobalStyling checks if any global styling options are specified
func (c *Command) hasGlobalStyling(args map[string]string) bool {
	_, hasStyle := args["style"]
	_, hasColor := args["color"]
	_, hasBackground := args["background"]
	_, hasBold := args["bold"]
	_, hasItalic := args["italic"]
	_, hasUnderline := args["underline"]

	return hasStyle || hasColor || hasBackground || hasBold || hasItalic || hasUnderline
}

// highlightKeywords highlights specific keywords in the text using theme styles
func (c *Command) highlightKeywords(text string, keywords []string, theme *services.Theme) string {
	result := text

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		// Escape special regex characters in the keyword
		escaped := regexp.QuoteMeta(keyword)

		// Create regex pattern for whole words (but allow backslash prefix for commands)
		pattern := fmt.Sprintf(`(\b%s\b|\\%s\b)`, escaped, escaped)
		re := regexp.MustCompile(pattern)

		// Replace matches with styled versions using theme's keyword style
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return theme.Keyword.Render(match)
		})
	}

	return result
}

// highlightNeuroShellSyntax applies syntax highlighting for NeuroShell-specific patterns
func (c *Command) highlightNeuroShellSyntax(text string, theme *services.Theme) string {
	result := text

	// Highlight variables: ${variable_name}
	variablePattern := regexp.MustCompile(`\$\{[^}]+\}`)
	result = variablePattern.ReplaceAllStringFunc(result, func(match string) string {
		return theme.Variable.Render(match)
	})

	// Highlight commands: \command (but only if not already styled)
	commandPattern := regexp.MustCompile(`\\[a-zA-Z_][a-zA-Z0-9_-]*`)
	result = commandPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Check if this text is already styled (contains ANSI sequences)
		if strings.Contains(match, "\x1b[") {
			return match // Already styled, don't re-style
		}
		return theme.Command.Render(match)
	})

	return result
}

// applyGlobalStyling applies global style options to the entire text using theme and custom options
func (c *Command) applyGlobalStyling(text string, args map[string]string, theme *services.Theme) string {
	style := lipgloss.NewStyle()

	// Apply boolean styling options
	if boldStr, exists := args["bold"]; exists {
		if bold, err := strconv.ParseBool(boldStr); err == nil && bold {
			style = style.Bold(true)
		}
	}
	if italicStr, exists := args["italic"]; exists {
		if italic, err := strconv.ParseBool(italicStr); err == nil && italic {
			style = style.Italic(true)
		}
	}
	if underlineStr, exists := args["underline"]; exists {
		if underline, err := strconv.ParseBool(underlineStr); err == nil && underline {
			style = style.Underline(true)
		}
	}

	// Apply custom colors
	if color, exists := args["color"]; exists {
		style = style.Foreground(lipgloss.Color(color))
	}
	if background, exists := args["background"]; exists {
		style = style.Background(lipgloss.Color(background))
	}

	// Apply named semantic styles from theme
	if styleOption, exists := args["style"]; exists {
		switch styleOption {
		case "bold":
			style = theme.Bold
		case "italic":
			style = theme.Italic
		case "underline":
			style = theme.Underline
		case "success":
			style = theme.Success
		case "error":
			style = theme.Error
		case "warning":
			style = theme.Warning
		case "info":
			style = theme.Info
		case "highlight":
			style = theme.Highlight
		}
	}

	return style.Render(text)
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&Command{}); err != nil {
		panic(fmt.Sprintf("failed to register render command: %v", err))
	}
}
