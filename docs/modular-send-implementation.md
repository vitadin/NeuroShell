# Modular Send Implementation Plan

## Overview

This document outlines the implementation of a modular `\send` command using neuro scripts that orchestrate minimal fundamental Go commands. This approach follows the principle that "golang-based commands should only be fundamental ones" while enabling powerful, maintainable LLM communication workflows.

## Architecture Philosophy

### Core Principle: Command Orchestration via Neuro Scripts

Inspired by the existing LLM architecture design principle: "*commands orchestrate, services execute*". In this model:
- **Neuro scripts orchestrate** the complete workflow
- **Fundamental Go commands execute** atomic operations  
- **No monolithic Go implementations** for complex features

### Data Flow Simplicity

**Key Insight**: The OpenAI Go library provides complete JSON abstraction:
- **Internal Storage**: Go structs (`neurotypes.ChatSession` with `[]neurotypes.Message`)
- **OpenAI Integration**: Library handles JSON serialization internally
- **No JSON manipulation needed** in neuro scripts

## Fundamental Go Commands (Phase 1)

### 1. `\llm-client-get` - Client Management
```neuro
\llm-client-get[provider=openai]
# Uses ${OPENAI_API_KEY} automatically
# Stores client ID in ${_client_id}
```
**Go Implementation:**
- Uses existing ClientFactory pattern from LLM architecture design
- Handles client creation, caching, API key validation
- Lazy initialization following established patterns

### 2. `\llm-call` - Core LLM Request
```neuro
\llm-call[client=${_client_id}, model=gpt-4, session=${session_id}]
# Reads entire ChatSession, makes API call, returns response
# Response stored in ${_output}
```
**Go Implementation:**
- Uses existing LLMService pattern (pure business logic)
- Takes ChatSession struct, converts to OpenAI format internally
- Handles both sync and streaming modes
- Direct OpenAI Go library integration

### 3. `\session-get-active` - Get Active Session
```neuro
\session-get-active
# Returns session ID in ${_session_id}
```
**Implementation:**
- **Already exists** in existing session commands
- Leverages existing ChatSessionService

### 4. `\session-add-usermsg` - Add User Message
```neuro
\session-add-usermsg[session=${session_id}] Hello, how are you?
# Adds user message to session
```
**Go Implementation:**
- Simple wrapper around existing ChatSessionService
- Adds message with role="user" to specified session
- Updates session timestamps

### 5. `\session-add-assistantmsg` - Add Assistant Response
```neuro
\session-add-assistantmsg[session=${session_id}] I'm doing well, thank you!
# Adds assistant response to session
```
**Go Implementation:**
- Simple wrapper around existing ChatSessionService
- Adds message with role="assistant" to specified session
- Updates session timestamps and message history variables

## Neuro Script Implementation (Phase 2)

### Main `send.neuro` Script (stdlib)
```neuro
%% Modular send implementation using fundamental building blocks
%% Usage: \send Hello, how are you?
%% This script orchestrates the complete LLM conversation workflow

%% 1. Get active session
\session-get-active
\set[session_id="${_session_id}"]

%% 2. Add user message to session
\session-add-usermsg[session=${session_id}] ${_1}

%% 3. Get LLM client (uses OPENAI_API_KEY environment variable)
\llm-client-get[provider=openai]
\set[client_id="${_client_id}"]

%% 4. Make LLM call (reads entire session, returns response)
\llm-call[client=${client_id}, model=${llm_model}, session=${session_id}]
\set[response="${_output}"]

%% 5. Add assistant response to session
\session-add-assistantmsg[session=${session_id}] ${response}

%% 6. Update message history variables and display response
\set[1="${response}"]
\render-markdown ${response}
```

### Streaming Version `send-stream.neuro`
```neuro
%% Streaming variant - same workflow with streaming LLM call
%% Usage: \send-stream Hello, how are you?

%% Steps 1-3: Same as send.neuro

%% 4. Make streaming LLM call
\llm-call[client=${client_id}, model=${llm_model}, session=${session_id}, stream=true]
\set[response="${_output}"]

%% 5-6: Same as send.neuro
```

## Implementation Benefits

### 1. **Atomic, Testable Components**
- Each Go command has single responsibility
- Easy to unit test individual operations  
- Clear interfaces and error handling

### 2. **Readable, Maintainable Scripts**
- 6-step workflow is easy to understand
- Professional users can modify individual steps
- No monolithic Go code to maintain

### 3. **Reusable Building Blocks**
- Commands can be used in other LLM workflows
- Support for multi-agent conversations
- Custom LLM interaction patterns

### 4. **Leverages Existing Infrastructure**
- Uses established service patterns (ClientFactory, LLMService, ChatSessionService)
- Compatible with existing session/, render/ commands
- Follows existing variable system conventions

