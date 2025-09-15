// Package context provides session state management and variable interpolation for NeuroShell.
package context

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// allowedGlobalVariables defines which global variables (starting with _) can be set by users
var allowedGlobalVariables = []string{
	"_style",
	"_reply_way",
	"_echo_command",
	"_render_markdown",
	"_default_command",
	"_stream",
	"_editor",
	"_session_autosave",
	"_completion_mode",
	// Shell prompt configuration variables
	"_prompt_lines_count",
	"_prompt_line1",
	"_prompt_line2",
	"_prompt_line3",
	"_prompt_line4",
	"_prompt_line5",
}

// NeuroContext implements the neurotypes.Context interface providing session state management.
// It maintains variables, message history, metadata, and chat sessions for NeuroShell sessions.
type NeuroContext struct {
	variables      map[string]string
	history        []neurotypes.Message
	sessionID      string
	scriptMetadata map[string]interface{}
	testMode       bool

	// Stack-based execution support
	stackCtx       StackSubcontext // Delegated stack management
	variablesMutex sync.RWMutex    // Protects variables map

	// Chat session storage
	chatSessions    map[string]*neurotypes.ChatSession // Session storage by ID
	sessionNameToID map[string]string                  // Name to ID mapping
	activeSessionID string                             // Currently active session ID

	// Active model tracking
	activeModelID string // Currently active model ID

	// Model storage (bidirectional mapping)
	models        map[string]*neurotypes.ModelConfig // Model storage by ID
	modelNameToID map[string]string                  // Name to ID mapping

	// Provider registry management
	providerRegistryCtx ProviderRegistrySubcontext // Delegated provider registry management
	modelIDToName       map[string]string          // ID to name mapping

	// LLM client management
	llmClientCtx LLMClientSubcontext // Delegated LLM client management

	// Command registry management
	commandRegistryCtx CommandRegistrySubcontext // Delegated command registry management

	// Default command configuration
	defaultCommand string // Command to use when input doesn't start with \\

	// Configuration management
	configurationCtx ConfigurationSubcontext // Delegated configuration management

	// Script metadata protection
	scriptMutex sync.RWMutex // Protects scriptMetadata map

	// Error state management
	errorStateCtx ErrorStateSubcontext // Delegated error state management

	// Read-only command management
	readOnlyOverrides map[string]bool // Dynamic overrides: true=readonly, false=writable
	readOnlyMutex     sync.RWMutex    // Protects readOnlyOverrides
}

