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

var speedTable = map[sdl.Keycode]time.Duration{
	sdl.K_1: time.Second / 1,
	sdl.K_2: time.Second / 2,
	sdl.K_3: time.Second / 4,
	sdl.K_4: time.Second / 10,
	sdl.K_5: time.Second / 30,
	sdl.K_6: time.Second / 60,
	sdl.K_7: time.Second / 80,
	sdl.K_8: time.Second / 120,
	sdl.K_9: time.Second / 144,
}

type handler interface {
	Handle(event sdl.Event, console *nes.Console) (handled bool, err error)
	Destroy() error
	Init(*engine, *nes.Console) error
	Visible() bool
	Update(*nes.Console, *engine)
	Render() error
}

type engine struct {
	audio *audioEngine

	paused bool

	// ticker       *time.Ticker
	ticker       *time.Ticker
	tickerChan   <-chan time.Time
	fpsMeter     *meter.Meter
	paintMeter   *meter.Meter
	consoleMeter *meter.Meter

	mainView      *gameView
	patternView   *patternView
	nametableView *nametableView

	viewsById   map[uint32]handler
	controllers []*sdl.GameController
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
	e.viewsById = map[uint32]handler{
		gameView.ID:      gameView,
		patternView.ID:   patternView,
		nametableView.ID: nametableView,
	}

	return e, nil
}

func (e *engine) run(ctx context.Context, console *nes.Console) error {
	for _, v := range e.viewsById {
		if err := v.Init(e, console); err != nil {
			return err
		}
	}

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	e.ticker = ticker
	e.tickerChan = ticker.C

	e.audio.play()

	defer func() {
		for _, w := range e.viewsById {
			w.Destroy()
		}
	}()
	defer func() {
		fmt.Println("done")
	}()

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
			// case <-e.tickerChan:
		default:
			if err := e.poll(console); err != nil {
				return err
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

func (e *engine) changeFramerate(d time.Duration, console *nes.Console) {
	e.ticker.Stop()
	e.ticker = time.NewTicker(d)
	e.tickerChan = e.ticker.C
	e.fpsMeter.Reset()
	e.mainView.SetFlashMsg(fmt.Sprintf("%d fps", int(1/d.Seconds())))
	// TODO: update apu sample rate
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

func (e *engine) poll(console *nes.Console) error {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch evt := event.(type) {
		case *sdl.QuitEvent:
			fmt.Println("quit")
			return errQuit

		case *sdl.DropEvent:
			if evt.File == "" {
				continue
			}

			cartridge, err := loadRom(evt.File)
			if err != nil {
				return err
			}
			console.Load(cartridge)

		case *sdl.ControllerDeviceEvent:
			e.onControllersChanged(evt)

		case *sdl.KeyboardEvent:
			if err := e.onKeyboarEvent(evt, console); err != nil {
				return err
			}
		case *sdl.WindowEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.TextEditingEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.TextInputEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.MouseMotionEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.MouseButtonEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.MouseWheelEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		case *sdl.UserEvent:
			e.viewsById[evt.WindowID].Handle(evt, console)
		default:
			for _, w := range e.viewsById {
				_, err := w.Handle(evt, console)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *engine) onControllersChanged(*sdl.ControllerDeviceEvent) {
	for _, ctrl := range e.controllers {
		ctrl.Close()
	}
	e.controllers = e.controllers[:0]

	for i := 0; i < sdl.NumJoysticks(); i++ {
		e.controllers = append(e.controllers, sdl.GameControllerOpen(i))
	}
}

func (e *engine) onKeyboarEvent(evt *sdl.KeyboardEvent, console *nes.Console) error {
	if entry, ok := speedTable[evt.Keysym.Sym]; evt.Type == sdl.KEYUP && ok {
		e.changeFramerate(entry, console)
		return nil
	}

	if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_SPACE {
		return e.pauseUnpause()
	}

	if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F1 {
		e.patternView.Toggle()
		return nil
	}

	if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F2 {
		e.nametableView.Toggle()
		return nil
	}

	_, err := e.viewsById[evt.WindowID].Handle(evt, console)
	return err
}

func (e *engine) render(console *nes.Console) error {
	paintStart := time.Now()
	for _, w := range e.viewsById {
		if !w.Visible() {
			continue
		}

		w.Update(console, e)

		if err := w.Render(); err != nil {
			return err
		}
	}
	e.paintMeter.Record(time.Since(paintStart))

	return nil
}
