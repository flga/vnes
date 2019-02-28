package nes

import (
	"io"
)

type Interrupt byte

const (
	None Interrupt = iota
	NMI
	IRQ
)

const (
	NMIAddr     = uint16(0xFFFA)
	ResetAddr   = uint16(0xFFFC)
	IRQ_BRKAddr = uint16(0xFFFE)

	stackHi = 0x0100
)

// Status are all the flags that represent the processor status.
type Status byte

const (
	// Carry flag.
	//
	// After ADC, this is the carry result of the addition.
	// After SBC or CMP, this flag will be set if no borrow was the result, or
	// alternatively a "greater than or equal" result.
	// After a shift instruction (ASL, LSR, ROL, ROR), this contains the bit
	// that was shifted out.
	//
	// Increment and decrement instructions do not affect the carry flag.
	// Can be set or cleared directly with SEC, CLC.
	Carry Status = 1 << iota

	// Zero flag is set when the result of an instruction is zero.
	Zero

	// InterruptDisable flag.
	//
	// When set, all interrupts except the NMI are inhibited.
	// Can be set or cleared directly with SEI, CLI.
	// Automatically set by the CPU when an IRQ is triggered, and restored
	// to its previous state by RTI.
	//
	// If the /IRQ line is low (IRQ pending) when this flag is cleared, an
	// interrupt will immediately be triggered.
	InterruptDisable

	// Decimal flag. On the NES, this flag has no effect.
	Decimal

	// Break flag.
	//
	// While there are only six flags in the processor status register within
	// the CPU, when transferred to the stack, there are two additional bits.
	//
	// These do not represent a register that can hold a value but can be used
	// to distinguish how the flags were pushed.
	//
	// Some 6502 references call this the "B flag", though it does not represent
	// an actual CPU register.
	//
	// Two interrupts (/IRQ and /NMI) and two instructions (PHP and BRK) push
	// the flags to the stack.
	//
	// In the byte pushed, Break is 1 if from an instruction (PHP or BRK) or 0
	// if from an interrupt line being pulled low (/IRQ or /NMI).
	//
	// Two instructions (PLP and RTI) pull a byte from the stack and set all the
	// flags. They ignore Unused and Break.
	//
	// The only way for an IRQ handler to distinguish /IRQ from BRK is to read
	// the flags byte from the stack and test Break.
	Break

	// Unused flag.
	Unused

	// Overflow flag.
	//
	// ADC, SBC, and CMP will set this flag if the signed result would be
	// invalid http://www.6502.org/tutorials/vflag.html, necessary for making
	// signed comparisons http://www.6502.org/tutorials/compare_beyond.html#5.
	//
	// BIT will load bit 6 of the addressed value directly into the V flag.
	// Can be cleared directly with CLV.
	// There is no corresponding set instruction.
	Overflow

	// Negative flag.
	//
	// After most instructions that have a value result, this flag will contain
	// bit 7 of that result.
	// BIT will load bit 7 of the addressed value directly into the N flag.
	Negative
)

type CPU struct {
	Cycles uint64

	// A, along with the arithmetic logic unit (ALU), supports using the status
	// register for carrying, overflow detection, and so on.
	A byte

	// X and Y are used for several addressing modes. They can be used as loop
	// counters easily, using INC/DEC and branch instructions.
	//
	// Not being the accumulator, they have limited addressing modes themselves
	// when loading and saving.
	X, Y byte

	// The program counter PC supports 65536 direct (unbanked) memory locations,
	// however not all values are sent to the cartridge.
	//
	// It can be accessed either by allowing CPU's internal fetch logic
	// increment the address bus, an interrupt (NMI, Reset, IRQ/BRQ), and using
	// the RTS/JMP/JSR/Branch instructions.
	PC uint16

	// The Stack Pointer can be accessed using interrupts, pulls, pushes, and
	// transfers.
	S byte

	// The Status Register has 6 bits used by the ALU but is byte-wide.
	// PHP, PLP, arithmetic, testing, and branch instructions can access this
	// register.
	//
	// See Status for more info.
	P Status

	debug     io.Writer
	interrupt Interrupt
}

func NewCPU(debug io.Writer) *CPU {
	return &CPU{
		debug: debug,
		P:     InterruptDisable | Unused,
		S:     0xFD,
		PC:    ResetAddr,
	}
}

func (c *CPU) Init(bus *SysBus) {
	c.PC = c.readAddress(bus, ResetAddr)
}

func (c *CPU) SetPC(pc uint16) {
	c.PC = pc
}

func (c *CPU) Reset(bus *SysBus) {
	c.P |= InterruptDisable
	c.S -= 3

	c.PC = c.readAddress(bus, ResetAddr)
}

func (c *CPU) Trigger(interrupt Interrupt) {
	if interrupt == IRQ && c.P&InterruptDisable > 0 {
		return
	}
	c.interrupt = interrupt
}

