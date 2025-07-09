# NeuroShell Standard Library Infrastructure Enhancement

## Overview

This document outlines the infrastructure enhancements required to implement a **meta-command system** for NeuroShell, where complex commands are defined as embedded NeuroShell scripts rather than hardcoded Go implementations.

### Vision

Transform NeuroShell from a shell with hardcoded commands into a **programmable shell platform** where:
- Complex behaviors emerge from simple, composable primitives
- Command logic lives in `.neuro` scripts, embedded in the binary
- Everything uses the same execution pipeline and variable system
- **True macro expansion** enables dynamic command generation
- Users can see, understand, and potentially override standard behaviors

### Key Benefits

- **Self-hosting**: NeuroShell commands implemented in NeuroShell itself
- **Macro-enabled**: Variables can contain entire commands for true programmability
- **Transparency**: Users can inspect what commands actually do
- **Modularity**: Complex commands built from simple building blocks
- **Extensibility**: Easy to add new "builtin" commands as scripts
- **Consistency**: All commands use same variable system and execution model
- **Maintainability**: Command logic in readable script format, not Go code

## 1. Enhanced Interpolation System (Macro Support)

### Current State
Interpolation only works within command parameters:
```neuro
\set[name="John"]
\echo Hello ${name}    # Interpolates to: \echo Hello John
```

### Required Enhancement

Implement **command-level macro expansion** where variables can contain entire commands:

```neuro
\set[c1="\set[c2=2]"]
\set[cmd="echo"]
\set[params="[style=bold]"]

${c1}                  # Expands to: \set[c2=2] 
${cmd} Hello World     # Expands to: \echo Hello World
\echo${params} Text    # Expands to: \echo[style=bold] Text
```

This transforms NeuroShell into a **true macro system** enabling:
- Variables containing entire commands
- Dynamic command generation based on context
- Complex command patterns abstracted into reusable macros
- Conditional command execution through variable expansion

### Implementation Requirements

#### 1.1 Enhanced Interpolation Pipeline

```
Current Pipeline:
Parse Command → Queue → Interpolate Parameters → Execute

Enhanced Pipeline:
Parse Command → Pre-Interpolate Entire Line → Parse Again → Queue → Interpolate Parameters → Execute
```

#### 1.2 Two-Phase Interpolation Service

```go
type EnhancedInterpolationService struct {
    variableService neurotypes.VariableService
}

// Phase 1: Command-level interpolation (before parsing)
func (s *EnhancedInterpolationService) InterpolateCommandLine(line string) (string, error) {
    // Apply variable interpolation to the entire command line
    // This can generate completely new commands
    return s.variableService.InterpolateVariables(line), nil
}

// Phase 2: Parameter-level interpolation (after parsing, before execution)  
func (s *EnhancedInterpolationService) InterpolateCommand(cmd *Command) (*Command, error) {
    // Existing functionality - interpolate command parameters
    // This happens after the command is parsed
    return s.interpolateParameters(cmd), nil
}
```

#### 1.3 Script Execution Integration

```go
func ExecuteScript(scriptPath string) error {
    // Phase 1: Load script content
    content, err := loadScript(scriptPath)
    if err != nil {
        return err
    }
    
    // Phase 2: Process each line with macro expansion
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        // Skip empty lines and comments
        if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
            continue
        }
        
        // PHASE 1: Command-level interpolation (macro expansion)
        interpolatedLine, err := interpolationService.InterpolateCommandLine(line)
        if err != nil {
            return fmt.Errorf("macro expansion failed for line '%s': %w", line, err)
        }
        
        // Parse the interpolated command
        cmd, err := parser.ParseInput(interpolatedLine)
        if err != nil {
            return fmt.Errorf("parse failed for interpolated line '%s': %w", interpolatedLine, err)
        }
        
        // PHASE 2: Parameter-level interpolation (existing functionality)
        interpolatedCmd, err := interpolationService.InterpolateCommand(cmd)
        if err != nil {
            return fmt.Errorf("parameter interpolation failed: %w", err)
        }
        
        // Execute the fully interpolated command
        err = commands.GetGlobalRegistry().Execute(interpolatedCmd.Name, interpolatedCmd.Options, interpolatedCmd.Message)
        if err != nil {
            return fmt.Errorf("command execution failed: %w", err)
        }
    }
    
    return nil
}
```

