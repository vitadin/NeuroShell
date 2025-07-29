// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ClientGetCommand implements the \llm-client-get command for getting/creating LLM clients.
// It provides LLM client creation and caching functionality for the NeuroShell environment.
type ClientGetCommand struct{}

// Name returns the command name "llm-client-get" for registration and lookup.
func (c *ClientGetCommand) Name() string {
	return "llm-client-get"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *ClientGetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-client-get command does.
func (c *ClientGetCommand) Description() string {
	return "Get or create LLM client for provider catalog ID"
}

// Usage returns the syntax and usage examples for the llm-client-get command.
func (c *ClientGetCommand) Usage() string {
	return "\\llm-client-get[key=api_key, provider_catalog_id=OAC|OAR|ORC|MSC|ANC|GMC] or \\llm-client-get (uses env vars)"
}

// HelpInfo returns structured help information for the llm-client-get command.
func (c *ClientGetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider_catalog_id",
				Description: "LLM provider catalog ID (OAC, OAR, ORC, MSC, ANC, GMC)",
				Required:    false,
				Type:        "string",
				Default:     "OAR",
			},
			{
				Name:        "key",
				Description: "API key for the provider (optional if environment variable is set)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\llm-client-get[provider_catalog_id=OAR, key=sk-...]",
				Description: "Get OpenAI reasoning client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider_catalog_id=OAC, key=sk-...]",
				Description: "Get OpenAI chat client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider_catalog_id=ORC, key=sk-or-...]",
				Description: "Get OpenRouter client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider_catalog_id=ANC, key=sk-ant-...]",
				Description: "Get Anthropic client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider_catalog_id=GMC, key=AIzaSy...]",
				Description: "Get Gemini client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[key=${OPENAI_API_KEY}]",
				Description: "Get OpenAI reasoning client using variable interpolation",
			},
			{
				Command:     "\\llm-client-get",
				Description: "Get OpenAI reasoning client using default catalog ID",
			},
		},
		Notes: []string{
			"Creates and caches LLM clients for subsequent use",
			"Client ID stored in ${_client_id} system variable",
			"Provider catalog IDs: OAC (OpenAI Chat), OAR (OpenAI Reasoning), ORC (OpenRouter), MSC (Moonshot), ANC (Anthropic), GMC (Gemini)",
			"API key resolution: parameter → environment variable",
			"Environment variables: OPENAI_API_KEY, OPENROUTER_API_KEY, MOONSHOT_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY",
			"Client configuration status stored in ${#client_configured}",
			"Provider catalog ID stored in ${#client_provider_catalog_id}",
			"Cached client count stored in ${#client_cache_count}",
		},
	}
}

// Execute creates or retrieves an LLM client for the specified provider catalog ID.
// API key can be provided explicitly or via environment variables.
func (c *ClientGetCommand) Execute(args map[string]string, _ string) error {
	// Determine provider catalog ID (from args or default to OAR)
	providerCatalogID := args["provider_catalog_id"]
	if providerCatalogID == "" {
		providerCatalogID = "OAR" // Default to OpenAI reasoning
	}

	// Get variable service for API key resolution
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get provider catalog service to validate catalog ID
	providerCatalogService, err := services.GetGlobalProviderCatalogService()
	if err != nil {
		return fmt.Errorf("provider catalog service not available: %w", err)
	}

	// Get valid catalog IDs dynamically
	validCatalogIDs, err := providerCatalogService.GetValidCatalogIDs()
	if err != nil {
		return fmt.Errorf("failed to get valid catalog IDs: %w", err)
	}

	// Validate provider catalog ID
	isValid := false
	for _, valid := range validCatalogIDs {
		if providerCatalogID == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("unsupported provider catalog ID '%s'. Supported IDs: %s", providerCatalogID, strings.Join(validCatalogIDs, ", "))
	}

	// Delegate to specialized commands for better key resolution
	switch providerCatalogID {
	case "OAC", "OAR": // OpenAI endpoints
		return c.delegateToOpenAIClientNew(args)
	case "ANC": // Anthropic
		return c.delegateToAnthropicClientNew(args)
	case "GMC": // Gemini
		return c.delegateToGeminiClientNew(args)
	case "ORC", "MSC": // OpenRouter, Moonshot - use direct factory
		return c.createDirectClient(providerCatalogID, args, variableService)
	default:
		return fmt.Errorf("unsupported provider catalog ID '%s'", providerCatalogID)
	}
}

