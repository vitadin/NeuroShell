// Package context provides session state management and variable interpolation for NeuroShell.
// It maintains variables, message history, execution queues, and system metadata across command executions.
package context

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"
	"sync"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// allowedGlobalVariables defines which global variables (starting with _) can be set by users
var allowedGlobalVariables = []string{
	"_style",
	"_reply_way",
	"_echo_commands",
	"_render_markdown",
}

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

	// LLM client storage
	llmClients map[string]neurotypes.LLMClient // LLM client storage by API key identifier

	// Command registry information
	registeredCommands map[string]bool                 // Track registered command names for autocomplete
	commandHelpInfo    map[string]*neurotypes.HelpInfo // Store detailed help info for autocomplete and help system
	commandMutex       sync.RWMutex                    // Protects registeredCommands and commandHelpInfo maps
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

		// Initialize LLM client storage
		llmClients: make(map[string]neurotypes.LLMClient),

		// Initialize command registry information
		registeredCommands: make(map[string]bool),
		commandHelpInfo:    make(map[string]*neurotypes.HelpInfo),
	}

	// Generate initial session ID (will be deterministic if test mode is set later)
	ctx.sessionID = ctx.generateSessionID()

	// Initialize whitelisted global variables with default values
	_ = ctx.SetSystemVariable("_style", "")

	return ctx
}

// NewTestContext creates a clean real context for testing purposes.
// This function is designed to be used by unit tests across the codebase.
func NewTestContext() neurotypes.Context {
	ResetGlobalContext()
	ctx := GetGlobalContext()
	ctx.SetTestMode(true)
	return ctx
}

// generateSessionID creates a session ID, deterministic in test mode
func (ctx *NeuroContext) generateSessionID() string {
	return testutils.GenerateSessionID(ctx)
}

// GetVariable retrieves a variable value by name, supporting both user and system variables.
func (ctx *NeuroContext) GetVariable(name string) (string, error) {
	// Handle special variables
	if value, ok := ctx.getSystemVariable(name); ok {
		return value, nil
	}

	// Handle user variables
	value, ok := ctx.variables[name]

	if ok {
		return value, nil
	}

	// Return empty string for non-existent variables
	return "", nil
}

// SetVariable sets a user variable, preventing modification of system variables.
func (ctx *NeuroContext) SetVariable(name string, value string) error {
	// Don't allow setting system variables with @ or # prefixes
	if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") {
		return fmt.Errorf("cannot set system variable: %s", name)
	}

	// For variables with _ prefix, check whitelist
	if strings.HasPrefix(name, "_") {
		allowed := false
		for _, allowedVar := range allowedGlobalVariables {
			if name == allowedVar {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("cannot set system variable: %s", name)
		}
	}

	ctx.variables[name] = value
	return nil
}

// SetVariableWithValidation is an alias for SetVariable for backward compatibility in tests.
func (ctx *NeuroContext) SetVariableWithValidation(name string, value string) error {
	return ctx.SetVariable(name, value)
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
	now := testutils.GetCurrentTime(ctx)
	return neurotypes.SessionState{
		ID:        ctx.sessionID,
		Variables: ctx.variables,
		History:   ctx.history,
		CreatedAt: now, // Deterministic time in test mode
		UpdatedAt: now,
	}
}

