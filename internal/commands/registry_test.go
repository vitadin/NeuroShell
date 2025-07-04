package commands

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// MockCommand implements Command interface for testing
type MockCommand struct {
	name        string
	parseMode   neurotypes.ParseMode
	description string
	usage       string
	executeFunc func(args map[string]string, input string, ctx neurotypes.Context) error
}

func NewMockCommand(name string) *MockCommand {
	return &MockCommand{
		name:        name,
		parseMode:   neurotypes.ParseModeKeyValue,
		description: fmt.Sprintf("Mock command: %s", name),
		usage:       fmt.Sprintf("Usage: \\%s", name),
		executeFunc: func(_ map[string]string, _ string, _ neurotypes.Context) error {
			return nil
		},
	}
}

func (m *MockCommand) Name() string {
	return m.name
}

func (m *MockCommand) ParseMode() neurotypes.ParseMode {
	return m.parseMode
}

func (m *MockCommand) Description() string {
	return m.description
}

func (m *MockCommand) Usage() string {
	return m.usage
}

func (m *MockCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	if m.executeFunc != nil {
		return m.executeFunc(args, input, ctx)
	}
	return nil
}

func (m *MockCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     m.Name(),
		Description: m.Description(),
		Usage:       m.Usage(),
		ParseMode:   m.ParseMode(),
		Examples:    []neurotypes.HelpExample{},
	}
}

func (m *MockCommand) SetParseMode(mode neurotypes.ParseMode) {
	m.parseMode = mode
}

func (m *MockCommand) SetExecuteFunc(fn func(args map[string]string, input string, ctx neurotypes.Context) error) {
	m.executeFunc = fn
}

