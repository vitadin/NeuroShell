package shellintegration

import (
	"strings"
	"testing"
)

func TestStreamParser_ParseOutput(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string // Multiple inputs to simulate streaming
		expected ParseResult
	}{
		{
			name:   "Simple command output without OSC",
			inputs: []string{"hello world\n"},
			expected: ParseResult{
				Output:       "hello world\n",
				Sequences:    []OSCSequence{},
				State:        StateIdle,
				IsComplete:   false,
				ExitCode:     0,
				HasNewOutput: true,
			},
		},
		{
			name:   "Command with prompt start",
			inputs: []string{"\033]133;A\007"},
			expected: ParseResult{
				Output:       "",
				Sequences:    []OSCSequence{{Type: "133;A", Raw: "\033]133;A\007"}},
				State:        StatePromptStart,
				IsComplete:   false,
				ExitCode:     0,
				HasNewOutput: false,
			},
		},
		{
			name:   "Command completion sequence",
			inputs: []string{"\033]133;D;0\007"},
			expected: ParseResult{
				Output:       "",
				Sequences:    []OSCSequence{{Type: "133;D;0", ExitCode: 0, Raw: "\033]133;D;0\007"}},
				State:        StateCommandEnd,
				IsComplete:   true,
				ExitCode:     0,
				HasNewOutput: false,
			},
		},
		{
			name:   "Mixed output and OSC sequences",
			inputs: []string{"Command output\n\033]133;D;0\007"},
			expected: ParseResult{
				Output:       "Command output\n",
				Sequences:    []OSCSequence{{Type: "133;D;0", ExitCode: 0, Raw: "\033]133;D;0\007"}},
				State:        StateCommandEnd,
				IsComplete:   true,
				ExitCode:     0,
				HasNewOutput: true,
			},
		},
		{
			name: "Multiple streaming inputs",
			inputs: []string{
				"\033]133;A\007",
				"$ ls\n",
				"\033]133;B\007",
				"\033]133;C\007",
				"file1.txt\nfile2.txt\n",
				"\033]133;D;0\007",
			},
			expected: ParseResult{
				Output:       "$ ls\nfile1.txt\nfile2.txt\n", // Should include command prompt
				Sequences:    []OSCSequence{{Type: "133;D;0", ExitCode: 0, Raw: "\033]133;D;0\007"}},
				State:        StateCommandEnd,
				IsComplete:   true,
				ExitCode:     0,
				HasNewOutput: false, // Last input was OSC sequence, not new output
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewStreamParser()
			var lastResult ParseResult

			// Process all inputs
			for _, input := range tt.inputs {
				lastResult = parser.ParseOutput([]byte(input))
			}

			// Check final result
			if lastResult.Output != tt.expected.Output {
				t.Errorf("ParseOutput() Output = %q, want %q", lastResult.Output, tt.expected.Output)
			}

			if lastResult.State != tt.expected.State {
				t.Errorf("ParseOutput() State = %v, want %v", lastResult.State, tt.expected.State)
			}

			if lastResult.IsComplete != tt.expected.IsComplete {
				t.Errorf("ParseOutput() IsComplete = %v, want %v", lastResult.IsComplete, tt.expected.IsComplete)
			}

			if lastResult.ExitCode != tt.expected.ExitCode {
				t.Errorf("ParseOutput() ExitCode = %v, want %v", lastResult.ExitCode, tt.expected.ExitCode)
			}

			if lastResult.HasNewOutput != tt.expected.HasNewOutput {
				t.Errorf("ParseOutput() HasNewOutput = %v, want %v", lastResult.HasNewOutput, tt.expected.HasNewOutput)
			}

			// Check that we have the expected number of sequences in final result
			if len(lastResult.Sequences) != len(tt.expected.Sequences) {
				t.Errorf("ParseOutput() got %d sequences, want %d", len(lastResult.Sequences), len(tt.expected.Sequences))
			}
		})
	}
}

func TestFilterOSCSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Text without OSC sequences",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Text with OSC sequence at start",
			input:    "\033]133;A\007hello world",
			expected: "hello world",
		},
		{
			name:     "Text with OSC sequence at end",
			input:    "hello world\033]133;D;0\007",
			expected: "hello world",
		},
		{
			name:     "Text with OSC sequence in middle",
			input:    "hello\033]133;C\007 world",
			expected: "hello world",
		},
		{
			name:     "Multiple OSC sequences",
			input:    "\033]133;A\007hello\033]133;B\007 world\033]133;D;0\007",
			expected: "hello world",
		},
		{
			name:     "OSC sequence with ST terminator",
			input:    "hello\033]133;A\033\\world",
			expected: "helloworld",
		},
		{
			name:     "Only OSC sequences",
			input:    "\033]133;A\007\033]133;B\007\033]133;D;0\007",
			expected: "",
		},
		{
			name:     "Mixed with other escape sequences (should preserve)",
			input:    "hello\033[31mred\033[0m\033]133;A\007world",
			expected: "hello\033[31mred\033[0mworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterOSCSequences(tt.input)
			if result != tt.expected {
				t.Errorf("FilterOSCSequences() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractOSCSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OSCSequence
	}{
		{
			name:     "No OSC sequences",
			input:    "hello world",
			expected: []OSCSequence{},
		},
		{
			name:  "Single OSC sequence",
			input: "\033]133;A\007",
			expected: []OSCSequence{
				{Type: "133;A", Raw: "\033]133;A\007"},
			},
		},
		{
			name:  "Multiple OSC sequences",
			input: "\033]133;A\007hello\033]133;D;0\007",
			expected: []OSCSequence{
				{Type: "133;A", Raw: "\033]133;A\007"},
				{Type: "133;D;0", ExitCode: 0, Raw: "\033]133;D;0\007"},
			},
		},
		{
			name:  "OSC sequence with ST terminator",
			input: "\033]133;B\033\\",
			expected: []OSCSequence{
				{Type: "133;B", Raw: "\033]133;B\033\\"},
			},
		},
		{
			name:  "Mixed terminators",
			input: "\033]133;A\007text\033]133;B\033\\more text",
			expected: []OSCSequence{
				{Type: "133;A", Raw: "\033]133;A\007"},
				{Type: "133;B", Raw: "\033]133;B\033\\"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractOSCSequences(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("ExtractOSCSequences() got %d sequences, want %d", len(result), len(tt.expected))
				return
			}

			for i, seq := range result {
				expected := tt.expected[i]
				if seq.Type != expected.Type {
					t.Errorf("ExtractOSCSequences() sequence %d Type = %v, want %v", i, seq.Type, expected.Type)
				}
				if seq.ExitCode != expected.ExitCode {
					t.Errorf("ExtractOSCSequences() sequence %d ExitCode = %v, want %v", i, seq.ExitCode, expected.ExitCode)
				}
				if seq.Raw != expected.Raw {
					t.Errorf("ExtractOSCSequences() sequence %d Raw = %v, want %v", i, seq.Raw, expected.Raw)
				}
			}
		})
	}
}

func TestLineByLineParser(t *testing.T) {
	input := "line1\n\033]133;A\007line2\nline3\033]133;D;0\007\nline4"
	reader := strings.NewReader(input)
	parser := NewLineByLineParser(reader)

	lines := []string{}
	results := []ParseResult{}

	for {
		line, result, err := parser.ReadLine()
		if err != nil {
			break
		}
		lines = append(lines, line)
		results = append(results, result)
	}

	expectedLines := []string{"line1", "\033]133;A\007line2", "line3\033]133;D;0\007", "line4"}

	if len(lines) != len(expectedLines) {
		t.Errorf("LineByLineParser got %d lines, want %d", len(lines), len(expectedLines))
		return
	}

	for i, line := range lines {
		if line != expectedLines[i] {
			t.Errorf("LineByLineParser line %d = %q, want %q", i, line, expectedLines[i])
		}
	}

	// Check that we detected OSC sequences in the appropriate results
	foundPromptStart := false
	foundCommandEnd := false

	for _, result := range results {
		for _, seq := range result.Sequences {
			if seq.Type == "133;A" {
				foundPromptStart = true
			}
			if seq.Type == "133;D;0" {
				foundCommandEnd = true
			}
		}
	}

	if !foundPromptStart {
		t.Error("LineByLineParser did not detect prompt start sequence")
	}

	if !foundCommandEnd {
		t.Error("LineByLineParser did not detect command end sequence")
	}
}

func TestStreamParser_Reset(t *testing.T) {
	parser := NewStreamParser()

	// Add some data
	parser.ParseOutput([]byte("test data\n"))
	parser.ParseOutput([]byte("\033]133;A\007"))

	// Verify data exists
	if parser.GetState() == StateIdle {
		t.Error("Expected parser to have non-idle state before reset")
	}

	// Reset and verify clean state
	parser.Reset()

	if parser.GetState() != StateIdle {
		t.Errorf("After reset, expected StateIdle, got %v", parser.GetState())
	}

	if len(parser.GetSequences()) != 0 {
		t.Errorf("After reset, expected 0 sequences, got %d", len(parser.GetSequences()))
	}

	if parser.GetRawOutput() != "" {
		t.Errorf("After reset, expected empty raw output, got %q", parser.GetRawOutput())
	}
}
