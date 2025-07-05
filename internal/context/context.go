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

	"neuroshell/pkg/neurotypes"
)

// NeuroContext implements the neurotypes.Context interface providing session state management.
// It maintains variables, message history, execution queues, metadata, and chat sessions for NeuroShell sessions.
type NeuroContext struct {
	variables      map[string]string
	history        []neurotypes.Message
	sessionID      string
	executionQueue []string
	scriptMetadata map[string]interface{}
	testMode       bool

	// Chat session storage
	chatSessions    map[string]*neurotypes.ChatSession // Session storage by ID
	sessionNameToID map[string]string                  // Name to ID mapping
	activeSessionID string                             // Currently active session ID

	// Model storage (bidirectional mapping)
	models        map[string]*neurotypes.ModelConfig // Model storage by ID
	modelNameToID map[string]string                  // Name to ID mapping
	modelIDToName map[string]string                  // ID to name mapping
}

// New creates a new NeuroContext with initialized maps and a unique session ID.
func New() *NeuroContext {
	ctx := &NeuroContext{
		variables:      make(map[string]string),
		history:        make([]neurotypes.Message, 0),
		sessionID:      "", // Will be set after we know test mode
		executionQueue: make([]string, 0),
		scriptMetadata: make(map[string]interface{}),
		testMode:       false,

		// Initialize chat session storage
		chatSessions:    make(map[string]*neurotypes.ChatSession),
		sessionNameToID: make(map[string]string),
		activeSessionID: "",

		// Initialize model storage
		models:        make(map[string]*neurotypes.ModelConfig),
		modelNameToID: make(map[string]string),
		modelIDToName: make(map[string]string),
	}

	// Generate initial session ID (will be deterministic if test mode is set later)
	ctx.sessionID = ctx.generateSessionID()
	return ctx
}

