package services

import (
	"fmt"
	"regexp"
	"strings"

	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
	"github.com/muesli/termenv"
)

// PromptColorService processes color markup in shell prompt templates.
// It converts semantic color markup like {{color:info}}text{{/color}} into ANSI escape codes
// using lipgloss and the theme system.
type PromptColorService struct {
	initialized  bool
	themeService *ThemeService
	colorRegex   *regexp.Regexp
	styleRegex   *regexp.Regexp
}

// NewPromptColorService creates a new PromptColorService instance.
func NewPromptColorService() *PromptColorService {
	return &PromptColorService{
		initialized: false,
		// Regex for color markup: {{color:semantic}}text{{/color}}
		colorRegex: regexp.MustCompile(`\{\{color:([^}]+)\}\}(.*?)\{\{/color\}\}`),
		// Regex for style markup: {{bold}}, {{italic}}, {{underline}}
		// Note: We'll handle matching closing tags in the processing function
		styleRegex: regexp.MustCompile(`\{\{(bold|italic|underline)\}\}(.*?)\{\{/(bold|italic|underline)\}\}`),
	}
}

// Name returns the service name "prompt_color" for registration.
func (p *PromptColorService) Name() string {
	return "prompt_color"
}

// Initialize sets up the service and obtains reference to ThemeService.
func (p *PromptColorService) Initialize() error {
	if p.initialized {
		return nil
	}

	// Get theme service for semantic color support
	registry := GetGlobalRegistry()
	themeServiceInterface, err := registry.GetService("theme")
	if err != nil {
		// Theme service is optional - we can still provide direct color support
		p.themeService = nil
	} else {
		p.themeService = themeServiceInterface.(*ThemeService)
	}

	p.initialized = true
	return nil
}

// GetServiceInfo returns service information for debugging.
func (p *PromptColorService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":          p.Name(),
		"initialized":   p.initialized,
		"type":          "prompt_color",
		"description":   "Processes color markup in shell prompt templates",
		"theme_support": p.themeService != nil,
	}
}

// ProcessColorMarkup processes color and style markup in a prompt string.
// Converts {{color:semantic}}text{{/color}} and {{bold}}text{{/bold}} into ANSI codes.
func (p *PromptColorService) ProcessColorMarkup(input string) string {
	if !p.initialized {
		return input // Return unchanged if not initialized
	}

	// Check if colors are disabled
	if lipgloss.ColorProfile() == termenv.Ascii {
		// Strip all markup but keep the text
		result := p.stripColorMarkup(input)
		result = p.stripStyleMarkup(result)
		return result
	}

	// Process color markup first
	result := p.processColorTags(input)

	// Then process style markup
	result = p.processStyleTags(result)

	return result
}

// processColorTags handles {{color:semantic}}text{{/color}} markup.
func (p *PromptColorService) processColorTags(input string) string {
	return p.colorRegex.ReplaceAllStringFunc(input, func(match string) string {
		matches := p.colorRegex.FindStringSubmatch(match)
		if len(matches) != 3 {
			return match // Return unchanged if regex doesn't match properly
		}

		colorSpec := matches[1]
		text := matches[2]

		style := p.createColorStyle(colorSpec)
		return style.Render(text)
	})
}

// processStyleTags handles {{bold}}, {{italic}}, {{underline}} markup.
func (p *PromptColorService) processStyleTags(input string) string {
	return p.styleRegex.ReplaceAllStringFunc(input, func(match string) string {
		matches := p.styleRegex.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match // Return unchanged if regex doesn't match properly
		}

		openingTag := matches[1]
		text := matches[2]
		closingTag := matches[3]

		// Validate that opening and closing tags match
		if openingTag != closingTag {
			return match // Return unchanged if tags don't match
		}

		style := p.createStyleFromType(openingTag)
		return style.Render(text)
	})
}

