// Package commands provides command registration and execution functionality for NeuroShell.
// It manages a global registry of commands and handles command lookup and execution.
package commands

import (
	"fmt"
	"sync"

	"neuroshell/pkg/neurotypes"
)

// Registry manages command registration and lookup for NeuroShell commands.
// It provides thread-safe registration and retrieval of commands by name.
type Registry struct {
	mu       sync.RWMutex
	commands map[string]neurotypes.Command
}

// NewRegistry creates a new command registry with an empty command map.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]neurotypes.Command),
	}
}

// Register adds a command to the registry. Returns an error if the command
// name is empty or if a command with the same name is already registered.
func (r *Registry) Register(cmd neurotypes.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cmd.Name() == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if _, exists := r.commands[cmd.Name()]; exists {
		return fmt.Errorf("command %s already registered", cmd.Name())
	}

	r.commands[cmd.Name()] = cmd
	return nil
}

// Unregister removes a command from the registry by name.
// This operation is thread-safe and will not error if the command doesn't exist.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.commands, name)
}

// Get retrieves a command by name. Returns the command and true if found,
// or nil and false if the command is not registered.
func (r *Registry) Get(name string) (neurotypes.Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetAll returns a slice containing all registered commands.
// The returned slice is a copy and can be safely modified.
func (r *Registry) GetAll() []neurotypes.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]neurotypes.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// Execute runs a command by name with the provided arguments, input, and context.
// Returns an error if the command is not found or if the command execution fails.
func (r *Registry) Execute(name string, args map[string]string, input string, ctx neurotypes.Context) error {
	cmd, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("unknown command: %s", name)
	}
	return cmd.Execute(args, input, ctx)
}

// GetParseMode returns the parse mode for a command by name.
// Returns ParseModeKeyValue as default if the command is not found.
func (r *Registry) GetParseMode(name string) neurotypes.ParseMode {
	cmd, exists := r.Get(name)
	if !exists {
		return neurotypes.ParseModeKeyValue // Default
	}
	return cmd.ParseMode()
}

// IsValidCommand checks if a command exists in the registry.
// Returns true if the command is registered, false otherwise.
func (r *Registry) IsValidCommand(name string) bool {
	_, exists := r.Get(name)
	return exists
}

// GlobalRegistry is the global command registry instance used throughout NeuroShell.
// Commands register themselves with this instance during initialization.
var GlobalRegistry = NewRegistry()
