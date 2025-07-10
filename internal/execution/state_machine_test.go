package execution

import (
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/parser"
)

// TestStateMachine_NewStateMachine tests state machine creation.
func TestStateMachine_NewStateMachine(t *testing.T) {
	ctx := context.New()
	config := DefaultConfig()

	sm := NewStateMachine(ctx, config)

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	if sm.context != ctx {
		t.Error("Expected context to be set correctly")
	}

	if sm.config != config {
		t.Error("Expected config to be set correctly")
	}

	if sm.interpolator == nil {
		t.Error("Expected interpolator to be initialized")
	}

	if sm.stateStack == nil {
		t.Error("Expected state stack to be initialized")
	}
}

// TestStateMachine_DefaultConfig tests default configuration.
func TestStateMachine_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.EchoCommands != false {
		t.Error("Expected EchoCommands to be false by default")
	}

	if config.MacroExpansion != true {
		t.Error("Expected MacroExpansion to be true by default")
	}

	if config.RecursionLimit != 50 {
		t.Error("Expected RecursionLimit to be 50 by default")
	}
}

// TestStateMachine_StateManagement tests basic state management.
func TestStateMachine_StateManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test initial state
	initialState := sm.getCurrentState()
	if initialState != StateReceived {
		t.Errorf("Expected initial state to be StateReceived, got %s", initialState.String())
	}

	// Test state transitions
	sm.setState(StateInterpolating)
	if sm.getCurrentState() != StateInterpolating {
		t.Error("Expected state to be StateInterpolating after setState")
	}

	sm.setState(StateParsing)
	if sm.getCurrentState() != StateParsing {
		t.Error("Expected state to be StateParsing after setState")
	}
}

// TestStateMachine_InputManagement tests input handling.
func TestStateMachine_InputManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	testInput := "\\echo Hello World"
	sm.setExecutionInput(testInput)

	if sm.getExecutionInput() != testInput {
		t.Errorf("Expected input to be '%s', got '%s'", testInput, sm.getExecutionInput())
	}
}

// TestStateMachine_ErrorHandling tests error management.
func TestStateMachine_ErrorHandling(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test no error initially
	if sm.getExecutionError() != nil {
		t.Error("Expected no error initially")
	}

	// Test setting error
	testError := &executionTestError{"test error"}
	sm.setExecutionError(testError)

	if sm.getExecutionError() != testError {
		t.Error("Expected error to be set correctly")
	}
}

// TestStateMachine_RecursionManagement tests recursion depth tracking.
func TestStateMachine_RecursionManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test initial recursion depth
	if sm.getRecursionDepth() != 0 {
		t.Error("Expected initial recursion depth to be 0")
	}

	// Test incrementing recursion depth
	sm.incrementRecursionDepth()
	if sm.getRecursionDepth() != 1 {
		t.Error("Expected recursion depth to be 1 after increment")
	}

	sm.incrementRecursionDepth()
	if sm.getRecursionDepth() != 2 {
		t.Error("Expected recursion depth to be 2 after second increment")
	}

	// Test resetting recursion depth
	sm.resetRecursionDepth()
	if sm.getRecursionDepth() != 0 {
		t.Error("Expected recursion depth to be 0 after reset")
	}
}

// TestStateMachine_CommandManagement tests parsed command handling.
func TestStateMachine_CommandManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test no command initially
	if sm.getParsedCommand() != nil {
		t.Error("Expected no parsed command initially")
	}

	// Test setting parsed command
	testCmd := &parser.Command{
		Name:    "echo",
		Message: "Hello World",
		Options: make(map[string]string),
	}

	sm.setParsedCommand(testCmd)
	if sm.getParsedCommand() != testCmd {
		t.Error("Expected parsed command to be set correctly")
	}
}

