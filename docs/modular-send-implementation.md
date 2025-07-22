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

## Component Independence Architecture

### Three Independent Components

The `\llm-call` command operates by synthesizing three completely independent components that do not rely on or know about each other:

1. **Client Component (`${_client_id}`)**
   - Manages API authentication and provider connection
   - Created via `\llm-client-get[provider=openai]`
   - Handles HTTP clients, API keys, rate limiting
   - Provider-agnostic authentication layer

2. **Model Configuration Component (`${model_id}`)**
   - Rich configuration containing model name, version, parameters
   - Includes temperature, max_tokens, stop sequences, model-specific settings
   - Created and managed via `internal/commands/model/` commands
   - Examples: `\model-new`, `\model-list`, `\model-get`, `\model-status`
   - Stores comprehensive model metadata, not just model names

3. **Session Component (`${session_id}`)**
   - Contains conversation history and context
   - Managed via `internal/commands/session/` commands
   - Includes messages, system prompts, timestamps
   - Independent of specific models or clients

### Why `model_id` Instead of Model Names

The `model_id` parameter references a **model configuration** rather than a raw model name because:

- **Rich Metadata**: Model configs contain model name, version, key parameters, stop sequences, and provider-specific settings
- **Reproducibility**: Same model configuration ensures consistent behavior across calls
- **Customization**: Users can create multiple configurations for the same base model with different parameters
- **Management**: Existing `\model-*` commands provide full lifecycle management of these configurations

### Component Synthesis in `\llm-call`

The `\llm-call` command:
1. **Retrieves** the client by `${_client_id}` (authentication layer)
2. **Retrieves** the model configuration by `${model_id}` (parameters and settings)  
3. **Retrieves** the session by `${session_id}` (conversation context)
4. **Synthesizes** these into a unified API call to the LLM provider
5. **Returns** the response without the components knowing about each other

This separation enables maximum flexibility - the same session can use different models, the same model can be used with different clients, and the same client can serve multiple sessions.

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
\llm-call[client=${_client_id}, model_id=${model_id}, session=${session_id}]
# Reads entire ChatSession, makes API call, returns response
# Response stored in ${_output}
```
**Go Implementation:**
- Uses existing LLMService pattern (pure business logic)
- Takes ChatSession struct, converts to OpenAI format internally
- Handles both sync and streaming modes
- Direct OpenAI Go library integration
- **Component Synthesis**: Combines three independent components (client, model config, session) into a unified API call

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

%% 4. Get or use model configuration (assumes ${model_id} is already set)
%% Alternative: \model-get-active to get active model ID
%% Alternative: \set[model_id="my-gpt4-config"] to use specific model config

%% 5. Make LLM call (reads entire session, returns response)
\llm-call[client=${client_id}, model_id=${model_id}, session=${session_id}]
\set[response="${_output}"]

%% 6. Add assistant response to session
\session-add-assistantmsg[session=${session_id}] ${response}

%% 7. Update message history variables and display response
\set[1="${response}"]
\render-markdown ${response}
```

### Streaming Version `send-stream.neuro`
```neuro
%% Streaming variant - same workflow with streaming LLM call
%% Usage: \send-stream Hello, how are you?

%% Steps 1-4: Same as send.neuro (session, user message, client, model config)

%% 5. Make streaming LLM call
\llm-call[client=${client_id}, model_id=${model_id}, session=${session_id}, stream=true]
\set[response="${_output}"]

%% 6-7: Same as send.neuro (add assistant response, update variables)
```

## Implementation Benefits

### 1. **Atomic, Testable Components**
- Each Go command has single responsibility
- Easy to unit test individual operations  
- Clear interfaces and error handling
- **Component Independence**: Client, model config, and session components are completely decoupled

### 2. **Readable, Maintainable Scripts**
- 7-step workflow is easy to understand
- Professional users can modify individual steps
- No monolithic Go code to maintain
- **Model Configuration Flexibility**: Users can switch between different model configs with same base model but different parameters

### 3. **Reusable Building Blocks**
- Commands can be used in other LLM workflows
- Support for multi-agent conversations
- Custom LLM interaction patterns
- **Model Configuration Reuse**: Same model configs can be used across different sessions and workflows

### 4. **Leverages Existing Infrastructure**
- Uses established service patterns (ClientFactory, LLMService, ChatSessionService, ModelService)
- Compatible with existing session/, model/, render/ commands
- Follows existing variable system conventions
- **Model Management Integration**: Full integration with existing `\model-*` command suite

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
- **Configuration**: `${model_id}` for model configuration selection
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

## Model Configuration Examples

### Creating and Using Model Configurations

```neuro
%% Create different model configurations for different use cases
\model-new[name=gpt4-creative] openai gpt-4 temperature=0.9 max_tokens=2000
\model-new[name=gpt4-analytical] openai gpt-4 temperature=0.1 max_tokens=1500
\model-new[name=gpt4-coding] openai gpt-4 temperature=0.2 max_tokens=4000 stop=["\n\n"]

%% Use different configurations in workflows
\set[model_id="gpt4-creative"]
\llm-call[client=${_client_id}, model_id=${model_id}, session=${session_id}] Write a creative story

\set[model_id="gpt4-analytical"]  
\llm-call[client=${_client_id}, model_id=${model_id}, session=${session_id}] Analyze this data

\set[model_id="gpt4-coding"]
\llm-call[client=${_client_id}, model_id=${model_id}, session=${session_id}] Review this code
```

