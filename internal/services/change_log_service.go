// Package services provides the change log service for NeuroShell.
// It manages embedded change log data and provides filtering capabilities.
package services

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"neuroshell/internal/data/embedded"
	"neuroshell/pkg/neurotypes"
)

// ChangeLogService provides access to embedded change log data with filtering and search capabilities.
// It implements the service pattern used throughout NeuroShell for accessing embedded resources.
type ChangeLogService struct {
	entries     []neurotypes.ChangeLogEntry
	initialized bool
}

// changeLogData represents the structure of the embedded change log YAML file.
type changeLogData struct {
	Entries []neurotypes.ChangeLogEntry `yaml:"entries"`
}

// NewChangeLogService creates a new change log service instance.
func NewChangeLogService() *ChangeLogService {
	return &ChangeLogService{
		initialized: false,
	}
}

// Name returns the service name "change_log" for registration.
func (s *ChangeLogService) Name() string {
	return "change_log"
}

// Initialize sets up the ChangeLogService for operation by loading embedded data.
func (s *ChangeLogService) Initialize() error {
	if err := s.loadEmbeddedData(); err != nil {
		return fmt.Errorf("failed to load embedded change log data: %w", err)
	}

	// Validate the data
	if err := s.Validate(); err != nil {
		return fmt.Errorf("change log data validation failed: %w", err)
	}

	s.initialized = true
	return nil
}

// loadEmbeddedData loads and parses the embedded change log YAML data.
func (s *ChangeLogService) loadEmbeddedData() error {
	var data changeLogData
	if err := yaml.Unmarshal(embedded.ChangeLogData, &data); err != nil {
		return fmt.Errorf("failed to parse change log YAML: %w", err)
	}

	s.entries = data.Entries
	return nil
}

// GetChangeLog returns all change log entries sorted by date (newest first).
func (s *ChangeLogService) GetChangeLog() ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	// Create a copy to avoid modifying the original data
	entries := make([]neurotypes.ChangeLogEntry, len(s.entries))
	copy(entries, s.entries)

	// Sort by date (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	return entries, nil
}

// GetChangeLogWithOrder returns all change log entries sorted by the specified order.
// order can be "asc" (oldest first) or "desc" (newest first).
func (s *ChangeLogService) GetChangeLogWithOrder(order string) ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	// Create a copy to avoid modifying the original data
	entries := make([]neurotypes.ChangeLogEntry, len(s.entries))
	copy(entries, s.entries)

	// Sort by date based on order
	if order == "desc" {
		// Sort by date (newest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Date > entries[j].Date
		})
	} else {
		// Sort by date (oldest first) - default behavior
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Date < entries[j].Date
		})
	}

	return entries, nil
}

// SearchChangeLog returns change log entries that match the search query.
// The search is case-insensitive and matches across ID, version, type, title, description, and impact fields.
func (s *ChangeLogService) SearchChangeLog(query string) ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	if query == "" {
		return s.GetChangeLog()
	}

	var matches []neurotypes.ChangeLogEntry
	queryLower := strings.ToLower(query)

	for _, entry := range s.entries {
		// Search across all relevant fields
		if strings.Contains(strings.ToLower(entry.ID), queryLower) ||
			strings.Contains(strings.ToLower(entry.Version), queryLower) ||
			strings.Contains(strings.ToLower(entry.Type), queryLower) ||
			strings.Contains(strings.ToLower(entry.Title), queryLower) ||
			strings.Contains(strings.ToLower(entry.Description), queryLower) ||
			strings.Contains(strings.ToLower(entry.Impact), queryLower) {
			matches = append(matches, entry)
		}
	}

	// Sort matches by date (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Date > matches[j].Date
	})

	return matches, nil
}

