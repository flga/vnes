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
type SysBus struct {
	Cartridge *Cartridge
	RAM       *RAM
	CPU       *CPU
	APU       *APU
	PPU       *PPU
	Ctrl1     *Controller
}

func (bus *SysBus) Read(address uint16) byte {
	if address < 0x2000 {
		return bus.RAM.Read(address)
	}

	if address >= 0x2000 && address <= 0x3FFF {
		return bus.PPU.ReadPort(address, bus.CPU)
	}

	if address == 0x4015 {
		return byte(bus.APU.ReadPort(address))
	}

	if address == 0x4016 {
		return byte(bus.Ctrl1.Read())
	}

	if address == 0x4017 {
		// TODO: controller 2
		return 0
	}

	if address == 0x4014 {
		return bus.PPU.ReadPort(address, bus.CPU)
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
		return bus.Cartridge.Read(address)
	}

	panic("erm...") //TODO
}

func (bus *SysBus) Write(address uint16, v byte) {
	if address < 0x2000 {
		bus.RAM.Write(address, v)
		return
	}

	if address < 0x4000 {
		bus.PPU.WritePort(address, v, bus.CPU)
		return
	}

	if address == 0x4014 {
		bus.PPU.WritePort(address, v, bus.CPU)
		return
	}

	if address < 0x4014 || address == 0x4015 || address == 0x4017 {
		bus.APU.WritePort(address, v)
		return
	}

	if address == 0x4016 {
		bus.Ctrl1.Write(v)
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
		bus.Cartridge.Write(address, v)
		return
		// bus.PrgROM[int(address-0x8000)%len(bus.PrgROM)] = v
		// return
	}
}

func (bus *SysBus) ReadAddress(address uint16) (value uint16, hi byte, lo byte) {
	lo = bus.Read(address)
	hi = bus.Read(address + 1)
	return uint16(hi)<<8 | uint16(lo), hi, lo
}

func (bus *SysBus) WriteAddress(address uint16, v uint16) {
	lo := byte(v & 0x00FF)
	hi := byte(v & 0xFF00 >> 8)
	bus.Write(address, lo)
	bus.Write(address+1, hi)
}