### 5. **Future-Proof Architecture**
- Easy to add new providers (Anthropic via OpenAI-compatible endpoints)
- Support for custom LLM workflows
- Plugin-style extensibility through neuro scripts

## Data Flow Architecture

### Simple Go Struct Pipeline
```
User Input → ChatSession (Go structs) → OpenAI Library → Response → ChatSession → Variables
```

**No JSON manipulation anywhere!**
- OpenAI Go library handles all serialization internally
- All data flows through type-safe Go structs
- Variable system handles string interpolation

### Variable Integration Pattern
- **Input**: `${_1}` for message content
- **Configuration**: `${llm_model}` for model selection
- **Session**: `${session_id}` for conversation context
- **Client**: `${_client_id}` for API client
- **Output**: `${_output}` for LLM response
- **History**: `${1}`, `${2}` for message history

## Implementation Timeline

### Week 1: Fundamental Go Commands
1. **`\llm-client-get`** - Client management using existing ClientFactory
2. **`\llm-call`** - Core API calls using existing LLMService
3. **`\session-add-usermsg`** - User message addition using ChatSessionService
4. **`\session-add-assistantmsg`** - Assistant message addition using ChatSessionService

### Week 2: Neuro Script Implementation
1. Create `stdlib/send.neuro` orchestrator script
2. Create `stdlib/send-stream.neuro` streaming variant
3. Integration testing with real OpenAI API
4. Validate against existing monolithic \send behavior

### Week 3: Testing & Polish
1. Comprehensive unit testing for Go commands
2. End-to-end testing of neuro script workflows
3. Error handling and edge case coverage
4. Performance validation and optimization

### Week 4: Documentation & Migration
1. User documentation and examples
2. Migration guide from monolithic approach
3. Community feedback integration
4. Production readiness validation

## Service Integration Details

### Command Registration Pattern
Following existing patterns in `internal/commands/builtin/`:
```go
type LlmClientGetCommand struct{}

func (c *LlmClientGetCommand) Name() string { return "llm-client-get" }
func (c *LlmClientGetCommand) Description() string { return "Get or create LLM client for provider" }
func (c *LlmClientGetCommand) Execute(args map[string]string, input string) error {
    // Implementation using existing ClientFactory service
}
```

### Service Access Pattern
Using established service registry:
```go
clientFactory, err := services.GetGlobalClientFactoryService()
llmService, err := services.GetGlobalLLMService()
chatSessionService, err := services.GetGlobalChatSessionService()
```

### Variable Result Pattern
Following BashService pattern:
```go
variableService, _ := services.GetGlobalVariableService()
_ = variableService.SetSystemVariable("_client_id", clientID)
_ = variableService.SetSystemVariable("_output", response)
```

## Success Criteria

1. **Complete Functionality Replacement**
   - All existing \send features supported
   - Both synchronous and streaming modes
   - Proper message history management

2. **Improved Maintainability**
   - Modular, single-responsibility components
   - Clear separation between Go and neuro script logic
   - Easy to test and debug individual steps

3. **Enhanced Extensibility**
   - Support for custom LLM workflows
   - Easy addition of new providers
   - Reusable building blocks for other features

4. **Performance Parity**
   - Response times comparable to monolithic implementation
   - Efficient client caching and session management
   - Minimal overhead from script orchestration

5. **Professional User Experience**
   - Transparent, understandable workflow
   - Ability to customize individual steps
   - Rich error messages and debugging information

## Migration Strategy

### Phase 1: Parallel Implementation
- Implement new commands alongside existing (disabled) monolithic code
- Ensure feature parity before switching
- Comprehensive testing in isolated environment

### Phase 2: Gradual Rollout
- Enable new neuro script implementation
- Monitor performance and user feedback
- Keep monolithic code as fallback during transition

### Phase 3: Cleanup
- Remove old monolithic implementation
- Optimize fundamental commands based on usage patterns
- Documentation and community education

## Future Enhancements

### Multi-Provider Support
- Anthropic Claude integration (via OpenAI-compatible endpoints)
- Custom provider scripts using the same building blocks
- Provider-specific optimizations and features

### Advanced Workflows
- Multi-agent conversation support
- Custom prompt templates and processing
- Integration with external tools and APIs

### Performance Optimizations
- Request batching and caching
- Streaming response optimization
- Resource usage monitoring and management

## Conclusion

This modular approach achieves the goal of eliminating monolithic Go implementations while maintaining full functionality and improving maintainability. The fundamental Go commands provide essential building blocks, while neuro scripts enable powerful, customizable workflows that professional users can understand and modify.

The key insight is that **simplicity in individual components enables complexity in composition** - atomic Go commands combined through neuro script orchestration create a flexible, powerful LLM communication system.