// Package stringprocessing provides text processing utilities for NeuroShell.
// This package contains functions for handling escape sequences and continuation markers
// used in shell-like input processing for markdown rendering and command parsing.
package stringprocessing

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// InterpretEscapeSequences converts escape sequences in a string to their actual characters.
// This function is extracted from the echo command to provide shared string processing.
func InterpretEscapeSequences(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\'", "'")
	return s
}

// CleanContinuationMarkers removes shell continuation markers from multiline input.
// This handles markers like "..." that shells use to indicate continuation lines.
// This function does NOT process escape sequences - that's handled separately.
func CleanContinuationMarkers(text string) string {
	// Split text into lines for processing
	lines := strings.Split(text, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Remove leading/trailing whitespace for pattern matching
		trimmed := strings.TrimSpace(line)

		// Skip lines that are just continuation markers
		if trimmed == "..." {
			continue
		}

		// Remove continuation markers at the beginning of lines
		if strings.HasPrefix(trimmed, "... ") {
			// Remove "... " prefix and trim the rest
			cleaned := strings.TrimPrefix(trimmed, "... ")
			cleaned = strings.TrimSpace(cleaned)
			cleanedLines = append(cleanedLines, cleaned)
		} else {
			// Keep the line as-is
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// ProcessTextForMarkdown processes text for markdown rendering with configurable options.
// This combines continuation marker cleanup with optional escape sequence processing.
func ProcessTextForMarkdown(text string, interpretEscapes bool) string {
	// Handle the mixed case where input might have both \n and actual newlines
	var processedText string

	if interpretEscapes {
		// First normalize \n to actual newlines, then clean continuation markers
		normalizedText := strings.ReplaceAll(text, "\\n", "\n")
		processedText = CleanContinuationMarkers(normalizedText)
	} else {
		// For raw mode, we need to be careful about \n handling
		// First, clean continuation markers without converting \n
		if strings.Contains(text, "\n") && !strings.Contains(text, "\\n") {
			// Input has actual newlines, clean continuation markers directly
			processedText = CleanContinuationMarkers(text)
		} else {
			// Input has \n escape sequences, temporarily convert to process markers
			tempText := strings.ReplaceAll(text, "\\n", "\n")
			cleanedText := CleanContinuationMarkers(tempText)
			// Convert back to escape sequences to preserve original format
			processedText = strings.ReplaceAll(cleanedText, "\n", "\\n")
		}
	}

	// Apply escape sequence interpretation if requested
	if interpretEscapes {
		processedText = InterpretEscapeSequences(processedText)
	}

	return processedText
}

// CaptureOutput captures stdout during function execution for testing purposes.
// This utility function redirects stdout to capture output from functions that write to it.
func CaptureOutput(fn func()) string {
	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Channel to receive output
	outputChan := make(chan string)

	// Start goroutine to read output
	go func() {
		defer close(outputChan)
		output, _ := io.ReadAll(r)
		outputChan <- string(output)
	}()

	// Execute function
	fn()

	// Restore stdout and close writer
	_ = w.Close()
	os.Stdout = oldStdout

	// Return captured output
	return <-outputChan
}

// IsTruthy determines if a string represents a truthy value with safety protections.
// This function provides shared condition evaluation logic for \if and \while commands.
//
// Condition evaluation rules (case-insensitive):
//   - Explicitly TRUTHY: 'true', '1', 'yes', 'on', 'enabled'
//   - Explicitly FALSY: 'false', '0', 'no', 'off', 'disabled'
//   - Empty strings: FALSY (including whitespace-only)
//   - Any other non-empty string: TRUTHY
//
// Safety protections:
//   - Returns error if condition string is too long (prevents infinite loops from malformed conditions)
func IsTruthy(value string) (bool, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	// Safety check: prevent infinite loops from malformed conditions
	// Malformed conditions (like from semicolon errors) tend to grow very long
	const maxConditionLength = 200
	if len(value) > maxConditionLength {
		return false, fmt.Errorf("condition too long (%d chars), possible malformed condition", len(value))
	}

	// Empty string is falsy
	if value == "" {
		return false, nil
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
		return true, nil
	}
	if falsyValues[value] {
		return false, nil
	}

	// Any non-empty string is considered truthy
	return true, nil
}
