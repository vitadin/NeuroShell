# Changelog Entry Template

## Overview

This document provides templates and guidelines for creating consistent changelog entries in `internal/data/embedded/change-logs/change-logs.yaml`. The changelog serves as the single source of truth for both the `\change-log-show` CLI command and GitHub release notes.

## Entry Template

```yaml
  - id: "CL[N]"                    # Sequential ID (CL3, CL4, CL5, etc.)
    version: "[MAJOR.MINOR.PATCH]" # Semantic version (e.g., "0.3.0")
    date: "YYYY-MM-DD"             # Release date
    type: "[TYPE]"                 # See types below
    title: "[Brief Summary]"       # One-line description (max 80 chars)
    description: "[Detailed Description]" # Comprehensive change description
    impact: "[User Impact]"        # How changes affect users
    files_changed: [               # List of modified files/directories
      "path/to/file1.go",
      "path/to/directory/*",
      "specific/file.yaml"
    ]
```

## Field Guidelines

### ID
- **Format**: `CL` followed by sequential number
- **Examples**: `CL3`, `CL4`, `CL5`
- **Requirements**: Must be unique and sequential

### Version
- **Format**: Semantic versioning (MAJOR.MINOR.PATCH)
- **Examples**: `"0.3.0"`, `"0.2.1"`, `"1.0.0"`
- **Requirements**: Must match the release tag (without `v` prefix)

### Date
- **Format**: ISO date format (YYYY-MM-DD)
- **Examples**: `"2025-08-15"`, `"2025-12-25"`
- **Requirements**: Should be the release date

### Type
Available entry types with their meanings:

- **`feature`**: New functionality or capabilities
- **`enhancement`**: Improvements to existing features
- **`bugfix`**: Bug fixes and corrections
- **`performance`**: Performance improvements
- **`security`**: Security-related changes
- **`testing`**: Testing infrastructure improvements
- **`refactor`**: Code refactoring without functional changes
- **`docs`**: Documentation updates
- **`chore`**: Maintenance tasks, dependency updates
- **`breaking`**: Breaking changes (consider major version bump)

### Title
- **Length**: Maximum 80 characters
- **Style**: Sentence case, no trailing period
- **Focus**: Summarize the most important change
- **Examples**:
  - `"Multi-Provider LLM Support and Enhanced Model Management"`
  - `"Session Import/Export with Comprehensive Validation"`
  - `"Performance Optimizations and Memory Usage Improvements"`

### Description
- **Length**: 2-4 sentences recommended
- **Content**: Technical details of changes
- **Focus**: What was implemented/changed
- **Style**: Clear, technical language for developers

### Impact
- **Length**: 1-3 sentences
- **Content**: User-facing benefits and changes
- **Focus**: How users experience the changes
- **Style**: User-friendly language

### Files Changed
- **Format**: Array of file paths or directory patterns
- **Examples**:
  - Specific files: `"internal/services/model_service.go"`
  - Directory patterns: `"internal/commands/llm/*"`
  - Test files: `"test/golden/model-*"`
- **Guidelines**: Include major changed files, use patterns for bulk changes

## Example Entries

### Feature Release Example
```yaml
  - id: "CL3"
    version: "0.3.0"
    date: "2025-08-15"
    type: "feature"
    title: "Advanced Model Parameter Validation and Provider Extensions"
    description: "Implemented comprehensive model parameter validation system with real-time feedback and error handling. Added support for provider-specific parameter constraints and custom validation rules. Enhanced model creation workflow with detailed parameter cards and usage examples."
    impact: "Users can now create models with confidence through guided parameter validation, reducing configuration errors and improving model setup experience. Provider-specific constraints ensure optimal model performance."
    files_changed: [
      "internal/services/parameter_validator_service.go",
      "internal/commands/model/*_new.go",
      "internal/data/embedded/models/*.yaml",
      "test/golden/model-parameter-*"
    ]
```

### Bugfix Release Example
```yaml
  - id: "CL4"
    version: "0.2.3"
    date: "2025-08-12"
    type: "bugfix"
    title: "Session Management Race Condition and Memory Leak Fixes"
    description: "Fixed critical race condition in concurrent session access that could cause data corruption. Resolved memory leak in session cleanup when using streaming responses. Improved error handling for malformed session data."
    impact: "Users experience more stable session management with improved performance and reliability, especially during concurrent operations and long-running streaming sessions."
    files_changed: [
      "internal/services/chat_session_service.go",
      "internal/context/context.go",
      "internal/services/*_client.go"
    ]
```

### Enhancement Release Example  
```yaml
  - id: "CL5"
    version: "0.3.1"
    date: "2025-08-20"
    type: "enhancement"
    title: "Improved CLI Performance and Enhanced Help System"
    description: "Optimized command parsing and variable interpolation for 40% faster CLI response times. Enhanced help system with categorized commands, better examples, and context-aware suggestions. Added command completion for common workflows."
    impact: "Users enjoy significantly faster CLI interactions and can discover features more easily through the improved help system and intelligent command suggestions."
    files_changed: [
      "internal/parser/command.go",
      "internal/services/autocomplete_service.go",
      "internal/commands/builtin/help.go",
      "internal/statemachine/interpolator.go"
    ]
```

## Best Practices

### Writing Guidelines
1. **Be Specific**: Avoid vague terms like "improved" or "fixed issues"
2. **User-Focused**: Explain benefits from the user's perspective in the impact section
3. **Technical Details**: Include enough technical information for developers
4. **Consistent Format**: Follow the template structure exactly
5. **Proper Grammar**: Use clear, grammatically correct English

### Version Planning
1. **Semantic Versioning**: Follow semver guidelines
   - MAJOR: Breaking changes
   - MINOR: New features (backward compatible)
   - PATCH: Bug fixes (backward compatible)
2. **Changelog First**: Write changelog entries before code changes when possible
3. **Release Coordination**: Ensure version matches tag and scripts/version.sh

### File Organization
1. **Newest First**: Add new entries at the top of the entries array
2. **Chronological Order**: Maintain date-based ordering
3. **ID Sequence**: Ensure IDs increment sequentially

## Integration with Release Process

The changelog entries are automatically used by:

1. **`\change-log-show` Command**: Displays formatted changelog in CLI
2. **GitHub Releases**: Extracts relevant entry for release notes
3. **GoReleaser**: Generates release descriptions with changelog content
4. **Documentation**: Provides historical record of project evolution

## Validation

Before committing changelog changes:

1. **Syntax Check**: Ensure YAML is valid
2. **ID Uniqueness**: Verify ID doesn't conflict with existing entries
3. **Version Format**: Confirm semantic versioning format
4. **Date Format**: Check ISO date format (YYYY-MM-DD)
5. **Field Completeness**: Ensure all required fields are present

## Common Mistakes to Avoid

1. **Duplicate IDs**: Always use the next sequential ID
2. **Version Mismatches**: Ensure consistency across tag, changelog, and scripts
3. **Missing Fields**: All template fields should be included
4. **Inconsistent Formatting**: Follow the exact YAML structure
5. **Non-descriptive Titles**: Avoid generic titles like "Bug fixes" or "Updates"
6. **Technical Jargon in Impact**: Keep user impact in plain language

## Tools and Automation

The release pipeline includes validation tools that check:
- YAML syntax correctness
- ID uniqueness and sequence
- Version format compliance
- Required field presence
- Chronological ordering

These tools run automatically during the release process to ensure changelog quality and consistency.