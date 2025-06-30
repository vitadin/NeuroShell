// Package services provides business logic services for NeuroShell operations.
// BashService manages PTY-based bash sessions for command execution.
package services

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
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

	// Initialize the bash session by waiting for it to be ready
	err = b.initializeBashSession(session)
	if err != nil {
		// Clean up session on initialization failure
		session.Active = false
		if session.PTY != nil {
			_ = session.PTY.Close()
		}
		if session.Process != nil && session.Process.Process != nil {
			_ = session.Process.Process.Kill()
			_ = session.Process.Wait()
		}
		return nil, fmt.Errorf("failed to initialize bash session: %w", err)
	}

	return session, nil
}

// initializeBashSession waits for bash to be ready and clears any startup output.
func (b *BashService) initializeBashSession(session *context.BashSession) error {
	logger.Debug("Initializing bash session", "session", session.Name)

	// Give bash a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Send a simple command to test if bash is ready and clear startup messages
	initMarker := fmt.Sprintf("NEURO_INIT_%d", time.Now().UnixNano())
	initCommand := fmt.Sprintf("echo '%s'\n", initMarker)

	logger.Debug("Sending initialization command", "command", initCommand)

	_, err := session.PTY.WriteString(initCommand)
	if err != nil {
		return fmt.Errorf("failed to write init command: %w", err)
	}

	// Read until we see our init marker, discarding all startup output
	scanner := bufio.NewScanner(session.PTY)
	timeout := time.After(5 * time.Second)
	done := make(chan bool)
	errChan := make(chan error)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logger.Debug("Init: read line", "line", line)

			if strings.Contains(line, initMarker) {
				logger.Debug("Bash session initialized successfully", "marker", initMarker)
				done <- true
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		} else {
			errChan <- fmt.Errorf("PTY closed during initialization")
		}
	}()

	select {
	case <-done:
		logger.Debug("Bash session ready", "session", session.Name)
		return nil
	case err := <-errChan:
		return fmt.Errorf("initialization failed: %w", err)
	case <-timeout:
		return fmt.Errorf("bash session initialization timed out")
	}
}

// executeInSession executes a command in an existing bash session.
func (b *BashService) executeInSession(session *context.BashSession, command string, options BashOptions) (string, error) {
	session.LastUsed = time.Now()

	// Generate unique marker for command completion detection
	marker := fmt.Sprintf("NEURO_CMD_DONE_%d_%d", time.Now().UnixNano(), rand.Intn(10000))

	// Create command sequence with proper separation and newlines
	// Use explicit newlines to ensure commands execute separately
	commandSequence := fmt.Sprintf("%s\necho '%s'\n", command, marker)

	logger.Debug("Writing command to PTY", "session", session.Name, "command", command, "sequence", commandSequence)

	_, err := session.PTY.WriteString(commandSequence)
	if err != nil {
		return "", fmt.Errorf("failed to write command to PTY: %w", err)
	}

	// Read output until we see the marker
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second // Default 2-second timeout instead of 5
	}

	return b.readUntilMarker(session.PTY, marker, timeout)
}

// readUntilMarker reads output from PTY until the completion marker is found.
func (b *BashService) readUntilMarker(reader io.Reader, marker string, timeout time.Duration) (string, error) {
	var output strings.Builder
	scanner := bufio.NewScanner(reader)

	timeoutChan := time.After(timeout)
	done := make(chan string)
	errChan := make(chan error)

	logger.Debug("Starting to read PTY output", "marker", marker, "timeout", timeout)

	go func() {
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++

			logger.Debug("Read PTY line", "line_num", lineCount, "content", line)

			// Check if this line contains our completion marker
			if strings.Contains(line, marker) {
				logger.Debug("Found completion marker, command finished", "marker", marker, "total_lines", lineCount)
				// Command completed - don't include the marker line in output
				done <- output.String()
				return
			}

			// Add line to output
			if output.Len() > 0 {
				output.WriteString("\n")
			}
			output.WriteString(line)
		}

		// If scanner exits (shouldn't happen in normal PTY usage)
		if err := scanner.Err(); err != nil {
			logger.Debug("Scanner error", "error", err)
			errChan <- err
		} else {
			logger.Debug("PTY closed unexpectedly")
			errChan <- fmt.Errorf("PTY closed unexpectedly")
		}
	}()

	select {
	case result := <-done:
		cleanedResult := b.cleanOutput(result)
		logger.Debug("Command completed successfully", "raw_output", result, "cleaned_output", cleanedResult)
		return cleanedResult, nil
	case err := <-errChan:
		logger.Debug("PTY read error", "error", err, "partial_output", output.String())
		return output.String(), fmt.Errorf("PTY read error: %w", err)
	case <-timeoutChan:
		logger.Debug("Command timed out", "timeout", timeout, "partial_output", output.String())
		return output.String(), fmt.Errorf("command timed out after %v", timeout)
	}
}

// cleanOutput removes only completion markers from command output, preserving all bash session output.
func (b *BashService) cleanOutput(output string) string {
	logger.Debug("Cleaning output", "raw_output", output)

	lines := strings.Split(output, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Only skip completion markers - keep everything else including bash prompts
		if strings.Contains(line, "NEURO_CMD_DONE_") || strings.Contains(line, "NEURO_INIT_") {
			logger.Debug("Skipping completion marker", "line", line)
			continue
		}

		// Keep all other output including bash prompts, shell messages, etc.
		cleanLines = append(cleanLines, line)
	}

	// Remove trailing empty lines only
	for len(cleanLines) > 0 && strings.TrimSpace(cleanLines[len(cleanLines)-1]) == "" {
		cleanLines = cleanLines[:len(cleanLines)-1]
	}

	result := strings.Join(cleanLines, "\n")
	logger.Debug("Output cleaned", "cleaned_output", result)
	return result
}
