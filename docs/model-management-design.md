# Model Management System Design for NeuroShell

## 1. Overview

NeuroShell's model management system enables professional CLI users to configure, manage, and utilize multiple LLM models with different parameters and providers. This system bridges the gap between raw API access and user-friendly model configuration, providing reproducible and scriptable LLM interactions.

### Vision
Create a model management system that treats LLM model configurations as first-class entities, similar to how traditional CLI tools manage profiles, configurations, and environments.

### Target Use Cases
- **Development Teams**: Different models for code review vs. creative brainstorming
- **Data Scientists**: Consistent model configurations across analysis workflows
- **Researchers**: Reproducible experiments with specific model parameters
- **Professionals**: Context-specific models for different types of work

## 2. Core Concepts and Architecture

### Model Definition
A **Model** is a configured LLM instance that combines:
- **API Provider** (e.g., openai, anthropic, local)
- **Base Model** (e.g., gpt-4, claude-3, llama-2)  
- **Parameters** (temperature, max_tokens, top_p, etc.)
- **User Configuration** (name, description, default status)

### Key Design Principles
- **CLI-First**: Familiar patterns for terminal professionals
- **Reproducible**: Model configurations can be scripted and versioned
- **Flexible**: Support multiple providers and extensible parameters
- **Memory-Based**: In-memory storage for this phase (no persistence)
- **Integrated**: Seamless integration with existing session and variable systems

### Relationship to Existing Concepts
```
API Provider (OpenAI) 
├── Base Model (gpt-4)
│   ├── Model Config A (temperature=0.1, name="precise")
│   ├── Model Config B (temperature=0.7, name="creative")
│   └── Model Config C (temperature=0.3, name="balanced")
└── Base Model (gpt-3.5-turbo)
    └── Model Config D (temperature=0.5, name="fast")

Sessions use Models:
Session "code-review" -> Model "precise"
Session "brainstorm" -> Model "creative"
```

## 3. Command Group Design: `\model-*`

Following the existing `session-*` pattern, introduce a `model-*` command group:

### Core Commands

#### `\model-new[provider, model, parameter=value, ...] name`
Create a new model configuration with specified parameters.

**Examples:**
```neuro
\model-new[provider=openai, model=gpt-4, temperature=0.7] gpt4-balanced
\model-new[provider=openai, model=gpt-4, temperature=0.1, max_tokens=2000] gpt4-precise
\model-new[provider=anthropic, model=claude-3-sonnet, temperature=0.5] claude-default
\model-new[provider=local, model=llama-2, temperature=0.3, context_length=4096] local-llama
```

#### `\model-list[sort=name|provider|created] [provider=filter]`
List all configured models with filtering and sorting options.

**Examples:**
```neuro
\model-list                           # List all models
\model-list[sort=provider]           # Sort by provider
\model-list[provider=openai]         # Show only OpenAI models
```

#### `\model-delete name`
Remove a model configuration.

**Examples:**
```neuro
\model-delete old-config
\model-delete gpt4-test
```

#### `\model-set-default name`
Set the default model for new sessions.

**Examples:**
```neuro
\model-set-default gpt4-balanced
\model-set-default claude-default
```

#### `\model-show name`
Display detailed configuration for a specific model.

**Examples:**
```neuro
\model-show gpt4-precise
\model-show claude-default
```

#### `\model-test name`
Test model connectivity and basic functionality.

**Examples:**
```neuro
\model-test gpt4-balanced
\model-test claude-default
```

### Command Integration with Sessions

#### Enhanced Session Commands
```neuro
# Create session with specific model
\session-new[model=gpt4-precise, system="You are a code reviewer"] review-session

# Create session with default model
\session-new code-analysis  # Uses system default model

# Send message with temporary model override
\send[model=claude-default] Get a different perspective on this issue
```

## 4. Data Structures and Types

### ModelConfig Type
```go
type ModelConfig struct {
    ID           string            `json:"id"`           // Unique identifier
    Name         string            `json:"name"`         // User-friendly name
    Provider     string            `json:"provider"`     // API provider (openai, anthropic, etc.)
    BaseModel    string            `json:"base_model"`   // Provider's model name
    Parameters   map[string]any    `json:"parameters"`   // Model parameters (temperature, etc.)
    Description  string            `json:"description"`  // Optional description
    IsDefault    bool              `json:"is_default"`   // Whether this is the default model
    CreatedAt    time.Time         `json:"created_at"`   
    UpdatedAt    time.Time         `json:"updated_at"`   
}
```

