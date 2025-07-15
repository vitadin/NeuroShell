# Stack-Based Execution Engine Design

## Overview

This document describes the refactoring of NeuroShell's execution engine from a **state-machine-driven model** to a **stack-based execution model** while preserving the proven command processing pipeline.

### Current vs Proposed Architecture

**Current Model (State-Machine-Driven)**:
- Complex state transitions control overall execution flow
- States: `StateReceived → StateInterpolating → StateParsing → StateResolving → {StateExecuting|StateScriptLoaded|StateTryResolving} → StateCompleted`
- Script and try commands have special execution states
- Commands interact with state machine directly

**Proposed Model (Stack-Based)**:
- **Stack Machine** controls overall execution flow
- **State Processor** handles individual command processing: `Interpolation → Parsing → Resolving → Execution`
- All commands (including scripts and try) use unified stack mechanism
- Commands interact only with services, never with execution engine directly

## Core Architecture

### Stack Machine Model

The execution engine becomes a **Stack Machine** where:

1. **Context holds execution stack** - each item is a raw command string (unparsed, with variables)
2. **Main execution loop**: `Pop command → Process command → Repeat until stack empty`
3. **Command processing pipeline**: `Interpolation → Parsing → Resolving → Execution`
4. **Commands manipulate stack** through StackService during their execution

```go
// Main Stack Machine Loop
func (sm *StackMachine) processStack() error {
    for !sm.context.GetStack().IsEmpty() {
        rawCommand := sm.context.GetStack().Pop()
        
        // Process individual command through state pipeline
        err := sm.processCommand(rawCommand)
        if err != nil && !sm.isInTryBlock() {
            return err // Normal error propagation
        }
        if err != nil && sm.isInTryBlock() {
            sm.handleTryError(err) // Try block error capture
            sm.skipToTryBlockEnd()
        }
    }
    return nil
}

// Individual Command Processing (preserves current proven pipeline)
func (sm *StackMachine) processCommand(rawCommand string) error {
    // 1. Variable Interpolation (StateInterpolating equivalent)
    interpolated, err := sm.interpolateVariables(rawCommand)
    if err != nil {
        return err
    }
    
    // 2. Command Parsing (StateParsing equivalent)
    parsed, err := sm.parseCommand(interpolated)
    if err != nil {
        return err
    }
    
    // 3. Command Resolution (StateResolving equivalent)
    resolved, err := sm.resolveCommand(parsed)
    if err != nil {
        return err
    }
    
    // 4. Command Execution (StateExecuting equivalent)
    return sm.executeCommand(resolved)
}
```

### Enhanced Context and Stack

```go
type StackContext struct {
    // Execution stack (LIFO order)
    executionStack []string
    
    // Try block management
    tryBlocks []TryBlockContext
    currentTryDepth int
    
    // Variable storage (unchanged)
    variables map[string]string
    systemVariables map[string]string
    
    // Session state (unchanged)
    sessionState SessionState
}

type TryBlockContext struct {
    ID string           // Unique identifier for this try block
    StartDepth int      // Stack depth when try block started
    ErrorCaptured bool  // Whether an error has been captured
}
```

### Enhanced Stack Service

```go
type StackService interface {
    // Basic stack operations
    PushCommand(command string)
    PushCommands(commands []string)
    PopCommand() (string, error)
    PeekCommand() (string, error)
    ClearStack()
    GetStackSize() int
    IsEmpty() bool
    
    // Try block support
    PushErrorBoundary(tryID string)
    IsInTryBlock() bool
    GetCurrentTryID() string
}
```

## Command Processing Pipeline

### 1. Variable Interpolation (StateInterpolating)
- **Purpose**: Expand `${variable}` expressions in raw commands
- **Input**: `"\echo Hello ${@user}"`
- **Output**: `"\echo Hello john"`
- **Unchanged**: Uses existing interpolation logic

### 2. Command Parsing (StateParsing)
- **Purpose**: Parse command structure and extract arguments
- **Input**: `"\echo Hello john"`
- **Output**: `{Command: "echo", Args: {}, Input: "Hello john"}`
- **Unchanged**: Uses existing parsing logic

### 3. Command Resolution (StateResolving)
- **Purpose**: Find command implementation (builtin → stdlib → user)
- **Input**: `{Command: "echo", ...}`
- **Output**: `{Type: CommandTypeBuiltin, Implementation: EchoCommand}`
- **Unchanged**: Uses existing resolution logic

### 4. Command Execution (StateExecuting)
- **Purpose**: Execute the resolved command
- **Input**: `{Implementation: EchoCommand, Args: {...}}`
- **Output**: Command side effects and potential stack manipulation
- **Enhanced**: Commands only interact with services

## Error Handling: Try Command Design

