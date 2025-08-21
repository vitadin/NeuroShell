package shell

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptPreviewCommand_Name(t *testing.T) {
	cmd := &PromptPreviewCommand{}
	assert.Equal(t, "shell-prompt-preview", cmd.Name())
}

func TestPromptPreviewCommand_Execute(t *testing.T) {
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

	cmd := &PromptPreviewCommand{}
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)
}

func TestPromptPreviewCommand_Execute_ServiceUnavailable(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry without shell_prompt service
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	defer func() { services.SetGlobalRegistry(oldServiceRegistry) }()

	cmd := &PromptPreviewCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell prompt service not available")
}

func TestPromptPreviewCommand_IsReadOnly(t *testing.T) {
	cmd := &PromptPreviewCommand{}
	assert.True(t, cmd.IsReadOnly())
}
