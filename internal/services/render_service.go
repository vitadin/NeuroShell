package services

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// RenderService provides theme management for NeuroShell styling.
// It maintains theme objects that commands can use for semantic styling.
type RenderService struct {
	initialized bool
	themes      map[string]*RenderTheme
}

// RenderTheme defines color schemes and styles for text rendering
type RenderTheme struct {
	Name       string
	Keyword    lipgloss.Style
	Variable   lipgloss.Style
	Command    lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
	Highlight  lipgloss.Style
	Bold       lipgloss.Style
	Italic     lipgloss.Style
	Underline  lipgloss.Style
	Background lipgloss.Style
}

// NewRenderService creates a new RenderService instance with default themes.
func NewRenderService() *RenderService {
	service := &RenderService{
		initialized: false,
		themes:      make(map[string]*RenderTheme),
	}
	// Initialize default themes
	service.initializeDefaultThemes()
	return service
}

// Name returns the service name "render" for registration.
func (r *RenderService) Name() string {
	return "render"
}

// Initialize sets up the RenderService for operation.
func (r *RenderService) Initialize(_ neurotypes.Context) error {
	r.initialized = true
	return nil
}

// initializeDefaultThemes sets up built-in color themes
func (r *RenderService) initializeDefaultThemes() {
	// Default theme - works well with both light and dark terminals
	r.themes["default"] = &RenderTheme{
		Name:       "default",
		Keyword:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}),
		Variable:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#6f42c1", Dark: "#a5a5ff"}),
		Command:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cf222e", Dark: "#f85149"}),
		Success:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1f883d", Dark: "#3fb950"}),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#d1242f", Dark: "#f85149"}),
		Warning:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#bf8700", Dark: "#d29922"}),
		Info:       lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}),
		Highlight:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"}).Background(lipgloss.AdaptiveColor{Light: "#0969da", Dark: "#58a6ff"}),
		Bold:       lipgloss.NewStyle().Bold(true),
		Italic:     lipgloss.NewStyle().Italic(true),
		Underline:  lipgloss.NewStyle().Underline(true),
		Background: lipgloss.NewStyle(),
	}

	// Dark theme - optimized for dark terminals
	r.themes["dark"] = &RenderTheme{
		Name:       "dark",
		Keyword:    lipgloss.NewStyle().Foreground(lipgloss.Color("#58a6ff")),
		Variable:   lipgloss.NewStyle().Foreground(lipgloss.Color("#a5a5ff")),
		Command:    lipgloss.NewStyle().Foreground(lipgloss.Color("#f85149")),
		Success:    lipgloss.NewStyle().Foreground(lipgloss.Color("#3fb950")),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.Color("#f85149")),
		Warning:    lipgloss.NewStyle().Foreground(lipgloss.Color("#d29922")),
		Info:       lipgloss.NewStyle().Foreground(lipgloss.Color("#58a6ff")),
		Highlight:  lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#58a6ff")),
		Bold:       lipgloss.NewStyle().Bold(true),
		Italic:     lipgloss.NewStyle().Italic(true),
		Underline:  lipgloss.NewStyle().Underline(true),
		Background: lipgloss.NewStyle(),
	}

	// Light theme - optimized for light terminals
	r.themes["light"] = &RenderTheme{
		Name:       "light",
		Keyword:    lipgloss.NewStyle().Foreground(lipgloss.Color("#0969da")),
		Variable:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6f42c1")),
		Command:    lipgloss.NewStyle().Foreground(lipgloss.Color("#cf222e")),
		Success:    lipgloss.NewStyle().Foreground(lipgloss.Color("#1f883d")),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.Color("#d1242f")),
		Warning:    lipgloss.NewStyle().Foreground(lipgloss.Color("#bf8700")),
		Info:       lipgloss.NewStyle().Foreground(lipgloss.Color("#0969da")),
		Highlight:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#0969da")),
		Bold:       lipgloss.NewStyle().Bold(true),
		Italic:     lipgloss.NewStyle().Italic(true),
		Underline:  lipgloss.NewStyle().Underline(true),
		Background: lipgloss.NewStyle(),
	}

	// Plain theme - no styling (returns text as-is)
	r.themes["plain"] = &RenderTheme{
		Name:       "plain",
		Keyword:    lipgloss.NewStyle(),
		Variable:   lipgloss.NewStyle(),
		Command:    lipgloss.NewStyle(),
		Success:    lipgloss.NewStyle(),
		Error:      lipgloss.NewStyle(),
		Warning:    lipgloss.NewStyle(),
		Info:       lipgloss.NewStyle(),
		Highlight:  lipgloss.NewStyle(),
		Bold:       lipgloss.NewStyle(),
		Italic:     lipgloss.NewStyle(),
		Underline:  lipgloss.NewStyle(),
		Background: lipgloss.NewStyle(),
	}
}

// GetAvailableThemes returns a list of available theme names
func (r *RenderService) GetAvailableThemes() []string {
	if !r.initialized {
		return []string{}
	}

	themes := make([]string, 0, len(r.themes))
	for name := range r.themes {
		themes = append(themes, name)
	}
	return themes
}

// GetTheme returns a specific theme by name
func (r *RenderService) GetTheme(name string) (*RenderTheme, bool) {
	if !r.initialized {
		return nil, false
	}

	theme, exists := r.themes[name]
	return theme, exists
}

// GetThemeByName retrieves a theme by name with support for aliases and case-insensitive matching.
// Supports aliases like "dark1" -> "dark". Always returns a valid theme object, never fails.
// For invalid themes, logs a warning and returns the plain theme.
func (r *RenderService) GetThemeByName(theme string) *RenderTheme {
	if !r.initialized {
		return r.GetDefaultTheme()
	}

	normalizedTheme := strings.ToLower(strings.TrimSpace(theme))

	switch normalizedTheme {
	case "", "plain":
		return r.themes["plain"]
	case "dark1", "dark":
		if themeObj, exists := r.themes["dark"]; exists {
			return themeObj
		}
		return r.themes["plain"]
	case "default":
		if themeObj, exists := r.themes["default"]; exists {
			return themeObj
		}
		return r.themes["plain"]
	case "light":
		if themeObj, exists := r.themes["light"]; exists {
			return themeObj
		}
		return r.themes["plain"]
	default:
		// Invalid theme - log warning and return plain theme
		logger.Debug("Invalid theme requested, using plain theme", "theme", theme, "available", r.GetAvailableThemes())
		return r.themes["plain"]
	}
}

// GetDefaultTheme returns the plain theme (no styling) for fallback scenarios.
func (r *RenderService) GetDefaultTheme() *RenderTheme {
	if !r.initialized {
		// Return a basic plain theme if service not initialized
		return &RenderTheme{
			Name:       "plain",
			Keyword:    lipgloss.NewStyle(),
			Variable:   lipgloss.NewStyle(),
			Command:    lipgloss.NewStyle(),
			Success:    lipgloss.NewStyle(),
			Error:      lipgloss.NewStyle(),
			Warning:    lipgloss.NewStyle(),
			Info:       lipgloss.NewStyle(),
			Highlight:  lipgloss.NewStyle(),
			Bold:       lipgloss.NewStyle(),
			Italic:     lipgloss.NewStyle(),
			Underline:  lipgloss.NewStyle(),
			Background: lipgloss.NewStyle(),
		}
	}
	return r.themes["plain"]
}

func init() {
	// Register the RenderService with the global registry
	if err := GlobalRegistry.RegisterService(NewRenderService()); err != nil {
		panic(fmt.Sprintf("failed to register render service: %v", err))
	}
}
