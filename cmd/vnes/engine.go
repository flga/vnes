package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flga/nes/cmd/internal/gui"
	"github.com/flga/nes/cmd/internal/meter"
	"github.com/flga/nes/nes"

	"github.com/veandco/go-sdl2/sdl"
)

var errQuit = errors.New("quit requested")

type view interface {
	Title() string
	Init(*engine, *nes.Console) error
	Destroy() error
	Visible() bool
	Handle(event sdl.Event, engine *engine, console *nes.Console) (handled bool, err error)
	Update(*nes.Console, *engine)
	Render() error
}

type controllers []*sdl.GameController

func (c controllers) which(id sdl.JoystickID) int {
	for i, ctrl := range c {
		if ctrl.Joystick().InstanceID() == id {
			return i
		}
	}

	return 0
}

type engine struct {
	audio *audioEngine

	paused bool

	fpsMeter     *meter.Meter
	paintMeter   *meter.Meter
	consoleMeter *meter.Meter

	mainView      *gameView
	patternView   *patternView
	nametableView *nametableView

	// viewsById   map[uint32]handler
	views       []view
	controllers controllers
}

func newEngine(title string, zoom int, audio *audioEngine, fontCache gui.FontMap) (*engine, error) {
	e := &engine{
		audio:        audio,
		fpsMeter:     meter.New(10),
		paintMeter:   meter.New(10),
		consoleMeter: meter.New(10),
	}

	gameView, err := newGameView(title, zoom, fontCache)
	if err != nil {
		return nil, fmt.Errorf("newEngine: unable to create game window: %s", err)
	}

	patternView, err := newPatternView(zoom, fontCache)
	if err != nil {
		return nil, fmt.Errorf("newEngine: unable to create pattern window: %s", err)
	}

	nametableView, err := newNametableView(zoom/2, fontCache)
	if err != nil {
		return nil, fmt.Errorf("newEngine: unable to create nametable window: %s", err)
	}

	e.mainView = gameView
	e.patternView = patternView
	e.nametableView = nametableView
	e.views = []view{
		gameView,
		patternView,
		nametableView,
	}

	return e, nil
}

func (e *engine) run(ctx context.Context, console *nes.Console) error {
	for _, v := range e.views {
		if err := v.Init(e, console); err != nil {
			return fmt.Errorf("engine: run: unable to init view %s: %s", v.Title(), err)
		}

		defer v.Destroy()
	}

	if err := e.audio.play(); err != nil {
		return fmt.Errorf("engine: run: unable to start audio")
	}

	defer fmt.Println("engine: run: done")

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := e.poll(console); err != nil {
				if err == errQuit {
					return err
				}
				return fmt.Errorf("engine: poll: %s", err)
			}

			if !e.mainView.Visible() {
				return errQuit
			}

			if !e.paused {
				start := time.Now()
				console.StepFrame()
				e.consoleMeter.Record(time.Since(start))
			}

			if err := e.render(console); err != nil {
				return err
			}
			e.fpsMeter.Record(time.Since(start))
			start = time.Now()
		}
	}

	return nil
}

func (e *engine) poll(console *nes.Console) error {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch evt := event.(type) {

		case *sdl.QuitEvent:
			return errQuit

		case *sdl.ControllerDeviceEvent:
			for _, ctrl := range e.controllers {
				ctrl.Close()
			}
			e.controllers = e.controllers[:0]

			for i := 0; i < sdl.NumJoysticks(); i++ {
				ctrl := sdl.GameControllerOpen(i)
				e.controllers = append(e.controllers, ctrl)
			}

			return nil

		case *sdl.KeyboardEvent:
			if isKeyPress(evt, sdl.K_SPACE) {
				return e.pauseUnpause()
			}

			if isKeyPress(evt, sdl.K_F1) {
				e.patternView.Toggle()
				return nil
			}

			if isKeyPress(evt, sdl.K_F2) {
				e.nametableView.Toggle()
				return nil
			}

			return e.dispatch(evt, console)

		default:
			return e.dispatch(evt, console)
		}
	}

	return nil
}

func (e *engine) pauseUnpause() error {
	e.paused = !e.paused
	if e.paused {
		e.mainView.SetStatusMsg("paused")
		if err := e.audio.pause(); err != nil {
			return err
		}
	} else {
		e.mainView.SetStatusMsg("")
		e.mainView.SetFlashMsg("unpaused")
		if err := e.audio.play(); err != nil {
			return err
		}
	}

	return nil
}

func (e *engine) dispatch(evt sdl.Event, console *nes.Console) error {
	for _, v := range e.views {
		if handled, err := v.Handle(evt, e, console); handled {
			return err
		}
	}

	return nil
}

func (e *engine) render(console *nes.Console) error {
	paintStart := time.Now()
	for _, v := range e.views {
		if !v.Visible() {
			continue
		}

		v.Update(console, e)

		if err := v.Render(); err != nil {
			return err
		}
	}
	e.paintMeter.Record(time.Since(paintStart))

	return nil
}
