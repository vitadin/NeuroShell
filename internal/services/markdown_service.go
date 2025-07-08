package services

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"neuroshell/internal/logger"
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

// Initialize sets up the MarkdownService with default configuration.
func (m *MarkdownService) Initialize() error {
	// Create a terminal renderer with auto-style detection
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80), // Default word wrap
	)
	if err != nil {
		return fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	m.renderer = renderer
	m.initialized = true

	logger.Debug("MarkdownService initialized successfully")
	return nil
}

// Render renders markdown content to ANSI terminal output.
// It returns the rendered content as a string with ANSI escape sequences.
func (m *MarkdownService) Render(markdown string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	if strings.TrimSpace(markdown) == "" {
		return "", fmt.Errorf("markdown content cannot be empty")
	}

	// Clean shell continuation markers and process escape sequences before rendering
	cleanedMarkdown := m.cleanShellMarkers(markdown)
	processedMarkdown := m.processEscapeSequences(cleanedMarkdown)

	// Render the markdown content
	rendered, err := m.renderer.Render(processedMarkdown)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// RenderWithStyle renders markdown content with a specific style.
// Supported styles include: "auto", "dark", "light", "notty", "ascii"
func (m *MarkdownService) RenderWithStyle(markdown string, style string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	if strings.TrimSpace(markdown) == "" {
		return "", fmt.Errorf("markdown content cannot be empty")
	}

	// Clean shell continuation markers and process escape sequences before rendering
	cleanedMarkdown := m.cleanShellMarkers(markdown)
	processedMarkdown := m.processEscapeSequences(cleanedMarkdown)

	// Create a new renderer with the specified style
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		// Fall back to default renderer if style is not available
		logger.Debug("Failed to create renderer with style, falling back to default", "style", style, "error", err)
		return m.Render(markdown)
	}

	// Render the markdown content
	rendered, err := renderer.Render(processedMarkdown)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown with style '%s': %w", style, err)
	}

	return rendered, nil
}

// RenderWithTheme renders markdown content using NeuroShell's theme system.
// It maps the current theme to an appropriate Glamour style.
func (m *MarkdownService) RenderWithTheme(markdown string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("markdown service not initialized")
	}

	// Get the current theme from the variable service
	themeName := m.getCurrentTheme()

	// Map NeuroShell theme to Glamour style
	glamourStyle := m.mapThemeToGlamourStyle(themeName)

	// Render with the mapped style
	return m.RenderWithStyle(markdown, glamourStyle)
}

// SetWordWrap sets the word wrap width for markdown rendering.
func (m *MarkdownService) SetWordWrap(width int) error {
	if !m.initialized {
		return fmt.Errorf("markdown service not initialized")
	}

	if width <= 0 {
		return fmt.Errorf("word wrap width must be positive, got %d", width)
	}

	// Create a new renderer with updated word wrap
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return fmt.Errorf("failed to create renderer with word wrap %d: %w", width, err)
	}

	m.renderer = renderer
	logger.Debug("MarkdownService word wrap updated", "width", width)
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

// cleanShellMarkers removes shell continuation markers that appear in multi-line input.
// This handles markers like "..." that shells use to indicate continuation lines.
func (m *MarkdownService) cleanShellMarkers(text string) string {
	// Handle both \n newlines and actual newlines
	normalizedText := strings.ReplaceAll(text, "\\n", "\n")

	// Split text into lines for processing
	lines := strings.Split(normalizedText, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Remove leading/trailing whitespace for pattern matching
		trimmed := strings.TrimSpace(line)

		// Skip lines that are just continuation markers
		if trimmed == "..." {
			continue
		}

		// Remove continuation markers at the beginning of lines
		if strings.HasPrefix(trimmed, "... ") {
			// Remove "... " prefix and preserve the rest
			cleaned := strings.TrimPrefix(trimmed, "... ")
			cleanedLines = append(cleanedLines, cleaned)
		} else {
			// Keep the line as-is
			cleanedLines = append(cleanedLines, line)
		}
	}

	result := strings.Join(cleanedLines, "\n")

	// Convert back to escape sequences if the original input used them
	if !strings.Contains(text, "\n") && strings.Contains(text, "\\n") {
		result = strings.ReplaceAll(result, "\n", "\\n")
	}

	return result
}

// processEscapeSequences converts common escape sequences to their actual characters.
// This allows users to input \n for newlines, \t for tabs, etc.
func (m *MarkdownService) processEscapeSequences(text string) string {
	// Replace common escape sequences
	result := text
	result = strings.ReplaceAll(result, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\r", "\r")
	result = strings.ReplaceAll(result, "\\\n", "\n") // Handle literal backslash-n
	result = strings.ReplaceAll(result, "\\\\", "\\") // Handle escaped backslashes
	return result
}

// GetServiceInfo returns information about the markdown service.
func (m *MarkdownService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        m.Name(),
		"initialized": m.initialized,
		"styles":      m.GetAvailableStyles(),
		"description": "Markdown rendering service using Glamour library",
	}
}

func init() {
	// Register the MarkdownService with the global registry
	if err := GlobalRegistry.RegisterService(NewMarkdownService()); err != nil {
		panic(fmt.Sprintf("failed to register markdown service: %v", err))
	}
}
