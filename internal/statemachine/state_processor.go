// Package statemachine implements individual command state processing for the stack-based execution engine.
// The StateProcessor handles the command processing pipeline: Interpolation → Parsing → Resolving → Execution
package statemachine

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// StateProcessor handles individual command processing through the proven pipeline.
// It uses the existing interpolation and resolution logic.
type StateProcessor struct {
	// Context for state management
	context *context.NeuroContext
	// Integrated interpolation engine
	interpolator *CoreInterpolator
	// Command resolver
	resolver *CommandResolver
	// Configuration options
	config neurotypes.StateMachineConfig
	// Custom styled logger
	logger *log.Logger
	// Services
	stackService *services.StackService
}

// NewStateProcessor creates a new state processor with the given context and configuration.
func NewStateProcessor(ctx *context.NeuroContext, config neurotypes.StateMachineConfig) *StateProcessor {
	sp := &StateProcessor{
		context:      ctx,
		interpolator: NewCoreInterpolator(ctx),
		resolver:     NewCommandResolver(),
		config:       config,
		logger:       logger.NewStyledLogger("StateProcessor"),
	}

	// Initialize services
	var err error
	sp.stackService, err = services.GetGlobalStackService()
	if err != nil {
		sp.logger.Error("Failed to get stack service", "error", err)
	}

	return sp
}

// ProcessCommand processes a single command through the complete pipeline.
// This preserves the existing proven logic from the state machine processor.
func (sp *StateProcessor) ProcessCommand(rawCommand string) error {
	sp.logger.Debug("Processing command through pipeline", "command", rawCommand)

	// 1. Variable Interpolation (StateInterpolating equivalent)
	interpolated, err := sp.interpolateVariables(rawCommand)
	if err != nil {
		return fmt.Errorf("variable expansion failed: %w", err)
	}

	// 2. Command Parsing (StateParsing equivalent)
	parsed, err := sp.parseCommand(interpolated)
	if err != nil {
		return fmt.Errorf("command parsing failed: %w", err)
	}

	// 3. Command Resolution (StateResolving equivalent)
	resolved, err := sp.resolveCommand(parsed)
	if err != nil {
		return fmt.Errorf("command resolution failed: %w", err)
	}

	// 4. Command Execution (StateExecuting equivalent)
	return sp.executeCommand(resolved, parsed, interpolated)
}

// interpolateVariables handles variable and macro expansion.
// This uses the existing CoreInterpolator which already handles recursion limits properly.
func (sp *StateProcessor) interpolateVariables(input string) (string, error) {
	// Use the existing CoreInterpolator.InterpolateCommandLine method
	// which already handles recursion limits and all the complex logic
	expanded, hasVariables, err := sp.interpolator.InterpolateCommandLine(input)
	if err != nil {
		return "", err
	}

	// The CoreInterpolator already handles recursive expansion with safety limits
	// No need to reimplement this logic
	_ = hasVariables // We don't need to track this for the stack machine

	return expanded, nil
}

// parseCommand handles command structure parsing.
// This preserves the existing parsing logic from processParsing.
func (sp *StateProcessor) parseCommand(input string) (*parser.Command, error) {
	cmd := parser.ParseInputWithContext(input, sp.context)
	if cmd == nil {
		return nil, fmt.Errorf("failed to parse command: %s", input)
	}

	return cmd, nil
}

// resolveCommand handles command resolution through the priority system.
// This uses the standalone CommandResolver.
func (sp *StateProcessor) resolveCommand(parsedCmd *parser.Command) (*neurotypes.StateMachineResolvedCommand, error) {
	// Handle try command as a special case FIRST
	if parsedCmd.Name == "try" {
		return &neurotypes.StateMachineResolvedCommand{
			Name: "try",
			Type: neurotypes.CommandTypeTry,
		}, nil
	}

	// Use the command resolver for all other commands
	return sp.resolver.ResolveCommand(parsedCmd.Name)
}

