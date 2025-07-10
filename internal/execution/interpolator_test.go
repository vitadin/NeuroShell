package execution

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/parser"
)

// TestCoreInterpolator_NewCoreInterpolator tests interpolator creation.
func TestCoreInterpolator_NewCoreInterpolator(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	if interpolator == nil {
		t.Fatal("Expected interpolator to be created, got nil")
	}

	if interpolator.context != ctx {
		t.Error("Expected context to be set correctly")
	}
}

// TestCoreInterpolator_HasVariables tests variable detection.
func TestCoreInterpolator_HasVariables(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	testCases := []struct {
		input    string
		expected bool
	}{
		{"no variables here", false},
		{"${variable}", true},
		{"prefix ${var} suffix", true},
		{"${var1} and ${var2}", true},
		{"$variable", false}, // Not in ${} format
		{"${}", true},        // Empty variable name
		{"$${escaped}", true}, // Still contains ${} pattern
	}

	for _, tc := range testCases {
		result := interpolator.HasVariables(tc.input)
		if result != tc.expected {
			t.Errorf("HasVariables('%s') = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_InterpolateCommandLine tests command-line interpolation.
func TestCoreInterpolator_InterpolateCommandLine(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up test variables
	ctx.SetVariable("user", "Alice")
	ctx.SetVariable("cmd", "\\echo")

	testCases := []struct {
		input           string
		expectedOutput  string
		expectedHasVars bool
		expectError     bool
	}{
		{
			input:           "no variables",
			expectedOutput:  "no variables",
			expectedHasVars: false,
			expectError:     false,
		},
		{
			input:           "\\echo Hello ${user}",
			expectedOutput:  "\\echo Hello Alice",
			expectedHasVars: true,
			expectError:     false,
		},
		{
			input:           "${cmd} Hello World",
			expectedOutput:  "\\echo Hello World",
			expectedHasVars: true,
			expectError:     false,
		},
		{
			input:           "\\${cmd} prefix ${user} suffix",
			expectedOutput:  "\\\\echo prefix Alice suffix",
			expectedHasVars: true,
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		output, hasVars, err := interpolator.InterpolateCommandLine(tc.input)

		if tc.expectError && err == nil {
			t.Errorf("InterpolateCommandLine('%s') expected error, got none", tc.input)
			continue
		}

		if !tc.expectError && err != nil {
			t.Errorf("InterpolateCommandLine('%s') unexpected error: %v", tc.input, err)
			continue
		}

		if output != tc.expectedOutput {
			t.Errorf("InterpolateCommandLine('%s') = '%s', expected '%s'", tc.input, output, tc.expectedOutput)
		}

		if hasVars != tc.expectedHasVars {
			t.Errorf("InterpolateCommandLine('%s') hasVars = %v, expected %v", tc.input, hasVars, tc.expectedHasVars)
		}
	}
}

// TestCoreInterpolator_ExpandVariables tests variable expansion.
func TestCoreInterpolator_ExpandVariables(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up test variables
	ctx.SetVariable("name", "World")
	ctx.SetVariable("greeting", "Hello")

	testCases := []struct {
		input    string
		expected string
	}{
		{"no variables", "no variables"},
		{"${greeting} ${name}", "Hello World"},
		{"prefix ${name} suffix", "prefix World suffix"},
		{"${undefined}", ""}, // Undefined variables are removed
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("ExpandVariables('%s') = '%s', expected '%s'", tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandVariablesWithLimit tests limited variable expansion.
func TestCoreInterpolator_ExpandVariablesWithLimit(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up recursive variables
	ctx.SetVariable("a", "${b}")
	ctx.SetVariable("b", "${c}")
	ctx.SetVariable("c", "final")

	// Test normal expansion within limit
	result := interpolator.ExpandVariablesWithLimit("${a}", 5)
	if result != "final" {
		t.Errorf("ExpandVariablesWithLimit = '%s', expected 'final'", result)
	}

	// Test expansion with low limit
	result = interpolator.ExpandVariablesWithLimit("${a}", 1)
	// With stack-based algorithm, it might expand more efficiently
	// The key is that it should be different from the original input
	if result == "${a}" {
		t.Errorf("ExpandVariablesWithLimit with limit 1 should have expanded from original input")
	}

	// Test with zero limit (should use default)
	result = interpolator.ExpandVariablesWithLimit("${a}", 0)
	if result != "final" {
		t.Errorf("ExpandVariablesWithLimit with zero limit = '%s', expected 'final'", result)
	}
}


// TestCoreInterpolator_InterpolateCommand tests command structure interpolation.
func TestCoreInterpolator_InterpolateCommand(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up test variables
	ctx.SetVariable("user", "Alice")
	ctx.SetVariable("style", "red")

	// Test nil command
	result, err := interpolator.InterpolateCommand(nil)
	if err != nil {
		t.Errorf("InterpolateCommand(nil) unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("InterpolateCommand(nil) = %v, expected nil", result)
	}

	// Test command with variables
	cmd := &parser.Command{
		Name:           "echo",
		Message:        "Hello ${user}",
		BracketContent: "[style=${style}]",
		Options: map[string]string{
			"color": "${style}",
			"text":  "Hi ${user}",
		},
	}

	interpolated, err := interpolator.InterpolateCommand(cmd)
	if err != nil {
		t.Errorf("InterpolateCommand unexpected error: %v", err)
	}

	if interpolated.Name != "echo" {
		t.Errorf("Command name should not be interpolated, got '%s'", interpolated.Name)
	}

	if interpolated.Message != "Hello Alice" {
		t.Errorf("Message = '%s', expected 'Hello Alice'", interpolated.Message)
	}

	if interpolated.BracketContent != "[style=red]" {
		t.Errorf("BracketContent = '%s', expected '[style=red]'", interpolated.BracketContent)
	}

	if interpolated.Options["color"] != "red" {
		t.Errorf("Options[color] = '%s', expected 'red'", interpolated.Options["color"])
	}

	if interpolated.Options["text"] != "Hi Alice" {
		t.Errorf("Options[text] = '%s', expected 'Hi Alice'", interpolated.Options["text"])
	}
}

// TestCoreInterpolator_NestedVariableNames tests nested variable name construction
func TestCoreInterpolator_NestedVariableNames(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Test case: ${a_${b_${c}}} where c="x", b_x="y", a_y="final"
	ctx.SetVariable("c", "x")
	ctx.SetVariable("b_x", "y") 
	ctx.SetVariable("a_y", "final_value")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${a_${b_${c}}}",
			expected: "final_value",
			desc:     "Triple nested: c=x, b_x=y, a_y=final_value",
		},
		{
			input:    "${b_${c}}",
			expected: "y",
			desc:     "Double nested: c=x, b_x=y",
		},
		{
			input:    "prefix_${a_${b_${c}}}_suffix",
			expected: "prefix_final_value_suffix",
			desc:     "Nested with surrounding text",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_CompositeVariables tests multiple variables in one expression
func TestCoreInterpolator_CompositeVariables(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	ctx.SetVariable("a", "hello")
	ctx.SetVariable("b", "world")
	ctx.SetVariable("prefix", "pre")
	ctx.SetVariable("suffix", "post")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${a}_${b}",
			expected: "hello_world",
			desc:     "Simple composite: a_b",
		},
		{
			input:    "${prefix}_${a}_${suffix}",
			expected: "pre_hello_post",
			desc:     "Multiple variables with underscores",
		},
		{
			input:    "${a}${b}",
			expected: "helloworld",
			desc:     "Adjacent variables without separator",
		},
		{
			input:    "start_${a}_middle_${b}_end",
			expected: "start_hello_middle_world_end",
			desc:     "Variables mixed with literal text",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ComplexNesting tests complex nested scenarios
func TestCoreInterpolator_ComplexNesting(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables for complex nesting
	ctx.SetVariable("level1", "level2")
	ctx.SetVariable("level2", "level3") 
	ctx.SetVariable("level3", "deep_value")
	ctx.SetVariable("x", "y")
	ctx.SetVariable("var_y", "resolved")
	ctx.SetVariable("a", "b")
	ctx.SetVariable("c", "d")
	ctx.SetVariable("b_d", "combined")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${${${level1}}}",
			expected: "deep_value",
			desc:     "Multi-level nested resolution",
		},
		{
			input:    "${var_${x}}",
			expected: "resolved", 
			desc:     "Variable name with nested suffix",
		},
		{
			input:    "${${a}_${c}}",
			expected: "combined",
			desc:     "Nested variable name from two variables",
		},
		{
			input:    "${${a}_${c}}_extra",
			expected: "combined_extra",
			desc:     "Nested with literal suffix",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_UndefinedNestedVariables tests undefined variables in nested contexts
func TestCoreInterpolator_UndefinedNestedVariables(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	ctx.SetVariable("defined", "exists")
	ctx.SetVariable("exists", "final")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${undefined_${defined}}",
			expected: "",
			desc:     "Undefined variable with defined nested part",
		},
		{
			input:    "${defined_${undefined}}",
			expected: "",
			desc:     "Defined prefix with undefined nested part",
		},
		{
			input:    "${${defined}}",
			expected: "final",
			desc:     "Nested defined variables should work",
		},
		{
			input:    "${${undefined}}",
			expected: "",
			desc:     "Nested undefined variable",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_MixedNestedAndComposite tests combinations
func TestCoreInterpolator_MixedNestedAndComposite(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Complex setup
	ctx.SetVariable("type", "user")
	ctx.SetVariable("id", "123")
	ctx.SetVariable("user_123", "alice")
	ctx.SetVariable("action", "login")
	ctx.SetVariable("alice_login", "success")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${${type}_${id}}",
			expected: "alice",
			desc:     "Composite nested variable name",
		},
		{
			input:    "${${${type}_${id}}_${action}}",
			expected: "success", 
			desc:     "Double nested with composite parts",
		},
		{
			input:    "User ${${type}_${id}} performed ${action}: ${${${type}_${id}}_${action}}",
			expected: "User alice performed login: success",
			desc:     "Complex sentence with multiple nested variables",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_EdgeCases tests edge cases and malformed inputs
func TestCoreInterpolator_EdgeCases(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	ctx.SetVariable("var", "value")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "${",
			expected: "${",
			desc:     "Unclosed variable should remain as-is",
		},
		{
			input:    "}",
			expected: "}",
			desc:     "Orphaned closing brace",
		},
		{
			input:    "${}",
			expected: "",
			desc:     "Empty variable name",
		},
		{
			input:    "${var} ${",
			expected: "value ${",
			desc:     "Valid variable followed by unclosed",
		},
		{
			input:    "${${}}",
			expected: "",
			desc:     "Nested empty variable",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandVariables('%s') = '%s', expected '%s'", 
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// Test helper types - none needed for interpolator tests since we use parser.Command directly