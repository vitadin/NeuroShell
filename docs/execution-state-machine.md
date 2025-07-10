# NeuroShell Execution State Machine Design

## Vision: Execution as State Machine with Integrated Interpolation

Transform NeuroShell's execution pipeline into a unified **state machine** that operates on context, where each execution step is a **state** with defined **transitions**, and interpolation is a core part of the state machine logic rather than a separate service.

## Current Problems

### Existing Issues:
- **Code Duplication**: ScriptService vs ScriptExecutor (~90% overlap)
- **Limited executeCommand**: Only handles builtin commands
- **No Priority Resolution**: Missing stdlib and user command support  
- **Architecture Inconsistency**: Mixed approaches for execution
- **Service Overhead**: Interpolation as separate service adds unnecessary complexity
- **No Recursive Execution**: Variables containing commands not fully supported

### Architecture Violations:
- ScriptExecutor imports both services and commands (orchestration layer)
- Interpolation service accessed through service registry in hot execution path
- executeCommand function lacks extensibility for stdlib/user commands

## True State Machine Design

### Core States:
```go
type ExecutionState int

const (
    StateReceived     ExecutionState = iota  // Initial state: line/command received
    StateInterpolating                      // Expanding variables/macros
    StateParsing                           // Parsing command structure  
    StateResolving                         // Finding command (builtin → stdlib → user)
    StateExecuting                         // Running builtin commands
    StateScriptLoaded                      // Script content loaded, ready to process
    StateScriptExecuting                   // Processing script lines natively
    StateCompleted                         // Execution finished (success)
    StateError                            // Execution failed
)
```

### State Machine Interface:
```go
type ExecutionStateMachine struct {
    context      *context.NeuroContext
    
    // Integrated interpolation (not a separate service)
    interpolator *CoreInterpolator
    
    // Configuration
    echoCommands   bool
    macroExpansion bool
    recursionLimit int
}

// State machine operations - state is stored in context
func (sm *ExecutionStateMachine) Execute(input string) error
func (sm *ExecutionStateMachine) DetermineNextState() ExecutionState
func (sm *ExecutionStateMachine) ProcessCurrentState() error
func (sm *ExecutionStateMachine) GetCurrentState() ExecutionState
func (sm *ExecutionStateMachine) SetState(state ExecutionState)
```

### State Transitions:
```
StateReceived → StateInterpolating (if contains variables)
StateReceived → StateParsing (if no variables)
StateInterpolating → StateReceived (recursive: expanded line re-enters)
StateParsing → StateResolving (command parsed successfully)
StateResolving → StateExecuting (if builtin command found)
StateResolving → StateScriptLoaded (if stdlib/user script found)
StateExecuting → StateCompleted (builtin command success)
StateScriptLoaded → StateScriptExecuting (script ready to process)
StateScriptExecuting → StateReceived (recursive: each script line re-enters)
StateScriptExecuting → StateCompleted (script finished)
Any State → StateError (on failure)
StateCompleted/StateError → StateReceived (for next execution)
```

## Integrated Interpolation Engine

### Core Principle: Move Interpolation Into State Machine Core

**Why integrate interpolation:**
1. **Fundamental Operation**: Variable expansion is core to execution, not an optional service
2. **Performance**: Eliminates service lookup overhead in hot path
3. **State Awareness**: Interpolation needs to know execution state for recursion control
4. **Simplicity**: Direct access to context, no service layer indirection

### CoreInterpolator:
```go
type CoreInterpolator struct {
    context *context.NeuroContext
}

// Core interpolation methods (no longer a service)
func (ci *CoreInterpolator) InterpolateCommandLine(line string) (string, bool, error)
func (ci *CoreInterpolator) InterpolateCommand(cmd *parser.Command) (*parser.Command, error)
func (ci *CoreInterpolator) HasVariables(line string) bool
func (ci *CoreInterpolator) ExpandVariables(text string) (string, error)
```

## Implementation Plan

