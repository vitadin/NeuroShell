package statemachine

import (
	"bufio"
	"fmt"
	"strings"

	"neuroshell/pkg/neurotypes"
)

// processScriptLoaded handles setup after a script has been loaded.
func (sm *StateMachine) processScriptLoaded() error {
	resolved := sm.getResolvedCommand()
	parsedCmd := sm.getParsedCommand()

	if resolved == nil || (resolved.Type != neurotypes.CommandTypeStdlib && resolved.Type != neurotypes.CommandTypeUser) {
		return fmt.Errorf("no script to load")
	}

	// Setup script parameters
	err := sm.setupScriptParameters(parsedCmd.Options, parsedCmd.Message, parsedCmd.Name)
	if err != nil {
		return fmt.Errorf("failed to setup script parameters: %w", err)
	}

	// Parse script content into executable lines
	lines := sm.parseScriptIntoLines(resolved.ScriptContent)
	sm.setScriptLines(lines)
	sm.setCurrentScriptLine(0)

	return nil
}

// processScriptExecuting handles line-by-line script execution.
func (sm *StateMachine) processScriptExecuting() error {
	lines := sm.getScriptLines()
	currentLineIndex := sm.getCurrentScriptLine()

	if currentLineIndex >= len(lines) {
		// Script finished - cleanup parameters
		err := sm.cleanupScriptParameters()
		if err != nil {
			sm.logger.Error("Failed to cleanup script parameters", "error", err)
		}
		return nil // Will transition to StateCompleted
	}

	// Get current line to execute
	line := lines[currentLineIndex]
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "%%") {
		sm.setCurrentScriptLine(currentLineIndex + 1)
		return nil // Stay in StateScriptExecuting for next line
	}

	// Output command line with %%> prefix if echo_commands is enabled
	if sm.config.EchoCommands {
		fmt.Printf("%%%%> %s\n", line)
	}

	// Save current execution state for recursive call
	savedState := sm.saveExecutionState()

	// Execute script line through state machine recursively
	// This line will go through: StateReceived → StateInterpolating → ... → StateCompleted
	err := sm.Execute(line)

	// Restore execution state after recursive call
	sm.restoreExecutionState(savedState)

	if err != nil {
		return fmt.Errorf("script failed at line %d: %w", currentLineIndex+1, err)
	}

	// Move to next line
	sm.setCurrentScriptLine(currentLineIndex + 1)

	return nil // Stay in StateScriptExecuting for next line
}

// parseScriptIntoLines parses script content into executable lines, handling multiline commands.
func (sm *StateMachine) parseScriptIntoLines(scriptContent string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(scriptContent))

	for scanner.Scan() {
		line := scanner.Text()

		// Handle multiline continuation with ...
		if strings.HasSuffix(strings.TrimSpace(line), "...") {
			// Accumulate multiline command
			var multilineBuilder []string
			multilineBuilder = append(multilineBuilder, line)

			// Continue reading lines until we find one that doesn't end with ...
			for scanner.Scan() {
				nextLine := scanner.Text()
				multilineBuilder = append(multilineBuilder, nextLine)

				if !strings.HasSuffix(strings.TrimSpace(nextLine), "...") {
					break
				}
			}

			// Join and clean up multiline command
			multilineCommand := strings.Join(multilineBuilder, "\n")
			multilineCommand = strings.ReplaceAll(multilineCommand, "...\n", " ")
			multilineCommand = strings.ReplaceAll(multilineCommand, "...", " ")
			lines = append(lines, strings.TrimSpace(multilineCommand))
		} else {
			lines = append(lines, line)
		}
	}

	return lines
}

// setupScriptParameters sets up parameters for script execution.
func (sm *StateMachine) setupScriptParameters(args map[string]string, input string, commandName string) error {
	// Use the state machine's context directly for parameter setup
	ctx := sm.context

	// Set standard script parameters
	_ = ctx.SetSystemVariable("_0", commandName) // Command name
	_ = ctx.SetSystemVariable("_1", input)       // Input parameter
	_ = ctx.SetSystemVariable("_*", input)       // All positional args

	// Set named arguments as variables (use regular SetVariable for user-defined names)
	var namedArgs []string
	for key, value := range args {
		_ = ctx.SetVariable(key, value)
		namedArgs = append(namedArgs, key+"="+value)
	}
	_ = ctx.SetSystemVariable("_@", strings.Join(namedArgs, " "))

	return nil
}

// cleanupScriptParameters removes script parameters after execution.
func (sm *StateMachine) cleanupScriptParameters() error {
	// Use the state machine's context directly for parameter cleanup
	ctx := sm.context

	// Clear standard script parameters
	parameterNames := []string{"_0", "_1", "_*", "_@"}
	for _, name := range parameterNames {
		_ = ctx.SetSystemVariable(name, "")
	}

	return nil
}
