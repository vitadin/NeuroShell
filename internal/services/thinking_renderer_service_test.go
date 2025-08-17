package services

import (
	"strings"
	"testing"

	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/lipgloss"
)

// MockRenderConfig provides a test implementation of RenderConfig interface.
type MockRenderConfig struct {
	showThinking  bool
	thinkingStyle string
	compactMode   bool
	maxWidth      int
	theme         string
	styles        map[string]lipgloss.Style
}

// NewMockRenderConfig creates a new MockRenderConfig with default values.
func NewMockRenderConfig() *MockRenderConfig {
	return &MockRenderConfig{
		showThinking:  true,
		thinkingStyle: "full",
		compactMode:   false,
		maxWidth:      120,
		theme:         "default",
		styles: map[string]lipgloss.Style{
			"info":       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
			"italic":     lipgloss.NewStyle().Italic(true),
			"background": lipgloss.NewStyle(),
			"warning":    lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
			"highlight":  lipgloss.NewStyle().Foreground(lipgloss.Color("159")),
			"bold":       lipgloss.NewStyle().Bold(true),
			"underline":  lipgloss.NewStyle().Underline(true),
		},
	}
}

func (m *MockRenderConfig) GetStyle(element string) lipgloss.Style {
	if style, exists := m.styles[element]; exists {
		return style
	}
	return lipgloss.NewStyle()
}

func (m *MockRenderConfig) GetTheme() string {
	return m.theme
}

func (m *MockRenderConfig) IsCompactMode() bool {
	return m.compactMode
}

func (m *MockRenderConfig) GetMaxWidth() int {
	return m.maxWidth
}

func (m *MockRenderConfig) ShowThinking() bool {
	return m.showThinking
}

func (m *MockRenderConfig) GetThinkingStyle() string {
	return m.thinkingStyle
}

// TestThinkingRendererService_Name tests the service name method.
func TestThinkingRendererService_Name(t *testing.T) {
	service := NewThinkingRendererService()
	expected := "thinking-renderer"
	actual := service.Name()

	if actual != expected {
		t.Errorf("Expected service name %q, got %q", expected, actual)
	}
}

// TestThinkingRendererService_Initialize tests service initialization.
func TestThinkingRendererService_Initialize(t *testing.T) {
	service := NewThinkingRendererService()

	// Initially not initialized
	if service.IsInitialized() {
		t.Error("Service should not be initialized initially")
	}

	// Initialize service
	err := service.Initialize()
	if err != nil {
		t.Errorf("Initialize() returned error: %v", err)
	}

	// Should be initialized now
	if !service.IsInitialized() {
		t.Error("Service should be initialized after Initialize()")
	}
}

// TestThinkingRendererService_RenderThinkingBlocks_EmptyBlocks tests rendering with no blocks.
func TestThinkingRendererService_RenderThinkingBlocks_EmptyBlocks(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()
	config := NewMockRenderConfig()

	result := service.RenderThinkingBlocks(nil, config)
	if result != "" {
		t.Errorf("Expected empty result for nil blocks, got %q", result)
	}

	result = service.RenderThinkingBlocks([]neurotypes.ThinkingBlock{}, config)
	if result != "" {
		t.Errorf("Expected empty result for empty blocks, got %q", result)
	}
}

// TestThinkingRendererService_RenderThinkingBlocks_ShowThinkingDisabled tests rendering with thinking disabled.
func TestThinkingRendererService_RenderThinkingBlocks_ShowThinkingDisabled(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()
	config := NewMockRenderConfig()
	config.showThinking = false

	blocks := []neurotypes.ThinkingBlock{
		{Provider: "anthropic", Type: "thinking", Content: "Test thinking content"},
	}

	result := service.RenderThinkingBlocks(blocks, config)
	if result != "" {
		t.Errorf("Expected empty result when thinking disabled, got %q", result)
	}
}