func (c *CPU) Execute(bus *SysBus, ppu *PPU) uint64 {
	oldCycles := c.Cycles

	c.handleInterrupts(bus)

	initialPc := c.PC

	opCode := c.read(bus, c.PC)
	c.PC++

	inst := instructions[opCode]
	intermediateAddr, addr := c.resolveAddress(bus, inst)

	if c.debug != nil {
		//TODO: rework disassembly/tracing
		disassemble(c.debug, bus, initialPc, c.A, c.X, c.Y, byte(c.P), c.S, inst, intermediateAddr, addr, oldCycles, ppu)
	}

	switch opCode {
	case 0x04, 0x0C, 0x14, 0x1A, 0x1C, 0x34, 0x3A, 0x3C, 0x44, 0x54, 0x5A,
		0x5C, 0x64, 0x74, 0x7A, 0x7C, 0x80, 0x82, 0x89, 0xC2, 0xD4, 0xDA,
		0xDC, 0xE2, 0xEA, 0xF4, 0xFA, 0xFC:
		c.NOP(bus, inst.Mode, addr)
	case 0x61, 0x65, 0x69, 0x6D, 0x71, 0x75, 0x79, 0x7D:
		c.ADC(bus, inst.Mode, addr)
	case 0x93, 0x9F:
		c.AHX(bus, inst.Mode, addr)
	case 0x4B:
		c.ALR(bus, inst.Mode, addr)
	case 0x0B, 0x2B:
		c.ANC(bus, inst.Mode, addr)
	case 0x21, 0x25, 0x29, 0x2D, 0x31, 0x35, 0x39, 0x3D:
		c.AND(bus, inst.Mode, addr)
	case 0x6B:
		c.ARR(bus, inst.Mode, addr)
	case 0x06, 0x0A, 0x0E, 0x16, 0x1E:
		c.ASL(bus, inst.Mode, addr)
	case 0xCB:
		c.AXS(bus, inst.Mode, addr)
	case 0x90:
		c.BCC(bus, inst.Mode, addr)
	case 0xB0:
		c.BCS(bus, inst.Mode, addr)
	case 0xF0:
		c.BEQ(bus, inst.Mode, addr)
	case 0x24, 0x2C:
		c.BIT(bus, inst.Mode, addr)
	case 0x30:
		c.BMI(bus, inst.Mode, addr)
	case 0xD0:
		c.BNE(bus, inst.Mode, addr)
	case 0x10:
		c.BPL(bus, inst.Mode, addr)
	case 0x00:
		c.BRK(bus, inst.Mode, addr)
	case 0x50:
		c.BVC(bus, inst.Mode, addr)
	case 0x70:
		c.BVS(bus, inst.Mode, addr)
	case 0x18:
		c.CLC(bus, inst.Mode, addr)
	case 0xD8:
		c.CLD(bus, inst.Mode, addr)
	case 0x58:
		c.CLI(bus, inst.Mode, addr)
	case 0xB8:
		c.CLV(bus, inst.Mode, addr)
	case 0xC1, 0xC5, 0xC9, 0xCD, 0xD1, 0xD5, 0xD9, 0xDD:
		c.CMP(bus, inst.Mode, addr)
	case 0xE0, 0xE4, 0xEC:
		c.CPX(bus, inst.Mode, addr)
	case 0xC0, 0xC4, 0xCC:
		c.CPY(bus, inst.Mode, addr)
	case 0xC3, 0xC7, 0xCF, 0xD3, 0xD7, 0xDB, 0xDF:
		c.DCP(bus, inst.Mode, addr)
	case 0xC6, 0xCE, 0xD6, 0xDE:
		c.DEC(bus, inst.Mode, addr)
	case 0xCA:
		c.DEX(bus, inst.Mode, addr)
	case 0x88:
		c.DEY(bus, inst.Mode, addr)
	case 0x41, 0x45, 0x49, 0x4D, 0x51, 0x55, 0x59, 0x5D:
		c.EOR(bus, inst.Mode, addr)
	case 0xE6, 0xEE, 0xF6, 0xFE:
		c.INC(bus, inst.Mode, addr)
	case 0xE8:
		c.INX(bus, inst.Mode, addr)
	case 0xC8:
		c.INY(bus, inst.Mode, addr)
	case 0xE3, 0xE7, 0xEF, 0xF3, 0xF7, 0xFB, 0xFF:
		c.ISC(bus, inst.Mode, addr)
	case 0x4C, 0x6C:
		c.JMP(bus, inst.Mode, addr)
	case 0x20:
		c.JSR(bus, inst.Mode, addr)
	case 0x02, 0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x92, 0xB2, 0xD2, 0xF2:
		c.KIL(bus, inst.Mode, addr)
	case 0xBB:
		c.LAS(bus, inst.Mode, addr)
	case 0xA3, 0xA7, 0xAB, 0xAF, 0xB3, 0xB7, 0xBF:
		c.LAX(bus, inst.Mode, addr)
	case 0xA1, 0xA5, 0xA9, 0xAD, 0xB1, 0xB5, 0xB9, 0xBD:
		c.LDA(bus, inst.Mode, addr)
	case 0xA2, 0xA6, 0xAE, 0xB6, 0xBE:
		c.LDX(bus, inst.Mode, addr)
	case 0xA0, 0xA4, 0xAC, 0xB4, 0xBC:
		c.LDY(bus, inst.Mode, addr)
	case 0x46, 0x4A, 0x4E, 0x56, 0x5E:
		c.LSR(bus, inst.Mode, addr)
	case 0x01, 0x05, 0x09, 0x0D, 0x11, 0x15, 0x19, 0x1D:
		c.ORA(bus, inst.Mode, addr)
	case 0x48:
		c.PHA(bus, inst.Mode, addr)
	case 0x08:
		c.PHP(bus, inst.Mode, addr)
	case 0x68:
		c.PLA(bus, inst.Mode, addr)
	case 0x28:
		c.PLP(bus, inst.Mode, addr)
	case 0x23, 0x27, 0x2F, 0x33, 0x37, 0x3B, 0x3F:
		c.RLA(bus, inst.Mode, addr)
	case 0x26, 0x2A, 0x2E, 0x36, 0x3E:
		c.ROL(bus, inst.Mode, addr)
	case 0x66, 0x6A, 0x6E, 0x76, 0x7E:
		c.ROR(bus, inst.Mode, addr)
	case 0x63, 0x67, 0x6F, 0x73, 0x77, 0x7B, 0x7F:
		c.RRA(bus, inst.Mode, addr)
	case 0x40:
		c.RTI(bus, inst.Mode, addr)
	case 0x60:
		c.RTS(bus, inst.Mode, addr)
	case 0x83, 0x87, 0x8F, 0x97:
		c.SAX(bus, inst.Mode, addr)
	case 0xE1, 0xE5, 0xE9, 0xEB, 0xED, 0xF1, 0xF5, 0xF9, 0xFD:
		c.SBC(bus, inst.Mode, addr)
	case 0x38:
		c.SEC(bus, inst.Mode, addr)
	case 0xF8:
		c.SED(bus, inst.Mode, addr)
	case 0x78:
		c.SEI(bus, inst.Mode, addr)
	case 0x9E:
		c.SHX(bus, inst.Mode, addr)
	case 0x9C:
		c.SHY(bus, inst.Mode, addr)
	case 0x03, 0x07, 0x0F, 0x13, 0x17, 0x1B, 0x1F:
		c.SLO(bus, inst.Mode, addr)
	case 0x43, 0x47, 0x4F, 0x53, 0x57, 0x5B, 0x5F:
		c.SRE(bus, inst.Mode, addr)
	case 0x81, 0x85, 0x8D, 0x91, 0x95, 0x99, 0x9D:
		c.STA(bus, inst.Mode, addr)
	case 0x86, 0x8E, 0x96:
		c.STX(bus, inst.Mode, addr)
	case 0x84, 0x8C, 0x94:
		c.STY(bus, inst.Mode, addr)
	case 0x9B:
		c.TAS(bus, inst.Mode, addr)
	case 0xAA:
		c.TAX(bus, inst.Mode, addr)
	case 0xA8:
		c.TAY(bus, inst.Mode, addr)
	case 0xBA:
		c.TSX(bus, inst.Mode, addr)
	case 0x8A:
		c.TXA(bus, inst.Mode, addr)
	case 0x9A:
		c.TXS(bus, inst.Mode, addr)
	case 0x98:
		c.TYA(bus, inst.Mode, addr)
	case 0x8B:
		c.XAA(bus, inst.Mode, addr)
	}

	return c.Cycles - oldCycles
}

