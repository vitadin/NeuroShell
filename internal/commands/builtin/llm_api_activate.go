package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// LLMAPIActivateCommand implements the \llm-api-activate command for setting active API keys.
// It allows users to explicitly choose which API key to use for each provider.
type LLMAPIActivateCommand struct{}

// Name returns the command name "llm-api-activate" for registration and lookup.
func (c *LLMAPIActivateCommand) Name() string {
	return "llm-api-activate"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *LLMAPIActivateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-api-activate command does.
func (c *LLMAPIActivateCommand) Description() string {
	return "Activate an API key for a specific provider"
}

// Usage returns the syntax and usage examples for the llm-api-activate command.
func (c *LLMAPIActivateCommand) Usage() string {
	return "\\llm-api-activate[provider=<name>, key=<source.KEY_NAME>]"
}

// HelpInfo returns structured help information for the llm-api-activate command.
func (c *LLMAPIActivateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "Provider name (openai, anthropic, openrouter, moonshot, gemini)",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "key",
				Description: "Source-prefixed key name (e.g., os.A_OPENAI_KEY, config.OPENAI_API_KEY)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\llm-api-activate[provider=openai, key=os.A_OPENAI_KEY]",
				Description: "Activate an OpenAI key from OS environment variables",
			},
			{
				Command:     "\\llm-api-activate[provider=anthropic, key=config.ANTHROPIC_API_KEY]",
				Description: "Activate an Anthropic key from config .env file",
			},
			{
				Command:     "\\llm-api-activate[provider=moonshot, key=local.MOONSHOT_KEY]",
				Description: "Activate a Moonshot key from local .env file",
			},
			{
				Command:     "\\llm-api-activate[provider=gemini, key=os.GOOGLE_API_KEY]",
				Description: "Activate a Gemini key from OS environment variables",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#active_{provider}_key",
				Description: "Contains the activated API key value for the specified provider",
				Type:        "system_metadata",
				Example:     "#active_openai_key = \"sk-1234...\"",
			},
		},
		Notes: []string{
			"Use \\llm-api-show to see available keys and their source-prefixed names",
			"The key must exist as a collected variable (run \\llm-api-show first)",
		},
	}
}

// Execute activates the specified API key for the given provider.
func (c *LLMAPIActivateCommand) Execute(args map[string]string, _ string) error {
	// Validate required arguments
	provider := args["provider"]
	key := args["key"]

	if provider == "" {
		return fmt.Errorf("provider is required. Usage: %s", c.Usage())
	}

	if key == "" {
		return fmt.Errorf("key is required. Usage: %s", c.Usage())
	}

	// Get configuration service to validate provider
	configService, err := services.GetGlobalConfigurationService()
	if err != nil {
		return fmt.Errorf("configuration service not available: %w", err)
	}

	validProviders := configService.GetSupportedProviders()
	providerValid := false
	for _, validProvider := range validProviders {
		if provider == validProvider {
			providerValid = true
			break
		}
	}

	if !providerValid {
		return fmt.Errorf("invalid provider '%s'. Valid providers: %s",
			provider, strings.Join(validProviders, ", "))
	}

	// Get variable service
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Validate that the key exists as a collected variable
	keyValue, err := variableService.Get(key)
	if err != nil {
		return fmt.Errorf("key '%s' not found. Run \\llm-api-show to see available keys", key)
	}

	if keyValue == "" {
		return fmt.Errorf("key '%s' is empty. Run \\llm-api-show to see available keys", key)
	}

	// Validate key format (should be source.ORIGINAL_NAME)
	if !c.isValidKeyFormat(key) {
		return fmt.Errorf("invalid key format '%s'. Expected format: source.KEY_NAME (e.g., os.A_OPENAI_KEY)", key)
	}

	// Store the actual API key value in the active key system variable
	activeVarName := "#active_" + provider + "_key"
	err = variableService.SetSystemVariable(activeVarName, keyValue)
	if err != nil {
		return fmt.Errorf("failed to set active key: %w", err)
	}

	// Provide simple user feedback
	fmt.Printf("✓ Activated %s API key for %s provider\n", key, provider)
	fmt.Printf("  ${%s} now contains the API key value\n", activeVarName)

	return nil
}

// isValidKeyFormat checks if the key follows the expected source.ORIGINAL_NAME format
func (c *LLMAPIActivateCommand) isValidKeyFormat(key string) bool {
	// Should have format: source.ORIGINAL_NAME
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return false
	}

	source := parts[0]
	originalName := parts[1]

	// Validate source
	validSources := []string{"os", "config", "local"}
	sourceValid := false
	for _, validSource := range validSources {
		if source == validSource {
			sourceValid = true
			break
		}
	}

	if !sourceValid {
		return false
	}

	// Original name should not be empty
	if originalName == "" {
		return false
	}

	return true
}

func init() {
	if err := commands.GlobalRegistry.Register(&LLMAPIActivateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-api-activate command: %v", err))
	}
}
