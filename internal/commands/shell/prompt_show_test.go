package shell

import (
	"testing"

	"neuroshell/internal/context"

	"github.com/stretchr/testify/assert"
)

func TestPromptShowCommand_Name(t *testing.T) {
	cmd := &PromptShowCommand{}
	assert.Equal(t, "shell-prompt-show", cmd.Name())
}

func TestPromptShowCommand_Execute(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Pre-configure some prompt settings
	_ = ctx.SetVariable("_prompt_lines_count", "2")
	_ = ctx.SetVariable("_prompt_line1", "test line 1")
	_ = ctx.SetVariable("_prompt_line2", "test line 2")

	cmd := &PromptShowCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)
}

func TestPromptShowCommand_IsReadOnly(t *testing.T) {
	cmd := &PromptShowCommand{}
	assert.True(t, cmd.IsReadOnly())
}
