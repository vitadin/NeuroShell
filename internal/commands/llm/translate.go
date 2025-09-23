// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// TranslateCommand implements the \translate command for text translation.
// It supports multiple translation providers and customization options.
type TranslateCommand struct{}

// Name returns the command name "translate" for registration and lookup.
func (c *TranslateCommand) Name() string {
	return "translate"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *TranslateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the translate command does.
func (c *TranslateCommand) Description() string {
	return "Translate text using AI translation services with customizable options"
}

// Usage returns the syntax and usage examples for the translate command.
func (c *TranslateCommand) Usage() string {
	return `\translate[translator=provider, source=lang, target=lang, tone=style] text to translate

Examples:
  \translate Hello world                                    %% Auto-detect source, translate to system default
  \translate[target=spanish] Hello world                   %% Translate to Spanish
  \translate[source=english, target=french] Hello world    %% Explicit source and target
  \translate[translator=zai] Hello world                   %% Use specific provider
  \translate[target=japanese, tone=formal] Hello world     %% With tone specification
  \translate[target=german, tone=casual, context=business] Hello world  %% Business context

Options:
  translator - Translation provider (zai, deepl, google) [default: zai]
  source     - Source language (auto-detect if not specified)
  target     - Target language [default: english]
  tone       - Translation tone/style (formal, casual, literary, technical)
  context    - Additional context for translation (business, medical, legal, casual)
  preserve   - Elements to preserve (formatting, names, technical_terms)`
}

// HelpInfo returns structured help information for the translate command.
func (c *TranslateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "translator",
				Description: "Translation service provider (zai, deepl, google)",
				Required:    false,
				Type:        "string",
				Default:     "zai",
			},
			{
				Name:        "source",
				Description: "Source language (auto-detect if not specified)",
				Required:    false,
				Type:        "string",
				Default:     "auto",
			},
			{
				Name:        "target",
				Description: "Target language for translation",
				Required:    false,
				Type:        "string",
				Default:     "english",
			},
			{
				Name:        "tone",
				Description: "Translation tone/style (formal, casual, literary, technical, conversational)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "context",
				Description: "Domain context for specialized translation (business, medical, legal, technical, academic, casual)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "preserve",
				Description: "Elements to preserve during translation (formatting, names, technical_terms, urls, numbers)",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\translate Hello, how are you?`,
				Description: "Basic translation to default target language",
			},
			{
				Command:     `\translate[target=spanish] Hello, how are you?`,
				Description: "Translate to Spanish",
			},
			{
				Command:     `\translate[source=french, target=english] Bonjour, comment allez-vous?`,
				Description: "Translate from French to English",
			},
			{
				Command:     `\translate[target=japanese, tone=formal] Please review this document`,
				Description: "Formal translation to Japanese",
			},
			{
				Command:     `\translate[translator=deepl, target=german, context=business] Our quarterly results show growth`,
				Description: "Business context translation using DeepL",
			},
		},
	}
}

// Execute processes the translate command with the provided options and text.
func (c *TranslateCommand) Execute(options map[string]string, input string) error {
	if !services.GetGlobalRegistry().HasService("variable") {
		return fmt.Errorf("variable service not available")
	}

	// Get text to translate
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("no text provided for translation")
	}

	textToTranslate := strings.TrimSpace(input)

	// Extract and validate options
	translator := getOption(options, "translator", "zai")
	source := getOption(options, "source", "auto")
	target := getOption(options, "target", "english")
	tone := getOption(options, "tone", "")
	context := getOption(options, "context", "")
	preserve := getOption(options, "preserve", "")

	// Validate translator
	validTranslators := map[string]bool{
		"zai":    true,
		"deepl":  true,
		"google": true,
	}
	if !validTranslators[translator] {
		return fmt.Errorf("unsupported translator '%s'. Supported: zai, deepl, google", translator)
	}

	// Log translation request
	logger.Debug("Translation request initiated",
		"translator", translator,
		"source", source,
		"target", target,
		"tone", tone,
		"context", context,
		"preserve", preserve,
		"text_length", len(textToTranslate))

	// Create translation configuration
	translationConfig := map[string]interface{}{
		"translator":      translator,
		"source_language": source,
		"target_language": target,
		"text":            textToTranslate,
	}

	// Add optional parameters if specified
	if tone != "" {
		translationConfig["tone"] = tone
	}
	if context != "" {
		translationConfig["context"] = context
	}
	if preserve != "" {
		translationConfig["preserve"] = preserve
	}

	// Future: Store configuration in variables for potential command chaining
	// This will be implemented when the actual translation service is added

	// For now, just output the configuration - actual translation will be implemented later
	fmt.Printf("🌐 Translation Configuration:\n")
	fmt.Printf("   Provider: %s\n", translator)
	fmt.Printf("   Source: %s → Target: %s\n", source, target)
	if tone != "" {
		fmt.Printf("   Tone: %s\n", tone)
	}
	if context != "" {
		fmt.Printf("   Context: %s\n", context)
	}
	if preserve != "" {
		fmt.Printf("   Preserve: %s\n", preserve)
	}
	fmt.Printf("   Text: %s\n", textToTranslate)
	fmt.Printf("\n💡 Translation service integration will be implemented in the next phase.\n")

	logger.Info("Translation command executed successfully",
		"translator", translator,
		"source", source,
		"target", target,
		"text_length", len(textToTranslate))

	return nil
}

// getOption retrieves an option value with a fallback default.
func getOption(options map[string]string, key, defaultValue string) string {
	if value, exists := options[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// IsReadOnly returns false as the translate command modifies system state (sets variables).
func (c *TranslateCommand) IsReadOnly() bool {
	return false
}

// init registers the TranslateCommand with the global command registry.
func init() {
	if err := commands.GetGlobalRegistry().Register(&TranslateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register translate command: %v", err))
	}
}
