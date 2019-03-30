package main

import (
	"fmt"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
)

type audioEngine struct {
	AudioChan <-chan float32

	freeFuncs    []func() error
	envelope     *envelope
	streamParams portaudio.StreamParameters
	stream       *portaudio.Stream
	lastSample   float32
}

func (a *audioEngine) deferFn(fn func() error) {
	a.freeFuncs = append(a.freeFuncs, fn)
}

func (a *audioEngine) quit() error {
	a.envelope.close()

	var errorList *errorList
	for i := len(a.freeFuncs) - 1; i >= 0; i-- {
		errorList = errorList.Add(a.freeFuncs[i]())
	}

	return errorList.Errorf("audioEngine.quit: %s", errorList)
}

func (a *audioEngine) init(lowLatency bool) error {
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("audioEngine.init: unable to initialize portaudio: %s", err)
	}
	a.deferFn(portaudio.Terminate)

	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return fmt.Errorf("audioEngine.init: unable to get default host api: %s", err)
	}

	if lowLatency {
		a.streamParams = portaudio.LowLatencyParameters(nil, host.DefaultOutputDevice)
	} else {
		a.streamParams = portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	}

	a.streamParams.SampleRate = 48000
	a.streamParams.FramesPerBuffer = 2048

	fmt.Println("SampleRate", a.streamParams.SampleRate)

	a.envelope = newEnvelope(float32(a.streamParams.SampleRate/2.0), float32(a.streamParams.SampleRate/2.0))

	stream, err := portaudio.OpenStream(a.streamParams, a.audioCallback)
	if err != nil {
		return fmt.Errorf("audioEngine.init: unable to open stream: %s", err)
	}
	a.stream = stream
	a.deferFn(a.stream.Close)

	return nil
}

func (a *audioEngine) play() error {
	a.envelope.open()
	if err := a.stream.Start(); err != nil {
		return fmt.Errorf("audioEngine.play: unable to start stream: %s", err)
	}
	return nil
}

func (a *audioEngine) pause() error {
	a.envelope.close()
	if err := a.stream.Stop(); err != nil {
		return fmt.Errorf("audioEngine.pause: unable to stop stream: %s", err)
	}
	return nil
}

func (a *audioEngine) audioCallback(out []float32) {
	channels := a.streamParams.Output.Channels

	for i := 0; i < len(out); i += channels {
		var f float32
		select {
		case f = <-a.AudioChan:
		default:
		}
		f *= a.envelope.gain()
		out[i] = f
		out[i+channels-1] = f
	}
}

const (
	envOpen int32 = iota
	envSustain
	envClose
)

type envelope struct {
	state      int32
	attackRate float32
	decayRate  float32
	step       float32
}

func newEnvelope(attackDurSamples, decayDurSamples float32) *envelope {
	return &envelope{
		attackRate: 1.0 / attackDurSamples,
		decayRate:  1.0 / decayDurSamples,
	}
}

func (e *envelope) gain() float32 {
	s := atomic.LoadInt32(&e.state)
	switch s {
	case envOpen:
		e.step = e.step + e.attackRate
		if e.step >= 1.0 {
			e.step = 1.0
			atomic.StoreInt32(&e.state, envSustain)
		}
	case envClose:
		e.step = e.step - e.decayRate
		if e.step <= 0.01 {
			e.step = 0.0
			atomic.StoreInt32(&e.state, envSustain)
		}
	}
	return e.step
}

func (e *envelope) open() {
	atomic.StoreInt32(&e.state, envOpen)
}

func (e *envelope) close() {
	atomic.StoreInt32(&e.state, envClose)
}

func (e *envelope) closed() bool {
	return atomic.LoadInt32(&e.state) == envClose
}
