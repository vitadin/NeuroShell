package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestRenderCommand_BasicFunctionality(t *testing.T) {
	cmd := &RenderCommand{}

	assert.Equal(t, "render", cmd.Name())
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
	assert.NotEmpty(t, cmd.Description())
	assert.NotEmpty(t, cmd.Usage())
}

func TestRenderCommand_Execute_BasicStyling(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	ctx := context.GetGlobalContext()

	tests := []struct {
		name  string
		args  map[string]string
		input string
	}{
		{
			name:  "simple text rendering",
			args:  map[string]string{},
			input: "Hello world",
		},
		{
			name: "text with theme",
			args: map[string]string{
				"theme": "dark",
			},
			input: "Styled text",
		},
		{
			name: "text with bold style",
			args: map[string]string{
				"style": "bold",
			},
			input: "Bold text",
		},
		{
			name: "text with boolean options",
			args: map[string]string{
				"bold":      "true",
				"italic":    "true",
				"underline": "true",
			},
			input: "Multi-styled text",
		},
		{
			name: "text with color",
			args: map[string]string{
				"color":      "#ff0000",
				"background": "#0000ff",
			},
			input: "Colorful text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			require.NoError(t, err)

			// Check that result was stored in _output
			output, err := ctx.GetVariable("_output")
			require.NoError(t, err)
			assert.Contains(t, output, tt.input, "Original text should be preserved in output")
		})
	}
}

func TestRenderCommand_Execute_KeywordHighlighting(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	ctx := context.GetGlobalContext()

	tests := []struct {
		name     string
		args     map[string]string
		input    string
		keywords []string
	}{
		{
			name: "highlight single keyword",
			args: map[string]string{
				"keywords": "[\\get]",
			},
			input:    "Use \\get command",
			keywords: []string{"\\get"},
		},
		{
			name: "highlight multiple keywords",
			args: map[string]string{
				"keywords": "[\\get, \\set]",
			},
			input:    "Use \\get and \\set commands",
			keywords: []string{"\\get", "\\set"},
		},
		{
			name: "highlight with spaces in array",
			args: map[string]string{
				"keywords": "[ \\get , \\set , \\vars ]",
			},
			input:    "Commands: \\get, \\set, \\vars",
			keywords: []string{"\\get", "\\set", "\\vars"},
		},
		{
			name: "highlight with quoted keywords",
			args: map[string]string{
				"keywords": "[\"\\get\", '\\set']",
			},
			input:    "Use \\get and \\set",
			keywords: []string{"\\get", "\\set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			require.NoError(t, err)

			// Check that result was stored in _output
			output, err := ctx.GetVariable("_output")
			require.NoError(t, err)

			// All keywords should be present in output
			for _, keyword := range tt.keywords {
				assert.Contains(t, output, keyword, "Keyword should be present in output")
			}
		})
	}
}

func TestRenderCommand_Execute_VariableInterpolation(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	ctx := context.GetGlobalContext()

	// Set up test variables
	err := ctx.SetVariable("test_cmd", "\\session-new")
	require.NoError(t, err)

	tests := []struct {
		name  string
		args  map[string]string
		input string
	}{
		{
			name: "interpolated keywords",
			args: map[string]string{
				"keywords": "[${test_cmd}, \\get]",
			},
			input: "Use \\session-new and \\get commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			require.NoError(t, err)

			// Check that result was stored
			output, err := ctx.GetVariable("_output")
			require.NoError(t, err)
			assert.NotEmpty(t, output)
		})
	}
}

func TestRenderCommand_Execute_OutputVariable(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	ctx := context.GetGlobalContext()

	tests := []struct {
		name      string
		args      map[string]string
		input     string
		outputVar string
		systemVar bool
	}{
		{
			name:      "default output variable",
			args:      map[string]string{},
			input:     "test text",
			outputVar: "_output",
			systemVar: true,
		},
		{
			name: "custom user variable",
			args: map[string]string{
				"to": "my_styled_text",
			},
			input:     "test text",
			outputVar: "my_styled_text",
			systemVar: false,
		},
		{
			name: "system error variable",
			args: map[string]string{
				"to": "_error",
			},
			input:     "error text",
			outputVar: "_error",
			systemVar: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			require.NoError(t, err)

			// Check that result was stored in correct variable
			output, err := ctx.GetVariable(tt.outputVar)
			require.NoError(t, err)
			assert.Contains(t, output, tt.input)
		})
	}
}

