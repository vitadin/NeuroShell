package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	neuroshellcontext "neuroshell/internal/context"
)

// BashService provides system command execution operations for NeuroShell contexts.
type BashService struct {
	initialized bool
	timeout     time.Duration
}

// NewBashService creates a new BashService instance with a default timeout of 30 seconds.
func NewBashService() *BashService {
	return &BashService{
		initialized: false,
		timeout:     30 * time.Second,
	}
}

// Name returns the service name "bash" for registration.
func (b *BashService) Name() string {
	return "bash"
}

// Initialize sets up the BashService for operation.
func (b *BashService) Initialize() error {
	b.initialized = true
	return nil
}

// SetTimeout configures the execution timeout for bash commands.
func (b *BashService) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// Execute runs a bash command and returns the output, error, and exit status.
// It also sets the _output variable and @error, @status system variables in the global context.
func (b *BashService) Execute(command string) (string, string, int, error) {
	if !b.initialized {
		return "", "", -1, fmt.Errorf("bash service not initialized")
	}

	if strings.TrimSpace(command) == "" {
		return "", "", -1, fmt.Errorf("empty command provided")
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	// Execute command using bash -c
	cmd := exec.CommandContext(ctxWithTimeout, "bash", "-c", command)

	// Set up separate pipes for stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	err := cmd.Run()

	// Get the output
	stdout := stdoutBuf.Bytes()
	stderr := stderrBuf.Bytes()
	var exitCode int

	if err != nil {
		// Handle different types of errors
		if exitError, ok := err.(*exec.ExitError); ok {
			// Command ran but returned non-zero exit code
			exitCode = exitError.ExitCode()
		} else if ctxWithTimeout.Err() == context.DeadlineExceeded || strings.Contains(err.Error(), "deadline exceeded") {
			// Command timed out
			stderr = []byte(fmt.Sprintf("command timed out after %v", b.timeout))
			exitCode = -1
		} else {
			// Other execution errors (command not found, etc.)
			// For these errors, add to stderr if it's empty
			if len(stderr) == 0 {
				stderr = []byte(err.Error())
			}
			exitCode = -1
		}
	} else {
		exitCode = 0
	}

	stdoutStr := strings.TrimRight(string(stdout), "\n")
	stderrStr := strings.TrimRight(string(stderr), "\n")

	// Set only _output variable - error state is now managed by the framework
	globalCtx := neuroshellcontext.GetGlobalContext()
	if neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext); ok {
		_ = neuroCtx.SetSystemVariable("_output", stdoutStr)
		// Note: @error and @status are now set by the stack machine framework
		// based on command execution results, not by individual services
	}

	return stdoutStr, stderrStr, exitCode, nil
}
