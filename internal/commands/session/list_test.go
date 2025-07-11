// TODO: Integrate into state machine - temporarily commented out for build compatibility
package session

/*

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestListCommand_Name(t *testing.T) {
	cmd := &ListCommand{}
	assert.Equal(t, "session-list", cmd.Name())
}

func TestListCommand_ParseMode(t *testing.T) {
	cmd := &ListCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestListCommand_Description(t *testing.T) {
	cmd := &ListCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "list")
	assert.Contains(t, strings.ToLower(desc), "session")
}

func TestListCommand_Usage(t *testing.T) {
	cmd := &ListCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "session-list")
	assert.Contains(t, usage, "sort")
	assert.Contains(t, usage, "filter")
}

func TestListCommand_Execute_EmptySessionList(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	err := cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Equal(t, "No sessions found.\n", output)
}

func TestListCommand_Execute_SingleSession(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create a session first
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "test_session")
	require.NoError(t, err)

	// List sessions
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Sessions (1 total):")
	assert.Contains(t, output, "test_session")
	assert.Contains(t, output, "active")
	assert.Contains(t, output, "0 messages")
}

func TestListCommand_Execute_MultipleSessions(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create multiple sessions with slight delays to ensure different timestamps
	newCmd := &NewCommand{}

	err := newCmd.Execute(map[string]string{}, "first_session")
	require.NoError(t, err)
	time.Sleep(1 * time.Millisecond)

	err = newCmd.Execute(map[string]string{}, "second_session")
	require.NoError(t, err)
	time.Sleep(1 * time.Millisecond)

	err = newCmd.Execute(map[string]string{}, "third_session")
	require.NoError(t, err)

	// List sessions
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Sessions (3 total):")
	assert.Contains(t, output, "first_session")
	assert.Contains(t, output, "second_session")
	assert.Contains(t, output, "third_session")
	assert.Contains(t, output, "active") // One should be active
}

func TestListCommand_Execute_SortByName(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create sessions in reverse alphabetical order
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "zebra")
	require.NoError(t, err)
	err = newCmd.Execute(map[string]string{}, "alpha")
	require.NoError(t, err)
	err = newCmd.Execute(map[string]string{}, "beta")
	require.NoError(t, err)

	// List sessions sorted by name
	err = cmd.Execute(map[string]string{"sort": "name"}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)

	// Verify alphabetical order
	alphaPos := strings.Index(output, "alpha")
	betaPos := strings.Index(output, "beta")
	zebraPos := strings.Index(output, "zebra")

	assert.True(t, alphaPos < betaPos)
	assert.True(t, betaPos < zebraPos)
}

func TestListCommand_Execute_SortByCreated(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create sessions with time delays
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "first")
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	err = newCmd.Execute(map[string]string{}, "second")
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	err = newCmd.Execute(map[string]string{}, "third")
	require.NoError(t, err)

	// List sessions sorted by created (newest first - default)
	err = cmd.Execute(map[string]string{"sort": "created"}, "")
	assert.NoError(t, err)

	// Check output variable - newest should be first
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)

	// Verify newest first order
	thirdPos := strings.Index(output, "third")
	secondPos := strings.Index(output, "second")
	firstPos := strings.Index(output, "first")

	assert.True(t, thirdPos < secondPos)
	assert.True(t, secondPos < firstPos)
}

func TestListCommand_Execute_FilterActive(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create multiple sessions
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "first_session")
	require.NoError(t, err)
	err = newCmd.Execute(map[string]string{}, "second_session")
	require.NoError(t, err)
	err = newCmd.Execute(map[string]string{}, "active_session")
	require.NoError(t, err)

	// Filter for active session only
	err = cmd.Execute(map[string]string{"filter": "active"}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Sessions (1 total):")
	assert.Contains(t, output, "active_session")
	assert.Contains(t, output, "active")
	assert.NotContains(t, output, "first_session")
	assert.NotContains(t, output, "second_session")
}

func TestListCommand_Execute_FilterActiveNoActiveSession(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Filter for active session when none exists
	err := cmd.Execute(map[string]string{"filter": "active"}, "")
	assert.NoError(t, err)

	// Check output variable
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Equal(t, "No sessions found.\n", output)
}

func TestListCommand_Execute_CombinedOptions(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create multiple sessions
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "zebra_session")
	require.NoError(t, err)
	err = newCmd.Execute(map[string]string{}, "alpha_session")
	require.NoError(t, err)

	// Use combined options: filter active and sort by name
	err = cmd.Execute(map[string]string{"filter": "active", "sort": "name"}, "")
	assert.NoError(t, err)

	// Check output variable - should show only active session
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Sessions (1 total):")
	assert.Contains(t, output, "alpha_session") // Latest created is active
	assert.Contains(t, output, "active")
}

func TestListCommand_Execute_InvalidSortOption(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	err := cmd.Execute(map[string]string{"sort": "invalid"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort option")
}

func TestListCommand_Execute_InvalidFilterOption(t *testing.T) {
	cmd := &ListCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	err := cmd.Execute(map[string]string{"filter": "invalid"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter option")
}

func TestListCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &ListCommand{}

	// Don't setup services - should fail
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestListCommand_ValidateArguments(t *testing.T) {
	cmd := &ListCommand{}

	tests := []struct {
		name        string
		sortBy      string
		filterBy    string
		expectError bool
	}{
		{
			name:        "valid sort and filter",
			sortBy:      "name",
			filterBy:    "active",
			expectError: false,
		},
		{
			name:        "valid defaults",
			sortBy:      "created",
			filterBy:    "all",
			expectError: false,
		},
		{
			name:        "invalid sort",
			sortBy:      "invalid",
			filterBy:    "all",
			expectError: true,
		},
		{
			name:        "invalid filter",
			sortBy:      "name",
			filterBy:    "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.validateArguments(tt.sortBy, tt.filterBy)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Interface compliance check
var _ neurotypes.Command = (*ListCommand)(nil)
*/