// createColorStyle creates a lipgloss style based on color specification.
func (p *PromptColorService) createColorStyle(colorSpec string) lipgloss.Style {
	// Try semantic colors first if theme service is available
	if p.themeService != nil {
		if semanticStyle := p.trySemanticColor(colorSpec); semanticStyle != nil {
			return *semanticStyle
		}
	}

	// Handle direct colors (hex, named colors)
	return p.createDirectColorStyle(colorSpec)
}

// trySemanticColor attempts to get a style from the theme service.
func (p *PromptColorService) trySemanticColor(semantic string) *lipgloss.Style {
	if p.themeService == nil {
		return nil
	}

	// Get the active theme (default to "default" theme)
	theme, exists := p.themeService.GetTheme("default")
	if !exists {
		return nil
	}

	// Map semantic names to theme styles
	switch semantic {
	case "info":
		return &theme.Info
	case "success":
		return &theme.Success
	case "warning":
		return &theme.Warning
	case "error":
		return &theme.Error
	case "command":
		return &theme.Command
	case "variable":
		return &theme.Variable
	case "keyword":
		return &theme.Keyword
	case "highlight":
		return &theme.Highlight
	case "bold":
		return &theme.Bold
	case "italic":
		return &theme.Italic
	case "underline":
		return &theme.Underline
	}

	return nil
}

// createDirectColorStyle creates a style with direct color specification.
func (p *PromptColorService) createDirectColorStyle(colorSpec string) lipgloss.Style {
	style := lipgloss.NewStyle()

	// Handle hex colors
	if strings.HasPrefix(colorSpec, "#") {
		return style.Foreground(lipgloss.Color(colorSpec))
	}

	// Handle named colors
	switch strings.ToLower(colorSpec) {
	case "red":
		return style.Foreground(lipgloss.Color("1"))
	case "green":
		return style.Foreground(lipgloss.Color("2"))
	case "yellow":
		return style.Foreground(lipgloss.Color("3"))
	case "blue":
		return style.Foreground(lipgloss.Color("4"))
	case "magenta", "purple":
		return style.Foreground(lipgloss.Color("5"))
	case "cyan":
		return style.Foreground(lipgloss.Color("6"))
	case "white":
		return style.Foreground(lipgloss.Color("7"))
	case "gray", "grey":
		return style.Foreground(lipgloss.Color("8"))
	case "bright-red":
		return style.Foreground(lipgloss.Color("9"))
	case "bright-green":
		return style.Foreground(lipgloss.Color("10"))
	case "bright-yellow":
		return style.Foreground(lipgloss.Color("11"))
	case "bright-blue":
		return style.Foreground(lipgloss.Color("12"))
	case "bright-magenta", "bright-purple":
		return style.Foreground(lipgloss.Color("13"))
	case "bright-cyan":
		return style.Foreground(lipgloss.Color("14"))
	case "bright-white":
		return style.Foreground(lipgloss.Color("15"))
	default:
		// Try to parse as ANSI color number or return unstyled
		return style.Foreground(lipgloss.Color(colorSpec))
	}
}

// createStyleFromType creates a lipgloss style for text styling.
func (p *PromptColorService) createStyleFromType(styleType string) lipgloss.Style {
	switch styleType {
	case "bold":
		return lipgloss.NewStyle().Bold(true)
	case "italic":
		return lipgloss.NewStyle().Italic(true)
	case "underline":
		return lipgloss.NewStyle().Underline(true)
	default:
		return lipgloss.NewStyle()
	}
}

// stripColorMarkup removes color markup while preserving the text content.
func (p *PromptColorService) stripColorMarkup(input string) string {
	return p.colorRegex.ReplaceAllString(input, "$2")
}

// stripStyleMarkup removes style markup while preserving the text content.
func (p *PromptColorService) stripStyleMarkup(input string) string {
	return p.styleRegex.ReplaceAllStringFunc(input, func(match string) string {
		matches := p.styleRegex.FindStringSubmatch(match)
		if len(matches) >= 3 {
			return matches[2] // Return just the text content
		}
		return match // Return unchanged if no match
	})
}

