package builtin

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestSetCommand_Name(t *testing.T) {
	cmd := &SetCommand{}
	assert.Equal(t, "set", cmd.Name())
}

func TestSetCommand_ParseMode(t *testing.T) {
	cmd := &SetCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestSetCommand_Description(t *testing.T) {
	cmd := &SetCommand{}
	assert.Equal(t, "Set a variable", cmd.Description())
}

func TestSetCommand_Usage(t *testing.T) {
	cmd := &SetCommand{}
	assert.Equal(t, "\\set[var=value] or \\set var value", cmd.Usage())
}

func TestSetCommand_Execute_BracketSyntax(t *testing.T) {
	cmd := &SetCommand{}

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		wantErr        bool
		errMsg         string
		expectedVars   map[string]string
		expectedOutput string
	}{
		{
			name:           "set single variable",
			args:           map[string]string{"testvar": "testvalue"},
			input:          "",
			expectedVars:   map[string]string{"testvar": "testvalue"},
			expectedOutput: "Setting testvar = testvalue\n",
		},
		{
			name: "set multiple variables",
			args: map[string]string{
				"var1": "value1",
				"var2": "value2",
			},
			input: "",
			expectedVars: map[string]string{
				"var1": "value1",
				"var2": "value2",
			},
			// Note: order may vary due to map iteration
		},
		{
			name:           "set variable with empty value",
			args:           map[string]string{"emptyvar": ""},
			input:          "",
			expectedVars:   map[string]string{"emptyvar": ""},
			expectedOutput: "Setting emptyvar = \n",
		},
		{
			name:           "set variable with special characters",
			args:           map[string]string{"special": "value with spaces & symbols!"},
			input:          "",
			expectedVars:   map[string]string{"special": "value with spaces & symbols!"},
			expectedOutput: "Setting special = value with spaces & symbols!\n",
		},
		{
			name:           "set variable with quotes",
			args:           map[string]string{"quoted": "\"quoted value\""},
			input:          "",
			expectedVars:   map[string]string{"quoted": "\"quoted value\""},
			expectedOutput: "Setting quoted = \"quoted value\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutils.NewMockContext()
			context.SetGlobalContext(ctx)

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(tt.args, tt.input)

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify variables were set
				for expectedVar, expectedValue := range tt.expectedVars {
					actualValue, err := ctx.GetVariable(expectedVar)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, actualValue)
				}

				// For single variable tests, check exact output
				if len(tt.args) == 1 && tt.expectedOutput != "" {
					assert.Equal(t, tt.expectedOutput, outputStr)
				} else if len(tt.args) > 1 {
					// For multiple variables, just check that all variables are mentioned
					for varName := range tt.args {
						assert.Contains(t, outputStr, fmt.Sprintf("Setting %s =", varName))
					}
				}
			}
		})
	}
}

func TestSetCommand_Execute_SpaceSyntax(t *testing.T) {
	cmd := &SetCommand{}

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		wantErr        bool
		errMsg         string
		expectedVar    string
		expectedValue  string
		expectedOutput string
	}{
		{
			name:           "set variable with space syntax",
			args:           map[string]string{},
			input:          "testvar testvalue",
			expectedVar:    "testvar",
			expectedValue:  "testvalue",
			expectedOutput: "Setting testvar = testvalue\n",
		},
		{
			name:           "set variable with multiple words in value",
			args:           map[string]string{},
			input:          "testvar this is a multi word value",
			expectedVar:    "testvar",
			expectedValue:  "this is a multi word value",
			expectedOutput: "Setting testvar = this is a multi word value\n",
		},
		{
			name:           "set variable with single word (no value)",
			args:           map[string]string{},
			input:          "testvar",
			expectedVar:    "testvar",
			expectedValue:  "",
			expectedOutput: "Setting testvar = \n",
		},
		{
			name:           "set variable with spaces around",
			args:           map[string]string{},
			input:          "  testvar   testvalue  ",
			expectedVar:    "testvar",
			expectedValue:  "testvalue  ",
			expectedOutput: "Setting testvar = testvalue  \n",
		},
		{
			name:           "set variable with quoted value",
			args:           map[string]string{},
			input:          "testvar \"quoted value\"",
			expectedVar:    "testvar",
			expectedValue:  "\"quoted value\"",
			expectedOutput: "Setting testvar = \"quoted value\"\n",
		},
		{
			name:           "set variable with special characters",
			args:           map[string]string{},
			input:          "special_var value!@#$%^&*()",
			expectedVar:    "special_var",
			expectedValue:  "value!@#$%^&*()",
			expectedOutput: "Setting special_var = value!@#$%^&*()\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutils.NewMockContext()
			context.SetGlobalContext(ctx)

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(tt.args, tt.input)

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify variable was set
				actualValue, err := ctx.GetVariable(tt.expectedVar)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, actualValue)

				// Check output
				assert.Equal(t, tt.expectedOutput, outputStr)
			}
		})
	}
}

