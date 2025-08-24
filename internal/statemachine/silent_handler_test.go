package statemachine

import (
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSilentHandler(t *testing.T) {
	// Setup global context and services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Create registry and register services
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Register required services
	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()
	assert.NotNil(t, silentHandler)
	assert.NotNil(t, silentHandler.logger)
	assert.NotNil(t, silentHandler.stackService)
}

func TestSilentHandler_GenerateUniqueSilentID(t *testing.T) {
	// Setup global context and services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test ID generation
	id1 := silentHandler.GenerateUniqueSilentID()
	assert.NotEmpty(t, id1)
	assert.Contains(t, id1, "silent_id_")

	// Enter a silent block and generate another ID
	silentHandler.EnterSilentBlock("test_silent")
	id2 := silentHandler.GenerateUniqueSilentID()
	assert.NotEmpty(t, id2)
	assert.Contains(t, id2, "silent_id_")
	assert.NotEqual(t, id1, id2)
}

func TestSilentHandler_GenerateUniqueSilentID_NoStackService(t *testing.T) {
	silentHandler := &SilentHandler{} // No services initialized

	id := silentHandler.GenerateUniqueSilentID()
	assert.Equal(t, "silent_id_0", id)
}

func TestSilentHandler_PushSilentBoundary(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test pushing silent boundary
	silentID := "test_silent_1"
	targetCommand := "\\echo test command"

	silentHandler.PushSilentBoundary(silentID, targetCommand)

	// Check stack contents (LIFO order)
	stack := stackService.PeekStack()
	assert.Len(t, stack, 3)

	// Verify order: SILENT_BOUNDARY_START should be first (top of stack)
	assert.Equal(t, "SILENT_BOUNDARY_START:"+silentID, stack[0])
	assert.Equal(t, targetCommand, stack[1])
	assert.Equal(t, "SILENT_BOUNDARY_END:"+silentID, stack[2])
}

func TestSilentHandler_PushSilentBoundary_NoStackService(_ *testing.T) {
	silentHandler := &SilentHandler{} // No services initialized

	// Should not panic when no stack service
	silentHandler.PushSilentBoundary("test_silent", "echo test")
}

func TestSilentHandler_EnterSilentBlock(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test entering silent block
	silentID := "test_silent_enter"
	silentHandler.EnterSilentBlock(silentID)

	// Verify we're in silent block
	assert.True(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, silentID, silentHandler.GetCurrentSilentID())
}

func TestSilentHandler_EnterSilentBlock_NoStackService(_ *testing.T) {
	silentHandler := NewSilentHandler() // Initialize with logger but no stack service
	silentHandler.stackService = nil    // Explicitly set to nil for testing

	// Should not panic when no stack service
	silentHandler.EnterSilentBlock("test_silent")
}

func TestSilentHandler_ExitSilentBlock(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Enter silent block first
	silentID := "test_silent_exit"
	silentHandler.EnterSilentBlock(silentID)
	assert.True(t, silentHandler.IsInSilentBlock())

	// Exit silent block
	silentHandler.ExitSilentBlock(silentID)

	// Verify we're no longer in silent block
	assert.False(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, "", silentHandler.GetCurrentSilentID())
}

func TestSilentHandler_ExitSilentBlock_NoStackService(_ *testing.T) {
	silentHandler := NewSilentHandler() // Initialize with logger but no stack service
	silentHandler.stackService = nil    // Explicitly set to nil for testing

	// Should not panic when no stack service
	silentHandler.ExitSilentBlock("test_silent")
}

func TestSilentHandler_IsInSilentBlock(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Initially not in silent block
	assert.False(t, silentHandler.IsInSilentBlock())

	// Enter silent block
	silentHandler.EnterSilentBlock("test_silent")
	assert.True(t, silentHandler.IsInSilentBlock())

	// Exit silent block
	silentHandler.ExitSilentBlock("test_silent")
	assert.False(t, silentHandler.IsInSilentBlock())
}

func TestSilentHandler_IsInSilentBlock_NoStackService(t *testing.T) {
	silentHandler := &SilentHandler{} // No services initialized

	assert.False(t, silentHandler.IsInSilentBlock())
}

func TestSilentHandler_GetCurrentSilentID(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Initially no silent ID
	assert.Equal(t, "", silentHandler.GetCurrentSilentID())

	// Enter silent block
	silentID := "test_current_silent"
	silentHandler.EnterSilentBlock(silentID)
	assert.Equal(t, silentID, silentHandler.GetCurrentSilentID())

	// Enter nested silent block
	nestedSilentID := "nested_silent"
	silentHandler.EnterSilentBlock(nestedSilentID)
	assert.Equal(t, nestedSilentID, silentHandler.GetCurrentSilentID())

	// Exit nested silent block
	silentHandler.ExitSilentBlock(nestedSilentID)
	assert.Equal(t, silentID, silentHandler.GetCurrentSilentID())

	// Exit original silent block
	silentHandler.ExitSilentBlock(silentID)
	assert.Equal(t, "", silentHandler.GetCurrentSilentID())
}

func TestSilentHandler_GetCurrentSilentID_NoStackService(t *testing.T) {
	silentHandler := &SilentHandler{} // No services initialized

	assert.Equal(t, "", silentHandler.GetCurrentSilentID())
}

func TestSilentHandler_IsSilentBoundaryMarker(t *testing.T) {
	silentHandler := NewSilentHandler()

	tests := []struct {
		name       string
		command    string
		isBoundary bool
		silentID   string
		isStart    bool
	}{
		{
			name:       "silent boundary start",
			command:    "SILENT_BOUNDARY_START:test_silent_1",
			isBoundary: true,
			silentID:   "test_silent_1",
			isStart:    true,
		},
		{
			name:       "silent boundary end",
			command:    "SILENT_BOUNDARY_END:test_silent_1",
			isBoundary: true,
			silentID:   "test_silent_1",
			isStart:    false,
		},
		{
			name:       "regular command",
			command:    "\\echo hello",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "similar but not boundary",
			command:    "SILENT_BOUNDARY_MIDDLE:test_silent_1",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "empty command",
			command:    "",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "error boundary marker",
			command:    "ERROR_BOUNDARY_START:try_1",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isBoundary, silentID, isStart := silentHandler.IsSilentBoundaryMarker(tt.command)
			assert.Equal(t, tt.isBoundary, isBoundary)
			assert.Equal(t, tt.silentID, silentID)
			assert.Equal(t, tt.isStart, isStart)
		})
	}
}

