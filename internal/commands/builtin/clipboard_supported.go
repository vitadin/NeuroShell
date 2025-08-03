//go:build !linux

package builtin

import "golang.design/x/clipboard"

// clipboardAvailable indicates if clipboard functionality is available on this platform
const clipboardAvailable = true

// initClipboard initializes the clipboard library
func initClipboard() error {
	return clipboard.Init()
}

// writeToClipboard writes text to the system clipboard
func writeToClipboard(text string) error {
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}
