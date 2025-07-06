// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CatalogCommand implements the \model-catalog command for listing available LLM models.
// It provides access to the embedded model catalog with filtering and search capabilities.
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
	return "List available LLM models from embedded catalog"
}

// Usage returns the syntax and usage examples for the model-catalog command.
func (c *CatalogCommand) Usage() string {
	return `\model-catalog[provider=openai|anthropic|all, sort=name|provider, search=query]

Examples:
  \model-catalog                              %% List all available models (default: sorted by provider)
  \model-catalog[provider=openai]             %% List OpenAI models only
  \model-catalog[provider=anthropic]          %% List Anthropic models only
  \model-catalog[sort=name]                   %% Sort models alphabetically by name
  \model-catalog[search=gpt-4]                %% Search for models containing "gpt-4"
  \model-catalog[search=CS4]                  %% Search by model ID (case-insensitive)
  \model-catalog[provider=openai,sort=name]   %% OpenAI models sorted by name
  \model-catalog[search=claude,sort=name]     %% Search for Claude models, sorted by name
  
Options:
  provider - Filter by provider: openai, anthropic, all (default: all)
  sort     - Sort order: name (alphabetical), provider (by provider then name)
  search   - Search query to filter models by ID, name, display name, or description
  
Note: Options can be combined. Default sort is by provider.
      Model catalog is stored in ${_output} variable.
      Shows model ID, display name, provider, capabilities, and context window.
      Model IDs are shown in format: [ID] Display Name (model_name)`
}

// HelpInfo returns structured help information for the model-catalog command.
func (c *CatalogCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-catalog[provider=openai|anthropic|all, sort=name|provider, search=query]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "Filter by provider: openai, anthropic, all",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
			{
				Name:        "sort",
				Description: "Sort order: name (alphabetical), provider (by provider then name)",
				Required:    false,
				Type:        "string",
				Default:     "provider",
			},
			{
				Name:        "search",
				Description: "Search query to filter models by ID, name, display name, or description",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-catalog",
				Description: "List all available models sorted by provider",
			},
			{
				Command:     "\\model-catalog[provider=openai]",
				Description: "List OpenAI models only",
			},
			{
				Command:     "\\model-catalog[search=gpt-4]",
				Description: "Search for models containing 'gpt-4'",
			},
			{
				Command:     "\\model-catalog[search=CS4]",
				Description: "Search by model ID (case-insensitive)",
			},
			{
				Command:     "\\model-catalog[provider=anthropic,sort=name]",
				Description: "List Anthropic models sorted alphabetically",
			},
		},
		Notes: []string{
			"Options can be combined (e.g., provider=openai,sort=name)",
			"Default sort is by provider, then by name within each provider",
			"Model catalog output is stored in ${_output} variable",
			"Shows model ID, display name, provider, capabilities, context window, and deprecation status",
			"Model IDs are displayed in format: [ID] Display Name (model_name)",
			"Embedded catalog includes popular models from OpenAI and Anthropic",
			"Search is case-insensitive and matches ID, name, display name, or description",
			"Model IDs can be used with \\model-new[catalog_id=<ID>] for easy model creation",
		},
	}
}

