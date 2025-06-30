package shellintegration

import (
	"testing"
	"time"
)

func TestCommandTracker_CreateSession(t *testing.T) {
	tracker := NewCommandTracker()

	sessionName := "test-session"
	state := tracker.CreateSession(sessionName)

	if state.SessionName != sessionName {
		t.Errorf("CreateSession() SessionName = %v, want %v", state.SessionName, sessionName)
	}

	if state.CurrentState != StateIdle {
		t.Errorf("CreateSession() CurrentState = %v, want %v", state.CurrentState, StateIdle)
	}

	if !state.IsActive {
		t.Error("CreateSession() session should be active")
	}

	if state.Parser == nil {
		t.Error("CreateSession() should initialize parser")
	}

	// Verify session is stored
	retrieved, exists := tracker.GetSession(sessionName)
	if !exists {
		t.Error("CreateSession() session not found after creation")
	}

	if retrieved != state {
		t.Error("CreateSession() retrieved session does not match created session")
	}
}

func TestCommandTracker_GetSession(t *testing.T) {
	tracker := NewCommandTracker()

	// Test non-existent session
	_, exists := tracker.GetSession("non-existent")
	if exists {
		t.Error("GetSession() should return false for non-existent session")
	}

	// Create and retrieve session
	sessionName := "test-session"
	created := tracker.CreateSession(sessionName)

	retrieved, exists := tracker.GetSession(sessionName)
	if !exists {
		t.Error("GetSession() should return true for existing session")
	}

	if retrieved != created {
		t.Error("GetSession() should return the same session instance")
	}
}

func TestCommandTracker_RemoveSession(t *testing.T) {
	tracker := NewCommandTracker()

	// Test removing non-existent session
	removed := tracker.RemoveSession("non-existent")
	if removed {
		t.Error("RemoveSession() should return false for non-existent session")
	}

	// Create and remove session
	sessionName := "test-session"
	tracker.CreateSession(sessionName)

	removed = tracker.RemoveSession(sessionName)
	if !removed {
		t.Error("RemoveSession() should return true for existing session")
	}

	// Verify session is removed
	_, exists := tracker.GetSession(sessionName)
	if exists {
		t.Error("RemoveSession() session should not exist after removal")
	}
}

func TestCommandTracker_ProcessOutput(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test processing output for non-existent session
	_, err := tracker.ProcessOutput(sessionName, []byte("test"))
	if err == nil {
		t.Error("ProcessOutput() should return error for non-existent session")
	}

	// Create session and process output
	tracker.CreateSession(sessionName)

	tests := []struct {
		name     string
		input    []byte
		expected CommandState
		complete bool
	}{
		{
			name:     "Regular output",
			input:    []byte("hello world\n"),
			expected: StateIdle,
			complete: false,
		},
		{
			name:     "Prompt start",
			input:    []byte("\033]133;A\007"),
			expected: StatePromptStart,
			complete: false,
		},
		{
			name:     "Command start",
			input:    []byte("\033]133;B\007"),
			expected: StateCommandStart,
			complete: false,
		},
		{
			name:     "Output start",
			input:    []byte("\033]133;C\007"),
			expected: StateOutputStart,
			complete: false,
		},
		{
			name:     "Command completion",
			input:    []byte("\033]133;D;0\007"),
			expected: StateCommandEnd,
			complete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tracker.ProcessOutput(sessionName, tt.input)
			if err != nil {
				t.Errorf("ProcessOutput() error = %v", err)
				return
			}

			if result.State != tt.expected {
				t.Errorf("ProcessOutput() State = %v, want %v", result.State, tt.expected)
			}

			if result.IsComplete != tt.complete {
				t.Errorf("ProcessOutput() IsComplete = %v, want %v", result.IsComplete, tt.complete)
			}

			if result.SessionName != sessionName {
				t.Errorf("ProcessOutput() SessionName = %v, want %v", result.SessionName, sessionName)
			}
		})
	}
}