// TestThinkingRendererService_RenderThinkingBlocks_NotInitialized tests rendering without initialization.
func TestThinkingRendererService_RenderThinkingBlocks_NotInitialized(t *testing.T) {
	service := NewThinkingRendererService()
	config := NewMockRenderConfig()

	blocks := []neurotypes.ThinkingBlock{
		{Provider: "anthropic", Type: "thinking", Content: "Test content"},
	}

	result := service.RenderThinkingBlocks(blocks, config)
	if result != "" {
		t.Errorf("Expected empty result for uninitialized service, got %q", result)
	}
}

// TestThinkingRendererService_RenderSingleBlock_ProviderSpecific tests provider-specific rendering.
func TestThinkingRendererService_RenderSingleBlock_ProviderSpecific(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()
	config := NewMockRenderConfig()

	testCases := []struct {
		provider      string
		expectedLabel string
		description   string
	}{
		{"anthropic", "Claude's internal reasoning:", "Anthropic provider"},
		{"Anthropic", "Claude's internal reasoning:", "Anthropic provider (case insensitive)"},
		{"gemini", "Gemini's thought process:", "Gemini provider"},
		{"Gemini", "Gemini's thought process:", "Gemini provider (case insensitive)"},
		{"openai", "OpenAI reasoning trace:", "OpenAI provider"},
		{"OpenAI", "OpenAI reasoning trace:", "OpenAI provider (case insensitive)"},
		{"unknown", "Internal processing:", "Unknown provider fallback"},
		{"", "Internal processing:", "Empty provider fallback"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			block := neurotypes.ThinkingBlock{
				Provider: tc.provider,
				Type:     "thinking",
				Content:  "Test content",
			}

			result := service.RenderSingleBlock(block, config)
			if !strings.Contains(result, tc.expectedLabel) {
				t.Errorf("Expected result to contain %q for provider %q, got: %s", tc.expectedLabel, tc.provider, result)
			}
		})
	}
}

