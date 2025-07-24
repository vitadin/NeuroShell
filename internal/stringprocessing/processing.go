// Package stringprocessing provides text processing utilities for NeuroShell.
// This package contains functions for handling escape sequences and continuation markers
// used in shell-like input processing for markdown rendering and command parsing.
package stringprocessing

import (
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

// WithSuppressedOutput suppresses stdout during function execution.
// This utility function redirects stdout to discard all output, preserving stderr for errors.
// It's designed for the \silent command to suppress all fmt.Print* output globally.
func WithSuppressedOutput(fn func() error) error {
	// Save original stdout
	oldStdout := os.Stdout

	// Open /dev/null for discarding output
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	// Redirect stdout to /dev/null
	os.Stdout = devNull

	// Execute function and capture any error
	fnErr := fn()

	// Restore stdout and close /dev/null
	_ = devNull.Close()
	os.Stdout = oldStdout

	// Return the function's error (not redirection errors)
	return fnErr
}
