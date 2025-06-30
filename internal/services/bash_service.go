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
	"neuroshell/internal/logger"
	"neuroshell/internal/shellintegration"
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

	// Execute command in session with hybrid detection (OSC + marker fallback)
	output, err := b.executeInSessionWithOSC(session, command, options, neuroCtx)
	if err != nil {
		// If hybrid detection fails completely, try one last fallback
		logger.Debug("Hybrid detection failed, trying emergency fallback", "session", sessionName, "error", err)
		emergencyOutput, emergencyErr := b.executeInSession(session, command, options)
		if emergencyErr != nil {
			return "", fmt.Errorf("failed to execute command in session %s (hybrid: %v, emergency: %v)", sessionName, err, emergencyErr)
		}
		return emergencyOutput, nil
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

// getOrCreateSession gets an existing session or creates a new one with shell integration.
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

// createBashSession creates a new bash session with PTY and shell integration.
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

	// Initialize shell integration tracking (will be set up properly when connected to context)
	session.CommandTracker = nil // Will be initialized when used with context
	session.SessionState = nil   // Will be initialized when used with context

	// Initialize shell integration for the session
	err = b.setupShellIntegration(session)
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
		return nil, fmt.Errorf("failed to setup shell integration: %w", err)
	}

	return session, nil
}

// setupShellIntegration configures bash session with OSC 133 shell integration.
func (b *BashService) setupShellIntegration(session *context.BashSession) error {
	logger.Debug("Setting up shell integration", "session", session.Name)

	// Give bash a moment to start up
	time.Sleep(100 * time.Millisecond)

	// Generate and log the shell integration script for debugging
	integrationScript := shellintegration.ShellIntegrationScript()
	logger.Debug("Generated shell integration script", "session", session.Name, "script_length", len(integrationScript))
	
	// Log first few lines of script for debugging
	scriptLines := strings.Split(integrationScript, "\n")
	if len(scriptLines) > 0 {
		logger.Debug("Script preview", "session", session.Name, "first_line", scriptLines[0])
	}

	// Send the script
	_, err := session.PTY.WriteString(integrationScript + "\n")
	if err != nil {
		logger.Debug("Failed to write shell integration script", "session", session.Name, "error", err)
		return fmt.Errorf("failed to write shell integration script: %w", err)
	}

	// Send a test command to verify integration
	testCommand := "echo 'NEURO_INTEGRATION_TEST'\n"
	_, err = session.PTY.WriteString(testCommand)
	if err != nil {
		logger.Debug("Failed to write test command", "session", session.Name, "error", err)
		return fmt.Errorf("failed to write test command: %w", err)
	}

	// Wait for the integration to be set up by looking for the initial prompt start
	timeout := time.After(5 * time.Second)
	done := make(chan bool)
	errChan := make(chan error)

	parser := shellintegration.NewStreamParser()

	go func() {
		scanner := bufio.NewScanner(session.PTY)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			logger.Debug("Shell integration setup: read line", "line", strings.TrimSpace(line))

			result := parser.ParseOutput([]byte(line))

			// Look for any OSC sequence indicating integration is working
			if len(result.Sequences) > 0 {
				logger.Debug("Shell integration setup complete", "sequences", len(result.Sequences))
				done <- true
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		} else {
			errChan <- fmt.Errorf("PTY closed during shell integration setup")
		}
	}()

	select {
	case <-done:
		logger.Debug("Shell integration ready", "session", session.Name)
		return b.validateShellIntegration(session)
	case err := <-errChan:
		logger.Debug("Shell integration setup failed", "error", err)
		// Don't fail - session can still work without integration
		return nil
	case <-timeout:
		logger.Debug("Shell integration setup timed out - using validation", "session", session.Name)
		// Try validation anyway - might still work
		return b.validateShellIntegration(session)
	}
}