func (c *CPU) clock() {
	c.Cycles++
}

func (c *CPU) read(bus *SysBus, address uint16) byte {
	c.clock()
	return bus.Read(address)
}

func (c *CPU) readAddress(bus *SysBus, address uint16) uint16 {
	c.clock()

	addr, _, _ := bus.ReadAddress(address)
	return addr
}

func (c *CPU) write(bus *SysBus, address uint16, value byte) {
	if address == OAMDMA {
		c.dmaTransfer(bus, value)
		return
	}

	c.clock()
	bus.Write(address, value)
}

func (c *CPU) dmaTransfer(bus *SysBus, address byte) {
	addr := uint16(address) << 8
	for i := 0; i < 256; i++ {
		c.clock()
		v := bus.Read(addr)

		c.clock()
		bus.Write(OAMDMA, v)

		addr++
	}

	if c.Cycles&1 == 1 {
		c.clock()
	}
}

func (c *CPU) resolveAddress(bus *SysBus, inst Instruction) (intermediateAddr, address uint16) {
	switch inst.Mode {
	case Accumulator:
		_ = c.read(bus, c.PC)
		return 0, 0

	case Implied:
		_ = c.read(bus, c.PC)
		return 0, 0

	case Immediate:
		pc := c.PC
		c.PC++
		return 0, pc

	case Absolute:
		lo := c.read(bus, c.PC)
		c.PC++

		hi := c.read(bus, c.PC)
		c.PC++

		return 0, uint16(hi)<<8 | uint16(lo)

	case ZeroPage:
		addr := c.read(bus, c.PC)
		c.PC++

		return 0, uint16(addr)

	case ZeroPageIndexedX:
		addr := c.read(bus, c.PC)
		c.PC++

		_ = c.read(bus, uint16(addr)) + c.X

		return 0, uint16(addr + c.X) //let it overflow

	case ZeroPageIndexedY:
		addr := c.read(bus, c.PC)
		c.PC++

		_ = c.read(bus, uint16(addr)) + c.Y

		return 0, uint16(addr + c.Y) //let it overflow

	case IndexedX:
		switch inst.Kind {
		case Read:
			lo := c.read(bus, c.PC)
			c.PC++

			hi := c.read(bus, c.PC)
			c.PC++

			if (lo + c.X) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.X))
			}

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.X)

		case ReadModWrite, Write:
			lo := c.read(bus, c.PC)
			c.PC++

			hi := c.read(bus, c.PC)
			c.PC++

			_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.X))

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.X)
		}

	case IndexedY:
		switch inst.Kind {
		case Read:
			lo := c.read(bus, c.PC)
			c.PC++

			hi := c.read(bus, c.PC)
			c.PC++

			if (lo + c.Y) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.Y))
			}

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.Y)

		case Write, ReadModWrite:
			lo := c.read(bus, c.PC)
			c.PC++

			hi := c.read(bus, c.PC)
			c.PC++

			addr := uint16(hi)<<8 | uint16(lo) + uint16(c.Y)
			_ = c.read(bus, addr)

			return 0, addr
		}

	case Relative:
		operand := c.read(bus, c.PC)
		c.PC++

		return 0, c.PC + uint16(int8(operand))

	case PreIndexedIndirect:
		pointer := c.read(bus, c.PC)
		c.PC++

		_ = c.read(bus, uint16(pointer)) + c.X

		pointer = pointer + c.X // let it overflow
		lo := c.read(bus, uint16(pointer))
		hi := c.read(bus, uint16(pointer+1)) // let it overflow

		return uint16(pointer), uint16(hi)<<8 | uint16(lo)

	case PostIndexedIndirect:
		switch inst.Kind {
		case Read:
			pointer := c.read(bus, c.PC)
			c.PC++

			lo := c.read(bus, uint16(pointer))
			hi := c.read(bus, uint16(pointer+1))

			if (lo + c.Y) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.Y))
			}

			addr := uint16(hi)<<8 | uint16(lo)
			return addr, addr + uint16(c.Y)

		case Write, ReadModWrite:
			pointer := c.read(bus, c.PC)
			c.PC++

			lo := c.read(bus, uint16(pointer))
			hi := c.read(bus, uint16(pointer+1))

			_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.Y))

			addr := uint16(hi)<<8 | uint16(lo)
			return addr, addr + uint16(c.Y)
		}

	case Indirect:
		pointerlo := c.read(bus, c.PC)
		c.PC++

		pointerhi := c.read(bus, c.PC)
		c.PC++

		pointer := uint16(pointerhi)<<8 | uint16(pointerlo)
		lo := c.read(bus, pointer)
		hi := c.read(bus, pointer&0xFF00|uint16(byte(pointer)+1))

		return pointer, uint16(hi)<<8 | uint16(lo)
	}

	return 0, 0
}

