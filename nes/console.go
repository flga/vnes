package nes

import (
	"fmt"
	"image"
	"io"
	"time"
)

const (
	ramSize          = 2048
	ppuRegistersSize = 8
	ioRegistersSize  = 32
	expRomSize       = 8160
	sramSize         = 8192
	prgBankSize      = 16384
	prgRomSize       = 16384 * 2 //TODO
)

type Console struct {
	Cartridge   *Cartridge
	CPU         *CPU
	PPU         *PPU
	Controller1 *Controller

	bus       *SysBus
	frameTime time.Duration
}

func NewConsole(c *Cartridge, pc uint16, debugOut io.Writer) *Console {
	if c.Mapper != 0 {
		panic(fmt.Sprintf("unsupported mapper %d", c.Mapper))
	}

	ram := make([]byte, ramSize)
	ctrl1 := &Controller{}

	ppu := &PPU{
		Cartridge: c,
	}

	cpu := NewCPU(debugOut)

	bus := &SysBus{
		Cartridge: c,
		RAM:       ram,
		CPU:       cpu,
		PPU:       ppu,
		Ctrl1:     ctrl1,
	}

	ppu.Init()
	ppu.Bus = bus // TODO: untangle

	cpu.Init(bus)
	if pc != 0 {
		cpu.SetPC(pc)
	}
	cpu.Cycles = 7 //TODO

	return &Console{
		Cartridge:   c,
		CPU:         cpu,
		PPU:         ppu,
		Controller1: ctrl1,
		bus:         bus,
	}
}

func (c *Console) FrameTime() time.Duration {
	return c.frameTime
}

func (c *Console) Reset() {
	c.CPU.Reset(c.bus)
}

func (c *Console) Step() {
	cycles := c.CPU.Execute(c.bus, c.PPU)
	for i := uint64(0); i < cycles; i++ {
		c.PPU.Tick(c.CPU)
		c.PPU.Tick(c.CPU)
		c.PPU.Tick(c.CPU)
	}
}

func (c *Console) StepFrame() {
	start := time.Now()
	frame := c.PPU.Frame
	for frame == c.PPU.Frame {
		c.Step()
	}
	c.frameTime = time.Since(start)

}
func (c *Console) StepScanline() {
	start := time.Now()
	scan := c.PPU.ScanLine
	for scan == c.PPU.ScanLine {
		c.Step()
	}
	c.frameTime = time.Since(start)
}

func (c *Console) Buffer() *image.RGBA {
	return c.PPU.buffer
}

func (c *Console) Read(addr uint16) byte {
	return c.bus.Read(addr)
}

func (c *Console) Write(addr uint16, v byte) {
	c.bus.Write(addr, v)
}
