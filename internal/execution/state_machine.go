package execution

import (
	"bufio"
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
)

// StateMachine implements the core state machine for NeuroShell command execution.
// It provides a unified execution pipeline that handles builtin commands, stdlib scripts,
// and user scripts through well-defined state transitions with integrated interpolation.
type StateMachine struct {
	// Direct reference to the context for state management
	context *context.NeuroContext
	// Integrated interpolation engine (not a separate service)
	interpolator *CoreInterpolator
	// Stdlib script loader for embedded scripts
	stdlibLoader *embedded.StdlibLoader
	// Configuration options
	config Config
	// State stack for recursive execution
	stateStack []StateSnapshot

	// Internal execution state (Phase 1 implementation)
	// These will be moved to context in Phase 2
	currentState      State
	executionInput    string
	executionError    error
	recursionDepth    int
	parsedCommand     *parser.Command
	resolvedCommand   *ResolvedCommand
	scriptLines       []string
	currentScriptLine int
	tryMode           bool // Track when we're in try mode
	tryCompleted      bool // Track when a try command completed successfully
}

// NewStateMachine creates a new state machine with the given context and configuration.
func NewStateMachine(ctx *context.NeuroContext, config Config) *StateMachine {
	return &StateMachine{
		context:      ctx,
		interpolator: NewCoreInterpolator(ctx),
		stdlibLoader: embedded.NewStdlibLoader(),
		config:       config,
		stateStack:   make([]StateSnapshot, 0),
	}
}

// NewStateMachineWithDefaults creates a new state machine with default configuration.
func NewStateMachineWithDefaults(ctx *context.NeuroContext) *StateMachine {
	return NewStateMachine(ctx, DefaultConfig())
}

// Execute is the main entry point for the state machine execution.
// It processes the input through the complete state machine pipeline until completion or error.
func (sm *StateMachine) Execute(input string) error {
	// Initialize execution state in context
	sm.initializeExecution(input)

	logger.Debug("State machine execution started", "input", input)

	// Main state machine loop
	for {
		currentState := sm.getCurrentState()
		logger.Debug("State machine processing", "state", currentState.String())

		// Check for terminal states
		if currentState == StateCompleted || currentState == StateError {
			break
		}

		// Process current state
		if err := sm.ProcessCurrentState(); err != nil {
			sm.setExecutionError(err)
			// In try mode, errors go to StateTryError instead of StateError
			if sm.tryMode {
				sm.setState(StateTryError)
				// Continue to next iteration to process StateTryError
				continue
			}
			sm.setState(StateError)
			break
		}

		// Determine and transition to next state
		nextState := sm.DetermineNextState()
		sm.setState(nextState)

		// Safety check to prevent infinite loops
		if nextState == currentState {
			logger.Error("State machine stuck in infinite loop", "state", currentState.String())
			return fmt.Errorf("state machine stuck in state: %s", currentState.String())
		}
	}

	// Handle successful completion of try command
	if sm.getCurrentState() == StateCompleted && sm.tryCompleted {
		sm.tryCompleted = false
		// Set success variables for try command
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
	}

	// Return any execution error
	// In try mode, StateTryError should return nil (success)
	if sm.getCurrentState() == StateTryError {
		return nil
	}
	return sm.getExecutionError()
}

// ProcessCurrentState handles the logic for the current execution state.
func (sm *StateMachine) ProcessCurrentState() error {
	currentState := sm.getCurrentState()
	switch currentState {
	case StateReceived:
		return sm.processReceived()
	case StateInterpolating:
		return sm.processInterpolating()
	case StateParsing:
		return sm.processParsing()
	case StateResolving:
		return sm.processResolving()
	case StateExecuting:
		return sm.processExecuting()
	case StateScriptLoaded:
		return sm.processScriptLoaded()
	case StateScriptExecuting:
		return sm.processScriptExecuting()
	case StateTryError:
		return sm.processTryError()
	default:
		return fmt.Errorf("unknown state: %s", currentState.String())
	}
}