func (c *CPU) handleInterrupts(bus *SysBus) {
	switch c.interrupt {
	case NMI:
		c.handleNMI(bus)
	case IRQ:
		c.handleIRQ(bus)
	}

	c.interrupt = None
}

// NMI - Non-Maskable Interrupt
func (c *CPU) handleNMI(bus *SysBus) {
	c.pushAddress(bus, c.PC)
	c.push(bus, byte(c.P|Unused))

	c.PC = c.readAddress(bus, NMIAddr)

	// TODO: how do these 2 cycles get spent?
	c.clock()
	c.clock()
}

// IRQ - IRQ Interrupt
func (c *CPU) handleIRQ(bus *SysBus) {
	if c.P&InterruptDisable > 0 {
		return
	}

	c.pushAddress(bus, c.PC)
	c.push(bus, byte(c.P|Unused))

	c.PC = c.readAddress(bus, IRQ_BRKAddr)

	// TODO: how do these 2 cycles get spent?
	c.clock()
	c.clock()

	c.P |= InterruptDisable
}

func (c *CPU) push(bus *SysBus, v byte) {
	stackLo := uint16(c.S)
	c.write(bus, stackHi|stackLo, v)
	c.S--
}

func (c *CPU) pull(bus *SysBus) byte {
	c.S++
	stackLo := uint16(c.S)
	return c.read(bus, stackHi|stackLo)
}

func (c *CPU) pushAddress(bus *SysBus, value uint16) {
	hi := byte(value >> 8)
	lo := byte(value & 0xFF)

	c.push(bus, hi)
	c.push(bus, lo)
}

func (c *CPU) pullAddress(bus *SysBus) uint16 {
	lo := uint16(c.pull(bus))
	hi := uint16(c.pull(bus))

	return hi<<8 | lo
}

func (c *CPU) updateZero(v byte) {
	if v == 0 {
		c.P |= Zero
	} else {
		c.P &^= Zero
	}
}

func (c *CPU) updateNegative(v byte) {
	if v&0x80 > 0 {
		c.P |= Negative
	} else {
		c.P &^= Negative
	}
}

func (c *CPU) compare(a, b byte) {
	if a >= b {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}

	if a == b {
		c.P |= Zero
	} else {
		c.P &^= Zero
	}
	c.updateNegative(a - b)
}

func (c *CPU) dec(v byte) byte {
	r := v - 1
	c.updateZero(r)
	c.updateNegative(r)
	return r
}

func (c *CPU) inc(v byte) byte {
	r := v + 1
	c.updateZero(r)
	c.updateNegative(r)
	return r
}

func (c *CPU) add(v byte) {
	a := uint16(c.A)
	b := uint16(v)
	carry := uint16(c.P & Carry)

	result := a + b + carry

	if result&0x0100 > 0 {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}

	if a&0x80 == b&0x80 && a&0x80 != result&0x80 {
		c.P |= Overflow
	} else {
		c.P &^= Overflow
	}

	c.A = byte(result)
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

func (c *CPU) asl(v byte) byte {
	if v&0x80 > 0 {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}
	v = v << 1
	c.updateZero(v)
	c.updateNegative(v)
	return v
}

func (c *CPU) rol(v byte) byte {
	var carries bool
	if v&0x80 > 0 {
		carries = true
	}
	v = v << 1
	v |= byte(c.P & Carry)

	if carries {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}
	c.updateZero(v)
	c.updateNegative(v)

	return v
}

func (c *CPU) lsr(v byte) byte {
	if v&1 > 0 {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}
	v = v >> 1
	c.updateZero(v)
	c.updateNegative(v)
	return v
}

func (c *CPU) ror(v byte) byte {
	var carries bool
	if v&1 > 0 {
		carries = true
	}

	v = v >> 1
	if c.P&Carry > 0 {
		v |= 0x80
	}

	if carries {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}
	c.updateZero(v)
	c.updateNegative(v)

	return v
}

func (c *CPU) branch(addr uint16) {
	if c.PC&0xFF00 != addr&0xFF00 {
		c.clock()
	}

	c.clock()
	c.PC = addr
}

// BRK - Force Interrupt
//
// The BRK instruction forces the generation of an interrupt request.
// The program counter and processor status are pushed on the stack then the
// IRQ interrupt vector at $FFFE/F is loaded into the PC and the break flag in
// the status set to one.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Set to 1
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BRK(bus *SysBus, mode AddressingMode, addr uint16) {
	c.pushAddress(bus, c.PC+1)

	status := c.P
	status |= Unused
	status |= Break
	c.push(bus, byte(status))

	c.PC = c.readAddress(bus, IRQ_BRKAddr)
}

// NOP - No Operation
//
// The NOP instruction causes no changes to the processor other than the normal
// incrementing of the program counter to the next instruction.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) NOP(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode != Implied {
		c.read(bus, addr)
	}
}

