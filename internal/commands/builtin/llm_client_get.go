package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// LLMClientGetCommand implements the \llm-client-get command for getting/creating LLM clients.
// It provides LLM client creation and caching functionality for the NeuroShell environment.
type LLMClientGetCommand struct{}

// Name returns the command name "llm-client-get" for registration and lookup.
func (c *LLMClientGetCommand) Name() string {
	return "llm-client-get"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *LLMClientGetCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-client-get command does.
func (c *LLMClientGetCommand) Description() string {
	return "Get or create LLM client for provider"
}

// Usage returns the syntax and usage examples for the llm-client-get command.
func (c *LLMClientGetCommand) Usage() string {
	return "\\llm-client-get[key=api_key, provider=openai] or \\llm-client-get (uses env vars)"
}

// HelpInfo returns structured help information for the llm-client-get command.
func (c *LLMClientGetCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "LLM provider name (openai, anthropic)",
				Required:    false,
				Type:        "string",
				Default:     "openai",
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
				Command:     "\\llm-client-get[provider=openai, key=sk-...]",
				Description: "Get OpenAI client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[key=${OPENAI_API_KEY}]",
				Description: "Get OpenAI client using variable interpolation",
			},
			{
				Command:     "\\llm-client-get",
				Description: "Get OpenAI client using OPENAI_API_KEY environment variable",
			},
		},
		Notes: []string{
			"Creates and caches LLM clients for subsequent use",
			"Client ID stored in ${_client_id} system variable",
			"API key can be provided explicitly or via environment variables:",
			"  - OPENAI_API_KEY for OpenAI provider",
			"  - ANTHROPIC_API_KEY for Anthropic provider (when supported)",
			"Client configuration status stored in ${#client_configured}",
			"Provider name stored in ${#client_provider}",
			"Cached client count stored in ${#client_cache_count}",
		},
	}
}

// Execute creates or retrieves an LLM client for the specified provider.
// API key can be provided explicitly or via environment variables.
func (c *LLMClientGetCommand) Execute(args map[string]string, _ string) error {
	// Determine provider (from args or default to openai)
	provider := args["provider"]
	if provider == "" {
		provider = "openai" // Default to openai provider
	}

	// Get variable service first for API key resolution
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get API key - first from args, then from environment variable
	apiKey := args["key"]
	if apiKey == "" {
		// Get API key from environment variable via VariableService
		// This follows proper architecture: Command -> Service -> Context -> OS
		switch provider {
		case "openai":
			apiKey = variableService.GetEnv("OPENAI_API_KEY")
		case "anthropic":
			apiKey = variableService.GetEnv("ANTHROPIC_API_KEY")
		default:
			return fmt.Errorf("unsupported provider '%s'. Supported providers: openai, anthropic", provider)
		}

		// If still no API key found, return error
		if apiKey == "" {
			envVarName := "OPENAI_API_KEY"
			if provider == "anthropic" {
				envVarName = "ANTHROPIC_API_KEY"
			}
			return fmt.Errorf("API key not found. Please provide key parameter or set %s environment variable. Usage: %s", envVarName, c.Usage())
		}
	}

	// Get client factory service
	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		return fmt.Errorf("client factory service not available: %w", err)
	}

	// Get or create client with ID from service (service handles ID generation)
	client, clientID, err := clientFactory.GetClientWithID(provider, apiKey)
	if err != nil {
		return fmt.Errorf("failed to get client for provider %s: %w", provider, err)
	}

	// Set result variables (graceful degradation - we already have the service)
	_ = variableService.SetSystemVariable("_client_id", clientID)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("LLM client ready: %s", clientID))

	// Set metadata variables
	_ = variableService.SetSystemVariable("#client_provider", provider)
	_ = variableService.SetSystemVariable("#client_configured", fmt.Sprintf("%t", client.IsConfigured()))
	_ = variableService.SetSystemVariable("#client_cache_count", fmt.Sprintf("%d", clientFactory.GetCachedClientCount()))

	// Output success message
	fmt.Printf("LLM client ready: %s (configured: %t)\n", clientID, client.IsConfigured())

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&LLMClientGetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-client-get command: %v", err))
	}
}
