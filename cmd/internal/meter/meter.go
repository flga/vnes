package meter

import (
	"math"
	"time"
)

const DefaultBufferLen = 50

type Meter struct {
	times []float64
	head  int
}

func New(bufferLength int) *Meter {
	return &Meter{
		times: make([]float64, bufferLength),
	}
}

func (m *Meter) init() {
	if m.times == nil {
		m.times = make([]float64, DefaultBufferLen)
	}
}

func (m *Meter) Reset() {
	m.init()

	m.head = 0
	for i := 0; i < len(m.times); i++ {
		m.times[i] = 0
	}
}

func (m *Meter) Tps() int {
	m.init()

	var sum float64
	for i := 0; i < len(m.times); i++ {
		sum += m.times[i]
	}
	divisor := len(m.times)
	if m.head < len(m.times) {
		divisor = m.head
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

func (m *Meter) Ms() float64 {
	m.init()

	var sum float64
	for i := 0; i < len(m.times); i++ {
		sum += m.times[i]
	}
	divisor := len(m.times)
	if m.head < len(m.times) {
		divisor = m.head
	}
	avg := sum / float64(divisor)
	if avg < 0 {
		avg = 1
	}

	return avg * 1000
}

func (m *Meter) Record(d time.Duration) {
	m.times[m.head%len(m.times)] = d.Seconds()
	m.head++
}
