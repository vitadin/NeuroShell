package shell

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptCommand_Name(t *testing.T) {
	cmd := &PromptCommand{}
	assert.Equal(t, "shell-prompt", cmd.Name())
}

func TestPromptCommand_Execute_SetLines(t *testing.T) {
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

	cmd := &PromptCommand{}
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

func TestPromptCommand_Execute_SetLineTemplates(t *testing.T) {
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

	cmd := &PromptCommand{}
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

func TestPromptCommand_Execute_InvalidLines(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry
	cmd := &PromptCommand{}

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

func TestPromptCommand_IsReadOnly(t *testing.T) {
	cmd := &PromptCommand{}
	assert.False(t, cmd.IsReadOnly())
}
