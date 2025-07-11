package statemachine

import (
	"fmt"

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
	// This will be implemented when we extend the context

	return nil, fmt.Errorf("command not found: %s", commandName)
}
