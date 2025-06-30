// Package context provides session state management and variable interpolation for NeuroShell.
// It maintains variables, message history, execution queues, and system metadata across command executions.
package context

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"neuroshell/pkg/types"
)

// NeuroContext implements the types.Context interface providing session state management.
// It maintains variables, message history, execution queues, and metadata for NeuroShell sessions.
type NeuroContext struct {
	variables      map[string]string
	history        []types.Message
	sessionID      string
	executionQueue []string
	scriptMetadata map[string]interface{}
	testMode       bool
}

// New creates a new NeuroContext with initialized maps and a unique session ID.
func New() *NeuroContext {
	return &NeuroContext{
		variables:      make(map[string]string),
		history:        make([]types.Message, 0),
		sessionID:      fmt.Sprintf("session_%d", time.Now().Unix()),
		executionQueue: make([]string, 0),
		scriptMetadata: make(map[string]interface{}),
		testMode:       false,
	}
}

// GetVariable retrieves a variable value by name, supporting both user and system variables.
func (ctx *NeuroContext) GetVariable(name string) (string, error) {
	// Handle special variables
	if value, ok := ctx.getSystemVariable(name); ok {
		return value, nil
	}

	// Handle user variables
	if value, ok := ctx.variables[name]; ok {
		return value, nil
	}

	return "", fmt.Errorf("variable %s not found", name)
}

// SetVariable sets a user variable, preventing modification of system variables.
func (ctx *NeuroContext) SetVariable(name string, value string) error {
	// Don't allow setting system variables
	if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_") {
		return fmt.Errorf("cannot set system variable: %s", name)
	}

	ctx.variables[name] = value
	return nil
}

// GetMessageHistory returns the last n messages from the conversation history.
func (ctx *NeuroContext) GetMessageHistory(n int) []types.Message {
	if n <= 0 || n > len(ctx.history) {
		return ctx.history
	}

	start := len(ctx.history) - n
	return ctx.history[start:]
}

// GetSessionState returns the complete session state including variables and history.
func (ctx *NeuroContext) GetSessionState() types.SessionState {
	return types.SessionState{
		ID:        ctx.sessionID,
		Variables: ctx.variables,
		History:   ctx.history,
		CreatedAt: time.Now(), // Simplified for now
		UpdatedAt: time.Now(),
	}
}

func (ctx *NeuroContext) getSystemVariable(name string) (string, bool) {
	switch name {
	case "@pwd":
		if pwd, err := os.Getwd(); err == nil {
			return pwd, true
		}
	case "@user":
		if u, err := user.Current(); err == nil {
			return u.Username, true
		}
	case "@home":
		if home, err := os.UserHomeDir(); err == nil {
			return home, true
		}
	case "@date":
		return time.Now().Format("2006-01-02"), true
	case "@time":
		return time.Now().Format("15:04:05"), true
	case "@os":
		return fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")), true
	case "#session_id":
		return ctx.sessionID, true
	case "#message_count":
		return fmt.Sprintf("%d", len(ctx.history)), true
	case "#test_mode":
		if ctx.testMode {
			return "true", true
		}
		return "false", true
	}

	// Handle message history variables: ${1}, ${2}, etc.
	if matched, _ := regexp.MatchString(`^\d+$`, name); matched {
		// TODO: Implement message history access
		return fmt.Sprintf("message_%s_placeholder", name), true
	}

	return "", false
}

// InterpolateVariables replaces ${variable} placeholders in text with their values.
func (ctx *NeuroContext) InterpolateVariables(text string) string {
	// Early exit optimization - if no variables detected, return as-is
	if !strings.Contains(text, "${") {
		return text
	}

	// Iterative nested interpolation with safety limit
	maxIterations := 10 // Prevent infinite loops
	for i := 0; i < maxIterations; i++ {
		before := text
		text = ctx.interpolateOnce(text)

		// If no changes or no more variables, we're done
		if text == before || !strings.Contains(text, "${") {
			break
		}
	}

	return text
}

// interpolateOnce performs a single pass of variable interpolation
func (ctx *NeuroContext) interpolateOnce(text string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name (remove ${})
		varName := match[2 : len(match)-1]

		if value, err := ctx.GetVariable(varName); err == nil {
			return value
		}

		// Graceful handling: missing variables become empty string
		return ""
	})
}

// QueueCommand adds a command to the execution queue.
func (ctx *NeuroContext) QueueCommand(command string) {
	ctx.executionQueue = append(ctx.executionQueue, command)
}

// DequeueCommand removes and returns the first command from the queue.
func (ctx *NeuroContext) DequeueCommand() (string, bool) {
	if len(ctx.executionQueue) == 0 {
		return "", false
	}

	command := ctx.executionQueue[0]
	ctx.executionQueue = ctx.executionQueue[1:]
	return command, true
}

// GetQueueSize returns the number of commands in the execution queue.
func (ctx *NeuroContext) GetQueueSize() int {
	return len(ctx.executionQueue)
}

// ClearQueue removes all commands from the execution queue.
func (ctx *NeuroContext) ClearQueue() {
	ctx.executionQueue = make([]string, 0)
}

// PeekQueue returns a copy of the execution queue without modifying it.
func (ctx *NeuroContext) PeekQueue() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(ctx.executionQueue))
	copy(result, ctx.executionQueue)
	return result
}

// SetScriptMetadata stores metadata associated with script execution.
func (ctx *NeuroContext) SetScriptMetadata(key string, value interface{}) {
	ctx.scriptMetadata[key] = value
}

// GetScriptMetadata retrieves metadata by key, returning the value and existence flag.
func (ctx *NeuroContext) GetScriptMetadata(key string) (interface{}, bool) {
	value, exists := ctx.scriptMetadata[key]
	return value, exists
}

// ClearScriptMetadata removes all script metadata.
func (ctx *NeuroContext) ClearScriptMetadata() {
	ctx.scriptMetadata = make(map[string]interface{})
}

// SetTestMode enables or disables test mode for deterministic behavior.
func (ctx *NeuroContext) SetTestMode(testMode bool) {
	ctx.testMode = testMode
}

// IsTestMode returns whether test mode is currently enabled.
func (ctx *NeuroContext) IsTestMode() bool {
	return ctx.testMode
}
