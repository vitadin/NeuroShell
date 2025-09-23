// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ZaiTranslateCommand implements the \zai-translate command for direct ZAI API translation.
// It provides comprehensive access to ZAI's translation API features including multiple strategies,
// glossary support, and custom configurations.
type ZaiTranslateCommand struct{}

// Name returns the command name "zai-translate" for registration and lookup.
func (c *ZaiTranslateCommand) Name() string {
	return "zai-translate"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *ZaiTranslateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the zai-translate command does.
func (c *ZaiTranslateCommand) Description() string {
	return "Translate text using ZAI's general translation API with advanced features"
}

// Usage returns the syntax and usage examples for the zai-translate command.
func (c *ZaiTranslateCommand) Usage() string {
	return `\zai-translate[source=lang, target=lang, strategy=type, instruction="hint", glossary="id"] text to translate

Examples:
  \zai-translate Hello world                                    %% Auto-detect source, translate to Chinese
  \zai-translate[target=en] 你好世界                            %% Translate to English
  \zai-translate[source=en, target=ja] Hello world             %% English to Japanese
  \zai-translate[strategy=paraphrase] Hello world              %% Use paraphrase translation
  \zai-translate[strategy=three_step, target=fr] Hello world   %% Multi-step translation to French
  \zai-translate[instruction="formal business tone"] Hello     %% With style guidance

Options:
  source      - Source language code (auto, en, zh-CN, ja, etc.) [default: auto]
  target      - Target language code (zh-CN, en, ja, fr, de, etc.) [default: zh-CN]
  strategy    - Translation strategy (general, paraphrase, two_step, three_step, reflection) [default: general]
  instruction - Translation style or terminology hints for general strategy
  glossary    - Glossary ID for custom terminology`
}

// HelpInfo returns structured help information for the zai-translate command.
func (c *ZaiTranslateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "source",
				Description: "Source language (auto-detect, en, zh-CN, ja, fr, de, es, ru, etc.)",
				Required:    false,
				Type:        "string",
				Default:     "auto",
			},
			{
				Name:        "target",
				Description: "Target language (zh-CN, en, ja, fr, de, es, ru, etc.)",
				Required:    false,
				Type:        "string",
				Default:     "zh-CN",
			},
			{
				Name:        "strategy",
				Description: "Translation strategy (general, paraphrase, two_step, three_step, reflection)",
				Required:    false,
				Type:        "string",
				Default:     "general",
			},
			{
				Name:        "instruction",
				Description: "Style hints for general strategy (e.g., 'formal business tone', 'casual')",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "glossary",
				Description: "Glossary ID for custom terminology",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\zai-translate Hello, how are you?`,
				Description: "Basic translation with auto-detection to Chinese",
			},
			{
				Command:     `\zai-translate[target=en] 你好，你好吗？`,
				Description: "Translate Chinese to English",
			},
			{
				Command:     `\zai-translate[source=en, target=ja, strategy=paraphrase] Please help me`,
				Description: "Paraphrase translation from English to Japanese",
			},
			{
				Command:     `\zai-translate[strategy=three_step, target=fr] Technical documentation`,
				Description: "High-quality three-step translation to French",
			},
			{
				Command:     `\zai-translate[instruction="formal business tone", target=de] Hello colleagues`,
				Description: "Business-style translation to German",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Translated text result",
				Type:        "command_output",
				Example:     "你好世界",
			},
			{
				Name:        "_translation_id",
				Description: "ZAI translation task ID",
				Type:        "command_output",
				Example:     "trans_abc123",
			},
			{
				Name:        "_source_detected",
				Description: "Auto-detected source language",
				Type:        "command_output",
				Example:     "en",
			},
			{
				Name:        "_tokens_used",
				Description: "Total tokens consumed in translation",
				Type:        "command_output",
				Example:     "156",
			},
		},
		Notes: []string{
			"Requires ZAI API key: set Z_DOT_AI_API_KEY or ZAI_API_KEY environment variable",
			"Supports 40+ languages including specialized variants (zh-TW, en-US, etc.)",
			"Strategy comparison:",
			"  • general: Fast, good for most content",
			"  • paraphrase: More natural phrasing",
			"  • two_step: Better accuracy via intermediate steps",
			"  • three_step: Highest quality for critical content",
			"  • reflection: Self-correcting translation",
			"Instruction parameter works only with 'general' strategy",
			"Auto-detected source language stored in _source_detected",
			"Translation cost and token usage stored in _tokens_used",
		},
	}
}

// Execute processes the zai-translate command with ZAI API integration.
func (c *ZaiTranslateCommand) Execute(options map[string]string, input string) error {
	if !services.GetGlobalRegistry().HasService("variable") {
		return fmt.Errorf("variable service not available")
	}

	if !services.GetGlobalRegistry().HasService("http_request") {
		return fmt.Errorf("http request service not available")
	}

	// Get text to translate - if empty, show help
	if strings.TrimSpace(input) == "" {
		fmt.Printf("Usage: %s\n\n", c.Usage())
		return nil
	}

	textToTranslate := strings.TrimSpace(input)

	// Extract and validate options
	source := getOption(options, "source", "auto")
	target := getOption(options, "target", "zh-CN")
	strategy := getOption(options, "strategy", "general")
	instruction := getOption(options, "instruction", "")
	glossary := getOption(options, "glossary", "")

	// Validate strategy
	validStrategies := map[string]bool{
		"general":    true,
		"paraphrase": true,
		"two_step":   true,
		"three_step": true,
		"reflection": true,
	}
	if !validStrategies[strategy] {
		return fmt.Errorf("unsupported strategy '%s'. Supported: general, paraphrase, two_step, three_step, reflection", strategy)
	}

	// Check for ZAI API key
	variableService, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// Get ZAI API key from OS environment variable (try both names)
	apiKey, _ := variableService.(*services.VariableService).Get("os.Z_DOT_AI_API_KEY")
	if apiKey == "" {
		apiKey, _ = variableService.(*services.VariableService).Get("os.ZAI_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("ZAI API key not found. Please set Z_DOT_AI_API_KEY or ZAI_API_KEY environment variable")
	}

	logger.Debug("ZAI translation request initiated",
		"source", source,
		"target", target,
		"strategy", strategy,
		"instruction", instruction,
		"glossary", glossary,
		"text_length", len(textToTranslate))

	// Build ZAI API request
	requestBody := map[string]interface{}{
		"agent_id": "general_translation",
		"stream":   false,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": textToTranslate,
					},
				},
			},
		},
		"custom_variables": map[string]interface{}{
			"source_lang": source,
			"target_lang": target,
			"strategy":    strategy,
		},
	}

	// Add optional parameters
	customVars := requestBody["custom_variables"].(map[string]interface{})
	if glossary != "" {
		customVars["glossary"] = glossary
	}

	// Add strategy-specific config
	if strategy == "general" && instruction != "" {
		customVars["strategy_config"] = map[string]interface{}{
			"general": map[string]interface{}{
				"suggestion": instruction,
			},
		}
	}

	// Convert to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request JSON: %w", err)
	}

	// Get HTTP service
	httpService, err := services.GetGlobalRegistry().GetService("http_request")
	if err != nil {
		return fmt.Errorf("failed to get HTTP service: %w", err)
	}

	httpRequestService := httpService.(*services.HTTPRequestService)

	// Prepare headers
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	// Display translation status
	fmt.Printf("ZAI Translating (%s → %s", source, target)
	if strategy != "general" {
		fmt.Printf(", %s strategy", strategy)
	}
	fmt.Printf(")...\n")

	if instruction != "" {
		fmt.Printf("Style guidance: %s\n", instruction)
	}

	fmt.Printf("\nOriginal: %s\n", textToTranslate)

	// Make HTTP request to ZAI API
	response, err := httpRequestService.Post("https://api.z.ai/api/v1/agents", string(requestJSON), headers)
	if err != nil {
		logger.Error("ZAI API request failed", "error", err)
		return fmt.Errorf("ZAI API request failed: %w", err)
	}

	if response.StatusCode != 200 {
		logger.Error("ZAI API returned error", "status_code", response.StatusCode, "body", response.Body)
		return fmt.Errorf("ZAI API error (status %d): %s", response.StatusCode, response.Body)
	}

	// Parse response
	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &apiResponse); err != nil {
		return fmt.Errorf("failed to parse ZAI API response: %w", err)
	}

	// Extract translation result
	choices, ok := apiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fmt.Errorf("no translation choices in ZAI API response")
	}

	choice := choices[0].(map[string]interface{})
	messages, ok := choice["messages"].([]interface{})
	if !ok || len(messages) == 0 {
		return fmt.Errorf("no messages in ZAI API response")
	}

	message := messages[0].(map[string]interface{})
	content, ok := message["content"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no content in ZAI API response")
	}

	translatedText, ok := content["text"].(string)
	if !ok {
		return fmt.Errorf("no text in ZAI API response content")
	}

	// Display result
	fmt.Printf("Translation: %s\n", translatedText)

	// Store results in variables
	if err := variableService.(*services.VariableService).SetSystemVariable("_output", translatedText); err != nil {
		logger.Error("Failed to set _output variable", "error", err)
	}

	if taskID, ok := apiResponse["id"].(string); ok {
		if err := variableService.(*services.VariableService).SetSystemVariable("_translation_id", taskID); err != nil {
			logger.Error("Failed to set _translation_id variable", "error", err)
		}
	}

	if source == "auto" {
		// For auto-detection, we'd need to parse the actual detected language from response
		// For now, store the original source parameter
		if err := variableService.(*services.VariableService).SetSystemVariable("_source_detected", source); err != nil {
			logger.Error("Failed to set _source_detected variable", "error", err)
		}
	}

	// Store token usage if available
	if usage, ok := apiResponse["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			if err := variableService.(*services.VariableService).SetSystemVariable("_tokens_used", fmt.Sprintf("%.0f", totalTokens)); err != nil {
				logger.Error("Failed to set _tokens_used variable", "error", err)
			}
		}
	}

	logger.Debug("ZAI translation completed successfully",
		"source", source,
		"target", target,
		"strategy", strategy,
		"text_length", len(textToTranslate),
		"result_length", len(translatedText))

	return nil
}

// IsReadOnly returns false as the command sets variables.
func (c *ZaiTranslateCommand) IsReadOnly() bool {
	return false
}

// init registers the ZaiTranslateCommand with the global command registry.
func init() {
	if err := commands.GetGlobalRegistry().Register(&ZaiTranslateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register zai-translate command: %v", err))
	}
}
