package nes

// AddressingMode represents the various addressing modes of the NES.
//
//	The following content is taken from
//	http://www.thealmightyguru.com/Games/Hacking/Wiki/index.php/Addressing_Modes
//	and is here only for ease of use.
//
//	This content is available under GNU Free Documentation License 1.2 unless otherwise noted.
//	http://www.gnu.org/licenses/old-licenses/fdl-1.2.txt
//
// Instructions need operands on which to work. There are various ways of
// indicating where the processor is to get these operands. The different
// methods used to do this are called addressing modes.
//
//
//
// Immediate Addressing
//
// Immediate addressing is used when the operand's value is given in the
// instruction itself. In 6502 assembly, this is indicated by a pound sign, "#",
// before the operand.
//
// In this example, the first instruction uses Immediate addressing, while the
// second instruction uses ZeroPage addressing. In the first line, the
// accumulator is loaded with the number 7, but in the second line, the
// accumulator is loaded with whatever value is stored in memory address $0007.
//
//	 0001:A9 07     LDA #$07        ; Load A with the value of $07.
//	 0002:A5 07     LDA $07         ; Load A with whatever value is in memory address $0007.
//
// The following opcodes support Immediate addressing:
//
//	ADC    AND    CMP    CPX    CPY    EOR    LDA    LDX    LDY    ORA    SBC
//
//
//
// ZeroPage Addressing
//
// In ZeroPage addressing, the operand is a memory address rather than a value.
// As the name suggests, the memory must be on the zero-page of memory
// (addresses $0000-$00FF). You only need to supply the low byte of the memory
// address, the $00 high byte is automatically added by the processor.
//
// ZeroPage addressing is very similar to Absolute addressing, except that
// absolute addressing requires a full 2-byte address, and can access the full
// range of the processor's memory ($0000-$FFFF).
//
// This example uses ZeroPage addressing to load the accumulator with whatever
// value is stored in memory address $00EE. Notice that the operand doesn't
// include $00.
//
//	0001:A5 EE     LDA $EE         ; Load A with whatever value is in memory address $00EE.
//
// Since ZeroPage addressing is essentially the same as Absolute addressing,
// they use the same opcodes.
//
// The following opcodes support ZeroPage addressing:
//
//	ADC    AND    ASL    BIT    CMP    CPX    CPY    DEC    EOR    INC    LDA
//	LDX    LDY    LSR    ORA    ROL    ROR    SBC    STA    STX    STY
//
//
//
// Absolute Addressing
//
// In Absolute addressing, the operand is a memory address rather than a value.
// Absolute addressing is very similar to ZeroPage addressing, except that
// absolute addressing requires a full 2-byte address, and can access the full
// range of the processor's memory ($0000-$FFFF).
//
// This example uses Absolute addressing to load the accumulator with whatever
// value is stored in memory address $16A0. Notice that the address is stored
// in little-endian.
//
//	0001:AD A0 16  LDA $16A0       ; Load A with whatever value is in memory address $16A0.
//
// Since Absolute addressing is essentially the same as ZeroPage addressing,
// they use the same opcodes.
//
// The following opcodes support Absolute addressing:
//
//	ADC    AND    ASL    BIT    CMP    CPX    CPY    DEC    EOR    INC    LDA
//	LDX    LDY    LSR    ORA    ROL    ROR    SBC    STA    STX    STY
//
//
//
// Relative Addressing
//
// Relative addressing is used on the various Branch-On-Condition true.
// A 1-byte signed operand is added to the program counter, and the program
// continues execution from the new address. Because this value is signed,
// values #00-#7F are positive, and values #FF-#80 are negative.
//
// Keep in mind that the program counter will be set to the address after the
// branch instruction, so take this into account when calculating true
// new position.
//
// Since branching works by checking a relevant status bit, make sure it is true
// to the proper value prior to calling the branch true.
// This is usually done with a CMP instruction.
//
// If you need to move the program counter to a location greater or less than
// 127 bytes away from the current location, make a nearby JMP instruction,
// and move the program counter to the JMP line.
//
// This example creates a countdown loop. Memory address $50 is loaded with #10,
// and then decreased. We then check if the Zero Flag has been set, which will
// only occur when we decrease all the way down to 0. If we haven't reached
// zero, we go back to $0005 and decrease it again. If we have hit 0, we jump
// past the line that would send us back up, escaping the loop.
//
//	0001:A9 01     LDA #$10        ; Loads A with #10.
//	0003:85 50     STA $0050       ; Stores A into $0050.
//	0005:C6 50     DEC $0050       ; Decrement $0050.
//	0007:D0 08     BEQ $04         ; If the Zero Flag is set, JMP from our current location ($09),
//	                                 plus the operand ($04) to the address $000C.
//	0009:4C 05 00  JMP $0005       ; Jump back to $0005, creating a countdown loop.
//	000C:00        BRK             ; Done.
//
// The following opcodes support Relative addressing:
//
//	BCC    BCS    BEQ    BMI    BNE    BPL    BVC    BVS
//
//
//
// Implied Addressing
//
// Implied addressing occurs when there is no operand. The addressing mode is
// implied by the instruction.
//
// This example uses Implied addressing to transfer the value stored in the
// accumulator to the X register.
//
//	0001:AA        TAX             ; Transfer A to X.
//
// The following opcodes support Implied addressing:
//
//	BRK    CLC    CLD    CLI    CLV    DEX    DEY    INX    INY    NOP    PHA
//	PHP    PLA    PLP    RTI    RTS    TAX    TAY    TSX    TXA    TXS    TYT
//
//
//
// Accumulator Addressing
//
// Accumulator addressing is a special type of Implied addressing that
// only addresses the accumulator.
//
// This example uses Accumulator addressing to transfer the value stored in the
// accumulator to the X register.
//
//	0001:0A        ASL             ; Bit-shift the accumulator left.
//
// The following opcodes support Accumulator addressing:
//
//	ASL    LSR    ROL    ROR
//
//
//
// IndexedX and IndexedY Addressing
//
// In this mode, the address is added to the value either by the X or Y index
// register. Most opcodes support the X index register for the offset, but an
// additional handful also support the Y index register.
//
// The benefit of Indexed addressing is that you can quickly loop through memory
// by simply increasing or decreasing the offset.
//
// This example uses Indexed addressing to store #$7F into all the memory from
// 1000-10FF. Line 0005 uses Indexed addressing.
//
//	0001:A9 7F     LDA #$7F         ; Load A with 7F.
//	0003:A2 FF     LDX #$FF         ; Load X with FF.
//	0005:9D 00 10  STA $1000,X      ; Store A into $1000 offset with X.
//	0008:CA        DEX              ; Decrement X.
//	0008:30 FB     BPL $0005        ; Goto $0005 while X is not -1.
//
// The following opcodes support IndexedX addressing:
//
//	ADC    AND    ASL    CMP    DEC    EOR    INC    LDA    LDY    LSR    ORA
//	ROL    ROR    SBC    STA
//
// The following opcodes support IndexedY addressing:
//
//	ADC    AND    CMP    EOR    LDA    LDY    ORA    SBC    STA
//
//
//
// ZeroPageIndexedX and ZeroPageIndexedY Addressing
//
// In this mode, the address is added to the value either by the X or Y index
// register. Most opcodes support the X index register for the offset, but an
// additional handful also support the Y index register.
//
// The benefit of ZeroPageIndexed addressing is that you can quickly loop
// through memory by simply increasing or decreasing the offset.
//
// This example uses ZeroPageIndexed addressing to store #$7F into all the
// memory from 0000-00FF. Line 0005 uses ZeroPageIndexed addressing.
//
//	0001:A9 7F     LDA #$7F         ; Load A with 7F.
//	0003:A2 FF     LDX #$FF         ; Load X with FF.
//	0005:95 00     STA $00,X        ; Store A into $0000 offset with X.
//	0007:CA        DEX              ; Decrement X.
//	0008:30 FB     BPL $0005        ; Goto $0005 while X is not -1.
//
// The following opcodes support ZeroPageIndexedX addressing:
//
//	ADC    AND    ASL    CMP    DEC    EOR    INC    LDA    LDY    LSR    ORA
//	ROL    ROR    SBC    STA    STY
//
// The following opcodes support ZeroPageIndexedY addressing:
//
//	LDX    STX
//
//
//
// Indirect Addressing
//
// Indirect addressing reads a memory location from a 2-byte pointer.
//
// This particular addressing mode is used only by a special type of JMP.
// A pointer is first stored in memory, and read for the address rather than
// using an absolute position.
//
// There are two other indirect addressing modes, PreIndexedIndirect and
// PostIndexedIndirect.
//
// Parenthesis are used to signify that the opcode is reading a pointer rather
// than an absolute value.
//
// This example compares the usual absolute addressing of the JMP opcode to the
// Indirect addressing mode.
//
//	0001:4C 04 00  JMP $0004       ; This example moves the program counter to $0004
//	                               ; and resumes execution. This is ABSOLUTE addressing.
//
//	0004:A9 05     LDA #$22        ; The first four lines create a pointer value of
//	0006:85 00     STA $20         ; $1122 and store it into $0020-$0021.
//	0008:A9 20     LDA #$11
//	000A:85 01     STA $21
//	000C:6C 20 00  JMP ($0020)     ; Reads the location found in $0020-$0021, which
//	                               ; we just set to $1122, then moves the program
//	                               ; counter to the location. This is INDIRECT addressing.
//	                               ; Like all addresses, the pointer uses little-endian.
//
// The following opcodes support Indirect addressing:
//
//	JMP
//
//
//
// PreIndexedIndirect Addressing
//
// This mode accepts a zero-page address and adds the contents of the X register
// to get an address. The address is expected to contain a 2-byte pointer to a
// memory address (ordered in little-endian).
//
// The indirection is indicated by parenthesis in assembly language.
//
// This instruction is a three step process.
//
//	1) Sum the operand and the X register to get the address to read from.
//	2) Read the 2-byte address.
//	3) Return the value found in the address.
//
// Keep in mind that this only operates on the zero page. The processor uses
// wrap-around addition to sum up the operand and the X register.
// Thus, if you use an operand of $FF and an X register of $10, the result is
// $0F, not $110.
//
// This example loads the memory with a pointer and sets the X register to read
// the value from the pointer's address.
//
//	0001:A9 05     LDA #$05        ; The first four lines create a pointer value of
//	0003:85 00     STA $46         ; $2005 and store it into $0046-$0047.
//	0005:A9 20     LDA #$20
//	0007:85 01     STA $47
//	0009:A9 FF     LDA #$FF        ; Load A with #FF.
//	0011:8D 05 20  STA $2005       ; Store A into address $2005.
//	0014:A2 05     LDX #06         ; Load the X register with #06.
//	0016:A1 40     LDA ($40, X)    ; Read the memory address located at $40 + X ($46).
//	                               ; The address there is $2005, which contains the value #FF,
//	                               ; So A will be loaded with #FF.
//
// The following opcodes support PreIndexedIndirect addressing:
//
//	ADC    AND    CMP    EOR    LDA    ORA    SBC    STA
//
//
//
// PostIndexedIndirect Addressing
//
// This mode accepts an adress and adds the Y register after reading from
// memory. The address is expected to contain a 2-byte pointer to a memory
// address (ordered in little-endian).
//
// The indirection is indicated by parenthesis in assembly language.
// Notice that, unlike PreIndexedIndirect, they don't encompass the Y,
// signifying that the addition occurs after the read.
//
// This instruction is a three step process.
//
//	1) Read the 2-byte address based on the operand.
//	2) Sum the address and the Y register to get the offset address.
//	3) Return the value found in the offset address.
//
// Note: I'm not sure what happens when you use #FF as you operand.
// I would assume, since this is a zero-page operation, that the next byte would
// be read from #00, but it may also be read from #100.
//
// This example loads the memory with a pointer and sets the Y register to read
// the value from the pointer's address.
//
//	0001:A9 05     LDA #$15        ; The first four lines create a pointer value of
//	0003:85 00     STA $46         ; $3215 and store it into $0046-$0047.
//	0005:A9 20     LDA #$32
//	0007:85 01     STA $47
//	0009:A9 FF     LDA #$FF        ; Load A with #FF.
//	0011:8D 15 32  STA $3219       ; Store A into address $3219.
//	0014:A2 05     LDY #04         ; Load the Y register with #04.
//	0016:B1 40     LDA ($46), Y    ; Read the memory address located at $46.
//	                               ; The address there is $3215. Then, add the Y register to it, which gives us $3219.
//	                               ; So A will be loaded with #FF.
//
// The following opcodes support PostIndexedIndirect addressing:
//
// 	ADC    AND    CMP    EOR    LDA    ORA    SBC    STA
//
type addressingMode byte

