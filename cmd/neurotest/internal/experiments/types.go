// Package experiments provides functionality for recording and running real-world experiments.
package experiments

import "time"

// ExperimentMetadata contains metadata about an experiment recording
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

// ExperimentSummary contains a summary of all recordings for an experiment
type ExperimentSummary struct {
	ExperimentName   string               `json:"experiment_name"`
	TotalRecordings  int                  `json:"total_recordings"`
	LatestRecording  time.Time            `json:"latest_recording"`
	RecordingHistory []ExperimentMetadata `json:"recording_history"`
}
