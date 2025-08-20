// Package context provides session state management and variable interpolation for NeuroShell.
package context

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"

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
}

// TryBlockContext represents the context for a try block with error boundaries
type TryBlockContext struct {
	ID            string // Unique identifier for this try block
	StartDepth    int    // Stack depth when try block started
	ErrorCaptured bool   // Whether an error has been captured
}

// SilentBlockContext represents the context for a silent block with output suppression
type SilentBlockContext struct {
	ID         string // Unique identifier for this silent block
	StartDepth int    // Stack depth when silent block started
}

// NeuroContext implements the neurotypes.Context interface providing session state management.
// It maintains variables, message history, metadata, and chat sessions for NeuroShell sessions.
type NeuroContext struct {
	variables        map[string]string
	history          []neurotypes.Message
	sessionID        string
	scriptMetadata   map[string]interface{}
	testMode         bool
	testEnvOverrides map[string]string // Test-specific environment variable overrides

	// Stack-based execution support
	executionStack     []string             // Execution stack (LIFO order)
	tryBlocks          []TryBlockContext    // Try block management
	currentTryDepth    int                  // Current try block depth
	silentBlocks       []SilentBlockContext // Silent block management
	currentSilentDepth int                  // Current silent block depth
	stackMutex         sync.RWMutex         // Protects executionStack, tryBlocks, and silentBlocks

	// Chat session storage
	chatSessions    map[string]*neurotypes.ChatSession // Session storage by ID
	sessionNameToID map[string]string                  // Name to ID mapping
	activeSessionID string                             // Currently active session ID

	// Active model tracking
	activeModelID string // Currently active model ID

	// Model storage (bidirectional mapping)
	models        map[string]*neurotypes.ModelConfig // Model storage by ID
	modelNameToID map[string]string                  // Name to ID mapping

	// Provider registry - central source of truth for supported providers
	supportedProviders  []string          // Supported LLM provider names (lowercase)
	providerEnvPrefixes []string          // Environment variable prefixes for provider detection
	modelIDToName       map[string]string // ID to name mapping

	// LLM client storage
	llmClients map[string]neurotypes.LLMClient // LLM client storage by client ID (provider:hash format)

	// Command registry information
	registeredCommands map[string]bool                 // Track registered command names for autocomplete
	commandHelpInfo    map[string]*neurotypes.HelpInfo // Store detailed help info for autocomplete and help system
	commandMutex       sync.RWMutex                    // Protects registeredCommands and commandHelpInfo maps

	// Default command configuration
	defaultCommand string // Command to use when input doesn't start with \\

	// Configuration management
	configMap   map[string]string // Configuration key-value store
	configMutex sync.RWMutex      // Protects configMap

	// Script metadata protection
	scriptMutex sync.RWMutex // Protects scriptMetadata map

	// Error state management
	lastStatus      string       // Last command's exit status
	lastError       string       // Last command's error message
	currentStatus   string       // Current command's exit status (0 = success, non-zero = error)
	currentError    string       // Current command's error message
	errorStateMutex sync.RWMutex // Protects error state fields

	// Read-only command management
	readOnlyOverrides map[string]bool // Dynamic overrides: true=readonly, false=writable
	readOnlyMutex     sync.RWMutex    // Protects readOnlyOverrides
}

