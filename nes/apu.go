package nes

import (
	"fmt"
	"io"
	"math"

	"github.com/go-audio/wav"
)

var lengthTable = []byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

var pulseDutyTables = [][]byte{
	{0, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0},
	{1, 0, 0, 1, 1, 1, 1, 1},
}

var triangleTable = []byte{
	15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

var noiseFreqTable = []uint16{
	4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068,
}

var pulseTable [31]float32
var tndTable [203]float32

func init() {
	for i := 0; i < 31; i++ {
		pulseTable[i] = 95.52 / (8128.0/float32(i) + 100)
	}
	for i := 0; i < 203; i++ {
		tndTable[i] = 163.67 / (24329.0/float32(i) + 100)
	}
}

type pulse struct {
	enabled bool
	channel byte

	dutyTable       byte
	envelopeLoop    bool
	lengthEnabled   bool
	envelopeEnabled bool
	envelopeV       byte

	sweepTimer   byte
	sweepNegate  bool
	sweepShift   byte
	sweepReload  bool
	sweepEnabled bool

	sweepCounter byte

	freqTimer     uint16
	lengthCounter byte
	freqCounter   uint16
	dutyCounter   byte
	envelopeReset bool

	envelopeHiddenVol byte
	envelopeCounter   byte
}

func (p *pulse) writePort(addr uint16, v byte) {
	switch addr {
	case 0x4000: //DDLC VVVV
		p.dutyTable = v >> 6
		p.envelopeLoop = v>>5&1 == 1
		p.lengthEnabled = v>>5&1 == 0
		p.envelopeEnabled = v>>4&1 == 0
		p.envelopeV = v & 0xF

	case 0x4001: //EPPP NSSS
		p.sweepTimer = v >> 4 & 7
		p.sweepNegate = v>>3&1 == 1
		p.sweepShift = v & 7
		p.sweepReload = true
		p.sweepEnabled = v>>7&1 == 1 && p.sweepShift != 0

	case 0x4002: //TTTT TTTT
		p.freqTimer = p.freqTimer&0xFF00 | uint16(v)

	case 0x4003: //LLLL LTTT
		p.freqTimer = uint16(v&7)<<8 | p.freqTimer&0x00FF

		if p.enabled {
			p.lengthCounter = lengthTable[v>>3]
		}
		// phase is also reset here  (important for games like SMB)
		p.freqCounter = p.freqTimer
		p.dutyCounter = 0

		// envelope is also flagged for reset here
		p.envelopeReset = true

	case 0x4015: //---D NT21
		p.enabled = v>>p.channel&1 == 1

		if !p.enabled {
			p.lengthCounter = 0
		}
	}
}

func (p *pulse) clockFreq() {
	if p.freqCounter > 0 {
		p.freqCounter--
	} else {
		p.freqCounter = p.freqTimer
		p.dutyCounter = (p.dutyCounter + 1) & 7
	}
}

func (p *pulse) clockEnvelope() {
	if p.envelopeReset {
		p.envelopeReset = false
		p.envelopeHiddenVol = 0xF
		p.envelopeCounter = p.envelopeV
		return
	}
	if p.envelopeCounter > 0 {
		p.envelopeCounter--
		return
	}

	p.envelopeCounter = p.envelopeV
	if p.envelopeHiddenVol > 0 {
		p.envelopeHiddenVol--
	} else if p.envelopeLoop {
		p.envelopeHiddenVol = 0xF
	}
}

func (p *pulse) clockSweep() {
	if p.sweepReload {
		p.sweepCounter = p.sweepTimer
		// note there's an edge case here -- see http://wiki.nesdev.com/w/index.php/APU_Sweep
		// for details.  You can probably ignore it for now

		p.sweepReload = false
		return
	}

	if p.sweepCounter > 0 {
		p.sweepCounter--
		return
	}

	p.sweepCounter = p.sweepTimer
	if p.sweepEnabled && !p.isSweepForcingSilence() {
		shift := p.freqTimer >> p.sweepShift
		var offset uint16
		if p.channel == 0 {
			offset = 1
		}
		if p.sweepNegate {
			p.freqTimer -= shift + offset
		} else {
			p.freqTimer += shift
		}
	}

	// sweep := func() {
	// 	delta := p.freqTimer >> p.sweepShift
	// 	if p.sweepNegate {
	// 		p.freqTimer -= delta
	// 		if p.channel == 0 {
	// 			p.freqTimer--
	// 		}
	// 	} else {
	// 		p.freqTimer += delta
	// 	}
	// }

	// if p.sweepReload {
	// 	if p.sweepEnabled && p.sweepCounter == 0 {
	// 		sweep()
	// 	}
	// 	p.sweepCounter = p.sweepTimer
	// 	p.sweepReload = false
	// } else if p.sweepCounter > 0 {
	// 	p.sweepCounter--
	// } else {
	// 	if p.sweepEnabled {
	// 		sweep()
	// 	}
	// 	p.sweepCounter = p.sweepTimer
	// }
}

func (p *pulse) clockLengthCounter() {
	if p.lengthEnabled && p.lengthCounter > 0 {
		p.lengthCounter--
	}
}

func (p *pulse) isSweepForcingSilence() bool {
	if p.freqTimer < 8 {
		return true
	}
	if !p.sweepNegate && p.freqTimer+(p.freqTimer>>p.sweepShift) >= 0x800 {
		return true
	}

	return false
}

func (p *pulse) sample() byte {
	dutyHigh := pulseDutyTables[p.dutyTable][p.dutyCounter] != 0
	active := p.lengthCounter != 0
	if p.enabled && dutyHigh && active && !p.isSweepForcingSilence() {
		// output current volume
		if p.envelopeEnabled {
			return p.envelopeHiddenVol
		}
		return p.envelopeV
	}

	// low duty, or channel is silent
	return 0
}

type triangle struct {
	enabled bool

	linearControl bool
	lengthEnabled bool
	linearLoad    byte
	freqTimer     uint16
	lengthCounter byte
	linearReload  bool

	freqCounter   uint16
	linearCounter byte

	triStep byte
}

func (t *triangle) writePort(addr uint16, v byte) {
	switch addr {
	case 0x4008: //CRRR RRRR
		t.linearControl = v>>7&1 == 1
		t.lengthEnabled = v>>7&1 == 0
		t.linearLoad = v &^ 0x80

	case 0x4009: //---- ----
		// unused
	case 0x400A: //TTTT TTTT
		t.freqTimer = t.freqTimer&0xFF00 | uint16(v)

	case 0x400B: //LLLL LTTT
		t.freqTimer = uint16(v&7)<<8 | t.freqTimer&0x00FF

		if t.enabled {
			t.lengthCounter = lengthTable[v>>3]
		}
		// t.freqCounter = t.freqTimer //TODO?

		t.linearReload = true
	case 0x4015: //---D NT21
		t.enabled = v>>2&1 == 1
		if !t.enabled {
			t.lengthCounter = 0
		}
	}
}

func (t *triangle) ultrasonic() bool {
	return t.freqTimer < 2 && t.freqCounter == 0
}

func (t *triangle) clockFreq() {
	if t.lengthCounter == 0 || t.linearCounter == 0 || t.ultrasonic() {
		return
	}

	// if t.freqCounter > 0 {
	// 	t.freqCounter--
	// } else {
	// 	t.freqCounter = t.freqTimer
	// 	t.triStep = (t.triStep + 1) & 0x1F // tri-step bound to 00..1F range
	// }
	if t.freqCounter > 0 {
		t.freqCounter--
		return
	}

	t.freqCounter = t.freqTimer
	if t.lengthCounter > 0 && t.linearCounter > 0 {
		t.triStep = (t.triStep + 1) % 32
	}
}

func (t *triangle) clockLinear() {
	if t.linearReload {
		t.linearCounter = t.linearLoad
	} else if t.linearCounter > 0 {
		t.linearCounter--
	}

	if !t.linearControl {
		t.linearReload = false
	}
}

func (t *triangle) clockLengthCounter() {
	if t.lengthEnabled && t.lengthCounter > 0 {
		t.lengthCounter--
	}
}

func (t *triangle) sample() byte {
	// if t.lengthCounter == 0 {
	// 	return triangleTable[0]
	// }

	return triangleTable[t.triStep]
}

type noise struct {
	enabled bool

	envelopeLoop    bool
	lengthEnabled   bool
	envelopeEnabled bool
	envelopeV       byte

	freqTimer     uint16
	lengthCounter byte
	freqCounter   uint16
	dutyCounter   byte
	envelopeReset bool
	shiftMode     byte

	register          uint16
	envelopeHiddenVol byte
	envelopeCounter   byte
}

func (n *noise) writePort(addr uint16, v byte) {
	switch addr {
	case 0x400C: //--LC VVVV
		n.envelopeLoop = v>>5&1 == 1
		n.lengthEnabled = v>>5&1 == 0
		n.envelopeEnabled = v>>4&1 == 0
		n.envelopeV = v & 0xF

	case 0x400D: //---- ----
		// unused
	case 0x400E: //L--- PPPP
		n.freqTimer = noiseFreqTable[v&0x0F] // see http://wiki.nesdev.com/w/index.php/APU_Noise for freq table
		n.shiftMode = v >> 7

	case 0x400F: //LLLL L---
		if n.enabled {
			n.lengthCounter = lengthTable[v>>3]
		}

		// envelope is also flagged for reset here
		n.envelopeReset = true

	case 0x4015: //---D NT21
		n.enabled = v>>3&1 == 1
		if !n.enabled {
			n.lengthCounter = 0
		}
	}
}

func (n *noise) clockFreq() {
	if n.freqCounter > 0 {
		n.freqCounter--
	} else {
		n.freqCounter = n.freqTimer

		if n.shiftMode == 1 {
			n.register |= (n.register>>6 ^ n.register&1) << 15
		} else {
			n.register |= (n.register>>1 ^ n.register&1) << 15
		}
		n.register >>= 1
	}
}

func (n *noise) clockEnvelope() {
	if n.envelopeReset {
		n.envelopeReset = false
		n.envelopeHiddenVol = 0xF
		n.envelopeCounter = n.envelopeV
		return
	}
	if n.envelopeCounter > 0 {
		n.envelopeCounter--
		return
	}

	n.envelopeCounter = n.envelopeV
	if n.envelopeHiddenVol > 0 {
		n.envelopeHiddenVol--
	} else if n.envelopeLoop {
		n.envelopeHiddenVol = 0xF
	}
}

func (n *noise) clockLengthCounter() {
	if n.lengthEnabled && n.lengthCounter > 0 {
		n.lengthCounter--
	}
}

func (n *noise) sample() byte {
	outputIsLow := n.register&1 == 0
	active := n.lengthCounter != 0
	if outputIsLow && active {
		// output current volume
		if n.envelopeEnabled {
			return n.envelopeHiddenVol
		}
		return n.envelopeV
	}

	// high shift output, or channel is silent
	return 0
}

type apu struct {
	seqResetDelay int8
	pulse0        *pulse
	pulse1        *pulse
	triangle      *triangle
	noise         *noise

	sequencerMode    byte
	irqEnabled       bool
	sequencerCounter uint16
	irqPending       bool

	last4017Write byte

	mixer *mixer
}

func newApu(bufferSize int, freq float32, makeFile func(channel string) (io.WriteSeeker, error)) *apu {
	return &apu{
		pulse0: &pulse{
			channel:       0,
			lengthEnabled: true,
		},
		pulse1: &pulse{
			channel:       1,
			lengthEnabled: true,
		},
		triangle: &triangle{
			lengthEnabled: true,
		},
		noise: &noise{
			register:      1,
			lengthEnabled: true,
		},
		mixer: newMixer(bufferSize, freq, makeFile),
	}
}

func (a *apu) channel() <-chan float32 {
	return a.mixer.Output
}

func (a *apu) readPort(addr uint16) byte {
	switch addr {
	case 0x4015: // IF-D NT21
		ret := byte(0)

		if a.pulse0.lengthCounter != 0 {
			ret |= 0x01
		}
		if a.pulse1.lengthCounter != 0 {
			ret |= 0x02
		}
		if a.triangle.lengthCounter != 0 {
			ret |= 0x04
		}
		if a.noise.lengthCounter != 0 {
			ret |= 0x08
		}

		if a.irqPending {
			ret |= 0x40
		}

		// ... DMC IRQ state read back here

		a.irqPending = false // IRQ acknowledged on $4015 read

		return ret
	}

	return 0
}

func (a *apu) writePort(addr uint16, v byte) {
	switch addr {
	case 0x4000, 0x4001, 0x4002, 0x4003:
		a.pulse0.writePort(addr, v)

	case 0x4004, 0x4005, 0x4006, 0x4007:
		a.pulse1.writePort(addr-0x0004, v)

	case 0x4008, 0x4009, 0x400A, 0x400B:
		a.triangle.writePort(addr, v)

	case 0x400C, 0x400D, 0x400E, 0x400F:
		a.noise.writePort(addr, v)

	case 0x4015:
		a.pulse0.writePort(addr, v)
		a.pulse1.writePort(addr, v)
		a.triangle.writePort(addr, v)
		a.noise.writePort(addr, v)

	case 0x4017: //MI-- ----
		a.sequencerMode = v >> 7 // switch between 5-step (1) and 4-step (0) mode
		a.irqEnabled = v>>6 == 0
		if a.sequencerMode == 0 {
			a.seqResetDelay = 4
		} else {
			a.seqResetDelay = 0
		}
		// a.sequencerCounter = 0 // see: http://wiki.nesdev.com/w/index.php/APU_Frame_Counterq
		// for example, this will be 3728.5 apu cycles, or 7457 CPU cycles.
		// It might be easier to work in CPU cycles so you don't have to deal with
		// half cycles.

		if a.sequencerMode == 1 {
			a.clockQuarterFrame()
			a.clockHalfFrame()
		}
		if !a.irqEnabled {
			a.irqPending = false // acknowledge Frame IRQ
		}
		a.last4017Write = v
	}
}

func (a *apu) clockFC(c *cpu) {
	switch a.sequencerMode {
	case 0:
		switch a.sequencerCounter {
		case 0:
			if a.irqEnabled {
				c.trigger(irq)
				a.irqPending = true
			}
		case 7457:
			a.clockQuarterFrame()
		case 14913:
			a.clockQuarterFrame()
			a.clockHalfFrame()
		case 22371:
			a.clockQuarterFrame()
		case 29828:
			if a.irqEnabled {
				c.trigger(irq)
				a.irqPending = true
			}
		case 29829:
			a.clockQuarterFrame()
			a.clockHalfFrame()
			if a.irqEnabled {
				c.trigger(irq)
				a.irqPending = true
			}
		}

		a.sequencerCounter++
		if a.sequencerCounter == 29830 {
			a.sequencerCounter = 0
		}

	case 1:
		switch a.sequencerCounter {
		case 0:
		case 7457:
			a.clockQuarterFrame()
		case 14913:
			a.clockQuarterFrame()
			a.clockHalfFrame()
		case 22371:
			a.clockQuarterFrame()
		case 29829:
		case 37281:
			a.clockQuarterFrame()
			a.clockHalfFrame()
		}
		a.sequencerCounter++
		if a.sequencerCounter == 37282 {
			a.sequencerCounter = 0
		}
	}

}

func (a *apu) clockQuarterFrame() {
	a.pulse0.clockEnvelope()
	a.pulse1.clockEnvelope()
	a.triangle.clockLinear()
	a.noise.clockEnvelope()
}

func (a *apu) clockHalfFrame() {
	a.pulse0.clockSweep()
	a.pulse0.clockLengthCounter()

	a.pulse1.clockSweep()
	a.pulse1.clockLengthCounter()

	a.triangle.clockLengthCounter()

	a.noise.clockLengthCounter()
}

func (a *apu) clock(c *cpu) {
	if a.seqResetDelay > 0 {
		a.seqResetDelay--
	} else if a.seqResetDelay == 0 {
		a.sequencerCounter = 0
		a.seqResetDelay = -1
	}
	if c.cycles&1 == 1 {
		a.pulse0.clockFreq()
		a.pulse1.clockFreq()
		a.noise.clockFreq()
	}
	a.triangle.clockFreq()

	a.clockFC(c)

	a.mixer.mix(
		a.pulse0.sample(),
		a.pulse1.sample(),
		a.triangle.sample(),
		a.noise.sample(),
		0, //TODO: a.dmc.sample()
	)

}

func (a *apu) reset() {
	a.writePort(0x4015, 0)
	a.writePort(0x4017, a.last4017Write)
}

type mixer struct {
	Output chan float32

	p0 *channel
	p1 *channel
	t  *channel
	n  *channel
	d  *channel
	m  *channel

	filters []filter
	cycles  uint64
	divider uint64
}

func newMixer(bufferSize int, freq float32, makeFile func(channel string) (io.WriteSeeker, error)) *mixer {
	return &mixer{
		Output:  make(chan float32, bufferSize),
		divider: uint64(cpuFreq / float64(freq)),
		filters: []filter{
			highpass(freq, 90),
			highpass(freq, 440),
			lowpass(freq, 14000),
		},
		p0: newChannel("pulse_0", freq, makeFile),
		p1: newChannel("pulse_1", freq, makeFile),
		t:  newChannel("triangle", freq, makeFile),
		n:  newChannel("noise", freq, makeFile),
		d:  newChannel("dmc", freq, makeFile),
		m:  newChannel("mix", freq, makeFile),
	}
}

func (m *mixer) startRecording() error {
	fmt.Println("startRecording")
	if err := m.p0.startRecording(); err != nil {
		return err
	}
	if err := m.p1.startRecording(); err != nil {
		return err
	}
	if err := m.t.startRecording(); err != nil {
		return err
	}
	if err := m.n.startRecording(); err != nil {
		return err
	}
	if err := m.d.startRecording(); err != nil {
		return err
	}
	if err := m.m.startRecording(); err != nil {
		return err
	}

	return nil
}

func (m *mixer) pauseRecording() {
	fmt.Println("pauseRecording")
	m.p0.pauseRecording()
	m.p1.pauseRecording()
	m.t.pauseRecording()
	m.n.pauseRecording()
	m.d.pauseRecording()
	m.m.pauseRecording()
}

func (m *mixer) unpauseRecording() {
	fmt.Println("unpauseRecording")
	m.p0.unpauseRecording()
	m.p1.unpauseRecording()
	m.t.unpauseRecording()
	m.n.unpauseRecording()
	m.d.unpauseRecording()
	m.m.unpauseRecording()
}

func (m *mixer) stopRecording() error {
	fmt.Println("stopRecording")
	if err := m.p0.stopRecording(); err != nil {
		return err
	}
	if err := m.p1.stopRecording(); err != nil {
		return err
	}
	if err := m.t.stopRecording(); err != nil {
		return err
	}
	if err := m.n.stopRecording(); err != nil {
		return err
	}
	if err := m.d.stopRecording(); err != nil {
		return err
	}
	if err := m.m.stopRecording(); err != nil {
		return err
	}

	return nil
}

func (m *mixer) mix(p0, p1, t, n, d byte) {

	if m.cycles%m.divider == 0 { //TODO: 0 or 1?
		m.p0.process(pulseTable[p0+0] + tndTable[0])
		m.p1.process(pulseTable[0+p1] + tndTable[0])
		m.t.process(pulseTable[0] + tndTable[3*t])
		m.n.process(pulseTable[0] + tndTable[2*n])
		m.d.process(pulseTable[0] + tndTable[d])
		out := pulseTable[p0+p1] + tndTable[3*t+2*n+d]
		for _, f := range m.filters {
			out = f(out)
		}
		m.m.process(out)
		m.Output <- out
	}

	m.cycles++
}

type channel struct {
	name      string
	recording bool
	paused    bool
	freq      float32
	makeFile  func(channel string) (io.WriteSeeker, error)
	enc       *wav.Encoder
}

func newChannel(name string, freq float32, makeFile func(channel string) (io.WriteSeeker, error)) *channel {
	return &channel{
		name:     name,
		freq:     freq,
		makeFile: makeFile,
	}
}

func (c *channel) createEncoder() error {
	fmt.Println(c.name, "createEncoder")
	f, err := c.makeFile(c.name)
	if err != nil {
		return err
	}

	c.enc = wav.NewEncoder(f, int(c.freq), 32, 1, 0x0003)

	return nil
}

func (c *channel) process(preMix float32) error {
	if !c.recording || c.paused {
		return nil
	}

	if err := c.enc.WriteFrame(preMix); err != nil {
		return err
	}

	return nil
}

func (c *channel) startRecording() error {
	var err error
	if c.recording == false {
		err = c.createEncoder()
	}
	c.recording = true
	c.paused = false
	return err
}

func (c *channel) pauseRecording() {
	if c.paused {
		c.unpauseRecording()
		return
	}
	c.paused = true
}

func (c *channel) unpauseRecording() {
	c.paused = false
}

func (c *channel) stopRecording() error {
	if !c.recording {
		return nil
	}

	c.recording = false
	c.paused = false

	if err := c.enc.Close(); err != nil {
		return err
	}

	return nil
}

type filter func(float32) float32

func lowpass(sampleRate, cutoff float32) filter {
	rc := 1.0 / (2.0 * math.Pi * cutoff)
	dt := 1.0 / sampleRate
	alpha := dt / (rc + dt)

	var prev float32
	return func(x float32) float32 {
		ret := alpha*x + (1.0-alpha)*prev
		prev = ret
		return ret
	}
}

func highpass(sampleRate, cutoff float32) filter {
	rc := 1.0 / (2.0 * math.Pi * cutoff)
	dt := 1.0 / sampleRate
	alpha := rc / (rc + dt)

	var prev, prevx float32
	return func(x float32) float32 {
		ret := alpha*prev + alpha*(x-prevx)
		prev = ret
		prevx = x
		return ret
	}
}
