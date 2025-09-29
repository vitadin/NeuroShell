# Claude Code Integration Design Document

**Version:** 1.0
**Date:** 2025-09-28
**Status:** Draft for Review

## Table of Contents

1. [Overview](#overview)
2. [Architecture Decisions](#architecture-decisions)
3. [Service Architecture](#service-architecture)
4. [Command Design](#command-design)
5. [Variable Integration](#variable-integration)
6. [User Workflows](#user-workflows)
7. [Implementation Plan](#implementation-plan)
8. [Error Handling](#error-handling)
9. [Testing Strategy](#testing-strategy)
10. [Open Questions](#open-questions)

## Overview

This document outlines the integration of Claude Code CLI into NeuroShell, enabling users to leverage Claude Code's capabilities through NeuroShell's command interface while maintaining separation of concerns and respecting both tools' architectures.

### Goals

- **Seamless Integration**: Users can interact with Claude Code through familiar NeuroShell commands
- **Non-Blocking Operations**: Long-running Claude Code operations don't block NeuroShell
- **Variable System Integration**: Claude Code responses integrate with NeuroShell's variable system
- **Session Management**: Proper lifecycle management of Claude Code sessions
- **Authentication Handling**: Secure and flexible API key management

### Requirements

- Claude Code CLI must be pre-installed on the system
- NeuroShell acts as a client to Claude Code's headless mode
- Operations must be asynchronous by default with optional synchronous mode
- Full integration with NeuroShell's variable interpolation system

## Architecture Decisions

### Session Architecture Decision

After analyzing the existing NeuroShell session structure (`ChatSession`), we chose **Separate Claude Code Session Management** for the following reasons:

1. **Clean Separation**: Claude Code sessions have different lifecycle, persistence, and state requirements
2. **No Breaking Changes**: Existing NeuroShell sessions remain unchanged
3. **Flexibility**: Can map between session types when needed
4. **Future-Proof**: Easier to extend without affecting core session functionality

#### Claude Code Session Structure

```go
type ClaudeCodeSession struct {
    ID                string    // Internal NeuroShell ID
    ClaudeSessionID   string    // Claude Code's session ID
    NeuroSessionID    string    // Optional: linked NeuroShell session
    Status            string    // idle|busy|error
    CurrentJobID      string    // Active job if any
    LastResponse      string    // Last complete response
    WorkingDirectory  string    // CC working directory
    Model             string    // Active model
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

#### Relationship with NeuroShell Sessions

- **Independent**: Claude Code sessions can exist without NeuroShell sessions
- **Linkable**: Optional linking via `NeuroSessionID` field
- **Persistent**: Claude Code session state persists across NeuroShell restarts
- **Discoverable**: All CC sessions are listable and manageable

### Authentication Architecture

Claude Code authentication is handled through multiple layers:

1. **Environment Variables**: `ANTHROPIC_API_KEY`
2. **NeuroShell Variables**: `${os.ANTHROPIC_API_KEY}`
3. **Secure Storage**: Platform keychain when available
4. **Session-Only Keys**: Temporary keys for specific sessions

## Service Architecture

### ClaudeCodeService

Located at `internal/services/claudecode_service.go`:

```go
type ClaudeCodeService struct {
    // Process management
    process        *exec.Cmd
    stdin          io.WriteCloser
    stdout         io.ReadCloser
    stderr         io.ReadCloser

    // State management
    sessions       map[string]*ClaudeCodeSession
    jobs           map[string]*ClaudeCodeJob
    authConfig     *ClaudeCodeAuth

    // Control
    initialized    bool
    mu             sync.RWMutex

    // Communication
    responseChans  map[string]chan *ClaudeCodeResponse
    eventBus       chan *ClaudeCodeEvent
}

type ClaudeCodeJob struct {
    ID         string
    SessionID  string
    Command    string
    Input      string
    Status     JobStatus
    Output     *strings.Builder
    StreamChan chan string
    Error      error
    StartedAt  time.Time
    EndedAt    *time.Time
    Progress   float64
}

type ClaudeCodeAuth struct {
    Method    string    // env|variable|keychain|session
    KeySource string    // Where the key comes from
    IsValid   bool      // Last validation result
    ExpiresAt *time.Time
}
```

### Service Responsibilities

1. **Process Lifecycle**: Start/stop/monitor Claude Code daemon
2. **Session Management**: Create, track, and cleanup Claude Code sessions
3. **Job Queue**: Manage async operations with proper queuing
4. **Authentication**: Handle API key validation and rotation
5. **Output Streaming**: Real-time output streaming for long operations
6. **Error Recovery**: Handle process crashes and connection failures

### ClaudeCodeSubcontext

Following NeuroShell's pattern of focused interfaces:

```go
type ClaudeCodeSubcontext interface {
    // Session management
    InitSession(opts InitOptions) (*ClaudeCodeSession, error)
    GetSession(id string) (*ClaudeCodeSession, error)
    GetActiveSession() (*ClaudeCodeSession, error)
    SetActiveSession(id string) error
    ListSessions() []*ClaudeCodeSession
    DeleteSession(id string) error

    // Job management
    SubmitJob(sessionID, command string, opts JobOptions) (*ClaudeCodeJob, error)
    GetJob(jobID string) (*ClaudeCodeJob, error)
    ListJobs(filter JobFilter) []*ClaudeCodeJob
    WaitForJob(jobID string, timeout time.Duration) error
    CancelJob(jobID string) error

    // Output handling
    StreamJob(jobID string) (<-chan string, error)
    GetJobOutput(jobID string) (string, error)
    GetJobStatus(jobID string) (JobStatus, error)

    // Authentication
    CheckAuth() (*ClaudeCodeAuth, error)
    SetAuth(key string, persistent bool) error
    ClearAuth() error
}
```

## Command Design

Following NeuroShell's established patterns, with emphasis on simplicity and bash-command-like behavior.

### Core Commands

#### Session Initialization

```neuro
# Initialize Claude Code (lazy init on first use)
\cc-init[model="sonnet", verbose=false, dirs="./src,./tests", permission_mode="plan"]
# Sets: ${#cc_status}, ${#cc_session_id}, ${#cc_ready}
```

**Options:**
- `model`: Claude model to use (sonnet|opus|haiku)
- `verbose`: Enable verbose output
- `dirs`: Additional working directories
- `permission_mode`: Claude Code permission mode (ask|plan|auto)

**Variables Set:**
- `${#cc_status}`: Service status (ready|error|starting)
- `${#cc_session_id}`: Active Claude Code session ID
- `${#cc_ready}`: Boolean indicating readiness

#### Basic Message Sending

```neuro
# Send message (async by default)
\cc Explain the authentication flow in this codebase
# Sets: ${#cc_job_id}, starts background processing

# Send message (synchronous mode)
\cc[wait=true, timeout=30] What files implement user authentication?
# Blocks until complete, sets: ${_cc_output}, ${_cc_tools_used}
```

**Options:**
- `wait`: Block until completion (default: false)
- `timeout`: Timeout in seconds for wait mode
- `stream`: Show streaming output (default: false with wait=true)

**Variables Set:**
- `${#cc_job_id}`: Job ID for tracking
- `${_cc_output}`: Complete response (when done)
- `${_cc_tools_used}`: JSON array of tools used
- `${_cc_files_modified}`: List of files modified

#### Job Management

```neuro
# Check job status
\cc-status[job="${#cc_job_id}"]
# Sets: ${_cc_job_status}, ${_cc_job_progress}

# Wait for job completion
\cc-wait[job="${#cc_job_id}", timeout=60]
# Blocks until done, sets: ${_cc_output}, ${_cc_error}

# Get job output (non-blocking)
\cc-get[job="${#cc_job_id}"]
# Sets: ${_cc_output}, ${_cc_status}, ${_cc_partial}

# Stream job output
\cc-stream[job="${#cc_job_id}", follow=true]
# Displays streaming output, updates ${_cc_stream}

# List active jobs
\cc-jobs[status="running", session="${#cc_session_id}"]
# Sets: ${_cc_active_jobs} as JSON array
```

#### Conversation Management

```neuro
# Continue previous conversation
\cc-continue Please add error handling to that function
# Uses current session context, sets: ${#cc_job_id}

# Resume specific Claude Code session
\cc-resume[session="abc123"] Continue working on the refactor
# Loads previous context, sets: ${#cc_session_id}

# Start new conversation thread
\cc-new[name="feature-auth"] Analyze the authentication system
# Creates new session, sets: ${#cc_session_id}
```

### Authentication Commands

```neuro
# Check authentication status
\cc-auth
# Sets: ${#cc_auth_status}, ${#cc_auth_method}, ${#cc_auth_valid}

# Configure API key (persistent)
\cc-auth[key="${api_key}", persist=true]
# Stores in keychain/config, sets: ${#cc_auth_status}

# Use temporary key (session only)
\cc-auth[key="${temp_key}", session_only=true]
# Only for current session

# Clear stored authentication
\cc-auth[clear=true]
# Removes stored credentials
```

### Slash Command Execution

Following the bash command pattern for simplicity:

```neuro
# Execute Claude Code slash commands (simple pattern)
\cc-slash review-pr 123
# Direct execution like \bash, sets: ${#cc_job_id}

# With additional arguments
\cc-slash[wait=true] commit "Add new authentication feature"
# Block until complete

# Pass complex arguments
\cc-slash search-files "authentication.*\\.py"
# Search with regex pattern
```

**Supported Slash Commands:**
- `review-pr <number>`: Review pull request
- `commit <message>`: Create commit with message
- `search-files <pattern>`: Search for files matching pattern
- `explain <file>`: Explain specific file
- `fix <issue>`: Fix specific issue
- `test <component>`: Generate tests for component

### Advanced Commands

#### Input Piping

```neuro
# Pipe content to Claude (like bash)
\cat config.yaml | \cc-pipe Explain this configuration
# Sets: ${#cc_pipe_job_id}

# Pipe with processing options
\bash[git diff HEAD~1] | \cc-pipe[wait=true] Review these changes
# Synchronous processing of git diff
```

#### Configuration Management

```neuro
# Configure Claude Code settings
\cc-config[permission_mode="plan", allowed_tools="Read,Bash(ls:*)"]
# Updates current session configuration

# Set global defaults
\cc-config[global=true, model="opus", verbose=true]
# Affects new sessions

# Show current configuration
\cc-config[show=true]
# Sets: ${_cc_config} as JSON
```

#### Session Management

```neuro
# List all Claude Code sessions
\cc-sessions[active_only=false]
# Sets: ${_cc_sessions} as JSON array

# Switch active session
\cc-switch[session="session-id"]
# Changes active CC session

# Link to NeuroShell session
\cc-link[neuro_session="${#session_id}"]
# Associates CC session with NeuroShell session

# Export session data
\cc-export[session="${#cc_session_id}", file="session.json", format="claude"]
# Saves CC session in Claude Code format

# Import session data
\cc-import[file="session.json"]
# Loads CC session, sets: ${#cc_session_id}
```

#### Utility Commands

```neuro
# Cancel running job
\cc-cancel[job="${#cc_job_id}"]
# Cancels specific job

# Kill all jobs in session
\cc-kill[session="${#cc_session_id}"]
# Emergency stop

# Health check
\cc-health
# Sets: ${#cc_health}, ${#cc_version}, ${#cc_uptime}

# Extract code blocks from response
\cc-extract[job="${#cc_job_id}", type="code", lang="python"]
# Sets: ${_cc_code_blocks} as array

# Extract file modifications
\cc-extract[job="${#cc_job_id}", type="files"]
# Sets: ${_cc_modified_files} as array
```

## Variable Integration

### Variable Naming Convention

Claude Code integration follows NeuroShell's systematic variable naming:

| Prefix | Type | Examples | Description |
|--------|------|----------|-------------|
| `#cc_*` | Metadata | `${#cc_session_id}`, `${#cc_job_id}` | System metadata |
| `_cc_*` | Output | `${_cc_output}`, `${_cc_error}` | Command outputs |
| `@cc_*` | System | `${@cc_version}`, `${@cc_uptime}` | System information |

### Core Variables

#### Session Variables
- `${#cc_session_id}`: Active Claude Code session ID
- `${#cc_status}`: Overall service status (ready|busy|error|starting)
- `${#cc_ready}`: Boolean indicating service readiness
- `${#cc_model}`: Currently active model

#### Job Variables
- `${#cc_job_id}`: Last submitted job ID
- `${#cc_job_status}`: Job status (pending|running|completed|failed|cancelled)
- `${#cc_job_progress}`: Job progress percentage (0-100)
- `${#cc_active_jobs}`: Count of active jobs

#### Authentication Variables
- `${#cc_auth_status}`: Authentication status (valid|invalid|missing|expired)
- `${#cc_auth_method}`: How authentication was provided (env|keychain|variable)
- `${#cc_auth_valid}`: Boolean indicating valid authentication

#### Output Variables
- `${_cc_output}`: Last complete response from Claude Code
- `${_cc_error}`: Last error message
- `${_cc_stream}`: Current streaming output buffer
- `${_cc_partial}`: Partial response (when job still running)

#### Tool and File Variables
- `${_cc_tools_used}`: JSON array of tools used in last response
- `${_cc_files_modified}`: Array of files modified by Claude Code
- `${_cc_code_blocks}`: Extracted code blocks from response
- `${_cc_file_diffs}`: File changes made

#### System Variables
- `${@cc_version}`: Claude Code CLI version
- `${@cc_uptime}`: Service uptime in seconds
- `${@cc_pid}`: Process ID of Claude Code daemon

### Variable Lifecycle

1. **Initialization**: Core variables set on `\cc-init`
2. **Job Submission**: Job variables set on command execution
3. **Progress Updates**: Status variables updated during execution
4. **Completion**: Output variables populated when job finishes
5. **Cleanup**: Temporary variables cleared on session end

## User Workflows

### Workflow 1: Basic Interactive Development

```neuro
# Start a development session
\session-new[name="auth-feature"]
\cc-init[model="sonnet", dirs="./backend,./frontend"]

# Ask for analysis
\cc Analyze the current authentication system and suggest improvements
\cc-stream[follow=true]  # Watch response in real-time

# Continue the conversation
\cc-continue Can you implement the OAuth2 integration you suggested?
\cc-wait[timeout=120]   # Wait for implementation

# Review the changes
\echo Claude modified: ${_cc_files_modified}
\cc-extract[type="code", lang="python"]
\echo Found ${#cc_code_blocks_count} code blocks

# Save the session
\session-save[name="auth-with-claude"]
```

### Workflow 2: Parallel Code Reviews

```neuro
# Set up for multiple reviews
\cc-init[model="opus", permission_mode="plan"]

# Start multiple review jobs
\cc-slash review-pr 101
\set[review1="${#cc_job_id}"]

\cc-slash review-pr 102
\set[review2="${#cc_job_id}"]

\cc-slash review-pr 103
\set[review3="${#cc_job_id}"]

# Monitor progress
\cc-jobs[status="running"]
\echo Active reviews: ${_cc_active_jobs}

# Collect results as they complete
\cc-get[job="${review1}"]
\if[condition="${_cc_job_status}" equals="completed"]
  \write[file="review-101.md"] ${_cc_output}
\endif

\cc-get[job="${review2}"]
\if[condition="${_cc_job_status}" equals="completed"]
  \write[file="review-102.md"] ${_cc_output}
\endif

# Wait for all to complete
\cc-wait[job="${review3}"]
\write[file="review-103.md"] ${_cc_output}
```

### Workflow 3: Integration with NeuroShell LLM

```neuro
# Use both NeuroShell LLM and Claude Code together
\session-new[name="feature-planning"]

# Plan with NeuroShell LLM
\send Create a detailed plan for implementing user role management

# Use Claude Code for implementation
\cc-init
\cc-link[neuro_session="${#session_id}"]
\cc Based on this plan: "${1}", implement the user role system

# Iterate between both
\send Review the implementation: "${_cc_output}"
\cc-continue Address these review comments: "${1}"

# Save combined session
\session-save[name="role-management-complete"]
```

### Workflow 4: Automated Documentation Generation

```neuro
#!/usr/bin/env neuro
# Script: generate-docs.neuro

\cc-init[model="sonnet"]

# Get all Python files
\bash[find ./src -name "*.py" -type f | head -10]
\set[files="${_output}"]

# Process each file
\foreach[var="file", in="${files}"]
  \echo Processing: ${file}
  \cat ${file} | \cc-pipe Generate comprehensive documentation for this Python module
  \cc-wait[timeout=60]

  # Extract the file name for output
  \bash[basename ${file} .py]
  \set[filename="${_output}"]

  \write[file="docs/${filename}.md"] ${_cc_output}
\endfor

\echo Documentation generation complete for ${#files_count} files
```

### Workflow 5: Code Refactoring Assistant

```neuro
# Refactoring session with incremental improvements
\cc-init[model="opus", permission_mode="ask"]

# Start with analysis
\cc Analyze the code structure in ./src/auth/ and identify refactoring opportunities
\cc-wait

# Review suggestions
\render-markdown[display_only=true] ${_cc_output}

# Implement specific refactoring
\cc-continue Please refactor the UserService class to use dependency injection
\cc-stream[follow=true]

# Validate changes
\bash[python -m pytest tests/test_auth.py]
\if[condition="${@status}" equals="0"]
  \echo ✅ Tests pass after refactoring
  \cc-continue Great! Now please refactor the AuthController following the same pattern
\else
  \echo ❌ Tests failed, reverting changes
  \bash[git checkout -- src/auth/]
\endif
```

### Workflow 6: Research and Implementation

```neuro
# Research phase with Claude Code
\cc-init[dirs="./docs,./src,./tests"]

# Research best practices
\cc Research modern authentication patterns for Python web applications. Consider our current FastAPI setup.
\cc-wait

# Store research
\write[file="research/auth-patterns.md"] ${_cc_output}

# Implementation planning
\cc Based on the research, create a detailed implementation plan for upgrading our auth system
\cc-wait

# Extract action items
\cc-extract[type="tasks"]
\write[file="tasks/auth-upgrade.json"] ${_cc_extracted_tasks}

# Begin implementation
\cc Let's start with the first task. Implement JWT token refresh mechanism.
\cc-stream[follow=true]
```

## Implementation Plan

### Phase 1: Foundation (Week 1)

**Goal**: Basic Claude Code integration with synchronous operations

#### Tasks:
- [ ] Create `ClaudeCodeService` in `internal/services/`
- [ ] Implement process lifecycle management (start/stop/monitor)
- [ ] Add `\cc-init` command with basic options
- [ ] Implement synchronous `\cc` command
- [ ] Basic variable storage (`${_cc_output}`, `${#cc_session_id}`)
- [ ] Error handling for missing Claude Code installation

#### Deliverables:
- Working `\cc-init` and `\cc` commands
- Basic process management
- Synchronous message sending and response handling
- Initial test suite

#### Success Criteria:
- User can send messages to Claude Code and receive responses
- Proper error messages when Claude Code is not installed
- Variables are correctly set and accessible

### Phase 2: Authentication (Week 1)

**Goal**: Secure and flexible authentication handling

#### Tasks:
- [ ] Implement `\cc-auth` command suite
- [ ] Environment variable checking (`ANTHROPIC_API_KEY`)
- [ ] NeuroShell variable integration (`${os.ANTHROPIC_API_KEY}`)
- [ ] Secure credential storage (keychain/config)
- [ ] Authentication validation and status checking
- [ ] Session-only authentication support

#### Deliverables:
- Complete authentication system
- Secure key storage
- Auth status checking
- Documentation for key management

#### Success Criteria:
- Users can configure API keys securely
- Multiple authentication methods work correctly
- Clear feedback on authentication status

### Phase 3: Async Operations (Week 2)

**Goal**: Non-blocking operations with job tracking

#### Tasks:
- [ ] Implement job queue system
- [ ] Add background job monitoring
- [ ] Create `\cc-wait`, `\cc-get`, `\cc-status` commands
- [ ] Implement output streaming (`\cc-stream`)
- [ ] Job cancellation and cleanup
- [ ] Progress tracking and reporting

#### Deliverables:
- Async job system
- Job monitoring commands
- Streaming output support
- Job lifecycle management

#### Success Criteria:
- Long-running operations don't block NeuroShell
- Users can monitor job progress
- Multiple concurrent jobs work correctly

### Phase 4: Session Management (Week 2)

**Goal**: Complete session lifecycle and persistence

#### Tasks:
- [ ] Multiple Claude Code session support
- [ ] Session persistence across restarts
- [ ] `\cc-sessions`, `\cc-switch` commands
- [ ] Integration with NeuroShell sessions (`\cc-link`)
- [ ] Session import/export functionality
- [ ] Session cleanup and garbage collection

#### Deliverables:
- Multi-session support
- Session persistence
- NeuroShell integration
- Session management UI

#### Success Criteria:
- Users can manage multiple Claude Code sessions
- Sessions persist across NeuroShell restarts
- Clean integration with NeuroShell session system

### Phase 5: Advanced Features (Week 3)

**Goal**: Power user features and automation support

#### Tasks:
- [ ] `\cc-slash` command for Claude slash commands
- [ ] `\cc-pipe` for input piping
- [ ] `\cc-config` for permissions and settings
- [ ] `\cc-extract` for parsing responses
- [ ] Advanced variable integration
- [ ] Batch operation support

#### Deliverables:
- Slash command support
- Input piping system
- Configuration management
- Response parsing tools

#### Success Criteria:
- All Claude Code slash commands are accessible
- Complex workflows can be automated
- Response data is easily extractable

### Phase 6: Polish and Documentation (Week 3)

**Goal**: Production-ready integration with excellent UX

#### Tasks:
- [ ] Auto-completion for Claude Code commands
- [ ] Progress indicators and status displays
- [ ] Error recovery and retry mechanisms
- [ ] Performance optimization
- [ ] Comprehensive documentation
- [ ] Integration test suite
- [ ] User guide and examples

#### Deliverables:
- Polished user experience
- Complete documentation
- Robust error handling
- Performance benchmarks

#### Success Criteria:
- Commands have helpful auto-completion
- Clear progress indication for long operations
- Comprehensive error messages and recovery
- Full documentation with examples

### Development Milestones

#### Milestone 1 (End of Week 1): Basic Integration
- Users can send messages to Claude Code
- Authentication is properly configured
- Basic error handling works

#### Milestone 2 (End of Week 2): Async Operations
- Non-blocking operations with job tracking
- Multiple concurrent Claude Code sessions
- Session persistence

#### Milestone 3 (End of Week 3): Feature Complete
- All planned commands implemented
- Advanced features working
- Production-ready quality

#### Milestone 4 (End of Week 4): Production Ready
- Complete documentation
- Performance optimized
- Comprehensive test coverage

## Error Handling

### Error Categories

#### 1. Installation and Environment Errors

**Claude Code Not Installed:**
```neuro
\cc-init
# Error: Claude Code CLI not found. Please install it first:
# npm install -g @anthropic/claude-code
# Sets: ${_cc_error} = "claude_not_installed"
```

**Version Compatibility:**
```neuro
\cc-init
# Error: Claude Code version 1.2.0 is not supported. Please upgrade to 1.5.0+
# Sets: ${_cc_error} = "version_incompatible"
```

#### 2. Authentication Errors

**Missing API Key:**
```neuro
\cc-init
# Error: No Anthropic API key found. Configure with: \cc-auth[key="your-key"]
# Sets: ${_cc_error} = "auth_missing", ${#cc_auth_status} = "missing"
```

**Invalid API Key:**
```neuro
\cc Hello world
# Error: Invalid API key. Please check your credentials with \cc-auth
# Sets: ${_cc_error} = "auth_invalid", ${#cc_auth_status} = "invalid"
```

**Expired API Key:**
```neuro
\cc Hello world
# Error: API key has expired. Please update your credentials.
# Sets: ${_cc_error} = "auth_expired", ${#cc_auth_status} = "expired"
```

#### 3. Process Management Errors

**Process Startup Failure:**
```neuro
\cc-init
# Error: Failed to start Claude Code process. Check permissions and installation.
# Sets: ${_cc_error} = "process_start_failed"
```

**Process Crash:**
```neuro
\cc Hello world
# Error: Claude Code process crashed. Attempting restart...
# Auto-restart attempted, sets: ${_cc_error} = "process_crashed"
```

**Communication Timeout:**
```neuro
\cc[wait=true, timeout=10] Complex analysis task
# Error: Operation timed out after 10 seconds
# Sets: ${_cc_error} = "timeout", ${#cc_job_status} = "timeout"
```

#### 4. Job Management Errors

**Job Not Found:**
```neuro
\cc-get[job="invalid-job-id"]
# Error: Job 'invalid-job-id' not found
# Sets: ${_cc_error} = "job_not_found"
```

**Job Cancellation:**
```neuro
\cc-cancel[job="${#cc_job_id}"]
# Success: Job cancelled successfully
# Sets: ${#cc_job_status} = "cancelled"
```

**Resource Limits:**
```neuro
\cc Start 10 complex analysis tasks
# Error: Too many concurrent jobs (limit: 5). Wait for jobs to complete.
# Sets: ${_cc_error} = "job_limit_exceeded"
```

#### 5. Session Management Errors

**Session Not Found:**
```neuro
\cc-switch[session="invalid-session"]
# Error: Claude Code session 'invalid-session' not found
# Sets: ${_cc_error} = "session_not_found"
```

**Session Corruption:**
```neuro
\cc-resume[session="corrupted-session"]
# Error: Session data corrupted, creating new session
# Sets: ${_cc_error} = "session_corrupted", creates new session
```

### Error Recovery Strategies

#### 1. Automatic Recovery

**Process Restart:**
- Detect process crashes through health checks
- Automatically restart Claude Code process
- Restore active sessions when possible
- Notify user of restart and any lost state

**Authentication Refresh:**
- Detect expired tokens
- Attempt to refresh from stored credentials
- Fallback to environment variables
- Prompt user for new credentials if needed

**Connection Recovery:**
- Implement exponential backoff for connection retries
- Queue operations during temporary disconnections
- Resume queued operations after reconnection

#### 2. User-Guided Recovery

**Configuration Issues:**
```neuro
\cc-health
# Diagnosis: Configuration issue detected
# Suggestion: Run \cc-config[reset=true] to restore defaults
```

**Permission Problems:**
```neuro
\cc Modify the authentication system
# Error: Permission denied for file modification
# Suggestion: Run \cc-config[permission_mode="ask"] to enable interactive permissions
```

#### 3. Graceful Degradation

**Limited Functionality:**
- When Claude Code unavailable, provide helpful error messages
- Suggest alternative NeuroShell commands when appropriate
- Maintain NeuroShell functionality independent of Claude Code

**Partial Results:**
- Save partial outputs when jobs are interrupted
- Allow users to resume from last successful state
- Provide options to retry or start fresh

### Error Logging and Debugging

#### Debug Mode

```neuro
\cc-init[verbose=true, debug=true]
# Enables detailed logging for troubleshooting
# Sets: ${#cc_debug_mode} = true
```

#### Log Access

```neuro
\cc-logs[lines=50, level="error"]
# Shows recent error logs
# Sets: ${_cc_logs} with log entries
```

#### Health Checks

```neuro
\cc-health[detailed=true]
# Comprehensive system check
# Sets: ${_cc_health_report} with detailed status
```

## Testing Strategy

### Unit Testing

#### Service Layer Tests

**ClaudeCodeService Tests:**
```go
func TestClaudeCodeService_ProcessLifecycle(t *testing.T)
func TestClaudeCodeService_JobManagement(t *testing.T)
func TestClaudeCodeService_SessionManagement(t *testing.T)
func TestClaudeCodeService_AuthenticationHandling(t *testing.T)
func TestClaudeCodeService_ErrorRecovery(t *testing.T)
```

**Job Queue Tests:**
```go
func TestJobQueue_ConcurrentJobs(t *testing.T)
func TestJobQueue_Prioritization(t *testing.T)
func TestJobQueue_Cancellation(t *testing.T)
func TestJobQueue_ResourceLimits(t *testing.T)
```

#### Command Tests

**Command Execution Tests:**
```go
func TestCCInitCommand_BasicInit(t *testing.T)
func TestCCCommand_SyncMessage(t *testing.T)
func TestCCCommand_AsyncMessage(t *testing.T)
func TestCCAuthCommand_KeyManagement(t *testing.T)
func TestCCSlashCommand_Execution(t *testing.T)
```

**Variable Integration Tests:**
```go
func TestCCCommands_VariableSettings(t *testing.T)
func TestCCCommands_VariableInterpolation(t *testing.T)
func TestCCCommands_ErrorVariables(t *testing.T)
```

### Integration Testing

#### Claude Code Process Integration

**Real Process Tests:**
```go
func TestIntegration_ClaudeCodeStartup(t *testing.T)
func TestIntegration_MessageSending(t *testing.T)
func TestIntegration_SessionPersistence(t *testing.T)
func TestIntegration_SlashCommands(t *testing.T)
```

**Mock Process Tests:**
```go
func TestIntegration_MockedClaudeCode(t *testing.T)
func TestIntegration_ErrorScenarios(t *testing.T)
func TestIntegration_TimeoutHandling(t *testing.T)
```

#### Authentication Integration

**API Key Tests:**
```go
func TestIntegration_EnvironmentAuth(t *testing.T)
func TestIntegration_KeychainAuth(t *testing.T)
func TestIntegration_VariableAuth(t *testing.T)
func TestIntegration_AuthValidation(t *testing.T)
```

### End-to-End Testing

#### Workflow Tests

**Golden File Tests:**
```bash
# Test basic workflow
./bin/neurotest record cc-basic-workflow
./bin/neurotest run cc-basic-workflow

# Test async operations
./bin/neurotest record cc-async-operations
./bin/neurotest run cc-async-operations

# Test error scenarios
./bin/neurotest record cc-error-handling
./bin/neurotest run cc-error-handling
```

**Performance Tests:**
```go
func TestE2E_ConcurrentJobs(t *testing.T)
func TestE2E_LongRunningOperations(t *testing.T)
func TestE2E_MemoryUsage(t *testing.T)
func TestE2E_ResponseTime(t *testing.T)
```

#### Real-World Scenarios

**Development Workflow Tests:**
- Code analysis and suggestion workflow
- Refactoring assistance workflow
- Documentation generation workflow
- Code review workflow

**Error Recovery Tests:**
- Process crash recovery
- Network disconnection recovery
- Invalid input handling
- Resource exhaustion scenarios

### Test Infrastructure

#### Mock Claude Code Service

```go
type MockClaudeCodeService struct {
    responses map[string]string
    delays    map[string]time.Duration
    errors    map[string]error
}

func (m *MockClaudeCodeService) SendMessage(msg string) (*Response, error)
func (m *MockClaudeCodeService) SimulateSlowResponse(delay time.Duration)
func (m *MockClaudeCodeService) SimulateError(err error)
```

#### Test Fixtures

```
tests/
├── fixtures/
│   ├── claude_responses/
│   │   ├── code_analysis.json
│   │   ├── refactoring_suggestions.json
│   │   └── error_responses.json
│   ├── sessions/
│   │   ├── basic_session.json
│   │   └── complex_session.json
│   └── configs/
│       ├── test_config.json
│       └── auth_configs.json
```

#### Continuous Integration

**Test Matrix:**
- Go versions: 1.21, 1.22
- Operating systems: Ubuntu, macOS, Windows
- Claude Code versions: Latest stable, Latest beta

**Test Stages:**
1. **Fast Tests**: Unit tests, basic integration (< 2 minutes)
2. **Medium Tests**: Complex integration, mock E2E (< 10 minutes)
3. **Slow Tests**: Real Claude Code E2E (< 30 minutes)
4. **Nightly Tests**: Performance, stress testing (< 2 hours)

## Open Questions

### Technical Architecture

1. **Process Lifecycle**: Should the Claude Code daemon persist across NeuroShell sessions, or start fresh each time?
   - **Consideration**: Persistent daemon maintains context but uses more resources
   - **Recommendation**: Configurable behavior with persistent as default

2. **Session State Synchronization**: How should we handle Claude Code session state when NeuroShell restarts?
   - **Consideration**: Full state sync vs. lightweight reconnection
   - **Recommendation**: Store session IDs and attempt reconnection, fall back to new session

3. **Concurrent Request Handling**: Should we limit concurrent Claude Code operations to prevent API rate limiting?
   - **Consideration**: User productivity vs. API compliance
   - **Recommendation**: Configurable limits with sensible defaults (5 concurrent jobs)

### User Experience

4. **Output Formatting**: Should we automatically parse and format Claude's responses (code blocks, markdown)?
   - **Consideration**: Convenience vs. raw output control
   - **Recommendation**: Provide both options - raw output by default, formatting commands available

5. **Progress Indication**: How detailed should progress feedback be for long-running operations?
   - **Consideration**: Information richness vs. UI clutter
   - **Recommendation**: Minimal by default with verbose option

6. **History Integration**: Should Claude Code conversations appear in NeuroShell's `\history` command?
   - **Consideration**: Unified history vs. tool separation
   - **Recommendation**: Separate histories with optional unified view

### Integration Boundaries

7. **Error Boundary**: When Claude Code fails, should it affect NeuroShell's stability?
   - **Consideration**: Isolation vs. tight integration
   - **Recommendation**: Complete isolation - Claude Code failures never crash NeuroShell

8. **Configuration Management**: Should Claude Code settings be managed through NeuroShell config or remain separate?
   - **Consideration**: Unified configuration vs. tool autonomy
   - **Recommendation**: Hybrid approach - auth and integration settings in NeuroShell, Claude Code settings remain separate

9. **Model Selection**: Should model selection be per-session or global?
   - **Consideration**: Flexibility vs. simplicity
   - **Recommendation**: Per-session with global default

### Performance and Scalability

10. **Memory Management**: How should we handle memory usage for long-running sessions with large outputs?
    - **Consideration**: Memory usage vs. history preservation
    - **Recommendation**: Configurable output retention with automatic cleanup

11. **Response Caching**: Should we cache Claude Code responses for identical inputs?
    - **Consideration**: Performance vs. freshness of responses
    - **Recommendation**: Optional caching for expensive operations

12. **Resource Monitoring**: Should we monitor and report Claude Code resource usage?
    - **Consideration**: Transparency vs. complexity
    - **Recommendation**: Basic monitoring available via `\cc-health`

### Future Extensions

13. **Plugin Architecture**: Should we design for third-party Claude Code extensions?
    - **Consideration**: Extensibility vs. security
    - **Recommendation**: Design with extension points but implement security carefully

14. **Multi-User Support**: How should Claude Code integration work in multi-user NeuroShell environments?
    - **Consideration**: User isolation vs. shared resources
    - **Recommendation**: Per-user authentication and session isolation

15. **Alternative AI Tools**: Should the architecture support other AI tools beyond Claude Code?
    - **Consideration**: Generalization vs. specialization
    - **Recommendation**: Design interfaces that could support other tools but focus on Claude Code initially

---

## Conclusion

This design document provides a comprehensive foundation for integrating Claude Code into NeuroShell. The proposed architecture maintains clean separation between the tools while enabling powerful combined workflows.

Key principles guiding this design:

1. **Respect Tool Boundaries**: Each tool maintains its core identity and capabilities
2. **Async-First**: Long operations never block the user interface
3. **Variable Integration**: Seamless data flow between tools through NeuroShell's variable system
4. **Error Resilience**: Failures in one tool don't affect the other
5. **User Choice**: Flexible authentication and configuration options

The implementation plan provides a clear path from basic integration to advanced features, ensuring users can benefit from the integration early while development continues.

This document should be reviewed and refined based on team feedback before implementation begins.
