package execution

import (
	"fmt"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/parser"
)

// BenchmarkCoreInterpolator_SimpleVariable benchmarks basic variable expansion
func BenchmarkCoreInterpolator_SimpleVariable(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("user", "Alice")
	text := "Hello ${user}!"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_MultipleVariables benchmarks multiple variable expansion
func BenchmarkCoreInterpolator_MultipleVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("user", "Alice")
	ctx.SetVariable("action", "running")
	ctx.SetVariable("object", "tests")
	text := "${user} is ${action} ${object} successfully"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_NestedVariables benchmarks recursive variable expansion
func BenchmarkCoreInterpolator_NestedVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("a", "${b}")
	ctx.SetVariable("b", "${c}")
	ctx.SetVariable("c", "final_value")
	text := "Result: ${a}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_DeepNesting benchmarks deep recursive expansion
func BenchmarkCoreInterpolator_DeepNesting(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	// Create a chain of 8 nested variables
	ctx.SetVariable("var1", "${var2}")
	ctx.SetVariable("var2", "${var3}")
	ctx.SetVariable("var3", "${var4}")
	ctx.SetVariable("var4", "${var5}")
	ctx.SetVariable("var5", "${var6}")
	ctx.SetVariable("var6", "${var7}")
	ctx.SetVariable("var7", "${var8}")
	ctx.SetVariable("var8", "deep_value")
	text := "Deep: ${var1}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_NoVariables benchmarks text without variables
func BenchmarkCoreInterpolator_NoVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	text := "This is plain text without any variables to expand"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_LargeText benchmarks large text with variables
func BenchmarkCoreInterpolator_LargeText(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("service", "NeuroShell")
	ctx.SetVariable("version", "1.0.0")
	
	// Create a large text with variables scattered throughout
	baseText := "Welcome to ${service} version ${version}. "
	largeText := ""
	for i := 0; i < 100; i++ {
		largeText += baseText
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(largeText)
	}
}

// BenchmarkCoreInterpolator_UndefinedVariables benchmarks undefined variable handling
func BenchmarkCoreInterpolator_UndefinedVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	text := "Value: ${undefined_var} and ${another_undefined}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_HasVariables benchmarks variable detection
func BenchmarkCoreInterpolator_HasVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	testCases := []string{
		"no variables here",
		"${variable}",
		"prefix ${var} suffix",
		"multiple ${var1} and ${var2}",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, text := range testCases {
			_ = interpolator.HasVariables(text)
		}
	}
}

// BenchmarkCoreInterpolator_InterpolateCommandLine benchmarks command-line interpolation
func BenchmarkCoreInterpolator_InterpolateCommandLine(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("cmd", "echo")
	ctx.SetVariable("message", "Hello World")
	commandLine := "\\${cmd} ${message}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = interpolator.InterpolateCommandLine(commandLine)
	}
}

// BenchmarkCoreInterpolator_InterpolateCommand benchmarks command structure interpolation
func BenchmarkCoreInterpolator_InterpolateCommand(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("user", "Alice")
	ctx.SetVariable("style", "red")
	
	cmd := &parser.Command{
		Name:           "echo",
		Message:        "Hello ${user}",
		BracketContent: "[style=${style}]",
		Options: map[string]string{
			"color": "${style}",
			"text":  "Hi ${user}",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interpolator.InterpolateCommand(cmd)
	}
}

// BenchmarkCoreInterpolator_SystemVariables benchmarks system variable expansion
func BenchmarkCoreInterpolator_SystemVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	text := "User: ${@user}, Date: ${@date}, Session: ${#session_id}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_MixedContent benchmarks realistic mixed content
func BenchmarkCoreInterpolator_MixedContent(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("user", "developer")
	ctx.SetVariable("project", "NeuroShell")
	ctx.SetVariable("task", "interpolation")
	
	text := "\\echo[style=green] ${user} is working on ${project} ${task} module at ${@date}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// Benchmark comparison with different iteration limits
func BenchmarkCoreInterpolator_WithLimit1(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("a", "${b}")
	ctx.SetVariable("b", "${c}")
	ctx.SetVariable("c", "final")
	text := "${a}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariablesWithLimit(text, 1)
	}
}

func BenchmarkCoreInterpolator_WithLimit5(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("a", "${b}")
	ctx.SetVariable("b", "${c}")
	ctx.SetVariable("c", "final")
	text := "${a}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariablesWithLimit(text, 5)
	}
}

func BenchmarkCoreInterpolator_WithLimit10(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("a", "${b}")
	ctx.SetVariable("b", "${c}")
	ctx.SetVariable("c", "final")
	text := "${a}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariablesWithLimit(text, 10)
	}
}

// Memory allocation benchmarks
func BenchmarkCoreInterpolator_MemoryAllocation(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("var", "value")
	text := "prefix ${var} suffix"
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// Benchmark with varying text sizes
func BenchmarkCoreInterpolator_SmallText(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("x", "y")
	text := "${x}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

func BenchmarkCoreInterpolator_MediumText(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("service", "NeuroShell")
	ctx.SetVariable("user", "developer")
	text := "Welcome to ${service}! Hello ${user}, ready to work?"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

func BenchmarkCoreInterpolator_LargeTextExpansion(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("var", "expanded_value")
	
	// Create text with 1000 characters and 10 variables
	text := ""
	for i := 0; i < 10; i++ {
		text += fmt.Sprintf("This is section %d with variable ${var} and some additional text to make it longer. ", i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_NestedVariableNames benchmarks nested variable name construction
func BenchmarkCoreInterpolator_NestedVariableNames(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("c", "x")
	ctx.SetVariable("b_x", "y") 
	ctx.SetVariable("a_y", "final_value")
	text := "${a_${b_${c}}}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_CompositeVariables benchmarks composite variable expansion
func BenchmarkCoreInterpolator_CompositeVariables(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("a", "hello")
	ctx.SetVariable("b", "world")
	ctx.SetVariable("c", "test")
	text := "${a}_${b}_${c}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_DeepNesting benchmarks very deep nesting
func BenchmarkCoreInterpolator_VeryDeepNesting(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	// ${${${${level1}}}}
	ctx.SetVariable("level1", "level2")
	ctx.SetVariable("level2", "level3") 
	ctx.SetVariable("level3", "level4")
	ctx.SetVariable("level4", "deep_value")
	text := "${${${${level1}}}}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// BenchmarkCoreInterpolator_ComplexMixed benchmarks complex mixed scenarios
func BenchmarkCoreInterpolator_ComplexMixed(b *testing.B) {
	ctx := context.New()
	interpolator := NewCoreInterpolator(ctx)
	
	ctx.SetVariable("type", "user")
	ctx.SetVariable("id", "123")
	ctx.SetVariable("user_123", "alice")
	ctx.SetVariable("action", "login")
	ctx.SetVariable("alice_login", "success")
	text := "User ${${type}_${id}} performed ${action}: ${${${type}_${id}}_${action}}"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interpolator.ExpandVariables(text)
	}
}

// Benchmark helper types - none needed since we use parser.Command directly