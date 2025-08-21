package shell

import (
	"testing"

	"neuroshell/internal/context"

	"github.com/stretchr/testify/assert"
)

func TestPromptPresetCommand_Name(t *testing.T) {
	cmd := &PromptPresetCommand{}
	assert.Equal(t, "shell-prompt-preset", cmd.Name())
}

func TestPromptPresetCommand_Execute_Minimal(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
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

func TestPromptPresetCommand_Execute_Default(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
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

func TestPromptPresetCommand_Execute_Developer(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
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

func TestPromptPresetCommand_Execute_Powerline(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
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

func TestPromptPresetCommand_Execute_InvalidStyle(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
	options := map[string]string{"style": "invalid"}

	err := cmd.Execute(options, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown preset style")
}

func TestPromptPresetCommand_Execute_MissingStyle(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	cmd := &PromptPresetCommand{}
	options := map[string]string{} // No style option

	err := cmd.Execute(options, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "style option is required")
}

func TestPromptPresetCommand_IsReadOnly(t *testing.T) {
	cmd := &PromptPresetCommand{}
	assert.False(t, cmd.IsReadOnly())
}
