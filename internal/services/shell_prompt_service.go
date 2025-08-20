package services

import (
	"fmt"
	"strconv"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// ShellPromptService manages shell prompt configuration and template retrieval.
// It follows the three-layer architecture by being stateless and interacting only with Context.
// This service retrieves prompt templates but does NOT perform variable interpolation.
// Interpolation is handled at the shell layer using the state machine.
type ShellPromptService struct {
	initialized bool
}

// NewShellPromptService creates a new ShellPromptService instance.
func NewShellPromptService() *ShellPromptService {
	return &ShellPromptService{
		initialized: false,
	}
}

// Name returns the service name "shell_prompt" for registration.
func (s *ShellPromptService) Name() string {
	return "shell_prompt"
}

// Initialize sets up the service with default prompt configuration if needed.
func (s *ShellPromptService) Initialize() error {
	if s.initialized {
		return nil
	}

	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Set default prompt configuration if not already set
	// This ensures a fallback configuration exists
	if linesCount, _ := ctx.GetVariable("_prompt_lines_count"); linesCount == "" {
		// Set single-line prompt as default to maintain backward compatibility
		if err := ctx.SetSystemVariable("_prompt_lines_count", "1"); err != nil {
			return fmt.Errorf("failed to set default prompt lines count: %w", err)
		}
		if err := ctx.SetSystemVariable("_prompt_line1", "neuro> "); err != nil {
			return fmt.Errorf("failed to set default prompt line: %w", err)
		}
	}

	s.initialized = true
	return nil
}

// GetServiceInfo returns service information for debugging and monitoring.
func (s *ShellPromptService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        s.Name(),
		"initialized": s.initialized,
		"type":        "shell_prompt",
		"description": "Shell prompt configuration management",
	}
}

// GetPromptLines retrieves raw prompt templates from context without interpolation.
// Returns a slice of template strings where each string represents one line of the prompt.
// Variable interpolation should be performed by the caller using the state machine.
func (s *ShellPromptService) GetPromptLines() ([]string, error) {
	if !s.initialized {
		return nil, fmt.Errorf("shell prompt service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Get number of prompt lines to display
	countStr, _ := ctx.GetVariable("_prompt_lines_count")
	if countStr == "" {
		// Fallback to single line if not configured
		countStr = "1"
	}

	count := 1 // default
	if n, err := strconv.Atoi(countStr); err == nil && n >= 1 && n <= 5 {
		count = n
	}

	// Retrieve raw templates (NO interpolation happens here)
	lines := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		varName := fmt.Sprintf("_prompt_line%d", i)
		line, _ := ctx.GetVariable(varName)
		if line == "" {
			if i == 1 {
				// Always provide a fallback for the first line
				line = "neuro> "
			} else {
				// Skip empty lines beyond the first
				continue
			}
		}
		lines = append(lines, line)
	}

	// Ensure we always have at least one line
	if len(lines) == 0 {
		lines = append(lines, "neuro> ")
	}

	return lines, nil
}

// GetPromptLinesCount returns the configured number of prompt lines.
func (s *ShellPromptService) GetPromptLinesCount() (int, error) {
	if !s.initialized {
		return 0, fmt.Errorf("shell prompt service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)
	countStr, _ := ctx.GetVariable("_prompt_lines_count")
	if countStr == "" {
		return 1, nil // default
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 5 {
		return 1, nil // default
	}

	return count, nil
}

// ValidatePromptConfiguration checks if the current prompt configuration is valid.
func (s *ShellPromptService) ValidatePromptConfiguration() error {
	if !s.initialized {
		return fmt.Errorf("shell prompt service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext().(*neuroshellcontext.NeuroContext)

	// Check lines count
	countStr, _ := ctx.GetVariable("_prompt_lines_count")
	if countStr == "" {
		return fmt.Errorf("prompt lines count not configured")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fmt.Errorf("invalid prompt lines count format: %s", countStr)
	}

	if count < 1 || count > 5 {
		return fmt.Errorf("prompt lines count must be between 1 and 5, got: %d", count)
	}

	// Check that at least the first line is configured
	firstLine, _ := ctx.GetVariable("_prompt_line1")
	if firstLine == "" {
		return fmt.Errorf("first prompt line must be configured")
	}

	return nil
}

// Interface compliance check
var _ neurotypes.Service = (*ShellPromptService)(nil)
