package statemachine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/pkg/neurotypes"
)

// resolveCommand attempts to resolve a command name to a builtin command or script.
func (sm *StateMachine) resolveCommand(commandName string) (*neurotypes.StateMachineResolvedCommand, error) {
	// Priority 1: Try builtin commands (highest priority)
	if builtinCmd, exists := commands.GetGlobalRegistry().Get(commandName); exists {
		return &neurotypes.StateMachineResolvedCommand{
			Name:           commandName,
			Type:           neurotypes.CommandTypeBuiltin,
			BuiltinCommand: builtinCmd,
		}, nil
	}

	// Priority 2: Try stdlib scripts (medium priority)
	if sm.stdlibLoader.ScriptExists(commandName) {
		scriptContent, err := sm.stdlibLoader.LoadScript(commandName)
		if err != nil {
			sm.logger.Error("Failed to load stdlib script", "command", commandName, "error", err)
		} else {
			return &neurotypes.StateMachineResolvedCommand{
				Name:          commandName,
				Type:          neurotypes.CommandTypeStdlib,
				ScriptContent: scriptContent,
				ScriptPath:    sm.stdlibLoader.GetScriptPath(commandName),
			}, nil
		}
	}

	// Priority 3: Try user scripts (lowest priority)
	// Check if this is a file path (contains .neuro suffix)
	if strings.HasSuffix(commandName, ".neuro") {
		sm.logger.Debug("Detected file path command", "command", commandName)
		return sm.resolveUserFilePath(commandName)
	}
	// TODO: Add name-based user script resolution (~/.neuro/scripts/, etc.)

	return nil, fmt.Errorf("unknown command: %s", commandName)
}

// resolveUserFilePath resolves a file path to a user script command.
// It handles both relative and absolute paths with security checks.
func (sm *StateMachine) resolveUserFilePath(filePath string) (*neurotypes.StateMachineResolvedCommand, error) {
	sm.logger.Debug("Resolving user file path", "path", filePath)

	// Path resolution with security checks
	resolvedPath, err := sm.resolvePathSafely(filePath)
	if err != nil {
		sm.logger.Error("Path resolution failed", "path", filePath, "error", err)
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	sm.logger.Debug("Path resolved", "original", filePath, "resolved", resolvedPath)

	// Load and validate file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		sm.logger.Error("File read failed", "path", resolvedPath, "error", err)
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	sm.logger.Debug("File loaded successfully", "path", resolvedPath, "content_length", len(content))

	return &neurotypes.StateMachineResolvedCommand{
		Name:          filePath,                   // Original user input
		Type:          neurotypes.CommandTypeUser, // Existing type
		ScriptContent: string(content),            // Reuse existing field
		ScriptPath:    resolvedPath,               // Reuse existing field
	}, nil
}

// resolvePathSafely resolves file paths with security checks to prevent directory traversal
// and handles both relative and absolute paths appropriately.
func (sm *StateMachine) resolvePathSafely(filePath string) (string, error) {
	// Security: Prevent directory traversal attacks
	if strings.Contains(filePath, "..") {
		return "", fmt.Errorf("parent directory access not allowed")
	}

	// Handle absolute paths as-is
	if filepath.IsAbs(filePath) {
		if _, err := os.Stat(filePath); err != nil {
			return "", fmt.Errorf("file not found: %s", filePath)
		}
		return filePath, nil
	}

	// Resolve relative paths against current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine current directory: %w", err)
	}

	resolved := filepath.Join(cwd, filePath)
	if _, err := os.Stat(resolved); err != nil {
		return "", fmt.Errorf("file not found: %s", resolved)
	}

	return resolved, nil
}