func TestSilentHandler_NestedSilentBlocks(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test nested silent blocks
	outerSilentID := "outer_silent"
	innerSilentID := "inner_silent"

	// Enter outer silent block
	silentHandler.EnterSilentBlock(outerSilentID)
	assert.True(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, outerSilentID, silentHandler.GetCurrentSilentID())

	// Enter inner silent block
	silentHandler.EnterSilentBlock(innerSilentID)
	assert.True(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, innerSilentID, silentHandler.GetCurrentSilentID())

	// Exit inner silent block
	silentHandler.ExitSilentBlock(innerSilentID)
	assert.True(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, outerSilentID, silentHandler.GetCurrentSilentID())

	// Exit outer silent block
	silentHandler.ExitSilentBlock(outerSilentID)
	assert.False(t, silentHandler.IsInSilentBlock())
	assert.Equal(t, "", silentHandler.GetCurrentSilentID())
}

func TestSilentHandler_MultipleSilentBlocksSequential(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test multiple sequential silent blocks
	silentIDs := []string{"silent_1", "silent_2", "silent_3"}

	for _, silentID := range silentIDs {
		// Enter silent block
		silentHandler.EnterSilentBlock(silentID)
		assert.True(t, silentHandler.IsInSilentBlock())
		assert.Equal(t, silentID, silentHandler.GetCurrentSilentID())

		// Exit silent block
		silentHandler.ExitSilentBlock(silentID)
		assert.False(t, silentHandler.IsInSilentBlock())
		assert.Equal(t, "", silentHandler.GetCurrentSilentID())
	}
}

