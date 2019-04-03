package nes

import (
	"io"
)

const cpuFreq float64 = 1789773

type interrupt byte

const (
	none interrupt = iota
	nmi
	nmiNext
	irq
)

const (
	nmiAddr    = uint16(0xFFFA)
	resetAddr  = uint16(0xFFFC)
	irqBrkAddr = uint16(0xFFFE)

	stackHi = 0x0100
)

// status are all the flags that represent the processor status.
type status byte

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
	carry status = 1 << iota

	// Zero flag is set when the result of an instruction is zero.
	zero

	// InterruptDisable flag.
	//
	// When set, all interrupts except the NMI are inhibited.
	// Can be set or cleared directly with SEI, CLI.
	// Automatically set by the cpu when an IRQ is triggered, and restored
	// to its previous state by RTI.
	//
	// If the /IRQ line is low (IRQ pending) when this flag is cleared, an
	// interrupt will immediately be triggered.
	interruptDisable

	// Decimal flag. On the NES, this flag has no effect.
	decimal

	// Break flag.
	//
	// While there are only six flags in the processor status register within
	// the cpu, when transferred to the stack, there are two additional bits.
	//
	// These do not represent a register that can hold a value but can be used
	// to distinguish how the flags were pushed.
	//
	// Some 6502 references call this the "B flag", though it does not represent
	// an actual cpu register.
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
	brk

	// Unused flag.
	unused

	// Overflow flag.
	//
	// ADC, SBC, and CMP will set this flag if the signed result would be
	// invalid http://www.6502.org/tutorials/vflag.html, necessary for making
	// signed comparisons http://www.6502.org/tutorials/compare_beyond.html#5.
	//
	// BIT will load bit 6 of the addressed value directly into the V flag.
	// Can be cleared directly with CLV.
	// There is no corresponding set instruction.
	overflow

	// Negative flag.
	//
	// After most instructions that have a value result, this flag will contain
	// bit 7 of that result.
	// BIT will load bit 7 of the addressed value directly into the N flag.
	negative
)

type cpu struct {
	cycles uint64

	// A, along with the arithmetic logic unit (ALU), supports using the status
	// register for carrying, overflow detection, and so on.
	a byte

	// X and Y are used for several addressing modes. They can be used as loop
	// counters easily, using INC/DEC and branch instructions.
	//
	// Not being the accumulator, they have limited addressing modes themselves
	// when loading and saving.
	x, y byte

	// The program counter PC supports 65536 direct (unbanked) memory locations,
	// however not all values are sent to the cartridge.
	//
	// It can be accessed either by allowing cpu's internal fetch logic
	// increment the address bus, an interrupt (NMI, Reset, IRQ/BRQ), and using
	// the RTS/JMP/JSR/Branch instructions.
	pc uint16

	// The Stack Pointer can be accessed using interrupts, pulls, pushes, and
	// transfers.
	s byte

	// The Status Register has 6 bits used by the ALU but is byte-wide.
	// PHP, PLP, arithmetic, testing, and branch instructions can access this
	// register.
	//
	// See Status for more info.
	p status

	debug     io.Writer
	interrupt interrupt

	pputemp *ppu
	aputemp *apu
}

func newCpu(debug io.Writer, ppu *ppu, apu *apu) *cpu {
	return &cpu{
		debug:   debug,
		p:       interruptDisable | unused,
		s:       0xFD,
		pc:      resetAddr,
		pputemp: ppu,
		aputemp: apu,
	}
}

func (c *cpu) init(bus *sysBus) {
	c.pc = c.readAddress(bus, resetAddr)
}

func (c *cpu) setPC(pc uint16) {
	c.pc = pc
}

func (c *cpu) reset(bus *sysBus) {
	c.p |= interruptDisable
	c.s -= 3

	c.pc = c.readAddress(bus, resetAddr)
}

func (c *cpu) trigger(interrupt interrupt) {
	if interrupt == irq && c.p&interruptDisable > 0 {
		return
	}

	c.interrupt = interrupt
}

