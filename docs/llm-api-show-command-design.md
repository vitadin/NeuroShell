# `\llm-api-show` Command Design Specification

## Overview
Create a new command `\llm-api-show` that uses an enhanced Configuration Service to collect API keys from multiple sources and stores them with source prefixes. Users explicitly activate keys via `\llm-api-activate` command, with active keys stored as metadata variables.

## Command Design

### 1. Command Structure
- **Name**: `llm-api-show` 
- **Location**: `/internal/commands/builtin/llm_api_show.go`
- **Parse Mode**: `ParseModeKeyValue` for options
- **Integration**: Auto-register in builtin package `init.go`

### 2. Core Functionality
**Configuration Service Role**: Pure collector - scan all sources, collect keys containing provider names
**Command Role**: 
- Display collected keys with source attribution
- Show which keys are currently active
- Provide user interface for key management

### 3. Command Options
```
\llm-api-show[provider=openai|anthropic|all]
```

**Options:**
- `provider`: Filter by provider (`openai`, `anthropic`, `openrouter`, `moonshot`, `all`) - default: `all`

**Note**: Removed `format` option per requirements

### 4. Enhanced Configuration Service Architecture

**Multi-Source Collection Strategy**:
```go
type APIKeySource struct {
    Source       string // "os", "config", "local"
    OriginalName string // "A_OPENAI_KEY", "OPENAI_API_KEY"
    Value        string // actual key
    Provider     string // "openai" (detected)
}

// Configuration Service scans three sources:
// (a) OS environment variables - scan all env vars, collect if contains provider name
// (b) Config folder .env file (~/.neuroshell/.env)
// (c) Local .env file (./.env)

func (cs *ConfigurationService) GetAllAPIKeys() ([]APIKeySource, error) {
    providers := []string{"openai", "anthropic", "openrouter", "moonshot"}
    var keys []APIKeySource
    
    // Scan OS environment
    for _, env := range os.Environ() {
        for _, provider := range providers {
            if strings.Contains(strings.ToLower(env), provider) {
                keys = append(keys, APIKeySource{
                    Source: "os",
                    OriginalName: extractName(env),
                    Value: extractValue(env),
                    Provider: provider,
                })
            }
        }
    }
    
    // Scan config and local .env files similarly
    return keys, nil
}
```

### 5. Variable Storage Strategy

**Source-Prefixed Collection Variables**:
```
os.A_OPENAI_KEY = "sk-123..."           # From OS env A_OPENAI_KEY
os.OPENAI_API_KEY = "sk-456..."         # From OS env OPENAI_API_KEY  
config.OPENAI_API_KEY = "sk-789..."     # From config .env file
local.ANTHROPIC_KEY = "sk-abc..."       # From local .env file
os.MY_OPENAI_WORK_KEY = "sk-def..."     # From OS env MY_OPENAI_WORK_KEY
```