// New creates a new NeuroContext with initialized maps and a unique session ID.
func New() *NeuroContext {
	ctx := &NeuroContext{
		variables:        make(map[string]string),
		history:          make([]neurotypes.Message, 0),
		sessionID:        "", // Will be set after we know test mode
		scriptMetadata:   make(map[string]interface{}),
		testMode:         false,
		testEnvOverrides: make(map[string]string),

		// Initialize stack-based execution support
		executionStack:     make([]string, 0),
		tryBlocks:          make([]TryBlockContext, 0),
		currentTryDepth:    0,
		silentBlocks:       make([]SilentBlockContext, 0),
		currentSilentDepth: 0,

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

		// Initialize provider registry
		supportedProviders:  []string{"openai", "anthropic", "openrouter", "moonshot", "gemini"},
		providerEnvPrefixes: []string{"NEURO_", "OPENAI_", "ANTHROPIC_", "MOONSHOT_", "GOOGLE_"},

		// Initialize command registry information
		registeredCommands: make(map[string]bool),
		commandHelpInfo:    make(map[string]*neurotypes.HelpInfo),

		// Initialize configuration management
		configMap: make(map[string]string),

		// Initialize read-only command management
		readOnlyOverrides: make(map[string]bool),

		// Initialize default command
		defaultCommand: "echo", // Default to echo for development convenience

		// Initialize error state management (start with success state)
		lastStatus:    "0",
		lastError:     "",
		currentStatus: "0",
		currentError:  "",
	}

	// Generate initial session ID (will be deterministic if test mode is set later)
	ctx.sessionID = ctx.generateSessionID()

	// Initialize whitelisted global variables with default values
	_ = ctx.SetSystemVariable("_style", "")
	_ = ctx.SetSystemVariable("_default_command", "echo")

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
		case "@status":
			ctx.errorStateMutex.RLock()
			defer ctx.errorStateMutex.RUnlock()
			return ctx.currentStatus, true
		case "@error":
			ctx.errorStateMutex.RLock()
			defer ctx.errorStateMutex.RUnlock()
			return ctx.currentError, true
		case "@last_status":
			ctx.errorStateMutex.RLock()
			defer ctx.errorStateMutex.RUnlock()
			return ctx.lastStatus, true
		case "@last_error":
			ctx.errorStateMutex.RLock()
			defer ctx.errorStateMutex.RUnlock()
			return ctx.lastError, true
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
		ctx.errorStateMutex.RLock()
		defer ctx.errorStateMutex.RUnlock()
		return ctx.currentStatus, true
	case "@error":
		ctx.errorStateMutex.RLock()
		defer ctx.errorStateMutex.RUnlock()
		return ctx.currentError, true
	case "@last_status":
		ctx.errorStateMutex.RLock()
		defer ctx.errorStateMutex.RUnlock()
		return ctx.lastStatus, true
	case "@last_error":
		ctx.errorStateMutex.RLock()
		defer ctx.errorStateMutex.RUnlock()
		return ctx.lastError, true
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
				// Get variable value and push to stack
				value, _ := ctx.GetVariable(varName)
				stack = append(stack, value)
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
	if ctx.IsTestMode() {
		return ctx.getTestEnvValue(key)
	}
	return os.Getenv(key)
}