// SEC - Set Carry Flag
// C = 1
//
// Set the carry flag to one.
//
// Processor Status after use:
// C	Carry Flag			Set to 1
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) SEC(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P |= Carry
}

// CLC - Clear Carry Flag
// C = 0
//
// Set the carry flag to zero.
//
// Processor Status after use:
// C	Carry Flag			Set to 0
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) CLC(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P &^= Carry
}

// SED - Set Decimal Flag
// D = 1
//
// Set the decimal mode flag to one.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Set to 1
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) SED(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P |= Decimal
}

// CLD - Clear Decimal Mode
// D = 0
//
// Sets the decimal mode flag to zero.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Set to 0
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) CLD(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P &^= Decimal
}

// SEI - Set Interrupt Disable
// I = 1
//
// Set the interrupt disable flag to one.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Set to 1
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) SEI(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P |= InterruptDisable
}

// CLI - Clear Interrupt Disable
// I = 0
//
// Clears the interrupt disable flag allowing normal interrupt requests to be serviced.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Set to 0
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) CLI(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P &^= InterruptDisable
}

// CLV - Clear Overflow Flag
// V = 0
//
// Clears the overflow flag.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Set to 0
// N	Negative Flag		Not affected
func (c *CPU) CLV(bus *SysBus, mode AddressingMode, addr uint16) {
	c.P &^= Overflow
}

// STA - Store Accumulator
// M = A
//
// Stores the contents of the accumulator into memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) STA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.write(bus, addr, c.A)
}

// STX - Store X Register
// M = X
//
// Stores the contents of the X register into memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) STX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.write(bus, addr, c.X)
}

// STY - Store Y Register
// M = Y
//
// Stores the contents of the Y register into memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) STY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.write(bus, addr, c.Y)
}

// LDA - Load Accumulator
// A,Z,N = M
//
// Loads a byte of memory into the accumulator setting the zero and negative
// flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of A is set
func (c *CPU) LDA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A = c.read(bus, addr)
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// LDX - Load X Register
// X,Z,N = M
//
// Loads a byte of memory into the X register setting the zero and negative
// flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if X = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of X is set
func (c *CPU) LDX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.X = c.read(bus, addr)
	c.updateZero(c.X)
	c.updateNegative(c.X)
}

// LDY - Load Y Register
// Y,Z,N = M
//
// Loads a byte of memory into the Y register setting the zero and negative
// flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if Y = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of Y is set
func (c *CPU) LDY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.Y = c.read(bus, addr)
	c.updateZero(c.Y)
	c.updateNegative(c.Y)
}

// TAX - Transfer Accumulator to X
// X = A
//
// Copies the current contents of the accumulator into the X register and sets
// the zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if X = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of X is set
func (c *CPU) TAX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.X = c.A
	c.updateZero(c.X)
	c.updateNegative(c.X)
}

// TAY - Transfer Accumulator to Y
// Y = A
//
// Copies the current contents of the accumulator into the Y register and sets
// the zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if Y = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of Y is set
func (c *CPU) TAY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.Y = c.A
	c.updateZero(c.Y)
	c.updateNegative(c.Y)
}

// TSX - Transfer Stack Pointer to X
// X = S
//
// Copies the current contents of the stack register into the X register and
// sets the zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if X = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of X is set
func (c *CPU) TSX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.X = c.S
	c.updateZero(c.X)
	c.updateNegative(c.X)
}

// TXA - Transfer X to Accumulator
// A = X
//
// Copies the current contents of the X register into the accumulator and sets
// the zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of A is set
func (c *CPU) TXA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A = c.X
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// TXS - Transfer X to Stack Pointer
// S = X
//
// Copies the current contents of the X register into the stack register.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) TXS(bus *SysBus, mode AddressingMode, addr uint16) {
	c.S = c.X
}

// TYA - Transfer Y to Accumulator
// A = Y
//
// Copies the current contents of the Y register into the accumulator and sets
// the zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of A is set
func (c *CPU) TYA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A = c.Y
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// PHA - Push Accumulator
//
// Pushes a copy of the accumulator on to the stack.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) PHA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.push(bus, c.A)
}

// PHP - Push Processor Status
//
// Pushes a copy of the status flags on to the stack.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) PHP(bus *SysBus, mode AddressingMode, addr uint16) {
	status := c.P
	status |= Break
	status |= Unused
	c.push(bus, byte(status))
}