// IsColorSupported returns true if the terminal supports colors.
func (p *PromptColorService) IsColorSupported() bool {
	return lipgloss.ColorProfile() != termenv.Ascii
}

// CommandHighlighter implements readline.Painter to highlight command prefixes
type CommandHighlighter struct {
	colorService *PromptColorService
	commandRegex *regexp.Regexp
}

// CreateCommandHighlighter returns a readline.Painter that highlights command prefixes
func (p *PromptColorService) CreateCommandHighlighter() readline.Painter {
	return &CommandHighlighter{
		colorService: p,
		// Regex to match command prefix: \command or \command[options]
		commandRegex: regexp.MustCompile(`^\\([a-zA-Z][a-zA-Z0-9-]*(?:\[[^\]]*\])?)`),
	}
}

// Paint implements readline.Painter interface to highlight command prefixes
func (h *CommandHighlighter) Paint(line []rune, _ int) []rune {
	// Convert runes to string for processing
	input := string(line)

	// Skip highlighting if colors are not supported
	if !h.colorService.IsColorSupported() {
		return line
	}

	// Parse command into three parts: command name, options, and remaining text
	commandName, options, remainingText := h.parseCommandParts(input)
	if commandName == "" {
		// No command found, return original line
		return line
	}

	// Apply three-part highlighting
	highlighted := h.applyThreePartHighlighting(commandName, options, remainingText)

	// Convert back to runes
	return []rune(highlighted)
}

// parseCommandParts parses input into three parts: command name, options, and remaining text
func (h *CommandHighlighter) parseCommandParts(input string) (commandName, options, remainingText string) {
	// Regex to match command name and optional options: \command[options]
	commandRegex := regexp.MustCompile(`^\\([a-zA-Z][a-zA-Z0-9-]*)(\[[^\]]*\])?(.*)$`)
	matches := commandRegex.FindStringSubmatch(input)

	if len(matches) < 2 {
		// No command found
		return "", "", ""
	}

	// Extract parts
	commandName = "\\" + matches[1] // Include the backslash prefix
	options = matches[2]            // Options in brackets (could be empty)
	remainingText = matches[3]      // Everything after the command and options

	return commandName, options, remainingText
}

// applyThreePartHighlighting applies different colors to command name, options, and remaining text
func (h *CommandHighlighter) applyThreePartHighlighting(commandName, options, remainingText string) string {
	const (
		// Basic ANSI color codes - safe and widely supported
		ansiBrightBlue  = "\033[94m" // Bright blue for command names
		ansiBrightGreen = "\033[92m" // Bright green for options
		ansiYellow      = "\033[33m" // Yellow as fallback for options
		ansiCyan        = "\033[36m" // Cyan as fallback for commands
		ansiReset       = "\033[0m"  // Reset formatting
	)

	// Choose colors based on theme if available, otherwise use defaults
	commandColor := ansiBrightBlue
	optionsColor := ansiBrightGreen

	if h.colorService.themeService != nil {
		// Try to get semantic colors from theme
		if theme, exists := h.colorService.themeService.GetTheme("default"); exists {
			// Use theme command color if available
			if theme.Command.GetForeground() != lipgloss.Color("") {
				commandColor = ansiCyan
			}
			// Use theme variable color for options if available
			if theme.Variable.GetForeground() != lipgloss.Color("") {
				optionsColor = ansiYellow
			}
		}
	}

	// Build highlighted string
	result := commandColor + commandName + ansiReset
	if options != "" {
		result += optionsColor + options + ansiReset
	}
	result += remainingText

	return result
}

// Interface compliance check
var _ neurotypes.Service = (*PromptColorService)(nil)

func init() {
	// Register the PromptColorService with the global registry
	if err := GlobalRegistry.RegisterService(NewPromptColorService()); err != nil {
		panic(fmt.Sprintf("failed to register prompt color service: %v", err))
	}
}