// generateSessionID creates a session ID, deterministic in test mode
func (ctx *NeuroContext) generateSessionID() string {
	if ctx.testMode {
		return "session_1609459200" // Fixed timestamp for test mode
	}
	return fmt.Sprintf("session_%d", time.Now().Unix())
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

// SetSystemVariable sets a system variable, allowing internal app use only.
// This method is for internal use by the application and should not be exposed to users.
func (ctx *NeuroContext) SetSystemVariable(name string, value string) error {
	// Only allow setting system variables (prefixed with @, #, or _)
	if !strings.HasPrefix(name, "@") && !strings.HasPrefix(name, "#") && !strings.HasPrefix(name, "_") {
		return fmt.Errorf("SetSystemVariable can only set system variables (prefixed with @, #, or _), got: %s", name)
	}

	ctx.variables[name] = value
	return nil
}

// GetMessageHistory returns the last n messages from the conversation history.
func (ctx *NeuroContext) GetMessageHistory(n int) []neurotypes.Message {
	if n <= 0 || n > len(ctx.history) {
		return ctx.history
	}

	start := len(ctx.history) - n
	return ctx.history[start:]
}

// GetSessionState returns the complete session state including variables and history.
func (ctx *NeuroContext) GetSessionState() neurotypes.SessionState {
	return neurotypes.SessionState{
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
		// Check if there's a stored chat session ID first
		if value, ok := ctx.variables["#session_id"]; ok {
			return value, true
		}
		// Fall back to shell session ID
		return ctx.sessionID, true
	case "#message_count":
		// Check if there's a stored session message count first
		if value, ok := ctx.variables["#message_count"]; ok {
			return value, true
		}
		// Fall back to NeuroShell context history count
		return fmt.Sprintf("%d", len(ctx.history)), true
	case "#test_mode":
		if ctx.testMode {
			return "true", true
		}
		return "false", true
	case "#session_name":
		// Look for stored session name variable
		if value, ok := ctx.variables["#session_name"]; ok {
			return value, true
		}
		return "", false
	case "#system_prompt":
		// Look for stored system prompt variable
		if value, ok := ctx.variables["#system_prompt"]; ok {
			return value, true
		}
		return "", false
	case "#session_created":
		// Look for stored session creation time variable
		if value, ok := ctx.variables["#session_created"]; ok {
			return value, true
		}
		return "", false
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
	result := strings.Builder{}
	i := 0

	for i < len(text) {
		// Look for ${
		if i < len(text)-1 && text[i] == '$' && text[i+1] == '{' {
			// Find the matching closing brace
			braceCount := 1
			start := i + 2 // Position after ${
			end := start

			for end < len(text) && braceCount > 0 {
				switch text[end] {
				case '{':
					braceCount++
				case '}':
					braceCount--
				}
				if braceCount > 0 {
					end++
				}
			}

			if braceCount == 0 {
				// Found matching closing brace
				varName := text[start:end]

				// Special case: empty variable name should be left as-is
				if varName == "" {
					result.WriteString("${}")
				} else if value, err := ctx.GetVariable(varName); err == nil {
					result.WriteString(value)
				}
				// If variable doesn't exist, write nothing (empty string)

				i = end + 1 // Move past the closing brace
			} else {
				// No matching closing brace, treat as literal text
				result.WriteByte(text[i])
				i++
			}
		} else {
			// Regular character
			result.WriteByte(text[i])
			i++
		}
	}

	return result.String()
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
	// Regenerate session ID to be deterministic in test mode
	ctx.sessionID = ctx.generateSessionID()
}

// IsTestMode returns whether test mode is currently enabled.
func (ctx *NeuroContext) IsTestMode() bool {
	return ctx.testMode
}

// GetAllVariables returns all variables including both user variables and computed system variables.
func (ctx *NeuroContext) GetAllVariables() map[string]string {
	result := make(map[string]string)

	// Add all stored variables (both user and system variables)
	for name, value := range ctx.variables {
		result[name] = value
	}

	// Add computed system variables
	systemVars := []string{"@pwd", "@user", "@home", "@date", "@time", "@os", "#session_id", "#message_count", "#test_mode"}
	for _, varName := range systemVars {
		if value, ok := ctx.getSystemVariable(varName); ok {
			result[varName] = value
		}
	}

	return result
}

// GetChatSessions returns all chat sessions stored in the context.
func (ctx *NeuroContext) GetChatSessions() map[string]*neurotypes.ChatSession {
	return ctx.chatSessions
}

// SetChatSessions sets the chat sessions map in the context.
func (ctx *NeuroContext) SetChatSessions(sessions map[string]*neurotypes.ChatSession) {
	ctx.chatSessions = sessions
}

// GetSessionNameToID returns the session name to ID mapping.
func (ctx *NeuroContext) GetSessionNameToID() map[string]string {
	return ctx.sessionNameToID
}

// SetSessionNameToID sets the session name to ID mapping in the context.
func (ctx *NeuroContext) SetSessionNameToID(nameToID map[string]string) {
	ctx.sessionNameToID = nameToID
}

// GetActiveSessionID returns the currently active session ID.
func (ctx *NeuroContext) GetActiveSessionID() string {
	return ctx.activeSessionID
}

// SetActiveSessionID sets the currently active session ID.
func (ctx *NeuroContext) SetActiveSessionID(sessionID string) {
	ctx.activeSessionID = sessionID
}

// GetModels returns all model configurations stored in the context.
func (ctx *NeuroContext) GetModels() map[string]*neurotypes.ModelConfig {
	return ctx.models
}

// SetModels sets the model configurations map in the context.
func (ctx *NeuroContext) SetModels(models map[string]*neurotypes.ModelConfig) {
	ctx.models = models
}

// GetModelNameToID returns the model name to ID mapping.
func (ctx *NeuroContext) GetModelNameToID() map[string]string {
	return ctx.modelNameToID
}

// SetModelNameToID sets the model name to ID mapping in the context.
func (ctx *NeuroContext) SetModelNameToID(nameToID map[string]string) {
	ctx.modelNameToID = nameToID
}

// GetModelIDToName returns the model ID to name mapping.
func (ctx *NeuroContext) GetModelIDToName() map[string]string {
	return ctx.modelIDToName
}

// SetModelIDToName sets the model ID to name mapping in the context.
func (ctx *NeuroContext) SetModelIDToName(idToName map[string]string) {
	ctx.modelIDToName = idToName
}

// ModelNameExists checks if a model name already exists in the context.
func (ctx *NeuroContext) ModelNameExists(name string) bool {
	_, exists := ctx.modelNameToID[name]
	return exists
}

// ModelIDExists checks if a model ID already exists in the context.
func (ctx *NeuroContext) ModelIDExists(id string) bool {
	_, exists := ctx.models[id]
	return exists
}
