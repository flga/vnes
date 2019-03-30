package nes

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	ppuRegistersSize = 8
	ioRegistersSize  = 32
	expRomSize       = 8160
	sramSize         = 8192
	prgBankSize      = 16384
	prgRomSize       = 16384 * 2 //TODO
)

type Console struct {
	Cartridge   *Cartridge
	RAM         *RAM
	CPU         *CPU
	APU         *APU
	PPU         *PPU
	Controller1 *Controller

	bus *SysBus

	openFiles []*os.File
}

func NewConsole(pc uint16, debugOut io.Writer) *Console {
	console := &Console{}
	makeFile := func(channel string) (io.WriteSeeker, error) {
		name := "TODO"
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		f, err := ioutil.TempFile(dir, strings.TrimSuffix(path.Base(name), path.Ext(name))+"_"+channel+"_*.wav")
		if err != nil {
			return nil, err
		}

		console.openFiles = append(console.openFiles, f)
		return f, nil
	}

	ram := NewRAM()
	ctrl1 := &Controller{}

	ppu := NewPPU()
	apu := NewAPU(4096, 48000, makeFile) // TODO: parameterize sample rate
	cpu := NewCPU(debugOut, ppu, apu)

	bus := &SysBus{
		RAM:   ram,
		CPU:   cpu,
		APU:   apu,
		PPU:   ppu,
		Ctrl1: ctrl1,
	}

	if pc != 0 {
		cpu.SetPC(pc)
	}
	cpu.Cycles = 7 //TODO

	console.RAM = ram
	console.CPU = cpu
	console.APU = apu
	console.PPU = ppu
	console.Controller1 = ctrl1
	console.bus = bus

	return console
}

func (c *Console) Empty() bool {
	return c.Cartridge == nil
}

func (c *Console) Load(cartridge *Cartridge) {
	first := c.Cartridge == nil
	c.Cartridge = cartridge
	c.bus.Cartridge = cartridge
	c.PPU.Cartridge = cartridge

	if first {
		c.CPU.Init(c.bus)
		return
	}

	c.Reset()
}

func (c *Console) StartRecording() error {
	return c.APU.mixer.startRecording()
}

func (c *Console) PauseRecording() {
	c.APU.mixer.pauseRecording()
}

func (c *Console) UnpauseRecording() {
	c.APU.mixer.unpauseRecording()
}

func (c *Console) StopRecording() error {
	return c.APU.mixer.stopRecording()
}

func (c *Console) Close() error {
	if err := c.StopRecording(); err != nil {
		return err
	}

	var err error
	for _, f := range c.openFiles {
		err = f.Close()
	}

	return err
}

func (c *Console) Reset() {
	c.CPU.Reset(c.bus)
	c.APU.Reset()
}

func (c *Console) StepFrame() {
	if c.Empty() {
		return
	}

	frame := c.PPU.Frame
	for frame == c.PPU.Frame {
		c.CPU.Execute(c.bus, c.PPU)
	}
	// fmt.Println(c.PPU.bufferHead)
}

func (c *Console) Buffer() []byte {
	return c.PPU.buffer
}

func (c *Console) Read(addr uint16) byte {
	return c.bus.Read(addr)
}

func (c *Console) Write(addr uint16, v byte) {
	c.bus.Write(addr, v)
}
