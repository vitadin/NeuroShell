# Configuration Service Integration with Client Factory

## Overview

This document outlines a comprehensive refactoring plan to integrate the Configuration Service with the Client Factory Service while maintaining proper three-layer architecture and leveraging NeuroShell's powerful variable system.

## Current Problems

### 1. Hard-coded Configuration in ClientFactory

The ClientFactory service contains hard-coded values that should be configurable:

```go
// Hard-coded base URLs and headers
case "anthropic":
    config := OpenAICompatibleConfig{
        BaseURL: "https://api.anthropic.com/v1",
        Headers: map[string]string{"anthropic-version": "2023-06-01"},
    }
case "gemini":
    client = NewGeminiClient(apiKey)
```

### 2. Architecture Violations

- **Service-to-service dependencies**: ClientFactory should not determine API keys
- **Duplicate logic**: `GetClientForProvider` and `GetClientWithID` have identical switch statements
- **Mixed responsibilities**: ClientFactory both configures and creates clients

### 3. Limited Configurability

- Base URLs cannot be overridden via environment variables
- Provider-specific headers are hard-coded
- No way to add custom providers without code changes

## Solution Architecture

### Core Principle: Command-Orchestrated Configuration

```
┌─────────────────────────────────────────────────────────────┐
│                     Command Layer                           │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │ \llm-client-get │ ───┤ 1. Get provider config          │ │
│  │    Command      │    │ 2. Get API key                  │ │
│  │                 │    │ 3. Create client with config    │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
           │                          │
           ▼                          ▼
┌─────────────────┐         ┌─────────────────┐
│ Configuration   │         │ ClientFactory   │
│ Service         │         │ Service         │
│                 │         │                 │
│ • Get provider  │         │ • Create client │
│   base URL      │         │   with config   │
│ • Get headers   │         │ • Cache client  │
│ • Get API key   │         │ • Return ID     │
└─────────────────┘         └─────────────────┘
           │                          │
           ▼                          ▼
┌─────────────────────────────────────────────────────────────┐
│                     Context Layer                           │
│ • Configuration map  • Environment variables               │
│ • Client cache      • File operations                      │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Plan

### Phase 1: Configuration Service Enhancement

#### 1.1 Add Provider Configuration Methods

```go
// Provider configuration structure
type ProviderConfig struct {
    Name      string
    BaseURL   string
    Headers   map[string]string
    Endpoint  string
    APIKeyEnv string
}

// New methods in ConfigurationService
func (c *ConfigurationService) GetProviderConfig(provider string) (*ProviderConfig, error)
func (c *ConfigurationService) GetProviderBaseURL(provider string) (string, error)
func (c *ConfigurationService) GetProviderHeaders(provider string) (map[string]string, error)
func (c *ConfigurationService) GetProviderAPIKey(provider string) (string, error)
```

#### 1.2 Default Provider Configurations

Add to Context layer's `LoadDefaults()`:

```go
defaults := map[string]string{
    // OpenAI
    "NEURO_OPENAI_BASE_URL":    "https://api.openai.com/v1",
    "NEURO_OPENAI_ENDPOINT":    "/chat/completions",
    
    // Anthropic
    "NEURO_ANTHROPIC_BASE_URL":  "https://api.anthropic.com/v1",
    "NEURO_ANTHROPIC_ENDPOINT":  "/messages",
    "NEURO_ANTHROPIC_HEADERS":   "anthropic-version=2023-06-01",
    
    // Gemini
    "NEURO_GEMINI_BASE_URL":     "https://generativelanguage.googleapis.com/v1beta",
    "NEURO_GEMINI_ENDPOINT":     "/models",
}
```

#### 1.3 Configuration Overrides

Support environment variable overrides:

```bash
# Override base URLs
export NEURO_OPENAI_BASE_URL="https://my-proxy.com/v1"
export NEURO_GEMINI_BASE_URL="https://custom-gemini.com/v1"

# Override headers
export NEURO_ANTHROPIC_HEADERS="Custom-Header=value,Another=value2"
```

### Phase 2: ClientFactory Refactoring

#### 2.1 Remove Hard-coded Configuration

Replace provider switch statements with configuration-driven approach:

```go
// Old: Hard-coded switch statements (REMOVE)
switch provider {
case "openai":
    client = NewOpenAIClient(apiKey)
case "anthropic":
    client = NewAnthropicClient(apiKey)
case "gemini":
    client = NewGeminiClient(apiKey)
// ...
}

