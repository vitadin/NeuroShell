// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// OpenAIClientNewCommand implements the \openai-client-new command for creating OpenAI clients.
// It provides automatic key resolution and supports both regular and reasoning models.
type OpenAIClientNewCommand struct{}

// Name returns the command name "openai-client-new" for registration and lookup.
func (c *OpenAIClientNewCommand) Name() string {
	return "openai-client-new"
}

// ParseMode returns ParseModeKeyValue for optional bracket parameter parsing.
func (c *OpenAIClientNewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the openai-client-new command does.
func (c *OpenAIClientNewCommand) Description() string {
	return "Create new OpenAI client with automatic key resolution and reasoning model support"
}

// Usage returns the syntax and usage examples for the openai-client-new command.
func (c *OpenAIClientNewCommand) Usage() string {
	return "\\openai-client-new[key=api_key, client_type=OAC|OAR] or \\openai-client-new (uses active key)"
}

// HelpInfo returns structured help information for the openai-client-new command.
func (c *OpenAIClientNewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "key",
				Description: "OpenAI API key (optional if #active_openai_key is set)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "client_type",
				Description: "OpenAI client type: OAC (chat only) or OAR (reasoning/dual-mode)",
				Required:    false,
				Type:        "string",
				Default:     "OAR",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\openai-client-new[key=sk-proj-...]",
				Description: "Create OpenAI reasoning client with explicit API key",
			},
			{
				Command:     "\\openai-client-new[client_type=OAC, key=sk-proj-...]",
				Description: "Create OpenAI chat-only client with explicit API key",
			},
			{
				Command:     "\\openai-client-new[client_type=OAR]",
				Description: "Create OpenAI reasoning client using #active_openai_key",
			},
			{
				Command:     "\\openai-client-new[key=${OPENAI_API_KEY}]",
				Description: "Create OpenAI reasoning client using variable interpolation",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_client_id",
				Description: "Contains the created OpenAI client ID",
				Type:        "system_output",
				Example:     "_client_id = \"openai:a1b2c3d4\"",
			},
			{
				Name:        "_output",
				Description: "Contains success message with client details",
				Type:        "system_output",
				Example:     "_output = \"OpenAI client ready: openai:a1b2c3d4\"",
			},
			{
				Name:        "#client_provider",
				Description: "Contains the provider name (always 'openai')",
				Type:        "system_metadata",
				Example:     "#client_provider = \"openai\"",
			},
			{
				Name:        "#client_configured",
				Description: "Contains client configuration status",
				Type:        "system_metadata",
				Example:     "#client_configured = \"true\"",
			},
			{
				Name:        "#client_reasoning_support",
				Description: "Indicates if client supports reasoning models",
				Type:        "system_metadata",
				Example:     "#client_reasoning_support = \"true\"",
			},
		},
		Notes: []string{
			"Key resolution priority: 1) key parameter, 2) #active_openai_key, 3) OPENAI_API_KEY env var",
			"Use \\llm-api-activate[provider=openai, key=...] to set #active_openai_key",
			"Client types: OAC (chat-only /chat/completions), OAR (dual-mode /chat/completions + /responses)",
			"OAR clients automatically route to appropriate endpoint based on model parameters",
			"Default client_type is OAR for backward compatibility",
			"Use OAC for dedicated chat-only testing, OAR for reasoning or dual-mode testing",
		},
	}
}

// Execute creates a new OpenAI client with automatic key resolution.
func (c *OpenAIClientNewCommand) Execute(args map[string]string, _ string) error {
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

	// Determine client type (default to OAR for backward compatibility)
	clientType := args["client_type"]
	if clientType == "" {
		clientType = "OAR"
	}

	// Validate client type
	if clientType != "OAC" && clientType != "OAR" {
		return fmt.Errorf("invalid client_type '%s'. Must be 'OAC' (chat-only) or 'OAR' (reasoning/dual-mode)", clientType)
	}

	// Create OpenAI client with specified type
	client, clientID, err := clientFactory.GetClientWithID(clientType, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_client_id", clientID)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("OpenAI client ready: %s (type: %s)", clientID, clientType))

	// Set metadata variables
	_ = variableService.SetSystemVariable("#client_provider", "openai")
	_ = variableService.SetSystemVariable("#client_configured", fmt.Sprintf("%t", client.IsConfigured()))
	_ = variableService.SetSystemVariable("#client_type", clientType)
	reasoningSupported := clientType == "OAR"
	_ = variableService.SetSystemVariable("#client_reasoning_support", fmt.Sprintf("%t", reasoningSupported))
	_ = variableService.SetSystemVariable("#client_cache_count", fmt.Sprintf("%d", clientFactory.GetCachedClientCount()))

	// Output success message
	clientTypeDesc := "chat-only"
	if clientType == "OAR" {
		clientTypeDesc = "reasoning/dual-mode"
	}
	fmt.Printf("OpenAI client ready: %s (type: %s - %s, configured: %t)\n", clientID, clientType, clientTypeDesc, client.IsConfigured())

	return nil
}

// resolveAPIKey resolves the API key using priority order:
// 1. User-provided key parameter
// 2. #active_openai_key system variable
// 3. OPENAI_API_KEY environment variable
func (c *OpenAIClientNewCommand) resolveAPIKey(args map[string]string, variableService *services.VariableService) (string, error) {
	// Priority 1: User-provided key parameter
	if key := args["key"]; key != "" {
		return key, nil
	}

	// Priority 2: #active_openai_key system variable
	if activeKey, err := variableService.Get("#active_openai_key"); err == nil && activeKey != "" {
		return activeKey, nil
	}

	// Priority 3: OPENAI_API_KEY environment variable
	if envKey := variableService.GetEnv("OPENAI_API_KEY"); envKey != "" {
		return envKey, nil
	}

	// No key found
	return "", fmt.Errorf("no API key found. Use one of:\n" +
		"  1. \\openai-client-new[key=your_api_key]\n" +
		"  2. \\llm-api-activate[provider=openai, key=source.KEY_NAME] (sets #active_openai_key)\n" +
		"  3. Set OPENAI_API_KEY environment variable")
}

func init() {
	if err := commands.GlobalRegistry.Register(&OpenAIClientNewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register openai-client-new command: %v", err))
	}
}
