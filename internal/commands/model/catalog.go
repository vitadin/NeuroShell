// Package model provides model management commands for NeuroShell.
// This file implements the \model-catalog command for discovering and searching LLM models.
package model

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CatalogCommand implements the \model-catalog command for discovering available LLM models.
// It provides model search, filtering, and auto-creation functionality for seamless integration with \model-new.
type CatalogCommand struct{}

// Name returns the command name "model-catalog" for registration and lookup.
func (c *CatalogCommand) Name() string {
	return "model-catalog"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *CatalogCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-catalog command does.
func (c *CatalogCommand) Description() string {
	return "Discover and search available LLM models from various providers"
}

// Usage returns the syntax and usage examples for the model-catalog command.
func (c *CatalogCommand) Usage() string {
	return `\model-catalog[provider=name, pattern=search, sort=field]

Examples:
  \model-catalog                           %% List all available models
  \model-catalog[provider=openai]          %% Show only OpenAI models  
  \model-catalog[pattern=gpt]              %% Search for models containing "gpt"
  \model-catalog[provider=anthropic, sort=context_length] %% Anthropic models by context length
  \model-catalog[pattern=gpt-4-turbo]      %% Find specific model (auto-creates if single result)
  
Options:
  provider - Filter by provider (openai, anthropic, local)
  pattern  - Search pattern for model names/descriptions (supports fuzzy matching)
  sort     - Sort field: name, context_length, pricing_tier, provider, release_date
  
Auto-Creation:
  When exactly one model matches your search, it's automatically created in your model registry
  with default parameters. Use the resulting model immediately or customize with \model-new[from_id=...].
  
Result Variables:
  _catalog_count     - Number of matching models
  _catalog_provider  - Provider (if all results from same provider)  
  _catalog_model_id  - Model ID (if exactly one result, references auto-created model)
  _catalog_models    - Comma-separated list of all matching model IDs`
}

// HelpInfo returns structured help information for the model-catalog command.
func (c *CatalogCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-catalog[provider=name, pattern=search, sort=field]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "Filter by provider (e.g., openai, anthropic, local)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "pattern",
				Description: "Search pattern for model names/descriptions",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "sort",
				Description: "Sort field: name, context_length, pricing_tier, provider, release_date",
				Required:    false,
				Type:        "string",
				Default:     "provider",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-catalog",
				Description: "List all available models from all providers",
			},
			{
				Command:     "\\model-catalog[provider=openai]",
				Description: "Show only OpenAI models",
			},
			{
				Command:     "\\model-catalog[pattern=gpt-4]",
				Description: "Search for GPT-4 variants (auto-creates if single match)",
			},
			{
				Command:     "\\model-catalog[provider=anthropic, sort=context_length]",
				Description: "Show Anthropic models sorted by context length",
			},
		},
		Notes: []string{
			"Single search results are automatically created as models for immediate use",
			"Use pattern search to discover models and avoid typos in model names",
			"Sort options help explore models by different characteristics",
			"Result variables enable seamless integration with \\model-new command",
			"Auto-created models use catalog default parameters but can be customized",
		},
	}
}

// Execute searches the model catalog and optionally auto-creates models for single results.
func (c *CatalogCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get required services
	catalogService, err := c.getCatalogService()
	if err != nil {
		return fmt.Errorf("catalog service not available: %w", err)
	}

	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	modelService, err := c.getModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	// Parse search options
	searchOptions := neurotypes.CatalogSearchOptions{
		Provider: args["provider"],
		Pattern:  args["pattern"],
		Sort:     args["sort"],
	}

	// If input is provided, use it as pattern (allows both \model-catalog[pattern=x] and \model-catalog x)
	if input != "" {
		if searchOptions.Pattern != "" {
			return fmt.Errorf("cannot specify pattern both as option and input parameter")
		}
		searchOptions.Pattern = input
	}

	// Interpolate variables in search options
	if searchOptions.Provider != "" {
		searchOptions.Provider, err = variableService.InterpolateString(searchOptions.Provider)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in provider: %w", err)
		}
	}

	if searchOptions.Pattern != "" {
		searchOptions.Pattern, err = variableService.InterpolateString(searchOptions.Pattern)
		if err != nil {
			return fmt.Errorf("failed to interpolate variables in pattern: %w", err)
		}
	}

	// Perform search
	result, err := catalogService.SearchModels(searchOptions)
	if err != nil {
		return fmt.Errorf("failed to search catalog: %w", err)
	}

	// Display results
	if err := c.displayResults(result, searchOptions); err != nil {
		return fmt.Errorf("failed to display results: %w", err)
	}

	// Set result variables
	if err := c.setResultVariables(result, variableService, ctx); err != nil {
		return fmt.Errorf("failed to set result variables: %w", err)
	}

	// Auto-create model if exactly one result
	if result.Count == 1 {
		autoModelID, err := c.autoCreateModel(result.Models[0], searchOptions.Pattern, catalogService, modelService, ctx)
		if err != nil {
			return fmt.Errorf("failed to auto-create model: %w", err)
		}

		// Update catalog_model_id variable to point to auto-created model
		if err := variableService.SetSystemVariable("_catalog_model_id", autoModelID); err != nil {
			return fmt.Errorf("failed to set catalog model ID variable: %w", err)
		}
	}

	return nil
}

