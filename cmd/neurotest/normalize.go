// Package main provides normalization functions for neurotest golden file testing.
// This module handles smart comparison by replacing variable content with placeholders.
package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// NormalizationPattern represents a pattern for normalizing output
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
	// UUID pattern (for session IDs, etc.)
	uuidPattern := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

	// Timestamp patterns (various formats)
	timestampPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?`), // ISO 8601
		regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`),                                 // Common format
		regexp.MustCompile(`\d{10,13}`),                                                           // Unix timestamp
	}

	// Path patterns (absolute paths)
	pathPattern := regexp.MustCompile(`(?:[A-Za-z]:\\|/)[^\s<>"]*`)

	// Username patterns (common in paths)
	userPattern := regexp.MustCompile(`/Users/[^/\s]+|/home/[^/\s]+|C:\\Users\\[^\\s]+`)

	// Add UUID pattern
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "UUID",
		Pattern: uuidPattern,
		MinLen:  36,
		MaxLen:  36,
	})

	// Add timestamp patterns
	for i, pattern := range timestampPatterns {
		ne.patterns = append(ne.patterns, NormalizationPattern{
			Name:    fmt.Sprintf("TIMESTAMP_%d", i),
			Pattern: pattern,
			MinLen:  8,
			MaxLen:  30,
		})
	}

	// Add path pattern
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "PATH",
		Pattern: pathPattern,
		MinLen:  3,
		MaxLen:  500,
	})

	// Add user pattern
	ne.patterns = append(ne.patterns, NormalizationPattern{
		Name:    "USER",
		Pattern: userPattern,
		MinLen:  5,
		MaxLen:  100,
	})
}

// NormalizeOutput normalizes output by replacing known patterns with placeholders
func (ne *NormalizationEngine) NormalizeOutput(output string) string {
	normalized := output

	// Apply each normalization pattern
	for _, pattern := range ne.patterns {
		normalized = pattern.Pattern.ReplaceAllStringFunc(normalized, func(match string) string {
			// Check if match length is within expected range
			if len(match) >= pattern.MinLen && len(match) <= pattern.MaxLen {
				return fmt.Sprintf("{{%s}}", pattern.Name)
			}
			return match
		})
	}

	return normalized
}

// IsPlaceholderLine checks if a line contains placeholder syntax
func (ne *NormalizationEngine) IsPlaceholderLine(line string) bool {
	placeholderPattern := regexp.MustCompile(`\{\{[^}]+\}\}`)
	return placeholderPattern.MatchString(line)
}

// ParsePlaceholder parses a placeholder and returns its type and constraints
func (ne *NormalizationEngine) ParsePlaceholder(placeholder string) (string, int, int, error) {
	// Remove {{ and }}
	content := strings.TrimPrefix(strings.TrimSuffix(placeholder, "}}"), "{{")

	// Check if it has length constraints like PLACEHOLDER:10:15
	parts := strings.Split(content, ":")
	if len(parts) == 1 {
		// Simple placeholder without constraints
		return parts[0], 0, 1000, nil
	} else if len(parts) == 3 {
		// Placeholder with min:max constraints
		minLen := 0
		maxLen := 1000

		if parts[1] != "" {
			if _, err := fmt.Sscanf(parts[1], "%d", &minLen); err != nil {
				return "", 0, 0, fmt.Errorf("invalid min length: %s", parts[1])
			}
		}

		if parts[2] != "" {
			if _, err := fmt.Sscanf(parts[2], "%d", &maxLen); err != nil {
				return "", 0, 0, fmt.Errorf("invalid max length: %s", parts[2])
			}
		}

		return parts[0], minLen, maxLen, nil
	}

	return "", 0, 0, fmt.Errorf("invalid placeholder format: %s", placeholder)
}

// MatchLineWithPlaceholders compares a line containing placeholders with actual output
func (ne *NormalizationEngine) MatchLineWithPlaceholders(expected, actual string) bool {
	// If no placeholders, do exact match
	if !ne.IsPlaceholderLine(expected) {
		return strings.TrimSpace(expected) == strings.TrimSpace(actual)
	}

	// Find all placeholders in expected line
	placeholderPattern := regexp.MustCompile(`\{\{[^}]+\}\}`)
	placeholders := placeholderPattern.FindAllString(expected, -1)

	// Create regex pattern by replacing placeholders
	regexPattern := regexp.QuoteMeta(expected)
	for _, placeholder := range placeholders {
		placeholderType, minLen, maxLen, err := ne.ParsePlaceholder(placeholder)
		if err != nil {
			continue
		}

		// Create regex for this placeholder
		var placeholderRegex string
		switch placeholderType {
		case "TIMESTAMP":
			placeholderRegex = `\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?|\d{10,13}`
		case "UUID":
			placeholderRegex = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
		case "PATH":
			placeholderRegex = `(?:[A-Za-z]:\\|/)[^\s<>"]*.{0,500}`
		case "USER":
			placeholderRegex = `/Users/[^/\s]+|/home/[^/\s]+|C:\\Users\\[^\\s]+`
		case "PLACEHOLDER":
			placeholderRegex = fmt.Sprintf(`.{%d,%d}`, minLen, maxLen)
		default:
			placeholderRegex = `.+`
		}

		// Replace the quoted placeholder with regex
		quotedPlaceholder := regexp.QuoteMeta(placeholder)
		regexPattern = strings.ReplaceAll(regexPattern, quotedPlaceholder, placeholderRegex)
	}

	// Compile and match
	compiledPattern, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		return false
	}

	return compiledPattern.MatchString(strings.TrimSpace(actual))
}

// CompareWithPlaceholders compares expected output (with placeholders) against actual output
func (ne *NormalizationEngine) CompareWithPlaceholders(expected, actual string) bool {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	// If different number of lines, they don't match
	if len(expectedLines) != len(actualLines) {
		return false
	}

	// Compare line by line
	for i, expectedLine := range expectedLines {
		if i >= len(actualLines) {
			return false
		}

		if !ne.MatchLineWithPlaceholders(expectedLine, actualLines[i]) {
			return false
		}
	}

	return true
}

// GetCurrentTimestamp returns current timestamp for testing
func (ne *NormalizationEngine) GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
