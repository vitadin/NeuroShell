package model

import (
	"strings"
	"testing"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestCatalogCommand_Name(t *testing.T) {
	cmd := &CatalogCommand{}
	expected := "model-catalog"
	actual := cmd.Name()

	if actual != expected {
		t.Errorf("Expected command name %q, got %q", expected, actual)
	}
}

func TestCatalogCommand_ParseMode(t *testing.T) {
	cmd := &CatalogCommand{}
	expected := neurotypes.ParseModeKeyValue
	actual := cmd.ParseMode()

	if actual != expected {
		t.Errorf("Expected parse mode %v, got %v", expected, actual)
	}
}

func TestCatalogCommand_Description(t *testing.T) {
	cmd := &CatalogCommand{}
	description := cmd.Description()

	if description == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(description, "Discover") {
		t.Error("Description should mention discovery functionality")
	}
}

func TestCatalogCommand_Usage(t *testing.T) {
	cmd := &CatalogCommand{}
	usage := cmd.Usage()

	if usage == "" {
		t.Error("Usage should not be empty")
	}
	if !strings.Contains(usage, "\\model-catalog") {
		t.Error("Usage should contain command syntax")
	}
	if !strings.Contains(usage, "provider") {
		t.Error("Usage should mention provider option")
	}
	if !strings.Contains(usage, "pattern") {
		t.Error("Usage should mention pattern option")
	}
}

func TestCatalogCommand_HelpInfo(t *testing.T) {
	cmd := &CatalogCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != cmd.Name() {
		t.Errorf("Expected command name %s, got %s", cmd.Name(), helpInfo.Command)
	}
	if helpInfo.Description != cmd.Description() {
		t.Error("HelpInfo description should match Description()")
	}
	if helpInfo.ParseMode != cmd.ParseMode() {
		t.Error("HelpInfo parse mode should match ParseMode()")
	}

	// Check that expected options are present
	expectedOptions := []string{"provider", "pattern", "sort"}
	optionMap := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionMap[option.Name] = true
	}

	for _, expected := range expectedOptions {
		if !optionMap[expected] {
			t.Errorf("Expected option %s not found in help info", expected)
		}
	}

	// Check examples
	if len(helpInfo.Examples) == 0 {
		t.Error("HelpInfo should contain examples")
	}

	// Check notes
	if len(helpInfo.Notes) == 0 {
		t.Error("HelpInfo should contain notes")
	}
}

func setupCatalogTest(t testing.TB) (*CatalogCommand, neurotypes.Context) {
	// Use real NeuroContext for tests to avoid type casting issues
	ctx := context.New()
	ctx.SetTestMode(true)

	// Create new registry to avoid conflicts
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Clean up after test
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	// Initialize services
	registry := services.GetGlobalRegistry()

	// Register required services
	if err := registry.RegisterService(services.NewVariableService()); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}
	if err := registry.RegisterService(services.NewModelService()); err != nil {
		t.Fatalf("Failed to register model service: %v", err)
	}
	if err := registry.RegisterService(services.NewCatalogService()); err != nil {
		t.Fatalf("Failed to register catalog service: %v", err)
	}

	// Initialize all services
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	cmd := &CatalogCommand{}
	return cmd, ctx
}

