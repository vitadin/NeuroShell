// Package execution provides the core state machine infrastructure for NeuroShell command execution.
// It implements a unified execution pipeline that handles builtin commands, stdlib scripts,
// and user scripts through a well-defined state machine with integrated interpolation.
package execution

import (
	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

// State represents the current state of command execution in the state machine.
type State int

const (
	// StateReceived - Initial state: command line received and ready for processing
	StateReceived State = iota
	// StateInterpolating - Expanding variables and macros in the command line
	StateInterpolating
	// StateParsing - Parsing command structure and extracting components
	StateParsing
	// StateResolving - Finding command in builtin registry, stdlib, or user scripts
	StateResolving
	// StateExecuting - Running builtin commands through the command registry
	StateExecuting
	// StateScriptLoaded - Script content loaded and ready for line-by-line processing
	StateScriptLoaded
	// StateScriptExecuting - Processing script lines through recursive state machine calls
	StateScriptExecuting
	// StateCompleted - Execution finished successfully
	StateCompleted
	// StateError - Execution failed with an error
	StateError
)

// String returns a human-readable representation of the execution state.
func (s State) String() string {
	switch s {
	case StateReceived:
		return "Received"
	case StateInterpolating:
		return "Interpolating"
	case StateParsing:
		return "Parsing"
	case StateResolving:
		return "Resolving"
	case StateExecuting:
		return "Executing"
	case StateScriptLoaded:
		return "ScriptLoaded"
	case StateScriptExecuting:
		return "ScriptExecuting"
	case StateCompleted:
		return "Completed"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// CommandType represents the type of command being executed.
type CommandType int

const (
	// CommandTypeBuiltin - Go-implemented command in the registry (highest priority)
	CommandTypeBuiltin CommandType = iota
	// CommandTypeStdlib - Embedded .neuro script in the binary (medium priority)
	CommandTypeStdlib
	// CommandTypeUser - User-defined .neuro script (lowest priority)
	CommandTypeUser
)

// String returns a human-readable representation of the command type.
func (ct CommandType) String() string {
	switch ct {
	case CommandTypeBuiltin:
		return "Builtin"
	case CommandTypeStdlib:
		return "Stdlib"
	case CommandTypeUser:
		return "User"
	default:
		return "Unknown"
	}
}

// ResolvedCommand represents a command that has been resolved through the priority system.
type ResolvedCommand struct {
	// Name of the command
	Name string
	// Type indicates whether this is a builtin, stdlib, or user command
	Type CommandType
	// BuiltinCommand is populated for builtin commands
	BuiltinCommand neurotypes.Command
	// ScriptContent is populated for stdlib and user scripts
	ScriptContent string
	// ScriptPath is the path to the script file (for user scripts)
	ScriptPath string
}

// StateSnapshot captures the complete execution state for recursive calls.
// This allows the state machine to handle script-to-script calls by saving and restoring
// the execution context when entering and exiting nested script executions.
type StateSnapshot struct {
	// Current execution state
	State State
	// Input being processed
	Input string
	// Parsed command structure
	ParsedCommand *parser.Command
	// Resolved command information
	ResolvedCommand *ResolvedCommand
	// Script execution state
	ScriptLines []string
	CurrentLine int
	// Recursion tracking
	RecursionDepth int
	// Error state
	Error error
}

// Config holds configuration options for the state machine.
type Config struct {
	// EchoCommands controls whether to output command lines with %%> prefix
	EchoCommands bool
	// MacroExpansion enables command-level macro expansion
	MacroExpansion bool
	// RecursionLimit sets maximum recursion depth for nested script calls
	RecursionLimit int
}

// DefaultConfig returns sensible default configuration for the state machine.
func DefaultConfig() Config {
	return Config{
		EchoCommands:   false,
		MacroExpansion: true,
		RecursionLimit: 50,
	}
}
