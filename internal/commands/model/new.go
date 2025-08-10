// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// NewCommand implements the \model-new command for creating new model configurations.
// It provides model creation functionality with configurable parameters for different LLM providers.
type NewCommand struct{}

// Name returns the command name "model-new" for registration and lookup.
func (c *NewCommand) Name() string {
	return "model-new"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *NewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-new command does.
func (c *NewCommand) Description() string {
	return "Create new LLM model configuration"
}

// Usage returns the syntax and usage examples for the model-new command.
func (c *NewCommand) Usage() string {
	return `\model-new[catalog_id=<ID>, temperature=0.7, max_tokens=1000, thinking_budget=1024, ...] model_name

Examples:
  \model-new[catalog_id=CS4] my-claude                                   %% Create from catalog (Claude Sonnet 4)
  \model-new[catalog_id=CS4, temperature=0.8] creative-claude             %% Create from catalog with custom parameters
  \model-new[catalog_id=GM25F, thinking_budget=2048] reasoning-model     %% Create Gemini Flash with thinking mode
  \model-new[catalog_id=GM25FL, thinking_budget=0] fast-model            %% Create Gemini Flash Lite with thinking disabled
  \model-new[catalog_id=GM25P, thinking_budget=-1] dynamic-model         %% Create Gemini Pro with dynamic thinking
  \model-new[catalog_id=O3] my-o3                                       %% Create OpenAI o3 (delegates to \\openai-model-new)
  \model-new[catalog_id=CO4, max_tokens=4000] analysis-opus              %% Create Claude Opus 4 with custom max tokens

Required Options:
  catalog_id - Short model ID from catalog (e.g., CS4, O3, CO37, GM25F) - auto-populates provider and base_model

Optional Parameters:
  temperature - Controls randomness (0.0-1.0, default varies by provider)
  max_tokens - Maximum tokens to generate (positive integer)
  top_p - Nucleus sampling parameter (0.0-1.0)
  top_k - Top-k sampling parameter (positive integer)
  presence_penalty - Presence penalty (-2.0 to 2.0)
  frequency_penalty - Frequency penalty (-2.0 to 2.0)
  thinking_budget - Thinking tokens budget for Gemini models (-1=dynamic, 0=disabled, positive=fixed)
  description - Human-readable description of the model configuration

Note: Model name is required and taken from the input parameter.
      Model names must be unique and cannot contain spaces.
      Use \model-catalog to see available catalog IDs.
      Provider and base_model are auto-populated from the model catalog.
      thinking_budget is only supported by Gemini 2.5 models (Pro, Flash, Flash Lite).
      Additional provider-specific parameters can be passed and will be stored.`
}

// HelpInfo returns structured help information for the model-new command.
func (c *NewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-new[catalog_id=<ID>, ...] model_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "catalog_id",
				Description: "Short model ID from catalog (e.g., CS4, O3, CO37, GM25F) - auto-populates provider and base_model",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "temperature",
				Description: "Controls randomness in model output (0.0-1.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "max_tokens",
				Description: "Maximum number of tokens to generate",
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
			{
				Name:        "thinking_budget",
				Description: "Thinking tokens budget for Gemini models (-1=dynamic, 0=disabled, positive=fixed)",
				Required:    false,
				Type:        "int",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-new[catalog_id=CS4] my-claude",
				Description: "Create model from catalog (Claude Sonnet 4)",
			},
			{
				Command:     "\\model-new[catalog_id=O3, temperature=0.8] creative-model",
				Description: "Create from catalog with custom parameters",
			},
			{
				Command:     "\\model-new[catalog_id=GM25F, thinking_budget=2048] reasoning-model",
				Description: "Create Gemini Flash with fixed thinking budget",
			},
			{
				Command:     "\\model-new[catalog_id=GM25FL, thinking_budget=0] fast-model",
				Description: "Create Gemini Flash Lite with thinking disabled",
			},
			{
				Command:     "\\model-new[catalog_id=GM25P, thinking_budget=-1] dynamic-model",
				Description: "Create Gemini Pro with dynamic thinking",
			},
			{
				Command:     "\\model-new[catalog_id=O3] my-o3",
				Description: "Create OpenAI o3 model (delegates to openai-model-new)",
			},
			{
				Command:     "\\model-new[catalog_id=CO4, max_tokens=4000] analysis-opus",
				Description: "Create Claude Opus 4 with custom max tokens",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#model_id",
				Description: "Unique identifier of the created model",
				Type:        "system_metadata",
				Example:     "550e8400-e29b-41d4",
			},
			{
				Name:        "#model_name",
				Description: "Name of the created model",
				Type:        "system_metadata",
				Example:     "my-claude",
			},
			{
				Name:        "#model_provider",
				Description: "Provider of the created model",
				Type:        "system_metadata",
				Example:     "anthropic",
			},
			{
				Name:        "#model_base",
				Description: "Base model identifier",
				Type:        "system_metadata",
				Example:     "claude-3-sonnet",
			},
			{
				Name:        "#model_created",
				Description: "Model creation timestamp",
				Type:        "system_metadata",
				Example:     "2024-01-15 14:30:25",
			},
			{
				Name:        "#model_param_count",
				Description: "Number of parameters configured",
				Type:        "system_metadata",
				Example:     "3",
			},
			{
				Name:        "#model_description",
				Description: "Model description (if provided)",
				Type:        "system_metadata",
				Example:     "Creative writing model",
			},
			{
				Name:        "_output",
				Description: "Command result message",
				Type:        "command_output",
				Example:     "Created model 'my-claude' (ID: 550e8400, Provider: anthropic, Base: claude-3-sonnet)",
			},
		},
		Notes: []string{
			"Model name is required and taken from the input parameter",
			"Model names must be unique and cannot contain spaces",
			"catalog_id is required - use \\model-catalog to see available catalog IDs",
			"Provider and base_model are auto-populated from the model catalog",
			"thinking_budget is only supported by Gemini 2.5 models (Pro, Flash, Flash Lite)",
			"thinking_budget values: -1=dynamic, 0=disabled, positive=fixed token count",
			"Each Gemini model has different thinking_budget ranges (see model catalog)",
			"OpenAI models are delegated to \\openai-model-new for specialized handling",
			"Variables in model name and parameters are interpolated",
			"Additional provider-specific parameters can be included",
		},
	}
}