// executeCommand handles command execution based on the resolved command type.
// This preserves and enhances the existing execution logic.
func (sp *StateProcessor) executeCommand(resolved *neurotypes.StateMachineResolvedCommand, parsedCmd *parser.Command, input string) error {
	switch resolved.Type {
	case neurotypes.CommandTypeTry:
		return sp.executeTryCommand(parsedCmd, input)
	case neurotypes.CommandTypeBuiltin:
		return sp.executeBuiltinCommand(resolved, parsedCmd, input)
	case neurotypes.CommandTypeStdlib, neurotypes.CommandTypeUser:
		return sp.executeScriptCommand(resolved, parsedCmd)
	default:
		return fmt.Errorf("unknown command type: %v", resolved.Type)
	}
}

// executeTryCommand handles try command execution with error boundary markers.
func (sp *StateProcessor) executeTryCommand(parsedCmd *parser.Command, _ string) error {
	// Extract target command from the try command message
	targetCommand := strings.TrimSpace(parsedCmd.Message)

	// Create try handler for this operation
	tryHandler := NewTryHandler()

	if targetCommand == "" {
		// Empty try command - set success variables
		tryHandler.SetupEmptyTryCommand()
		return nil
	}

	// Generate unique try ID and push boundary
	tryID := tryHandler.GenerateUniqueTryID()
	tryHandler.PushTryBoundary(tryID, targetCommand)

	return nil
}

// executeBuiltinCommand handles execution of builtin commands.
// This preserves the existing execution logic from processExecuting.
func (sp *StateProcessor) executeBuiltinCommand(resolved *neurotypes.StateMachineResolvedCommand, parsedCmd *parser.Command, _ string) error {
	if resolved.BuiltinCommand == nil {
		return fmt.Errorf("no builtin command to execute")
	}

	// Execute the builtin command
	err := resolved.BuiltinCommand.Execute(parsedCmd.Options, parsedCmd.Message)
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// executeScriptCommand handles execution of script commands (stdlib and user).
// This preserves the existing script execution logic.
func (sp *StateProcessor) executeScriptCommand(resolved *neurotypes.StateMachineResolvedCommand, parsedCmd *parser.Command) error {
	if resolved.ScriptContent == "" {
		return fmt.Errorf("no script content to execute")
	}

	// Set up parameter variables before script execution
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		sp.logger.Error("Failed to get variable service", "error", err)
	} else {
		// Special parameter variables (using SetSystemVariable for system-level variables)
		_ = variableService.SetSystemVariable("_0", resolved.Name)                         // Command name
		_ = variableService.SetSystemVariable("_1", parsedCmd.Message)                     // Message/first positional arg
		_ = variableService.SetSystemVariable("_*", parsedCmd.Message)                     // All positional args (same as _1)
		_ = variableService.SetSystemVariable("_@", sp.formatNamedArgs(parsedCmd.Options)) // Named args

		// Individual named parameters (these are user-level variables)
		for key, value := range parsedCmd.Options {
			_ = variableService.Set(key, value)
		}
	}

	// Parse script lines
	lines := strings.Split(resolved.ScriptContent, "\n")
	scriptLines := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments (%% for neuro-style comments)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			scriptLines = append(scriptLines, trimmed)
		}
	}

	if len(scriptLines) == 0 {
		return nil // Empty script is successful
	}

	// Push script lines to stack in reverse order (LIFO execution)
	if sp.stackService != nil {
		for i := len(scriptLines) - 1; i >= 0; i-- {
			sp.stackService.PushCommand(scriptLines[i])
		}
	}

	return nil
}

// formatNamedArgs formats named arguments as a comma-separated string.
func (sp *StateProcessor) formatNamedArgs(options map[string]string) string {
	if len(options) == 0 {
		return ""
	}
	var parts []string
	for key, value := range options {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, ",")
}

// SetConfig updates the configuration.
func (sp *StateProcessor) SetConfig(config neurotypes.StateMachineConfig) {
	sp.config = config
}