### Provider Interface
```go
type LLMProvider interface {
    Name() string
    SupportedModels() []string
    DefaultParameters() map[string]any
    ValidateConfig(config ModelConfig) error
    CreateClient(config ModelConfig, apiKey string) (LLMClient, error)
}

type LLMClient interface {
    SendMessage(message string, history []Message) (*Message, error)
    ValidateConnection() error
    GetModelInfo() ModelInfo
}

type ModelInfo struct {
    Provider      string
    ModelName     string
    MaxTokens     int
    SupportsTools bool
    CostPerToken  float64
}
```

### Common Model Parameters
```go
// Standard parameters across providers
type StandardParameters struct {
    Temperature   *float64 `json:"temperature,omitempty"`   // 0.0-1.0
    MaxTokens     *int     `json:"max_tokens,omitempty"`    // Max output tokens
    TopP          *float64 `json:"top_p,omitempty"`         // 0.0-1.0
    TopK          *int     `json:"top_k,omitempty"`         // For some providers
    PresencePenalty *float64 `json:"presence_penalty,omitempty"` // -2.0 to 2.0
    FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"` // -2.0 to 2.0
}
```

## 5. Service Architecture

### ModelService
**Responsibilities:**
- Manage model configurations in memory
- Validate model parameters against provider capabilities
- Provide model lookup and default handling
- Integrate with existing variable system
- Handle model lifecycle (create, update, delete)

**Key Methods:**
```go
type ModelService interface {
    CreateModel(config ModelConfig) error
    GetModel(name string) (ModelConfig, error)
    ListModels(filters ModelFilters) []ModelConfig
    DeleteModel(name string) error
    SetDefaultModel(name string) error
    GetDefaultModel() (ModelConfig, error)
    ValidateModel(config ModelConfig) error
    TestModel(name string) error
}
```

### Enhanced ChatSessionService
**Changes:**
- Accept ModelConfig instead of hardcoded settings
- Support per-session model overrides
- Maintain backward compatibility with existing sessions
- Handle model-specific message formatting

### ProviderRegistry
**Responsibilities:**
- Register and manage LLM providers
- Validate provider-specific configurations
- Create provider clients with API keys
- Support for extensible provider ecosystem

**Built-in Providers:**
- **OpenAI Provider**: GPT-4, GPT-3.5, etc.
- **Anthropic Provider**: Claude models
- **Local Provider**: Local model servers (Ollama, etc.)
- **Mock Provider**: For testing and development

## 6. Integration Points

### Session Integration
Models integrate seamlessly with the existing session system:

```neuro
# Create session with specific model
\session-new[model=gpt4-precise, system="You are a code reviewer"] review-session

# Use default model for session
\session-new data-analysis  # Uses current default model

