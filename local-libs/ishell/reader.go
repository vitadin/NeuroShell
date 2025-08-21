package ishell

import (
	"bytes"
	"strings"
	"sync"

	"github.com/chzyer/readline"
)

type (
	lineString struct {
		line string
		err  error
	}

	shellReader struct {
		scanner      *readline.Instance
		consumers    chan lineString
		reading      bool
		readingMulti bool
		buf          *bytes.Buffer
		prompt       string
		multiPrompt  string
		showPrompt   bool
		completer    readline.AutoCompleter
		defaultInput string
		promptPrefix []string // Lines to display before the actual prompt
		sync.Mutex
	}
)

// rlPrompt returns the proper prompt for readline based on showPrompt and
// prompt members.
func (s *shellReader) rlPrompt() string {
	if s.showPrompt {
		if s.readingMulti {
			return s.multiPrompt
		}
		return s.prompt
	}
	return ""
}

func (s *shellReader) readPasswordErr() (string, error) {
	prompt := ""
	if s.buf.Len() > 0 {
		prompt = s.buf.String()
		s.buf.Truncate(0)
	}
	password, err := s.scanner.ReadPassword(prompt)
	return string(password), err
}

func (s *shellReader) readPassword() string {
	password, _ := s.readPasswordErr()
	return password
}

func (s *shellReader) setMultiMode(use bool) {
	s.readingMulti = use
}

// setPromptPrefix sets the prefix lines to display before the actual prompt
func (s *shellReader) setPromptPrefix(lines []string) {
	s.promptPrefix = lines
}

func (s *shellReader) readLine(consumer chan lineString) {
	s.Lock()
	defer s.Unlock()

	// already reading
	if s.reading {
		return
	}
	s.reading = true
	// start reading

	// Display prompt prefix lines (multi-line prompt support)
	if len(s.promptPrefix) > 0 && s.showPrompt {
		// Add separator newline before first prompt line for multi-line prompts only
		s.scanner.Stdout().Write([]byte("\n"))
		for _, line := range s.promptPrefix {
			// Use ANSI codes to ensure proper display
			// Clear line and print prefix
			s.scanner.Stdout().Write([]byte("\r\033[K" + line + "\n"))
		}
	}

	// detect if print is called to
	// prevent readline lib from clearing line.
	// use the last line as prompt.
	// TODO find better way.
	shellPrompt := s.prompt
	prompt := s.rlPrompt()
	if s.buf.Len() > 0 {
		lines := strings.Split(s.buf.String(), "\n")
		if p := lines[len(lines)-1]; strings.TrimSpace(p) != "" {
			prompt = p
		}
		s.buf.Truncate(0)
	}

	// use printed statement as prompt
	s.scanner.SetPrompt(prompt)

	line, err := s.scanner.ReadlineWithDefault(s.defaultInput)

	// reset prompt
	s.scanner.SetPrompt(shellPrompt)

	ls := lineString{string(line), err}
	consumer <- ls
	s.reading = false
}
