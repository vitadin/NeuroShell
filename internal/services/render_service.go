package services

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
	"neuroshell/internal/data/embedded"
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

// NewRenderService creates a new RenderService instance with themes loaded from YAML.
func NewRenderService() *RenderService {
	service := &RenderService{
		initialized: false,
		themes:      make(map[string]*RenderTheme),
	}
	// Load themes from embedded YAML files
	service.loadThemesFromYAML()
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

// loadThemesFromYAML loads themes from embedded YAML files
func (r *RenderService) loadThemesFromYAML() {
	// Load individual theme files
	themeFiles := map[string][]byte{
		"default": embedded.DefaultThemeData,
		"dark":    embedded.DarkThemeData,
		"light":   embedded.LightThemeData,
		"plain":   embedded.PlainThemeData,
	}

	for themeName, themeData := range themeFiles {
		theme, err := r.loadThemeFile(themeData)
		if err != nil {
			logger.Error("Failed to load theme", "theme", themeName, "error", err)
			// Fall back to creating a basic plain theme
			r.themes[themeName] = r.createFallbackTheme(themeName)
			continue
		}
		r.themes[themeName] = theme
	}

	// Ensure we always have a plain theme as fallback
	if _, exists := r.themes["plain"]; !exists {
		r.themes["plain"] = r.createFallbackTheme("plain")
	}
}

// loadThemeFile loads and parses an individual theme file from embedded YAML data.
func (r *RenderService) loadThemeFile(data []byte) (*RenderTheme, error) {
	var themeFile neurotypes.ThemeFile

	if err := yaml.Unmarshal(data, &themeFile); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	// Convert ThemeConfig to RenderTheme
	return r.convertThemeConfig(&themeFile.ThemeConfig), nil
}

// convertThemeConfig converts a ThemeConfig from YAML to a RenderTheme with lipgloss styles.
func (r *RenderService) convertThemeConfig(config *neurotypes.ThemeConfig) *RenderTheme {
	return &RenderTheme{
		Name:       config.Name,
		Keyword:    r.createStyle(config.Styles.Keyword),
		Variable:   r.createStyle(config.Styles.Variable),
		Command:    r.createStyle(config.Styles.Command),
		Success:    r.createStyle(config.Styles.Success),
		Error:      r.createStyle(config.Styles.Error),
		Warning:    r.createStyle(config.Styles.Warning),
		Info:       r.createStyle(config.Styles.Info),
		Highlight:  r.createStyle(config.Styles.Highlight),
		Bold:       r.createStyle(config.Styles.Bold),
		Italic:     r.createStyle(config.Styles.Italic),
		Underline:  r.createStyle(config.Styles.Underline),
		Background: r.createStyle(config.Styles.Background),
	}
}

// createStyle converts a StyleConfig to a lipgloss.Style.
func (r *RenderService) createStyle(config neurotypes.StyleConfig) lipgloss.Style {
	style := lipgloss.NewStyle()

	// Handle foreground color
	if config.Foreground != nil {
		if color := r.parseColor(config.Foreground); color != nil {
			style = style.Foreground(color)
		}
	}

	// Handle background color
	if config.Background != nil {
		if color := r.parseColor(config.Background); color != nil {
			style = style.Background(color)
		}
	}

	// Handle text decorations
	if config.Bold != nil && *config.Bold {
		style = style.Bold(true)
	}
	if config.Italic != nil && *config.Italic {
		style = style.Italic(true)
	}
	if config.Underline != nil && *config.Underline {
		style = style.Underline(true)
	}
	if config.Strikethrough != nil && *config.Strikethrough {
		style = style.Strikethrough(true)
	}

	return style
}

// parseColor parses a color value that can be a string, AdaptiveColor, or map.
func (r *RenderService) parseColor(colorValue interface{}) lipgloss.TerminalColor {
	switch v := colorValue.(type) {
	case string:
		// Simple color string
		return lipgloss.Color(v)
	case map[string]interface{}:
		// Check if it's an adaptive color with light/dark keys
		if light, hasLight := v["light"].(string); hasLight {
			if dark, hasDark := v["dark"].(string); hasDark {
				return lipgloss.AdaptiveColor{Light: light, Dark: dark}
			}
		}
		return nil
	default:
		return nil
	}
}

// createFallbackTheme creates a basic plain theme for fallback scenarios.
func (r *RenderService) createFallbackTheme(name string) *RenderTheme {
	return &RenderTheme{
		Name:       name,
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
