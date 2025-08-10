package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// diffNeuroRCTest shows differences in .neurorc startup test output
func diffNeuroRCTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Showing diff for .neurorc startup test: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the test configuration file
	configFile, err := findNeuroRCTestConfig(testName)
	if err != nil {
		return err
	}

	// Parse test configuration
	config, err := parseNeuroRCTestConfig(configFile)
	if err != nil {
		return err
	}

	// Check if expected file exists
	expectedFile := filepath.Join(testDir, "neurorc", testName+".expected")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s", expectedFile)
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output: %w", err)
	}
	expected := string(expectedBytes)

	// Create isolated test environment
	testEnv, err := createNeuroRCTestEnvironment(config)
	if err != nil {
		return err
	}
	defer cleanupNeuroRCTestEnvironment(testEnv)

	// Run the .neurorc startup test
	actual, err := runNeuroRCShellTest(testEnv, config)
	if err != nil {
		return fmt.Errorf("failed to run .neurorc startup test: %w", err)
	}

	// Show enhanced diff
	showDetailedDiff(expected, actual, testName)

	return nil
}

// findNeuroRCTestConfig finds the test configuration file for a .neurorc test
func findNeuroRCTestConfig(testName string) (string, error) {
	// Try different possible locations
	candidates := []string{
		filepath.Join(testDir, "neurorc", testName+".neurorc-test"),
		filepath.Join("test/neurorc", testName+".neurorc-test"),
		testName + ".neurorc-test",
		testName,
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf(".neurorc test config not found for test: %s\nTried: %v", testName, candidates)
}

// parseNeuroRCTestConfig parses a .neurorc test configuration file
func parseNeuroRCTestConfig(configFile string) (*NeuroRCTestConfig, error) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &NeuroRCTestConfig{
		Name:        filepath.Base(strings.TrimSuffix(configFile, ".neurorc-test")),
		Description: "Basic .neurorc startup test",
		Setup:       NeuroRCTestSetup{},
		CLIFlags:    []string{},
		EnvVars:     make(map[string]string),
		InputSeq:    []string{"\\exit"},
		ExpectedOut: []string{},
		ExpectedNot: []string{},
		Timeout:     30,
	}

	// Simple line-by-line parser for the test config format
	lines := strings.Split(string(content), "\n")
	var currentSection string
	var currentContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section markers
		if strings.HasPrefix(line, "---") {
			if currentSection != "" {
				err := applyConfigSection(config, currentSection, currentContent.String())
				if err != nil {
					return nil, fmt.Errorf("failed to parse section %s: %w", currentSection, err)
				}
				currentContent.Reset()
			}
			currentSection = strings.TrimSpace(strings.TrimPrefix(line, "---"))
			continue
		}

		// Content lines
		if currentSection != "" {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// Process final section
	if currentSection != "" {
		err := applyConfigSection(config, currentSection, currentContent.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse section %s: %w", currentSection, err)
		}
	}

	return config, nil
}

// applyConfigSection applies a parsed configuration section to the config
func applyConfigSection(config *NeuroRCTestConfig, section, content string) error {
	switch section {
	case "description":
		config.Description = content
	case "working_dir_neurorc":
		config.Setup.WorkingDirNeuroRC = content
	case "home_dir_neurorc":
		config.Setup.HomeDirNeuroRC = content
	case "cli_flags":
		config.CLIFlags = strings.Fields(content)
	case "env_vars":
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				config.EnvVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	case "input_sequence":
		config.InputSeq = strings.Split(content, "\n")
	case "expected_contains":
		for _, line := range strings.Split(content, "\n") {
			if strings.TrimSpace(line) != "" {
				config.ExpectedOut = append(config.ExpectedOut, strings.TrimSpace(line))
			}
		}
	case "expected_not_contains":
		for _, line := range strings.Split(content, "\n") {
			if strings.TrimSpace(line) != "" {
				config.ExpectedNot = append(config.ExpectedNot, strings.TrimSpace(line))
			}
		}
	case "custom_files":
		if config.Setup.CustomFiles == nil {
			config.Setup.CustomFiles = make(map[string]string)
		}
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				filename := strings.TrimSpace(parts[0])
				fileContent := strings.TrimSpace(parts[1])
				// Replace literal \n with actual newlines in file content
				fileContent = strings.ReplaceAll(fileContent, "\\n", "\n")
				config.Setup.CustomFiles[filename] = fileContent
			}
		}
	case "timeout":
		// Parse timeout if needed
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
	return nil
}

