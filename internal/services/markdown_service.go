package services

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"neuroshell/internal/logger"
	"neuroshell/internal/stringprocessing"
)

// MarkdownService provides markdown rendering capabilities for NeuroShell using Glamour.
// It supports multiple rendering styles and integrates with the theme system for consistent styling.
type MarkdownService struct {
	initialized bool
	renderer    *glamour.TermRenderer
}

// NewMarkdownService creates a new MarkdownService instance.
func NewMarkdownService() *MarkdownService {
	return &MarkdownService{
		initialized: false,
		renderer:    nil,
	}
}

// Name returns the service name "markdown" for registration.
func (m *MarkdownService) Name() string {
	return "markdown"
}

// Initialize sets up the MarkdownService with enhanced UTF-8 and terminal support.
func (m *MarkdownService) Initialize() error {
	// Ensure UTF-8 environment variables are set for proper terminal detection
	m.ensureUTF8Environment()

	// Create a terminal renderer with UTF-8 support
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),  // Auto-detect terminal style
		glamour.WithWordWrap(80), // Default word wrap with UTF-8 awareness
	)
	if err != nil {
		return fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	m.renderer = renderer
	m.initialized = true

	logger.Debug("MarkdownService initialized successfully with UTF-8 support")
	return nil
}

// Render renders markdown content to ANSI terminal output.
// It returns the rendered content as a string with ANSI escape sequences.
// By default, it interprets escape sequences (legacy behavior).
func (m *MarkdownService) Render(markdown string) (string, error) {
	return m.RenderWithOptions(markdown, true)
}

// RenderWithOptions renders markdown content with configurable string processing and UTF-8 safety.
// interpretEscapes controls whether escape sequences like \n are converted to actual characters.
func (m *MarkdownService) RenderWithOptions(markdown string, interpretEscapes bool) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	if strings.TrimSpace(markdown) == "" {
		return "", fmt.Errorf("markdown content cannot be empty")
	}

	// Validate and sanitize UTF-8 input to prevent encoding issues
	sanitizedMarkdown := m.sanitizeUTF8(markdown)

	// Process text using the shared utility functions
	processedMarkdown := stringprocessing.ProcessTextForMarkdown(sanitizedMarkdown, interpretEscapes)

	// Render the markdown content
	rendered, err := m.renderer.Render(processedMarkdown)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	// Validate output is valid UTF-8
	if !utf8.ValidString(rendered) {
		logger.Debug("Rendered markdown contains invalid UTF-8, attempting to fix")
		rendered = m.fixInvalidUTF8(rendered)
	}

	return rendered, nil
}

// RenderWithStyle renders markdown content with a specific style.
// Supported styles include: "auto", "dark", "light", "notty", "ascii"
// By default, it interprets escape sequences (legacy behavior).
func (m *MarkdownService) RenderWithStyle(markdown string, style string) (string, error) {
	return m.RenderWithStyleAndOptions(markdown, style, true)
}

// RenderWithStyleAndOptions renders markdown content with a specific style and configurable string processing with UTF-8 safety.
func (m *MarkdownService) RenderWithStyleAndOptions(markdown string, style string, interpretEscapes bool) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	if strings.TrimSpace(markdown) == "" {
		return "", fmt.Errorf("markdown content cannot be empty")
	}

	// Validate and sanitize UTF-8 input to prevent encoding issues
	sanitizedMarkdown := m.sanitizeUTF8(markdown)

	// Process text using the shared utility functions
	processedMarkdown := stringprocessing.ProcessTextForMarkdown(sanitizedMarkdown, interpretEscapes)

	// Create a new renderer with the specified style and UTF-8 support
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		// Fall back to default renderer if style is not available
		logger.Debug("Failed to create renderer with style, falling back to default", "style", style, "error", err)
		return m.RenderWithOptions(markdown, interpretEscapes)
	}

	// Render the markdown content
	rendered, err := renderer.Render(processedMarkdown)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown with style '%s': %w", style, err)
	}

	// Validate output is valid UTF-8
	if !utf8.ValidString(rendered) {
		logger.Debug("Rendered markdown contains invalid UTF-8, attempting to fix")
		rendered = m.fixInvalidUTF8(rendered)
	}

	return rendered, nil
}

// RenderWithTheme renders markdown content using NeuroShell's theme system.
// It maps the current theme to an appropriate Glamour style.
// By default, it interprets escape sequences (legacy behavior).
func (m *MarkdownService) RenderWithTheme(markdown string) (string, error) {
	return m.RenderWithThemeAndOptions(markdown, true)
}

// RenderWithThemeAndOptions renders markdown content using NeuroShell's theme system with configurable string processing.
func (m *MarkdownService) RenderWithThemeAndOptions(markdown string, interpretEscapes bool) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	// Get the current theme from the variable service
	themeName := m.getCurrentTheme()

	// Map NeuroShell theme to Glamour style
	glamourStyle := m.mapThemeToGlamourStyle(themeName)

	// Render with the mapped style and options
	return m.RenderWithStyleAndOptions(markdown, glamourStyle, interpretEscapes)
}