// DetermineNextState determines the next state based on the current state and execution context.
func (sm *StateMachine) DetermineNextState() State {
	currentState := sm.getCurrentState()
	switch currentState {
	case StateReceived:
		input := sm.getExecutionInput()
		if sm.interpolator.HasVariables(input) {
			return StateInterpolating
		}
		return StateParsing
	case StateInterpolating:
		return StateReceived // Recursive re-entry with expanded input
	case StateParsing:
		return StateResolving
	case StateResolving:
		// Handle try command recursive re-entry (like interpolation)
		if sm.tryMode {
			// Set flag to track that we're executing a try command
			sm.tryCompleted = true
			sm.tryMode = false
			return StateReceived // Recursive re-entry with try target
		}

		// Determine if builtin command or script
		if sm.getResolvedBuiltinCommand() != nil {
			return StateExecuting
		}
		if sm.getScriptContent() != "" {
			return StateScriptLoaded
		}
		return StateError
	case StateExecuting:
		return StateCompleted
	case StateScriptLoaded:
		return StateScriptExecuting
	case StateScriptExecuting:
		// Check if more script lines to process
		if sm.hasMoreScriptLines() {
			return StateReceived // Recursive: next script line re-enters
		}
		return StateCompleted
	case StateTryError:
		return StateCompleted
	default:
		return StateError
	}
}

// initializeExecution sets up the initial state for execution.
func (sm *StateMachine) initializeExecution(input string) {
	// Reset execution state
	sm.setState(StateReceived)
	sm.setExecutionInput(input)
	sm.setExecutionError(nil)
	sm.resetRecursionDepth()
	sm.clearExecutionData()
}

// State management methods - Phase 1 implementation using internal fields
func (sm *StateMachine) getCurrentState() State {
	return sm.currentState
}

func (sm *StateMachine) setState(state State) {
	sm.currentState = state
	logger.Debug("State transition", "new_state", state.String())
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
	if sm.resolvedCommand != nil && sm.resolvedCommand.Type == CommandTypeBuiltin {
		return sm.resolvedCommand.BuiltinCommand
	}
	return nil
}

func (sm *StateMachine) getScriptContent() string {
	if sm.resolvedCommand != nil && (sm.resolvedCommand.Type == CommandTypeStdlib || sm.resolvedCommand.Type == CommandTypeUser) {
		return sm.resolvedCommand.ScriptContent
	}
	return ""
}

func (sm *StateMachine) hasMoreScriptLines() bool {
	return sm.currentScriptLine < len(sm.scriptLines)
}