// PLA - Pull Accumulator
//
// Pulls an 8 bit value from the stack and into the accumulator. The zero and
// negative flags are set as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of A is set
func (c *CPU) PLA(bus *SysBus, mode AddressingMode, addr uint16) {
	a := c.pull(bus)

	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	c.A = a
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// PLP - Pull Processor Status
//
// Pulls an 8 bit value from the stack and into the processor flags. The
// flags will take on new states as determined by the value pulled.
//
// Processor Status after use:
// C	Carry Flag	Set from stack
// Z	Zero Flag	Set from stack
// I	Interrupt Disable	Set from stack
// D	Decimal Mode Flag	Set from stack
// B	Break Command	Set from stack
// V	Overflow Flag	Set from stack
// N	Negative Flag	Set from stack
func (c *CPU) PLP(bus *SysBus, mode AddressingMode, addr uint16) {
	p := c.pull(bus)

	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	c.P = Status(p)
	c.P &^= Break //TODO figure out if we can just turn it off instead of actually ignoring
	c.P |= Unused
}

// DEC - Decrement Memory
// M,Z,N = M-1
//
// Subtracts one from the value held at a specified memory location setting the
// zero and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if result is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) DEC(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.dec(v))
}

// DEX - Decrement X Register
// X,Z,N = X-1
//
// Subtracts one from the X register setting the zero and negative flags as
// appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if X is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of X is set
func (c *CPU) DEX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.X = c.dec(c.X)
}

// DEY - Decrement Y Register
// Y,Z,N = Y-1
//
// Subtracts one from the Y register setting the zero and negative flags as
// appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if Y is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of Y is set
func (c *CPU) DEY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.Y = c.dec(c.Y)
}

// INC - Increment Memory
// M,Z,N = M+1
//
// Adds one to the value held at a specified memory location setting the zero
// and negative flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if result is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) INC(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.inc(v))
}

// INX - Increment X Register
// X,Z,N = X+1
//
// Adds one to the X register setting the zero and negative flags as
// appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if X is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of X is set
func (c *CPU) INX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.X = c.inc(c.X)
}

// INY - Increment Y Register
// Y,Z,N = Y+1
//
// Adds one to the Y register setting the zero and negative flags as
// appropriate.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if Y is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of Y is set
func (c *CPU) INY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.Y = c.inc(c.Y)
}

// ADC - Add with Carry
// A,Z,C,N = A+M+C
//
// This instruction adds the contents of a memory location to the accumulator
// together with the carry bit. If overflow occurs the carry bit is set,
// this enables multiple byte addition to be performed.
//
// Processor Status after use:
// C	Carry Flag			Set if overflow in bit 7
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Set if sign bit is incorrect
// N	Negative Flag		Set if bit 7 set
func (c *CPU) ADC(bus *SysBus, mode AddressingMode, addr uint16) {
	c.add(c.read(bus, addr))
}

// SBC - Subtract with Carry
// A,Z,C,N = A-M-(1-C)
//
// This instruction subtracts the contents of a memory location to the
// accumulator together with the not of the carry bit. If overflow occurs the
// carry bit is clear, this enables multiple byte subtraction to be performed.
//
// Processor Status after use:
// C	Carry Flag			Clear if overflow in bit 7
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Set if sign bit is incorrect
// N	Negative Flag		Set if bit 7 set
func (c *CPU) SBC(bus *SysBus, mode AddressingMode, addr uint16) {
	c.add(c.read(bus, addr) ^ 0xFF)
}

// ASL - Arithmetic Shift Left
// A,Z,C,N = M*2 or M,Z,C,N = M*2
//
// This operation shifts all the bits of the accumulator or memory contents one
// bit left. Bit 0 is set to 0 and bit 7 is placed in the carry flag. The effect
// of this operation is to multiply the memory contents by 2 (ignoring 2's
// complement considerations), setting the carry if the result will not fit in
// 8 bits.
//
// Processor Status after use:
// C	Carry Flag			Set to contents of old bit 7
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) ASL(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode == Accumulator {
		c.A = c.asl(c.A)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.asl(v))
}

// AND - Logical AND
// A,Z,N = A&M
//
// A logical AND is performed, bit by bit, on the accumulator contents using
// the contents of a byte of memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 set
func (c *CPU) AND(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A &= c.read(bus, addr)
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// EOR - Exclusive OR
// A,Z,N = A^M
//
// An exclusive OR is performed, bit by bit, on the accumulator contents using
// the contents of a byte of memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 set
func (c *CPU) EOR(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A ^= c.read(bus, addr)
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// LSR - Logical Shift Right
// A,C,Z,N = A/2 or M,C,Z,N = M/2
//
// Each of the bits in A or M is shifted one place to the right. The bit that
// was in bit 0 is shifted into the carry flag. Bit 7 is set to zero.
//
// Processor Status after use:
// C	Carry Flag			Set to contents of old bit 0
// Z	Zero Flag			Set if result = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) LSR(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode == Accumulator {
		c.A = c.lsr(c.A)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.lsr(v))
}

// ROL - Rotate Left
//
// Move each of the bits in either A or M one place to the left. Bit 0 is filled
// with the current value of the carry flag whilst the old bit 7 becomes the new
// carry flag value.
//
// Processor Status after use:
// C	Carry Flag			Set to contents of old bit 7
// Z	Zero Flag			Set if result = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) ROL(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode == Accumulator {
		c.A = c.rol(c.A)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.rol(v))
}

// ROR - Rotate Right
//
// Move each of the bits in either A or M one place to the right. Bit 7 is
// filled with the current value of the carry flag whilst the old bit 0 becomes
// the new carry flag value.
//
// Processor Status after use:
// C	Carry Flag			Set to contents of old bit 0
// Z	Zero Flag			Set if result = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) ROR(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode == Accumulator {
		c.A = c.ror(c.A)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.ror(v))
}

