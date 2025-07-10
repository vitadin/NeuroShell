// Package builtin provides built-in NeuroShell commands that are available by default.
// This file selectively registers builtin commands during state machine transition.
package builtin

import (
	"fmt"
	"neuroshell/internal/commands"
)

// TEMPORARY STATE MACHINE TRANSITION ISOLATION:
// During the transition to state machine execution, we are only registering
// the echo command to isolate risk. Other commands remain in the codebase
// but are not registered, allowing gradual migration one command at a time.
//
// To re-enable other commands:
// 1. Add their registration calls in the init() function below
// 2. Uncomment their test lines in justfile
// 3. Test each command thoroughly through the state machine

func init() {
	// Register only the echo command during state machine transition
	if err := commands.GlobalRegistry.Register(&EchoCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register echo command: %v", err))
	}

	// TODO: Gradually re-enable other commands after echo is proven stable:
	// commands.GlobalRegistry.Register(&SetCommand{})
	// commands.GlobalRegistry.Register(&GetCommand{})
	// commands.GlobalRegistry.Register(&BashCommand{})
	// commands.GlobalRegistry.Register(&SendCommand{})
	// commands.GlobalRegistry.Register(&SendStreamCommand{})
	// commands.GlobalRegistry.Register(&SendSyncCommand{})
	// commands.GlobalRegistry.Register(&HelpCommand{})
	// commands.GlobalRegistry.Register(&ExitCommand{})
	// commands.GlobalRegistry.Register(&VarsCommand{})
	// commands.GlobalRegistry.Register(&TryCommand{})
	// commands.GlobalRegistry.Register(&CheckCommand{})
	// commands.GlobalRegistry.Register(&RunCommand{})
	// commands.GlobalRegistry.Register(&EditorCommand{})
}
