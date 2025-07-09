# LLM Architecture Design

## Overview

This document outlines the architectural design for NeuroShell's LLM integration system. The design follows command-orchestrated, dependency-injected principles to achieve better separation of concerns, testability, and extensibility.

## Current Problem

The existing LLM architecture has several issues:
- **Tight coupling**: LLMService directly manages OpenAI client creation
- **Startup dependency**: Application fails to start without `OPENAI_API_KEY` environment variable
- **Service interdependency**: LLMService knows about and calls other services directly
- **Provider lock-in**: Hard to extend to other LLM providers (Anthropic, Claude, etc.)
- **Testing complexity**: Difficult to mock due to hidden dependencies

## Proposed Architecture

### Core Principles

1. **Command Orchestration**: Commands coordinate between services, not services calling other services
2. **Dependency Injection**: All dependencies are explicit parameters, no hidden service calls
3. **Interface Abstraction**: Services work with interfaces, not concrete implementations
4. **Lazy Initialization**: Resources created only when actually needed

### Component Relationships

```
API Key → Client → Multiple Models
   ↓         ↓         ↓
   1:1      1:N      Per Request
```

- **API Key**: Uniquely identifies and authenticates a client
- **Client**: Authenticated connection that can use multiple models
- **Model**: Specified per-request (GPT-4, GPT-3.5-turbo, etc.)

### Architecture Layers

#### 1. Command Layer (Orchestration)
**Responsibility**: Coordinate between services and manage data flow

```go
func (c *SendCommand) Execute(args map[string]string, input string) error {
    // 1. Gather data from various services
    modelConfig := modelService.GetActiveModel()
    session := chatSessionService.GetActiveSession()
    
    // 2. Determine API key (from model config, user config, or env)
    apiKey := c.determineAPIKey(modelConfig)
    
    // 3. Get appropriate client
    client := clientFactory.GetClient(apiKey)
    
    // 4. Call LLM service with all data
    response := llmService.SendCompletion(client, session, modelConfig, input)
    
    // 5. Handle response (update session, variables, etc.)
    chatSessionService.AddMessage(session.ID, "assistant", response)
    variableService.UpdateMessageHistory(session)
    
    return nil
}
```

#### 2. LLM Service (Pure Business Logic)
**Responsibility**: Handle LLM interaction logic without external dependencies

```go
type LLMService interface {
    SendCompletion(client LLMClient, session *ChatSession, model *ModelConfig, message string) (string, error)
    StreamCompletion(client LLMClient, session *ChatSession, model *ModelConfig, message string) (<-chan StreamChunk, error)
}

type LLMServiceImpl struct {
    // No service dependencies - only business logic
}

func (s *LLMServiceImpl) SendCompletion(client LLMClient, session *ChatSession, model *ModelConfig, message string) (string, error) {
    // Pure business logic:
    // - Convert session to API format
    // - Apply model parameters
    // - Call client
    // - Process response
    return client.SendChatCompletion(session, model)
}
```

#### 3. Client Factory Service
**Responsibility**: Create and manage LLM clients based on API keys

```go
type ClientFactory interface {
    GetClient(apiKey string) (LLMClient, error)
    GetClientForProvider(provider, apiKey string) (LLMClient, error)
}

type ClientFactoryImpl struct {
    clients map[string]LLMClient // Cache clients by API key
}

func (f *ClientFactoryImpl) GetClient(apiKey string) (LLMClient, error) {
    if client, exists := f.clients[apiKey]; exists {
        return client, nil
    }
    
    // Create new client (lazy initialization)
    client := NewOpenAIClient(apiKey)
    f.clients[apiKey] = client
    return client, nil
}
```

#### 4. LLM Client Interface
**Responsibility**: Abstract LLM provider implementations