func TestRegistry_NewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.commands)
	assert.Equal(t, 0, len(registry.commands))
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		command neurotypes.Command
		wantErr bool
		errMsg  string
	}{
		{
			name:    "register valid command",
			command: NewMockCommand("test"),
			wantErr: false,
		},
		{
			name:    "register another command",
			command: NewMockCommand("another"),
			wantErr: false,
		},
		{
			name:    "register command with empty name",
			command: NewMockCommand(""),
			wantErr: true,
			errMsg:  "command name cannot be empty",
		},
	}

	registry := NewRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.command)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify command was registered
				cmd, exists := registry.Get(tt.command.Name())
				assert.True(t, exists)
				assert.Equal(t, tt.command, cmd)
			}
		})
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()
	cmd1 := NewMockCommand("duplicate")
	cmd2 := NewMockCommand("duplicate")

	// Register first command
	err := registry.Register(cmd1)
	assert.NoError(t, err)

	// Try to register command with same name
	err = registry.Register(cmd2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command duplicate already registered")

	// Verify original command is still registered
	cmd, exists := registry.Get("duplicate")
	assert.True(t, exists)
	assert.Equal(t, cmd1, cmd)
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	cmd := NewMockCommand("test")

	// Register command
	err := registry.Register(cmd)
	require.NoError(t, err)

	// Verify it exists
	_, exists := registry.Get("test")
	assert.True(t, exists)

	// Unregister it
	registry.Unregister("test")

	// Verify it's gone
	_, exists = registry.Get("test")
	assert.False(t, exists)

	// Unregistering non-existent command should not panic
	registry.Unregister("nonexistent")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	cmd := NewMockCommand("test")

	// Register command
	err := registry.Register(cmd)
	require.NoError(t, err)

	tests := []struct {
		name        string
		commandName string
		wantExists  bool
		wantCommand neurotypes.Command
	}{
		{
			name:        "get existing command",
			commandName: "test",
			wantExists:  true,
			wantCommand: cmd,
		},
		{
			name:        "get non-existing command",
			commandName: "nonexistent",
			wantExists:  false,
			wantCommand: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, exists := registry.Get(tt.commandName)

			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				assert.Equal(t, tt.wantCommand, gotCmd)
			} else {
				assert.Nil(t, gotCmd)
			}
		})
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	commands := registry.GetAll()
	assert.Equal(t, 0, len(commands))

	// Register some commands
	testCommands := []neurotypes.Command{
		NewMockCommand("cmd1"),
		NewMockCommand("cmd2"),
		NewMockCommand("cmd3"),
	}

	for _, cmd := range testCommands {
		err := registry.Register(cmd)
		require.NoError(t, err)
	}

	// Get all commands
	allCommands := registry.GetAll()
	assert.Equal(t, len(testCommands), len(allCommands))

	// Verify all commands are present (order may vary)
	commandMap := make(map[string]neurotypes.Command)
	for _, cmd := range allCommands {
		commandMap[cmd.Name()] = cmd
	}

	for _, expectedCmd := range testCommands {
		actualCmd, exists := commandMap[expectedCmd.Name()]
		assert.True(t, exists)
		assert.Equal(t, expectedCmd, actualCmd)
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	// Create a command with custom execute function
	executed := false
	var capturedArgs map[string]string
	var capturedInput string
	var capturedContext neurotypes.Context

	cmd := NewMockCommand("test")
	cmd.SetExecuteFunc(func(args map[string]string, input string, ctx neurotypes.Context) error {
		executed = true
		capturedArgs = args
		capturedInput = input
		capturedContext = ctx
		return nil
	})

	err := registry.Register(cmd)
	require.NoError(t, err)

	// Test successful execution
	args := map[string]string{"key": "value"}
	input := "test input"

	err = registry.Execute("test", args, input, ctx)
	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Equal(t, args, capturedArgs)
	assert.Equal(t, input, capturedInput)
	assert.Equal(t, ctx, capturedContext)

	// Test execution of non-existent command
	err = registry.Execute("nonexistent", args, input, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command: nonexistent")
}

func TestRegistry_Execute_CommandError(t *testing.T) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	// Create a command that returns an error
	expectedError := fmt.Errorf("command execution failed")
	cmd := NewMockCommand("failing")
	cmd.SetExecuteFunc(func(_ map[string]string, _ string, _ neurotypes.Context) error {
		return expectedError
	})

	err := registry.Register(cmd)
	require.NoError(t, err)

	// Test that command error is propagated
	err = registry.Execute("failing", nil, "", ctx)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestRegistry_GetParseMode(t *testing.T) {
	registry := NewRegistry()

	// Create commands with different parse modes
	keyValueCmd := NewMockCommand("keyvalue")
	keyValueCmd.SetParseMode(neurotypes.ParseModeKeyValue)

	rawCmd := NewMockCommand("raw")
	rawCmd.SetParseMode(neurotypes.ParseModeRaw)

	err := registry.Register(keyValueCmd)
	require.NoError(t, err)
	err = registry.Register(rawCmd)
	require.NoError(t, err)

	tests := []struct {
		name         string
		commandName  string
		expectedMode neurotypes.ParseMode
	}{
		{
			name:         "key-value parse mode",
			commandName:  "keyvalue",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "raw parse mode",
			commandName:  "raw",
			expectedMode: neurotypes.ParseModeRaw,
		},
		{
			name:         "non-existent command defaults to key-value",
			commandName:  "nonexistent",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := registry.GetParseMode(tt.commandName)
			assert.Equal(t, tt.expectedMode, mode)
		})
	}
}