func TestSetCommand_Execute_PrioritizeBracketSyntax(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	// When both args and input are provided, args (bracket syntax) should take priority
	args := map[string]string{"bracketvar": "bracketvalue"}
	input := "spacevar spacevalue"

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(args, input)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Should only set bracket syntax variable
	bracketValue, err := ctx.GetVariable("bracketvar")
	assert.NoError(t, err)
	assert.Equal(t, "bracketvalue", bracketValue)

	// Space syntax variable should not be set
	_, err = ctx.GetVariable("spacevar")
	assert.Error(t, err)

	// Output should only mention bracket variable
	assert.Contains(t, outputStr, "Setting bracketvar = bracketvalue")
	assert.NotContains(t, outputStr, "spacevar")
}

func TestSetCommand_Execute_ContextError(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	// Set up context to return an error
	ctx.SetSetVariableError(fmt.Errorf("context error"))

	tests := []struct {
		name  string
		args  map[string]string
		input string
	}{
		{
			name:  "bracket syntax with context error",
			args:  map[string]string{"testvar": "testvalue"},
			input: "",
		},
		{
			name:  "space syntax with context error",
			args:  map[string]string{},
			input: "testvar testvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to set variable")
			assert.Contains(t, err.Error(), "context error")
		})
	}
}

func TestSetCommand_Execute_EmptyInputAndArgs(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	err := cmd.Execute(map[string]string{}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestSetCommand_Execute_VariableOverwrite(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"existing": "oldvalue",
	})
	context.SetGlobalContext(ctx)

	args := map[string]string{"existing": "newvalue"}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(args, "")

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Verify variable was overwritten
	actualValue, err := ctx.GetVariable("existing")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", actualValue)

	// Check output
	expectedOutput := "Setting existing = newvalue\n"
	assert.Equal(t, expectedOutput, outputStr)
}

func TestSetCommand_Execute_SpecialVariableNames(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	specialNames := []struct {
		name  string
		value string
	}{
		{"var_with_underscores", "underscore_value"},
		{"var-with-dashes", "dash-value"},
		{"var123", "numeric_value"},
		{"UPPERCASE_VAR", "upper_value"},
		{"mixedCaseVar", "mixed_value"},
		{"var.with.dots", "dot_value"},
	}

	for _, test := range specialNames {
		t.Run(fmt.Sprintf("set_%s", test.name), func(t *testing.T) {
			args := map[string]string{test.name: test.value}

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(args, "")

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			assert.NoError(t, err)

			// Verify variable was set
			actualValue, err := ctx.GetVariable(test.name)
			assert.NoError(t, err)
			assert.Equal(t, test.value, actualValue)

			// Check output
			expectedOutput := fmt.Sprintf("Setting %s = %s\n", test.name, test.value)
			assert.Equal(t, expectedOutput, outputStr)
		})
	}
}

