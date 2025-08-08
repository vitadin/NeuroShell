package services

import (
	"fmt"
	"strings"

	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"gopkg.in/yaml.v3"
)

// ThemeService provides theme management for NeuroShell styling.
// It maintains theme objects that commands can use for semantic styling.
type ThemeService struct {
	initialized bool
	themes      map[string]*Theme
}

// Theme defines color schemes and styles for text rendering
type Theme struct {
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
	List       lipgloss.Style
}

// NewThemeService creates a new ThemeService instance with themes loaded from YAML.
func NewThemeService() *ThemeService {
	service := &ThemeService{
		initialized: false,
		themes:      make(map[string]*Theme),
	}
	// Load themes from embedded YAML files
	service.loadThemesFromYAML()
	return service
}

// Name returns the service name "theme" for registration.
func (t *ThemeService) Name() string {
	return "theme"
}

// Initialize sets up the ThemeService for operation.
func (t *ThemeService) Initialize() error {
	t.initialized = true
	return nil
}

// loadThemesFromYAML loads themes from embedded YAML files
func (t *ThemeService) loadThemesFromYAML() {
	// Load individual theme files
	themeFiles := map[string][]byte{
		"default": embedded.DefaultThemeData,
		"dark":    embedded.DarkThemeData,
		"light":   embedded.LightThemeData,
		"plain":   embedded.PlainThemeData,
	}

	for themeName, themeData := range themeFiles {
		theme, err := t.loadThemeFile(themeData)
		if err != nil {
			logger.Error("Failed to load theme", "theme", themeName, "error", err)
			// Fall back to creating a basic plain theme
			t.themes[themeName] = t.createFallbackTheme(themeName)
			continue
		}
		t.themes[themeName] = theme
	}

	// Ensure we always have a plain theme as fallback
	if _, exists := t.themes["plain"]; !exists {
		t.themes["plain"] = t.createFallbackTheme("plain")
	}
}

// loadThemeFile loads and parses an individual theme file from embedded YAML data.
func (t *ThemeService) loadThemeFile(data []byte) (*Theme, error) {
	var themeFile neurotypes.ThemeFile

	if err := yaml.Unmarshal(data, &themeFile); err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	// Convert ThemeConfig to Theme
	return t.convertThemeConfig(&themeFile.ThemeConfig), nil
}

// convertThemeConfig converts a ThemeConfig from YAML to a Theme with lipgloss styles.
func (t *ThemeService) convertThemeConfig(config *neurotypes.ThemeConfig) *Theme {
	return &Theme{
		Name:       config.Name,
		Keyword:    t.createStyle(config.Styles.Keyword),
		Variable:   t.createStyle(config.Styles.Variable),
		Command:    t.createStyle(config.Styles.Command),
		Success:    t.createStyle(config.Styles.Success),
		Error:      t.createStyle(config.Styles.Error),
		Warning:    t.createStyle(config.Styles.Warning),
		Info:       t.createStyle(config.Styles.Info),
		Highlight:  t.createStyle(config.Styles.Highlight),
		Bold:       t.createStyle(config.Styles.Bold),
		Italic:     t.createStyle(config.Styles.Italic),
		Underline:  t.createStyle(config.Styles.Underline),
		Background: t.createStyle(config.Styles.Background),
		List:       t.createStyle(config.Styles.List),
	}
}

// createStyle converts a StyleConfig to a lipgloss.Style.
func (t *ThemeService) createStyle(config neurotypes.StyleConfig) lipgloss.Style {
	style := lipgloss.NewStyle()

	// Handle foreground color
	if config.Foreground != nil {
		if color := t.parseColor(config.Foreground); color != nil {
			style = style.Foreground(color)
		}
	}

	// Handle background color
	if config.Background != nil {
		if color := t.parseColor(config.Background); color != nil {
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
func (t *ThemeService) parseColor(colorValue interface{}) lipgloss.TerminalColor {
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
func (t *ThemeService) createFallbackTheme(name string) *Theme {
	return &Theme{
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
		List:       lipgloss.NewStyle(),
	}
}

// GetAvailableThemes returns a list of available theme names
func (t *ThemeService) GetAvailableThemes() []string {
	if !t.initialized {
		return []string{}
	}

	themes := make([]string, 0, len(t.themes))
	for name := range t.themes {
		themes = append(themes, name)
	}
	return themes
}

// GetTheme returns a specific theme by name
func (t *ThemeService) GetTheme(name string) (*Theme, bool) {
	if !t.initialized {
		return nil, false
	}

	theme, exists := t.themes[name]
	return theme, exists
}

// GetThemeByName retrieves a theme by name with support for aliases and case-insensitive matching.
// Supports aliases like "dark1" -> "dark". Always returns a valid theme object, never fails.
// For invalid themes, logs a warning and returns the plain theme.
func (t *ThemeService) GetThemeByName(theme string) *Theme {
	if !t.initialized {
		return t.GetDefaultTheme()
	}

	normalizedTheme := strings.ToLower(strings.TrimSpace(theme))

	switch normalizedTheme {
	case "", "plain":
		return t.themes["plain"]
	case "dark1", "dark":
		if themeObj, exists := t.themes["dark"]; exists {
			return themeObj
		}
		return t.themes["plain"]
	case "default":
		if themeObj, exists := t.themes["default"]; exists {
			return themeObj
		}
		return t.themes["plain"]
	case "light":
		if themeObj, exists := t.themes["light"]; exists {
			return themeObj
		}
		return t.themes["plain"]
	default:
		// Invalid theme - log warning and return plain theme
		logger.Debug("Invalid theme requested, using plain theme", "theme", theme, "available", t.GetAvailableThemes())
		return t.themes["plain"]
	}
}

// GetDefaultTheme returns the plain theme (no styling) for fallback scenarios.
func (t *ThemeService) GetDefaultTheme() *Theme {
	if !t.initialized {
		// Return a basic plain theme if service not initialized
		return &Theme{
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
			List:       lipgloss.NewStyle(),
		}
	}
	return t.themes["plain"]
}

// CreateList creates a new list with theme styling applied
func (t *Theme) CreateList() *list.List {
	return list.New().EnumeratorStyle(t.List)
}

// CreateSimpleList creates a simple list from string array
func (t *Theme) CreateSimpleList(items []string) *list.List {
	l := t.CreateList()
	for _, item := range items {
		l.Item(item)
	}
	return l
}

// CreateGroupedList creates a nested list from grouped data
func (t *Theme) CreateGroupedList(groups map[string][]string) *list.List {
	var items []interface{}
	for groupName, groupItems := range groups {
		if len(groupItems) > 0 {
			items = append(items, t.Warning.Render(groupName))
			subList := t.CreateList()
			for _, item := range groupItems {
				subList.Item(item)
			}
			items = append(items, subList)
		}
	}
	return list.New(items...).EnumeratorStyle(t.List)
}

// CreateVariableList creates a formatted list for variables with name-value pairs
func (t *Theme) CreateVariableList(vars map[string]string) *list.List {
	l := t.CreateList()
	for name, value := range vars {
		formattedVar := fmt.Sprintf("%s = %s",
			t.Variable.Render(name),
			t.Info.Render(value))
		l.Item(formattedVar)
	}
	return l
}

func init() {
	// Register the ThemeService with the global registry
	if err := GlobalRegistry.RegisterService(NewThemeService()); err != nil {
		panic(fmt.Sprintf("failed to register theme service: %v", err))
	}
}
