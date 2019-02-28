package nes

const ramSize = 2048

type RAM struct {
	data []byte
}

func NewRAM() *RAM {
	return &RAM{
		data: make([]byte, ramSize),
	}
}

func (r *RAM) Read(address uint16) byte {
	return r.data[address%ramSize]
}

func (r *RAM) Write(address uint16, value byte) {
	r.data[address%ramSize] = value
}
