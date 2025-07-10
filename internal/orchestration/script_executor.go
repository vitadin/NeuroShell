// Package orchestration provides workflow orchestration for NeuroShell operations.
// This package contains centralized logic for coordinating multiple services
// to accomplish complex operations like script execution.
package orchestration

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"neuroshell/internal/commands"
	_ "neuroshell/internal/commands/assert" // Import assert commands (init functions)
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
)

// ExecuteScript orchestrates the complete script execution workflow with enhanced macro support.
// This function now uses the enhanced interpolation system by default, enabling command-level
// macro expansion where variables can contain entire commands.
//
// Enhanced Pipeline:
//
//	Load Line → Interpolate Command Line (Macro Expansion) → Parse → Execute
//
// This function consolidates the script execution logic that was previously
// duplicated between batch mode and the \run command, ensuring consistent
// behavior and maintainability.
//
// Parameters:
//   - scriptPath: Path to the .neuro script file to execute
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScript(scriptPath string) error {
	// Use the enhanced macro-enabled script execution by default
	return ExecuteScriptWithMacros(scriptPath)
}

// ExecuteScriptContent executes script content directly without reading from a file.
// This is useful for embedded scripts or dynamically generated content.
//
// Parameters:
//   - content: The script content to execute
//   - scriptName: Name for logging/debugging purposes
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScriptContent(content, scriptName string) error {
	return ExecuteScriptContentWithMacros(content, scriptName)
}

// ExecuteScriptLegacy provides the original script execution method without macro support.
// This function is maintained for backward compatibility and debugging purposes.
// It uses the original queue-based execution pipeline.
//
// Original Pipeline:
//
//	Load Script → Queue Commands → Parse → Interpolate Parameters → Execute
//
// Parameters:
//   - scriptPath: Path to the .neuro script file to execute
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScriptLegacy(scriptPath string) error {
	logger.Debug("Starting script execution", "script", scriptPath)

	// Phase 1: Get required services from global registry
	ss, err := services.GetGlobalScriptService()
	if err != nil {
		return fmt.Errorf("script service not available: %w", err)
	}

	es, err := services.GetGlobalExecutorService()
	if err != nil {
		return fmt.Errorf("executor service not available: %w", err)
	}

	is, err := services.GetGlobalInterpolationService()
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	vs, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Phase 2: Load script file into execution queue
	if err := ss.LoadScript(scriptPath); err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	logger.Debug("Script loaded successfully", "script", scriptPath)

	// Phase 3: Execute all commands in the queue
	commandCount := 0
	for {
		// Get next command from queue
		cmd, err := es.GetNextCommand()
		if err != nil {
			return fmt.Errorf("failed to get next command: %w", err)
		}
		if cmd == nil {
			break // No more commands
		}

		commandCount++
		logger.Debug("Executing command", "number", commandCount, "command", cmd.Name, "message", cmd.Message)

		// Output command line with %%> prefix if echo_commands is enabled
		if echoCommands, _ := vs.Get("_echo_commands"); echoCommands == "true" {
			// Use original text if available, otherwise fall back to reconstructed
			originalText := cmd.OriginalText
			if originalText == "" {
				originalText = cmd.String()
			}
			fmt.Printf("%%%%> %s\n", originalText)
		}

		// Interpolate command using interpolation service
		interpolatedCmd, err := is.InterpolateCommand(cmd)
		if err != nil {
			// Mark execution error for tracking
			if markErr := es.MarkExecutionError(err, cmd.String()); markErr != nil {
				logger.Error("Failed to mark execution error", "error", markErr)
			}
			return fmt.Errorf("interpolation failed for command %d: %w", commandCount, err)
		}

		// Prepare input for execution
		cmdInput := interpolatedCmd.Message

		// Execute command using builtin registry
		err = commands.GetGlobalRegistry().Execute(interpolatedCmd.Name, interpolatedCmd.Options, cmdInput)

		if err != nil {
			// Mark execution error and return
			if markErr := es.MarkExecutionError(err, cmd.String()); markErr != nil {
				logger.Error("Failed to mark execution error", "error", markErr)
			}
			return fmt.Errorf("command execution failed for command %d (%s): %w", commandCount, interpolatedCmd.Name, err)
		}

		// Mark command as successfully executed
		if err := es.MarkCommandExecuted(); err != nil {
			logger.Error("Failed to mark command as executed", "error", err)
		}

		logger.Debug("Command executed successfully", "number", commandCount, "command", interpolatedCmd.Name)
	}

	// Phase 4: Mark successful completion
	if err := es.MarkExecutionComplete(); err != nil {
		logger.Error("Failed to mark execution complete", "error", err)
	}

	// Phase 5: Set success status variables for caller access
	if err := vs.SetSystemVariable("_status", "0"); err != nil {
		logger.Debug("Failed to set _status variable", "error", err)
	}

	successMessage := fmt.Sprintf("Script %s executed successfully (%d commands)", scriptPath, commandCount)
	if err := vs.SetSystemVariable("_output", successMessage); err != nil {
		logger.Debug("Failed to set _output variable", "error", err)
	}

	logger.Debug("Script execution completed successfully", "script", scriptPath, "commands_executed", commandCount)
	return nil
}