**Active Key Metadata Variables (# prefix)**:
```
#active_openai_key = "os.A_OPENAI_KEY"      # Points to which source variable
#active_anthropic_key = "local.ANTHROPIC_KEY" 
#active_moonshot_key = "config.MOONSHOT_API_KEY"
```

**User Access Pattern**:
- Direct: `${#active_openai_key}` resolves to actual API key value
- Alias: `\set[my_key="${#active_openai_key}"]` for convenience

### 6. Output Format

**Table Format**:
```
LLM API Keys Found:
┌─────────────────────────┬─────────────────┬────────┐
│ Variable Name           │ API Key         │ Status │
├─────────────────────────┼─────────────────┼────────┤
│ os.A_OPENAI_KEY         │ sk-...123       │        │
│ os.OPENAI_API_KEY       │ sk-...456       │ ACTIVE │
│ config.OPENAI_API_KEY   │ sk-...789       │        │
│ local.ANTHROPIC_KEY     │ sk-...abc       │ ACTIVE │
│ os.MY_OPENAI_WORK_KEY   │ sk-...def       │        │
└─────────────────────────┴─────────────────┴────────┘

Active keys: ${#active_openai_key}, ${#active_anthropic_key}
Activate key: \llm-api-activate[provider=openai, key=os.A_OPENAI_KEY]
```

**Provider Filtering**:
```bash
> \llm-api-show[provider=openai]
LLM API Keys Found (OpenAI only):
┌─────────────────────────┬─────────────────┬────────┐
│ Variable Name           │ API Key         │ Status │
├─────────────────────────┼─────────────────┼────────┤
│ os.A_OPENAI_KEY         │ sk-...123       │        │
│ os.OPENAI_API_KEY       │ sk-...456       │ ACTIVE │
│ config.OPENAI_API_KEY   │ sk-...789       │        │
│ os.MY_OPENAI_WORK_KEY   │ sk-...def       │        │
└─────────────────────────┴─────────────────┴────────┘

Active key: ${#active_openai_key}
```

### 7. API Key Activation Command

**New Command: `\llm-api-activate`**
```bash
\llm-api-activate[provider=openai, key=os.A_OPENAI_KEY]
```

**Execution Flow**:
```bash
> \llm-api-activate[provider=openai, key=os.A_OPENAI_KEY]
✓ Activated os.A_OPENAI_KEY for OpenAI provider
  ${#active_openai_key} = "os.A_OPENAI_KEY"
  Access via: ${#active_openai_key} or set alias: \set[my_key="${#active_openai_key}"]
```

**Implementation**:
- Validates that the specified key exists in collected variables
- Sets `#active_{provider}_key` metadata variable to point to source variable
- Provides clear feedback on activation

### 8. Implementation Plan

**Phase 1: Enhanced Configuration Service**
- Extend `GetAllAPIKeys()` method to scan three sources (OS env, config .env, local .env)
- Implement provider name detection in variable names
- Return `APIKeySource` structs with source attribution
- Add validation for minimum key length (10+ chars)

**Phase 2: \llm-api-show Command**  
- Create command with provider filtering
- Store collected keys as `{source}.{ORIGINAL_NAME}` variables
- Display table with source attribution and active status
- Show masked API keys (first 3 + last 3 chars)
- Check for existing `#active_{provider}_key` metadata to show status

**Phase 3: \llm-api-activate Command**
- Create activation command with provider and key parameters
- Validate key exists in collected variables
- Set `#active_{provider}_key` metadata variable
- Provide clear user feedback

**Phase 4: Testing**
- Unit tests for Configuration Service multi-source collection
- Integration tests for both commands
- E2E tests with mock environment variables and .env files
- Variable resolution and access pattern tests

### 9. Division of Responsibilities

**Enhanced Configuration Service**:
- **Pure Collector**: Scan OS env vars, config .env, local .env
- **No Prioritization**: Collect all keys containing provider names
- **Source Attribution**: Return keys with source labels (os/config/local)
- **No Activation Logic**: Does not decide which key is active

**\llm-api-show Command**:
- Call Configuration Service to get all collected keys
- Store keys as `{source}.{ORIGINAL_NAME}` variables
- Display with source attribution and active status
- Apply provider filtering for display
- Show masked keys for security

**\llm-api-activate Command**:
- **User Control**: Let users explicitly choose active keys
- Set `#active_{provider}_key` metadata variables
- Validate key existence before activation
- Provide clear activation feedback

### 10. User Workflow
1. `\llm-api-show` → displays all collected API keys with source attribution
2. `\llm-api-show[provider=openai]` → shows only OpenAI keys  
3. Keys automatically stored as `{source}.{ORIGINAL_NAME}` variables
4. `\llm-api-activate[provider=openai, key=os.A_OPENAI_KEY]` → set active key
5. Use active keys via `${#active_openai_key}` in commands
6. Create aliases: `\set[my_key="${#active_openai_key}"]` for convenience

### 11. Security Considerations
- **Display Masking**: Show first 3 and last 3 characters only (`sk-...xyz`)
- **Full Storage**: Store complete API keys in variables for functionality
- **Source Transparency**: Clear attribution of where keys come from
- **Minimum Length**: Only process keys with adequate length (10+ chars)
- **Explicit Activation**: Users must explicitly choose active keys

### 12. Integration Points
- **Enhanced Configuration Service**: `GetAllAPIKeys()` for multi-source collection
- **Variable Service**: Store both collected keys and active metadata variables
- **Theme Service**: Professional table formatting with current theme  
- **Help System**: Comprehensive help for both commands

### 13. Architecture Benefits
- **Separation of Concerns**: Collection vs. activation clearly separated
- **Full Transparency**: Users see exactly what's found where
- **Explicit Control**: No hidden priority logic, users choose active keys
- **Professional UX**: Similar to AWS profiles, Git contexts
- **Flexible Naming**: Support for descriptive environment variable names
- **Easy Debugging**: Clear source attribution for troubleshooting

This architecture provides a clean separation between discovery (Configuration Service) and management (user commands), giving users full control over their API key configuration.