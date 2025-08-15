package output

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// CodeRenderer provides professional code block rendering using charm.sh libraries.
type CodeRenderer struct {
	glamourRenderer *glamour.TermRenderer
	style           lipgloss.Style
	inlineStyle     lipgloss.Style
	styleProvider   StyleProvider
	available       bool
}

// NewCodeRenderer creates a new code renderer with professional styling.
func NewCodeRenderer(styleProvider StyleProvider) *CodeRenderer {
	// Determine theme from StyleProvider if available
	themeStyle := "auto" // default to auto
	if styleProvider != nil && styleProvider.IsAvailable() {
		themeStyle = styleProvider.GetThemeType()
	}

	// Create glamour renderer for markdown code blocks with theme-aware styling
	var glamourRenderer *glamour.TermRenderer
	var err error

	// Try with the theme from StyleProvider first
	if themeStyle != "" && themeStyle != "auto" {
		glamourRenderer, err = glamour.NewTermRenderer(
			glamour.WithStylePath(themeStyle),
			glamour.WithWordWrap(80),
		)
	}

	// Fallback to auto-detection if theme-specific rendering fails
	if glamourRenderer == nil || err != nil {
		glamourRenderer, err = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
			glamour.WithEnvironmentConfig(), // Use environment variables for terminal detection
		)
	}

	// Final fallback to dark theme if all else fails
	if err != nil {
		glamourRenderer, err = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"), // Use dark theme as final fallback
			glamour.WithWordWrap(80),
		)
		if err != nil {
			// Complete failure - disable glamour
			glamourRenderer = nil
		}
	}

	// Create lipgloss styles for code blocks with enhanced visibility
	codeBlockStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("blue")). // more visible blue border
		Background(lipgloss.Color("236")).        // slightly lighter background
		Foreground(lipgloss.Color("white"))       // bright text

	// Style for inline code with enhanced visibility
	inlineCodeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("cyan")).
		Padding(0, 1)

	return &CodeRenderer{
		glamourRenderer: glamourRenderer,
		style:           codeBlockStyle,
		inlineStyle:     inlineCodeStyle,
		styleProvider:   styleProvider,
		available:       true,
	}
}

// RenderCodeBlock renders a code block with professional styling and optional syntax highlighting.
func (c *CodeRenderer) RenderCodeBlock(code, language string) string {
	if !c.available {
		return c.renderPlainCodeBlock(code)
	}

	// Always try glamour first for ALL code blocks (including single-line)
	if c.glamourRenderer != nil {
		// Create markdown code block
		var markdown string
		if language != "" {
			markdown = "```" + language + "\n" + code + "\n```"
		} else {
			markdown = "```\n" + code + "\n```"
		}

		// Render with glamour
		rendered, err := c.glamourRenderer.Render(markdown)
		if err == nil && strings.TrimSpace(rendered) != "" {
			return strings.TrimSpace(rendered)
		}
	}

	// Only fallback to plain rendering if glamour completely fails
	return c.renderPlainCodeBlock(code)
}

// RenderInlineCode renders inline code with subtle styling.
func (c *CodeRenderer) RenderInlineCode(code string) string {
	if !c.available {
		return "`" + code + "`" // fallback to backticks
	}

	// Try lipgloss styling first
	styled := c.inlineStyle.Render(code)

	// If styling didn't work (same as input), add simple visual markers
	if styled == code {
		return "`" + code + "`"
	}

	return styled
}

// renderPlainCodeBlock provides fallback rendering for code blocks.
func (c *CodeRenderer) renderPlainCodeBlock(code string) string {
	lines := strings.Split(code, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			result[i] = "  " + line // Indent code lines
		} else {
			result[i] = line // Keep empty lines as-is
		}
	}

	return strings.Join(result, "\n")
}

// IsAvailable returns whether the code renderer is available.
func (c *CodeRenderer) IsAvailable() bool {
	return c.available
}
