// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// AnthropicClientNewCommand implements the \anthropic-client-new command for creating Anthropic clients.
// It provides automatic key resolution and supports extended thinking mode.
type AnthropicClientNewCommand struct{}

// Name returns the command name "anthropic-client-new" for registration and lookup.
func (c *AnthropicClientNewCommand) Name() string {
	return "anthropic-client-new"
}

// ParseMode returns ParseModeKeyValue for optional bracket parameter parsing.
func (c *AnthropicClientNewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the anthropic-client-new command does.
func (c *AnthropicClientNewCommand) Description() string {
	return "Create new Anthropic client with automatic key resolution and extended thinking support"
}

// Usage returns the syntax and usage examples for the anthropic-client-new command.
func (c *AnthropicClientNewCommand) Usage() string {
	return "\\anthropic-client-new[key=api_key] or \\anthropic-client-new (uses active key)"
}

// HelpInfo returns structured help information for the anthropic-client-new command.
func (c *AnthropicClientNewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "key",
				Description: "Anthropic API key (optional if #active_anthropic_key is set)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\anthropic-client-new[key=sk-ant-api03-...]",
				Description: "Create Anthropic client with explicit API key",
			},
			{
				Command:     "\\anthropic-client-new",
				Description: "Create Anthropic client using #active_anthropic_key",
			},
			{
				Command:     "\\anthropic-client-new[key=${ANTHROPIC_API_KEY}]",
				Description: "Create Anthropic client using variable interpolation",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_client_id",
				Description: "Contains the created Anthropic client ID",
				Type:        "system_output",
				Example:     "_client_id = \"anthropic:a1b2c3d4\"",
			},
			{
				Name:        "_output",
				Description: "Contains success message with client details",
				Type:        "system_output",
				Example:     "_output = \"Anthropic client ready: anthropic:a1b2c3d4\"",
			},
			{
				Name:        "#client_provider",
				Description: "Contains the provider name (always 'anthropic')",
				Type:        "system_metadata",
				Example:     "#client_provider = \"anthropic\"",
			},
			{
				Name:        "#client_configured",
				Description: "Contains client configuration status",
				Type:        "system_metadata",
				Example:     "#client_configured = \"true\"",
			},
			{
				Name:        "#client_thinking_support",
				Description: "Indicates if client supports extended thinking mode",
				Type:        "system_metadata",
				Example:     "#client_thinking_support = \"true\"",
			},
		},
		Notes: []string{
			"Key resolution priority: 1) key parameter, 2) #active_anthropic_key, 3) ANTHROPIC_API_KEY env var",
			"Use \\llm-api-activate[provider=anthropic, key=...] to set #active_anthropic_key",
			"Supports extended thinking with thinking_budget parameter in model configurations",
			"Extended thinking enables Claude's step-by-step reasoning capabilities",
			"Client handles both regular text and thinking/redacted_thinking blocks transparently",
		},
	}
}

// Execute creates a new Anthropic client with automatic key resolution.
func (c *AnthropicClientNewCommand) Execute(args map[string]string, _ string) error {
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

	// Create Anthropic client
	client, clientID, err := clientFactory.GetClientWithID("ANC", apiKey)
	if err != nil {
		return fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_client_id", clientID)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("Anthropic client ready: %s", clientID))

	// Set metadata variables
	_ = variableService.SetSystemVariable("#client_provider", "anthropic")
	_ = variableService.SetSystemVariable("#client_configured", fmt.Sprintf("%t", client.IsConfigured()))
	_ = variableService.SetSystemVariable("#client_thinking_support", "true") // Anthropic supports extended thinking
	_ = variableService.SetSystemVariable("#client_cache_count", fmt.Sprintf("%d", clientFactory.GetCachedClientCount()))

	// Output success message
	fmt.Printf("Anthropic client ready: %s (configured: %t, extended thinking: supported)\n", clientID, client.IsConfigured())

	return nil
}

// resolveAPIKey resolves the API key using priority order:
// 1. User-provided key parameter
// 2. #active_anthropic_key system variable
// 3. ANTHROPIC_API_KEY environment variable
func (c *AnthropicClientNewCommand) resolveAPIKey(args map[string]string, variableService *services.VariableService) (string, error) {
	// Priority 1: User-provided key parameter
	if key := args["key"]; key != "" {
		return key, nil
	}

	// Priority 2: #active_anthropic_key system variable
	if activeKey, err := variableService.Get("#active_anthropic_key"); err == nil && activeKey != "" {
		return activeKey, nil
	}

	// Priority 3: ANTHROPIC_API_KEY environment variable
	if envKey := variableService.GetEnv("ANTHROPIC_API_KEY"); envKey != "" {
		return envKey, nil
	}

	// No key found
	return "", fmt.Errorf("no API key found. Use one of:\n" +
		"  1. \\anthropic-client-new[key=your_api_key]\n" +
		"  2. \\llm-api-activate[provider=anthropic, key=source.KEY_NAME] (sets #active_anthropic_key)\n" +
		"  3. Set ANTHROPIC_API_KEY environment variable")
}

// IsReadOnly returns false as the llm command modifies system state.
func (c *AnthropicClientNewCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GlobalRegistry.Register(&AnthropicClientNewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register anthropic-client-new command: %v", err))
	}
}
