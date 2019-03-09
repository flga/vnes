package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/ttf"

	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	fontSmall  *ttf.Font
	fontMedium *ttf.Font
	fontLarge  *ttf.Font
)

type window interface {
	handle(sdl.Event, *nes.Console) error
	render(c *nes.Console, meter *fpsMeter) error
	toggle()
	free() error
	visible() bool
}

const zoom = 4

func init() {
	runtime.LockOSThread()
}

func run(filename string, console *nes.Console) error {
	// init gl
	{
		if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
			return fmt.Errorf("unable to init sdl: %s", err)
		}
		defer sdl.Quit()
	}

	// init ttf
	{
		if err := ttf.Init(); err != nil {
			return fmt.Errorf("unable to init sdl ttf: %s", err)
		}
		defer ttf.Quit()

		var err error
		fontSmall, err = ttf.OpenFont("assets/runescape_uf.ttf", 16)
		if err != nil {
			return fmt.Errorf("unable to open small font: %s", err)
		}
		defer fontSmall.Close()

		fontMedium, err = ttf.OpenFont("assets/runescape_uf.ttf", 32)
		if err != nil {
			return fmt.Errorf("unable to open medium font: %s", err)
		}
		defer fontMedium.Close()

		fontLarge, err = ttf.OpenFont("assets/runescape_uf.ttf", 64)
		if err != nil {
			return fmt.Errorf("unable to open large font: %s", err)
		}
		defer fontLarge.Close()
	}

	gameView, err := newGameView(strings.TrimSuffix(path.Base(filename), path.Ext(filename)), zoom)
	if err != nil {
		return fmt.Errorf("unable to create game window: %s", err)
	}

	patternView, err := newPatternView(zoom)
	if err != nil {
		return fmt.Errorf("unable to create pattern window: %s", err)
	}

	nametableView, err := newNametableView(zoom / 2)
	if err != nil {
		return fmt.Errorf("unable to create nametable window: %s", err)
	}

	windows := map[uint32]window{
		gameView.id:      gameView,
		patternView.id:   patternView,
		nametableView.id: nametableView,
	}

	running := true
	paused := false
	fpsMeter := newFPSMeter(200)

	speedTable := [...]time.Duration{
		time.Second / 1,
		time.Second / 2,
		time.Second / 4,
		time.Second / 10,
		time.Second / 30,
		time.Second / 60,
		time.Second / 80,
		time.Second / 120,
		time.Second / 144,
	}
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()
	tickerChan := ticker.C

	quit := func() {
		running = false
		for _, w := range windows {
			w.free()
		}
	}

	var controllers []*sdl.GameController
	start := time.Now()
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
						ticker.Stop()
						ticker = time.NewTicker(speedTable[0])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("1 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_2 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[1])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("2 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_3 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[2])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("4 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_4 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[3])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("10 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_5 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[4])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("30 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_6 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[5])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("60 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_7 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[6])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("80 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_8 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[7])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("120 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_9 {
						ticker.Stop()
						ticker = time.NewTicker(speedTable[8])
						tickerChan = ticker.C
						fpsMeter.reset()
						gameView.setFlashMsg("144 fps")
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_SPACE {
						paused = !paused
						if paused {
							gameView.setStatusMsg("paused")
						} else {
							gameView.setStatusMsg("")
						}
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F1 {
						patternView.toggle()
					} else if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_F2 {
						nametableView.toggle()
					} else {
						windows[evt.WindowID].handle(evt, console)
					}
				case *sdl.WindowEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.TextEditingEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.TextInputEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.MouseMotionEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.MouseButtonEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.MouseWheelEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.DropEvent:
					windows[evt.WindowID].handle(evt, console)
				case *sdl.UserEvent:
					windows[evt.WindowID].handle(evt, console)
				default:
					for _, w := range windows {
						w.handle(evt, console)
					}
				}
			}
		}

		visible := false
		for _, w := range windows {
			if w.visible() {
				visible = true
			}
		}
		if !visible || !gameView.visible() {
			quit()
			break Main
		}

		select {
		case <-tickerChan:
			if !paused {
				console.StepFrame()
			}
			for _, w := range windows {
				if !w.visible() {
					continue
				}
				w.render(console, fpsMeter)
			}
			fpsMeter.record(time.Since(start))
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

	if err := run(os.Args[1], console); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
