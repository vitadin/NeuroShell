package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
)

func TestChatSessionService_GenerateDefaultSessionName(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and initialize service
	service := NewChatSessionService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test first auto-generated name
	name1 := service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 1", name1)

	// Create a session with that name to make it unavailable
	_, err = service.CreateSession(name1, "test prompt", "")
	require.NoError(t, err)

	// Test second auto-generated name should be incremented
	name2 := service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 2", name2)

	// Create that session too
	_, err = service.CreateSession(name2, "test prompt", "")
	require.NoError(t, err)

	// Test third auto-generated name
	name3 := service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 3", name3)
}

func TestChatSessionService_GenerateDefaultSessionName_FallbackPatterns(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and initialize service
	service := NewChatSessionService()
	err := service.Initialize()
	require.NoError(t, err)

	// Fill up Session 1-5 to test fallback to Chat pattern
	for i := 1; i <= 5; i++ {
		sessionName := "Session " + string(rune('0'+i))
		_, err = service.CreateSession(sessionName, "test prompt", "")
		require.NoError(t, err)
	}

	// Should fall back to Chat pattern
	name := service.GenerateDefaultSessionName()

	// Should still find an available Session number or move to Chat pattern
	// Since we only created Session 1-5, Session 6 should be available
	assert.Equal(t, "Session 6", name)
}

func TestChatSessionService_GenerateDefaultSessionName_UninitializedService(t *testing.T) {
	// Test uninitialized service behavior
	service := NewChatSessionService()
	// Don't call Initialize()

	name := service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 1", name)
}

func TestChatSessionService_GenerateDefaultSessionName_AllSessionPatternsTaken(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and initialize service
	service := NewChatSessionService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create sessions to test pattern progression
	// Take Session 1 and Session 2
	_, err = service.CreateSession("Session 1", "test", "")
	require.NoError(t, err)
	_, err = service.CreateSession("Session 2", "test", "")
	require.NoError(t, err)

	// Next should be Session 3
	name := service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 3", name)

	// Take Chat 1 to test Chat fallback later
	_, err = service.CreateSession("Chat 1", "test", "")
	require.NoError(t, err)

	// Should still get Session 3 since we skipped it
	name = service.GenerateDefaultSessionName()
	assert.Equal(t, "Session 3", name)
}
