// Package commands provides script-based command implementations.
// This file implements ScriptCommand which wraps .neuro scripts as
// executable commands that integrate with the command system.
package commands

import (
	"bufio"
	"fmt"
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ScriptCommand implements the Command interface for script-based commands.
// It wraps .neuro script content and executes it as a command, handling
// parameter passing and integration with the macro expansion system.
type ScriptCommand struct {
	name        string
	content     string
	commandType neurotypes.CommandType
	description string
	usage       string

	// Cached help info to avoid parsing script multiple times
	helpInfo *neurotypes.HelpInfo
}

// NewScriptCommand creates a new ScriptCommand from script content.
// The script content should be valid NeuroShell script syntax.
func NewScriptCommand(name, content string, commandType neurotypes.CommandType) *ScriptCommand {
	return &ScriptCommand{
		name:        name,
		content:     content,
		commandType: commandType,
	}
}

// Name implements the Command interface.
func (sc *ScriptCommand) Name() string {
	return sc.name
}

// ParseMode implements the Command interface.
// Script commands use KeyValue parsing by default, but this can be
// overridden by script metadata comments.
func (sc *ScriptCommand) ParseMode() neurotypes.ParseMode {
	// Check for parse mode directive in script comments
	if mode := sc.extractParseMode(); mode != neurotypes.ParseModeKeyValue {
		return mode
	}
	return neurotypes.ParseModeKeyValue
}

// Description implements the Command interface.
// Extracts description from script comments or provides a default.
func (sc *ScriptCommand) Description() string {
	if sc.description != "" {
		return sc.description
	}

	// Try to extract description from script comments
	if desc := sc.extractDescription(); desc != "" {
		sc.description = desc
		return desc
	}

	// Default description based on command type
	switch sc.commandType {
	case neurotypes.CommandTypeStdlib:
		sc.description = fmt.Sprintf("Standard library command: %s", sc.name)
	case neurotypes.CommandTypeUser:
		sc.description = fmt.Sprintf("User-defined command: %s", sc.name)
	default:
		sc.description = fmt.Sprintf("Script command: %s", sc.name)
	}

	return sc.description
}

// Usage implements the Command interface.
// Extracts usage information from script comments or provides a default.
func (sc *ScriptCommand) Usage() string {
	if sc.usage != "" {
		return sc.usage
	}

	// Try to extract usage from script comments
	if usage := sc.extractUsage(); usage != "" {
		sc.usage = usage
		return usage
	}

	// Default usage pattern
	sc.usage = fmt.Sprintf("\\%s [options] <message>", sc.name)
	return sc.usage
}

// HelpInfo implements the Command interface.
// Parses script comments to extract comprehensive help information.
func (sc *ScriptCommand) HelpInfo() neurotypes.HelpInfo {
	if sc.helpInfo != nil {
		return *sc.helpInfo
	}

	// Parse help information from script content
	helpInfo := neurotypes.HelpInfo{
		Command:     sc.name,
		Description: sc.Description(),
		Usage:       sc.Usage(),
		ParseMode:   sc.ParseMode(),
		Options:     sc.extractOptions(),
		Examples:    sc.extractExamples(),
		Notes:       sc.extractNotes(),
	}

	sc.helpInfo = &helpInfo
	return helpInfo
}

// Execute implements the Command interface.
// Sets up parameters, executes the script content, and cleans up.
func (sc *ScriptCommand) Execute(args map[string]string, input string) error {
	logger.Debug("Executing script command", "name", sc.name, "type", sc.commandType.String())

	// Setup script parameters
	err := sc.setupParameters(args, input)
	if err != nil {
		return fmt.Errorf("failed to setup parameters for script %s: %w", sc.name, err)
	}

	// Ensure cleanup happens regardless of execution outcome
	defer func() {
		if cleanupErr := sc.cleanupParameters(); cleanupErr != nil {
			logger.Error("Failed to cleanup script parameters", "script", sc.name, "error", cleanupErr)
		}
	}()

	// Execute script content using the enhanced script execution pipeline
	err = sc.executeScriptContent()
	if err != nil {
		return fmt.Errorf("script execution failed for %s: %w", sc.name, err)
	}

	logger.Debug("Script command executed successfully", "name", sc.name)
	return nil
}

// setupParameters configures script variables based on command arguments.
// This implements the parameter variable system using ${_0}, ${_1}, etc.
func (sc *ScriptCommand) setupParameters(args map[string]string, input string) error {
	vs, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Set standard script parameters
	err = vs.SetSystemVariable("_0", sc.name) // Command name
	if err != nil {
		return fmt.Errorf("failed to set _0 parameter: %w", err)
	}

	err = vs.SetSystemVariable("_1", input) // Input message
	if err != nil {
		return fmt.Errorf("failed to set _1 parameter: %w", err)
	}

	err = vs.SetSystemVariable("_*", input) // All positional args
	if err != nil {
		return fmt.Errorf("failed to set _* parameter: %w", err)
	}

	// Set named arguments as individual variables
	var namedArgs []string
	for key, value := range args {
		err = vs.SetSystemVariable(key, value)
		if err != nil {
			return fmt.Errorf("failed to set parameter %s: %w", key, err)
		}
		namedArgs = append(namedArgs, key+"="+value)
	}

	// Set all named args as a single variable
	err = vs.SetSystemVariable("_@", strings.Join(namedArgs, " "))
	if err != nil {
		return fmt.Errorf("failed to set _@ parameter: %w", err)
	}

	logger.Debug("Script parameters setup completed", "script", sc.name, "args_count", len(args))
	return nil
}

// cleanupParameters removes script parameters after execution.
// This prevents parameter variables from leaking between script executions.
func (sc *ScriptCommand) cleanupParameters() error {
	vs, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Clear standard script parameters by setting them to empty strings
	parameterNames := []string{"_0", "_1", "_*", "_@"}
	for _, name := range parameterNames {
		err = vs.SetSystemVariable(name, "")
		if err != nil {
			logger.Debug("Failed to cleanup parameter", "parameter", name, "error", err)
		}
	}

	// Note: Named parameters (args) are left as-is since they might be user variables
	// The script itself should manage cleanup of named parameters if needed

	return nil
}

// executeScriptContent executes the script content using the enhanced
// script execution pipeline with macro expansion support.
func (sc *ScriptCommand) executeScriptContent() error {
	// Get required services
	is, err := services.GetGlobalInterpolationService()
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	vs, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Process script content line by line with macro expansion
	scanner := bufio.NewScanner(strings.NewReader(sc.content))
	lineNum := 0
	commandCount := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}

		commandCount++
		logger.Debug("Processing script content line", "script", sc.name, "line_number", lineNum, "original", line)

		// PHASE 1: Command-level interpolation (macro expansion)
		expandedLine, err := is.InterpolateCommandLine(line)
		if err != nil {
			return fmt.Errorf("macro expansion failed at line %d '%s': %w", lineNum, line, err)
		}

		// Log expansion if it occurred
		if line != expandedLine {
			logger.Debug("Macro expansion applied", "script", sc.name, "line", lineNum, "original", line, "expanded", expandedLine)
		}

		// Parse the expanded command line
		cmd := parser.ParseInput(expandedLine)
		if cmd == nil {
			return fmt.Errorf("failed to parse expanded command at line %d: '%s'", lineNum, expandedLine)
		}

		logger.Debug("Command parsed", "script", sc.name, "line", lineNum, "command", cmd.Name, "message", cmd.Message)

		// Execute the command using builtin registry
		err = GetGlobalRegistry().Execute(cmd.Name, cmd.Options, cmd.Message)

		if err != nil {
			return fmt.Errorf("command execution failed at line %d (%s): %w", lineNum, cmd.Name, err)
		}

		logger.Debug("Command executed successfully", "script", sc.name, "line", lineNum, "command", cmd.Name)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading script content: %w", err)
	}

	// Set success status variables
	if err := vs.SetSystemVariable("_status", "0"); err != nil {
		logger.Debug("Failed to set _status variable", "error", err)
	}

	successMessage := fmt.Sprintf("Script content %s executed successfully with macros (%d commands)", sc.name, commandCount)
	if err := vs.SetSystemVariable("_output", successMessage); err != nil {
		logger.Debug("Failed to set _output variable", "error", err)
	}

	logger.Debug("Enhanced script content execution completed successfully", "script", sc.name, "commands_executed", commandCount)
	return nil
}

