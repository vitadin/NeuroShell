package builtin

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestVarsCommand_Name(t *testing.T) {
	cmd := &VarsCommand{}
	if got := cmd.Name(); got != "vars" {
		t.Errorf("VarsCommand.Name() = %v, want %v", got, "vars")
	}
}

func TestVarsCommand_ParseMode(t *testing.T) {
	cmd := &VarsCommand{}
	if got := cmd.ParseMode(); got != neurotypes.ParseModeKeyValue {
		t.Errorf("VarsCommand.ParseMode() = %v, want %v", got, neurotypes.ParseModeKeyValue)
	}
}

func TestVarsCommand_Description(t *testing.T) {
	cmd := &VarsCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("VarsCommand.Description() should not be empty")
	}
	if !strings.Contains(desc, "variable") {
		t.Errorf("VarsCommand.Description() = %v, should contain 'variable'", desc)
	}
}

func TestVarsCommand_Usage(t *testing.T) {
	cmd := &VarsCommand{}
	usage := cmd.Usage()
	if usage == "" {
		t.Error("VarsCommand.Usage() should not be empty")
	}
	if !strings.Contains(usage, "\\vars") {
		t.Errorf("VarsCommand.Usage() = %v, should contain '\\vars'", usage)
	}
}

func TestVarsCommand_Execute_NoVariables(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Should show system variables (always present) but no user variables
	if !strings.Contains(output, "System Variables:") {
		t.Errorf("Expected 'System Variables:' section, got: %s", output)
	}
	if strings.Contains(output, "User Variables:") {
		t.Errorf("Should not show 'User Variables:' section when no user variables exist, got: %s", output)
	}
	if !strings.Contains(output, "Total:") {
		t.Errorf("Expected total count in output, got: %s", output)
	}
}

func TestVarsCommand_Execute_WithUserVariables(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Set some user variables
	require.NoError(t, ctx.SetVariable("name", "John"))
	require.NoError(t, ctx.SetVariable("project", "NeuroShell"))

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Check that user variables are displayed
	if !strings.Contains(output, "User Variables:") {
		t.Errorf("Expected 'User Variables:' section, got: %s", output)
	}
	if !strings.Contains(output, "name") || !strings.Contains(output, "John") {
		t.Errorf("Expected variable 'name = John' in output, got: %s", output)
	}
	if !strings.Contains(output, "project") || !strings.Contains(output, "NeuroShell") {
		t.Errorf("Expected variable 'project = NeuroShell' in output, got: %s", output)
	}
}