### Phase 1: Create State Machine Core

**New Execution State Machine** (`internal/execution/state_machine.go`):

```go
package execution

type ExecutionStateMachine struct {
    // State machine core
    context        *context.NeuroContext
    interpolator   *CoreInterpolator
    
    // Configuration
    echoCommands   bool
    recursionLimit int
}

// Main execution method - true state machine
func (sm *ExecutionStateMachine) Execute(input string) error {
    // Initialize execution in context
    sm.context.SetExecutionState(StateReceived)
    sm.context.SetExecutionInput(input)
    sm.context.SetExecutionError(nil)
    sm.context.ResetExecutionRecursionDepth()
    
    for {
        currentState := sm.context.GetExecutionState()
        if currentState == StateCompleted || currentState == StateError {
            break
        }
        
        if err := sm.ProcessCurrentState(); err != nil {
            sm.context.SetExecutionError(err)
            sm.context.SetExecutionState(StateError)
            break
        }
        
        nextState := sm.DetermineNextState()
        sm.context.SetExecutionState(nextState)
    }
    
    return sm.context.GetExecutionError()
}

// State processor - handles current state logic
func (sm *ExecutionStateMachine) ProcessCurrentState() error {
    currentState := sm.context.GetExecutionState()
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
    default:
        return fmt.Errorf("unknown state: %v", currentState)
    }
}

// State transition logic - determines next state based on current state
func (sm *ExecutionStateMachine) DetermineNextState() ExecutionState {
    currentState := sm.context.GetExecutionState()
    switch currentState {
    case StateReceived:
        input := sm.context.GetExecutionInput()
        if sm.interpolator.HasVariables(input) {
            return StateInterpolating
        }
        return StateParsing
    case StateInterpolating:
        return StateReceived  // Recursive re-entry with expanded input
    case StateParsing:
        return StateResolving
    case StateResolving:
        // Determine if builtin command or script
        if sm.context.GetExecutionResolvedBuiltinCommand() != nil {
            return StateExecuting
        }
        if sm.context.GetExecutionScriptContent() != "" {
            return StateScriptLoaded
        }
        return StateError
    case StateExecuting:
        return StateCompleted
    case StateScriptLoaded:
        return StateScriptExecuting
    case StateScriptExecuting:
        // Check if more script lines to process
        if sm.context.HasMoreScriptLines() {
            return StateReceived  // Recursive: next script line re-enters
        }
        return StateCompleted
    default:
        return StateError
    }
}
```

### Phase 2: State-Specific Processing Methods

**State Processors**:

```go
func (sm *ExecutionStateMachine) processReceived() error {
    // Log input, prepare for processing
    input := sm.context.GetExecutionInput()
    logger.Debug("State: Received", "input", input)
    return nil
}

func (sm *ExecutionStateMachine) processInterpolating() error {
    input := sm.context.GetExecutionInput()
    logger.Debug("State: Interpolating", "input", input)
    
    expanded, hasVariables, err := sm.interpolator.InterpolateCommandLine(input)
    if err != nil {
        return fmt.Errorf("interpolation failed: %w", err)
    }
    
    if hasVariables {
        // Check recursion limit
        recursionDepth := sm.context.GetExecutionRecursionDepth()
        if recursionDepth >= sm.recursionLimit {
            return fmt.Errorf("recursion limit exceeded")
        }
        
        // Increment recursion depth and set up for recursive re-entry
        sm.context.IncrementExecutionRecursionDepth()
        sm.context.SetExecutionInput(expanded)
        logger.Debug("Macro expansion - recursive re-entry", "original", input, "expanded", expanded)
    }
    
    return nil
}

func (sm *ExecutionStateMachine) processParsing() error {
    input := sm.context.GetExecutionInput()
    logger.Debug("State: Parsing", "input", input)
    
    cmd := parser.ParseInput(input)
    if cmd == nil {
        return fmt.Errorf("failed to parse command: %s", input)
    }
    
    sm.context.SetExecutionParsedCommand(cmd)
    return nil
}

func (sm *ExecutionStateMachine) processResolving() error {
    parsedCmd := sm.context.GetExecutionParsedCommand()
    logger.Debug("State: Resolving", "command", parsedCmd.Name)
    
    // Priority-based resolution: builtin → stdlib → user
    resolved, err := sm.resolveCommand(parsedCmd.Name)
    if err != nil {
        return fmt.Errorf("command not found: %s", parsedCmd.Name)
    }
    
    sm.context.SetExecutionResolvedCommand(resolved)
    return nil
}

func (sm *ExecutionStateMachine) processExecuting() error {
    builtinCmd := sm.context.GetExecutionResolvedBuiltinCommand()
    parsedCmd := sm.context.GetExecutionParsedCommand()
    input := sm.context.GetExecutionInput()
    
    logger.Debug("State: Executing", "command", parsedCmd.Name, "type", "builtin")
    
    if sm.echoCommands {
        fmt.Printf("%%> %s\n", input)
    }
    
    return builtinCmd.Execute(parsedCmd.Options, parsedCmd.Message)
}

func (sm *ExecutionStateMachine) processScriptLoaded() error {
    parsedCmd := sm.context.GetExecutionParsedCommand()
    scriptContent := sm.context.GetExecutionScriptContent()
    scriptType := sm.context.GetExecutionScriptType()
    
    logger.Debug("State: ScriptLoaded", "command", parsedCmd.Name, "type", scriptType.String())
    
    // Setup script parameters using the same system as ScriptCommand
    err := sm.setupScriptParameters(parsedCmd.Options, parsedCmd.Message)
    if err != nil {
        return fmt.Errorf("failed to setup script parameters: %w", err)
    }
    
    // Parse script content into executable lines
    lines := sm.parseScriptIntoLines(scriptContent)
    sm.context.SetExecutionScriptLines(lines)
    sm.context.SetExecutionScriptCurrentLine(0)
    
    return nil
}

func (sm *ExecutionStateMachine) processScriptExecuting() error {
    lines := sm.context.GetExecutionScriptLines()
    currentLineIndex := sm.context.GetExecutionScriptCurrentLine()
    parsedCmd := sm.context.GetExecutionParsedCommand()
    
    logger.Debug("State: ScriptExecuting", "script", parsedCmd.Name, "line", currentLineIndex+1, "total", len(lines))
    
    if currentLineIndex >= len(lines) {
        // Script finished - cleanup parameters
        err := sm.cleanupScriptParameters()
        if err != nil {
            logger.Error("Failed to cleanup script parameters", "error", err)
        }
        return nil  // Will transition to StateCompleted
    }
    
    // Get current line to execute
    line := lines[currentLineIndex]
    line = strings.TrimSpace(line)
    
    // Skip empty lines and comments
    if line == "" || strings.HasPrefix(line, "%%") {
        sm.context.SetExecutionScriptCurrentLine(currentLineIndex + 1)
        return nil  // Stay in StateScriptExecuting for next line
    }
    
    if sm.echoCommands {
        fmt.Printf("%%> %s\n", line)
    }
    
    // Save current execution state for recursive call
    savedState := sm.context.SaveExecutionState()
    
    // Execute script line through state machine recursively
    // This line will go through: StateReceived → StateInterpolating → ... → StateCompleted
    err := sm.Execute(line)
    
    // Restore execution state after recursive call
    sm.context.RestoreExecutionState(savedState)
    
    if err != nil {
        return fmt.Errorf("script line execution failed at line %d: %w", currentLineIndex+1, err)
    }
    
    // Move to next line
    sm.context.SetExecutionScriptCurrentLine(currentLineIndex + 1)
    
    return nil  // Stay in StateScriptExecuting for next line
}
```

### Phase 3: Command Resolution Integration

**Native Script Resolution in State Machine**:

The state machine now handles script resolution natively without creating ScriptCommand wrapper objects:

```go
func (sm *ExecutionStateMachine) processResolving() error {
    parsedCmd := sm.context.GetExecutionParsedCommand()
    commandName := parsedCmd.Name
    
    logger.Debug("State: Resolving", "command", commandName)
    
    // Priority 1: Try builtin commands (highest priority)
    if builtinCmd, exists := commands.GetGlobalRegistry().Get(commandName); exists {
        sm.context.SetExecutionResolvedBuiltinCommand(builtinCmd)
        logger.Debug("Resolved to builtin command", "command", commandName)
        return nil  // Will transition to StateExecuting
    }
    
    // Priority 2: Try stdlib scripts (medium priority)
    if scriptContent := sm.context.GetStdlibScript(commandName); scriptContent != "" {
        sm.context.SetExecutionScriptContent(scriptContent)
        sm.context.SetExecutionScriptType(CommandTypeStdlib)
        logger.Debug("Resolved to stdlib script", "command", commandName)
        return nil  // Will transition to StateScriptLoaded
    }
    
    // Priority 3: Try user scripts (lowest priority)
    if scriptContent := sm.context.GetUserScript(commandName); scriptContent != "" {
        sm.context.SetExecutionScriptContent(scriptContent)
        sm.context.SetExecutionScriptType(CommandTypeUser)
        logger.Debug("Resolved to user script", "command", commandName)
        return nil  // Will transition to StateScriptLoaded
    }
    
    return fmt.Errorf("command not found: %s", commandName)
}

// Helper methods for script handling
func (sm *ExecutionStateMachine) parseScriptIntoLines(scriptContent string) []string {
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

func (sm *ExecutionStateMachine) setupScriptParameters(args map[string]string, input string) error {
    vs, err := services.GetGlobalVariableService()
    if err != nil {
        return fmt.Errorf("variable service not available: %w", err)
    }
    
    parsedCmd := sm.context.GetExecutionParsedCommand()
    
    // Set standard script parameters (same as ScriptCommand)
    vs.SetSystemVariable("_0", parsedCmd.Name)  // Command name
    vs.SetSystemVariable("_1", input)           // Input parameter
    vs.SetSystemVariable("_*", input)           // All positional args
    
    // Set named arguments as variables
    var namedArgs []string
    for key, value := range args {
        vs.SetSystemVariable(key, value)
        namedArgs = append(namedArgs, key+"="+value)
    }
    vs.SetSystemVariable("_@", strings.Join(namedArgs, " "))
    
    return nil
}

func (sm *ExecutionStateMachine) cleanupScriptParameters() error {
    vs, err := services.GetGlobalVariableService()
    if err != nil {
        return fmt.Errorf("variable service not available: %w", err)
    }
    
    // Clear standard script parameters
    parameterNames := []string{"_0", "_1", "_*", "_@"}
    for _, name := range parameterNames {
        vs.SetSystemVariable(name, "")
    }
    
    return nil
}
```

### Phase 4: Integration Points

**ExecutionEngine Service** (thin wrapper around state machine):

```go
type ExecutionEngineService struct {
    stateMachine *execution.ExecutionStateMachine
}

func (e *ExecutionEngineService) Execute(line string) error {
    return e.stateMachine.Execute(line)
}

func (e *ExecutionEngineService) ExecuteScript(scriptPath string) error {
    // Load script content and execute line by line using state machine
    content, err := os.ReadFile(scriptPath)
    if err != nil {
        return err
    }
    
    scanner := bufio.NewScanner(strings.NewReader(string(content)))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "%%") {
            continue
        }
        
        if err := e.stateMachine.Execute(line); err != nil {
            return err
        }
    }
    
    return scanner.Err()
}
```

**Replace executeCommand in shell/handler.go**:

```go
func executeCommand(c *ishell.Context, cmd *parser.Command) {
    engine, err := services.GetGlobalExecutionEngine()
    if err != nil {
        c.Printf("Error: execution engine not available: %s\n", err.Error())
        return
    }
    
    // Convert parsed command back to string for state machine
    cmdLine := cmd.String()
    
    err = engine.Execute(cmdLine)
    if err != nil {
        c.Printf("Error: %s\n", err.Error())
    }
}
```

## Integrated Interpolation Implementation

**CoreInterpolator** (`internal/execution/interpolator.go`):

```go
type CoreInterpolator struct {
    context *context.NeuroContext
}

func (ci *CoreInterpolator) InterpolateCommandLine(line string) (string, bool, error) {
    // Direct context access, no service layer
    if !ci.HasVariables(line) {
        return line, false, nil
    }
    
    expanded, err := ci.ExpandVariables(line)
    return expanded, true, err
}

func (ci *CoreInterpolator) HasVariables(line string) bool {
    // Check for ${...} patterns
    return strings.Contains(line, "${")
}

func (ci *CoreInterpolator) ExpandVariables(text string) (string, error) {
    // Variable expansion logic with direct context access
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    
    result := re.ReplaceAllStringFunc(text, func(match string) string {
        varName := match[2 : len(match)-1] // Remove ${ and }
        
        if value, err := ci.context.GetVariable(varName); err == nil {
            return value
        }
        
        // Return original if variable not found
        return match
    })
    
    return result, nil
}
```

## Benefits of True State Machine Design

1. **Clear State Transitions**: Each step is explicit with defined transitions
2. **Recursive Capability**: Natural handling of macro expansion through state re-entry
3. **Integrated Interpolation**: Core functionality, not separate service overhead
4. **Extensible**: Easy to add new states (debugging, profiling, caching)
5. **Testable**: Each state can be tested independently
6. **Debuggable**: Clear state tracking and transition logging
7. **Performance**: Direct context access, minimal overhead
8. **Architecture Compliance**: Service wrapper maintains command-service-context flow

## Command Resolution Priority

The state machine implements the command resolution priority:

1. **Builtin Commands** (highest priority)
   - Go-implemented commands in the registry
   - Always resolved first

2. **Stdlib Scripts** (medium priority)  
   - Embedded .neuro scripts in the binary
   - Loaded on startup and stored in context

3. **User Scripts** (lowest priority)
   - User-defined .neuro scripts
   - Loaded from user directories

## Context Integration

### Required Context Methods

The context needs to be extended to store execution state and data:

```go
// Execution state management
func (ctx *NeuroContext) SetExecutionState(state ExecutionState)
func (ctx *NeuroContext) GetExecutionState() ExecutionState

// Execution data storage
func (ctx *NeuroContext) SetExecutionInput(input string)
func (ctx *NeuroContext) GetExecutionInput() string

func (ctx *NeuroContext) SetExecutionParsedCommand(cmd *parser.Command)
func (ctx *NeuroContext) GetExecutionParsedCommand() *parser.Command

func (ctx *NeuroContext) SetExecutionError(err error)
func (ctx *NeuroContext) GetExecutionError() error

// Command resolution storage
func (ctx *NeuroContext) SetExecutionResolvedBuiltinCommand(cmd neurotypes.Command)
func (ctx *NeuroContext) GetExecutionResolvedBuiltinCommand() neurotypes.Command

func (ctx *NeuroContext) SetExecutionScriptContent(content string)
func (ctx *NeuroContext) GetExecutionScriptContent() string

func (ctx *NeuroContext) SetExecutionScriptType(scriptType CommandType)
func (ctx *NeuroContext) GetExecutionScriptType() CommandType

// Script execution state
func (ctx *NeuroContext) SetExecutionScriptLines(lines []string)
func (ctx *NeuroContext) GetExecutionScriptLines() []string

func (ctx *NeuroContext) SetExecutionScriptCurrentLine(line int)
func (ctx *NeuroContext) GetExecutionScriptCurrentLine() int

func (ctx *NeuroContext) HasMoreScriptLines() bool

// Execution state save/restore for recursive calls
func (ctx *NeuroContext) SaveExecutionState() ExecutionStateSnapshot
func (ctx *NeuroContext) RestoreExecutionState(snapshot ExecutionStateSnapshot)

// Recursion tracking
func (ctx *NeuroContext) ResetExecutionRecursionDepth()
func (ctx *NeuroContext) IncrementExecutionRecursionDepth()
func (ctx *NeuroContext) GetExecutionRecursionDepth() int

// Script storage (for stdlib/user commands)
func (ctx *NeuroContext) SetStdlibScript(name, content string) error
func (ctx *NeuroContext) GetStdlibScript(name string) string
func (ctx *NeuroContext) SetUserScript(name, content string) error
func (ctx *NeuroContext) GetUserScript(name string) string
```

