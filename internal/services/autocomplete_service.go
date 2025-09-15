package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
)

// CompletionItem represents a single completion suggestion with optional description.
type CompletionItem struct {
	Text        string // The completion text
	Description string // Optional description for enhanced modes
	Category    string // Category for grouping (commands, variables, files, etc.)
}

// AutoCompleteService provides intelligent tab completion for NeuroShell commands and syntax.
// It implements the readline.AutoCompleter interface to integrate with ishell.
type AutoCompleteService struct {
	initialized        bool
	commandRegistryCtx neuroshellcontext.CommandRegistrySubcontext
}

// NewAutoCompleteService creates a new AutoCompleteService instance.
func NewAutoCompleteService() *AutoCompleteService {
	return &AutoCompleteService{
		initialized: false,
	}
}

// Name returns the service name "autocomplete" for registration.
func (a *AutoCompleteService) Name() string {
	return "autocomplete"
}

// Initialize sets up the AutoCompleteService for operation.
func (a *AutoCompleteService) Initialize() error {
	ctx := neuroshellcontext.GetGlobalContext()
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("global context is not a NeuroContext")
	}
	a.commandRegistryCtx = neuroshellcontext.NewCommandRegistrySubcontextFromContext(neuroCtx)
	a.initialized = true
	return nil
}

// getCompletionMode retrieves the current completion mode from the global context.
func (a *AutoCompleteService) getCompletionMode() string {
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return "tab"
	}

	if mode, err := globalCtx.GetVariable("_completion_mode"); err == nil {
		return mode
	}
	return "tab"
}

// Do implements the readline.AutoCompleter interface.
// It analyzes the current input line and cursor position to provide relevant completions.
func (a *AutoCompleteService) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	if !a.initialized {
		return nil, 0
	}

	// Convert runes to string for easier processing
	lineStr := string(line)

	// Find the word being completed (from last space or start to cursor)
	wordStart := a.findWordStart(lineStr, pos)
	wordEnd := pos

	// Ensure wordEnd doesn't exceed string length
	if wordEnd > len(lineStr) {
		wordEnd = len(lineStr)
	}

	// Extract the word being completed
	currentWord := ""
	if wordStart < wordEnd {
		currentWord = lineStr[wordStart:wordEnd]
	}

	// Get completion mode
	completionMode := a.getCompletionMode()

	// Get completions based on context
	completionItems := a.getCompletionItems(lineStr, pos, currentWord)

	// Format completions based on mode
	suggestions := a.formatCompletions(completionItems, currentWord, completionMode)

	return suggestions, len(currentWord)
}

// formatCompletions converts CompletionItems to readline format based on the completion mode.
func (a *AutoCompleteService) formatCompletions(items []CompletionItem, currentWord, mode string) [][]rune {
	var suggestions [][]rune

	switch mode {
	case "enhanced":
		// Enhanced mode: provides smart context-aware completions (files, sessions, models)
		// Currently uses same display format as basic mode due to readline limitations
		suggestions = a.formatEnhancedCompletions(items, currentWord)
	case "tab":
		fallthrough
	default:
		// Traditional TAB mode: basic completion logic and display
		suggestions = a.formatBasicCompletions(items, currentWord)
	}

	return suggestions
}

// formatBasicCompletions formats completions for traditional TAB mode.
func (a *AutoCompleteService) formatBasicCompletions(items []CompletionItem, currentWord string) [][]rune {
	var suggestions [][]rune
	for _, item := range items {
		if strings.HasPrefix(item.Text, currentWord) {
			// Return the part that should be added to complete the word
			suffix := strings.TrimPrefix(item.Text, currentWord)
			suggestions = append(suggestions, []rune(suffix))
		}
	}
	return suggestions
}

// formatEnhancedCompletions formats completions for enhanced mode.
// Currently simplified to avoid readline display issues with descriptions and category headers.
// Future enhancement: integrate with a custom display system for richer completion information.
func (a *AutoCompleteService) formatEnhancedCompletions(items []CompletionItem, currentWord string) [][]rune {
	// For now, enhanced mode provides the same clean completions as basic mode
	// The main benefit is the enhanced completion logic (smart file/session/model completions)
	// rather than display formatting which has readline library limitations
	return a.formatBasicCompletions(items, currentWord)
}

// findWordStart finds the start position of the word being completed.
// It handles special cases like variable references (${var}) to provide better completions.
func (a *AutoCompleteService) findWordStart(line string, pos int) int {
	// Check if we're in a variable reference context
	if pos >= 2 && a.isInVariableReference(line, pos) {
		// Find the start of the variable reference (${)
		for i := pos - 1; i >= 1; i-- {
			if i >= 1 && line[i-1] == '$' && line[i] == '{' {
				return i - 1 // Return position of '$'
			}
		}
	}

	// Standard word boundary detection
	// Start from cursor position and work backwards
	for i := pos - 1; i >= 0; i-- {
		// Safety check to prevent index out of bounds
		if i >= len(line) {
			continue
		}
		char := line[i]
		// Stop at spaces, brackets, or special characters that typically separate words
		if char == ' ' || char == '[' || char == ']' || char == ',' || char == '=' {
			return i + 1
		}
	}
	return 0
}

