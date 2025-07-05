// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"strconv"
	"time"

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
	return `\model-new[provider=provider_name, base_model=model_name, temperature=0.7, max_tokens=1000, ...] model_name
\model-new[from_id=existing_model_id, temperature=0.8, ...] new_model_name

Examples:
  \model-new[provider=openai, base_model=gpt-4] my-gpt4                    %% Create OpenAI GPT-4 model
  \model-new[provider=anthropic, base_model=claude-3-sonnet] claude-work   %% Create Anthropic Claude model
  \model-new[provider=openai, base_model=gpt-3.5-turbo, temperature=0.9] creative-gpt  %% Custom temperature
  \model-new[from_id=${_catalog_model_id}, temperature=0.1] precise-model  %% Create from catalog discovery
  \model-new[from_id=existing-model, max_tokens=2000] custom-model         %% Clone and customize existing model
  
Creation Methods:
  Method 1 - From Provider/Base Model:
    provider - LLM provider name (e.g., openai, anthropic, local) [REQUIRED]
    base_model - Provider's model identifier (e.g., gpt-4, claude-3-sonnet) [REQUIRED]
    
  Method 2 - From Existing Model:
    from_id - ID of existing model to clone/customize [REQUIRED]
    
Optional Parameters (both methods):
  temperature - Controls randomness (0.0-1.0, default varies by provider)
  max_tokens - Maximum tokens to generate (positive integer)
  top_p - Nucleus sampling parameter (0.0-1.0)
  top_k - Top-k sampling parameter (positive integer)
  presence_penalty - Presence penalty (-2.0 to 2.0)
  frequency_penalty - Frequency penalty (-2.0 to 2.0)
  description - Human-readable description of the model configuration
  
Note: Model name is required and taken from the input parameter.
      Model names must be unique and cannot contain spaces.
      Cannot combine from_id with provider/base_model options.
      Additional provider-specific parameters can be passed and will be stored.`
}

// HelpInfo returns structured help information for the model-new command.
func (c *NewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-new[provider=provider_name, base_model=model_name, ...] model_name or \\model-new[from_id=existing_model_id, ...] new_model_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "LLM provider name (e.g., openai, anthropic, local) [Method 1]",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "base_model",
				Description: "Provider's model identifier (e.g., gpt-4, claude-3-sonnet) [Method 1]",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "from_id",
				Description: "ID of existing model to clone/customize [Method 2]",
				Required:    false,
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
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-new[provider=openai, base_model=gpt-4] my-gpt4",
				Description: "Create OpenAI GPT-4 model configuration",
			},
			{
				Command:     "\\model-new[provider=anthropic, base_model=claude-3-sonnet] claude-work",
				Description: "Create Anthropic Claude model configuration",
			},
			{
				Command:     "\\model-new[provider=openai, base_model=gpt-3.5-turbo, temperature=0.9] creative-gpt",
				Description: "Create model with custom temperature setting",
			},
			{
				Command:     "\\model-new[from_id=${_catalog_model_id}, temperature=0.1] precise-model",
				Description: "Create model from catalog discovery with custom temperature",
			},
			{
				Command:     "\\model-new[from_id=existing-model, max_tokens=2000] custom-model",
				Description: "Clone existing model with custom token limit",
			},
		},
		Notes: []string{
			"Model name is required and taken from the input parameter",
			"Model names must be unique and cannot contain spaces",
			"Use either provider+base_model OR from_id, not both",
			"from_id enables cloning/templating from existing models or catalog discoveries",
			"Variables in model name and parameters are interpolated",
			"Additional provider-specific parameters can be included",
			"Created model ID and metadata are stored in system variables",
		},
	}
}

// Execute creates a new model configuration with the specified parameters.
// The input parameter is used as the model name (required).
// Creation methods: 1) provider+base_model, 2) from_id
// Optional parameters: temperature, max_tokens, top_p, top_k, presence_penalty, frequency_penalty, description
func (c *NewCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get model service
	modelService, err := c.getModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	// Get variable service for interpolation and storing result variables
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse model name
	modelName := input
	if modelName == "" {
		return fmt.Errorf("model name is required\\n\\nUsage: %s", c.Usage())
	}

	// Interpolate model name
	modelName, err = variableService.InterpolateString(modelName)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in model name: %w", err)
	}

	// Determine creation method
	fromID := args["from_id"]
	provider := args["provider"]
	baseModel := args["base_model"]

	// Check for conflicting options
	if fromID != "" && (provider != "" || baseModel != "") {
		return fmt.Errorf("cannot combine from_id with provider/base_model options\\n\\nUsage: %s", c.Usage())
	}

	// Route to appropriate creation method
	if fromID != "" {
		return c.executeFromExisting(args, modelName, fromID, variableService, modelService, ctx)
	}
	return c.executeFromProvider(args, modelName, provider, baseModel, variableService, modelService, ctx)
}

