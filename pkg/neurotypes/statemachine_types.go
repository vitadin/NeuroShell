// Package neurotypes defines core data types and interfaces for the NeuroShell state machine.
// This package contains fundamental types used throughout the state machine execution pipeline.
package neurotypes

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
	// StateTryResolving - Handling try command setup and target extraction
	StateTryResolving
	// StateExecuting - Running builtin commands through the command registry
	StateExecuting
	// StateTryExecuting - Executing try target command with error capture
	StateTryExecuting
	// StateScriptLoaded - Script content loaded and ready for line-by-line processing
	StateScriptLoaded
	// StateScriptExecuting - Processing script lines through recursive state machine calls
	StateScriptExecuting
	// StateTryCompleted - Try command finished, set success/error variables
	StateTryCompleted
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
	case StateTryResolving:
		return "TryResolving"
	case StateExecuting:
		return "Executing"
	case StateTryExecuting:
		return "TryExecuting"
	case StateScriptLoaded:
		return "ScriptLoaded"
	case StateScriptExecuting:
		return "ScriptExecuting"
	case StateTryCompleted:
		return "TryCompleted"
	case StateCompleted:
		return "Completed"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// StateMachineResolvedCommand represents a command that has been resolved through the priority system.
// This is specific to the state machine's needs and extends the base ResolvedCommand concept.
type StateMachineResolvedCommand struct {
	// Name of the command
	Name string
	// Type indicates whether this is a builtin, stdlib, or user command
	Type CommandType
	// BuiltinCommand is populated for builtin commands
	BuiltinCommand Command
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
	// Parsed command structure (using interface{} to avoid import cycle)
	ParsedCommand interface{}
	// Resolved command information
	ResolvedCommand *StateMachineResolvedCommand
	// Script execution state
	ScriptLines []string
	CurrentLine int
	// Recursion tracking
	RecursionDepth int
	// Error state
	Error error
}

// StateMachineConfig holds configuration options for the state machine.
type StateMachineConfig struct {
	// EchoCommands controls whether to output command lines with %%> prefix
	EchoCommands bool
	// MacroExpansion enables command-level macro expansion
	MacroExpansion bool
	// RecursionLimit sets maximum recursion depth for nested script calls
	RecursionLimit int
}

// DefaultStateMachineConfig returns sensible default configuration for the state machine.
func DefaultStateMachineConfig() StateMachineConfig {
	return StateMachineConfig{
		EchoCommands:   false,
		MacroExpansion: true,
		RecursionLimit: 50,
	}
}
