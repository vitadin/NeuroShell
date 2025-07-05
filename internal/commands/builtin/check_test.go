package builtin

import (
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// MockService for testing - implements the Service interface
type MockService struct {
	name        string
	initialized bool
	initError   error
}

func (m *MockService) Name() string {
	return m.name
}

func (m *MockService) Initialize(_ neurotypes.Context) error {
	if m.initError != nil {
		return m.initError
	}
	m.initialized = true
	return nil
}

func TestCheckCommand_Name(t *testing.T) {
	cmd := &CheckCommand{}
	if cmd.Name() != "check" {
		t.Errorf("Expected command name 'check', got '%s'", cmd.Name())
	}
}

func TestCheckCommand_ParseMode(t *testing.T) {
	cmd := &CheckCommand{}
	if cmd.ParseMode() != neurotypes.ParseModeKeyValue {
		t.Errorf("Expected ParseModeKeyValue, got %v", cmd.ParseMode())
	}
}

func TestCheckCommand_Description(t *testing.T) {
	cmd := &CheckCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if !strings.Contains(desc, "service") {
		t.Error("Description should mention 'service'")
	}
}

func TestCheckCommand_Usage(t *testing.T) {
	cmd := &CheckCommand{}
	usage := cmd.Usage()
	if !strings.Contains(usage, "\\check") {
		t.Error("Usage should contain '\\check'")
	}
}

func TestCheckCommand_HelpInfo(t *testing.T) {
	cmd := &CheckCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != "check" {
		t.Errorf("Expected command 'check', got '%s'", helpInfo.Command)
	}

	if len(helpInfo.Options) == 0 {
		t.Error("HelpInfo should have options")
	}

	if len(helpInfo.Examples) == 0 {
		t.Error("HelpInfo should have examples")
	}

	// Check for specific options
	hasServiceOption := false
	hasQuietOption := false
	for _, option := range helpInfo.Options {
		if option.Name == "service" {
			hasServiceOption = true
		}
		if option.Name == "quiet" {
			hasQuietOption = true
		}
	}

	if !hasServiceOption {
		t.Error("HelpInfo should have 'service' option")
	}
	if !hasQuietOption {
		t.Error("HelpInfo should have 'quiet' option")
	}
}

func TestCheckCommand_Execute_AllServices(t *testing.T) {
	// Create a test registry with mock services
	registry := services.NewRegistry()

	// Register variable service (needed for setting result variables)
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Register mock services
	mockService1 := &MockService{name: "test1", initialized: true}
	mockService2 := &MockService{name: "test2", initialized: true}

	if err := registry.RegisterService(mockService1); err != nil {
		t.Fatalf("Failed to register mock service 1: %v", err)
	}
	if err := registry.RegisterService(mockService2); err != nil {
		t.Fatalf("Failed to register mock service 2: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services
	ctx := context.New() // Use real NeuroContext for system variables
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Create and execute check command
	cmd := &CheckCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "", ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check result variables
	status, err := ctx.GetVariable("_check_status")
	if err != nil {
		t.Fatalf("Failed to get _check_status: %v", err)
	}
	if status != "success" {
		t.Errorf("Expected status 'success', got '%s'", status)
	}

	totalServices, err := ctx.GetVariable("_check_total_services")
	if err != nil {
		t.Fatalf("Failed to get _check_total_services: %v", err)
	}
	if totalServices != "3" {
		t.Errorf("Expected total services '3', got '%s'", totalServices)
	}

	failedCount, err := ctx.GetVariable("_check_failed_count")
	if err != nil {
		t.Fatalf("Failed to get _check_failed_count: %v", err)
	}
	if failedCount != "0" {
		t.Errorf("Expected failed count '0', got '%s'", failedCount)
	}
}

func TestCheckCommand_Execute_SpecificService(t *testing.T) {
	// Create a test registry with mock services
	registry := services.NewRegistry()

	// Register variable service (needed for setting result variables)
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Register mock services
	mockService1 := &MockService{name: "test1", initialized: true}
	mockService2 := &MockService{name: "test2", initialized: true}

	if err := registry.RegisterService(mockService1); err != nil {
		t.Fatalf("Failed to register mock service 1: %v", err)
	}
	if err := registry.RegisterService(mockService2); err != nil {
		t.Fatalf("Failed to register mock service 2: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services
	ctx := context.New() // Use real NeuroContext for system variables
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Create and execute check command for specific service
	cmd := &CheckCommand{}
	args := map[string]string{
		"service": "test1",
	}

	err := cmd.Execute(args, "", ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check result variables
	status, err := ctx.GetVariable("_check_status")
	if err != nil {
		t.Fatalf("Failed to get _check_status: %v", err)
	}
	if status != "success" {
		t.Errorf("Expected status 'success', got '%s'", status)
	}

	totalServices, err := ctx.GetVariable("_check_total_services")
	if err != nil {
		t.Fatalf("Failed to get _check_total_services: %v", err)
	}
	if totalServices != "1" {
		t.Errorf("Expected total services '1', got '%s'", totalServices)
	}

	output, err := ctx.GetVariable("_check_output")
	if err != nil {
		t.Fatalf("Failed to get _check_output: %v", err)
	}
	if !strings.Contains(output, "test1") {
		t.Errorf("Expected output to contain 'test1', got '%s'", output)
	}
}

func TestCheckCommand_Execute_NonExistentService(t *testing.T) {
	// Create a test registry with mock services
	registry := services.NewRegistry()

	// Register variable service (needed for setting result variables)
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Register mock services
	mockService1 := &MockService{name: "test1", initialized: true}

	if err := registry.RegisterService(mockService1); err != nil {
		t.Fatalf("Failed to register mock service 1: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services
	ctx := context.New() // Use real NeuroContext for system variables
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Create and execute check command for non-existent service
	cmd := &CheckCommand{}
	args := map[string]string{
		"service": "nonexistent",
	}

	err := cmd.Execute(args, "", ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check result variables
	status, err := ctx.GetVariable("_check_status")
	if err != nil {
		t.Fatalf("Failed to get _check_status: %v", err)
	}
	if status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", status)
	}

	failedCount, err := ctx.GetVariable("_check_failed_count")
	if err != nil {
		t.Fatalf("Failed to get _check_failed_count: %v", err)
	}
	if failedCount != "1" {
		t.Errorf("Expected failed count '1', got '%s'", failedCount)
	}

	failedServices, err := ctx.GetVariable("_check_failed_services")
	if err != nil {
		t.Fatalf("Failed to get _check_failed_services: %v", err)
	}
	if failedServices != "nonexistent" {
		t.Errorf("Expected failed services 'nonexistent', got '%s'", failedServices)
	}
}

func TestCheckCommand_Execute_QuietMode(t *testing.T) {
	// Create a test registry with mock services
	registry := services.NewRegistry()

	// Register variable service (needed for setting result variables)
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Register mock services
	mockService1 := &MockService{name: "test1", initialized: true}

	if err := registry.RegisterService(mockService1); err != nil {
		t.Fatalf("Failed to register mock service 1: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services
	ctx := context.New() // Use real NeuroContext for system variables
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Create and execute check command in quiet mode
	cmd := &CheckCommand{}
	args := map[string]string{
		"quiet": "true",
	}

	err := cmd.Execute(args, "", ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check that result variables are still set
	status, err := ctx.GetVariable("_check_status")
	if err != nil {
		t.Fatalf("Failed to get _check_status: %v", err)
	}
	if status != "success" {
		t.Errorf("Expected status 'success', got '%s'", status)
	}

	// Output should still be generated for variables
	output, err := ctx.GetVariable("_check_output")
	if err != nil {
		t.Fatalf("Failed to get _check_output: %v", err)
	}
	if output == "" {
		t.Error("Output should not be empty even in quiet mode")
	}
}

func TestCheckCommand_checkSingleService(t *testing.T) {
	// Create a test registry with mock services
	registry := services.NewRegistry()

	// Register mock services
	mockService1 := &MockService{name: "test1", initialized: true}

	if err := registry.RegisterService(mockService1); err != nil {
		t.Fatalf("Failed to register mock service 1: %v", err)
	}

	// Test checking existing service
	cmd := &CheckCommand{}
	result := cmd.checkSingleService("test1", registry)

	if result.Name != "test1" {
		t.Errorf("Expected name 'test1', got '%s'", result.Name)
	}
	if !result.Available {
		t.Error("Service should be available")
	}
	if !result.Initialized {
		t.Error("Service should be initialized")
	}
	if result.Error != "" {
		t.Errorf("Expected no error, got '%s'", result.Error)
	}

	// Test checking non-existent service
	result = cmd.checkSingleService("nonexistent", registry)

	if result.Name != "nonexistent" {
		t.Errorf("Expected name 'nonexistent', got '%s'", result.Name)
	}
	if result.Available {
		t.Error("Service should not be available")
	}
	if result.Error == "" {
		t.Error("Expected error for non-existent service")
	}
}

func TestCheckCommand_generateOutput(t *testing.T) {
	cmd := &CheckCommand{}

	results := []ServiceCheckResult{
		{Name: "service1", Available: true, Initialized: true},
		{Name: "service2", Available: false, Initialized: false, Error: "not found"},
		{Name: "service3", Available: true, Initialized: false},
	}

	output := cmd.generateOutput(results)

	if !strings.Contains(output, "service1") {
		t.Error("Output should contain 'service1'")
	}
	if !strings.Contains(output, "service2") {
		t.Error("Output should contain 'service2'")
	}
	if !strings.Contains(output, "service3") {
		t.Error("Output should contain 'service3'")
	}
	if !strings.Contains(output, "✓") {
		t.Error("Output should contain success checkmark")
	}
	if !strings.Contains(output, "✗") {
		t.Error("Output should contain failure cross")
	}
	if !strings.Contains(output, "not found") {
		t.Error("Output should contain error message")
	}
}

func TestCheckCommand_setResultVariables(t *testing.T) {
	cmd := &CheckCommand{}
	ctx := context.New() // Use real NeuroContext for system variables

	// Set up variable service in global registry for testing
	registry := services.NewRegistry()
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	results := []ServiceCheckResult{
		{Name: "service1", Available: true, Initialized: true},
		{Name: "service2", Available: false, Initialized: false, Error: "not found"},
	}

	err := cmd.setResultVariables(results, ctx)
	if err != nil {
		t.Fatalf("setResultVariables failed: %v", err)
	}

	// Check all expected variables are set
	expectedVars := []string{
		"_check_status",
		"_check_output",
		"_check_failed_services",
		"_check_total_services",
		"_check_failed_count",
	}

	for _, varName := range expectedVars {
		value, err := ctx.GetVariable(varName)
		if err != nil {
			t.Errorf("Failed to get variable '%s': %v", varName, err)
		}
		if value == "" {
			t.Errorf("Variable '%s' should not be empty", varName)
		}
	}

	// Check specific values
	status, _ := ctx.GetVariable("_check_status")
	if status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", status)
	}

	failedCount, _ := ctx.GetVariable("_check_failed_count")
	if failedCount != "1" {
		t.Errorf("Expected failed count '1', got '%s'", failedCount)
	}

	totalServices, _ := ctx.GetVariable("_check_total_services")
	if totalServices != "2" {
		t.Errorf("Expected total services '2', got '%s'", totalServices)
	}

	failedServices, _ := ctx.GetVariable("_check_failed_services")
	if failedServices != "service2" {
		t.Errorf("Expected failed services 'service2', got '%s'", failedServices)
	}
}

func TestCheckCommand_isServiceInitialized(t *testing.T) {
	cmd := &CheckCommand{}

	// Test with mock service
	mockService := &MockService{name: "test", initialized: true}

	// Currently this method always returns true since services in registry are assumed initialized
	// This test verifies the current behavior
	result := cmd.isServiceInitialized(mockService)
	if !result {
		t.Error("Service should be considered initialized")
	}

	// Test with uninitialized service
	uninitializedService := &MockService{name: "test", initialized: false}
	result = cmd.isServiceInitialized(uninitializedService)
	if !result {
		t.Error("Service should be considered initialized (current implementation)")
	}
}

// Test edge cases and error conditions
func TestCheckCommand_Execute_EmptyRegistry(t *testing.T) {
	// Create an empty registry
	registry := services.NewRegistry()

	// Register variable service (needed for setting result variables)
	variableService := services.NewVariableService()
	if err := registry.RegisterService(variableService); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}

	// Set the test registry as global
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)
	services.SetGlobalRegistry(registry)

	// Initialize services (empty registry)
	ctx := context.New() // Use real NeuroContext for system variables
	if err := registry.InitializeAll(ctx); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Create and execute check command
	cmd := &CheckCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "", ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check result variables
	status, err := ctx.GetVariable("_check_status")
	if err != nil {
		t.Fatalf("Failed to get _check_status: %v", err)
	}
	if status != "success" {
		t.Errorf("Expected status 'success' for empty registry, got '%s'", status)
	}

	totalServices, err := ctx.GetVariable("_check_total_services")
	if err != nil {
		t.Fatalf("Failed to get _check_total_services: %v", err)
	}
	if totalServices != "1" {
		t.Errorf("Expected total services '1', got '%s'", totalServices)
	}
}
