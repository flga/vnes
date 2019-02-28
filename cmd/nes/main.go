package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/veandco/go-sdl2/ttf"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

const zoom = 4

func init() {
	runtime.LockOSThread()
}

type window interface {
	Handle(sdl.Event, *nes.Console) error
	Render(*nes.Console, time.Duration) error
	Toggle()
	Free() error
	Visible() bool
}

func run(console *nes.Console) error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return fmt.Errorf("unable to init sdl: %s", err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		return fmt.Errorf("unable to init sdl ttf: %s", err)
	}

	gameWin, gameId, err := newGameWindow(zoom, "foobar")
	if err != nil {
		return fmt.Errorf("unable to create game window: %s", err)
	}

	patternWin, patternId, err := newPatternWindow(zoom)
	if err != nil {
		return fmt.Errorf("unable to create pattern window: %s", err)
	}

	nametableWin, nametableId, err := newNametableWindow(zoom)
	if err != nil {
		return fmt.Errorf("unable to create nametable window: %s", err)
	}

	windows := map[uint32]window{
		gameId:      gameWin,
		patternId:   patternWin,
		nametableId: nametableWin,
	}

	running := true

	speedTable := [...]time.Duration{16 * time.Millisecond, 200 * time.Microsecond, 1000 * time.Microsecond, 200000 * time.Microsecond, 1 * time.Second}
	ticker := time.NewTicker(speedTable[0])
	defer ticker.Stop()

	tickerChan := ticker.C

	quit := func() {
		running = false
		for _, w := range windows {
			w.Free()
		}
	}

	paused := false
	var controllers []*sdl.GameController

Main:
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if event != nil {
				switch evt := event.(type) {
				case *sdl.ControllerDeviceEvent:
					for _, ctrl := range controllers {
						ctrl.Close()
					}
					controllers = controllers[:0]

					for i := 0; i < sdl.NumJoysticks(); i++ {
						controllers = append(controllers, sdl.GameControllerOpen(i))
					}
				case *sdl.QuitEvent:
					quit()
					break Main
				case *sdl.KeyboardEvent:
					if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_1 {
						ticker = time.NewTicker(speedTable[0])
						tickerChan = ticker.C
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_2 {
						ticker = time.NewTicker(speedTable[1])
						tickerChan = ticker.C
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_3 {
						ticker = time.NewTicker(speedTable[2])
						tickerChan = ticker.C
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_4 {
						ticker = time.NewTicker(speedTable[3])
						tickerChan = ticker.C
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_5 {
						ticker = time.NewTicker(speedTable[4])
						tickerChan = ticker.C
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_SPACE {
						paused = !paused
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F1 {
						patternWin.Toggle()
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F2 {
						nametableWin.Toggle()
					} else {
						windows[evt.WindowID].Handle(evt, console)
					}
				case *sdl.WindowEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.TextEditingEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.TextInputEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.MouseMotionEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.MouseButtonEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.MouseWheelEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.DropEvent:
					windows[evt.WindowID].Handle(evt, console)
				case *sdl.UserEvent:
					windows[evt.WindowID].Handle(evt, console)
				default:
					for _, w := range windows {
						w.Handle(evt, console)
					}
				}
			}
		}

		visible := false
		for _, w := range windows {
			if w.Visible() {
				visible = true
			}
		}
		if !visible || !gameWin.Visible() {
			quit()
			break Main
		}

		start := time.Now()
		select {
		case <-tickerChan:
			if !paused {
				console.StepFrame()
			}
			dur := time.Since(start)
			for _, w := range windows {
				if !w.Visible() {
					continue
				}
				w.Render(console, dur)
			}
			start = time.Now()
		}

	}

	return nil
}

func main() {
	testRom, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("unable to open rom: %s", err))
	}
	cartridge, err := nes.LoadINES(testRom)

	if cartridge.Mapper != 0 {
		panic(fmt.Sprintf("Unexpected mapper %d\n", cartridge.Mapper))
	}

	var out io.Writer
	if len(os.Args) > 2 && os.Args[2] == "--trace" {
		out = os.Stdout
	}

	console := nes.NewConsole(cartridge, 0, out)

	if err := run(console); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func resize(window *sdl.Window, minWidth, minHeight float64, surface *sdl.Rect) {
	ww, hh := window.GetSize()
	width := float64(ww)
	height := float64(hh)
	var x, y float64 = 0, 0

	ow, oh := width, height
	height = math.Floor(width * (minHeight / minWidth))
	if height > oh {
		width = math.Floor(oh * (minWidth / minHeight))
		height = math.Floor(width * (minHeight / minWidth))
	}

	if width > ow {
		x = (width - ow) / 2
	} else {
		x = (ow - width) / 2
	}
	if height > oh {
		y = (height - oh) / 2
	} else {
		y = (oh - height) / 2
	}

	surface.W = int32(width)
	surface.H = int32(height)
	surface.X = int32(x)
	surface.Y = int32(y)
}
