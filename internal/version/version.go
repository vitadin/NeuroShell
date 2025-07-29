// Package version provides centralized version management for NeuroShell.
// It supports semantic versioning, build-time injection, and version validation.
package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// Build information that can be set at compile time via -ldflags
var (
	// Version is the semantic version of the application
	Version = "0.2.0"

	// GitCommit is the git commit hash when the binary was built
	GitCommit = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
)

// versionCodenames maps version strings to their animal codenames
// Progression based on neural system complexity
var versionCodenames = map[string]string{
	"0.1.0": "Hydra",    // Simple nerve net, basic neural processing
	"0.2.0": "Planaria", // Simple brain, basic learning ability
	"0.3.0": "Aplysia",  // Sea slug, famous in neuroscience research
	"0.4.0": "Octopus",  // Highly intelligent invertebrate
	"0.5.0": "Corvus",   // Crow - exceptional bird intelligence
	"0.6.0": "Rattus",   // Rat - standard neuroscience model
	"0.7.0": "Macaca",   // Macaque - advanced primate cognition
	"0.8.0": "Pan",      // Chimpanzee - tool use, social intelligence
	"0.9.0": "Tursiops", // Dolphin - self-awareness, complex communication
	"1.0.0": "Sapiens",  // Human-level intelligence milestone
	"2.0.0": "Synthia",  // Synthetic/artificial intelligence beyond biological
}

// Info represents comprehensive version information
type Info struct {
	Version   string          `json:"version"`
	Codename  string          `json:"codename"`
	GitCommit string          `json:"gitCommit"`
	BuildDate string          `json:"buildDate"`
	GoVersion string          `json:"goVersion"`
	Platform  string          `json:"platform"`
	SemVer    *semver.Version `json:"-"`
}

// GetVersion returns the current version string
func GetVersion() string {
	return Version
}

// GetCodename returns the codename for the current version
func GetCodename() string {
	return GetCodenameForVersion(Version)
}

// GetBaseVersion returns the base version (major.minor.patch) without build metadata
func GetBaseVersion() string {
	sv, err := semver.NewVersion(Version)
	if err != nil {
		return Version
	}
	return fmt.Sprintf("%d.%d.%d", sv.Major(), sv.Minor(), sv.Patch())
}

// GetBuildMetadata returns the build metadata part of the version (after +)
func GetBuildMetadata() string {
	sv, err := semver.NewVersion(Version)
	if err != nil {
		return ""
	}
	return sv.Metadata()
}

// GetCommitCount returns the commit count from the version build metadata
func GetCommitCount() int {
	// For versions like 0.2.0+123.abc1234, parse the build metadata
	sv, err := semver.NewVersion(Version)
	if err != nil {
		return 0
	}

	metadata := sv.Metadata()
	if metadata == "" {
		return 0
	}

	// Split by dots and try to parse the first part as commit count
	parts := strings.Split(metadata, ".")
	if len(parts) > 0 {
		var commitCount int
		if _, err := fmt.Sscanf(parts[0], "%d", &commitCount); err == nil && commitCount > 0 {
			return commitCount
		}
	}
	return 0
}

// GetCodenameForVersion returns the codename for a specific version
// Handles patch versions by using the major.minor.0 base version
func GetCodenameForVersion(version string) string {
	// First try exact match
	if codename, exists := versionCodenames[version]; exists {
		return codename
	}

	// Parse the version to handle patch versions
	sv, err := semver.NewVersion(version)
	if err != nil {
		return ""
	}

	// Try major.minor.0 format for patch versions
	baseVersion := fmt.Sprintf("%d.%d.0", sv.Major(), sv.Minor())
	if codename, exists := versionCodenames[baseVersion]; exists {
		return codename
	}

	return ""
}

