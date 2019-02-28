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
type AddressingMode byte

const (
	// Immediate adressing is used when the operand's 1-byte value is given in
	// the instruction itself.
	Immediate AddressingMode = iota

	// ZeroPage adressing, requires a 1-byte address and can only access the
	// zero-page range ($0000-$00FF).
	ZeroPage

	// Absolute addressing requires a full 2-byte address and can access the
	// full range ($0000-$FFFF).
	Absolute

	// Relative addressing is used on the various Branch-On-true
	// instructions.
	//
	// A 1-byte signed operand is added to the program counter, and the program
	// continues execution from the new address. Because this value is signed,
	// values #00-#7F are positive, and values #FF-#80 are negative.
	Relative

	// Implied addressing occurs when there is no operand. The addressing mode
	// is implied by the instruction.
	Implied

	// Accumulator addressing is a special type of Implied addressing that only
	// addresses the accumulator.
	Accumulator

	// IndexedX addressing works like Absolute but uses the X register as an
	// offset.
	//
	// This mode takes an extra cycle for write instructions (or for page
	// wrapping on read instructions) called the "oops" cycle.
	// Read the AddressingModes type docs for more info.
	IndexedX

	// IndexedY addressing works like Absolute but uses the Y register as an
	// offset.
	//
	// This mode takes an extra cycle for write instructions (or for page
	// wrapping on read instructions) called the "oops" cycle.
	// Read the AddressingModes type docs for more info.
	IndexedY

	// ZeroPageIndexedX addressing works like ZeroPage but uses the X register
	// as an offset.
	ZeroPageIndexedX

	// ZeroPageIndexedY addressing works like ZeroPage but uses the Y register
	// as an offset.
	ZeroPageIndexedY

	// Indirect addressing reads a memory location from a two-byte pointer.
	Indirect

	// PreIndexedIndirect addressing accepts a zero-page address and adds the
	// contents of the X register to get an address.
	//
	// The address is expected to contain a 2-byte pointer to a memory address
	// (ordered in little-endian).
	PreIndexedIndirect

	// PostIndexedIndirect accepts an address and adds the Y register after
	// reading from memory.
	//
	// The address is expected to contain a 2-byte pointer to a memory address
	// (ordered in little-endian).
	PostIndexedIndirect
)

type InstructionKind byte

const (
	_ InstructionKind = iota
	Read
	Write
	ReadModWrite
)

type Instruction struct {
	OpCode     byte
	Name       string
	Mode       AddressingMode
	Kind       InstructionKind
	Size       byte
	Cycles     byte
	PageCycles byte
	Illegal    bool
}

