package nes

import (
	"fmt"
	"io"
	"strings"
)

// TODO: rework this
func disassemble(out io.Writer, bus *sysBus,
	inst_pc uint16, a, x, y, p, sp byte,
	inst instruction, intermediateAddr, resolvedAddr uint16, cycles uint64, ppu *ppu) {
	var strlen int

	n, _ := fmt.Fprintf(out, "%04X  ", inst_pc)
	strlen += n

	if inst.size == 1 {
		n, _ := fmt.Fprintf(out, "%02X      ", inst.opCode)
		strlen += n
	} else if inst.size == 2 {
		n, _ := fmt.Fprintf(out, "%02X %02X   ", inst.opCode, bus.read(inst_pc+1))
		strlen += n
	} else if inst.size == 3 {
		n, _ := fmt.Fprintf(out, "%02X %02X %02X", inst.opCode, bus.read(inst_pc+1), bus.read(inst_pc+2))
		strlen += n
	}

	if inst.illegal {
		n, _ := fmt.Fprint(out, " *")
		strlen += n
	} else {
		n, _ := fmt.Fprint(out, "  ")
		strlen += n
	}

	n, _ = fmt.Fprint(out, inst.name, " ")
	strlen += n

	switch inst.mode {
	case accumulator:
		n, _ := fmt.Fprint(out, "A")
		strlen += n
	case implied:
	default:
		var arg uint16
		switch inst.mode {
		case immediate, zeroPage, zeroPageIndexedX, zeroPageIndexedY, preIndexedIndirect, postIndexedIndirect:
			arg = uint16(bus.read(inst_pc + 1))
		case absolute, indirect, indexedX, indexedY:
			arg = uint16(bus.read(inst_pc+1)) | uint16(bus.read(inst_pc+2))<<8
		case relative:
			arg = resolvedAddr
		}

		n, _ := fmt.Fprintf(out, addressingFormats[inst.mode], arg)
		strlen += n
	}

	// // DEBUG INFO
	// switch inst.mode {
	// case Indirect:
	// 	n, _ := fmt.Fprintf(out, " = %04X", resolvedAddr)
	// 	strlen += n
	// case ZeroPage, Absolute:
	// 	if inst.name != "JMP" && inst.name != "JSR" {
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
		col, scanLine = ppu.dot, ppu.scanline
	}
	fmt.Fprintf(out, "A:%02X X:%02X Y:%02X P:%02X SP:%02X PPU:%3d,%3d CYC:%d\n", a, x, y, p, sp, col, scanLine, cycles)
	// fmt.Fprintf(out, "A:%02X X:%02X Y:%02X P:%02X SP:%02X CYC:%d\n", a, x, y, p, sp, cycles /* , frame */)
}

var addressingFormats = map[addressingMode]string{
	immediate:           "#$%02X",    // #aa
	absolute:            "$%04X",     // aaaa
	zeroPage:            "$%02X",     // aa
	implied:             "",          //
	indirect:            "($%04X)",   // (aaaa)
	indexedX:            "$%04X,X",   // aaaa,X
	indexedY:            "$%04X,Y",   // aaaa,Y
	zeroPageIndexedX:    "$%02X,X",   // aa,X
	zeroPageIndexedY:    "$%02X,Y",   // aa,Y
	preIndexedIndirect:  "($%02X,X)", // (aa,X)
	postIndexedIndirect: "($%02X),Y", // (aa),Y
	relative:            "$%04X",     // aaaa
	accumulator:         "A",         // A
}