// Execute lists available LLM models with optional filtering, sorting, and searching.
// Options:
//   - provider: openai|anthropic|all (default: all)
//   - sort: name|provider (default: provider)
//   - search: query string for filtering (optional)
func (c *CatalogCommand) Execute(args map[string]string, _ string) error {
	// Get model catalog service
	modelCatalogService, err := services.GetGlobalModelCatalogService()
	if err != nil {
		return fmt.Errorf("model catalog service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse arguments
	provider := args["provider"]
	if provider == "" {
		provider = "all" // default provider filter
	}
	sortBy := args["sort"]
	if sortBy == "" {
		sortBy = "provider" // default sort
	}
	searchQuery := args["search"]

	// Validate arguments
	if err := c.validateArguments(provider, sortBy); err != nil {
		return err
	}

	// Get models based on provider filter
	var models []neurotypes.ModelCatalogEntry
	if provider == "all" {
		models, err = modelCatalogService.GetModelCatalog()
	} else {
		models, err = modelCatalogService.GetModelCatalogByProvider(provider)
	}
	if err != nil {
		return fmt.Errorf("failed to get model catalog: %w", err)
	}

	// Build model-to-provider mapping
	modelToProvider := make(map[string]string)
	for _, supportedProvider := range []string{"anthropic", "openai"} {
		providerModels, err := modelCatalogService.GetModelCatalogByProvider(supportedProvider)
		if err != nil {
			return fmt.Errorf("failed to get %s models for mapping: %w", supportedProvider, err)
		}
		for _, model := range providerModels {
			modelToProvider[model.Name] = supportedProvider
		}
	}

	// Apply search filter if provided
	if searchQuery != "" {
		models, err = c.filterModelsBySearch(models, searchQuery)
		if err != nil {
			return fmt.Errorf("failed to search models: %w", err)
		}
	}

	// Apply sorting
	c.sortModels(models, sortBy, provider, modelToProvider)

	// Format output
	output := c.formatModelCatalog(models, provider, sortBy, searchQuery, modelToProvider)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", output); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print the catalog
	fmt.Print(output)

	return nil
}

// validateArguments checks if the provided provider and sort options are valid.
func (c *CatalogCommand) validateArguments(provider, sortBy string) error {
	validProviders := map[string]bool{
		"all":       true,
		"openai":    true,
		"anthropic": true,
	}
	if !validProviders[provider] {
		return fmt.Errorf("invalid provider option '%s'. Valid options: all, openai, anthropic", provider)
	}

	validSorts := map[string]bool{
		"name":     true,
		"provider": true,
	}
	if !validSorts[sortBy] {
		return fmt.Errorf("invalid sort option '%s'. Valid options: name, provider", sortBy)
	}

	return nil
}

// filterModelsBySearch filters models based on the search query.
func (c *CatalogCommand) filterModelsBySearch(models []neurotypes.ModelCatalogEntry, query string) ([]neurotypes.ModelCatalogEntry, error) {
	var matches []neurotypes.ModelCatalogEntry
	queryLower := strings.ToLower(query)

	for _, model := range models {
		// Search in model ID, name, display name, and description
		if strings.Contains(strings.ToLower(model.ID), queryLower) ||
			strings.Contains(strings.ToLower(model.Name), queryLower) ||
			strings.Contains(strings.ToLower(model.DisplayName), queryLower) ||
			strings.Contains(strings.ToLower(model.Description), queryLower) {
			matches = append(matches, model)
		}
	}

	return matches, nil
}

// sortModels sorts the model list according to the specified criteria.
func (c *CatalogCommand) sortModels(models []neurotypes.ModelCatalogEntry, sortBy, provider string, modelToProvider map[string]string) {
	switch sortBy {
	case "name":
		sort.Slice(models, func(i, j int) bool {
			return strings.ToLower(models[i].DisplayName) < strings.ToLower(models[j].DisplayName)
		})
	case "provider":
		sort.Slice(models, func(i, j int) bool {
			// First sort by provider (if showing all providers)
			if provider == "all" {
				providerI := c.getProviderFromModel(models[i], modelToProvider)
				providerJ := c.getProviderFromModel(models[j], modelToProvider)
				if providerI != providerJ {
					return providerI < providerJ
				}
			}
			// Then sort by display name within provider
			return strings.ToLower(models[i].DisplayName) < strings.ToLower(models[j].DisplayName)
		})
	}
}

// getProviderFromModel determines the provider for a model entry using the model catalog data.
func (c *CatalogCommand) getProviderFromModel(model neurotypes.ModelCatalogEntry, modelToProvider map[string]string) string {
	if provider, exists := modelToProvider[model.Name]; exists {
		return provider
	}
	return "unknown"
}

// formatModelCatalog formats the model catalog for display.
func (c *CatalogCommand) formatModelCatalog(models []neurotypes.ModelCatalogEntry, provider, sortBy, searchQuery string, modelToProvider map[string]string) string {
	if len(models) == 0 {
		searchText := ""
		if searchQuery != "" {
			searchText = fmt.Sprintf(" matching '%s'", searchQuery)
		}
		providerText := ""
		if provider != "all" {
			providerText = fmt.Sprintf(" from %s", provider)
		}
		return fmt.Sprintf("No models found%s%s.\n", providerText, searchText)
	}

	var result strings.Builder

	// Header
	headerParts := []string{"Model Catalog"}
	if provider != "all" {
		headerParts = append(headerParts, fmt.Sprintf("(%s)", c.toTitle(provider)))
	}
	if searchQuery != "" {
		headerParts = append(headerParts, fmt.Sprintf("- Search: '%s'", searchQuery))
	}
	headerParts = append(headerParts, fmt.Sprintf("(%d models)", len(models)))

	result.WriteString(fmt.Sprintf("%s:\n", strings.Join(headerParts, " ")))

	// Group by provider if showing all providers
	if provider == "all" && sortBy == "provider" {
		currentProvider := ""
		for _, model := range models {
			modelProvider := c.getProviderFromModel(model, modelToProvider)
			if modelProvider != currentProvider {
				if currentProvider != "" {
					result.WriteString("\n")
				}
				result.WriteString(fmt.Sprintf("%s Models:\n", c.toTitle(modelProvider)))
				currentProvider = modelProvider
			}
			result.WriteString(c.formatModelEntry(model, true, modelToProvider))
		}
	} else {
		// Simple list format
		for _, model := range models {
			result.WriteString(c.formatModelEntry(model, provider == "all", modelToProvider))
		}
	}

	return result.String()
}

// formatModelEntry formats a single model entry for display.
func (c *CatalogCommand) formatModelEntry(model neurotypes.ModelCatalogEntry, showProvider bool, modelToProvider map[string]string) string {
	var parts []string

	// Model ID, display name and name
	modelInfo := fmt.Sprintf("[%s] %s (%s)", model.ID, model.DisplayName, model.Name)
	parts = append(parts, modelInfo)

	// Provider (if showing all providers)
	if showProvider {
		provider := c.getProviderFromModel(model, modelToProvider)
		parts = append(parts, fmt.Sprintf("Provider: %s", provider))
	}

	// Context window and max output tokens
	if model.ContextWindow > 0 {
		contextInfo := fmt.Sprintf("Context: %s tokens", c.formatNumber(model.ContextWindow))
		if model.MaxOutputTokens != nil && *model.MaxOutputTokens != model.ContextWindow {
			contextInfo += fmt.Sprintf(" (max output: %s tokens)", c.formatNumber(*model.MaxOutputTokens))
		}
		parts = append(parts, contextInfo)
	}

	// Capabilities
	if len(model.Capabilities) > 0 {
		capabilitiesInfo := fmt.Sprintf("Capabilities: %s", strings.Join(model.Capabilities, ", "))
		parts = append(parts, capabilitiesInfo)
	}

	// Modalities
	if len(model.Modalities) > 0 {
		modalitiesInfo := fmt.Sprintf("Modalities: %s", strings.Join(model.Modalities, ", "))
		parts = append(parts, modalitiesInfo)
	}

	// Pricing information
	if model.Pricing != nil {
		pricingInfo := fmt.Sprintf("Pricing: $%.2f/1M input, $%.2f/1M output tokens",
			model.Pricing.InputPerMToken, model.Pricing.OutputPerMToken)
		parts = append(parts, pricingInfo)
	}

	// Features
	if model.Features != nil {
		var features []string
		if model.Features.Streaming != nil && *model.Features.Streaming {
			features = append(features, "streaming")
		}
		if model.Features.FunctionCalling != nil && *model.Features.FunctionCalling {
			features = append(features, "function-calling")
		}
		if model.Features.StructuredOutputs != nil && *model.Features.StructuredOutputs {
			features = append(features, "structured-outputs")
		}
		if model.Features.Vision != nil && *model.Features.Vision {
			features = append(features, "vision")
		}
		if model.Features.FineTuning != nil && *model.Features.FineTuning {
			features = append(features, "fine-tuning")
		}
		if len(features) > 0 {
			featuresInfo := fmt.Sprintf("Features: %s", strings.Join(features, ", "))
			parts = append(parts, featuresInfo)
		}
	}

	// Tools
	if len(model.Tools) > 0 {
		toolsInfo := fmt.Sprintf("Tools: %s", strings.Join(model.Tools, ", "))
		parts = append(parts, toolsInfo)
	}

	// Knowledge cutoff
	if model.KnowledgeCutoff != nil {
		parts = append(parts, fmt.Sprintf("Knowledge cutoff: %s", *model.KnowledgeCutoff))
	}

	// Reasoning tokens
	if model.ReasoningTokens != nil && *model.ReasoningTokens {
		parts = append(parts, "Reasoning tokens supported")
	}

	// Snapshots
	if len(model.Snapshots) > 0 {
		snapshotsInfo := fmt.Sprintf("Snapshots: %s", strings.Join(model.Snapshots, ", "))
		parts = append(parts, snapshotsInfo)
	}

	// Deprecation status
	if model.Deprecated {
		parts = append(parts, "DEPRECATED")
	}

	// Format as indented entry
	firstLine := fmt.Sprintf("  %s\n", parts[0])
	result := firstLine
	for _, part := range parts[1:] {
		result += fmt.Sprintf("    %s\n", part)
	}
	if len(model.Description) > 0 {
		result += fmt.Sprintf("    Description: %s\n", model.Description)
	}

	return result
}

// formatNumber formats large numbers with commas for readability.
func (c *CatalogCommand) formatNumber(num int) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	}
	if num < 1000000 {
		return fmt.Sprintf("%d,%.3d", num/1000, num%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", num/1000000, (num%1000000)/1000, num%1000)
}

// toTitle converts the first character of a string to uppercase.
func (c *CatalogCommand) toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func init() {
	if err := commands.GlobalRegistry.Register(&CatalogCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-catalog command: %v", err))
	}
}