// isInVariableReference checks if the cursor position is within a variable reference (${...})
func (a *AutoCompleteService) isInVariableReference(line string, pos int) bool {
	// Look backwards to find the most recent ${ and }
	lastDollarBrace := -1
	lastCloseBrace := -1

	for i := 0; i < pos && i < len(line)-1; i++ {
		if line[i] == '$' && line[i+1] == '{' {
			lastDollarBrace = i
		} else if line[i] == '}' {
			lastCloseBrace = i
		}
	}

	// We're in a variable reference if the last ${ is after the last }
	return lastDollarBrace > lastCloseBrace
}

// getCompletionItems analyzes the input context and returns appropriate completion items.
func (a *AutoCompleteService) getCompletionItems(line string, pos int, currentWord string) []CompletionItem {
	// Analyze the context to determine what kind of completion is needed

	// Priority 1: Check if we're in a variable reference context
	if a.isInVariableReference(line, pos) || strings.HasPrefix(currentWord, "${") {
		return a.getVariableCompletionItems(currentWord)
	}

	// Priority 2: Check if we're after \help command for command name completion
	if a.isAfterHelpCommand(line, pos) {
		return a.getHelpCommandCompletionItems(currentWord)
	}

	// Priority 3: Check if we're completing a command name (starts with \)
	if strings.HasPrefix(currentWord, "\\") {
		return a.getCommandCompletionItems(currentWord)
	}

	// Priority 4: Check if we're inside brackets for option completion
	if a.isInsideBrackets(line, pos) {
		return a.getOptionCompletionItems(line, pos, currentWord)
	}

	// Priority 5: Check for command-specific smart completions
	commandName := a.extractCommandNameFromLine(line)
	if commandName != "" {
		if smartCompletions := a.getSmartCompletions(commandName, line, pos, currentWord); len(smartCompletions) > 0 {
			return smartCompletions
		}
	}

	// Priority 6: Check if we're at the beginning of input (no \ prefix)
	if pos == 0 || (pos > 0 && line[0] != '\\') {
		// Complete with common commands
		return a.getCommandCompletionItems("\\" + currentWord)
	}

	return make([]CompletionItem, 0)
}

// getSmartCompletions provides context-aware completions for specific commands.
func (a *AutoCompleteService) getSmartCompletions(commandName, _ string, _ int, currentWord string) []CompletionItem {
	switch commandName {
	case "cat", "write":
		// File path completion for file-related commands
		return a.getFilePathCompletions(currentWord)
	case "session-activate", "session-delete", "session-show", "session-copy":
		// Session name completion for session commands
		return a.getSessionNameCompletions(currentWord)
	case "model-activate", "model-delete":
		// Model name completion for model commands
		return a.getModelNameCompletions(currentWord)
	}
	return []CompletionItem{}
}

// getFilePathCompletions returns file and directory completions.
func (a *AutoCompleteService) getFilePathCompletions(currentWord string) []CompletionItem {
	var completions []CompletionItem

	// Handle absolute vs relative paths
	var basePath string
	var searchPattern string

	if strings.HasPrefix(currentWord, "/") {
		// Absolute path
		basePath = filepath.Dir(currentWord)
		searchPattern = filepath.Base(currentWord)
		if basePath == "." {
			basePath = "/"
		}
	} else {
		// Relative path - use current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return completions
		}

		if strings.Contains(currentWord, "/") {
			basePath = filepath.Join(cwd, filepath.Dir(currentWord))
			searchPattern = filepath.Base(currentWord)
		} else {
			basePath = cwd
			searchPattern = currentWord
		}
	}

	// Read directory contents
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return completions
	}

	// Filter entries that match the pattern
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless explicitly requested
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(searchPattern, ".") {
			continue
		}

		if strings.HasPrefix(name, searchPattern) {
			var completion string
			switch {
			case strings.HasPrefix(currentWord, "/"):
				completion = filepath.Join(basePath, name)
			case strings.Contains(currentWord, "/"):
				completion = filepath.Join(filepath.Dir(currentWord), name)
			default:
				completion = name
			}

			// Add trailing slash for directories
			description := "file"
			if entry.IsDir() {
				completion += "/"
				description = "directory"
			}

			completions = append(completions, CompletionItem{
				Text:        completion,
				Description: description,
				Category:    "files",
			})
		}
	}

	// Sort by name
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	return completions
}

