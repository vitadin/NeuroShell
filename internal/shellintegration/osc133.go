// Package shellintegration provides modern shell integration protocols for NeuroShell.
// It implements OSC 133 sequences for reliable command completion detection.
package shellintegration

import (
	"strconv"
	"strings"
)

// OSC 133 sequence constants for shell integration
const (
	// ESC = ASCII 27 (0x1B)
	ESC = "\033"
	// BEL = ASCII 7 (0x07) - Bell character used as string terminator
	BEL = "\007"
	// OSC = Operating System Command prefix
	OSC = ESC + "]"
	// ST = String Terminator (alternative to BEL)
	ST = ESC + "\\"

	// OSC 133 command codes
	OSC133_PROMPT_START  = "133;A" // Mark prompt start
	OSC133_COMMAND_START = "133;B" // Mark command start
	OSC133_OUTPUT_START  = "133;C" // Mark command output start
	OSC133_COMMAND_END   = "133;D" // Mark command end with exit code
)

// CommandState represents the current state of command execution
type CommandState int

const (
	StateIdle CommandState = iota
	StatePromptStart
	StateCommandStart
	StateOutputStart
	StateCommandEnd
)

func (s CommandState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StatePromptStart:
		return "PromptStart"
	case StateCommandStart:
		return "CommandStart"
	case StateOutputStart:
		return "OutputStart"
	case StateCommandEnd:
		return "CommandEnd"
	default:
		return "Unknown"
	}
}

// OSCSequence represents a parsed OSC sequence
type OSCSequence struct {
	Type     string // e.g., "133;A", "133;D"
	ExitCode int    // Only valid for command end sequences
	Raw      string // Original sequence text
}

// ShellIntegrationScript returns the bash script to set up shell integration
func ShellIntegrationScript() string {
	return `
# NeuroShell Integration Setup
__neuro_command_started=0
__neuro_in_command=0

# Prompt start marker
__neuro_prompt_start() {
    printf '\033]133;A\007'
}

# Command start detection
__neuro_pre_command() {
    if [ "$__neuro_command_started" -eq 0 ] && [ "$__neuro_in_command" -eq 0 ]; then
        __neuro_command_started=1
        __neuro_in_command=1
        printf '\033]133;B\007'
        printf '\033]133;C\007'
    fi
}

# Command end detection
__neuro_post_command() {
    local exit_code=$?
    if [ "$__neuro_command_started" -eq 1 ]; then
        printf '\033]133;D;%s\007' "$exit_code"
        __neuro_command_started=0
        __neuro_in_command=0
    fi
    __neuro_prompt_start
    return $exit_code
}

# Set up traps and prompt
trap '__neuro_pre_command' DEBUG
PROMPT_COMMAND="__neuro_post_command${PROMPT_COMMAND:+; $PROMPT_COMMAND}"

# Send initial prompt start
__neuro_prompt_start
`
}

// ParseOSCSequence parses an OSC sequence from input text
func ParseOSCSequence(text string) (*OSCSequence, bool) {
	// Look for OSC sequences: ESC]133;X[;data]BEL or ESC]133;X[;data]ST
	if !strings.HasPrefix(text, OSC+"133;") {
		return nil, false
	}

	var terminator string
	var content string

	// Check for BEL terminator
	if idx := strings.Index(text, BEL); idx != -1 {
		terminator = BEL
		content = text[len(OSC):idx]
	} else if idx := strings.Index(text, ST); idx != -1 {
		// Check for ST terminator
		terminator = ST
		content = text[len(OSC):idx]
	} else {
		// No terminator found
		return nil, false
	}

	// Parse the sequence content
	parts := strings.Split(content, ";")
	if len(parts) < 2 || parts[0] != "133" {
		return nil, false
	}

	seq := &OSCSequence{
		Type: content,
		Raw:  text[:strings.Index(text, terminator)+len(terminator)],
	}

	// Parse exit code for command end sequences
	if parts[1] == "D" && len(parts) >= 3 {
		if exitCode, err := strconv.Atoi(parts[2]); err == nil {
			seq.ExitCode = exitCode
		}
	}

	return seq, true
}

// GetCommandState returns the command state for a given OSC sequence type
func GetCommandState(oscType string) CommandState {
	switch {
	case oscType == OSC133_PROMPT_START:
		return StatePromptStart
	case oscType == OSC133_COMMAND_START:
		return StateCommandStart
	case oscType == OSC133_OUTPUT_START:
		return StateOutputStart
	case strings.HasPrefix(oscType, "133;D"):
		return StateCommandEnd
	default:
		return StateIdle
	}
}

// IsCommandComplete checks if the given OSC sequence indicates command completion
func IsCommandComplete(seq *OSCSequence) bool {
	return seq != nil && strings.HasPrefix(seq.Type, "133;D")
}

// FormatOSCSequence creates an OSC sequence string
func FormatOSCSequence(command string, data ...string) string {
	sequence := OSC + "133;" + command
	if len(data) > 0 {
		sequence += ";" + strings.Join(data, ";")
	}
	sequence += BEL
	return sequence
}