// validateShellIntegration validates that shell integration is working properly
func (b *BashService) validateShellIntegration(session *context.BashSession) error {
	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	errChan := make(chan error)
	
	parser := shellintegration.NewStreamParser()
	var allOutput strings.Builder
	oscSequenceFound := false
	testOutputFound := false

	go func() {
		scanner := bufio.NewScanner(session.PTY)
		lineCount := 0
		
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			allOutput.WriteString(line + "\n")
			
			logger.Debug("Integration validation", "session", session.Name, "line", lineCount, "content", line)
			
			// Check for test output
			if strings.Contains(line, "NEURO_INTEGRATION_TEST") {
				testOutputFound = true
				logger.Debug("Test output found", "session", session.Name)
			}
			
			// Parse for OSC sequences
			result := parser.ParseOutput([]byte(line + "\n"))
			
			if len(result.Sequences) > 0 {
				for _, seq := range result.Sequences {
					logger.Debug("OSC sequence detected", "session", session.Name, "type", seq.Type, "raw", seq.Raw)
				}
				oscSequenceFound = true
			}
			
			// Consider integration successful if we see OSC sequences and test output
			if oscSequenceFound && testOutputFound {
				logger.Debug("Shell integration validation successful", "session", session.Name)
				done <- true
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		} else {
			errChan <- fmt.Errorf("PTY closed during validation")
		}
	}()

	select {
	case <-done:
		logger.Debug("Shell integration validated successfully", "session", session.Name)
		return nil
		
	case err := <-errChan:
		logger.Debug("Shell integration validation failed", "session", session.Name, "error", err, "output", allOutput.String())
		// Don't fail completely - session can work with fallback detection
		return nil
		
	case <-timeout:
		output := allOutput.String()
		logger.Debug("Shell integration validation timed out", "session", session.Name, "osc_found", oscSequenceFound, "test_found", testOutputFound, "output", output)
		
		// Check if we at least got the test output (bash is working)
		if testOutputFound {
			logger.Debug("Bash is working but OSC integration may have failed", "session", session.Name)
		} else {
			logger.Debug("No test output - bash session may have issues", "session", session.Name)
		}
		
		// Continue anyway - we'll use fallback detection
		return nil
	}
}

// executeInSessionWithOSC executes a command using OSC 133 shell integration for completion detection.
func (b *BashService) executeInSessionWithOSC(session *context.BashSession, command string, options BashOptions, ctx *context.NeuroContext) (string, error) {
	session.LastUsed = time.Now()

	logger.Debug("Executing command with OSC detection", "session", session.Name, "command", command)

	// Get or create command tracker session
	tracker := ctx.GetCommandTracker()
	trackerSession, exists := tracker.GetSession(session.Name)
	if !exists {
		trackerSession = tracker.CreateSession(session.Name)
		logger.Debug("Created new tracker session", "session", session.Name)
	}

	// Write command to PTY
	commandWithNewline := command + "\n"
	_, err := session.PTY.WriteString(commandWithNewline)
	if err != nil {
		return "", fmt.Errorf("failed to write command to PTY: %w", err)
	}

	// Set timeout - shorter default for faster response when OSC fails
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second // Fast timeout for better user experience
	}

	return b.executeWithHybridDetection(session, trackerSession, tracker, command, timeout)
}

// executeWithHybridDetection tries OSC detection first, then falls back to marker detection
func (b *BashService) executeWithHybridDetection(session *context.BashSession, trackerSession *shellintegration.SessionState, tracker *shellintegration.CommandTracker, command string, timeout time.Duration) (string, error) {
	logger.Debug("Starting hybrid detection", "session", session.Name, "command", command, "timeout", timeout)
	
	// Try OSC detection first with optimized timeout for completion detection
	oscTimeout := 500 * time.Millisecond
	if timeout < oscTimeout {
		oscTimeout = timeout / 3
	}
	
	logger.Debug("Trying OSC detection first", "session", session.Name, "osc_timeout", oscTimeout)
	
	output, err := b.readWithOSCDetectionTimeout(session, trackerSession, tracker, oscTimeout)
	if err == nil {
		logger.Debug("OSC detection successful", "session", session.Name)
		return output, nil
	}
	
	// OSC detection failed, fallback to marker detection
	logger.Debug("OSC detection failed, falling back to marker detection", "session", session.Name, "error", err)
	
	// Calculate remaining timeout
	remainingTimeout := timeout - oscTimeout
	if remainingTimeout <= 0 {
		remainingTimeout = 500 * time.Millisecond // Fast fallback timeout
	}
	
	return b.executeWithMarkerDetectionNoCommandWrite(session, command, remainingTimeout)
}

