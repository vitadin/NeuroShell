# API Service Refactoring Plan

## Overview

This document outlines the refactoring plan for NeuroShell's API service layer to transition from a switch-case based provider routing design to a clean, maintainable interface-based architecture.

## Current Problems

### 1. Switch-Case Anti-Pattern

The current `APIService` uses switch-case statements for provider routing:

```go
func (a *APIService) SendMessage(provider, model, message string, options map[string]any) (*APIResponse, error) {
    switch strings.ToLower(provider) {
    case "openai":
        return a.sendOpenAIMessage(model, message, options)
    case "anthropic":
        return nil, fmt.Errorf("anthropic provider not yet implemented")
    default:
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }
}
```

**Problems:**
- Violates Open/Closed Principle - adding providers requires modifying existing code
- Creates tight coupling between APIService and provider implementations
- Makes the service increasingly complex as providers are added
- Hard to unit test individual provider logic

### 2. Single Responsibility Violation

The current `APIService` handles multiple responsibilities:
- HTTP client management and connectivity checking
- Provider-specific client initialization (OpenAI, Anthropic)
- Message sending logic for each provider
- Model listing for each provider
- Response format standardization

This violates the Single Responsibility Principle and makes the service difficult to maintain.

### 3. Maintenance and Extensibility Issues

- **Code Duplication**: Similar patterns repeated for each provider method
- **Hard to Test**: Provider-specific logic is embedded in the main service
- **Configuration Complexity**: Different providers have different configuration needs
- **Scalability Problems**: Adding features requires touching the main service class

## Proposed Architecture

### Core Design Principles

1. **Interface Segregation**: Clear, focused interfaces for different responsibilities
2. **Dependency Injection**: Services depend on interfaces, not implementations
3. **Open/Closed Principle**: Easy to add providers without modifying existing code
4. **Single Responsibility**: Each service has one clear purpose

### Interface Definitions

```go
// LLMProvider defines the contract for all LLM provider implementations
type LLMProvider interface {
    // Provider identification
    Name() string
    
    // Core messaging operations
    SendMessage(model, message string, options map[string]any) (*APIResponse, error)
    SendMessageStreaming(model, message string, options map[string]any, callback func(string)) error
    
    // Model operations
    ListModels() ([]ModelInfo, error)
    
    // Health and connectivity
    CheckConnectivity() error
}

// ClientFactory creates and configures provider clients
type ClientFactory interface {
    CreateClient(config *ProviderConfig) (LLMProvider, error)
    SupportedProvider() string
    ValidateConfig(config *ProviderConfig) error
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
    Provider    string
    APIKey      string
    BaseURL     string
    Timeout     time.Duration
    Parameters  map[string]interface{}
}
```

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      APIService                            │
│  • High-level API coordination                             │
│  • Request routing via ProviderRegistryService             │
│  • Response standardization                                │
│  • Backward compatibility layer                            │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│               ProviderRegistryService                       │
│  • Provider factory registration and discovery             │
│  • Provider instance management and caching                │
│  • Thread-safe provider lookup                             │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                ClientManagerService                        │
│  • Provider client creation and lifecycle                  │
│  • Configuration management (context/environment)          │
│  • Client caching and connection pooling                   │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                   Provider Implementations                 │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │  OpenAIProvider │  │AnthropicProvider│  ...             │
│  │  (LLMProvider)  │  │  (LLMProvider)  │                  │
│  └─────────────────┘  └─────────────────┘                  │
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │OpenAIClientFact.│  │AnthropicClientF.│  ...             │
│  │ (ClientFactory) │  │ (ClientFactory) │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Roadmap

### Phase 1: Foundation (Core Interfaces)

**Deliverables:**
- Create `internal/services/providers/` package
- Define `LLMProvider` interface
- Define `ClientFactory` interface
- Define `ProviderConfig` struct
- Create basic error types for provider operations

**Estimated Effort:** 1-2 days

