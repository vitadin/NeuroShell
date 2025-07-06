package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// RenderHelp renders structured help information with professional styling
func (r *RenderService) RenderHelp(helpInfo neurotypes.HelpInfo, styled bool) (string, error) {
	if !r.initialized {
		return "", fmt.Errorf("render service not initialized")
	}

	// If not styled, render as plain text
	if !styled {
		return r.renderHelpPlainText(helpInfo), nil
	}

	// Render with styling
	return r.renderHelpStyled(helpInfo)
}

// renderHelpPlainText renders help information as plain text (backward compatibility)
func (r *RenderService) renderHelpPlainText(helpInfo neurotypes.HelpInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Command: %s\n", helpInfo.Command))
	result.WriteString(fmt.Sprintf("Description: %s\n", helpInfo.Description))
	result.WriteString(fmt.Sprintf("Usage: %s\n", helpInfo.Usage))
	result.WriteString(fmt.Sprintf("Parse Mode: %s\n", r.parseModeToString(helpInfo.ParseMode)))

	if len(helpInfo.Options) > 0 {
		result.WriteString("\nOptions:\n")
		for _, option := range helpInfo.Options {
			defaultStr := ""
			if option.Default != "" {
				defaultStr = fmt.Sprintf(" (default: %s)", option.Default)
			}
			requiredStr := ""
			if option.Required {
				requiredStr = " (required)"
			}
			result.WriteString(fmt.Sprintf("  %s - %s%s%s\n", option.Name, option.Description, defaultStr, requiredStr))
		}
	}

	if len(helpInfo.Examples) > 0 {
		result.WriteString("\nExamples:\n")
		for _, example := range helpInfo.Examples {
			result.WriteString(fmt.Sprintf("  %s\n", example.Command))
			if example.Description != "" {
				result.WriteString("    %% " + example.Description + "\n")
			}
		}
	}

	if len(helpInfo.Notes) > 0 {
		result.WriteString("\nNotes:\n")
		for _, note := range helpInfo.Notes {
			result.WriteString(fmt.Sprintf("  %s\n", note))
		}
	}

	return result.String()
}

// renderHelpStyled renders help information with professional styling
func (r *RenderService) renderHelpStyled(helpInfo neurotypes.HelpInfo) (string, error) {
	theme, exists := r.GetTheme("default")
	if !exists {
		return "", fmt.Errorf("default theme not found")
	}

	var result strings.Builder

	// Title with border
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Command.GetForeground()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Info.GetForeground())

	title := fmt.Sprintf("Command: %s", helpInfo.Command)
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Description
	descStyle := theme.Info.Bold(false)
	result.WriteString(descStyle.Render("Description: "))
	result.WriteString(helpInfo.Description)
	result.WriteString("\n\n")

	// Usage with syntax highlighting
	usageStyle := theme.Bold.Foreground(theme.Success.GetForeground())
	result.WriteString(usageStyle.Render("Usage: "))
	styledUsage := r.highlightNeuroShellSyntax(helpInfo.Usage, theme)
	result.WriteString(styledUsage)
	result.WriteString("\n\n")

	// Parse Mode
	parseModeStyle := theme.Info.Bold(false)
	result.WriteString(parseModeStyle.Render("Parse Mode: "))
	result.WriteString(r.parseModeToString(helpInfo.ParseMode))
	result.WriteString("\n")

	// Options section
	if len(helpInfo.Options) > 0 {
		result.WriteString("\n")
		optionHeaderStyle := theme.Bold.Foreground(theme.Warning.GetForeground())
		result.WriteString(optionHeaderStyle.Render("Options:"))
		result.WriteString("\n")

		for _, option := range helpInfo.Options {
			optionNameStyle := theme.Variable.Bold(true)
			result.WriteString("  ")
			result.WriteString(optionNameStyle.Render(option.Name))
			result.WriteString(" - ")
			result.WriteString(option.Description)

			if option.Default != "" {
				defaultStyle := theme.Info.Italic(true)
				result.WriteString(defaultStyle.Render(fmt.Sprintf(" (default: %s)", option.Default)))
			}
			if option.Required {
				requiredStyle := theme.Error.Bold(true)
				result.WriteString(requiredStyle.Render(" (required)"))
			}
			result.WriteString("\n")
		}
	}

	// Examples section
	if len(helpInfo.Examples) > 0 {
		result.WriteString("\n")
		exampleHeaderStyle := theme.Bold.Foreground(theme.Success.GetForeground())
		result.WriteString(exampleHeaderStyle.Render("Examples:"))
		result.WriteString("\n")

		for _, example := range helpInfo.Examples {
			result.WriteString("  ")
			styledExample := r.highlightNeuroShellSyntax(example.Command, theme)
			result.WriteString(styledExample)
			result.WriteString("\n")
			if example.Description != "" {
				commentStyle := theme.Info.Italic(true)
				result.WriteString("    ")
				result.WriteString(commentStyle.Render("%% " + example.Description))
				result.WriteString("\n")
			}
		}
	}

	// Notes section
	if len(helpInfo.Notes) > 0 {
		result.WriteString("\n")
		noteHeaderStyle := theme.Bold.Foreground(theme.Warning.GetForeground())
		result.WriteString(noteHeaderStyle.Render("Notes:"))
		result.WriteString("\n")

		for _, note := range helpInfo.Notes {
			noteStyle := theme.Info.Italic(true)
			result.WriteString("  ")
			result.WriteString(noteStyle.Render(note))
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// parseModeToString converts parse mode enum to readable string
func (r *RenderService) parseModeToString(mode neurotypes.ParseMode) string {
	switch mode {
	case neurotypes.ParseModeKeyValue:
		return "Key-Value (supports [key=value] syntax)"
	case neurotypes.ParseModeRaw:
		return "Raw (passes input directly without parsing)"
	case neurotypes.ParseModeWithOptions:
		return "With Options (supports additional options)"
	default:
		return "Unknown"
	}
}

func init() {
	// Register the RenderService with the global registry
	if err := GlobalRegistry.RegisterService(NewRenderService()); err != nil {
		panic(fmt.Sprintf("failed to register render service: %v", err))
	}
}
