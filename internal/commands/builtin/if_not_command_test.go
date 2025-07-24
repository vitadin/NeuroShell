package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
	"neuroshell/pkg/stringprocessing"
)

func TestIfNotCommand_Name(t *testing.T) {
	cmd := &IfNotCommand{}
	assert.Equal(t, "if-not", cmd.Name())
}

func TestIfNotCommand_Description(t *testing.T) {
	cmd := &IfNotCommand{}
	assert.NotEmpty(t, cmd.Description())
	assert.Contains(t, cmd.Description(), "Conditionally")
}

func TestIfNotCommand_Usage(t *testing.T) {
	cmd := &IfNotCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\if-not")
	assert.Contains(t, usage, "condition")
}

func TestIfNotCommand_HelpInfo(t *testing.T) {
	cmd := &IfNotCommand{}
	help := cmd.HelpInfo()
	assert.Equal(t, "if-not", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)
	assert.NotEmpty(t, help.Examples)
	assert.NotEmpty(t, help.Notes)
}

func TestIfNotCommand_ParseMode(t *testing.T) {
	cmd := &IfNotCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestIfNotCommand_EvaluateCondition(t *testing.T) {
	cmd := &IfNotCommand{}

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

func TestIfNotCommand_IsTruthy(t *testing.T) {
	// Test that if-not uses the same IsTruthy logic as stringprocessing
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
			result := stringprocessing.IsTruthy(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIfNotCommand_Execute_MissingCondition(t *testing.T) {
	cmd := &IfNotCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "some command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition parameter is required")
}

func TestIfNotCommand_Execute_FalseCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was pushed to stack (condition is false, so if-not executes)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 1, stackSize)

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#if_not_result")
	assert.Equal(t, "false", varValue)
}

func TestIfNotCommand_Execute_TrueCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was NOT queued (condition is true, so if-not does not execute)
	queueSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, queueSize)

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#if_not_result")
	assert.Equal(t, "true", varValue)
}

func TestIfNotCommand_Execute_VariableCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	stackService := services.NewStackService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(stackService)
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	// Note: In real execution, the state machine would have already expanded ${debug_mode} to "false"
	// This test simulates that the \if-not command receives the pre-expanded value
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was pushed to stack (condition is false, so if-not executes)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 1, stackSize)
}

func TestIfNotCommand_Execute_InterpolatedSystemVariable(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	stackService := services.NewStackService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(stackService)
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	// Note: In real execution, the state machine would have already expanded ${#test_mode} to ""
	// This test simulates that the \if-not command receives the pre-expanded value
	args := map[string]string{"condition": ""}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if command was pushed to stack (empty condition is falsy, so if-not executes)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 1, stackSize)
}

func TestIfNotCommand_Execute_EmptyCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "")
	assert.NoError(t, err)

	// Check if command was NOT queued (empty command)
	queueSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, queueSize)
}

func TestIfNotCommand_Execute_WhitespaceCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "   \t\n  ")
	assert.NoError(t, err)

	// Check if command was NOT queued (whitespace only)
	queueSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, queueSize)
}

func TestIfNotCommand_Execute_MissingServices(t *testing.T) {
	cmd := &IfNotCommand{}
	args := map[string]string{"condition": "false"}

	// This should not panic even if services are missing
	err := cmd.Execute(args, "\\set[var=test]")
	assert.NoError(t, err)
}

func TestIfNotCommand_Execute_InterpolatedConditions(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	varService := services.NewVariableService()
	stackService := services.NewStackService()
	_ = services.GetGlobalRegistry().RegisterService(varService)
	_ = services.GetGlobalRegistry().RegisterService(stackService)
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &IfNotCommand{}

	// Test with falsy condition (state machine would have expanded ${flag1} to "false")
	args := map[string]string{"condition": "false"}
	err := cmd.Execute(args, "\\set[var1=value1]")
	assert.NoError(t, err)
	assert.Equal(t, 1, concreteCtx.GetStackSize()) // if-not executes on false

	// Clear stack
	concreteCtx.ClearStack()

	// Test with truthy condition (state machine would have expanded ${flag2} to "true")
	args = map[string]string{"condition": "true"}
	err = cmd.Execute(args, "\\set[var2=value2]")
	assert.NoError(t, err)
	assert.Equal(t, 0, concreteCtx.GetStackSize()) // if-not does not execute on true

	// Clear stack
	concreteCtx.ClearStack()

	// Test with empty condition (state machine would have expanded ${empty_var} to "")
	args = map[string]string{"condition": ""}
	err = cmd.Execute(args, "\\set[var3=value3]")
	assert.NoError(t, err)
	assert.Equal(t, 1, concreteCtx.GetStackSize()) // if-not executes on empty (falsy)
}