// displayResults outputs the search results in a user-friendly format.
func (c *CatalogCommand) displayResults(result *neurotypes.CatalogSearchResult, options neurotypes.CatalogSearchOptions) error {
	if result.Count == 0 {
		fmt.Printf("No models found")
		if options.Provider != "" {
			fmt.Printf(" for provider '%s'", options.Provider)
		}
		if options.Pattern != "" {
			fmt.Printf(" matching pattern '%s'", options.Pattern)
		}
		fmt.Println()
		return nil
	}

	// Header
	if result.Count == 1 {
		fmt.Printf("Found 1 model")
	} else {
		fmt.Printf("Found %d models", result.Count)
	}

	if options.Provider != "" {
		fmt.Printf(" from %s", options.Provider)
	}
	if options.Pattern != "" {
		fmt.Printf(" matching '%s'", options.Pattern)
	}
	fmt.Println(":")
	fmt.Println()

	// List models
	for _, model := range result.Models {
		// Format: "provider/model-id - Model Name - Description"
		fmt.Printf("  %s/%s - %s", model.Provider, model.ID, model.Name)

		// Add key details
		details := []string{}
		if model.ContextLength > 0 {
			if model.ContextLength >= 1000 {
				details = append(details, fmt.Sprintf("%.0fK context", float64(model.ContextLength)/1000))
			} else {
				details = append(details, fmt.Sprintf("%d context", model.ContextLength))
			}
		}
		if model.PricingTier != "" {
			details = append(details, model.PricingTier)
		}

		if len(details) > 0 {
			fmt.Printf(" (%s)", strings.Join(details, ", "))
		}
		fmt.Println()

		// Show description if not too long
		if len(model.Description) > 0 && len(model.Description) <= 80 {
			fmt.Printf("    %s\n", model.Description)
		}
	}

	// Auto-creation notice for single results
	if result.Count == 1 {
		model := result.Models[0]
		// Generate the same name that will be used in auto-creation
		catalogService, _ := c.getCatalogService()
		autoName := catalogService.GenerateAutoModelName(options.Pattern, model)
		fmt.Println()
		fmt.Printf("Auto-created model '%s' with default parameters.\n", autoName)
		fmt.Printf("Use: \\session-new[model=%s] or \\model-new[from_id=${_catalog_model_id}]\n", autoName)
	}

	fmt.Println()
	fmt.Printf("Search completed in %v\n", result.QueryTime)

	return nil
}

// setResultVariables sets the standard result variables for scripting integration.
func (c *CatalogCommand) setResultVariables(result *neurotypes.CatalogSearchResult, variableService *services.VariableService, _ neurotypes.Context) error {
	// Always set count and models list
	if err := variableService.SetSystemVariable("_catalog_count", strconv.Itoa(result.Count)); err != nil {
		return err
	}

	modelIDs := make([]string, len(result.Models))
	for i, model := range result.Models {
		modelIDs[i] = model.ID
	}
	if err := variableService.SetSystemVariable("_catalog_models", strings.Join(modelIDs, ",")); err != nil {
		return err
	}

	// Set provider if all results from same provider
	if result.Provider != "" {
		if err := variableService.SetSystemVariable("_catalog_provider", result.Provider); err != nil {
			return err
		}
	}

	// Set model ID if exactly one result (will be updated after auto-creation)
	if result.ModelID != "" {
		if err := variableService.SetSystemVariable("_catalog_model_id", result.ModelID); err != nil {
			return err
		}
	}

	return nil
}

// autoCreateModel creates a model configuration from a catalog entry for immediate use.
func (c *CatalogCommand) autoCreateModel(catalogModel neurotypes.CatalogModel, searchPattern string, catalogService *services.CatalogService, modelService *services.ModelService, _ neurotypes.Context) (string, error) {
	// Generate model name
	modelName := catalogService.GenerateAutoModelName(searchPattern, catalogModel)

	// Check if model with this name already exists
	if _, err := modelService.GetModelByName(modelName); err == nil {
		// Model already exists, return its ID
		existingModel, _ := modelService.GetModelByName(modelName)
		return existingModel.ID, nil
	}

	// Create new model from catalog entry
	modelConfig, err := catalogService.AutoCreateModelFromCatalog(catalogModel, modelName)
	if err != nil {
		return "", fmt.Errorf("failed to create model config from catalog: %w", err)
	}

	// Store in model registry using CreateModel (which handles storage internally)
	createdModel, err := modelService.CreateModel(
		modelConfig.Name,
		modelConfig.Provider,
		modelConfig.BaseModel,
		modelConfig.Parameters,
		modelConfig.Description,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store auto-created model: %w", err)
	}

	return createdModel.ID, nil
}

// getCatalogService retrieves the catalog service from the global registry.
func (c *CatalogCommand) getCatalogService() (*services.CatalogService, error) {
	service, err := services.GetGlobalRegistry().GetService("catalog")
	if err != nil {
		return nil, err
	}

	catalogService, ok := service.(*services.CatalogService)
	if !ok {
		return nil, fmt.Errorf("catalog service has incorrect type")
	}

	return catalogService, nil
}

// getVariableService retrieves the variable service from the global registry.
func (c *CatalogCommand) getVariableService() (*services.VariableService, error) {
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

// getModelService retrieves the model service from the global registry.
func (c *CatalogCommand) getModelService() (*services.ModelService, error) {
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

func init() {
	if err := commands.GlobalRegistry.Register(&CatalogCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-catalog command: %v", err))
	}
}
