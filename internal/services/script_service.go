package services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// ScriptService handles loading and parsing of .neuro script files.
type ScriptService struct {
	initialized bool
}

// NewScriptService creates a new ScriptService instance.
func NewScriptService() *ScriptService {
	return &ScriptService{
		initialized: false,
	}
}

// Name returns the service name "script" for registration.
func (s *ScriptService) Name() string {
	return "script"
}

// Initialize sets up the ScriptService for operation.
func (s *ScriptService) Initialize(_ neurotypes.Context) error {
	s.initialized = true
	return nil
}

// LoadScript reads a script file and queues commands in the context
func (s *ScriptService) LoadScript(filepath string, ctx neurotypes.Context) error {
	if !s.initialized {
		return fmt.Errorf("script service not initialized")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open script file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Cast to NeuroContext to access queue methods
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	// Store script metadata
	neuroCtx.SetScriptMetadata("current_file", filepath)
	neuroCtx.SetScriptMetadata("line_count", 0)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Queue the command for execution
		neuroCtx.QueueCommand(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading script file: %w", err)
	}

	// Update metadata
	neuroCtx.SetScriptMetadata("line_count", lineNum)
	neuroCtx.SetScriptMetadata("commands_queued", neuroCtx.GetQueueSize())

	return nil
}

// GetScriptMetadata returns script execution information from context
func (s *ScriptService) GetScriptMetadata(ctx neurotypes.Context) (map[string]interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("script service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	metadata := make(map[string]interface{})

	if file, exists := neuroCtx.GetScriptMetadata("current_file"); exists {
		metadata["current_file"] = file
	}
	if lineCount, exists := neuroCtx.GetScriptMetadata("line_count"); exists {
		metadata["line_count"] = lineCount
	}
	if commandsQueued, exists := neuroCtx.GetScriptMetadata("commands_queued"); exists {
		metadata["commands_queued"] = commandsQueued
	}

	metadata["queue_size"] = neuroCtx.GetQueueSize()

	return metadata, nil
}
