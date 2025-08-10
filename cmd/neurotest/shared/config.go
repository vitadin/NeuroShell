// Package shared provides common configuration and utilities for neurotest.
package shared

// Config holds the global configuration for neurotest
type Config struct {
	TestDir     string
	NeuroCmd    string
	Verbose     bool
	TestTimeout int
}

// Default configuration values
const (
	DefaultTestDir     = "test/golden"
	DefaultNeuroCmd    = "neuro"
	DefaultTestTimeout = 30
)

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		TestDir:     DefaultTestDir,
		NeuroCmd:    DefaultNeuroCmd,
		Verbose:     false,
		TestTimeout: DefaultTestTimeout,
	}
}