### Try Command Error Boundary System

The `\try` command implements error isolation using **error boundary markers** on the stack:

```go
// TryCommand.Execute() implementation
func (c *TryCommand) Execute(args map[string]string, input string) error {
    stackService, err := services.GetGlobalStackService()
    if err != nil {
        return err
    }
    
    targetCommand := extractTargetCommand(args, input)
    tryID := generateUniqueTryID() // "try_id_123"
    
    // Push error boundary markers around target command
    stackService.PushCommand("ERROR_BOUNDARY_END:" + tryID)
    stackService.PushCommand(targetCommand)
    stackService.PushCommand("ERROR_BOUNDARY_START:" + tryID)
    
    return nil
}
```

### Concrete Error Handling Example

**Scenario**: `\try \bash echo "test" && exit 1`

#### Step 1: Try Command Execution
```go
// Stack after TryCommand.Execute():
[
  "ERROR_BOUNDARY_END:try_id_123",     // End marker  
  "\bash echo \"test\" && exit 1",      // Target command
  "ERROR_BOUNDARY_START:try_id_123",   // Start marker
]
```

#### Step 2: Stack Processing
```go
// Pop "ERROR_BOUNDARY_START:try_id_123"
sm.enterTryBlock("try_id_123")

// Pop "\bash echo \"test\" && exit 1"  
err := sm.processCommand("\bash echo \"test\" && exit 1")
// Command processing:
// 1. Interpolation: No variables → "\bash echo \"test\" && exit 1"
// 2. Parsing: {Command: "bash", Input: "echo \"test\" && exit 1"}
// 3. Resolution: BashCommand
// 4. Execution: BashCommand.Execute() → fails with exit code 1
// Returns: error "command failed with exit code 1"

// Error occurs in try block
sm.handleTryError(error("command failed with exit code 1"))
sm.skipToTryBlockEnd()
```

#### Step 3: Error Capture
```go
func (sm *StackMachine) handleTryError(err error) {
    variableService, _ := services.GetGlobalVariableService()
    
    // Set error variables exactly like current implementation
    variableService.SetSystemVariable("_status", "1")
    variableService.SetSystemVariable("_error", "command failed with exit code 1")
    
    // Preserve output from before failure (_output = "test" from echo)
    currentOutput, _ := variableService.GetVariable("_output")
    // _output remains "test"
}

func (sm *StackMachine) skipToTryBlockEnd() {
    currentTryID := sm.getCurrentTryID() // "try_id_123"
    
    for !sm.stack.IsEmpty() {
        command := sm.stack.Pop()
        if command == "ERROR_BOUNDARY_END:"+currentTryID {
            sm.exitTryBlock(currentTryID)
            return
        }
        // Skip any other commands in the try block
    }
}
```

#### Step 4: Final State
```go
// Variables after try completion:
_status = "1"                               // Command failed
_error = "command failed with exit code 1"  // Exact error message  
_output = "test"                            // Output preserved

// Try command succeeded in capturing the error
// Stack processing continues normally
```

### Nested Try Support

The error boundary system supports nested try blocks:

```go
// Example: \try \try \bash exit 1
// Stack after both try commands:
[
  "ERROR_BOUNDARY_END:try_id_456",    // Outer try end
  "ERROR_BOUNDARY_END:try_id_123",    // Inner try end  
  "\bash exit 1",                     // Failing command
  "ERROR_BOUNDARY_START:try_id_123",  // Inner try start
  "ERROR_BOUNDARY_START:try_id_456",  // Outer try start
]

// Processing:
// 1. Enter outer try block (try_id_456)
// 2. Enter inner try block (try_id_123) 
// 3. Command fails, inner try captures error
// 4. Inner try completes successfully
// 5. Outer try sees successful inner try (no error to capture)
```

## Command Interface Refactoring

### Service-Only Interaction Pattern

Commands can **only** interact with services, never with the execution engine directly:

```go
// Current pattern (problematic)
func (c *IfCommand) Execute(args map[string]string, input string) error {
    // Command directly accesses global services
    if stackService, err := services.GetGlobalStackService(); err == nil {
        stackService.PushCommand(input)
    }
    // No separation between command and execution engine
}

// Proposed pattern (clean separation)
func (c *IfCommand) Execute(args map[string]string, input string) error {
    // Command only uses services provided to it
    condition := args["condition"]
    result := c.evaluateCondition(condition)
    
    if result && strings.TrimSpace(input) != "" {
        stackService, err := services.GetGlobalStackService()
        if err != nil {
            return err
        }
        stackService.PushCommand(input)
    }
    
    return nil
}
```

### Updated Command Examples

