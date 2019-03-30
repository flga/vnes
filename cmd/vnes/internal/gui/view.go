package gui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/flga/nes/cmd/vnes/internal/errors"
	"github.com/flga/nes/nes"

	"github.com/veandco/go-sdl2/sdl"
)

type View struct {
	ID    uint32
	title string

	width  int32
	height int32
	scale  int32

	visible    bool
	fullscreen bool
	vsync      bool

	// FlashMsg  string
	// FlashTTL  time.Time
	// StatusMsg string

	window   *sdl.Window
	Renderer *Renderer
	Rect     *sdl.Rect

	FontMap FontMap
}

func NewView(title string, w, h, scale int, windowOptions, rendererOptions uint32, blendMode sdl.BlendMode, fontCache FontMap) (*View, error) {
	v := &View{
		title:      title,
		width:      int32(w),
		height:     int32(h),
		scale:      int32(scale),
		visible:    windowOptions&sdl.WINDOW_SHOWN > 0,
		fullscreen: windowOptions&sdl.WINDOW_FULLSCREEN > 0 || windowOptions&sdl.WINDOW_FULLSCREEN_DESKTOP > 0,
		vsync:      rendererOptions&sdl.RENDERER_PRESENTVSYNC > 0,
	}

	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, int32(w*scale), int32(h*scale), windowOptions)
	if err != nil {
		return nil, v.Errorf("unable to create window: %s", err)
	}

	renderer, err := newRenderer(window, int32(w), int32(h), rendererOptions)
	if err != nil {
		return nil, v.Errorf("unable to create renderer: %s", err)
	}

	if err = renderer.SetDrawBlendMode(blendMode); err != nil {
		return nil, v.Errorf("unable to set draw blend mode: %s", err)
	}

	id, err := window.GetID()
	if err != nil {
		return nil, v.Errorf("unable to get window id: %s", err)
	}

	v.ID = id
	v.window = window
	v.Renderer = renderer
	v.Rect = &sdl.Rect{
		X: 0,
		Y: 0,
		W: int32(w * scale),
		H: int32(h * scale),
	}
	v.FontMap = fontCache

	return v, nil
}

func (v *View) Error(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%q: %s", v.title, err)
}

func (v *View) Errorf(format string, args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("%q: %s", v.title, fmt.Sprintf(format, args...))
}

func (v *View) Destroy() error {
	return errors.NewList(v.Renderer.Destroy(), v.window.Destroy())
}

func (v *View) Focused() bool {
	return v.window.GetFlags()&sdl.WINDOW_INPUT_FOCUS > 0
}

func (v *View) Visible() bool {
	return v.visible
}

func (v *View) Fullscreen() bool {
	return v.fullscreen
}

func (v *View) VSync() bool {
	return v.vsync
}

func (v *View) Raise() {
	v.window.Raise()
}

func (v *View) Show() {
	// sdl seems to get confused, at this point it thinks the window is visible
	// and .Show() will noop, so we hide it (again) so that sdl updates its
	// status and actually shows the window afterwards
	v.window.Hide()
	v.visible = true

	v.window.Show()
	v.window.Raise()
}

func (v *View) Hide() {
	v.visible = false
	v.window.Hide()
}

func (v *View) Toggle() {
	switch {
	case !v.visible:
		v.Show()
	case v.visible && v.Focused():
		v.Hide()
	case v.visible && !v.Focused():
		v.Raise()
	}
}

func (v *View) resize() {
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

	v.Rect.W = int32(width)
	v.Rect.H = int32(height)
	v.Rect.X = int32(x)
	v.Rect.Y = int32(y)
}

func (v *View) Handle(event sdl.Event, console *nes.Console) (handled bool, err error) {
	switch evt := event.(type) {
	case *sdl.WindowEvent:
		if evt.Event == sdl.WINDOWEVENT_CLOSE {
			v.Hide()
			return true, nil
		}

		if evt.Event == sdl.WINDOWEVENT_RESIZED {
			v.resize()
			return true, nil
		}

	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F11 {
			return true, v.ToggleFullscreen()
		}
	}

	return false, nil
}

func (v *View) ToggleFullscreen() error {
	if v.fullscreen {
		v.fullscreen = false
		return v.window.SetFullscreen(0)
	}

	v.fullscreen = true
	return v.window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
}

func (v *View) ToggleVSync() error {
	v.vsync = !v.vsync
	if v.vsync {
		return sdl.GLSetSwapInterval(1)
	}

	return sdl.GLSetSwapInterval(0)
}

func (v *View) Clear(c color.RGBA) error {
	if err := v.Renderer.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		return v.Errorf("unable to set draw color: %s", err)
	}

	if err := v.Renderer.Clear(); err != nil {
		return v.Errorf("unable to clear renderer: %s", err)
	}

	return nil
}

func (v *View) Paint() {
	v.Renderer.Present()
}
