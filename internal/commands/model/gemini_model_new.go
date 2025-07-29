// Package model provides commands for managing LLM model configurations.
package model

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// GeminiModelNewCommand implements the \gemini-model-new command for creating Gemini-specific models.
// It handles Gemini thinking models and their specialized thinking_budget parameters.
type GeminiModelNewCommand struct{}

// Name returns the command name "gemini-model-new" for registration and lookup.
func (c *GeminiModelNewCommand) Name() string {
	return "gemini-model-new"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *GeminiModelNewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of the gemini-model-new command.
func (c *GeminiModelNewCommand) Description() string {
	return "Create Gemini model configurations with thinking support"
}

// Usage returns the syntax and usage examples for the gemini-model-new command.
func (c *GeminiModelNewCommand) Usage() string {
	return `\gemini-model-new[catalog_id=<ID>, thinking_budget=<budget>, temperature=0.7, max_tokens=1000, ...] model_name

Examples:
  \gemini-model-new[catalog_id=GM25F] my-flash                               %% Create Gemini Flash with defaults
  \gemini-model-new[catalog_id=GM25F, thinking_budget=2048] reasoning-flash  %% Create Flash with fixed thinking budget
  \gemini-model-new[catalog_id=GM25F, thinking_budget=0] fast-flash          %% Create Flash with thinking disabled
  \gemini-model-new[catalog_id=GM25F, thinking_budget=-1] dynamic-flash      %% Create Flash with dynamic thinking
  \gemini-model-new[catalog_id=GM25P, thinking_budget=16384] thorough-pro    %% Create Pro with large thinking budget
  \gemini-model-new[catalog_id=GM25FL, temperature=0.9] creative-lite        %% Create Flash Lite with high temperature

Options:
  catalog_id - Short model ID from catalog (GM25F, GM25P, GM25FL, etc.)
  thinking_budget - Thinking tokens budget (-1=dynamic, 0=disabled, positive=fixed token count)
  temperature - Sampling temperature (0.0-1.0)
  max_tokens - Maximum completion tokens
  top_p - Nucleus sampling parameter (0.0-1.0)
  top_k - Top-k sampling parameter (positive integer)
  presence_penalty - Presence penalty (-2.0 to 2.0)
  frequency_penalty - Frequency penalty (-2.0 to 2.0)
  description - Human-readable description

Note: catalog_id is required.
      thinking_budget ranges vary by model:
      - GM25F (Flash): 0-24576 tokens, can disable
      - GM25P (Pro): 128-32768 tokens, cannot disable
      - GM25FL (Flash Lite): varies by model
      Use \model-catalog[provider=gemini] to see available Gemini models.`
}

// HelpInfo returns structured help information for the gemini-model-new command.
func (c *GeminiModelNewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\gemini-model-new[catalog_id=<ID>, thinking_budget=<budget>, ...] model_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "catalog_id",
				Description: "Short model ID from catalog (GM25F, GM25P, GM25FL, etc.)",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "thinking_budget",
				Description: "Thinking tokens budget (-1=dynamic, 0=disabled, positive=fixed token count)",
				Required:    false,
				Type:        "int",
				Default:     "-1",
			},
			{
				Name:        "temperature",
				Description: "Sampling temperature (0.0-1.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "max_tokens",
				Description: "Maximum completion tokens",
				Required:    false,
				Type:        "int",
			},
			{
				Name:        "top_p",
				Description: "Nucleus sampling parameter (0.0-1.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "top_k",
				Description: "Top-k sampling parameter",
				Required:    false,
				Type:        "int",
			},
			{
				Name:        "presence_penalty",
				Description: "Presence penalty (-2.0 to 2.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "frequency_penalty",
				Description: "Frequency penalty (-2.0 to 2.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "description",
				Description: "Human-readable description of the model configuration",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\gemini-model-new[catalog_id=GM25F] my-flash",
				Description: "Create Gemini Flash with default settings",
			},
			{
				Command:     "\\gemini-model-new[catalog_id=GM25F, thinking_budget=2048] reasoning-flash",
				Description: "Create Flash with fixed thinking budget",
			},
			{
				Command:     "\\gemini-model-new[catalog_id=GM25F, thinking_budget=0] fast-flash",
				Description: "Create Flash with thinking disabled",
			},
			{
				Command:     "\\gemini-model-new[catalog_id=GM25P, thinking_budget=-1] dynamic-pro",
				Description: "Create Pro with dynamic thinking",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_model_id",
				Description: "Contains the created model's unique ID",
				Type:        "system_output",
				Example:     "_model_id = \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"",
			},
			{
				Name:        "_model_name",
				Description: "Contains the created model's name",
				Type:        "system_output",
				Example:     "_model_name = \"my-flash\"",
			},
			{
				Name:        "#active_model_name",
				Description: "Contains the active model name (automatically set)",
				Type:        "system_metadata",
				Example:     "#active_model_name = \"my-flash\"",
			},
			{
				Name:        "#active_model_id",
				Description: "Contains the active model ID (automatically set)",
				Type:        "system_metadata",
				Example:     "#active_model_id = \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"",
			},
			{
				Name:        "#active_model_provider",
				Description: "Contains the active model provider (always 'gemini')",
				Type:        "system_metadata",
				Example:     "#active_model_provider = \"gemini\"",
			},
		},
		Notes: []string{
			"catalog_id is required",
			"Use \\model-catalog[provider=gemini] to see available Gemini models",
			"thinking_budget ranges vary by model - see model catalog for details",
			"thinking_budget: -1=dynamic, 0=disabled, positive=fixed token count",
			"Some models cannot disable thinking (thinking_budget=0 not allowed)",
			"Model is automatically activated after creation",
			"Variables in model name and parameters are interpolated",
		},
	}
}

