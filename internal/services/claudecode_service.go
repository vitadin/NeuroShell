package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/logger"

	"github.com/charmbracelet/log"
)

// ClaudeCodeService provides Claude Code CLI integration for NeuroShell.
// It manages Claude Code process lifecycle, session management, and job execution.
type ClaudeCodeService struct {
	// Process management
	process       *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	processCtx    context.Context
	cancelProcess context.CancelFunc

	// State management
	sessions      map[string]*ClaudeCodeSession
	jobs          map[string]*ClaudeCodeJob
	activeSession string

	// Control
	initialized bool
	mu          sync.RWMutex
	logger      *log.Logger

	// Communication
	responseChans map[string]chan *ClaudeCodeResponse
	eventBus      chan *ClaudeCodeEvent
	jobCounter    int64

	// Function fields for testing (can be overridden)
	isClaudeCodeInstalled func() bool
	getClaudeCodeVersion  func() (string, error)
	StartProcess          func(InitOptions) error

	// Cleanup state
	eventBusClosed bool
}

// ClaudeCodeSession represents a Claude Code session state
type ClaudeCodeSession struct {
	ID               string    `json:"id"`
	ClaudeSessionID  string    `json:"claude_session_id"`
	NeuroSessionID   string    `json:"neuro_session_id,omitempty"`
	Status           string    `json:"status"` // idle|busy|error
	CurrentJobID     string    `json:"current_job_id,omitempty"`
	LastResponse     string    `json:"last_response,omitempty"`
	WorkingDirectory string    `json:"working_directory"`
	Model            string    `json:"model"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ClaudeCodeJob represents an async operation
type ClaudeCodeJob struct {
	ID         string              `json:"id"`
	SessionID  string              `json:"session_id"`
	Command    string              `json:"command"`
	Input      string              `json:"input"`
	Status     ClaudeCodeJobStatus `json:"status"`
	Output     strings.Builder     `json:"-"`
	StreamChan chan string         `json:"-"`
	Error      error               `json:"-"`
	StartedAt  time.Time           `json:"started_at"`
	EndedAt    *time.Time          `json:"ended_at,omitempty"`
	Progress   float64             `json:"progress"`
}

// ClaudeCodeJobStatus represents job execution status
type ClaudeCodeJobStatus string

// Job status constants for tracking Claude Code job states
const (
	JobStatusPending   ClaudeCodeJobStatus = "pending"
	JobStatusRunning   ClaudeCodeJobStatus = "running"
	JobStatusCompleted ClaudeCodeJobStatus = "completed"
	JobStatusFailed    ClaudeCodeJobStatus = "failed"
	JobStatusCancelled ClaudeCodeJobStatus = "cancelled"
	JobStatusTimeout   ClaudeCodeJobStatus = "timeout"
)

// ClaudeCodeResponse represents a response from Claude Code
type ClaudeCodeResponse struct {
	JobID         string                 `json:"job_id"`
	Type          string                 `json:"type"` // message|tool_use|error|complete
	Content       string                 `json:"content"`
	ToolsUsed     []string               `json:"tools_used,omitempty"`
	FilesModified []string               `json:"files_modified,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ClaudeCodeEvent represents system events
type ClaudeCodeEvent struct {
	Type      string    `json:"type"` // session_start|session_end|job_start|job_end|error
	SessionID string    `json:"session_id"`
	JobID     string    `json:"job_id,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ClaudeCodeAuth represents authentication state
type ClaudeCodeAuth struct {
	Method    string     `json:"method"`     // env|variable|keychain|session
	KeySource string     `json:"key_source"` // Where the key comes from
	IsValid   bool       `json:"is_valid"`   // Last validation result
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// InitOptions represents initialization options for Claude Code
type InitOptions struct {
	Model          string   `json:"model"`
	Verbose        bool     `json:"verbose"`
	Directories    []string `json:"directories"`
	PermissionMode string   `json:"permission_mode"`
}

// JobOptions represents options for job execution
type JobOptions struct {
	Timeout   time.Duration `json:"timeout"`
	Stream    bool          `json:"stream"`
	SessionID string        `json:"session_id"`
}

// NewClaudeCodeService creates a new ClaudeCodeService instance.
func NewClaudeCodeService() *ClaudeCodeService {
	service := &ClaudeCodeService{
		sessions:      make(map[string]*ClaudeCodeSession),
		jobs:          make(map[string]*ClaudeCodeJob),
		responseChans: make(map[string]chan *ClaudeCodeResponse),
		eventBus:      make(chan *ClaudeCodeEvent, 100),
		initialized:   false,
		logger:        logger.NewStyledLogger("ClaudeCodeService"),
		jobCounter:    0,
	}

	// Initialize function fields with default implementations
	service.isClaudeCodeInstalled = service.defaultIsClaudeCodeInstalled
	service.getClaudeCodeVersion = service.defaultGetClaudeCodeVersion
	service.StartProcess = service.defaultStartProcess

	return service
}

// Name returns the service name "claudecode" for registration.
func (c *ClaudeCodeService) Name() string {
	return "claudecode"
}

// Initialize sets up the ClaudeCodeService for operation.
func (c *ClaudeCodeService) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Check if Claude Code is installed
	if !c.isClaudeCodeInstalled() {
		return fmt.Errorf("claude code CLI not found - please install it first: npm install -g @anthropic/claude-code")
	}

	// Verify Claude Code version compatibility
	version, err := c.getClaudeCodeVersion()
	if err != nil {
		return fmt.Errorf("failed to get claude code version: %w", err)
	}

	c.logger.Debug("Found Claude Code version", "version", version)

	c.initialized = true
	return nil
}

// defaultIsClaudeCodeInstalled checks if Claude Code CLI is available
func (c *ClaudeCodeService) defaultIsClaudeCodeInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// getClaudeCodeVersion retrieves the Claude Code CLI version
func (c *ClaudeCodeService) defaultGetClaudeCodeVersion() (string, error) {
	cmd := exec.Command("claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// StartProcess starts the Claude Code daemon process
func (c *ClaudeCodeService) defaultStartProcess(opts InitOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.process != nil {
		return fmt.Errorf("claude code process already running")
	}

	// Create process context for cleanup
	c.processCtx, c.cancelProcess = context.WithCancel(context.Background())

	// Build command arguments
	args := []string{}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts.Verbose {
		args = append(args, "--verbose")
	}
	if opts.PermissionMode != "" {
		args = append(args, "--permission-mode", opts.PermissionMode)
	}
	for _, dir := range opts.Directories {
		args = append(args, "--add-dir", dir)
	}

	// Add output format for programmatic usage
	args = append(args, "--output-format", "json")

	// Start Claude Code process
	c.process = exec.CommandContext(c.processCtx, "claude", args...)

	// Set up pipes
	stdin, err := c.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdout

	stderr, err := c.process.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderr

	// Start the process
	if err := c.process.Start(); err != nil {
		return fmt.Errorf("failed to start claude code process: %w", err)
	}

	// Start output processing goroutines
	go c.processOutput()
	go c.processErrors()

	c.logger.Info("Claude Code process started", "pid", c.process.Process.Pid)
	return nil
}

// StopProcess stops the Claude Code daemon process
func (c *ClaudeCodeService) StopProcess() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.process == nil {
		return nil
	}

	// Cancel the process context
	if c.cancelProcess != nil {
		c.cancelProcess()
	}

	// Close stdin to signal shutdown
	if c.stdin != nil {
		_ = c.stdin.Close()
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- c.process.Wait()
	}()

	select {
	case err := <-done:
		c.logger.Info("Claude Code process stopped", "error", err)
	case <-time.After(5 * time.Second):
		// Force kill if it doesn't stop gracefully
		if err := c.process.Process.Kill(); err != nil {
			c.logger.Error("Failed to kill Claude Code process", "error", err)
		}
		c.logger.Warn("Claude Code process force killed after timeout")
	}

	// Clean up
	c.process = nil
	c.stdin = nil
	c.stdout = nil
	c.stderr = nil

	return nil
}

// processOutput handles stdout from Claude Code process
func (c *ClaudeCodeService) processOutput() {
	if c.stdout == nil {
		return
	}

	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		c.logger.Debug("Claude Code output", "line", line)

		// Try to parse as JSON response
		var response ClaudeCodeResponse
		if err := json.Unmarshal([]byte(line), &response); err == nil {
			c.handleResponse(&response)
		} else {
			// Handle non-JSON output (might be streaming text)
			c.handleStreamingOutput(line)
		}
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("Error reading Claude Code stdout", "error", err)
	}
}

// processErrors handles stderr from Claude Code process
func (c *ClaudeCodeService) processErrors() {
	if c.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		c.logger.Error("Claude Code stderr", "line", line)
	}

	if err := scanner.Err(); err != nil {
		c.logger.Error("Error reading Claude Code stderr", "error", err)
	}
}

// handleResponse processes structured responses from Claude Code
func (c *ClaudeCodeService) handleResponse(response *ClaudeCodeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update job status if job ID is provided
	if response.JobID != "" {
		if job, exists := c.jobs[response.JobID]; exists {
			job.Output.WriteString(response.Content)

			// Send to stream channel if open
			if job.StreamChan != nil {
				select {
				case job.StreamChan <- response.Content:
				default:
					// Channel might be full or closed
				}
			}

			// Update job status based on response type
			switch response.Type {
			case "complete":
				job.Status = JobStatusCompleted
				now := time.Now()
				job.EndedAt = &now
				job.Progress = 100.0
			case "error":
				job.Status = JobStatusFailed
				now := time.Now()
				job.EndedAt = &now
			}

			// Store tools used and files modified
			if len(response.ToolsUsed) > 0 {
				c.setJobVariable(response.JobID, "_cc_tools_used", strings.Join(response.ToolsUsed, ","))
			}
			if len(response.FilesModified) > 0 {
				c.setJobVariable(response.JobID, "_cc_files_modified", strings.Join(response.FilesModified, ","))
			}
		}
	}

	// Send to response channel if someone is waiting
	if ch, exists := c.responseChans[response.JobID]; exists {
		select {
		case ch <- response:
		default:
			// Channel might be full or closed
		}
	}
}

// handleStreamingOutput processes streaming text output
func (c *ClaudeCodeService) handleStreamingOutput(line string) {
	// This would handle streaming output that isn't JSON
	// For now, just log it
	c.logger.Debug("Streaming output", "line", line)
}

// setJobVariable sets a variable in the global context related to a job
func (c *ClaudeCodeService) setJobVariable(_, varName, value string) {
	ctx := neuroshellcontext.GetGlobalContext()
	if ctx == nil {
		return
	}

	if neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext); ok {
		// Set as system variable since it starts with _
		_ = neuroCtx.SetSystemVariable(varName, value)
	}
}

// CreateSession creates a new Claude Code session
func (c *ClaudeCodeService) CreateSession(opts InitOptions) (*ClaudeCodeSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("claude code service not initialized")
	}

	// Ensure process is running
	if c.process == nil {
		if err := c.StartProcess(opts); err != nil {
			return nil, err
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate session ID
	sessionID := fmt.Sprintf("cc-session-%d", time.Now().UnixNano())

	// Create session
	session := &ClaudeCodeSession{
		ID:               sessionID,
		ClaudeSessionID:  "", // Will be set when Claude Code responds
		Status:           "idle",
		WorkingDirectory: ".", // Default to current directory
		Model:            opts.Model,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	c.sessions[sessionID] = session

	// Set as active session if it's the first one
	if c.activeSession == "" {
		c.activeSession = sessionID
	}

	c.logger.Info("Created Claude Code session", "session_id", sessionID)
	return session, nil
}

// GetSession retrieves a session by ID
func (c *ClaudeCodeService) GetSession(sessionID string) (*ClaudeCodeSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// GetActiveSession returns the currently active session
func (c *ClaudeCodeService) GetActiveSession() (*ClaudeCodeSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.activeSession == "" {
		return nil, fmt.Errorf("no active session")
	}

	return c.GetSession(c.activeSession)
}

// SetActiveSession sets the active session
func (c *ClaudeCodeService) SetActiveSession(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	c.activeSession = sessionID
	return nil
}

// ListSessions returns all sessions
func (c *ClaudeCodeService) ListSessions() []*ClaudeCodeSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]*ClaudeCodeSession, 0, len(c.sessions))
	for _, session := range c.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// SubmitJob submits a job for execution
func (c *ClaudeCodeService) SubmitJob(sessionID, command, input string, opts JobOptions) (*ClaudeCodeJob, error) {
	if !c.initialized {
		return nil, fmt.Errorf("claude code service not initialized")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify session exists
	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Generate job ID
	c.jobCounter++
	jobID := fmt.Sprintf("cc-job-%d", c.jobCounter)

	// Create job
	job := &ClaudeCodeJob{
		ID:        jobID,
		SessionID: sessionID,
		Command:   command,
		Input:     input,
		Status:    JobStatusPending,
		StartedAt: time.Now(),
		Progress:  0.0,
	}

	// Set up stream channel if requested
	if opts.Stream {
		job.StreamChan = make(chan string, 100)
	}

	c.jobs[jobID] = job

	// For testing: Add some mock output to demonstrate the service is working
	job.Output.WriteString(fmt.Sprintf("Mock output for command: %s\nInput: %s\n", command, input))

	// Update session status
	session.Status = "busy"
	session.CurrentJobID = jobID
	session.UpdatedAt = time.Now()

	// Submit to Claude Code (this is a simplified implementation)
	// In reality, we'd need to format the request properly for Claude Code
	go c.executeJob(job)

	c.logger.Info("Submitted job", "job_id", jobID, "session_id", sessionID)
	return job, nil
}

// executeJob executes a job (simplified implementation)
func (c *ClaudeCodeService) executeJob(job *ClaudeCodeJob) {
	// Update job status
	c.mu.Lock()
	job.Status = JobStatusRunning
	c.mu.Unlock()

	// This is a simplified implementation - in reality we'd send the command to Claude Code
	// For now, simulate some work
	time.Sleep(100 * time.Millisecond)

	// Simulate response
	response := &ClaudeCodeResponse{
		JobID:   job.ID,
		Type:    "complete",
		Content: fmt.Sprintf("Response to: %s", job.Input),
	}

	c.handleResponse(response)

	// Update session status
	c.mu.Lock()
	if session, exists := c.sessions[job.SessionID]; exists {
		session.Status = "idle"
		session.CurrentJobID = ""
		session.LastResponse = response.Content
		session.UpdatedAt = time.Now()
	}
	c.mu.Unlock()
}

// GetJob retrieves a job by ID
func (c *ClaudeCodeService) GetJob(jobID string) (*ClaudeCodeJob, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	job, exists := c.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return job, nil
}

// WaitForJob waits for a job to complete
func (c *ClaudeCodeService) WaitForJob(jobID string, timeout time.Duration) error {
	job, err := c.GetJob(jobID)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("job %s timed out after %v", jobID, timeout)
}

// GetJobOutput returns the current output of a job
func (c *ClaudeCodeService) GetJobOutput(jobID string) (string, error) {
	job, err := c.GetJob(jobID)
	if err != nil {
		return "", err
	}

	return job.Output.String(), nil
}

// CheckAuth checks the current authentication status
func (c *ClaudeCodeService) CheckAuth() (*ClaudeCodeAuth, error) {
	auth := &ClaudeCodeAuth{
		Method:  "env",
		IsValid: false,
	}

	// Check for API key in environment
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		auth.KeySource = "ANTHROPIC_API_KEY"
		auth.IsValid = true // Simplified - would normally validate the key
		return auth, nil
	}

	// Check in NeuroShell variables
	ctx := neuroshellcontext.GetGlobalContext()
	if ctx != nil {
		if apiKey, err := ctx.GetVariable("os.ANTHROPIC_API_KEY"); err == nil && apiKey != "" {
			auth.Method = "variable"
			auth.KeySource = "os.ANTHROPIC_API_KEY"
			auth.IsValid = true
			return auth, nil
		}
	}

	return auth, fmt.Errorf("no valid authentication found")
}

// Health returns service health information
func (c *ClaudeCodeService) Health() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := map[string]interface{}{
		"initialized":       c.initialized,
		"process_running":   c.process != nil,
		"active_sessions":   len(c.sessions),
		"active_jobs":       len(c.jobs),
		"active_session_id": c.activeSession,
	}

	if c.process != nil {
		health["process_pid"] = c.process.Process.Pid
	}

	// Add version info if available
	if version, err := c.getClaudeCodeVersion(); err == nil {
		health["claude_version"] = version
	}

	return health
}

// Cleanup shuts down the service gracefully
func (c *ClaudeCodeService) Cleanup() error {
	c.logger.Info("Shutting down Claude Code service")

	// Stop the process if running
	if err := c.StopProcess(); err != nil {
		c.logger.Error("Error stopping Claude Code process", "error", err)
	}

	// Close event bus and response channels
	c.mu.Lock()
	if !c.eventBusClosed {
		close(c.eventBus)
		c.eventBusClosed = true
	}
	for _, ch := range c.responseChans {
		close(ch)
	}
	c.responseChans = make(map[string]chan *ClaudeCodeResponse)
	c.mu.Unlock()

	return nil
}
