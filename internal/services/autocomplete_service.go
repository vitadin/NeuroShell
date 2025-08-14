package services

import (
	"sort"
	"strings"

	"neuroshell/internal/context"
)

// AutoCompleteService provides intelligent tab completion for NeuroShell commands and syntax.
// It implements the readline.AutoCompleter interface to integrate with ishell.
type AutoCompleteService struct {
	initialized bool
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
	a.initialized = true
	return nil
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

	// Get completions based on context
	completions := a.getCompletions(lineStr, pos, currentWord)

	// Convert completions to readline format
	var suggestions [][]rune
	for _, completion := range completions {
		if strings.HasPrefix(completion, currentWord) {
			// Return the part that should be added to complete the word
			suffix := strings.TrimPrefix(completion, currentWord)
			suggestions = append(suggestions, []rune(suffix))
		}
	}

	return suggestions, len(currentWord)
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

// getCompletions analyzes the input context and returns appropriate completions.
func (a *AutoCompleteService) getCompletions(line string, pos int, currentWord string) []string {
	// Analyze the context to determine what kind of completion is needed

	// Priority 1: Check if we're in a variable reference context
	if a.isInVariableReference(line, pos) || strings.HasPrefix(currentWord, "${") {
		return a.getVariableCompletions(currentWord)
	}

	// Priority 2: Check if we're after \help command for command name completion
	if a.isAfterHelpCommand(line, pos) {
		return a.getHelpCommandCompletions(currentWord)
	}

	// Priority 3: Check if we're completing a command name (starts with \)
	if strings.HasPrefix(currentWord, "\\") {
		return a.getCommandCompletions(currentWord)
	}

	// Priority 4: Check if we're inside brackets for option completion
	if a.isInsideBrackets(line, pos) {
		return a.getOptionCompletions(line, pos, currentWord)
	}

	// Priority 5: Check if we're at the beginning of input (no \ prefix)
	if pos == 0 || (pos > 0 && line[0] != '\\') {
		// Complete with common commands
		return a.getCommandCompletions("\\" + currentWord)
	}

	return make([]string, 0)
}

// getCommandCompletions returns completions for command names.
func (a *AutoCompleteService) getCommandCompletions(prefix string) []string {
	// Remove the \ prefix for matching
	commandPrefix := strings.TrimPrefix(prefix, "\\")

	// Get all registered commands from global context
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return []string{}
	}

	neuroCtx, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return []string{}
	}

	commandList := neuroCtx.GetRegisteredCommands()

	var completions []string
	for _, cmdName := range commandList {
		if strings.HasPrefix(cmdName, commandPrefix) {
			completions = append(completions, "\\"+cmdName)
		}
	}

	// Sort completions alphabetically
	sort.Strings(completions)

	// Ensure we return an empty slice instead of nil
	if completions == nil {
		return make([]string, 0)
	}
	return completions
}

// getHelpCommandCompletions returns command name completions for use after \help.
// Unlike getCommandCompletions, this returns command names without the backslash prefix.
func (a *AutoCompleteService) getHelpCommandCompletions(prefix string) []string {
	// Get all registered commands from global context
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return []string{}
	}

	neuroCtx, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return []string{}
	}

	commandList := neuroCtx.GetRegisteredCommands()

	var completions []string
	for _, cmdName := range commandList {
		if strings.HasPrefix(cmdName, prefix) {
			completions = append(completions, cmdName) // No backslash prefix for help context
		}
	}

	// Sort completions alphabetically
	sort.Strings(completions)

	// Ensure we return an empty slice instead of nil
	if completions == nil {
		return make([]string, 0)
	}
	return completions
}

// getVariableCompletions returns completions for variable references.
func (a *AutoCompleteService) getVariableCompletions(prefix string) []string {
	// Extract the variable name being completed
	varStart := strings.LastIndex(prefix, "${")
	if varStart == -1 {
		return make([]string, 0)
	}

	// Extract the partial variable name after ${
	varPrefix := prefix[varStart+2:] // Skip "${

	// Remove any closing brace if present (for cases like "${abc}" where user is editing)
	if closeBraceIdx := strings.Index(varPrefix, "}"); closeBraceIdx != -1 {
		varPrefix = varPrefix[:closeBraceIdx]
	}

	// Get all variables from context
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return make([]string, 0)
	}

	_, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return make([]string, 0)
	}

	// Get variable service to access all variables
	variableService, err := GetGlobalVariableService()
	if err != nil {
		return make([]string, 0)
	}

	allVars, err := variableService.GetAllVariables()
	if err != nil {
		return make([]string, 0)
	}

	var completions []string
	for varName := range allVars {
		if strings.HasPrefix(varName, varPrefix) {
			// Build the complete variable reference with proper formatting
			completion := prefix[:varStart] + "${" + varName + "}"
			completions = append(completions, completion)
		}
	}

	// Sort completions alphabetically for consistent ordering
	sort.Strings(completions)
	return completions
}

// getOptionCompletions returns completions for command options inside brackets.
func (a *AutoCompleteService) getOptionCompletions(line string, _ int, currentWord string) []string {
	// Parse the command name from incomplete input
	commandName := a.extractCommandNameFromLine(line)
	if commandName == "" {
		return make([]string, 0)
	}

	// Check if the command exists in the global context
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return make([]string, 0)
	}

	neuroCtx, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return make([]string, 0)
	}

	// Get command help info from context
	commandHelpInfo, exists := neuroCtx.GetCommandHelpInfo(commandName)
	if !exists {
		return make([]string, 0)
	}

	// Get completions based on command options
	var completions []string
	for _, option := range commandHelpInfo.Options {
		optionName := option.Name

		// Check if this option matches the current word prefix
		if strings.HasPrefix(optionName, currentWord) {
			// For boolean options, add just the name
			if option.Type == "bool" {
				completions = append(completions, optionName)
			} else {
				// For other types, add name with = suffix
				completions = append(completions, optionName+"=")
			}
		}
	}

	sort.Strings(completions)
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
