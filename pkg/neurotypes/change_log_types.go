// Package neurotypes defines change log system types for NeuroShell.
// This file contains types for change log entries, statistics, and related data structures.
package neurotypes

// ChangeLogEntry represents a single change log entry with structured metadata.
// Each entry describes a specific change, bug fix, feature addition, or other modification.
type ChangeLogEntry struct {
	ID           string   `yaml:"id" json:"id"`                                 // Unique identifier (e.g., "CL001")
	Version      string   `yaml:"version" json:"version"`                       // Version this change belongs to
	Date         string   `yaml:"date" json:"date"`                             // Date of the change (YYYY-MM-DD format)
	Type         string   `yaml:"type" json:"type"`                             // Type of change (bugfix, feature, enhancement, etc.)
	Title        string   `yaml:"title" json:"title"`                           // Brief title of the change
	Description  string   `yaml:"description" json:"description"`               // Detailed description of what changed
	Impact       string   `yaml:"impact" json:"impact"`                         // Impact on users or system behavior
	FilesChanged []string `yaml:"files_changed" json:"files_changed,omitempty"` // List of files that were modified
}

// ChangeLogStats provides statistics about change log entries.
// Used for generating summary information about development activity.
type ChangeLogStats struct {
	TotalEntries  int            `json:"total_entries"`  // Total number of change log entries
	TypeCounts    map[string]int `json:"type_counts"`    // Count of entries by type (bugfix, feature, etc.)
	VersionCounts map[string]int `json:"version_counts"` // Count of entries by version
}

// ChangeLogType represents valid types for change log entries.
type ChangeLogType string

const (
	// ChangeLogTypeBugfix represents bug fixes and error corrections
	ChangeLogTypeBugfix ChangeLogType = "bugfix"

	// ChangeLogTypeFeature represents new features and functionality
	ChangeLogTypeFeature ChangeLogType = "feature"

	// ChangeLogTypeEnhancement represents improvements to existing features
	ChangeLogTypeEnhancement ChangeLogType = "enhancement"

	// ChangeLogTypeTesting represents testing-related changes
	ChangeLogTypeTesting ChangeLogType = "testing"

	// ChangeLogTypeRefactor represents code refactoring without functional changes
	ChangeLogTypeRefactor ChangeLogType = "refactor"

	// ChangeLogTypeDocs represents documentation changes
	ChangeLogTypeDocs ChangeLogType = "docs"

	// ChangeLogTypeChore represents maintenance and build system changes
	ChangeLogTypeChore ChangeLogType = "chore"
)

// String returns the string representation of a ChangeLogType.
func (t ChangeLogType) String() string {
	return string(t)
}

// IsValid checks if a change log type is valid.
func (t ChangeLogType) IsValid() bool {
	switch t {
	case ChangeLogTypeBugfix, ChangeLogTypeFeature, ChangeLogTypeEnhancement,
		ChangeLogTypeTesting, ChangeLogTypeRefactor, ChangeLogTypeDocs, ChangeLogTypeChore:
		return true
	default:
		return false
	}
}