// getTestEnvValue returns test mode appropriate values for environment variables.
func (ctx *NeuroContext) getTestEnvValue(key string) string {
	// Check for test-specific overrides first
	if value, exists := ctx.testEnvOverrides[key]; exists {
		return value
	}

	// Default test values
	switch key {
	case "OPENAI_API_KEY":
		return "test-openai-key"
	case "ANTHROPIC_API_KEY":
		return "test-anthropic-key"
	case "GOOGLE_API_KEY":
		return "test-google-key"
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
	// Handle invalid input
	if n < 1 {
		return "", fmt.Errorf("invalid message index %d: must be >= 1", n)
	}

	// Check if there's an active session
	if ctx.activeSessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	// Get the active session
	session, exists := ctx.chatSessions[ctx.activeSessionID]
	if !exists {
		return "", fmt.Errorf("active session %s not found", ctx.activeSessionID)
	}

	// Check if we have enough messages
	messageCount := len(session.Messages)
	if messageCount == 0 {
		return "", fmt.Errorf("session has no messages")
	}
	if n > messageCount {
		return "", fmt.Errorf("message index %d out of bounds: session has only %d messages", n, messageCount)
	}

	// Get the Nth most recent message (1-based indexing)
	// messages[len-1] is most recent, messages[len-2] is 2nd most recent, etc.
	messageIndex := messageCount - n
	return session.Messages[messageIndex].Content, nil
}

// GetNthChronologicalMessage returns the Nth message from the active session in chronological order.
// N=1 is the first message, N=2 is the second message, etc.
// Returns the message content and an error if the message cannot be retrieved.
func (ctx *NeuroContext) GetNthChronologicalMessage(n int) (string, error) {
	// Handle invalid input
	if n < 1 {
		return "", fmt.Errorf("invalid message index %d: must be >= 1", n)
	}

	// Check if there's an active session
	if ctx.activeSessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	// Get the active session
	session, exists := ctx.chatSessions[ctx.activeSessionID]
	if !exists {
		return "", fmt.Errorf("active session %s not found", ctx.activeSessionID)
	}

	// Check if we have enough messages
	messageCount := len(session.Messages)
	if messageCount == 0 {
		return "", fmt.Errorf("session has no messages")
	}
	if n > messageCount {
		return "", fmt.Errorf("message index %d out of bounds: session has only %d messages", n, messageCount)
	}

	// Get the Nth chronological message (1-based indexing)
	// messages[0] is first, messages[1] is second, etc.
	messageIndex := n - 1
	return session.Messages[messageIndex].Content, nil
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
	client, exists := ctx.llmClients[clientID]
	return client, exists
}

// SetLLMClient stores an LLM client by client ID (provider:hash format)
func (ctx *NeuroContext) SetLLMClient(clientID string, client neurotypes.LLMClient) {
	ctx.llmClients[clientID] = client
}

// GetLLMClientCount returns the number of cached LLM clients (for testing/debugging)
func (ctx *NeuroContext) GetLLMClientCount() int {
	return len(ctx.llmClients)
}

// GetAllLLMClients returns a copy of all cached LLM clients (for client lookup)
func (ctx *NeuroContext) GetAllLLMClients() map[string]neurotypes.LLMClient {
	clients := make(map[string]neurotypes.LLMClient)
	for id, client := range ctx.llmClients {
		clients[id] = client
	}
	return clients
}

// ClearLLMClients removes all cached LLM clients (for testing/debugging)
func (ctx *NeuroContext) ClearLLMClients() {
	ctx.llmClients = make(map[string]neurotypes.LLMClient)
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

// Stack operations for stack-based execution engine

// PushCommand adds a single command to the execution stack
func (ctx *NeuroContext) PushCommand(command string) {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()
	ctx.executionStack = append(ctx.executionStack, command)
}

// PushCommands adds multiple commands to the execution stack
func (ctx *NeuroContext) PushCommands(commands []string) {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()
	ctx.executionStack = append(ctx.executionStack, commands...)
}

// PopCommand removes and returns the last command from the stack (LIFO)
func (ctx *NeuroContext) PopCommand() (string, bool) {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	if len(ctx.executionStack) == 0 {
		return "", false
	}

	lastIndex := len(ctx.executionStack) - 1
	command := ctx.executionStack[lastIndex]
	ctx.executionStack = ctx.executionStack[:lastIndex]
	return command, true
}

// PeekCommand returns the next command without removing it from the stack
func (ctx *NeuroContext) PeekCommand() (string, bool) {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()

	if len(ctx.executionStack) == 0 {
		return "", false
	}

	return ctx.executionStack[len(ctx.executionStack)-1], true
}

// ClearStack removes all commands from the execution stack
func (ctx *NeuroContext) ClearStack() {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()
	ctx.executionStack = make([]string, 0)
}

// GetStackSize returns the number of commands in the execution stack
func (ctx *NeuroContext) GetStackSize() int {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return len(ctx.executionStack)
}

// IsStackEmpty returns true if the stack is empty
func (ctx *NeuroContext) IsStackEmpty() bool {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return len(ctx.executionStack) == 0
}

// PeekStack returns a copy of the execution stack without modifying it
// Returns the stack in reverse order (top to bottom, LIFO order)
func (ctx *NeuroContext) PeekStack() []string {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()

	result := make([]string, len(ctx.executionStack))
	// Copy in reverse order to show stack from top to bottom
	for i, cmd := range ctx.executionStack {
		result[len(ctx.executionStack)-1-i] = cmd
	}
	return result
}

// Try block support methods

// PushErrorBoundary pushes error boundary markers for try blocks
func (ctx *NeuroContext) PushErrorBoundary(tryID string) {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	// Create try block context
	tryBlock := TryBlockContext{
		ID:            tryID,
		StartDepth:    len(ctx.executionStack),
		ErrorCaptured: false,
	}

	ctx.tryBlocks = append(ctx.tryBlocks, tryBlock)
	ctx.currentTryDepth++
}

// PopErrorBoundary removes the most recent try block context
func (ctx *NeuroContext) PopErrorBoundary() {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	if len(ctx.tryBlocks) > 0 {
		ctx.tryBlocks = ctx.tryBlocks[:len(ctx.tryBlocks)-1]
		ctx.currentTryDepth--
	}
}

// IsInTryBlock returns true if currently inside a try block
func (ctx *NeuroContext) IsInTryBlock() bool {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return len(ctx.tryBlocks) > 0
}

// GetCurrentTryID returns the ID of the current try block
func (ctx *NeuroContext) GetCurrentTryID() string {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()

	if len(ctx.tryBlocks) == 0 {
		return ""
	}

	return ctx.tryBlocks[len(ctx.tryBlocks)-1].ID
}

// GetCurrentTryDepth returns the current try block depth
func (ctx *NeuroContext) GetCurrentTryDepth() int {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return ctx.currentTryDepth
}

// SetTryErrorCaptured marks the current try block as having captured an error
func (ctx *NeuroContext) SetTryErrorCaptured() {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	if len(ctx.tryBlocks) > 0 {
		ctx.tryBlocks[len(ctx.tryBlocks)-1].ErrorCaptured = true
	}
}

// IsTryErrorCaptured returns true if the current try block has captured an error
func (ctx *NeuroContext) IsTryErrorCaptured() bool {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()

	if len(ctx.tryBlocks) == 0 {
		return false
	}

	return ctx.tryBlocks[len(ctx.tryBlocks)-1].ErrorCaptured
}

// PushSilentBoundary pushes silent boundary markers for silent blocks
func (ctx *NeuroContext) PushSilentBoundary(silentID string) {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	// Create silent block context
	silentBlock := SilentBlockContext{
		ID:         silentID,
		StartDepth: len(ctx.executionStack),
	}

	ctx.silentBlocks = append(ctx.silentBlocks, silentBlock)
	ctx.currentSilentDepth++
}

// PopSilentBoundary removes the most recent silent block context
func (ctx *NeuroContext) PopSilentBoundary() {
	ctx.stackMutex.Lock()
	defer ctx.stackMutex.Unlock()

	if len(ctx.silentBlocks) > 0 {
		ctx.silentBlocks = ctx.silentBlocks[:len(ctx.silentBlocks)-1]
		ctx.currentSilentDepth--
	}
}

// IsInSilentBlock returns true if currently inside a silent block
func (ctx *NeuroContext) IsInSilentBlock() bool {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return len(ctx.silentBlocks) > 0
}

// GetCurrentSilentID returns the ID of the current silent block
func (ctx *NeuroContext) GetCurrentSilentID() string {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()

	if len(ctx.silentBlocks) == 0 {
		return ""
	}

	return ctx.silentBlocks[len(ctx.silentBlocks)-1].ID
}

// GetCurrentSilentDepth returns the current silent block depth
func (ctx *NeuroContext) GetCurrentSilentDepth() int {
	ctx.stackMutex.RLock()
	defer ctx.stackMutex.RUnlock()
	return ctx.currentSilentDepth
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
	ctx.testEnvOverrides[key] = value
}

// SetEnvVariable sets an environment variable, respecting test mode.
// In test mode, this sets a test environment override.
// In production mode, this sets an actual OS environment variable.
func (ctx *NeuroContext) SetEnvVariable(key, value string) error {
	if ctx.IsTestMode() {
		// In test mode, set test environment override
		ctx.SetTestEnvOverride(key, value)
		return nil
	}

	// In production mode, set actual OS environment variable
	return os.Setenv(key, value)
}

// GetEnvVariable retrieves an environment variable value, respecting test mode.
// This is a pure function that only gets the environment variable without side effects.
func (ctx *NeuroContext) GetEnvVariable(key string) string {
	return ctx.GetEnv(key)
}

// ClearTestEnvOverride removes a test-specific environment variable override.
func (ctx *NeuroContext) ClearTestEnvOverride(key string) {
	delete(ctx.testEnvOverrides, key)
}

// ClearAllTestEnvOverrides removes all test-specific environment variable overrides.
func (ctx *NeuroContext) ClearAllTestEnvOverrides() {
	ctx.testEnvOverrides = make(map[string]string)
}

// GetTestEnvOverrides returns a copy of all test environment variable overrides.
func (ctx *NeuroContext) GetTestEnvOverrides() map[string]string {
	overrides := make(map[string]string)
	for k, v := range ctx.testEnvOverrides {
		overrides[k] = v
	}
	return overrides
}

// File system operations for configuration service

// ReadFile reads the contents of a file, supporting test mode isolation.
func (ctx *NeuroContext) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file with the specified permissions, supporting test mode isolation.
func (ctx *NeuroContext) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// FileExists checks if a file exists at the given path.
func (ctx *NeuroContext) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetUserConfigDir returns the user's configuration directory.
// In test mode, returns a temporary directory to avoid polluting the user's system.
func (ctx *NeuroContext) GetUserConfigDir() (string, error) {
	if ctx.testMode {
		// In test mode, return a predictable test path
		return "/tmp/neuroshell-test-config", nil
	}

	// Get XDG config home or fall back to ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configHome, "neuroshell"), nil
}

// GetWorkingDir returns the current working directory.
func (ctx *NeuroContext) GetWorkingDir() (string, error) {
	if ctx.testMode {
		return "/tmp/neuroshell-test-workdir", nil
	}
	return os.Getwd()
}

// MkdirAll creates a directory path with the specified permissions, including any necessary parents.
func (ctx *NeuroContext) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Configuration management methods

// GetConfigMap returns a copy of the configuration map.
func (ctx *NeuroContext) GetConfigMap() map[string]string {
	ctx.configMutex.RLock()
	defer ctx.configMutex.RUnlock()

	result := make(map[string]string)
	for key, value := range ctx.configMap {
		result[key] = value
	}
	return result
}

// SetConfigMap replaces the entire configuration map.
func (ctx *NeuroContext) SetConfigMap(configMap map[string]string) {
	ctx.configMutex.Lock()
	defer ctx.configMutex.Unlock()

	ctx.configMap = make(map[string]string)
	for key, value := range configMap {
		ctx.configMap[key] = value
	}
}

// GetConfigValue retrieves a configuration value by key.
func (ctx *NeuroContext) GetConfigValue(key string) (string, bool) {
	ctx.configMutex.RLock()
	defer ctx.configMutex.RUnlock()

	value, exists := ctx.configMap[key]
	return value, exists
}

// SetConfigValue sets a configuration value.
func (ctx *NeuroContext) SetConfigValue(key, value string) {
	ctx.configMutex.Lock()
	defer ctx.configMutex.Unlock()

	ctx.configMap[key] = value
}

// Configuration loading methods (Context layer responsibilities)

// LoadDefaults sets up default configuration values.
func (ctx *NeuroContext) LoadDefaults() error {
	defaults := map[string]string{
		"NEURO_LOG_LEVEL": "info",
		"NEURO_TIMEOUT":   "30s",
	}

	for key, value := range defaults {
		ctx.SetConfigValue(key, value)
	}

	return nil
}

// LoadConfigDotEnv loads .env file from the user's config directory (~/.config/neuroshell/.env).
func (ctx *NeuroContext) LoadConfigDotEnv() error {
	configDir, err := ctx.GetUserConfigDir()
	if err != nil {
		// Config directory access failure is not fatal
		return nil
	}

	envPath := filepath.Join(configDir, ".env")
	if !ctx.FileExists(envPath) {
		// Missing config .env file is not an error
		return nil
	}

	return ctx.loadDotEnvFile(envPath)
}

// LoadLocalDotEnv loads .env file from the current working directory.
func (ctx *NeuroContext) LoadLocalDotEnv() error {
	workDir, err := ctx.GetWorkingDir()
	if err != nil {
		// Working directory access failure is not fatal
		return nil
	}

	envPath := filepath.Join(workDir, ".env")
	if !ctx.FileExists(envPath) {
		// Missing local .env file is not an error
		return nil
	}

	return ctx.loadDotEnvFile(envPath)
}

// LoadEnvironmentVariables loads specific prefixed environment variables into context configuration map.
// This has the highest priority and will override all file-based configuration.
func (ctx *NeuroContext) LoadEnvironmentVariables(prefixes []string) error {
	// In test mode, check test environment overrides first
	if ctx.IsTestMode() {
		testOverrides := ctx.GetTestEnvOverrides()
		for key, value := range testOverrides {
			for _, prefix := range prefixes {
				if strings.HasPrefix(key, prefix) {
					ctx.SetConfigValue(key, value)
					break
				}
			}
		}
	}

	// Then check actual OS environment variables
	environ := os.Environ()
	for _, env := range environ {
		for _, prefix := range prefixes {
			if strings.HasPrefix(env, prefix) {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					// Use context.GetEnv to respect test mode overrides
					value := ctx.GetEnv(key)

					// Store in configuration map (highest priority)
					ctx.SetConfigValue(key, value)
				}
				break // Found matching prefix, no need to check others
			}
		}
	}

	return nil
}

// loadDotEnvFile loads a specific .env file and stores all values in context configuration map.
// This is a private helper method used by LoadConfigDotEnv and LoadLocalDotEnv.
func (ctx *NeuroContext) loadDotEnvFile(envPath string) error {
	data, err := ctx.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file %s: %w", envPath, err)
	}

	// Parse .env file
	envMap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse .env file %s: %w", envPath, err)
	}

	// Store all values in context configuration map
	for key, value := range envMap {
		ctx.SetConfigValue(key, value)
	}

	return nil
}

