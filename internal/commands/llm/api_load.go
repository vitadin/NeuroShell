// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

// APILoadCommand implements the \llm-api-load command for loading and displaying API keys.
// It loads API keys from multiple sources (OS env, config .env, local .env) with source attribution.
type APILoadCommand struct{}

// Name returns the command name "llm-api-load" for registration and lookup.
func (c *APILoadCommand) Name() string {
	return "llm-api-load"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *APILoadCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-api-load command does.
func (c *APILoadCommand) Description() string {
	return "Load and display API-related variables from multiple sources with intelligent filtering and masking"
}

// Usage returns the syntax and usage examples for the llm-api-load command.
func (c *APILoadCommand) Usage() string {
	return "\\llm-api-load[provider=openai|anthropic|gemini|all]"
}

// HelpInfo returns structured help information for the llm-api-load command.
func (c *APILoadCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "Filter by provider (openai, anthropic, gemini, all)",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\llm-api-load",
				Description: "Load and display all API-related variables from all sources and providers",
			},
			{
				Command:     "\\llm-api-load[provider=openai]",
				Description: "Load and display only OpenAI-related API variables",
			},
			{
				Command:     "\\llm-api-load[provider=anthropic]",
				Description: "Load and display only Anthropic-related API variables",
			},
			{
				Command:     "\\llm-api-load[provider=gemini]",
				Description: "Load and display only Gemini-related API variables",
			},
		},
		Notes: []string{
			"Variables are collected from OS environment variables, config .env, and local .env files",
			"Only API-related variables are shown based on intelligent filtering:",
			"  • Variables containing provider names: openai, anthropic, gemini, google",
			"  • Variables containing API keywords: api, key, secret (case-insensitive)",
			"Keys are stored as source-prefixed variables (e.g., os.OPENAI_API_KEY)",
			"Active keys are marked with ACTIVE status when set via \\llm-api-activate",
			"Use \\silent \\llm-api-load to load keys without displaying them",
			"API keys are masked showing only first 3 and last 3 characters for security",
		},
	}
}

// Execute loads and displays API keys with source attribution and masked values.
func (c *APILoadCommand) Execute(args map[string]string, _ string) error {
	// Get configuration service
	configService, err := services.GetGlobalConfigurationService()
	if err != nil {
		return fmt.Errorf("configuration service not available: %w", err)
	}

	// Get variable service for storing keys and checking active status
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse provider filter
	provider := args["provider"]
	if provider == "" {
		provider = "all"
	}

	// Get all configuration variables from configuration service
	allKeys, err := configService.GetAllAPIKeys()
	if err != nil {
		return fmt.Errorf("failed to get configuration variables: %w", err)
	}

	// Get supported providers from configuration service
	supportedProviders := configService.GetSupportedProviders()

	// Filter for API-related variables using smart keyword detection
	var apiKeys []services.APIKeySource
	for _, key := range allKeys {
		isAPI, detectedProvider := stringprocessing.IsAPIRelated(key.OriginalName, supportedProviders)
		if isAPI {
			// Set the detected provider
			key.Provider = detectedProvider
			apiKeys = append(apiKeys, key)
		}
	}

	// Filter by provider if specified
	var filteredKeys []services.APIKeySource
	for _, key := range apiKeys {
		if provider == "all" || key.Provider == provider {
			filteredKeys = append(filteredKeys, key)
		}
	}

	// Store keys as source-prefixed variables
	for _, key := range filteredKeys {
		variableName := key.Source + "." + key.OriginalName
		err := variableService.Set(variableName, key.Value)
		if err != nil {
			return fmt.Errorf("failed to store variable %s: %w", variableName, err)
		}
	}

	// Display results
	c.displayAPIKeys(filteredKeys, variableService, provider)

	return nil
}

// displayAPIKeys formats and displays the API keys using markdown table rendered by glamour
func (c *APILoadCommand) displayAPIKeys(keys []services.APIKeySource, variableService *services.VariableService, providerFilter string) {
	if len(keys) == 0 {
		if providerFilter == "all" {
			fmt.Println("No API keys found in any source.")
		} else {
			fmt.Printf("No API keys found for provider '%s'.\n", providerFilter)
		}
		return
	}

	// Create markdown table
	markdown := c.createMarkdownTable(keys, variableService, providerFilter)

	// Get markdown service to render the table
	markdownService, err := services.GetGlobalMarkdownService()
	if err != nil {
		// Fallback to plain text if markdown service unavailable
		fmt.Print(markdown)
		return
	}

	// Render markdown table with glamour
	rendered, err := markdownService.RenderWithTheme(markdown)
	if err != nil {
		// Fallback to plain text if rendering fails
		fmt.Print(markdown)
		return
	}

	fmt.Print(rendered)
}

