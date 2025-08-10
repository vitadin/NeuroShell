// Package experiments provides functionality for recording and running real-world experiments.
package experiments

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"neuroshell/cmd/neurotest/shared"
)

// Recorder handles experiment recording
type Recorder struct {
	config *shared.Config
}

// NewRecorder creates a new experiment recorder
func NewRecorder(config *shared.Config) *Recorder {
	return &Recorder{config: config}
}

// RecordExperiment records a real-world experiment execution
func (r *Recorder) RecordExperiment(experimentName string) error {
	if r.config.Verbose {
		fmt.Printf("Recording experiment: %s\n", experimentName)
	}

	scriptPath, err := FindExperimentScript(experimentName)
	if err != nil {
		return fmt.Errorf("failed to find experiment script: %w", err)
	}

	sessionID := GenerateSessionID()
	fmt.Printf("Recording experiment %s with session ID: %s\n", experimentName, sessionID)

	output, duration, err := RunExperimentScript(scriptPath)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}

		if r.config.Verbose {
			fmt.Printf("Experiment completed with exit code %d: %v\n", exitCode, err)
		}
	}

	recordingPath, err := r.saveExperimentRecording(experimentName, scriptPath, output, duration, exitCode, sessionID)
	if err != nil {
		return fmt.Errorf("failed to save recording: %w", err)
	}

	fmt.Printf("Experiment recorded successfully at: %s\n", recordingPath)
	return nil
}

// RecordAllExperiments records all available experiments
func (r *Recorder) RecordAllExperiments() error {
	experiments, err := FindAllExperiments()
	if err != nil {
		return fmt.Errorf("failed to find experiments: %w", err)
	}

	if len(experiments) == 0 {
		fmt.Println("No experiments found in examples/experiments/")
		return nil
	}

	fmt.Printf("Found %d experiments to record\n", len(experiments))

	var failed []string
	for i, experiment := range experiments {
		fmt.Printf("\n[%d/%d] Recording experiment: %s\n", i+1, len(experiments), experiment)

		if err := r.RecordExperiment(experiment); err != nil {
			fmt.Printf("Failed to record %s: %v\n", experiment, err)
			failed = append(failed, experiment)
		} else {
			fmt.Printf("Successfully recorded: %s\n", experiment)
		}
	}

	fmt.Printf("\nRecording complete. Success: %d, Failed: %d\n",
		len(experiments)-len(failed), len(failed))

	if len(failed) > 0 {
		fmt.Printf("Failed experiments: %v\n", failed)
		return fmt.Errorf("failed to record %d experiments", len(failed))
	}

	return nil
}

// saveExperimentRecording saves experiment output and metadata
func (r *Recorder) saveExperimentRecording(experimentName, scriptPath, output string, duration time.Duration, exitCode int, sessionID string) (string, error) {
	recordingDir, err := EnsureRecordingsDir(experimentName)
	if err != nil {
		return "", err
	}

	// Save output
	outputFile := filepath.Join(recordingDir, sessionID+".output")
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	// Create metadata
	metadata := ExperimentMetadata{
		ExperimentName: experimentName,
		SessionID:      sessionID,
		Timestamp:      time.Now(),
		Duration:       duration,
		ExitCode:       exitCode,
		ScriptPath:     scriptPath,
		Environment:    GetRelevantEnvVars(),
		OutputFile:     outputFile,
	}

	// Save metadata
	metadataFile := filepath.Join(recordingDir, sessionID+".metadata.json")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, metadataJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write metadata file: %w", err)
	}

	// Update experiment summary
	if err := r.updateExperimentSummary(experimentName, metadata); err != nil {
		// Log error but don't fail the recording
		fmt.Printf("Warning: failed to update experiment summary: %v\n", err)
	}

	return outputFile, nil
}

// updateExperimentSummary updates the summary file for an experiment
func (r *Recorder) updateExperimentSummary(experimentName string, newRecording ExperimentMetadata) error {
	recordingDir := filepath.Join("experiments", "recordings", experimentName)
	summaryFile := filepath.Join(recordingDir, "summary.json")

	var summary ExperimentSummary

	// Load existing summary if it exists
	if data, err := os.ReadFile(summaryFile); err == nil {
		if err := json.Unmarshal(data, &summary); err != nil {
			return fmt.Errorf("failed to parse existing summary: %w", err)
		}
	} else {
		// Initialize new summary
		summary = ExperimentSummary{
			ExperimentName: experimentName,
		}
	}

	// Update summary
	summary.TotalRecordings++
	summary.LatestRecording = newRecording.Timestamp
	summary.RecordingHistory = append(summary.RecordingHistory, newRecording)

	// Save updated summary
	summaryJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	return os.WriteFile(summaryFile, summaryJSON, 0644)
}

// GenerateSessionID creates a unique session identifier
func GenerateSessionID() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("session_%s_%d", timestamp, time.Now().UnixNano()%10000)
}

// EnsureRecordingsDir creates the recordings directory for an experiment
func EnsureRecordingsDir(experimentName string) (string, error) {
	recordingDir := filepath.Join("experiments", "recordings", experimentName)
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create recording directory: %w", err)
	}
	return recordingDir, nil
}

// GetRelevantEnvVars returns environment variables relevant to experiments
func GetRelevantEnvVars() map[string]string {
	relevantVars := []string{
		"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_API_KEY",
		"NEURO_LOG_LEVEL", "NEURO_CONFIG_DIR",
	}

	envMap := make(map[string]string)
	for _, varName := range relevantVars {
		if value := os.Getenv(varName); value != "" {
			// Don't store actual API keys for security
			if strings.Contains(varName, "API_KEY") {
				envMap[varName] = "***REDACTED***"
			} else {
				envMap[varName] = value
			}
		}
	}

	return envMap
}
