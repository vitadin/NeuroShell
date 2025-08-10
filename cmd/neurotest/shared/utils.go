// Package shared provides common utilities for neurotest.
package shared

import (
	"path/filepath"
	"strings"
)

// FindScript locates a test script in the test directory
func FindScript(testName, testDir string) (string, error) {
	scriptPath := filepath.Join(testDir, testName+".neuro")
	return scriptPath, nil
}

// CleanOutput normalizes neuro command output for consistent comparison
func CleanOutput(output string) string {
	lines := strings.Split(output, "\n")

	// Only remove trailing newlines from the entire output,
	// but preserve trailing spaces within lines as they may be meaningful
	// (e.g., "Setting _style = " for empty variable assignments)
	cleaned := strings.Join(lines, "\n")
	cleaned = strings.TrimRight(cleaned, "\n")

	return cleaned
}

// FindAllFiles finds all files with the specified extension in a directory
func FindAllFiles(dir, extension string) ([]string, error) {
	pattern := filepath.Join(dir, "*"+extension)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var basenames []string
	for _, match := range matches {
		basename := filepath.Base(match)
		basenames = append(basenames, strings.TrimSuffix(basename, extension))
	}

	return basenames, nil
}
