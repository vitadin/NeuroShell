# Multi-Line Shell Prompt System for NeuroShell

## Overview

This document describes the design and implementation of a flexible, multi-line shell prompt system for NeuroShell that displays rich contextual information while strictly adhering to NeuroShell's three-layer architecture.

## Architecture

### Core Components

#### 1. ShellPromptService
- **Location**: `internal/services/shell_prompt_service.go`
- **Purpose**: Retrieve prompt configuration from context layer
- **Key Principle**: No interpolation, only template retrieval
- **Layer**: Service Layer (interacts only with Context)

#### 2. Prompt Variables in Context
- **Location**: `internal/context/context.go`
- **Variables**: `_prompt_line1` through `_prompt_line5`, `_prompt_lines_count`
- **Storage**: Added to `allowedGlobalVariables` array
- **Access**: Standard variable get/set mechanisms

#### 3. Shell Integration
- **Location**: `cmd/neuro/main.go`
- **Responsibility**: Orchestrate between service and state machine
- **Interpolation**: Handled by state machine at shell layer
- **Display**: Dynamic prompt generation before each input

#### 4. Shell Prompt Commands
- **Location**: `internal/commands/builtin/shell_prompt.go`
- **Commands**:
  - `\shell-prompt` - Main configuration command
  - `\shell-prompt-show` - Display current configuration
  - `\shell-prompt-preset` - Apply preset configurations
  - `\shell-prompt-preview` - Preview with current context

## Configuration

### Basic Usage

```neuro
# Set single-line prompt (classic mode)
\set[_prompt_lines_count=1]
\set[_prompt_line1="neuro> "]

# Set two-line prompt (recommended)
\set[_prompt_lines_count=2]
\set[_prompt_line1="${@user}@neuro:${@pwd} [${#session_name:-no-session}]"]
\set[_prompt_line2="❯ "]

# Set three-line prompt (power user)
\set[_prompt_lines_count=3]
\set[_prompt_line1="┌─[${@time}] [${#active_model:-none}]"]
\set[_prompt_line2="├─[${@pwd}] [${#session_name}:${#message_count}]"]
\set[_prompt_line3="└➤ "]
```

### Available Variables

System variables available for prompt customization:
- `${@pwd}` - Current working directory
- `${@user}` - Current username
- `${@home}` - Home directory
- `${@date}` - Current date
- `${@time}` - Current time
- `${@os}` - Operating system
- `${#session_name}` - Active session name
- `${#session_id}` - Active session ID
- `${#message_count}` - Number of messages in session
- `${#active_model}` - Currently active model
- `${@status}` - Last command status
- `${@error}` - Last error message

### Preset Configurations

#### Minimal
```neuro
\shell-prompt-preset[style=minimal]
# Sets:
# _prompt_lines_count=1
# _prompt_line1="> "
```

#### Default
```neuro
\shell-prompt-preset[style=default]
# Sets:
# _prompt_lines_count=2
# _prompt_line1="${@pwd} [${#session_name:-no-session}]"
# _prompt_line2="neuro> "
```

#### Developer
```neuro
\shell-prompt-preset[style=developer]
# Sets:
# _prompt_lines_count=2
# _prompt_line1="${@pwd} (${#git_branch}) ${@status}"
# _prompt_line2="❯ "
```

#### Powerline
```neuro
\shell-prompt-preset[style=powerline]
# Sets:
# _prompt_lines_count=3
# _prompt_line1="┌─[${@user}@${@hostname}]-[${@time}]"
# _prompt_line2="├─[${#session_name}:${#message_count}]-[${#active_model}]"
# _prompt_line3="└─➤ "
```

## Implementation Details

### Service Layer Implementation

```go
package services

import (
    "fmt"
    "strconv"
    "neuroshell/internal/context"
)

type ShellPromptService struct {
    initialized bool
}

func NewShellPromptService() *ShellPromptService {
    return &ShellPromptService{}
}

func (s *ShellPromptService) Name() string {
    return "shell_prompt"
}

func (s *ShellPromptService) Initialize() error {
    s.initialized = true
    return nil
}

// GetPromptLines retrieves raw prompt templates from context
// NO interpolation happens here - that's the shell layer's job
func (s *ShellPromptService) GetPromptLines() ([]string, error) {
    if !s.initialized {
        return nil, fmt.Errorf("shell prompt service not initialized")
    }
    
    ctx := context.GetGlobalContext()
    
    // Get number of lines to display
    countStr, _ := ctx.GetVariable("_prompt_lines_count")
    count := 1  // default to single line
    if n, err := strconv.Atoi(countStr); err == nil && n >= 1 && n <= 5 {
        count = n
    }
    
    // Retrieve raw templates (NO interpolation)
    lines := make([]string, 0, count)
    for i := 1; i <= count; i++ {
        line, _ := ctx.GetVariable(fmt.Sprintf("_prompt_line%d", i))
        if line == "" && i == 1 {
            line = "neuro> "  // fallback default
        }
        lines = append(lines, line)
    }
    
    return lines, nil
}
```

### Shell Layer Integration

In `cmd/neuro/main.go`, modify the readline configuration:

