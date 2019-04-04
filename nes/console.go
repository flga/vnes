package nes

import (
	"fmt"
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
	cartridge   *cartridge
	ram         *ram
	cpu         *cpu
	apu         *apu
	ppu         *ppu
	controller1 *controller
	controller2 *controller

	bus *sysBus

	openFiles []*os.File
}

func NewConsole(sampleRate float32, pc uint16, debugOut io.Writer) *Console {
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

	ram := newRam()
	ctrl1 := &controller{}
	ctrl2 := &controller{}

	ppu := newPpu()
	apu := newApu(4096, sampleRate, makeFile)
	cpu := newCpu(debugOut, ppu, apu)

	bus := &sysBus{
		ram:   ram,
		cpu:   cpu,
		apu:   apu,
		ppu:   ppu,
		ctrl1: ctrl1,
		ctrl2: ctrl2,
	}

	if pc != 0 {
		cpu.setPC(pc)
	}
	cpu.cycles = 7 //TODO

	console.ram = ram
	console.cpu = cpu
	console.apu = apu
	console.ppu = ppu
	console.controller1 = ctrl1
	console.controller2 = ctrl2
	console.bus = bus

	return console
}

func (c *Console) Empty() bool {
	return c.cartridge == nil
}

func (c *Console) load(cartridge *cartridge) {
	first := c.cartridge == nil
	c.cartridge = cartridge
	c.bus.cartridge = cartridge
	c.ppu.cartridge = cartridge

	if first {
		c.cpu.init(c.bus)
		return
	}

	c.Reset()
}

func (c *Console) LoadPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open rom: %s", err)
	}
	defer f.Close()

	cart, err := loadRom(f)
	if err != nil {
		return err
	}

	c.load(cart)
	return nil
}

func (c *Console) LoadRom(rom io.Reader) error {
	cart, err := loadRom(rom)
	if err != nil {
		return err
	}

	c.load(cart)
	return nil
}

func (c *Console) StartRecording() error {
	return c.apu.mixer.startRecording()
}

func (c *Console) PauseRecording() {
	c.apu.mixer.pauseRecording()
}

func (c *Console) UnpauseRecording() {
	c.apu.mixer.unpauseRecording()
}

func (c *Console) StopRecording() error {
	return c.apu.mixer.stopRecording()
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
	c.cpu.reset(c.bus)
	c.apu.reset()
}

func (c *Console) StepFrame() {
	if c.Empty() {
		return
	}

	frame := c.ppu.frame
	for frame == c.ppu.frame {
		c.cpu.execute(c.bus)
	}
}

func (c *Console) Press(ctrl int, button Button) {
	switch ctrl {
	case 0:
		c.controller1.press(button)
	case 1:
		c.controller2.press(button)
	}
}

func (c *Console) Release(ctrl int, button Button) {
	switch ctrl {
	case 0:
		c.controller1.release(button)
	case 1:
		c.controller2.release(button)
	}
}

func (c *Console) Buffer() []byte {
	return c.ppu.buffer
}

func (c *Console) AudioChannel() <-chan float32 {
	return c.apu.channel()
}

func (c *Console) DrawNametables(buf []byte) {
	c.ppu.drawNametables(buf)
}

func (c *Console) DrawPatternTables(buf []byte, palette byte) {
	c.ppu.drawPatternTables(buf, palette)
}

func (c *Console) Read(addr uint16) byte {
	return c.bus.read(addr)
}

func (c *Console) Write(addr uint16, v byte) {
	c.bus.write(addr, v)
}
