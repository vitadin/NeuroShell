package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestIfCommand_Name(t *testing.T) {
	cmd := &IfCommand{}
	assert.Equal(t, "if", cmd.Name())
}

func TestIfCommand_Description(t *testing.T) {
	cmd := &IfCommand{}
	assert.NotEmpty(t, cmd.Description())
	assert.Contains(t, cmd.Description(), "Conditionally")
}

func TestIfCommand_Usage(t *testing.T) {
	cmd := &IfCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\if")
	assert.Contains(t, usage, "condition")
}

func TestIfCommand_HelpInfo(t *testing.T) {
	cmd := &IfCommand{}
	help := cmd.HelpInfo()
	assert.Equal(t, "if", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)
	assert.NotEmpty(t, help.Examples)
	assert.NotEmpty(t, help.Notes)
}

func TestIfCommand_ParseMode(t *testing.T) {
	cmd := &IfCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestIfCommand_EvaluateCondition(t *testing.T) {
	cmd := &IfCommand{}

	tests := []struct {
		name      string
		condition string
		expected  bool
	}{
		{"true string", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed case", "True", true},
		{"1 string", "1", true},
		{"yes string", "yes", true},
		{"YES uppercase", "YES", true},
		{"on string", "on", true},
		{"enabled string", "enabled", true},

		{"false string", "false", false},
		{"FALSE uppercase", "FALSE", false},
		{"0 string", "0", false},
		{"no string", "no", false},
		{"NO uppercase", "NO", false},
		{"off string", "off", false},
		{"disabled string", "disabled", false},

		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"random string", "hello", true},
		{"random uppercase", "HELLO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.evaluateCondition(tt.condition)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfCommand_IsTruthy(t *testing.T) {
	cmd := &IfCommand{}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty", "", false},
		{"whitespace", "   ", false},
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"on", "on", true},
		{"enabled", "enabled", true},
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"0", "0", false},
		{"no", "no", false},
		{"NO", "NO", false},
		{"off", "off", false},
		{"disabled", "disabled", false},
		{"random", "hello", true},
		{"random uppercase", "HELLO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.isTruthy(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfCommand_Execute_MissingCondition(t *testing.T) {
	cmd := &IfCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "some command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition parameter is required")
}

func TestIfCommand_Execute_TrueCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewQueueService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was queued
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 1, queueSize)

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#if_result")
	assert.Equal(t, "true", varValue)
}

func TestIfCommand_Execute_FalseCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewQueueService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfCommand{}
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was NOT queued
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 0, queueSize)

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#if_result")
	assert.Equal(t, "false", varValue)
}

func TestIfCommand_Execute_VariableCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	queueService := services.NewQueueService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(queueService)
	_ = services.GetGlobalRegistry().InitializeAll()

	// Set a variable
	_ = varService.Set("debug_mode", "true")

	cmd := &IfCommand{}
	args := map[string]string{"condition": "${debug_mode}"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was queued
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 1, queueSize)
}

func TestIfCommand_Execute_SystemVariableCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	queueService := services.NewQueueService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(queueService)
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfCommand{}
	args := map[string]string{"condition": "#test_mode"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was queued (test_mode should be true in test context)
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 1, queueSize)
}

func TestIfCommand_Execute_EmptyCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewQueueService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "")
	assert.NoError(t, err)

	// Check if command was NOT queued (empty command)
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 0, queueSize)
}

func TestIfCommand_Execute_WhitespaceCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewQueueService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "   \t\n  ")
	assert.NoError(t, err)

	// Check if command was NOT queued (whitespace only)
	queueSize := concreteCtx.GetQueueSize()
	assert.Equal(t, 0, queueSize)
}

func TestIfCommand_Execute_MissingServices(t *testing.T) {
	cmd := &IfCommand{}
	args := map[string]string{"condition": "true"}

	// This should not panic even if services are missing
	err := cmd.Execute(args, "\\set[var=test]")
	assert.NoError(t, err)
}

func TestIfCommand_Execute_ComplexCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	queueService := services.NewQueueService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(queueService)
	_ = services.GetGlobalRegistry().InitializeAll()

	// Set variables
	_ = varService.Set("flag1", "true")
	_ = varService.Set("flag2", "false")
	_ = varService.Set("empty_var", "")

	cmd := &IfCommand{}

	// Test with true variable
	args := map[string]string{"condition": "${flag1}"}
	err := cmd.Execute(args, "\\set[var1=value1]")
	assert.NoError(t, err)
	assert.Equal(t, 1, concreteCtx.GetQueueSize())

	// Clear queue
	concreteCtx.ClearQueue()

	// Test with false variable
	args = map[string]string{"condition": "${flag2}"}
	err = cmd.Execute(args, "\\set[var2=value2]")
	assert.NoError(t, err)
	assert.Equal(t, 0, concreteCtx.GetQueueSize())

	// Clear queue
	concreteCtx.ClearQueue()

	// Test with empty variable
	args = map[string]string{"condition": "${empty_var}"}
	err = cmd.Execute(args, "\\set[var3=value3]")
	assert.NoError(t, err)
	assert.Equal(t, 0, concreteCtx.GetQueueSize())
}
