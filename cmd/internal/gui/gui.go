package gui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

type Component interface {
	tag() string
	Update(*View)
	Draw(*View) error
	Enabled() bool
	Enable()
	Disable()
	Toggle()
}

type Layers []Layer

func (ll Layers) New(c ...Component) Layers {
	return append(ll, Layer(c))
}

func (ll Layers) Find(tag string) Component {
	for _, l := range ll {
		for _, c := range l {
			if c.tag() == tag {
				return c
			}
		}
	}

	return nil
}

func (ll Layers) Update(v *View) {
	for _, l := range ll {
		for _, c := range l {
			c.Update(v)
		}
	}
}

func (ll Layers) Draw(v *View) error {
	for _, l := range ll {
		for _, c := range l {
			if err := c.Draw(v); err != nil {
				return err
			}
		}
	}

	return nil
}

type Layer []Component

type Padding struct {
	Top, Right, Bottom, Left int32
}

type Margin Padding

type Align byte

const (
	Left Align = 1 << iota
	Right
	Center
	Top
	Middle
	Bottom
)

func anchor(rect *sdl.Rect, a Align, target *sdl.Rect, m Margin) {
	switch a {
	case Top | Left:
		rect.Y = target.Y + m.Top
		rect.X = target.X + m.Left
	case Top | Center:
		rect.Y = target.Y + m.Top
		rect.X = target.X + target.W/2 - rect.W/2
	case Top | Right:
		rect.Y = target.Y + m.Top
		rect.X = target.X + target.W - rect.W - m.Right

	case Middle | Left:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + m.Left
	case Middle | Center:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + target.W/2 - rect.W/2
	case Middle | Right:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + target.W - rect.W - m.Right

	case Bottom | Left:
		rect.Y = target.Y + target.H - rect.H - m.Bottom
		rect.X = target.X + m.Left
	case Bottom | Center:
		rect.Y = target.Y + target.H - rect.H - m.Bottom
		rect.X = target.X + target.W/2 - rect.W/2
	case Bottom | Right:
		rect.Y = target.Y + target.H - rect.H - m.Bottom
		rect.X = target.X + target.W - rect.W - m.Right
	}
}

func drawRect(renderer *Renderer, rect *sdl.Rect, c color.RGBA) error {
	if err := renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return fmt.Errorf("DrawRect: unable to set color: %s", err)
	}
	if err := renderer.FillRect(rect); err != nil {
		return fmt.Errorf("DrawRect: unable to render rect: %s", err)
	}
	return nil
}

func round32(f float32) int32 {
	return int32(math.Round(float64(f)))
}

func mini32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func maxi32(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}
