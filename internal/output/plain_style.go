package output

import (
	"fmt"
	"strings"
)

// PlainTextStyle implements TextStyle for plain text output without any styling.
// This is used as a fallback when no StyleProvider is available.
type PlainTextStyle struct {
	prefix string // Optional prefix for semantic meaning
}

// InlineCodeTextStyle implements TextStyle for inline code with professional rendering.
type InlineCodeTextStyle struct {
	renderer *CodeRenderer
}

// NewInlineCodeTextStyle creates a new inline code text style with professional renderer.
func NewInlineCodeTextStyle(styleProvider StyleProvider) *InlineCodeTextStyle {
	return &InlineCodeTextStyle{
		renderer: NewCodeRenderer(styleProvider),
	}
}

// Render implements TextStyle.Render for inline code with charm.sh professional rendering.
func (i *InlineCodeTextStyle) Render(text string) string {
	if i.renderer != nil && i.renderer.IsAvailable() {
		return i.renderer.RenderInlineCode(text)
	}

	// Fallback to backticks
	return "`" + text + "`"
}

// CodeBlockTextStyle implements TextStyle for code block formatting with professional rendering.
type CodeBlockTextStyle struct {
	renderer *CodeRenderer
}

// NewCodeBlockTextStyle creates a new code block text style with professional renderer.
func NewCodeBlockTextStyle(styleProvider StyleProvider) *CodeBlockTextStyle {
	return &CodeBlockTextStyle{
		renderer: NewCodeRenderer(styleProvider),
	}
}

// Render implements TextStyle.Render for code blocks with charm.sh professional rendering.
func (c *CodeBlockTextStyle) Render(text string) string {
	if c.renderer != nil && c.renderer.IsAvailable() {
		// Try to detect language from content (basic heuristics)
		language := c.detectLanguage(text)
		return c.renderer.RenderCodeBlock(text, language)
	}

	// Fallback to plain rendering
	return c.renderPlain(text)
}

// renderPlain provides fallback plain text rendering.
func (c *CodeBlockTextStyle) renderPlain(text string) string {
	lines := strings.Split(text, "\n")
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

// detectLanguage provides basic language detection for syntax highlighting.
func (c *CodeBlockTextStyle) detectLanguage(text string) string {
	// Basic heuristics for NeuroShell commands and common languages
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "\\") {
		return "bash" // NeuroShell commands look similar to bash
	}
	if strings.Contains(text, "func ") && strings.Contains(text, "{") {
		return "go"
	}
	if strings.Contains(text, "def ") || strings.Contains(text, "import ") {
		return "python"
	}
	if strings.Contains(text, "const ") || strings.Contains(text, "function ") {
		return "javascript"
	}

	return "" // No language detected, use plain rendering
}

// NewPlainTextStyle creates a new plain text style with an optional prefix.
func NewPlainTextStyle(prefix string) *PlainTextStyle {
	return &PlainTextStyle{prefix: prefix}
}

// Render implements TextStyle.Render for plain text output.
// It simply returns the text with an optional prefix.
func (p *PlainTextStyle) Render(text string) string {
	if p.prefix != "" {
		return p.prefix + text
	}
	return text
}

// PlainStyleProvider implements StyleProvider for plain text output.
// This is used when no theme service is available or when plain mode is forced.
type PlainStyleProvider struct {
	available bool
}

// NewPlainStyleProvider creates a new plain style provider.
func NewPlainStyleProvider() *PlainStyleProvider {
	return &PlainStyleProvider{available: true}
}

// GetStyle implements StyleProvider.GetStyle for plain text styles with semantic prefixes.
func (p *PlainStyleProvider) GetStyle(semantic string) TextStyle {
	switch semantic {
	case "success":
		return NewPlainTextStyle("✓ ")
	case "warning":
		return NewPlainTextStyle("⚠ ")
	case "error":
		return NewPlainTextStyle("✗ ")
	case "info":
		return NewPlainTextStyle("ℹ ")
	case "command":
		return NewPlainTextStyle("\\")
	case "variable":
		return NewPlainTextStyle("$")
	case "code":
		return NewInlineCodeTextStyle(p)
	case "code_block":
		return NewCodeBlockTextStyle(p)
	case "comment":
		return NewPlainTextStyle("")
	default:
		return NewPlainTextStyle("")
	}
}

// IsAvailable implements StyleProvider.IsAvailable.
func (p *PlainStyleProvider) IsAvailable() bool {
	return p.available
}

// GetThemeType implements StyleProvider.GetThemeType.
func (p *PlainStyleProvider) GetThemeType() string {
	return "auto" // Plain style provider uses auto-detection
}

// String returns a string representation for debugging.
func (p *PlainStyleProvider) String() string {
	return fmt.Sprintf("PlainStyleProvider{available: %t}", p.available)
}
