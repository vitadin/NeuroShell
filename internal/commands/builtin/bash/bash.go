// Package bash provides PTY-based bash command execution for NeuroShell.
// It manages persistent bash sessions with full terminal support.
package bash

import (
	"fmt"
	"strings"
	"time"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// Command implements the \bash command for executing system commands in PTY sessions.
// It provides persistent bash sessions with full terminal support and session management.
type Command struct{}

// Name returns the command name "bash" for registration and lookup.
func (c *Command) Name() string {
	return "bash"
}

// ParseMode returns ParseModeWithOptions to support option parsing.
func (c *Command) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeWithOptions
}

// Description returns a brief description of what the bash command does.
func (c *Command) Description() string {
	return "Execute commands in persistent bash sessions with PTY support"
}

// Usage returns the syntax and usage examples for the bash command.
func (c *Command) Usage() string {
	return `\bash[session="name", new=true, timeout="30s", env="KEY=value", cwd="/path", interactive=true, capture=true] command

Options:
  session="name"    - Session name (default: "default")
  new=true         - Force create new session
  timeout="30s"    - Command timeout (e.g., "30s", "5m")
  env="KEY=value"  - Set environment variable
  cwd="/path"      - Set working directory
  interactive=true - Enter interactive mode
  capture=true     - Capture output to ${_output}

Examples:
  \bash ls -la                              # Execute in default session
  \bash[session="analysis"] cd /data        # Use named session
  \bash[session="test", new=true] pwd       # Force new session
  \bash[timeout="10s"] long-running-cmd     # With timeout
  \bash[env="DEBUG=1"] ./script.sh          # With environment`
}

// Execute runs the bash command with the provided options and input.
func (c *Command) Execute(options map[string]string, input string, ctx neurotypes.Context) error {
	// Get services from global registry (commands only interact with services)
	bashService, err := services.GlobalRegistry.GetService("bash")
	if err != nil {
		return fmt.Errorf("bash service not available: %w", err)
	}

	variableService, err := services.GlobalRegistry.GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Cast services to their concrete types
	bs, ok := bashService.(*services.BashService)
	if !ok {
		return fmt.Errorf("invalid bash service type")
	}

	vs, ok := variableService.(*services.VariableService)
	if !ok {
		return fmt.Errorf("invalid variable service type")
	}

	// Parse options
	bashOptions, err := c.parseOptions(options)
	if err != nil {
		return fmt.Errorf("failed to parse options: %w", err)
	}

	// Validate command input
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("no command provided. Usage: %s", c.Usage())
	}

	// Handle interactive mode
	if bashOptions.Interactive {
		return c.handleInteractiveMode(bashOptions.SessionName, bs, ctx)
	}

	// Execute command in session
	output, err := bs.ExecuteCommand(bashOptions.SessionName, input, bashOptions, ctx)
	if err != nil {
		// Set error status
		if setErr := vs.SetSystemVariable("_status", "1", ctx); setErr != nil {
			fmt.Printf("Warning: failed to set _status variable: %v\n", setErr)
		}
		if setErr := vs.SetSystemVariable("_output", fmt.Sprintf("Error: %v", err), ctx); setErr != nil {
			fmt.Printf("Warning: failed to set _output variable: %v\n", setErr)
		}
		return fmt.Errorf("command execution failed: %w", err)
	}

	// Set success status and output if capture is enabled
	if bashOptions.CaptureOutput {
		if err := vs.SetSystemVariable("_status", "0", ctx); err != nil {
			// Log error but don't fail the command
			fmt.Printf("Warning: failed to set _status variable: %v\n", err)
		}

		if err := vs.SetSystemVariable("_output", output, ctx); err != nil {
			// Log error but don't fail the command
			fmt.Printf("Warning: failed to set _output variable: %v\n", err)
		}
	}

	// Output is already printed in real-time by OSC detection
	// No need to print again here

	return nil
}

// parseOptions parses the command options into BashOptions struct.
func (c *Command) parseOptions(options map[string]string) (services.BashOptions, error) {
	bashOptions := services.BashOptions{
		SessionName:   "default", // Default session name
		ForceNew:      false,
		Timeout:       0, // No timeout by default
		Environment:   make(map[string]string),
		WorkingDir:    "",
		Interactive:   false,
		CaptureOutput: true, // Default to capturing output
	}

	for key, value := range options {
		switch key {
		case "session":
			if value == "" {
				return bashOptions, fmt.Errorf("session name cannot be empty")
			}
			bashOptions.SessionName = value

		case "new":
			if value == "true" || value == "1" {
				bashOptions.ForceNew = true
			}

		case "timeout":
			if value != "" {
				timeout, err := time.ParseDuration(value)
				if err != nil {
					return bashOptions, fmt.Errorf("invalid timeout format '%s': %w", value, err)
				}
				bashOptions.Timeout = timeout
			}

		case "env":
			if value != "" {
				parts := strings.SplitN(value, "=", 2)
				if len(parts) != 2 {
					return bashOptions, fmt.Errorf("invalid environment variable format '%s', expected KEY=value", value)
				}
				bashOptions.Environment[parts[0]] = parts[1]
			}

		case "cwd":
			bashOptions.WorkingDir = value

		case "interactive":
			if value == "true" || value == "1" {
				bashOptions.Interactive = true
			}

		case "capture":
			if value == "false" || value == "0" {
				bashOptions.CaptureOutput = false
			}

		default:
			return bashOptions, fmt.Errorf("unknown option: %s", key)
		}
	}

	return bashOptions, nil
}

// handleInteractiveMode enters interactive mode for a bash session.
func (c *Command) handleInteractiveMode(_ string, _ *services.BashService, _ neurotypes.Context) error {
	// For now, return an error indicating interactive mode is not yet implemented
	return fmt.Errorf("interactive mode not yet implemented - this feature is planned for a future release")
}

func init() {
	if err := commands.GlobalRegistry.Register(&Command{}); err != nil {
		panic(fmt.Sprintf("failed to register bash command: %v", err))
	}
}
