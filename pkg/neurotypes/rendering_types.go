// Package neurotypes defines rendering-related interfaces and types for NeuroShell's thinking renderer system.
package neurotypes

import "github.com/charmbracelet/lipgloss"

// ThinkingRenderer defines the interface for rendering thinking blocks from LLM providers.
// This interface allows for pluggable rendering implementations that can adapt to different
// themes and styling requirements without direct service dependencies.
type ThinkingRenderer interface {
	// RenderThinkingBlocks renders a collection of thinking blocks using the provided configuration.
	// It returns a formatted string ready for display in the terminal.
	RenderThinkingBlocks(blocks []ThinkingBlock, config RenderConfig) string

	// RenderSingleBlock renders an individual thinking block using the provided configuration.
	// This method is useful for rendering blocks one at a time or for testing individual blocks.
	RenderSingleBlock(block ThinkingBlock, config RenderConfig) string

	// GetSupportedProviders returns a list of LLM providers that this renderer can handle.
	// This allows for provider-specific rendering optimizations.
	GetSupportedProviders() []string
}

// RenderConfig defines the configuration interface for thinking block rendering.
// This interface abstracts theme and styling dependencies, allowing commands to provide
// styling information without the renderer needing direct access to theme services.
type RenderConfig interface {
	// GetStyle returns the lipgloss style for a specific semantic element.
	// Common elements include: "info", "warning", "highlight", "background", "italic".
	GetStyle(element string) lipgloss.Style

	// GetTheme returns the name of the current theme being used.
	// This can be used for theme-specific rendering decisions.
	GetTheme() string

	// IsCompactMode returns true if rendering should use minimal vertical space.
	// In compact mode, thinking blocks should be rendered with reduced padding and spacing.
	IsCompactMode() bool

	// GetMaxWidth returns the maximum width for rendered content.
	// This allows for proper text wrapping and layout within terminal constraints.
	GetMaxWidth() int

	// ShowThinking returns true if thinking blocks should be rendered.
	// This allows for complete suppression of thinking content when desired.
	ShowThinking() bool

	// GetThinkingStyle returns the style preference for thinking blocks.
	// Valid values: "full", "summary", "hidden"
	GetThinkingStyle() string
}

// ThemeProvider defines the interface for providing theme-based styling.
// This interface allows the thinking renderer to access theme styles without
// direct dependency on the theme service implementation.
type ThemeProvider interface {
	// GetStyle returns the lipgloss style for a specific theme element.
	// This method maps semantic elements to actual styled components.
	GetStyle(element string) lipgloss.Style

	// GetTheme returns the name of the currently active theme.
	GetTheme() string

	// IsAvailable returns true if the theme provider is ready to provide styles.
	IsAvailable() bool
}

// RenderingOptions defines optional parameters for thinking block rendering.
// This struct allows for future extensibility without breaking existing interfaces.
type RenderingOptions struct {
	// MaxLines limits the number of lines displayed per thinking block.
	// Zero means no limit.
	MaxLines int

	// ShowProvider determines whether to display the provider name (e.g., "Claude", "Gemini").
	ShowProvider bool

	// IndentLevel specifies the indentation level for nested thinking blocks.
	IndentLevel int

	// PreviewOnly renders only a preview/summary of long thinking blocks.
	PreviewOnly bool

	// BorderStyle specifies the border style for thinking blocks.
	// Valid values: "none", "simple", "rounded", "thick"
	BorderStyle string
}
