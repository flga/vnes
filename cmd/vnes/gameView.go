package main

import (
	"fmt"

	"github.com/flga/nes/nes"

	"github.com/veandco/go-sdl2/sdl"
)

type gameView struct {
	*view

	showGrid bool
}

func newGameView(title string, scale int) (*gameView, error) {
	var w, h = 256, 240

	view, err := newView(title, w, h, scale, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE, sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, fmt.Errorf("unable to create game view: %s", err)
	}

	return &gameView{
		view: view,
	}, nil
}

func (v *gameView) visible() bool {
	return v.view.visible
}

func (v *gameView) handle(event sdl.Event, console *nes.Console) error {
	handled, err := v.view.handle(event)
	if handled {
		return err
	}

	press := func(b nes.Button, pressed bool) {
		if pressed {
			console.Controller1.Press(b)
		} else {
			console.Controller1.Release(b)
		}
	}

	switch evt := event.(type) {

	case *sdl.ControllerButtonEvent:
		switch evt.Button {
		case sdl.CONTROLLER_BUTTON_GUIDE:
			if evt.Type != sdl.CONTROLLERBUTTONUP {
				break
			}
			console.Reset()
		case sdl.CONTROLLER_BUTTON_A:
			press(nes.A, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_B:
			press(nes.B, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_START:
			press(nes.Start, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_BACK:
			press(nes.Select, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_DPAD_UP:
			press(nes.Up, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
			press(nes.Down, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
			press(nes.Left, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
			press(nes.Right, evt.Type == sdl.CONTROLLERBUTTONDOWN)
		}
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			v.showGrid = !v.showGrid
		}

		switch evt.Keysym.Sym {
		case sdl.K_r:
			if evt.Type != sdl.KEYUP {
				break
			}
			console.Reset()
		case sdl.K_a:
			press(nes.A, evt.Type == sdl.KEYDOWN)
		case sdl.K_z:
			press(nes.B, evt.Type == sdl.KEYDOWN)
		case sdl.K_s:
			press(nes.Start, evt.Type == sdl.KEYDOWN)
		case sdl.K_x:
			press(nes.Select, evt.Type == sdl.KEYDOWN)
		case sdl.K_UP:
			press(nes.Up, evt.Type == sdl.KEYDOWN)
		case sdl.K_DOWN:
			press(nes.Down, evt.Type == sdl.KEYDOWN)
		case sdl.K_LEFT:
			press(nes.Left, evt.Type == sdl.KEYDOWN)
		case sdl.K_RIGHT:
			press(nes.Right, evt.Type == sdl.KEYDOWN)
		}
	}

	return nil
}

func (v *gameView) render(console *nes.Console, meter *fpsMeter) error {
	if !v.visible() {
		return nil
	}

	if err := v.clear(black); err != nil {
		return v.errorf("unable to clear view: %s", err)
	}

	// draw main view
	if err := drawRGBA(v.view, console.Buffer().Pix); err != nil {
		return v.errorf("unable to draw game: %s", err)
	}

	fpsInfo := fmt.Sprintf(
		"%.fms, %d fps",
		console.FrameTime().Seconds()*1000,
		meter.fps(),
	)

	if err := drawMessage(
		v.view,
		fpsInfo,
		fontSmall,
		padding{top: 10, right: 10, bottom: 10, left: 10},
		margin{top: 10, right: 10},
		topRight,
		white,
		black128,
	); err != nil {
		return v.errorf("unable to render fps info: %s", err)
	}

	if err := v.drawStatus(fontLarge); err != nil {
		return v.errorf("unable to draw status: %s", err)
	}

	// draw grid
	if v.showGrid {
		if err := drawGrid(v.view, 240, 256, sdl.Rect{}, false, white64); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 30, 32, sdl.Rect{}, false, white128); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
		if err := drawGrid(v.view, 8, 8, sdl.Rect{H: v.rect.W}, true, white); err != nil {
			return v.errorf("unable to draw grid: %s", err)
		}
	}

	v.paint()
	return nil
}
