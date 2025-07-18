package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestWhileCommand_Name(t *testing.T) {
	cmd := &WhileCommand{}
	assert.Equal(t, "while", cmd.Name())
}

func TestWhileCommand_Description(t *testing.T) {
	cmd := &WhileCommand{}
	assert.NotEmpty(t, cmd.Description())
	assert.Contains(t, cmd.Description(), "Repeatedly")
}

func TestWhileCommand_Usage(t *testing.T) {
	cmd := &WhileCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\while")
	assert.Contains(t, usage, "condition")
}

func TestWhileCommand_HelpInfo(t *testing.T) {
	cmd := &WhileCommand{}
	help := cmd.HelpInfo()
	assert.Equal(t, "while", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)
	assert.NotEmpty(t, help.Examples)
	assert.NotEmpty(t, help.Notes)
}

func TestWhileCommand_ParseMode(t *testing.T) {
	cmd := &WhileCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestWhileCommand_EvaluateCondition(t *testing.T) {
	cmd := &WhileCommand{}

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
			result, err := cmd.evaluateCondition(tt.condition)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestWhileCommand_IsTruthy removed - now using shared stringprocessing.IsTruthy function
// Tests are in internal/stringprocessing/processing_test.go

func TestWhileCommand_Execute_MissingCondition(t *testing.T) {
	cmd := &WhileCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "some command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition parameter is required")
}

func TestWhileCommand_Execute_TrueCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if both the while command and the target command were pushed to stack
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 2, stackSize)

	// Check the order: target command should be on top, while command underneath
	commands := concreteCtx.PeekStack()
	assert.Equal(t, "\\set[var=test_value]", commands[0])      // Top of stack (executes first)
	assert.Contains(t, commands[1], "\\while[condition=true]") // While command for next iteration

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#while_result")
	assert.Equal(t, "true", varValue)
}

func TestWhileCommand_Execute_FalseCondition(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "false"}

	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Check if no commands were pushed to stack
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, stackSize)

	// Check if result was stored
	varValue, _ := concreteCtx.GetVariable("#while_result")
	assert.Equal(t, "false", varValue)
}

func TestWhileCommand_Execute_EmptyCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "")
	assert.NoError(t, err)

	// Check if no commands were pushed to stack (empty command)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, stackSize)
}

func TestWhileCommand_Execute_WhitespaceCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "   \t\n  ")
	assert.NoError(t, err)

	// Check if no commands were pushed to stack (whitespace only)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 0, stackSize)
}

func TestWhileCommand_Execute_MissingServices(t *testing.T) {
	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	// This should not panic even if services are missing
	err := cmd.Execute(args, "\\set[var=test]")
	assert.NoError(t, err)
}

func TestWhileCommand_Execute_StackOverflowProtection(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Set a very low stack limit for testing
	_ = concreteCtx.SetSystemVariable("_max_stack_depth", "5")

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	// Fill the stack close to the limit
	for i := 0; i < 3; i++ {
		concreteCtx.PushCommand("dummy command")
	}

	// This should be prevented by stack overflow protection
	err := cmd.Execute(args, "\\set[var=test_value]")
	assert.NoError(t, err)

	// Stack should not have grown beyond the limit
	stackSize := concreteCtx.GetStackSize()
	assert.LessOrEqual(t, stackSize, 5)
}

func TestWhileCommand_Execute_InterpolatedConditions(t *testing.T) {
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

	cmd := &WhileCommand{}

	// Test with truthy condition (state machine would have expanded ${flag1} to "true")
	args := map[string]string{"condition": "true"}
	err := cmd.Execute(args, "\\set[var1=value1]")
	assert.NoError(t, err)
	assert.Equal(t, 2, concreteCtx.GetStackSize()) // while command + target command

	// Clear stack
	concreteCtx.ClearStack()

	// Test with falsy condition (state machine would have expanded ${flag2} to "false")
	args = map[string]string{"condition": "false"}
	err = cmd.Execute(args, "\\set[var2=value2]")
	assert.NoError(t, err)
	assert.Equal(t, 0, concreteCtx.GetStackSize())

	// Clear stack
	concreteCtx.ClearStack()

	// Test with empty condition (state machine would have expanded ${empty_var} to "")
	args = map[string]string{"condition": ""}
	err = cmd.Execute(args, "\\set[var3=value3]")
	assert.NoError(t, err)
	assert.Equal(t, 0, concreteCtx.GetStackSize())
}

func TestWhileCommand_Execute_LoopIteration(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	ctx.SetTestMode(true)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &WhileCommand{}
	args := map[string]string{"condition": "true"}

	err := cmd.Execute(args, "\\echo hello")
	assert.NoError(t, err)

	// Check that the stack contains exactly what we expect
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 2, stackSize)

	commands := concreteCtx.PeekStack()

	// First command to execute should be the target command
	assert.Equal(t, "\\echo hello", commands[0])

	// Second command should be the while command for next iteration
	assert.Contains(t, commands[1], "\\while[condition=true]")
	assert.Contains(t, commands[1], "\\echo hello")
}