### Phase 2: Provider Registry Service

**Deliverables:**
- Create `ProviderRegistryService`
- Implement factory registration and discovery
- Implement provider instance management
- Add thread-safe provider lookup
- Create provider registry tests

**Key Features:**
```go
type ProviderRegistryService struct {
    factories map[string]ClientFactory
    instances map[string]LLMProvider
    mutex     sync.RWMutex
}

func (p *ProviderRegistryService) RegisterFactory(factory ClientFactory) error
func (p *ProviderRegistryService) GetProvider(providerName string) (LLMProvider, error)
func (p *ProviderRegistryService) ListProviders() []string
```

**Estimated Effort:** 2-3 days

### Phase 3: Client Manager Service

**Deliverables:**
- Create `ClientManagerService`
- Implement configuration management from context/environment
- Implement client caching and lifecycle management
- Handle provider-specific configuration needs
- Create client manager tests

**Key Features:**
```go
type ClientManagerService struct {
    registry     *ProviderRegistryService
    clients      map[string]LLMProvider
    configs      map[string]*ProviderConfig
    mutex        sync.RWMutex
}

func (c *ClientManagerService) GetClient(provider string) (LLMProvider, error)
func (c *ClientManagerService) UpdateConfig(provider string, config *ProviderConfig) error
func (c *ClientManagerService) RefreshClient(provider string) error
```

**Estimated Effort:** 2-3 days

### Phase 4: Provider Implementations

**Deliverables:**
- Create `OpenAIProvider` implementing `LLMProvider`
- Create `OpenAIClientFactory` implementing `ClientFactory`
- Extract and refactor existing OpenAI logic
- Implement proper error handling and logging
- Create comprehensive provider tests

**Implementation Example:**
```go
type OpenAIProvider struct {
    client  *openai.Client
    config  *ProviderConfig
}

func (o *OpenAIProvider) SendMessage(model, message string, options map[string]any) (*APIResponse, error) {
    // Clean OpenAI-specific implementation
}

type OpenAIClientFactory struct{}

func (f *OpenAIClientFactory) CreateClient(config *ProviderConfig) (LLMProvider, error) {
    // Factory logic for OpenAI client creation
}
```

**Estimated Effort:** 3-4 days

### Phase 5: APIService Refactoring

**Deliverables:**
- Refactor `APIService` to use provider registry
- Remove all switch-case provider routing
- Implement provider delegation pattern
- Maintain backward compatibility
- Update APIService tests

**New APIService Implementation:**
```go
type APIService struct {
    initialized    bool
    clientManager  *ClientManagerService
    httpClient     *http.Client  // Keep for connectivity checks
    timeout        time.Duration
}

func (a *APIService) SendMessage(provider, model, message string, options map[string]any) (*APIResponse, error) {
    if !a.initialized {
        return nil, fmt.Errorf("api service not initialized")
    }
    
    client, err := a.clientManager.GetClient(provider)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider client: %w", err)
    }
    
    return client.SendMessage(model, message, options)
}
```

**Estimated Effort:** 2-3 days

### Phase 6: Integration and Testing

**Deliverables:**
- Update service registry to include new services
- Ensure all existing tests pass
- Add integration tests for provider interactions
- Performance testing and optimization
- Documentation updates

**Service Registration:**
```go
// In registry initialization
registry.RegisterService(NewProviderRegistryService())
registry.RegisterService(NewClientManagerService())

// Register provider factories
providerRegistry.RegisterFactory(&OpenAIClientFactory{})
// Future: providerRegistry.RegisterFactory(&AnthropicClientFactory{})
```

**Estimated Effort:** 2-3 days

## Migration Strategy

### Backward Compatibility

The refactoring maintains full backward compatibility:

1. **Public API Unchanged**: All existing `APIService` public methods remain the same
2. **Service Registration**: Existing service registry patterns continue to work
3. **Configuration**: Current context/environment variable patterns preserved
4. **Error Handling**: Error messages and types remain consistent

