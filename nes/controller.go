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

type Controller struct {
	buttons [8]Button
	head    byte
	strobe  byte
}

func (c *Controller) Read() Button {
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

func (c *Controller) Write(value byte) {
	c.strobe = value
	if c.strobe&1 == 1 {
		c.head = 0
	}
}

func (c *Controller) Press(button Button) {
	c.buttons[button] = 1
}

func (c *Controller) Release(button Button) {
	c.buttons[button] = 0
}
