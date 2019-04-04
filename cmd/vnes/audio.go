package main

import (
	"fmt"
	"sync/atomic"

	"github.com/flga/nes/cmd/internal/errors"

	"github.com/gordonklaus/portaudio"
)

type audioEngine struct {
	audioChan <-chan float32

	envelope     *envelope
	streamParams portaudio.StreamParameters
	stream       *portaudio.Stream
}

func (a *audioEngine) quit() error {
	a.envelope.close()

	err := errors.NewList(
		a.stream.Stop(),
		a.stream.Close(),
		portaudio.Terminate(),
	)

	if err != nil {
		return fmt.Errorf("audioEngine.quit: %s", err)
	}

	return nil
}

func (a *audioEngine) init(lowLatency bool) error {
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("audioEngine.init: unable to initialize portaudio: %s", err)
	}

	host, err := portaudio.DefaultHostApi()
	if err != nil {
		return fmt.Errorf("audioEngine.init: unable to get default host api: %s", err)
	}

	if lowLatency {
		a.streamParams = portaudio.LowLatencyParameters(nil, host.DefaultOutputDevice)
	} else {
		a.streamParams = portaudio.HighLatencyParameters(nil, host.DefaultOutputDevice)
	}

	a.streamParams.FramesPerBuffer = 256

	a.envelope = newEnvelope(float32(a.streamParams.SampleRate))

	stream, err := portaudio.OpenStream(a.streamParams, a.audioCallback)
	if err != nil {
		return fmt.Errorf("audioEngine.init: unable to open stream: %s", err)
	}
	a.stream = stream

	return nil
}

func (a *audioEngine) sampleRate() float64 {
	return a.streamParams.SampleRate
}

func (a *audioEngine) setChannel(c <-chan float32) {
	a.audioChan = c
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
		case f = <-a.audioChan:
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
	step       float32
}

func newEnvelope(durSamples float32) *envelope {
	return &envelope{
		attackRate: 1.0 / durSamples,
	}
}

func (e *envelope) gain() float32 {
	s := atomic.LoadInt32(&e.state)
	switch s {
	case envOpen:
		e.step += e.attackRate
		if e.step >= 1.0 {
			e.step = 1.0
			atomic.StoreInt32(&e.state, envSustain)
		}
	case envClose:
		e.step = 0.0
		atomic.StoreInt32(&e.state, envSustain)
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