// saveExecutionState captures the current state for recursive calls.
func (sm *StateMachine) saveExecutionState() StateSnapshot {
	snapshot := StateSnapshot{
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
func (sm *StateMachine) restoreExecutionState(snapshot StateSnapshot) {
	sm.currentState = snapshot.State
	sm.executionInput = snapshot.Input
	sm.parsedCommand = snapshot.ParsedCommand
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

// parseScriptIntoLines parses script content into executable lines, handling multiline commands.
func (sm *StateMachine) parseScriptIntoLines(scriptContent string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(scriptContent))

	for scanner.Scan() {
		line := scanner.Text()

		// Handle multiline continuation with ...
		if strings.HasSuffix(strings.TrimSpace(line), "...") {
			// Accumulate multiline command
			var multilineBuilder []string
			multilineBuilder = append(multilineBuilder, line)

			// Continue reading lines until we find one that doesn't end with ...
			for scanner.Scan() {
				nextLine := scanner.Text()
				multilineBuilder = append(multilineBuilder, nextLine)

				if !strings.HasSuffix(strings.TrimSpace(nextLine), "...") {
					break
				}
			}

			// Join and clean up multiline command
			multilineCommand := strings.Join(multilineBuilder, "\n")
			multilineCommand = strings.ReplaceAll(multilineCommand, "...\n", " ")
			multilineCommand = strings.ReplaceAll(multilineCommand, "...", " ")
			lines = append(lines, strings.TrimSpace(multilineCommand))
		} else {
			lines = append(lines, line)
		}
	}

	return lines
}

// setupScriptParameters sets up parameters for script execution.
func (sm *StateMachine) setupScriptParameters(args map[string]string, input string, commandName string) error {
	// Use the state machine's context directly for parameter setup
	ctx := sm.context

	// Set standard script parameters
	_ = ctx.SetSystemVariable("_0", commandName) // Command name
	_ = ctx.SetSystemVariable("_1", input)       // Input parameter
	_ = ctx.SetSystemVariable("_*", input)       // All positional args

	// Set named arguments as variables (use regular SetVariable for user-defined names)
	var namedArgs []string
	for key, value := range args {
		_ = ctx.SetVariable(key, value)
		namedArgs = append(namedArgs, key+"="+value)
	}
	_ = ctx.SetSystemVariable("_@", strings.Join(namedArgs, " "))

	return nil
}

// cleanupScriptParameters removes script parameters after execution.
func (sm *StateMachine) cleanupScriptParameters() error {
	// Use the state machine's context directly for parameter cleanup
	ctx := sm.context

	// Clear standard script parameters
	parameterNames := []string{"_0", "_1", "_*", "_@"}
	for _, name := range parameterNames {
		_ = ctx.SetSystemVariable(name, "")
	}

	return nil
}

// resolveCommand attempts to resolve a command name to a builtin command or script.
func (sm *StateMachine) resolveCommand(commandName string) (*ResolvedCommand, error) {
	// Priority 1: Try builtin commands (highest priority)
	if builtinCmd, exists := commands.GetGlobalRegistry().Get(commandName); exists {
		return &ResolvedCommand{
			Name:           commandName,
			Type:           CommandTypeBuiltin,
			BuiltinCommand: builtinCmd,
		}, nil
	}

	// Priority 2: Try stdlib scripts (medium priority)
	if sm.stdlibLoader.ScriptExists(commandName) {
		scriptContent, err := sm.stdlibLoader.LoadScript(commandName)
		if err != nil {
			logger.Error("Failed to load stdlib script", "command", commandName, "error", err)
		} else {
			return &ResolvedCommand{
				Name:          commandName,
				Type:          CommandTypeStdlib,
				ScriptContent: scriptContent,
				ScriptPath:    sm.stdlibLoader.GetScriptPath(commandName),
			}, nil
		}
	}

	// Priority 3: Try user scripts (lowest priority)
	// This will be implemented when we extend the context

	return nil, fmt.Errorf("command not found: %s", commandName)
}

// State processor methods - each handles the logic for a specific execution state

// processReceived handles the initial state when a command is received.
func (sm *StateMachine) processReceived() error {
	input := sm.getExecutionInput()
	logger.Debug("State: Received", "input", input)

	// Log the input for debugging purposes
	if input == "" {
		return fmt.Errorf("empty input received")
	}

	// This state primarily serves as entry point and logging
	// The actual processing decision is made in DetermineNextState
	return nil
}

// processInterpolating handles variable and macro expansion.
func (sm *StateMachine) processInterpolating() error {
	input := sm.getExecutionInput()
	logger.Debug("State: Interpolating", "input", input)

	expanded, hasVariables, err := sm.interpolator.InterpolateCommandLine(input)
	if err != nil {
		return fmt.Errorf("interpolation failed: %w", err)
	}

	if hasVariables {
		// Check recursion limit
		recursionDepth := sm.getRecursionDepth()
		if recursionDepth >= sm.config.RecursionLimit {
			return fmt.Errorf("recursion limit exceeded (%d)", sm.config.RecursionLimit)
		}

		// Increment recursion depth and set up for recursive re-entry
		sm.incrementRecursionDepth()
		sm.setExecutionInput(expanded)
		logger.Debug("Macro expansion - recursive re-entry", "original", input, "expanded", expanded, "depth", recursionDepth+1)
	}

	return nil
}

// processParsing handles command structure parsing.
func (sm *StateMachine) processParsing() error {
	input := sm.getExecutionInput()
	logger.Debug("State: Parsing", "input", input)

	cmd := parser.ParseInput(input)
	if cmd == nil {
		return fmt.Errorf("failed to parse command: %s", input)
	}

	sm.setParsedCommand(cmd)
	logger.Debug("Command parsed successfully", "name", cmd.Name, "message", cmd.Message)
	return nil
}

// processResolving handles command resolution through the priority system.
func (sm *StateMachine) processResolving() error {
	parsedCmd := sm.getParsedCommand()
	if parsedCmd == nil {
		return fmt.Errorf("no parsed command available")
	}

	logger.Debug("State: Resolving", "command", parsedCmd.Name)

	// Handle try command - similar to interpolation pattern
	if parsedCmd.Name == "try" {
		targetCommand := strings.TrimSpace(parsedCmd.Message)
		if targetCommand == "" {
			// Empty try command - set success variables and complete
			_ = sm.context.SetSystemVariable("_status", "0")
			_ = sm.context.SetSystemVariable("_error", "")
			logger.Debug("Empty try command - setting success variables")
			return nil
		}

		// Set try mode and new input (like interpolation does)
		sm.tryMode = true
		sm.setExecutionInput(targetCommand)

		logger.Debug("Try command detected - recursing", "target", targetCommand)
		return nil // DetermineNextState will handle transition back to StateReceived
	}

	// Priority-based resolution: builtin → stdlib → user
	resolved, err := sm.resolveCommand(parsedCmd.Name)
	if err != nil {
		return fmt.Errorf("command not found: %s", parsedCmd.Name)
	}

	sm.setResolvedCommand(resolved)
	logger.Debug("Command resolved successfully", "command", parsedCmd.Name, "type", resolved.Type.String())
	return nil
}

// processExecuting handles execution of builtin commands.
func (sm *StateMachine) processExecuting() error {
	resolved := sm.getResolvedCommand()
	parsedCmd := sm.getParsedCommand()
	input := sm.getExecutionInput()

	if resolved == nil || resolved.Type != CommandTypeBuiltin {
		return fmt.Errorf("no builtin command to execute")
	}

	logger.Debug("State: Executing", "command", parsedCmd.Name, "type", "builtin")

	// Output command line with %%> prefix if echo_commands is enabled
	if sm.config.EchoCommands {
		fmt.Printf("%%%%> %s\n", input)
	}

	// Execute the builtin command
	err := resolved.BuiltinCommand.Execute(parsedCmd.Options, parsedCmd.Message)
	if err != nil {
		return fmt.Errorf("builtin command execution failed: %w", err)
	}

	logger.Debug("Builtin command executed successfully", "command", parsedCmd.Name)
	return nil
}

// processTryError handles errors in try mode by capturing them as variables.
func (sm *StateMachine) processTryError() error {
	// Set error variables from the execution error
	err := sm.getExecutionError()
	logger.Debug("processTryError called", "error", err, "errorIsNil", err == nil)

	if err != nil {
		_ = sm.context.SetSystemVariable("_status", "1")
		_ = sm.context.SetSystemVariable("_error", err.Error())
		logger.Debug("Set error variables", "status", "1", "error", err.Error())
	} else {
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
		logger.Debug("Set success variables", "status", "0", "error", "")
	}

	// Reset try mode completely when handling error
	sm.tryMode = false

	logger.Debug("Try error captured as variables", "error", err)
	return nil
}

// processScriptLoaded handles setup after a script has been loaded.
func (sm *StateMachine) processScriptLoaded() error {
	resolved := sm.getResolvedCommand()
	parsedCmd := sm.getParsedCommand()

	if resolved == nil || (resolved.Type != CommandTypeStdlib && resolved.Type != CommandTypeUser) {
		return fmt.Errorf("no script to load")
	}

	logger.Debug("State: ScriptLoaded", "command", parsedCmd.Name, "type", resolved.Type.String())

	// Setup script parameters
	err := sm.setupScriptParameters(parsedCmd.Options, parsedCmd.Message, parsedCmd.Name)
	if err != nil {
		return fmt.Errorf("failed to setup script parameters: %w", err)
	}

	// Parse script content into executable lines
	lines := sm.parseScriptIntoLines(resolved.ScriptContent)
	sm.setScriptLines(lines)
	sm.setCurrentScriptLine(0)

	logger.Debug("Script loaded and parsed", "command", parsedCmd.Name, "lines", len(lines))
	return nil
}

// processScriptExecuting handles line-by-line script execution.
func (sm *StateMachine) processScriptExecuting() error {
	lines := sm.getScriptLines()
	currentLineIndex := sm.getCurrentScriptLine()
	parsedCmd := sm.getParsedCommand()

	logger.Debug("State: ScriptExecuting", "script", parsedCmd.Name, "line", currentLineIndex+1, "total", len(lines))

	if currentLineIndex >= len(lines) {
		// Script finished - cleanup parameters
		err := sm.cleanupScriptParameters()
		if err != nil {
			logger.Error("Failed to cleanup script parameters", "error", err)
		}
		logger.Debug("Script execution completed", "script", parsedCmd.Name)
		return nil // Will transition to StateCompleted
	}

	// Get current line to execute
	line := lines[currentLineIndex]
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "%%") {
		sm.setCurrentScriptLine(currentLineIndex + 1)
		return nil // Stay in StateScriptExecuting for next line
	}

	// Output command line with %%> prefix if echo_commands is enabled
	if sm.config.EchoCommands {
		fmt.Printf("%%%%> %s\n", line)
	}

	// Save current execution state for recursive call
	savedState := sm.saveExecutionState()

	// Execute script line through state machine recursively
	// This line will go through: StateReceived → StateInterpolating → ... → StateCompleted
	err := sm.Execute(line)

	// Restore execution state after recursive call
	sm.restoreExecutionState(savedState)

	if err != nil {
		return fmt.Errorf("script line execution failed at line %d: %w", currentLineIndex+1, err)
	}

	// Move to next line
	sm.setCurrentScriptLine(currentLineIndex + 1)

	logger.Debug("Script line executed successfully", "script", parsedCmd.Name, "line", currentLineIndex+1)
	return nil // Stay in StateScriptExecuting for next line
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

func (sm *StateMachine) setResolvedCommand(resolved *ResolvedCommand) {
	sm.resolvedCommand = resolved
}

func (sm *StateMachine) getResolvedCommand() *ResolvedCommand {
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