// New creates a new NeuroContext with initialized maps and a unique session ID.
func New() *NeuroContext {
	ctx := &NeuroContext{
		variables:      make(map[string]string),
		history:        make([]neurotypes.Message, 0),
		sessionID:      "", // Will be set after we know test mode
		scriptMetadata: make(map[string]interface{}),
		testMode:       false,

		// Initialize stack-based execution support
		stackCtx: NewStackSubcontext(),

		// Initialize chat session storage
		chatSessions:    make(map[string]*neurotypes.ChatSession),
		sessionNameToID: make(map[string]string),
		activeSessionID: "",

		// Initialize model storage
		models:        make(map[string]*neurotypes.ModelConfig),
		modelNameToID: make(map[string]string),
		modelIDToName: make(map[string]string),

		// Initialize LLM client management
		llmClientCtx: NewLLMClientSubcontext(),

		// Initialize provider registry management
		providerRegistryCtx: NewProviderRegistrySubcontext(),

		// Initialize command registry management
		commandRegistryCtx: NewCommandRegistrySubcontext(),

		// Initialize configuration management
		configurationCtx: NewConfigurationSubcontext(), // Parent context will be set after construction

		// Initialize read-only command management
		readOnlyOverrides: make(map[string]bool),

		// Initialize default command
		defaultCommand: "echo", // Default to echo for development convenience

		// Initialize error state management
		errorStateCtx: NewErrorStateSubcontext(),
	}

	// Generate initial session ID (will be deterministic if test mode is set later)
	ctx.sessionID = ctx.generateSessionID()

	// Initialize whitelisted global variables with default values
	_ = ctx.SetSystemVariable("_style", "")
	_ = ctx.SetSystemVariable("_default_command", "echo")
	_ = ctx.SetSystemVariable("_completion_mode", "tab")

	// Set up parent context reference for subcontexts
	ctx.configurationCtx.SetParentContext(ctx)

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

	// Handle user variables with mutex protection
	ctx.variablesMutex.RLock()
	defer ctx.variablesMutex.RUnlock()

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

	// Set variable with mutex protection
	ctx.variablesMutex.Lock()
	defer ctx.variablesMutex.Unlock()

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

	// Set variable with mutex protection
	ctx.variablesMutex.Lock()
	defer ctx.variablesMutex.Unlock()

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
	// Acquire read lock for variables access
	ctx.variablesMutex.RLock()
	defer ctx.variablesMutex.RUnlock()
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
		case "@status":
			status, _ := ctx.errorStateCtx.GetCurrentErrorState()
			return status, true
		case "@error":
			_, errorMsg := ctx.errorStateCtx.GetCurrentErrorState()
			return errorMsg, true
		case "@last_status":
			status, _ := ctx.errorStateCtx.GetLastErrorState()
			return status, true
		case "@last_error":
			_, errorMsg := ctx.errorStateCtx.GetLastErrorState()
			return errorMsg, true
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
	case "@status":
		status, _ := ctx.errorStateCtx.GetCurrentErrorState()
		return status, true
	case "@error":
		_, errorMsg := ctx.errorStateCtx.GetCurrentErrorState()
		return errorMsg, true
	case "@last_status":
		status, _ := ctx.errorStateCtx.GetLastErrorState()
		return status, true
	case "@last_error":
		_, errorMsg := ctx.errorStateCtx.GetLastErrorState()
		return errorMsg, true
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
	case "#session_display":
		// Return formatted session display for prompts
		sessionName, ok := ctx.variables["#session_name"]
		if ok && sessionName != "" {
			return " [" + sessionName + "]", true
		}
		return "", false
	case "#session_display_color":
		// Return formatted session display for colored prompts
		sessionName, ok := ctx.variables["#session_name"]
		if ok && sessionName != "" {
			return " [{{color:yellow}}" + sessionName + "{{/color}}]", true
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
		// Use lazy resolution - get the Nth most recent message from active session
		if n, err := strconv.Atoi(name); err == nil {
			value, err := ctx.GetNthRecentMessage(n)
			if err != nil {
				// Return empty string and false to indicate the variable doesn't exist
				// This maintains backward compatibility with existing tests and behavior
				return "", false
			}
			return value, true
		}
		return "", false
	}

	// Handle chronological message history variables: ${.1}, ${.2}, etc.
	if matched, _ := regexp.MatchString(`^\.(\d+)$`, name); matched {
		// Extract the number after the dot
		if matches := regexp.MustCompile(`^\.(\d+)$`).FindStringSubmatch(name); len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil {
				value, err := ctx.GetNthChronologicalMessage(n)
				if err != nil {
					// Return empty string and false to indicate the variable doesn't exist
					// This maintains same behavior as reverse order variables
					return "", false
				}
				return value, true
			}
		}
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

// interpolateOnce performs a single pass of variable interpolation with UTF-8 safety
// This preserves the original business logic but uses rune-based processing for UTF-8 safety
func (ctx *NeuroContext) interpolateOnce(text string) string {
	var stack []string
	pending := false

	// Convert to runes for proper UTF-8 handling
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		// Look for ${
		switch {
		case i < len(runes)-1 && runes[i] == '$' && runes[i+1] == '{':
			stack = append(stack, "${")
			pending = true
			i++ // Skip the '{'
		case runes[i] == '}' && pending:
			// Pop back to "${" marker to extract variable name
			varName := ""
			for len(stack) > 0 && stack[len(stack)-1] != "${" {
				varName = stack[len(stack)-1] + varName
				stack = stack[:len(stack)-1]
			}
			// Remove the "${" marker
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

			// Special case: empty variable name should be left as-is
			if varName == "" {
				stack = append(stack, "${}")
			} else {
				// Check for default value syntax: varname:-default
				var actualVarName, defaultValue string
				if strings.Contains(varName, ":-") {
					parts := strings.SplitN(varName, ":-", 2)
					actualVarName = parts[0]
					defaultValue = parts[1]
				} else {
					actualVarName = varName
					defaultValue = ""
				}

				// Get variable value and push to stack
				value, err := ctx.GetVariable(actualVarName)
				if err != nil || value == "" {
					// Use default value if variable doesn't exist or is empty
					stack = append(stack, defaultValue)
				} else {
					stack = append(stack, value)
				}
			}
			pending = false
		default:
			// Regular character or "}" when not pending - use rune to preserve UTF-8
			stack = append(stack, string(runes[i]))
		}
	}

	// Join all stack elements to form final result
	return strings.Join(stack, "")
}

