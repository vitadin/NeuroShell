# Send Command Refactoring Plan

## Overview

This document outlines the plan to refactor the monolithic `\send` command implementation from Go code to modular neuro scripts. The goal is to create a maintainable, testable, and extensible architecture that leverages HTTP transparency and existing NeuroShell building blocks.

## Current State Analysis

### Problems with Existing Implementation

The current `\send` command (currently disabled in `internal/commands/builtin/send*.go`) has several architectural issues:

1. **Monolithic Design**: Single 267-line Go file handling all aspects of LLM communication
2. **Hard to Maintain**: Complex pipeline with 11 distinct steps tightly coupled
3. **Limited Extensibility**: Hard to add new LLM providers or customize behavior
4. **Poor Testability**: Monolithic structure makes unit testing difficult
5. **SSE Complexity**: Streaming implementation adds unnecessary complexity

### Current Architecture Steps

The existing `\send` command performs these steps:
1. Input validation and service acquisition
2. Active session retrieval/creation  
3. Model config resolution
4. API key determination and client creation
5. Sync/stream LLM interaction based on `_reply_way` variable
6. Response handling and variable updates
7. Message history management

## New Architecture: HTTP-First, Modular Approach

### Core Philosophy

1. **HTTP Transparency**: Users see and control actual API calls
2. **Provider Flexibility**: Support any LLM provider via HTTP
3. **Reuse Existing**: Leverage existing session/, render/ commands
4. **Single Responsibility**: Each script handles one concern
5. **Educational Value**: Users learn HTTP API patterns
6. **Debugging Friendly**: Clear visibility into each step

### Phase 1: Universal HTTP Building Block

#### HttpService Implementation
- Support GET, POST, PUT, DELETE with full header/body control
- Store response in `_output`, status in `_status`, headers in `_headers`
- Support JSON and text responses with proper error handling
- Timeout and retry logic

#### \http Command Interface
```neuro
# Syntax example
\http[method=POST, url=https://api.openai.com/v1/chat/completions, headers=Authorization:Bearer ${api_key}] request_body

# Variable interpolation for all parameters
\set[api_url="https://api.openai.com/v1/chat/completions"]
\set[auth_header="Authorization:Bearer ${OPENAI_API_KEY}"]
\http[method=POST, url=${api_url}, headers=${auth_header}] ${request_json}
```

**Benefits:**
- Enable full LLM API control for any provider
- Support custom headers, authentication, request bodies
- Variable interpolation for dynamic requests
- Consistent response variable patterns

### Phase 2: Three-Step Modular Scripts

#### Core Building Block Scripts (stdlib/)

**1. `_send_prepare_request.neuro`**
- **Responsibility**: Build LLM API request payload
- **Input**: User message, conversation history, model settings
- **Output**: JSON request body in `${_output}`
- **Features**:
  - Handle different provider formats (OpenAI vs Anthropic JSON structure)
  - Include conversation context from session
  - Set model parameters (temperature, max_tokens, etc.)

```neuro
# Example usage
\_send_prepare_request[provider=openai, model=gpt-4] What is the weather like?
# Results in ${_output} containing formatted JSON request
```

**2. `_send_http_call.neuro`**
- **Responsibility**: Execute HTTP API call to LLM provider
- **Input**: Prepared request body, provider configuration
- **Output**: Raw API response in `${_output}`
- **Features**:
  - Use `\http` for actual API communication
  - Parse response JSON and extract assistant message
  - Handle HTTP errors and API rate limits
  - Support both sync and streaming endpoints

```neuro
# Example usage
\_send_http_call[provider=openai, url=${api_url}] ${prepared_request}
# Results in ${_output} containing API response
```

**3. `_send_render_response.neuro`**
- **Responsibility**: Format and display LLM response
- **Input**: Raw API response, rendering preferences
- **Output**: Formatted display to user
- **Features**:
  - Use existing `\render-markdown` or `\render` commands
  - Apply theming and formatting based on user preferences
  - Extract message content from API response
  - Handle different response formats

```neuro
# Example usage
\_send_render_response[format=markdown, theme=dark] ${api_response}
# Displays formatted response to user
```

#### Main Orchestrator Script

**4. `send.neuro` (Replaces monolithic Go implementation)**
- **Responsibility**: Orchestrate the complete send pipeline
- **Features**:
  - Call the three building block scripts in sequence
  - Use existing `\session-*` commands for session management
  - Update message history variables using VariableService patterns
  - Handle error cases and fallbacks
  - Provider-agnostic orchestration