// extractDescription extracts description from script comments.
// Looks for comments like "%% Description: This command does X"
func (sc *ScriptCommand) extractDescription() string {
	lines := strings.Split(sc.content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% Description:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Description:"))
		}
		if strings.HasPrefix(trimmed, "%% Desc:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Desc:"))
		}
	}
	return ""
}

// extractUsage extracts usage information from script comments.
// Looks for comments like "%% Usage: \command [options] <message>"
func (sc *ScriptCommand) extractUsage() string {
	lines := strings.Split(sc.content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% Usage:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Usage:"))
		}
	}
	return ""
}

// extractParseMode extracts parse mode from script comments.
// Looks for comments like "%% ParseMode: raw"
func (sc *ScriptCommand) extractParseMode() neurotypes.ParseMode {
	lines := strings.Split(sc.content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% ParseMode:") {
			mode := strings.TrimSpace(strings.TrimPrefix(trimmed, "%% ParseMode:"))
			switch strings.ToLower(mode) {
			case "raw":
				return neurotypes.ParseModeRaw
			case "withoptions":
				return neurotypes.ParseModeWithOptions
			default:
				return neurotypes.ParseModeKeyValue
			}
		}
	}
	return neurotypes.ParseModeKeyValue
}

// extractOptions extracts option information from script comments.
// Looks for comments like "%% Option: name - description (type, default: value)"
func (sc *ScriptCommand) extractOptions() []neurotypes.HelpOption {
	var options []neurotypes.HelpOption
	lines := strings.Split(sc.content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% Option:") {
			optionText := strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Option:"))
			if option := parseOptionComment(optionText); option != nil {
				options = append(options, *option)
			}
		}
	}

	return options
}

