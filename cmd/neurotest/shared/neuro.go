// Package shared provides common utilities for neurotest.
package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckNeuroCommand verifies that the neuro command is available
func CheckNeuroCommand(neuroCmd string) error {
	originalCmd := neuroCmd

	// If user explicitly provided a command other than "neuro", try it as-is first
	if originalCmd != "neuro" {
		// Handle absolute paths
		if filepath.IsAbs(originalCmd) {
			if _, err := os.Stat(originalCmd); err == nil {
				return nil
			}
			return fmt.Errorf("neuro command not found at specified path: %s", originalCmd)
		}

		// Handle relative paths (containing path separators)
		if filepath.Dir(originalCmd) != "." {
			absPath, err := filepath.Abs(originalCmd)
			if err == nil {
				if _, err := os.Stat(absPath); err == nil {
					return nil
				}
			}
			return fmt.Errorf("neuro command not found at relative path: %s", originalCmd)
		}

		// If it's just a command name, fall through to candidates search
	}

	// Try common locations for the neuro binary
	candidates := []string{
		"./bin/neuro", // Local build
		"bin/neuro",   // Local build (alternative path)
		"neuro",       // PATH lookup
	}

	// If user specified something other than "neuro", try it first
	if originalCmd != "neuro" {
		candidates = append([]string{originalCmd}, candidates...)
	}

	for _, candidate := range candidates {
		if candidate == "neuro" {
			// Check if it's in PATH
			if _, err := exec.LookPath(candidate); err == nil {
				return nil
			}
		} else {
			// Check if file exists
			if _, err := os.Stat(candidate); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("neuro command not found. Tried: %v", candidates)
}

// RunNeuroScript executes a neuro script and returns its output
func RunNeuroScript(scriptPath, neuroCmd string, _ int) (string, error) {
	if err := CheckNeuroCommand(neuroCmd); err != nil {
		return "", err
	}

	// Determine the actual command to use
	actualCmd := neuroCmd
	if neuroCmd == "neuro" {
		// Try to find the best available option
		if _, err := os.Stat("./bin/neuro"); err == nil {
			actualCmd = "./bin/neuro"
		} else if _, err := os.Stat("bin/neuro"); err == nil {
			actualCmd = "bin/neuro"
		}
	}

	// Always use --log-level error to suppress INFO messages in test output
	cmd := exec.Command(actualCmd, "--test-mode", "--log-level", "error", "batch", scriptPath)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunNeuroCFlag executes a neuro script using the -c flag and returns its output
func RunNeuroCFlag(scriptPath, neuroCmd string, _ int) (string, error) {
	if err := CheckNeuroCommand(neuroCmd); err != nil {
		return "", err
	}

	// Read the script content
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read script file: %w", err)
	}

	// Determine the actual command to use
	actualCmd := neuroCmd
	if neuroCmd == "neuro" {
		// Try to find the best available option
		if _, err := os.Stat("./bin/neuro"); err == nil {
			actualCmd = "./bin/neuro"
		} else if _, err := os.Stat("bin/neuro"); err == nil {
			actualCmd = "bin/neuro"
		}
	}

	// Execute with -c flag and test mode, passing the script content directly
	// The -c flag will handle creating a temporary file and using batch processing
	cmd := exec.Command(actualCmd, "--test-mode", "--log-level", "error", "-c", string(scriptContent))
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	return string(output), err
}
