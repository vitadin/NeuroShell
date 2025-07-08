# API Service Refactoring Plan - Three-Layer Architecture Compliant

## Overview

This document outlines the refactoring plan for NeuroShell's API service to transition from a switch-case based provider routing design to a clean, maintainable interface-based architecture that **strictly adheres to the three-layer architecture** defined in CLAUDE.md.

## Current Architecture Violations

### 1. Direct Global Context Access

The current `APIService` violates the three-layer architecture by directly accessing the global context:

```go
// services/api_service.go:155
ctx := neuroshellcontext.GetGlobalContext()
```

**Violation**: Services should only interact with the context passed during initialization, not access global state.

### 2. Direct OS Environment Access

The service directly accesses OS environment variables:

```go
// services/api_service.go:109
if timeoutEnv := os.Getenv("API_TIMEOUT"); timeoutEnv != "" {
```

**Violation**: Only the Context Layer should interface with the OS. Services should receive configuration through the context.

### 3. State Storage in Service Layer

The service maintains internal state that belongs in the Context Layer:

```go
type APIService struct {
    initialized  bool
    httpClient   *http.Client
    timeout      time.Duration
    endpoints    map[string]string
    openaiClient *openai.Client
}
```

**Violation**: Services should be stateless. All state should be managed by the Context Layer.

### 4. Switch-Case Anti-Pattern

The current implementation uses switch-case for provider routing, violating the Open/Closed Principle.

## Three-Layer Architecture Principles

According to CLAUDE.md, NeuroShell follows a strict three-layer architecture:

```
┌──────────────────────────────────┐
│         Command Layer            │
│  (\send, \set, \get, etc.)      │
├──────────────────────────────────┤
│        Service Layer             │
│  (Stateless business logic)      │
├──────────────────────────────────┤
│        Context Layer             │
│  (All state and resources)       │
└──────────────────────────────────┘
```

**Key Rules**:
1. **Context Layer**: Holds ALL state and resources
2. **Service Layer**: Stateless business logic only
3. **Command Layer**: Orchestrates services
4. **No Cross-Layer Dependencies**: Layers can only interact through defined interfaces
5. **Services Don't Know About Commands**: Services are independent
6. **Commands Don't Access Context**: Commands only use services

## Proposed Architecture

### Context Layer Extensions

Extend the Context interface to support API provider management:

```go
// Context interface extensions for API provider support
type Context interface {
    // Existing methods...
    
    // Provider Configuration Management
    GetProviderConfig(provider string) (*ProviderConfig, error)
    SetProviderConfig(provider string, config *ProviderConfig) error
    ListProviderConfigs() map[string]*ProviderConfig
    
    // Provider Client Management (cached instances)
    GetProviderClient(provider string) (interface{}, error)
    SetProviderClient(provider string, client interface{}) error
    ClearProviderClient(provider string) error
    
    // Factory Registry Management
    RegisterProviderFactory(provider string, factory ClientFactory) error
    GetProviderFactory(provider string) (ClientFactory, error)
    ListProviderFactories() []string
    
    // API-related Environment Access (Context handles OS interface)
    GetAPIEnvironmentVariable(key string) (string, bool)
    GetAPITimeout() time.Duration
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

### Core Interfaces (Keep These)

These interfaces are well-designed and should be retained:

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

// Standard response format
type APIResponse struct {
    Content     string
    Provider    string
    Model       string
    TokensUsed  int
    Metadata    map[string]interface{}
}
```

### Service Layer Design (Stateless)

The APIService becomes a pure orchestration layer with no state:

```go
// APIService - stateless orchestration of API operations
type APIService struct {
    // No state! Pure business logic only
}

// Initialize registers default provider factories with the context
func (a *APIService) Initialize(ctx Context) error {
    // Register built-in provider factories
    if err := ctx.RegisterProviderFactory("openai", &OpenAIClientFactory{}); err != nil {
        return fmt.Errorf("failed to register OpenAI factory: %w", err)
    }
    
    // Future providers
    // ctx.RegisterProviderFactory("anthropic", &AnthropicClientFactory{})
    
    return nil
}

// SendMessage orchestrates sending a message through the appropriate provider
func (a *APIService) SendMessage(ctx Context, provider, model, message string, options map[string]any) (*APIResponse, error) {
    // Get or create provider client
    client, err := a.getOrCreateClient(ctx, provider)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider client: %w", err)
    }
    
    // Delegate to provider
    response, err := client.SendMessage(model, message, options)
    if err != nil {
        return nil, fmt.Errorf("provider error: %w", err)
    }
    
    // Update context with usage metrics
    if response.TokensUsed > 0 {
        ctx.SetVariable("#tokens_used", fmt.Sprintf("%d", response.TokensUsed))
    }
    
    return response, nil
}

// ListModels lists available models for a provider
func (a *APIService) ListModels(ctx Context, provider string) ([]ModelInfo, error) {
    client, err := a.getOrCreateClient(ctx, provider)
    if err != nil {
        return nil, err
    }
    
    return client.ListModels()
}

// CheckConnectivity tests provider connectivity
func (a *APIService) CheckConnectivity(ctx Context, provider string) error {
    client, err := a.getOrCreateClient(ctx, provider)
    if err != nil {
        return err
    }
    
    return client.CheckConnectivity()
}

// getOrCreateClient is a private helper for client management
func (a *APIService) getOrCreateClient(ctx Context, provider string) (LLMProvider, error) {
    // Try to get cached client from context
    if cached, err := ctx.GetProviderClient(provider); err == nil && cached != nil {
        if client, ok := cached.(LLMProvider); ok {
            return client, nil
        }
    }
    
    // Get factory from context
    factory, err := ctx.GetProviderFactory(provider)
    if err != nil {
        return nil, fmt.Errorf("no factory registered for provider %s: %w", provider, err)
    }
    
    // Get or create configuration
    config, err := a.getProviderConfig(ctx, provider)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider config: %w", err)
    }
    
    // Create new client
    client, err := factory.CreateClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }
    
    // Cache in context for reuse
    if err := ctx.SetProviderClient(provider, client); err != nil {
        // Log warning but continue - caching failure shouldn't block operation
        fmt.Printf("Warning: failed to cache provider client: %v\n", err)
    }
    
    return client, nil
}

// getProviderConfig assembles configuration from context
func (a *APIService) getProviderConfig(ctx Context, provider string) (*ProviderConfig, error) {
    // Check if explicit config exists
    if config, err := ctx.GetProviderConfig(provider); err == nil && config != nil {
        return config, nil
    }
    
    // Build config from context variables and environment
    config := &ProviderConfig{
        Provider: provider,
        Timeout:  ctx.GetAPITimeout(),
    }
    
    // Get API key (check context variable first, then environment)
    varName := fmt.Sprintf("%s_api_key", provider)
    if apiKey, err := ctx.GetVariable(varName); err == nil && apiKey != "" {
        config.APIKey = apiKey
    } else {
        envVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
        if apiKey, found := ctx.GetAPIEnvironmentVariable(envVar); found {
            config.APIKey = apiKey
        }
    }
    
    // Get base URL if needed
    if baseURL, err := ctx.GetVariable(fmt.Sprintf("%s_base_url", provider)); err == nil && baseURL != "" {
        config.BaseURL = baseURL
    }
    
    return config, nil
}
```

### Provider Implementations

Example OpenAI provider implementation:

```go
// OpenAIProvider implements LLMProvider for OpenAI
type OpenAIProvider struct {
    client *openai.Client
    config *ProviderConfig
}

func (o *OpenAIProvider) Name() string {
    return "openai"
}

func (o *OpenAIProvider) SendMessage(model, message string, options map[string]any) (*APIResponse, error) {
    // Implementation using OpenAI SDK
    req := openai.ChatCompletionRequest{
        Model: model,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleUser, Content: message},
        },
    }
    
    resp, err := o.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    return &APIResponse{
        Content:    resp.Choices[0].Message.Content,
        Provider:   "openai",
        Model:      model,
        TokensUsed: resp.Usage.TotalTokens,
    }, nil
}

// OpenAIClientFactory creates OpenAI provider instances
type OpenAIClientFactory struct{}

func (f *OpenAIClientFactory) SupportedProvider() string {
    return "openai"
}

func (f *OpenAIClientFactory) CreateClient(config *ProviderConfig) (LLMProvider, error) {
    if config.APIKey == "" {
        return nil, fmt.Errorf("OpenAI API key is required")
    }
    
    clientConfig := openai.DefaultConfig(config.APIKey)
    if config.BaseURL != "" {
        clientConfig.BaseURL = config.BaseURL
    }
    
    return &OpenAIProvider{
        client: openai.NewClientWithConfig(clientConfig),
        config: config,
    }, nil
}

func (f *OpenAIClientFactory) ValidateConfig(config *ProviderConfig) error {
    if config.APIKey == "" {
        return fmt.Errorf("API key is required")
    }
    return nil
}
```