func TestCommandTracker_IsCommandRunning(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test non-existent session
	running := tracker.IsCommandRunning("non-existent")
	if running {
		t.Error("IsCommandRunning() should return false for non-existent session")
	}

	// Create session
	tracker.CreateSession(sessionName)

	// Initially should not be running
	running = tracker.IsCommandRunning(sessionName)
	if running {
		t.Error("IsCommandRunning() should return false for idle session")
	}

	// Set command start state
	tracker.ProcessOutput(sessionName, []byte("\033]133;B\007"))
	running = tracker.IsCommandRunning(sessionName)
	if !running {
		t.Error("IsCommandRunning() should return true for command start state")
	}

	// Set output start state
	tracker.ProcessOutput(sessionName, []byte("\033]133;C\007"))
	running = tracker.IsCommandRunning(sessionName)
	if !running {
		t.Error("IsCommandRunning() should return true for output start state")
	}

	// Complete command
	tracker.ProcessOutput(sessionName, []byte("\033]133;D;0\007"))
	running = tracker.IsCommandRunning(sessionName)
	if running {
		t.Error("IsCommandRunning() should return false for completed command")
	}
}

func TestCommandTracker_IsCommandComplete(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test non-existent session
	complete := tracker.IsCommandComplete("non-existent")
	if complete {
		t.Error("IsCommandComplete() should return false for non-existent session")
	}

	// Create session
	tracker.CreateSession(sessionName)

	// Initially should be complete (idle)
	complete = tracker.IsCommandComplete(sessionName)
	if !complete {
		t.Error("IsCommandComplete() should return true for idle session")
	}

	// Start command
	tracker.ProcessOutput(sessionName, []byte("\033]133;B\007"))
	complete = tracker.IsCommandComplete(sessionName)
	if complete {
		t.Error("IsCommandComplete() should return false for running command")
	}

	// Complete command
	tracker.ProcessOutput(sessionName, []byte("\033]133;D;0\007"))
	complete = tracker.IsCommandComplete(sessionName)
	if !complete {
		t.Error("IsCommandComplete() should return true for completed command")
	}
}

func TestCommandTracker_GetLastExitCode(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test non-existent session
	_, exists := tracker.GetLastExitCode("non-existent")
	if exists {
		t.Error("GetLastExitCode() should return false for non-existent session")
	}

	// Create session
	tracker.CreateSession(sessionName)

	// Initially should have exit code 0
	exitCode, exists := tracker.GetLastExitCode(sessionName)
	if !exists {
		t.Error("GetLastExitCode() should return true for existing session")
	}
	if exitCode != 0 {
		t.Errorf("GetLastExitCode() initial exit code = %v, want 0", exitCode)
	}

	// Complete command with exit code 1
	tracker.ProcessOutput(sessionName, []byte("\033]133;D;1\007"))
	exitCode, exists = tracker.GetLastExitCode(sessionName)
	if !exists {
		t.Error("GetLastExitCode() should return true for existing session after command")
	}
	if exitCode != 1 {
		t.Errorf("GetLastExitCode() exit code = %v, want 1", exitCode)
	}
}

func TestCommandTracker_OutputBuffer(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test non-existent session
	_, exists := tracker.GetOutputBuffer("non-existent")
	if exists {
		t.Error("GetOutputBuffer() should return false for non-existent session")
	}

	// Create session
	tracker.CreateSession(sessionName)

	// Initially should have empty buffer
	buffer, exists := tracker.GetOutputBuffer(sessionName)
	if !exists {
		t.Error("GetOutputBuffer() should return true for existing session")
	}
	if len(buffer) != 0 {
		t.Errorf("GetOutputBuffer() initial buffer length = %v, want 0", len(buffer))
	}

	// Add some output
	tracker.ProcessOutput(sessionName, []byte("line1\n"))
	tracker.ProcessOutput(sessionName, []byte("line2\n"))

	buffer, exists = tracker.GetOutputBuffer(sessionName)
	if !exists {
		t.Error("GetOutputBuffer() should return true for existing session")
	}
	if len(buffer) != 2 {
		t.Errorf("GetOutputBuffer() buffer length = %v, want 2", len(buffer))
	}

	// Clear buffer
	err := tracker.ClearOutputBuffer(sessionName)
	if err != nil {
		t.Errorf("ClearOutputBuffer() error = %v", err)
	}

	buffer, exists = tracker.GetOutputBuffer(sessionName)
	if !exists {
		t.Error("GetOutputBuffer() should return true for existing session after clear")
	}
	if len(buffer) != 0 {
		t.Errorf("GetOutputBuffer() buffer length after clear = %v, want 0", len(buffer))
	}

	// Test clear non-existent session
	err = tracker.ClearOutputBuffer("non-existent")
	if err == nil {
		t.Error("ClearOutputBuffer() should return error for non-existent session")
	}
}