// ORA - Logical Inclusive OR
// A,Z,N = A|M
//
// An inclusive OR is performed, bit by bit, on the accumulator contents using
// the contents of a byte of memory.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if A = 0
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 set
func (c *CPU) ORA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.A |= c.read(bus, addr)
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// BIT - Bit Test
// A & M, N = M7, V = M6
//
// This instruction is used to test if one or more bits are set in a target
// memory location. The mask pattern in A is ANDed with the value in memory to
// set or clear the zero flag, but the result is not kept. Bits 7 and 6 of the
// value from memory are copied into the N and V flags.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Set if the result if the AND is zero
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Set to bit 6 of the memory value
// N	Negative Flag		Set to bit 7 of the memory value
func (c *CPU) BIT(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)

	c.updateNegative(v)
	c.updateZero(c.A & v)

	if v&0x40 > 0 {
		c.P |= Overflow
	} else {
		c.P &^= Overflow
	}
}

// CMP - Compare
// Z,C,N = A-M
//
// This instruction compares the contents of the accumulator with another memory
// held value and sets the zero and carry flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Set if A >= M
// Z	Zero Flag			Set if A = M
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) CMP(bus *SysBus, mode AddressingMode, addr uint16) {
	c.compare(c.A, c.read(bus, addr))
}

// CPX - Compare X Register
// Z,C,N = X-M
//
// This instruction compares the contents of the X register with another memory
// held value and sets the zero and carry flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Set if X >= M
// Z	Zero Flag			Set if X = M
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) CPX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.compare(c.X, c.read(bus, addr))
}

// CPY - Compare Y Register
// Z,C,N = Y-M
//
// This instruction compares the contents of the Y register with another memory
// held value and sets the zero and carry flags as appropriate.
//
// Processor Status after use:
// C	Carry Flag			Set if Y >= M
// Z	Zero Flag			Set if Y = M
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Set if bit 7 of the result is set
func (c *CPU) CPY(bus *SysBus, mode AddressingMode, addr uint16) {
	c.compare(c.Y, c.read(bus, addr))
}

// BCC - Branch if Carry Clear
//
// If the carry flag is clear then add the relative displacement to the program
// counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BCC(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Carry > 0 {
		return
	}

	c.branch(addr)
}

// BCS - Branch if Carry Set
//
// If the carry flag is set then add the relative displacement to the program
// counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BCS(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Carry == 0 {
		return
	}

	c.branch(addr)
}

// BVC - Branch if Overflow Clear
//
// If the overflow flag is clear then add the relative displacement to the
// program counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BVC(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Overflow > 0 {
		return
	}

	c.branch(addr)
}

// BVS - Branch if Overflow Set
//
// If the overflow flag is set then add the relative displacement to the
// program counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BVS(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Overflow == 0 {
		return
	}

	c.branch(addr)
}

// BEQ - Branch if Equal
//
// If the zero flag is set then add the relative displacement to the program
// counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BEQ(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Zero == 0 {
		return
	}

	c.branch(addr)
}

// BNE - Branch if Not Equal
//
// If the zero flag is clear then add the relative displacement to the program
// counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BNE(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Zero > 0 {
		return
	}

	c.branch(addr)
}

// BMI - Branch if Minus
//
// If the negative flag is set then add the relative displacement to the program
// counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BMI(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Negative == 0 {
		return
	}

	c.branch(addr)
}

// BPL - Branch if Positive
//
// If the negative flag is clear then add the relative displacement to the
// program counter to cause a branch to a new location.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) BPL(bus *SysBus, mode AddressingMode, addr uint16) {
	if c.P&Negative > 0 {
		return
	}

	c.branch(addr)
}

// JMP - Jump
//
// Sets the program counter to the address specified by the operand.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) JMP(bus *SysBus, mode AddressingMode, addr uint16) {
	c.PC = addr
}

// JSR - Jump to Subroutine
//
// The JSR instruction pushes the address (minus one) of the return point on to
// the stack and then sets the program counter to the target memory address.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) JSR(bus *SysBus, mode AddressingMode, addr uint16) {
	c.clock()

	c.pushAddress(bus, c.PC-1)
	c.PC = addr
}

// RTI - Return from Interrupt
//
// The RTI instruction is used at the end of an interrupt processing routine.
// It pulls the processor flags from the stack followed by the program counter.
//
// Processor Status after use:
// C	Carry Flag			Set from stack
// Z	Zero Flag			Set from stack
// I	Interrupt Disable	Set from stack
// D	Decimal Mode Flag	Set from stack
// B	Break Command		Set from stack
// V	Overflow Flag		Set from stack
// N	Negative Flag		Set from stack
func (c *CPU) RTI(bus *SysBus, mode AddressingMode, addr uint16) {
	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	p := c.pull(bus)

	c.P = Status(p) & ^Break
	c.P |= Unused

	c.PC = c.pullAddress(bus)
}

// RTS - Return from Subroutine
//
// The RTS instruction is used at the end of a subroutine to return to the
// calling routine. It pulls the program counter (minus one) from the stack.
//
// Processor Status after use:
// C	Carry Flag			Not affected
// Z	Zero Flag			Not affected
// I	Interrupt Disable	Not affected
// D	Decimal Mode Flag	Not affected
// B	Break Command		Not affected
// V	Overflow Flag		Not affected
// N	Negative Flag		Not affected
func (c *CPU) RTS(bus *SysBus, mode AddressingMode, addr uint16) {
	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	pclo := uint16(c.pull(bus))
	pchi := uint16(c.pull(bus))

	c.clock()
	c.PC = pchi<<8 | pclo + 1
}

// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================
// ====================================================================================================================================

// Equivalent to AND #i then LSR A. Some sources call this "ASR"; we do not
// follow this out of confusion with the mnemonic for a pseudoinstruction that
// combines CMP #$80 (or ANC #$FF) then ROR. Note that ALR #$FE acts like LSR
// followed by CLC.
func (c *CPU) ALR(bus *SysBus, mode AddressingMode, addr uint16) {
	c.AND(bus, mode, addr)
	c.LSR(bus, Accumulator, addr)
}

// Does AND #i, setting N and Z flags based on the result. Then it copies N
// (bit 7) to C. ANC #$FF could be useful for sign-extending, much like
// CMP #$80. ANC #$00 acts like LDA #$00 followed by CLC.
func (c *CPU) ANC(bus *SysBus, mode AddressingMode, addr uint16) {
	c.AND(bus, mode, addr)

	if c.P&Negative > 0 {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}
}

// Similar to AND #i then ROR A, except sets the flags differently. N and Z are
// normal, but C is bit 6 and V is bit 6 xor bit 5. A fast way to perform signed
// division by 4 is: CMP #$80; ARR #$FF; ROR. This can be extended to larger
// powers of two.
func (c *CPU) ARR(bus *SysBus, mode AddressingMode, addr uint16) {
	c.AND(bus, mode, addr)
	c.ROR(bus, Accumulator, addr)

	if (c.A>>6)&1 > 0 {
		c.P |= Carry
	} else {
		c.P &^= Carry
	}

	if ((c.A>>6)&1)^((c.A>>5)&1) > 0 {
		c.P |= Overflow
	} else {
		c.P &^= Overflow
	}
}

// Sets X to {(A AND X) - #value without borrow}, and updates NZC. One might use
// TXA AXS #-element_size to iterate through an array of structures or other
// elements larger than a byte, where the 6502 architecture usually prefers a
// structure of arrays. For example, TXA AXS #$FC could step to the next OAM
// entry or to the next APU channel, saving one byte and four cycles over four
// INXs. Also called SBX.
func (c *CPU) AXS(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("AXS wat") //SBC without carry void asx()
}

// Shortcut for LDA value then TAX. Saves a byte and two cycles and allows use
// of the X register with the (d),Y addressing mode. Notice that the immediate
// is missing; the opcode that would have been LAX is affected by line noise on
// the data bus. MOS 6502: even the bugs have bugs.
func (c *CPU) LAX(bus *SysBus, mode AddressingMode, addr uint16) {
	if mode == Immediate {
		panic("LAX Immediate")
	}

	c.LDA(bus, mode, addr)
	c.TAX(bus, mode, addr)
}

// Stores the bitwise AND of A and X. As with STA and STX, no flags are affected
func (c *CPU) SAX(bus *SysBus, mode AddressingMode, addr uint16) {
	c.write(bus, addr, c.A&c.X)
}

// Equivalent to DEC value then CMP value, except supporting more addressing
// modes. LDA #$FF followed by DCP can be used to check if the decrement
// underflows, which is useful for multi-byte decrements.
func (c *CPU) DCP(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.dec(v)
	c.write(bus, addr, v)
	c.compare(c.A, v)
}

// Equivalent to INC value then SBC value, except supporting more addressing
// modes.
func (c *CPU) ISC(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.inc(v)
	c.write(bus, addr, v)
	c.add(v ^ 0xFF)
}

// Equivalent to ROL value then AND value, except supporting more addressing
// modes. LDA #$FF followed by RLA is an efficient way to rotate a variable
// while also loading it in A.
func (c *CPU) RLA(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.rol(v)
	c.write(bus, addr, v)

	c.A &= v
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// Equivalent to ROR value then ADC value, except supporting more addressing
// modes. Essentially this computes A + value / 2, where value is 9-bit and the
// division is rounded up.
func (c *CPU) RRA(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.ror(v)
	c.write(bus, addr, v)
	c.add(v)
}

// Equivalent to ASL value then ORA value, except supporting more addressing
// modes. LDA #0 followed by SLO is an efficient way to shift a variable while
// also loading it in A.
func (c *CPU) SLO(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.asl(v)
	c.write(bus, addr, v)

	c.A |= v
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

// Equivalent to LSR value then EOR value, except supporting more addressing
// modes. LDA #0 followed by SRE is an efficient way to shift a variable while
// also loading it in A.
func (c *CPU) SRE(bus *SysBus, mode AddressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.lsr(v)
	c.write(bus, addr, v)

	c.A ^= v
	c.updateZero(c.A)
	c.updateNegative(c.A)
}

func (c *CPU) KIL(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("KIL NOT IMPLEMENTED")
}
func (c *CPU) XAA(bus *SysBus, mode AddressingMode, addr uint16) {
	c.TXA(bus, mode, addr)
	c.AND(bus, mode, addr)
}
func (c *CPU) AHX(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("AHX NOT IMPLEMENTED")
}
func (c *CPU) TAS(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("TAS NOT IMPLEMENTED")
}
func (c *CPU) SHY(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("SHY NOT IMPLEMENTED")
}
func (c *CPU) SHX(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("SHX NOT IMPLEMENTED")
}
func (c *CPU) LAS(bus *SysBus, mode AddressingMode, addr uint16) {
	panic("LAS NOT IMPLEMENTED")
}
