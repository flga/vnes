package gui

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

func DrawRect(renderer *Renderer, rect *sdl.Rect, c color.RGBA) error {
	if err := renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return fmt.Errorf("DrawRect: unable to set color: %s", err)
	}
	if err := renderer.FillRect(rect); err != nil {
		return fmt.Errorf("DrawRect: unable to render rect: %s", err)
	}
	return nil
}
