//go:build !minimal

package builtin

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CheckCommand implements the \check command for checking service availability and initialization status.
// It provides diagnostics for service registry health and individual service status.
type CheckCommand struct{}

// Name returns the command name "check" for registration and lookup.
func (c *CheckCommand) Name() string {
	return "check"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *CheckCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the check command does.
func (c *CheckCommand) Description() string {
	return "Check service initialization status and availability"
}

// Usage returns the syntax and usage examples for the check command.
func (c *CheckCommand) Usage() string {
	return "\\check[service=name, all=true, quiet=true]"
}

// HelpInfo returns structured help information for the check command.
func (c *CheckCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "service",
				Description: "Specific service name to check (e.g., variable, help, bash)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "all",
				Description: "Check all registered services",
				Required:    false,
				Type:        "bool",
				Default:     "true",
			},
			{
				Name:        "quiet",
				Description: "Suppress output and only set result variables",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\check",
				Description: "Check all services and display status",
			},
			{
				Command:     "\\check[service=variable]",
				Description: "Check only the variable service",
			},
			{
				Command:     "\\check[service=help]",
				Description: "Check only the help service",
			},
			{
				Command:     "\\check[quiet=true]",
				Description: "Check all services quietly, only set result variables",
			},
			{
				Command:     "\\check[service=bash, quiet=true]",
				Description: "Check bash service quietly",
			},
		},
		Notes: []string{
			"Sets result variables: ${_check_status}, ${_check_output}, ${_check_failed_services}",
			"Also sets: ${_check_total_services}, ${_check_failed_count}",
			"Available services: variable, help, bash, render, editor, script, executor, chat_session",
			"Service status: [OK] available/initialized, [FAIL] not available/not initialized",
		},
	}
}

// ServiceCheckResult represents the result of checking a single service.
type ServiceCheckResult struct {
	Name        string
	Available   bool
	Initialized bool
	Error       string
}

// Execute checks service availability and initialization status.
func (c *CheckCommand) Execute(args map[string]string, _ string) error {

	// Parse arguments
	serviceName := args["service"]
	quiet := args["quiet"] == "true"

	// Get the service registry
	registry := services.GetGlobalRegistry()
	if registry == nil {
		return fmt.Errorf("service registry not available")
	}

	var results []ServiceCheckResult
	var err error

	// Check specific service or all services
	switch {
	case serviceName != "":
		// Check specific service
		result := c.checkSingleService(serviceName, registry)
		results = append(results, result)
	default:
		// Check all services (default behavior)
		results, err = c.checkAllServices(registry)
		if err != nil {
			return fmt.Errorf("failed to check services: %w", err)
		}
	}

	// Set result variables
	if err := c.setResultVariables(results); err != nil {
		return fmt.Errorf("failed to set result variables: %w", err)
	}

	// Display results if not quiet
	if !quiet {
		c.displayResults(results)
	}

	return nil
}

// checkSingleService checks the status of a single service.
func (c *CheckCommand) checkSingleService(serviceName string, registry *services.Registry) ServiceCheckResult {
	result := ServiceCheckResult{
		Name:        serviceName,
		Available:   false,
		Initialized: false,
	}

	// Try to get the service
	service, err := registry.GetService(serviceName)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Available = true

	// Check if service is initialized by trying to call a method
	// Most services implement this pattern where they check initialized flag
	result.Initialized = c.isServiceInitialized(service)

	return result
}

// checkAllServices checks the status of all registered services.
func (c *CheckCommand) checkAllServices(registry *services.Registry) ([]ServiceCheckResult, error) {
	allServices := registry.GetAllServices()
	results := make([]ServiceCheckResult, 0, len(allServices))

	// Get service names and sort them for consistent output
	serviceNames := make([]string, 0, len(allServices))
	for name := range allServices {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)

	// Check each service
	for _, name := range serviceNames {
		service := allServices[name]
		result := ServiceCheckResult{
			Name:        name,
			Available:   true,
			Initialized: c.isServiceInitialized(service),
		}
		results = append(results, result)
	}

	return results, nil
}

// isServiceInitialized checks if a service is initialized.
// This is a heuristic approach since not all services expose initialized state directly.
func (c *CheckCommand) isServiceInitialized(_ neurotypes.Service) bool {
	// Try to use reflection or type assertions to check initialized state
	// For now, we'll assume all services that are successfully retrieved are initialized
	// since they go through InitializeAll() during startup

	// We could extend this in the future to check service-specific initialized flags
	// by adding a method to the Service interface or using type assertions

	return true // Services in the registry are assumed to be initialized
}

// setResultVariables sets the result variables based on the check results.
func (c *CheckCommand) setResultVariables(results []ServiceCheckResult) error {
	// Count failures
	failedCount := 0
	failedServices := make([]string, 0)

	for _, result := range results {
		if !result.Available || !result.Initialized {
			failedCount++
			failedServices = append(failedServices, result.Name)
		}
	}

	// Set status
	status := "success"
	if failedCount > 0 {
		status = "failed"
	}

	// Generate output
	output := c.generateOutput(results)

	// Get variable service to set system variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Set system variables using the variable service
	if err := variableService.SetSystemVariable("_check_status", status); err != nil {
		return err
	}
	if err := variableService.SetSystemVariable("_check_output", output); err != nil {
		return err
	}
	if err := variableService.SetSystemVariable("_check_failed_services", strings.Join(failedServices, ",")); err != nil {
		return err
	}
	if err := variableService.SetSystemVariable("_check_total_services", fmt.Sprintf("%d", len(results))); err != nil {
		return err
	}
	if err := variableService.SetSystemVariable("_check_failed_count", fmt.Sprintf("%d", failedCount)); err != nil {
		return err
	}

	return nil
}

// generateOutput creates a formatted output string from the results.
func (c *CheckCommand) generateOutput(results []ServiceCheckResult) string {
	var output strings.Builder

	for _, result := range results {
		status := "[OK]"
		if !result.Available || !result.Initialized {
			status = "[FAIL]"
		}

		statusDesc := "available/initialized"
		if !result.Available {
			statusDesc = "not available"
		} else if !result.Initialized {
			statusDesc = "not initialized"
		}

		output.WriteString(fmt.Sprintf("%s %s - %s", status, result.Name, statusDesc))
		if result.Error != "" {
			output.WriteString(fmt.Sprintf(" (%s)", result.Error))
		}
		output.WriteString("\n")
	}

	return strings.TrimSuffix(output.String(), "\n")
}

// displayResults displays the check results to the console.
func (c *CheckCommand) displayResults(results []ServiceCheckResult) {
	if len(results) == 0 {
		fmt.Println("No services found to check.")
		return
	}

	fmt.Println("Service Status Check:")
	fmt.Println("====================")

	successCount := 0
	for _, result := range results {
		status := "[OK]"
		statusDesc := "available/initialized"

		if !result.Available || !result.Initialized {
			status = "[FAIL]"
			if !result.Available {
				statusDesc = "not available"
			} else if !result.Initialized {
				statusDesc = "not initialized"
			}
		} else {
			successCount++
		}

		fmt.Printf("  %s %-20s - %s", status, result.Name, statusDesc)
		if result.Error != "" {
			fmt.Printf(" (%s)", result.Error)
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Printf("Summary: %d/%d services healthy\n", successCount, len(results))

	if successCount < len(results) {
		fmt.Println("Some services are not available or not initialized.")
	}
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CheckCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register check command: %v", err))
	}
}
