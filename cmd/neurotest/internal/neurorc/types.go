// Package neurorc provides testing functionality for .neurorc startup behavior.
package neurorc

// TestConfig represents the configuration for a .neurorc test case
type TestConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Setup       TestSetup         `json:"setup"`
	CLIFlags    []string          `json:"cli_flags"`
	EnvVars     map[string]string `json:"env_vars"`
	InputSeq    []string          `json:"input_sequence"`
	ExpectedOut []string          `json:"expected_contains"`
	ExpectedNot []string          `json:"expected_not_contains"`
	Timeout     int               `json:"timeout_seconds"`
}

// TestSetup represents the setup configuration for test environment
type TestSetup struct {
	WorkingDirNeuroRC string            `json:"working_dir_neurorc"`
	HomeDirNeuroRC    string            `json:"home_dir_neurorc"`
	CustomFiles       map[string]string `json:"custom_files"` // filename -> content
}

// TestEnvironment represents an isolated test environment
type TestEnvironment struct {
	TempDir     string
	WorkingDir  string
	HomeDir     string
	ConfigFiles map[string]string
}
