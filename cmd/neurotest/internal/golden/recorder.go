// Package golden provides golden file testing functionality.
package golden

import (
	"fmt"
	"os"
	"path/filepath"

	"neuroshell/cmd/neurotest/internal/normalize"
	"neuroshell/cmd/neurotest/shared"
)

// Recorder handles recording golden file test cases
type Recorder struct {
	config     *shared.Config
	normalizer *normalize.NormalizationEngine
}

// NewRecorder creates a new golden file test recorder
func NewRecorder(config *shared.Config) *Recorder {
	return &Recorder{
		config:     config,
		normalizer: normalize.NewNormalizationEngine(),
	}
}

// RecordTest records a new test case by running a .neuro script and saving output
func (r *Recorder) RecordTest(testName string) error {
	if r.config.Verbose {
		fmt.Printf("Recording test: %s\n", testName)
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
	if err := os.WriteFile(expectedPath, []byte(cleanedOutput), 0644); err != nil {
		return fmt.Errorf("failed to write expected file: %w", err)
	}

	if r.config.Verbose {
		fmt.Printf("Recorded expected output for test: %s\n", testName)
	}

	return nil
}

// AcceptTest updates the golden file for a test case with current output
func (r *Recorder) AcceptTest(testName string) error {
	return r.RecordTest(testName) // Same implementation
}

// RecordCFlagTest records a test case using the -c flag
func (r *Recorder) RecordCFlagTest(testName string) error {
	if r.config.Verbose {
		fmt.Printf("Recording -c flag test: %s\n", testName)
	}

	scriptPath, err := shared.FindScript(testName, r.config.TestDir)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("test script not found: %s", scriptPath)
	}

	output, err := shared.RunNeuroCFlag(scriptPath, r.config.NeuroCmd, r.config.TestTimeout)
	if err != nil {
		if r.config.Verbose {
			fmt.Printf("Command failed with error: %v\nOutput: %s\n", err, output)
		}
	}

	cleanedOutput := r.cleanOutput(output)

	expectedPath := filepath.Join(r.config.TestDir, testName+".c.expected")
	if err := os.WriteFile(expectedPath, []byte(cleanedOutput), 0644); err != nil {
		return fmt.Errorf("failed to write -c expected file: %w", err)
	}

	if r.config.Verbose {
		fmt.Printf("Recorded -c flag expected output for test: %s\n", testName)
	}

	return nil
}

// cleanOutput normalizes output for recording
func (r *Recorder) cleanOutput(output string) string {
	cleaned := shared.CleanOutput(output)
	return r.normalizer.NormalizeOutput(cleaned)
}
