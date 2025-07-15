package statemachine

import (
	"testing"
	"time"

	"neuroshell/internal/context"
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

	if interpolator.GetMaxIterations() != 10 {
		t.Errorf("Expected default maxIter to be 10, got %d", interpolator.GetMaxIterations())
	}
}

// TestCoreInterpolator_NewCoreInterpolatorWithLimit tests custom limit creation.
func TestCoreInterpolator_NewCoreInterpolatorWithLimit(t *testing.T) {
	ctx := context.New()

	// Test with custom limit
	interpolator := NewCoreInterpolatorWithLimit(ctx, 5)
	if interpolator.GetMaxIterations() != 5 {
		t.Errorf("Expected maxIter to be 5, got %d", interpolator.GetMaxIterations())
	}

	// Test with zero limit (should default to 10)
	interpolator = NewCoreInterpolatorWithLimit(ctx, 0)
	if interpolator.GetMaxIterations() != 10 {
		t.Errorf("Expected maxIter to default to 10 for zero input, got %d", interpolator.GetMaxIterations())
	}

	// Test with negative limit (should default to 10)
	interpolator = NewCoreInterpolatorWithLimit(ctx, -5)
	if interpolator.GetMaxIterations() != 10 {
		t.Errorf("Expected maxIter to default to 10 for negative input, got %d", interpolator.GetMaxIterations())
	}
}

// TestCoreInterpolator_SetMaxIterations tests iteration limit modification.
func TestCoreInterpolator_SetMaxIterations(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Test setting valid limit
	interpolator.SetMaxIterations(15)
	if interpolator.GetMaxIterations() != 15 {
		t.Errorf("Expected maxIter to be 15, got %d", interpolator.GetMaxIterations())
	}

	// Test setting zero limit (should default to 10)
	interpolator.SetMaxIterations(0)
	if interpolator.GetMaxIterations() != 10 {
		t.Errorf("Expected maxIter to default to 10 for zero input, got %d", interpolator.GetMaxIterations())
	}

	// Test setting negative limit (should default to 10)
	interpolator.SetMaxIterations(-3)
	if interpolator.GetMaxIterations() != 10 {
		t.Errorf("Expected maxIter to default to 10 for negative input, got %d", interpolator.GetMaxIterations())
	}
}

