package builtin

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellPromptCommand_Name(t *testing.T) {
	cmd := &ShellPromptCommand{}
	assert.Equal(t, "shell-prompt", cmd.Name())
}

func TestShellPromptCommand_Execute_SetLines(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	err := services.GetGlobalRegistry().RegisterService(services.NewShellPromptService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() { services.SetGlobalRegistry(oldServiceRegistry) }()

	cmd := &ShellPromptCommand{}
	options := map[string]string{
		"lines": "3",
	}

	err = cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify lines count was set
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "3", linesCount)
}

func TestShellPromptCommand_Execute_SetLineTemplates(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	err := services.GetGlobalRegistry().RegisterService(services.NewShellPromptService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() { services.SetGlobalRegistry(oldServiceRegistry) }()

	cmd := &ShellPromptCommand{}
	options := map[string]string{
		"lines": "2",
		"line1": "${@pwd} [${#session_name}]",
		"line2": "❯ ",
	}

	err = cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify templates were set
	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "${@pwd} [${#session_name}]", line1)

	line2, err := ctx.GetVariable("_prompt_line2")
	assert.NoError(t, err)
	assert.Equal(t, "❯ ", line2)
}

func TestShellPromptCommand_Execute_InvalidLines(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry
	cmd := &ShellPromptCommand{}

	tests := []struct {
		name        string
		linesValue  string
		expectError bool
	}{
		{"valid 1", "1", false},
		{"valid 5", "5", false},
		{"invalid 0", "0", true},
		{"invalid 6", "6", true},
		{"invalid text", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := map[string]string{"lines": tt.linesValue}
			err := cmd.Execute(options, "")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShellPromptShowCommand_Name(t *testing.T) {
	cmd := &ShellPromptShowCommand{}
	assert.Equal(t, "shell-prompt-show", cmd.Name())
}

func TestShellPromptShowCommand_Execute(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	// Pre-configure some prompt settings
	_ = ctx.SetVariable("_prompt_lines_count", "2")
	_ = ctx.SetVariable("_prompt_line1", "test line 1")
	_ = ctx.SetVariable("_prompt_line2", "test line 2")

	cmd := &ShellPromptShowCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)
}

func TestShellPromptPresetCommand_Name(t *testing.T) {
	cmd := &ShellPromptPresetCommand{}
	assert.Equal(t, "shell-prompt-preset", cmd.Name())
}

func TestShellPromptPresetCommand_Execute_Minimal(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{"style": "minimal"}

	err := cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify minimal preset was applied
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "1", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "> ", line1)
}

func TestShellPromptPresetCommand_Execute_Default(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{"style": "default"}

	err := cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify default preset was applied
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "${@pwd} [${#session_name:-no-session}]", line1)

	line2, err := ctx.GetVariable("_prompt_line2")
	assert.NoError(t, err)
	assert.Equal(t, "neuro> ", line2)
}

func TestShellPromptPresetCommand_Execute_Developer(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{"style": "developer"}

	err := cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify developer preset was applied
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "${@pwd} ${@status}", line1)

	line2, err := ctx.GetVariable("_prompt_line2")
	assert.NoError(t, err)
	assert.Equal(t, "❯ ", line2)
}

func TestShellPromptPresetCommand_Execute_Powerline(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{"style": "powerline"}

	err := cmd.Execute(options, "")
	assert.NoError(t, err)

	// Verify powerline preset was applied
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "3", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "┌─[${@user}@${@hostname:-local}]-[${@time}]", line1)

	line2, err := ctx.GetVariable("_prompt_line2")
	assert.NoError(t, err)
	assert.Equal(t, "├─[${#session_name:-no-session}:${#message_count:-0}]-[${#active_model:-none}]", line2)

	line3, err := ctx.GetVariable("_prompt_line3")
	assert.NoError(t, err)
	assert.Equal(t, "└─➤ ", line3)
}

func TestShellPromptPresetCommand_Execute_InvalidStyle(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{"style": "invalid"}

	err := cmd.Execute(options, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown preset style")
}

func TestShellPromptPresetCommand_Execute_MissingStyle(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry

	cmd := &ShellPromptPresetCommand{}
	options := map[string]string{} // No style option

	err := cmd.Execute(options, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "style option is required")
}

func TestShellPromptPreviewCommand_Name(t *testing.T) {
	cmd := &ShellPromptPreviewCommand{}
	assert.Equal(t, "shell-prompt-preview", cmd.Name())
}

func TestShellPromptPreviewCommand_Execute(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	err := services.GetGlobalRegistry().RegisterService(services.NewShellPromptService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() { services.SetGlobalRegistry(oldServiceRegistry) }()

	// Pre-configure some prompt settings
	_ = ctx.SetVariable("_prompt_lines_count", "2")
	_ = ctx.SetVariable("_prompt_line1", "Line 1: ${@user}")
	_ = ctx.SetVariable("_prompt_line2", "Line 2> ")

	cmd := &ShellPromptPreviewCommand{}
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)
}

func TestShellPromptPreviewCommand_Execute_ServiceUnavailable(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry without shell_prompt service
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	defer func() { services.SetGlobalRegistry(oldServiceRegistry) }()

	cmd := &ShellPromptPreviewCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell prompt service not available")
}