// Execute creates a new Gemini model configuration.
func (c *GeminiModelNewCommand) Execute(args map[string]string, input string) error {
	// Validate model name (input parameter)
	modelName := strings.TrimSpace(input)
	if modelName == "" {
		return fmt.Errorf("model name is required. Usage: %s", c.Usage())
	}

	// Get required services
	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	catalogService, err := services.GetGlobalModelCatalogService()
	if err != nil {
		return fmt.Errorf("model catalog service not available: %w", err)
	}

	// Determine base model and provider
	provider := "gemini"
	var baseModel string
	var catalogModel *neurotypes.ModelCatalogEntry

	catalogID := args["catalog_id"]

	if catalogID == "" {
		return fmt.Errorf("catalog_id is required")
	}

	// Look up model in catalog
	entry, err := catalogService.GetModelByID(catalogID)
	if err != nil {
		return fmt.Errorf("failed to find model with catalog_id '%s': %w", catalogID, err)
	}

	if entry.Provider != "gemini" {
		return fmt.Errorf("catalog_id '%s' is not a Gemini model (provider: %s). Use \\gemini-model-new only for Gemini models", catalogID, entry.Provider)
	}

	catalogModel = &entry
	baseModel = entry.Name

	// Parse and validate parameters
	parameters := make(map[string]any)
	if err := c.parseGeminiParameters(args, parameters, catalogModel); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Get description
	description := args["description"]
	if description == "" && catalogModel != nil {
		// Only set default description if no description argument was provided at all
		if _, exists := args["description"]; !exists {
			description = fmt.Sprintf("Gemini %s model", catalogModel.DisplayName)
		}
	}

	// Validate parameters
	if err := modelService.ValidateModelParameters(parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create model using the proper service method
	createdModel, err := modelService.CreateModelWithGlobalContext(modelName, provider, baseModel, parameters, description, catalogID)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_model_id", createdModel.ID)
	_ = variableService.SetSystemVariable("_model_name", createdModel.Name)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)", createdModel.Name, createdModel.ID[:8], createdModel.Provider, createdModel.BaseModel))

	// Set metadata variables (consistent with regular model-new command)
	_ = variableService.SetSystemVariable("#model_id", createdModel.ID)
	_ = variableService.SetSystemVariable("#model_name", createdModel.Name)
	_ = variableService.SetSystemVariable("#model_provider", createdModel.Provider)
	_ = variableService.SetSystemVariable("#model_base", createdModel.BaseModel)
	_ = variableService.SetSystemVariable("#model_created", createdModel.CreatedAt.Format("2006-01-02 15:04:05"))
	_ = variableService.SetSystemVariable("#model_param_count", fmt.Sprintf("%d", len(createdModel.Parameters)))

	// Set model description if present
	if createdModel.Description != "" {
		_ = variableService.SetSystemVariable("#model_description", createdModel.Description)
	}

	// Auto-push model activation command to stack service for seamless UX (following model-new tradition)
	// Use precise ID-based activation to avoid any ambiguity
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := fmt.Sprintf("\\silent \\model-activate[id=true] %s", createdModel.ID)
		stackService.PushCommand(activateCommand)
	}

	// Output success message
	fmt.Printf("Created model '%s' (ID: %s, Provider: %s, Base: %s)\n", createdModel.Name, createdModel.ID[:8], createdModel.Provider, createdModel.BaseModel)

	return nil
}

// parseGeminiParameters parses and validates Gemini-specific parameters.
func (c *GeminiModelNewCommand) parseGeminiParameters(args map[string]string, parameters map[string]any, catalogModel *neurotypes.ModelCatalogEntry) error {
	// Parse thinking_budget (Gemini-specific)
	if thinkingBudget, exists := args["thinking_budget"]; exists {
		thinkingBudgetInt, err := strconv.Atoi(thinkingBudget)
		if err != nil {
			return fmt.Errorf("invalid thinking_budget value: %s", thinkingBudget)
		}

		// Validate thinking_budget using catalog information
		if err := c.validateThinkingBudget(thinkingBudgetInt, catalogModel); err != nil {
			return fmt.Errorf("invalid thinking_budget: %w", err)
		}

		parameters["thinking_budget"] = thinkingBudgetInt
	}

	// Parse standard parameters
	if err := c.parseStandardParameters(args, parameters); err != nil {
		return err
	}

	// Add any other string parameters that aren't specially handled
	excludedParams := map[string]bool{
		"catalog_id": true, "description": true,
		"thinking_budget": true,
		"temperature":     true, "max_tokens": true, "top_p": true, "top_k": true,
		"presence_penalty": true, "frequency_penalty": true,
	}

	for key, value := range args {
		if !excludedParams[key] {
			parameters[key] = value
		}
	}

	return nil
}

