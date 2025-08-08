# Neuro Shell Design Document

## 1. Overview

Neuro is a specialized shell environment designed for seamless interaction with LLM agents. It bridges the gap between traditional command-line interfaces and modern AI assistants, enabling professionals to create reproducible, automated workflows that combine system operations with LLM capabilities.

### Vision
Create a shell that treats LLM interactions as first-class operations, similar to how traditional shells handle file operations and process management.

### Target Users
- Data scientists and statisticians familiar with Stata-like environments
- Software engineers working with AI tools like Claude Code
- Researchers who need reproducible AI-assisted workflows
- Professionals automating complex analytical tasks

## 2. Core Features

### Phase 1 Features (MVP)
1. **Command System**: Prefix-based commands with `\command[options] message` syntax
2. **Variable System**: Session-scoped variables with interpolation
3. **Session Management**: Create, save, load, and manage conversation contexts
4. **Script Execution**: Run `.neuro` files for reproducible workflows
5. **Shell Integration**: Execute system commands via `\bash`
6. **History Tracking**: Access conversation and command history

### Core Commands (Top 10)
1. `\send` - Send message to LLM agent
2. `\set[var=value]` - Set variable
3. `\get[var]` - Get variable value
4. `\new` - Start new session
5. `\history[n=5]` - View recent exchanges
6. `\clear` - Clear current session
7. `\save[name="session1"]` - Save session state
8. `\load[name="session1"]` - Load saved session
9. `\list` - List all variables
10. `\help[command]` - Show command help

## 3. Syntax Design

### Command Syntax
```
\command[opt1, opt2, key=value, ...] user message
```

### Variable System

#### Naming Convention
```
${1}, ${2}, ${3}...     → Message history (1=latest agent response)
${_output}              → Command outputs (underscore prefix)
${@pwd}                 → System information (@ prefix)
${#tokens_used}         → Metadata (# prefix)
${user_variable}        → User-defined (no special prefix)
```

#### Special Variables
- **Message History**: `${1}` (latest agent), `${2}` (latest user), etc.
- **Command Returns**: `${_output}`, `${_error}`, `${_status}`, `${_elapsed}`
- **System Variables**: `${@pwd}`, `${@user}`, `${@home}`, `${@date}`, `${@os}`
- **Metadata**: `${#model}`, `${#tokens_used}`, `${#session_id}`, `${#message_count}`

### Script Files (.neuro)
```neuro
# Example script
\set[data="sales_2024.csv"]
\bash[python preprocess.py ${data}]
\send Analyze the processed data in ${_output}
\save[name="analysis_${@date}"]
```

## 4. Design Principles

### 1. **Simplicity First**
- Start with minimal command set
- No pipelines or complex chaining initially
- Session-scoped variables only

### 2. **Stata-Inspired**
- Commands operate on implicit state
- Return values in predictable variables
- Line-by-line script execution

### 3. **Reproducibility**
- All workflows can be saved as scripts
- Sessions are fully serializable
- Clear variable naming prevents collisions

### 4. **Safety**
- Sandboxed bash execution with timeouts
- No shell expansion in commands
- Clear separation between system and user variables

### 5. **Extensibility**
- Service-based architecture for new features
- Plugin system planned for future
- Clear interfaces between layers

## 5. Software Architecture

### Three-Layer Architecture

NeuroShell implements a clean separation of concerns through three distinct layers:

**Command Layer** (Top) - User-facing commands like `\send`, `\set`, `\get`, `\session-new`, `\model-activate` that parse user input and orchestrate operations across services.

**Service Layer** (Middle) - Core business logic including VariableService for state management, ChatSessionService for conversation handling, ModelService for LLM model management, ClientFactoryService for LLM client creation, and BashService for system command execution.

**Context Layer** (Bottom) - Global state management including variable storage, session persistence, LLM client connections, and system interface abstractions.

### Layer Responsibilities

#### Context Layer
- **Purpose**: Holds all state and resources
- **Components**:
  - Variable storage (HashMap)
  - Message history (Ring buffer)
  - Session state
  - Agent connections
  - OS interface

#### Service Layer
- **Purpose**: Business logic and operations
- **Core Services**:
  1. **VariableService**: Get/Set/List/Interpolate variables
  2. **SessionService**: New/Save/Load/Clear sessions
  3. **AgentService**: Send messages, manage LLM connections
  4. **HistoryService**: Track commands and conversations
  5. **ScriptService**: Load and execute .neuro files
  6. **BashService**: Execute system commands safely

#### Command Layer
- **Purpose**: Parse and execute user commands
- **Responsibilities**:
  - Command parsing
  - Service orchestration
  - Error handling
  - Result formatting

### Key Interfaces

