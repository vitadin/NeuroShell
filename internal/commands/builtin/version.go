package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/internal/version"
	"neuroshell/pkg/neurotypes"
)

// VersionCommand implements the \version command for displaying NeuroShell version information.
// It shows formatted version info and stores detailed components in system variables.
type VersionCommand struct{}

// Name returns the command name "version" for registration and lookup.
func (c *VersionCommand) Name() string {
	return "version"
}

// ParseMode returns ParseModeKeyValue for consistency with other builtin commands.
func (c *VersionCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the version command does.
func (c *VersionCommand) Description() string {
	return "Show NeuroShell version information and store details in system variables"
}

// Usage returns the syntax and usage examples for the version command.
func (c *VersionCommand) Usage() string {
	return "\\version"
}

// HelpInfo returns structured help information for the version command.
func (c *VersionCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\version",
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for this command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\version",
				Description: "Show version information and store details in system variables",
			},
			{
				Command:     "\\version\\n\\echo Version: ${#version}",
				Description: "Show version and then display the stored version string",
			},
			{
				Command:     "\\version\\n\\echo Codename: ${#version_codename}",
				Description: "Show version and then display the stored codename",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#version",
				Description: "Full version string with build metadata",
				Type:        "version_info",
				Example:     "#version = \"0.2.0+530.a455fa8\"",
			},
			{
				Name:        "#version_base",
				Description: "Base semantic version (major.minor.patch)",
				Type:        "version_info",
				Example:     "#version_base = \"0.2.0\"",
			},
			{
				Name:        "#version_codename",
				Description: "Version codename based on neural complexity progression",
				Type:        "version_info",
				Example:     "#version_codename = \"Planaria\"",
			},
			{
				Name:        "#version_commit",
				Description: "Git commit hash (short form)",
				Type:        "version_info",
				Example:     "#version_commit = \"a455fa8\"",
			},
			{
				Name:        "#version_commit_count",
				Description: "Number of commits since base version tag",
				Type:        "version_info",
				Example:     "#version_commit_count = \"530\"",
			},
			{
				Name:        "#version_build_date",
				Description: "Date when the binary was built",
				Type:        "version_info",
				Example:     "#version_build_date = \"2025-07-29\"",
			},
			{
				Name:        "#version_build_metadata",
				Description: "Build metadata from semantic version",
				Type:        "version_info",
				Example:     "#version_build_metadata = \"530.a455fa8\"",
			},
			{
				Name:        "#version_go_version",
				Description: "Go version used to build the binary",
				Type:        "version_info",
				Example:     "#version_go_version = \"go1.24.4\"",
			},
			{
				Name:        "#version_platform",
				Description: "Target platform (OS/architecture)",
				Type:        "version_info",
				Example:     "#version_platform = \"darwin/arm64\"",
			},
		},
		Notes: []string{
			"Displays formatted version information to console",
			"Automatically stores detailed version components in system variables with # prefix",
			"All version details are accessible via ${#version_*} variables for scripting",
			"Codename follows neural complexity progression (Hydra → Planaria → ... → Sapiens → Synthia)",
			"Build metadata includes commit count and hash for tracking incremental changes",
		},
	}
}

// Execute displays version information and stores detailed components in system variables.
func (c *VersionCommand) Execute(_ map[string]string, _ string) error {
	// Get variable service for storing system variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get version information
	info, err := version.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get version information: %w", err)
	}

	// Store detailed version components in system variables
	systemVars := map[string]string{
		"#version":                version.GetVersion(),
		"#version_base":           version.GetBaseVersion(),
		"#version_codename":       version.GetCodename(),
		"#version_commit":         info.GitCommit,
		"#version_build_date":     info.BuildDate,
		"#version_build_metadata": version.GetBuildMetadata(),
		"#version_go_version":     info.GoVersion,
		"#version_platform":       info.Platform,
	}

	// Add commit count as string
	commitCount := version.GetCommitCount()
	systemVars["#version_commit_count"] = fmt.Sprintf("%d", commitCount)

	// Set all system variables
	for varName, value := range systemVars {
		if err := variableService.SetSystemVariable(varName, value); err != nil {
			// Log error but continue - don't fail the command for storage issues
			fmt.Printf("Warning: failed to set %s: %v\n", varName, err)
		}
	}

	// Display formatted version information
	formattedVersion := version.GetFormattedVersion()
	fmt.Println(formattedVersion)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&VersionCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register version command: %v", err))
	}
}