var instructions = [256]Instruction{
	Instruction{OpCode: 0x00, Name: "BRK", Size: 2, Cycles: 7, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x01, Name: "ORA", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x02, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x03, Name: "SLO", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x04, Name: "NOP", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x05, Name: "ORA", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x06, Name: "ASL", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x07, Name: "SLO", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x08, Name: "PHP", Size: 1, Cycles: 3, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x09, Name: "ORA", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x0A, Name: "ASL", Size: 1, Cycles: 2, PageCycles: 0, Mode: Accumulator, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x0B, Name: "ANC", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0x0C, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x0D, Name: "ORA", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x0E, Name: "ASL", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x0F, Name: "SLO", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x10, Name: "BPL", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0x11, Name: "ORA", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x12, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x13, Name: "SLO", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x14, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x15, Name: "ORA", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x16, Name: "ASL", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x17, Name: "SLO", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x18, Name: "CLC", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x19, Name: "ORA", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x1A, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x1B, Name: "SLO", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x1C, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x1D, Name: "ORA", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x1E, Name: "ASL", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x1F, Name: "SLO", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x20, Name: "JSR", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Illegal: false},
	Instruction{OpCode: 0x21, Name: "AND", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x22, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x23, Name: "RLA", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x24, Name: "BIT", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x25, Name: "AND", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x26, Name: "ROL", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x27, Name: "RLA", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x28, Name: "PLP", Size: 1, Cycles: 4, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x29, Name: "AND", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x2A, Name: "ROL", Size: 1, Cycles: 2, PageCycles: 0, Mode: Accumulator, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x2B, Name: "ANC", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0x2C, Name: "BIT", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x2D, Name: "AND", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x2E, Name: "ROL", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x2F, Name: "RLA", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x30, Name: "BMI", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0x31, Name: "AND", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x32, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x33, Name: "RLA", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x34, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x35, Name: "AND", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x36, Name: "ROL", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x37, Name: "RLA", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x38, Name: "SEC", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x39, Name: "AND", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x3A, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x3B, Name: "RLA", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x3C, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x3D, Name: "AND", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x3E, Name: "ROL", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x3F, Name: "RLA", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x40, Name: "RTI", Size: 1, Cycles: 6, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x41, Name: "EOR", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x42, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x43, Name: "SRE", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x44, Name: "NOP", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x45, Name: "EOR", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x46, Name: "LSR", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x47, Name: "SRE", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x48, Name: "PHA", Size: 1, Cycles: 3, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x49, Name: "EOR", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x4A, Name: "LSR", Size: 1, Cycles: 2, PageCycles: 0, Mode: Accumulator, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x4B, Name: "ALR", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0x4C, Name: "JMP", Size: 3, Cycles: 3, PageCycles: 0, Mode: Absolute, Illegal: false},
	Instruction{OpCode: 0x4D, Name: "EOR", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x4E, Name: "LSR", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x4F, Name: "SRE", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x50, Name: "BVC", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0x51, Name: "EOR", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x52, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x53, Name: "SRE", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x54, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x55, Name: "EOR", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x56, Name: "LSR", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x57, Name: "SRE", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x58, Name: "CLI", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x59, Name: "EOR", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x5A, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x5B, Name: "SRE", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x5C, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x5D, Name: "EOR", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x5E, Name: "LSR", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x5F, Name: "SRE", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x60, Name: "RTS", Size: 1, Cycles: 6, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x61, Name: "ADC", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x62, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x63, Name: "RRA", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x64, Name: "NOP", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x65, Name: "ADC", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x66, Name: "ROR", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x67, Name: "RRA", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x68, Name: "PLA", Size: 1, Cycles: 4, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x69, Name: "ADC", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x6A, Name: "ROR", Size: 1, Cycles: 2, PageCycles: 0, Mode: Accumulator, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x6B, Name: "ARR", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0x6C, Name: "JMP", Size: 3, Cycles: 5, PageCycles: 0, Mode: Indirect, Illegal: false},
	Instruction{OpCode: 0x6D, Name: "ADC", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x6E, Name: "ROR", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x6F, Name: "RRA", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x70, Name: "BVS", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0x71, Name: "ADC", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x72, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x73, Name: "RRA", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x74, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x75, Name: "ADC", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x76, Name: "ROR", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x77, Name: "RRA", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x78, Name: "SEI", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x79, Name: "ADC", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x7A, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x7B, Name: "RRA", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x7C, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x7D, Name: "ADC", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0x7E, Name: "ROR", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0x7F, Name: "RRA", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0x80, Name: "NOP", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x81, Name: "STA", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x82, Name: "NOP", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x83, Name: "SAX", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x84, Name: "STY", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x85, Name: "STA", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x86, Name: "STX", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x87, Name: "SAX", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x88, Name: "DEY", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x89, Name: "NOP", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0x8A, Name: "TXA", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x8B, Name: "XAA", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0x8C, Name: "STY", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x8D, Name: "STA", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x8E, Name: "STX", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x8F, Name: "SAX", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x90, Name: "BCC", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0x91, Name: "STA", Size: 2, Cycles: 6, PageCycles: 0, Mode: PostIndexedIndirect, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x92, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0x93, Name: "AHX", Size: 0, Cycles: 6, PageCycles: 0, Mode: PostIndexedIndirect, Illegal: true},
	Instruction{OpCode: 0x94, Name: "STY", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x95, Name: "STA", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x96, Name: "STX", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedY, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x97, Name: "SAX", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedY, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x98, Name: "TYA", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x99, Name: "STA", Size: 3, Cycles: 5, PageCycles: 0, Mode: IndexedY, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x9A, Name: "TXS", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0x9B, Name: "TAS", Size: 0, Cycles: 5, PageCycles: 0, Mode: IndexedY, Illegal: true},
	Instruction{OpCode: 0x9C, Name: "SHY", Size: 0, Cycles: 5, PageCycles: 0, Mode: IndexedX, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x9D, Name: "STA", Size: 3, Cycles: 5, PageCycles: 0, Mode: IndexedX, Kind: Write, Illegal: false},
	Instruction{OpCode: 0x9E, Name: "SHX", Size: 0, Cycles: 5, PageCycles: 0, Mode: IndexedY, Kind: Write, Illegal: true},
	Instruction{OpCode: 0x9F, Name: "AHX", Size: 0, Cycles: 5, PageCycles: 0, Mode: IndexedY, Illegal: true},
	Instruction{OpCode: 0xA0, Name: "LDY", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA1, Name: "LDA", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA2, Name: "LDX", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA3, Name: "LAX", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xA4, Name: "LDY", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA5, Name: "LDA", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA6, Name: "LDX", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xA7, Name: "LAX", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xA8, Name: "TAY", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xA9, Name: "LDA", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xAA, Name: "TAX", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xAB, Name: "LAX", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xAC, Name: "LDY", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xAD, Name: "LDA", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xAE, Name: "LDX", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xAF, Name: "LAX", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xB0, Name: "BCS", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0xB1, Name: "LDA", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xB2, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0xB3, Name: "LAX", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xB4, Name: "LDY", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xB5, Name: "LDA", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xB6, Name: "LDX", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xB7, Name: "LAX", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedY, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xB8, Name: "CLV", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xB9, Name: "LDA", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xBA, Name: "TSX", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xBB, Name: "LAS", Size: 0, Cycles: 4, PageCycles: 1, Mode: IndexedY, Illegal: true},
	Instruction{OpCode: 0xBC, Name: "LDY", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xBD, Name: "LDA", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xBE, Name: "LDX", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xBF, Name: "LAX", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xC0, Name: "CPY", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: false},
	Instruction{OpCode: 0xC1, Name: "CMP", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xC2, Name: "NOP", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xC3, Name: "DCP", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xC4, Name: "CPY", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Illegal: false},
	Instruction{OpCode: 0xC5, Name: "CMP", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xC6, Name: "DEC", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xC7, Name: "DCP", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xC8, Name: "INY", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xC9, Name: "CMP", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xCA, Name: "DEX", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xCB, Name: "AXS", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: true},
	Instruction{OpCode: 0xCC, Name: "CPY", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Illegal: false},
	Instruction{OpCode: 0xCD, Name: "CMP", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xCE, Name: "DEC", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xCF, Name: "DCP", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xD0, Name: "BNE", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0xD1, Name: "CMP", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xD2, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0xD3, Name: "DCP", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xD4, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xD5, Name: "CMP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xD6, Name: "DEC", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xD7, Name: "DCP", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xD8, Name: "CLD", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xD9, Name: "CMP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xDA, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xDB, Name: "DCP", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xDC, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xDD, Name: "CMP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xDE, Name: "DEC", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xDF, Name: "DCP", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xE0, Name: "CPX", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Illegal: false},
	Instruction{OpCode: 0xE1, Name: "SBC", Size: 2, Cycles: 6, PageCycles: 0, Mode: PreIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xE2, Name: "NOP", Size: 0, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xE3, Name: "ISB", Size: 2, Cycles: 8, PageCycles: 0, Mode: PreIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xE4, Name: "CPX", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Illegal: false},
	Instruction{OpCode: 0xE5, Name: "SBC", Size: 2, Cycles: 3, PageCycles: 0, Mode: ZeroPage, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xE6, Name: "INC", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xE7, Name: "ISB", Size: 2, Cycles: 5, PageCycles: 0, Mode: ZeroPage, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xE8, Name: "INX", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xE9, Name: "SBC", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xEA, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xEB, Name: "SBC", Size: 2, Cycles: 2, PageCycles: 0, Mode: Immediate, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xEC, Name: "CPX", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Illegal: false},
	Instruction{OpCode: 0xED, Name: "SBC", Size: 3, Cycles: 4, PageCycles: 0, Mode: Absolute, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xEE, Name: "INC", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xEF, Name: "ISB", Size: 3, Cycles: 6, PageCycles: 0, Mode: Absolute, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xF0, Name: "BEQ", Size: 2, Cycles: 2, PageCycles: 1, Mode: Relative, Illegal: false},
	Instruction{OpCode: 0xF1, Name: "SBC", Size: 2, Cycles: 5, PageCycles: 1, Mode: PostIndexedIndirect, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xF2, Name: "KIL", Size: 0, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: true},
	Instruction{OpCode: 0xF3, Name: "ISB", Size: 2, Cycles: 8, PageCycles: 0, Mode: PostIndexedIndirect, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xF4, Name: "NOP", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xF5, Name: "SBC", Size: 2, Cycles: 4, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xF6, Name: "INC", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xF7, Name: "ISB", Size: 2, Cycles: 6, PageCycles: 0, Mode: ZeroPageIndexedX, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xF8, Name: "SED", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Illegal: false},
	Instruction{OpCode: 0xF9, Name: "SBC", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedY, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xFA, Name: "NOP", Size: 1, Cycles: 2, PageCycles: 0, Mode: Implied, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xFB, Name: "ISB", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedY, Kind: ReadModWrite, Illegal: true},
	Instruction{OpCode: 0xFC, Name: "NOP", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: true},
	Instruction{OpCode: 0xFD, Name: "SBC", Size: 3, Cycles: 4, PageCycles: 1, Mode: IndexedX, Kind: Read, Illegal: false},
	Instruction{OpCode: 0xFE, Name: "INC", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: false},
	Instruction{OpCode: 0xFF, Name: "ISB", Size: 3, Cycles: 7, PageCycles: 0, Mode: IndexedX, Kind: ReadModWrite, Illegal: true},
}