func (c *cpu) execute(bus *sysBus) uint64 {
	oldCycles := c.cycles

	c.handleInterrupts(bus)

	initialPc := c.pc

	opCode := c.read(bus, c.pc)
	c.pc++

	inst := instructions[opCode]
	intermediateAddr, addr := c.resolveAddress(bus, inst)

	if c.debug != nil {
		//TODO: rework disassembly/tracing
		disassemble(c.debug, bus, initialPc, c.a, c.x, c.y, byte(c.p), c.s, inst, intermediateAddr, addr, oldCycles, c.pputemp)
	}

	switch opCode {
	case 0x04, 0x0C, 0x14, 0x1A, 0x1C, 0x34, 0x3A, 0x3C, 0x44, 0x54, 0x5A,
		0x5C, 0x64, 0x74, 0x7A, 0x7C, 0x80, 0x82, 0x89, 0xC2, 0xD4, 0xDA,
		0xDC, 0xE2, 0xEA, 0xF4, 0xFA, 0xFC:
		c.nop(bus, inst.mode, addr)
	case 0x61, 0x65, 0x69, 0x6D, 0x71, 0x75, 0x79, 0x7D:
		c.adc(bus, inst.mode, addr)
	case 0x93, 0x9F:
		c.ahx(bus, inst.mode, addr)
	case 0x4B:
		c.alr(bus, inst.mode, addr)
	case 0x0B, 0x2B:
		c.anc(bus, inst.mode, addr)
	case 0x21, 0x25, 0x29, 0x2D, 0x31, 0x35, 0x39, 0x3D:
		c.and(bus, inst.mode, addr)
	case 0x6B:
		c.arr(bus, inst.mode, addr)
	case 0x06, 0x0A, 0x0E, 0x16, 0x1E:
		c.asl(bus, inst.mode, addr)
	case 0xCB:
		c.axs(bus, inst.mode, addr)
	case 0x90:
		c.bcc(bus, inst.mode, addr)
	case 0xB0:
		c.bcs(bus, inst.mode, addr)
	case 0xF0:
		c.beq(bus, inst.mode, addr)
	case 0x24, 0x2C:
		c.bit(bus, inst.mode, addr)
	case 0x30:
		c.bmi(bus, inst.mode, addr)
	case 0xD0:
		c.bne(bus, inst.mode, addr)
	case 0x10:
		c.bpl(bus, inst.mode, addr)
	case 0x00:
		c.brk(bus, inst.mode, addr)
	case 0x50:
		c.bvc(bus, inst.mode, addr)
	case 0x70:
		c.bvs(bus, inst.mode, addr)
	case 0x18:
		c.clc(bus, inst.mode, addr)
	case 0xD8:
		c.cld(bus, inst.mode, addr)
	case 0x58:
		c.cli(bus, inst.mode, addr)
	case 0xB8:
		c.clv(bus, inst.mode, addr)
	case 0xC1, 0xC5, 0xC9, 0xCD, 0xD1, 0xD5, 0xD9, 0xDD:
		c.cmp(bus, inst.mode, addr)
	case 0xE0, 0xE4, 0xEC:
		c.cpx(bus, inst.mode, addr)
	case 0xC0, 0xC4, 0xCC:
		c.cpy(bus, inst.mode, addr)
	case 0xC3, 0xC7, 0xCF, 0xD3, 0xD7, 0xDB, 0xDF:
		c.dcp(bus, inst.mode, addr)
	case 0xC6, 0xCE, 0xD6, 0xDE:
		c.dec(bus, inst.mode, addr)
	case 0xCA:
		c.dex(bus, inst.mode, addr)
	case 0x88:
		c.dey(bus, inst.mode, addr)
	case 0x41, 0x45, 0x49, 0x4D, 0x51, 0x55, 0x59, 0x5D:
		c.eor(bus, inst.mode, addr)
	case 0xE6, 0xEE, 0xF6, 0xFE:
		c.inc(bus, inst.mode, addr)
	case 0xE8:
		c.inx(bus, inst.mode, addr)
	case 0xC8:
		c.iny(bus, inst.mode, addr)
	case 0xE3, 0xE7, 0xEF, 0xF3, 0xF7, 0xFB, 0xFF:
		c.isc(bus, inst.mode, addr)
	case 0x4C, 0x6C:
		c.jmp(bus, inst.mode, addr)
	case 0x20:
		c.jsr(bus, inst.mode, addr)
	case 0x02, 0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x92, 0xB2, 0xD2, 0xF2:
		c.kil(bus, inst.mode, addr)
	case 0xBB:
		c.las(bus, inst.mode, addr)
	case 0xA3, 0xA7, 0xAB, 0xAF, 0xB3, 0xB7, 0xBF:
		c.lax(bus, inst.mode, addr)
	case 0xA1, 0xA5, 0xA9, 0xAD, 0xB1, 0xB5, 0xB9, 0xBD:
		c.lda(bus, inst.mode, addr)
	case 0xA2, 0xA6, 0xAE, 0xB6, 0xBE:
		c.ldx(bus, inst.mode, addr)
	case 0xA0, 0xA4, 0xAC, 0xB4, 0xBC:
		c.ldy(bus, inst.mode, addr)
	case 0x46, 0x4A, 0x4E, 0x56, 0x5E:
		c.lsr(bus, inst.mode, addr)
	case 0x01, 0x05, 0x09, 0x0D, 0x11, 0x15, 0x19, 0x1D:
		c.ora(bus, inst.mode, addr)
	case 0x48:
		c.pha(bus, inst.mode, addr)
	case 0x08:
		c.php(bus, inst.mode, addr)
	case 0x68:
		c.pla(bus, inst.mode, addr)
	case 0x28:
		c.plp(bus, inst.mode, addr)
	case 0x23, 0x27, 0x2F, 0x33, 0x37, 0x3B, 0x3F:
		c.rla(bus, inst.mode, addr)
	case 0x26, 0x2A, 0x2E, 0x36, 0x3E:
		c.rol(bus, inst.mode, addr)
	case 0x66, 0x6A, 0x6E, 0x76, 0x7E:
		c.ror(bus, inst.mode, addr)
	case 0x63, 0x67, 0x6F, 0x73, 0x77, 0x7B, 0x7F:
		c.rra(bus, inst.mode, addr)
	case 0x40:
		c.rti(bus, inst.mode, addr)
	case 0x60:
		c.rts(bus, inst.mode, addr)
	case 0x83, 0x87, 0x8F, 0x97:
		c.sax(bus, inst.mode, addr)
	case 0xE1, 0xE5, 0xE9, 0xEB, 0xED, 0xF1, 0xF5, 0xF9, 0xFD:
		c.sbc(bus, inst.mode, addr)
	case 0x38:
		c.sec(bus, inst.mode, addr)
	case 0xF8:
		c.sed(bus, inst.mode, addr)
	case 0x78:
		c.sei(bus, inst.mode, addr)
	case 0x9E:
		c.shx(bus, inst.mode, addr)
	case 0x9C:
		c.shy(bus, inst.mode, addr)
	case 0x03, 0x07, 0x0F, 0x13, 0x17, 0x1B, 0x1F:
		c.slo(bus, inst.mode, addr)
	case 0x43, 0x47, 0x4F, 0x53, 0x57, 0x5B, 0x5F:
		c.sre(bus, inst.mode, addr)
	case 0x81, 0x85, 0x8D, 0x91, 0x95, 0x99, 0x9D:
		c.sta(bus, inst.mode, addr)
	case 0x86, 0x8E, 0x96:
		c.stx(bus, inst.mode, addr)
	case 0x84, 0x8C, 0x94:
		c.sty(bus, inst.mode, addr)
	case 0x9B:
		c.tas(bus, inst.mode, addr)
	case 0xAA:
		c.tax(bus, inst.mode, addr)
	case 0xA8:
		c.tay(bus, inst.mode, addr)
	case 0xBA:
		c.tsx(bus, inst.mode, addr)
	case 0x8A:
		c.txa(bus, inst.mode, addr)
	case 0x9A:
		c.txs(bus, inst.mode, addr)
	case 0x98:
		c.tya(bus, inst.mode, addr)
	case 0x8B:
		c.xaa(bus, inst.mode, addr)
	}

	return c.cycles - oldCycles
}

