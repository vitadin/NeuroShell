// Package testutils provides deterministic generators and utility functions for NeuroShell testing.
// These utilities ensure consistent test output while maintaining production format compatibility.
package testutils

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"neuroshell/pkg/neurotypes"
)

var (
	// Thread-safe counter for deterministic ID generation
	idCounter uint64
	idMutex   sync.Mutex
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
// In test mode, returns fixed time: 2025-01-01T00:00:00Z
// In production mode, returns time.Now()
func GetCurrentTime(ctx neurotypes.Context) time.Time {
	if ctx.IsTestMode() {
		// Return fixed deterministic time
		return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
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

// ResetTestCounters resets the deterministic counters for testing.
// This should only be called from test code to ensure consistent test runs.
func ResetTestCounters() {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter = 0
}
