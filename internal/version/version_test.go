// Package version_test provides tests for version management functionality.
package version

import (
	"testing"
)

func TestGetCodenameForVersion(t *testing.T) {
	tests := []struct {
		name             string
		version          string
		expectedCodename string
	}{
		{
			name:             "exact match for 0.2.0",
			version:          "0.2.0",
			expectedCodename: "Planaria",
		},
		{
			name:             "patch version 0.2.1 should use 0.2.0 codename",
			version:          "0.2.1",
			expectedCodename: "Planaria",
		},
		{
			name:             "exact match for 0.2.5",
			version:          "0.2.5",
			expectedCodename: "Dendro",
		},
		{
			name:             "patch version 0.2.99 should use 0.2.0 codename",
			version:          "0.2.99",
			expectedCodename: "Planaria",
		},
		{
			name:             "future version 1.0.0 (codename commented out)",
			version:          "1.0.0",
			expectedCodename: "",
		},
		{
			name:             "future patch version 1.0.1 (codename commented out)",
			version:          "1.0.1",
			expectedCodename: "",
		},
		{
			name:             "version without codename",
			version:          "0.10.0",
			expectedCodename: "",
		},
		{
			name:             "patch version without base codename",
			version:          "0.10.5",
			expectedCodename: "",
		},
		{
			name:             "invalid version",
			version:          "invalid",
			expectedCodename: "",
		},
		{
			name:             "prerelease version should use base codename",
			version:          "0.2.0-alpha.1",
			expectedCodename: "Planaria",
		},
		{
			name:             "patch prerelease version should use base codename",
			version:          "0.2.3-beta.2",
			expectedCodename: "Planaria",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCodenameForVersion(tt.version)
			if result != tt.expectedCodename {
				t.Errorf("GetCodenameForVersion(%q) = %q, want %q", tt.version, result, tt.expectedCodename)
			}
		})
	}
}

func TestGetCodename(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name             string
		version          string
		expectedCodename string
	}{
		{
			name:             "current version 0.2.0",
			version:          "0.2.0",
			expectedCodename: "Planaria",
		},
		{
			name:             "current version 0.2.1",
			version:          "0.2.1",
			expectedCodename: "Planaria",
		},
		{
			name:             "current version without codename",
			version:          "0.10.0",
			expectedCodename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := GetCodename()
			if result != tt.expectedCodename {
				t.Errorf("GetCodename() with Version=%q = %q, want %q", tt.version, result, tt.expectedCodename)
			}
		})
	}
}

func TestVersionCodenames(t *testing.T) {
	// Test that all currently defined codenames are accessible
	// (Future codenames are commented out until their respective releases)
	expectedCodenames := map[string]string{
		"0.1.0": "Hydra",
		"0.2.0": "Planaria",
		"0.2.3": "Planaria",
		"0.2.4": "Planaria",
		"0.2.5": "Dendro",
	}

	for version, expectedCodename := range expectedCodenames {
		t.Run("version_"+version, func(t *testing.T) {
			result := GetCodenameForVersion(version)
			if result != expectedCodename {
				t.Errorf("GetCodenameForVersion(%q) = %q, want %q", version, result, expectedCodename)
			}
		})
	}
}

func TestGetInfoWithCodename(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	Version = "0.2.0"

	info, err := GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	if info.Version != "0.2.0" {
		t.Errorf("GetInfo().Version = %q, want %q", info.Version, "0.2.0")
	}

	if info.Codename != "Planaria" {
		t.Errorf("GetInfo().Codename = %q, want %q", info.Codename, "Planaria")
	}
}

func TestValidateVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "valid version",
			version:     "1.2.3",
			expectError: false,
		},
		{
			name:        "valid version with prerelease",
			version:     "1.2.3-alpha.1",
			expectError: false,
		},
		{
			name:        "invalid version",
			version:     "invalid",
			expectError: true,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			err := ValidateVersion()
			if tt.expectError && err == nil {
				t.Errorf("ValidateVersion() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateVersion() unexpected error: %v", err)
			}
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "stable version",
			version:  "1.2.3",
			expected: false,
		},
		{
			name:     "prerelease alpha",
			version:  "1.2.3-alpha.1",
			expected: true,
		},
		{
			name:     "prerelease beta",
			version:  "1.2.3-beta.2",
			expected: true,
		},
		{
			name:     "prerelease rc",
			version:  "1.2.3-rc.1",
			expected: true,
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := IsPrerelease()
			if result != tt.expected {
				t.Errorf("IsPrerelease() with version %q = %v, want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name      string
		gitCommit string
		buildDate string
		expected  bool
	}{
		{
			name:      "development build - unknown commit",
			gitCommit: "unknown",
			buildDate: "2023-01-01",
			expected:  true,
		},
		{
			name:      "development build - unknown date",
			gitCommit: "abc1234",
			buildDate: "unknown",
			expected:  true,
		},
		{
			name:      "development build - both unknown",
			gitCommit: "unknown",
			buildDate: "unknown",
			expected:  true,
		},
		{
			name:      "production build",
			gitCommit: "abc1234",
			buildDate: "2023-01-01",
			expected:  false,
		},
	}

	// Save original values
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	defer func() {
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GitCommit = tt.gitCommit
			BuildDate = tt.buildDate
			result := IsDevelopment()
			if result != tt.expected {
				t.Errorf("IsDevelopment() with GitCommit=%q, BuildDate=%q = %v, want %v",
					tt.gitCommit, tt.buildDate, result, tt.expected)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
		hasError bool
	}{
		{
			name:     "v1 less than v2",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
			hasError: false,
		},
		{
			name:     "v1 greater than v2",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
			hasError: false,
		},
		{
			name:     "v1 equals v2",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
			hasError: false,
		},
		{
			name:     "prerelease comparison",
			v1:       "1.0.0-alpha.1",
			v2:       "1.0.0",
			expected: -1,
			hasError: false,
		},
		{
			name:     "invalid v1",
			v1:       "invalid",
			v2:       "1.0.0",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid v2",
			v1:       "1.0.0",
			v2:       "invalid",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)
			if tt.hasError && err == nil {
				t.Errorf("CompareVersions(%q, %q) expected error but got none", tt.v1, tt.v2)
			}
			if !tt.hasError && err != nil {
				t.Errorf("CompareVersions(%q, %q) unexpected error: %v", tt.v1, tt.v2, err)
			}
			if !tt.hasError && result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestSetBuildInfo(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	defer func() {
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	testVersion := "1.2.3"
	testCommit := "abc1234"
	testDate := "2023-01-01"

	SetBuildInfo(testVersion, testCommit, testDate)

	if Version != testVersion {
		t.Errorf("SetBuildInfo() Version = %q, want %q", Version, testVersion)
	}
	if GitCommit != testCommit {
		t.Errorf("SetBuildInfo() GitCommit = %q, want %q", GitCommit, testCommit)
	}
	if BuildDate != testDate {
		t.Errorf("SetBuildInfo() BuildDate = %q, want %q", BuildDate, testDate)
	}
}

func TestGetBuildTime(t *testing.T) {
	// Save original build date
	originalBuildDate := BuildDate
	defer func() {
		BuildDate = originalBuildDate
	}()

	tests := []struct {
		name           string
		buildDate      string
		expectError    bool
		expectedFormat string
	}{
		{
			name:           "RFC3339 format",
			buildDate:      "2023-01-01T12:00:00Z",
			expectError:    false,
			expectedFormat: "2006-01-02T15:04:05Z",
		},
		{
			name:           "date only format",
			buildDate:      "2023-01-01",
			expectError:    false,
			expectedFormat: "2006-01-02",
		},
		{
			name:           "datetime format",
			buildDate:      "2023-01-01 12:00:00",
			expectError:    false,
			expectedFormat: "2006-01-02 15:04:05",
		},
		{
			name:        "unknown build date",
			buildDate:   "unknown",
			expectError: true,
		},
		{
			name:        "empty build date",
			buildDate:   "",
			expectError: true,
		},
		{
			name:        "invalid format",
			buildDate:   "invalid-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuildDate = tt.buildDate
			result, err := GetBuildTime()

			if tt.expectError && err == nil {
				t.Errorf("GetBuildTime() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("GetBuildTime() unexpected error: %v", err)
			}
			if !tt.expectError && !result.IsZero() {
				// Verify the time was parsed correctly by formatting it back
				formatted := result.Format(tt.expectedFormat)
				if formatted != tt.buildDate {
					t.Errorf("GetBuildTime() parsed time incorrectly, got %q, want %q", formatted, tt.buildDate)
				}
			}
		})
	}
}

func TestGetBaseVersion(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "standard version",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with build metadata",
			version:  "0.2.0+123.abc1234",
			expected: "0.2.0",
		},
		{
			name:     "version with prerelease",
			version:  "1.2.3-alpha.1",
			expected: "1.2.3",
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := GetBaseVersion()
			if result != tt.expected {
				t.Errorf("GetBaseVersion() with version %q = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetBuildMetadata(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "version with build metadata",
			version:  "0.2.0+123.abc1234",
			expected: "123.abc1234",
		},
		{
			name:     "version without build metadata",
			version:  "1.2.3",
			expected: "",
		},
		{
			name:     "version with complex metadata",
			version:  "1.0.0+build.1.sha.abc1234",
			expected: "build.1.sha.abc1234",
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := GetBuildMetadata()
			if result != tt.expected {
				t.Errorf("GetBuildMetadata() with version %q = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetCommitCount(t *testing.T) {
	// Save original version
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	tests := []struct {
		name     string
		version  string
		expected int
	}{
		{
			name:     "version with commit count",
			version:  "0.2.0+123.abc1234",
			expected: 123,
		},
		{
			name:     "version with zero commit count",
			version:  "0.2.0+0.abc1234",
			expected: 0,
		},
		{
			name:     "standard version without commit count",
			version:  "1.2.3",
			expected: 0,
		},
		{
			name:     "version with non-numeric commit count",
			version:  "0.2.0+alpha.abc1234",
			expected: 0,
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			result := GetCommitCount()
			if result != tt.expected {
				t.Errorf("GetCommitCount() with version %q = %d, want %d", tt.version, result, tt.expected)
			}
		})
	}
}