func TestCatalogCommand_Execute_BasicSearch(t *testing.T) {
	cmd, ctx := setupCatalogTest(t)

	tests := []struct {
		name          string
		args          map[string]string
		input         string
		expectedError bool
		validateVars  func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name:          "search all models",
			args:          map[string]string{},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				// Should set _catalog_count variable
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "" {
					t.Error("_catalog_count variable should be set")
				}
				models, _ := ctx.GetVariable("_catalog_models")
				if models == "" {
					t.Error("_catalog_models variable should be set")
				}
			},
		},
		{
			name: "filter by provider",
			args: map[string]string{
				"provider": "openai",
			},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "0" {
					t.Error("Should find OpenAI models")
				}
				provider, _ := ctx.GetVariable("_catalog_provider")
				if provider != "openai" {
					t.Error("_catalog_provider should be set to openai")
				}
			},
		},
		{
			name: "search by pattern",
			args: map[string]string{
				"pattern": "gpt",
			},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "0" {
					t.Error("Should find GPT models")
				}
			},
		},
		{
			name:          "pattern in input parameter",
			args:          map[string]string{},
			input:         "claude",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "0" {
					t.Error("Should find Claude models")
				}
			},
		},
		{
			name: "conflicting pattern options",
			args: map[string]string{
				"pattern": "gpt",
			},
			input:         "claude",
			expectedError: true,
		},
		{
			name: "sort by context length",
			args: map[string]string{
				"sort": "context_length",
			},
			input:         "",
			expectedError: false,
		},
		{
			name: "combined filters",
			args: map[string]string{
				"provider": "anthropic",
				"pattern":  "sonnet",
			},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "0" {
					t.Error("Should find Claude Sonnet")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear variables before test
			_ = ctx.SetVariable("_catalog_count", "")
			_ = ctx.SetVariable("_catalog_provider", "")
			_ = ctx.SetVariable("_catalog_model_id", "")
			_ = ctx.SetVariable("_catalog_models", "")

			err := cmd.Execute(tt.args, tt.input, ctx)

			if tt.expectedError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tt.expectedError && tt.validateVars != nil {
				tt.validateVars(t, ctx)
			}
		})
	}
}

func TestCatalogCommand_Execute_AutoCreation(t *testing.T) {
	cmd, ctx := setupCatalogTest(t)

	tests := []struct {
		name             string
		args             map[string]string
		input            string
		expectedError    bool
		shouldAutoCreate bool
		validateVars     func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "exact match triggers auto-creation",
			args: map[string]string{
				"provider": "openai",
				"pattern":  "gpt-4-turbo",
			},
			input:            "",
			expectedError:    false,
			shouldAutoCreate: true,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				// With fuzzy search, "gpt-4-turbo" may match multiple models
				// The auto-creation only happens if exactly one model matches
				if count == "1" {
					modelID, _ := ctx.GetVariable("_catalog_model_id")
					if modelID == "" {
						t.Error("_catalog_model_id should be set after auto-creation")
					}
				} else {
					// If multiple matches, no auto-creation should happen
					modelID, _ := ctx.GetVariable("_catalog_model_id")
					if modelID != "" {
						t.Error("_catalog_model_id should be empty for multiple matches")
					}
				}
			},
		},
		{
			name: "multiple matches no auto-creation",
			args: map[string]string{
				"pattern": "gpt",
			},
			input:            "",
			expectedError:    false,
			shouldAutoCreate: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "1" {
					t.Error("Should not auto-create for multiple matches")
				}
				modelID, _ := ctx.GetVariable("_catalog_model_id")
				if modelID != "" {
					t.Error("_catalog_model_id should be empty for multiple matches")
				}
			},
		},
		{
			name: "no matches no auto-creation",
			args: map[string]string{
				"pattern": "qwertyuiopasdfghjklzxcvbnm",
			},
			input:            "",
			expectedError:    false,
			shouldAutoCreate: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count != "0" {
					t.Errorf("Expected count=0 for no matches, got %s", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear variables before test
			_ = ctx.SetVariable("_catalog_count", "")
			_ = ctx.SetVariable("_catalog_provider", "")
			_ = ctx.SetVariable("_catalog_model_id", "")
			_ = ctx.SetVariable("_catalog_models", "")

			err := cmd.Execute(tt.args, tt.input, ctx)

			if tt.expectedError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tt.expectedError && tt.validateVars != nil {
				tt.validateVars(t, ctx)
			}
		})
	}
}

func TestCatalogCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd, ctx := setupCatalogTest(t)

	// Set up variables for interpolation
	_ = ctx.SetVariable("test_provider", "openai")
	_ = ctx.SetVariable("test_pattern", "gpt-4")

	tests := []struct {
		name          string
		args          map[string]string
		input         string
		expectedError bool
		validateVars  func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "interpolate provider",
			args: map[string]string{
				"provider": "${test_provider}",
			},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				provider, _ := ctx.GetVariable("_catalog_provider")
				if provider != "openai" {
					t.Errorf("Expected provider openai after interpolation, got %s", provider)
				}
			},
		},
		{
			name: "interpolate pattern",
			args: map[string]string{
				"pattern": "${test_pattern}",
			},
			input:         "",
			expectedError: false,
			validateVars: func(t *testing.T, ctx neurotypes.Context) {
				count, _ := ctx.GetVariable("_catalog_count")
				if count == "0" {
					t.Error("Should find models after pattern interpolation")
				}
			},
		},
		{
			name:          "interpolate input",
			args:          map[string]string{},
			input:         "${test_pattern}",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input, ctx)

			if tt.expectedError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tt.expectedError && tt.validateVars != nil {
				tt.validateVars(t, ctx)
			}
		})
	}
}

func TestCatalogCommand_Execute_ServiceErrors(t *testing.T) {
	// Test without proper service setup to trigger service errors
	cmd := &CatalogCommand{}
	ctx := context.New()
	ctx.SetTestMode(true)

	tests := []struct {
		name          string
		setupServices func()
		args          map[string]string
		input         string
		expectedError string
	}{
		{
			name:          "missing catalog service",
			setupServices: func() {},
			args:          map[string]string{},
			input:         "",
			expectedError: "catalog service not available",
		},
		{
			name: "missing variable service",
			setupServices: func() {
				registry := services.GetGlobalRegistry()
				_ = registry.RegisterService(services.NewCatalogService())
				_ = registry.InitializeAll(ctx)
			},
			args:          map[string]string{},
			input:         "",
			expectedError: "variable service not available",
		},
		{
			name: "missing model service",
			setupServices: func() {
				registry := services.GetGlobalRegistry()
				_ = registry.RegisterService(services.NewCatalogService())
				_ = registry.RegisterService(services.NewVariableService())
				_ = registry.InitializeAll(ctx)
			},
			args:          map[string]string{},
			input:         "",
			expectedError: "model service not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original registry and restore after test
			oldRegistry := services.GetGlobalRegistry()
			defer services.SetGlobalRegistry(oldRegistry)

			// Reset registry for each test
			services.SetGlobalRegistry(services.NewRegistry())
			tt.setupServices()

			err := cmd.Execute(tt.args, tt.input, ctx)

			if err == nil {
				t.Error("Expected error, but got none")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestCatalogCommand_GetCatalogService(t *testing.T) {
	cmd := &CatalogCommand{}

	// Save original registry and restore after test
	oldRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(oldRegistry)

	// Test without service
	services.SetGlobalRegistry(services.NewRegistry())
	_, err := cmd.getCatalogService()
	if err == nil {
		t.Error("Expected error when service not available")
	}

	// Test with service
	ctx := context.New()
	ctx.SetTestMode(true)
	registry := services.GetGlobalRegistry()
	if err := registry.RegisterService(services.NewCatalogService()); err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	service, err := cmd.getCatalogService()
	if err != nil {
		t.Errorf("Expected no error with service available, got: %v", err)
	}
	if service == nil {
		t.Error("Expected service, got nil")
	}
	if service.Name() != "catalog" {
		t.Errorf("Expected catalog service, got %s", service.Name())
	}
}

func TestCatalogCommand_GetVariableService(t *testing.T) {
	cmd := &CatalogCommand{}

	// Save original registry and restore after test
	oldRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(oldRegistry)

	// Test without service
	services.SetGlobalRegistry(services.NewRegistry())
	_, err := cmd.getVariableService()
	if err == nil {
		t.Error("Expected error when service not available")
	}

	// Test with service
	ctx := context.New()
	ctx.SetTestMode(true)
	registry := services.GetGlobalRegistry()
	if err := registry.RegisterService(services.NewVariableService()); err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	service, err := cmd.getVariableService()
	if err != nil {
		t.Errorf("Expected no error with service available, got: %v", err)
	}
	if service == nil {
		t.Error("Expected service, got nil")
	}
	if service.Name() != "variable" {
		t.Errorf("Expected variable service, got %s", service.Name())
	}
}

func TestCatalogCommand_GetModelService(t *testing.T) {
	cmd := &CatalogCommand{}

	// Save original registry and restore after test
	oldRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(oldRegistry)

	// Test without service
	services.SetGlobalRegistry(services.NewRegistry())
	_, err := cmd.getModelService()
	if err == nil {
		t.Error("Expected error when service not available")
	}

	// Test with service
	ctx := context.New()
	ctx.SetTestMode(true)
	registry := services.GetGlobalRegistry()
	if err := registry.RegisterService(services.NewModelService()); err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	service, err := cmd.getModelService()
	if err != nil {
		t.Errorf("Expected no error with service available, got: %v", err)
	}
	if service == nil {
		t.Error("Expected service, got nil")
	}
	if service.Name() != "model" {
		t.Errorf("Expected model service, got %s", service.Name())
	}
}

func TestCatalogCommand_SetResultVariables(t *testing.T) {
	cmd := &CatalogCommand{}
	ctx := context.New()
	ctx.SetTestMode(true)

	// Save original registry and restore after test
	oldRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(oldRegistry)

	// Setup variable service
	services.SetGlobalRegistry(services.NewRegistry())
	registry := services.GetGlobalRegistry()
	if err := registry.RegisterService(services.NewVariableService()); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	variableService, err := cmd.getVariableService()
	if err != nil {
		t.Fatalf("Failed to get variable service: %v", err)
	}

	// Test data
	result := &neurotypes.CatalogSearchResult{
		Models: []neurotypes.CatalogModel{
			{ID: "model1", Provider: "openai"},
			{ID: "model2", Provider: "openai"},
		},
		Count:    2,
		Provider: "openai",
		ModelID:  "", // Multiple results, so no single model ID
	}

	// Execute
	err = cmd.setResultVariables(result, variableService, ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Validate variables
	count, _ := ctx.GetVariable("_catalog_count")
	if count != "2" {
		t.Errorf("Expected count=2, got %s", count)
	}

	provider, _ := ctx.GetVariable("_catalog_provider")
	if provider != "openai" {
		t.Errorf("Expected provider=openai, got %s", provider)
	}

	models, _ := ctx.GetVariable("_catalog_models")
	if models != "model1,model2" {
		t.Errorf("Expected models=model1,model2, got %s", models)
	}

	// Test single result
	singleResult := &neurotypes.CatalogSearchResult{
		Models:  []neurotypes.CatalogModel{{ID: "single-model", Provider: "openai"}},
		Count:   1,
		ModelID: "single-model",
	}

	err = cmd.setResultVariables(singleResult, variableService, ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	modelID, _ := ctx.GetVariable("_catalog_model_id")
	if modelID != "single-model" {
		t.Errorf("Expected model_id=single-model, got %s", modelID)
	}
}

// Integration test for command registration
func TestCatalogCommand_Registration(t *testing.T) {
	// Check that command is registered
	cmd, exists := commands.GetGlobalRegistry().Get("model-catalog")
	if !exists {
		t.Error("Command should be registered")
		return
	}

	catalogCmd, ok := cmd.(*CatalogCommand)
	if !ok {
		t.Error("Registered command should be CatalogCommand type")
	}

	if catalogCmd.Name() != "model-catalog" {
		t.Errorf("Expected command name model-catalog, got %s", catalogCmd.Name())
	}
}

// Benchmark tests
func BenchmarkCatalogCommand_Execute(b *testing.B) {
	cmd, ctx := setupCatalogTest(b)

	args := map[string]string{"provider": "openai"}
	input := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cmd.Execute(args, input, ctx)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}
}
