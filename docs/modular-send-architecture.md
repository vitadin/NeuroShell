# Modular Send Command Architecture

## Vision

Transform the monolithic `\send` command into a set of focused, composable sub-commands orchestrated by a neuro script, following single responsibility principle and leveraging the existing state machine execution engine.

## Overview

### Current Problem

The existing `\send` command implementation (currently disabled) has several architectural issues:
- **Monolithic Logic**: 11-step execution pipeline in a single function (270+ lines)
- **Service Orchestration**: Complex coordination of 5+ services in command logic
- **Testing Complexity**: Difficult to test individual steps in isolation
- **Maintenance Burden**: Hard to debug, modify, or extend individual components
- **Code Duplication**: Similar patterns repeated across sync/stream variants

### Proposed Solution

**Core Principle**: Break the 11-step monolithic send process into focused sub-commands, each handling one specific responsibility, then orchestrate them via a `send.neuro` stdlib script.

**Architecture Benefits**:
- **Modularity**: Each sub-command has single, testable responsibility
- **Composability**: Sub-commands reusable in other contexts
- **Silent Operation**: Clean separation of data processing and presentation
- **State Machine Leverage**: Full utilization of existing execution engine
- **Variable System Power**: Rich inter-command communication via NeuroShell variables

## Architecture Design

### Three-Layer Approach

1. **Builtin Commands Layer**: Atomic, stateless operations (6 commands)
2. **Stdlib Scripts Layer**: Composable logic and error handling (2+ scripts)
3. **Orchestration Layer**: High-level command composition (`send.neuro`)

### Variable-Driven Communication

Commands communicate exclusively through NeuroShell's variable system:
- **Input Variables**: Commands read required data from variables
- **Output Variables**: Commands set results/status in variables
- **Silent Execution**: No console output except errors
- **State Preservation**: Variables maintain state between command executions

## Sub-Commands Specification

### Builtin Commands (Self-Contained & Stateless)

#### 1. `\_session-get-active` - Session Management

**Purpose**: Get active session or create auto session if none exists

**Inputs**: None (reads from session service)

**Outputs**:
- `_session_id`: Active session identifier
- `_session_name`: Session name
- `_session_status`: "found" or "created"

**Error Handling**: Returns error if session creation fails

**Implementation Notes**:
- Uses existing ChatSessionService
- Creates "auto" session as fallback
- No console output except errors

```go
func (c *SessionGetActiveCommand) Execute(args map[string]string, input string) error {
    chatSessionService, err := services.GetGlobalChatSessionService()
    if err != nil {
        return fmt.Errorf("failed to get chat session service: %w", err)
    }
    
    activeSession, err := chatSessionService.GetActiveSession()
    if err != nil {
        // Create auto session fallback
        session, createErr := chatSessionService.CreateSession("auto", "", "")
        if createErr != nil {
            return fmt.Errorf("no active session and failed to create auto: %w", createErr)
        }
        activeSession = session
        variableService.SetSystemVariable("_session_status", "created")
    } else {
        variableService.SetSystemVariable("_session_status", "found")
    }
    
    variableService.SetSystemVariable("_session_id", activeSession.ID)
    variableService.SetSystemVariable("_session_name", activeSession.Name)
    return nil
}
```

#### 2. `\_model-get-config` - Model Configuration

**Purpose**: Get active model configuration from ModelService

**Inputs**: None (reads from model service)

**Outputs**:
- `_model_name`: Active model name (e.g., "gpt-4")
- `_model_provider`: Provider name (e.g., "openai")
- `_model_params`: JSON-encoded model parameters

**Error Handling**: Returns error if no model configured

**Implementation Notes**:
- Uses existing ModelService
- Encodes parameters as JSON for complex data
- No console output except errors

#### 3. `\_client-setup` - Client & API Key Management

**Purpose**: Determine API key and validate LLM client setup

**Inputs**:
- `_model_provider`: Provider name (required)

**Outputs**:
- `_client_ready`: "true" if client configured, "false" otherwise
- `_api_key_status`: "configured", "missing", or "invalid"
- `_api_key_source`: "model_config", "env_var", or "user_config"

**Error Handling**: Returns error only for system failures, not missing API keys

**Implementation Notes**:
- Uses existing ClientFactoryService
- Handles API key resolution priority: model config → env vars → user config
- Validates client without making actual requests
- Sets status variables for caller decision-making

#### 4. `\_session-add-message` - Message Management

**Purpose**: Add single message to session (user or assistant)

**Inputs**:
- `role`: "user" or "assistant" (required)
- `content`: Message content (required)
- `session_id`: Target session (optional, uses `_session_id` if not provided)

