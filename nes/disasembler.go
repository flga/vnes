package nes

import (
	"fmt"
	"io"
	"strings"
)

// TODO: rework this
func disassemble(out io.Writer, bus *SysBus,
	inst_pc uint16, a, x, y, p, sp byte,
	inst Instruction, intermediateAddr, resolvedAddr uint16, cycles uint64, ppu *PPU) {
	var strlen int

	n, _ := fmt.Fprintf(out, "%04X  ", inst_pc)
	strlen += n

	if inst.Size == 1 {
		n, _ := fmt.Fprintf(out, "%02X      ", inst.OpCode)
		strlen += n
	} else if inst.Size == 2 {
		n, _ := fmt.Fprintf(out, "%02X %02X   ", inst.OpCode, bus.Read(inst_pc+1))
		strlen += n
	} else if inst.Size == 3 {
		n, _ := fmt.Fprintf(out, "%02X %02X %02X", inst.OpCode, bus.Read(inst_pc+1), bus.Read(inst_pc+2))
		strlen += n
	}

	if inst.Illegal {
		n, _ := fmt.Fprint(out, " *")
		strlen += n
	} else {
		n, _ := fmt.Fprint(out, "  ")
		strlen += n
	}

	n, _ = fmt.Fprint(out, inst.Name, " ")
	strlen += n

	switch inst.Mode {
	case Accumulator:
		n, _ := fmt.Fprint(out, "A")
		strlen += n
	case Implied:
	default:
		var arg uint16
		switch inst.Mode {
		case Immediate, ZeroPage, ZeroPageIndexedX, ZeroPageIndexedY, PreIndexedIndirect, PostIndexedIndirect:
			arg = uint16(bus.Read(inst_pc + 1))
		case Absolute, Indirect, IndexedX, IndexedY:
			arg = uint16(bus.Read(inst_pc+1)) | uint16(bus.Read(inst_pc+2))<<8
		case Relative:
			arg = resolvedAddr
		}

		n, _ := fmt.Fprintf(out, addressingFormats[inst.Mode], arg)
		strlen += n
	}

	// // DEBUG INFO
	// switch inst.Mode {
	// case Indirect:
	// 	n, _ := fmt.Fprintf(out, " = %04X", resolvedAddr)
	// 	strlen += n
	// case ZeroPage, Absolute:
	// 	if inst.Name != "JMP" && inst.Name != "JSR" {
	// 		n, _ := fmt.Fprintf(out, " = %02X", bus.Read(resolvedAddr))
	// 		strlen += n
	// 	}
	// case IndexedY, IndexedX:
	// 	n, _ := fmt.Fprintf(out, " @ %04X = %02X", resolvedAddr, bus.Read(resolvedAddr))
	// 	strlen += n
	// case ZeroPageIndexedY, ZeroPageIndexedX:
	// 	n, _ := fmt.Fprintf(out, " @ %02X = %02X", resolvedAddr, bus.Read(resolvedAddr))
	// 	strlen += n

	// case PreIndexedIndirect:
	// 	n, _ := fmt.Fprintf(out, " @ %02X = %04X = %02X", intermediateAddr, resolvedAddr, bus.Read(resolvedAddr))
	// 	strlen += n
	// case PostIndexedIndirect:
	// 	n, _ := fmt.Fprintf(out, " = %04X @ %04X = %02X", intermediateAddr, resolvedAddr, bus.Read(resolvedAddr))
	// 	strlen += n
	// }

	fmt.Fprint(out, strings.Repeat(" ", 48-strlen))
	var col, scanLine int
	if ppu != nil {
		col, scanLine = ppu.Dot, ppu.ScanLine
	}
	fmt.Fprintf(out, "A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3d,%3d CYC:%d\n", a, x, y, p, sp, col, scanLine, cycles)
	// fmt.Fprintf(out, "A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%d\n", a, x, y, p, sp, cycles /* , frame */)
}

var addressingFormats = map[AddressingMode]string{
	Immediate:           "#$%02X",    // #aa
	Absolute:            "$%04X",     // aaaa
	ZeroPage:            "$%02X",     // aa
	Implied:             "",          //
	Indirect:            "($%04X)",   // (aaaa)
	IndexedX:            "$%04X,X",   // aaaa,X
	IndexedY:            "$%04X,Y",   // aaaa,Y
	ZeroPageIndexedX:    "$%02X,X",   // aa,X
	ZeroPageIndexedY:    "$%02X,Y",   // aa,Y
	PreIndexedIndirect:  "($%02X,X)", // (aa,X)
	PostIndexedIndirect: "($%02X),Y", // (aa),Y
	Relative:            "$%04X",     // aaaa
	Accumulator:         "A",         // A
}
