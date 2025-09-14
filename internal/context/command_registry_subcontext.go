package context

import (
	"sync"

	"neuroshell/pkg/neurotypes"
)

// CommandRegistrySubcontext defines the interface for command registry management functionality.
// This includes command registration, help information storage, and autocomplete support.
type CommandRegistrySubcontext interface {
	// Command registration
	RegisterCommand(commandName string)
	RegisterCommandWithInfo(cmd neurotypes.Command)
	RegisterCommandWithInfoAndType(cmd neurotypes.Command, cmdType neurotypes.CommandType)
	UnregisterCommand(commandName string)

	// Command lookup
	IsCommandRegistered(commandName string) bool
	GetRegisteredCommands() []string

	// Help information management
	GetCommandHelpInfo(commandName string) (*neurotypes.HelpInfo, bool)
	GetAllCommandHelpInfo() map[string]*neurotypes.HelpInfo
}

// commandRegistrySubcontext implements the CommandRegistrySubcontext interface.
type commandRegistrySubcontext struct {
	// Command registry information
	registeredCommands map[string]bool                 // Track registered command names for autocomplete
	commandHelpInfo    map[string]*neurotypes.HelpInfo // Store detailed help info for autocomplete and help system
	commandMutex       sync.RWMutex                    // Protects registeredCommands and commandHelpInfo maps
}

// NewCommandRegistrySubcontext creates a new CommandRegistrySubcontext instance.
func NewCommandRegistrySubcontext() CommandRegistrySubcontext {
	return &commandRegistrySubcontext{
		registeredCommands: make(map[string]bool),
		commandHelpInfo:    make(map[string]*neurotypes.HelpInfo),
	}
}

// NewCommandRegistrySubcontextFromContext creates a CommandRegistrySubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's command registry subcontext.
func NewCommandRegistrySubcontextFromContext(ctx *NeuroContext) CommandRegistrySubcontext {
	return ctx.commandRegistryCtx
}

// RegisterCommand registers a command name for autocomplete functionality.
func (c *commandRegistrySubcontext) RegisterCommand(commandName string) {
	c.commandMutex.Lock()
	defer c.commandMutex.Unlock()
	c.registeredCommands[commandName] = true
}

// RegisterCommandWithInfo registers a command with its metadata for help and autocomplete.
func (c *commandRegistrySubcontext) RegisterCommandWithInfo(cmd neurotypes.Command) {
	c.commandMutex.Lock()
	defer c.commandMutex.Unlock()

	commandName := cmd.Name()
	c.registeredCommands[commandName] = true

	// Store command help information
	helpInfo := cmd.HelpInfo()
	c.commandHelpInfo[commandName] = &helpInfo
}

// RegisterCommandWithInfoAndType registers a command with its metadata and type.
func (c *commandRegistrySubcontext) RegisterCommandWithInfoAndType(cmd neurotypes.Command, _ neurotypes.CommandType) {
	c.commandMutex.Lock()
	defer c.commandMutex.Unlock()

	commandName := cmd.Name()
	c.registeredCommands[commandName] = true

	// Store command help information
	helpInfo := cmd.HelpInfo()
	c.commandHelpInfo[commandName] = &helpInfo
}

// UnregisterCommand removes a command name from the autocomplete registry.
func (c *commandRegistrySubcontext) UnregisterCommand(commandName string) {
	c.commandMutex.Lock()
	defer c.commandMutex.Unlock()
	delete(c.registeredCommands, commandName)
	delete(c.commandHelpInfo, commandName)
}

// GetRegisteredCommands returns a list of all registered command names.
func (c *commandRegistrySubcontext) GetRegisteredCommands() []string {
	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()

	commands := make([]string, 0, len(c.registeredCommands))
	for commandName := range c.registeredCommands {
		commands = append(commands, commandName)
	}
	return commands
}

// IsCommandRegistered checks if a command name is registered.
func (c *commandRegistrySubcontext) IsCommandRegistered(commandName string) bool {
	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()
	return c.registeredCommands[commandName]
}

// GetCommandHelpInfo returns the help information for a specific command.
func (c *commandRegistrySubcontext) GetCommandHelpInfo(commandName string) (*neurotypes.HelpInfo, bool) {
	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()
	info, exists := c.commandHelpInfo[commandName]
	return info, exists
}

// GetAllCommandHelpInfo returns all registered command help information.
func (c *commandRegistrySubcontext) GetAllCommandHelpInfo() map[string]*neurotypes.HelpInfo {
	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()

	result := make(map[string]*neurotypes.HelpInfo)
	for name, info := range c.commandHelpInfo {
		result[name] = info
	}
	return result
}