// LoadEnvironmentVariablesWithPrefix loads OS environment variables with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (ctx *NeuroContext) LoadEnvironmentVariablesWithPrefix(sourcePrefix string) error {
	// In test mode, only load test environment overrides for clean testing
	if ctx.IsTestMode() {
		testOverrides := ctx.GetTestEnvOverrides()
		for key, value := range testOverrides {
			prefixedKey := sourcePrefix + key
			ctx.SetConfigValue(prefixedKey, value)
		}
		return nil // Don't load OS environment variables in test mode
	}

	// Load actual OS environment variables with prefix (production mode only)
	environ := os.Environ()
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := ctx.GetEnv(key) // Respect test mode overrides
			prefixedKey := sourcePrefix + key
			ctx.SetConfigValue(prefixedKey, value)
		}
	}

	return nil
}

// LoadConfigDotEnvWithPrefix loads config .env file with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (ctx *NeuroContext) LoadConfigDotEnvWithPrefix(sourcePrefix string) error {
	configDir, err := ctx.GetUserConfigDir()
	if err != nil {
		return nil // Config directory access failure is not fatal
	}

	envPath := filepath.Join(configDir, ".env")
	if !ctx.FileExists(envPath) {
		return nil // Missing config .env file is not an error
	}

	return ctx.loadDotEnvFileWithPrefix(envPath, sourcePrefix)
}

