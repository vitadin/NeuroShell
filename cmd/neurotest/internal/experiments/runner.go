// Package experiments provides functionality for recording and running real-world experiments.
package experiments

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"neuroshell/cmd/neurotest/shared"
)

// Runner handles experiment execution
type Runner struct {
	config *shared.Config
}

// NewRunner creates a new experiment runner
func NewRunner(config *shared.Config) *Runner {
	return &Runner{config: config}
}

// RunExperiment runs an experiment and compares with a specific recording
func (r *Runner) RunExperiment(experimentName, sessionID string) error {
	if r.config.Verbose {
		fmt.Printf("Running experiment: %s with session: %s\n", experimentName, sessionID)
	}

	scriptPath, err := FindExperimentScript(experimentName)
	if err != nil {
		return fmt.Errorf("failed to find experiment script: %w", err)
	}

	// Find the recording to compare against
	recordingDir := filepath.Join("experiments", "recordings", experimentName)
	recordingFile := filepath.Join(recordingDir, sessionID+".output")

	if _, err := os.Stat(recordingFile); os.IsNotExist(err) {
		return fmt.Errorf("recording not found: %s", recordingFile)
	}

	// Run the experiment
	actualOutput, duration, err := RunExperimentScript(scriptPath)
	if err != nil && r.config.Verbose {
		fmt.Printf("Experiment failed with error: %v\nOutput: %s\n", err, actualOutput)
	}

	// Read expected output
	expectedContent, err := os.ReadFile(recordingFile)
	if err != nil {
		return fmt.Errorf("failed to read recording file: %w", err)
	}

	expectedOutput := strings.TrimSpace(string(expectedContent))
	actualOutput = strings.TrimSpace(actualOutput)

	if expectedOutput != actualOutput {
		fmt.Printf("Experiment output differs from recording %s\n", sessionID)
		fmt.Printf("Expected length: %d, Actual length: %d\n", len(expectedOutput), len(actualOutput))
		return fmt.Errorf("experiment output differs from recording")
	}

	fmt.Printf("Experiment %s matches recording %s (duration: %v)\n", experimentName, sessionID, duration)
	return nil
}

// FindExperimentScript locates an experiment script
func FindExperimentScript(experimentName string) (string, error) {
	// Check in examples/experiments/ directory
	experimentsDir := "examples/experiments"

	// Try different possible locations and extensions
	candidates := []string{
		filepath.Join(experimentsDir, experimentName, experimentName+".neuro"),
		filepath.Join(experimentsDir, experimentName+".neuro"),
		filepath.Join(experimentsDir, experimentName, "run.neuro"),
		filepath.Join(experimentsDir, experimentName, "experiment.neuro"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("experiment script not found for: %s", experimentName)
}

// RunExperimentScript runs an experiment script with real LLM API calls
func RunExperimentScript(scriptPath string) (string, time.Duration, error) {
	start := time.Now()

	cmd := exec.Command("./bin/neuro", "batch", scriptPath)
	// Don't set test mode - we want real API calls for experiments
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	return string(output), duration, err
}

// FindAllExperiments finds all available experiments
func FindAllExperiments() ([]string, error) {
	experimentsDir := "examples/experiments"

	if _, err := os.Stat(experimentsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("experiments directory not found: %s", experimentsDir)
	}

	entries, err := os.ReadDir(experimentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read experiments directory: %w", err)
	}

	var experiments []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if directory contains experiment files
			dirPath := filepath.Join(experimentsDir, entry.Name())
			if hasExperimentFiles(dirPath) {
				experiments = append(experiments, entry.Name())
			}
		} else if strings.HasSuffix(entry.Name(), ".neuro") {
			// Single file experiment
			name := strings.TrimSuffix(entry.Name(), ".neuro")
			experiments = append(experiments, name)
		}
	}

	return experiments, nil
}

// hasExperimentFiles checks if a directory contains experiment files
func hasExperimentFiles(dirPath string) bool {
	files := []string{"run.neuro", "experiment.neuro"}
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(dirPath, file)); err == nil {
			return true
		}
	}
	// Check if directory name matches a .neuro file
	dirName := filepath.Base(dirPath)
	if _, err := os.Stat(filepath.Join(dirPath, dirName+".neuro")); err == nil {
		return true
	}
	return false
}