// executeWithMarkerDetection uses the legacy marker-based detection as fallback
func (b *BashService) executeWithMarkerDetection(session *context.BashSession, command string, timeout time.Duration) (string, error) {
	logger.Debug("Using marker detection fallback", "session", session.Name)
	
	// Generate unique marker for command completion detection
	marker := fmt.Sprintf("NEURO_CMD_DONE_%d", time.Now().UnixNano())
	
	// Don't re-execute the command, just add a marker to detect when it's done
	markerCommand := fmt.Sprintf("echo '%s'\n", marker)
	
	_, err := session.PTY.WriteString(markerCommand)
	if err != nil {
		return "", fmt.Errorf("failed to write marker command to PTY: %w", err)
	}
	
	return b.readUntilMarker(session.PTY, marker, timeout)
}

// executeWithMarkerDetectionNoCommandWrite uses marker-based detection without re-executing the command
func (b *BashService) executeWithMarkerDetectionNoCommandWrite(session *context.BashSession, command string, timeout time.Duration) (string, error) {
	logger.Debug("Using marker detection fallback without re-executing command", "session", session.Name)
	
	// Generate unique marker for command completion detection
	marker := fmt.Sprintf("NEURO_CMD_DONE_%d", time.Now().UnixNano())
	
	// Only add a marker to detect when the already-executed command is done
	markerCommand := fmt.Sprintf("echo '%s'\n", marker)
	
	_, err := session.PTY.WriteString(markerCommand)
	if err != nil {
		return "", fmt.Errorf("failed to write marker command to PTY: %w", err)
	}
	
	return b.readUntilMarker(session.PTY, marker, timeout)
}

// readWithOSCDetectionTimeout is like readWithOSCDetection but with a specific timeout
func (b *BashService) readWithOSCDetectionTimeout(session *context.BashSession, trackerSession *shellintegration.SessionState, tracker *shellintegration.CommandTracker, timeout time.Duration) (string, error) {
	var output strings.Builder
	
	timeoutChan := time.After(timeout)
	done := make(chan string)
	errChan := make(chan error)
	
	logger.Debug("Starting OSC detection with timeout", "session", session.Name, "timeout", timeout)
	
	go func() {
		scanner := bufio.NewScanner(session.PTY)
		lineCount := 0
		
		for scanner.Scan() {
			line := scanner.Text()
			lineCount++
			
			logger.Debug("Read PTY line with OSC", "line_num", lineCount, "content", line)
			
			// Process the line through command tracker for OSC detection
			result, err := tracker.ProcessOutput(session.Name, []byte(line+"\n"))
			if err != nil {
				logger.Debug("Error processing output", "error", err)
				continue
			}
			
			// Debug OSC sequence detection
			if result.StateChanged {
				logger.Debug("OSC state changed", "session", session.Name, "old_state", "unknown", "new_state", result.State)
			}
			
			// Display output in real-time (honest output) - use NewOutput to avoid duplication
			if result.HasNewOutput && result.NewOutput != "" {
				cleanOutput := b.filterRealtimeOutput(result.NewOutput)
				if strings.TrimSpace(cleanOutput) != "" {
					fmt.Print(cleanOutput)
					output.WriteString(cleanOutput)
				}
			}
			
			// Check if command is complete
			if result.IsComplete {
				logger.Debug("Command completed via OSC detection", "session", session.Name, "exit_code", result.ExitCode, "total_lines", lineCount, "state", result.State)
				done <- output.String()
				return
			}
		}
		
		// Scanner error or PTY closed
		if err := scanner.Err(); err != nil {
			logger.Debug("Scanner error in OSC detection", "error", err)
			errChan <- err
		} else {
			logger.Debug("PTY closed during OSC detection")
			errChan <- fmt.Errorf("PTY closed unexpectedly")
		}
	}()
	
	select {
	case result := <-done:
		logger.Debug("OSC detection completed successfully", "output_length", len(result))
		return result, nil
	case err := <-errChan:
		logger.Debug("OSC detection failed with PTY error", "error", err, "partial_output", output.String())
		return output.String(), fmt.Errorf("OSC detection failed: %w", err)
	case <-timeoutChan:
		logger.Debug("OSC detection timed out", "timeout", timeout, "partial_output", output.String())
		return output.String(), fmt.Errorf("OSC detection timed out after %v", timeout)
	}
}