func (c *cpu) clock() {
	c.cycles++
	c.pputemp.tick(c)
	c.pputemp.tick(c)
	c.pputemp.tick(c)
	c.aputemp.clock(c)
}

func (c *cpu) read(bus *sysBus, address uint16) byte {
	c.clock()
	v := bus.read(address)
	return v
}

func (c *cpu) readAddress(bus *sysBus, address uint16) uint16 {
	c.clock()
	lo := bus.read(address)
	c.clock()
	hi := bus.read(address + 1)

	addr := uint16(hi)<<8 | uint16(lo)

	return addr
}

func (c *cpu) write(bus *sysBus, address uint16, value byte) {
	if address == oamDmaAddr {
		c.dmaTransfer(bus, value)
		return
	}

	c.clock()
	bus.write(address, value)
}

func (c *cpu) dmaTransfer(bus *sysBus, address byte) {
	addr := uint16(address) << 8
	for i := 0; i < 256; i++ {
		c.clock()
		v := bus.read(addr)

		c.clock()
		bus.write(oamDmaAddr, v)

		addr++
	}

	if c.cycles&1 == 1 {
		c.clock()
	}
}

func (c *cpu) resolveAddress(bus *sysBus, inst instruction) (intermediateAddr, address uint16) {
	switch inst.mode {
	case accumulator:
		_ = c.read(bus, c.pc)
		return 0, 0

	case implied:
		_ = c.read(bus, c.pc)
		return 0, 0

	case immediate:
		pc := c.pc
		c.pc++
		return 0, pc

	case absolute:
		lo := c.read(bus, c.pc)
		c.pc++

		hi := c.read(bus, c.pc)
		c.pc++

		return 0, uint16(hi)<<8 | uint16(lo)

	case zeroPage:
		addr := c.read(bus, c.pc)
		c.pc++

		return 0, uint16(addr)

	case zeroPageIndexedX:
		addr := c.read(bus, c.pc)
		c.pc++

		_ = c.read(bus, uint16(addr)) + c.x

		return 0, uint16(addr + c.x) //let it overflow

	case zeroPageIndexedY:
		addr := c.read(bus, c.pc)
		c.pc++

		_ = c.read(bus, uint16(addr)) + c.y

		return 0, uint16(addr + c.y) //let it overflow

	case indexedX:
		switch inst.kind {
		case read:
			lo := c.read(bus, c.pc)
			c.pc++

			hi := c.read(bus, c.pc)
			c.pc++

			if (lo + c.x) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.x))
			}

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.x)

		case readModWrite, write:
			lo := c.read(bus, c.pc)
			c.pc++

			hi := c.read(bus, c.pc)
			c.pc++

			_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.x))

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.x)
		}

	case indexedY:
		switch inst.kind {
		case read:
			lo := c.read(bus, c.pc)
			c.pc++

			hi := c.read(bus, c.pc)
			c.pc++

			if (lo + c.y) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.y))
			}

			return 0, uint16(hi)<<8 | uint16(lo) + uint16(c.y)

		case write, readModWrite:
			lo := c.read(bus, c.pc)
			c.pc++

			hi := c.read(bus, c.pc)
			c.pc++

			addr := uint16(hi)<<8 | uint16(lo) + uint16(c.y)
			_ = c.read(bus, addr)

			return 0, addr
		}

	case relative:
		operand := c.read(bus, c.pc)
		c.pc++

		return 0, c.pc + uint16(int8(operand))

	case preIndexedIndirect:
		pointer := c.read(bus, c.pc)
		c.pc++

		_ = c.read(bus, uint16(pointer)) + c.x

		pointer = pointer + c.x // let it overflow
		lo := c.read(bus, uint16(pointer))
		hi := c.read(bus, uint16(pointer+1)) // let it overflow

		return uint16(pointer), uint16(hi)<<8 | uint16(lo)

	case postIndexedIndirect:
		switch inst.kind {
		case read:
			pointer := c.read(bus, c.pc)
			c.pc++

			lo := c.read(bus, uint16(pointer))
			hi := c.read(bus, uint16(pointer+1))

			if (lo + c.y) < lo {
				_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.y))
			}

			addr := uint16(hi)<<8 | uint16(lo)
			return addr, addr + uint16(c.y)

		case write, readModWrite:
			pointer := c.read(bus, c.pc)
			c.pc++

			lo := c.read(bus, uint16(pointer))
			hi := c.read(bus, uint16(pointer+1))

			_ = c.read(bus, uint16(hi)<<8|uint16(lo+c.y))

			addr := uint16(hi)<<8 | uint16(lo)
			return addr, addr + uint16(c.y)
		}

	case indirect:
		pointerlo := c.read(bus, c.pc)
		c.pc++

		pointerhi := c.read(bus, c.pc)
		c.pc++

		pointer := uint16(pointerhi)<<8 | uint16(pointerlo)
		lo := c.read(bus, pointer)
		hi := c.read(bus, pointer&0xFF00|uint16(byte(pointer)+1))

		return pointer, uint16(hi)<<8 | uint16(lo)
	}

	return 0, 0
}

