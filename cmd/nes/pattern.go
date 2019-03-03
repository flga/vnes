package main

import (
	"fmt"
	"image"
	"time"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type patternWindow struct {
	baseWidth  int32
	baseHeight int32
	visible    bool
	window     *sdl.Window
	renderer   *sdl.Renderer
	tex        *sdl.Texture
	buf        *image.RGBA
	rect       *sdl.Rect
	showGrid   bool
	paletteNum byte
}

func newPatternWindow(scale int32) (*patternWindow, uint32, error) {
	var baseWidth, baseHeight int32 = 256, 128

	window, renderer, err := sdl.CreateWindowAndRenderer(baseWidth*scale, baseHeight*scale, sdl.WINDOW_HIDDEN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create pattern window: %s", err)
	}
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	id, err := window.GetID()
	if err != nil {
		return nil, 0, fmt.Errorf("unable to get pattern window id: %s", err)
	}

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, baseWidth, baseHeight)
	if err != nil {
		return nil, id, fmt.Errorf("unable to create pattern texture: %s", err)
	}

	buf := image.NewRGBA(image.Rect(0, 0, int(baseWidth), int(baseHeight)))
	rect := &sdl.Rect{X: 0, Y: 0, W: baseWidth * scale, H: baseHeight * scale}

	window.SetTitle("pattern table")

	return &patternWindow{
		baseWidth:  baseWidth,
		baseHeight: baseHeight,
		window:     window,
		renderer:   renderer,
		tex:        tex,
		buf:        buf,
		rect:       rect,
	}, id, nil
}

func (w *patternWindow) Render(console *nes.Console, _ time.Duration) error {
	console.PPU.DrawPatternTables(w.buf, w.paletteNum)

	pixels, _, err := w.tex.Lock(nil)
	if err != nil {
		return fmt.Errorf("unable to lock pattern texture: %s", err)
	}

	copy(pixels, w.buf.Pix)
	w.tex.Unlock()

	if err := w.renderer.Clear(); err != nil {
		return fmt.Errorf("unable to clear pattern renderer: %s", err)
	}

	if err := w.renderer.Copy(w.tex, nil, w.rect); err != nil {
		return fmt.Errorf("unable to copy pattern: %s", err)
	}

	if w.showGrid {
		w.renderer.SetDrawColor(255, 255, 255, 255/2)
		for i := int32(0); i < 16; i++ {
			cellHeight := i * w.rect.H / 16
			w.renderer.DrawLine(
				w.rect.X, w.rect.Y+cellHeight,
				w.rect.X+w.rect.W, w.rect.Y+cellHeight,
			)
		}
		for i := int32(0); i < 32; i++ {
			cellWidth := i * w.rect.W / 32
			w.renderer.DrawLine(
				w.rect.X+cellWidth, w.rect.Y,
				w.rect.X+cellWidth, w.rect.Y+w.rect.H,
			)
		}

		w.renderer.SetDrawColor(0x13, 0xE0, 0xD7, 255)
		w.renderer.DrawLine(
			w.rect.X+w.rect.W/2, w.rect.Y,
			w.rect.X+w.rect.W/2, w.rect.Y+w.rect.H,
		)
	}

	w.renderer.SetDrawColor(255, 0, 0, 255)
	w.renderer.Present()
	return nil
}

func (w *patternWindow) Handle(event sdl.Event, console *nes.Console) error {
	// fmt.Printf("%T, handle %T\n", w, event)
	switch evt := event.(type) {
	case *sdl.WindowEvent:
		if evt.Event == sdl.WINDOWEVENT_CLOSE {
			w.hide()
		}

		if evt.Event == sdl.WINDOWEVENT_RESIZED {
			resize(w.window, float64(w.baseWidth), float64(w.baseHeight), w.rect)
		}
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			w.showGrid = !w.showGrid
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_UP {
			if w.paletteNum == 7 {
				w.paletteNum = 0
			} else {
				w.paletteNum++
			}
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_DOWN {
			if w.paletteNum == 0 {
				w.paletteNum = 7
			} else {
				w.paletteNum--
			}
		}
	}

	return nil
}

func (w *patternWindow) Visible() bool {
	return w.visible
}

func (w *patternWindow) show() {
	// sdl seems to get confused, at this point it thinks the window is visible
	// and .Show() will noop, so we hide it (again) so that sdl updates its
	// status and actually shows the window afterwards
	w.hide()

	w.visible = true
	w.window.Show()
	w.window.Raise()
}

func (w *patternWindow) hide() {
	w.visible = false
	w.window.Hide()
}

func (w *patternWindow) Toggle() {
	if w.visible {
		w.hide()
	} else {
		w.show()
	}
}

func (w *patternWindow) Free() error {
	if w.tex != nil {
		if err := w.tex.Destroy(); err != nil {
			return err
		}
	}
	if w.renderer != nil {
		if err := w.renderer.Destroy(); err != nil {
			return err
		}
	}
	if w.window != nil {
		if err := w.window.Destroy(); err != nil {
			return err
		}
	}

	return nil
}
