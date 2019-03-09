package main

import (
	"math"
	"time"
)

type fpsMeter struct {
	frameTimes []float64
	head       int
}

func newFPSMeter(length int) *fpsMeter {
	return &fpsMeter{
		frameTimes: make([]float64, length),
	}
}

func (f *fpsMeter) reset() {
	f.head = 0
	for i := 0; i < len(f.frameTimes); i++ {
		f.frameTimes[i] = 0
	}
}

func (f *fpsMeter) fps() int {
	var sum float64
	for i := 0; i < len(f.frameTimes); i++ {
		sum += f.frameTimes[i]
	}
	divisor := len(f.frameTimes)
	if f.head < len(f.frameTimes) {
		divisor = f.head
	}
	avg := sum / float64(divisor)
	if avg < 0 {
		avg = 1
	}

	fps := int(math.Round(1.0 / avg))
	if fps <= 0 {
		return 0
	}
	return fps
}

func (f *fpsMeter) record(d time.Duration) {
	f.frameTimes[f.head%len(f.frameTimes)] = d.Seconds()
	f.head++
}