// New: Configuration-driven approach
func (f *ClientFactoryService) GetClientWithConfig(
    provider, apiKey, baseURL string, 
    headers map[string]string, 
    endpoint string,
) (neurotypes.LLMClient, string, error)
```

#### 2.2 Unified Client Creation

Consolidate duplicate logic:

```go
// Single method replaces GetClientForProvider and GetClientWithID
func (f *ClientFactoryService) CreateClient(config ClientConfig) (neurotypes.LLMClient, string, error) {
    clientID := f.generateClientID(config.Provider, config.APIKey)
    
    // Check cache
    if client, exists := ctx.GetLLMClient(clientID); exists {
        return client, clientID, nil
    }
    
    // Create client based on provider type
    var client neurotypes.LLMClient
    switch config.ProviderType {
    case "openai":
        client = NewOpenAIClient(config.APIKey)
    case "anthropic":
        client = NewAnthropicClient(config.APIKey)
    case "gemini":
        client = NewGeminiClient(config.APIKey)
    default:
        return nil, "", fmt.Errorf("unsupported provider type: %s", config.ProviderType)
    }
    
    // Cache and return
    ctx.SetLLMClient(clientID, client)
    return client, clientID, nil
}
```

### Phase 3: Command Integration

#### 3.1 Enhanced llm-client-get Command

```go
func (c *LLMClientGetCommand) Execute(args map[string]string, _ string) error {
    provider := args["provider"]
    if provider == "" {
        provider = "openai"
    }
    
    // Get configuration service
    configService, err := services.GetGlobalConfigurationService()
    if err != nil {
        return fmt.Errorf("configuration service not available: %w", err)
    }
    
    // Get provider configuration
    providerConfig, err := configService.GetProviderConfig(provider)
    if err != nil {
        return fmt.Errorf("failed to get provider config: %w", err)
    }
    
    // Get API key (args override config/env)
    apiKey := args["key"]
    if apiKey == "" {
        apiKey, err = configService.GetProviderAPIKey(provider)
        if err != nil {
            return fmt.Errorf("API key not found: %w", err)
        }
    }
    
    // Allow base URL override from args
    baseURL := args["base_url"]
    if baseURL == "" {
        baseURL = providerConfig.BaseURL
    }
    
    // Get client factory service
    clientFactory, err := services.GetGlobalClientFactoryService()
    if err != nil {
        return fmt.Errorf("client factory service not available: %w", err)
    }
    
    // Create client config
    clientConfig := ClientConfig{
        Provider:     provider,
        ProviderType: providerConfig.Type,
        APIKey:       apiKey,
        BaseURL:      baseURL,
        Headers:      providerConfig.Headers,
        Endpoint:     providerConfig.Endpoint,
    }
    
    // Create client
    client, clientID, err := clientFactory.CreateClient(clientConfig)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }
    
    // Set variables...
    return nil
}
```

#### 3.2 Provider-Specific Commands (Optional)

For advanced use cases, create specialized commands:

```go
// \llm-client-get-openai[key=..., base_url=...]
// \llm-client-get-anthropic[key=..., version=...]
// \llm-client-get-gemini[key=..., api_version=...]
```

### Phase 4: Variable System Integration

#### 4.1 Configuration Variables

Expose configuration as variables:

```bash
# Base URLs
${NEURO_OPENAI_BASE_URL}
${NEURO_ANTHROPIC_BASE_URL}
${NEURO_GEMINI_BASE_URL}

# Headers
${NEURO_ANTHROPIC_HEADERS}
${NEURO_GEMINI_HEADERS}

# API Keys (from configuration service)
${NEURO_OPENAI_API_KEY}
${NEURO_ANTHROPIC_API_KEY}
${NEURO_GEMINI_API_KEY}
```

#### 4.2 Variable Interpolation Support

Commands support variable interpolation:

```bash
# Use custom base URL
\llm-client-get[provider=openai, base_url=${MY_OPENAI_PROXY}]

# Use variable for API key
\llm-client-get[provider=anthropic, key=${MY_ANTHROPIC_KEY}]

# Complex configuration
\set[proxy_url="https://my-proxy.com/v1"]
\llm-client-get[provider=openai, base_url=${proxy_url}]
```

## Configuration Examples

### Environment Variables

```bash
# Override default base URLs
export NEURO_OPENAI_BASE_URL="https://my-openai-proxy.com/v1"
export NEURO_GEMINI_BASE_URL="https://custom-gemini.ai/v1"

# Custom headers
export NEURO_ANTHROPIC_HEADERS="Custom-App=MyApp,Version=1.0"