// createMarkdownTable creates a markdown table for API keys with proper formatting
func (c *APILoadCommand) createMarkdownTable(keys []services.APIKeySource, variableService *services.VariableService, providerFilter string) string {
	var result strings.Builder

	// Title
	title := "# LLM API Keys Found"
	if providerFilter != "all" {
		// Capitalize first letter manually
		providerTitle := strings.ToUpper(providerFilter[:1]) + providerFilter[1:]
		title = fmt.Sprintf("# LLM API Keys Found (%s only)", providerTitle)
	}
	result.WriteString(title)
	result.WriteString("\n\n")

	// Sort keys by variable name for deterministic output
	sort.Slice(keys, func(i, j int) bool {
		varI := keys[i].Source + "." + keys[i].OriginalName
		varJ := keys[j].Source + "." + keys[j].OriginalName
		return varI < varJ
	})

	// Create markdown table
	result.WriteString("| Variable Name | API Key | Status |\n")
	result.WriteString("|---------------|---------|--------|\n")

	// Add table rows
	for _, key := range keys {
		varName := key.Source + "." + key.OriginalName
		maskedKey := c.maskAPIKey(key.Value)
		status := c.getKeyStatus(key.Provider, varName, variableService)

		// Format status with appropriate styling
		var statusCell string
		switch status {
		case "ACTIVE":
			statusCell = "**🟢 " + status + "**" // Bold with green indicator for active status
		case "INACTIVE":
			statusCell = status // Plain text for inactive status
		default:
			statusCell = " " // Empty cell fallback
		}

		result.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", varName, maskedKey, statusCell))
	}

	// Add usage information
	result.WriteString("\n")
	if providerFilter == "all" {
		activeKeys := c.getActiveKeysInfo(variableService)
		if len(activeKeys) > 0 {
			result.WriteString(fmt.Sprintf("**Active keys:** %s\n\n", strings.Join(activeKeys, ", ")))
		}
	} else {
		activeKey := c.getActiveKeyForProvider(providerFilter, variableService)
		if activeKey != "" {
			result.WriteString(fmt.Sprintf("**Active key:** %s\n\n", activeKey))
		}
	}

	return result.String()
}

// maskAPIKey masks an API key showing only first 3 and last 3 characters
func (c *APILoadCommand) maskAPIKey(apiKey string) string {
	if len(apiKey) <= 6 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
}

// getKeyStatus checks if a key is currently active for its provider
func (c *APILoadCommand) getKeyStatus(provider, varName string, variableService *services.VariableService) string {
	activeVarName := "#active_" + provider + "_key"
	activeKeyValue, err := variableService.Get(activeVarName)
	if err != nil || activeKeyValue == "" {
		return "INACTIVE"
	}

	// Get the actual key value to compare
	currentKeyValue, err := variableService.Get(varName)
	if err != nil || currentKeyValue == "" {
		return "INACTIVE"
	}

	// If the active key value matches this key's value, it's active
	if activeKeyValue == currentKeyValue {
		return "ACTIVE"
	}
	return "INACTIVE"
}

// getActiveKeysInfo returns a list of active key metadata variable names
func (c *APILoadCommand) getActiveKeysInfo(variableService *services.VariableService) []string {
	// Get configuration service to access provider list
	configService, err := services.GetGlobalConfigurationService()
	if err != nil {
		return []string{} // Return empty slice if service unavailable
	}

	providers := configService.GetSupportedProviders()
	var activeKeys []string

	for _, provider := range providers {
		activeVarName := "#active_" + provider + "_key"
		activeKey, err := variableService.Get(activeVarName)
		if err == nil && activeKey != "" {
			activeKeys = append(activeKeys, "${"+activeVarName+"}")
		}
	}

	return activeKeys
}

// getActiveKeyForProvider returns the active key metadata variable for a specific provider
func (c *APILoadCommand) getActiveKeyForProvider(provider string, variableService *services.VariableService) string {
	activeVarName := "#active_" + provider + "_key"
	activeKey, err := variableService.Get(activeVarName)
	if err == nil && activeKey != "" {
		return "${" + activeVarName + "}"
	}
	return ""
}

func init() {
	if err := commands.GlobalRegistry.Register(&APILoadCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-api-load command: %v", err))
	}
}