```go
// createDynamicPrompt generates the prompt string with interpolation
func createDynamicPrompt() string {
    // Get prompt templates from service
    promptService, err := services.GetGlobalRegistry().GetService("shell_prompt")
    if err != nil {
        return "neuro> "  // fallback
    }
    
    shellPrompt := promptService.(*services.ShellPromptService)
    lines, err := shellPrompt.GetPromptLines()
    if err != nil {
        return "neuro> "  // fallback
    }
    
    // Get context for interpolation
    ctx := shell.GetGlobalContext()
    
    // Interpolate each line using context's interpolation
    var promptBuilder strings.Builder
    for i, template := range lines {
        // Use context's InterpolateVariables method
        interpolated, err := ctx.InterpolateVariables(template)
        if err != nil {
            interpolated = template  // use raw template on error
        }
        
        promptBuilder.WriteString(interpolated)
        if i < len(lines)-1 {
            promptBuilder.WriteString("\n")
        }
    }
    
    return promptBuilder.String()
}

// Update readline config to use dynamic prompt
func createCustomReadlineConfig() *readline.Config {
    cfg := &readline.Config{
        Prompt:      createDynamicPrompt(),  // Dynamic prompt
        HistoryFile: "/tmp/neuro_history",
    }
    // ... rest of configuration
}
```

### Command Implementation

```go
package builtin

import (
    "fmt"
    "strconv"
    "neuroshell/internal/context"
    "neuroshell/pkg/neurotypes"
)

type ShellPromptCommand struct{}

func (c *ShellPromptCommand) Name() string {
    return "shell-prompt"
}

func (c *ShellPromptCommand) Description() string {
    return "Configure the shell prompt display"
}

func (c *ShellPromptCommand) Execute(options map[string]string, args string) error {
    ctx := context.GetGlobalContext()
    
    // Handle setting number of lines
    if lines, ok := options["lines"]; ok {
        if n, err := strconv.Atoi(lines); err == nil && n >= 1 && n <= 5 {
            ctx.SetVariable("_prompt_lines_count", lines)
            fmt.Printf("Prompt lines set to %s\n", lines)
        } else {
            return fmt.Errorf("invalid lines value: must be 1-5")
        }
    }
    
    // Handle setting specific line
    for i := 1; i <= 5; i++ {
        key := fmt.Sprintf("line%d", i)
        if template, ok := options[key]; ok {
            varName := fmt.Sprintf("_prompt_line%d", i)
            ctx.SetVariable(varName, template)
            fmt.Printf("Prompt line %d set\n", i)
        }
    }
    
    return nil
}
```

## Architectural Principles

### Layer Separation
1. **Context Layer**: Stores all prompt configuration as variables
2. **Service Layer**: Only retrieves templates, no interpolation
3. **Shell Layer**: Orchestrates service and interpolation
4. **Command Layer**: Only modifies context variables

### No Cross-Layer Violations
- Services do NOT call other services
- Services do NOT perform interpolation
- Commands do NOT directly call services
- Interpolation happens ONLY at shell layer

### Performance Considerations
- Template retrieval: < 1ms
- Interpolation: < 5ms
- Total prompt generation: < 10ms target
- Caching: Consider caching static elements

## Migration Path

### Phase 1: Core Implementation
1. Add prompt variables to `allowedGlobalVariables`
2. Implement ShellPromptService
3. Update main.go with dynamic prompt
4. Test basic functionality

### Phase 2: Commands
1. Implement shell-prompt command
2. Add preset support
3. Create show and preview commands
4. Update help documentation

### Phase 3: Enhancements
1. Add smart path truncation
2. Implement conditional elements
3. Integrate with themes
4. Optimize performance

## Testing Strategy

### Unit Tests
- ShellPromptService.GetPromptLines (template retrieval only)
- Context variable storage and retrieval
- Command parsing and execution

### Integration Tests
- Prompt display with various configurations
- Variable interpolation in prompts
- Multi-line display on different terminals

### Performance Tests
- Prompt generation benchmark
- Memory usage with complex templates
- Stress test with rapid prompt updates

## Default Configuration

In `internal/data/embedded/stdlib/system-init.neuro`:
```neuro
%% Default shell prompt configuration
\silent {
    %% Check if user has already configured prompt
    \if-not[condition="${_prompt_lines_count}"] {
        %% Two-line prompt by default for new installations
        \set[_prompt_lines_count=2]
        \set[_prompt_line1="${@pwd} [${#session_name:-no-session}]"]
        \set[_prompt_line2="neuro> "]
    }
}
```

## Backward Compatibility

- Default behavior remains single-line "neuro> " if not configured
- Existing .neurorc files continue to work
- Users can opt-in to multi-line via commands or .neurorc
- No breaking changes to existing functionality

## Future Enhancements

### Planned Features
- Git branch integration
- Virtual environment indicators
- SSH connection status
- Docker/Kubernetes context
- Custom status indicators

### Conditional Display
- Show elements only when relevant
- Hide session info when no session
- Display git info only in repositories
- Adaptive width based on terminal size

### Theme Integration
- Color support based on current theme
- Style variables for prompt elements
- Accessibility modes (high contrast, no color)