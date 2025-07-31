// Package main provides experiment-specific functionality for neurotest.
// This module handles real-world experiment execution with actual LLM API calls.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	experimentsDir = "examples/experiments"
	recordingsDir  = "experiments/recordings"
)

// ExperimentMetadata contains information about an experiment recording
type ExperimentMetadata struct {
	ExperimentName string            `json:"experiment_name"`
	SessionID      string            `json:"session_id"`
	Timestamp      time.Time         `json:"timestamp"`
	Duration       time.Duration     `json:"duration"`
	ExitCode       int               `json:"exit_code"`
	ScriptPath     string            `json:"script_path"`
	Environment    map[string]string `json:"environment"`
	OutputFile     string            `json:"output_file"`
}

// ExperimentSummary contains summary information for an experiment folder
type ExperimentSummary struct {
	ExperimentName   string               `json:"experiment_name"`
	TotalRecordings  int                  `json:"total_recordings"`
	LatestRecording  time.Time            `json:"latest_recording"`
	RecordingHistory []ExperimentMetadata `json:"recording_history"`
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Generate 4 random bytes for short ID
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-only if random generation fails
		return timestamp + "_fallback"
	}
	shortID := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%s_%s", timestamp, shortID)
}

// findExperimentScript locates a .neuro script file in the experiments directory
func findExperimentScript(experimentName string) (string, error) {
	// Try different possible locations in experiments directory
	candidates := []string{
		filepath.Join(experimentsDir, experimentName+".neuro"),
		filepath.Join(experimentsDir, "openai", experimentName+".neuro"),
		filepath.Join(experimentsDir, "gemini", experimentName+".neuro"),
		filepath.Join(experimentsDir, "anthropic", experimentName+".neuro"),
	}

	// Also try nested directories
	err := filepath.Walk(experimentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".neuro") {
			nameWithoutExt := strings.TrimSuffix(info.Name(), ".neuro")
			if nameWithoutExt == experimentName {
				candidates = append(candidates, path)
			}
		}
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to search experiments directory: %w", err)
	}

	// Check each candidate
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("experiment script not found: %s\nTried: %v", experimentName, candidates)
}

// findAllExperiments discovers all available experiment scripts
func findAllExperiments() ([]string, error) {
	var experiments []string
	experimentSet := make(map[string]bool) // To avoid duplicates

	if _, err := os.Stat(experimentsDir); os.IsNotExist(err) {
		return experiments, nil // No experiments directory
	}

	err := filepath.Walk(experimentsDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".neuro") {
			// Extract experiment name (filename without extension)
			experimentName := strings.TrimSuffix(info.Name(), ".neuro")
			if !experimentSet[experimentName] {
				experiments = append(experiments, experimentName)
				experimentSet[experimentName] = true
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover experiments: %w", err)
	}

	return experiments, nil
}

// runExperimentScript executes a .neuro script WITHOUT test mode (real API calls)
func runExperimentScript(scriptPath string) (string, time.Duration, error) {
	start := time.Now()

	// Convert to absolute path for consistent execution
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Use neuro batch command to run the script (NO --test-mode for real API calls)
	cmd := exec.Command(neurocmd, "batch", absPath)

	// Set environment variables for experiment execution
	cmd.Env = append(os.Environ(),
		"NEURO_LOG_LEVEL=info", // Keep some logging for debugging
	)

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		return string(output), duration, fmt.Errorf("experiment execution failed: %w", err)
	}

	// Clean up the output by removing shell startup messages
	cleanOutput := cleanNeuroOutput(string(output))
	return cleanOutput, duration, nil
}

// ensureRecordingsDir creates the recordings directory structure
func ensureRecordingsDir(experimentName string) (string, error) {
	experimentDir := filepath.Join(recordingsDir, experimentName)

	if err := os.MkdirAll(experimentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create recordings directory: %w", err)
	}

	return experimentDir, nil
}

// saveExperimentRecording saves experiment output and metadata
func saveExperimentRecording(experimentName, scriptPath, output string, duration time.Duration, exitCode int) (string, error) {
	// Ensure recordings directory exists
	experimentDir, err := ensureRecordingsDir(experimentName)
	if err != nil {
		return "", err
	}

	// Generate unique session ID
	sessionID := generateSessionID()
	outputFile := sessionID + ".expected"
	outputPath := filepath.Join(experimentDir, outputFile)

	// Save the output
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("failed to save experiment output: %w", err)
	}

	// Create metadata
	metadata := ExperimentMetadata{
		ExperimentName: experimentName,
		SessionID:      sessionID,
		Timestamp:      time.Now(),
		Duration:       duration,
		ExitCode:       exitCode,
		ScriptPath:     scriptPath,
		Environment:    getRelevantEnvVars(),
		OutputFile:     outputFile,
	}

	// Save individual metadata
	metadataPath := filepath.Join(experimentDir, sessionID+".metadata.json")
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	// Update experiment summary
	if err := updateExperimentSummary(experimentName, metadata); err != nil {
		// Don't fail the whole operation if summary update fails
		fmt.Printf("Warning: failed to update experiment summary: %v\n", err)
	}

	return sessionID, nil
}

// getRelevantEnvVars extracts environment variables relevant to experiments
func getRelevantEnvVars() map[string]string {
	relevantKeys := []string{
		"OPENAI_API_KEY",
		"GOOGLE_API_KEY",
		"ANTHROPIC_API_KEY",
		"NEURO_LOG_LEVEL",
		"PATH",
	}

	env := make(map[string]string)
	for _, key := range relevantKeys {
		if value := os.Getenv(key); value != "" {
			// Mask API keys for security
			if strings.Contains(strings.ToUpper(key), "API_KEY") {
				if len(value) > 8 {
					env[key] = value[:4] + "..." + value[len(value)-4:]
				} else {
					env[key] = "***"
				}
			} else {
				env[key] = value
			}
		}
	}

	return env
}

// updateExperimentSummary updates the summary file for an experiment
func updateExperimentSummary(experimentName string, newRecording ExperimentMetadata) error {
	experimentDir := filepath.Join(recordingsDir, experimentName)
	summaryPath := filepath.Join(experimentDir, "metadata.json")

	var summary ExperimentSummary

	// Load existing summary if it exists
	if data, err := os.ReadFile(summaryPath); err == nil {
		if err := json.Unmarshal(data, &summary); err != nil {
			// Ignore unmarshal errors and start with empty summary
			summary = ExperimentSummary{}
		}
	}

	// Update summary
	summary.ExperimentName = experimentName
	summary.TotalRecordings++
	summary.LatestRecording = newRecording.Timestamp
	summary.RecordingHistory = append(summary.RecordingHistory, newRecording)

	// Keep only last 50 recordings in history to prevent file from growing too large
	if len(summary.RecordingHistory) > 50 {
		summary.RecordingHistory = summary.RecordingHistory[len(summary.RecordingHistory)-50:]
	}

	// Save updated summary
	summaryBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	return os.WriteFile(summaryPath, summaryBytes, 0644)
}