const (
	// Immediate adressing is used when the operand's 1-byte value is given in
	// the instruction itself.
	immediate addressingMode = iota

	// ZeroPage adressing, requires a 1-byte address and can only access the
	// zero-page range ($0000-$00FF).
	zeroPage

	// Absolute addressing requires a full 2-byte address and can access the
	// full range ($0000-$FFFF).
	absolute

	// Relative addressing is used on the various Branch-On-true
	// instructions.
	//
	// A 1-byte signed operand is added to the program counter, and the program
	// continues execution from the new address. Because this value is signed,
	// values #00-#7F are positive, and values #FF-#80 are negative.
	relative

	// Implied addressing occurs when there is no operand. The addressing mode
	// is implied by the instruction.
	implied

	// Accumulator addressing is a special type of Implied addressing that only
	// addresses the accumulator.
	accumulator

	// IndexedX addressing works like Absolute but uses the X register as an
	// offset.
	//
	// This mode takes an extra cycle for write instructions (or for page
	// wrapping on read instructions) called the "oops" cycle.
	// Read the AddressingModes type docs for more info.
	indexedX

	// IndexedY addressing works like Absolute but uses the Y register as an
	// offset.
	//
	// This mode takes an extra cycle for write instructions (or for page
	// wrapping on read instructions) called the "oops" cycle.
	// Read the AddressingModes type docs for more info.
	indexedY

	// ZeroPageIndexedX addressing works like ZeroPage but uses the X register
	// as an offset.
	zeroPageIndexedX

	// ZeroPageIndexedY addressing works like ZeroPage but uses the Y register
	// as an offset.
	zeroPageIndexedY

	// Indirect addressing reads a memory location from a two-byte pointer.
	indirect

	// PreIndexedIndirect addressing accepts a zero-page address and adds the
	// contents of the X register to get an address.
	//
	// The address is expected to contain a 2-byte pointer to a memory address
	// (ordered in little-endian).
	preIndexedIndirect

	// PostIndexedIndirect accepts an address and adds the Y register after
	// reading from memory.
	//
	// The address is expected to contain a 2-byte pointer to a memory address
	// (ordered in little-endian).
	postIndexedIndirect
)