// extractExamples extracts usage examples from script comments.
// Looks for comments like "%% Example: \command arg - does something"
func (sc *ScriptCommand) extractExamples() []neurotypes.HelpExample {
	var examples []neurotypes.HelpExample
	lines := strings.Split(sc.content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% Example:") {
			exampleText := strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Example:"))
			if example := parseExampleComment(exampleText); example != nil {
				examples = append(examples, *example)
			}
		}
	}

	return examples
}

// extractNotes extracts additional notes from script comments.
// Looks for comments like "%% Note: Important information"
func (sc *ScriptCommand) extractNotes() []string {
	var notes []string
	lines := strings.Split(sc.content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%% Note:") {
			note := strings.TrimSpace(strings.TrimPrefix(trimmed, "%% Note:"))
			if note != "" {
				notes = append(notes, note)
			}
		}
	}

	return notes
}

// parseOptionComment parses an option comment into a HelpOption struct.
// Expected format: "name - description (type, default: value)"
func parseOptionComment(text string) *neurotypes.HelpOption {
	// Simple parsing - can be enhanced later
	parts := strings.SplitN(text, " - ", 2)
	if len(parts) != 2 {
		return nil
	}

	return &neurotypes.HelpOption{
		Name:        strings.TrimSpace(parts[0]),
		Description: strings.TrimSpace(parts[1]),
		Required:    false,    // Default to optional
		Type:        "string", // Default type
	}
}

// parseExampleComment parses an example comment into a HelpExample struct.
// Expected format: "command - description"
func parseExampleComment(text string) *neurotypes.HelpExample {
	parts := strings.SplitN(text, " - ", 2)
	if len(parts) != 2 {
		return &neurotypes.HelpExample{
			Command:     text,
			Description: "",
		}
	}

	return &neurotypes.HelpExample{
		Command:     strings.TrimSpace(parts[0]),
		Description: strings.TrimSpace(parts[1]),
	}
}
