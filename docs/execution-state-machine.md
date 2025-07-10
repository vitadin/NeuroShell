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
    StateExecuting                         // Running the resolved command
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
StateResolving → StateExecuting (command found)
StateExecuting → StateCompleted (success)
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
        return StateExecuting
    case StateExecuting:
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
    resolvedCmd := sm.context.GetExecutionResolvedCommand()
    parsedCmd := sm.context.GetExecutionParsedCommand()
    input := sm.context.GetExecutionInput()
    
    logger.Debug("State: Executing", "command", resolvedCmd.Name, "type", resolvedCmd.Type)
    
    if sm.echoCommands {
        fmt.Printf("%%> %s\n", input)
    }
    
    return resolvedCmd.Command.Execute(parsedCmd.Options, parsedCmd.Message)
}
```

### Phase 3: Command Resolution Integration

**Priority Resolution in State Machine**:

```go
func (sm *ExecutionStateMachine) resolveCommand(name string) (*ResolvedCommand, error) {
    // 1. Try builtin commands (highest priority)
    if cmd, exists := commands.GetGlobalRegistry().Get(name); exists {
        return &ResolvedCommand{
            Name:    name,
            Type:    CommandTypeBuiltin,
            Source:  "builtin",
            Command: cmd,
        }, nil
    }
    
    // 2. Try stdlib scripts (medium priority)
    if scriptContent := sm.context.GetStdlibScript(name); scriptContent != "" {
        scriptCmd := NewScriptCommand(name, scriptContent, CommandTypeStdlib)
        return &ResolvedCommand{
            Name:    name,
            Type:    CommandTypeStdlib,
            Source:  "stdlib",
            Command: scriptCmd,
        }, nil
    }
    
    // 3. Try user scripts (lowest priority)
    if scriptContent := sm.context.GetUserScript(name); scriptContent != "" {
        scriptCmd := NewScriptCommand(name, scriptContent, CommandTypeUser)
        return &ResolvedCommand{
            Name:    name,
            Type:    CommandTypeUser,
            Source:  "user",
            Command: scriptCmd,
        }, nil
    }
    
    return nil, fmt.Errorf("command not found: %s", name)
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

func (ctx *NeuroContext) SetExecutionResolvedCommand(resolved *ResolvedCommand)
func (ctx *NeuroContext) GetExecutionResolvedCommand() *ResolvedCommand

func (ctx *NeuroContext) SetExecutionError(err error)
func (ctx *NeuroContext) GetExecutionError() error

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

## Future Extensions

The state machine design enables future enhancements:

1. **Debugging State**: Add StateDebugging for step-through execution
2. **Profiling State**: Add StateProfiling for performance measurement
3. **Caching State**: Add StateCaching for command result caching
4. **Conditional Execution**: Add states for if/then/else logic
5. **Loop Execution**: Add states for iterative command execution

This design truly captures the essence of a state machine while seamlessly integrating interpolation as a fundamental execution capability and maintaining compatibility with the existing command-service-context architecture.