### Component Independence in Practice

```neuro
%% Same session can use different models
\set[session_id="analysis-work"]
\set[client_id="${_openai_client}"]

%% Start with analytical model
\set[model_id="gpt4-analytical"]
\llm-call[client=${client_id}, model_id=${model_id}, session=${session_id}] Analyze trends

%% Switch to creative model for same session
\set[model_id="gpt4-creative"] 
\llm-call[client=${client_id}, model_id=${model_id}, session=${session_id}] Create a visualization idea

%% Same model config can be used across different sessions
\set[session_id="coding-work"]
\set[model_id="gpt4-coding"]
\llm-call[client=${client_id}, model_id=${model_id}, session=${session_id}] Review this function
```

This demonstrates how the three components (client, model config, session) operate independently while being composed by `\llm-call` to create diverse LLM interactions.

## Service Implementation Gaps Analysis

### Current Service Status for `\llm-call` Support

After analyzing the three core services that manage the independent components, the following gaps have been identified:

#### ✅ **LLMService** - COMPLETE
**Status:** Ready for `\llm-call` implementation
- ✅ `SendCompletion(client, session, model, message)` - exactly what `\llm-call` needs
- ✅ `StreamCompletion(client, session, model, message)` - for streaming support
- ✅ Pure business logic design - takes all three components as parameters
- ✅ No service dependencies - perfect for component synthesis

#### ✅ **ChatSessionService** - COMPLETE  
**Status:** Ready for `\llm-call` implementation
- ✅ `GetSessionByNameOrID(nameOrID)` - supports both session names and IDs
- ✅ `GetActiveSession()` - for default session resolution when no session specified
- ✅ Complete session lifecycle management
- ✅ All necessary methods exist for session component retrieval

#### ✅ **ModelService** - COMPLETE (with limitation)
**Status:** Ready for `\llm-call` implementation, minor limitation noted
- ✅ `GetModelByNameWithGlobalContext(name)` - retrieves stored model configurations
- ✅ `GetActiveModelConfigWithGlobalContext()` - provides default model fallback
- ⚠️ **Limitation**: Active model returns synthetic default rather than stored model config
- ⚠️ **Note**: Limitation does not block `\llm-call` functionality

#### ❌ **ClientFactoryService** - CRITICAL GAP
**Status:** Missing essential method for `\llm-call` implementation

**MISSING CRITICAL METHOD**: `GetClientByID(clientID string) (LLMClient, error)`

**Current Methods:**
- ✅ `GetClientWithID(provider, apiKey)` - creates client and returns client ID  
- ✅ `GetClientForProvider(provider, apiKey)` - creates client without ID tracking
- ❌ **Gap**: No method to retrieve existing client using client ID

**The Problem:**
1. `\llm-client-get` command stores client ID in `${_client_id}` system variable
2. `\llm-call` command needs to retrieve the cached client using that client ID
3. **Critical Gap**: No service method exists to get client by its ID

### Required ClientFactoryService Enhancement

#### Client Storage Architecture Decision
**Storage Key Format**: `provider:hashed-api-key` (e.g., `"openai:a1b2c3d4"`)
- **Rationale**: This format serves as both cache key and client ID
- **Benefit**: Enables direct lookup without additional mapping structures
- **Consistency**: Same format used for caching and client identification

#### Required Method Addition
```go
// GetClientByID retrieves a cached LLM client by its client ID.
// Client ID format: "provider:hashed-api-key" (e.g., "openai:a1b2c3d4")
func (f *ClientFactoryService) GetClientByID(clientID string) (neurotypes.LLMClient, error) {
    if !f.initialized {
        return nil, fmt.Errorf("client factory service not initialized")
    }
    
    // Validate client ID format
    if !strings.Contains(clientID, ":") {
        return nil, fmt.Errorf("invalid client ID format: %s", clientID)
    }
    
    // Direct lookup using client ID as cache key
    f.mutex.RLock()
    defer f.mutex.RUnlock()
    
    if client, exists := f.clients[clientID]; exists {
        return client, nil
    }
    
    return nil, fmt.Errorf("client with ID '%s' not found in cache", clientID)
}
```

#### Service Method Integration
The `GetClientByID` method integrates seamlessly with existing architecture:
- **Storage**: Uses same `f.clients[clientID]` map as existing methods
- **Caching**: Leverages existing cache management and thread safety
- **ID Generation**: Compatible with existing `generateClientID()` method
- **Lookup**: Direct O(1) lookup using client ID as map key

### Implementation Priority
1. **Critical**: Add `GetClientByID` method to ClientFactoryService
2. **Optional**: Enhance ModelService to support stored default models (future improvement)

This single method addition completes the service layer requirements for `\llm-call` implementation, enabling the command to retrieve all three components (client, model config, session) and synthesize them into unified LLM API calls.