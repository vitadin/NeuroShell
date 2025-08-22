// Package testutils provides deterministic generators and utility functions for NeuroShell testing.
// These utilities ensure consistent test output while maintaining production format compatibility.
package testutils

import (
	"fmt"
	"sync"
	"time"

	"neuroshell/pkg/neurotypes"

	"github.com/google/uuid"
)

var (
	// Thread-safe counter for deterministic ID generation
	idCounter uint64
	idMutex   sync.Mutex

	// Thread-safe counter for deterministic timestamp generation
	timeCounter int64
	timeMutex   sync.Mutex
)

// GenerateUUID generates a UUID that is deterministic in test mode but random in production.
// In test mode, returns UUIDs in format: 00000001-0000-4000-8000-000000000001, 00000002-0000-4000-8000-000000000002, etc.
// In production mode, returns standard random UUIDs.
func GenerateUUID(ctx neurotypes.Context) string {
	if ctx.IsTestMode() {
		return getDeterministicUUID()
	}
	return uuid.New().String()
}

// GetCurrentTime returns the current time, deterministic in test mode but real in production.
// In test mode, returns incrementing time starting from 2025-01-01T00:00:00Z
// In production mode, returns time.Now()
func GetCurrentTime(ctx neurotypes.Context) time.Time {
	if ctx.IsTestMode() {
		// Return deterministic but incrementing time for proper sorting
		return getDeterministicTime()
	}
	return time.Now()
}

// FormatDateForDisplay formats a time for display in YYYY-MM-DD format.
// Uses deterministic time in test mode, current time in production.
func FormatDateForDisplay(ctx neurotypes.Context) string {
	return GetCurrentTime(ctx).Format("2006-01-02")
}

// GenerateSessionID generates a deterministic session ID for test mode.
// In test mode, returns format: session_1609459200 (fixed timestamp)
// In production mode, returns format: session_<unix_timestamp>
func GenerateSessionID(ctx neurotypes.Context) string {
	if ctx.IsTestMode() {
		// Use fixed timestamp for deterministic session IDs
		return "session_1609459200" // 2021-01-01 00:00:00 UTC
	}
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

// getDeterministicUUID generates a deterministic UUID maintaining UUID v4 format.
// Returns UUIDs like: 00000001-0000-4000-8000-000000000001, 00000002-0000-4000-8000-000000000002
func getDeterministicUUID() string {
	idMutex.Lock()
	defer idMutex.Unlock()

	idCounter++

	// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// Where 4 indicates version 4, and y is 8, 9, a, or b (we use 8 for simplicity)
	return fmt.Sprintf("%08x-0000-4000-8000-%012x", idCounter, idCounter)
}

// getDeterministicTime generates incrementing deterministic timestamps for test mode.
// Each call returns a time that is 1 second later than the previous call.
// First call: 2025-01-01T00:00:00Z, second call: 2025-01-01T00:00:01Z, etc.
func getDeterministicTime() time.Time {
	timeMutex.Lock()
	defer timeMutex.Unlock()

	timeCounter++

	// Base time: 2025-01-01T00:00:00Z + timeCounter seconds
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return baseTime.Add(time.Duration(timeCounter) * time.Second)
}

var (
	// Thread-safe counter for deterministic temporary file names
	tempFileCounter uint64
	tempFileMutex   sync.Mutex
)

// GenerateTempFilePath generates a deterministic temporary file path for test mode.
// In test mode, returns: /tmp/neuro-command-1.neuro, /tmp/neuro-command-2.neuro, etc.
// In production mode, returns empty string to indicate caller should use standard temp file generation
func GenerateTempFilePath(ctx neurotypes.Context) string {
	if ctx.IsTestMode() {
		tempFileMutex.Lock()
		defer tempFileMutex.Unlock()

		tempFileCounter++
		// Use fixed path for deterministic test results across all platforms
		// This ensures golden files are identical on macOS, Linux, Windows, etc.
		return fmt.Sprintf("/tmp/neuro-command-%d.neuro", tempFileCounter)
	}
	// Return empty string to indicate caller should use standard temp file generation
	return ""
}

// ResetTestCounters resets the deterministic counters for testing.
// This should only be called from test code to ensure consistent test runs.
func ResetTestCounters() {
	idMutex.Lock()
	timeMutex.Lock()
	tempFileMutex.Lock()
	defer idMutex.Unlock()
	defer timeMutex.Unlock()
	defer tempFileMutex.Unlock()

	idCounter = 0
	timeCounter = 0
	tempFileCounter = 0
}
