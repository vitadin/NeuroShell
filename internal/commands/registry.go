package commands

import (
	"fmt"
	"sync"

	"neuroshell/pkg/types"
)

// Registry manages command registration and lookup
type Registry struct {
	mu       sync.RWMutex
	commands map[string]types.Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]types.Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd types.Command) error {
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

// Unregister removes a command from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.commands, name)
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (types.Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetAll returns all registered commands
func (r *Registry) GetAll() []types.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]types.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// Execute runs a command by name
func (r *Registry) Execute(name string, args map[string]string, input string, ctx types.Context) error {
	cmd, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("unknown command: %s", name)
	}
	return cmd.Execute(args, input, ctx)
}

// GetParseMode returns the parse mode for a command
func (r *Registry) GetParseMode(name string) types.ParseMode {
	cmd, exists := r.Get(name)
	if !exists {
		return types.ParseModeKeyValue // Default
	}
	return cmd.ParseMode()
}

// IsValidCommand checks if a command exists
func (r *Registry) IsValidCommand(name string) bool {
	_, exists := r.Get(name)
	return exists
}

// Global registry instance
var GlobalRegistry = NewRegistry()