// TestStateMachine_ScriptManagement tests script line handling.
func TestStateMachine_ScriptManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test no script lines initially
	if sm.getScriptLines() != nil {
		t.Error("Expected no script lines initially")
	}

	if sm.getCurrentScriptLine() != 0 {
		t.Error("Expected current script line to be 0 initially")
	}

	if sm.hasMoreScriptLines() {
		t.Error("Expected no more script lines initially")
	}

	// Test setting script lines
	testLines := []string{"\\echo Line 1", "\\echo Line 2", "\\echo Line 3"}
	sm.setScriptLines(testLines)

	if len(sm.getScriptLines()) != 3 {
		t.Error("Expected 3 script lines to be set")
	}

	if !sm.hasMoreScriptLines() {
		t.Error("Expected to have more script lines")
	}

	// Test advancing through script lines
	sm.setCurrentScriptLine(1)
	if sm.getCurrentScriptLine() != 1 {
		t.Error("Expected current script line to be 1")
	}

	sm.setCurrentScriptLine(3)
	if sm.hasMoreScriptLines() {
		t.Error("Expected no more script lines at end")
	}
}

// TestStateMachine_StateSnapshot tests state save/restore.
func TestStateMachine_StateSnapshot(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Set up some state
	sm.setState(StateInterpolating)
	sm.setExecutionInput("test input")
	sm.incrementRecursionDepth()
	sm.incrementRecursionDepth()

	testCmd := &parser.Command{Name: "test", Message: "test message"}
	sm.setParsedCommand(testCmd)

	// Save state
	snapshot := sm.saveExecutionState()

	// Verify snapshot content
	if snapshot.State != StateInterpolating {
		t.Error("Expected snapshot to capture current state")
	}

	if snapshot.Input != "test input" {
		t.Error("Expected snapshot to capture input")
	}

	if snapshot.RecursionDepth != 2 {
		t.Error("Expected snapshot to capture recursion depth")
	}

	if snapshot.ParsedCommand != testCmd {
		t.Error("Expected snapshot to capture parsed command")
	}

	// Modify state
	sm.setState(StateCompleted)
	sm.setExecutionInput("modified input")
	sm.resetRecursionDepth()
	sm.setParsedCommand(nil)

	// Restore state
	sm.restoreExecutionState(snapshot)

	// Verify restoration
	if sm.getCurrentState() != StateInterpolating {
		t.Error("Expected state to be restored")
	}

	if sm.getExecutionInput() != "test input" {
		t.Error("Expected input to be restored")
	}

	if sm.getRecursionDepth() != 2 {
		t.Error("Expected recursion depth to be restored")
	}

	if sm.getParsedCommand() != testCmd {
		t.Error("Expected parsed command to be restored")
	}
}

// TestStateMachine_ClearExecutionData tests data clearing.
func TestStateMachine_ClearExecutionData(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Set up some data
	testCmd := &parser.Command{Name: "test"}
	testResolved := &ResolvedCommand{Name: "test", Type: CommandTypeBuiltin}
	testLines := []string{"line1", "line2"}

	sm.setParsedCommand(testCmd)
	sm.setResolvedCommand(testResolved)
	sm.setScriptLines(testLines)
	sm.setCurrentScriptLine(1)

	// Clear data
	sm.clearExecutionData()

	// Verify clearing
	if sm.getParsedCommand() != nil {
		t.Error("Expected parsed command to be cleared")
	}

	if sm.getResolvedCommand() != nil {
		t.Error("Expected resolved command to be cleared")
	}

	if sm.getScriptLines() != nil {
		t.Error("Expected script lines to be cleared")
	}

	if sm.getCurrentScriptLine() != 0 {
		t.Error("Expected current script line to be reset to 0")
	}
}