```go
type LLMClient interface {
    SendChatCompletion(session *ChatSession, model *ModelConfig) (string, error)
    StreamChatCompletion(session *ChatSession, model *ModelConfig) (<-chan StreamChunk, error)
    GetProviderName() string
    IsConfigured() bool
}

type OpenAIClient struct {
    client *openai.Client
    apiKey string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
    return &OpenAIClient{
        apiKey: apiKey,
        // client created lazily on first use
    }
}

func (c *OpenAIClient) SendChatCompletion(session *ChatSession, model *ModelConfig) (string, error) {
    if c.client == nil {
        if c.apiKey == "" {
            return "", fmt.Errorf("OpenAI API key not configured")
        }
        c.client = openai.NewClient(option.WithAPIKey(c.apiKey))
    }
    
    // Convert and send request
    // ...
}
```

### Data Flow Example

#### `\send` Command Execution Flow

1. **Input Processing**: Command receives user message
2. **Service Coordination**:
   ```go
   // Get model configuration
   modelConfig := modelService.GetActiveModel()
   
   // Get chat session
   session := chatSessionService.GetActiveSession()
   
   // Add user message to session
   chatSessionService.AddMessage(session.ID, "user", input)
   ```

3. **Client Acquisition**:
   ```go
   // Determine API key source (model config, user config, env var)
   apiKey := c.determineAPIKey(modelConfig)
   
   // Get appropriate client
   client := clientFactory.GetClient(apiKey)
   ```

4. **LLM Interaction**:
   ```go
   // Call LLM service with explicit dependencies
   response := llmService.SendCompletion(client, session, modelConfig, input)
   ```

5. **Result Processing**:
   ```go
   // Update session with response
   chatSessionService.AddMessage(session.ID, "assistant", response)
   
   // Update message history variables
   variableService.UpdateMessageHistory(session)
   ```

### Context Storage Enhancement

Add LLM client storage to NeuroContext:

```go
type NeuroContext struct {
    // ... existing fields ...
    
    // LLM client storage
    llmClients map[string]LLMClient // key = API key identifier
}

// Access methods
func (ctx *NeuroContext) GetLLMClient(apiKey string) (LLMClient, bool)
func (ctx *NeuroContext) SetLLMClient(apiKey string, client LLMClient)
```

## Benefits

### 1. **Separation of Concerns**
- Commands handle orchestration
- Services focus on single responsibilities
- Clear boundaries between layers

### 2. **Improved Testability**
- Easy to mock individual components
- No hidden dependencies
- Pure functions with explicit parameters

### 3. **Extensibility**
- Easy to add new LLM providers (Anthropic, Claude, etc.)
- Provider-specific clients implement common interface
- Model configurations independent of client implementation

### 4. **Configuration Flexibility**
- API keys can come from multiple sources
- No startup dependency on environment variables
- Lazy client creation only when needed

### 5. **Error Handling**
- Clear errors when API keys are missing
- Validation happens at command execution time
- User-friendly error messages with guidance

### 6. **Resource Management**
- Clients cached and reused
- No unnecessary client recreation
- Memory efficient

## Migration Strategy

### Phase 1: Create New Interfaces
1. Define `LLMClient` interface
2. Define refined `LLMService` interface
3. Create `ClientFactory` interface

### Phase 2: Implement Client Layer
1. Implement `OpenAIClient` with lazy initialization
2. Implement `ClientFactory` service
3. Add client storage to NeuroContext

### Phase 3: Refactor LLM Service
1. Remove service dependencies from LLMService
2. Update methods to take explicit parameters
3. Focus on pure business logic

### Phase 4: Update Commands
1. Refactor `\send`, `\send-sync`, `\send-stream` commands
2. Implement orchestration logic in commands
3. Update error handling and user feedback

### Phase 5: Service Registration
1. Register ClientFactory in service registry
2. Remove LLM service initialization dependency on API keys
3. Update application startup flow

## Future Enhancements

### Multi-Provider Support
- Anthropic Claude client implementation
- Google Gemini client implementation
- Provider-specific model catalogs

### Advanced Configuration
- Per-model API key configuration
- Organization-level client management
- Usage tracking and rate limiting

### Plugin Architecture
- Dynamic provider loading
- Custom client implementations
- Third-party LLM integrations

## Conclusion

This architecture provides a solid foundation for scalable, testable, and maintainable LLM integration. It solves the immediate API key startup issue while preparing the codebase for future multi-provider support and advanced features.

The key insight is that **commands orchestrate, services execute** - this principle leads to cleaner code, better testing, and more flexible architecture.