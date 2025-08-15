package output

import (
	"os"
	"sync"
)

// Global printer instance for convenience functions
var (
	globalPrinter *Printer
	globalMu      sync.RWMutex
)

// Initialize sets up the global printer with default settings.
func init() {
	globalPrinter = NewPrinter()
}

// SetGlobalPrinter sets the global printer instance.
// This allows configuration of the global output behavior.
func SetGlobalPrinter(printer *Printer) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalPrinter = printer
}

// GetGlobalPrinter returns the current global printer instance.
func GetGlobalPrinter() *Printer {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalPrinter
}

// ConfigureGlobal configures the global printer with the given options.
func ConfigureGlobal(options ...Option) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalPrinter = NewPrinter(options...)
}

// Global convenience functions that use the global printer instance.
// These provide direct replacements for fmt.Print* functions.

// Print outputs text using the global printer.
func Print(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Print(text)
}

// Printf outputs formatted text using the global printer.
func Printf(format string, args ...interface{}) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Printf(format, args...)
}

// Println outputs text with newline using the global printer.
func Println(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Println(text)
}

// Info outputs informational text using the global printer.
func Info(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Info(text)
}

// Success outputs success text using the global printer.
func Success(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Success(text)
}

// Warning outputs warning text using the global printer.
func Warning(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Warning(text)
}

// Error outputs error text using the global printer.
func Error(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Error(text)
}

// Command outputs command text using the global printer.
func Command(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Command(text)
}

// Variable outputs variable text using the global printer.
func Variable(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Variable(text)
}

// Keyword outputs keyword text using the global printer.
func Keyword(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Keyword(text)
}

// Highlight outputs highlighted text using the global printer.
func Highlight(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Highlight(text)
}

// Bold outputs bold text using the global printer.
func Bold(text string) {
	globalMu.RLock()
	printer := globalPrinter
	globalMu.RUnlock()
	printer.Bold(text)
}

// IsTerminal checks if the output is going to a terminal.
// This is useful for deciding whether to use colors or not.
func IsTerminal() bool {
	// Check if stdout is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice
}

// SupportsColor returns true if the output supports color.
// This is a simple heuristic based on terminal detection and environment variables.
func SupportsColor() bool {
	// Check common environment variables that indicate color support
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if term := os.Getenv("TERM"); term != "" {
		// Basic terminal types that support color
		colorTerms := []string{"xterm", "xterm-256color", "screen", "tmux"}
		for _, colorTerm := range colorTerms {
			if term == colorTerm {
				return true
			}
		}
	}

	// Fall back to terminal check
	return IsTerminal()
}
