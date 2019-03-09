package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type view struct {
	id    uint32
	title string

	width  int32
	height int32
	scale  int32

	visible    bool
	fullscreen bool

	flashMsg  string
	flashTTL  time.Time
	statusMsg string

	window   *sdl.Window
	renderer *sdl.Renderer
	rect     *sdl.Rect
	texture  *sdl.Texture

	freeFuncs []func() error
}

func newView(title string, w, h, scale int, options uint32, blendMode sdl.BlendMode) (*view, error) {
	v := &view{
		title:      title,
		width:      int32(w),
		height:     int32(h),
		scale:      int32(scale),
		visible:    options&sdl.WINDOW_SHOWN > 0,
		fullscreen: options&sdl.WINDOW_FULLSCREEN > 0 || options&sdl.WINDOW_FULLSCREEN_DESKTOP > 0,
	}

	window, renderer, err := sdl.CreateWindowAndRenderer(int32(w*scale), int32(h*scale), options)
	if err != nil {
		return nil, v.errorf("unable to create window: %s", err)
	}
	v.deferFn(window.Destroy)
	v.deferFn(renderer.Destroy)

	window.SetTitle(title)

	if err = renderer.SetDrawBlendMode(blendMode); err != nil {
		return nil, v.errorf("unable to set draw blend mode: %s", err)
	}

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		return nil, v.errorf("unable to create texture: %s", err)
	}
	v.deferFn(texture.Destroy)

	id, err := window.GetID()
	if err != nil {
		return nil, v.errorf("unable to get window id: %s", err)
	}

	v.id = id
	v.window = window
	v.renderer = renderer
	v.texture = texture
	v.rect = &sdl.Rect{
		X: 0,
		Y: 0,
		W: int32(w * scale),
		H: int32(h * scale),
	}

	return v, nil
}

func (v *view) error(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%q: %s", v.title, err)
}

func (v *view) errorf(format string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%q: %s: %s", v.title, format, err)
}

func (v *view) deferFn(f func() error) {
	v.freeFuncs = append(v.freeFuncs, f)
}

func (v *view) free() error {
	for i := len(v.freeFuncs) - 1; i >= 0; i-- {
		err := v.freeFuncs[i]()
		if err != nil {
			return v.error(err)
		}
	}

	return nil
}

func (v *view) focused() bool {
	return v.window.GetFlags()&sdl.WINDOW_INPUT_FOCUS > 0
}

func (v *view) raise() {
	v.window.Raise()
}

func (v *view) show() {
	// sdl seems to get confused, at this point it thinks the window is visible
	// and .Show() will noop, so we hide it (again) so that sdl updates its
	// status and actually shows the window afterwards
	v.window.Hide()
	v.visible = true

	v.window.Show()
	v.window.Raise()
}

func (v *view) hide() {
	v.visible = false
	v.window.Hide()
}

func (v *view) toggle() {
	switch {
	case !v.visible:
		v.show()
	case v.visible && v.focused():
		v.hide()
	case v.visible && !v.focused():
		v.raise()
	}
}

func (v *view) resize() {
	minHeight := float64(v.height)
	minWidth := float64(v.width)

	wf, hf := v.window.GetSize()
	width := float64(wf)
	height := float64(hf)
	var x, y float64 = 0, 0

	origW, origH := width, height
	height = math.Floor(width * (minHeight / minWidth))
	if height > origH {
		width = math.Floor(origH * (minWidth / minHeight))
		height = math.Floor(width * (minHeight / minWidth))
	}

	if width > origW {
		x = (width - origW) / 2
	} else {
		x = (origW - width) / 2
	}
	if height > origH {
		y = (height - origH) / 2
	} else {
		y = (origH - height) / 2
	}

	v.rect.W = int32(width)
	v.rect.H = int32(height)
	v.rect.X = int32(x)
	v.rect.Y = int32(y)
}

func (v *view) handle(event sdl.Event) (handled bool, err error) {
	switch evt := event.(type) {
	case *sdl.WindowEvent:
		if evt.Event == sdl.WINDOWEVENT_CLOSE {
			v.hide()
			return true, nil
		}

		if evt.Event == sdl.WINDOWEVENT_RESIZED {
			v.resize()
			return true, nil
		}

	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F11 {
			if v.fullscreen {
				v.window.SetFullscreen(0)
				v.fullscreen = false
				return true, nil
			} else {
				v.window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
				v.fullscreen = true
				return true, nil
			}
		}
	}

	return false, nil
}

func (v *view) setFlashMsg(m string) {
	v.flashMsg = m
	v.flashTTL = time.Now().Add(2 * time.Second)
}

func (v *view) setStatusMsg(m string) {
	v.statusMsg = m
}

func (v *view) drawStatus(font *ttf.Font) error {
	if v.statusMsg != "" {
		err := drawMessage(
			v,
			v.statusMsg,
			font,
			padding{top: 10, right: 10, bottom: 10, left: 10},
			margin{},
			centerCenter,
			color.RGBA{255, 255, 255, 255},
			color.RGBA{0, 0, 0, 200},
		)
		return v.errorf("unable to display status msg: %s", err)
	}

	if time.Now().Before(v.flashTTL) {
		err := drawMessage(
			v,
			v.flashMsg,
			font,
			padding{top: 10, right: 10, bottom: 10, left: 10},
			margin{},
			centerCenter,
			color.RGBA{255, 255, 255, 255},
			color.RGBA{0, 0, 0, 200},
		)
		return v.errorf("unable to display flash msg: %s", err)
	}
	return nil
}

func (v *view) clear(c color.RGBA) error {
	if err := v.renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return v.errorf("unable to set draw color: %s", err)
	}

	if err := v.renderer.Clear(); err != nil {
		return v.errorf("unable to clear renderer: %s", err)
	}

	return nil
}

func (v *view) paint() {
	v.renderer.Present()
}
