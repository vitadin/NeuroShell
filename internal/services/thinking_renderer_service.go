package services

import (
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// ThinkingRendererService provides simple rendering capabilities for thinking blocks.
// This service focuses on clean, uniform display of thinking content across all providers.
type ThinkingRendererService struct {
	initialized bool
}

// NewThinkingRendererService creates a new ThinkingRendererService instance.
func NewThinkingRendererService() *ThinkingRendererService {
	return &ThinkingRendererService{
		initialized: false,
	}
}

// Name returns the service name "thinking_renderer" for registration.
func (t *ThinkingRendererService) Name() string {
	return "thinking_renderer"
}

// Initialize sets up the ThinkingRendererService.
func (t *ThinkingRendererService) Initialize() error {
	logger.ServiceOperation("thinking_renderer", "initialize", "starting")
	t.initialized = true
	logger.ServiceOperation("thinking_renderer", "initialize", "completed")
	return nil
}

// RenderThinkingBlocks renders thinking blocks with simple, uniform formatting.
// This provides a clean separation between thinking content and regular text.
func (t *ThinkingRendererService) RenderThinkingBlocks(blocks []neurotypes.ThinkingBlock) string {
	if !t.initialized {
		logger.Error("ThinkingRendererService not initialized")
		return ""
	}

	if len(blocks) == 0 {
		return ""
	}

	var result strings.Builder

	for _, block := range blocks {
		// Simple, consistent formatting for all providers
		result.WriteString("\n🤔 **Thinking:**\n")
		result.WriteString(block.Content)
		result.WriteString("\n\n")

		logger.Debug("Thinking block rendered", "provider", block.Provider, "type", block.Type, "content_length", len(block.Content))
	}

	return result.String()
}

// IsInitialized returns true if the service has been initialized.
func (t *ThinkingRendererService) IsInitialized() bool {
	return t.initialized
}
