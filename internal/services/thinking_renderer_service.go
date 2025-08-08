package services

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// ThinkingRendererService provides sophisticated rendering capabilities for thinking blocks.
// This service implements provider-specific, theme-aware rendering with pluggable architecture.
type ThinkingRendererService struct {
	initialized bool
}

// NewThinkingRendererService creates a new ThinkingRendererService instance.
func NewThinkingRendererService() *ThinkingRendererService {
	return &ThinkingRendererService{
		initialized: false,
	}
}

// Name returns the service name "thinking-renderer" for registration.
func (t *ThinkingRendererService) Name() string {
	return "thinking-renderer"
}

// Initialize sets up the ThinkingRendererService.
func (t *ThinkingRendererService) Initialize() error {
	logger.ServiceOperation("thinking-renderer", "initialize", "starting")
	t.initialized = true
	logger.ServiceOperation("thinking-renderer", "initialize", "completed")
	return nil
}

// RenderThinkingBlocks renders thinking blocks with theme-aware, provider-specific formatting.
// This provides sophisticated styling that adapts to the current theme and provider.
func (t *ThinkingRendererService) RenderThinkingBlocks(blocks []neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	if !t.initialized {
		logger.Error("ThinkingRendererService not initialized")
		return ""
	}

	if len(blocks) == 0 {
		return ""
	}

	// Check if thinking should be shown at all
	if !config.ShowThinking() {
		return ""
	}

	var result strings.Builder

	for _, block := range blocks {
		rendered := t.RenderSingleBlock(block, config)
		result.WriteString(rendered)
		logger.Debug("Thinking block rendered", "provider", block.Provider, "type", block.Type, "content_length", len(block.Content))
	}

	return result.String()
}

// RenderSingleBlock renders an individual thinking block with provider-specific styling.
func (t *ThinkingRendererService) RenderSingleBlock(block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	if !t.initialized {
		return ""
	}

	// Route to provider-specific rendering
	switch strings.ToLower(block.Provider) {
	case "anthropic":
		return t.renderAnthropicBlock(block, config)
	case "gemini":
		return t.renderGeminiBlock(block, config)
	case "openai":
		return t.renderOpenAIBlock(block, config)
	default:
		return t.renderGenericBlock(block, config)
	}
}

// GetSupportedProviders returns the list of providers this renderer supports.
func (t *ThinkingRendererService) GetSupportedProviders() []string {
	return []string{"anthropic", "gemini", "openai", "generic"}
}

// RenderThinkingBlocksLegacy provides backward compatibility for the old interface.
// This method uses default configuration settings for rendering.
// Deprecated: Use RenderThinkingBlocks(blocks, config) instead.
func (t *ThinkingRendererService) RenderThinkingBlocksLegacy(blocks []neurotypes.ThinkingBlock) string {
	// Create a default configuration
	defaultConfig := &DefaultRenderConfig{
		showThinking:  true,
		thinkingStyle: "full",
		compactMode:   false,
		maxWidth:      80,
		theme:         "default",
	}

	return t.RenderThinkingBlocks(blocks, defaultConfig)
}

// IsInitialized returns true if the service has been initialized.
func (t *ThinkingRendererService) IsInitialized() bool {
	return t.initialized
}

// renderAnthropicBlock renders a thinking block from Anthropic (Claude) with subtle styling.
func (t *ThinkingRendererService) renderAnthropicBlock(block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	return t.createThinkingBox("Claude's internal reasoning:", block, config)
}

// renderGeminiBlock renders a thinking block from Google Gemini with subtle styling.
func (t *ThinkingRendererService) renderGeminiBlock(block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	return t.createThinkingBox("Gemini's thought process:", block, config)
}

// renderOpenAIBlock renders a thinking block from OpenAI with subtle styling.
func (t *ThinkingRendererService) renderOpenAIBlock(block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	providerLabel := "OpenAI reasoning trace:"
	if block.Type == "reasoning" {
		providerLabel = "OpenAI reasoning trace:"
	}
	return t.createThinkingBox(providerLabel, block, config)
}

// renderGenericBlock renders a thinking block from an unknown provider with default styling.
func (t *ThinkingRendererService) renderGenericBlock(block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	return t.createThinkingBox("Internal processing:", block, config)
}

