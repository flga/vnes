package main

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type padding struct {
	top, right, bottom, left int32
}

type margin padding

type anchorMode byte

const (
	topLeft anchorMode = iota
	topCenter
	topRight
	centerLeft
	centerCenter
	centerRight
	bottomLeft
	bottomCenter
	bottomRight
)

func anchor(rect *sdl.Rect, a anchorMode, target *sdl.Rect, m margin) {
	switch a {
	case topLeft:
		rect.Y = target.Y + m.top
		rect.X = target.X + m.left
	case topCenter:
		rect.Y = target.Y + m.top
		rect.X = target.X + target.W/2 - rect.W/2
	case topRight:
		rect.Y = target.Y + m.top
		rect.X = target.X + target.W - rect.W - m.right

	case centerLeft:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + m.left
	case centerCenter:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + target.W/2 - rect.W/2
	case centerRight:
		rect.Y = target.Y + target.H/2 - rect.H/2
		rect.X = target.X + target.W - rect.W - m.right

	case bottomLeft:
		rect.Y = target.Y + target.H - rect.H - m.bottom
		rect.X = target.X + m.left
	case bottomCenter:
		rect.Y = target.Y + target.H - rect.H - m.bottom
		rect.X = target.X + target.W/2 - rect.W/2
	case bottomRight:
		rect.Y = target.Y + target.H - rect.H - m.bottom
		rect.X = target.X + target.W - rect.W - m.right
	}
}

func drawGrid(v *view, rows, cols int32, rect sdl.Rect, borders bool, c color.RGBA) error {
	if err := v.renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return fmt.Errorf("drawGrid: unable to set draw color: %s", err)
	}

	if rect.W == 0 {
		rect.W = v.rect.W
	}
	if rect.H == 0 {
		rect.H = v.rect.H
	}

	var b int32
	if borders {
		b = 1
	}
	for i := 1 - b; i < cols+b; i++ {
		x0 := rect.X + v.rect.X + i*rect.W/cols
		x1 := x0
		y0 := rect.Y + v.rect.Y
		y1 := y0 + v.rect.H
		if err := v.renderer.DrawLine(x0, y0, x1, y1); err != nil {
			return fmt.Errorf("drawGrid: unable to draw cols: %s", err)
		}
	}

	for i := 1 - b; i < rows+b; i++ {
		x0 := rect.X + v.rect.X
		x1 := x0 + v.rect.W
		y0 := rect.Y + v.rect.Y + i*rect.H/rows
		y1 := y0
		if err := v.renderer.DrawLine(x0, y0, x1, y1); err != nil {
			return fmt.Errorf("drawGrid: unable to draw rows: %s", err)
		}
	}

	return nil
}

func drawMessage(v *view, m string, font *ttf.Font, pad padding, marg margin, position anchorMode, foreground, background color.RGBA) error {
	surface, err := font.RenderUTF8Blended(m, sdl.Color{R: foreground.R, G: foreground.G, B: foreground.B, A: foreground.A})
	if err != nil {
		return fmt.Errorf("drawMessage: unable to create message surface: %s", err)
	}
	defer surface.Free()

	bgrect := &sdl.Rect{
		W: surface.W + pad.left + pad.right,
		H: surface.H + pad.top + pad.bottom,
	}
	anchor(bgrect, position, v.rect, marg)

	if err := drawRect(v, bgrect, background); err != nil {
		return fmt.Errorf("drawMessage: unable to draw background: %s", err)
	}

	texture, err := v.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return fmt.Errorf("drawMessage: unable to create message texture: %s", err)
	}
	defer texture.Destroy()

	msgRect := &sdl.Rect{
		W: surface.W,
		H: surface.H,
	}
	anchor(msgRect, position, v.rect, margin{
		top:    pad.top + marg.top,
		right:  pad.right + marg.right,
		bottom: pad.bottom + marg.bottom,
		left:   pad.left + marg.left,
	})

	if err := v.renderer.Copy(texture, nil, msgRect); err != nil {
		return fmt.Errorf("drawMessage: unable to render message: %s", err)
	}

	return nil
}

func drawRect(v *view, rect *sdl.Rect, c color.RGBA) error {
	if err := v.renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return fmt.Errorf("drawRect: unable to set color: %s", err)
	}
	if err := v.renderer.FillRect(rect); err != nil {
		return fmt.Errorf("drawRect: unable to render rect: %s", err)
	}
	return nil
}

func drawRGBA(v *view, data []byte) error {
	pixels, _, err := v.texture.Lock(nil)
	if err != nil {
		return fmt.Errorf("unable to lock main texture: %s", err)
	}

	copy(pixels, data)
	v.texture.Unlock()

	if err := v.renderer.Copy(v.texture, nil, v.rect); err != nil {
		return fmt.Errorf("unable to copy main texture: %s", err)
	}

	return nil
}
