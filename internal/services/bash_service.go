package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"neuroshell/pkg/neurotypes"
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
func (b *BashService) Initialize(_ neurotypes.Context) error {
	b.initialized = true
	return nil
}

// SetTimeout configures the execution timeout for bash commands.
func (b *BashService) SetTimeout(timeout time.Duration) {
	b.timeout = timeout
}

// Execute runs a bash command and returns the output, error, and exit status.
// It also sets the _output, _error, and _status variables using the variable service.
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

	// Set system variables for command results
	if variableService, err := getVariableService(); err == nil {
		_ = variableService.SetSystemVariable("_output", stdoutStr)
		_ = variableService.SetSystemVariable("_error", stderrStr)
		_ = variableService.SetSystemVariable("_status", fmt.Sprintf("%d", exitCode))
	}

	return stdoutStr, stderrStr, exitCode, nil
}

// getVariableService is a helper function to get the variable service from the global registry
func getVariableService() (*VariableService, error) {
	service, err := GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}
