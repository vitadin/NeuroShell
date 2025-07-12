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

```
┌──────────────────────────────────┐
│         Command Layer            │
│  (\send, \set, \get, etc.)      │
├──────────────────────────────────┤
│        Service Layer             │
│  ┌─────────┐ ┌─────────┐       │
│  │Variable │ │Session  │       │
│  │Service  │ │Service  │  ...  │
│  └─────────┘ └─────────┘       │
├──────────────────────────────────┤
│        Context Layer             │
│  • State Store (variables)       │
│  • Session State                 │
│  • Agent Connection              │
│  • OS Interface                  │
└──────────────────────────────────┘
```

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
// Simplified interface examples
type Context interface {
    GetVariable(name string) (string, error)
    SetVariable(name string, value string) error
    GetMessageHistory(n int) []Message
    GetSessionState() SessionState
}

type Service interface {
    Name() string
    Initialize(ctx Context) error
}

type Command interface {
    Name() string
    Execute(args []string, input string, services ServiceRegistry) error
}
```

### Implementation Technologies
- **Language**: Go
- **UI Framework**: charm.sh libraries (bubbletea, lipgloss)
- **Shell Framework**: ishell or go-prompt
- **Parser**: participle for command syntax
- **Configuration**: viper

## 6. Development Phases

### Phase 1: Core Shell (MVP)
- Basic command system
- Variable interpolation
- Session management
- Script execution

### Phase 2: Enhanced Integration
- Multi-agent support
- Advanced templating
- Conditional execution
- Error recovery

### Phase 3: Ecosystem
- Plugin system
- Package manager for scripts
- Cloud session sync
- Team collaboration features

## 7. Example Workflows

### Data Analysis
```neuro
\bash[ls data/*.csv > ${_output}]
\set[files="${_output}"]
\send Analyze trends across these files: ${files}
\set[analysis="${1}"]
\send Create visualizations for: ${analysis}
\save[name="quarterly_report"]
```

### Code Review
```neuro
\templates/code_review.neuro
\bash[git diff main > ${_output}]
\send Review these changes: ${_output}
\extract[pattern="LGTM|Needs work", from="${1}", to="decision"]
\bash[echo "${decision}" > review_status.txt]
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

## 10.

Everytime run testing, make sure you add `EDITOR=echo` before calling testing command (e.g. `go test`)
