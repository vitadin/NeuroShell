// Package output provides a unified console output system for NeuroShell.
// It uses dependency injection to support optional styling while maintaining clean architecture.
package output

// StyleProvider is the interface that styling services (like ThemeService) implement
// to provide styled text rendering capabilities.
// The output package depends only on this interface, not on concrete services.
type StyleProvider interface {
	// GetStyle returns a TextStyle for the given semantic type.
	// Semantic types include: "info", "success", "warning", "error", "command", "variable", etc.
	GetStyle(semantic string) TextStyle

	// IsAvailable returns true if the style provider is ready to provide styles.
	// This allows the output system to gracefully fall back to plain text.
	IsAvailable() bool

	// GetThemeType returns the theme type for code rendering (e.g., "dark", "light", "auto").
	// This allows code renderers to select appropriate styling.
	GetThemeType() string
}

// TextStyle represents the capability to render text with styling.
// This interface is implemented by lipgloss.Style or other styling systems.
type TextStyle interface {
	// Render applies styling to the given text and returns the styled result.
	Render(text string) string
}

// Mode defines different output modes the printer can operate in.
type Mode int

const (
	// ModeAuto automatically detects the best output mode based on context
	ModeAuto Mode = iota

	// ModeStyled forces styled output (with colors, formatting)
	ModeStyled

	// ModePlain forces plain text output (no colors, minimal formatting)
	ModePlain

	// ModeJSON outputs structured JSON for machine consumption
	ModeJSON
)

// SemanticType defines the semantic meaning of output for consistent styling.
type SemanticType string

const (
	// SemanticPlain represents plain text without any semantic meaning.
	SemanticPlain SemanticType = "plain"
	// SemanticInfo represents informational text.
	SemanticInfo SemanticType = "info"
	// SemanticSuccess represents success or completion text.
	SemanticSuccess SemanticType = "success"
	// SemanticWarning represents warning text.
	SemanticWarning SemanticType = "warning"
	// SemanticError represents error text.
	SemanticError SemanticType = "error"

	// SemanticCommand represents command or executable text.
	SemanticCommand SemanticType = "command"
	// SemanticVariable represents variable or parameter text.
	SemanticVariable SemanticType = "variable"
	// SemanticKeyword represents keyword or reserved word text.
	SemanticKeyword SemanticType = "keyword"

	// SemanticHighlight represents highlighted or emphasized text.
	SemanticHighlight SemanticType = "highlight"
	// SemanticBold represents bold text styling.
	SemanticBold SemanticType = "bold"
	// SemanticItalic represents italic text styling.
	SemanticItalic SemanticType = "italic"
	// SemanticUnderline represents underlined text styling.
	SemanticUnderline SemanticType = "underline"

	// SemanticCode represents inline code text.
	SemanticCode SemanticType = "code"
	// SemanticCodeBlock represents multi-line code block text.
	SemanticCodeBlock SemanticType = "code_block"
	// SemanticComment represents comment text.
	SemanticComment SemanticType = "comment"
)
