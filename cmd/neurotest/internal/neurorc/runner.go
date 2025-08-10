// Package neurorc provides testing functionality for .neurorc startup behavior.
package neurorc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neuroshell/cmd/neurotest/shared"
)

// Runner handles NeuroRC test execution
type Runner struct {
	config *shared.Config
}

// NewRunner creates a new NeuroRC test runner
func NewRunner(config *shared.Config) *Runner {
	return &Runner{config: config}
}

// RunTest runs a .neurorc startup test case
func (r *Runner) RunTest(testName string) error {
	if r.config.Verbose {
		fmt.Printf("Running .neurorc test: %s\n", testName)
	}

	configFile, err := FindTestConfig(testName)
	if err != nil {
		return fmt.Errorf("failed to find test config: %w", err)
	}

	testConfig, err := ParseTestConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}

	// Create isolated test environment
	testEnv, err := CreateTestEnvironment(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create test environment: %w", err)
	}
	defer CleanupTestEnvironment(testEnv)

	// Run the test
	actualOutput, err := RunShellTest(testEnv, testConfig)
	if err != nil {
		return fmt.Errorf("failed to run shell test: %w", err)
	}

	actualOutput = CleanOutput(actualOutput)

	// Read expected output
	expectedPath := filepath.Join(r.config.TestDir, "neurorc", testName+".expected")
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return fmt.Errorf("failed to read expected file %s: %w", expectedPath, err)
	}

	expectedOutput := strings.TrimRight(string(expectedContent), "\n")

	if actualOutput != expectedOutput {
		return fmt.Errorf("test failed: output doesn't match expected")
	}

	if r.config.Verbose {
		fmt.Printf(".neurorc test passed: %s\n", testName)
	}

	return nil
}