### Gradual Migration Approach

1. **Phase-by-Phase Implementation**: Each phase can be implemented and tested independently
2. **Feature Flags**: New architecture can be enabled/disabled during development
3. **Parallel Testing**: Both old and new implementations can run side-by-side during testing
4. **Incremental Deployment**: Providers can be migrated one at a time

### Testing Strategy

1. **Unit Tests**: Each service and provider tested in isolation
2. **Integration Tests**: Full provider workflow testing
3. **Regression Tests**: Ensure all existing functionality continues to work
4. **Performance Tests**: Verify no performance degradation

## Benefits and Trade-offs

### Benefits

1. **Maintainability**: Clean separation of concerns, easy to understand and modify
2. **Extensibility**: Adding new providers requires no changes to existing code
3. **Testability**: Easy to mock and test individual components
4. **Scalability**: Architecture scales well with new providers and features
5. **Code Quality**: Eliminates switch-case anti-patterns and reduces duplication

### Trade-offs

1. **Initial Complexity**: More moving parts in the initial implementation
2. **Learning Curve**: Developers need to understand the new architecture
3. **Development Time**: Significant upfront investment (estimated 12-18 days)
4. **Memory Overhead**: Slightly more memory usage due to additional service layers

### Performance Considerations

1. **Provider Caching**: Clients are cached to avoid initialization overhead
2. **Registry Lookup**: Fast O(1) provider lookup via hash map
3. **Minimal Overhead**: Interface-based dispatch has negligible performance impact
4. **Connection Pooling**: Provider implementations can optimize connections independently

## Code Examples

### Example: Adding a New Provider

```go
// 1. Implement the provider
type ClaudeProvider struct {
    client *anthropic.Client
    config *ProviderConfig
}

func (c *ClaudeProvider) Name() string { return "anthropic" }

func (c *ClaudeProvider) SendMessage(model, message string, options map[string]any) (*APIResponse, error) {
    // Anthropic-specific implementation
}

// 2. Implement the factory
type ClaudeClientFactory struct{}

func (f *ClaudeClientFactory) CreateClient(config *ProviderConfig) (LLMProvider, error) {
    // Create and configure Anthropic client
}

func (f *ClaudeClientFactory) SupportedProvider() string { return "anthropic" }

// 3. Register the factory (in initialization code)
providerRegistry.RegisterFactory(&ClaudeClientFactory{})
```

### Example: Using the New Architecture

```go
// Commands can use the same APIService interface
apiService, err := services.GetGlobalAPIService()
if err != nil {
    return err
}

// Provider routing is handled automatically
response, err := apiService.SendMessage("openai", "gpt-4", "Hello!", nil)
if err != nil {
    return err
}

// The implementation automatically:
// 1. Looks up the provider via registry
// 2. Gets/creates the appropriate client
// 3. Delegates to the provider implementation
// 4. Returns standardized response
```

## Future Considerations

### Planned Extensions

1. **Provider Plugins**: Dynamic loading of provider implementations
2. **Circuit Breakers**: Automatic failover and retry logic
3. **Metrics Collection**: Per-provider performance and usage metrics
4. **Rate Limiting**: Provider-specific rate limiting and throttling
5. **Provider Chaining**: Fallback providers for reliability

### Technology Evolution

1. **WebAssembly Providers**: Support for WASM-based provider implementations
2. **gRPC Support**: Provider implementations using gRPC for better performance
3. **Event-Driven Architecture**: Async messaging patterns for streaming responses
4. **Configuration Management**: External configuration management integration

## Conclusion

This refactoring transforms the API service from a monolithic, switch-based design to a clean, extensible, interface-driven architecture. While requiring significant upfront investment, the benefits in maintainability, extensibility, and code quality make this a worthwhile improvement for the long-term evolution of NeuroShell.

The phased implementation approach allows for gradual migration while maintaining backward compatibility and reducing implementation risk.