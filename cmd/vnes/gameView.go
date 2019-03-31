package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/flga/nes/cmd/internal/gui"
	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

var controllerMapping = map[uint8]nes.Button{
	sdl.CONTROLLER_BUTTON_A:          nes.A,
	sdl.CONTROLLER_BUTTON_B:          nes.B,
	sdl.CONTROLLER_BUTTON_START:      nes.Start,
	sdl.CONTROLLER_BUTTON_BACK:       nes.Select,
	sdl.CONTROLLER_BUTTON_DPAD_UP:    nes.Up,
	sdl.CONTROLLER_BUTTON_DPAD_DOWN:  nes.Down,
	sdl.CONTROLLER_BUTTON_DPAD_LEFT:  nes.Left,
	sdl.CONTROLLER_BUTTON_DPAD_RIGHT: nes.Right,
}

var keyboardMapping = map[sdl.Keycode]nes.Button{
	sdl.K_a:      nes.A,
	sdl.K_z:      nes.B,
	sdl.K_RETURN: nes.Start,
	sdl.K_RSHIFT: nes.Select,
	sdl.K_UP:     nes.Up,
	sdl.K_DOWN:   nes.Down,
	sdl.K_LEFT:   nes.Left,
	sdl.K_RIGHT:  nes.Right,
}

type gameView struct {
	*gui.View

	// game           *gui.Image
	// fps            *gui.Message
	// status         *gui.Status
	// gameMenu       *gui.Menu
	// gridList       gui.GridList
	layers         gui.Layers
	recording      bool
	pauseRecording bool
}

func newGameView(title string, scale int, fontMap gui.FontMap) (*gameView, error) {
	var w, h = 256, 240

	view, err := gui.NewView(
		title,
		w,
		h,
		scale,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
		sdl.RENDERER_PRESENTVSYNC|sdl.RENDERER_ACCELERATED,
		sdl.BLENDMODE_BLEND,
		fontMap,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create game view: %s", err)
	}

	v := &gameView{View: view}

	return v, nil
}

func (v *gameView) Init(engine *engine, console *nes.Console) error {
	font, ok := v.FontMap["RuneScape UF"]
	if !ok {
		return fmt.Errorf("font %q not found", "RuneScape UF")
	}

	v.layers = v.layers.New(
		&gui.Background{
			Tag:      "background",
			UpdateFn: func(r *gui.Background) { r.RGBA8888 = console.Buffer() },
		},
		&gui.Message{
			Tag:      "screensaver",
			Disabled: false,
			UpdateFn: func(m *gui.Message) {
				if console.Empty() {
					m.Text = "DROP A ROM"
				} else {
					m.Text = ""
				}
			},
			Font:       font,
			Size:       64,
			Position:   gui.CenterCenter,
			Foreground: white,
			Background: black,
		},
	)

	v.layers = v.layers.New(
		&gui.GridList{
			Tag:      "grid",
			Disabled: true,
			List: []*gui.Grid{
				&gui.Grid{Rows: 240, Cols: 256, Color: white64, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
				&gui.Grid{Rows: 30, Cols: 32, Color: white128, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
				&gui.Grid{Rows: 8, Cols: 8, Square: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
				&gui.Grid{Rows: 1, Cols: 1, Borders: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
			},
		},
		&gui.Message{
			Tag:      "fps",
			Disabled: false,
			UpdateFn: func(m *gui.Message) {
				m.Text = fmt.Sprintf(
					"%.fms - %.fms, %d fps",
					engine.paintMeter.Ms(),
					engine.consoleMeter.Ms(),
					engine.fpsMeter.Tps(),
				)
			},
			Font:       font,
			Size:       16,
			Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
			Margin:     gui.Margin{Top: 10, Right: 10},
			Position:   gui.TopRight,
			Foreground: white,
			Background: black128,
		},
		&gui.Status{
			Tag: "status",
			Message: &gui.Message{
				Font:       font,
				Size:       64,
				Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
				Position:   gui.CenterCenter,
				Foreground: white,
				Background: black128,
			},
		},
	)

	v.layers = v.layers.New(
		&gui.Menu{
			Tag:        "menu",
			Disabled:   true,
			Position:   gui.CenterCenter,
			Margin:     gui.Margin{Top: 30, Right: 30, Bottom: 30, Left: 30},
			Background: black,
			Backdrop:   black128,
			Items: []gui.MenuItem{
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "Fullscreen",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 0, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						UpdateFn: func() string { return boolToStr(v.Fullscreen()) },
						Font:     font,
						Size:     32,
						Padding:  gui.Padding{Top: 0, Right: 0, Bottom: 5, Left: 15},
						Color:    white,
						Hover:    lightBlue,
					},
					Callback: func() error { return v.ToggleFullscreen() },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "VSync",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						UpdateFn: func() string { return boolToStr(v.VSync()) },
						Font:     font,
						Size:     32,
						Padding:  gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:    white,
						Hover:    lightBlue,
					},
					Callback: func() error { return v.ToggleVSync() },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "Volume",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "3",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Volume"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "Filter Output",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "no",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Filter Output"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "Channels",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Channels"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "    Mute Pulse 1",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "yes",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Mute Pulse 1"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "    Mute Pulse 2",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "yes",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Mute Pulse 2"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "    Mute Triangle",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "no",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Mute Triangle"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "    Mute Noise",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "no",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Mute Noise"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						Text:    "    Mute DMC",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "yes",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Mute DMC"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						UpdateFn: func() string {
							if v.recording {
								return "Stop Recording"
							} else {
								return "Start Recording"
							}
						},
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 5, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 5, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Stop Recording"); return nil },
				},
				gui.MenuItem{
					Label: gui.Cell{
						UpdateFn: func() string {
							if !v.recording {
								return ""
							}
							if v.pauseRecording {
								return "Unpause Recording"
							} else {
								return "Pause Recording"
							}
						},
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 15, Bottom: 0, Left: 0},
						Color:   white,
						Hover:   lightBlue,
					},
					Value: gui.Cell{
						Text:    "",
						Font:    font,
						Size:    32,
						Padding: gui.Padding{Top: 5, Right: 0, Bottom: 0, Left: 15},
						Color:   white,
						Hover:   lightBlue,
					},
					Callback: func() error { fmt.Println("Pause Recording"); return nil },
				},
			},
		},
	)

	return nil
}

