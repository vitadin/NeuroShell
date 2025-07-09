// Package neurotypes defines command resolution interfaces for NeuroShell.
// This file contains types and interfaces for priority-based command resolution,
// enabling commands to be sourced from builtin Go code, embedded stdlib scripts,
// or user-defined scripts.
package neurotypes

// CommandType defines the source type of a resolved command.
type CommandType int

const (
	// CommandTypeBuiltin represents commands implemented in Go code.
	// These have the highest priority and are always resolved first.
	CommandTypeBuiltin CommandType = iota

	// CommandTypeStdlib represents commands from embedded stdlib scripts.
	// These have medium priority and are resolved after builtin commands.
	CommandTypeStdlib

	// CommandTypeUser represents commands from user-defined scripts.
	// These have the lowest priority and are resolved last.
	CommandTypeUser
)

// String returns a human-readable string representation of the CommandType.
func (ct CommandType) String() string {
	switch ct {
	case CommandTypeBuiltin:
		return "builtin"
	case CommandTypeStdlib:
		return "stdlib"
	case CommandTypeUser:
		return "user"
	default:
		return "unknown"
	}
}

// CommandInfo holds metadata about a command for help display and autocomplete.
type CommandInfo struct {
	Name        string
	Description string
	Usage       string
	ParseMode   ParseMode
	CommandType CommandType
}

// ResolvedCommand represents a command that has been resolved through the
// priority-based command resolution system. It contains metadata about
// the command's source and type.
type ResolvedCommand struct {
	// Name is the command name as it would be invoked by the user
	Name string

	// Type indicates whether this is a builtin, stdlib, or user command
	Type CommandType

	// Source provides information about where the command comes from:
	// - "builtin" for Go-implemented commands
	// - Filesystem path for stdlib or user scripts
	Source string

	// Command is the actual command implementation that can be executed
	Command Command
}

// CommandResolver defines the interface for priority-based command resolution.
// It provides methods to resolve commands by name with fallback through
// different command sources (builtin → stdlib → user).
type CommandResolver interface {
	// ResolveCommand resolves a command by name using priority-based lookup.
	// It searches through builtin commands first, then stdlib scripts,
	// then user scripts, returning the first match found.
	ResolveCommand(name string) (*ResolvedCommand, error)

	// HasBuiltinCommand checks if a command exists in the builtin registry.
	HasBuiltinCommand(name string) bool

	// HasStdlibCommand checks if a command exists in the embedded stdlib.
	HasStdlibCommand(name string) bool

	// HasUserCommand checks if a command exists in the user script directory.
	HasUserCommand(name string) bool

	// ListCommands returns all available commands grouped by type.
	// The returned map keys are command names, values are command types.
	ListCommands() map[string]CommandType

	// GetCommandInfo returns detailed information about a resolved command.
	// This includes help information, source location, and command type.
	GetCommandInfo(name string) (*ResolvedCommand, error)
}

// ScriptLoader defines the interface for loading command scripts from
// various sources (embedded filesystem, user directories, etc.).
type ScriptLoader interface {
	// LoadScript loads the content of a script by filename.
	// Returns the script content as a string or an error if not found.
	LoadScript(filename string) (string, error)

	// ListAvailableScripts returns a list of all available script names
	// (without the .neuro extension) that can be loaded.
	ListAvailableScripts() ([]string, error)

	// ScriptExists checks if a script with the given name exists.
	ScriptExists(name string) bool

	// GetScriptPath returns the full path or identifier for a script.
	// For embedded scripts, this might be a virtual path.
	GetScriptPath(name string) string
}

// ScriptParameterProvider defines the interface for setting up and cleaning up
// script parameters when executing script-based commands.
type ScriptParameterProvider interface {
	// SetupParameters configures script variables based on command arguments.
	// This maps command arguments to script variables like ${_0}, ${_1}, etc.
	SetupParameters(commandName string, args map[string]string, input string) error

	// CleanupParameters removes script parameters after execution.
	// This ensures no parameter variables leak between script executions.
	CleanupParameters() error

	// GetParameterMapping returns the current parameter variable mapping.
	// Useful for debugging and testing parameter setup.
	GetParameterMapping() map[string]string
}