func (c *cpu) handleInterrupts(bus *sysBus) {
	switch c.interrupt {
	case nmi:
		c.handleNmi(bus)
		c.interrupt = none
	case nmiNext:
		// skip NMI now, handle it next instr
		c.interrupt = nmi
	case irq:
		c.handleIrq(bus)
		c.interrupt = none
	}

}

// NMI - Non-Maskable Interrupt
func (c *cpu) handleNmi(bus *sysBus) {
	c.pushAddress(bus, c.pc)
	c.push(bus, byte(c.p|unused))

	c.pc = c.readAddress(bus, nmiAddr)

	// TODO: how do these 2 cycles get spent?
	c.clock()
	c.clock()
}

// IRQ - IRQ Interrupt
func (c *cpu) handleIrq(bus *sysBus) {
	if c.p&interruptDisable > 0 {
		return
	}

	c.pushAddress(bus, c.pc)
	c.push(bus, byte(c.p|unused))

	c.pc = c.readAddress(bus, irqBrkAddr)

	// TODO: how do these 2 cycles get spent?
	c.clock()
	c.clock()

	c.p |= interruptDisable
}

func (c *cpu) push(bus *sysBus, v byte) {
	stackLo := uint16(c.s)
	c.write(bus, stackHi|stackLo, v)
	c.s--
}