// createNeuroRCTestEnvironment creates an isolated test environment
func createNeuroRCTestEnvironment(config *NeuroRCTestConfig) (*NeuroRCTestEnvironment, error) {
	// Create main temporary directory
	tempDir, err := os.MkdirTemp("", "neurotest-neurorc-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create working directory and home directory
	workingDir := filepath.Join(tempDir, "working")
	homeDir := filepath.Join(tempDir, "home")

	if err := os.MkdirAll(workingDir, 0755); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create home directory: %w", err)
	}

	testEnv := &NeuroRCTestEnvironment{
		TempDir:     tempDir,
		WorkingDir:  workingDir,
		HomeDir:     homeDir,
		ConfigFiles: make(map[string]string),
	}

	// Create .neurorc files as specified in config
	if config.Setup.WorkingDirNeuroRC != "" {
		neuroRCPath := filepath.Join(workingDir, ".neurorc")
		if err := os.WriteFile(neuroRCPath, []byte(config.Setup.WorkingDirNeuroRC), 0644); err != nil {
			cleanupNeuroRCTestEnvironment(testEnv)
			return nil, fmt.Errorf("failed to create working dir .neurorc: %w", err)
		}
		testEnv.ConfigFiles["working/.neurorc"] = neuroRCPath
	}

	if config.Setup.HomeDirNeuroRC != "" {
		neuroRCPath := filepath.Join(homeDir, ".neurorc")
		if err := os.WriteFile(neuroRCPath, []byte(config.Setup.HomeDirNeuroRC), 0644); err != nil {
			cleanupNeuroRCTestEnvironment(testEnv)
			return nil, fmt.Errorf("failed to create home dir .neurorc: %w", err)
		}
		testEnv.ConfigFiles["home/.neurorc"] = neuroRCPath
	}

	// Create custom files if specified
	for filename, content := range config.Setup.CustomFiles {
		customFilePath := filepath.Join(workingDir, filename)
		if err := os.MkdirAll(filepath.Dir(customFilePath), 0755); err != nil {
			cleanupNeuroRCTestEnvironment(testEnv)
			return nil, fmt.Errorf("failed to create custom file directory: %w", err)
		}
		if err := os.WriteFile(customFilePath, []byte(content), 0644); err != nil {
			cleanupNeuroRCTestEnvironment(testEnv)
			return nil, fmt.Errorf("failed to create custom file %s: %w", filename, err)
		}
		testEnv.ConfigFiles[filename] = customFilePath
	}

	return testEnv, nil
}

// cleanupNeuroRCTestEnvironment removes the isolated test environment
func cleanupNeuroRCTestEnvironment(testEnv *NeuroRCTestEnvironment) {
	if testEnv != nil && testEnv.TempDir != "" {
		_ = os.RemoveAll(testEnv.TempDir)
	}
}

// runNeuroRCShellTest runs the shell test in the isolated environment
func runNeuroRCShellTest(testEnv *NeuroRCTestEnvironment, config *NeuroRCTestConfig) (string, error) {
	// Prepare the command arguments
	args := []string{"shell", "--test-mode", "--no-color"}
	args = append(args, config.CLIFlags...)

	// Prepare input sequence
	inputSeq := strings.Join(config.InputSeq, "\n") + "\n"

	// Create the command
	cmd := exec.Command(neurocmd, args...)

	// Set working directory
	cmd.Dir = testEnv.WorkingDir

	// Set up environment variables
	cmd.Env = []string{
		"HOME=" + testEnv.HomeDir,
		"NEURO_LOG_LEVEL=fatal", // Minimize log noise
		"NO_COLOR=1",            // Disable colors for consistent output
		"PATH=" + os.Getenv("PATH"),
		"TERM=xterm", // Set consistent terminal
	}

	// Add custom environment variables from config
	for key, value := range config.EnvVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	// Set up stdin for scripted input
	cmd.Stdin = strings.NewReader(inputSeq)

	// Set timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Run command with timeout
	done := make(chan error, 1)
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		done <- cmdErr
	}()

	select {
	case err := <-done:
		if err != nil {
			// Don't treat exit errors as fatal - the shell might exit normally
			if _, ok := err.(*exec.ExitError); !ok {
				return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
			}
		}
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return "", fmt.Errorf("command timed out after %v", timeout)
	}

	// Clean up the output
	cleanOutput := cleanNeuroRCOutput(string(output))
	return cleanOutput, nil
}

// cleanNeuroRCOutput removes shell startup messages and other noise specific to .neurorc tests
func cleanNeuroRCOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string

	skipPatterns := []string{
		"INFO ",
		"DEBUG ",
		"WARN ",
		"ERROR ",
		"Neuro Shell v",
		"Licensed under LGPL",
		"Type '\\help' for Neuro commands",
		"Licensed under",
		"Type",
	}

	for _, line := range lines {
		shouldSkip := false
		for _, pattern := range skipPatterns {
			if strings.Contains(line, pattern) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}