# API keys
export NEURO_OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="AIza..."
```

### .env Files

```bash
# ~/.config/neuroshell/.env
NEURO_OPENAI_BASE_URL=https://corporate-proxy.com/openai/v1
NEURO_ANTHROPIC_HEADERS=Corporate-ID=12345,Department=AI

# ./.env (project-specific)
NEURO_GEMINI_BASE_URL=https://dev-gemini.company.com/v1
GOOGLE_API_KEY=dev-key-123
```

### Command Usage

```bash
# Use defaults
\llm-client-get[provider=openai]

# Override base URL
\llm-client-get[provider=openai, base_url=https://custom.openai.com/v1]

# Variable interpolation
\set[my_key="${OPENAI_API_KEY}"]
\llm-client-get[provider=openai, key=${my_key}]

# Complex configuration
\set[proxy_base="https://proxy.company.com"]
\llm-client-get[provider=gemini, base_url=${proxy_base}/gemini/v1]
```

## Migration Strategy

### Phase 1: Backward Compatibility

1. Keep existing `GetClientForProvider` method
2. Add new configuration-driven methods alongside
3. Update commands to use new methods
4. Existing code continues working

### Phase 2: Gradual Migration

1. Add deprecation warnings to old methods
2. Update all internal code to use new methods
3. Document migration path for external users

### Phase 3: Clean Up

1. Remove deprecated methods
2. Simplify internal architecture
3. Update all documentation

## Benefits

### ✅ Architectural Benefits

- **Clean separation of concerns**: Services don't know about each other
- **Command orchestration**: Commands coordinate multiple services properly
- **Configurable by design**: No more hard-coded values
- **Variable system leverage**: Full integration with NeuroShell variables

### ✅ User Benefits

- **Flexible configuration**: Override any provider setting via environment
- **Variable interpolation**: Use NeuroShell's powerful variable system
- **Custom providers**: Easy to add via configuration
- **Corporate environments**: Support for proxies and custom headers

### ✅ Developer Benefits

- **Maintainable code**: Configuration changes don't require code changes
- **Testable architecture**: Easy to mock and test individual components
- **Extensible design**: New providers added via configuration, not code
- **Professional patterns**: Follows dependency injection and SRP principles

## Implementation Checklist

### Configuration Service Enhancement
- [ ] Add `ProviderConfig` structure
- [ ] Implement `GetProviderConfig()` method
- [ ] Add default provider configurations
- [ ] Support configuration overrides
- [ ] Add provider-specific API key resolution

### ClientFactory Refactoring
- [ ] Create `ClientConfig` structure
- [ ] Implement unified `CreateClient()` method
- [ ] Remove hard-coded switch statements
- [ ] Consolidate duplicate logic
- [ ] Add configuration validation

### Command Integration
- [ ] Update `llm-client-get` command
- [ ] Add configuration service orchestration
- [ ] Support parameter overrides
- [ ] Maintain backward compatibility
- [ ] Add comprehensive error handling

### Variable System Integration
- [ ] Expose configuration as variables
- [ ] Support variable interpolation in commands
- [ ] Add configuration variable documentation
- [ ] Test variable override scenarios

### Testing
- [ ] Unit tests for configuration service methods
- [ ] Integration tests for command orchestration
- [ ] End-to-end tests for variable interpolation
- [ ] Migration compatibility tests
- [ ] Configuration override tests

### Documentation
- [ ] Update command documentation
- [ ] Add configuration examples
- [ ] Document migration path
- [ ] Add troubleshooting guide
- [ ] Update architecture documentation

## Future Enhancements

### Dynamic Provider Registration

Support for registering new providers at runtime:

```bash
\provider-register[
  name=custom-ai,
  type=openai-compatible,
  base_url=https://custom-ai.com/v1,
  headers=Custom-Key=value
]
```

### Configuration Validation

Add configuration validation and health checks:

```bash
\config-validate[provider=openai]  # Validate OpenAI configuration
\config-test[provider=anthropic]   # Test API connectivity
```

### Advanced Variable Integration

Support for complex variable operations:

```bash
\set[providers="openai,anthropic,gemini"]
\for[provider in ${providers}] {
  \llm-client-get[provider=${provider}]
}
```

## Conclusion

This refactoring plan transforms the hard-coded, tightly-coupled client factory into a flexible, configuration-driven system that properly follows NeuroShell's three-layer architecture while leveraging its powerful variable system. The result is a more maintainable, testable, and user-friendly LLM client management system.