// Package commands provides enhanced command registration and resolution functionality.
// This file implements priority-based command resolution supporting builtin, stdlib,
// and user-defined commands with proper precedence handling.
package commands

import (
	"fmt"
	"sync"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// EnhancedCommandRegistry implements priority-based command resolution.
// It maintains separate registries for different command types and resolves
// commands using the priority order: builtin → stdlib → user.
type EnhancedCommandRegistry struct {
	mu sync.RWMutex

	// builtinRegistry contains Go-implemented commands (highest priority)
	builtinRegistry *Registry

	// stdlibCommands contains commands loaded from embedded stdlib scripts
	stdlibCommands map[string]neurotypes.Command

	// userCommands contains commands loaded from user script directories
	userCommands map[string]neurotypes.Command

	// scriptLoaders for different script sources
	stdlibLoader neurotypes.ScriptLoader
	userLoader   neurotypes.ScriptLoader
}

// NewEnhancedCommandRegistry creates a new enhanced registry with the existing
// builtin registry as the base. This ensures backward compatibility while
// adding enhanced resolution capabilities.
func NewEnhancedCommandRegistry(builtinRegistry *Registry) *EnhancedCommandRegistry {
	return &EnhancedCommandRegistry{
		builtinRegistry: builtinRegistry,
		stdlibCommands:  make(map[string]neurotypes.Command),
		userCommands:    make(map[string]neurotypes.Command),
	}
}

// SetStdlibLoader configures the script loader for embedded stdlib scripts.
func (r *EnhancedCommandRegistry) SetStdlibLoader(loader neurotypes.ScriptLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stdlibLoader = loader
}

// SetUserLoader configures the script loader for user-defined scripts.
func (r *EnhancedCommandRegistry) SetUserLoader(loader neurotypes.ScriptLoader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.userLoader = loader
}

// RegisterStdlibCommand registers a command loaded from the stdlib.
// This is typically called during startup when discovering embedded scripts.
func (r *EnhancedCommandRegistry) RegisterStdlibCommand(cmd neurotypes.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd.Name() == "" {
		return fmt.Errorf("stdlib command name cannot be empty")
	}

	r.stdlibCommands[cmd.Name()] = cmd

	// Register the command with info in global context for autocomplete and help
	if globalCtx := context.GetGlobalContext(); globalCtx != nil {
		if neuroCtx, ok := globalCtx.(*context.NeuroContext); ok {
			neuroCtx.RegisterCommandWithInfoAndType(cmd, neurotypes.CommandTypeStdlib)
		}
	}

	return nil
}

// RegisterUserCommand registers a command loaded from user scripts.
func (r *EnhancedCommandRegistry) RegisterUserCommand(cmd neurotypes.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd.Name() == "" {
		return fmt.Errorf("user command name cannot be empty")
	}

	r.userCommands[cmd.Name()] = cmd

	// Register the command with info in global context for autocomplete and help
	if globalCtx := context.GetGlobalContext(); globalCtx != nil {
		if neuroCtx, ok := globalCtx.(*context.NeuroContext); ok {
			neuroCtx.RegisterCommandWithInfoAndType(cmd, neurotypes.CommandTypeUser)
		}
	}

	return nil
}

// ResolveCommand implements the CommandResolver interface.
// It searches for commands using the priority order: builtin → stdlib → user.
func (r *EnhancedCommandRegistry) ResolveCommand(name string) (*neurotypes.ResolvedCommand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Priority 1: Check builtin commands first
	if cmd, exists := r.builtinRegistry.Get(name); exists {
		return &neurotypes.ResolvedCommand{
			Name:    name,
			Type:    neurotypes.CommandTypeBuiltin,
			Source:  "builtin",
			Command: cmd,
		}, nil
	}

	// Priority 2: Check stdlib commands
	if cmd, exists := r.stdlibCommands[name]; exists {
		source := "stdlib"
		if r.stdlibLoader != nil {
			source = r.stdlibLoader.GetScriptPath(name)
		}
		return &neurotypes.ResolvedCommand{
			Name:    name,
			Type:    neurotypes.CommandTypeStdlib,
			Source:  source,
			Command: cmd,
		}, nil
	}

	// Priority 3: Check user commands (lowest priority)
	if cmd, exists := r.userCommands[name]; exists {
		source := "user"
		if r.userLoader != nil {
			source = r.userLoader.GetScriptPath(name)
		}
		return &neurotypes.ResolvedCommand{
			Name:    name,
			Type:    neurotypes.CommandTypeUser,
			Source:  source,
			Command: cmd,
		}, nil
	}

	return nil, fmt.Errorf("command not found: %s", name)
}

// HasBuiltinCommand implements the CommandResolver interface.
func (r *EnhancedCommandRegistry) HasBuiltinCommand(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.builtinRegistry.Get(name)
	return exists
}

// HasStdlibCommand implements the CommandResolver interface.
func (r *EnhancedCommandRegistry) HasStdlibCommand(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.stdlibCommands[name]
	return exists
}

// HasUserCommand implements the CommandResolver interface.
func (r *EnhancedCommandRegistry) HasUserCommand(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.userCommands[name]
	return exists
}

// ListCommands implements the CommandResolver interface.
// Returns all available commands grouped by type.
func (r *EnhancedCommandRegistry) ListCommands() map[string]neurotypes.CommandType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make(map[string]neurotypes.CommandType)

	// Add builtin commands
	for _, cmd := range r.builtinRegistry.GetAll() {
		commands[cmd.Name()] = neurotypes.CommandTypeBuiltin
	}

	// Add stdlib commands
	for name := range r.stdlibCommands {
		commands[name] = neurotypes.CommandTypeStdlib
	}

	// Add user commands
	for name := range r.userCommands {
		commands[name] = neurotypes.CommandTypeUser
	}

	return commands
}

