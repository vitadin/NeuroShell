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
	return "Get or create LLM client for provider"
}

// Usage returns the syntax and usage examples for the llm-client-get command.
func (c *ClientGetCommand) Usage() string {
	return "\\llm-client-get[key=api_key, provider=openai|openrouter|moonshot|anthropic|gemini] or \\llm-client-get (uses env vars)"
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
				Name:        "provider",
				Description: "LLM provider name (openai, openrouter, moonshot, anthropic, gemini)",
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
				Command:     "\\llm-client-get[provider=openrouter, key=sk-or-...]",
				Description: "Get OpenRouter client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider=moonshot, key=sk-...]",
				Description: "Get Moonshot client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider=anthropic, key=sk-ant-...]",
				Description: "Get Anthropic client with explicit API key",
			},
			{
				Command:     "\\llm-client-get[provider=gemini, key=AIzaSy...]",
				Description: "Get Gemini client with explicit API key",
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
			"  - OPENROUTER_API_KEY for OpenRouter provider",
			"  - MOONSHOT_API_KEY for Moonshot provider",
			"  - ANTHROPIC_API_KEY for Anthropic provider",
			"  - GOOGLE_API_KEY for Gemini provider",
			"OpenRouter configuration is automatically set for NeuroShell",
			"Client configuration status stored in ${#client_configured}",
			"Provider name stored in ${#client_provider}",
			"Cached client count stored in ${#client_cache_count}",
		},
	}
}

// Execute creates or retrieves an LLM client for the specified provider.
// API key can be provided explicitly or via environment variables.
func (c *ClientGetCommand) Execute(args map[string]string, _ string) error {
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

	// Get configuration service for provider validation
	configService, err := services.GetGlobalConfigurationService()
	if err != nil {
		return fmt.Errorf("configuration service not available: %w", err)
	}

	// Validate provider
	supportedProviders := configService.GetSupportedProviders()
	isValidProvider := false
	for _, supportedProvider := range supportedProviders {
		if provider == supportedProvider {
			isValidProvider = true
			break
		}
	}

	if !isValidProvider {
		return fmt.Errorf("unsupported provider '%s'. Supported providers: %s", provider, strings.Join(supportedProviders, ", "))
	}

	// For Gemini provider, delegate to specialized command with better key resolution
	if provider == "gemini" {
		return c.delegateToGeminiClientNew(args)
	}

	// Get API key - first from args, then from environment variable
	apiKey := args["key"]
	if apiKey == "" {
		// Get API key from environment variable via VariableService
		// This follows proper architecture: Command -> Service -> Context -> OS
		switch provider {
		case "openai":
			apiKey = variableService.GetEnv("OPENAI_API_KEY")
		case "openrouter":
			apiKey = variableService.GetEnv("OPENROUTER_API_KEY")
		case "moonshot":
			apiKey = variableService.GetEnv("MOONSHOT_API_KEY")
		case "anthropic":
			apiKey = variableService.GetEnv("ANTHROPIC_API_KEY")
		case "gemini":
			apiKey = variableService.GetEnv("GOOGLE_API_KEY")
		default:
			// This should never happen due to provider validation above
			return fmt.Errorf("unsupported provider '%s' in switch statement", provider)
		}

		// If still no API key found, return error
		if apiKey == "" {
			var envVarName string
			switch provider {
			case "openai":
				envVarName = "OPENAI_API_KEY"
			case "openrouter":
				envVarName = "OPENROUTER_API_KEY"
			case "moonshot":
				envVarName = "MOONSHOT_API_KEY"
			case "anthropic":
				envVarName = "ANTHROPIC_API_KEY"
			case "gemini":
				envVarName = "GOOGLE_API_KEY"
			default:
				envVarName = "API_KEY"
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

// delegateToGeminiClientNew handles Gemini provider by delegating to the specialized command.
// This leverages the robust key resolution system in gemini-client-new.
func (c *ClientGetCommand) delegateToGeminiClientNew(args map[string]string) error {
	// Get stack service to push the command for execution
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		return fmt.Errorf("stack service not available: %w", err)
	}

	// Build the gemini-client-new command with optional key parameter
	var commandStr string
	if key, exists := args["key"]; exists && key != "" {
		commandStr = fmt.Sprintf("\\gemini-client-new[key=%s]", key)
	} else {
		commandStr = "\\gemini-client-new"
	}

	// Push the command to the stack for execution
	stackService.PushCommand(commandStr)

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ClientGetCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-client-get command: %v", err))
	}
}
