package shellintegration

import (
	"strings"
	"testing"
)

func TestParseOSCSequence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *OSCSequence
		valid    bool
	}{
		{
			name:  "Prompt start sequence",
			input: "\033]133;A\007",
			expected: &OSCSequence{
				Type: "133;A",
				Raw:  "\033]133;A\007",
			},
			valid: true,
		},
		{
			name:  "Command start sequence",
			input: "\033]133;B\007",
			expected: &OSCSequence{
				Type: "133;B",
				Raw:  "\033]133;B\007",
			},
			valid: true,
		},
		{
			name:  "Output start sequence",
			input: "\033]133;C\007",
			expected: &OSCSequence{
				Type: "133;C",
				Raw:  "\033]133;C\007",
			},
			valid: true,
		},
		{
			name:  "Command end with exit code 0",
			input: "\033]133;D;0\007",
			expected: &OSCSequence{
				Type:     "133;D;0",
				ExitCode: 0,
				Raw:      "\033]133;D;0\007",
			},
			valid: true,
		},
		{
			name:  "Command end with exit code 1",
			input: "\033]133;D;1\007",
			expected: &OSCSequence{
				Type:     "133;D;1",
				ExitCode: 1,
				Raw:      "\033]133;D;1\007",
			},
			valid: true,
		},
		{
			name:  "Sequence with ST terminator",
			input: "\033]133;A\033\\",
			expected: &OSCSequence{
				Type: "133;A",
				Raw:  "\033]133;A\033\\",
			},
			valid: true,
		},
		{
			name:     "Invalid sequence - not OSC 133",
			input:    "\033]134;A\007",
			expected: nil,
			valid:    false,
		},
		{
			name:     "Invalid sequence - no terminator",
			input:    "\033]133;A",
			expected: nil,
			valid:    false,
		},
		{
			name:     "Invalid sequence - wrong prefix",
			input:    "\033[133;A\007",
			expected: nil,
			valid:    false,
		},
		{
			name:     "Not an OSC sequence",
			input:    "regular text",
			expected: nil,
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := ParseOSCSequence(tt.input)

			if valid != tt.valid {
				t.Errorf("ParseOSCSequence() valid = %v, want %v", valid, tt.valid)
				return
			}

			if !tt.valid {
				if result != nil {
					t.Errorf("ParseOSCSequence() expected nil result for invalid input, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("ParseOSCSequence() expected result, got nil")
				return
			}

			if result.Type != tt.expected.Type {
				t.Errorf("ParseOSCSequence() Type = %v, want %v", result.Type, tt.expected.Type)
			}

			if result.ExitCode != tt.expected.ExitCode {
				t.Errorf("ParseOSCSequence() ExitCode = %v, want %v", result.ExitCode, tt.expected.ExitCode)
			}

			if result.Raw != tt.expected.Raw {
				t.Errorf("ParseOSCSequence() Raw = %v, want %v", result.Raw, tt.expected.Raw)
			}
		})
	}
}

func TestGetCommandState(t *testing.T) {
	tests := []struct {
		name     string
		oscType  string
		expected CommandState
	}{
		{
			name:     "Prompt start",
			oscType:  "133;A",
			expected: StatePromptStart,
		},
		{
			name:     "Command start",
			oscType:  "133;B",
			expected: StateCommandStart,
		},
		{
			name:     "Output start",
			oscType:  "133;C",
			expected: StateOutputStart,
		},
		{
			name:     "Command end without exit code",
			oscType:  "133;D",
			expected: StateCommandEnd,
		},
		{
			name:     "Command end with exit code",
			oscType:  "133;D;0",
			expected: StateCommandEnd,
		},
		{
			name:     "Command end with error exit code",
			oscType:  "133;D;1",
			expected: StateCommandEnd,
		},
		{
			name:     "Unknown sequence",
			oscType:  "133;X",
			expected: StateIdle,
		},
		{
			name:     "Invalid sequence",
			oscType:  "134;A",
			expected: StateIdle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommandState(tt.oscType)
			if result != tt.expected {
				t.Errorf("GetCommandState() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsCommandComplete(t *testing.T) {
	tests := []struct {
		name     string
		sequence *OSCSequence
		expected bool
	}{
		{
			name: "Command end sequence",
			sequence: &OSCSequence{
				Type:     "133;D;0",
				ExitCode: 0,
			},
			expected: true,
		},
		{
			name: "Command end with error",
			sequence: &OSCSequence{
				Type:     "133;D;1",
				ExitCode: 1,
			},
			expected: true,
		},
		{
			name: "Prompt start sequence",
			sequence: &OSCSequence{
				Type: "133;A",
			},
			expected: false,
		},
		{
			name: "Command start sequence",
			sequence: &OSCSequence{
				Type: "133;B",
			},
			expected: false,
		},
		{
			name: "Output start sequence",
			sequence: &OSCSequence{
				Type: "133;C",
			},
			expected: false,
		},
		{
			name:     "Nil sequence",
			sequence: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCommandComplete(tt.sequence)
			if result != tt.expected {
				t.Errorf("IsCommandComplete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatOSCSequence(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		data     []string
		expected string
	}{
		{
			name:     "Prompt start",
			command:  "A",
			data:     nil,
			expected: "\033]133;A\007",
		},
		{
			name:     "Command end with exit code",
			command:  "D",
			data:     []string{"0"},
			expected: "\033]133;D;0\007",
		},
		{
			name:     "Command end with error",
			command:  "D",
			data:     []string{"1"},
			expected: "\033]133;D;1\007",
		},
		{
			name:     "Multiple data parts",
			command:  "X",
			data:     []string{"part1", "part2"},
			expected: "\033]133;X;part1;part2\007",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOSCSequence(tt.command, tt.data...)
			if result != tt.expected {
				t.Errorf("FormatOSCSequence() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShellIntegrationScript(t *testing.T) {
	script := ShellIntegrationScript()

	// Check that the script contains expected components
	expectedComponents := []string{
		"__neuro_command_started",
		"__neuro_in_command",
		"__neuro_prompt_start",
		"__neuro_pre_command",
		"__neuro_post_command",
		"trap '__neuro_pre_command' DEBUG",
		"PROMPT_COMMAND=",
		OSC + OSC133_PROMPT_START + BEL,
		OSC + OSC133_COMMAND_START + BEL,
		OSC + OSC133_OUTPUT_START + BEL,
		OSC + OSC133_COMMAND_END,
	}

	for _, component := range expectedComponents {
		if !strings.Contains(script, component) {
			t.Errorf("ShellIntegrationScript() missing component: %s", component)
		}
	}

	// Ensure the script is not empty
	if len(script) == 0 {
		t.Error("ShellIntegrationScript() returned empty script")
	}
}

func TestCommandStateString(t *testing.T) {
	tests := []struct {
		state    CommandState
		expected string
	}{
		{StateIdle, "Idle"},
		{StatePromptStart, "PromptStart"},
		{StateCommandStart, "CommandStart"},
		{StateOutputStart, "OutputStart"},
		{StateCommandEnd, "CommandEnd"},
		{CommandState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("CommandState.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
