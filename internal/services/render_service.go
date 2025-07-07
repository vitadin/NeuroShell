package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// RenderService provides text rendering and styling operations for NeuroShell.
// It uses lipgloss for terminal styling and supports keyword highlighting with themes.
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

// RenderOptions contains configuration for rendering text
type RenderOptions struct {
	Keywords   []string
	Theme      string
	Style      string
	Color      string
	Background string
	Bold       bool
	Italic     bool
	Underline  bool
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
}

// RenderText applies styling to text based on the provided options
func (r *RenderService) RenderText(text string, options RenderOptions) (string, error) {
	if !r.initialized {
		return "", fmt.Errorf("render service not initialized")
	}

	// Get theme
	theme, exists := r.themes[options.Theme]
	if !exists {
		theme = r.themes["default"] // Fallback to default theme
	}

	// Start with the input text
	result := text

	// Apply keyword highlighting first (so it takes precedence)
	if len(options.Keywords) > 0 {
		result = r.highlightKeywords(result, options.Keywords, theme)
	}

	// Apply NeuroShell-specific highlighting
	result = r.highlightNeuroShellSyntax(result, theme)

	// Apply global styling if specified
	if options.Style != "" || options.Color != "" || options.Background != "" ||
		options.Bold || options.Italic || options.Underline {
		result = r.applyGlobalStyling(result, options, theme)
	}

	return result, nil
}

// highlightKeywords highlights specific keywords in the text
func (r *RenderService) highlightKeywords(text string, keywords []string, theme *RenderTheme) string {
	result := text

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		// Escape special regex characters in the keyword
		escaped := regexp.QuoteMeta(keyword)

		// Create regex pattern for whole words (but allow backslash prefix for commands)
		pattern := fmt.Sprintf(`(\b%s\b|\\%s\b)`, escaped, escaped)
		re := regexp.MustCompile(pattern)

		// Replace matches with styled versions
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return theme.Keyword.Render(match)
		})
	}

	return result
}

// highlightNeuroShellSyntax applies syntax highlighting for NeuroShell-specific patterns
func (r *RenderService) highlightNeuroShellSyntax(text string, theme *RenderTheme) string {
	result := text

	// Highlight variables: ${variable_name}
	variablePattern := regexp.MustCompile(`\$\{[^}]+\}`)
	result = variablePattern.ReplaceAllStringFunc(result, func(match string) string {
		return theme.Variable.Render(match)
	})

	// Highlight commands: \command (but only if not already styled)
	commandPattern := regexp.MustCompile(`\\[a-zA-Z_][a-zA-Z0-9_-]*`)
	result = commandPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Check if this text is already styled (contains ANSI sequences)
		if strings.Contains(match, "\x1b[") {
			return match // Already styled, don't re-style
		}
		return theme.Command.Render(match)
	})

	return result
}

// applyGlobalStyling applies global style options to the entire text
func (r *RenderService) applyGlobalStyling(text string, options RenderOptions, theme *RenderTheme) string {
	style := lipgloss.NewStyle()

	// Apply basic styling
	if options.Bold {
		style = style.Bold(true)
	}
	if options.Italic {
		style = style.Italic(true)
	}
	if options.Underline {
		style = style.Underline(true)
	}

	// Apply colors
	if options.Color != "" {
		style = style.Foreground(lipgloss.Color(options.Color))
	}
	if options.Background != "" {
		style = style.Background(lipgloss.Color(options.Background))
	}

	// Apply named styles
	switch options.Style {
	case "bold":
		style = theme.Bold
	case "italic":
		style = theme.Italic
	case "underline":
		style = theme.Underline
	case "success":
		style = theme.Success
	case "error":
		style = theme.Error
	case "warning":
		style = theme.Warning
	case "info":
		style = theme.Info
	case "highlight":
		style = theme.Highlight
	}

	return style.Render(text)
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

// RenderWithTheme is a convenience method to render text with a specific theme
func (r *RenderService) RenderWithTheme(text string, keywords []string, themeName string) (string, error) {
	options := RenderOptions{
		Keywords: keywords,
		Theme:    themeName,
	}
	return r.RenderText(text, options)
}

// HighlightKeywords is a convenience method for simple keyword highlighting
func (r *RenderService) HighlightKeywords(text string, keywords []string) (string, error) {
	return r.RenderWithTheme(text, keywords, "default")
}

// GetThemeByName retrieves a theme by name with support for aliases and case-insensitive matching.
// Supports aliases like "dark1" -> "dark" and returns nil for plain text themes ("" or "plain").
// Returns the theme object and a boolean indicating success.
func (r *RenderService) GetThemeByName(theme string) (*RenderTheme, bool) {
	if !r.initialized {
		return nil, false
	}

	normalizedTheme := strings.ToLower(strings.TrimSpace(theme))

	switch normalizedTheme {
	case "", "plain":
		return nil, true // No theme (plain text)
	case "dark1", "dark":
		return r.GetTheme("dark")
	case "default":
		return r.GetTheme("default")
	default:
		// Invalid theme - log warning and return false
		logger.Debug("Invalid theme requested", "theme", theme, "available", r.GetAvailableThemes())
		return nil, false
	}
}

func init() {
	// Register the RenderService with the global registry
	if err := GlobalRegistry.RegisterService(NewRenderService()); err != nil {
		panic(fmt.Sprintf("failed to register render service: %v", err))
	}
}