func (c *cpu) pull(bus *sysBus) byte {
	c.s++
	stackLo := uint16(c.s)
	return c.read(bus, stackHi|stackLo)
}

func (c *cpu) pushAddress(bus *sysBus, value uint16) {
	hi := byte(value >> 8)
	lo := byte(value & 0xFF)

	c.push(bus, hi)
	c.push(bus, lo)
}

func (c *cpu) pullAddress(bus *sysBus) uint16 {
	lo := uint16(c.pull(bus))
	hi := uint16(c.pull(bus))

	return hi<<8 | lo
}

func (c *cpu) updateZero(v byte) {
	if v == 0 {
		c.p |= zero
	} else {
		c.p &^= zero
	}
}

func (c *cpu) updateNegative(v byte) {
	if v&0x80 > 0 {
		c.p |= negative
	} else {
		c.p &^= negative
	}
}

func (c *cpu) compare(a, b byte) {
	if a >= b {
		c.p |= carry
	} else {
		c.p &^= carry
	}

	if a == b {
		c.p |= zero
	} else {
		c.p &^= zero
	}
	c.updateNegative(a - b)
}

func (c *cpu) doDec(v byte) byte {
	r := v - 1
	c.updateZero(r)
	c.updateNegative(r)
	return r
}

func (c *cpu) doInc(v byte) byte {
	r := v + 1
	c.updateZero(r)
	c.updateNegative(r)
	return r
}

func (c *cpu) doAdd(v byte) {
	a := uint16(c.a)
	b := uint16(v)
	crry := uint16(c.p & carry)

	result := a + b + crry

	if result&0x0100 > 0 {
		c.p |= carry
	} else {
		c.p &^= carry
	}

	if a&0x80 == b&0x80 && a&0x80 != result&0x80 {
		c.p |= overflow
	} else {
		c.p &^= overflow
	}

	c.a = byte(result)
	c.updateZero(c.a)
	c.updateNegative(c.a)
}

func (c *cpu) doAsl(v byte) byte {
	if v&0x80 > 0 {
		c.p |= carry
	} else {
		c.p &^= carry
	}
	v = v << 1
	c.updateZero(v)
	c.updateNegative(v)
	return v
}

func (c *cpu) doRol(v byte) byte {
	var carries bool
	if v&0x80 > 0 {
		carries = true
	}
	v = v << 1
	v |= byte(c.p & carry)

	if carries {
		c.p |= carry
	} else {
		c.p &^= carry
	}
	c.updateZero(v)
	c.updateNegative(v)

	return v
}

func (c *cpu) doLsr(v byte) byte {
	if v&1 > 0 {
		c.p |= carry
	} else {
		c.p &^= carry
	}
	v = v >> 1
	c.updateZero(v)
	c.updateNegative(v)
	return v
}

func (c *cpu) doRor(v byte) byte {
	var carries bool
	if v&1 > 0 {
		carries = true
	}

	v = v >> 1
	if c.p&carry > 0 {
		v |= 0x80
	}

	if carries {
		c.p |= carry
	} else {
		c.p &^= carry
	}
	c.updateZero(v)
	c.updateNegative(v)

	return v
}

