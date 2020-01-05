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

package instructions

import "fmt"

// AddressingMode describes the method data for the instruction should be received
type AddressingMode int

// List of supported addressing modes
const (
	Implied AddressingMode = iota
	Immediate
	Relative // relative addressing is used for branch instructions

	Absolute // sometimes called absolute addressing
	ZeroPage
	Indirect // indirect addressing (with no indexing) is only for JMP instructions

	PreIndexedIndirect  // uses X register
	PostIndexedIndirect // uses Y register
	AbsoluteIndexedX
	AbsoluteIndexedY
	IndexedZeroPageX
	IndexedZeroPageY // only used for LDX
)

// EffectCategory categorises an instruction by the effect it has
type EffectCategory int

// List of effect categories
const (
	Read EffectCategory = iota
	Write
	RMW

	// the following three effects have a variable effect on the program
	// counter, depending on the instruction's precise operand
	Flow
	Subroutine
	Interrupt
)

// Definition defines each instruction in the instruction set; one per instruction
type Definition struct {
	OpCode         uint8
	Mnemonic       string
	Bytes          int
	Cycles         int
	AddressingMode AddressingMode
	PageSensitive  bool
	Effect         EffectCategory
}

// String returns a single instruction definition as a string.
func (defn Definition) String() string {
	if defn.Mnemonic == "" {
		return "undecoded instruction"
	}
	return fmt.Sprintf("%02x %s +%dbytes (%d cycles) [mode=%d pagesens=%t effect=%d]", defn.OpCode, defn.Mnemonic, defn.Bytes, defn.Cycles, defn.AddressingMode, defn.PageSensitive, defn.Effect)
}