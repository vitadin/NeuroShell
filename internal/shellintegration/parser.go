// Package shellintegration provides escape sequence parsing for PTY output
package shellintegration

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// StreamParser parses PTY output stream for OSC sequences while preserving output
type StreamParser struct {
	buffer       bytes.Buffer
	sequences    []OSCSequence
	rawOutput    bytes.Buffer
	cleanOutput  bytes.Buffer // Accumulated clean output
	state        CommandState
	lastExitCode int
}

// NewStreamParser creates a new stream parser
func NewStreamParser() *StreamParser {
	return &StreamParser{
		state: StateIdle,
	}
}

// ParseOutput parses output from PTY and extracts OSC sequences
func (p *StreamParser) ParseOutput(data []byte) ParseResult {
	p.buffer.Write(data)
	p.rawOutput.Write(data)

	return p.processBuffer()
}

// ParseResult contains the result of parsing PTY output
type ParseResult struct {
	Output       string        // Accumulated clean output without OSC sequences
	NewOutput    string        // Just the new clean output from this parse call
	Sequences    []OSCSequence // Detected OSC sequences
	State        CommandState  // Current command state
	IsComplete   bool          // Whether command is complete
	ExitCode     int           // Exit code if command is complete
	HasNewOutput bool          // Whether there's new output to display
}

// processBuffer processes the internal buffer for OSC sequences
func (p *StreamParser) processBuffer() ParseResult {
	content := p.buffer.String()
	var newCleanOutput strings.Builder
	var newSequences []OSCSequence
	hasNewOutput := false

	i := 0
	for i < len(content) {
		// Look for OSC sequence start
		if i+len(ESC+"]133") <= len(content) && content[i:i+len(ESC+"]133")] == (ESC+"]133") {
			// Find the end of the sequence (BEL or ST)
			seqEnd := -1

			// Look for BEL terminator
			if belPos := strings.Index(content[i:], BEL); belPos != -1 {
				seqEnd = i + belPos + len(BEL)
			} else if stPos := strings.Index(content[i:], ST); stPos != -1 {
				// Look for ST terminator
				seqEnd = i + stPos + len(ST)
			}

			if seqEnd != -1 {
				// Extract and parse the sequence
				seqText := content[i:seqEnd]
				if seq, ok := ParseOSCSequence(seqText); ok {
					newSequences = append(newSequences, *seq)
					p.updateState(seq)
				}

				// Skip past the sequence
				i = seqEnd
				continue
			}
		}

		// Regular character - add to clean output
		newCleanOutput.WriteByte(content[i])
		hasNewOutput = true
		i++
	}

	// Add new clean output to accumulated output
	newCleanText := newCleanOutput.String()
	if newCleanText != "" {
		p.cleanOutput.WriteString(newCleanText)
	}

	// Clear the buffer after processing
	p.buffer.Reset()

	// Update sequences list
	p.sequences = append(p.sequences, newSequences...)

	// Check if command is complete (either command end or prompt start indicates completion)
	isComplete := p.state == StateCommandEnd || p.state == StatePromptStart

	return ParseResult{
		Output:       p.cleanOutput.String(), // Return accumulated clean output
		NewOutput:    newCleanText,           // Return only new clean output from this call
		Sequences:    newSequences,           // Return only new sequences
		State:        p.state,
		IsComplete:   isComplete,
		ExitCode:     p.lastExitCode,
		HasNewOutput: hasNewOutput,
	}
}

// updateState updates the parser state based on OSC sequence
func (p *StreamParser) updateState(seq *OSCSequence) {
	newState := GetCommandState(seq.Type)

	if newState == StateCommandEnd {
		p.lastExitCode = seq.ExitCode
	}

	p.state = newState
}

// Reset resets the parser state
func (p *StreamParser) Reset() {
	p.buffer.Reset()
	p.rawOutput.Reset()
	p.cleanOutput.Reset()
	p.sequences = nil
	p.state = StateIdle
	p.lastExitCode = 0
}

// GetRawOutput returns the raw output including OSC sequences
func (p *StreamParser) GetRawOutput() string {
	return p.rawOutput.String()
}

// GetSequences returns all detected OSC sequences
func (p *StreamParser) GetSequences() []OSCSequence {
	return p.sequences
}

// GetState returns the current command state
func (p *StreamParser) GetState() CommandState {
	return p.state
}

// LineByLineParser provides line-by-line parsing with OSC sequence detection
type LineByLineParser struct {
	scanner *bufio.Scanner
	parser  *StreamParser
}

// NewLineByLineParser creates a new line-by-line parser
func NewLineByLineParser(reader io.Reader) *LineByLineParser {
	return &LineByLineParser{
		scanner: bufio.NewScanner(reader),
		parser:  NewStreamParser(),
	}
}

// ReadLine reads the next line and processes it for OSC sequences
func (p *LineByLineParser) ReadLine() (string, ParseResult, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", ParseResult{}, err
		}
		return "", ParseResult{}, io.EOF
	}

	line := p.scanner.Text()
	result := p.parser.ParseOutput([]byte(line + "\n"))

	return line, result, nil
}

// Reset resets the line-by-line parser
func (p *LineByLineParser) Reset() {
	p.parser.Reset()
}

// FilterOSCSequences removes OSC sequences from text while preserving other content
func FilterOSCSequences(text string) string {
	var result strings.Builder

	i := 0
	for i < len(text) {
		// Look for OSC sequence start
		if i+len(ESC+"]133") <= len(text) && text[i:i+len(ESC+"]133")] == (ESC+"]133") {
			// Find the end of the sequence
			seqEnd := -1

			if belPos := strings.Index(text[i:], BEL); belPos != -1 {
				seqEnd = i + belPos + len(BEL)
			} else if stPos := strings.Index(text[i:], ST); stPos != -1 {
				seqEnd = i + stPos + len(ST)
			}

			if seqEnd != -1 {
				// Skip the OSC sequence
				i = seqEnd
				continue
			}
		}

		// Regular character
		result.WriteByte(text[i])
		i++
	}

	return result.String()
}

// ExtractOSCSequences extracts all OSC sequences from text
func ExtractOSCSequences(text string) []OSCSequence {
	var sequences []OSCSequence

	i := 0
	for i < len(text) {
		if i+len(ESC+"]133") <= len(text) && text[i:i+len(ESC+"]133")] == (ESC+"]133") {
			seqEnd := -1

			if belPos := strings.Index(text[i:], BEL); belPos != -1 {
				seqEnd = i + belPos + len(BEL)
			} else if stPos := strings.Index(text[i:], ST); stPos != -1 {
				seqEnd = i + stPos + len(ST)
			}

			if seqEnd != -1 {
				seqText := text[i:seqEnd]
				if seq, ok := ParseOSCSequence(seqText); ok {
					sequences = append(sequences, *seq)
				}
				i = seqEnd
				continue
			}
		}

		i++
	}

	return sequences
}
