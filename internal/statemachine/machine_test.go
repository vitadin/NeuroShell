package statemachine

import (
	"testing"

	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/builtin"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// setupTestEnvironment sets up test context, command registry and services
func setupTestEnvironment(t *testing.T) func() {
	// Create a fresh test context
	testCtx := context.New()
	testCtx.SetTestMode(true)

	// Set as global context
	context.SetGlobalContext(testCtx)

	// Clear and reinitialize registries using thread-safe functions
	services.SetGlobalRegistry(services.NewRegistry())
	commands.SetGlobalRegistry(commands.NewRegistry())

	// Register builtin commands manually since we cleared the registry
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SetCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.GetCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.HelpCommand{}))
	require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.ExitCommand{}))
	// Send commands commented out during state machine transition
	// require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendCommand{}))
	// require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendSyncCommand{}))
	// require.NoError(t, commands.GetGlobalRegistry().Register(&builtin.SendStreamCommand{}))

	// Initialize services
	if err := services.GetGlobalRegistry().RegisterService(services.NewVariableService()); err != nil {
		t.Fatalf("Failed to register variable service: %v", err)
	}
	if err := services.GetGlobalRegistry().RegisterService(services.NewStackService()); err != nil {
		t.Fatalf("Failed to register stack service: %v", err)
	}
	if err := services.GetGlobalRegistry().RegisterService(services.NewExecutorService()); err != nil {
		t.Fatalf("Failed to register executor service: %v", err)
	}
	if err := services.GetGlobalRegistry().RegisterService(services.NewMockLLMService()); err != nil {
		t.Fatalf("Failed to register LLM service: %v", err)
	}
	if err := services.GetGlobalRegistry().InitializeAll(); err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	return func() {
		// Reset global context
		context.ResetGlobalContext()
	}
}

// TestStateMachine_NewStateMachine tests state machine creation.
func TestStateMachine_NewStateMachine(t *testing.T) {
	ctx := context.New()
	config := neurotypes.DefaultStateMachineConfig()

	sm := NewStateMachine(ctx, config)

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	if sm.context != ctx {
		t.Error("Expected context to be set correctly")
	}

	if sm.GetConfig().RecursionLimit != config.RecursionLimit {
		t.Error("Expected config to be set correctly")
	}
}

// TestStateMachine_NewStateMachineWithDefaults tests creation with defaults.
func TestStateMachine_NewStateMachineWithDefaults(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	// Test default configuration is applied
	defaultConfig := neurotypes.DefaultStateMachineConfig()
	if sm.GetConfig().RecursionLimit != defaultConfig.RecursionLimit {
		t.Error("Expected default recursion limit to be applied")
	}
}

// TestStateMachine_Execute tests basic command execution.
func TestStateMachine_Execute(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.GetGlobalContext().(*context.NeuroContext)
	sm := NewStateMachineWithDefaults(ctx)

	// Test simple command execution
	err := sm.Execute("\\set[test=value]")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the variable was set
	value, err := ctx.GetVariable("test")
	if err != nil {
		t.Errorf("Expected variable to be set, got error: %v", err)
	}
	if value != "value" {
		t.Errorf("Expected variable value to be 'value', got '%s'", value)
	}

	// Test empty input should return an error
	err = sm.Execute("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}