func TestSetCommand_Execute_LargeValues(t *testing.T) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	// Test with large value
	largeValue := make([]byte, 1000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	largeValueStr := string(largeValue)

	args := map[string]string{"large_var": largeValueStr}

	// Capture stdout
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(args, "")

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	assert.NoError(t, err)

	// Verify variable was set
	actualValue, err := ctx.GetVariable("large_var")
	assert.NoError(t, err)
	assert.Equal(t, largeValueStr, actualValue)
}

// Benchmark tests
func BenchmarkSetCommand_Execute_BracketSyntax(b *testing.B) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)
	args := map[string]string{"benchvar": "benchvalue"}

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

func BenchmarkSetCommand_Execute_SpaceSyntax(b *testing.B) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)
	input := "benchvar benchvalue"

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, input)
	}
}

func BenchmarkSetCommand_Execute_MultipleVariables(b *testing.B) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)
	args := map[string]string{
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
		"var4": "value4",
		"var5": "value5",
	}

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

func BenchmarkSetCommand_Execute_LargeValue(b *testing.B) {
	cmd := &SetCommand{}
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	// Create large value
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	args := map[string]string{"large_var": string(largeValue)}

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		_ = devNull.Close()
		os.Stdout = originalStdout
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

func TestSetCommand_Execute_WhitelistedGlobalVariables(t *testing.T) {
	cmd := &SetCommand{}

	tests := []struct {
		name           string
		args           map[string]string
		input          string
		wantErr        bool
		errMsg         string
		expectedVars   map[string]string
		expectedOutput string
	}{
		{
			name:           "set whitelisted _style variable",
			args:           map[string]string{"_style": "dark"},
			input:          "",
			expectedVars:   map[string]string{"_style": "dark"},
			expectedOutput: "Setting _style = dark\n",
		},
		{
			name:           "set _style to empty string",
			args:           map[string]string{"_style": ""},
			input:          "",
			expectedVars:   map[string]string{"_style": ""},
			expectedOutput: "Setting _style = \n",
		},
		{
			name:           "set _style with space syntax",
			args:           map[string]string{},
			input:          "_style light",
			expectedVars:   map[string]string{"_style": "light"},
			expectedOutput: "Setting _style = light\n",
		},
		{
			name:    "try to set non-whitelisted _secret variable",
			args:    map[string]string{"_secret": "value"},
			input:   "",
			wantErr: true,
			errMsg:  "cannot set system variable: _secret",
		},
		{
			name:    "try to set non-whitelisted _config variable",
			args:    map[string]string{"_config": "value"},
			input:   "",
			wantErr: true,
			errMsg:  "cannot set system variable: _config",
		},
		{
			name:    "try to set @pwd system variable",
			args:    map[string]string{"@pwd": "/tmp"},
			input:   "",
			wantErr: true,
			errMsg:  "cannot set system variable: @pwd",
		},
		{
			name:    "try to set #session_id system variable",
			args:    map[string]string{"#session_id": "fake"},
			input:   "",
			wantErr: true,
			errMsg:  "cannot set system variable: #session_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use real context for whitelist testing, not mock context
			ctx := context.New()
			setupSetTestRegistry(t, ctx)

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(tt.args, tt.input)

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify variables were set
				for expectedVar, expectedValue := range tt.expectedVars {
					actualValue, err := ctx.GetVariable(expectedVar)
					assert.NoError(t, err)
					assert.Equal(t, expectedValue, actualValue)
				}

				// Check output
				if tt.expectedOutput != "" {
					assert.Equal(t, tt.expectedOutput, outputStr)
				}
			}
		})
	}
}

func TestSetCommand_Execute_StyleVariableDefaultInitialization(t *testing.T) {
	// Test that _style is initialized to empty string by default
	ctx := testutils.NewMockContext()
	context.SetGlobalContext(ctx)

	// _style should exist and be empty by default
	value, err := ctx.GetVariable("_style")
	assert.NoError(t, err)
	assert.Equal(t, "", value)
}

func setupSetTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}
