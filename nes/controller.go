package nes

type Button byte

const (
	A Button = iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
)

type controller struct {
	buttons [8]Button
	head    byte
	strobe  byte
}

func (c *controller) read() Button {
	var value Button
	if c.head < 8 {
		value = c.buttons[c.head]
	} else {
		value = 0
	}
	c.head++
	if c.strobe&1 == 1 {
		c.head = 0
	}
	return value
}

func (c *controller) write(value byte) {
	c.strobe = value
	if c.strobe&1 == 1 {
		c.head = 0
	}
}

func (c *controller) press(button Button) {
	c.buttons[button] = 1
}

func (c *controller) release(button Button) {
	c.buttons[button] = 0
}
