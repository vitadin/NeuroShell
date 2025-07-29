// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// GeminiClientNewCommand implements the \gemini-client-new command for creating Gemini clients.
// It provides automatic key resolution from #active_gemini_key system variable.
type GeminiClientNewCommand struct{}

// Name returns the command name "gemini-client-new" for registration and lookup.
func (c *GeminiClientNewCommand) Name() string {
	return "gemini-client-new"
}

// ParseMode returns ParseModeKeyValue for optional bracket parameter parsing.
func (c *GeminiClientNewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the gemini-client-new command does.
func (c *GeminiClientNewCommand) Description() string {
	return "Create new Gemini client with automatic key resolution"
}

// Usage returns the syntax and usage examples for the gemini-client-new command.
func (c *GeminiClientNewCommand) Usage() string {
	return "\\gemini-client-new[key=api_key] or \\gemini-client-new (uses active key)"
}

// HelpInfo returns structured help information for the gemini-client-new command.
func (c *GeminiClientNewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "key",
				Description: "Google API key for Gemini (optional if #active_gemini_key is set)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\gemini-client-new[key=AIzaSy...]",
				Description: "Create Gemini client with explicit API key",
			},
			{
				Command:     "\\gemini-client-new",
				Description: "Create Gemini client using #active_gemini_key",
			},
			{
				Command:     "\\gemini-client-new[key=${GOOGLE_API_KEY}]",
				Description: "Create Gemini client using variable interpolation",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_client_id",
				Description: "Contains the created Gemini client ID",
				Type:        "system_output",
				Example:     "_client_id = \"gemini:a1b2c3d4\"",
			},
			{
				Name:        "_output",
				Description: "Contains success message with client details",
				Type:        "system_output",
				Example:     "_output = \"Gemini client ready: gemini:a1b2c3d4\"",
			},
			{
				Name:        "#client_provider",
				Description: "Contains the provider name (always 'gemini')",
				Type:        "system_metadata",
				Example:     "#client_provider = \"gemini\"",
			},
			{
				Name:        "#client_configured",
				Description: "Contains client configuration status",
				Type:        "system_metadata",
				Example:     "#client_configured = \"true\"",
			},
		},
		Notes: []string{
			"Key resolution priority: 1) key parameter, 2) #active_gemini_key, 3) GOOGLE_API_KEY env var",
			"Use \\llm-api-activate[provider=gemini, key=...] to set #active_gemini_key",
			"Creates and caches Gemini clients for subsequent use",
			"Client configuration status stored in ${#client_configured}",
		},
	}
}

// Execute creates a new Gemini client with automatic key resolution.
func (c *GeminiClientNewCommand) Execute(args map[string]string, _ string) error {
	// Get variable service for key resolution
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Resolve API key with priority order
	apiKey, err := c.resolveAPIKey(args, variableService)
	if err != nil {
		return err
	}

	// Get client factory service
	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		return fmt.Errorf("client factory service not available: %w", err)
	}

	// Create Gemini client
	client, clientID, err := clientFactory.GetClientWithID("GMC", apiKey)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_client_id", clientID)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("Gemini client ready: %s", clientID))

	// Set metadata variables
	_ = variableService.SetSystemVariable("#client_provider", "gemini")
	_ = variableService.SetSystemVariable("#client_configured", fmt.Sprintf("%t", client.IsConfigured()))
	_ = variableService.SetSystemVariable("#client_cache_count", fmt.Sprintf("%d", clientFactory.GetCachedClientCount()))

	// Output success message
	fmt.Printf("Gemini client ready: %s (configured: %t)\n", clientID, client.IsConfigured())

	return nil
}

// resolveAPIKey resolves the API key using priority order:
// 1. User-provided key parameter
// 2. #active_gemini_key system variable
// 3. GOOGLE_API_KEY environment variable
func (c *GeminiClientNewCommand) resolveAPIKey(args map[string]string, variableService *services.VariableService) (string, error) {
	// Priority 1: User-provided key parameter
	if key := args["key"]; key != "" {
		return key, nil
	}

	// Priority 2: #active_gemini_key system variable
	if activeKey, err := variableService.Get("#active_gemini_key"); err == nil && activeKey != "" {
		return activeKey, nil
	}

	// Priority 3: GOOGLE_API_KEY environment variable
	if envKey := variableService.GetEnv("GOOGLE_API_KEY"); envKey != "" {
		return envKey, nil
	}

	// No key found
	return "", fmt.Errorf("no API key found. Use one of:\n" +
		"  1. \\gemini-client-new[key=your_api_key]\n" +
		"  2. \\llm-api-activate[provider=gemini, key=source.KEY_NAME] (sets #active_gemini_key)\n" +
		"  3. Set GOOGLE_API_KEY environment variable")
}

func init() {
	if err := commands.GlobalRegistry.Register(&GeminiClientNewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register gemini-client-new command: %v", err))
	}
}
