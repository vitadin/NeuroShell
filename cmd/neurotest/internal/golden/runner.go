// Package golden provides golden file testing functionality.
package golden

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neuroshell/cmd/neurotest/internal/normalize"
	"neuroshell/cmd/neurotest/shared"
)

// Runner handles running golden file tests
type Runner struct {
	config     *shared.Config
	normalizer *normalize.NormalizationEngine
}

// NewRunner creates a new golden file test runner
func NewRunner(config *shared.Config) *Runner {
	return &Runner{
		config:     config,
		normalizer: normalize.NewNormalizationEngine(),
	}
}

// RunTest runs a specific test case and compares with expected output
func (r *Runner) RunTest(testName string) error {
	if r.config.Verbose {
		fmt.Printf("Running test: %s\n", testName)
	}

	scriptPath, err := shared.FindScript(testName, r.config.TestDir)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("test script not found: %s", scriptPath)
	}

	output, err := shared.RunNeuroScript(scriptPath, r.config.NeuroCmd, r.config.TestTimeout)
	if err != nil {
		if r.config.Verbose {
			fmt.Printf("Command failed with error: %v\nOutput: %s\n", err, output)
		}
	}

	cleanedOutput := r.cleanOutput(output)

	expectedPath := filepath.Join(r.config.TestDir, testName+".expected")
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return fmt.Errorf("failed to read expected file %s: %w", expectedPath, err)
	}

	expectedOutput := strings.TrimRight(string(expectedContent), "\n")

	if !r.normalizer.CompareWithPlaceholders(expectedOutput, cleanedOutput) {
		return fmt.Errorf("test failed: output doesn't match expected")
	}

	if r.config.Verbose {
		fmt.Printf("Test passed: %s\n", testName)
	}

	return nil
}

// RunAllTests runs all tests in the test directory
func (r *Runner) RunAllTests() error {
	tests, err := shared.FindAllFiles(r.config.TestDir, ".neuro")
	if err != nil {
		return fmt.Errorf("failed to find tests: %w", err)
	}

	var failedTests []string
	passedTests := 0

	for _, test := range tests {
		if err := r.RunTest(test); err != nil {
			failedTests = append(failedTests, test)
			fmt.Printf("FAIL %s: %v\n", test, err)
		} else {
			passedTests++
			fmt.Printf("PASS %s\n", test)
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed\n", passedTests, len(failedTests))

	if len(failedTests) > 0 {
		return fmt.Errorf("tests failed: %v", failedTests)
	}

	return nil
}

// cleanOutput normalizes output for comparison
func (r *Runner) cleanOutput(output string) string {
	cleaned := shared.CleanOutput(output)
	return r.normalizer.NormalizeOutput(cleaned)
}