// TestStateMachine_DetermineNextState tests state transition logic.
func TestStateMachine_DetermineNextState(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test StateReceived with variables
	sm.setState(StateReceived)
	sm.setExecutionInput("\\echo ${user}")
	nextState := sm.DetermineNextState()
	if nextState != StateInterpolating {
		t.Errorf("Expected StateInterpolating, got %s", nextState.String())
	}

	// Test StateReceived without variables
	sm.setState(StateReceived)
	sm.setExecutionInput("\\echo hello")
	nextState = sm.DetermineNextState()
	if nextState != StateParsing {
		t.Errorf("Expected StateParsing, got %s", nextState.String())
	}

	// Test StateInterpolating
	sm.setState(StateInterpolating)
	nextState = sm.DetermineNextState()
	if nextState != StateReceived {
		t.Errorf("Expected StateReceived, got %s", nextState.String())
	}

	// Test StateParsing
	sm.setState(StateParsing)
	nextState = sm.DetermineNextState()
	if nextState != StateResolving {
		t.Errorf("Expected StateResolving, got %s", nextState.String())
	}

	// Test StateExecuting
	sm.setState(StateExecuting)
	nextState = sm.DetermineNextState()
	if nextState != StateCompleted {
		t.Errorf("Expected StateCompleted, got %s", nextState.String())
	}

	// Test StateScriptLoaded
	sm.setState(StateScriptLoaded)
	nextState = sm.DetermineNextState()
	if nextState != StateScriptExecuting {
		t.Errorf("Expected StateScriptExecuting, got %s", nextState.String())
	}

	// Test StateScriptExecuting with more lines
	sm.setState(StateScriptExecuting)
	sm.setScriptLines([]string{"line1", "line2"})
	sm.setCurrentScriptLine(0)
	nextState = sm.DetermineNextState()
	if nextState != StateReceived {
		t.Errorf("Expected StateReceived, got %s", nextState.String())
	}

	// Test StateScriptExecuting without more lines
	sm.setState(StateScriptExecuting)
	sm.setCurrentScriptLine(2)
	nextState = sm.DetermineNextState()
	if nextState != StateCompleted {
		t.Errorf("Expected StateCompleted, got %s", nextState.String())
	}
}

// TestStateMachine_ParseScriptIntoLines tests script parsing.
func TestStateMachine_ParseScriptIntoLines(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test simple script
	script := `\echo Line 1
\echo Line 2
\echo Line 3`

	lines := sm.parseScriptIntoLines(script)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "\\echo Line 1" {
		t.Errorf("Expected first line to be '\\echo Line 1', got '%s'", lines[0])
	}

	// Test multiline command
	scriptWithMultiline := `\echo Line 1...
continued...
end
\echo Line 2`

	lines = sm.parseScriptIntoLines(scriptWithMultiline)
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines after multiline processing, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Line 1") || !strings.Contains(lines[0], "continued") || !strings.Contains(lines[0], "end") {
		t.Errorf("Expected multiline to be joined, got '%s'", lines[0])
	}
}

// Test helper types and functions

type executionTestError struct {
	message string
}

func (e *executionTestError) Error() string {
	return e.message
}

// TestState_String tests string representation of states.
func TestState_String(t *testing.T) {
	states := map[State]string{
		StateReceived:        "Received",
		StateInterpolating:   "Interpolating",
		StateParsing:         "Parsing",
		StateResolving:       "Resolving",
		StateExecuting:       "Executing",
		StateScriptLoaded:    "ScriptLoaded",
		StateScriptExecuting: "ScriptExecuting",
		StateCompleted:       "Completed",
		StateError:           "Error",
	}

	for state, expected := range states {
		if state.String() != expected {
			t.Errorf("Expected %s.String() to be '%s', got '%s'", state, expected, state.String())
		}
	}

	// Test unknown state
	unknownState := State(999)
	if unknownState.String() != "Unknown" {
		t.Errorf("Expected unknown state to return 'Unknown', got '%s'", unknownState.String())
	}
}