// ExecuteScriptWithMacros executes a script file with enhanced command-level macro expansion support.
// This function provides the enhanced interpolation system where variables can contain entire commands.
//
// Enhanced Pipeline:
//
//	Load Line → Interpolate Command Line (Macro Expansion) → Parse → Execute
//
// Examples of macro expansion:
//
//	\set[cmd="echo"]
//	${cmd} Hello World    # Expands to: \echo Hello World before parsing
//
//	\set[debug_cmd="\echo[style=red] DEBUG:"]
//	${debug_cmd} Message  # Expands to: \echo[style=red] DEBUG: Message
//
// Parameters:
//   - scriptPath: Path to the .neuro script file to execute
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScriptWithMacros(scriptPath string) error {
	logger.Debug("Starting enhanced script execution with macro support", "script", scriptPath)

	// Get required services
	is, err := services.GetGlobalInterpolationService()
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	vs, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Open and read script file
	file, err := os.Open(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to open script file: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	commandCount := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}

		// Handle multiline continuation with ...
		if strings.HasSuffix(line, "...") {
			// Start accumulating multiline command
			var lines []string
			lines = append(lines, line)

			// Continue reading lines until we find one that doesn't end with ...
			for scanner.Scan() {
				lineNum++
				nextLine := strings.TrimSpace(scanner.Text())

				// Skip empty lines and comments in multiline context
				if nextLine == "" || strings.HasPrefix(nextLine, "%%") {
					continue
				}

				lines = append(lines, nextLine)

				// If this line doesn't end with ..., we're done
				if !strings.HasSuffix(nextLine, "...") {
					break
				}
			}

			// Join all lines with newlines and remove continuation markers
			multilineCommand := strings.Join(lines, "\n")

			// Remove multiline continuation markers (same logic as original script service)
			multilineCommand = strings.ReplaceAll(multilineCommand, "...\n", " ")
			multilineCommand = strings.ReplaceAll(multilineCommand, "...", " ")

			// Use the processed multiline command
			line = strings.TrimSpace(multilineCommand)
		}

		commandCount++
		logger.Debug("Processing script line", "line_number", lineNum, "original", line)

		// PHASE 1: Command-level interpolation (macro expansion)
		expandedLine, err := is.InterpolateCommandLine(line)
		if err != nil {
			return fmt.Errorf("macro expansion failed at line %d '%s': %w", lineNum, line, err)
		}

		// Log expansion if it occurred
		if line != expandedLine {
			logger.Debug("Macro expansion applied", "line", lineNum, "original", line, "expanded", expandedLine)
		}

		// Output command line with %%> prefix if echo_commands is enabled
		if echoCommands, _ := vs.Get("_echo_commands"); echoCommands == "true" {
			// Show original line for script debugging
			fmt.Printf("%%%%> %s\n", line)

			// If macro expansion occurred, also show the expanded version for macro debugging
			if line != expandedLine {
				fmt.Printf("%%%%> [expanded] %s\n", expandedLine)
			}
		}

		// Parse the expanded command line
		cmd := parser.ParseInput(expandedLine)
		if cmd == nil {
			return fmt.Errorf("failed to parse expanded command at line %d: '%s'", lineNum, expandedLine)
		}

		logger.Debug("Command parsed", "line", lineNum, "command", cmd.Name, "message", cmd.Message)

		// Execute the command using builtin registry
		// All variables should already be resolved by the command-level interpolation
		err = commands.GetGlobalRegistry().Execute(cmd.Name, cmd.Options, cmd.Message)

		if err != nil {
			return fmt.Errorf("command execution failed at line %d (%s): %w", lineNum, cmd.Name, err)
		}

		logger.Debug("Command executed successfully", "line", lineNum, "command", cmd.Name)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading script file: %w", err)
	}

	// Set success status variables
	if err := vs.SetSystemVariable("_status", "0"); err != nil {
		logger.Debug("Failed to set _status variable", "error", err)
	}

	successMessage := fmt.Sprintf("Script %s executed successfully with macros (%d commands)", scriptPath, commandCount)
	if err := vs.SetSystemVariable("_output", successMessage); err != nil {
		logger.Debug("Failed to set _output variable", "error", err)
	}

	logger.Debug("Enhanced script execution completed successfully", "script", scriptPath, "commands_executed", commandCount)
	return nil
}