### Command Layer Integration

Example of how commands use the refactored APIService:

```go
// SendCommand demonstrates proper three-layer interaction
type SendCommand struct{}

func (c *SendCommand) Execute(args []string, input string, services ServiceRegistry) error {
    // Get services from registry
    apiService := services.GetService("api").(*APIService)
    variableService := services.GetService("variable").(*VariableService)
    
    // Parse arguments to get provider and model
    provider := "openai" // default or from args
    model := "gpt-4"     // default or from args
    
    // Get context from service registry (passed during command execution)
    ctx := services.GetContext()
    
    // Send message through API service
    response, err := apiService.SendMessage(ctx, provider, model, input, nil)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    
    // Store response in context variables
    if err := variableService.Set(ctx, "1", response.Content); err != nil {
        return fmt.Errorf("failed to store response: %w", err)
    }
    
    // Display response
    fmt.Println(response.Content)
    
    return nil
}
```

## Implementation Roadmap

### Phase 1: Context Layer Extensions (2-3 days)

1. Extend Context interface with API provider methods
2. Implement provider configuration storage
3. Implement provider client caching
4. Implement factory registry
5. Add environment variable access through context

### Phase 2: Core Interfaces and Providers (3-4 days)

1. Define `LLMProvider` and `ClientFactory` interfaces
2. Implement `OpenAIProvider` and `OpenAIClientFactory`
3. Create provider-specific error types
4. Add comprehensive unit tests

### Phase 3: APIService Refactoring (2-3 days)

1. Remove all state from APIService
2. Implement stateless orchestration methods
3. Remove direct context and environment access
4. Update all service methods to receive context as parameter
5. Add service tests with mocked context

### Phase 4: Command Integration (1-2 days)

1. Update commands to pass context to service methods
2. Remove any direct context access from commands
3. Update command tests

### Phase 5: Migration and Testing (2-3 days)

1. Migrate existing configurations
2. Ensure backward compatibility
3. Run integration tests
4. Performance testing
5. Documentation updates

## Benefits of This Architecture

### 1. **Strict Layer Separation**
- Context owns all state
- Services are pure business logic
- Commands orchestrate without knowing implementation details

### 2. **Testability**
- Services can be tested with mock contexts
- No global state dependencies
- Clear interfaces for mocking

### 3. **Maintainability**
- Adding providers doesn't modify existing code
- Clear responsibilities for each layer
- Consistent with other NeuroShell services

### 4. **Extensibility**
- New providers via factory registration
- New features without architectural changes
- Plugin-ready architecture

## Migration Strategy

### Step 1: Implement Context Extensions
- Add new methods to Context interface
- Implement in existing context implementation
- No breaking changes

### Step 2: Create New APIService
- Implement as `APIServiceV2` initially
- Run parallel with existing service
- Gradual migration of commands

### Step 3: Provider Migration
- Start with OpenAI (most used)
- Add Anthropic and others incrementally
- Each provider is independent

### Step 4: Deprecate Old Service
- Mark old APIService as deprecated
- Provide migration guide
- Remove after transition period

## Example: Adding a New Provider

With this architecture, adding a new provider is simple:

```go
// 1. Implement the provider
type AnthropicProvider struct {
    client *anthropic.Client
    config *ProviderConfig
}

func (a *AnthropicProvider) SendMessage(model, message string, options map[string]any) (*APIResponse, error) {
    // Anthropic-specific implementation
}

// 2. Implement the factory
type AnthropicClientFactory struct{}

func (f *AnthropicClientFactory) CreateClient(config *ProviderConfig) (LLMProvider, error) {
    // Create Anthropic client
}

// 3. Register in APIService.Initialize()
ctx.RegisterProviderFactory("anthropic", &AnthropicClientFactory{})
```

No other code changes required!

## Conclusion

This refactoring plan transforms the API service to strictly follow NeuroShell's three-layer architecture while maintaining all the benefits of the interface-based design. The key improvements are:

1. **Complete removal of state from services**
2. **All configuration through Context Layer**
3. **No cross-layer violations**
4. **Consistent with existing service patterns**
5. **Maintains extensibility and testability**

The architecture ensures that each layer has clear, focused responsibilities, making the system more maintainable and easier to understand while providing a solid foundation for future enhancements.