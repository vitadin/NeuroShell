// Package shellintegration provides command state tracking for shell integration
package shellintegration

import (
	"fmt"
	"sync"
	"time"
)

// CommandTracker tracks the state of commands in bash sessions
type CommandTracker struct {
	mutex    sync.RWMutex
	sessions map[string]*SessionState
}

// SessionState represents the state of a bash session
type SessionState struct {
	SessionName    string
	CurrentState   CommandState
	LastCommand    string
	LastExitCode   int
	CommandStarted time.Time
	CommandEnded   time.Time
	OutputBuffer   []string
	IsActive       bool
	Parser         *StreamParser
}

// NewCommandTracker creates a new command tracker
func NewCommandTracker() *CommandTracker {
	return &CommandTracker{
		sessions: make(map[string]*SessionState),
	}
}

// CreateSession creates a new session state
func (t *CommandTracker) CreateSession(sessionName string) *SessionState {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	state := &SessionState{
		SessionName:  sessionName,
		CurrentState: StateIdle,
		IsActive:     true,
		Parser:       NewStreamParser(),
		OutputBuffer: make([]string, 0),
	}

	t.sessions[sessionName] = state
	return state
}

// GetSession returns the session state for the given session name
func (t *CommandTracker) GetSession(sessionName string) (*SessionState, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	return state, exists
}

// RemoveSession removes a session from tracking
func (t *CommandTracker) RemoveSession(sessionName string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.sessions[sessionName]; exists {
		delete(t.sessions, sessionName)
		return true
	}
	return false
}

// ProcessOutput processes output for a session and updates its state
func (t *CommandTracker) ProcessOutput(sessionName string, data []byte) (*ProcessResult, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	// Parse the output
	parseResult := state.Parser.ParseOutput(data)

	// Update session state based on parse result
	previousState := state.CurrentState
	state.CurrentState = parseResult.State

	// Handle state transitions
	if err := t.handleStateTransition(state, previousState, parseResult); err != nil {
		return nil, fmt.Errorf("failed to handle state transition: %w", err)
	}

	// Add output to buffer if there's new content
	if parseResult.HasNewOutput && parseResult.Output != "" {
		state.OutputBuffer = append(state.OutputBuffer, parseResult.Output)
	}

	result := &ProcessResult{
		SessionName:  sessionName,
		Output:       parseResult.Output,
		NewOutput:    parseResult.NewOutput,
		State:        parseResult.State,
		IsComplete:   parseResult.IsComplete,
		ExitCode:     parseResult.ExitCode,
		HasNewOutput: parseResult.HasNewOutput,
		StateChanged: previousState != parseResult.State,
	}

	return result, nil
}

// ProcessResult contains the result of processing output for a session
type ProcessResult struct {
	SessionName  string
	Output       string
	NewOutput    string
	State        CommandState
	IsComplete   bool
	ExitCode     int
	HasNewOutput bool
	StateChanged bool
}

// handleStateTransition handles state transitions and updates session metadata
func (t *CommandTracker) handleStateTransition(state *SessionState, previousState CommandState, parseResult ParseResult) error {
	switch state.CurrentState {
	case StateCommandStart:
		if previousState != StateCommandStart {
			state.CommandStarted = time.Now()
			state.OutputBuffer = make([]string, 0) // Clear previous output
		}

	case StateCommandEnd:
		if previousState != StateCommandEnd {
			state.CommandEnded = time.Now()
			state.LastExitCode = parseResult.ExitCode
		}

	case StatePromptStart:
		// Ready for next command - could clear output buffer here if desired
		// For now, we'll keep the output for potential capture

	case StateOutputStart:
		// Command is generating output

	default:
		// StateIdle or unknown state
	}

	return nil
}

// IsCommandRunning returns true if the session has a command currently running
func (t *CommandTracker) IsCommandRunning(sessionName string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return false
	}

	return state.CurrentState == StateCommandStart || state.CurrentState == StateOutputStart
}

// IsCommandComplete returns true if the last command in the session is complete
func (t *CommandTracker) IsCommandComplete(sessionName string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return false
	}

	return state.CurrentState == StateCommandEnd || state.CurrentState == StatePromptStart || state.CurrentState == StateIdle
}

// GetLastExitCode returns the exit code of the last completed command
func (t *CommandTracker) GetLastExitCode(sessionName string) (int, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return 0, false
	}

	return state.LastExitCode, true
}

// GetOutputBuffer returns the accumulated output for a session
func (t *CommandTracker) GetOutputBuffer(sessionName string) ([]string, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return nil, false
	}

	// Return a copy of the buffer
	buffer := make([]string, len(state.OutputBuffer))
	copy(buffer, state.OutputBuffer)

	return buffer, true
}

// ClearOutputBuffer clears the output buffer for a session
func (t *CommandTracker) ClearOutputBuffer(sessionName string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return fmt.Errorf("session %s not found", sessionName)
	}

	state.OutputBuffer = make([]string, 0)
	return nil
}

// GetSessionInfo returns detailed information about a session
func (t *CommandTracker) GetSessionInfo(sessionName string) (*SessionInfo, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	info := &SessionInfo{
		SessionName:    state.SessionName,
		CurrentState:   state.CurrentState,
		LastCommand:    state.LastCommand,
		LastExitCode:   state.LastExitCode,
		CommandStarted: state.CommandStarted,
		CommandEnded:   state.CommandEnded,
		IsActive:       state.IsActive,
		OutputLines:    len(state.OutputBuffer),
		IsRunning:      state.CurrentState == StateCommandStart || state.CurrentState == StateOutputStart,
	}

	if !state.CommandStarted.IsZero() && !state.CommandEnded.IsZero() {
		info.LastCommandDuration = state.CommandEnded.Sub(state.CommandStarted)
	}

	return info, nil
}

// SessionInfo contains information about a session
type SessionInfo struct {
	SessionName         string
	CurrentState        CommandState
	LastCommand         string
	LastExitCode        int
	CommandStarted      time.Time
	CommandEnded        time.Time
	LastCommandDuration time.Duration
	IsActive            bool
	OutputLines         int
	IsRunning           bool
}

// ListSessions returns information about all tracked sessions
func (t *CommandTracker) ListSessions() []SessionInfo {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	sessions := make([]SessionInfo, 0, len(t.sessions))

	for _, state := range t.sessions {
		info := SessionInfo{
			SessionName:    state.SessionName,
			CurrentState:   state.CurrentState,
			LastCommand:    state.LastCommand,
			LastExitCode:   state.LastExitCode,
			CommandStarted: state.CommandStarted,
			CommandEnded:   state.CommandEnded,
			IsActive:       state.IsActive,
			OutputLines:    len(state.OutputBuffer),
			IsRunning:      state.CurrentState == StateCommandStart || state.CurrentState == StateOutputStart,
		}

		if !state.CommandStarted.IsZero() && !state.CommandEnded.IsZero() {
			info.LastCommandDuration = state.CommandEnded.Sub(state.CommandStarted)
		}

		sessions = append(sessions, info)
	}

	return sessions
}

// Reset resets the state for a session
func (t *CommandTracker) Reset(sessionName string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	state, exists := t.sessions[sessionName]
	if !exists {
		return fmt.Errorf("session %s not found", sessionName)
	}

	state.CurrentState = StateIdle
	state.LastCommand = ""
	state.LastExitCode = 0
	state.CommandStarted = time.Time{}
	state.CommandEnded = time.Time{}
	state.OutputBuffer = make([]string, 0)
	state.Parser.Reset()

	return nil
}