### Context Storage Fields

The context struct needs these additional fields:

```go
type NeuroContext struct {
    // ... existing fields ...
    
    // Execution state machine data
    executionState      ExecutionState
    executionInput      string
    executionParsedCmd  *parser.Command
    executionResolvedCmd *ResolvedCommand
    executionError      error
    executionRecursionDepth int
    executionMutex      sync.RWMutex
    
    // Script storage for command resolution
    stdlibScripts      map[string]string  // name -> content
    userScripts        map[string]string  // name -> content
    scriptStorageMutex sync.RWMutex
}
```

## File Structure

### New Files:
- `internal/execution/state_machine.go` - Core state machine implementation
- `internal/execution/interpolator.go` - Integrated interpolation engine
- `internal/execution/types.go` - State machine types and constants
- `internal/services/execution_engine_service.go` - Service wrapper

### Modified Files:
- `internal/shell/handler.go` - Use state machine for execution
- `internal/context/context.go` - Add stdlib/user script storage methods
- `cmd/neuro/main.go` - Use state machine for batch execution
- `internal/services/script_service.go` - Add direct execution using state machine

### Deprecated/Removed:
- `internal/services/interpolation_service.go` - Functionality moved to core
- `internal/orchestration/script_executor.go` - Replaced by state machine

## Script-to-Script Execution Flow

The state machine design naturally handles scripts calling other scripts through its recursive execution model.

### How Script Calls Work

When a script calls another script, the state machine processes it through the same execution pipeline:

#### 1. Initial Script Execution
When a script is first executed, the state machine follows this path:
```
StateReceived → StateInterpolating → StateParsing → StateResolving → StateScriptLoaded → StateScriptExecuting
```

#### 2. Script Line Processing
In `StateScriptExecuting`, each script line is processed individually:
- The state machine gets the current script line (e.g., `\send Hello ${user}`)
- It calls `sm.Execute(line)` **recursively** - the same state machine processes this line

#### 3. When Script Line Contains Another Script Call
If a script line calls another script (e.g., `\another-script param1`), here's what happens:

```go
// In processScriptExecuting()
line := "\\another-script param1"  // Current script line

// Save current execution state 
savedState := sm.context.SaveExecutionState()

// Execute script line through state machine recursively
err := sm.Execute(line)  // This line re-enters the state machine
```

#### 4. Recursive State Machine Processing
The recursive `sm.Execute(line)` call processes the script invocation:

```
StateReceived (line="\\another-script param1")
    ↓
StateInterpolating (expand any variables in parameters)
    ↓  
StateParsing (parse the command structure)
    ↓
StateResolving (find "another-script" - resolves to another script)
    ↓
StateScriptLoaded (loads content of "another-script")
    ↓
StateScriptExecuting (processes each line of "another-script")
```

#### 5. Nested Script Execution
The nested script (`another-script`) now executes its lines:
- Each line in `another-script` goes through the **same recursive process**
- If `another-script` calls yet another script, it creates another level of recursion
- Each recursive call has its own execution state stored in context

#### 6. State Management During Recursion
The key insight is how execution state is managed:

