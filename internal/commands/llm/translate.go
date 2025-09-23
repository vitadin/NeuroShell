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
	return `\translate[translator=provider, source=lang, target=lang, instruction="custom instructions"] text to translate

Examples:
  \translate Hello world                                    %% Auto-detect source, translate to English (via ZAI)
  \translate[target=spanish] Hello world                   %% Translate to Spanish
  \translate[source=english, target=french] Hello world    %% Explicit source and target
  \translate[translator=deepl] Hello world                 %% Use DeepL provider
  \translate[target=japanese, instruction="formal tone"] Hello world     %% With custom instructions
  \translate[target=german, instruction="business style"] Hello world    %% Business style instruction

ZAI-specific options (when translator=zai):
  \translate[strategy=paraphrase] Hello world              %% Use paraphrase strategy
  \translate[strategy=three_step] Hello world              %% High-quality multi-step translation
  \translate[glossary="tech_terms"] Hello world            %% Use custom glossary

Options:
  translator  - Translation provider (zai, deepl, google) [default: zai]
  source      - Source language (auto-detect if not specified)
  target      - Target language [default: english]
  instruction - Custom instructions for translation style, tone, context, etc.

ZAI-only options:
  strategy    - Translation strategy (general, paraphrase, two_step, three_step, reflection)
  glossary    - Glossary ID for custom terminology`
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
				Name:        "instruction",
				Description: "Custom instructions for translation style, tone, context, etc.",
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
				Command:     `\translate[target=japanese, instruction="the tone should be formal"] Please review this document`,
				Description: "Formal translation to Japanese with custom instructions",
			},
			{
				Command:     `\translate[translator=deepl, target=german, instruction="make it business style"] Our quarterly results show growth`,
				Description: "Business style translation using DeepL",
			},
		},
	}
}

// Execute processes the translate command with the provided options and text.
func (c *TranslateCommand) Execute(options map[string]string, input string) error {
	if !services.GetGlobalRegistry().HasService("variable") {
		return fmt.Errorf("variable service not available")
	}

	// Get text to translate - if empty, show help
	if strings.TrimSpace(input) == "" {
		fmt.Printf("Usage: %s\n\n", c.Usage())
		return nil
	}

	textToTranslate := strings.TrimSpace(input)

	// Extract and validate options
	translator := getOption(options, "translator", "zai")
	source := getOption(options, "source", "auto")
	target := getOption(options, "target", "english")
	instruction := getOption(options, "instruction", "")

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
		"instruction", instruction,
		"text_length", len(textToTranslate))

	// Delegate to specific translator implementations
	if translator == "zai" {
		return c.delegateToZaiTranslate(options, source, target, instruction, textToTranslate)
	}

	// For other translators (deepl, google), show placeholder for now
	fmt.Printf("Translating (%s → %s via %s)...\n", source, target, translator)

	if instruction != "" {
		fmt.Printf("Instruction: %s\n", instruction)
	}

	// This is a placeholder - actual translation will replace this
	fmt.Printf("\nOriginal: %s\n", textToTranslate)
	fmt.Printf("Translation: [Placeholder - %s translation will be implemented here]\n", translator)

	logger.Debug("Translation command executed successfully",
		"translator", translator,
		"source", source,
		"target", target,
		"instruction", instruction,
		"text_length", len(textToTranslate))

	return nil
}

// delegateToZaiTranslate delegates translation to the zai-translate command via stack service
func (c *TranslateCommand) delegateToZaiTranslate(originalOptions map[string]string, source, target, instruction, text string) error {
	// Get stack service for delegation
	stackService, err := services.GetGlobalStackService()
	if err != nil {
		// Fallback to placeholder if stack service is not available
		logger.Debug("stack service not available, using placeholder", "error", err)
		return c.showZaiPlaceholder(source, target, instruction, text)
	}

	// Build zai-translate command with options
	command := "\\zai-translate"

	// Map translate options to zai-translate options
	zaiOptions := make(map[string]string)

	// Map source language
	zaiOptions["source"] = source

	// Map target language, with conversion for zai defaults
	if target == "english" {
		zaiOptions["target"] = "en"
	} else {
		zaiOptions["target"] = target
	}

	// Map instruction parameter directly
	if instruction != "" {
		zaiOptions["instruction"] = instruction
	}

	// Copy any zai-specific options from original
	zaiSpecificOptions := []string{"strategy", "glossary"}
	for _, opt := range zaiSpecificOptions {
		if value, exists := originalOptions[opt]; exists {
			zaiOptions[opt] = value
		}
	}

	// Add options if provided
	if len(zaiOptions) > 0 {
		command += "["
		first := true
		for key, value := range zaiOptions {
			if !first {
				command += ","
			}
			command += key + "=" + value
			first = false
		}
		command += "]"
	}

	// Add the message input
	command += " " + text

	logger.Debug("Delegating to zai-translate command via stack service",
		"original_options", originalOptions,
		"zai_options", zaiOptions,
		"command", command,
		"text_length", len(text))

	// Delegate to zai-translate via stack service
	stackService.PushCommand(command)

	return nil
}

// showZaiPlaceholder shows placeholder output when zai-translate is not available
func (c *TranslateCommand) showZaiPlaceholder(source, target, instruction, text string) error {
	fmt.Printf("ZAI Translating (%s → %s)...\n", source, target)

	if instruction != "" {
		fmt.Printf("Instruction: %s\n", instruction)
	}

	fmt.Printf("\nOriginal: %s\n", text)
	fmt.Printf("Translation: [ZAI translation will be implemented here]\n")

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
