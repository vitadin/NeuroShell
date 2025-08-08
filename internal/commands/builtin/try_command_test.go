package builtin

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
)

func TestTryCommand_Name(t *testing.T) {
	cmd := &TryCommand{}
	assert.Equal(t, "try", cmd.Name())
}

func TestTryCommand_Description(t *testing.T) {
	cmd := &TryCommand{}
	assert.NotEmpty(t, cmd.Description())
}

func TestTryCommand_Usage(t *testing.T) {
	cmd := &TryCommand{}
	assert.NotEmpty(t, cmd.Usage())
}

func TestTryCommand_ParseMode(t *testing.T) {
	cmd := &TryCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestTryCommand_HelpInfo(t *testing.T) {
	cmd := &TryCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "try", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeRaw, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestTryCommand_Execute_EmptyCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &TryCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "")
	assert.NoError(t, err)

	// Check if success variables were set
	status, _ := concreteCtx.GetVariable("_status")
	assert.Equal(t, "0", status)

	errorVar, _ := concreteCtx.GetVariable("_error")
	assert.Equal(t, "", errorVar)

	output, _ := concreteCtx.GetVariable("_output")
	assert.Equal(t, "", output)

	// Stack should be empty since no command was pushed
	assert.Equal(t, 0, concreteCtx.GetStackSize())
}

func TestTryCommand_Execute_WithTargetCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &TryCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "\\echo Hello World")
	assert.NoError(t, err)

	// Check if error boundary markers and command were pushed to stack
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 3, stackSize) // START marker, command, END marker

	// Check the stack contents (in reverse order due to LIFO)
	stack := concreteCtx.PeekStack()
	assert.Equal(t, 3, len(stack))
	assert.Contains(t, stack[0], "ERROR_BOUNDARY_START:")
	assert.Equal(t, "\\echo Hello World", stack[1])
	assert.Contains(t, stack[2], "ERROR_BOUNDARY_END:")

	// The markers should have the same try ID
	assert.Contains(t, stack[0], "try_id_")
	assert.Contains(t, stack[2], "try_id_")
	// Extract try ID from both markers and verify they match
	startMarker := stack[0]
	endMarker := stack[2]
	startID := startMarker[len("ERROR_BOUNDARY_START:"):]
	endID := endMarker[len("ERROR_BOUNDARY_END:"):]
	assert.Equal(t, startID, endID)
}

func TestTryCommand_Execute_WhitespaceCommand(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &TryCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "   \t\n   ")
	assert.NoError(t, err)

	// Whitespace-only command should be treated as empty
	status, _ := concreteCtx.GetVariable("_status")
	assert.Equal(t, "0", status)

	errorVar, _ := concreteCtx.GetVariable("_error")
	assert.Equal(t, "", errorVar)

	output, _ := concreteCtx.GetVariable("_output")
	assert.Equal(t, "", output)

	// Stack should be empty
	assert.Equal(t, 0, concreteCtx.GetStackSize())
}

func TestTryCommand_Execute_MultipleCommands(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &TryCommand{}
	args := map[string]string{}

	// Execute first try command
	err := cmd.Execute(args, "\\set[var1=value1]")
	assert.NoError(t, err)

	// Execute second try command
	err = cmd.Execute(args, "\\echo Test")
	assert.NoError(t, err)

	// Should have 6 items on stack (3 for each try command)
	stackSize := concreteCtx.GetStackSize()
	assert.Equal(t, 6, stackSize)

	// Check that different try commands have different IDs
	stack := concreteCtx.PeekStack()
	assert.Equal(t, 6, len(stack))

	// First try command markers (most recent, at top of stack)
	firstStartMarker := stack[0]
	firstEndMarker := stack[2]

	// Second try command markers (older, lower in stack)
	secondStartMarker := stack[3]
	secondEndMarker := stack[5]

	// Extract try IDs
	firstStartID := firstStartMarker[len("ERROR_BOUNDARY_START:"):]
	firstEndID := firstEndMarker[len("ERROR_BOUNDARY_END:"):]
	secondStartID := secondStartMarker[len("ERROR_BOUNDARY_START:"):]
	secondEndID := secondEndMarker[len("ERROR_BOUNDARY_END:"):]

	// Same try command should have matching IDs
	assert.Equal(t, firstStartID, firstEndID)
	assert.Equal(t, secondStartID, secondEndID)

	// Different try commands should have different IDs
	assert.NotEqual(t, firstStartID, secondStartID)
}

func TestTryCommand_Execute_MissingServices(t *testing.T) {
	// Setup test context without services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup empty registry (no services)
	services.SetGlobalRegistry(services.NewRegistry())

	cmd := &TryCommand{}
	args := map[string]string{}

	err := cmd.Execute(args, "\\echo Hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stack service not available")
}

func TestTryCommand_Execute_TargetCommandTrimming(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	_ = services.GetGlobalRegistry().InitializeAll()

	cmd := &TryCommand{}
	args := map[string]string{}

	// Test command with leading/trailing whitespace
	err := cmd.Execute(args, "  \t\\echo Hello World\n  ")
	assert.NoError(t, err)

	// Check that the command was trimmed properly
	stack := concreteCtx.PeekStack()
	assert.Equal(t, 3, len(stack))
	assert.Equal(t, "\\echo Hello World", stack[1]) // Command should be trimmed
}