func TestVarsCommand_Execute_TypeFilter_User(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Set mixed variables
	require.NoError(t, ctx.SetVariable("user_var", "value1"))
	require.NoError(t, ctx.SetSystemVariable("_system_var", "value2"))

	args := map[string]string{"type": "user"}

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(args, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Should only show user variables
	if !strings.Contains(output, "user_var") {
		t.Errorf("Expected user variable 'user_var' in output, got: %s", output)
	}
	if strings.Contains(output, "_system_var") {
		t.Errorf("Should not show system variable '_system_var' in user filter, got: %s", output)
	}
	if strings.Contains(output, "System Variables:") {
		t.Errorf("Should not show 'System Variables:' section with user filter, got: %s", output)
	}
}

func TestVarsCommand_Execute_TypeFilter_System(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Set mixed variables
	require.NoError(t, ctx.SetVariable("user_var", "value1"))
	require.NoError(t, ctx.SetSystemVariable("_system_var", "value2"))

	args := map[string]string{"type": "system"}

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(args, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Should only show system variables
	if !strings.Contains(output, "_system_var") {
		t.Errorf("Expected system variable '_system_var' in output, got: %s", output)
	}
	if strings.Contains(output, "user_var") {
		t.Errorf("Should not show user variable 'user_var' in system filter, got: %s", output)
	}
	if strings.Contains(output, "User Variables:") {
		t.Errorf("Should not show 'User Variables:' section with system filter, got: %s", output)
	}
}

func TestVarsCommand_Execute_PatternFilter(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Set variables with different patterns
	require.NoError(t, ctx.SetVariable("test_var1", "value1"))
	require.NoError(t, ctx.SetVariable("test_var2", "value2"))
	require.NoError(t, ctx.SetVariable("other_var", "value3"))

	args := map[string]string{"pattern": "^test_"}

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(args, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Should only show variables matching pattern
	if !strings.Contains(output, "test_var1") {
		t.Errorf("Expected 'test_var1' in pattern-filtered output, got: %s", output)
	}
	if !strings.Contains(output, "test_var2") {
		t.Errorf("Expected 'test_var2' in pattern-filtered output, got: %s", output)
	}
	if strings.Contains(output, "other_var") {
		t.Errorf("Should not show 'other_var' in pattern-filtered output, got: %s", output)
	}
}

func TestVarsCommand_Execute_InvalidRegex(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	args := map[string]string{"pattern": "[invalid"}

	err := cmd.Execute(args, "")
	if err == nil {
		t.Error("VarsCommand.Execute() should return error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid regex pattern") {
		t.Errorf("Expected 'invalid regex pattern' error, got: %v", err)
	}
}

func TestVarsCommand_Execute_CombinedFilters(t *testing.T) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Setup test registry with variable service
	setupVarsTestRegistry(t, ctx)

	// Set mixed variables
	require.NoError(t, ctx.SetVariable("user_test", "value1"))
	require.NoError(t, ctx.SetVariable("user_other", "value2"))
	require.NoError(t, ctx.SetSystemVariable("_test_var", "value3"))

	args := map[string]string{
		"pattern": "test",
		"type":    "user",
	}

	// Capture stdout
	output := captureOutput(func() {
		err := cmd.Execute(args, "")
		if err != nil {
			t.Errorf("VarsCommand.Execute() error = %v, want nil", err)
		}
	})

	// Should only show user variables matching pattern
	if !strings.Contains(output, "user_test") {
		t.Errorf("Expected 'user_test' in filtered output, got: %s", output)
	}
	if strings.Contains(output, "user_other") {
		t.Errorf("Should not show 'user_other' (doesn't match pattern), got: %s", output)
	}
	if strings.Contains(output, "_test_var") {
		t.Errorf("Should not show '_test_var' (system variable), got: %s", output)
	}
}

func TestVarsCommand_MatchesTypeFilter(t *testing.T) {
	cmd := &VarsCommand{}

	tests := []struct {
		name     string
		varName  string
		varType  string
		expected bool
	}{
		{"user var with user filter", "normal_var", "user", true},
		{"user var with system filter", "normal_var", "system", false},
		{"user var with all filter", "normal_var", "all", true},
		{"@ system var with system filter", "@pwd", "system", true},
		{"@ system var with user filter", "@pwd", "user", false},
		{"# system var with system filter", "#session_id", "system", true},
		{"_ system var with system filter", "_status", "system", true},
		{"unknown type defaults to all", "any_var", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.matchesTypeFilter(tt.varName, tt.varType)
			if result != tt.expected {
				t.Errorf("matchesTypeFilter(%s, %s) = %v, want %v", tt.varName, tt.varType, result, tt.expected)
			}
		})
	}
}

func TestVarsCommand_DisplayVariables_SmartTruncation(t *testing.T) {
	cmd := &VarsCommand{}

	tests := []struct {
		name         string
		value        string
		shouldShow   []string // Strings that should appear in output
		shouldntShow []string // Strings that should NOT appear in output
	}{
		{
			name:       "short value unchanged",
			value:      "short value",
			shouldShow: []string{"short value"},
		},
		{
			name:         "long value smart truncated",
			value:        strings.Repeat("a", 100) + "END",
			shouldShow:   []string{"aaa", "...", "END", "(length: 103 chars)"},
			shouldntShow: []string{strings.Repeat("a", 50)}, // Middle part should not appear
		},
		{
			name:       "value with newlines",
			value:      "Line 1\nLine 2\nLine 3",
			shouldShow: []string{"Line 1\\nLine 2\\nLine 3"},
		},
		{
			name:       "value with tabs",
			value:      "Text\twith\ttabs",
			shouldShow: []string{"Text\\twith\\ttabs"},
		},
		{
			name:         "very long value with special chars",
			value:        "Start" + strings.Repeat("x", 90) + "\n\tEnd",
			shouldShow:   []string{"Start", "...", "\\n\\tEnd", "(length: 100 chars)"},
			shouldntShow: []string{strings.Repeat("x", 40)}, // Middle part should not appear
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]string{
				"test_var": tt.value,
			}

			// Create a plain theme for testing
			plainTheme := &services.Theme{
				Name: "plain",
			}

			// Capture stdout
			output := captureOutput(func() {
				cmd.displayVariables(vars, plainTheme)
			})

			// Check that expected strings appear
			for _, expected := range tt.shouldShow {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, output)
				}
			}

			// Check that unexpected strings don't appear
			for _, unexpected := range tt.shouldntShow {
				if strings.Contains(output, unexpected) {
					t.Errorf("Expected output to NOT contain '%s', got: %s", unexpected, output)
				}
			}
		})
	}
}

