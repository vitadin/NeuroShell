package output

import "fmt"

// PlainTextStyle implements TextStyle for plain text output without any styling.
// This is used as a fallback when no StyleProvider is available.
type PlainTextStyle struct {
	prefix string // Optional prefix for semantic meaning
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
	default:
		return NewPlainTextStyle("")
	}
}

// IsAvailable implements StyleProvider.IsAvailable.
func (p *PlainStyleProvider) IsAvailable() bool {
	return p.available
}

// String returns a string representation for debugging.
func (p *PlainStyleProvider) String() string {
	return fmt.Sprintf("PlainStyleProvider{available: %t}", p.available)
}