// ExecuteScriptContentWithMacros executes script content directly with enhanced command-level macro expansion support.
// This function is similar to ExecuteScriptWithMacros but operates on content rather than files.
//
// Enhanced Pipeline:
//
//	Load Line → Interpolate Command Line (Macro Expansion) → Parse → Execute
//
// Parameters:
//   - content: The script content to execute
//   - scriptName: Name for logging/debugging purposes
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScriptContentWithMacros(content, scriptName string) error {
	logger.Debug("Starting enhanced script content execution with macro support", "script", scriptName)

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
	scanner := bufio.NewScanner(strings.NewReader(content))
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
		logger.Debug("Processing script content line", "script", scriptName, "line_number", lineNum, "original", line)

		// PHASE 1: Command-level interpolation (macro expansion)
		expandedLine, err := is.InterpolateCommandLine(line)
		if err != nil {
			return fmt.Errorf("macro expansion failed at line %d '%s': %w", lineNum, line, err)
		}

		// Log expansion if it occurred
		if line != expandedLine {
			logger.Debug("Macro expansion applied", "script", scriptName, "line", lineNum, "original", line, "expanded", expandedLine)
		}

		// Output command line with %%> prefix if echo_commands is enabled
		if echoCommands, _ := vs.Get("_echo_commands"); echoCommands == "true" {
			// Show original line for script debugging
			fmt.Printf("%%%%> %s\n", line)

			// If macro expansion occurred, also show the expanded version for macro debugging
			if line != expandedLine {
				fmt.Printf("%%%%> [expanded] %s\n", expandedLine)
			}
		}

		// Parse the expanded command line
		cmd := parser.ParseInput(expandedLine)
		if cmd == nil {
			return fmt.Errorf("failed to parse expanded command at line %d: '%s'", lineNum, expandedLine)
		}

		logger.Debug("Command parsed", "script", scriptName, "line", lineNum, "command", cmd.Name, "message", cmd.Message)

		// Execute the command using builtin registry
		// All variables should already be resolved by the command-level interpolation
		err = commands.GetGlobalRegistry().Execute(cmd.Name, cmd.Options, cmd.Message)

		if err != nil {
			return fmt.Errorf("command execution failed at line %d (%s): %w", lineNum, cmd.Name, err)
		}

		logger.Debug("Command executed successfully", "script", scriptName, "line", lineNum, "command", cmd.Name)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading script content: %w", err)
	}

	// Set success status variables
	if err := vs.SetSystemVariable("_status", "0"); err != nil {
		logger.Debug("Failed to set _status variable", "error", err)
	}

	successMessage := fmt.Sprintf("Script content %s executed successfully with macros (%d commands)", scriptName, commandCount)
	if err := vs.SetSystemVariable("_output", successMessage); err != nil {
		logger.Debug("Failed to set _output variable", "error", err)
	}

	logger.Debug("Enhanced script content execution completed successfully", "script", scriptName, "commands_executed", commandCount)
	return nil
}