```go
type NeuroContext struct {
    // Current execution state (for active execution)
    executionState      ExecutionState
    executionInput      string
    executionParsedCmd  *parser.Command
    // ... other current execution data
    
    // Stack of saved states for recursive calls
    executionStateStack []ExecutionStateSnapshot
}

// Save state before recursive call
savedState := sm.context.SaveExecutionState()

// Restore state after recursive call completes
sm.context.RestoreExecutionState(savedState)
```

#### 7. Complete Flow Example

**Original Script (`main-script.neuro`):**
```neuro
\set[user="Alice"]
\helper-script ${user}
\echo Done with helper
```

**Helper Script (`helper-script.neuro`):**
```neuro
\echo Processing user: ${_1}
\utility-script format
\echo Helper complete
```

**Execution Flow:**
1. **Main script starts**: `StateScriptExecuting` for `main-script`
2. **Line 1**: `\set[user="Alice"]` → recursive execution → builtin command → completes
3. **Line 2**: `\helper-script ${user}` → recursive execution begins
   - Interpolates to `\helper-script Alice`
   - Resolves to `helper-script.neuro`
   - **Nested StateScriptExecuting** for `helper-script`
4. **Helper Line 1**: `\echo Processing user: ${_1}` → recursive execution → builtin command
5. **Helper Line 2**: `\utility-script format` → **Another recursive execution**
   - Creates third level of nesting
   - Executes `utility-script.neuro` completely
   - Returns to helper script
6. **Helper Line 3**: `\echo Helper complete` → completes helper script
7. **Back to main script Line 3**: `\echo Done with helper` → completes main script

#### 8. State Stack Management

The context maintains a stack of execution states:

```go
// Before each recursive call
func (ctx *NeuroContext) SaveExecutionState() ExecutionStateSnapshot {
    snapshot := ExecutionStateSnapshot{
        State:           ctx.executionState,
        Input:          ctx.executionInput,
        ParsedCmd:      ctx.executionParsedCmd,
        ScriptLines:    ctx.executionScriptLines,
        CurrentLine:    ctx.executionScriptCurrentLine,
        RecursionDepth: ctx.executionRecursionDepth,
    }
    ctx.executionStateStack = append(ctx.executionStateStack, snapshot)
    return snapshot
}

// After recursive call completes
func (ctx *NeuroContext) RestoreExecutionState(snapshot ExecutionStateSnapshot) {
    ctx.executionState = snapshot.State
    ctx.executionInput = snapshot.Input
    ctx.executionParsedCmd = snapshot.ParsedCmd
    ctx.executionScriptLines = snapshot.ScriptLines
    ctx.executionScriptCurrentLine = snapshot.CurrentLine
    ctx.executionRecursionDepth = snapshot.RecursionDepth
    // Pop from stack
    ctx.executionStateStack = ctx.executionStateStack[:len(ctx.executionStateStack)-1]
}
```

### Key Benefits of Recursive Script Design

1. **Natural Recursion**: Scripts calling scripts is handled naturally through the same state machine
2. **State Isolation**: Each recursive call has isolated execution state
3. **Parameter Passing**: Script parameters (`${_0}`, `${_1}`, etc.) are set up properly for each script level
4. **Error Propagation**: Errors in nested scripts properly bubble up to calling scripts
5. **Unlimited Nesting**: Scripts can call scripts that call scripts (limited only by recursion limit)
6. **Consistent Behavior**: All script executions follow the same state machine logic

This design treats script execution as a **first-class operation** in the state machine, making script-to-script calls as natural as any other command execution.

## Future Extensions

The state machine design enables future enhancements:

1. **Debugging State**: Add StateDebugging for step-through execution
2. **Profiling State**: Add StateProfiling for performance measurement
3. **Caching State**: Add StateCaching for command result caching
4. **Conditional Execution**: Add states for if/then/else logic
5. **Loop Execution**: Add states for iterative command execution

This design truly captures the essence of a state machine while seamlessly integrating interpolation as a fundamental execution capability and maintaining compatibility with the existing command-service-context architecture.