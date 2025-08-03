//go:build linux

package builtin

import "fmt"

// clipboardAvailable indicates if clipboard functionality is available on this platform
const clipboardAvailable = false

// initClipboard returns an error indicating clipboard is not available
func initClipboard() error {
	return fmt.Errorf("clipboard not available on this platform (Linux without X11)")
}

// writeToClipboard returns an error indicating clipboard is not available
func writeToClipboard(text string) error {
	return fmt.Errorf("clipboard not available on this platform (Linux without X11)")
}