// TestThinkingRendererService_GetSupportedProviders tests supported providers list.
func TestThinkingRendererService_GetSupportedProviders(t *testing.T) {
	service := NewThinkingRendererService()

	expected := []string{"anthropic", "gemini", "openai", "generic"}
	actual := service.GetSupportedProviders()

	if len(actual) != len(expected) {
		t.Errorf("Expected %d supported providers, got %d", len(expected), len(actual))
	}

	for _, expectedProvider := range expected {
		found := false
		for _, actualProvider := range actual {
			if actualProvider == expectedProvider {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %q not found in supported providers: %v", expectedProvider, actual)
		}
	}
}

// Note: RenderThinkingBlocksLegacy test has been removed.
// Use TestThinkingRendererService_RenderThinkingBlocks_MultipleBlocks for testing.

// TestThinkingRendererService_RenderSingleBlock_CompactMode tests compact mode rendering.
func TestThinkingRendererService_RenderSingleBlock_CompactMode(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()

	config := NewMockRenderConfig()
	config.compactMode = true

	block := neurotypes.ThinkingBlock{
		Provider: "anthropic",
		Type:     "thinking",
		Content:  "Compact mode test content",
	}

	result := service.RenderSingleBlock(block, config)

	// Should contain the content
	if !strings.Contains(result, "Compact mode test content") {
		t.Error("Compact mode rendering should contain the thinking content")
	}

	// Test that it's different from normal mode
	config.compactMode = false
	normalResult := service.RenderSingleBlock(block, config)

	if result == normalResult {
		t.Error("Compact mode rendering should be different from normal mode")
	}
}

// TestThinkingRendererService_RenderSingleBlock_ThinkingStyles tests different thinking styles.
func TestThinkingRendererService_RenderSingleBlock_ThinkingStyles(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()

	longContent := strings.Repeat("This is a long thinking process that should be summarized in summary mode. ", 10)

	block := neurotypes.ThinkingBlock{
		Provider: "anthropic",
		Type:     "thinking",
		Content:  longContent,
	}

	// Test full style (default)
	config := NewMockRenderConfig()
	config.thinkingStyle = "full"
	fullResult := service.RenderSingleBlock(block, config)

	// Test summary style
	config.thinkingStyle = "summary"
	summaryResult := service.RenderSingleBlock(block, config)

	// Summary should be shorter than full
	if len(summaryResult) >= len(fullResult) {
		t.Error("Summary rendering should be shorter than full rendering for long content")
	}

	// Both should contain the provider label
	if !strings.Contains(fullResult, "Claude's internal reasoning:") {
		t.Error("Full rendering should contain provider label")
	}
	if !strings.Contains(summaryResult, "Claude's internal reasoning:") {
		t.Error("Summary rendering should contain provider label")
	}
}

// TestThinkingRendererService_createContentSummary tests content summarization.
func TestThinkingRendererService_createContentSummary(t *testing.T) {
	service := NewThinkingRendererService()

	// Test short content (should remain unchanged)
	shortContent := "Short content"
	result := service.createContentSummary(shortContent)
	if result != shortContent {
		t.Errorf("Short content should remain unchanged, got %q", result)
	}

	// Test long content (should be summarized)
	longContent := strings.Repeat("This is a very long thinking process that goes on and on with lots of details. ", 10)
	result = service.createContentSummary(longContent)

	if len(result) >= len(longContent) {
		t.Error("Long content should be summarized to shorter length")
	}

	if !strings.HasSuffix(result, "...") {
		t.Error("Summarized content should end with '...'")
	}
}

// TestThinkingRendererService_wrapContent tests content wrapping.
func TestThinkingRendererService_wrapContent(t *testing.T) {
	service := NewThinkingRendererService()

	// Test with zero width (should return original)
	content := "This is a test content"
	result := service.wrapContent(content, 0)
	if result != content {
		t.Error("Zero width should return original content")
	}

	// Test with negative width (should return original)
	result = service.wrapContent(content, -1)
	if result != content {
		t.Error("Negative width should return original content")
	}

	// Test wrapping
	longLine := "This is a very long line that should be wrapped at word boundaries when the width limit is reached"
	result = service.wrapContent(longLine, 20)

	lines := strings.Split(result, "\n")
	if len(lines) <= 1 {
		t.Error("Long content should be wrapped into multiple lines")
	}

	// Check that no line exceeds the width limit (with some tolerance for word boundaries)
	for i, line := range lines {
		if len(line) > 25 { // Some tolerance for word boundaries
			t.Errorf("Line %d exceeds reasonable width limit: %q (length: %d)", i, line, len(line))
		}
	}
}

// TestThinkingRendererService_RenderSingleBlock_MaxWidth tests width constraints.
func TestThinkingRendererService_RenderSingleBlock_MaxWidth(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()

	config := NewMockRenderConfig()
	config.maxWidth = 50 // Small width to force wrapping

	longContent := "This is a very long thinking process content that should definitely be wrapped when rendered with a narrow width constraint applied to the display"

	block := neurotypes.ThinkingBlock{
		Provider: "anthropic",
		Type:     "thinking",
		Content:  longContent,
	}

	result := service.RenderSingleBlock(block, config)

	// Should contain the content
	if !strings.Contains(result, "This is a very long thinking") {
		t.Error("Rendered result should contain the thinking content")
	}

	// Test with unlimited width
	config.maxWidth = 0
	unlimitedResult := service.RenderSingleBlock(block, config)

	// Results should be different (one wrapped, one not)
	if result == unlimitedResult {
		t.Error("Width-constrained rendering should differ from unlimited width")
	}
}

// TestThinkingRendererService_RenderSingleBlock_DifferentThemes tests theme handling.
func TestThinkingRendererService_RenderSingleBlock_DifferentThemes(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()

	block := neurotypes.ThinkingBlock{
		Provider: "anthropic",
		Type:     "thinking",
		Content:  "Theme test content",
	}

	// Test with default theme
	config := NewMockRenderConfig()
	config.theme = "default"
	defaultResult := service.RenderSingleBlock(block, config)

	// Test with plain theme
	config.theme = "plain"
	plainResult := service.RenderSingleBlock(block, config)

	// Both should contain the content and provider label
	if !strings.Contains(defaultResult, "Theme test content") {
		t.Error("Default theme rendering should contain content")
	}
	if !strings.Contains(plainResult, "Theme test content") {
		t.Error("Plain theme rendering should contain content")
	}
	if !strings.Contains(defaultResult, "Claude's internal reasoning:") {
		t.Error("Default theme rendering should contain provider label")
	}
	if !strings.Contains(plainResult, "Claude's internal reasoning:") {
		t.Error("Plain theme rendering should contain provider label")
	}
}

// TestThinkingRendererService_RenderThinkingBlocks_MultipleBlocks tests rendering multiple blocks.
func TestThinkingRendererService_RenderThinkingBlocks_MultipleBlocks(t *testing.T) {
	service := NewThinkingRendererService()
	_ = service.Initialize()
	config := NewMockRenderConfig()

	blocks := []neurotypes.ThinkingBlock{
		{Provider: "anthropic", Type: "thinking", Content: "First thinking block"},
		{Provider: "gemini", Type: "thinking", Content: "Second thinking block"},
		{Provider: "openai", Type: "reasoning", Content: "Third reasoning block"},
	}

	result := service.RenderThinkingBlocks(blocks, config)

	// Should contain all provider labels
	if !strings.Contains(result, "Claude's internal reasoning:") {
		t.Error("Result should contain Anthropic provider label")
	}
	if !strings.Contains(result, "Gemini's thought process:") {
		t.Error("Result should contain Gemini provider label")
	}
	if !strings.Contains(result, "OpenAI reasoning trace:") {
		t.Error("Result should contain OpenAI provider label")
	}

	// Should contain all content
	if !strings.Contains(result, "First thinking block") {
		t.Error("Result should contain first block content")
	}
	if !strings.Contains(result, "Second thinking block") {
		t.Error("Result should contain second block content")
	}
	if !strings.Contains(result, "Third reasoning block") {
		t.Error("Result should contain third block content")
	}
}

// TestThinkingRendererService_DefaultRenderConfig tests the default render config.
func TestThinkingRendererService_DefaultRenderConfig(t *testing.T) {
	config := &DefaultRenderConfig{
		ShowThinkingEnabled: true,
		ThinkingStyleValue:  "full",
		CompactModeEnabled:  false,
		MaxWidthValue:       80,
		ThemeValue:          "default",
	}

	// Test interface methods
	if !config.ShowThinking() {
		t.Error("Default config should show thinking")
	}
	if config.GetThinkingStyle() != "full" {
		t.Error("Default thinking style should be 'full'")
	}
	if config.IsCompactMode() {
		t.Error("Default config should not be in compact mode")
	}
	if config.GetMaxWidth() != 80 {
		t.Error("Default max width should be 80")
	}
	if config.GetTheme() != "default" {
		t.Error("Default theme should be 'default'")
	}

	// Test style methods - verify they return styles without comparison
	infoStyle := config.GetStyle("info")
	infoRendered := infoStyle.Render("test")
	if infoRendered == "" {
		t.Error("Info style should be able to render text")
	}

	italicStyle := config.GetStyle("italic")
	italicRendered := italicStyle.Render("test")
	if italicRendered == "" {
		t.Error("Italic style should be able to render text")
	}

	// Test that unknown elements return a usable style
	unknownStyle := config.GetStyle("unknown")
	unknownRendered := unknownStyle.Render("test")
	if unknownRendered != "test" {
		t.Error("Unknown style should return plain rendering")
	}
}