// parseStandardParameters parses standard model parameters.
func (c *GeminiModelNewCommand) parseStandardParameters(args map[string]string, parameters map[string]any) error {
	// Parse temperature
	if temp, exists := args["temperature"]; exists {
		tempFloat, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			return fmt.Errorf("invalid temperature value: %s", temp)
		}
		if tempFloat < 0.0 || tempFloat > 1.0 {
			return fmt.Errorf("temperature must be between 0.0 and 1.0: %f", tempFloat)
		}
		parameters["temperature"] = tempFloat
	}

	// Parse max_tokens
	if maxTokens, exists := args["max_tokens"]; exists {
		maxTokensInt, err := strconv.Atoi(maxTokens)
		if err != nil {
			return fmt.Errorf("invalid max_tokens value: %s", maxTokens)
		}
		if maxTokensInt <= 0 {
			return fmt.Errorf("max_tokens must be positive: %d", maxTokensInt)
		}
		parameters["max_tokens"] = maxTokensInt
	}

	// Parse top_p
	if topP, exists := args["top_p"]; exists {
		topPFloat, err := strconv.ParseFloat(topP, 64)
		if err != nil {
			return fmt.Errorf("invalid top_p value: %s", topP)
		}
		if topPFloat < 0.0 || topPFloat > 1.0 {
			return fmt.Errorf("top_p must be between 0.0 and 1.0: %f", topPFloat)
		}
		parameters["top_p"] = topPFloat
	}

	// Parse top_k
	if topK, exists := args["top_k"]; exists {
		topKInt, err := strconv.Atoi(topK)
		if err != nil {
			return fmt.Errorf("invalid top_k value: %s", topK)
		}
		if topKInt <= 0 {
			return fmt.Errorf("top_k must be positive: %d", topKInt)
		}
		parameters["top_k"] = topKInt
	}

	// Parse presence_penalty
	if presPenalty, exists := args["presence_penalty"]; exists {
		presPenaltyFloat, err := strconv.ParseFloat(presPenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid presence_penalty value: %s", presPenalty)
		}
		if presPenaltyFloat < -2.0 || presPenaltyFloat > 2.0 {
			return fmt.Errorf("presence_penalty must be between -2.0 and 2.0: %f", presPenaltyFloat)
		}
		parameters["presence_penalty"] = presPenaltyFloat
	}

	// Parse frequency_penalty
	if freqPenalty, exists := args["frequency_penalty"]; exists {
		freqPenaltyFloat, err := strconv.ParseFloat(freqPenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid frequency_penalty value: %s", freqPenalty)
		}
		if freqPenaltyFloat < -2.0 || freqPenaltyFloat > 2.0 {
			return fmt.Errorf("frequency_penalty must be between -2.0 and 2.0: %f", freqPenaltyFloat)
		}
		parameters["frequency_penalty"] = freqPenaltyFloat
	}

	return nil
}

// validateThinkingBudget validates thinking_budget parameter for Gemini models using catalog information.
func (c *GeminiModelNewCommand) validateThinkingBudget(thinkingBudget int, catalogModel *neurotypes.ModelCatalogEntry) error {
	// Only validate for Gemini models
	if catalogModel.Provider != "gemini" {
		return nil
	}

	// Check if model supports thinking
	if catalogModel.Features == nil || catalogModel.Features.ThinkingSupported == nil || !*catalogModel.Features.ThinkingSupported {
		return fmt.Errorf("model %s does not support thinking mode", catalogModel.Name)
	}

	// Parse thinking range if available
	if catalogModel.Features.ThinkingRange != nil {
		rangeParts := strings.Split(*catalogModel.Features.ThinkingRange, "-")
		if len(rangeParts) == 2 {
			minRange, err1 := strconv.Atoi(rangeParts[0])
			maxRange, err2 := strconv.Atoi(rangeParts[1])
			if err1 == nil && err2 == nil {
				// Special case for -1 (dynamic thinking)
				if thinkingBudget == -1 {
					return nil // Dynamic thinking is always valid
				}

				// Special case for 0 (disabled thinking)
				if thinkingBudget == 0 {
					if catalogModel.Features.ThinkingCanDisable != nil && *catalogModel.Features.ThinkingCanDisable {
						return nil // Disabling is allowed
					}
					return fmt.Errorf("thinking cannot be disabled for model %s (thinking_budget=0 not allowed)", catalogModel.Name)
				}

				// Validate range for positive values
				if thinkingBudget < minRange || thinkingBudget > maxRange {
					return fmt.Errorf("thinking_budget %d is outside valid range %s for model %s", thinkingBudget, *catalogModel.Features.ThinkingRange, catalogModel.Name)
				}
			}
		}
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&GeminiModelNewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register gemini-model-new command: %v", err))
	}
}