### 1.4 Macro Examples

#### Basic Command Generation
```neuro
\set[operation="set"]
\set[variable="name"]
\set[value="John"]

\${operation}[${variable}="${value}"]
# Expands to: \set[name="John"]
```

#### Conditional Command Generation
```neuro
\set[debug_mode="true"]
\set[debug_cmd="\echo[style=red] DEBUG:"]
\set[normal_cmd="#"]

${debug_mode=true?debug_cmd:normal_cmd} Starting process
# When debug_mode=true, expands to: \echo[style=red] DEBUG: Starting process
# When debug_mode=false, expands to: # Starting process
```

#### Dynamic Command Building
```neuro
\set[provider="openai"]
\set[model="gpt-4"]
\set[llm_command="\llm-send[provider=${provider},model=${model}]"]

${llm_command} How are you?
# Expands to: \llm-send[provider=openai,model=gpt-4] How are you?
```

## 2. Enhanced Command Resolution System

### Current State
Commands are resolved through a single Go-based registry (`commands.GlobalRegistry`).

### Required Enhancement

Implement **priority-based command resolution** after macro expansion:

```
User Input: \send Hello
After Macro Expansion: \send Hello (unchanged)

Resolution Order:
1. Builtin Go Commands (highest priority)
   → Check commands.GlobalRegistry.Get("send")
   
2. Embedded Standard Library Scripts (medium priority)
   → Check embedded filesystem for "stdlib/send.neuro"
   
3. User Script Directory (lowest priority) 
   → Check filesystem for "~/.neuro/scripts/send.neuro"
```

### Implementation Requirements

#### 2.1 Enhanced Command Registry Interface

```go
type CommandResolver interface {
    // Resolve command by name with priority-based lookup
    ResolveCommand(name string) (Command, error)
    
    // Check if command exists at specific priority level
    HasBuiltinCommand(name string) bool
    HasStdlibCommand(name string) bool
    HasUserCommand(name string) bool
}

type ResolvedCommand struct {
    Name     string
    Type     CommandType  // Builtin, Stdlib, User
    Source   string       // Path or "builtin"
    Command  Command      // Actual command implementation
}

type CommandType int
const (
    CommandTypeBuiltin CommandType = iota
    CommandTypeStdlib
    CommandTypeUser
)
```

#### 2.2 Integration with Enhanced Execution Pipeline

```go
func ExecuteCommand(commandName string, args map[string]string, input string) error {
    // Resolve command using priority-based lookup
    resolved, err := commandResolver.ResolveCommand(commandName)
    if err != nil {
        return fmt.Errorf("command not found: %s", commandName)
    }
    
    // Execute based on command type
    switch resolved.Type {
    case CommandTypeBuiltin:
        return resolved.Command.Execute(args, input)
    case CommandTypeStdlib, CommandTypeUser:
        return executeScriptCommand(resolved, args, input)
    }
    
    return fmt.Errorf("unknown command type")
}
```

## 3. Embedded Script Loading Infrastructure

### Current State
NeuroShell can load and execute `.neuro` scripts from filesystem using `ScriptService`.

### Required Enhancement

Add capability to load scripts from **embedded filesystem** within the binary.

### Implementation Requirements

#### 3.1 Embedded Filesystem Structure

```
internal/
  data/
    embedded/
      stdlib/              # NeuroShell Standard Library
        send.neuro         # Enhanced send command
        analyze.neuro      # Data analysis workflows
        review.neuro       # Code review workflows
        deploy.neuro       # Deployment automation
        test.neuro         # Testing workflows
        debug.neuro        # Debugging utilities
```

#### 3.2 Go Embed Integration

