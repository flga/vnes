package gui

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

var _ Component = &Message{}

type Message struct {
	Tag      string
	UpdateFn func(m *Message)

	Disabled   bool
	Text       string
	Font       *Font
	Size       int
	Align      Align
	Padding    Padding
	Margin     Margin
	Position   Align
	Foreground color.RGBA
	Background color.RGBA

	viewRect sdl.Rect
}

func (m *Message) tag() string {
	return m.Tag
}

func (m *Message) Enabled() bool {
	return !m.Disabled
}

func (m *Message) Enable() {
	m.Disabled = false
}

func (m *Message) Disable() {
	m.Disabled = true
}

func (m *Message) Toggle() {
	m.Disabled = !m.Disabled
}

func (m *Message) Update(v *View) {
	if m.Disabled {
		return
	}

	if m.UpdateFn != nil {
		m.UpdateFn(m)
	}

	m.viewRect = *v.rect
}

func (m *Message) Draw(v *View) error {
	if m.Disabled || m.Text == "" {
		return nil
	}

	textW, textH := m.Font.Bounds(m.Text, m.Size)

	bgrect := &sdl.Rect{
		W: textW + m.Padding.Left + m.Padding.Right,
		H: textH + m.Padding.Top + m.Padding.Bottom,
	}
	anchor(bgrect, m.Position, &m.viewRect, m.Margin)

	if err := drawRect(v.renderer, bgrect, m.Background); err != nil {
		return fmt.Errorf("drawMessage: unable to draw background: %s", err)
	}

	msgRect := &sdl.Rect{
		W: textW,
		H: textH,
	}
	anchor(msgRect, m.Position, &m.viewRect, Margin{
		Top:    m.Padding.Top + m.Margin.Top,
		Right:  m.Padding.Right + m.Margin.Right,
		Bottom: m.Padding.Bottom + m.Margin.Bottom,
		Left:   m.Padding.Left + m.Margin.Left,
	})

	if _, _, err := v.renderer.DrawText(m.Text, m.Font, m.Size, m.Align, m.Foreground, msgRect); err != nil {
		return fmt.Errorf("drawMessage: unable to render message: %s", err)
	}

	return nil
}