func (c *cpu) branch(addr uint16) {
	if c.pc&0xFF00 != addr&0xFF00 {
		c.clock()
	}

	c.clock()
	c.pc = addr
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
func (c *cpu) brk(bus *sysBus, mode addressingMode, addr uint16) {
	c.pushAddress(bus, c.pc+1)

	status := c.p
	status |= unused
	status |= brk
	c.push(bus, byte(status))
	c.p |= interruptDisable

	c.pc = c.readAddress(bus, irqBrkAddr)
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
func (c *cpu) nop(bus *sysBus, mode addressingMode, addr uint16) {
	if mode != implied {
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
func (c *cpu) sec(bus *sysBus, mode addressingMode, addr uint16) {
	c.p |= carry
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
func (c *cpu) clc(bus *sysBus, mode addressingMode, addr uint16) {
	c.p &^= carry
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
func (c *cpu) sed(bus *sysBus, mode addressingMode, addr uint16) {
	c.p |= decimal
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
func (c *cpu) cld(bus *sysBus, mode addressingMode, addr uint16) {
	c.p &^= decimal
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
func (c *cpu) sei(bus *sysBus, mode addressingMode, addr uint16) {
	c.p |= interruptDisable
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
func (c *cpu) cli(bus *sysBus, mode addressingMode, addr uint16) {
	c.p &^= interruptDisable
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
func (c *cpu) clv(bus *sysBus, mode addressingMode, addr uint16) {
	c.p &^= overflow
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
func (c *cpu) sta(bus *sysBus, mode addressingMode, addr uint16) {
	c.write(bus, addr, c.a)
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
func (c *cpu) stx(bus *sysBus, mode addressingMode, addr uint16) {
	c.write(bus, addr, c.x)
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
func (c *cpu) sty(bus *sysBus, mode addressingMode, addr uint16) {
	c.write(bus, addr, c.y)
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
func (c *cpu) lda(bus *sysBus, mode addressingMode, addr uint16) {
	c.a = c.read(bus, addr)
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) ldx(bus *sysBus, mode addressingMode, addr uint16) {
	c.x = c.read(bus, addr)
	c.updateZero(c.x)
	c.updateNegative(c.x)
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
func (c *cpu) ldy(bus *sysBus, mode addressingMode, addr uint16) {
	c.y = c.read(bus, addr)
	c.updateZero(c.y)
	c.updateNegative(c.y)
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
func (c *cpu) tax(bus *sysBus, mode addressingMode, addr uint16) {
	c.x = c.a
	c.updateZero(c.x)
	c.updateNegative(c.x)
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
func (c *cpu) tay(bus *sysBus, mode addressingMode, addr uint16) {
	c.y = c.a
	c.updateZero(c.y)
	c.updateNegative(c.y)
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
func (c *cpu) tsx(bus *sysBus, mode addressingMode, addr uint16) {
	c.x = c.s
	c.updateZero(c.x)
	c.updateNegative(c.x)
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
func (c *cpu) txa(bus *sysBus, mode addressingMode, addr uint16) {
	c.a = c.x
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) txs(bus *sysBus, mode addressingMode, addr uint16) {
	c.s = c.x
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
func (c *cpu) tya(bus *sysBus, mode addressingMode, addr uint16) {
	c.a = c.y
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) pha(bus *sysBus, mode addressingMode, addr uint16) {
	c.push(bus, c.a)
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
func (c *cpu) php(bus *sysBus, mode addressingMode, addr uint16) {
	status := c.p
	status |= brk
	status |= unused
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
func (c *cpu) pla(bus *sysBus, mode addressingMode, addr uint16) {
	// TODO: this cycle should be spent in pull. read the docs
	c.clock()
	a := c.pull(bus)

	c.a = a
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) plp(bus *sysBus, mode addressingMode, addr uint16) {

	// TODO: this cycle should be spent in pull. read the docs
	c.clock()
	p := c.pull(bus)

	c.p = status(p)
	c.p &^= brk //TODO figure out if we can just turn it off instead of actually ignoring
	c.p |= unused
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
func (c *cpu) dec(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doDec(v))
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
func (c *cpu) dex(bus *sysBus, mode addressingMode, addr uint16) {
	c.x = c.doDec(c.x)
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
func (c *cpu) dey(bus *sysBus, mode addressingMode, addr uint16) {
	c.y = c.doDec(c.y)
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
func (c *cpu) inc(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doInc(v))
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
func (c *cpu) inx(bus *sysBus, mode addressingMode, addr uint16) {
	c.x = c.doInc(c.x)
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
func (c *cpu) iny(bus *sysBus, mode addressingMode, addr uint16) {
	c.y = c.doInc(c.y)
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
func (c *cpu) adc(bus *sysBus, mode addressingMode, addr uint16) {
	c.doAdd(c.read(bus, addr))
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
func (c *cpu) sbc(bus *sysBus, mode addressingMode, addr uint16) {
	c.doAdd(c.read(bus, addr) ^ 0xFF)
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
func (c *cpu) asl(bus *sysBus, mode addressingMode, addr uint16) {
	if mode == accumulator {
		c.a = c.doAsl(c.a)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doAsl(v))
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
func (c *cpu) and(bus *sysBus, mode addressingMode, addr uint16) {
	c.a &= c.read(bus, addr)
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) eor(bus *sysBus, mode addressingMode, addr uint16) {
	c.a ^= c.read(bus, addr)
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) lsr(bus *sysBus, mode addressingMode, addr uint16) {
	if mode == accumulator {
		c.a = c.doLsr(c.a)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doLsr(v))
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
func (c *cpu) rol(bus *sysBus, mode addressingMode, addr uint16) {
	if mode == accumulator {
		c.a = c.doRol(c.a)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doRol(v))
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
func (c *cpu) ror(bus *sysBus, mode addressingMode, addr uint16) {
	if mode == accumulator {
		c.a = c.doRor(c.a)
		return
	}

	v := c.read(bus, addr)
	c.write(bus, addr, v)
	c.write(bus, addr, c.doRor(v))
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
func (c *cpu) ora(bus *sysBus, mode addressingMode, addr uint16) {
	c.a |= c.read(bus, addr)
	c.updateZero(c.a)
	c.updateNegative(c.a)
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
func (c *cpu) bit(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)

	c.updateNegative(v)
	c.updateZero(c.a & v)

	if v&0x40 > 0 {
		c.p |= overflow
	} else {
		c.p &^= overflow
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
func (c *cpu) cmp(bus *sysBus, mode addressingMode, addr uint16) {
	c.compare(c.a, c.read(bus, addr))
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
func (c *cpu) cpx(bus *sysBus, mode addressingMode, addr uint16) {
	c.compare(c.x, c.read(bus, addr))
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
func (c *cpu) cpy(bus *sysBus, mode addressingMode, addr uint16) {
	c.compare(c.y, c.read(bus, addr))
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
func (c *cpu) bcc(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&carry > 0 {
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
func (c *cpu) bcs(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&carry == 0 {
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
func (c *cpu) bvc(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&overflow > 0 {
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
func (c *cpu) bvs(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&overflow == 0 {
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
func (c *cpu) beq(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&zero == 0 {
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
func (c *cpu) bne(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&zero > 0 {
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
func (c *cpu) bmi(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&negative == 0 {
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
func (c *cpu) bpl(bus *sysBus, mode addressingMode, addr uint16) {
	if c.p&negative > 0 {
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
func (c *cpu) jmp(bus *sysBus, mode addressingMode, addr uint16) {
	c.pc = addr
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
func (c *cpu) jsr(bus *sysBus, mode addressingMode, addr uint16) {
	c.clock()

	c.pushAddress(bus, c.pc-1)
	c.pc = addr
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
func (c *cpu) rti(bus *sysBus, mode addressingMode, addr uint16) {
	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	p := c.pull(bus)

	c.p = status(p) & ^brk
	c.p |= unused

	c.pc = c.pullAddress(bus)
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
func (c *cpu) rts(bus *sysBus, mode addressingMode, addr uint16) {
	// TODO: this cycle should be spent in pull. read the docs
	c.clock()

	pclo := uint16(c.pull(bus))
	pchi := uint16(c.pull(bus))

	c.clock()
	c.pc = pchi<<8 | pclo + 1
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
func (c *cpu) alr(bus *sysBus, mode addressingMode, addr uint16) {
	c.and(bus, mode, addr)
	c.lsr(bus, accumulator, addr)
}

// Does AND #i, setting N and Z flags based on the result. Then it copies N
// (bit 7) to C. ANC #$FF could be useful for sign-extending, much like
// CMP #$80. ANC #$00 acts like LDA #$00 followed by CLC.
func (c *cpu) anc(bus *sysBus, mode addressingMode, addr uint16) {
	c.and(bus, mode, addr)

	if c.p&negative > 0 {
		c.p |= carry
	} else {
		c.p &^= carry
	}
}

// Similar to AND #i then ROR A, except sets the flags differently. N and Z are
// normal, but C is bit 6 and V is bit 6 xor bit 5. A fast way to perform signed
// division by 4 is: CMP #$80; ARR #$FF; ROR. This can be extended to larger
// powers of two.
func (c *cpu) arr(bus *sysBus, mode addressingMode, addr uint16) {
	c.and(bus, mode, addr)
	c.ror(bus, accumulator, addr)

	if (c.a>>6)&1 > 0 {
		c.p |= carry
	} else {
		c.p &^= carry
	}

	if ((c.a>>6)&1)^((c.a>>5)&1) > 0 {
		c.p |= overflow
	} else {
		c.p &^= overflow
	}
}

// Sets X to {(A AND X) - #value without borrow}, and updates NZC. One might use
// TXA AXS #-element_size to iterate through an array of structures or other
// elements larger than a byte, where the 6502 architecture usually prefers a
// structure of arrays. For example, TXA AXS #$FC could step to the next OAM
// entry or to the next APU channel, saving one byte and four cycles over four
// INXs. Also called SBX.
func (c *cpu) axs(bus *sysBus, mode addressingMode, addr uint16) {
	panic("AXS wat") //SBC without carry void asx()
}

// Shortcut for LDA value then TAX. Saves a byte and two cycles and allows use
// of the X register with the (d),Y addressing mode. Notice that the immediate
// is missing; the opcode that would have been LAX is affected by line noise on
// the data bus. MOS 6502: even the bugs have bugs.
func (c *cpu) lax(bus *sysBus, mode addressingMode, addr uint16) {
	if mode == immediate {
		panic("LAX Immediate")
	}

	c.lda(bus, mode, addr)
	c.tax(bus, mode, addr)
}

// Stores the bitwise AND of A and X. As with STA and STX, no flags are affected
func (c *cpu) sax(bus *sysBus, mode addressingMode, addr uint16) {
	c.write(bus, addr, c.a&c.x)
}

// Equivalent to DEC value then CMP value, except supporting more addressing
// modes. LDA #$FF followed by DCP can be used to check if the decrement
// underflows, which is useful for multi-byte decrements.
func (c *cpu) dcp(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doDec(v)
	c.write(bus, addr, v)
	c.compare(c.a, v)
}

// Equivalent to INC value then SBC value, except supporting more addressing
// modes.
func (c *cpu) isc(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doInc(v)
	c.write(bus, addr, v)
	c.doAdd(v ^ 0xFF)
}

// Equivalent to ROL value then AND value, except supporting more addressing
// modes. LDA #$FF followed by RLA is an efficient way to rotate a variable
// while also loading it in A.
func (c *cpu) rla(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doRol(v)
	c.write(bus, addr, v)

	c.a &= v
	c.updateZero(c.a)
	c.updateNegative(c.a)
}

// Equivalent to ROR value then ADC value, except supporting more addressing
// modes. Essentially this computes A + value / 2, where value is 9-bit and the
// division is rounded up.
func (c *cpu) rra(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doRor(v)
	c.write(bus, addr, v)
	c.doAdd(v)
}

// Equivalent to ASL value then ORA value, except supporting more addressing
// modes. LDA #0 followed by SLO is an efficient way to shift a variable while
// also loading it in A.
func (c *cpu) slo(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doAsl(v)
	c.write(bus, addr, v)

	c.a |= v
	c.updateZero(c.a)
	c.updateNegative(c.a)
}

// Equivalent to LSR value then EOR value, except supporting more addressing
// modes. LDA #0 followed by SRE is an efficient way to shift a variable while
// also loading it in A.
func (c *cpu) sre(bus *sysBus, mode addressingMode, addr uint16) {
	v := c.read(bus, addr)
	c.write(bus, addr, v)

	v = c.doLsr(v)
	c.write(bus, addr, v)

	c.a ^= v
	c.updateZero(c.a)
	c.updateNegative(c.a)
}

func (c *cpu) kil(bus *sysBus, mode addressingMode, addr uint16) {
	panic("KIL NOT IMPLEMENTED")
}
func (c *cpu) xaa(bus *sysBus, mode addressingMode, addr uint16) {
	c.txa(bus, mode, addr)
	c.and(bus, mode, addr)
}
func (c *cpu) ahx(bus *sysBus, mode addressingMode, addr uint16) {
	panic("AHX NOT IMPLEMENTED")
}
func (c *cpu) tas(bus *sysBus, mode addressingMode, addr uint16) {
	panic("TAS NOT IMPLEMENTED")
}
func (c *cpu) shy(bus *sysBus, mode addressingMode, addr uint16) {
	panic("SHY NOT IMPLEMENTED")
}
func (c *cpu) shx(bus *sysBus, mode addressingMode, addr uint16) {
	panic("SHX NOT IMPLEMENTED")
}
func (c *cpu) las(bus *sysBus, mode addressingMode, addr uint16) {
	panic("LAS NOT IMPLEMENTED")
}
