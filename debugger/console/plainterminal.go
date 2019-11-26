package console

import (
	"fmt"
	"gopher2600/gui"
	"io"
	"os"
)

// PlainTerminal is the default, most basic terminal interface. It keeps the
// terminal in whatever mode it started, probably cooked mode. As such, it
// offers only rudimentary editing facility and little control over output.
type PlainTerminal struct {
	input    io.Reader
	output   io.Writer
	silenced bool
}

// Initialise perfoms any setting up required for the terminal
func (pt *PlainTerminal) Initialise() error {
	pt.input = os.Stdin
	pt.output = os.Stdout
	return nil
}

// CleanUp perfoms any cleaning up required for the terminal
func (pt *PlainTerminal) CleanUp() {
}

// RegisterTabCompleter adds an implementation of TabCompleter to the terminal
func (pt *PlainTerminal) RegisterTabCompleter(TabCompleter) {
}

// UserPrint is the plain terminal print routine
func (pt PlainTerminal) UserPrint(style Style, s string, a ...interface{}) {
	if pt.silenced && style != StyleError {
		return
	}

	switch style {
	case StyleError:
		s = fmt.Sprintf("* %s", s)
	case StyleHelp:
		s = fmt.Sprintf("  %s", s)
	}

	s = fmt.Sprintf(s, a...)
	pt.output.Write([]byte(s))

	if style != StylePrompt {
		pt.output.Write([]byte("\n"))
	}
}

// UserRead is the plain terminal read routine
func (pt PlainTerminal) UserRead(input []byte, prompt Prompt, _ chan gui.Event, _ func(gui.Event) error) (int, error) {
	if pt.silenced {
		return 0, nil
	}

	pt.UserPrint(prompt.Style, prompt.Content)

	n, err := pt.input.Read(input)
	if err != nil {
		return n, err
	}
	return n, nil
}

// IsInteractive implements the console.UserInput interface
func (pt *PlainTerminal) IsInteractive() bool {
	return true
}

// Silence implemented the console.UserOutput interface
func (pt *PlainTerminal) Silence(silenced bool) {
	pt.silenced = silenced
}