```go
//go:embed stdlib/*.neuro
var stdlibFS embed.FS

type StdlibLoader struct {
    fs embed.FS
}

func NewStdlibLoader() *StdlibLoader {
    return &StdlibLoader{fs: stdlibFS}
}

func (s *StdlibLoader) LoadScript(filename string) (string, error) {
    content, err := s.fs.ReadFile("stdlib/" + filename)
    if err != nil {
        return "", fmt.Errorf("stdlib script not found: %s", filename)
    }
    return string(content), nil
}

func (s *StdlibLoader) ListAvailableScripts() ([]string, error) {
    entries, err := s.fs.ReadDir("stdlib")
    if err != nil {
        return nil, err
    }
    
    var scripts []string
    for _, entry := range entries {
        if strings.HasSuffix(entry.Name(), ".neuro") {
            name := strings.TrimSuffix(entry.Name(), ".neuro")
            scripts = append(scripts, name)
        }
    }
    return scripts, nil
}
```

#### 3.3 Script Command Wrapper

```go
type ScriptCommand struct {
    name       string
    scriptPath string
    content    string
}

func NewScriptCommand(name, content string) *ScriptCommand {
    return &ScriptCommand{
        name:    name,
        content: content,
    }
}

func (s *ScriptCommand) Execute(args map[string]string, input string) error {
    // Set script parameters as variables
    err := s.setupParameters(args, input)
    if err != nil {
        return err
    }
    
    // Execute script content with enhanced interpolation
    err = s.executeScriptWithMacros(s.content)
    if err != nil {
        return err
    }
    
    // Clean up parameters
    return s.cleanupParameters()
}
```

## 4. Script Parameter Passing System

### Current State
Scripts can access variables but have no standard way to receive command parameters.

### Required Enhancement

Implement **standardized parameter passing** from command invocation to script execution, compatible with the macro system.

### Implementation Requirements

#### 4.1 Parameter Variable System

```
Command: \send[model=gpt-4] Hello, how are you?

Available in script:
${0}     = "send"                    # Command name
${1}     = "Hello, how are you?"     # Input parameter
${model} = "gpt-4"                   # Named argument
${*}     = "Hello, how are you?"     # All positional args
${@}     = "model=gpt-4"             # All named args
```

#### 4.2 Parameter Setup Integration

```go
func (s *ScriptCommand) setupParameters(args map[string]string, input string) error {
    vs, err := services.GetGlobalVariableService()
    if err != nil {
        return err
    }
    
    // Set standard script parameters (compatible with macro system)
    vs.SetSystemVariable("0", s.name)           // Command name
    vs.SetSystemVariable("1", input)            // Input parameter
    vs.SetSystemVariable("*", input)            // All positional args
    
    // Set named arguments as variables (enables macro expansion)
    var namedArgs []string
    for key, value := range args {
        vs.SetSystemVariable(key, value)
        namedArgs = append(namedArgs, key+"="+value)
    }
    vs.SetSystemVariable("@", strings.Join(namedArgs, " "))
    
    return nil
}
```

#### 4.3 Macro-Enabled Parameter Usage

```neuro
# In stdlib/send.neuro
# Received parameters: ${0}=send, ${1}=Hello, ${model}=gpt-4

\set[command_name="${0}"]
\set[user_message="${1}"]  
\set[selected_model="${model}"]

# Use macro expansion for dynamic commands
\set[llm_cmd="\llm-send[model=${selected_model}]"]
${llm_cmd} ${user_message}
```

## 5. Implementation Phases

### Phase 1: Enhanced Interpolation System (Week 1)
1. **Command-Level Interpolation Service**
   - Implement two-phase interpolation (command-level + parameter-level)
   - Add macro expansion capabilities to InterpolationService
   - Test with simple macro examples

2. **Script Execution Pipeline Enhancement** 
   - Integrate command-level interpolation into script executor
   - Update script processing to handle macro expansion
   - Add comprehensive error handling for macro failures

3. **Macro System Testing**
   - Test basic command generation via variables
   - Test conditional command generation
   - Verify parameter interpolation still works correctly

### Phase 2: Enhanced Command Resolution (Week 2)
1. **Priority-Based Command Registry**
   - Implement CommandResolver with priority-based lookup
   - Add ScriptCommand wrapper for embedded scripts
   - Update command registration system

2. **Command Resolution Integration**
   - Integrate enhanced resolution into execution pipeline
   - Test builtin command precedence over scripts
   - Add fallback error handling