# Override model for specific messages within a session
\send[model=claude-default] Get another perspective on this
\send[model=gpt4-creative] Brainstorm creative solutions
```

### Variable System Extensions
New metadata variables:
- `${#model}` - Current session's model name
- `${#model_provider}` - Current model's provider
- `${#model_temperature}` - Current model's temperature
- `${#model_max_tokens}` - Current model's max tokens
- `${#default_model}` - System default model name
- `${#available_models}` - Comma-separated list of configured models

### Script Integration
```neuro
# Reproducible model configurations in scripts
\model-new[provider=openai, model=gpt-4, temperature=0.2, max_tokens=1000] analysis-model
\model-set-default analysis-model

# Use in session workflow
\session-new[model=analysis-model, system="You are a data analyst"] data-session
\bash cat quarterly_sales.csv
\send Analyze these quarterly sales trends: ${_output}
\set[insights="${1}"]

# Log model info for reproducibility
\echo Analysis completed with model: ${#model} (${#model_provider})
\echo Parameters: temp=${#model_temperature}, max_tokens=${#model_max_tokens}
```

### Command Result Variables
Each model command sets result variables for scripting:

```neuro
# \model-new sets:
\model-new[provider=openai, model=gpt-4] test-model
# ${_model_created} = "test-model"
# ${_model_provider} = "openai"
# ${_model_base_model} = "gpt-4"

# \model-list sets:
\model-list[provider=openai]
# ${_model_count} = "3"
# ${_model_list} = "model1,model2,model3"

# \model-test sets:
\model-test gpt4-balanced
# ${_model_test_status} = "success" | "failed"
# ${_model_test_latency} = "1.2s"
# ${_model_test_error} = "" | "error message"
```

## 7. Implementation Structure

### File Organization
```
internal/
├── commands/
│   └── model/              # New model command group
│       ├── new.go          # \model-new command
│       ├── list.go         # \model-list command  
│       ├── delete.go       # \model-delete command
│       ├── show.go         # \model-show command
│       ├── set_default.go  # \model-set-default command
│       └── test.go         # \model-test command
├── services/
│   ├── model_service.go    # Model management service
│   └── provider_registry.go # LLM provider registry
├── providers/              # LLM provider implementations
│   ├── openai/
│   │   ├── provider.go     # OpenAI provider implementation
│   │   ├── client.go       # OpenAI API client
│   │   └── models.go       # OpenAI model definitions
│   ├── anthropic/
│   │   ├── provider.go     # Anthropic provider implementation
│   │   ├── client.go       # Anthropic API client
│   │   └── models.go       # Anthropic model definitions
│   ├── local/
│   │   ├── provider.go     # Local provider (Ollama, etc.)
│   │   └── client.go       # Local API client
│   └── mock/
│       ├── provider.go     # Mock provider for testing
│       └── client.go       # Mock client implementation
└── pkg/neurotypes/
    └── model_types.go      # Model-related type definitions
```

### Command Pattern Consistency
All model commands follow established NeuroShell patterns:
- **Naming**: `\model-action` format
- **Parsing**: `ParseModeKeyValue` for options
- **Help**: Comprehensive `HelpInfo` with examples
- **Variables**: Set result variables for scripting
- **Errors**: Consistent error handling and messages

## 8. Example User Workflows

### Initial Setup Workflow
```neuro
# Set up different models for different use cases
\model-new[provider=openai, model=gpt-4, temperature=0.1, max_tokens=2000] precise
\model-new[provider=openai, model=gpt-4, temperature=0.7, max_tokens=1500] creative  
\model-new[provider=anthropic, model=claude-3-sonnet, temperature=0.5] claude
\model-new[provider=local, model=llama-2, temperature=0.3] local-fast

# Set default and verify setup
\model-set-default precise
\model-list
\model-test precise
```

### Development Workflow
```neuro
# Code review session with precise model
\session-new[model=precise, system="You are a senior code reviewer"] review
\bash git diff HEAD~1
\send Review this code change: ${_output}
\set[review_feedback="${1}"]

# Switch to creative model for brainstorming in same script
\session-new[model=creative, system="You are a creative product manager"] brainstorm
\send Generate 10 innovative features based on this feedback: ${review_feedback}

# Use local model for quick questions
\send[model=local-fast] Quick syntax check: is this Go code valid?
```

### Scripted Analysis Pipeline
```neuro
# Configure analysis model
\model-new[provider=openai, model=gpt-4, temperature=0.3, max_tokens=3000] analyst
\model-set-default analyst

# Create session and run analysis
\session-new[system="You are a data analyst expert"] quarterly-review
\bash python extract_data.py --quarter Q4
\send Analyze this quarterly data for trends and insights: ${_output}
\set[analysis="${1}"]

# Generate summary with different model for comparison
\send[model=claude] Provide a second opinion on this analysis: ${analysis}
\set[second_opinion="${1}"]

# Save results with model metadata
\echo Report generated with:
\echo Primary model: ${#model} (temp=${#model_temperature})
\echo Secondary model: claude
\bash echo "${analysis}" > analysis_${@date}_${#model}.txt
\bash echo "${second_opinion}" > analysis_${@date}_claude.txt
```

### Multi-Provider Comparison
```neuro
# Test same prompt across different models
\set[prompt="Explain quantum computing in simple terms"]

\send[model=precise] ${prompt}
\set[gpt4_response="${1}"]

\send[model=claude] ${prompt}
\set[claude_response="${1}"]

\send[model=local-fast] ${prompt}
\set[local_response="${1}"]

# Compare responses
\echo GPT-4 Response: ${gpt4_response}
\echo Claude Response: ${claude_response}
\echo Local Response: ${local_response}
```

## 9. Testing Strategy

### Unit Tests
- **Model Configuration**: Validation, serialization, parameter handling
- **Provider Registry**: Provider registration, lookup, validation
- **Command Parsing**: Argument validation, error handling
- **Service Integration**: Model service operations, default handling

### Integration Tests
- **Provider Integration**: Real API calls to supported providers
- **Session Integration**: Model usage within sessions
- **Variable Integration**: Model metadata in variable system
- **Command Workflow**: Complete command sequences

### End-to-End Tests
```neuro
# Test files for each major workflow
test/golden/model-basic/              # Basic model operations
test/golden/model-session-integration/ # Models with sessions
test/golden/model-variables/          # Model metadata variables
test/golden/model-scripting/          # Complex scripted workflows
test/golden/model-error-handling/     # Error conditions
```

### Performance Tests
- **Model Creation**: Performance with many model configurations
- **Provider Switching**: Latency when switching between providers
- **Memory Usage**: Memory efficiency of in-memory storage

## 10. Error Handling and Validation

### Model Configuration Validation
```neuro
# Invalid provider
\model-new[provider=invalid] test
# Error: Provider 'invalid' not supported. Available: openai, anthropic, local

# Invalid parameters
\model-new[provider=openai, model=gpt-4, temperature=2.0] test
# Error: temperature must be between 0.0 and 1.0 for openai provider

# Duplicate name
\model-new[provider=openai, model=gpt-4] existing-name
# Error: Model 'existing-name' already exists. Use \model-delete first or choose different name.
```

### Runtime Error Handling
```neuro
# Model not found
\send[model=nonexistent] Hello
# Error: Model 'nonexistent' not found. Use \model-list to see available models.

# API connection issues
\model-test failing-model
# ${_model_test_status} = "failed"
# ${_model_test_error} = "API connection failed: invalid API key"

# Provider-specific errors
\send[model=local-offline] Hello
# Error: Local provider connection failed. Ensure local model server is running.
```

## 11. Future Extensibility

### Phase 2 Enhancements
- **Persistence**: Save model configurations to disk/database
- **API Key Management**: Secure storage and management of API keys
- **Usage Tracking**: Monitor token usage and costs per model
- **Model Metrics**: Track performance, latency, success rates
- **Batch Operations**: Bulk model configuration operations
- **Model Templates**: Predefined model configurations for common use cases

### Provider Ecosystem
- **Plugin Architecture**: Custom provider plugins
- **Community Registry**: Shared model configurations
- **Auto-Discovery**: Automatic detection of available models
- **Provider Capabilities**: Dynamic capability detection and validation
- **Cost Management**: Real-time cost tracking and budgets

### Advanced Features
- **Model Chaining**: Sequential model operations
- **Ensemble Models**: Combine responses from multiple models
- **A/B Testing**: Compare model performance on same inputs
- **Model Versioning**: Track and manage model configuration versions
- **Performance Optimization**: Caching, request batching, connection pooling

### Integration Opportunities
- **CI/CD Integration**: Model configurations in version control
- **Team Collaboration**: Shared model configurations
- **Monitoring**: Integration with observability tools
- **Compliance**: Audit trails and compliance reporting

## 12. Migration and Backward Compatibility

### Existing Session Support
- Current sessions without model specification use system default
- Existing `\send` commands work unchanged with default model
- Session format remains compatible, with optional model field

### Gradual Adoption
```neuro
# Users can start simple
\session-new basic-session
\send Hello  # Uses default model (auto-configured)

# Then adopt model management gradually
\model-new[provider=openai, model=gpt-4, temperature=0.5] my-model
\model-set-default my-model

# Advanced users get full control
\model-new[provider=openai, model=gpt-4, temperature=0.1] precise
\model-new[provider=anthropic, model=claude-3] claude
\session-new[model=precise] code-review
\send[model=claude] Get second opinion
```

This design provides a comprehensive foundation for LLM model management while maintaining NeuroShell's core principles of simplicity, reproducibility, and professional CLI experience.