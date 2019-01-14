package disassembly

import (
	"fmt"
	"gopher2600/errors"
	"gopher2600/hardware/cpu"
	"gopher2600/hardware/cpu/result"
	"gopher2600/hardware/memory"
	"gopher2600/symbols"
	"io"
	"strings"
)

// Disassembly represents the annotated disassembly of a 6502 binary
type Disassembly struct {
	// symbols used to build disassembly output
	Symtable *symbols.Table

	// SequencePoints contains the list of program counter values. listed in
	// order so can be used to index program map to produce complete
	// disassembly
	SequencePoints []uint16

	// table of instruction results. index with contents of sequencePoints
	Program map[uint16]*result.Instruction
}

// ParseMemory disassembles an existing memory instance. uses a new cpu
// instance which has no side effects, so it's safe to use with "live" memory
func (dsm *Disassembly) ParseMemory(mem *memory.VCSMemory, symtable *symbols.Table) error {
	dsm.Symtable = symtable
	dsm.Program = make(map[uint16]*result.Instruction)
	dsm.SequencePoints = make([]uint16, 0, mem.Cart.Memtop()-mem.Cart.Origin())

	// create a new non-branching CPU to disassemble memory
	mc, err := cpu.NewCPU(mem)
	if err != nil {
		return err
	}
	mc.NoSideEffects = true

	// start disassembly at reset point
	mc.LoadPC(memory.AddressReset)

	for {
		ir, err := mc.ExecuteInstruction(func(ir *result.Instruction) {})

		// filter out some errors
		if err != nil {
			switch err := err.(type) {
			case errors.GopherError:
				switch err.Errno {
				case errors.ProgramCounterCycled:
					// reached end of memory, exit loop with no errors
					// TODO: handle multi-bank ROMS
					return nil
				case errors.NullInstruction:
					// we've encountered a null instruction. ignore
					continue
				case errors.UnimplementedInstruction:
					// ignore unimplemented instructions
					continue
				case errors.UnreadableAddress:
					// ignore unreadable addresses
					continue
				default:
					return err
				}
			default:
				return err
			}
		}

		// check validity
		err = ir.IsValid()
		if err != nil {
			return err
		}

		// add instruction result to disassembly result. an instruction result
		// of nil means that the part of the program just read by the CPU does
		// not contain valid instructions (maybe the assembler reasoned that
		// the code is unreachable)
		dsm.SequencePoints = append(dsm.SequencePoints, ir.Address)
		dsm.Program[ir.Address] = ir
	}
}

// NewDisassembly initialises a new partial emulation and returns a
// disassembly from the supplied cartridge filename. - useful for one-shot
// disassemblies, like the gopher2600 "disasm" mode
func NewDisassembly(cartridgeFilename string) (*Disassembly, error) {
	// ignore errors caused by loading of symbols table
	symtable, err := symbols.ReadSymbolsFile(cartridgeFilename)
	if err != nil {
		fmt.Println(err)
		symtable, err = symbols.StandardSymbolTable()
		if err != nil {
			return nil, err
		}
	}

	mem, err := memory.NewVCSMemory()
	if err != nil {
		return nil, err
	}

	err = mem.Cart.Attach(cartridgeFilename)
	if err != nil {
		return nil, err
	}

	dsm := new(Disassembly)
	err = dsm.ParseMemory(mem, symtable)
	if err != nil {
		return dsm, err
	}

	return dsm, nil
}

// Dump writes the entire disassembly to the write interface
func (dsm *Disassembly) Dump(output io.Writer) {
	for _, pc := range dsm.SequencePoints {
		output.Write([]byte(dsm.Program[pc].GetString(dsm.Symtable, result.StyleFull)))
		output.Write([]byte("\n"))
	}
}

// Grep searches the disassembly dump for search string. case sensitive
func (dsm *Disassembly) Grep(search string, output io.Writer, caseSensitive bool) {
	var s, m string

	if !caseSensitive {
		search = strings.ToUpper(search)
	}

	for _, pc := range dsm.SequencePoints {
		s = dsm.Program[pc].GetString(dsm.Symtable, result.StyleBrief)
		if !caseSensitive {
			m = strings.ToUpper(s)
		} else {
			m = s
		}

		if strings.Contains(m, search) {
			output.Write([]byte(s))
			output.Write([]byte("\n"))
		}
	}
}