// executeFromExisting creates a model by cloning an existing model with optional parameter overrides.
func (c *NewCommand) executeFromExisting(args map[string]string, modelName, fromID string, variableService *services.VariableService, modelService *services.ModelService, ctx neurotypes.Context) error {
	// Interpolate from_id
	fromID, err := variableService.InterpolateString(fromID)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in from_id: %w", err)
	}

	// Get base model configuration
	baseModel, err := modelService.GetModel(fromID)
	if err != nil {
		return fmt.Errorf("base model with ID '%s' not found: %w", fromID, err)
	}

	// Create new model by copying base model
	newModel := *baseModel // Copy struct
	newModel.ID = c.generateModelID()
	newModel.Name = modelName
	newModel.CreatedAt = time.Now()

	// Update description to indicate cloning
	if newModel.Description == "" {
		newModel.Description = fmt.Sprintf("Cloned from model '%s'", baseModel.Name)
	} else {
		newModel.Description = fmt.Sprintf("Cloned from '%s': %s", baseModel.Name, newModel.Description)
	}

	// Apply parameter overrides
	if err := c.applyParameterOverrides(&newModel, args, variableService, ctx); err != nil {
		return fmt.Errorf("failed to apply parameter overrides: %w", err)
	}

	// Validate parameters
	if err := modelService.ValidateModelParameters(newModel.Parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Store model using CreateModel (which handles storage internally)
	createdModel, err := modelService.CreateModel(
		newModel.Name,
		newModel.Provider,
		newModel.BaseModel,
		newModel.Parameters,
		newModel.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to store model: %w", err)
	}

	// Update result variables
	if err := c.updateModelVariables(createdModel, variableService, ctx); err != nil {
		return fmt.Errorf("failed to update model variables: %w", err)
	}

	// Print confirmation
	fmt.Printf("Created model '%s' (ID: %s) from base model '%s'\n",
		createdModel.Name, createdModel.ID[:8], baseModel.Name)

	return nil
}

// executeFromProvider creates a model from provider and base model specifications.
func (c *NewCommand) executeFromProvider(args map[string]string, modelName, provider, baseModel string, variableService *services.VariableService, modelService *services.ModelService, ctx neurotypes.Context) error {
	// Validate required parameters
	if provider == "" {
		return fmt.Errorf("provider is required for provider-based creation\\n\\nUsage: %s", c.Usage())
	}

	if baseModel == "" {
		return fmt.Errorf("base_model is required for provider-based creation\\n\\nUsage: %s", c.Usage())
	}

	// Interpolate variables in string parameters
	var err error
	provider, err = variableService.InterpolateString(provider)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in provider: %w", err)
	}

	baseModel, err = variableService.InterpolateString(baseModel)
	if err != nil {
		return fmt.Errorf("failed to interpolate variables in base_model: %w", err)
	}

	// Parse optional parameters
	parameters := make(map[string]any)
	description := ""

	// Handle description separately
	if desc, exists := args["description"]; exists {
		description, err = variableService.InterpolateString(desc)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in description: %w", err)
		}
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
	model, err := modelService.CreateModel(modelName, provider, baseModel, parameters, description)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Update model-related variables
	if err := c.updateModelVariables(model, variableService, ctx); err != nil {
		return fmt.Errorf("failed to update model variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)",
		model.Name, model.ID[:8], model.Provider, model.BaseModel)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// generateModelID generates a unique ID for a new model.
func (c *NewCommand) generateModelID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// applyParameterOverrides applies parameter overrides from args to an existing model configuration.
func (c *NewCommand) applyParameterOverrides(model *neurotypes.ModelConfig, args map[string]string, variableService *services.VariableService, _ neurotypes.Context) error {
	// Handle description override
	if desc, exists := args["description"]; exists {
		interpolatedDesc, err := variableService.InterpolateString(desc)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in description: %w", err)
		}
		model.Description = interpolatedDesc
	}

	// Parse and apply parameter overrides
	parameterOverrides := make(map[string]any)
	if err := c.parseParameters(args, parameterOverrides); err != nil {
		return fmt.Errorf("failed to parse parameter overrides: %w", err)
	}

	// Apply overrides to existing parameters
	if model.Parameters == nil {
		model.Parameters = make(map[string]any)
	}

	for key, value := range parameterOverrides {
		model.Parameters[key] = value
	}

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
		"provider": true, "base_model": true, "description": true,
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
func (c *NewCommand) updateModelVariables(model *neurotypes.ModelConfig, variableService *services.VariableService, _ neurotypes.Context) error {
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

// getModelService retrieves the model service from the global registry.
func (c *NewCommand) getModelService() (*services.ModelService, error) {
	service, err := services.GetGlobalRegistry().GetService("model")
	if err != nil {
		return nil, err
	}

	modelService, ok := service.(*services.ModelService)
	if !ok {
		return nil, fmt.Errorf("model service has incorrect type")
	}

	return modelService, nil
}

// getVariableService retrieves the variable service from the global registry.
func (c *NewCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-new command: %v", err))
	}
}