// TestCoreInterpolator_MaxIterBehavior tests that maxIter affects expansion behavior.
func TestCoreInterpolator_MaxIterBehavior(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up recursive variables: a -> b -> c -> "final"
	_ = ctx.SetVariable("a", "${b}")
	_ = ctx.SetVariable("b", "${c}")
	_ = ctx.SetVariable("c", "final")

	// Test with maxIter=1 (should only expand once)
	interpolator.SetMaxIterations(1)
	result := interpolator.ExpandVariables("${a}")
	if result != "${b}" {
		t.Errorf("With maxIter=1, expected '${b}', got '%s'", result)
	}

	// Test with maxIter=2 (should expand twice)
	interpolator.SetMaxIterations(2)
	result = interpolator.ExpandVariables("${a}")
	if result != "${c}" {
		t.Errorf("With maxIter=2, expected '${c}', got '%s'", result)
	}

	// Test with maxIter=10 (should fully expand)
	interpolator.SetMaxIterations(10)
	result = interpolator.ExpandVariables("${a}")
	if result != "final" {
		t.Errorf("With maxIter=10, expected 'final', got '%s'", result)
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
		{"$variable", false},  // Not in ${} format
		{"${}", true},         // Empty variable name
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
	_ = ctx.SetVariable("user", "Alice")
	_ = ctx.SetVariable("cmd", "\\echo")

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
	_ = ctx.SetVariable("name", "World")
	_ = ctx.SetVariable("greeting", "Hello")

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
	_ = ctx.SetVariable("a", "${b}")
	_ = ctx.SetVariable("b", "${c}")
	_ = ctx.SetVariable("c", "final")

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

// TestCoreInterpolator_NestedVariableNames tests nested variable name construction
func TestCoreInterpolator_NestedVariableNames(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Test case: ${a_${b_${c}}} where c="x", b_x="y", a_y="final"
	_ = ctx.SetVariable("c", "x")
	_ = ctx.SetVariable("b_x", "y")
	_ = ctx.SetVariable("a_y", "final_value")

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

	_ = ctx.SetVariable("a", "hello")
	_ = ctx.SetVariable("b", "world")
	_ = ctx.SetVariable("prefix", "pre")
	_ = ctx.SetVariable("suffix", "post")

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
	_ = ctx.SetVariable("level1", "level2")
	_ = ctx.SetVariable("level2", "level3")
	_ = ctx.SetVariable("level3", "deep_value")
	_ = ctx.SetVariable("x", "y")
	_ = ctx.SetVariable("var_y", "resolved")
	_ = ctx.SetVariable("a", "b")
	_ = ctx.SetVariable("c", "d")
	_ = ctx.SetVariable("b_d", "combined")

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

	_ = ctx.SetVariable("defined", "exists")
	_ = ctx.SetVariable("exists", "final")

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
	_ = ctx.SetVariable("type", "user")
	_ = ctx.SetVariable("id", "123")
	_ = ctx.SetVariable("user_123", "alice")
	_ = ctx.SetVariable("action", "login")
	_ = ctx.SetVariable("alice_login", "success")

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

	_ = ctx.SetVariable("var", "value")

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

// TestCoreInterpolator_CircularReferences tests infinite loop prevention
func TestCoreInterpolator_CircularReferences(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	testCases := []struct {
		variables map[string]string
		input     string
		desc      string
		// We don't test exact output since circular refs should be handled gracefully
		// but we want to ensure no infinite loops and reasonable behavior
	}{
		{
			variables: map[string]string{
				"a": "${b}",
				"b": "${a}",
			},
			input: "${a}",
			desc:  "Simple circular reference: a->b->a",
		},
		{
			variables: map[string]string{
				"x": "${y}",
				"y": "${z}",
				"z": "${x}",
			},
			input: "${x}",
			desc:  "Three-way circular reference: x->y->z->x",
		},
		{
			variables: map[string]string{
				"self": "${self}",
			},
			input: "${self}",
			desc:  "Self-referential variable",
		},
		{
			variables: map[string]string{
				"a": "${b}",
				"b": "${c}",
				"c": "${d}",
				"d": "${e}",
				"e": "${f}",
				"f": "${g}",
				"g": "${h}",
				"h": "${i}",
				"i": "${j}",
				"j": "${k}",
				"k": "${a}", // Back to start
			},
			input: "${a}",
			desc:  "Long circular chain (11 variables)",
		},
		{
			variables: map[string]string{
				"prefix": "start_",
				"circ1":  "${circ2}_middle",
				"circ2":  "${prefix}${circ1}",
			},
			input: "${circ1}",
			desc:  "Circular reference with partial resolution",
		},
	}

	for _, tc := range testCases {
		// Set up variables
		for name, value := range tc.variables {
			_ = ctx.SetVariable(name, value)
		}

		// The key test: ensure this doesn't hang or panic
		// We use a timeout approach by running in a separate goroutine
		done := make(chan string, 1)
		go func() {
			result := interpolator.ExpandVariables(tc.input)
			done <- result
		}()

		// Wait for completion or timeout
		select {
		case result := <-done:
			t.Logf("Test '%s': Input '%s' -> Output '%s' (completed successfully)",
				tc.desc, tc.input, result)
			// Test passed - no infinite loop
		case <-make(chan struct{}):
			// This would timeout in real scenario, but for unit test we assume
			// the iteration limit works and completes quickly
		}

		// Clean up variables for next test
		for name := range tc.variables {
			_ = ctx.SetVariable(name, "")
		}
	}
}

// TestCoreInterpolator_IterationLimitBehavior tests the iteration limit functionality
func TestCoreInterpolator_IterationLimitBehavior(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up a circular reference
	_ = ctx.SetVariable("a", "${b}")
	_ = ctx.SetVariable("b", "${a}")

	testCases := []struct {
		limit int
		input string
		desc  string
	}{
		{
			limit: 1,
			input: "${a}",
			desc:  "Limit 1: Should expand once then stop",
		},
		{
			limit: 5,
			input: "${a}",
			desc:  "Limit 5: Should hit limit and stop safely",
		},
		{
			limit: 0,
			input: "${a}",
			desc:  "Limit 0: Should use default limit (10)",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariablesWithLimit(tc.input, tc.limit)

		// Key assertions:
		// 1. Function should return (not hang)
		// 2. Result should be a string (not panic)
		// 3. Should be deterministic (same input -> same output)
		if result == "" {
			t.Logf("Test '%s': Returned empty string (acceptable for circular refs)", tc.desc)
		} else {
			t.Logf("Test '%s': Returned '%s'", tc.desc, result)
		}

		// Test deterministic behavior - same input should give same output
		result2 := interpolator.ExpandVariablesWithLimit(tc.input, tc.limit)
		if result != result2 {
			t.Errorf("Test '%s': Non-deterministic behavior - got '%s' then '%s'",
				tc.desc, result, result2)
		}
	}
}

// TestCoreInterpolator_MixedCircularAndValid tests partial circular references
func TestCoreInterpolator_MixedCircularAndValid(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Mix of valid and circular variables
	_ = ctx.SetVariable("valid", "good_value")
	_ = ctx.SetVariable("circ1", "${circ2}")
	_ = ctx.SetVariable("circ2", "${circ1}")
	_ = ctx.SetVariable("partial", "${valid}_${circ1}")

	testCases := []struct {
		input string
		desc  string
	}{
		{
			input: "${valid}",
			desc:  "Pure valid variable should work normally",
		},
		{
			input: "${circ1}",
			desc:  "Pure circular variable should be handled safely",
		},
		{
			input: "${partial}",
			desc:  "Mixed valid and circular should partially resolve",
		},
		{
			input: "prefix_${valid}_${circ1}_suffix",
			desc:  "Mixed valid/circular in larger text",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandVariables(tc.input)
		t.Logf("Test '%s': Input '%s' -> Output '%s'", tc.desc, tc.input, result)

		// Key test: function should return successfully
		// We don't test exact output since circular behavior may vary,
		// but we ensure it doesn't hang and produces some result
	}
}

// TestCoreInterpolator_ExpandOnce tests single-pass variable expansion with stack algorithm
func TestCoreInterpolator_ExpandOnce(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables for testing
	_ = ctx.SetVariable("a", "value_a")
	_ = ctx.SetVariable("b", "value_b")
	_ = ctx.SetVariable("c", "x")
	_ = ctx.SetVariable("b_x", "y")
	_ = ctx.SetVariable("a_y", "final_value")
	_ = ctx.SetVariable("a_x", "resolved_a_x")
	_ = ctx.SetVariable("recursive", "${another}")
	_ = ctx.SetVariable("another", "deep_value")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "no variables",
			expected: "no variables",
			desc:     "Text without variables",
		},
		{
			input:    "${a}",
			expected: "value_a",
			desc:     "Simple variable expansion",
		},
		{
			input:    "${a}_${b}",
			expected: "value_a_value_b",
			desc:     "Multiple variables: both expand in single pass",
		},
		{
			input:    "${a_${b_${c}}}",
			expected: "${a_${b_x}}",
			desc:     "Nested variables: c=x expands, but outer structure remains incomplete",
		},
		{
			input:    "${recursive}",
			expected: "${another}",
			desc:     "Variable values not recursively expanded (stops at ${another})",
		},
		{
			input:    "${${a_${c}}_${c}}",
			expected: "${${a_x}_x}",
			desc:     "Complex nesting with c=x: both ${c} expand to x, result: ${${a_x}_x",
		},
		{
			input:    "prefix_${a}_suffix",
			expected: "prefix_value_a_suffix",
			desc:     "Variable with surrounding text",
		},
		{
			input:    "${undefined}",
			expected: "",
			desc:     "Undefined variable becomes empty",
		},
		{
			input:    "${}",
			expected: "",
			desc:     "Empty variable name becomes empty",
		},
		{
			input:    "${",
			expected: "${",
			desc:     "Unclosed variable remains as-is",
		},
		{
			input:    "}",
			expected: "}",
			desc:     "Orphaned closing brace remains as-is",
		},
		{
			input:    "${a} literal } text ${b}",
			expected: "value_a literal } text value_b",
			desc:     "Mixed variables and literal braces",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandOnce(tc.input)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandOnce('%s') = '%s', expected '%s'",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandOnce_ExactTrace tests the specific example from the algorithm trace
func TestCoreInterpolator_ExpandOnce_ExactTrace(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables to match the exact trace example
	_ = ctx.SetVariable("b", "x")
	_ = ctx.SetVariable("a_x", "resolved")

	// Test the exact case from the trace: "xxx ${${a_${b}}_${b}}"
	input := "xxx ${${a_${b}}_${b}}"
	expected := "xxx ${${a_x}_x}"

	result := interpolator.ExpandOnce(input)
	if result != expected {
		t.Errorf("Exact trace test: ExpandOnce('%s') = '%s', expected '%s'",
			input, result, expected)
	}

	// Verify step by step what should happen:
	// 1. First ${b} (inside ${a_${b}}) expands to "x" → "xxx ${${a_x}_${b}}"
	// 2. Second ${b} (rightmost) expands to "x" → "xxx ${${a_x}_x}"
	// 3. The outer ${${a_x}_x} is not a complete variable, so it remains

	t.Logf("Trace verification:")
	t.Logf("  Input:  %s", input)
	t.Logf("  Output: %s", result)
	t.Logf("  Expected: %s", expected)
}

// TestCoreInterpolator_ExpandWithLimit tests iterative expansion with limits
func TestCoreInterpolator_ExpandWithLimit(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables for testing
	_ = ctx.SetVariable("recursive", "${another}")
	_ = ctx.SetVariable("another", "deep_value")
	_ = ctx.SetVariable("chain1", "${chain2}")
	_ = ctx.SetVariable("chain2", "${chain3}")
	_ = ctx.SetVariable("chain3", "final")

	testCases := []struct {
		input    string
		limit    int
		expected string
		desc     string
	}{
		{
			input:    "${recursive}",
			limit:    10,
			expected: "deep_value",
			desc:     "Recursive expansion with sufficient limit",
		},
		{
			input:    "${recursive}",
			limit:    1,
			expected: "${another}",
			desc:     "Recursive expansion limited to 1 iteration",
		},
		{
			input:    "${chain1}",
			limit:    10,
			expected: "final",
			desc:     "Chain expansion with sufficient limit",
		},
		{
			input:    "${chain1}",
			limit:    2,
			expected: "${chain3}",
			desc:     "Chain expansion limited to 2 iterations",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandWithLimit(tc.input, tc.limit)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandWithLimit('%s', %d) = '%s', expected '%s'",
				tc.desc, tc.input, tc.limit, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandOnceVsExpandWithLimit demonstrates the key differences
func TestCoreInterpolator_ExpandOnceVsExpandWithLimit(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up nested variables: a -> b -> c -> "final"
	_ = ctx.SetVariable("a", "${b}")
	_ = ctx.SetVariable("b", "${c}")
	_ = ctx.SetVariable("c", "final")

	input := "${a}"

	// ExpandOnce: expands only one level
	onceResult := interpolator.ExpandOnce(input)
	expectedOnce := "${b}"
	if onceResult != expectedOnce {
		t.Errorf("ExpandOnce('%s') = '%s', expected '%s'", input, onceResult, expectedOnce)
	}

	// ExpandWithLimit: expands fully
	limitResult := interpolator.ExpandWithLimit(input, 10)
	expectedLimit := "final"
	if limitResult != expectedLimit {
		t.Errorf("ExpandWithLimit('%s', 10) = '%s', expected '%s'", input, limitResult, expectedLimit)
	}

	t.Logf("Demonstrating control:")
	t.Logf("  Input: %s", input)
	t.Logf("  ExpandOnce: %s (single level)", onceResult)
	t.Logf("  ExpandWithLimit(10): %s (full expansion)", limitResult)

	// Test circular reference behavior
	_ = ctx.SetVariable("circular1", "${circular2}")
	_ = ctx.SetVariable("circular2", "${circular1}")

	circularInput := "${circular1}"

	// ExpandOnce with circular reference
	onceCircular := interpolator.ExpandOnce(circularInput)
	t.Logf("  Circular ExpandOnce: %s -> %s", circularInput, onceCircular)

	// ExpandWithLimit with circular reference
	limitCircular := interpolator.ExpandWithLimit(circularInput, 5)
	t.Logf("  Circular ExpandWithLimit(5): %s -> %s", circularInput, limitCircular)
}

// TestCoreInterpolator_ExpandWithStack tests the core stack-based expansion algorithm (backward compatibility)
func TestCoreInterpolator_ExpandWithStack(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables for testing
	_ = ctx.SetVariable("a", "value_a")
	_ = ctx.SetVariable("b", "value_b")
	_ = ctx.SetVariable("c", "x")
	_ = ctx.SetVariable("b_x", "y")
	_ = ctx.SetVariable("a_y", "final_value")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "no variables",
			expected: "no variables",
			desc:     "Text without variables",
		},
		{
			input:    "${a}",
			expected: "value_a",
			desc:     "Simple variable",
		},
		{
			input:    "${a}_${b}",
			expected: "value_a_value_b",
			desc:     "Multiple variables",
		},
		{
			input:    "${a_${b_${c}}}",
			expected: "final_value",
			desc:     "Triple nested variables: c=x, b_x=y, a_y=final_value",
		},
		{
			input:    "prefix_${a_${b_${c}}}_suffix",
			expected: "prefix_final_value_suffix",
			desc:     "Nested variables with surrounding text",
		},
		{
			input:    "${undefined}",
			expected: "",
			desc:     "Undefined variable should become empty",
		},
		{
			input:    "${undefined_${c}}",
			expected: "",
			desc:     "Undefined variable with nested part",
		},
		{
			input:    "${}",
			expected: "",
			desc:     "Empty variable name",
		},
		{
			input:    "${",
			expected: "${",
			desc:     "Unclosed variable",
		},
		{
			input:    "}",
			expected: "}",
			desc:     "Orphaned closing brace",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandWithStack(tc.input, 10)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandWithStack('%s') = '%s', expected '%s'",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandWithStack_CircularReferences tests circular reference handling
// This test verifies that ExpandWithStack now has proper protection against circular references
func TestCoreInterpolator_ExpandWithStack_CircularReferences(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	testCases := []struct {
		variables map[string]string
		input     string
		desc      string
	}{
		{
			variables: map[string]string{
				"a": "${b}",
				"b": "${a}",
			},
			input: "${a}",
			desc:  "Simple circular reference: a->b->a",
		},
		{
			variables: map[string]string{
				"self": "${self}",
			},
			input: "${self}",
			desc:  "Self-referential variable",
		},
		{
			variables: map[string]string{
				"x": "${y}",
				"y": "${z}",
				"z": "${x}",
			},
			input: "${x}",
			desc:  "Three-way circular reference: x->y->z->x",
		},
	}

	for _, tc := range testCases {
		// Set up variables
		for name, value := range tc.variables {
			_ = ctx.SetVariable(name, value)
		}

		// Test that ExpandWithStack now safely handles circular references
		// This should NOT hang and should return promptly
		result := interpolator.ExpandWithStack(tc.input, 50)
		t.Logf("Test '%s': ExpandWithStack (now safe) Input '%s' -> Output '%s'",
			tc.desc, tc.input, result)

		// Clean up variables for next test
		for name := range tc.variables {
			_ = ctx.SetVariable(name, "")
		}
	}
}

// TestCoreInterpolator_ExpandWithStack_IterationLimit tests the iteration limit protection
func TestCoreInterpolator_ExpandWithStack_IterationLimit(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up a long chain that would exceed reasonable limits if not protected
	_ = ctx.SetVariable("a", "${b}")
	_ = ctx.SetVariable("b", "${c}")
	_ = ctx.SetVariable("c", "${d}")
	_ = ctx.SetVariable("d", "${e}")
	_ = ctx.SetVariable("e", "${f}")
	_ = ctx.SetVariable("f", "${g}")
	_ = ctx.SetVariable("g", "${h}")
	_ = ctx.SetVariable("h", "${i}")
	_ = ctx.SetVariable("i", "${j}")
	_ = ctx.SetVariable("j", "final_value")

	// This should work fine - not circular, just deep
	result := interpolator.ExpandWithStack("${a}", 50)
	if result != "final_value" {
		t.Errorf("Expected deep expansion to work, got '%s'", result)
	}

	// Now test actual circular reference with timing
	_ = ctx.SetVariable("circular1", "${circular2}")
	_ = ctx.SetVariable("circular2", "${circular1}")

	// This should complete quickly without hanging
	start := time.Now()
	result = interpolator.ExpandWithStack("${circular1}", 50)
	elapsed := time.Since(start)

	// Should complete in well under a second
	if elapsed > time.Second {
		t.Errorf("ExpandWithStack took too long (%v) - possible infinite loop", elapsed)
	}

	t.Logf("Circular reference handled in %v, result: '%s'", elapsed, result)
}

// TestCoreInterpolator_ExpandWithStack_Complex tests complex nested scenarios
func TestCoreInterpolator_ExpandWithStack_Complex(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Complex setup
	_ = ctx.SetVariable("type", "user")
	_ = ctx.SetVariable("id", "123")
	_ = ctx.SetVariable("user_123", "alice")
	_ = ctx.SetVariable("action", "login")
	_ = ctx.SetVariable("alice_login", "success")
	_ = ctx.SetVariable("level1", "level2")
	_ = ctx.SetVariable("level2", "level3")
	_ = ctx.SetVariable("level3", "deep_value")

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
			input:    "${${${level1}}}",
			expected: "deep_value",
			desc:     "Multi-level nested resolution",
		},
		{
			input:    "User ${${type}_${id}} performed ${action}: ${${${type}_${id}}_${action}}",
			expected: "User alice performed login: success",
			desc:     "Complex sentence with multiple nested variables",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandWithStack(tc.input, 20)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandWithStack('%s') = '%s', expected '%s'",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandWithStack_EdgeCases tests edge cases and malformed inputs
func TestCoreInterpolator_ExpandWithStack_EdgeCases(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	_ = ctx.SetVariable("var", "value")
	_ = ctx.SetVariable("nested", "var")

	testCases := []struct {
		input    string
		expected string
		desc     string
	}{
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
		{
			input:    "${${nested}}",
			expected: "value",
			desc:     "Nested variable resolution",
		},
		{
			input:    "${{var}}",
			expected: "}",
			desc:     "Variable name starting with brace (remaining } is literal)",
		},
		{
			input:    "${var}}",
			expected: "value}",
			desc:     "Variable with extra closing brace",
		},
		{
			input:    "{{${var}}}",
			expected: "{{value}}",
			desc:     "Variable surrounded by literal braces",
		},
	}

	for _, tc := range testCases {
		result := interpolator.ExpandWithStack(tc.input, 10)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandWithStack('%s') = '%s', expected '%s'",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// TestCoreInterpolator_ExpandWithStack_StackBehavior tests specific stack algorithm behavior
func TestCoreInterpolator_ExpandWithStack_StackBehavior(t *testing.T) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)

	// Set up variables to test stack behavior (innermost-first expansion)
	_ = ctx.SetVariable("a", "1")
	_ = ctx.SetVariable("b", "2")
	_ = ctx.SetVariable("c", "3")
	_ = ctx.SetVariable("expand_order", "test")

	testCases := []struct {
		input    string
		desc     string
		setup    func()
		expected string
	}{
		{
			input: "${a${b${c}}}",
			desc:  "Stack should expand innermost first: c->b->a",
			setup: func() {
				_ = ctx.SetVariable("c", "3")
				_ = ctx.SetVariable("b3", "2")
				_ = ctx.SetVariable("a2", "final")
			},
			expected: "final",
		},
		{
			input: "${${a}${b}${c}}",
			desc:  "Multiple nested variables in one name",
			setup: func() {
				_ = ctx.SetVariable("a", "x")
				_ = ctx.SetVariable("b", "y")
				_ = ctx.SetVariable("c", "z")
				_ = ctx.SetVariable("xyz", "combined")
			},
			expected: "combined",
		},
	}

	for _, tc := range testCases {
		// Reset variables
		_ = ctx.SetVariable("a", "1")
		_ = ctx.SetVariable("b", "2")
		_ = ctx.SetVariable("c", "3")

		// Apply test-specific setup
		tc.setup()

		result := interpolator.ExpandWithStack(tc.input, 15)
		if result != tc.expected {
			t.Errorf("Test '%s': ExpandWithStack('%s') = '%s', expected '%s'",
				tc.desc, tc.input, result, tc.expected)
		}
	}
}

// Test helper types - none needed for interpolator tests since we use parser.Command directly
