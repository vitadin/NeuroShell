// Package printing provides common printing utilities for command implementations.
package printing

import (
	"neuroshell/internal/output"
	"neuroshell/internal/services"
)

// NewDefaultPrinter creates a printer with theme service if available,
// falling back to plain style if theme service is unavailable.
// This is a convenience function commonly used by commands.
func NewDefaultPrinter() *output.Printer {
	// Try to get theme service as style provider
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// Fall back to plain style provider
		return output.NewPrinter(output.WithStyles(output.NewPlainStyleProvider()))
	}

	return output.NewPrinter(output.WithStyles(themeService))
}