// GetCommandInfo implements the CommandResolver interface.
// Returns detailed information about a resolved command.
func (r *EnhancedCommandRegistry) GetCommandInfo(name string) (*neurotypes.ResolvedCommand, error) {
	// Delegate to ResolveCommand which already provides all the information we need
	return r.ResolveCommand(name)
}

// Execute provides a compatibility method that resolves and executes a command.
// This maintains compatibility with the existing Execute interface while
// using the enhanced resolution system.
func (r *EnhancedCommandRegistry) Execute(name string, args map[string]string, input string) error {
	resolved, err := r.ResolveCommand(name)
	if err != nil {
		return err
	}

	return resolved.Command.Execute(args, input)
}

// GetParseMode provides a compatibility method for getting command parse modes.
// This maintains compatibility with the existing registry interface.
func (r *EnhancedCommandRegistry) GetParseMode(name string) neurotypes.ParseMode {
	resolved, err := r.ResolveCommand(name)
	if err != nil {
		return neurotypes.ParseModeKeyValue // Default fallback
	}

	return resolved.Command.ParseMode()
}

// IsValidCommand provides a compatibility method for command validation.
// This maintains compatibility with the existing registry interface.
func (r *EnhancedCommandRegistry) IsValidCommand(name string) bool {
	_, err := r.ResolveCommand(name)
	return err == nil
}

// RefreshStdlibCommands reloads all stdlib commands from the script loader.
// This is useful for development or when the embedded scripts change.
func (r *EnhancedCommandRegistry) RefreshStdlibCommands() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stdlibLoader == nil {
		return fmt.Errorf("no stdlib loader configured")
	}

	// Clear existing stdlib commands
	r.stdlibCommands = make(map[string]neurotypes.Command)

	// Reload from script loader
	scripts, err := r.stdlibLoader.ListAvailableScripts()
	if err != nil {
		return fmt.Errorf("failed to list stdlib scripts: %w", err)
	}

	for _, scriptName := range scripts {
		content, err := r.stdlibLoader.LoadScript(scriptName + ".neuro")
		if err != nil {
			// Log error but continue with other scripts
			continue
		}

		// Create script command
		cmd := NewScriptCommand(scriptName, content, neurotypes.CommandTypeStdlib)
		r.stdlibCommands[scriptName] = cmd
	}

	return nil
}

// RefreshUserCommands reloads all user commands from the script loader.
func (r *EnhancedCommandRegistry) RefreshUserCommands() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.userLoader == nil {
		return fmt.Errorf("no user loader configured")
	}

	// Clear existing user commands
	r.userCommands = make(map[string]neurotypes.Command)

	// Reload from script loader
	scripts, err := r.userLoader.ListAvailableScripts()
	if err != nil {
		return fmt.Errorf("failed to list user scripts: %w", err)
	}

	for _, scriptName := range scripts {
		content, err := r.userLoader.LoadScript(scriptName + ".neuro")
		if err != nil {
			// Log error but continue with other scripts
			continue
		}

		// Create script command
		cmd := NewScriptCommand(scriptName, content, neurotypes.CommandTypeUser)
		r.userCommands[scriptName] = cmd
	}

	return nil
}