**Outputs**:
- `_message_added`: "true" if successful, "false" otherwise
- `_message_id`: Unique message identifier
- `_message_index`: Position in session (0-based)

**Error Handling**: Returns error if session not found or invalid role

**Implementation Notes**:
- Uses existing ChatSessionService
- Supports both explicit session_id and variable-based session
- No console output except errors

#### 5. `\_llm-send` - LLM Interaction Core

**Purpose**: Execute actual LLM request (sync/stream based on `_reply_way`)

**Inputs**:
- `message`: User message to send (required)
- `_session_id`: Session context (required variable)
- `_model_name`: Model to use (required variable)
- `_model_provider`: Provider for client (required variable)
- `_reply_way`: "sync" or "stream" (optional, defaults to "sync")

**Outputs**:
- `_llm_response`: Complete response text
- `_llm_status`: "success", "error", or "partial"
- `_llm_tokens_used`: Token count if available

**Error Handling**: Returns error for LLM failures, sets status variables

**Implementation Notes**:
- Uses existing LLMService and ClientFactory
- Handles both streaming and synchronous modes
- **Critical**: No console output - response stored in variables only
- Streaming mode still collects complete response in variable

#### 6. `\_variables-update-history` - Message History Management

**Purpose**: Update message history variables (${1}, ${2}, etc.) from session

**Inputs**:
- `_session_id`: Session to read from (required variable)

**Outputs**:
- Updates `${1}` through `${10}`: Recent messages (${1} = latest assistant response)
- `_history_updated`: "true" if successful
- `_history_count`: Number of messages processed

**Error Handling**: Returns error if session not found

**Implementation Notes**:
- Uses existing VariableService.UpdateMessageHistoryVariables()
- Bypasses normal `\set` restrictions for system variables
- Maintains existing variable numbering convention
- No console output except errors

### Stdlib Scripts (Composable Logic)

#### 1. `session-ensure-active.neuro` - High-level Session Management

**Purpose**: Combines session-get-active with error handling and user feedback

**Usage**: Can be used by other commands needing session setup

**Implementation**:
```neuro
%% High-level session management with user feedback
\_session-get-active
\try[\echo "Error: Failed to get or create session: ${_error}"]
${_session_id!=?:${_last}}

%% Provide user feedback about session status
${_session_status==created?\echo "Created new auto session: ${_session_name}":}
${_session_status==found?\echo "Using active session: ${_session_name}":}
```

#### 2. `model-validate-config.neuro` - Model Configuration Validation

**Purpose**: Validates model config and provides user guidance

**Usage**: Could be reused by other LLM-related commands

**Implementation**:
```neuro
%% Validate model configuration with user guidance
\_model-get-config
\try[\echo "Error: No model configured. Use \\model-new to create one."]
${_model_provider!=?:${_last}}

\_client-setup
${_client_ready!=true?\echo "Warning: LLM client not ready. Check API key for ${_model_provider}":}
```

## Send Command Orchestration

### `send.neuro` - Main Implementation

**Location**: `internal/data/embedded/stdlib/send.neuro`

**Purpose**: Orchestrates all sub-commands to provide complete send functionality

**Implementation**:
```neuro
%% Send command implementation via sub-command orchestration
%% Usage: \send <message>
%% Handles the complete LLM interaction pipeline

%% 1. Validate input
\set[user_message="${_1}"]
${user_message!=?\echo "Error: Message required for send command":}

%% 2. Ensure active session (with user feedback)
\_session-get-active
${_session_id!=?\echo "Error: ${_error}":}

%% 3. Get and validate model configuration
\_model-get-config
${_model_provider!=?\echo "Error: ${_error}":}

%% 4. Setup LLM client
\_client-setup
${_client_ready!=true?\echo "Error: LLM client not configured for ${_model_provider}. Check API key.":}

%% 5. Add user message to session
\_session-add-message[role=user, content="${user_message}"]
${_message_added!=true?\echo "Error: Failed to add user message: ${_error}":}

%% 6. Send LLM request (silent - response stored in _llm_response)
\_llm-send[message="${user_message}"]
${_llm_status!=success?\echo "Error: LLM request failed: ${_error}":}

%% 7. Add assistant response to session
\_session-add-message[role=assistant, content="${_llm_response}"]

%% 8. Update message history variables
\_variables-update-history

%% 9. Display response (using echo for now, render-markdown later)
\echo ${_llm_response}
```

### Auto-Send Integration

#### State Machine Enhancement

**Location**: `internal/statemachine/machine.go`

**Purpose**: Automatically convert non-command input to send commands