// LoadLocalDotEnvWithPrefix loads local .env file with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (ctx *NeuroContext) LoadLocalDotEnvWithPrefix(sourcePrefix string) error {
	workDir, err := ctx.GetWorkingDir()
	if err != nil {
		return nil // Working directory access failure is not fatal
	}

	envPath := filepath.Join(workDir, ".env")
	if !ctx.FileExists(envPath) {
		return nil // Missing local .env file is not an error
	}

	return ctx.loadDotEnvFileWithPrefix(envPath, sourcePrefix)
}

// loadDotEnvFileWithPrefix loads a .env file and stores values with a source prefix.
func (ctx *NeuroContext) loadDotEnvFileWithPrefix(envPath, sourcePrefix string) error {
	data, err := ctx.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file %s: %w", envPath, err)
	}

	// Parse .env file
	envMap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse .env file %s: %w", envPath, err)
	}

	// Store all values with source prefix in context configuration map
	for key, value := range envMap {
		prefixedKey := sourcePrefix + key
		ctx.SetConfigValue(prefixedKey, value)
	}

	return nil
}

// Provider Registry Methods

// GetSupportedProviders returns the list of supported LLM provider names.
// This is the central source of truth for all provider-related functionality.
func (ctx *NeuroContext) GetSupportedProviders() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(ctx.supportedProviders))
	copy(result, ctx.supportedProviders)
	return result
}

