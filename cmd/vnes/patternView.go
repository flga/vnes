package main

import (
	"fmt"
	"image"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type patternView struct {
	*view

	buf        *image.RGBA
	showGrid   bool
	paletteNum byte
}

func newPatternView(scale int) (*patternView, error) {
	w, h := 256, 128

	view, err := newView("vnes - pattern tables", w, h, scale, sdl.WINDOW_HIDDEN|sdl.WINDOW_RESIZABLE, sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, fmt.Errorf("unable to create pattern table view: %s", err)
	}

	return &patternView{
		view: view,
		buf:  image.NewRGBA(image.Rect(0, 0, w, h)),
	}, nil
}

func (v *patternView) visible() bool {
	return v.view.visible
}

func (v *patternView) handle(event sdl.Event, console *nes.Console) error {
	handled, err := v.view.handle(event)
	if handled {
		return err
	}

	switch evt := event.(type) {
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			v.showGrid = !v.showGrid
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_UP {
			if v.paletteNum == 7 {
				v.paletteNum = 0
			} else {
				v.paletteNum++
			}
			v.setFlashMsg(fmt.Sprintf("palette %d", v.paletteNum))
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_DOWN {
			if v.paletteNum == 0 {
				v.paletteNum = 7
			} else {
				v.paletteNum--
			}
			v.setFlashMsg(fmt.Sprintf("palette %d", v.paletteNum))
		}
	}

	return nil
}

func (v *patternView) render(console *nes.Console, _ *fpsMeter) error {
	if !v.visible() {
		return nil
	}

	if err := v.clear(black); err != nil {
		return v.errorf("unable to clear view: %s", err)
	}

	// draw main view
	console.PPU.DrawPatternTables(v.buf, v.paletteNum)
	if err := drawRGBA(v.view, v.buf.Pix); err != nil {
		return v.errorf("unable to draw pattern tables: %s", err)
	}

	if err := v.drawStatus(fontLarge); err != nil {
		return v.errorf("unable to draw status: %s", err)
	}

	// draw grid
	if v.showGrid {
		if err := drawGrid(v.view, 16*8, 32*8, sdl.Rect{}, false, white64); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 16, 32, sdl.Rect{}, false, white128); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 1, 2, sdl.Rect{}, true, white); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
	}

	v.paint()
	return nil
}
