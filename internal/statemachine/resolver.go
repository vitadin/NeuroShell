package statemachine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/log"
)

// CommandResolver handles command resolution through the priority system.
type CommandResolver struct {
	stdlibLoader *embedded.StdlibLoader
	logger       *log.Logger
}

// NewCommandResolver creates a new command resolver.
func NewCommandResolver() *CommandResolver {
	return &CommandResolver{
		stdlibLoader: embedded.NewStdlibLoader(),
		logger:       logger.NewStyledLogger("CommandResolver"),
	}
}

// ResolveCommand attempts to resolve a command name to a builtin command or script.
// Priority: builtin → stdlib → user scripts
func (r *CommandResolver) ResolveCommand(commandName string) (*neurotypes.StateMachineResolvedCommand, error) {
	// Priority 1: Try builtin commands (highest priority)
	if builtinCmd, exists := commands.GetGlobalRegistry().Get(commandName); exists {
		return &neurotypes.StateMachineResolvedCommand{
			Name:           commandName,
			Type:           neurotypes.CommandTypeBuiltin,
			BuiltinCommand: builtinCmd,
		}, nil
	}

	// Priority 2: Try stdlib scripts (medium priority)
	if r.stdlibLoader.ScriptExists(commandName) {
		scriptContent, err := r.stdlibLoader.LoadScript(commandName)
		if err != nil {
			r.logger.Error("Failed to load stdlib script", "command", commandName, "error", err)
		} else {
			return &neurotypes.StateMachineResolvedCommand{
				Name:          commandName,
				Type:          neurotypes.CommandTypeStdlib,
				ScriptContent: scriptContent,
				ScriptPath:    r.stdlibLoader.GetScriptPath(commandName),
			}, nil
		}
	}

	// Priority 3: Try user scripts (lowest priority)
	if strings.HasSuffix(commandName, ".neuro") {
		r.logger.Debug("Detected file path command", "command", commandName)
		return r.resolveUserFilePath(commandName)
	}

	return nil, fmt.Errorf("unknown command: %s", commandName)
}

// resolveUserFilePath resolves a file path to a user script command.
func (r *CommandResolver) resolveUserFilePath(filePath string) (*neurotypes.StateMachineResolvedCommand, error) {
	r.logger.Debug("Resolving user file path", "path", filePath)

	// Path resolution with security checks
	resolvedPath, err := r.resolvePathSafely(filePath)
	if err != nil {
		r.logger.Error("Path resolution failed", "path", filePath, "error", err)
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	r.logger.Debug("Path resolved", "original", filePath, "resolved", resolvedPath)

	// Load and validate file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		r.logger.Error("File read failed", "path", resolvedPath, "error", err)
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	r.logger.Debug("File loaded successfully", "path", resolvedPath, "content_length", len(content))

	return &neurotypes.StateMachineResolvedCommand{
		Name:          filePath,
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: string(content),
		ScriptPath:    resolvedPath,
	}, nil
}

// resolvePathSafely resolves file paths with security checks.
func (r *CommandResolver) resolvePathSafely(filePath string) (string, error) {
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