// TestStateMachine_StdlibScriptResolution tests resolving stdlib scripts.
func TestStateMachine_StdlibScriptResolution(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test resolving an existing stdlib script
	resolved, err := sm.resolveCommand("test-script")
	if err != nil {
		t.Fatalf("Expected to resolve stdlib script 'test-script', got error: %v", err)
	}

	if resolved.Name != "test-script" {
		t.Errorf("Expected resolved command name to be 'test-script', got '%s'", resolved.Name)
	}

	if resolved.Type != CommandTypeStdlib {
		t.Errorf("Expected resolved command type to be CommandTypeStdlib, got %s", resolved.Type.String())
	}

	if resolved.ScriptContent == "" {
		t.Error("Expected resolved command to have script content")
	}

	if resolved.ScriptPath == "" {
		t.Error("Expected resolved command to have script path")
	}

	// Test resolving another stdlib script
	resolved, err = sm.resolveCommand("enhanced-echo")
	if err != nil {
		t.Fatalf("Expected to resolve stdlib script 'enhanced-echo', got error: %v", err)
	}

	if resolved.Type != CommandTypeStdlib {
		t.Errorf("Expected resolved command type to be CommandTypeStdlib, got %s", resolved.Type.String())
	}

	// Test resolving non-existent script
	_, err = sm.resolveCommand("non-existent-script")
	if err == nil {
		t.Error("Expected error when resolving non-existent script")
	}
}

// TestStateMachine_StdlibScriptExecution tests executing stdlib scripts through state machine.
func TestStateMachine_StdlibScriptExecution(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test that stdlib script goes through the correct state transitions
	// We'll test the state transitions without full execution due to missing builtin commands in test

	// Start with a stdlib script
	sm.initializeExecution("\\test-script Hello World")

	// Should start in StateReceived
	if sm.getCurrentState() != StateReceived {
		t.Errorf("Expected initial state to be StateReceived, got %s", sm.getCurrentState().String())
	}

	// Process through the states
	err := sm.ProcessCurrentState() // StateReceived
	if err != nil {
		t.Fatalf("StateReceived processing failed: %v", err)
	}

	// Move to next state (should be StateParsing since no variables)
	sm.setState(sm.DetermineNextState())
	if sm.getCurrentState() != StateParsing {
		t.Errorf("Expected state to be StateParsing, got %s", sm.getCurrentState().String())
	}

	err = sm.ProcessCurrentState() // StateParsing
	if err != nil {
		t.Fatalf("StateParsing processing failed: %v", err)
	}

	// Move to next state (should be StateResolving)
	sm.setState(sm.DetermineNextState())
	if sm.getCurrentState() != StateResolving {
		t.Errorf("Expected state to be StateResolving, got %s", sm.getCurrentState().String())
	}

	err = sm.ProcessCurrentState() // StateResolving
	if err != nil {
		t.Fatalf("StateResolving processing failed: %v", err)
	}

	// Verify that command was resolved as stdlib script
	resolved := sm.getResolvedCommand()
	if resolved == nil {
		t.Fatal("Expected resolved command to be set")
	}

	if resolved.Type != CommandTypeStdlib {
		t.Errorf("Expected resolved command type to be CommandTypeStdlib, got %s", resolved.Type.String())
	}

	// Move to next state (should be StateScriptLoaded)
	sm.setState(sm.DetermineNextState())
	if sm.getCurrentState() != StateScriptLoaded {
		t.Errorf("Expected state to be StateScriptLoaded, got %s", sm.getCurrentState().String())
	}
}

// TestStateMachine_ScriptParameterPassing tests script parameter setup and cleanup.
func TestStateMachine_ScriptParameterPassing(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test parameter setup
	args := map[string]string{
		"greeting": "Hello",
		"style":    "bold",
	}
	input := "World"
	commandName := "test-script"

	err := sm.setupScriptParameters(args, input, commandName)
	if err != nil {
		t.Fatalf("Script parameter setup failed: %v", err)
	}

	// Verify standard parameters were set
	if value, _ := ctx.GetVariable("_0"); value != commandName {
		t.Errorf("Expected _0 to be '%s', got '%s'", commandName, value)
	}

	if value, _ := ctx.GetVariable("_1"); value != input {
		t.Errorf("Expected _1 to be '%s', got '%s'", input, value)
	}

	if value, _ := ctx.GetVariable("_*"); value != input {
		t.Errorf("Expected _* to be '%s', got '%s'", input, value)
	}

	// Verify named parameters were set
	if value, _ := ctx.GetVariable("greeting"); value != "Hello" {
		t.Errorf("Expected greeting to be 'Hello', got '%s'", value)
	}

	if value, _ := ctx.GetVariable("style"); value != "bold" {
		t.Errorf("Expected style to be 'bold', got '%s'", value)
	}

	// Verify _@ contains named args
	if value, _ := ctx.GetVariable("_@"); !strings.Contains(value, "greeting=Hello") {
		t.Errorf("Expected _@ to contain 'greeting=Hello', got '%s'", value)
	}

	// Test parameter cleanup
	err = sm.cleanupScriptParameters()
	if err != nil {
		t.Fatalf("Script parameter cleanup failed: %v", err)
	}

	// Verify standard parameters were cleared
	if value, _ := ctx.GetVariable("_0"); value != "" {
		t.Errorf("Expected _0 to be cleared, got '%s'", value)
	}

	if value, _ := ctx.GetVariable("_1"); value != "" {
		t.Errorf("Expected _1 to be cleared, got '%s'", value)
	}

	// Note: Named parameters (greeting, style) are not automatically cleared by cleanupScriptParameters
	// They remain in the context as user variables, which is correct behavior
}