func TestCommandTracker_GetSessionInfo(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test non-existent session
	_, err := tracker.GetSessionInfo("non-existent")
	if err == nil {
		t.Error("GetSessionInfo() should return error for non-existent session")
	}

	// Create session
	tracker.CreateSession(sessionName)

	info, err := tracker.GetSessionInfo(sessionName)
	if err != nil {
		t.Errorf("GetSessionInfo() error = %v", err)
		return
	}

	if info.SessionName != sessionName {
		t.Errorf("GetSessionInfo() SessionName = %v, want %v", info.SessionName, sessionName)
	}

	if info.CurrentState != StateIdle {
		t.Errorf("GetSessionInfo() CurrentState = %v, want %v", info.CurrentState, StateIdle)
	}

	if !info.IsActive {
		t.Error("GetSessionInfo() session should be active")
	}

	if info.IsRunning {
		t.Error("GetSessionInfo() session should not be running initially")
	}
}

func TestCommandTracker_ListSessions(t *testing.T) {
	tracker := NewCommandTracker()

	// Initially should have no sessions
	sessions := tracker.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("ListSessions() initial count = %v, want 0", len(sessions))
	}

	// Create multiple sessions
	sessionNames := []string{"session1", "session2", "session3"}
	for _, name := range sessionNames {
		tracker.CreateSession(name)
	}

	sessions = tracker.ListSessions()
	if len(sessions) != len(sessionNames) {
		t.Errorf("ListSessions() count = %v, want %v", len(sessions), len(sessionNames))
	}

	// Verify all sessions are listed
	found := make(map[string]bool)
	for _, session := range sessions {
		found[session.SessionName] = true
	}

	for _, name := range sessionNames {
		if !found[name] {
			t.Errorf("ListSessions() missing session %v", name)
		}
	}
}

func TestCommandTracker_Reset(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"

	// Test reset non-existent session
	err := tracker.Reset("non-existent")
	if err == nil {
		t.Error("Reset() should return error for non-existent session")
	}

	// Create session and add some state
	tracker.CreateSession(sessionName)
	tracker.ProcessOutput(sessionName, []byte("test output\n"))
	tracker.ProcessOutput(sessionName, []byte("\033]133;B\007"))

	// Verify state exists
	info, _ := tracker.GetSessionInfo(sessionName)
	if info.CurrentState == StateIdle {
		t.Error("Session should have non-idle state before reset")
	}

	// Reset session
	err = tracker.Reset(sessionName)
	if err != nil {
		t.Errorf("Reset() error = %v", err)
	}

	// Verify state is reset
	info, _ = tracker.GetSessionInfo(sessionName)
	if info.CurrentState != StateIdle {
		t.Errorf("Reset() CurrentState = %v, want %v", info.CurrentState, StateIdle)
	}

	if info.LastExitCode != 0 {
		t.Errorf("Reset() LastExitCode = %v, want 0", info.LastExitCode)
	}

	if info.OutputLines != 0 {
		t.Errorf("Reset() OutputLines = %v, want 0", info.OutputLines)
	}
}

func TestSessionInfo_LastCommandDuration(t *testing.T) {
	tracker := NewCommandTracker()
	sessionName := "test-session"
	tracker.CreateSession(sessionName)

	// Simulate command execution timing
	tracker.ProcessOutput(sessionName, []byte("\033]133;B\007"))
	time.Sleep(10 * time.Millisecond) // Small delay
	tracker.ProcessOutput(sessionName, []byte("\033]133;D;0\007"))

	info, err := tracker.GetSessionInfo(sessionName)
	if err != nil {
		t.Errorf("GetSessionInfo() error = %v", err)
		return
	}

	if info.LastCommandDuration <= 0 {
		t.Error("GetSessionInfo() LastCommandDuration should be positive")
	}

	if info.LastCommandDuration > time.Second {
		t.Error("GetSessionInfo() LastCommandDuration seems too long")
	}
}