func TestRegistry_IsValidCommand(t *testing.T) {
	registry := NewRegistry()
	cmd := NewMockCommand("valid")

	err := registry.Register(cmd)
	require.NoError(t, err)

	tests := []struct {
		name        string
		commandName string
		expected    bool
	}{
		{
			name:        "valid command",
			commandName: "valid",
			expected:    true,
		},
		{
			name:        "invalid command",
			commandName: "invalid",
			expected:    false,
		},
		{
			name:        "empty command name",
			commandName: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.IsValidCommand(tt.commandName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test concurrent access
func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Number of goroutines
	numGoroutines := 10
	commandsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < commandsPerGoroutine; j++ {
				cmdName := fmt.Sprintf("cmd_%d_%d", id, j)
				cmd := NewMockCommand(cmdName)

				err := registry.Register(cmd)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all commands were registered
	allCommands := registry.GetAll()
	expectedCount := numGoroutines * commandsPerGoroutine
	assert.Equal(t, expectedCount, len(allCommands))

	// Test concurrent retrieval
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < commandsPerGoroutine; j++ {
				cmdName := fmt.Sprintf("cmd_%d_%d", id, j)

				cmd, exists := registry.Get(cmdName)
				assert.True(t, exists)
				assert.Equal(t, cmdName, cmd.Name())

				// Test IsValidCommand
				valid := registry.IsValidCommand(cmdName)
				assert.True(t, valid)

				// Test GetParseMode
				mode := registry.GetParseMode(cmdName)
				assert.Equal(t, neurotypes.ParseModeKeyValue, mode)
			}
		}(i)
	}

	wg.Wait()
}

// Test registry state consistency
func TestRegistry_StateConsistency(t *testing.T) {
	registry := NewRegistry()

	// Test empty state
	assert.Equal(t, 0, len(registry.GetAll()))
	assert.False(t, registry.IsValidCommand("nonexistent"))

	// Register command
	cmd := NewMockCommand("test")
	err := registry.Register(cmd)
	assert.NoError(t, err)

	// Test state after registration
	assert.Equal(t, 1, len(registry.GetAll()))
	assert.True(t, registry.IsValidCommand("test"))

	retrievedCmd, exists := registry.Get("test")
	assert.True(t, exists)
	assert.Equal(t, cmd, retrievedCmd)

	// Test duplicate registration fails
	err = registry.Register(NewMockCommand("test"))
	assert.Error(t, err)

	// State should remain unchanged
	assert.Equal(t, 1, len(registry.GetAll()))
	retrievedCmd, exists = registry.Get("test")
	assert.True(t, exists)
	assert.Equal(t, cmd, retrievedCmd)

	// Test unregistration
	registry.Unregister("test")
	assert.Equal(t, 0, len(registry.GetAll()))
	assert.False(t, registry.IsValidCommand("test"))
}

// Test GlobalRegistry
func TestGlobalRegistry(t *testing.T) {
	assert.NotNil(t, GlobalRegistry)

	// Test basic functionality (note: this modifies global state)
	originalCount := len(GlobalRegistry.GetAll())

	cmd := NewMockCommand("global_test")
	err := GlobalRegistry.Register(cmd)
	assert.NoError(t, err)

	assert.True(t, GlobalRegistry.IsValidCommand("global_test"))

	retrievedCmd, exists := GlobalRegistry.Get("global_test")
	assert.True(t, exists)
	assert.Equal(t, cmd, retrievedCmd)

	// Clean up
	GlobalRegistry.Unregister("global_test")
	assert.Equal(t, originalCount, len(GlobalRegistry.GetAll()))
}

// Benchmark tests
func BenchmarkRegistry_Register(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewMockCommand(fmt.Sprintf("cmd_%d", i))
		_ = registry.Register(cmd)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	registry := NewRegistry()

	// Pre-register commands
	for i := 0; i < 1000; i++ {
		cmd := NewMockCommand(fmt.Sprintf("cmd_%d", i))
		_ = registry.Register(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmdName := fmt.Sprintf("cmd_%d", i%1000)
		_, _ = registry.Get(cmdName)
	}
}

func BenchmarkRegistry_Execute(b *testing.B) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	cmd := NewMockCommand("bench")
	_ = registry.Register(cmd)

	args := map[string]string{"key": "value"}
	input := "test input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Execute("bench", args, input, ctx)
	}
}

func BenchmarkRegistry_GetAll(b *testing.B) {
	registry := NewRegistry()

	// Pre-register commands
	for i := 0; i < 100; i++ {
		cmd := NewMockCommand(fmt.Sprintf("cmd_%d", i))
		_ = registry.Register(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetAll()
	}
}
