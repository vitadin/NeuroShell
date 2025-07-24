// Package stringprocessing provides utilities for string processing and manipulation.
// It includes functions for boolean evaluation, text parsing, and condition checking
// that are commonly used across NeuroShell commands.
package stringprocessing

import "strings"

// IsTruthy determines if a string represents a truthy value using NeuroShell's boolean evaluation rules.
// This function is used by conditional commands like \if and \if-not for consistent boolean logic.
//
// EVALUATION RULES (case-insensitive):
//   - Explicitly TRUTHY: 'true', '1', 'yes', 'on', 'enabled'
//   - Explicitly FALSY: 'false', '0', 'no', 'off', 'disabled'
//   - Empty strings: FALSY ("" or whitespace-only)
//   - Any other non-empty string: TRUTHY (including Unicode characters)
//
// Examples:
//
//	IsTruthy("true")     -> true
//	IsTruthy("false")    -> false
//	IsTruthy("")         -> false
//	IsTruthy("hello")    -> true
//	IsTruthy("0")        -> false
//	IsTruthy("1")        -> true
//	IsTruthy("  ")       -> false (whitespace-only)
//	IsTruthy("🌟")       -> true (Unicode)
func IsTruthy(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))

	// Empty string is falsy
	if value == "" {
		return false
	}

	// Common truthy values
	truthyValues := map[string]bool{
		"true":    true,
		"1":       true,
		"yes":     true,
		"on":      true,
		"enabled": true,
	}

	// Common falsy values
	falsyValues := map[string]bool{
		"false":    true,
		"0":        true,
		"no":       true,
		"off":      true,
		"disabled": true,
	}

	// Check explicit truthy/falsy values
	if truthyValues[value] {
		return true
	}
	if falsyValues[value] {
		return false
	}

	// Any non-empty string is considered truthy
	return true
}

// IsFalsy is the logical inverse of IsTruthy for convenience and readability.
// It returns true when IsTruthy returns false, and vice versa.
//
// This function is provided for semantic clarity when the falsy check is more natural:
//
//	if stringprocessing.IsFalsy(condition) { ... }
//
// vs:
//
//	if !stringprocessing.IsTruthy(condition) { ... }
func IsFalsy(value string) bool {
	return !IsTruthy(value)
}
