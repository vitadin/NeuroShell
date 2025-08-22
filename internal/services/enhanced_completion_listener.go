package services

import (
	"fmt"
	"os"
	"strings"
)

// EnhancedCompletionListener implements readline.Listener to provide automatic completion suggestions
// when the completion mode is set to "enhanced". It monitors input changes and displays suggestions
// without requiring the user to press TAB.
type EnhancedCompletionListener struct {
	autoCompleteService *AutoCompleteService
	lastSuggestions     []string
	displayActive       bool
}

// NewEnhancedCompletionListener creates a new enhanced completion listener.
func NewEnhancedCompletionListener(autoCompleteService *AutoCompleteService) *EnhancedCompletionListener {
	return &EnhancedCompletionListener{
		autoCompleteService: autoCompleteService,
		lastSuggestions:     make([]string, 0),
		displayActive:       false,
	}
}

// OnChange implements readline.Listener interface. It's called on every keystroke.
func (l *EnhancedCompletionListener) OnChange(line []rune, pos int, _ rune) (newLine []rune, newPos int, ok bool) {
	// Check if enhanced mode is active
	if !l.isEnhancedModeActive() {
		return line, pos, true
	}

	// Convert input to string for processing
	lineStr := string(line)

	// Only show suggestions for meaningful input (avoid showing on empty lines or single characters)
	if pos < 2 || strings.TrimSpace(lineStr) == "" {
		l.clearSuggestions()
		return line, pos, true
	}

	// Get completion suggestions using the existing autocomplete logic
	suggestions := l.getSuggestions(lineStr, pos)

	// Update display if suggestions changed
	if l.suggestionsChanged(suggestions) {
		l.updateSuggestionDisplay(suggestions, lineStr, pos)
		l.lastSuggestions = suggestions
	}

	// Return original line unchanged - we only display suggestions, don't modify input
	return line, pos, true
}

// isEnhancedModeActive checks if the completion mode is set to "enhanced"
func (l *EnhancedCompletionListener) isEnhancedModeActive() bool {
	if l.autoCompleteService == nil {
		return false
	}

	mode := l.autoCompleteService.getCompletionMode()
	return mode == "enhanced"
}

// getSuggestions gets completion suggestions for the current input
func (l *EnhancedCompletionListener) getSuggestions(line string, pos int) []string {
	if l.autoCompleteService == nil {
		return []string{}
	}

	// Find the word being completed
	wordStart := l.autoCompleteService.findWordStart(line, pos)
	wordEnd := pos
	if wordEnd > len(line) {
		wordEnd = len(line)
	}

	currentWord := ""
	if wordStart < wordEnd {
		currentWord = line[wordStart:wordEnd]
	}

	// Get completion items using existing logic
	items := l.autoCompleteService.getCompletionItems(line, pos, currentWord)

	// Convert to string array, limiting to reasonable number for display
	maxSuggestions := 5
	suggestions := make([]string, 0, maxSuggestions)

	for i, item := range items {
		if i >= maxSuggestions {
			break
		}
		if strings.HasPrefix(item.Text, currentWord) {
			suggestions = append(suggestions, item.Text)
		}
	}

	return suggestions
}

// suggestionsChanged checks if the suggestions have changed since last time
func (l *EnhancedCompletionListener) suggestionsChanged(newSuggestions []string) bool {
	if len(newSuggestions) != len(l.lastSuggestions) {
		return true
	}

	for i, suggestion := range newSuggestions {
		if i >= len(l.lastSuggestions) || suggestion != l.lastSuggestions[i] {
			return true
		}
	}

	return false
}

// updateSuggestionDisplay shows the suggestions to the user
func (l *EnhancedCompletionListener) updateSuggestionDisplay(suggestions []string, _ string, _ int) {
	if len(suggestions) == 0 {
		l.clearSuggestions()
		return
	}

	// For now, we'll use a simple approach: print suggestions to stderr so they appear above the prompt
	// This is a basic implementation - could be enhanced with better positioning and formatting
	l.displaySuggestions(suggestions)
}

// displaySuggestions shows the suggestions in a formatted way
func (l *EnhancedCompletionListener) displaySuggestions(suggestions []string) {
	if len(suggestions) == 0 {
		return
	}

	// Use a simple approach: just print suggestions above the current line
	// Save cursor position, move up, print suggestions, restore position
	fmt.Fprintf(os.Stderr, "\033[s")  // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[1A") // Move up one line
	fmt.Fprintf(os.Stderr, "\033[K")  // Clear line

	// Format suggestions nicely
	if len(suggestions) == 1 {
		fmt.Fprintf(os.Stderr, "💡 %s", suggestions[0])
	} else {
		fmt.Fprintf(os.Stderr, "💡 %s", strings.Join(suggestions, " • "))
		if len(suggestions) > 3 {
			fmt.Fprintf(os.Stderr, " (+%d more)", len(suggestions)-3)
		}
	}

	fmt.Fprintf(os.Stderr, "\033[u") // Restore cursor position
	l.displayActive = true
}

// clearSuggestions clears the suggestion display
func (l *EnhancedCompletionListener) clearSuggestions() {
	if !l.displayActive {
		return
	}

	// Clear the suggestions display
	fmt.Fprintf(os.Stderr, "\033[s")  // Save cursor position
	fmt.Fprintf(os.Stderr, "\033[1A") // Move up one line
	fmt.Fprintf(os.Stderr, "\033[K")  // Clear line
	fmt.Fprintf(os.Stderr, "\033[u")  // Restore cursor position

	l.displayActive = false
	l.lastSuggestions = make([]string, 0)
}
