package nes

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	trainerLen = 512
	prgMul     = 1024 * 16
	chrMul     = 1024 * 8
)

const (
	rc1MirrorModeVertical = 1 << iota
	rc1SaveRAM
	rc1Trainer
	rc1FourScreen
)

var (
	inesMagic  = []byte{'N', 'E', 'S', 0x1A}
	errNoMagic = errors.New("nes: invalid magic in header")
)

type mirrorMode int

const (
	horizontal mirrorMode = iota
	vertical
	quad
)

type cartridge struct {
	mirrorMode mirrorMode
	saveRAM    bool //TODO
	fourScreen bool
	mapper     byte

	trainer []byte
	prg     []byte
	chr     []byte
}

func loadRom(r io.Reader) (*cartridge, error) {
	type header struct {
		// String "NES^Z" used to recognize .NES files.
		Magic [4]byte

		// Number of 16kB ROM banks.
		ROMBanks byte

		// Number of 8kB VROM banks.
		CHROMBanks byte

		// 76543210
		// ||||||||
		// |||||||+- Mirroring: 0: horizontal (vertical arrangement)
		// |||||||                 (CIRAM A10 = PPU A11)
		// |||||||              1: vertical (horizontal arrangement)
		// |||||||                 (CIRAM A10 = PPU A10)
		// ||||||+-- 1: Cartridge contains battery-backed
		// ||||||       PRG RAM ($6000-7FFF) or other persistent memory
		// |||||+--- 1: 512-byte trainer at $7000-$71FF (stored before PRG data)
		// ||||+---- 1: Ignore mirroring control or above mirroring bit;
		// ||||         instead provide four-screen VRAM
		// ++++----- Lower nybble of mapper number
		ROMControl1 byte

		// 76543210
		// ||||||||
		// |||||||+- VS Unisystem
		// ||||||+-- PlayChoice10, 8KB of Hint Screen data stored after CHR data
		// ||||++--- If equal to 2, flags 8-15 are in NES 2.0 format
		// ++++----- Upper nybble of mapper number
		ROMControl2 byte

		// Number of 8kB RAM banks. For compatibility with the previous
		// versions of the .NES format, assume 1x8kB RAM page when this
		// byte is zero.
		PRGRAMSize byte

		// Reserved, must be zeroes!
		_ [7]byte
	}
	var h header
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("nes: unable to read header: %s", err)
	}

	if !bytes.Equal(h.Magic[:], inesMagic) {
		return nil, errNoMagic
	}

	var trainer []byte
	if h.ROMControl1&rc1Trainer > 0 {
		trainer = make([]byte, trainerLen)
		if _, err := io.ReadFull(r, trainer); err != nil {
			return nil, err
		}
	}

	prg := make([]byte, int(h.ROMBanks)*prgMul)
	if _, err := io.ReadFull(r, prg); err != nil {
		return nil, err
	}

	var chr []byte
	if h.CHROMBanks == 0 {
		chr = make([]byte, chrMul)
	} else {
		chr = make([]byte, int(h.CHROMBanks)*chrMul)
		if _, err := io.ReadFull(r, chr); err != nil {
			return nil, err
		}
	}

	mirrorMode := horizontal
	if h.ROMControl1&rc1MirrorModeVertical > 0 {
		mirrorMode = vertical
	}

	fourScreen := h.ROMControl1&rc1FourScreen > 0
	if fourScreen {
		mirrorMode = quad
	}

	saveRAM := h.ROMControl1&rc1SaveRAM > 0

	mapper := h.ROMControl1>>4 | (h.ROMControl2 & 0xF0)

	return &cartridge{
		mirrorMode: mirrorMode,
		saveRAM:    saveRAM,
		trainer:    trainer,
		fourScreen: fourScreen,
		mapper:     mapper,
		prg:        prg,
		chr:        chr,
	}, nil
}

func (c *cartridge) read(address uint16) byte {
	switch {
	case address < 0x2000:
		// fmt.Printf("%04X\n", address)
		return c.chr[address]
	case address >= 0x8000:
		return c.prg[int(address-0x8000)%len(c.prg)]
	case address >= 0x6000:
		// TODO: SRAM
	default:
		log.Fatalf("unhandled cartridge read at address: 0x%04X", address)
	}
	return 0

}

func (c *cartridge) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// c.CHR[address] = value
	case address >= 0x8000:
		// c.PRG[int(address-0x8000)%len(c.PRG)] = value
	case address >= 0x6000:
		// TODO: SRAM
	default:
		log.Fatalf("unhandled cartridge write at address: 0x%04X", address)
	}

}
