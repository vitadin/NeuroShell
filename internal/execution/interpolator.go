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
	// Maximum iterations for recursive expansion to prevent infinite loops
	maxIter int
}

// NewCoreInterpolator creates a new CoreInterpolator with direct context access and default settings.
func NewCoreInterpolator(ctx *context.NeuroContext) *CoreInterpolator {
	return &CoreInterpolator{
		context: ctx,
		maxIter: 10, // Default safe limit for recursive expansion
	}
}

// NewCoreInterpolatorWithLimit creates a new CoreInterpolator with a custom iteration limit.
func NewCoreInterpolatorWithLimit(ctx *context.NeuroContext, maxIter int) *CoreInterpolator {
	if maxIter <= 0 {
		maxIter = 10 // Ensure a safe default
	}
	return &CoreInterpolator{
		context: ctx,
		maxIter: maxIter,
	}
}

// SetMaxIterations updates the maximum iteration limit for recursive expansion.
func (ci *CoreInterpolator) SetMaxIterations(maxIter int) {
	if maxIter <= 0 {
		maxIter = 10 // Ensure a safe default
	}
	ci.maxIter = maxIter
}

// GetMaxIterations returns the current maximum iteration limit.
func (ci *CoreInterpolator) GetMaxIterations() int {
	return ci.maxIter
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
// Uses the interpolator's configured maxIter limit.
func (ci *CoreInterpolator) ExpandVariables(text string) string {
	return ci.ExpandWithLimit(text, ci.maxIter)
}

// ExpandOnce performs a single pass of stack-based variable expansion.
// It expands all complete variables found in the input during one scan, but does not 
// recursively expand the variable values themselves.
//
// Algorithm uses a stack to hold string fragments and a PENDING flag to track variable state:
// - On "${": Push "${" to stack, set PENDING=true  
// - On "}" with PENDING=true: Pop back to "${", extract variable name, expand it, set PENDING=false
// - On "}" with PENDING=false: Treat as literal character
// - Other characters: Push to stack as-is
//
// Examples:
//   Input: "${a}" where a="hello" -> Output: "hello"
//   Input: "${a}" where a="${b}" -> Output: "${b}" (value not recursively expanded)
//   Input: "${a}_${b}" where a="x", b="y" -> Output: "x_y" (both variables expanded)
//   Input: "${${a_${b}}_${b}}" where b="x" -> Output: "${${a_x}_x}" (only complete vars)
//
// This function provides fine-grained control over expansion depth.
func (ci *CoreInterpolator) ExpandOnce(text string) string {
	var stack []string
	pending := false
	
	for i := 0; i < len(text); i++ {
		// Look for ${
		if i < len(text)-1 && text[i] == '$' && text[i+1] == '{' {
			stack = append(stack, "${")
			pending = true
			i++ // Skip the '{'
		} else if text[i] == '}' && pending {
			// Pop back to "${" marker to extract variable name
			varName := ""
			for len(stack) > 0 && stack[len(stack)-1] != "${" {
				varName = stack[len(stack)-1] + varName
				stack = stack[:len(stack)-1]
			}
			// Remove the "${" marker
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			
			// Get variable value and push to stack
			value := ci.getVariableValue(varName)
			stack = append(stack, value)
			pending = false
		} else {
			// Regular character or "}" when not pending
			stack = append(stack, string(text[i]))
		}
	}
	
	// Join all stack elements to form final result
	return strings.Join(stack, "")
}

// ExpandWithLimit performs iterative variable expansion with a maximum iteration limit.
// It calls ExpandOnce repeatedly until no more variables are found or the iteration limit is reached.
// This provides protection against circular references while allowing full recursive expansion.
//
// Examples:
//   Input: "${a}" where a="${b}", b="final" with limit=10 -> Output: "final"
//   Input: "${a}" where a="${b}", b="${a}" with limit=10 -> Output: "${a}" or "${b}" (circular ref stopped)
//
// This function gives users control over how deep the expansion should go and provides
// safety against infinite loops from circular variable references.
func (ci *CoreInterpolator) ExpandWithLimit(text string, maxIterations int) string {
	if maxIterations <= 0 {
		maxIterations = 10 // Default safe limit
	}
	
	for iteration := 0; iteration < maxIterations; iteration++ {
		textBefore := text
		text = ci.ExpandOnce(text)
		
		// If no change occurred, we're done (no more variables or circular reference)
		if text == textBefore {
			break
		}
		
		logger.Debug("Variable expansion iteration", "iteration", iteration+1, "result", text)
	}
	
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
	return ci.ExpandWithLimit(text, maxIterations)
}

// ExpandWithStack is a backward compatibility wrapper around ExpandWithLimit.
// Deprecated: Use ExpandOnce for single-pass expansion or ExpandWithLimit for iterative expansion.
func (ci *CoreInterpolator) ExpandWithStack(text string, maxIterations int) string {
	return ci.ExpandWithLimit(text, maxIterations)
}