// getSessionNameCompletions returns available session names for completion.
func (a *AutoCompleteService) getSessionNameCompletions(currentWord string) []CompletionItem {
	var completions []CompletionItem

	// Get chat session service
	sessionService, err := GetGlobalRegistry().GetService("chatsession")
	if err != nil {
		return completions
	}

	chatSession, ok := sessionService.(*ChatSessionService)
	if !ok {
		return completions
	}

	// Get list of available sessions
	sessions := chatSession.ListSessions()

	// Filter sessions that match the current word
	for _, session := range sessions {
		if strings.HasPrefix(session.Name, currentWord) {
			completions = append(completions, CompletionItem{
				Text:        session.Name,
				Description: "chat session",
				Category:    "sessions",
			})
		}
	}

	// Sort alphabetically
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	return completions
}

// getModelNameCompletions returns available model names for completion.
func (a *AutoCompleteService) getModelNameCompletions(currentWord string) []CompletionItem {
	var completions []CompletionItem

	// Get model service
	modelService, err := GetGlobalRegistry().GetService("model")
	if err != nil {
		return completions
	}

	modelSvc, ok := modelService.(*ModelService)
	if !ok {
		return completions
	}

	// Get list of available models
	models, err := modelSvc.ListModelsWithGlobalContext()
	if err != nil {
		return completions
	}

	// Filter models that match the current word
	for modelName, modelConfig := range models {
		if strings.HasPrefix(modelName, currentWord) {
			description := fmt.Sprintf("model: %s", modelConfig.CatalogID)
			completions = append(completions, CompletionItem{
				Text:        modelName,
				Description: description,
				Category:    "models",
			})
		}
	}

	// Sort alphabetically
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	return completions
}

// getCommandCompletionItems returns completion items for command names with descriptions.
func (a *AutoCompleteService) getCommandCompletionItems(prefix string) []CompletionItem {
	if !a.initialized {
		return []CompletionItem{}
	}

	// Remove the \ prefix for matching
	commandPrefix := strings.TrimPrefix(prefix, "\\")

	commandList := a.commandRegistryCtx.GetRegisteredCommands()

	var completions []CompletionItem
	for _, cmdName := range commandList {
		if strings.HasPrefix(cmdName, commandPrefix) {
			// Get command description from help info
			description := ""
			if helpInfo, exists := a.commandRegistryCtx.GetCommandHelpInfo(cmdName); exists {
				description = helpInfo.Description
			}

			completions = append(completions, CompletionItem{
				Text:        "\\" + cmdName,
				Description: description,
				Category:    "commands",
			})
		}
	}

	// Sort completions alphabetically by text
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	return completions
}

// getHelpCommandCompletionItems returns command name completion items for use after \help.
// Unlike getCommandCompletionItems, this returns command names without the backslash prefix.
func (a *AutoCompleteService) getHelpCommandCompletionItems(prefix string) []CompletionItem {
	if !a.initialized {
		return []CompletionItem{}
	}

	commandList := a.commandRegistryCtx.GetRegisteredCommands()

	var completions []CompletionItem
	for _, cmdName := range commandList {
		if strings.HasPrefix(cmdName, prefix) {
			// Get command description from help info
			description := ""
			if helpInfo, exists := a.commandRegistryCtx.GetCommandHelpInfo(cmdName); exists {
				description = helpInfo.Description
			}

			completions = append(completions, CompletionItem{
				Text:        cmdName, // No backslash prefix for help context
				Description: description,
				Category:    "commands",
			})
		}
	}

	// Sort completions alphabetically by text
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})

	return completions
}

// getVariableCompletionItems returns completion items for variable references.
func (a *AutoCompleteService) getVariableCompletionItems(prefix string) []CompletionItem {
	// Extract the variable name being completed
	varStart := strings.LastIndex(prefix, "${")
	if varStart == -1 {
		return make([]CompletionItem, 0)
	}

	// Extract the partial variable name after ${
	varPrefix := prefix[varStart+2:] // Skip "${

	// Remove any closing brace if present (for cases like "${abc}" where user is editing)
	if closeBraceIdx := strings.Index(varPrefix, "}"); closeBraceIdx != -1 {
		varPrefix = varPrefix[:closeBraceIdx]
	}

	// Get all variables from context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return make([]CompletionItem, 0)
	}

	_, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return make([]CompletionItem, 0)
	}

	// Get variable service to access all variables
	variableService, err := GetGlobalVariableService()
	if err != nil {
		return make([]CompletionItem, 0)
	}

	allVars, err := variableService.GetAllVariables()
	if err != nil {
		return make([]CompletionItem, 0)
	}

	var completions []CompletionItem
	for varName, varValue := range allVars {
		if strings.HasPrefix(varName, varPrefix) {
			// Build the complete variable reference with proper formatting
			completion := prefix[:varStart] + "${" + varName + "}"

			// Create a brief description based on variable type and value
			description := a.getVariableDescription(varName, varValue)

			completions = append(completions, CompletionItem{
				Text:        completion,
				Description: description,
				Category:    "variables",
			})
		}
	}

	// Sort completions alphabetically by variable name
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})
	return completions
}

