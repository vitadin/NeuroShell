package statemachine

import (
	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

// State management methods - Phase 1 implementation using internal fields
func (sm *StateMachine) getCurrentState() neurotypes.State {
	return sm.currentState
}

func (sm *StateMachine) setState(state neurotypes.State) {
	sm.currentState = state
}

func (sm *StateMachine) getExecutionInput() string {
	return sm.executionInput
}

func (sm *StateMachine) setExecutionInput(input string) {
	sm.executionInput = input
}

func (sm *StateMachine) setExecutionError(err error) {
	sm.executionError = err
}

func (sm *StateMachine) getExecutionError() error {
	return sm.executionError
}

func (sm *StateMachine) resetRecursionDepth() {
	sm.recursionDepth = 0
}

func (sm *StateMachine) clearExecutionData() {
	sm.parsedCommand = nil
	sm.resolvedCommand = nil
	sm.scriptLines = nil
	sm.currentScriptLine = 0
}

func (sm *StateMachine) getResolvedBuiltinCommand() interface{} {
	if sm.resolvedCommand != nil && sm.resolvedCommand.Type == neurotypes.CommandTypeBuiltin {
		return sm.resolvedCommand.BuiltinCommand
	}
	return nil
}

func (sm *StateMachine) getScriptContent() string {
	if sm.resolvedCommand != nil && (sm.resolvedCommand.Type == neurotypes.CommandTypeStdlib || sm.resolvedCommand.Type == neurotypes.CommandTypeUser) {
		return sm.resolvedCommand.ScriptContent
	}
	return ""
}

func (sm *StateMachine) hasMoreScriptLines() bool {
	return sm.currentScriptLine < len(sm.scriptLines)
}

// saveExecutionState captures the current state for recursive calls.
func (sm *StateMachine) saveExecutionState() neurotypes.StateSnapshot {
	snapshot := neurotypes.StateSnapshot{
		State:           sm.currentState,
		Input:           sm.executionInput,
		ParsedCommand:   sm.parsedCommand,
		ResolvedCommand: sm.resolvedCommand,
		ScriptLines:     sm.scriptLines,
		CurrentLine:     sm.currentScriptLine,
		RecursionDepth:  sm.recursionDepth,
		Error:           sm.executionError,
	}
	sm.stateStack = append(sm.stateStack, snapshot)
	return snapshot
}

// restoreExecutionState restores a previously saved execution state.
func (sm *StateMachine) restoreExecutionState(snapshot neurotypes.StateSnapshot) {
	sm.currentState = snapshot.State
	sm.executionInput = snapshot.Input
	// Convert interface{} back to *parser.Command
	if cmd, ok := snapshot.ParsedCommand.(*parser.Command); ok {
		sm.parsedCommand = cmd
	}
	sm.resolvedCommand = snapshot.ResolvedCommand
	sm.scriptLines = snapshot.ScriptLines
	sm.currentScriptLine = snapshot.CurrentLine
	sm.recursionDepth = snapshot.RecursionDepth
	sm.executionError = snapshot.Error

	// Pop from stack
	if len(sm.stateStack) > 0 {
		sm.stateStack = sm.stateStack[:len(sm.stateStack)-1]
	}
}

// Execution state accessor methods - Phase 1 implementation using internal fields

func (sm *StateMachine) getRecursionDepth() int {
	return sm.recursionDepth
}

func (sm *StateMachine) incrementRecursionDepth() {
	sm.recursionDepth++
}

func (sm *StateMachine) setParsedCommand(cmd *parser.Command) {
	sm.parsedCommand = cmd
}

func (sm *StateMachine) getParsedCommand() *parser.Command {
	return sm.parsedCommand
}

func (sm *StateMachine) setResolvedCommand(resolved *neurotypes.StateMachineResolvedCommand) {
	sm.resolvedCommand = resolved
}

func (sm *StateMachine) getResolvedCommand() *neurotypes.StateMachineResolvedCommand {
	return sm.resolvedCommand
}

func (sm *StateMachine) setScriptLines(lines []string) {
	sm.scriptLines = lines
}

func (sm *StateMachine) getScriptLines() []string {
	return sm.scriptLines
}

func (sm *StateMachine) setCurrentScriptLine(line int) {
	sm.currentScriptLine = line
}

func (sm *StateMachine) getCurrentScriptLine() int {
	return sm.currentScriptLine
}
