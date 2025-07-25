# `\llm-api-show` Command Design Specification

## Overview
Create a new command `\llm-api-show` that delegates API key discovery to the Configuration Service and stores them in NeuroShell variables with optional variable name filtering.

## Command Design

### 1. Command Structure
- **Name**: `llm-api-show` 
- **Location**: `/internal/commands/builtin/llm_api_show.go`
- **Parse Mode**: `ParseModeKeyValue` for options
- **Integration**: Auto-register in builtin package `init.go`

### 2. Core Functionality
**Delegation**: Let Configuration Service handle all API key searching logic
**Command Role**: 
- Call Configuration Service methods to find API keys
- Map found keys to user-friendly variable names
- Store actual API keys in NeuroShell variables
- Display masked keys for UI

### 3. Command Options
```
\llm-api-show[provider=openai|anthropic|all, format=table|list]
```

**Options:**
- `provider`: Filter by provider (`openai`, `anthropic`, `openrouter`, `moonshot`, `all`) - default: `all`
- `format`: Display format (`table`, `list`) - default: `table`

**Filtering Logic**: Applies only to variable names being displayed/stored, not to API key values

### 4. Implementation Strategy

**Configuration Service Interaction**:
```go
// For each provider, call Configuration Service
providers := []string{"openai", "anthropic", "openrouter", "moonshot"}
for _, provider := range providers {
    apiKey, err := configService.GetAPIKey(provider)
    if err == nil && apiKey != "" {
        // Store in variable: provider + "_key"
        variableName := provider + "_key"
        variableService.Set(variableName, apiKey)
        // Add to display list with masked value
    }
}

// Also check for generic key
if genericKey, err := configService.GetAPIKey(""); err == nil {
    variableService.Set("neuro_key", genericKey)
}
```

### 5. Provider Filtering
**Filter Application**: Only affects which variable names are processed/displayed
- `provider=openai`: Only look for OpenAI keys, only show `openai_key` variable
- `provider=anthropic`: Only look for Anthropic keys, only show `anthropic_key` variable  
- `provider=all`: Look for all provider keys, show all found variables

**Configuration Service Does**: All the searching logic (NEURO_OPENAI_API_KEY, OPENAI_API_KEY, etc.)
**Command Does**: Filter which providers to ask about, map results to variables

### 6. Output Format

**Table Format** (default):
```
LLM API Keys Configuration:
┌─────────────────────────┬─────────────────┐
│ Variable Name           │ API Key         │
├─────────────────────────┼─────────────────┤
│ openai_key              │ sk-...xyz       │
│ anthropic_key           │ sk-...abc       │
│ openrouter_key          │ sk-...def       │
└─────────────────────────┴─────────────────┘

API keys stored in variables. Use ${openai_key}, ${anthropic_key}, etc.
```

**List Format**:
```
LLM API Keys Configuration:
• openai_key: sk-...xyz
• anthropic_key: sk-...abc  
• openrouter_key: sk-...def

API keys stored in variables. Use ${openai_key}, ${anthropic_key}, etc.
```

### 7. Variable Storage & Display
**Storage**: Store actual API keys in user variables (`openai_key`, `anthropic_key`, etc.)
**Display**: Show first 3 and last 3 chars for UI safety (`sk-...xyz`)
**Access**: Users can use `${openai_key}` in subsequent commands

### 8. Implementation Plan

**Step 1**: Create command file with:
- Standard command methods (Name, Description, Usage, HelpInfo)
- Execute method that calls `configService.GetAPIKey(provider)` for each provider
- Variable name mapping logic (`provider` → `provider_key`)
- Display masking (3+3 chars)
- Provider filtering for variable names
- Theme-based formatting

**Step 2**: Tests:
- Unit tests for variable mapping and filtering
- Integration tests calling Configuration Service
- Mock Configuration Service responses
- Variable storage verification

**Step 3**: E2E tests:
- Basic functionality with real config
- Provider filtering
- Format options
- Variable access patterns

**Step 4**: Registration in builtin package

### 9. Division of Responsibilities

**Configuration Service**:
- Find API keys using existing search patterns
- Handle NEURO_ prefixes, legacy formats, etc.
- Return actual API key values

**\llm-api-show Command**:
- Call Configuration Service for each provider
- Apply provider filtering to variable names
- Map results to user-friendly variable names  
- Store actual keys in NeuroShell variables
- Display masked keys with formatting
- Handle UI/UX and user interaction

### 10. User Workflow
1. `\llm-api-show` → shows all found API keys
2. `\llm-api-show[provider=openai]` → shows only OpenAI keys
3. Command stores keys in variables: `openai_key`, `anthropic_key`, etc.
4. User can use `${openai_key}` in other commands
5. Display shows masked keys for security

### 11. Security Considerations
- **Display Masking**: Only show first 3 and last 3 characters in output
- **Variable Storage**: Store complete API keys in variables for functionality
- **No System Variables**: Use user variables without special prefixes
- **Minimum Length**: Only process keys with adequate length (10+ chars)

### 12. Integration Points
- **Configuration Service**: `GetAPIKey(provider)` for each provider
- **Variable Service**: Store actual API keys in user variables
- **Theme Service**: Professional formatting with current theme  
- **Help System**: Comprehensive help and examples

This approach properly delegates API key searching to Configuration Service while keeping the command focused on variable management and user interface.