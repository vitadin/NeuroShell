# Modular Send Implementation Plan

## Overview

This document outlines the implementation of a modular `\send` command using neuro scripts that orchestrate minimal fundamental Go commands. This approach follows the principle that "golang-based commands should only be fundamental ones" while enabling powerful, maintainable LLM communication workflows.

## Architecture Philosophy

### Core Principle: Seamless User Experience

The architecture is designed around providing a seamless chat experience where users:

1. **Activate a model first** - Users specify their desired model configuration
2. **Start chatting immediately** - Type `\send` without worrying about setup
3. **Automatic component creation** - System handles client and session creation transparently

### UX Design Philosophy

**Workflow Principle**: "*Model → Chat*"
- Users activate a model configuration (containing provider, model name, parameters)
- Model info determines which client to use automatically
- Session creation is handled automatically when user starts chatting
- No manual client or session management required

### Command Orchestration via Neuro Scripts

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

## Integrated Client Management Architecture

### Stack Service Integration for Seamless UX

Following the established pattern from `if_command.go`, model commands now automatically push `\llm-client-get` commands to the stack service, ensuring clients are created whenever models are activated:

**Implementation Pattern:**
```go
// From if_command.go:121 - Push commands to stack service
if stackService, err := services.GetGlobalStackService(); err == nil {
    stackService.PushCommand("\llm-client-get[provider=" + modelProvider + "]")
}
```

### Enhanced Model Commands with Auto-Client Creation

#### `\model-activate` Enhancement
When users activate a model, the command automatically:
1. Reads model configuration to determine provider
2. Pushes `\llm-client-get[provider=...]` to stack service
3. Sets `#active_model_name` global variable for send.neuro

#### `\model-new` Enhancement  
When users create a new model, the command automatically:
1. Creates model configuration with provider info
2. Pushes `\llm-client-get[provider=...]` to stack service
3. Sets appropriate global variables

This eliminates the need for users to manually manage clients - they simply activate a model and start chatting.

## Fundamental Go Commands (Phase 1)

### 1. `\llm-client-get` - Client Management (Auto-Called)
```neuro
\llm-client-get[provider=openai]
# Uses ${OPENAI_API_KEY} automatically
# Stores client ID in ${_client_id}
# Now called automatically by model commands
```
**Go Implementation:**
- Uses existing ClientFactory pattern from LLM architecture design
- Handles client creation, caching, API key validation
- Lazy initialization following established patterns
- **Client caching is efficient** - multiple calls are lightweight

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

### 3. `\session-activate` - Smart Session Management
```neuro
\session-activate
# Auto-activates existing session or prompts creation
# Returns session ID in ${_session_id}
```
**Implementation:**
- **Already exists** in existing session commands
- Smart logic: shows active session or auto-activates latest
- Returns empty `${_session_id}` when no sessions exist
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

### Simplified `send.neuro` Script (stdlib)
```neuro
%% Simplified send implementation leveraging variable service and auto-client creation
%% Usage: \send Hello, how are you?
%% Assumes user has activated a model first (seamless UX design)

%% 1. Ensure we have an active session (create if needed)
\session-activate
\get[_session_id]
\if[condition="${_session_id}"] \session-new
\if[condition="${_session_id}"] \session-activate

%% 2. Add user message to session
\session-add-usermsg[session=${_session_id}] ${_1}

%% 3. Check for active model (required for seamless UX)
\get[#active_model_name]
\if[condition="${#active_model_name}"] \model-new[catalog_id=O4M] default_model
\if[condition="${#active_model_name}"] \model-activate default_model

%% 4. Get client ID (set automatically by model commands via stack service)
\get[_client_id]

%% 5. Make LLM call using active model configuration
\llm-call[client=${_client_id}, session=${_session_id}]

%% 6. Add assistant response to session and update variables
\session-add-assistantmsg[session=${_session_id}] ${_output}
\set[1="${_output}"]
\render-markdown ${_output}
```

**Key Architecture Improvements:**
- **Leverages Variable Service**: Uses `\get` to check global variables instead of reimplementing logic
- **Avoids Chaining `\if`**: Uses simple binary conditions for better readability
- **Auto-Client Creation**: Relies on model commands to handle client creation via stack service
- **Seamless Session Management**: Creates sessions automatically when needed
- **Default Model Fallback**: Creates default model if none active (better UX)

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

### Phase 1: Enhanced Model Commands (Stack Service Integration)
1. **`\model-activate` Enhancement** - Add auto-client creation via stack service
2. **`\model-new` Enhancement** - Add auto-client creation via stack service
3. **Stack Service Integration** - Follow `if_command.go` pattern for command pushing
4. **Variable Service Integration** - Ensure global variables are properly set

### Phase 2: Simplified Neuro Script Implementation
1. **Rewrite `stdlib/send.neuro`** - Use simplified logic with `\get` commands
2. **Avoid Complex Conditionals** - Replace chained `\if` with simple binary conditions
3. **Integration Testing** - Validate seamless UX workflow
4. **Default Model Handling** - Ensure graceful fallback behavior

### Phase 3: Testing & Validation
1. **Update All Test Cases** - Re-record with new simplified logic
2. **End-to-End Workflow Testing** - Validate complete user journey
3. **Error Handling Coverage** - Test edge cases and failure modes
4. **Performance Validation** - Ensure client caching efficiency

### Phase 4: User Experience Polish
1. **Seamless Workflow Documentation** - Document "Model → Chat" workflow
2. **Error Messages** - Improve user guidance for missing components
3. **Default Configuration** - Optimize out-of-box experience
4. **Integration Validation** - Ensure compatibility with existing commands

## Service Integration Details