// readWithOSCDetection reads PTY output with real-time display and OSC completion detection.
func (b *BashService) readWithOSCDetection(session *context.BashSession, trackerSession *shellintegration.SessionState, tracker *shellintegration.CommandTracker, timeout time.Duration) (string, error) {
	var output strings.Builder

	timeoutChan := time.After(timeout)
	done := make(chan string)
	errChan := make(chan error)

	logger.Debug("Starting OSC-based output reading", "session", session.Name, "timeout", timeout)

	go func() {
		scanner := bufio.NewScanner(session.PTY)
		lineCount := 0

		for scanner.Scan() {
			line := scanner.Text()
			lineCount++

			logger.Debug("Read PTY line with OSC", "line_num", lineCount, "content", line)

			// Process the line through command tracker for OSC detection
			result, err := tracker.ProcessOutput(session.Name, []byte(line+"\n"))
			if err != nil {
				logger.Debug("Error processing output", "error", err)
				continue
			}

			// Display output in real-time (honest output) - use NewOutput to avoid duplication
			if result.HasNewOutput && result.NewOutput != "" {
				cleanOutput := b.filterRealtimeOutput(result.NewOutput)
				if strings.TrimSpace(cleanOutput) != "" {
					fmt.Print(cleanOutput)
					output.WriteString(cleanOutput)
				}
			}

			// Check if command is complete
			if result.IsComplete {
				logger.Debug("Command completed via OSC detection", "session", session.Name, "exit_code", result.ExitCode, "total_lines", lineCount)
				done <- output.String()
				return
			}
		}

		// Scanner error or PTY closed
		if err := scanner.Err(); err != nil {
			logger.Debug("Scanner error in OSC detection", "error", err)
			errChan <- err
		} else {
			logger.Debug("PTY closed during OSC detection")
			errChan <- fmt.Errorf("PTY closed unexpectedly")
		}
	}()

	select {
	case result := <-done:
		logger.Debug("Command completed successfully with OSC", "output_length", len(result))
		return result, nil
	case err := <-errChan:
		logger.Debug("PTY read error with OSC", "error", err, "partial_output", output.String())
		return output.String(), fmt.Errorf("PTY read error: %w", err)
	case <-timeoutChan:
		logger.Debug("Command timed out with OSC", "timeout", timeout, "partial_output", output.String())
		return output.String(), fmt.Errorf("command timed out after %v", timeout)
	}
}

// executeInSession executes a command in an existing bash session.
func (b *BashService) executeInSession(session *context.BashSession, command string, options BashOptions) (string, error) {
	session.LastUsed = time.Now()

	// Generate unique marker for command completion detection (legacy fallback)
	marker := fmt.Sprintf("NEURO_CMD_DONE_%d", time.Now().UnixNano())

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
		timeout = 1 * time.Second // Default 1-second timeout for faster response
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

// filterRealtimeOutput filters both OSC sequences and markers from real-time output
func (b *BashService) filterRealtimeOutput(output string) string {
	// First filter OSC sequences
	cleanOutput := shellintegration.FilterOSCSequences(output)
	
	// Then filter marker lines
	lines := strings.Split(cleanOutput, "\n")
	var filteredLines []string
	
	for _, line := range lines {
		// Skip completion markers and bash command echoes containing markers
		if strings.Contains(line, "NEURO_CMD_DONE_") || 
		   strings.Contains(line, "NEURO_INIT_") ||
		   strings.Contains(line, "echo 'NEURO_CMD_DONE_") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	
	return strings.Join(filteredLines, "\n")
}
