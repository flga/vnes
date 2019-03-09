package main

import (
	"fmt"
	"image"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type nametableView struct {
	*view

	showGrid bool
	buf      *image.RGBA
}

func newNametableView(scale int) (*nametableView, error) {
	w, h := 256*2, 240*2

	view, err := newView("vnes - nametables", w, h, scale, sdl.WINDOW_HIDDEN|sdl.WINDOW_RESIZABLE, sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, fmt.Errorf("unable to create name table view: %s", err)
	}

	return &nametableView{
		view: view,
		buf:  image.NewRGBA(image.Rect(0, 0, w, h)),
	}, nil
}

func (v *nametableView) visible() bool {
	return v.view.visible
}

func (v *nametableView) handle(event sdl.Event, console *nes.Console) error {
	handled, err := v.view.handle(event)
	if handled {
		return err
	}

	switch evt := event.(type) {
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			v.showGrid = !v.showGrid
		}
	}

	return nil
}

func (v *nametableView) render(console *nes.Console, _ *fpsMeter) error {
	if !v.visible() {
		return nil
	}

	if err := v.clear(black); err != nil {
		return v.errorf("unable to clear view: %s", err)
	}

	// draw main view
	console.PPU.DrawNametables(v.buf)
	if err := drawRGBA(v.view, v.buf.Pix); err != nil {
		return v.errorf("unable to draw nametables: %s", err)
	}

	// draw grid
	if v.showGrid {
		if err := drawGrid(v.view, 60, 64, sdl.Rect{}, false, white64); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 8, 16, sdl.Rect{H: v.rect.W / 2}, false, white128); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 8, 16, sdl.Rect{H: v.rect.W / 2, Y: v.rect.H / 2}, false, white128); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 2, 2, sdl.Rect{}, true, white); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
	}

	v.paint()
	return nil
}