```neuro
# Main send.neuro script structure
%% Validate input and get active session
\session-get-active

%% Prepare the request
\_send_prepare_request[provider=${llm_provider}, model=${llm_model}] ${_1}

%% Make the API call
\_send_http_call[provider=${llm_provider}] ${_output}

%% Render the response
\_send_render_response[format=markdown] ${_output}

%% Update session and variables
\session-add-message[role=user] ${_1}
\session-add-message[role=assistant] ${assistant_response}
```

### Phase 3: Provider-Specific Extensions

#### Provider-Specific Scripts
- **`_send_openai.neuro`**: OpenAI-specific request formatting and configuration
- **`_send_anthropic.neuro`**: Anthropic-specific request formatting and configuration
- **`_send_custom.neuro`**: Template for custom provider integration

These scripts can override or extend the base building blocks for provider-specific optimizations.

## Integration with Existing Architecture

### Reuse Existing Commands

**Session Management** (internal/commands/session/):
- `\session-new`: Create new chat sessions
- `\session-list`: List and manage sessions
- No new session commands needed

**Rendering** (internal/commands/render/):
- `\render`: Text styling with themes
- `\render-markdown`: Markdown to ANSI terminal
- Perfect for formatting LLM responses

**Variable Management**:
- Use existing VariableService patterns
- Follow established variable naming conventions (`_output`, `${1}`, etc.)
- Message history variable updates

### Service Dependencies

The neuro scripts will leverage existing services:
1. **VariableService**: Variable interpolation and history updates
2. **ChatSessionService**: Session management via existing commands
3. **ModelService**: Model configuration (accessed via variables)
4. **ThemeService**: Consistent styling via render commands

## Implementation Benefits

### Developer Benefits
1. **Maintainability**: Each script has single responsibility
2. **Testability**: Independent testing of each component
3. **Extensibility**: Easy to add new providers or customize behavior
4. **Debugging**: Clear visibility into each step
5. **Reusability**: Building blocks can be used in other contexts

### User Benefits
1. **Transparency**: See actual HTTP API calls and responses
2. **Customization**: Override individual steps without full reimplementation
3. **Learning**: Understand HTTP protocols and API structures
4. **Flexibility**: Support any LLM provider that offers HTTP API
5. **Control**: Fine-grained control over request formatting and response handling

### System Benefits
1. **Consistency**: Follows established NeuroShell patterns
2. **Integration**: Seamless use of existing session and render commands
3. **Performance**: HTTP-only approach, no unnecessary abstractions
4. **Reliability**: Simpler error handling and recovery

## Implementation Priority

### Phase 1: Foundation (Week 1-2)
1. **HttpService + \http command**: Universal HTTP building block
2. **Basic testing**: Ensure HTTP functionality works with external APIs

### Phase 2: Core Scripts (Week 3-4)
1. **`_send_prepare_request.neuro`**: Basic request preparation
2. **`_send_http_call.neuro`**: HTTP execution and response parsing
3. **Simple integration testing**: End-to-end HTTP → LLM → Response

### Phase 3: Orchestration (Week 5)
1. **`send.neuro`**: Main orchestrator script
2. **Integration with session and render commands**
3. **Comprehensive testing with real LLM providers**

### Phase 4: Polish (Week 6)
1. **`_send_render_response.neuro`**: Response formatting
2. **Provider-specific optimizations**
3. **Documentation and examples**

## Success Criteria

1. **Functionality**: Complete replacement of existing \send behavior
2. **Performance**: Response times comparable to or better than Go implementation
3. **Reliability**: Robust error handling and recovery
4. **Extensibility**: Easy addition of new LLM providers
5. **Maintainability**: Clear, testable, single-responsibility components
6. **User Experience**: Transparent, customizable, educational

## Migration Strategy

1. **Parallel Development**: Build new scripts alongside disabled Go commands
2. **Gradual Testing**: Test individual components before full integration
3. **Feature Parity**: Ensure all existing \send features are supported
4. **Smooth Transition**: Enable new implementation when fully tested
5. **Documentation**: Provide migration guide for users

## Future Enhancements

Once the core architecture is stable, consider:
1. **Caching**: HTTP response caching for development workflows
2. **Retry Logic**: Intelligent retry with exponential backoff
3. **Batch Operations**: Multiple requests in single script
4. **Custom Templates**: User-defined request/response templates
5. **Monitoring**: Request/response logging and analytics