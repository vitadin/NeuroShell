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

func TestGetCommand_Name(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, "get", cmd.Name())
}

func TestGetCommand_ParseMode(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestGetCommand_Description(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, "Get a variable", cmd.Description())
}

func TestGetCommand_Usage(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, "\\get[var] or \\get var", cmd.Usage())
}

func TestGetCommand_Execute_BracketSyntax(t *testing.T) {
	cmd := &GetCommand{}

	tests := []struct {
		name          string
		args          map[string]string
		input         string
		setupVars     map[string]string
		expectedVar   string
		expectedValue string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "get existing variable",
			args:          map[string]string{"testvar": ""},
			input:         "",
			setupVars:     map[string]string{"testvar": "testvalue"},
			expectedVar:   "testvar",
			expectedValue: "testvalue",
			wantErr:       false,
		},
		{
			name:          "get system variable @user",
			args:          map[string]string{"@user": ""},
			input:         "",
			setupVars:     map[string]string{},
			expectedVar:   "@user",
			expectedValue: "testuser",
			wantErr:       false,
		},
		{
			name:          "get system variable #test_mode",
			args:          map[string]string{"#test_mode": ""},
			input:         "",
			setupVars:     map[string]string{},
			expectedVar:   "#test_mode",
			expectedValue: "true",
			wantErr:       false,
		},
		{
			name:      "get non-existent variable",
			args:      map[string]string{"nonexistent": ""},
			input:     "",
			setupVars: map[string]string{},
			wantErr:   true,
			errMsg:    "failed to get variable nonexistent",
		},
		{
			name:      "empty args and input",
			args:      map[string]string{},
			input:     "",
			setupVars: map[string]string{},
			wantErr:   true,
			errMsg:    "Usage:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutils.NewMockContextWithVars(tt.setupVars)
			setupGetTestRegistry(t, ctx)

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
				expectedOutput := fmt.Sprintf("%s = %s\n", tt.expectedVar, tt.expectedValue)
				assert.Equal(t, expectedOutput, outputStr)
			}
		})
	}
}

func TestGetCommand_Execute_SpaceSyntax(t *testing.T) {
	cmd := &GetCommand{}

	tests := []struct {
		name          string
		args          map[string]string
		input         string
		setupVars     map[string]string
		expectedVar   string
		expectedValue string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "get variable with space syntax",
			args:          map[string]string{},
			input:         "testvar",
			setupVars:     map[string]string{"testvar": "testvalue"},
			expectedVar:   "testvar",
			expectedValue: "testvalue",
			wantErr:       false,
		},
		{
			name:          "get system variable with space syntax",
			args:          map[string]string{},
			input:         "@pwd",
			setupVars:     map[string]string{},
			expectedVar:   "@pwd",
			expectedValue: "/test/pwd",
			wantErr:       false,
		},
		{
			name:          "get variable with extra spaces",
			args:          map[string]string{},
			input:         "  testvar  ",
			setupVars:     map[string]string{"testvar": "testvalue"},
			expectedVar:   "testvar",
			expectedValue: "testvalue",
			wantErr:       false,
		},
		{
			name:          "get first word from multi-word input",
			args:          map[string]string{},
			input:         "testvar extra words ignored",
			setupVars:     map[string]string{"testvar": "testvalue"},
			expectedVar:   "testvar",
			expectedValue: "testvalue",
			wantErr:       false,
		},
		{
			name:      "get non-existent variable with space syntax",
			args:      map[string]string{},
			input:     "nonexistent",
			setupVars: map[string]string{},
			wantErr:   true,
			errMsg:    "failed to get variable nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutils.NewMockContextWithVars(tt.setupVars)
			setupGetTestRegistry(t, ctx)

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
				expectedOutput := fmt.Sprintf("%s = %s\n", tt.expectedVar, tt.expectedValue)
				assert.Equal(t, expectedOutput, outputStr)
			}
		})
	}
}

func TestGetCommand_Execute_PrioritizeBracketSyntax(t *testing.T) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"bracketvar": "bracketvalue",
		"spacevar":   "spacevalue",
	})
	setupGetTestRegistry(t, ctx)

	// When both args and input are provided, args (bracket syntax) should take priority
	args := map[string]string{"bracketvar": ""}
	input := "spacevar"

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
	// Should use bracket syntax (bracketvar), not space syntax (spacevar)
	expectedOutput := "bracketvar = bracketvalue\n"
	assert.Equal(t, expectedOutput, outputStr)
}

func TestGetCommand_Execute_ContextError(t *testing.T) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContext()
	setupGetTestRegistry(t, ctx)

	// Set up context to return an error
	ctx.SetGetVariableError(fmt.Errorf("context error"))

	args := map[string]string{"testvar": ""}

	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get variable testvar")
	assert.Contains(t, err.Error(), "context error")
}

func TestGetCommand_Execute_EmptyVariableName(t *testing.T) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContext()
	setupGetTestRegistry(t, ctx)

	tests := []struct {
		name  string
		args  map[string]string
		input string
	}{
		{
			name:  "empty args and empty input",
			args:  map[string]string{},
			input: "",
		},
		{
			name:  "empty args and whitespace input",
			args:  map[string]string{},
			input: "   ",
		},
		{
			name:  "args with empty key",
			args:  map[string]string{"": ""},
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Usage:")
		})
	}
}

func TestGetCommand_Execute_VariableWithSpecialCharacters(t *testing.T) {
	cmd := &GetCommand{}

	specialVars := map[string]string{
		"var_with_underscores": "underscore_value",
		"var-with-dashes":      "dash-value",
		"var123":               "numeric_value",
		"UPPERCASE_VAR":        "upper_value",
	}

	ctx := testutils.NewMockContextWithVars(specialVars)
	setupGetTestRegistry(t, ctx)

	for varName, expectedValue := range specialVars {
		t.Run(fmt.Sprintf("get_%s", varName), func(t *testing.T) {
			args := map[string]string{varName: ""}

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
			expectedOutput := fmt.Sprintf("%s = %s\n", varName, expectedValue)
			assert.Equal(t, expectedOutput, outputStr)
		})
	}
}

func TestGetCommand_Execute_EmptyVariableValue(t *testing.T) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"empty_var": "",
	})
	setupGetTestRegistry(t, ctx)

	args := map[string]string{"empty_var": ""}

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
	expectedOutput := "empty_var = \n"
	assert.Equal(t, expectedOutput, outputStr)
}

// Benchmark tests
func BenchmarkGetCommand_Execute_BracketSyntax(b *testing.B) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"benchvar": "benchvalue",
	})

	// Setup for benchmark (simplified)
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().InitializeAll(ctx)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()
	args := map[string]string{"benchvar": ""}

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

func BenchmarkGetCommand_Execute_SpaceSyntax(b *testing.B) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"benchvar": "benchvalue",
	})

	// Setup for benchmark (simplified)
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().InitializeAll(ctx)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()
	input := "benchvar"

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, input)
	}
}

func BenchmarkGetCommand_Execute_SystemVariable(b *testing.B) {
	cmd := &GetCommand{}
	ctx := testutils.NewMockContext()

	// Setup for benchmark (simplified)
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().InitializeAll(ctx)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()
	args := map[string]string{"@user": ""}

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

// setupGetTestRegistry sets up a test environment with variable service
func setupGetTestRegistry(t *testing.T, ctx neurotypes.Context) {
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
