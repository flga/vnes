package nes

const ramSize = 2048

type ram struct {
	data []byte
}

func newRam() *ram {
	return &ram{
		data: make([]byte, ramSize),
	}
}

func (r *ram) read(address uint16) byte {
	return r.data[address%ramSize]
}

func (r *ram) write(address uint16, value byte) {
	r.data[address%ramSize] = value
}
