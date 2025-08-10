// Package normalize provides output normalization functionality for test comparisons.
package normalize

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"
)

// NormalizationPattern represents a pattern for normalizing test output
type NormalizationPattern struct {
	Name    string
	Pattern *regexp.Regexp
	MinLen  int
	MaxLen  int
}

// NormalizationEngine handles smart normalization of test output
type NormalizationEngine struct {
	patterns []NormalizationPattern
}

// NewNormalizationEngine creates a new normalization engine with built-in patterns
func NewNormalizationEngine() *NormalizationEngine {
	engine := &NormalizationEngine{}
	engine.initBuiltinPatterns()
	return engine
}

// initBuiltinPatterns initializes the built-in normalization patterns
func (ne *NormalizationEngine) initBuiltinPatterns() {
	// Session IDs and timestamps are deterministic in test mode, so no normalization needed
	// Temp file paths are also kept as-is to show actual CLI behavior

	// Username masking - replace machine-specific usernames with placeholder
	ne.addUsernameMasking()

	// Cross-platform error message normalization
	ne.addCrossPlatformPatterns()

	// Memory addresses
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "memory_address",
		Pattern: regexp.MustCompile(`0x[a-fA-F0-9]{8,16}`),
		MinLen:  10,
		MaxLen:  18,
	})

	// Process IDs
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "process_id",
		Pattern: regexp.MustCompile(`\bpid:\s*\d+\b`),
		MinLen:  6,
		MaxLen:  15,
	})
}

// addUsernameMasking adds username masking pattern based on current user
func (ne *NormalizationEngine) addUsernameMasking() {
	// Get current username from system
	currentUser, err := user.Current()
	if err != nil {
		// Fallback to environment variable if user.Current() fails
		if username := os.Getenv("USER"); username != "" {
			ne.addUsernamePattern(username)
		}
		return
	}

	if currentUser.Username != "" {
		ne.addUsernamePattern(currentUser.Username)
	}
}

// addUsernamePattern adds a pattern to mask the specified username
func (ne *NormalizationEngine) addUsernamePattern(username string) {
	// Only mask if username is not empty and has reasonable length
	if len(username) < 2 || len(username) > 50 {
		return
	}

	// Skip common/generic usernames that are not personally identifying
	if ne.isCommonUsername(username) {
		return
	}

	// Create pattern to match the exact username (word boundary to avoid partial matches)
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(username) + `\b`)

	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "username",
		Pattern: pattern,
		MinLen:  len(username),
		MaxLen:  len(username),
	})
}

// isCommonUsername checks if the username is a common/generic one that shouldn't be masked
func (ne *NormalizationEngine) isCommonUsername(username string) bool {
	commonUsernames := []string{
		// CI/CD and automation
		"runner", "ci", "github", "gitlab", "jenkins", "build", "deploy",
		// Generic system users
		"admin", "administrator", "user", "test", "testuser", "guest",
		// Common service accounts
		"root", "daemon", "nobody", "www-data", "nginx", "apache",
		// Development/testing
		"dev", "developer", "tester", "qa", "demo", "example",
		// Cloud/container common names
		"ubuntu", "centos", "debian", "alpine", "node", "app",
	}

	usernameLower := strings.ToLower(username)
	for _, common := range commonUsernames {
		if usernameLower == common {
			return true
		}
	}

	return false
}

// addCrossPlatformPatterns adds normalization for OS-specific differences
func (ne *NormalizationEngine) addCrossPlatformPatterns() {
	// ls command error messages differ between macOS and Linux
	// macOS: "ls: /path: No such file or directory"
	// Linux: "ls: cannot access '/path': No such file or directory"
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "ls_error_macos",
		Pattern: regexp.MustCompile(`ls: ([^:]+): No such file or directory`),
		MinLen:  10,
		MaxLen:  200,
	})

	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "ls_error_linux",
		Pattern: regexp.MustCompile(`ls: cannot access '([^']+)': No such file or directory`),
		MinLen:  10,
		MaxLen:  200,
	})

	// Exit status normalization for any non-zero exit status
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "non_zero_exit_status",
		Pattern: regexp.MustCompile(`Exit status: [1-9]\d*`),
		MinLen:  15,
		MaxLen:  30,
	})

	// Project path normalization (different between dev and CI environments)
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "project_path",
		Pattern: regexp.MustCompile(`(/Users/[^/]+/GolandProjects/NeuroShell|/home/[^/]+/project)`),
		MinLen:  10,
		MaxLen:  200,
	})
}

// NormalizeOutput normalizes the given output by replacing dynamic content with placeholders
func (ne *NormalizationEngine) NormalizeOutput(output string) string {
	normalized := output
	for _, pattern := range ne.patterns {
		placeholder := "<" + pattern.Name + ">"
		normalized = pattern.Pattern.ReplaceAllString(normalized, placeholder)
	}
	return normalized
}

// IsPlaceholderLine checks if a line contains a placeholder pattern
func (ne *NormalizationEngine) IsPlaceholderLine(line string) bool {
	return strings.Contains(line, "<") && strings.Contains(line, ">")
}

// ParsePlaceholder extracts placeholder information from a placeholder string
func (ne *NormalizationEngine) ParsePlaceholder(placeholder string) (string, int, int, error) {
	// Extract placeholder name between < and >
	start := strings.Index(placeholder, "<")
	end := strings.Index(placeholder, ">")
	if start == -1 || end == -1 || end <= start {
		return "", 0, 0, fmt.Errorf("invalid placeholder format")
	}

	name := placeholder[start+1 : end]

	// Find matching pattern
	for _, pattern := range ne.patterns {
		if pattern.Name == name {
			return name, pattern.MinLen, pattern.MaxLen, nil
		}
	}

	// Default values for unknown placeholders
	return name, 1, 100, nil
}

// MatchLineWithPlaceholders checks if an actual line matches an expected line with placeholders
func (ne *NormalizationEngine) MatchLineWithPlaceholders(expected, actual string) bool {
	if !ne.IsPlaceholderLine(expected) {
		return expected == actual
	}

	// For placeholder matching, use normalized comparison
	normalizedExpected := ne.NormalizeOutput(expected)
	normalizedActual := ne.NormalizeOutput(actual)
	return normalizedExpected == normalizedActual
}

// CompareWithPlaceholders compares two outputs considering placeholder patterns
func (ne *NormalizationEngine) CompareWithPlaceholders(expected, actual string) bool {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	if len(expectedLines) != len(actualLines) {
		return false
	}

	for i, expectedLine := range expectedLines {
		if !ne.MatchLineWithPlaceholders(expectedLine, actualLines[i]) {
			return false
		}
	}

	return true
}

// GetCurrentTimestamp returns the current timestamp in ISO format
func (ne *NormalizationEngine) GetCurrentTimestamp() string {
	return time.Now().Format("2006-01-02T15:04:05.000Z")
}