// SearchChangeLogWithOrder returns change log entries that match the search query, sorted by the specified order.
// order can be "asc" (oldest first) or "desc" (newest first).
func (s *ChangeLogService) SearchChangeLogWithOrder(query string, order string) ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	if query == "" {
		return s.GetChangeLogWithOrder(order)
	}

	var matches []neurotypes.ChangeLogEntry
	queryLower := strings.ToLower(query)

	for _, entry := range s.entries {
		// Search across all relevant fields
		if strings.Contains(strings.ToLower(entry.ID), queryLower) ||
			strings.Contains(strings.ToLower(entry.Version), queryLower) ||
			strings.Contains(strings.ToLower(entry.Type), queryLower) ||
			strings.Contains(strings.ToLower(entry.Title), queryLower) ||
			strings.Contains(strings.ToLower(entry.Description), queryLower) ||
			strings.Contains(strings.ToLower(entry.Impact), queryLower) {
			matches = append(matches, entry)
		}
	}

	// Sort matches by date based on order
	if order == "desc" {
		// Sort by date (newest first)
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Date > matches[j].Date
		})
	} else {
		// Sort by date (oldest first) - default behavior
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Date < matches[j].Date
		})
	}

	return matches, nil
}

// GetChangeLogByType returns change log entries filtered by type (bugfix, feature, enhancement, etc.).
func (s *ChangeLogService) GetChangeLogByType(entryType string) ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	var matches []neurotypes.ChangeLogEntry
	typeLower := strings.ToLower(entryType)

	for _, entry := range s.entries {
		if strings.ToLower(entry.Type) == typeLower {
			matches = append(matches, entry)
		}
	}

	// Sort by date (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Date > matches[j].Date
	})

	return matches, nil
}

// GetChangeLogByVersion returns change log entries for a specific version.
func (s *ChangeLogService) GetChangeLogByVersion(version string) ([]neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	var matches []neurotypes.ChangeLogEntry

	for _, entry := range s.entries {
		if entry.Version == version {
			matches = append(matches, entry)
		}
	}

	// Sort by date (newest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Date > matches[j].Date
	})

	return matches, nil
}

// GetEntryByID returns a specific change log entry by its ID.
func (s *ChangeLogService) GetEntryByID(id string) (*neurotypes.ChangeLogEntry, error) {
	if !s.initialized {
		return nil, fmt.Errorf("change log service not initialized")
	}

	for _, entry := range s.entries {
		if entry.ID == id {
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("change log entry with ID '%s' not found", id)
}

// GetChangeLogStats returns statistics about the change log entries.
func (s *ChangeLogService) GetChangeLogStats() neurotypes.ChangeLogStats {
	if !s.initialized {
		return neurotypes.ChangeLogStats{}
	}

	stats := neurotypes.ChangeLogStats{
		TotalEntries:  len(s.entries),
		TypeCounts:    make(map[string]int),
		VersionCounts: make(map[string]int),
	}

	for _, entry := range s.entries {
		stats.TypeCounts[entry.Type]++
		stats.VersionCounts[entry.Version]++
	}

	return stats
}

// Validate checks the integrity of the change log data.
func (s *ChangeLogService) Validate() error {
	idSet := make(map[string]bool)

	for i, entry := range s.entries {
		// Check for required fields
		if entry.ID == "" {
			return fmt.Errorf("entry %d: missing ID", i)
		}
		if entry.Title == "" {
			return fmt.Errorf("entry %s: missing title", entry.ID)
		}
		if entry.Type == "" {
			return fmt.Errorf("entry %s: missing type", entry.ID)
		}
		if entry.Date == "" {
			return fmt.Errorf("entry %s: missing date", entry.ID)
		}

		// Check for duplicate IDs
		if idSet[entry.ID] {
			return fmt.Errorf("duplicate entry ID: %s", entry.ID)
		}
		idSet[entry.ID] = true

		// Validate entry type
		validTypes := map[string]bool{
			"bugfix":      true,
			"feature":     true,
			"enhancement": true,
			"testing":     true,
			"refactor":    true,
			"docs":        true,
			"chore":       true,
		}
		if !validTypes[entry.Type] {
			return fmt.Errorf("entry %s: invalid type '%s'", entry.ID, entry.Type)
		}
	}

	return nil
}

// GetGlobalChangeLogService returns the global change log service instance from the service registry.
func GetGlobalChangeLogService() (*ChangeLogService, error) {
	service, err := GetGlobalRegistry().GetService("change_log")
	if err != nil {
		return nil, fmt.Errorf("change log service not available: %w", err)
	}
	return service.(*ChangeLogService), nil
}
