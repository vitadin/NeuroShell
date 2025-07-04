package builtin

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// RenderCommand implements the \render command for styling and highlighting text.
// It provides lipgloss-based text rendering with keyword highlighting and theme support.
type RenderCommand struct{}

// Name returns the command name "render" for registration and lookup.
func (c *RenderCommand) Name() string {
	return "render"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *RenderCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the render command does.
func (c *RenderCommand) Description() string {
	return "Style and highlight text using lipgloss with keyword support"
}

// Usage returns the syntax and usage examples for the render command.
func (c *RenderCommand) Usage() string {
	return "\\render[keywords=[\\get,\\set], style=bold, theme=dark, to=var] text to render"
}

// HelpInfo returns structured help information for the render command.
func (c *RenderCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\render[style=bold] Hello World",
				Description: "Render text with bold styling",
			},
		},
	}
}

// Execute applies styling and highlighting to text based on the provided options.
// Options:
//   - keywords: array of keywords to highlight (e.g., [\\get,\\set])
//   - style: named style (bold, italic, underline, success, error, warning, info, highlight)
//   - theme: color theme (default, dark, light)
//   - color: foreground color (hex code or color name)
//   - background: background color (hex code or color name)
//   - bold: make text bold (true/false)
//   - italic: make text italic (true/false)
//   - underline: make text underlined (true/false)
//   - to: variable to store result (default: ${_output})
//   - silent: suppress console output (true/false, default: false)
func (c *RenderCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	if input == "" {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Get render service
	renderService, err := c.getRenderService()
	if err != nil {
		return fmt.Errorf("render service not available: %w", err)
	}

	// Get variable service for storing result
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse render options
	options, err := c.parseRenderOptions(args, ctx)
	if err != nil {
		return fmt.Errorf("failed to parse render options: %w", err)
	}

	// Apply styling to the input text
	styledText, err := renderService.RenderText(input, options)
	if err != nil {
		return fmt.Errorf("failed to render text: %w", err)
	}

	// Store result in target variable
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
		// Store in system variable
		err = variableService.SetSystemVariable(targetVar, styledText, ctx)
	} else {
		// Store in user variable
		err = variableService.Set(targetVar, styledText, ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to store result in variable '%s': %w", targetVar, err)
	}

	// Parse silent option
	silentStr := args["silent"]
	silent := false
	if silentStr != "" {
		silent, err = strconv.ParseBool(silentStr)
		if err != nil {
			return fmt.Errorf("invalid value for silent option: %s (must be true or false)", silentStr)
		}
	}

	// Output to console unless silent mode is enabled
	if !silent {
		fmt.Print(styledText)
		// Only add newline if the styled text doesn't already end with one
		if len(styledText) > 0 && styledText[len(styledText)-1] != '\n' {
			fmt.Println()
		}
	}

	return nil
}

// parseRenderOptions parses command arguments into RenderOptions
func (c *RenderCommand) parseRenderOptions(args map[string]string, ctx neurotypes.Context) (services.RenderOptions, error) {
	options := services.RenderOptions{
		Theme: "default", // Default theme
	}

	// Parse keywords array
	if keywordsStr, exists := args["keywords"]; exists {
		keywords := parser.ParseArrayValue(keywordsStr)

		// Interpolate variables in keywords
		interpolationService, err := c.getInterpolationService()
		if err == nil {
			for i, keyword := range keywords {
				if interpolated, err := interpolationService.InterpolateString(keyword, ctx); err == nil {
					keywords[i] = interpolated
				}
			}
		}

		options.Keywords = keywords
	}

	// Parse theme
	if theme, exists := args["theme"]; exists {
		options.Theme = theme
	}

	// Parse style
	if style, exists := args["style"]; exists {
		options.Style = style
	}

	// Parse colors
	if color, exists := args["color"]; exists {
		options.Color = color
	}
	if background, exists := args["background"]; exists {
		options.Background = background
	}

	// Parse boolean options
	if boldStr, exists := args["bold"]; exists {
		if bold, err := strconv.ParseBool(boldStr); err == nil {
			options.Bold = bold
		}
	}
	if italicStr, exists := args["italic"]; exists {
		if italic, err := strconv.ParseBool(italicStr); err == nil {
			options.Italic = italic
		}
	}
	if underlineStr, exists := args["underline"]; exists {
		if underline, err := strconv.ParseBool(underlineStr); err == nil {
			options.Underline = underline
		}
	}

	return options, nil
}

// getRenderService retrieves the render service from the global registry
func (c *RenderCommand) getRenderService() (*services.RenderService, error) {
	service, err := services.GetGlobalRegistry().GetService("render")
	if err != nil {
		return nil, err
	}

	renderService, ok := service.(*services.RenderService)
	if !ok {
		return nil, fmt.Errorf("render service has incorrect type")
	}

	return renderService, nil
}

// getVariableService retrieves the variable service from the global registry
func (c *RenderCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}

// getInterpolationService retrieves the interpolation service from the global registry
func (c *RenderCommand) getInterpolationService() (*services.InterpolationService, error) {
	service, err := services.GetGlobalRegistry().GetService("interpolation")
	if err != nil {
		return nil, err
	}

	interpolationService, ok := service.(*services.InterpolationService)
	if !ok {
		return nil, fmt.Errorf("interpolation service has incorrect type")
	}

	return interpolationService, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&RenderCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register render command: %v", err))
	}
}