func TestRenderCommand_Execute_SilentMode(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	ctx := context.GetGlobalContext()

	tests := []struct {
		name   string
		args   map[string]string
		input  string
		silent bool
	}{
		{
			name: "normal output",
			args: map[string]string{
				"silent": "false",
			},
			input:  "visible text",
			silent: false,
		},
		{
			name: "silent output",
			args: map[string]string{
				"silent": "true",
			},
			input:  "hidden text",
			silent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			require.NoError(t, err)

			// Result should still be stored in variable regardless of silent mode
			output, err := ctx.GetVariable("_output")
			require.NoError(t, err)
			assert.Contains(t, output, tt.input)
		})
	}
}

func TestRenderCommand_Execute_ErrorCases(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	cmd := &RenderCommand{}
	context.GetGlobalContext()

	tests := []struct {
		name    string
		args    map[string]string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			args:    map[string]string{},
			input:   "",
			wantErr: true,
		},
		{
			name: "invalid boolean for silent",
			args: map[string]string{
				"silent": "maybe",
			},
			input:   "test",
			wantErr: true,
		},
		{
			name: "invalid boolean for bold",
			args: map[string]string{
				"bold": "yes",
			},
			input:   "test",
			wantErr: false, // Should not error, just ignore invalid boolean
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderCommand_ThemeHandling(t *testing.T) {
	// Setup services
	setupTestServices(t)
	defer cleanupTestServices()

	tests := []struct {
		name        string
		themeName   string
		expectPlain bool
		description string
	}{
		{
			name:        "default theme",
			themeName:   "default",
			expectPlain: false,
			description: "Should use default theme",
		},
		{
			name:        "dark theme",
			themeName:   "dark",
			expectPlain: false,
			description: "Should use dark theme",
		},
		{
			name:        "dark1 alias",
			themeName:   "dark1",
			expectPlain: false,
			description: "Should use dark theme via dark1 alias",
		},
		{
			name:        "plain theme",
			themeName:   "plain",
			expectPlain: true,
			description: "Should use plain theme (no styling)",
		},
		{
			name:        "empty theme",
			themeName:   "",
			expectPlain: true,
			description: "Should use plain theme when empty",
		},
		{
			name:        "invalid theme",
			themeName:   "invalid",
			expectPlain: true,
			description: "Should fall back to plain theme for invalid themes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			themeService, err := services.GetGlobalThemeService()
			require.NoError(t, err)

			theme := themeService.GetThemeByName(tt.themeName)
			assert.NotNil(t, theme, "Theme should never be nil")

			if tt.expectPlain {
				assert.Equal(t, "plain", theme.Name, tt.description)
			} else {
				assert.NotEqual(t, "plain", theme.Name, tt.description)
			}

			// Test that theme can be used to render text
			testText := "test"
			styledText := theme.Command.Render(testText)
			assert.NotEmpty(t, styledText, "Styled text should not be empty")
		})
	}
}

// Helper functions for test setup

func setupTestServices(t *testing.T) {
	// Reset global registry for testing
	services.GlobalRegistry = services.NewRegistry()

	// Register required services
	err := services.GlobalRegistry.RegisterService(services.NewThemeService())
	require.NoError(t, err)

	err = services.GlobalRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GlobalRegistry.RegisterService(services.NewInterpolationService())
	require.NoError(t, err)

	// Initialize services
	ctx := context.New()
	// Set the test context as global context
	context.SetGlobalContext(ctx)
	err = services.GlobalRegistry.InitializeAll(ctx)
	require.NoError(t, err)
}

func cleanupTestServices() {
	// Reset global context
	context.ResetGlobalContext()
}
