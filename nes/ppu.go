package nes

import (
	"fmt"
	"image/color"
	"log"
)

// ╔═════════════════╤═══════╤════════════════════════════╤════════════════╗
// ║ Address Range   │ Size  │ Purpose                    │ Kind           ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x0000 - 0x0FFF │ 4096  │ Pattern Table #0           │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Pattern Tables ║
// ║ 0x1000 - 0x1FFF │ 4096  │ Pattern Table #1           │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x2000 - 0x23BF │ 960   │ Name Table #0              │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Name Table #0  ║
// ║ 0x23C0 - 0x23FF │ 64    │ Attribute Table #0         │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x2400 - 0x27BF │ 960   │ Name Table #1              │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Name Table #1  ║
// ║ 0x27C0 - 0x27FF │ 64    │ Attribute Table #1         │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x2800 - 0x2BBF │ 960   │ Name Table #2              │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Name Table #2  ║
// ║ 0x2BC0 - 0x2BFF │ 64    │ Attribute Table #2         │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x2C00 - 0x2FBF │ 960   │ Name Table #3              │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Name Table #3  ║
// ║ 0x2FC0 - 0x2FFF │ 64    │ Attribute Table #3         │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x3000 - 0x3EFF │ 3840  │ Mirror of 0x2000-0x2EFF    │ Mirror         ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x3F00          │ 1     │ Universal background color │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F01 - 0x3F03 │ 3     │ Background palette #0      │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F04          │ 1     │ ??                         │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F05 - 0x3F07 │ 3     │ Background palette #1      │ Background     ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Palette        ║
// ║ 0x3F08          │ 1     │ ??                         │ Data           ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F09 - 0x3F0B │ 3     │ Background palette #2      │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F0C          │ 1     │ ??                         │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F0D - 0x3F0F │ 3     │ Background palette #3      │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x3F10          │ 1     │ Mirror of 0x3F00           │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F11 - 0x3F13 │ 3     │ Sprite palette #0          │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F14          │ 1     │ Mirror of 0x3F04           │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F15 - 0x3F17 │ 3     │ Sprite palette #1          │ Sprite         ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Palette        ║
// ║ 0x3F18          │ 1     │ Mirror of 0x3F08           │ Data           ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F19 - 0x3F1B │ 3     │ Sprite palette #2          │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F1C          │ 1     │ Mirror of 0x3F0C           │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤                ║
// ║ 0x3F1D - 0x3F1F │ 3     │ Sprite palette #3          │                ║
// ╠═════════════════╪═══════╪════════════════════════════╪════════════════╣
// ║ 0x3F20 - 0x3FFF │ 224   │ Mirrors of 0x3F00 - 0x3F1F │                ║
// ║╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌┼╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┤ Mirrors        ║
// ║ 0x4000 - 0xFFFF │ 49152 │ Mirrors of 0x0000 - 0x3FFF │                ║
// ╚═════════════════╧═══════╧════════════════════════════╧════════════════╝