// getVariableDescription creates a brief description for a variable based on its name and value.
func (a *AutoCompleteService) getVariableDescription(name, value string) string {
	// System variables (start with @)
	if strings.HasPrefix(name, "@") {
		return "system variable"
	}

	// Global configuration variables (start with _)
	if strings.HasPrefix(name, "_") {
		return "config variable"
	}

	// Command output variables (start with _)
	if strings.HasPrefix(name, "_output") || strings.HasPrefix(name, "_error") || strings.HasPrefix(name, "_status") {
		return "command output"
	}

	// Metadata variables (start with #)
	if strings.HasPrefix(name, "#") {
		return "metadata"
	}

	// Message history variables (numeric)
	if len(name) > 0 && name[0] >= '1' && name[0] <= '9' {
		return "message history"
	}

	// Truncate long values for description
	if len(value) > 30 {
		return fmt.Sprintf("%.30s...", value)
	} else if value != "" {
		return value
	}

	return "user variable"
}

// getOptionCompletionItems returns completion items for command options inside brackets.
func (a *AutoCompleteService) getOptionCompletionItems(line string, _ int, currentWord string) []CompletionItem {
	if !a.initialized {
		return make([]CompletionItem, 0)
	}

	// Parse the command name from incomplete input
	commandName := a.extractCommandNameFromLine(line)
	if commandName == "" {
		return make([]CompletionItem, 0)
	}

	// Get command help info from context
	commandHelpInfo, exists := a.commandRegistryCtx.GetCommandHelpInfo(commandName)
	if !exists {
		return make([]CompletionItem, 0)
	}

	// Get completions based on command options
	var completions []CompletionItem
	for _, option := range commandHelpInfo.Options {
		optionName := option.Name

		// Check if this option matches the current word prefix
		if strings.HasPrefix(optionName, currentWord) {
			var completion string
			var description string

			// For boolean options, add just the name
			if option.Type == "bool" {
				completion = optionName
				description = fmt.Sprintf("boolean option: %s", option.Description)
			} else {
				// For other types, add name with = suffix
				completion = optionName + "="
				description = fmt.Sprintf("%s option: %s", option.Type, option.Description)
			}

			completions = append(completions, CompletionItem{
				Text:        completion,
				Description: description,
				Category:    "options",
			})
		}
	}

	// Sort completions alphabetically by text
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Text < completions[j].Text
	})
	return completions
}

// extractCommandNameFromLine extracts the command name from a line, handling incomplete bracket input.
func (a *AutoCompleteService) extractCommandNameFromLine(line string) string {
	// Remove leading backslash if present
	line = strings.TrimPrefix(line, "\\")

	// Find the first bracket
	bracketIdx := strings.Index(line, "[")
	if bracketIdx == -1 {
		// No brackets, return the first word
		parts := strings.SplitN(line, " ", 2)
		return parts[0]
	}

	// Return the command name before the bracket
	return line[:bracketIdx]
}

// isInsideBrackets checks if the cursor position is inside command brackets.
func (a *AutoCompleteService) isInsideBrackets(line string, pos int) bool {
	// Look for the last [ before the cursor
	lastBracketOpen := -1
	lastBracketClose := -1

	for i := 0; i < pos && i < len(line); i++ {
		switch line[i] {
		case '[':
			lastBracketOpen = i
		case ']':
			lastBracketClose = i
		}
	}

	// We're inside brackets if the last [ is after the last ]
	return lastBracketOpen > lastBracketClose
}

// isAfterHelpCommand checks if the cursor is positioned after \help for command name completion.
// This handles both space syntax (\help command) and bracket syntax (\help[command]).
func (a *AutoCompleteService) isAfterHelpCommand(line string, pos int) bool {
	// Check for space syntax: \help followed by space(s)
	if strings.HasPrefix(line, "\\help ") {
		// Make sure cursor is after the space(s)
		spaceEnd := 6 // length of "\help "
		for spaceEnd < len(line) && line[spaceEnd] == ' ' {
			spaceEnd++
		}
		return pos >= spaceEnd
	}

	// Check for bracket syntax: \help[
	if strings.HasPrefix(line, "\\help[") && pos >= 6 {
		// Make sure we're inside the brackets or right after the opening bracket
		return a.isInsideBrackets(line, pos) || pos == 6
	}

	return false
}

func init() {
	// Register the AutoCompleteService with the global registry
	if err := GlobalRegistry.RegisterService(NewAutoCompleteService()); err != nil {
		panic(err)
	}
}
