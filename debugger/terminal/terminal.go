// This file is part of Gopher2600.
//
// Gopher2600 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Gopher2600 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gopher2600.  If not, see <https://www.gnu.org/licenses/>.
//
// *** NOTE: all historical versions of this file, as found in any
// git repository, are also covered by the licence, even when this
// notice is not present ***

package terminal

import (
	"gopher2600/gui"
)

// Prompt specifies the prompt text and the prompt style.
type Prompt struct {
	Content string
	Style   Style
}

// Input defines the operations required by an interface that allows input.
type Input interface {
	// the TermRead loop should listen (if possible) for events on eventChannel
	// and call eventHandler with the received event as the argument.
	TermRead(buffer []byte, prompt Prompt, eventChannel chan gui.Event, eventHandler func(gui.Event) error) (int, error)

	// IsInteractive() should return true for implementations that require user
	// interaction. implementations that don't require a user to interact with
	// the debugger should return false.
	IsInteractive() bool
}

// Output defines the operations required by an interface that allows output.
type Output interface {
	TermPrintLine(Style, string, ...interface{})
}

// Terminal defines the operations required by the debugger's command line interface.
type Terminal interface {
	Initialise() error
	CleanUp()

	// register the tab completion engine to use with the UserInput
	// implementation
	RegisterTabCompletion(TabCompletion)

	// Silence all input and output (except error messages)
	Silence(silenced bool)

	// Userinterfaces, by definition, embed the Input and Output interfaces
	Input
	Output
}

// TabCompletion defines the operations required for tab completion. A good
// implementation can be found in the commandline sub-package.
type TabCompletion interface {
	Complete(input string) string
	Reset()
}