// TestStateMachine_RecursiveStateStackManagement tests the state stack for script-to-script calls.
func TestStateMachine_RecursiveStateStackManagement(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Simulate parent script state
	sm.initializeExecution("\\test-script Hello")
	sm.setState(StateScriptExecuting)
	sm.setScriptLines([]string{"\\enhanced-echo World", "\\echo Done"})
	sm.setCurrentScriptLine(0)

	// Verify initial state
	if len(sm.stateStack) != 0 {
		t.Error("Expected empty state stack initially")
	}

	// Save state before recursive call (simulating processScriptExecuting)
	savedState := sm.saveExecutionState()

	// Verify state was saved to stack
	if len(sm.stateStack) != 1 {
		t.Errorf("Expected state stack length 1 after save, got %d", len(sm.stateStack))
	}

	// Simulate recursive call by setting up new execution state
	sm.initializeExecution("\\enhanced-echo World")
	sm.setState(StateScriptExecuting)
	sm.setScriptLines([]string{"\\echo Enhanced output"})
	sm.setCurrentScriptLine(0)

	// Verify current state reflects the nested script
	if sm.getExecutionInput() != "\\enhanced-echo World" {
		t.Errorf("Expected nested script input, got '%s'", sm.getExecutionInput())
	}

	if sm.getCurrentScriptLine() != 0 {
		t.Errorf("Expected nested script line 0, got %d", sm.getCurrentScriptLine())
	}

	// Simulate completion of nested script and restore parent state
	sm.restoreExecutionState(savedState)

	// Verify parent state was restored
	if sm.getExecutionInput() != "\\test-script Hello" {
		t.Errorf("Expected parent script input restored, got '%s'", sm.getExecutionInput())
	}

	if sm.getCurrentScriptLine() != 0 {
		t.Errorf("Expected parent script line 0, got %d", sm.getCurrentScriptLine())
	}

	if len(sm.getScriptLines()) != 2 {
		t.Errorf("Expected parent script to have 2 lines, got %d", len(sm.getScriptLines()))
	}

	if sm.getCurrentState() != StateScriptExecuting {
		t.Errorf("Expected parent state StateScriptExecuting, got %s", sm.getCurrentState().String())
	}

	// Verify state stack is back to empty
	if len(sm.stateStack) != 0 {
		t.Errorf("Expected empty state stack after restore, got %d", len(sm.stateStack))
	}
}

// TestCommandType_String tests string representation of command types.
func TestCommandType_String(t *testing.T) {
	types := map[CommandType]string{
		CommandTypeBuiltin: "Builtin",
		CommandTypeStdlib:  "Stdlib",
		CommandTypeUser:    "User",
	}

	for cmdType, expected := range types {
		if cmdType.String() != expected {
			t.Errorf("Expected %s.String() to be '%s', got '%s'", cmdType, expected, cmdType.String())
		}
	}

	// Test unknown type
	unknownType := CommandType(999)
	if unknownType.String() != "Unknown" {
		t.Errorf("Expected unknown type to return 'Unknown', got '%s'", unknownType.String())
	}
}
