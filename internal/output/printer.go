package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Printer is the main output handler that supports both plain and styled output.
// It uses dependency injection to optionally support styling while maintaining
// clean architecture with no service dependencies.
type Printer struct {
	styleProvider StyleProvider
	writer        io.Writer
	mode          Mode
	forcePlain    bool
	testMode      bool
	silent        bool
	prefix        string

	// Thread safety for concurrent output
	mu sync.Mutex
}

// NewPrinter creates a new Printer with the given options.
// By default, it writes to os.Stdout with automatic mode detection.
func NewPrinter(options ...Option) *Printer {
	p := &Printer{
		writer: os.Stdout,
		mode:   ModeAuto,
	}

	// Apply options
	for _, opt := range options {
		opt(p)
	}

	return p
}

// Print outputs text without any semantic styling.
// This is the basic replacement for fmt.Print.
func (p *Printer) Print(text string) {
	p.output(SemanticPlain, text, false)
}

// Printf outputs formatted text without any semantic styling.
// This is the basic replacement for fmt.Printf.
func (p *Printer) Printf(format string, args ...interface{}) {
	p.output(SemanticPlain, fmt.Sprintf(format, args...), false)
}

// Println outputs text with a newline without any semantic styling.
// This is the basic replacement for fmt.Println.
func (p *Printer) Println(text string) {
	p.output(SemanticPlain, text, true)
}

// Info outputs informational text with info styling.
func (p *Printer) Info(text string) {
	p.output(SemanticInfo, text, true)
}

// Success outputs success text with success styling (typically green).
func (p *Printer) Success(text string) {
	p.output(SemanticSuccess, text, true)
}

// Warning outputs warning text with warning styling (typically yellow).
func (p *Printer) Warning(text string) {
	p.output(SemanticWarning, text, true)
}

// Error outputs error text with error styling (typically red).
func (p *Printer) Error(text string) {
	p.output(SemanticError, text, true)
}

// Command outputs command text with command styling.
func (p *Printer) Command(text string) {
	p.output(SemanticCommand, text, false)
}

// Variable outputs variable text with variable styling.
func (p *Printer) Variable(text string) {
	p.output(SemanticVariable, text, false)
}

// Keyword outputs keyword text with keyword styling.
func (p *Printer) Keyword(text string) {
	p.output(SemanticKeyword, text, false)
}

// Highlight outputs text with highlight styling.
func (p *Printer) Highlight(text string) {
	p.output(SemanticHighlight, text, false)
}

// Bold outputs text with bold styling.
func (p *Printer) Bold(text string) {
	p.output(SemanticBold, text, false)
}

// Code outputs inline code text.
func (p *Printer) Code(text string) {
	p.output(SemanticCode, text, false)
}

// CodeBlock outputs a multi-line code block.
func (p *Printer) CodeBlock(text string) {
	p.output(SemanticCodeBlock, text, true)
}

// Comment outputs comment text (typically prefixed with %%).
func (p *Printer) Comment(text string) {
	p.output(SemanticComment, text, true)
}

// output is the core output method that handles all rendering logic.
func (p *Printer) output(semantic SemanticType, text string, addNewline bool) {
	if p.silent {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Prepare the text based on output mode
	var finalText string

	switch p.mode {
	case ModeJSON:
		finalText = p.renderJSON(semantic, text)
	case ModePlain, ModeAuto:
		finalText = p.renderText(semantic, text, addNewline)
	case ModeStyled:
		finalText = p.renderStyled(semantic, text, addNewline)
	}

	// Apply prefix if configured
	if p.prefix != "" {
		finalText = p.prefix + finalText
	}

	// Write to output
	_, _ = fmt.Fprint(p.writer, finalText) // Ignore write errors for output operations
}

// renderText renders text in plain or auto mode.
func (p *Printer) renderText(semantic SemanticType, text string, addNewline bool) string {
	var result string

	// Use styling if available and not forced to plain mode
	if !p.forcePlain && p.styleProvider != nil && p.styleProvider.IsAvailable() {
		style := p.styleProvider.GetStyle(string(semantic))
		result = style.Render(text)
	} else {
		// Fall back to plain text with semantic prefixes
		plainProvider := NewPlainStyleProvider()
		style := plainProvider.GetStyle(string(semantic))
		result = style.Render(text)
	}

	if addNewline && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// renderStyled renders text with forced styling.
func (p *Printer) renderStyled(semantic SemanticType, text string, addNewline bool) string {
	if p.styleProvider != nil && p.styleProvider.IsAvailable() {
		style := p.styleProvider.GetStyle(string(semantic))
		result := style.Render(text)
		if addNewline && !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		return result
	}

	// Fall back to plain if no styling available
	return p.renderText(semantic, text, addNewline)
}

// renderJSON renders output as structured JSON.
func (p *Printer) renderJSON(semantic SemanticType, text string) string {
	output := map[string]interface{}{
		"type":    semantic,
		"message": text,
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		// Fall back to plain text if JSON encoding fails
		return text + "\n"
	}

	return string(jsonBytes) + "\n"
}

// SetWriter changes the output writer. This is useful for testing or redirecting output.
func (p *Printer) SetWriter(writer io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer = writer
}

// SetMode changes the output mode.
func (p *Printer) SetMode(mode Mode) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mode = mode
}

// SetStyleProvider changes the style provider. Pass nil to disable styling.
func (p *Printer) SetStyleProvider(provider StyleProvider) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.styleProvider = provider
}

// IsStylable returns true if the printer can apply styles.
func (p *Printer) IsStylable() bool {
	return !p.forcePlain && p.styleProvider != nil && p.styleProvider.IsAvailable()
}

// String returns a string representation for debugging.
func (p *Printer) String() string {
	hasStyles := "no"
	if p.IsStylable() {
		hasStyles = "yes"
	}
	return fmt.Sprintf("Printer{mode: %v, styles: %s, writer: %T}", p.mode, hasStyles, p.writer)
}