### Phase 3: Embedded Script Loading (Week 3)
1. **Embedded Filesystem Integration**
   - Integrate Go embed filesystem
   - Create StdlibLoader service
   - Add script loading capabilities to command registry

2. **Script Parameter System**
   - Implement script parameter variable setup
   - Add parameter cleanup after execution
   - Test parameter passing with macro-enabled scripts

3. **End-to-End Integration Testing**
   - Test complete macro + resolution + embedded script pipeline
   - Verify backward compatibility with existing commands
   - Performance testing for macro expansion overhead

### Phase 4: Standard Library Foundation (Week 4)
1. **Core Infrastructure Scripts**
   - Create basic stdlib scripts leveraging macro system
   - Test complex macro-driven command generation
   - Validate script parameter passing works correctly

2. **Script Management Tools**
   - `\stdlib-list` - List available stdlib scripts
   - `\stdlib-show` - Display script content  
   - `\stdlib-help` - Enhanced help for script commands

## 6. Macro-Enabled Usage Examples

### 6.1 Dynamic Command Generation

**Script with macro-driven commands:**
```neuro
# Dynamic provider selection
\set[provider="${NEURO_PROVIDER}"]  # From environment
\set[model="${provider=openai?gpt-4:claude-3}"]
\set[send_cmd="\llm-send[provider=${provider},model=${model}]"]

# User message from parameter
\set[message="${1}"]

# Generate and execute dynamic command
${send_cmd} ${message}
```

### 6.2 Conditional Workflow Generation

**Debug-aware script:**
```neuro
# Set up conditional commands based on debug mode
\set[debug="${DEBUG_MODE}"]
\set[log_cmd="${debug=true?\echo[style=yellow]:# }"]  
\set[error_cmd="${debug=true?\echo[style=red]:# }"]

# Conditional logging throughout script
${log_cmd} Starting LLM preparation
\llm-prepare

${log_cmd} Sending message: ${1}
\llm-send ${1}

${log_cmd} Saving conversation
\llm-save
```

### 6.3 Template-Based Command Generation

**Workflow template system:**
```neuro
# Define command templates
\set[llm_template="\llm-send[provider=${provider},model=${model},temperature=${temp}]"]
\set[render_template="\render-markdown[style=${theme}]"]

# Configure environment
\set[provider="openai"]
\set[model="gpt-4"]
\set[temp="0.7"]
\set[theme="dark"]

# Generate configured commands
${llm_template} ${1}
${render_template} ${_llm_response}
```

## 7. Technical Specifications

### 7.1 Macro Expansion Rules

```
Expansion Order:
1. Variable resolution: ${var} → value
2. Command concatenation: \echo${var} → \echovalue  
3. Complex expressions: ${var1=value?var2:var3}
4. Nested expansion: ${${prefix}_${suffix}}

Safety Limits:
- Maximum expansion depth: 10 levels
- Maximum expanded line length: 10KB
- Circular reference detection
- Malformed variable syntax error handling
```

### 7.2 Performance Considerations

- **Macro Caching**: Cache expanded macros for repeated patterns
- **Lazy Expansion**: Only expand variables that exist  
- **Early Exit**: Skip expansion if no variables detected in line
- **Memory Limits**: Prevent runaway macro expansion

### 7.3 Error Handling Strategy

- **Macro Errors**: Clear error messages with original line context
- **Circular References**: Detect and report circular macro expansion
- **Missing Variables**: Configurable behavior (error vs empty string)
- **Syntax Errors**: Report both original and expanded command context

## 8. Backward Compatibility

### 8.1 Existing Script Compatibility

- All existing `.neuro` scripts continue to work unchanged
- Macro expansion is additive - no breaking changes
- Existing parameter interpolation remains identical
- Performance impact minimal for non-macro scripts

### 8.2 Migration Strategy

- Deploy macro system alongside existing infrastructure
- Gradually enable macro features in stdlib scripts
- Maintain existing command behavior during transition
- Comprehensive testing of existing script compatibility

---

This enhanced infrastructure transforms NeuroShell into a truly programmable shell platform with macro capabilities, enabling dynamic command generation while maintaining the simplicity and power of the core system. The macro system provides the foundation for sophisticated stdlib scripts that can adapt their behavior based on context, configuration, and user preferences.