package gui

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

var _ Component = &Menu{}

type Cell struct {
	UpdateFn func() string
	Text     string
	Font     *Font
	Size     int
	Padding  Padding
	Color    color.RGBA
	Hover    color.RGBA

	width  int32
	height int32
}

func (c *Cell) PaddedBounds() (w, h int32) {
	return c.PaddedWidth(), c.PaddedHeight()
}

func (c *Cell) PaddedWidth() int32 {
	return c.Padding.Left + c.width + c.Padding.Right
}

func (c *Cell) PaddedHeight() int32 {
	return c.Padding.Top + c.height + c.Padding.Bottom
}

func (c *Cell) Update(*View) {
	if c.UpdateFn != nil {
		c.Text = c.UpdateFn()
	}
	c.width, c.height = c.Font.Bounds(c.Text, c.Size)
}

type MenuItem struct {
	Label    Cell
	Value    Cell
	Callback func() error
}

func (item *MenuItem) Update(v *View) {
	item.Label.Update(v)
	item.Value.Update(v)
}

func (item *MenuItem) Visible() bool {
	lw := item.Label.width
	lh := item.Label.height
	vw := item.Value.width
	vh := item.Value.height

	return (lw != 0 && lh != 0) || (vw != 0 || vh != 0)
}

type Menu struct {
	Tag        string
	Disabled   bool
	Position   AnchorMode
	Margin     Margin
	Background color.RGBA
	Backdrop   color.RGBA
	Items      []MenuItem

	focus int
}

func (m *Menu) tag() string {
	return m.Tag
}

func (m *Menu) Enabled() bool {
	return !m.Disabled
}

func (m *Menu) Enable() {
	m.Disabled = false
}

func (m *Menu) Disable() {
	m.Disabled = true
}

func (m *Menu) Toggle() {
	m.Disabled = !m.Disabled
}

func (m *Menu) Update(v *View) {
	if m.Disabled {
		return
	}

	for i := 0; i < len(m.Items); i++ {
		m.Items[i].Update(v)
	}
}

func (m *Menu) Down() {
	if m.Disabled {
		return
	}

	m.focus++
	if m.focus >= len(m.Items) {
		m.focus = 0
	}

	if !m.Items[m.focus].Visible() {
		m.Down()
	}
}

func (m *Menu) Up() {
	if m.Disabled {
		return
	}

	m.focus--
	if m.focus < 0 {
		m.focus = len(m.Items) - 1
	}

	if !m.Items[m.focus].Visible() {
		m.Up()
	}
}

func (m *Menu) Activate() error {
	if m.Disabled {
		return nil
	}

	return m.Items[m.focus].Callback()
}

func (m *Menu) Draw(v *View) error {
	if m.Disabled {
		return nil
	}

	var (
		maxLabelWidth int32
		maxValueWidth int32
		height        int32
	)

	// compute column positions
	for i := 0; i < len(m.Items); i++ {
		item := m.Items[i]
		if !item.Visible() {
			continue
		}

		lw, lh := item.Label.PaddedBounds()
		vw, vh := item.Value.PaddedBounds()

		maxLabelWidth = maxi32(lw, maxLabelWidth)
		maxValueWidth = maxi32(vw, maxValueWidth)

		height += maxi32(lh, vh)
	}

	// draw background
	bgRect := &sdl.Rect{
		X: 0,
		Y: 0,
		W: m.Margin.Left + maxLabelWidth + maxValueWidth + m.Margin.Right,
		H: m.Margin.Top + height + m.Margin.Bottom,
	}
	viewport := v.Renderer.GetViewport()
	anchor(bgRect, m.Position, &viewport, m.Margin)

	if err := DrawRect(v.Renderer, nil, m.Backdrop); err != nil {
		return fmt.Errorf("menu: unable to draw overlay: %s", err)
	}

	if err := DrawRect(v.Renderer, bgRect, m.Background); err != nil {
		return fmt.Errorf("menu: unable to draw background: %s", err)
	}

	// draw menu
	x0 := bgRect.X + m.Margin.Left
	y0 := bgRect.Y + m.Margin.Top
	y := int32(0)

	for i := 0; i < len(m.Items); i++ {
		item := m.Items[i]

		if !item.Visible() {
			continue
		}

		labelColor := item.Label.Color
		valueColor := item.Value.Color
		if m.focus == i {
			labelColor = item.Label.Hover
			valueColor = item.Value.Hover
		}

		_, lh, err := v.Renderer.DrawText(item.Label.Text, item.Label.Font, item.Label.Size, labelColor, &sdl.Rect{
			X: x0 + item.Label.Padding.Left,
			Y: y0 + item.Label.Padding.Top + y,
		})
		if err != nil {
			return fmt.Errorf("menu: unable to draw label %q: %s", item.Label.Text, err)
		}

		_, vh, err := v.Renderer.DrawText(item.Value.Text, item.Value.Font, item.Value.Size, valueColor, &sdl.Rect{
			X: x0 + maxLabelWidth + item.Value.Padding.Left,
			Y: y0 + y + item.Value.Padding.Top,
		})
		if err != nil {
			return fmt.Errorf("menu: unable to draw value %q of label %q: %s", item.Value.Text, item.Label.Text, err)
		}

		y += maxi32(lh, vh) +
			maxi32(item.Label.Padding.Top, item.Value.Padding.Top) +
			maxi32(item.Label.Padding.Bottom, item.Value.Padding.Bottom)
	}

	return nil
}