// SetScriptMetadata stores metadata associated with script execution.
func (ctx *NeuroContext) SetScriptMetadata(key string, value interface{}) {
	ctx.scriptMutex.Lock()
	defer ctx.scriptMutex.Unlock()
	ctx.scriptMetadata[key] = value
}

// GetScriptMetadata retrieves metadata by key, returning the value and existence flag.
func (ctx *NeuroContext) GetScriptMetadata(key string) (interface{}, bool) {
	ctx.scriptMutex.RLock()
	defer ctx.scriptMutex.RUnlock()
	value, exists := ctx.scriptMetadata[key]
	return value, exists
}

// ClearScriptMetadata removes all script metadata.
func (ctx *NeuroContext) ClearScriptMetadata() {
	ctx.scriptMutex.Lock()
	defer ctx.scriptMutex.Unlock()
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
	return ctx.configurationCtx.GetEnv(key)
}

// GetAllVariables returns all variables including both user variables and computed system variables.
func (ctx *NeuroContext) GetAllVariables() map[string]string {
	result := make(map[string]string)

	// Add all stored variables (both user and system variables) with mutex protection
	ctx.variablesMutex.RLock()
	defer ctx.variablesMutex.RUnlock()

	for name, value := range ctx.variables {
		result[name] = value
	}

	// Add computed system variables
	systemVars := []string{"@pwd", "@user", "@home", "@date", "@time", "@os", "@status", "@error", "@last_status", "@last_error", "#session_id", "#message_count", "#test_mode"}
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

// GetNthRecentMessage returns the Nth most recent message from the active session.
// N=1 is the most recent message, N=2 is the previous message, etc.
// Returns the message content and an error if the message cannot be retrieved.
func (ctx *NeuroContext) GetNthRecentMessage(n int) (string, error) {
	// Delegate to session subcontext
	return NewSessionSubcontext(ctx).GetNthRecentMessage(n)
}

// GetNthChronologicalMessage returns the Nth message from the active session in chronological order.
// N=1 is the first message, N=2 is the second message, etc.
// Returns the message content and an error if the message cannot be retrieved.
func (ctx *NeuroContext) GetNthChronologicalMessage(n int) (string, error) {
	// Delegate to session subcontext
	return NewSessionSubcontext(ctx).GetNthChronologicalMessage(n)
}

// GetActiveModelID returns the currently active model ID.
func (ctx *NeuroContext) GetActiveModelID() string {
	return ctx.activeModelID
}

// SetActiveModelID sets the currently active model ID.
func (ctx *NeuroContext) SetActiveModelID(modelID string) {
	ctx.activeModelID = modelID
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

// GetLLMClient retrieves an LLM client by client ID (provider:hash format)
func (ctx *NeuroContext) GetLLMClient(clientID string) (neurotypes.LLMClient, bool) {
	return ctx.llmClientCtx.GetClient(clientID)
}

// SetLLMClient stores an LLM client by client ID (provider:hash format)
func (ctx *NeuroContext) SetLLMClient(clientID string, client neurotypes.LLMClient) {
	ctx.llmClientCtx.StoreClient(clientID, client)
}

// GetLLMClientCount returns the number of cached LLM clients (for testing/debugging)
func (ctx *NeuroContext) GetLLMClientCount() int {
	return len(ctx.llmClientCtx.GetAllClients())
}

// GetAllLLMClients returns a copy of all cached LLM clients (for client lookup)
func (ctx *NeuroContext) GetAllLLMClients() map[string]neurotypes.LLMClient {
	return ctx.llmClientCtx.GetAllClients()
}

// ClearLLMClients removes all cached LLM clients (for testing/debugging)
func (ctx *NeuroContext) ClearLLMClients() {
	ctx.llmClientCtx.ClearAllClients()
}

// RegisterCommand registers a command name for autocomplete functionality.
func (ctx *NeuroContext) RegisterCommand(commandName string) {
	ctx.commandRegistryCtx.RegisterCommand(commandName)
}

// RegisterCommandWithInfo registers a command with its metadata for help and autocomplete.
func (ctx *NeuroContext) RegisterCommandWithInfo(cmd neurotypes.Command) {
	ctx.commandRegistryCtx.RegisterCommandWithInfo(cmd)
}

// RegisterCommandWithInfoAndType registers a command with its metadata and type.
func (ctx *NeuroContext) RegisterCommandWithInfoAndType(cmd neurotypes.Command, cmdType neurotypes.CommandType) {
	ctx.commandRegistryCtx.RegisterCommandWithInfoAndType(cmd, cmdType)
}

// UnregisterCommand removes a command name from the autocomplete registry.
func (ctx *NeuroContext) UnregisterCommand(commandName string) {
	ctx.commandRegistryCtx.UnregisterCommand(commandName)
}

// GetRegisteredCommands returns a list of all registered command names.
func (ctx *NeuroContext) GetRegisteredCommands() []string {
	return ctx.commandRegistryCtx.GetRegisteredCommands()
}

// IsCommandRegistered checks if a command name is registered.
func (ctx *NeuroContext) IsCommandRegistered(commandName string) bool {
	return ctx.commandRegistryCtx.IsCommandRegistered(commandName)
}

// GetCommandHelpInfo returns the help information for a specific command.
func (ctx *NeuroContext) GetCommandHelpInfo(commandName string) (*neurotypes.HelpInfo, bool) {
	return ctx.commandRegistryCtx.GetCommandHelpInfo(commandName)
}

// GetAllCommandHelpInfo returns all registered command help information.
func (ctx *NeuroContext) GetAllCommandHelpInfo() map[string]*neurotypes.HelpInfo {
	return ctx.commandRegistryCtx.GetAllCommandHelpInfo()
}

// Stack operations for stack-based execution engine

// PushCommand adds a single command to the execution stack
func (ctx *NeuroContext) PushCommand(command string) {
	ctx.stackCtx.PushCommand(command)
}

// PushCommands adds multiple commands to the execution stack
func (ctx *NeuroContext) PushCommands(commands []string) {
	ctx.stackCtx.PushCommands(commands)
}

// PopCommand removes and returns the last command from the stack (LIFO)
func (ctx *NeuroContext) PopCommand() (string, bool) {
	return ctx.stackCtx.PopCommand()
}

// PeekCommand returns the next command without removing it from the stack
func (ctx *NeuroContext) PeekCommand() (string, bool) {
	return ctx.stackCtx.PeekCommand()
}

// ClearStack removes all commands from the execution stack
func (ctx *NeuroContext) ClearStack() {
	ctx.stackCtx.ClearStack()
}

// GetStackSize returns the number of commands in the execution stack
func (ctx *NeuroContext) GetStackSize() int {
	return ctx.stackCtx.GetStackSize()
}

// IsStackEmpty returns true if the stack is empty
func (ctx *NeuroContext) IsStackEmpty() bool {
	return ctx.stackCtx.IsStackEmpty()
}

// PeekStack returns a copy of the execution stack without modifying it
// Returns the stack in reverse order (top to bottom, LIFO order)
func (ctx *NeuroContext) PeekStack() []string {
	return ctx.stackCtx.PeekStack()
}

// Try block support methods

// PushErrorBoundary pushes error boundary markers for try blocks
func (ctx *NeuroContext) PushErrorBoundary(tryID string) {
	ctx.stackCtx.PushErrorBoundary(tryID)
}

// PopErrorBoundary removes the most recent try block context
func (ctx *NeuroContext) PopErrorBoundary() {
	ctx.stackCtx.PopErrorBoundary()
}

// IsInTryBlock returns true if currently inside a try block
func (ctx *NeuroContext) IsInTryBlock() bool {
	return ctx.stackCtx.IsInTryBlock()
}

// GetCurrentTryID returns the ID of the current try block
func (ctx *NeuroContext) GetCurrentTryID() string {
	return ctx.stackCtx.GetCurrentTryID()
}

// GetCurrentTryDepth returns the current try block depth
func (ctx *NeuroContext) GetCurrentTryDepth() int {
	return ctx.stackCtx.GetCurrentTryDepth()
}

// SetTryErrorCaptured marks the current try block as having captured an error
func (ctx *NeuroContext) SetTryErrorCaptured() {
	ctx.stackCtx.SetTryErrorCaptured()
}

// IsTryErrorCaptured returns true if the current try block has captured an error
func (ctx *NeuroContext) IsTryErrorCaptured() bool {
	return ctx.stackCtx.IsTryErrorCaptured()
}

// PushSilentBoundary pushes silent boundary markers for silent blocks
func (ctx *NeuroContext) PushSilentBoundary(silentID string) {
	ctx.stackCtx.PushSilentBoundary(silentID)
}

// PopSilentBoundary removes the most recent silent block context
func (ctx *NeuroContext) PopSilentBoundary() {
	ctx.stackCtx.PopSilentBoundary()
}

// IsInSilentBlock returns true if currently inside a silent block
func (ctx *NeuroContext) IsInSilentBlock() bool {
	return ctx.stackCtx.IsInSilentBlock()
}

// GetCurrentSilentID returns the ID of the current silent block
func (ctx *NeuroContext) GetCurrentSilentID() string {
	return ctx.stackCtx.GetCurrentSilentID()
}

// GetCurrentSilentDepth returns the current silent block depth
func (ctx *NeuroContext) GetCurrentSilentDepth() int {
	return ctx.stackCtx.GetCurrentSilentDepth()
}

// GetDefaultCommand returns the default command to use when input doesn't start with \\
func (ctx *NeuroContext) GetDefaultCommand() string {
	// Check if _default_command variable overrides the default
	if override, exists := ctx.getSystemVariable("_default_command"); exists && override != "" {
		return override
	}
	return ctx.defaultCommand
}

// SetDefaultCommand sets the default command to use when input doesn't start with \\
func (ctx *NeuroContext) SetDefaultCommand(command string) {
	ctx.defaultCommand = command
	_ = ctx.SetSystemVariable("_default_command", command)
}

// SetTestEnvOverride sets a test-specific environment variable override.
// This allows tests to control what GetEnv returns for specific keys without affecting the OS environment.
func (ctx *NeuroContext) SetTestEnvOverride(key, value string) {
	ctx.configurationCtx.SetTestEnvOverride(key, value)
}

// SetEnvVariable sets an environment variable, respecting test mode.
// In test mode, this sets a test environment override.
// In production mode, this sets an actual OS environment variable.
func (ctx *NeuroContext) SetEnvVariable(key, value string) error {
	return ctx.configurationCtx.SetEnvVariable(key, value)
}

// GetEnvVariable retrieves an environment variable value, respecting test mode.
// This is a pure function that only gets the environment variable without side effects.
func (ctx *NeuroContext) GetEnvVariable(key string) string {
	return ctx.configurationCtx.GetEnvVariable(key)
}

// ClearTestEnvOverride removes a test-specific environment variable override.
func (ctx *NeuroContext) ClearTestEnvOverride(key string) {
	ctx.configurationCtx.ClearTestEnvOverride(key)
}

// ClearAllTestEnvOverrides removes all test-specific environment variable overrides.
func (ctx *NeuroContext) ClearAllTestEnvOverrides() {
	ctx.configurationCtx.ClearAllTestEnvOverrides()
}

// GetTestEnvOverrides returns a copy of all test environment variable overrides.
func (ctx *NeuroContext) GetTestEnvOverrides() map[string]string {
	return ctx.configurationCtx.GetTestEnvOverrides()
}

// File system operations for configuration service

// ReadFile reads the contents of a file, supporting test mode isolation.
func (ctx *NeuroContext) ReadFile(path string) ([]byte, error) {
	return ctx.configurationCtx.ReadFile(path)
}

// WriteFile writes data to a file with the specified permissions, supporting test mode isolation.
func (ctx *NeuroContext) WriteFile(path string, data []byte, perm os.FileMode) error {
	return ctx.configurationCtx.WriteFile(path, data, perm)
}

// FileExists checks if a file exists at the given path.
func (ctx *NeuroContext) FileExists(path string) bool {
	return ctx.configurationCtx.FileExists(path)
}

// GetUserConfigDir returns the user's configuration directory.
// In test mode, returns a temporary directory to avoid polluting the user's system.
func (ctx *NeuroContext) GetUserConfigDir() (string, error) {
	return ctx.configurationCtx.GetUserConfigDir()
}

// GetWorkingDir returns the current working directory.
func (ctx *NeuroContext) GetWorkingDir() (string, error) {
	return ctx.configurationCtx.GetWorkingDir()
}

// MkdirAll creates a directory path with the specified permissions, including any necessary parents.
func (ctx *NeuroContext) MkdirAll(path string, perm os.FileMode) error {
	return ctx.configurationCtx.MkdirAll(path, perm)
}

// Configuration management methods

// GetConfigMap returns a copy of the configuration map.
func (ctx *NeuroContext) GetConfigMap() map[string]string {
	return ctx.configurationCtx.GetConfigMap()
}

// SetConfigMap replaces the entire configuration map.
func (ctx *NeuroContext) SetConfigMap(configMap map[string]string) {
	ctx.configurationCtx.SetConfigMap(configMap)
}

// GetConfigValue retrieves a configuration value by key.
func (ctx *NeuroContext) GetConfigValue(key string) (string, bool) {
	return ctx.configurationCtx.GetConfigValue(key)
}

// SetConfigValue sets a configuration value.
func (ctx *NeuroContext) SetConfigValue(key, value string) {
	ctx.configurationCtx.SetConfigValue(key, value)
}

// Configuration loading methods (Context layer responsibilities)

// LoadDefaults sets up default configuration values.
func (ctx *NeuroContext) LoadDefaults() error {
	return ctx.configurationCtx.LoadDefaults()
}

// LoadConfigDotEnv loads .env file from the user's config directory (~/.config/neuroshell/.env).
func (ctx *NeuroContext) LoadConfigDotEnv() error {
	return ctx.configurationCtx.LoadConfigDotEnv()
}

// LoadLocalDotEnv loads .env file from the current working directory.
func (ctx *NeuroContext) LoadLocalDotEnv() error {
	return ctx.configurationCtx.LoadLocalDotEnv()
}

// LoadEnvironmentVariables loads specific prefixed environment variables into context configuration map.
// This has the highest priority and will override all file-based configuration.
func (ctx *NeuroContext) LoadEnvironmentVariables(prefixes []string) error {
	return ctx.configurationCtx.LoadEnvironmentVariables(prefixes)
}

// LoadEnvironmentVariablesWithPrefix loads OS environment variables with a source prefix.
func (ctx *NeuroContext) LoadEnvironmentVariablesWithPrefix(sourcePrefix string) error {
	return ctx.configurationCtx.LoadEnvironmentVariablesWithPrefix(sourcePrefix)
}

// LoadConfigDotEnvWithPrefix loads config .env file with a source prefix.
func (ctx *NeuroContext) LoadConfigDotEnvWithPrefix(sourcePrefix string) error {
	return ctx.configurationCtx.LoadConfigDotEnvWithPrefix(sourcePrefix)
}

// LoadLocalDotEnvWithPrefix loads local .env file with a source prefix.
func (ctx *NeuroContext) LoadLocalDotEnvWithPrefix(sourcePrefix string) error {
	return ctx.configurationCtx.LoadLocalDotEnvWithPrefix(sourcePrefix)
}

// Provider Registry Methods

// GetSupportedProviders returns the list of supported LLM provider names.
// This is the central source of truth for all provider-related functionality.
func (ctx *NeuroContext) GetSupportedProviders() []string {
	return ctx.providerRegistryCtx.GetSupportedProviders()
}

// GetProviderEnvPrefixes returns the list of environment variable prefixes
// used for loading provider-specific configuration from the environment.
func (ctx *NeuroContext) GetProviderEnvPrefixes() []string {
	return ctx.providerRegistryCtx.GetProviderEnvPrefixes()
}

// IsValidProvider checks if a given provider name is supported.
// Provider comparison is case-insensitive.
func (ctx *NeuroContext) IsValidProvider(provider string) bool {
	return ctx.providerRegistryCtx.IsProviderSupported(provider)
}

// Error state management methods

// ResetErrorState resets the current error state to success (0/"") and moves current to last.
// This should be called before executing a new command.
func (ctx *NeuroContext) ResetErrorState() {
	ctx.errorStateCtx.ResetErrorState()
}

// SetErrorState sets the current error state based on command execution results.
// This should be called after command execution with the results.
func (ctx *NeuroContext) SetErrorState(status string, errorMsg string) {
	ctx.errorStateCtx.SetErrorState(status, errorMsg)
}

// GetCurrentErrorState returns the current error state (thread-safe read).
func (ctx *NeuroContext) GetCurrentErrorState() (status string, errorMsg string) {
	return ctx.errorStateCtx.GetCurrentErrorState()
}

// GetLastErrorState returns the last error state (thread-safe read).
func (ctx *NeuroContext) GetLastErrorState() (status string, errorMsg string) {
	return ctx.errorStateCtx.GetLastErrorState()
}

// SetCommandReadOnly sets or removes a read-only override for a specific command.
// This allows dynamic configuration of read-only status at runtime.
func (ctx *NeuroContext) SetCommandReadOnly(commandName string, readOnly bool) {
	ctx.readOnlyMutex.Lock()
	defer ctx.readOnlyMutex.Unlock()

	ctx.readOnlyOverrides[commandName] = readOnly
}

// RemoveCommandReadOnlyOverride removes any read-only override for a command,
// reverting to the command's self-declared IsReadOnly() status.
func (ctx *NeuroContext) RemoveCommandReadOnlyOverride(commandName string) {
	ctx.readOnlyMutex.Lock()
	defer ctx.readOnlyMutex.Unlock()

	delete(ctx.readOnlyOverrides, commandName)
}

// IsCommandReadOnly checks if a command is read-only by considering both
// the command's self-declared status and any dynamic overrides.
// Dynamic overrides take precedence over self-declared status.
func (ctx *NeuroContext) IsCommandReadOnly(cmd neurotypes.Command) bool {
	ctx.readOnlyMutex.RLock()
	defer ctx.readOnlyMutex.RUnlock()

	// Check for dynamic override first
	if override, exists := ctx.readOnlyOverrides[cmd.Name()]; exists {
		return override
	}

	// Fall back to command's self-declared read-only status
	return cmd.IsReadOnly()
}

// GetReadOnlyOverrides returns a copy of all current read-only overrides.
// This is useful for configuration services and debugging.
func (ctx *NeuroContext) GetReadOnlyOverrides() map[string]bool {
	ctx.readOnlyMutex.RLock()
	defer ctx.readOnlyMutex.RUnlock()

	// Return a copy to prevent external modification
	overrides := make(map[string]bool)
	for name, readOnly := range ctx.readOnlyOverrides {
		overrides[name] = readOnly
	}
	return overrides
}