// delegateToOpenAIClientNew handles OpenAI catalog IDs by delegating to the specialized command.
// This leverages the robust key resolution system and reasoning model support in openai-client-new.
func (c *ClientGetCommand) delegateToOpenAIClientNew(args map[string]string) error {
	// Create openai-client-new command and execute it directly
	openaiClientNewCmd := &OpenAIClientNewCommand{}

	// Prepare args for the delegated command
	delegateArgs := make(map[string]string)
	if key, exists := args["key"]; exists && key != "" {
		delegateArgs["key"] = key
	}

	// Execute the openai-client-new command directly
	return openaiClientNewCmd.Execute(delegateArgs, "")
}

// delegateToAnthropicClientNew handles Anthropic catalog ID by delegating to the specialized command.
// This leverages the robust key resolution system in anthropic-client-new.
func (c *ClientGetCommand) delegateToAnthropicClientNew(args map[string]string) error {
	// Create anthropic-client-new command and execute it directly
	anthropicClientNewCmd := &AnthropicClientNewCommand{}

	// Prepare args for the delegated command
	delegateArgs := make(map[string]string)
	if key, exists := args["key"]; exists && key != "" {
		delegateArgs["key"] = key
	}

	// Execute the anthropic-client-new command directly
	return anthropicClientNewCmd.Execute(delegateArgs, "")
}

// delegateToGeminiClientNew handles Gemini catalog ID by delegating to the specialized command.
// This leverages the robust key resolution system in gemini-client-new.
func (c *ClientGetCommand) delegateToGeminiClientNew(args map[string]string) error {
	// Create gemini-client-new command and execute it directly
	geminiClientNewCmd := &GeminiClientNewCommand{}

	// Prepare args for the delegated command
	delegateArgs := make(map[string]string)
	if key, exists := args["key"]; exists && key != "" {
		delegateArgs["key"] = key
	}

	// Execute the gemini-client-new command directly
	return geminiClientNewCmd.Execute(delegateArgs, "")
}

// createDirectClient creates clients directly using the factory for simple providers.
// Used for OpenRouter and Moonshot which don't have specialized commands yet.
func (c *ClientGetCommand) createDirectClient(providerCatalogID string, args map[string]string, variableService *services.VariableService) error {
	// Get API key - first from args, then from environment variable
	apiKey := args["key"]
	if apiKey == "" {
		// Map catalog ID to environment variable
		var envVarName string
		switch providerCatalogID {
		case "ORC": // OpenRouter
			envVarName = "OPENROUTER_API_KEY"
		case "MSC": // Moonshot
			envVarName = "MOONSHOT_API_KEY"
		}

		apiKey = variableService.GetEnv(envVarName)
		if apiKey == "" {
			return fmt.Errorf("API key not found. Please provide key parameter or set %s environment variable", envVarName)
		}
	}

	// Get client factory service
	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		return fmt.Errorf("client factory service not available: %w", err)
	}

	// Get or create client with ID from service
	client, clientID, err := clientFactory.GetClientWithID(providerCatalogID, apiKey)
	if err != nil {
		return fmt.Errorf("failed to get client for provider catalog ID %s: %w", providerCatalogID, err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_client_id", clientID)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("LLM client ready: %s (catalog ID: %s, configured: %t)", clientID, providerCatalogID, client.IsConfigured()))

	// Set metadata variables
	_ = variableService.SetSystemVariable("#client_provider_catalog_id", providerCatalogID)
	_ = variableService.SetSystemVariable("#client_configured", fmt.Sprintf("%t", client.IsConfigured()))
	_ = variableService.SetSystemVariable("#client_cache_count", fmt.Sprintf("%d", clientFactory.GetCachedClientCount()))

	// Output success message
	fmt.Printf("LLM client ready: %s (catalog ID: %s, configured: %t)\n", clientID, providerCatalogID, client.IsConfigured())

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ClientGetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-client-get command: %v", err))
	}
}