// GetInfo returns comprehensive version information
func GetInfo() (*Info, error) {
	// Parse semantic version
	sv, err := semver.NewVersion(Version)
	if err != nil {
		return nil, fmt.Errorf("invalid semantic version '%s': %w", Version, err)
	}

	return &Info{
		Version:   Version,
		Codename:  GetCodename(),
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		SemVer:    sv,
	}, nil
}

// GetFormattedVersion returns a nicely formatted version string
func GetFormattedVersion() string {
	info, err := GetInfo()
	if err != nil {
		return fmt.Sprintf("NeuroShell v%s (invalid version)", Version)
	}

	var parts []string

	// Format version with codename if available
	if info.Codename != "" {
		parts = append(parts, fmt.Sprintf("NeuroShell v%s '%s'", info.Version, info.Codename))
	} else {
		parts = append(parts, fmt.Sprintf("NeuroShell v%s", info.Version))
	}

	if info.GitCommit != "unknown" && info.GitCommit != "" {
		// Show short commit hash (7 characters)
		shortCommit := info.GitCommit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		parts = append(parts, fmt.Sprintf("commit %s", shortCommit))
	}

	if info.BuildDate != "unknown" && info.BuildDate != "" {
		parts = append(parts, fmt.Sprintf("built %s", info.BuildDate))
	}

	return strings.Join(parts, ", ")
}

// GetDetailedVersion returns detailed version information for debugging
func GetDetailedVersion() string {
	info, err := GetInfo()
	if err != nil {
		return fmt.Sprintf("NeuroShell v%s (error: %v)", Version, err)
	}

	var lines []string

	// Format version with codename if available
	if info.Codename != "" {
		lines = append(lines, fmt.Sprintf("NeuroShell v%s '%s'", info.Version, info.Codename))
		lines = append(lines, fmt.Sprintf("Codename: %s", info.Codename))
	} else {
		lines = append(lines, fmt.Sprintf("NeuroShell v%s", info.Version))
	}

	lines = append(lines, fmt.Sprintf("Git Commit: %s", info.GitCommit))
	lines = append(lines, fmt.Sprintf("Build Date: %s", info.BuildDate))

	// Show commit count and build metadata if available
	if commitCount := GetCommitCount(); commitCount > 0 {
		lines = append(lines, fmt.Sprintf("Commit Count: %d", commitCount))
	}
	if buildMeta := GetBuildMetadata(); buildMeta != "" {
		lines = append(lines, fmt.Sprintf("Build Metadata: %s", buildMeta))
	}

	lines = append(lines, fmt.Sprintf("Go Version: %s", info.GoVersion))
	lines = append(lines, fmt.Sprintf("Platform: %s", info.Platform))

	return strings.Join(lines, "\n")
}

// ValidateVersion validates that the current version is a valid semantic version
func ValidateVersion() error {
	_, err := semver.NewVersion(Version)
	if err != nil {
		return fmt.Errorf("invalid semantic version '%s': %w", Version, err)
	}
	return nil
}

// IsPrerelease returns true if the current version is a prerelease
func IsPrerelease() bool {
	sv, err := semver.NewVersion(Version)
	if err != nil {
		return false
	}
	return sv.Prerelease() != ""
}

// IsDevelopment returns true if this appears to be a development build
func IsDevelopment() bool {
	return GitCommit == "unknown" || BuildDate == "unknown"
}

// CompareVersions compares two version strings and returns:
// -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	sv1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version v1 '%s': %w", v1, err)
	}

	sv2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version v2 '%s': %w", v2, err)
	}

	return sv1.Compare(sv2), nil
}

// SetBuildInfo sets build information (used for testing)
func SetBuildInfo(version, gitCommit, buildDate string) {
	Version = version
	GitCommit = gitCommit
	BuildDate = buildDate
}

// GetBuildTime returns the build time as a time.Time if parseable
func GetBuildTime() (time.Time, error) {
	if BuildDate == "unknown" || BuildDate == "" {
		return time.Time{}, fmt.Errorf("build date not available")
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, BuildDate); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse build date '%s'", BuildDate)
}
