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

package debugger

// this file holds the functions/structures to be used when outputting to the
// terminal. The TermPrint functions of the Terminal interface should not be
// used directly.

import (
	"fmt"
	"gopher2600/debugger/terminal"
	"strings"
)

// all print operations from the debugger should be made with the this printLine()
// function. output will be normalised and sent to the attached terminal as
// required.
func (dbg *Debugger) printLine(sty terminal.Style, s string, a ...interface{}) {
	// resolve string placeholders for styles other than the help style. not
	// filtering the help style causes HELP output to fail; because the
	// commandline template uses fmt style placeholders.
	if sty != terminal.StyleHelp {
		s = fmt.Sprintf(s, a...)
	}

	// remove all trailing newlines, and return if the resulting string is empty
	s = strings.TrimRight(s, "\n")
	if len(s) == 0 {
		return
	}

	// split string if necessary
	t := strings.Split(s, "\n")
	for _, s := range t {
		dbg.term.TermPrintLine(sty, s)
	}
}

// styleWriter implements the io.Writer interface. it is useful for when an
// io.Writer is required and you want to direct the output to the terminal.
// allows the application of a single style.
type styleWriter struct {
	dbg   *Debugger
	style terminal.Style
}

func (dbg *Debugger) printStyle(sty terminal.Style) *styleWriter {
	return &styleWriter{
		dbg:   dbg,
		style: sty,
	}
}

func (wrt styleWriter) Write(p []byte) (n int, err error) {
	wrt.dbg.printLine(wrt.style, string(p))
	return len(p), nil
}

func (dbg *Debugger) printInstrument(mi fmt.Stringer) {
	dbg.printLine(terminal.StyleInstrument, "%s", mi.String())
}