// Execute creates a new model configuration with the specified parameters.
// The input parameter is used as the model name (required).
// Required options: provider, base_model
// Optional parameters: temperature, max_tokens, top_p, top_k, presence_penalty, frequency_penalty, description
func (c *NewCommand) Execute(args map[string]string, input string) error {
	// Get model service
	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Note: Variable interpolation is now handled by the state machine before commands execute

	// Get model catalog service for catalog_id support
	modelCatalogService, err := services.GetGlobalModelCatalogService()
	if err != nil {
		return fmt.Errorf("model catalog service not available: %w", err)
	}

	// Parse required arguments
	modelName := input
	catalogID := args["catalog_id"]

	// Validate required parameters
	if modelName == "" {
		return fmt.Errorf("model name is required\\n\\nUsage: %s", c.Usage())
	}

	// catalog_id is required
	if catalogID == "" {
		return fmt.Errorf("catalog_id is required\\n\\nUsage: %s", c.Usage())
	}

	// Look up model in catalog and auto-populate provider and base_model
	catalogEntry, err := modelCatalogService.GetModelByID(catalogID)
	if err != nil {
		return fmt.Errorf("failed to find model with catalog_id '%s': %w", catalogID, err)
	}
	catalogModel := &catalogEntry

	// Auto-populate provider and base_model from catalog
	provider := catalogModel.Provider
	baseModel := catalogModel.Name

	// For OpenAI provider, delegate to specialized command with better parameter handling
	if provider == "openai" {
		return c.delegateToOpenAIModelNew(args, input)
	}

	// For Gemini provider, delegate to specialized command with better thinking_budget handling
	if provider == "gemini" {
		return c.delegateToGeminiModelNew(args, input)
	}

	// Note: Variable interpolation for model name, provider, and base_model is handled by state machine

	// Parse optional parameters
	parameters := make(map[string]any)
	description := ""

	// Handle description separately
	if desc, exists := args["description"]; exists {
		description = desc
		// Note: Variable interpolation for description is handled by state machine
	}

	// Parse numeric parameters
	if err := c.parseParameters(args, parameters); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate parameters
	if err := modelService.ValidateModelParameters(parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create model configuration
	createdModel, err := modelService.CreateModelWithGlobalContext(modelName, provider, baseModel, parameters, description, catalogID)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Auto-push client creation command to stack service for seamless UX
	if stackService, err := services.GetGlobalStackService(); err == nil {
		clientCommand := c.generateClientNewCommand(createdModel.Provider)
		if clientCommand != "" {
			stackService.PushCommand(clientCommand)
		}
	}

	// Auto-push model activation command to stack service for seamless UX
	// Use precise ID-based activation to avoid any ambiguity
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := fmt.Sprintf("\\try \\silent \\model-activate[id=true] %s", createdModel.ID)
		stackService.PushCommand(activateCommand)
	}

	// Update model-related variables (not active model variables - that's done by model-activate)
	if err := c.updateModelVariables(createdModel, variableService); err != nil {
		return fmt.Errorf("failed to update model variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)",
		createdModel.Name, createdModel.ID[:8], createdModel.Provider, createdModel.BaseModel)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// parseParameters parses numeric and other typed parameters from string arguments.
func (c *NewCommand) parseParameters(args map[string]string, parameters map[string]any) error {
	// Parse temperature
	if temp, exists := args["temperature"]; exists {
		tempFloat, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			return fmt.Errorf("invalid temperature value: %s", temp)
		}
		parameters["temperature"] = tempFloat
	}

	// Parse max_tokens
	if maxTokens, exists := args["max_tokens"]; exists {
		maxTokensInt, err := strconv.Atoi(maxTokens)
		if err != nil {
			return fmt.Errorf("invalid max_tokens value: %s", maxTokens)
		}
		parameters["max_tokens"] = maxTokensInt
	}

	// Parse top_p
	if topP, exists := args["top_p"]; exists {
		topPFloat, err := strconv.ParseFloat(topP, 64)
		if err != nil {
			return fmt.Errorf("invalid top_p value: %s", topP)
		}
		parameters["top_p"] = topPFloat
	}

	// Parse top_k
	if topK, exists := args["top_k"]; exists {
		topKInt, err := strconv.Atoi(topK)
		if err != nil {
			return fmt.Errorf("invalid top_k value: %s", topK)
		}
		parameters["top_k"] = topKInt
	}

	// Parse presence_penalty
	if presencePenalty, exists := args["presence_penalty"]; exists {
		presencePenaltyFloat, err := strconv.ParseFloat(presencePenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid presence_penalty value: %s", presencePenalty)
		}
		parameters["presence_penalty"] = presencePenaltyFloat
	}

	// Parse frequency_penalty
	if frequencyPenalty, exists := args["frequency_penalty"]; exists {
		frequencyPenaltyFloat, err := strconv.ParseFloat(frequencyPenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid frequency_penalty value: %s", frequencyPenalty)
		}
		parameters["frequency_penalty"] = frequencyPenaltyFloat
	}

	// Add any other string parameters that aren't specially handled
	excludedParams := map[string]bool{
		"description": true, "catalog_id": true,
		"temperature": true, "max_tokens": true, "top_p": true, "top_k": true,
		"presence_penalty": true, "frequency_penalty": true,
	}

	for key, value := range args {
		if !excludedParams[key] {
			parameters[key] = value
		}
	}

	return nil
}

// updateModelVariables sets model-related system variables.
func (c *NewCommand) updateModelVariables(model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Set model variables
	variables := map[string]string{
		"#model_id":       model.ID,
		"#model_name":     model.Name,
		"#model_provider": model.Provider,
		"#model_base":     model.BaseModel,
		"#model_created":  model.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Add parameter count
	variables["#model_param_count"] = fmt.Sprintf("%d", len(model.Parameters))

	// Add description if present
	if model.Description != "" {
		variables["#model_description"] = model.Description
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// delegateToOpenAIModelNew handles OpenAI provider by delegating to the specialized command.
// This leverages the robust reasoning parameter handling in openai-model-new.
func (c *NewCommand) delegateToOpenAIModelNew(args map[string]string, input string) error {
	// Create openai-model-new command and execute it directly
	openaiModelNewCmd := &OpenAIModelNewCommand{}

	// Prepare args for the delegated command (exclude provider since openai-model-new is OpenAI-specific)
	delegateArgs := make(map[string]string)
	for key, value := range args {
		if key != "provider" {
			delegateArgs[key] = value
		}
	}

	// Execute the openai-model-new command directly
	return openaiModelNewCmd.Execute(delegateArgs, input)
}

// delegateToGeminiModelNew handles Gemini provider by delegating to the specialized command.
// This leverages the robust thinking_budget parameter handling in gemini-model-new.
func (c *NewCommand) delegateToGeminiModelNew(args map[string]string, input string) error {
	// Create gemini-model-new command and execute it directly
	geminiModelNewCmd := &GeminiModelNewCommand{}

	// Prepare args for the delegated command (exclude provider since gemini-model-new is Gemini-specific)
	delegateArgs := make(map[string]string)
	for key, value := range args {
		if key != "provider" {
			delegateArgs[key] = value
		}
	}

	// Execute the gemini-model-new command directly
	return geminiModelNewCmd.Execute(delegateArgs, input)
}

// generateClientNewCommand generates the appropriate client creation command for a provider.
func (c *NewCommand) generateClientNewCommand(provider string) string {
	switch provider {
	case "openai":
		return "\\try \\silent \\openai-client-new"
	case "anthropic":
		return "\\try \\silent \\anthropic-client-new"
	case "google":
		return "\\try \\silent \\gemini-client-new"
	default:
		// For unknown providers, return empty string (no client creation)
		return ""
	}
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-new command: %v", err))
	}
}