var palette [64]color.RGBA = [64]color.RGBA{
	color.RGBA{0x7C, 0x7C, 0x7C, 0xFF}, color.RGBA{0x00, 0x00, 0xFC, 0xFF},
	color.RGBA{0x00, 0x00, 0xBC, 0xFF}, color.RGBA{0x44, 0x28, 0xBC, 0xFF},
	color.RGBA{0x94, 0x00, 0x84, 0xFF}, color.RGBA{0xA8, 0x00, 0x20, 0xFF},
	color.RGBA{0xA8, 0x10, 0x00, 0xFF}, color.RGBA{0x88, 0x14, 0x00, 0xFF},
	color.RGBA{0x50, 0x30, 0x00, 0xFF}, color.RGBA{0x00, 0x78, 0x00, 0xFF},
	color.RGBA{0x00, 0x68, 0x00, 0xFF}, color.RGBA{0x00, 0x58, 0x00, 0xFF},
	color.RGBA{0x00, 0x40, 0x58, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
	color.RGBA{0x00, 0x00, 0x00, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
	color.RGBA{0xBC, 0xBC, 0xBC, 0xFF}, color.RGBA{0x00, 0x78, 0xF8, 0xFF},
	color.RGBA{0x00, 0x58, 0xF8, 0xFF}, color.RGBA{0x68, 0x44, 0xFC, 0xFF},
	color.RGBA{0xD8, 0x00, 0xCC, 0xFF}, color.RGBA{0xE4, 0x00, 0x58, 0xFF},
	color.RGBA{0xF8, 0x38, 0x00, 0xFF}, color.RGBA{0xE4, 0x5C, 0x10, 0xFF},
	color.RGBA{0xAC, 0x7C, 0x00, 0xFF}, color.RGBA{0x00, 0xB8, 0x00, 0xFF},
	color.RGBA{0x00, 0xA8, 0x00, 0xFF}, color.RGBA{0x00, 0xA8, 0x44, 0xFF},
	color.RGBA{0x00, 0x88, 0x88, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
	color.RGBA{0x00, 0x00, 0x00, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
	color.RGBA{0xF8, 0xF8, 0xF8, 0xFF}, color.RGBA{0x3C, 0xBC, 0xFC, 0xFF},
	color.RGBA{0x68, 0x88, 0xFC, 0xFF}, color.RGBA{0x98, 0x78, 0xF8, 0xFF},
	color.RGBA{0xF8, 0x78, 0xF8, 0xFF}, color.RGBA{0xF8, 0x58, 0x98, 0xFF},
	color.RGBA{0xF8, 0x78, 0x58, 0xFF}, color.RGBA{0xFC, 0xA0, 0x44, 0xFF},
	color.RGBA{0xF8, 0xB8, 0x00, 0xFF}, color.RGBA{0xB8, 0xF8, 0x18, 0xFF},
	color.RGBA{0x58, 0xD8, 0x54, 0xFF}, color.RGBA{0x58, 0xF8, 0x98, 0xFF},
	color.RGBA{0x00, 0xE8, 0xD8, 0xFF}, color.RGBA{0x78, 0x78, 0x78, 0xFF},
	color.RGBA{0x00, 0x00, 0x00, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
	color.RGBA{0xFC, 0xFC, 0xFC, 0xFF}, color.RGBA{0xA4, 0xE4, 0xFC, 0xFF},
	color.RGBA{0xB8, 0xB8, 0xF8, 0xFF}, color.RGBA{0xD8, 0xB8, 0xF8, 0xFF},
	color.RGBA{0xF8, 0xB8, 0xF8, 0xFF}, color.RGBA{0xF8, 0xA4, 0xC0, 0xFF},
	color.RGBA{0xF0, 0xD0, 0xB0, 0xFF}, color.RGBA{0xFC, 0xE0, 0xA8, 0xFF},
	color.RGBA{0xF8, 0xD8, 0x78, 0xFF}, color.RGBA{0xD8, 0xF8, 0x78, 0xFF},
	color.RGBA{0xB8, 0xF8, 0xB8, 0xFF}, color.RGBA{0xB8, 0xF8, 0xD8, 0xFF},
	color.RGBA{0x00, 0xFC, 0xFC, 0xFF}, color.RGBA{0xF8, 0xD8, 0xF8, 0xFF},
	color.RGBA{0x00, 0x00, 0x00, 0xFF}, color.RGBA{0x00, 0x00, 0x00, 0xFF},
}

const (
	ppuCtrlAddr   uint16 = 0x2000
	ppuMaskAddr   uint16 = 0x2001
	ppuStatusAddr uint16 = 0x2002
	oamAddrAddr   uint16 = 0x2003
	oamDataAddr   uint16 = 0x2004
	ppuScrollAddr uint16 = 0x2005
	ppuAddrAddr   uint16 = 0x2006
	ppuDataAddr   uint16 = 0x2007
	oamDmaAddr    uint16 = 0x4014
)

// VPHB SINN
// |||| ||||
// |||| ||++- Base nametable address
// |||| ||    (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
// |||| |+--- VRAM address increment per CPU read/write of PPUDATA
// |||| |     (0: add 1, going across; 1: add 32, going down)
// |||| +---- Sprite pattern table address for 8x8 sprites
// ||||       (0: $0000; 1: $1000; ignored in 8x16 mode)
// |||+------ Background pattern table address (0: $0000; 1: $1000)
// ||+------- Sprite size (0: 8x8 pixels; 1: 8x16 pixels)
// |+-------- PPU master/slave select
// |          (0: read backdrop from EXT pins; 1: output color on EXT pins)
// +--------- Generate an NMI at the start of the
//            vertical blanking interval (0: off; 1: on)
type ppuCtrl byte

const (
	// NametableAddress - VRAM address
	// 0 = $2000
	// 1 = $2400
	// 2 = $2800
	// 3 = $2C00
	nametableAddress ppuCtrl = 3

	// PPU Address Increment
	// 0 = Increment by 1
	// 1 = Increment by 32
	addressIncrement ppuCtrl = 1 << iota * 2

	// SpritePatternTableAddress - VRAM address
	// 0 = $0000
	// 1 = $1000
	spritePatternTableAddress

	// Background Pattern Table Address - VRAM address
	// 0 = $0000
	// 1 = $1000
	backgroundPatternTableAddress

	// SpriteSize
	// 0 = 8x8
	// 1 = 8x16
	spriteSize

	// MasterSlaveSelect - PPU Master/Slave Selection --+   Always write 0
	// 0 = Receive EXTBG                              +-- in unmodified
	// 1 = Send EXTBG                               --+   Control Deck
	masterSlaveSelect

	// GenerateNMI - Execute NMI on VBlank
	// 0 = Disabled
	// 1 = Enabled
	generateNMI
)

// BGRs bMmG
// |||| ||||
// |||| |||+- Greyscale (0: normal color, 1: produce a greyscale display)
// |||| ||+-- 1: Show background in leftmost 8 pixels of screen, 0: Hide
// |||| |+--- 1: Show sprites in leftmost 8 pixels of screen, 0: Hide
// |||| +---- 1: Show background
// |||+------ 1: Show sprites
// ||+------- Emphasize red
// |+-------- Emphasize green
// +--------- Emphasize blue
type ppuMask byte

const (
	// Greyscale - Display Type
	// 0 = Colour display
	// 1 = Monochrome display (all palette values ANDed with $30)
	greyscale ppuMask = 1 << iota

	// BackgroundClipping
	// 0 = BG invisible in left 8-pixel column
	// 1 = No clipping
	backgroundClipping

	// SpriteClipping
	// 0 = Sprites invisible in left 8-pixel column
	// 1 = No clipping
	spriteClipping

	// ShowBackground - Background Visibility
	// 0 = Background not displayed
	// 1 = Background visible
	showBackground

	// ShowSprites - Sprite Visibility
	// 0 = Sprites not displayed
	// 1 = Sprites visible
	showSprites

	emphasizeRed
	emphasizeGreen
	emphasizeBlue
)

// VSO. ....
// |||| ||||
// |||+-++++- Least significant bits previously written into a PPU register
// |||        (due to register not being updated for this address)
// ||+------- Sprite overflow. The intent was for this flag to be set
// ||         whenever more than eight sprites appear on a scanline, but a
// ||         hardware bug causes the actual behavior to be more complicated
// ||         and generate false positives as well as false negatives; see
// ||         PPU sprite evaluation. This flag is set during sprite
// ||         evaluation and cleared at dot 1 (the second dot) of the
// ||         pre-render line.
// |+-------- Sprite 0 Hit.  Set when a nonzero pixel of sprite 0 overlaps
// |          a nonzero background pixel; cleared at dot 1 of the pre-render
// |          line.  Used for raster timing.
// +--------- Vertical blank has started (0: not in vblank; 1: in vblank).
//            Set at dot 1 of line 241 (the line *after* the post-render
//            line); cleared after reading $2002 and at dot 1 of the
//            pre-render line.
type ppuStatus byte

const (
	// SpriteOverflow - Scanline Sprite Count
	// 0 = No scanline with more than eight (8) sprites
	// 1 = At least one line with more than 8 sprites since end of VBlank
	spriteOverflow ppuStatus = 0x20 << iota

	// Sprite0Hit
	// 0 = Sprite #0 not found
	// 1 = PPU has hit Sprite #0 since end of VBlank
	sprite0Hit

	// VerticalBlank
	// 0 = Not occuring
	// 1 = In VBlank
	verticalBlank
)

type ppu struct {
	cartridge *cartridge

	ctrl           ppuCtrl   // 0x2000 PPUCTRL
	mask           ppuMask   // 0x2001 PPUMASK
	status         ppuStatus // 0x2002 PPUSTATUS
	oamAddress     byte      // 0x2003 OAMADDR
	oamData        [256]byte // 0x2004 OAMDATA
	spritesInRange byte
	oamDataBuf     byte
	// secondaryOAMAddress byte
	secondaryOAMData [32]byte

	readBuffer byte // 0x2007 PPUDATA

	dot      int
	scanline int
	frame    uint64

	paletteData [32]byte
	nametable0  [1024]byte
	nametable1  [1024]byte
	nametable2  [1024]byte
	nametable3  [1024]byte

	// Current VRAM address (15 bits)
	v uint16
	// Temporary VRAM address (15 bits); can also be thought of as the address
	// of the top left onscreen tile.
	t uint16
	// Fine X scroll (3 bits)
	x byte
	// First or second write toggle (1 bit)
	w byte
	// Even/odd frame
	f byte

	addressBus  uint16
	registerBus byte

	nametableByte byte // Nametable byte
	attributeByte byte // Attribute table byte
	lowTileByte   byte // Tile bitmap low
	highTileByte  byte // Tile bitmap high

	lowTileRegister  uint16
	highTileRegister uint16
	lowAttrRegister  uint16
	highAttrRegister uint16

	sprite0Next bool
	nmiSent     bool
	suppressNMI bool

	// buffer *image.RGBA
	buffer []byte
}

func newPpu() *ppu {
	return &ppu{
		buffer: make([]byte, 256*240*4),
	}
}

func (p *ppu) spritePixel() (pixel, color, priority byte, spriteZero bool) {
	// TODO: 16px sprites
	outputX := byte(p.dot - 1)
	if p.mask&showSprites == 0 || (outputX < 8 && p.mask&spriteClipping == 0) {
		return 0, 0, 0, false
	}

	spriteHeight := uint16(p.spriteHeight())

	for i := byte(0); i < p.spritesInRange; i++ {
		y := p.secondaryOAMData[i*4] + 1 //TODO
		pattern := uint16(p.secondaryOAMData[i*4+1])
		attr := p.secondaryOAMData[i*4+2]
		x := p.secondaryOAMData[i*4+3]

		pal := attr & 0x03 << 2
		priority := attr >> 5 & 0x01
		flipX := attr>>6&0x01 > 0
		flipY := attr>>7&0x01 > 0

		if outputX < x || outputX > x+7 {
			continue
		}

		patternTable := p.spriteTable(pattern)
		patternY := uint16(p.scanline - int(y))
		patternX := outputX - x

		if !flipX {
			patternX = 7 - patternX
		}

		if flipY {
			patternY = spriteHeight - 1 - patternY
		}

		if patternY > 7 { // top sprite
			patternY += 8
		}

		if spriteHeight == 16 {
			pattern &= 0xFE
		}

		patternLo := p.read(patternTable + pattern*0x10 + patternY)
		patternHi := p.read(patternTable + pattern*0x10 + patternY + 8)

		pixLo := patternLo >> patternX & 0x01
		pixHi := patternHi >> patternX & 0x01 << 1

		pixel = pixLo | pixHi
		color = pixel | 0x10 | pal

		if pixel == 0 {
			continue
		}

		return pixel, color, priority, p.sprite0Next && i == 0
	}

	return 0, 0, 0, false
}

func (p *ppu) bgPixel() (pixel, color byte) {
	x := p.dot - 1

	if p.mask&showBackground == 0 || (x < 8 && p.mask&backgroundClipping == 0) {
		return 0, 0
	}

	bgPixelLo := byte(p.lowTileRegister >> (15 - p.x) & 0x1)
	bgPixelHi := byte(p.highTileRegister >> (15 - p.x) & 0x1)

	bgAttrLo := byte(p.lowAttrRegister >> (15 - p.x) & 0x1)
	bgAttrHi := byte(p.highAttrRegister >> (15 - p.x) & 0x1)
	attr := bgAttrHi<<1 | bgAttrLo

	pixel = bgPixelHi<<1 | bgPixelLo
	color = pixel | attr<<2
	return pixel, color
}

func (p *ppu) render() {
	bgPixel, bgColor := p.bgPixel()
	spPixel, spColor, priority, szero := p.spritePixel()

	// BG pixel	Sprite pixel	Priority	Output
	// 0			0				X			BG ($3F00)
	// 0			1-3				X			Sprite
	// 1-3			0				X			BG
	// 1-3			1-3				0			Sprite
	// 1-3			1-3				1			BG
	var col byte
	switch {
	case bgPixel == 0 && spPixel == 0:
		col = 0

	case bgPixel == 0 && spPixel != 0:
		col = spColor

	case bgPixel != 0 && spPixel == 0:
		col = bgColor

	case bgPixel != 0 && spPixel != 0 && priority == 0:
		// TODO: sprite 0 hit needs to check more stuff
		if szero && p.status&sprite0Hit == 0 && p.dot-1 != 255 {
			p.status |= sprite0Hit
		}
		col = spColor

	case bgPixel != 0 && spPixel != 0 && priority == 1:
		// TODO: sprite 0 hit needs to check more stuff
		if szero && p.status&sprite0Hit == 0 && p.dot-1 != 255 {
			p.status |= sprite0Hit
		}
		col = bgColor
	}

	paletteIdx := p.readPalette(uint16(col))
	// p.buffer.SetRGBA(p.dot-1, p.scanline, palette[paletteIdx])
	c := palette[paletteIdx]
	pos := p.scanline*256*4 + (p.dot-1)*4
	p.buffer[pos+0] = c.R
	p.buffer[pos+1] = c.G
	p.buffer[pos+2] = c.B
	p.buffer[pos+3] = c.A
}

func (p *ppu) tick(cpu *cpu) {
	renderingEnabled := p.renderingEnabled()
	preRender := p.scanline == 261
	visibleFrame := p.scanline < 240
	visibleDot := p.dot > 0 && p.dot < 257
	invisibleDot := p.dot > 320 && p.dot < 341
	opFrame := preRender || visibleFrame
	doOp := renderingEnabled && opFrame
	fetchDot := visibleDot || invisibleDot
	shiftDot := (p.dot > 0 && p.dot < 257) || (p.dot > 320 && p.dot < 337)

	// render
	if renderingEnabled && visibleFrame && visibleDot {
		p.render()
	}

	// shift
	if doOp && shiftDot {
		p.lowTileRegister <<= 1
		p.highTileRegister <<= 1
		p.lowAttrRegister <<= 1
		p.highAttrRegister <<= 1
	}

	// fetch
	if doOp && fetchDot {
		switch (p.dot - 1) % 8 {

		case 0:
			//load nametable address
			p.addressBus = 0x2000 | (p.v & 0x0FFF)
		case 1:
			// fetch nametable byte
			p.nametableByte = p.read(p.addressBus)

		case 2:
			// load attribute address
			p.addressBus = 0x23C0 | (p.v & 0x0C00) | ((p.v >> 4) & 0x38) | ((p.v >> 2) & 0x07)
		case 3:
			// fetch attribute byte

			// The quadrant is composed of GB and can be either 0, 1, 2 or 3.
			// Since each attribute is 2 bits, for quadrant 0 we shift the
			// attribute by 0, q1 we shift by 2, q2 we shift by 4, etc.
			g := p.v & 0x40 >> 5
			b := p.v & 0x02 >> 1
			shift := (g | b) << 1
			p.attributeByte = p.read(p.addressBus) >> shift & 0x03

		case 4:
			// load low tile address
			fineY := p.v >> 12 & 0x07 // pixel Y
			p.addressBus = p.backgroundTable() + uint16(p.nametableByte)*16 + fineY
		case 5:
			// fetch low tile byte
			p.lowTileByte = p.read(p.addressBus)

		case 6:
			// load high tile address
			fineY := p.v >> 12 & 0x07 // pixel Y
			p.addressBus = p.backgroundTable() + uint16(p.nametableByte)*16 + fineY + 8
		case 7:
			// fetch high tile byte
			p.highTileByte = p.read(p.addressBus)

			// load shift registers
			p.highTileRegister = p.highTileRegister&0xFF00 | uint16(p.highTileByte)
			p.lowTileRegister = p.lowTileRegister&0xFF00 | uint16(p.lowTileByte)

			p.highAttrRegister |= uint16(p.attributeByte >> 1 * 0xFF)
			p.lowAttrRegister |= uint16(p.attributeByte & 0x1 * 0xFF)

			p.incrementX()
		}
	}

	// update
	switch {
	case doOp && p.dot == 256:
		p.incrementY()
	case doOp && p.dot == 257:
		p.copyX()
	case renderingEnabled && preRender && p.dot >= 280 && p.dot <= 304:
		p.copyY()
	}

	if renderingEnabled && visibleFrame {
		p.evaluateSprites()
	} else {
		p.spritesInRange = 0
	}

	// flags
	switch {
	case p.scanline == 241 && p.dot == 1:
		p.status |= verticalBlank
		if !p.suppressNMI && p.ctrl&generateNMI > 0 {
			cpu.trigger(nmi)
		}

	case preRender && p.dot == 1:
		p.status &^= spriteOverflow
		p.status &^= sprite0Hit
		p.status &^= verticalBlank
	}

	if p.dot == 255 && p.scanline == 239 {
		p.frame++
	}

	// tick
	switch {
	case p.dot == 340 && preRender:
		p.dot = 0
		if p.frame&1 == 1 && p.mask&showBackground > 0 {
			p.dot = 1
		}
		p.scanline = 0
	case p.dot == 340:
		p.dot = 0
		p.scanline++
	default:
		p.dot++
	}
}

func (p *ppu) evaluateSprites() {
	// Cycles 1-64: Secondary OAM (32-byte buffer for current sprites on
	// scanline) is initialized to $FF - attempting to read $2004 will return
	// $FF. Internally, the clear operation is implemented by reading from the
	// OAM and writing into the secondary OAM as usual, only a signal is active
	// that makes the read always return $FF.
	// TODO: emulate cycles

	// if p.dot > 0 && p.dot < 65 {
	// 	// TODO: reads from 2004 in this range should return FF
	// 	p.oamDataBuf = 0xFF
	// 	p.secondaryOAMData[(p.dot-1)>>1] = p.oamDataBuf
	// 	return
	// }

	spriteHeight := p.spriteHeight()

	if p.dot == 256 {
		p.spritesInRange = 0
		p.sprite0Next = false
		secAddress := 0

		for i := 0; i < 64; i++ {
			y := p.oamData[i*4]
			row := p.scanline - int(y) //TODO

			// sprite not in range
			if row < 0 || row >= spriteHeight {
				continue
			}

			if p.spritesInRange < 8 {
				p.secondaryOAMData[secAddress*4] = p.oamData[i*4]
				p.secondaryOAMData[secAddress*4+1] = p.oamData[i*4+1]
				p.secondaryOAMData[secAddress*4+2] = p.oamData[i*4+2]
				p.secondaryOAMData[secAddress*4+3] = p.oamData[i*4+3]
				secAddress++
			}
			if i == 0 {
				p.sprite0Next = true
			}
			p.spritesInRange++

		}
		if p.spritesInRange > 8 {
			p.spritesInRange = 8
			p.status |= spriteOverflow
		}
	}
}

// func (p *ppu) buffer() *image.RGBA {
// 	return p.buffer
// }

func (p *ppu) readPort(address uint16, c *cpu) byte {
	if address < 0x4000 {
		address = 0x2000 + address%0x08
	}

	switch address {
	case ppuStatusAddr: // $2002
		result := p.registerBus&0x1F | byte(p.status)
		p.status &^= verticalBlank

		if p.scanline == 241 && p.dot <= 2 {
			p.suppressNMI = true
			result &^= byte(verticalBlank)
			c.trigger(none)
		} else {
			p.suppressNMI = false
		}
		// w:                  = 0
		p.w = 0
		return result

	case oamDataAddr: // $2004
		v := p.oamData[p.oamAddress]
		p.registerBus = v
		return v

	case ppuDataAddr: // $2007
		var ret byte
		if p.v >= 0x3F00 && p.v <= 0x3FFF {
			ret = p.read(p.v)
			// When you read from palette memory, the read buffer gets the contents
			// of the PPU address. Meaning if you read from $3F00 ... $3FFF, the
			// read buffer will get the value that is stored in $2F00 ... $2FFF,
			// because of PPU memory mirrorring.
			p.readBuffer = p.read(p.v - 0x1000)
		} else if p.v < 0x3F00 {
			ret = p.readBuffer
			p.readBuffer = p.read(p.v)
		}

		p.incrementV()

		p.registerBus = ret
		return ret
	}

	log.Printf("unexpected ppu port read: 0x%04X", address)
	// panic(fmt.Sprintf("unexpected ppu port read: 0x%04X", address))
	return byte(p.registerBus)
}

func (p *ppu) writePort(address uint16, value byte, cpu *cpu) {
	if address < 0x4000 {
		address = 0x2000 + address%0x08
	}
	p.registerBus = value

	switch address {
	case ppuCtrlAddr: // $2000
		prev := p.ctrl
		p.ctrl = ppuCtrl(value)

		if p.status&verticalBlank > 0 && p.ctrl&generateNMI > 0 && prev&generateNMI == 0 {
			cpu.trigger(nmiNext)
		}

		// t: ....BA.. ........ = d: ......BA
		d := uint16(value)
		p.t = p.t&0xF3FF | d&0x3<<10

	case ppuMaskAddr: // $2001
		// TODO: greyscale
		// TODO: emphasis
		p.mask = ppuMask(value)

	case oamAddrAddr: // $2003
		// TODO: OAMADDR is set to 0 during each of ticks 257-320 (the sprite
		// tile loading interval) of the pre-render and visible scanlines
		p.oamAddress = value

	case oamDataAddr: // $2004
		// For emulation purposes, it is probably best to completely ignore
		// writes during rendering.
		if p.currentlyRendering() {
			return
		}
		p.oamData[p.oamAddress] = value
		p.oamAddress++

	case ppuScrollAddr: // $2005
		d := uint16(value)
		if p.w == 0 {
			// t: ........ ...HGFED = d: HGFED...
			// x:               CBA = d: .....CBA
			// w:                   = 1
			p.t = p.t&0xFFE0 | d>>3
			p.x = value & 0x07
			p.w = 1
		} else {
			// t: .CBA..HG FED..... = d: HGFEDCBA
			// w:                   = 0
			fineY := d & 0x07 << 12  // CBA   = fine Y scroll
			coarseY := d & 0xF8 << 2 // HGFED = coarse Y scroll
			p.t = p.t&0x8C1F | fineY | coarseY
			p.w = 0
		}

	case ppuAddrAddr: // $2006
		d := uint16(value)
		if p.w == 0 {
			// t: ..FEDCBA ........ = d: ..FEDCBA
			// t: .X...... ........ = 0
			// w:                   = 1
			p.w = 1
			p.t = p.t&0xC0FF | d&0x3F<<8
			p.t &^= 0x4000
		} else {
			// t: ........ HGFEDCBA = d: HGFEDCBA
			// v                    = t
			// w:                   = 0
			p.t = p.t&0xFF00 | d
			p.v = p.t
			p.w = 0
		}

	case ppuDataAddr: // $2007
		p.write(p.v, value)
		p.incrementV()

	case oamDmaAddr: // $4014
		p.oamData[p.oamAddress] = value
		p.oamAddress++

	default:
		log.Printf("unexpected ppu port write: 0x%04X, 0x%02X", address, value)
		// panic(fmt.Sprintf("unexpected ppu port write: 0x%04X, 0x%02X", address, value))
	}
}

func (p *ppu) read(address uint16) byte {
	address %= 0x4000
	switch {
	case address < 0x2000:
		return p.cartridge.read(address)

	case address < 0x3F00:
		return p.readNametable(address)

	case address < 0x4000:
		return p.readPalette(address)

	}

	panic(fmt.Sprintf("unexpected ppu memory read: 0x%04X", address))
}

func (p *ppu) write(address uint16, value byte) {
	address %= 0x4000
	switch {
	case address < 0x2000:
		p.cartridge.write(address, value)

	case address < 0x3F00:
		p.writeNametable(address, value)

	case address < 0x4000:
		p.writePalette(address, value)

	default:
		panic(fmt.Sprintf("unexpected ppu memory write: 0x%04X, 0x%02X", address, value))
	}

}

func (p *ppu) writeDMA(v byte) {
	p.oamData[p.oamAddress] = v
	p.oamAddress++
}

func (p *ppu) readPalette(address uint16) byte {
	switch address {
	case 0x3F10, 0x3F14, 0x3F18, 0x3F1C:
		address -= 0x10
	}
	return p.paletteData[address%32]
}

func (p *ppu) writePalette(address uint16, value byte) {
	switch address {
	case 0x3F10, 0x3F14, 0x3F18, 0x3F1C:
		address -= 0x10
	}
	p.paletteData[address%32] = value
}

func (p *ppu) readNametable(addr uint16) byte {
	switch p.cartridge.mirrorMode {
	case horizontal:
		if addr < 0x2800 {
			return p.nametable0[addr%1024]
		} else {
			return p.nametable2[addr%1024]
		}
	case vertical:
		if addr < 0x2400 || (addr >= 0x2800 && addr < 0x2C00) {
			return p.nametable0[addr%1024]
		} else {
			return p.nametable1[addr%1024]
		}
	}

	return 0
}

func (p *ppu) writeNametable(addr uint16, val byte) {
	switch p.cartridge.mirrorMode {
	case horizontal:
		if addr < 0x2800 {
			p.nametable0[addr%1024] = val
			p.nametable1[addr%1024] = val
		} else {
			p.nametable2[addr%1024] = val
			p.nametable3[addr%1024] = val
		}
	case vertical:
		if addr < 0x2400 {
			p.nametable0[addr%1024] = val
			p.nametable2[addr%1024] = val
		} else {
			p.nametable1[addr%1024] = val
			p.nametable3[addr%1024] = val
		}
	}
}

func (p *ppu) incrementV() {
	if p.ctrl&addressIncrement > 0 {
		p.v += 32
	} else {
		p.v += 1
	}
}

// The coarse X component of v needs to be incremented when the next tile is
// reached. Bits 0-4 are incremented, with overflow toggling bit 10. This means
// that bits 0-4 count from 0 to 31 across a single nametable, and bit 10
// selects the current nametable horizontally.
func (p *ppu) incrementX() {
	coarseX := p.v & 0x001F

	if coarseX == 31 {
		p.v &^= 0x001F // coarse X = 0
		p.v ^= 0x0400  // switch horizontal nametable
		return
	}

	// increment coarse X
	p.v += 1
}

func (p *ppu) copyX() {
	// v: .....F.. ...EDCBA = t: .....F.. ...EDCBA
	p.v = p.v&^0x041F | p.t&0x041F
}

// If rendering is enabled, fine Y is incremented at dot 256 of each scanline,
// overflowing to coarse Y, and finally adjusted to wrap among the nametables
// vertically.
// Bits 12-14 are fine Y.
// Bits 5-9 are coarse Y.
// Bit 11 selects the vertical nametable.
func (p *ppu) incrementY() {

	// if fine Y < 7
	if p.v&0x7000 != 0x7000 {
		p.v += 0x1000 // increment fine Y
		return
	}

	p.v &^= 0x7000 // fine Y = 0

	coarseY := (p.v & 0x03E0) >> 5

	if coarseY == 29 {
		coarseY = 0
		p.v ^= 0x0800 // switch vertical nametable
	} else if coarseY == 31 {
		coarseY = 0
	} else {
		coarseY += 1
	}

	// put coarse Y back into v
	p.v = p.v&^0x03E0 | coarseY<<5
}

func (p *ppu) copyY() {
	// v: .IHGF.ED CBA..... = t: .IHGF.ED CBA.....
	p.v = p.v&^0x7BE0 | p.t&0x7BE0
}

func (p *ppu) backgroundTable() uint16 {
	if p.ctrl&backgroundPatternTableAddress > 0 {
		return 0x1000
	}
	return 0x0000
}

func (p *ppu) spriteTable(pattern uint16) uint16 {
	if p.ctrl&spriteSize > 0 {
		return pattern & 1 * 0x1000
	}

	if p.ctrl&spritePatternTableAddress > 0 {
		return 0x1000
	}

	return 0x0000
}

func (p *ppu) spriteHeight() int {
	if p.ctrl&spriteSize == 0 {
		return 8
	} else {
		return 16
	}
}

func (p *ppu) renderingEnabled() bool {
	return p.mask&showBackground > 0 || p.mask&showSprites > 0
}

func (p *ppu) currentlyRendering() bool {
	return p.renderingEnabled() && (p.scanline < 240 || p.scanline == 261)
}

func (p *ppu) drawPatternTables(buf []byte, paletteNum byte) {
	if p.cartridge == nil {
		return
	}

	draw := func(table uint16, xoffset int) {
		attr := paletteNum << 2

		for y := 0; y < 128; y++ {
			coarseY := y / 8
			fineY := uint16(y % 8)
			for tile := 0; tile < 16; tile++ {
				fineX := tile * 8
				patternNum := uint16(coarseY*16 + tile)

				patternLo := p.read(table + patternNum*16 + fineY)
				patternHi := p.read(table + patternNum*16 + fineY + 8)

				for pixel := 0; pixel < 8; pixel++ {
					pixello := patternLo & 0x80 >> 7
					pixelhi := patternHi & 0x80 >> 6
					patternLo <<= 1
					patternHi <<= 1
					paletteIndex := p.paletteData[attr|pixello|pixelhi]
					// buf.SetRGBA(xoffset+fineX+pixel, y, palette[paletteIndex])
					pos := y*128*2*4 + (xoffset+fineX+pixel)*4
					c := palette[paletteIndex]
					buf[pos+0] = c.R
					buf[pos+1] = c.G
					buf[pos+2] = c.B
					buf[pos+3] = c.A
				}
			}
		}
	}

	draw(0x0000, 0)
	draw(0x1000, 128)
}

func (p *ppu) drawNametables(buf []byte) {
	if p.cartridge == nil {
		return
	}

	draw := func(nametable, offsetX, offsetY uint16) {
		patternTable := p.backgroundTable()

		for y := uint16(0); y < 240; y++ {
			tileY := uint16(y / 8)

			patternY := uint16(y % 8)
			for tile := uint16(0); tile < 32; tile++ {
				nametableAddr := tileY*32 + tile
				tileX := tile * 8

				patternNum := uint16(p.read(nametable + nametableAddr))

				patternLo := p.read(patternTable + patternNum*16 + patternY)
				patternHi := p.read(patternTable + patternNum*16 + patternY + 8)

				attribute := p.read(nametable + 960 + (tileY/4)*8 + tile/4)

				top := tileY%4/2 == 0
				bot := tileY%4/2 == 1
				left := tile%4/2 == 0
				right := tile%4/2 == 1

				if top && left {
					attribute = attribute >> 0 & 0x03 << 2
				} else if top && right {
					attribute = attribute >> 2 & 0x03 << 2
				} else if bot && left {
					attribute = attribute >> 4 & 0x03 << 2
				} else if bot && right {
					attribute = attribute >> 6 & 0x03 << 2
				}

				for pixel := uint16(0); pixel < 8; pixel++ {
					pixello := patternLo & 0x80 >> 7
					pixelhi := patternHi & 0x80 >> 6
					patternLo <<= 1
					patternHi <<= 1
					color := p.paletteData[attribute|pixello|pixelhi]

					pos := int(offsetY+y)*256*2*4 + int(offsetX+tileX+pixel)*4
					c := palette[color]
					buf[pos+0] = c.R
					buf[pos+1] = c.G
					buf[pos+2] = c.B
					buf[pos+3] = c.A
				}
			}
		}
	}

	draw(0x2000, 0, 0)
	draw(0x2400, 256, 0)
	draw(0x2800, 0, 240)
	draw(0x2C00, 256, 240)
}

func (p *ppu) debugDumpSprites() {
	y := p.scanline

	if p.mask&showSprites == 0 {
		return
	}

	for i := 0; i < 256; i += 4 {
		spriteY := p.oamData[i] + 1
		patternNum := uint16(p.oamData[i+1])
		spriteAttr := p.oamData[i+2]
		x := p.oamData[i+3]

		flipV := spriteAttr&0x80 > 0
		flipH := spriteAttr&0x40 > 0
		priority := spriteAttr & 0x20
		colorUp := spriteAttr & 0x03 << 2
		_ = flipV
		_ = flipH

		row := uint16(byte(y) - spriteY)
		// sprite not in range
		if row < 0 || row > 7 {
			continue
		}

		// sprite line is not = to y
		if int(spriteY)+int(row) != y {
			continue
		}
		outputX := byte(p.dot - 1)
		if outputX < x || outputX > x+7 {
			continue
		}

		_ = priority
		if spriteY == 0 {
			continue
		}

		// TODO: bigger sprites
		var patternTable uint16
		if p.ctrl&spritePatternTableAddress > 0 {
			patternTable = 0x1000
		} else {
			patternTable = 0x0000
		}

		patternLo := p.read(patternTable + patternNum*16 + row)
		patternHi := p.read(patternTable + patternNum*16 + row + 8)

		for col := 0; col < 8; col++ {
			var pixello, pixelhi byte
			if !flipH {
				pixello = patternLo & 0x80 >> 7
				pixelhi = (patternHi & 0x80) >> 6
				patternLo <<= 1
				patternHi <<= 1
			} else {
				pixello = patternLo & 0x1
				pixelhi = (patternHi & 0x1) << 1
				patternLo >>= 1
				patternHi >>= 1
			}

			paletteIdx := p.paletteData[16+(colorUp|pixello|pixelhi)]

			if paletteIdx != 0 && priority == 0 {
				// p.buffer.SetRGBA(int(x)+col, int(spriteY)+int(row), palette[paletteIdx])
			}
		}
	}
}

// func (p *ppu) tick(cpu *CPU) {
// 	// TODO: OAMADDR is set to 0 during each of ticks 257-320 (the sprite tile loading interval) of the pre-render and visible scanlines.
// 	// TODO: The value of OAMADDR when sprite evaluation starts at tick 65 of the visible scanlines will determine where in OAM sprite evaluation starts, and hence which sprite gets treated as sprite 0
// 	var (
// 		width  = 256
// 		height = 240
// 	)
// 	if p.dot == 1 && p.scanline < 240 {

// 		p.renderBackground()
// 		p.renderSprites()
// 		// for j := 0; j < 256; j++ {
// 		// 	p.buffer.SetRGBA(j, p.scanline+1, color.RGBA{255, 255, 255, 255})
// 		// }
// 	}
// 	if p.dot >= width {
// 		// hblank = true
// 	}
// 	if p.dot == 340 {
// 		p.dot = 0
// 		p.scanline++
// 	} else {
// 		p.dot++
// 	}

// 	if p.scanline == height {
// 		// post render
// 	}
// 	if p.scanline > height && p.scanline < 261 {
// 		p.status |= VerticalBlank
// 		//TODO: read up on timing info instead of relying on bool
// 		if !p.nmiGenerated && p.ctrl&GenerateNMI > 0 {
// 			p.nmiGenerated = true
// 			cpu.trigger(NMI)
// 		}
// 	} else {
// 		p.nmiGenerated = false
// 		p.status &^= VerticalBlank
// 	}

// 	if p.scanline == 261 {
// 		// pre render
// 		if p.dot == 1 {
// 			p.status &^= SpriteOverflow
// 			p.status &^= Sprite0Hit
// 			p.status &^= VerticalBlank
// 		}
// 		p.scanline = 0
// 		p.frame++
// 	}
// }

// func (p *ppu) renderBackground() {
// 	nametable := 0x2000 * int(p.ctrl&NametableAddress)

// 	var patternTable uint16
// 	if p.ctrl&BackgroundPatternTableAddress > 0 {
// 		patternTable = 0x1000
// 	} else {
// 		patternTable = 0x0000
// 	}

// 	y := p.scanline // [0,240[

// 	tilesPerScanline := 32
// 	pixelsPerRow := 8
// 	tileY := y / 8
// 	patternY := uint16(y % 8)
// 	for tile := 0; tile < tilesPerScanline; tile++ {
// 		nametableAddr := tileY*tilesPerScanline + tile
// 		tileX := tile * 8

// 		patternNum := uint16(p.readNt(uint16(nametable + nametableAddr)))

// 		patternLo := p.Read(patternTable + patternNum*16 + patternY)
// 		patternHi := p.Read(patternTable + patternNum*16 + patternY + 8)

// 		attribute := p.readNt(uint16(nametable + 960 + (tileY/4)*8 + tile/4))
// 		top := tileY%4/2 == 0
// 		bot := tileY%4/2 == 1
// 		left := tile%4/2 == 0
// 		right := tile%4/2 == 1

// 		// 7654 3210
// 		// |||| ||++- Color bits 3-2 for top left quadrant of this byte
// 		// |||| ++--- Color bits 3-2 for top right quadrant of this byte
// 		// ||++------ Color bits 3-2 for bottom left quadrant of this byte
// 		// ++-------- Color bits 3-2 for bottom right quadrant of this byte
// 		if top && left {
// 			attribute >>= 0
// 		} else if top && right {
// 			attribute >>= 2
// 		} else if bot && left {
// 			attribute >>= 4
// 		} else if bot && right {
// 			attribute >>= 6
// 		}
// 		attribute = (attribute & 3) << 2
// 		for pixel := 0; pixel < pixelsPerRow; pixel++ {
// 			pixello := patternLo & 0x80 >> 7
// 			pixelhi := patternHi & 0x80 >> 6
// 			patternLo <<= 1
// 			patternHi <<= 1
// 			color := p.paletteData[attribute|pixello|pixelhi]
// 			p.buffer.SetRGBA(tileX+pixel, y, palette[color])
// 		}
// 	}
// }