// SetWordWrap sets the word wrap width for markdown rendering with UTF-8 character awareness.
func (m *MarkdownService) SetWordWrap(width int) error {
	if !m.initialized {
		return fmt.Errorf("markdown service not initialized")
	}

	if width <= 0 {
		return fmt.Errorf("word wrap width must be positive, got %d", width)
	}

	// Create a new renderer with updated word wrap and UTF-8 support
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return fmt.Errorf("failed to create renderer with word wrap %d: %w", width, err)
	}

	m.renderer = renderer
	logger.Debug("MarkdownService word wrap updated with UTF-8 support", "width", width)
	return nil
}

// getCurrentTheme gets the current theme from the variable service.
func (m *MarkdownService) getCurrentTheme() string {
	// Try to get the _style variable
	variableService, err := GetGlobalVariableService()
	if err != nil {
		logger.Debug("Failed to get variable service, using default theme")
		return "default"
	}

	themeValue, err := variableService.Get("_style")
	if err != nil {
		logger.Debug("Failed to get _style variable, using default theme")
		return "default"
	}

	if themeValue == "" {
		return "default"
	}

	return themeValue
}

// mapThemeToGlamourStyle maps NeuroShell theme names to Glamour styles.
func (m *MarkdownService) mapThemeToGlamourStyle(themeName string) string {
	// Check if colors should be disabled (--no-color flag or NO_COLOR environment variable)
	if lipgloss.ColorProfile() == termenv.Ascii {
		return "notty" // Plain text, no colors
	}

	switch strings.ToLower(themeName) {
	case "dark", "dark1":
		return "dark"
	case "light":
		return "light"
	case "plain":
		return "notty"
	case "default":
		return "auto"
	default:
		// For unknown themes, use auto-detection
		return "auto"
	}
}

// GetAvailableStyles returns a list of available Glamour styles.
func (m *MarkdownService) GetAvailableStyles() []string {
	return []string{
		"auto",  // Auto-detect based on terminal
		"dark",  // Dark theme
		"light", // Light theme
		"notty", // Plain text (no colors)
		"ascii", // ASCII-only styling
	}
}

// GetServiceInfo returns information about the markdown service.
func (m *MarkdownService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        m.Name(),
		"initialized": m.initialized,
		"styles":      m.GetAvailableStyles(),
		"description": "Markdown rendering service using Glamour library with UTF-8 support",
	}
}

// ensureUTF8Environment sets up environment variables for proper UTF-8 terminal handling.
func (m *MarkdownService) ensureUTF8Environment() {
	// Set LANG if not already set to ensure UTF-8 locale
	if os.Getenv("LANG") == "" {
		if err := os.Setenv("LANG", "en_US.UTF-8"); err != nil {
			logger.Debug("Failed to set LANG environment variable", "error", err)
		} else {
			logger.Debug("Set LANG environment variable for UTF-8 support")
		}
	}

	// Set LC_ALL if not set to ensure consistent UTF-8 handling
	if os.Getenv("LC_ALL") == "" {
		if err := os.Setenv("LC_ALL", "en_US.UTF-8"); err != nil {
			logger.Debug("Failed to set LC_ALL environment variable", "error", err)
		} else {
			logger.Debug("Set LC_ALL environment variable for UTF-8 support")
		}
	}

	// Ensure TERM supports UTF-8 if not already set
	if os.Getenv("TERM") == "" {
		if err := os.Setenv("TERM", "xterm-256color"); err != nil {
			logger.Debug("Failed to set TERM environment variable", "error", err)
		} else {
			logger.Debug("Set TERM environment variable for enhanced terminal support")
		}
	}
}

// sanitizeUTF8 cleans and validates UTF-8 input to prevent encoding issues.
func (m *MarkdownService) sanitizeUTF8(input string) string {
	// If input is already valid UTF-8, return as-is
	if utf8.ValidString(input) {
		return input
	}

	logger.Debug("Input contains invalid UTF-8, sanitizing")

	// Convert invalid UTF-8 sequences to replacement characters
	sanitized := strings.ToValidUTF8(input, "�")

	// Log if we had to make replacements
	if sanitized != input {
		logger.Debug("Replaced invalid UTF-8 sequences with replacement characters")
	}

	return sanitized
}

// fixInvalidUTF8 attempts to fix invalid UTF-8 in rendered output.
func (m *MarkdownService) fixInvalidUTF8(output string) string {
	// Convert any remaining invalid UTF-8 to replacement characters
	fixed := strings.ToValidUTF8(output, "�")

	// If we made changes, log it
	if fixed != output {
		logger.Debug("Fixed invalid UTF-8 in rendered output")
	}

	return fixed
}

func init() {
	// Register the MarkdownService with the global registry
	if err := GlobalRegistry.RegisterService(NewMarkdownService()); err != nil {
		panic(fmt.Sprintf("failed to register markdown service: %v", err))
	}
}
