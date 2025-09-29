// Package stringprocessing provides text processing utilities for NeuroShell.
// This package contains functions for handling escape sequences and continuation markers
// used in shell-like input processing for markdown rendering and command parsing.
package stringprocessing

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
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

// WithCapturedOutput captures stdout during function execution and returns both the output and any error.
// This utility function redirects stdout to capture all output while preserving stderr for errors.
// It's designed for automatic output capture in the command execution flow.
func WithCapturedOutput(fn func() error) (string, error) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// Redirect stdout to pipe
	os.Stdout = w

	// Channel to receive output
	outputChan := make(chan string, 1)

	// Start goroutine to read output
	go func() {
		defer close(outputChan)
		output, _ := io.ReadAll(r)
		outputChan <- string(output)
	}()

	// Execute function and capture any error
	fnErr := fn()

	// Close writer to signal end of writing
	_ = w.Close()

	// Restore stdout
	os.Stdout = oldStdout

	// Get captured output (this will block until the goroutine finishes reading)
	capturedOutput := <-outputChan

	// Close reader after reading is complete
	_ = r.Close()

	// Return both output and error
	return capturedOutput, fnErr
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

// IsAPIRelated checks if a variable name is API-related based on case-insensitive keyword matching.
// Returns true only if the name contains API-related keywords (api, key, secret).
// If true, also returns the detected provider name or "generic" if no provider is detected.
// This implements the filtering criteria: variables must contain API keywords to be considered API-related.
func IsAPIRelated(variableName string, providers []string) (bool, string) {
	nameLower := strings.ToLower(variableName)

	// API-related keywords (case-insensitive) - required for API detection
	apiKeywords := []string{"api", "key", "secret"}
	hasAPIKeyword := false
	for _, keyword := range apiKeywords {
		if strings.Contains(nameLower, keyword) {
			hasAPIKeyword = true
			break
		}
	}

	// If no API keywords found, not API-related regardless of provider names
	if !hasAPIKeyword {
		return false, ""
	}

	// Provider names (case-insensitive) - optional for provider detection
	for _, provider := range providers {
		providerLower := strings.ToLower(provider)
		if strings.Contains(nameLower, providerLower) {
			return true, provider
		}

		// Special case: GOOGLE_API_KEY should be detected as "gemini" provider
		if provider == "gemini" && strings.Contains(nameLower, "google") {
			return true, provider
		}
	}

	// Has API keywords but no specific provider detected
	return true, "generic"
}

// StringPtr returns a pointer to the given string value.
// This is a utility function for creating string pointers, commonly used in structs
// where fields are defined as *string to allow for optional/nullable values.
func StringPtr(s string) *string {
	return &s
}

// Float64Ptr returns a pointer to the given float64 value.
// This is a utility function for creating float64 pointers, commonly used in structs
// where fields are defined as *float64 to allow for optional/nullable values.
func Float64Ptr(f float64) *float64 {
	return &f
}

// IntPtr returns a pointer to the given int value.
// This is a utility function for creating int pointers, commonly used in structs
// where fields are defined as *int to allow for optional/nullable values.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
// This is a utility function for creating bool pointers, commonly used in structs
// where fields are defined as *bool to allow for optional/nullable values.
func BoolPtr(b bool) *bool {
	return &b
}

// StripANSIEscapeCodes removes ANSI escape sequences from a string using a mature library.
// This includes color codes, cursor positioning, and other terminal control sequences.
// This is useful for cleaning captured output to store as plain text.
func StripANSIEscapeCodes(input string) string {
	// Use the charmbracelet/x/ansi library which properly handles all ANSI escape codes
	return ansi.Strip(input)
}