#### If Command (Stack-Based)
```go
func (c *IfCommand) Execute(args map[string]string, input string) error {
    condition := args["condition"]
    if condition == "" {
        return fmt.Errorf("condition parameter is required")
    }
    
    result := c.evaluateCondition(condition)
    
    // Store result for debugging
    if variableService, err := services.GetGlobalVariableService(); err == nil {
        _ = variableService.SetSystemVariable("#if_result", strconv.FormatBool(result))
    }
    
    // Push command to stack if condition is true
    if result && strings.TrimSpace(input) != "" {
        if stackService, err := services.GetGlobalStackService(); err == nil {
            stackService.PushCommand(input)
        }
    }
    
    return nil
}
```

#### Script Execution (Stack-Based)
```go
func executeScript(scriptPath string) error {
    lines, err := loadScriptLines(scriptPath)
    if err != nil {
        return err
    }
    
    stackService, err := services.GetGlobalStackService()
    if err != nil {
        return err
    }
    
    // Push all script lines to stack in reverse order (LIFO execution)
    for i := len(lines) - 1; i >= 0; i-- {
        stackService.PushCommand(lines[i])
    }
    
    return nil
}
```

## Implementation Strategy

### Phase 1: Stack Infrastructure Enhancement
1. **Enhanced Context Stack**
   - Robust stack implementation with try block context tracking
   - Thread-safe operations with comprehensive introspection
   - Stack depth tracking for recursion prevention

2. **Enhanced Stack Service**
   - Complete stack manipulation API
   - Try block boundary management
   - Service registry integration

### Phase 2: Stack Machine Implementation
1. **Main Execution Loop**
   - Replace current state transition logic with stack processing
   - Integrate existing command processing states
   - Add try block error boundary handling

2. **State Processor Component**
   - Extract and preserve current interpolation logic
   - Extract and preserve current parsing logic
   - Extract and preserve current resolution logic
   - Enhance execution logic for service-only command interaction

### Phase 3: Command Interface Adaptation
1. **Service-Only Pattern**
   - Remove direct state machine dependencies from all commands
   - Update command interface to use only service interactions
   - Implement dependency injection for service access

2. **Command Refactoring**
   - Update `\if` command to use stack-based conditional execution
   - Implement `\try` command with error boundary system
   - Refactor remaining commands for service-only interaction

### Phase 4: Integration and Testing
1. **Replace Current Engine**
   - Swap out current state machine with stack machine
   - Ensure backward compatibility for all existing functionality
   - Maintain existing command syntax and behavior

2. **Comprehensive Testing**
   - Unit tests for stack machine components
   - Integration tests for command execution
   - Error handling tests for try command scenarios
   - Performance testing for stack operations

## File Organization

### Preserve `internal/statemachine/` Structure

Keep the existing folder structure but rename files to reflect the new architecture:

```
internal/statemachine/
├── stack_machine.go          # Main stack-based execution engine
├── state_processor.go        # Individual command state processing  
├── stack_context.go          # Enhanced context with stack and try block management
├── interpolator.go           # Variable interpolation (extracted from current)
├── parser.go                 # Command parsing (extracted from current)
├── resolver.go               # Command resolution (extracted from current)
├── executor.go               # Command execution (enhanced for service-only)
├── try_handler.go            # Try block error boundary management
└── utils.go                  # Utility functions
```

### Key Design Principles

1. **Preserve Proven Components**: Keep existing interpolation, parsing, and resolution logic
2. **Simplify Execution Model**: Replace complex state transitions with straightforward stack processing  
3. **Clean Separation**: Commands only interact with services, execution engine handles flow control
4. **Unified Command Execution**: All commands (builtin, script, try) use the same stack mechanism
5. **Robust Error Handling**: Try blocks provide comprehensive error isolation and capture

## Benefits

### Architectural Benefits
- **Simplified execution model**: Stack LIFO processing is easier to understand and debug
- **Unified command handling**: No special cases for scripts or try commands
- **Better separation of concerns**: Clear boundaries between commands, services, and execution engine
- **Enhanced testability**: Service-only command interaction enables comprehensive mocking

### Functional Benefits  
- **Predictable execution order**: Stack-based execution is deterministic and traceable
- **Robust error handling**: Try block boundaries provide reliable error isolation
- **Command composition**: Commands can easily push other commands for deferred execution
- **Recursive execution support**: Stack depth tracking prevents infinite loops

### Maintainability Benefits
- **Cleaner codebase**: Removal of complex state transition logic
- **Easier debugging**: Linear stack processing with clear execution traces
- **Modular architecture**: Components can be tested and modified independently
- **Future extensibility**: Stack-based model easily accommodates new command types

This stack-based execution engine provides a more robust, maintainable, and extensible foundation for NeuroShell while preserving all existing functionality and proven architectural components.