func (ctx *NeuroContext) getSystemVariable(name string) (string, bool) {
	// In test mode, return fixed values for consistency
	if ctx.testMode {
		switch name {
		case "@pwd":
			return "/test/pwd", true
		case "@user":
			return "testuser", true
		case "@home":
			return "/test/home", true
		case "@date":
			return "2024-01-01", true
		case "@time":
			return "12:00:00", true
		case "@os":
			return "test-os", true
		}
	}

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
		return testutils.GetCurrentTime(ctx).Format("2006-01-02"), true
	case "@time":
		return testutils.GetCurrentTime(ctx).Format("15:04:05"), true
	case "@os":
		return fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")), true
	case "#session_id":
		// Check if there's a stored chat session ID first
		value, ok := ctx.variables["#session_id"]
		if ok {
			return value, true
		}
		// Fall back to shell session ID
		return ctx.sessionID, true
	case "#message_count":
		// Check if there's a stored session message count first
		value, ok := ctx.variables["#message_count"]
		if ok {
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
		value, ok := ctx.variables["#session_name"]
		if ok {
			return value, true
		}
		return "", false
	case "#system_prompt":
		// Look for stored system prompt variable
		value, ok := ctx.variables["#system_prompt"]
		if ok {
			return value, true
		}
		return "", false
	case "#session_created":
		// Look for stored session creation time variable
		value, ok := ctx.variables["#session_created"]
		if ok {
			return value, true
		}
		return "", false
	}

	// Handle message history variables: ${1}, ${2}, etc.
	if matched, _ := regexp.MatchString(`^\d+$`, name); matched {
		// Check if the variable is stored in the regular variables map first
		value, ok := ctx.variables[name]
		if ok {
			return value, true
		}
		// Return empty string if not found (was previously returning placeholder)
		return "", false
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

// GetEnv retrieves environment variables, providing test mode appropriate values.
// In test mode, returns predefined test values. In normal mode, returns os.Getenv().
func (ctx *NeuroContext) GetEnv(key string) string {
	if ctx.IsTestMode() {
		return ctx.getTestEnvValue(key)
	}
	return os.Getenv(key)
}

// getTestEnvValue returns test mode appropriate values for environment variables.
func (ctx *NeuroContext) getTestEnvValue(key string) string {
	switch key {
	case "OPENAI_API_KEY":
		return "test-openai-key"
	case "ANTHROPIC_API_KEY":
		return "test-anthropic-key"
	case "EDITOR":
		return "test-editor"
	case "GOOS":
		return "test-os"
	case "GOARCH":
		return "test-arch"
	default:
		return ""
	}
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

// GetLLMClient retrieves an LLM client by API key identifier
func (ctx *NeuroContext) GetLLMClient(apiKey string) (neurotypes.LLMClient, bool) {
	client, exists := ctx.llmClients[apiKey]
	return client, exists
}

// SetLLMClient stores an LLM client by API key identifier
func (ctx *NeuroContext) SetLLMClient(apiKey string, client neurotypes.LLMClient) {
	ctx.llmClients[apiKey] = client
}

// RegisterCommand registers a command name for autocomplete functionality.
func (ctx *NeuroContext) RegisterCommand(commandName string) {
	ctx.commandMutex.Lock()
	defer ctx.commandMutex.Unlock()
	ctx.registeredCommands[commandName] = true
}

// RegisterCommandWithInfo registers a command with its metadata for help and autocomplete.
func (ctx *NeuroContext) RegisterCommandWithInfo(cmd neurotypes.Command) {
	ctx.commandMutex.Lock()
	defer ctx.commandMutex.Unlock()

	commandName := cmd.Name()
	ctx.registeredCommands[commandName] = true

	// Store command help information
	helpInfo := cmd.HelpInfo()
	ctx.commandHelpInfo[commandName] = &helpInfo
}

// RegisterCommandWithInfoAndType registers a command with its metadata and type.
func (ctx *NeuroContext) RegisterCommandWithInfoAndType(cmd neurotypes.Command, _ neurotypes.CommandType) {
	ctx.commandMutex.Lock()
	defer ctx.commandMutex.Unlock()

	commandName := cmd.Name()
	ctx.registeredCommands[commandName] = true

	// Store command help information
	helpInfo := cmd.HelpInfo()
	ctx.commandHelpInfo[commandName] = &helpInfo
}

// UnregisterCommand removes a command name from the autocomplete registry.
func (ctx *NeuroContext) UnregisterCommand(commandName string) {
	ctx.commandMutex.Lock()
	defer ctx.commandMutex.Unlock()
	delete(ctx.registeredCommands, commandName)
	delete(ctx.commandHelpInfo, commandName)
}

// GetRegisteredCommands returns a list of all registered command names.
func (ctx *NeuroContext) GetRegisteredCommands() []string {
	ctx.commandMutex.RLock()
	defer ctx.commandMutex.RUnlock()

	commands := make([]string, 0, len(ctx.registeredCommands))
	for commandName := range ctx.registeredCommands {
		commands = append(commands, commandName)
	}
	return commands
}

// IsCommandRegistered checks if a command name is registered.
func (ctx *NeuroContext) IsCommandRegistered(commandName string) bool {
	ctx.commandMutex.RLock()
	defer ctx.commandMutex.RUnlock()
	return ctx.registeredCommands[commandName]
}

// GetCommandHelpInfo returns the help information for a specific command.
func (ctx *NeuroContext) GetCommandHelpInfo(commandName string) (*neurotypes.HelpInfo, bool) {
	ctx.commandMutex.RLock()
	defer ctx.commandMutex.RUnlock()
	info, exists := ctx.commandHelpInfo[commandName]
	return info, exists
}

// GetAllCommandHelpInfo returns all registered command help information.
func (ctx *NeuroContext) GetAllCommandHelpInfo() map[string]*neurotypes.HelpInfo {
	ctx.commandMutex.RLock()
	defer ctx.commandMutex.RUnlock()

	result := make(map[string]*neurotypes.HelpInfo)
	for name, info := range ctx.commandHelpInfo {
		result[name] = info
	}
	return result
}
