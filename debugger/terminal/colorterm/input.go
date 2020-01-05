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

package colorterm

import (
	"gopher2600/debugger/terminal"
	"gopher2600/debugger/terminal/colorterm/ansi"
	"gopher2600/debugger/terminal/colorterm/easyterm"
	"gopher2600/errors"
	"gopher2600/gui"
	"unicode"
	"unicode/utf8"
)

// #cursor #keys #tab #completion

// TermRead implements the terminal.Terminal interface
func (ct *ColorTerminal) TermRead(input []byte, prompt terminal.Prompt, events chan gui.Event, eventHandler func(gui.Event) error) (int, error) {

	if ct.silenced {
		return 0, nil
	}

	// ctrl-c handling: currently, we put the terminal into rawmode and listen
	// for ctrl-c event using the readRune reader.

	ct.RawMode()
	defer ct.CanonicalMode()

	// er is used to store encoded runes (length of 4 should be enough)
	er := make([]byte, 4)

	inputLen := 0
	cursorPos := 0
	history := len(ct.commandHistory)

	// liveBuffInput is used to store the latest input when we scroll through
	// history - we don't want to lose what we've typed in case the user wants
	// to resume where we left off
	liveHistory := make([]byte, cap(input))
	liveHistoryLen := 0

	// the method for cursor placement is as follows:
	//	 for each iteration in the loop
	//		1. store current cursor position
	//		2. clear the current line
	//		3. output the prompt
	//		4. output the input buffer
	//		5. restore the cursor position
	//
	// for this to work we need to place the cursor in it's initial position
	// before we begin the loop
	ct.EasyTerm.TermPrint("\r")
	ct.EasyTerm.TermPrint(ansi.CursorMove(len(prompt.Content)))

	for {
		ct.EasyTerm.TermPrint(ansi.CursorStore)
		ct.TermPrintLine(prompt.Style, "%s%s", ansi.ClearLine, prompt.Content)
		ct.TermPrintLine(terminal.StyleInput, string(input[:inputLen]))
		ct.EasyTerm.TermPrint(ansi.CursorRestore)

		select {
		case event := <-events:
			// handle functions that are passsed on over interruptChannel. these can
			// be things like events from the television GUI. eg. mouse clicks,
			// key presses, etc.
			ct.EasyTerm.TermPrint(ansi.CursorStore)
			err := eventHandler(event)
			ct.EasyTerm.TermPrint(ansi.CursorRestore)
			if err != nil {
				return inputLen + 1, err
			}

		case readRune := <-ct.reader:
			if readRune.err != nil {
				return inputLen, readRune.err
			}

			switch readRune.r {
			case easyterm.KeyTab:
				if ct.tabCompletion != nil {
					s := ct.tabCompletion.Complete(string(input[:cursorPos]))

					// the difference in the length of the new input and the old
					// input
					d := len(s) - cursorPos

					if inputLen+d <= len(input) {
						// append everything after the cursor to the new string and copy
						// into input array
						s += string(input[cursorPos:])
						copy(input, []byte(s))

						// advance character to end of completed word
						ct.EasyTerm.TermPrint(ansi.CursorMove(d))
						cursorPos += d

						// note new used-length of input array
						inputLen += d
					}
				}

			case easyterm.KeyInterrupt:
				// CTRL-C -- note that there is a ctrl-c signal handler, set up in
				// debugger.Start(), that controls the main debugging loop. this
				// ctrl-c handler by contrast, controls the user input loop

				if inputLen > 0 {
					// clear current input
					inputLen = 0
					cursorPos = 0
					ct.EasyTerm.TermPrint("\r")
					ct.EasyTerm.TermPrint(ansi.CursorMove(len(prompt.Content)))
				} else {
					// there is no input so return UserInterrupt error
					ct.EasyTerm.TermPrint("\n")
					return inputLen + 1, errors.New(errors.UserInterrupt)
				}

			case easyterm.KeySuspend:
				// CTRL-Z
				return inputLen + 1, errors.New(errors.UserSuspend)

			case easyterm.KeyCarriageReturn:
				// CARRIAGE RETURN

				// check to see if input is the same as the last history entry
				newEntry := false
				if inputLen > 0 {
					newEntry = true
					if len(ct.commandHistory) > 0 {
						lastHistoryEntry := ct.commandHistory[len(ct.commandHistory)-1].input
						if len(lastHistoryEntry) == inputLen {
							newEntry = false
							for i := 0; i < inputLen; i++ {
								if input[i] != lastHistoryEntry[i] {
									newEntry = true
									break
								}
							}
						}
					}
				}

				// if input is not the same as the last history entry then append a
				// new entry to the history list
				if newEntry {
					nh := make([]byte, inputLen)
					copy(nh, input[:inputLen])
					ct.commandHistory = append(ct.commandHistory, command{input: nh})
				}

				ct.EasyTerm.TermPrint("\r\n")
				return inputLen + 1, nil

			case easyterm.KeyEsc:
				// ESCAPE SEQUENCE BEGIN
				readRune = <-ct.reader
				if readRune.err != nil {
					return inputLen, readRune.err
				}
				switch readRune.r {
				case easyterm.EscCursor:
					// CURSOR KEY
					readRune = <-ct.reader
					if readRune.err != nil {
						return inputLen, readRune.err
					}

					switch readRune.r {
					case easyterm.CursorUp:
						// move up through command history
						if len(ct.commandHistory) > 0 {
							// if we're at the end of the command history then store
							// the current input in liveBuffInput for possible later editing
							if history == len(ct.commandHistory) {
								copy(liveHistory, input[:inputLen])
								liveHistoryLen = inputLen
							}

							if history > 0 {
								history--
								l := len(ct.commandHistory[history].input)

								// length check in case input buffer is
								// shorted from when history entry was added
								if l < len(input) {
									copy(input, ct.commandHistory[history].input)
									inputLen = len(ct.commandHistory[history].input)
									ct.EasyTerm.TermPrint(ansi.CursorMove(inputLen - cursorPos))
									cursorPos = inputLen
								}
							}
						}
					case easyterm.CursorDown:
						// move down through command history
						if len(ct.commandHistory) > 0 {
							if history < len(ct.commandHistory)-1 {
								history++
								l := len(ct.commandHistory[history].input)
								if l < len(input) {
									copy(input, ct.commandHistory[history].input)
									inputLen = len(ct.commandHistory[history].input)
									ct.EasyTerm.TermPrint(ansi.CursorMove(inputLen - cursorPos))
									cursorPos = inputLen
								}
							} else if history == len(ct.commandHistory)-1 {
								history++

								// length check not really required because
								// liveHistroy should not ever be greater
								// in length than that of input buffer
								if liveHistoryLen < len(input) {
									copy(input, liveHistory)
									inputLen = liveHistoryLen
									ct.EasyTerm.TermPrint(ansi.CursorMove(inputLen - cursorPos))
									cursorPos = inputLen
								}
							}
						}
					case easyterm.CursorForward:
						// move forward through current command input
						if cursorPos < inputLen {
							ct.EasyTerm.TermPrint(ansi.CursorForwardOne)
							cursorPos++
						}
					case easyterm.CursorBackward:
						// move backward through current command input
						if cursorPos > 0 {
							ct.EasyTerm.TermPrint(ansi.CursorBackwardOne)
							cursorPos--
						}

					case easyterm.EscDelete:
						// DELETE
						if cursorPos < inputLen {
							copy(input[cursorPos:], input[cursorPos+1:])
							inputLen--
							history = len(ct.commandHistory)
						}

						// eat the third character in the sequence
						readRune = <-ct.reader

					case easyterm.EscHome:
						ct.EasyTerm.TermPrint(ansi.CursorMove(-cursorPos))
						cursorPos = 0

					case easyterm.EscEnd:
						ct.EasyTerm.TermPrint(ansi.CursorMove(inputLen - cursorPos))
						cursorPos = inputLen
					}
				}

			case easyterm.KeyBackspace:
				// BACKSPACE
				if cursorPos > 0 {
					copy(input[cursorPos-1:], input[cursorPos:])
					ct.EasyTerm.TermPrint(ansi.CursorBackwardOne)
					cursorPos--
					inputLen--
					history = len(ct.commandHistory)
				}

			default:
				if unicode.IsDigit(readRune.r) || unicode.IsLetter(readRune.r) || unicode.IsSpace(readRune.r) || unicode.IsPunct(readRune.r) || unicode.IsSymbol(readRune.r) {

					l := utf8.EncodeRune(er, readRune.r)

					// make sure we don't overflow the input buffer
					if cursorPos+l <= len(input) {
						ct.EasyTerm.TermPrint(ansi.CursorForwardOne)

						// insert new character into input stream at current cursor
						// position
						copy(input[cursorPos+l:], input[cursorPos:])
						copy(input[cursorPos:], er[:l])
						cursorPos++

						inputLen += l

						// make sure history pointer is at the end of the command
						// history array
						history = len(ct.commandHistory)
					}
				}
			}
		}
	}
}