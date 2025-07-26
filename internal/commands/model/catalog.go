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
	return `\model-catalog[provider=openai|anthropic|openrouter|moonshot|all, sort=name|provider, search=query]

Examples:
  \model-catalog                              %% List all available models (default: sorted by provider)
  \model-catalog[provider=openai]             %% List OpenAI models only
  \model-catalog[provider=anthropic]          %% List Anthropic models only
  \model-catalog[provider=openrouter]         %% List OpenRouter models only
  \model-catalog[provider=moonshot]           %% List Moonshot models only
  \model-catalog[sort=name]                   %% Sort models alphabetically by name
  \model-catalog[search=gpt-4]                %% Search for models containing "gpt-4"
  \model-catalog[search=CS4]                  %% Search by model ID (case-insensitive)
  \model-catalog[search=kimi]                 %% Search for Kimi models
  \model-catalog[provider=openai,sort=name]   %% OpenAI models sorted by name
  \model-catalog[search=claude,sort=name]     %% Search for Claude models, sorted by name

Options:
  provider - Filter by provider: openai, anthropic, openrouter, moonshot, all (default: all)
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
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Formatted catalog listing of available models",
				Type:        "command_output",
				Example:     "[CS4] Claude Sonnet 4 (claude-3-sonnet-20240229)\\n[O3] GPT-4 (gpt-4)...",
			},
		},
		Notes: []string{
			"Options can be combined (e.g., provider=openai,sort=name)",
			"Default sort is by provider, then by name within each provider",
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

	// Get theme object for styling
	themeObj := c.getThemeObject()

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
	for _, supportedProvider := range []string{"anthropic", "openai", "openrouter", "moonshot"} {
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

	// Format output with theme styling
	output := c.formatModelCatalog(models, provider, sortBy, searchQuery, modelToProvider, themeObj)

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
		"all":        true,
		"openai":     true,
		"anthropic":  true,
		"openrouter": true,
		"moonshot":   true,
	}
	if !validProviders[provider] {
		return fmt.Errorf("invalid provider option '%s'. Valid options: all, openai, anthropic, openrouter, moonshot", provider)
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
	// Use the Provider field directly from the model entry
	if model.Provider != "" {
		return model.Provider
	}
	// Fallback to mapping for legacy compatibility
	if provider, exists := modelToProvider[model.Name]; exists {
		return provider
	}
	return "unknown"
}

// formatModelCatalog formats the model catalog for display with theme styling.
func (c *CatalogCommand) formatModelCatalog(models []neurotypes.ModelCatalogEntry, provider, sortBy, searchQuery string, modelToProvider map[string]string, themeObj *services.Theme) string {
	if len(models) == 0 {
		searchText := ""
		if searchQuery != "" {
			searchText = fmt.Sprintf(" matching %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		}
		providerText := ""
		if provider != "all" {
			providerText = fmt.Sprintf(" from %s", themeObj.Keyword.Render(provider))
		}
		noModelsMsg := fmt.Sprintf("No models found%s%s.", providerText, searchText)
		return themeObj.Warning.Render(noModelsMsg) + "\n"
	}

	var result strings.Builder

	// Header with professional styling
	headerParts := []string{themeObj.Success.Render("Model Catalog")}
	if provider != "all" {
		providerPart := fmt.Sprintf("(%s)", themeObj.Keyword.Render(c.toTitle(provider)))
		headerParts = append(headerParts, providerPart)
	}
	if searchQuery != "" {
		searchPart := fmt.Sprintf("- Search: %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		headerParts = append(headerParts, searchPart)
	}
	countPart := fmt.Sprintf("(%s)", themeObj.Info.Render(fmt.Sprintf("%d models", len(models))))
	headerParts = append(headerParts, countPart)

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
				providerHeader := fmt.Sprintf("%s Models:", c.toTitle(modelProvider))
				result.WriteString(themeObj.Success.Render(providerHeader) + "\n")
				currentProvider = modelProvider
			}
			result.WriteString(c.formatModelEntry(model, true, modelToProvider, themeObj))
		}
	} else {
		// Simple list format
		for _, model := range models {
			result.WriteString(c.formatModelEntry(model, provider == "all", modelToProvider, themeObj))
		}
	}

	return result.String()
}

// formatModelEntry formats a single model entry for display with professional theme styling.
func (c *CatalogCommand) formatModelEntry(model neurotypes.ModelCatalogEntry, showProvider bool, modelToProvider map[string]string, themeObj *services.Theme) string {
	var result strings.Builder

	// Model header: [ID] Display Name (technical_name) with prominent catalog ID
	catalogID := themeObj.Highlight.Render(fmt.Sprintf("[%s]", model.ID))
	displayName := themeObj.Command.Render(model.DisplayName)
	technicalName := themeObj.Variable.Render(fmt.Sprintf("(%s)", model.Name))
	modelHeader := fmt.Sprintf("  %s %s %s\n", catalogID, displayName, technicalName)
	result.WriteString(modelHeader)

	// Provider (if showing all providers)
	if showProvider {
		provider := c.getProviderFromModel(model, modelToProvider)
		providerLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Provider:"),
			themeObj.Keyword.Render(provider))
		result.WriteString(providerLine)
	}

	// Context window and max output tokens
	if model.ContextWindow > 0 {
		contextValue := c.formatNumber(model.ContextWindow)
		contextInfo := fmt.Sprintf("    %s %s tokens",
			themeObj.Info.Render("Context:"),
			themeObj.Variable.Render(contextValue))
		if model.MaxOutputTokens != nil && *model.MaxOutputTokens != model.ContextWindow {
			maxOutputValue := c.formatNumber(*model.MaxOutputTokens)
			contextInfo += fmt.Sprintf(" (max output: %s tokens)",
				themeObj.Variable.Render(maxOutputValue))
		}
		result.WriteString(contextInfo + "\n")
	}

	// Capabilities
	if len(model.Capabilities) > 0 {
		capabilitiesList := make([]string, len(model.Capabilities))
		for i, cap := range model.Capabilities {
			capabilitiesList[i] = themeObj.Keyword.Render(cap)
		}
		capabilitiesLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Capabilities:"),
			strings.Join(capabilitiesList, ", "))
		result.WriteString(capabilitiesLine)
	}

	// Modalities
	if len(model.Modalities) > 0 {
		modalitiesList := make([]string, len(model.Modalities))
		for i, mod := range model.Modalities {
			modalitiesList[i] = themeObj.Variable.Render(mod)
		}
		modalitiesLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Modalities:"),
			strings.Join(modalitiesList, ", "))
		result.WriteString(modalitiesLine)
	}

	// Pricing information
	if model.Pricing != nil {
		inputPrice := themeObj.Variable.Render(fmt.Sprintf("$%.2f/1M", model.Pricing.InputPerMToken))
		outputPrice := themeObj.Variable.Render(fmt.Sprintf("$%.2f/1M", model.Pricing.OutputPerMToken))
		pricingLine := fmt.Sprintf("    %s %s input, %s output tokens\n",
			themeObj.Info.Render("Pricing:"), inputPrice, outputPrice)
		result.WriteString(pricingLine)
	}

	// Features
	if model.Features != nil {
		var features []string
		if model.Features.Streaming != nil && *model.Features.Streaming {
			features = append(features, themeObj.Keyword.Render("streaming"))
		}
		if model.Features.FunctionCalling != nil && *model.Features.FunctionCalling {
			features = append(features, themeObj.Keyword.Render("function-calling"))
		}
		if model.Features.StructuredOutputs != nil && *model.Features.StructuredOutputs {
			features = append(features, themeObj.Keyword.Render("structured-outputs"))
		}
		if model.Features.Vision != nil && *model.Features.Vision {
			features = append(features, themeObj.Keyword.Render("vision"))
		}
		if model.Features.FineTuning != nil && *model.Features.FineTuning {
			features = append(features, themeObj.Keyword.Render("fine-tuning"))
		}
		if len(features) > 0 {
			featuresLine := fmt.Sprintf("    %s %s\n",
				themeObj.Info.Render("Features:"),
				strings.Join(features, ", "))
			result.WriteString(featuresLine)
		}
	}

	// Tools
	if len(model.Tools) > 0 {
		toolsList := make([]string, len(model.Tools))
		for i, tool := range model.Tools {
			toolsList[i] = themeObj.Variable.Render(tool)
		}
		toolsLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Tools:"),
			strings.Join(toolsList, ", "))
		result.WriteString(toolsLine)
	}

	// Knowledge cutoff
	if model.KnowledgeCutoff != nil {
		cutoffLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Knowledge cutoff:"),
			themeObj.Variable.Render(*model.KnowledgeCutoff))
		result.WriteString(cutoffLine)
	}

	// Reasoning tokens
	if model.ReasoningTokens != nil && *model.ReasoningTokens {
		reasoningLine := fmt.Sprintf("    %s\n",
			themeObj.Success.Render("Reasoning tokens supported"))
		result.WriteString(reasoningLine)
	}

	// Snapshots
	if len(model.Snapshots) > 0 {
		snapshotsList := make([]string, len(model.Snapshots))
		for i, snapshot := range model.Snapshots {
			snapshotsList[i] = themeObj.Variable.Render(snapshot)
		}
		snapshotsLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Snapshots:"),
			strings.Join(snapshotsList, ", "))
		result.WriteString(snapshotsLine)
	}

	// Deprecation status
	if model.Deprecated {
		deprecatedLine := fmt.Sprintf("    %s\n",
			themeObj.Warning.Render("DEPRECATED"))
		result.WriteString(deprecatedLine)
	}

	// Description
	if len(model.Description) > 0 {
		descriptionLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Description:"),
			themeObj.Info.Render(model.Description))
		result.WriteString(descriptionLine)
	}

	return result.String()
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

// getThemeObject retrieves the theme object based on the _style variable
func (c *CatalogCommand) getThemeObject() *services.Theme {
	// Get _style variable for theme selection
	styleValue := ""
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		if value, err := variableService.Get("_style"); err == nil {
			styleValue = value
		}
	}

	// Get theme service and theme object (always returns valid theme)
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// This should rarely happen, but we need to return something
		panic(fmt.Sprintf("theme service not available: %v", err))
	}

	return themeService.GetThemeByName(styleValue)
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CatalogCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-catalog command: %v", err))
	}
}