### Stack Service Integration Pattern
Following the established pattern from `if_command.go:121-124` for pushing commands to execution queue:

```go
// Enhanced \model-activate command implementation
func (c *ModelActivateCommand) Execute(args map[string]string, input string) error {
    // ... existing model activation logic ...
    
    // Get model configuration to determine provider
    modelConfig, err := modelService.GetModelByNameWithGlobalContext(modelName)
    if err != nil {
        return fmt.Errorf("failed to get model configuration: %w", err)
    }
    
    // Auto-push client creation command to stack service
    if stackService, err := services.GetGlobalStackService(); err == nil {
        clientCommand := fmt.Sprintf("\\llm-client-get[provider=%s]", modelConfig.Provider)
        stackService.PushCommand(clientCommand)
    }
    
    // Set global variables for send.neuro
    variableService.SetSystemVariable("#active_model_name", modelName)
    
    return nil
}
```

### Variable Service Integration Pattern
Leveraging the powerful variable service for state tracking:

```go
// Enhanced variable management in model commands
variableService, _ := services.GetGlobalVariableService()

// Set tracking variables that send.neuro can query
_ = variableService.SetSystemVariable("#active_model_name", modelName)
_ = variableService.SetSystemVariable("#active_model_provider", provider)
_ = variableService.SetSystemVariable("#model_configured", "true")
```

### Service Access Pattern
Using established service registry:
```go
stackService, err := services.GetGlobalStackService()
variableService, err := services.GetGlobalVariableService()
modelService, err := services.GetGlobalModelService()
```

## Success Criteria

1. **Seamless User Experience**
   - **Model → Chat workflow**: Users activate model, then immediately start chatting
   - **Zero manual setup**: No explicit client or session management required
   - **Automatic component creation**: System handles all background setup transparently
   - **Graceful defaults**: Works out-of-box with sensible model and session defaults

2. **Simplified Architecture**
   - **Leverage Variable Service**: Use `\get` commands instead of reimplementing logic
   - **Stack Service Integration**: Model commands auto-create clients via command pushing
   - **Avoid Complex Conditionals**: Replace chained `\if` with simple binary conditions
   - **Service-Based State Management**: Global variables track active components

3. **Enhanced Maintainability**
   - **Clear Service Boundaries**: Each service handles its domain (model, session, client)
   - **Reduced Script Complexity**: Simplified send.neuro with better readability
   - **Integrated Component Management**: Models automatically create matching clients
   - **Consistent Error Handling**: Proper validation at service boundaries

4. **Performance Efficiency**
   - **Client Caching**: Leverage efficient client caching for repeated calls
   - **Minimal Overhead**: Stack service command pushing is lightweight
   - **Service Reuse**: Existing services handle all heavy lifting
   - **Optimized Workflows**: Reduced command chaining and complex logic

5. **Professional Integration**
   - **Existing Command Compatibility**: Works seamlessly with all existing commands
   - **Proper Variable Management**: Integrates with established variable system
   - **Service Pattern Adherence**: Follows established architectural patterns
   - **Future-Proof Design**: Easy to extend with new providers and features

## Migration Strategy

### Phase 1: Enhanced Model Commands
- Implement stack service integration in `\model-activate` and `\model-new`
- Add auto-client creation via command pushing
- Set proper global variables for state tracking
- Test integration with existing commands

### Phase 2: Simplified Send Script
- Rewrite `send.neuro` with simplified logic using `\get` commands
- Remove complex conditional chaining
- Leverage existing service capabilities
- Validate seamless user workflow

### Phase 3: Integration Testing
- Test complete **Model → Chat** workflow end-to-end
- Validate auto-client creation and session management
- Ensure proper error handling and edge case coverage
- Performance testing with client caching efficiency

### Phase 4: Production Rollout
- Deploy enhanced model commands and simplified send script
- Monitor user experience and performance metrics  
- Gather feedback on seamless workflow design
- Document best practices and usage patterns

## Future Enhancements

### Advanced Model Management
- **Model Profiles**: Pre-configured model collections for different use cases
- **Provider Auto-Detection**: Automatic provider determination from model names
- **Model Inheritance**: Base model configurations with customizable overrides
- **Context Length Optimization**: Automatic context management based on model limits

### Enhanced User Experience  
- **Model Recommendations**: Suggest optimal models based on conversation context
- **Session Templates**: Pre-configured session types with custom system prompts
- **Conversation Branching**: Support for parallel conversation threads
- **Multi-Modal Support**: Integration with image and file inputs

### Enterprise Features
- **Team Model Sharing**: Shared model configurations across team members
- **Usage Analytics**: Model and session usage tracking and reporting
- **Cost Optimization**: Automatic model selection based on cost/performance tradeoffs
- **Compliance Monitoring**: Audit trails for model usage and data handling

## Conclusion

This refined modular approach achieves the goal of creating a seamless LLM communication experience while maintaining architectural simplicity and leveraging existing service capabilities. The key innovations include:

### **Seamless UX Design**
The **"Model → Chat"** workflow eliminates manual setup complexity, allowing users to activate a model and immediately start conversing without worrying about clients or sessions.

### **Service Integration Excellence** 
By leveraging the stack service for auto-client creation and the variable service for state management, the architecture avoids reimplementing logic and follows established patterns.

### **Simplified Script Logic**
The new `send.neuro` uses `\get` commands and simple binary conditions instead of complex chaining, making it more readable and maintainable.

### **Efficient Resource Management**
Client caching efficiency and stack service command pushing ensure minimal performance overhead while providing maximum functionality.

The key architectural insight is that **service integration enables seamless user experience** - by having model commands automatically handle client creation and using the variable service for state tracking, we create a system where complexity is hidden from users while maintaining clean, maintainable code structure.

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