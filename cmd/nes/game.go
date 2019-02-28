package main

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/ttf"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type gameWindow struct {
	baseWidth  int32
	baseHeight int32
	visible    bool
	showGrid   bool
	window     *sdl.Window
	renderer   *sdl.Renderer
	tex        *sdl.Texture
	rect       *sdl.Rect
	font       *ttf.Font
}

func newGameWindow(scale int32, title string) (*gameWindow, uint32, error) {
	var baseWidth, baseHeight int32 = 256, 240

	window, renderer, err := sdl.CreateWindowAndRenderer(baseWidth*scale, baseHeight*scale, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create game window: %s", err)
	}
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	id, err := window.GetID()
	if err != nil {
		return nil, 0, fmt.Errorf("unable to get game window id: %s", err)
	}

	font, err := ttf.OpenFont("assets/runescape_uf.ttf", 6*int(scale))
	if err != nil {
		return nil, 0, fmt.Errorf("unable to open font: %s", err)
	}

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, baseWidth, baseHeight)
	if err != nil {
		return nil, id, fmt.Errorf("unable to create game texture: %s", err)
	}

	rect := &sdl.Rect{X: 0, Y: 0, W: baseWidth * scale, H: baseHeight * scale}

	window.SetTitle(title)

	return &gameWindow{
		baseWidth:  baseWidth,
		baseHeight: baseHeight,
		visible:    true,
		window:     window,
		renderer:   renderer,
		tex:        tex,
		font:       font,
		rect:       rect,
	}, id, nil
}

func (w *gameWindow) Render(console *nes.Console, fps time.Duration) error {
	buf := console.Buffer()

	pixels, _, err := w.tex.Lock(nil)
	if err != nil {
		return fmt.Errorf("unable to lock game texture: %s", err)
	}
	copy(pixels, buf.Pix)
	w.tex.Unlock()

	if err := w.renderer.Clear(); err != nil {
		return fmt.Errorf("unable to clear game renderer: %s", err)
	}

	if err := w.renderer.Copy(w.tex, nil, w.rect); err != nil {
		return fmt.Errorf("unable to copy game: %s", err)
	}

	fpsSur, err := w.font.RenderUTF8Solid(fmt.Sprintf("%.fms/%.fms", fps.Seconds()*1000, console.FrameTime().Seconds()*1000), sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return fmt.Errorf("unable to create fps surface: %s", err)
	}
	defer fpsSur.Free()

	fpsTex, err := w.renderer.CreateTextureFromSurface(fpsSur)
	if err != nil {
		return fmt.Errorf("unable to create fps texture: %s", err)
	}
	defer fpsTex.Destroy()

	if err := w.renderer.Copy(fpsTex, nil, &sdl.Rect{W: fpsSur.W, H: fpsSur.H, X: w.rect.X + w.rect.W - fpsSur.W - 10, Y: w.rect.Y + 10}); err != nil {
		return fmt.Errorf("unable to copy game: %s", err)
	}

	if w.showGrid {
		w.renderer.SetDrawColor(255, 255, 255, 255/2)
		for i := int32(0); i < 30; i++ {
			cellHeight := i * w.rect.H / 30
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
		for i := int32(0); i < 8; i++ {
			cellHeight := i * w.rect.W / 8
			w.renderer.DrawLine(
				w.rect.X, w.rect.Y+cellHeight,
				w.rect.X+w.rect.W, w.rect.Y+cellHeight,
			)
		}
		for i := int32(0); i < 8; i++ {
			cellWidth := i * w.rect.W / 8
			w.renderer.DrawLine(
				w.rect.X+cellWidth, w.rect.Y,
				w.rect.X+cellWidth, w.rect.Y+w.rect.H,
			)
		}
	}
	w.renderer.SetDrawColor(0, 0, 0, 255)
	w.renderer.Present()

	return nil
}

func (w *gameWindow) Handle(event sdl.Event, console *nes.Console) error {
	// fmt.Printf("%T, handle %T\n", w, event)
	press := func(b nes.Button, pressed bool) {
		if pressed {
			console.Controller1.Press(b)
		} else {
			console.Controller1.Release(b)
		}
	}

	switch evt := event.(type) {
	case *sdl.WindowEvent:
		if evt.Event == sdl.WINDOWEVENT_CLOSE {
			w.hide()
		}

		if evt.Event == sdl.WINDOWEVENT_RESIZED {
			resize(w.window, float64(w.baseWidth), float64(w.baseHeight), w.rect)
		}
	case *sdl.ControllerButtonEvent:
		if evt.Button == sdl.CONTROLLER_BUTTON_GUIDE {
			console.Reset()
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_A {
			press(nes.A, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_B {
			press(nes.B, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_START {
			press(nes.Start, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_BACK {
			press(nes.Select, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_DPAD_UP {
			press(nes.Up, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_DPAD_DOWN {
			press(nes.Down, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_DPAD_LEFT {
			press(nes.Left, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
		if evt.Button == sdl.CONTROLLER_BUTTON_DPAD_RIGHT {
			press(nes.Right, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			w.showGrid = !w.showGrid
		}

		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_r {
			console.Reset()
		}

		if evt.Keysym.Sym == sdl.K_a {
			press(nes.A, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_z {
			press(nes.B, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_s {
			press(nes.Start, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_x {
			press(nes.Select, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_UP {
			press(nes.Up, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_DOWN {
			press(nes.Down, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_LEFT {
			press(nes.Left, evt.Type == sdl.KEYDOWN)
		}
		if evt.Keysym.Sym == sdl.K_RIGHT {
			press(nes.Right, evt.Type == sdl.KEYDOWN)
		}
	}

	return nil
}

func (w *gameWindow) Visible() bool {
	return w.visible
}

func (w *gameWindow) show() {
	w.visible = true
	w.window.Show()
}

func (w *gameWindow) hide() {
	w.visible = false
	w.window.Hide()
}

func (w *gameWindow) Toggle() {
	if w.visible {
		w.hide()
	} else {
		w.show()
	}
}

func (w *gameWindow) Free() error {
	if w.font != nil {
		w.font.Close()
	}
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