// GetProviderEnvPrefixes returns the list of environment variable prefixes
// used for loading provider-specific configuration from the environment.
func (ctx *NeuroContext) GetProviderEnvPrefixes() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(ctx.providerEnvPrefixes))
	copy(result, ctx.providerEnvPrefixes)
	return result
}

// IsValidProvider checks if a given provider name is supported.
// Provider comparison is case-insensitive.
func (ctx *NeuroContext) IsValidProvider(provider string) bool {
	providerLower := strings.ToLower(provider)
	for _, supportedProvider := range ctx.supportedProviders {
		if providerLower == supportedProvider {
			return true
		}
	}
	return false
}

// Error state management methods

// ResetErrorState resets the current error state to success (0/"") and moves current to last.
// This should be called before executing a new command.
func (ctx *NeuroContext) ResetErrorState() {
	ctx.errorStateMutex.Lock()
	defer ctx.errorStateMutex.Unlock()

	// Move current error state to last
	ctx.lastStatus = ctx.currentStatus
	ctx.lastError = ctx.currentError

	// Reset current state to success
	ctx.currentStatus = "0"
	ctx.currentError = ""
}

// SetErrorState sets the current error state based on command execution results.
// This should be called after command execution with the results.
func (ctx *NeuroContext) SetErrorState(status string, errorMsg string) {
	ctx.errorStateMutex.Lock()
	defer ctx.errorStateMutex.Unlock()

	ctx.currentStatus = status
	ctx.currentError = errorMsg
}

// GetCurrentErrorState returns the current error state (thread-safe read).
func (ctx *NeuroContext) GetCurrentErrorState() (status string, errorMsg string) {
	ctx.errorStateMutex.RLock()
	defer ctx.errorStateMutex.RUnlock()

	return ctx.currentStatus, ctx.currentError
}

// GetLastErrorState returns the last error state (thread-safe read).
func (ctx *NeuroContext) GetLastErrorState() (status string, errorMsg string) {
	ctx.errorStateMutex.RLock()
	defer ctx.errorStateMutex.RUnlock()

	return ctx.lastStatus, ctx.lastError
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