**Implementation**:
```go
// In StateMachine.Execute() input preprocessing
func (sm *StateMachine) Execute(input string) error {
    trimmedInput := strings.TrimSpace(input)
    
    // Auto-convert non-command input to send command
    if trimmedInput != "" && !strings.HasPrefix(trimmedInput, "\\") {
        input = "\\send " + input
    }
    
    // Continue with normal state machine processing
    return sm.executeInternal(input)
}
```

**Benefits**:
- Seamless user experience: typing "hello" becomes "\\send hello"
- No visible transformation to user
- Maintains existing command functionality
- Works with all existing NeuroShell features

## Development Strategy

### Phase-by-Phase Implementation (Independent & Testable)

#### Phase 1: Core Infrastructure Commands
1. **Implement `\_session-get-active`** (stateless session management)
   - Create command structure and registration
   - Implement session service integration
   - Add comprehensive unit tests
   - Test error scenarios (no sessions, creation failures)

2. **Implement `\_model-get-config`** (stateless model configuration)
   - Create command with model service integration
   - Handle missing configuration scenarios
   - Add unit tests with mocked services
   - Test various model configuration formats

3. **Implement `\_client-setup`** (stateless client validation)
   - Integrate with ClientFactoryService
   - Handle API key resolution logic
   - Add comprehensive validation tests
   - Test different provider configurations

#### Phase 2: LLM Interaction Commands
4. **Implement `\_session-add-message`** (stateless message persistence)
   - Create message management command
   - Handle both user and assistant messages
   - Add tests for session integration
   - Test error scenarios (invalid sessions, malformed content)

5. **Implement `\_variables-update-history`** (stateless variable management)
   - Implement message history variable updates
   - Ensure proper ${1}-${10} numbering
   - Add tests for variable system integration
   - Test with various session message counts

6. **Implement `\_llm-send`** (stateless LLM interaction)
   - Create core LLM interaction command
   - Handle both sync and streaming modes
   - Ensure silent operation (no console output)
   - Add comprehensive LLM service tests

#### Phase 3: Stdlib Script Composition
7. **Create `session-ensure-active.neuro`** (high-level session management)
   - Implement user-friendly session management
   - Add error handling and user feedback
   - Test script execution through state machine
   - Validate variable flow between commands

8. **Create `model-validate-config.neuro`** (high-level model validation)
   - Implement model configuration validation
   - Add user guidance for missing configurations
   - Test integration with model and client commands
   - Validate error messaging and user experience

#### Phase 4: Send Command Orchestration
9. **Create `send.neuro`** stdlib script (orchestrates all sub-commands)
   - Implement complete send command logic
   - Add comprehensive error handling
   - Test end-to-end LLM interaction
   - Validate variable flow through entire pipeline

10. **Implement auto-send functionality** in state machine
    - Modify input preprocessing logic
    - Ensure seamless non-command input handling
    - Add tests for auto-conversion behavior
    - Validate backward compatibility

#### Phase 5: Integration & Migration
11. **Comprehensive integration testing**
    - Test with real LLM services (controlled environment)
    - Add golden file tests for regression protection
    - Test error scenarios and recovery
    - Performance testing with various session sizes

12. **Replace old send command registration**
    - Remove old send command from registry
    - Update documentation and help text
    - Ensure stdlib script is properly registered
    - Test command resolution priority

13. **Cleanup and documentation**
    - Remove old send command files
    - Update architecture documentation
    - Add user guide for new functionality
    - Create troubleshooting documentation

## Testing Strategy

### Individual Command Testing

#### Unit Tests
- **Service Mocking**: Each sub-command tested with mocked services
- **Variable Isolation**: Test variable setting/getting in isolation
- **Error Scenarios**: Comprehensive error condition testing
- **Silent Operation**: Verify no console output except errors

**Example Test Structure**:
```go
func TestSessionGetActive(t *testing.T) {
    // Test with existing session
    // Test with no session (auto-creation)
    // Test with session service errors
    // Test variable outputs
}
```

#### Integration Tests
- **Service Integration**: Test with real services in controlled environment
- **Variable Flow**: Test data flow between commands via variables
- **Error Propagation**: Test error handling through command chains
- **State Machine Integration**: Test commands through state machine execution

### Script Testing

#### Neuro Script Tests
- **End-to-End Testing**: Test complete send.neuro execution
- **Error Handling**: Test script error scenarios and recovery
- **Variable State**: Test variable state throughout script execution
- **User Feedback**: Test error messages and user guidance