```go
// Core service interface - all services implement this
type Service interface {
    Name() string
    Initialize() error
    GetServiceInfo() map[string]interface{}
}

// Command interface for builtin commands
type Command interface {
    Name() string
    Description() string
    Usage() string
    ParseMode() ParseMode
    Execute(options map[string]string, args []string, registry ServiceRegistry) error
}

// Global context for state management
type NeuroContext interface {
    GetVariable(name string) (string, error)
    SetVariable(name string, value string) error
    SetSystemVariable(name string, value string) error
    GetAllVariables() map[string]string
    InterpolateVariables(input string) (string, error)
}

// LLM client interface for provider abstraction
type LLMClient interface {
    GetProviderName() string
    IsConfigured() bool
    SendChatCompletion(messages []neurotypes.ChatMessage, model string, params map[string]interface{}) (*neurotypes.ChatResponse, error)
    StreamChatCompletion(messages []neurotypes.ChatMessage, model string, params map[string]interface{}) (<-chan neurotypes.ChatResponse, error)
}
```

### Implementation Technologies
- **Language**: Go
- **UI Framework**: charm.sh libraries (bubbletea, lipgloss)
- **Shell Framework**: ishell or go-prompt
- **Parser**: participle for command syntax
- **Configuration**: viper

## 6. Current Implementation Status

### ✅ Completed Features
- **Command System**: Full `\command[options] message` syntax with 50+ builtin commands
- **Multi-Provider LLM Support**: Anthropic Claude, OpenAI GPT/o1, Google Gemini, Moonshot Kimi
- **Advanced Session Management**: Create, activate, copy, edit, export/import sessions
- **Model Management**: Catalog-based model creation with provider-specific parameters
- **Variable System**: User, system (`@`), command output (`_`), and metadata (`#`) variables
- **Script Execution**: Batch mode with `.neuro` files and comprehensive stdlib
- **Shell Integration**: `\bash` command with timeout and variable capture
- **Thinking Blocks**: Provider-specific thinking block rendering for reasoning models
- **Testing Framework**: Golden file testing with 178+ end-to-end tests

### 🚧 Active Development
- **Performance Optimization**: Benchmarking and optimization of core services
- **Enhanced Error Handling**: Better error messages and recovery mechanisms
- **Documentation**: Comprehensive help system and user guides

### 🔮 Future Enhancements
- **Plugin System**: External command and service plugins
- **Web Interface**: Browser-based shell for remote access
- **Team Features**: Shared sessions and collaborative workflows
- **IDE Integration**: VSCode extension and editor plugins

## 7. Example Workflows

### Basic LLM Interaction
```neuro
%% Simple conversation with auto-model creation
\send Hello, can you help me with Python programming?
\send What's the difference between lists and tuples?
\session-show  %% View conversation history
```

### Multi-Provider Model Comparison
```neuro
%% Compare responses across different LLM providers
\anthropic-client-new claude-client
\model-new[catalog_id="CS4", thinking_budget=1000] claude-model
\model-activate claude-model
\send Explain quantum computing in simple terms

\gemini-client-new gemini-client  
\model-new[catalog_id="GM25P", thinking_budget=2048] gemini-model
\model-activate gemini-model
\send Explain quantum computing in simple terms

%% Compare the responses
\echo Claude response: ${1}
\echo Gemini response: ${3}
```

### Data Analysis Workflow
```neuro
%% Analyze data with LLM assistance
\session-new[system="You are a data analyst expert"] analysis-session
\bash[head -10 data.csv > ${_output}]
\send Here's a sample of my dataset: ${_output}. What insights can you provide?
\bash[python analyze.py > ${_output}]
\send Based on this analysis output: ${_output}, what recommendations do you have?
\session-export[format="json"] analysis_results.json
```

### Code Review Assistant
```neuro
%% Set up specialized session for code review
\session-new[system="You are a senior software engineer doing code review"] review-session
\bash[git diff HEAD~1 > ${_output}]
\send Please review these changes: ${_output}
\send What potential issues or improvements do you see?
\get[1]  %% Get the review feedback
\bash[echo "${1}" > code_review.md]  %% Save review to file
```

## 8. Success Metrics
- Command execution latency < 100ms
- Variable interpolation performance with 1000+ vars
- Script execution reliability > 99.9%
- Memory usage < 100MB for typical sessions
- Support for conversation histories > 1000 messages

## 9. Future Considerations
- Web API for remote execution
- Integration with popular IDEs
- Natural language command parsing
- Multi-modal support (images, files)
- Distributed session state

## 10. Development Workflow

### Building and Testing

For development, always run:

```bash
# Build all binaries with clean lint and format
just build

# Run comprehensive test suite
just test
```

### Testing Requirements

When running Go tests directly, always prefix with `EDITOR=echo` to prevent editor popups:

```bash
# Correct way to run Go tests
EDITOR=echo go test ./internal/services/...
EDITOR=echo go test -v ./internal/commands/builtin/...

# Available test commands
just test-unit        # Service and utility tests
just test-commands    # Command tests  
just test-e2e         # End-to-end golden file tests
just test-all-units   # All unit tests combined
```

### Code Quality

```bash
just format    # Format code and organize imports
just lint      # Run all linters and formatters
just imports   # Organize Go imports only
```

### End-to-End Testing

```bash
# Run specific e2e test
./bin/neurotest run send-basic-thinking

# Record new test case
./bin/neurotest record my-new-test

# Run all e2e tests
just test-e2e
```
