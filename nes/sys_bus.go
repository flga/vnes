package nes

// ╔═════════════════╤═══════╤═════════════════════════╤═══════════╗
// ║ Address Range   │ Size  │ Purpose                 │ Kind      ║
// ╠═════════════════╪═══════╪═════════════════════════╪═══════════╣
// ║ 0xC000 - 0xFFFF │ 16384 │ PRG-ROM UPPER BANK      │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤  PRG ROM  ║
// ║ 0x8000 - 0xBFFF │ 16384 │ PRG-ROM LOWER BANK      │           ║
// ╠═════════════════╪═══════╪═════════════════════════╪═══════════╣
// ║ 0x6000 - 0x7FFF │ 8192  │ SRAM                    │   SRAM    ║
// ╠═════════════════╪═══════╪═════════════════════════╪═══════════╣
// ║ 0x4020 - 0x5FFF │ 8160  │ EXPANSION ROM           │  EXP ROM  ║
// ╠═════════════════╪═══════╪═════════════════════════╪═══════════╣
// ║ 0x4000 - 0x401F │ 32    │ APU / I/0 REGISTERS     │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤           ║
// ║ 0x2008 - 0x3FFF │ 8184  │ MIRRORS 0x2000 - 0x2007 │  I/O REG  ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤           ║
// ║ 0x2000 - 0x2007 │ 8     │ PPU REGISTERS           │           ║
// ╠═════════════════╪═══════╪═════════════════════════╪═══════════╣
// ║ 0x1800 - 0x1FFF │ 2048  │                         │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┤                         │           ║
// ║ 0x1000 - 0x17FF │ 2048  │ MIRRORS 0x0000 - 0x07FF │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┤                         │           ║
// ║ 0x0800 - 0x0FFF │ 2048  │                         │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤    RAM    ║
// ║ 0x0200 - 0x07FF │ 1536  │ RAM                     │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤           ║
// ║ 0x0100 - 0x01FF │ 256   │ STACK                   │           ║
// ╟╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤           ║
// ║ 0x0000 - 0x00FF │ 256   │ ZERO PAGE               │           ║
// ╚═════════════════╧═══════╧═════════════════════════╧═══════════╝
type sysBus struct {
	cartridge *cartridge
	ram       *ram
	cpu       *cpu
	apu       *apu
	ppu       *ppu
	ctrl1     *controller
	ctrl2     *controller
}

func (bus *sysBus) read(address uint16) byte {
	if address < 0x2000 {
		return bus.ram.read(address)
	}

	if address >= 0x2000 && address <= 0x3FFF {
		return bus.ppu.readPort(address, bus.cpu)
	}

	if address == 0x4015 {
		return byte(bus.apu.readPort(address))
	}

	if address == 0x4016 {
		return byte(bus.ctrl1.read())
	}

	if address == 0x4017 {
		return byte(bus.ctrl2.read())
	}

	if address == 0x4014 {
		return bus.ppu.readPort(address, bus.cpu)
	}

	if address < 0x4020 {
		return 0xFF //TODO io registers
	}

	if address < 0x6000 {
		return 0 //TODO exp rom
	}

	if address < 0x8000 {
		return 0 //TODO sram
	}

	if address <= 0xFFFF {
		return bus.cartridge.read(address)
	}

	panic("erm...") //TODO
}

func (bus *sysBus) write(address uint16, v byte) {
	if address < 0x2000 {
		bus.ram.write(address, v)
		return
	}

	if address < 0x4000 {
		bus.ppu.writePort(address, v, bus.cpu)
		return
	}

	if address == 0x4014 {
		bus.ppu.writePort(address, v, bus.cpu)
		return
	}

	if address < 0x4014 || address == 0x4015 || address == 0x4017 {
		bus.apu.writePort(address, v)
		return
	}

	if address == 0x4016 {
		bus.ctrl1.write(v)
		bus.ctrl2.write(v)
		return
	}

	if address < 0x6000 {
		//TODO: ExpROM
		return
	}

	if address < 0x8000 {
		//TODO: SRAM
		return
	}

	if address <= 0xFFFF {
		bus.cartridge.write(address, v)
		return
		// bus.PrgROM[int(address-0x8000)%len(bus.PrgROM)] = v
		// return
	}
}

func (bus *sysBus) readAddress(address uint16) (value uint16, hi byte, lo byte) {
	lo = bus.read(address)
	hi = bus.read(address + 1)
	return uint16(hi)<<8 | uint16(lo), hi, lo
}

func (bus *sysBus) writeAddress(address uint16, v uint16) {
	lo := byte(v & 0x00FF)
	hi := byte(v & 0xFF00 >> 8)
	bus.write(address, lo)
	bus.write(address+1, hi)
}
