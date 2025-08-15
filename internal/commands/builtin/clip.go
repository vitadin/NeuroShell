package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ClipCommand implements the \clip command for copying text to the system clipboard.
// It copies everything after the command to the clipboard with variable interpolation.
type ClipCommand struct{}

// Name returns the command name "clip" for registration and lookup.
func (c *ClipCommand) Name() string {
	return "clip"
}

// ParseMode returns ParseModeRaw to handle literal text with variable interpolation.
func (c *ClipCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the clip command does.
func (c *ClipCommand) Description() string {
	return "Copy text to system clipboard"
}

// Usage returns the syntax and usage examples for the clip command.
func (c *ClipCommand) Usage() string {
	return "\\clip text to copy"
}

// HelpInfo returns structured help information for the clip command.
func (c *ClipCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\clip text to copy",
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\clip Hello World!",
				Description: "Copy literal text to clipboard",
			},
			{
				Command:     "\\clip Hello ${name}!",
				Description: "Copy text with variable interpolation",
			},
			{
				Command:     "\\clip ${_output}",
				Description: "Copy the last command output to clipboard",
			},
			{
				Command:     "\\clip ${1}",
				Description: "Copy the latest AI response to clipboard",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_clipboard",
				Description: "Fallback storage when system clipboard is unavailable",
				Type:        "system_output",
				Example:     "_clipboard = \"Hello World!\"",
			},
		},
		Notes: []string{
			"Copies text with full variable interpolation support",
			"Empty input shows warning and leaves clipboard unchanged",
			"Falls back to _clipboard variable if system clipboard unavailable",
			"Feedback shows number of characters copied",
		},
	}
}

// Execute copies the input text to the system clipboard with variable interpolation.
// Provides concise feedback showing character count and handles errors gracefully.
func (c *ClipCommand) Execute(_ map[string]string, input string) error {
	// Handle empty input - non-destructive approach
	if strings.TrimSpace(input) == "" {
		printer := c.createPrinter()
		printer.Warning("No content specified. Clipboard unchanged.")
		return nil
	}

	// Check if clipboard is available on this platform
	if !clipboardAvailable {
		return c.fallbackToVariable(input, "clipboard not available on this platform")
	}

	// Initialize clipboard - this is required for the library
	err := initClipboard()
	if err != nil {
		// Fallback: store in variable if clipboard unavailable
		return c.fallbackToVariable(input, fmt.Sprintf("clipboard initialization failed: %v", err))
	}

	// Copy to system clipboard
	err = writeToClipboard(input)
	if err != nil {
		return c.fallbackToVariable(input, fmt.Sprintf("failed to write to clipboard: %v", err))
	}

	// Provide success feedback with character count
	printer := c.createPrinter()
	printer.Success(fmt.Sprintf("Copied %d characters to clipboard", len(input)))
	return nil
}

// fallbackToVariable stores content in _clipboard variable when system clipboard is unavailable.
func (c *ClipCommand) fallbackToVariable(content, reason string) error {
	// Get variable service for fallback storage
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard and variable service unavailable: %s", reason)
	}

	// Store in _clipboard variable as fallback
	if err := variableService.SetSystemVariable("_clipboard", content); err != nil {
		return fmt.Errorf("failed to copy to clipboard and fallback storage failed: %s", reason)
	}

	// Inform user about fallback
	printer := c.createPrinter()
	printer.Warning(fmt.Sprintf("Failed to copy to clipboard: %s", reason))
	printer.Info(fmt.Sprintf("Stored %d characters in _clipboard variable", len(content)))
	return nil
}

// createPrinter creates a printer with theme service as style provider
func (c *ClipCommand) createPrinter() *output.Printer {
	// Try to get theme service as style provider
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// Fall back to plain style provider
		return output.NewPrinter(output.WithStyles(output.NewPlainStyleProvider()))
	}

	return output.NewPrinter(output.WithStyles(themeService))
}

func init() {
	if err := commands.GlobalRegistry.Register(&ClipCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register clip command: %v", err))
	}
}