type instructionKind byte

const (
	_ instructionKind = iota
	read
	write
	readModWrite
)

type instruction struct {
	opCode     byte
	name       string
	mode       addressingMode
	kind       instructionKind
	size       byte
	cycles     byte
	pageCycles byte
	illegal    bool
}

var instructions = [256]instruction{
	instruction{opCode: 0x00, name: "BRK", size: 2, cycles: 7, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x01, name: "ORA", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x02, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x03, name: "SLO", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x04, name: "NOP", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: true},
	instruction{opCode: 0x05, name: "ORA", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0x06, name: "ASL", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0x07, name: "SLO", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0x08, name: "PHP", size: 1, cycles: 3, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x09, name: "ORA", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0x0A, name: "ASL", size: 1, cycles: 2, pageCycles: 0, mode: accumulator, kind: readModWrite, illegal: false},
	instruction{opCode: 0x0B, name: "ANC", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0x0C, name: "NOP", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: true},
	instruction{opCode: 0x0D, name: "ORA", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0x0E, name: "ASL", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0x0F, name: "SLO", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0x10, name: "BPL", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0x11, name: "ORA", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x12, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x13, name: "SLO", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x14, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0x15, name: "ORA", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0x16, name: "ASL", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x17, name: "SLO", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x18, name: "CLC", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x19, name: "ORA", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0x1A, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0x1B, name: "SLO", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0x1C, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0x1D, name: "ORA", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0x1E, name: "ASL", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x1F, name: "SLO", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x20, name: "JSR", size: 3, cycles: 6, pageCycles: 0, mode: absolute, illegal: false},
	instruction{opCode: 0x21, name: "AND", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x22, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x23, name: "RLA", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x24, name: "BIT", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0x25, name: "AND", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0x26, name: "ROL", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0x27, name: "RLA", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0x28, name: "PLP", size: 1, cycles: 4, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x29, name: "AND", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0x2A, name: "ROL", size: 1, cycles: 2, pageCycles: 0, mode: accumulator, kind: readModWrite, illegal: false},
	instruction{opCode: 0x2B, name: "ANC", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0x2C, name: "BIT", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0x2D, name: "AND", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0x2E, name: "ROL", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0x2F, name: "RLA", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0x30, name: "BMI", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0x31, name: "AND", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x32, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x33, name: "RLA", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x34, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0x35, name: "AND", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0x36, name: "ROL", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x37, name: "RLA", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x38, name: "SEC", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x39, name: "AND", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0x3A, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0x3B, name: "RLA", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0x3C, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0x3D, name: "AND", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0x3E, name: "ROL", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x3F, name: "RLA", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x40, name: "RTI", size: 1, cycles: 6, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x41, name: "EOR", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x42, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x43, name: "SRE", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x44, name: "NOP", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: true},
	instruction{opCode: 0x45, name: "EOR", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0x46, name: "LSR", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0x47, name: "SRE", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0x48, name: "PHA", size: 1, cycles: 3, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x49, name: "EOR", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0x4A, name: "LSR", size: 1, cycles: 2, pageCycles: 0, mode: accumulator, kind: readModWrite, illegal: false},
	instruction{opCode: 0x4B, name: "ALR", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0x4C, name: "JMP", size: 3, cycles: 3, pageCycles: 0, mode: absolute, illegal: false},
	instruction{opCode: 0x4D, name: "EOR", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0x4E, name: "LSR", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0x4F, name: "SRE", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0x50, name: "BVC", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0x51, name: "EOR", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x52, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x53, name: "SRE", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x54, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0x55, name: "EOR", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0x56, name: "LSR", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x57, name: "SRE", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x58, name: "CLI", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x59, name: "EOR", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0x5A, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0x5B, name: "SRE", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0x5C, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0x5D, name: "EOR", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0x5E, name: "LSR", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x5F, name: "SRE", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x60, name: "RTS", size: 1, cycles: 6, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x61, name: "ADC", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x62, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x63, name: "RRA", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x64, name: "NOP", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: true},
	instruction{opCode: 0x65, name: "ADC", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0x66, name: "ROR", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0x67, name: "RRA", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0x68, name: "PLA", size: 1, cycles: 4, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x69, name: "ADC", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0x6A, name: "ROR", size: 1, cycles: 2, pageCycles: 0, mode: accumulator, kind: readModWrite, illegal: false},
	instruction{opCode: 0x6B, name: "ARR", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0x6C, name: "JMP", size: 3, cycles: 5, pageCycles: 0, mode: indirect, illegal: false},
	instruction{opCode: 0x6D, name: "ADC", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0x6E, name: "ROR", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0x6F, name: "RRA", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0x70, name: "BVS", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0x71, name: "ADC", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0x72, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x73, name: "RRA", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0x74, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0x75, name: "ADC", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0x76, name: "ROR", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x77, name: "RRA", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x78, name: "SEI", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x79, name: "ADC", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0x7A, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0x7B, name: "RRA", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0x7C, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0x7D, name: "ADC", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0x7E, name: "ROR", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0x7F, name: "RRA", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0x80, name: "NOP", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0x81, name: "STA", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: write, illegal: false},
	instruction{opCode: 0x82, name: "NOP", size: 0, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0x83, name: "SAX", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: write, illegal: true},
	instruction{opCode: 0x84, name: "STY", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: write, illegal: false},
	instruction{opCode: 0x85, name: "STA", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: write, illegal: false},
	instruction{opCode: 0x86, name: "STX", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: write, illegal: false},
	instruction{opCode: 0x87, name: "SAX", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: write, illegal: true},
	instruction{opCode: 0x88, name: "DEY", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x89, name: "NOP", size: 0, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0x8A, name: "TXA", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x8B, name: "XAA", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0x8C, name: "STY", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: write, illegal: false},
	instruction{opCode: 0x8D, name: "STA", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: write, illegal: false},
	instruction{opCode: 0x8E, name: "STX", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: write, illegal: false},
	instruction{opCode: 0x8F, name: "SAX", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: write, illegal: true},
	instruction{opCode: 0x90, name: "BCC", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0x91, name: "STA", size: 2, cycles: 6, pageCycles: 0, mode: postIndexedIndirect, kind: write, illegal: false},
	instruction{opCode: 0x92, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0x93, name: "AHX", size: 0, cycles: 6, pageCycles: 0, mode: postIndexedIndirect, illegal: true},
	instruction{opCode: 0x94, name: "STY", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: write, illegal: false},
	instruction{opCode: 0x95, name: "STA", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: write, illegal: false},
	instruction{opCode: 0x96, name: "STX", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedY, kind: write, illegal: false},
	instruction{opCode: 0x97, name: "SAX", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedY, kind: write, illegal: true},
	instruction{opCode: 0x98, name: "TYA", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x99, name: "STA", size: 3, cycles: 5, pageCycles: 0, mode: indexedY, kind: write, illegal: false},
	instruction{opCode: 0x9A, name: "TXS", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0x9B, name: "TAS", size: 0, cycles: 5, pageCycles: 0, mode: indexedY, illegal: true},
	instruction{opCode: 0x9C, name: "SHY", size: 0, cycles: 5, pageCycles: 0, mode: indexedX, kind: write, illegal: true},
	instruction{opCode: 0x9D, name: "STA", size: 3, cycles: 5, pageCycles: 0, mode: indexedX, kind: write, illegal: false},
	instruction{opCode: 0x9E, name: "SHX", size: 0, cycles: 5, pageCycles: 0, mode: indexedY, kind: write, illegal: true},
	instruction{opCode: 0x9F, name: "AHX", size: 0, cycles: 5, pageCycles: 0, mode: indexedY, illegal: true},
	instruction{opCode: 0xA0, name: "LDY", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0xA1, name: "LDA", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xA2, name: "LDX", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0xA3, name: "LAX", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: true},
	instruction{opCode: 0xA4, name: "LDY", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0xA5, name: "LDA", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0xA6, name: "LDX", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0xA7, name: "LAX", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: true},
	instruction{opCode: 0xA8, name: "TAY", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xA9, name: "LDA", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0xAA, name: "TAX", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xAB, name: "LAX", size: 0, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0xAC, name: "LDY", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0xAD, name: "LDA", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0xAE, name: "LDX", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0xAF, name: "LAX", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: true},
	instruction{opCode: 0xB0, name: "BCS", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0xB1, name: "LDA", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xB2, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0xB3, name: "LAX", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: true},
	instruction{opCode: 0xB4, name: "LDY", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0xB5, name: "LDA", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0xB6, name: "LDX", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedY, kind: read, illegal: false},
	instruction{opCode: 0xB7, name: "LAX", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedY, kind: read, illegal: true},
	instruction{opCode: 0xB8, name: "CLV", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xB9, name: "LDA", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0xBA, name: "TSX", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xBB, name: "LAS", size: 0, cycles: 4, pageCycles: 1, mode: indexedY, illegal: true},
	instruction{opCode: 0xBC, name: "LDY", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0xBD, name: "LDA", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0xBE, name: "LDX", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0xBF, name: "LAX", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: true},
	instruction{opCode: 0xC0, name: "CPY", size: 2, cycles: 2, pageCycles: 0, mode: immediate, illegal: false},
	instruction{opCode: 0xC1, name: "CMP", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xC2, name: "NOP", size: 0, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0xC3, name: "DCP", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0xC4, name: "CPY", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, illegal: false},
	instruction{opCode: 0xC5, name: "CMP", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0xC6, name: "DEC", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0xC7, name: "DCP", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0xC8, name: "INY", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xC9, name: "CMP", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0xCA, name: "DEX", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xCB, name: "AXS", size: 0, cycles: 2, pageCycles: 0, mode: immediate, illegal: true},
	instruction{opCode: 0xCC, name: "CPY", size: 3, cycles: 4, pageCycles: 0, mode: absolute, illegal: false},
	instruction{opCode: 0xCD, name: "CMP", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0xCE, name: "DEC", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0xCF, name: "DCP", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0xD0, name: "BNE", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0xD1, name: "CMP", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xD2, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0xD3, name: "DCP", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0xD4, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0xD5, name: "CMP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0xD6, name: "DEC", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0xD7, name: "DCP", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0xD8, name: "CLD", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xD9, name: "CMP", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0xDA, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0xDB, name: "DCP", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0xDC, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0xDD, name: "CMP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0xDE, name: "DEC", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0xDF, name: "DCP", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0xE0, name: "CPX", size: 2, cycles: 2, pageCycles: 0, mode: immediate, illegal: false},
	instruction{opCode: 0xE1, name: "SBC", size: 2, cycles: 6, pageCycles: 0, mode: preIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xE2, name: "NOP", size: 0, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0xE3, name: "ISB", size: 2, cycles: 8, pageCycles: 0, mode: preIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0xE4, name: "CPX", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, illegal: false},
	instruction{opCode: 0xE5, name: "SBC", size: 2, cycles: 3, pageCycles: 0, mode: zeroPage, kind: read, illegal: false},
	instruction{opCode: 0xE6, name: "INC", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: false},
	instruction{opCode: 0xE7, name: "ISB", size: 2, cycles: 5, pageCycles: 0, mode: zeroPage, kind: readModWrite, illegal: true},
	instruction{opCode: 0xE8, name: "INX", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xE9, name: "SBC", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: false},
	instruction{opCode: 0xEA, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: false},
	instruction{opCode: 0xEB, name: "SBC", size: 2, cycles: 2, pageCycles: 0, mode: immediate, kind: read, illegal: true},
	instruction{opCode: 0xEC, name: "CPX", size: 3, cycles: 4, pageCycles: 0, mode: absolute, illegal: false},
	instruction{opCode: 0xED, name: "SBC", size: 3, cycles: 4, pageCycles: 0, mode: absolute, kind: read, illegal: false},
	instruction{opCode: 0xEE, name: "INC", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: false},
	instruction{opCode: 0xEF, name: "ISB", size: 3, cycles: 6, pageCycles: 0, mode: absolute, kind: readModWrite, illegal: true},
	instruction{opCode: 0xF0, name: "BEQ", size: 2, cycles: 2, pageCycles: 1, mode: relative, illegal: false},
	instruction{opCode: 0xF1, name: "SBC", size: 2, cycles: 5, pageCycles: 1, mode: postIndexedIndirect, kind: read, illegal: false},
	instruction{opCode: 0xF2, name: "KIL", size: 0, cycles: 2, pageCycles: 0, mode: implied, illegal: true},
	instruction{opCode: 0xF3, name: "ISB", size: 2, cycles: 8, pageCycles: 0, mode: postIndexedIndirect, kind: readModWrite, illegal: true},
	instruction{opCode: 0xF4, name: "NOP", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: true},
	instruction{opCode: 0xF5, name: "SBC", size: 2, cycles: 4, pageCycles: 0, mode: zeroPageIndexedX, kind: read, illegal: false},
	instruction{opCode: 0xF6, name: "INC", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0xF7, name: "ISB", size: 2, cycles: 6, pageCycles: 0, mode: zeroPageIndexedX, kind: readModWrite, illegal: true},
	instruction{opCode: 0xF8, name: "SED", size: 1, cycles: 2, pageCycles: 0, mode: implied, illegal: false},
	instruction{opCode: 0xF9, name: "SBC", size: 3, cycles: 4, pageCycles: 1, mode: indexedY, kind: read, illegal: false},
	instruction{opCode: 0xFA, name: "NOP", size: 1, cycles: 2, pageCycles: 0, mode: implied, kind: read, illegal: true},
	instruction{opCode: 0xFB, name: "ISB", size: 3, cycles: 7, pageCycles: 0, mode: indexedY, kind: readModWrite, illegal: true},
	instruction{opCode: 0xFC, name: "NOP", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: true},
	instruction{opCode: 0xFD, name: "SBC", size: 3, cycles: 4, pageCycles: 1, mode: indexedX, kind: read, illegal: false},
	instruction{opCode: 0xFE, name: "INC", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: false},
	instruction{opCode: 0xFF, name: "ISB", size: 3, cycles: 7, pageCycles: 0, mode: indexedX, kind: readModWrite, illegal: true},
}
