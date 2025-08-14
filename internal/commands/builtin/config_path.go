package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ConfigPathCommand implements the \config-path command for displaying configuration file paths.
// It shows the configuration directory and status of loaded .env and .neurorc files.
type ConfigPathCommand struct{}

// Name returns the command name "config-path" for registration and lookup.
func (c *ConfigPathCommand) Name() string {
	return "config-path"
}

// ParseMode returns ParseModeKeyValue for consistency with other builtin commands.
func (c *ConfigPathCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the config-path command does.
func (c *ConfigPathCommand) Description() string {
	return "Display configuration file paths and their loading status"
}

// Usage returns the syntax and usage examples for the config-path command.
func (c *ConfigPathCommand) Usage() string {
	return "\\config-path"
}

// HelpInfo returns structured help information for the config-path command.
func (c *ConfigPathCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\config-path",
		ParseMode:   c.ParseMode(),
		Options:     []neurotypes.HelpOption{}, // No options for this command
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\config-path",
				Description: "Show configuration directory and loaded file paths",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#config_dir",
				Description: "Configuration directory path",
				Type:        "path_info",
				Example:     "#config_dir = \"/home/user/.config/neuroshell\"",
			},
			{
				Name:        "#config_dir_exists",
				Description: "Whether configuration directory exists",
				Type:        "path_info",
				Example:     "#config_dir_exists = \"true\"",
			},
			{
				Name:        "#config_env_path",
				Description: "Config .env file path (if exists)",
				Type:        "path_info",
				Example:     "#config_env_path = \"/home/user/.config/neuroshell/.env\"",
			},
			{
				Name:        "#config_env_loaded",
				Description: "Whether config .env was loaded",
				Type:        "path_info",
				Example:     "#config_env_loaded = \"true\"",
			},
			{
				Name:        "#local_env_path",
				Description: "Local .env file path (if exists)",
				Type:        "path_info",
				Example:     "#local_env_path = \"/current/dir/.env\"",
			},
			{
				Name:        "#local_env_loaded",
				Description: "Whether local .env was loaded",
				Type:        "path_info",
				Example:     "#local_env_loaded = \"false\"",
			},
			{
				Name:        "#neurorc_path",
				Description: "Executed .neurorc file path",
				Type:        "path_info",
				Example:     "#neurorc_path = \"/home/user/.neurorc\"",
			},
			{
				Name:        "#neurorc_executed",
				Description: "Whether .neurorc was executed",
				Type:        "path_info",
				Example:     "#neurorc_executed = \"true\"",
			},
		},
		Notes: []string{
			"Displays configuration directory and loaded file paths",
			"Shows loading status for each configuration file type",
			"All path information is stored in system variables with # prefix",
			"Paths are determined based on current test mode and environment",
			"Config .env is loaded from the configuration directory",
			"Local .env is loaded from the current working directory",
			".neurorc is loaded from home or current directory",
		},
	}
}

// Execute displays configuration file paths and their loading status.
func (c *ConfigPathCommand) Execute(_ map[string]string, _ string) error {
	// Get configuration service to retrieve path information
	configService, err := services.GetGlobalRegistry().GetService("configuration")
	if err != nil {
		return fmt.Errorf("configuration service not available: %w", err)
	}

	cs := configService.(*services.ConfigurationService)

	// Get variable service to read/write system variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get configuration paths from configuration service
	paths, err := cs.GetConfigurationPaths()
	if err != nil {
		return fmt.Errorf("failed to get configuration paths: %w", err)
	}

	// Get .neurorc information from system variables (set by main.go)
	neuroRCPath, _ := variableService.Get("#neurorc_path")
	neuroRCExecutedStr, _ := variableService.Get("#neurorc_executed")
	neuroRCExecuted := neuroRCExecutedStr == "true"

	// Store all path information in system variables for scripting access
	if paths.ConfigDir != "" {
		if err := variableService.SetSystemVariable("#config_dir", paths.ConfigDir); err != nil {
			fmt.Printf("Warning: failed to set #config_dir: %v\n", err)
		}
	}
	if err := variableService.SetSystemVariable("#config_dir_exists", fmt.Sprintf("%v", paths.ConfigDirExists)); err != nil {
		fmt.Printf("Warning: failed to set #config_dir_exists: %v\n", err)
	}

	if paths.ConfigEnvPath != "" {
		if err := variableService.SetSystemVariable("#config_env_path", paths.ConfigEnvPath); err != nil {
			fmt.Printf("Warning: failed to set #config_env_path: %v\n", err)
		}
	}
	if err := variableService.SetSystemVariable("#config_env_loaded", fmt.Sprintf("%v", paths.ConfigEnvLoaded)); err != nil {
		fmt.Printf("Warning: failed to set #config_env_loaded: %v\n", err)
	}

	if paths.LocalEnvPath != "" {
		if err := variableService.SetSystemVariable("#local_env_path", paths.LocalEnvPath); err != nil {
			fmt.Printf("Warning: failed to set #local_env_path: %v\n", err)
		}
	}
	if err := variableService.SetSystemVariable("#local_env_loaded", fmt.Sprintf("%v", paths.LocalEnvLoaded)); err != nil {
		fmt.Printf("Warning: failed to set #local_env_loaded: %v\n", err)
	}

	// Display configuration directory
	if paths.ConfigDir != "" {
		status := "not found"
		if paths.ConfigDirExists {
			status = "exists"
		}
		fmt.Printf("Config Directory: %s (%s)\n", paths.ConfigDir, status)
	}

	// Display config .env status
	if paths.ConfigEnvPath != "" {
		status := "not found"
		if paths.ConfigEnvLoaded {
			status = "loaded"
		}
		fmt.Printf("Config .env: %s (%s)\n", paths.ConfigEnvPath, status)
	}

	// Display local .env status
	if paths.LocalEnvPath != "" {
		status := "not found"
		if paths.LocalEnvLoaded {
			status = "loaded"
		}
		fmt.Printf("Local .env: %s (%s)\n", paths.LocalEnvPath, status)
	}

	// Display .neurorc status (from system variables)
	if neuroRCPath != "" {
		status := "not executed"
		if neuroRCExecuted {
			status = "executed"
		}
		fmt.Printf(".neurorc: %s (%s)\n", neuroRCPath, status)
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ConfigPathCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register config-path command: %v", err))
	}
}
