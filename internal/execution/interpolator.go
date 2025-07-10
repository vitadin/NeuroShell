package execution

import (
	"strings"

	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
)

// CoreInterpolator handles variable interpolation and macro expansion directly within the state machine.
// It implements a recursive interpolation system that scans for ${...} patterns and replaces them
// with variable values, using empty strings for undefined variables.
type CoreInterpolator struct {
	// Direct reference to the context for fast variable access
	context *context.NeuroContext
}

// NewCoreInterpolator creates a new CoreInterpolator with direct context access.
func NewCoreInterpolator(ctx *context.NeuroContext) *CoreInterpolator {
	return &CoreInterpolator{
		context: ctx,
	}
}

// HasVariables checks if the given text contains any variable references.
// This is used by the state machine to determine if interpolation is needed.
func (ci *CoreInterpolator) HasVariables(text string) bool {
	return strings.Contains(text, "${")
}

// InterpolateCommandLine performs command-level macro expansion on an entire command line.
// This enables variables to contain complete commands that are expanded before parsing.
//
// Examples:
//   \set[cmd="\echo"]
//   ${cmd} Hello World    # Expands to: \echo Hello World
//
//   \set[full_cmd="\echo[style=red] DEBUG:"]
//   ${full_cmd} Message   # Expands to: \echo[style=red] DEBUG: Message
//
// Returns:
//   - expanded: The text with variables expanded
//   - hasVariables: Whether any variables were found and expanded
//   - error: Any error that occurred during expansion
func (ci *CoreInterpolator) InterpolateCommandLine(line string) (string, bool, error) {
	if !ci.HasVariables(line) {
		return line, false, nil
	}

	logger.Debug("Command-level interpolation starting", "input", line)

	expanded := ci.ExpandVariables(line)
	hasVariables := line != expanded
	
	if hasVariables {
		logger.Debug("Command-level interpolation completed", "original", line, "expanded", expanded)
	}

	return expanded, hasVariables, nil
}

// InterpolateCommand interpolates all parts of a parsed command structure.
// This is used for parameter-level interpolation after the command has been parsed.
func (ci *CoreInterpolator) InterpolateCommand(cmd *parser.Command) (*parser.Command, error) {
	if cmd == nil {
		return nil, nil
	}

	// Create new command with interpolated values
	interpolatedCmd := &parser.Command{
		Name:           cmd.Name, // Don't interpolate command name
		Message:        ci.ExpandVariables(cmd.Message),
		BracketContent: ci.ExpandVariables(cmd.BracketContent),
		Options:        make(map[string]string),
		ParseMode:      cmd.ParseMode,
		OriginalText:   cmd.OriginalText,
	}

	// Interpolate option values
	for key, value := range cmd.Options {
		interpolatedCmd.Options[key] = ci.ExpandVariables(value)
	}

	return interpolatedCmd, nil
}

// ExpandVariables performs recursive variable expansion using a stack-based algorithm.
// It handles nested variables like ${a_${b_${c}}} by expanding innermost variables first.
// If a variable doesn't exist, it's replaced with an empty string.
// The process repeats recursively until no more variables are found.
func (ci *CoreInterpolator) ExpandVariables(text string) string {
	maxIterations := 10 // Prevent infinite loops
	
	for i := 0; i < maxIterations; i++ {
		if !ci.HasVariables(text) {
			break
		}
		
		before := text
		text = ci.expandWithStack(text)
		
		// If no change occurred, break to prevent infinite loop
		if text == before {
			break
		}
		
		logger.Debug("Variable expansion iteration", "iteration", i+1, "result", text)
	}
	
	return text
}

// expandWithStack performs stack-based variable expansion to handle nested variables.
// For ${a_${b_${c}}}, it expands ${c} first, then ${b_x}, then ${a_y}.
// This correctly handles nested variable names where inner variables form part of outer variable names.
func (ci *CoreInterpolator) expandWithStack(text string) string {
	var stack []int // Stack of opening ${} positions
	i := 0
	
	for i < len(text) {
		// Look for ${
		if i < len(text)-1 && text[i] == '$' && text[i+1] == '{' {
			// Push the position of the opening ${
			stack = append(stack, i)
			i += 2 // Skip past ${
		} else if text[i] == '}' && len(stack) > 0 {
			// Pop from stack to get the matching ${
			openPos := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			
			// Extract variable name between the matched ${ and }
			varName := text[openPos+2 : i]
			
			// Get variable value (empty string if not found)
			value := ci.getVariableValue(varName)
			
			// Replace the entire ${varName} with its value in the text
			// This is the key insight: we rebuild the text with the expansion
			newText := text[:openPos] + value + text[i+1:]
			
			// Restart parsing from the beginning with the new text
			// This ensures we handle cascading expansions correctly
			return ci.expandWithStack(newText)
		} else {
			i++
		}
	}
	
	// If we get here, no complete ${} pattern was found, return as-is
	return text
}

// getVariableValue retrieves the value of a variable from the context.
// Returns empty string if the variable doesn't exist (no error in macro system).
func (ci *CoreInterpolator) getVariableValue(varName string) string {
	// Handle empty variable name
	if varName == "" {
		return ""
	}
	
	// Get variable value from context
	value, _ := ci.context.GetVariable(varName)
	return value
}

// ExpandVariablesWithLimit performs variable expansion with a custom recursion limit.
// This is useful for preventing infinite recursion in macro expansion.
func (ci *CoreInterpolator) ExpandVariablesWithLimit(text string, maxIterations int) string {
	if maxIterations <= 0 {
		maxIterations = 10 // Default limit
	}
	
	for i := 0; i < maxIterations; i++ {
		if !ci.HasVariables(text) {
			break
		}
		
		before := text
		text = ci.expandWithStack(text)
		
		// If no change occurred, break to prevent infinite loop
		if text == before {
			break
		}
		
		logger.Debug("Variable expansion iteration", "iteration", i+1, "result", text)
	}
	
	return text
}