package services

import (
	"sort"
	"strings"

	"neuroshell/internal/commands"
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
func (a *AutoCompleteService) findWordStart(line string, pos int) int {
	// Start from cursor position and work backwards
	for i := pos - 1; i >= 0; i-- {
		char := line[i]
		// Stop at spaces, brackets, or special characters that typically separate words
		if char == ' ' || char == '[' || char == ']' || char == ',' || char == '=' {
			return i + 1
		}
	}
	return 0
}

// getCompletions analyzes the input context and returns appropriate completions.
func (a *AutoCompleteService) getCompletions(line string, pos int, currentWord string) []string {
	// Analyze the context to determine what kind of completion is needed

	// Check if we're completing a command name (starts with \)
	if strings.HasPrefix(currentWord, "\\") {
		return a.getCommandCompletions(currentWord)
	}

	// Check if we're completing a variable reference (${)
	if strings.Contains(currentWord, "${") {
		return a.getVariableCompletions(currentWord)
	}

	// Check if we're inside brackets for option completion
	if a.isInsideBrackets(line, pos) {
		return a.getOptionCompletions(line, pos, currentWord)
	}

	// Check if we're at the beginning of input (no \ prefix)
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

	// Get all registered commands
	registry := commands.GetGlobalRegistry()
	allCommands := registry.GetAll()

	var completions []string
	for _, cmd := range allCommands {
		cmdName := cmd.Name()
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

// getVariableCompletions returns completions for variable references.
func (a *AutoCompleteService) getVariableCompletions(prefix string) []string {
	// Extract the variable name being completed
	varStart := strings.LastIndex(prefix, "${")
	if varStart == -1 {
		return make([]string, 0)
	}

	varPrefix := prefix[varStart+2:] // Skip "${

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
			// Return the full variable reference with closing brace
			completions = append(completions, prefix[:varStart+2]+varName+"}")
		}
	}

	// Sort completions alphabetically
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

	// Get the command from registry
	registry := commands.GetGlobalRegistry()
	command, exists := registry.Get(commandName)
	if !exists {
		return make([]string, 0)
	}

	// Get the command's help info for options
	helpInfo := command.HelpInfo()

	var completions []string

	// Check if we're completing an option name (no = in current word)
	if !strings.Contains(currentWord, "=") {
		// Complete option names
		for _, option := range helpInfo.Options {
			optionName := option.Name
			if strings.HasPrefix(optionName, currentWord) {
				if option.Type == "bool" {
					// Boolean options don't need =
					completions = append(completions, optionName)
				} else {
					// Add = for value options
					completions = append(completions, optionName+"=")
				}
			}
		}
	} else {
		// We're completing an option value
		parts := strings.SplitN(currentWord, "=", 2)
		if len(parts) == 2 {
			optionName := parts[0]
			valuePrefix := parts[1]

			// Get context-specific completions for this option
			contextCompletions := a.getOptionValueCompletions(commandName, optionName, valuePrefix)
			for _, completion := range contextCompletions {
				completions = append(completions, optionName+"="+completion)
			}
		}
	}

	// Sort completions alphabetically
	sort.Strings(completions)
	return completions
}

// getOptionValueCompletions returns completions for specific option values.
func (a *AutoCompleteService) getOptionValueCompletions(commandName, optionName, valuePrefix string) []string {
	var completions []string

	// Context-specific completions based on command and option
	switch commandName {
	case "session":
		if optionName == "name" {
			// Complete with existing session names
			completions = a.getSessionNameCompletions(valuePrefix)
		}
	case "model":
		if optionName == "name" {
			// Complete with available model names
			completions = a.getModelNameCompletions(valuePrefix)
		}
	case "render":
		if optionName == "style" {
			// Complete with available theme names
			completions = a.getThemeNameCompletions(valuePrefix)
		}
	case "set":
		if optionName == "name" {
			// Complete with existing variable names
			completions = a.getVariableNameCompletions(valuePrefix)
		}
	}

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

// getSessionNameCompletions returns completions for session names.
func (a *AutoCompleteService) getSessionNameCompletions(prefix string) []string {
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return make([]string, 0)
	}

	_, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return make([]string, 0)
	}

	sessionNameToID := globalCtx.GetSessionNameToID()

	var completions []string
	for sessionName := range sessionNameToID {
		if strings.HasPrefix(sessionName, prefix) {
			completions = append(completions, sessionName)
		}
	}

	sort.Strings(completions)
	return completions
}

// getModelNameCompletions returns completions for model names.
func (a *AutoCompleteService) getModelNameCompletions(prefix string) []string {
	globalCtx := context.GetGlobalContext()
	if globalCtx == nil {
		return make([]string, 0)
	}

	_, ok := globalCtx.(*context.NeuroContext)
	if !ok {
		return make([]string, 0)
	}

	modelNameToID := globalCtx.GetModelNameToID()

	var completions []string
	for modelName := range modelNameToID {
		if strings.HasPrefix(modelName, prefix) {
			completions = append(completions, modelName)
		}
	}

	sort.Strings(completions)
	return completions
}

// getThemeNameCompletions returns completions for theme names.
func (a *AutoCompleteService) getThemeNameCompletions(prefix string) []string {
	themeService, err := GetGlobalThemeService()
	if err != nil {
		return make([]string, 0)
	}

	availableThemes := themeService.GetAvailableThemes()

	var completions []string
	for _, themeName := range availableThemes {
		if strings.HasPrefix(themeName, prefix) {
			completions = append(completions, themeName)
		}
	}

	sort.Strings(completions)
	return completions
}

// getVariableNameCompletions returns completions for variable names (without ${}).
func (a *AutoCompleteService) getVariableNameCompletions(prefix string) []string {
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
		if strings.HasPrefix(varName, prefix) {
			completions = append(completions, varName)
		}
	}

	sort.Strings(completions)
	return completions
}

func init() {
	// Register the AutoCompleteService with the global registry
	if err := GlobalRegistry.RegisterService(NewAutoCompleteService()); err != nil {
		panic(err)
	}
}