// createThinkingBox creates a styled thinking block using simple, clean formatting.
func (t *ThinkingRendererService) createThinkingBox(providerLabel string, block neurotypes.ThinkingBlock, config neurotypes.RenderConfig) string {
	// Process content based on style preference
	content := block.Content
	thinkingStyle := config.GetThinkingStyle()

	if thinkingStyle == "summary" {
		content = t.createContentSummary(content)
	}

	// Apply width constraints
	maxWidth := config.GetMaxWidth()
	if maxWidth > 0 && maxWidth < 120 {
		content = t.wrapContent(content, maxWidth-4) // Account for indentation
	}

	// Create simple, clean output without boxes or complex styling
	var result strings.Builder

	if config.IsCompactMode() {
		// Compact mode: single line format
		result.WriteString(fmt.Sprintf("[%s] %s\n", providerLabel, content))
	} else {
		// Normal mode: clean multi-line format with simple indentation
		result.WriteString("\n")
		result.WriteString(fmt.Sprintf("--- %s ---\n", providerLabel))

		// Indent content lines
		contentLines := strings.Split(content, "\n")
		for _, line := range contentLines {
			if strings.TrimSpace(line) != "" {
				result.WriteString("  " + line + "\n")
			}
		}

		// Add end marker
		result.WriteString("--- End thinking ---\n")
		result.WriteString("\n")
	}

	return result.String()
}

// createContentSummary creates a summary of long thinking content.
func (t *ThinkingRendererService) createContentSummary(content string) string {
	lines := strings.Split(content, "\n")

	// If content is short, return as-is
	if len(content) <= 200 && len(lines) <= 3 {
		return content
	}

	// Take first 150 characters or first 2 lines, whichever is shorter
	summary := content
	if len(content) > 150 {
		summary = content[:150]
		// Try to break at a word boundary
		if lastSpace := strings.LastIndex(summary, " "); lastSpace > 120 {
			summary = content[:lastSpace]
		}
		summary += "..."
	}

	// Limit to 2 lines max
	summaryLines := strings.Split(summary, "\n")
	if len(summaryLines) > 2 {
		summary = strings.Join(summaryLines[:2], "\n") + "..."
	}

	return summary
}

// wrapContent wraps content to fit within specified width.
func (t *ThinkingRendererService) wrapContent(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return content
	}

	words := strings.Fields(content)
	if len(words) == 0 {
		return content
	}

	var result strings.Builder
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+len(word)+1 > maxWidth {
			result.WriteString(strings.TrimSpace(currentLine.String()))
			result.WriteString("\n")
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add the last line
	if currentLine.Len() > 0 {
		result.WriteString(strings.TrimSpace(currentLine.String()))
	}

	return result.String()
}

// DefaultRenderConfig provides a basic implementation of RenderConfig for backward compatibility.
type DefaultRenderConfig struct {
	showThinking  bool
	thinkingStyle string
	compactMode   bool
	maxWidth      int
	theme         string
}

// GetStyle returns a basic lipgloss style for the given element.
func (c *DefaultRenderConfig) GetStyle(element string) lipgloss.Style {
	switch element {
	case "info":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Medium gray
	case "italic":
		return lipgloss.NewStyle().Italic(true)
	case "background":
		return lipgloss.NewStyle()
	default:
		return lipgloss.NewStyle()
	}
}

// GetTheme returns the theme name.
func (c *DefaultRenderConfig) GetTheme() string {
	return c.theme
}

// IsCompactMode returns whether compact mode is enabled.
func (c *DefaultRenderConfig) IsCompactMode() bool {
	return c.compactMode
}

// GetMaxWidth returns the maximum width for content.
func (c *DefaultRenderConfig) GetMaxWidth() int {
	return c.maxWidth
}

// ShowThinking returns whether thinking blocks should be displayed.
func (c *DefaultRenderConfig) ShowThinking() bool {
	return c.showThinking
}

// GetThinkingStyle returns the thinking display style.
func (c *DefaultRenderConfig) GetThinkingStyle() string {
	return c.thinkingStyle
}