func (v *gameView) SetFlashMsg(m string) {
	if status, ok := v.layers.Find("status").(*gui.Status); ok {
		status.SetFlashMsg(m, 2*time.Second)
	}
}

func (v *gameView) SetStatusMsg(m string) {
	if status, ok := v.layers.Find("status").(*gui.Status); ok {
		status.SetStatusMsg(m)
	}
}

func (v *gameView) Handle(event sdl.Event, console *nes.Console) (handled bool, err error) {
	if handled, err := v.View.Handle(event, console); handled || err != nil {
		return handled, err
	}

	press := func(b nes.Button, pressed bool) {
		fmt.Println(b, pressed)
		if pressed {
			console.Controller1.Press(b)
		} else {
			console.Controller1.Release(b)
		}
	}

	switch evt := event.(type) {

	case *sdl.ControllerButtonEvent:
		if evt.Button == sdl.CONTROLLER_BUTTON_Y && evt.Type == sdl.CONTROLLERBUTTONDOWN {
			fmt.Println("toggle menu (ctrl)")
			v.layers.Find("menu").Toggle()
			return true, nil
		}

		if v.layers.Find("menu").Enabled() {
			if evt.Type != sdl.CONTROLLERBUTTONDOWN {
				return false, nil
			}

			menu, ok := v.layers.Find("menu").(*gui.Menu)
			if !ok {
				return false, errors.New("menu not found")
			}

			switch evt.Button {
			case sdl.CONTROLLER_BUTTON_A:
				fmt.Println("Activate (ctrl)")
				return true, menu.Activate()
			case sdl.CONTROLLER_BUTTON_DPAD_UP:
				fmt.Println("Up (ctrl)")
				menu.Up()
			case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
				fmt.Println("Down (ctrl)")
				menu.Down()
			}

			return true, nil
		}

		if btn, ok := controllerMapping[evt.Button]; ok {
			fmt.Println("press (ctrl)")
			press(btn, evt.Type == sdl.CONTROLLERBUTTONDOWN)
			return true, nil
		}

		if evt.Button == sdl.CONTROLLER_BUTTON_X && evt.Type == sdl.CONTROLLERBUTTONDOWN {
			fmt.Println("reset (ctrl)")
			console.Reset()
			return true, nil
		}

	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_ESCAPE {
			fmt.Println("toggle menu (kb)")
			v.layers.Find("menu").Toggle()
			return true, nil
		}

		if v.layers.Find("menu").Enabled() {
			if evt.Type != sdl.KEYUP {
				return false, nil
			}

			menu, ok := v.layers.Find("menu").(*gui.Menu)
			if !ok {
				return false, errors.New("menu not found")
			}

			switch evt.Keysym.Sym {
			case sdl.K_LEFT:
				fmt.Println("Activate (kb)")
				return true, menu.Activate()
			case sdl.K_RIGHT:
				fmt.Println("Activate (kb)")
				return true, menu.Activate()
			case sdl.K_UP:
				fmt.Println("Up (kb)")
				menu.Up()
			case sdl.K_DOWN:
				fmt.Println("Down (kb)")
				menu.Down()
			}

			return true, nil
		}

		if btn, ok := keyboardMapping[evt.Keysym.Sym]; ok {
			fmt.Println("press (kb)")
			press(btn, evt.Type == sdl.KEYDOWN)
			return true, nil
		}

		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			fmt.Println("toggle grid (kb)")
			v.layers.Find("grid").Toggle()
			return true, nil
		}

		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_r {
			fmt.Println("reset (kb)")
			console.Reset()
			return true, nil
		}

		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_o {
			alt := evt.Keysym.Mod == sdl.KMOD_ALT ||
				evt.Keysym.Mod == sdl.KMOD_LALT ||
				evt.Keysym.Mod == sdl.KMOD_RALT
			if v.recording && alt {
				v.pauseRecording = !v.pauseRecording
				if v.pauseRecording {
					fmt.Println("pause (kb)")
					console.PauseRecording()
				} else {
					fmt.Println("unpause (kb)")
					console.UnpauseRecording()
				}
			} else {
				v.recording = !v.recording
				if v.recording {
					fmt.Println("start rec (kb)")
					if err := console.StartRecording(); err != nil {
						return true, err
					}
				} else {
					fmt.Println("stop rec (kb)")
					console.StopRecording()
				}
			}
			return true, nil
		}
	}

	return false, nil
}

func (v *gameView) Update(console *nes.Console, engine *engine) {
	v.layers.Update(v.View)
}

func (v *gameView) Render() error {
	if !v.Visible() {
		return nil
	}

	if err := v.Clear(black); err != nil {
		return v.Errorf("render: unable to clear view: %s", err)
	}

	if err := v.layers.Draw(v.View); err != nil {
		return v.Errorf("render: unable to draw: %s", err)
	}

	v.Paint()
	return nil
}

func boolToStr(v bool) string {
	if v {
		return "yes"
	}

	return "no"
}