func TestSilentHandler_BoundaryMarkerParsing_EdgeCases(t *testing.T) {
	silentHandler := NewSilentHandler()

	tests := []struct {
		name       string
		command    string
		isBoundary bool
		silentID   string
		isStart    bool
	}{
		{
			name:       "empty silent ID start",
			command:    "SILENT_BOUNDARY_START:",
			isBoundary: true,
			silentID:   "",
			isStart:    true,
		},
		{
			name:       "empty silent ID end",
			command:    "SILENT_BOUNDARY_END:",
			isBoundary: true,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "silent ID with special characters",
			command:    "SILENT_BOUNDARY_START:silent-id_123.test",
			isBoundary: true,
			silentID:   "silent-id_123.test",
			isStart:    true,
		},
		{
			name:       "silent ID with spaces",
			command:    "SILENT_BOUNDARY_START:silent id with spaces",
			isBoundary: true,
			silentID:   "silent id with spaces",
			isStart:    true,
		},
		{
			name:       "silent ID with colons",
			command:    "SILENT_BOUNDARY_END:silent:id:with:colons",
			isBoundary: true,
			silentID:   "silent:id:with:colons",
			isStart:    false,
		},
		{
			name:       "partial marker start",
			command:    "SILENT_BOUNDARY_START",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "partial marker end",
			command:    "SILENT_BOUNDARY_END",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
		{
			name:       "case sensitive - lowercase",
			command:    "silent_boundary_start:test",
			isBoundary: false,
			silentID:   "",
			isStart:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isBoundary, silentID, isStart := silentHandler.IsSilentBoundaryMarker(tt.command)
			assert.Equal(t, tt.isBoundary, isBoundary)
			assert.Equal(t, tt.silentID, silentID)
			assert.Equal(t, tt.isStart, isStart)
		})
	}
}

func TestSilentHandler_ComplexBoundaryScenario(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test complex scenario with multiple commands and boundaries
	commands := []string{
		"\\echo before silent",
		"\\echo inside first silent",
		"\\echo inside second silent",
		"\\echo after silent",
	}

	// Push a complex silent boundary structure
	silentID1 := "complex_silent_1"
	silentID2 := "complex_silent_2"

	// Push boundaries around commands
	silentHandler.PushSilentBoundary(silentID1, commands[1])
	silentHandler.PushSilentBoundary(silentID2, commands[2])

	// Verify stack structure
	stack := stackService.PeekStack()
	expectedStructure := []string{
		"SILENT_BOUNDARY_START:" + silentID2,
		commands[2],
		"SILENT_BOUNDARY_END:" + silentID2,
		"SILENT_BOUNDARY_START:" + silentID1,
		commands[1],
		"SILENT_BOUNDARY_END:" + silentID1,
	}

	assert.Len(t, stack, len(expectedStructure))
	for i, expected := range expectedStructure {
		assert.Equal(t, expected, stack[i], "Mismatch at index %d", i)
	}

	// Test boundary marker detection for each command in stack
	for _, cmd := range stack {
		isBoundary, id, isStart := silentHandler.IsSilentBoundaryMarker(cmd)
		if isBoundary {
			assert.NotEmpty(t, id)
			assert.True(t, id == silentID1 || id == silentID2)
			// isStart should be true for START markers, false for END markers
			expectedStart := strings.Contains(cmd, "SILENT_BOUNDARY_START:")
			assert.Equal(t, expectedStart, isStart)
		} else {
			assert.Contains(t, commands, cmd)
		}
	}
}

func TestSilentHandler_IDGeneration_WithDepth(t *testing.T) {
	// Setup services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	stackService := services.NewStackService()
	err := stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	silentHandler := NewSilentHandler()

	// Test ID generation at different nesting depths
	baseID := silentHandler.GenerateUniqueSilentID()
	assert.Contains(t, baseID, "silent_id_")

	// Enter first level
	silentHandler.EnterSilentBlock("level_1")
	level1ID := silentHandler.GenerateUniqueSilentID()
	assert.Contains(t, level1ID, "silent_id_")
	assert.NotEqual(t, baseID, level1ID)

	// Enter second level
	silentHandler.EnterSilentBlock("level_2")
	level2ID := silentHandler.GenerateUniqueSilentID()
	assert.Contains(t, level2ID, "silent_id_")
	assert.NotEqual(t, level1ID, level2ID)
	assert.NotEqual(t, baseID, level2ID)

	// Exit levels and verify IDs reflect depth changes
	silentHandler.ExitSilentBlock("level_2")
	backToLevel1ID := silentHandler.GenerateUniqueSilentID()
	assert.Contains(t, backToLevel1ID, "silent_id_")

	silentHandler.ExitSilentBlock("level_1")
	backToBaseID := silentHandler.GenerateUniqueSilentID()
	assert.Contains(t, backToBaseID, "silent_id_")
}