func TestVarsCommand_FormatValueForDisplay(t *testing.T) {
	cmd := &VarsCommand{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short value",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "exactly max length",
			input:    strings.Repeat("a", 80),
			expected: strings.Repeat("a", 80),
		},
		{
			name:     "just over max length",
			input:    strings.Repeat("a", 85),
			expected: strings.Repeat("a", 30) + "..." + strings.Repeat("a", 20) + " (length: 85 chars)",
		},
		{
			name:     "newlines replaced",
			input:    "line1\nline2\nline3",
			expected: "line1\\nline2\\nline3",
		},
		{
			name:     "tabs and carriage returns",
			input:    "text\twith\r\nvarious\twhitespace",
			expected: "text\\twith\\r\\nvarious\\twhitespace",
		},
		{
			name:     "very long with newlines",
			input:    "Start of message" + strings.Repeat("x", 70) + "\nEnd of message",
			expected: "Start of messagexxxxxxxxxxxxxx...xxxx\\nEnd of message (length: 101 chars)",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.formatValueForDisplay(tt.input)
			if result != tt.expected {
				t.Errorf("formatValueForDisplay(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkVarsCommand_Execute(b *testing.B) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Set up some test variables
	for i := 0; i < 100; i++ {
		require.NoError(b, ctx.SetVariable(fmt.Sprintf("var%d", i), fmt.Sprintf("value%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Discard output to avoid IO overhead in benchmark
		old := os.Stdout
		os.Stdout, _ = os.Open(os.DevNull)
		_ = cmd.Execute(map[string]string{}, "")
		os.Stdout = old
	}
}

func BenchmarkVarsCommand_PatternFilter(b *testing.B) {
	cmd := &VarsCommand{}
	ctx := context.New()

	// Set up test variables
	for i := 0; i < 100; i++ {
		require.NoError(b, ctx.SetVariable(fmt.Sprintf("test_var%d", i), fmt.Sprintf("value%d", i)))
		require.NoError(b, ctx.SetVariable(fmt.Sprintf("other_var%d", i), fmt.Sprintf("value%d", i)))
	}

	args := map[string]string{"pattern": "^test_"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Discard output to avoid IO overhead in benchmark
		old := os.Stdout
		os.Stdout, _ = os.Open(os.DevNull)
		_ = cmd.Execute(args, "")
		os.Stdout = old
	}
}

// setupVarsTestRegistry sets up a test environment with variable service
func setupVarsTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Register interpolation service
	err = services.GetGlobalRegistry().RegisterService(services.NewInterpolationService())
	require.NoError(t, err)

	// Register theme service
	err = services.GetGlobalRegistry().RegisterService(services.NewThemeService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*VarsCommand)(nil)
