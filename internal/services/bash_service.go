// Package services provides business logic services for NeuroShell operations.
// BashService manages PTY-based bash sessions for command execution.
package services

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// BashService manages PTY-based bash sessions for command execution.
type BashService struct {
	initialized bool
}

// NewBashService creates a new BashService instance.
func NewBashService() *BashService {
	return &BashService{
		initialized: false,
	}
}

// Name returns the service name for registration.
func (b *BashService) Name() string {
	return "bash"
}

// Initialize initializes the bash service.
func (b *BashService) Initialize(_ neurotypes.Context) error {
	b.initialized = true
	return nil
}

// ExecuteCommand executes a command in the specified bash session.
func (b *BashService) ExecuteCommand(sessionName, command string, options BashOptions, ctx neurotypes.Context) (string, error) {
	if !b.initialized {
		return "", fmt.Errorf("bash service not initialized")
	}

	// Cast to NeuroContext to access bash session methods
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	// Get or create session
	session, err := b.getOrCreateSession(sessionName, options, neuroCtx)
	if err != nil {
		return "", fmt.Errorf("failed to get or create session: %w", err)
	}

	// Execute command in session
	output, err := b.executeInSession(session, command, options)
	if err != nil {
		return "", fmt.Errorf("failed to execute command in session %s: %w", sessionName, err)
	}

	// Update session last used time
	session.LastUsed = time.Now()

	return output, nil
}

// BashOptions contains options for bash command execution.
type BashOptions struct {
	SessionName   string            // Session name (default: "default")
	ForceNew      bool              // Force create new session
	Timeout       time.Duration     // Command timeout
	Environment   map[string]string // Environment variables
	WorkingDir    string            // Working directory
	Interactive   bool              // Interactive mode
	CaptureOutput bool              // Capture output to variables
}

// CreateSession creates a new bash session with the given name and options.
func (b *BashService) CreateSession(sessionName string, options BashOptions, ctx neurotypes.Context) error {
	if !b.initialized {
		return fmt.Errorf("bash service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	// Check if session already exists and force new is not set
	if !options.ForceNew {
		if _, exists := neuroCtx.GetBashSession(sessionName); exists {
			return fmt.Errorf("session %s already exists, use new=true to force create", sessionName)
		}
	}

	// Create new session
	session, err := b.createBashSession(sessionName, options)
	if err != nil {
		return fmt.Errorf("failed to create bash session: %w", err)
	}

	// Store session in context (replacing if exists)
	neuroCtx.SetBashSession(sessionName, session)

	return nil
}

// GetSession retrieves a bash session by name.
func (b *BashService) GetSession(sessionName string, ctx neurotypes.Context) (*context.BashSession, error) {
	if !b.initialized {
		return nil, fmt.Errorf("bash service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	session, exists := neuroCtx.GetBashSession(sessionName)
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	return session, nil
}

// KillSession terminates a bash session by name.
func (b *BashService) KillSession(sessionName string, ctx neurotypes.Context) error {
	if !b.initialized {
		return fmt.Errorf("bash service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	removed := neuroCtx.RemoveBashSession(sessionName)
	if !removed {
		return fmt.Errorf("session %s not found", sessionName)
	}

	return nil
}

// ListSessions returns a list of all active bash session names.
func (b *BashService) ListSessions(ctx neurotypes.Context) ([]string, error) {
	if !b.initialized {
		return nil, fmt.Errorf("bash service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.ListBashSessions(), nil
}

// getOrCreateSession gets an existing session or creates a new one.
func (b *BashService) getOrCreateSession(sessionName string, options BashOptions, ctx *context.NeuroContext) (*context.BashSession, error) {
	// Try to get existing session
	if session, exists := ctx.GetBashSession(sessionName); exists && !options.ForceNew {
		// Check if session is still active
		if session.Active && session.Process != nil && session.Process.Process != nil {
			return session, nil
		}
		// Session is not active, remove it and create new one
		ctx.RemoveBashSession(sessionName)
	}

	// Create new session
	session, err := b.createBashSession(sessionName, options)
	if err != nil {
		return nil, err
	}

	// Store session in context
	ctx.SetBashSession(sessionName, session)

	return session, nil
}

// createBashSession creates a new bash session with PTY.
func (b *BashService) createBashSession(sessionName string, options BashOptions) (*context.BashSession, error) {
	// Create bash command
	cmd := exec.Command("bash")

	// Set up environment
	env := os.Environ()
	if options.Environment != nil {
		for key, value := range options.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	cmd.Env = env

	// Set working directory
	if options.WorkingDir != "" {
		cmd.Dir = options.WorkingDir
	}

	// Start with PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	// Create session
	session := &context.BashSession{
		Name:        sessionName,
		PTY:         ptmx,
		Process:     cmd,
		Environment: make(map[string]string),
		WorkingDir:  options.WorkingDir,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		Active:      true,
	}

	// Copy environment
	if options.Environment != nil {
		for key, value := range options.Environment {
			session.Environment[key] = value
		}
	}

	return session, nil
}

// executeInSession executes a command in an existing bash session.
func (b *BashService) executeInSession(session *context.BashSession, command string, options BashOptions) (string, error) {
	session.LastUsed = time.Now()

	// Write command to PTY
	_, err := session.PTY.WriteString(command + "\n")
	if err != nil {
		return "", fmt.Errorf("failed to write command to PTY: %w", err)
	}

	// Read output with timeout
	if options.Timeout > 0 {
		return b.readWithTimeout(session.PTY, options.Timeout)
	}

	return b.readOutput(session.PTY)
}

// readOutput reads output from PTY until a reasonable stopping point.
func (b *BashService) readOutput(reader io.Reader) (string, error) {
	var output strings.Builder
	scanner := bufio.NewScanner(reader)

	// Set a reasonable timeout to avoid hanging
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		// Read a reasonable amount of output
		lineCount := 0
		for scanner.Scan() && lineCount < 100 {
			line := scanner.Text()
			output.WriteString(line)
			output.WriteString("\n")
			lineCount++
		}
		done <- true
	}()

	select {
	case <-done:
		return output.String(), nil
	case <-timeout:
		return output.String(), nil // Return what we have so far
	}
}

// readWithTimeout reads output from PTY with a specific timeout.
func (b *BashService) readWithTimeout(reader io.Reader, timeout time.Duration) (string, error) {
	var output strings.Builder
	scanner := bufio.NewScanner(reader)

	timeoutChan := time.After(timeout)
	done := make(chan bool)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line)
			output.WriteString("\n")
		}
		done <- true
	}()

	select {
	case <-done:
		return output.String(), nil
	case <-timeoutChan:
		return output.String(), fmt.Errorf("command timed out after %v", timeout)
	}
}