#### Golden Tests
- **Regression Protection**: Record/replay tests for script execution
- **Output Verification**: Verify exact output for various scenarios
- **Error Message Stability**: Ensure consistent error messaging
- **Variable State Snapshots**: Test variable state at key points

### Auto-Send Testing

#### State Machine Tests
- **Input Conversion**: Test non-command input auto-conversion
- **Command Preservation**: Ensure existing commands unaffected
- **Edge Cases**: Test empty input, whitespace, special characters
- **Integration**: Test auto-send with complete LLM pipeline

## Technical Implementation Details

### Command Registration Pattern

```go
// Each internal command follows this pattern
type InternalCommand struct {
    name        string
    description string
    silent      bool  // No console output except errors
}

func (c *InternalCommand) Name() string {
    return c.name
}

func (c *InternalCommand) ParseMode() neurotypes.ParseMode {
    return neurotypes.ParseModeKeyValue
}

func (c *InternalCommand) Execute(args map[string]string, input string) error {
    // Stateless implementation
    // Variable-based input/output
    // Error-only console output
    return nil
}

// Registration in init()
func init() {
    commands.GetGlobalRegistry().Register(&InternalCommand{
        name: "_internal-command",
        description: "Internal command description",
        silent: true,
    })
}
```

### Variable Naming Conventions

#### Input Variables (Read by Commands)
- `_session_id`: Current session identifier
- `_model_provider`: LLM provider name
- `_reply_way`: Response mode ("sync" or "stream")

#### Output Variables (Set by Commands)
- `_session_id`: Session identifier (set by session commands)
- `_llm_response`: Complete LLM response text
- `_client_ready`: Client setup status ("true"/"false")
- `_message_added`: Message addition status ("true"/"false")

#### Status Variables (Error Handling)
- `_session_status`: "found", "created", "error"
- `_llm_status`: "success", "error", "partial"
- `_api_key_status`: "configured", "missing", "invalid"

### Error Handling Strategy

#### Command-Level Errors
- **Return Errors**: Commands return Go errors for system failures
- **Status Variables**: Set status variables for caller decision-making
- **No Fatal Exits**: Commands never exit the shell process

#### Script-Level Errors
- **Try Command**: Use `\try` for graceful error handling
- **Conditional Execution**: Use variable conditions for flow control
- **User Feedback**: Provide clear error messages with guidance

**Example Pattern**:
```neuro
\_session-get-active
${_session_id!=?\echo "Error: ${_error}. Use \\session-new to create a session.":}
```

## Benefits and Rationale

### Modularity Benefits
- **Single Responsibility**: Each command has one clear purpose
- **Independent Testing**: Commands can be tested in complete isolation
- **Reusable Components**: Sub-commands can be used in other contexts
- **Parallel Development**: Teams can work on different commands simultaneously

### State Machine Leverage
- **Execution Engine**: Full utilization of existing robust execution engine
- **Variable System**: Rich inter-command communication mechanism
- **Error Handling**: Existing error propagation and recovery mechanisms
- **Script Support**: Native support for neuro script orchestration

### Maintainability Improvements
- **Clear Separation**: Data processing vs. presentation clearly separated
- **Debugging**: Individual steps can be debugged in isolation
- **Extensibility**: Easy to add new functionality or modify existing steps
- **Documentation**: Each command has clear inputs, outputs, and purpose

### User Experience Enhancements
- **Auto-Send**: Seamless LLM interaction without explicit commands
- **Error Messages**: Clear, actionable error messages with guidance
- **Flexibility**: Users can combine sub-commands for custom workflows
- **Performance**: Efficient variable-based communication

## Future Enhancements

### Multi-Provider Support
- Additional `\_client-setup-*` commands for specific providers
- Provider-specific error handling and configuration
- Dynamic provider discovery and registration

### Advanced Features
- **Streaming Display**: Real-time response rendering with `\render-markdown`
- **Message Threading**: Support for threaded conversations
- **Response Caching**: Cache responses for performance optimization
- **Usage Tracking**: Token usage monitoring and reporting

### Plugin Architecture
- **Custom Sub-Commands**: Allow user-defined internal commands
- **Provider Plugins**: Third-party LLM provider integration
- **Processing Plugins**: Custom message pre/post-processing

## Conclusion

This modular architecture transforms the send command from a monolithic implementation into a composable, maintainable system that fully leverages NeuroShell's strengths:

- **Variable System**: Rich inter-command communication
- **State Machine**: Robust execution and error handling
- **Script Orchestration**: Flexible command composition
- **Testing**: Comprehensive isolation and integration testing

The result is a more reliable, maintainable, and extensible LLM interaction system that provides superior developer and user experiences while preserving all existing functionality.