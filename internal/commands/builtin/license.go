package builtin

import (
	"fmt"
	"os"
	"path/filepath"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// LicenseCommand implements the \license command for displaying NeuroShell license information.
// It shows the full license text from LICENSE file and stores license info in system variables.
type LicenseCommand struct{}

// Name returns the command name "license" for registration and lookup.
func (c *LicenseCommand) Name() string {
	return "license"
}

// ParseMode returns ParseModeKeyValue for consistency with other builtin commands.
func (c *LicenseCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the license command does.
func (c *LicenseCommand) Description() string {
	return "Display NeuroShell license information and store license details in system variables"
}

// Usage returns the syntax and usage examples for the license command.
func (c *LicenseCommand) Usage() string {
	return "\\license"
}

// HelpInfo returns structured help information for the license command.
func (c *LicenseCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\license",
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for this command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\license",
				Description: "Display full license text and store license info in system variables",
			},
			{
				Command:     "\\license\\n\\echo License: ${#license_name}",
				Description: "Display license and then show the stored license name",
			},
			{
				Command:     "\\license\\n\\echo URL: ${#license_url}",
				Description: "Display license and then show the license URL",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#license_name",
				Description: "Name of the license (LGPL v3)",
				Type:        "license_info",
				Example:     "#license_name = \"GNU Lesser General Public License v3.0\"",
			},
			{
				Name:        "#license_short_name",
				Description: "Short name of the license",
				Type:        "license_info",
				Example:     "#license_short_name = \"LGPL-3.0\"",
			},
			{
				Name:        "#license_url",
				Description: "URL to the full license text",
				Type:        "license_info",
				Example:     "#license_url = \"https://www.gnu.org/licenses/lgpl-3.0.html\"",
			},
			{
				Name:        "#license_file_path",
				Description: "Path to the local LICENSE file",
				Type:        "license_info",
				Example:     "#license_file_path = \"/path/to/NeuroShell/LICENSE\"",
			},
		},
		Notes: []string{
			"Displays the complete license text from the LICENSE file",
			"Automatically stores license metadata in system variables with # prefix",
			"License information is accessible via ${#license_*} variables for scripting",
			"LGPL v3 allows commercial linking while keeping modifications open source",
			"If LICENSE file is not found, displays basic license information",
		},
	}
}

// Execute displays license information and stores license details in system variables.
func (c *LicenseCommand) Execute(_ map[string]string, _ string) error {
	// Get variable service for storing system variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Find the LICENSE file - try common locations
	var licensePath string
	possiblePaths := []string{
		"LICENSE",
		"LICENSE.txt",
		"../LICENSE",
		"../LICENSE.txt",
		"../../LICENSE",
		"../../LICENSE.txt",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			licensePath, _ = filepath.Abs(path)
			break
		}
	}

	// Store license metadata in system variables
	systemVars := map[string]string{
		"#license_name":       "GNU Lesser General Public License v3.0",
		"#license_short_name": "LGPL-3.0",
		"#license_url":        "https://www.gnu.org/licenses/lgpl-3.0.html",
		"#license_file_path":  licensePath,
	}

	// Set all system variables
	for varName, value := range systemVars {
		if err := variableService.SetSystemVariable(varName, value); err != nil {
			// Log error but continue - don't fail the command for storage issues
			printer := printing.NewDefaultPrinter()
			printer.Warning(fmt.Sprintf("Failed to set %s: %v", varName, err))
		}
	}

	// Display license header
	printer := printing.NewDefaultPrinter()
	printer.Info("NeuroShell - GNU Lesser General Public License v3.0")
	printer.Info("=" + fmt.Sprintf("%*s", 60, ""))
	printer.Info("")

	// Try to read and display the full license file
	if licensePath != "" {
		licenseContent, err := os.ReadFile(licensePath)
		if err == nil {
			printer.Info(string(licenseContent))
			return nil
		}
		printer.Warning(fmt.Sprintf("Could not read LICENSE file at %s: %v", licensePath, err))
		printer.Info("")
	}

	// Fallback: display basic license information
	printer.Info("NeuroShell is licensed under the GNU Lesser General Public License v3.0 (LGPL-3.0)")
	printer.Info("")
	printer.Info("This means:")
	printer.Info("• You can use NeuroShell in commercial applications")
	printer.Info("• You can modify NeuroShell, but modifications must be open source")
	printer.Info("• You can link to NeuroShell from proprietary code")
	printer.Info("• You must preserve copyright notices")
	printer.Info("")
	printer.Pair("Full license text", systemVars["#license_url"])
	printer.Pair("License file", licensePath)

	return nil
}

// IsReadOnly returns true as the license command doesn't modify system state.
func (c *LicenseCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&LicenseCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register license command: %v", err))
	}
}
