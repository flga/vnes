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

var keyboardMapping = map[sdl.Keycode]struct {
	ctrl int
	btn  nes.Button
}{
	sdl.K_RETURN: {ctrl: 0, btn: nes.Start},
	sdl.K_z:      {ctrl: 0, btn: nes.Select},
	sdl.K_RSHIFT: {ctrl: 0, btn: nes.A},
	sdl.K_RCTRL:  {ctrl: 0, btn: nes.B},
	sdl.K_UP:     {ctrl: 0, btn: nes.Up},
	sdl.K_DOWN:   {ctrl: 0, btn: nes.Down},
	sdl.K_LEFT:   {ctrl: 0, btn: nes.Left},
	sdl.K_RIGHT:  {ctrl: 0, btn: nes.Right},

	sdl.K_v: {ctrl: 1, btn: nes.A},
	sdl.K_b: {ctrl: 1, btn: nes.B},
	sdl.K_w: {ctrl: 1, btn: nes.Up},
	sdl.K_s: {ctrl: 1, btn: nes.Down},
	sdl.K_a: {ctrl: 1, btn: nes.Left},
	sdl.K_d: {ctrl: 1, btn: nes.Right},
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
	font, ok := v.Font("RuneScape UF")
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
					m.Text = "DROP\nA\nROM"
				} else {
					m.Text = ""
				}
			},
			Font:       font,
			Size:       64,
			Align:      gui.Center,
			Position:   gui.Middle | gui.Center,
			Foreground: white,
			Background: black,
		},
	)

	v.layers = v.layers.New(
		&gui.GridList{
			Tag:      "grid",
			Disabled: true,
			List: []*gui.Grid{
				&gui.Grid{Rows: 240, Cols: 256, Color: white64, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
				&gui.Grid{Rows: 30, Cols: 32, Color: white128, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
				&gui.Grid{Rows: 8, Cols: 8, Square: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
				&gui.Grid{Rows: 1, Cols: 1, Borders: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
			},
		},
		&gui.Message{
			Tag:      "info",
			Disabled: true,
			UpdateFn: func(m *gui.Message) {
				renderer, wm := v.Info()
				m.Text = fmt.Sprintf(`Graphics
renderer: %s
sdl version: %d.%d.%d
vsync: on

Audio
audio device: %s
audio api: %s
sample rate: %.f
frames per buffer: %d
channels: %d
latency: %v

State
paused: %v
recording: %v
recording paused: %v

Timings
update: %.fms
render: %.fms
paint: %.fms
poll: %.fms
console: %.fms`,
					renderer.Name,
					wm.Version.Major,
					wm.Version.Minor,
					wm.Version.Patch,
					engine.audio.streamParams.Output.Device.Name,
					engine.audio.streamParams.Output.Device.HostApi.Name,
					engine.audio.streamParams.SampleRate,
					engine.audio.streamParams.FramesPerBuffer,
					engine.audio.streamParams.Output.Channels,
					engine.audio.streamParams.Output.Latency,
					engine.paused,
					v.recording,
					v.pauseRecording,
					engine.updateMeter.Ms(),
					engine.renderMeter.Ms(),
					engine.paintMeter.Ms(),
					engine.pollMeter.Ms(),
					engine.consoleMeter.Ms(),
				)
			},
			Font:       font,
			Size:       32,
			Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
			Margin:     gui.Margin{Top: 10, Left: 10},
			Position:   gui.Top | gui.Left,
			Foreground: white,
			Background: black128,
		},
		&gui.Message{
			Tag:      "fps",
			Disabled: false,
			UpdateFn: func(m *gui.Message) {
				m.Text = fmt.Sprintf("%d fps", engine.fpsMeter.Tps())
			},
			Font:       font,
			Size:       16,
			Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
			Margin:     gui.Margin{Top: 10, Right: 10},
			Position:   gui.Top | gui.Right,
			Foreground: white,
			Background: black128,
		},
		&gui.Status{
			Tag: "status",
			Message: &gui.Message{
				Font:       font,
				Size:       64,
				Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
				Position:   gui.Middle | gui.Center,
				Foreground: white,
				Background: black128,
			},
		},
	)

	v.layers = v.layers.New(
		&gui.Menu{
			Tag:        "menu",
			Disabled:   true,
			Position:   gui.Middle | gui.Center,
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
					// Callback: func() error { fmt.Println("Volume"); return nil },
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
					// Callback: func() error { fmt.Println("Filter Output"); return nil },
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
					// Callback: func() error { fmt.Println("Channels"); return nil },
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
					// Callback: func() error { fmt.Println("Mute Pulse 1"); return nil },
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
					// Callback: func() error { fmt.Println("Mute Pulse 2"); return nil },
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
					// Callback: func() error { fmt.Println("Mute Triangle"); return nil },
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
					// Callback: func() error { fmt.Println("Mute Noise"); return nil },
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
					// Callback: func() error { fmt.Println("Mute DMC"); return nil },
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
					// Callback: func() error { fmt.Println("Stop Recording"); return nil },
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
					// Callback: func() error { fmt.Println("Pause Recording"); return nil },
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

func (v *gameView) Handle(evt sdl.Event, engine *engine, console *nes.Console) (handled bool, err error) {
	if handled, err := v.View.Handle(evt); handled || err != nil {
		return handled, err
	}

	if evt, ok := gui.IsDropEvent(evt, sdl.DROPFILE, v.ID()); ok {
		return true, console.LoadPath(evt.File)
	}

	if !v.Focused() {
		return false, nil
	}

	if handled, err := v.handleGuiEvts(evt, engine); handled || err != nil {
		return handled, err
	}
	if handled, err := v.handleMediaEvts(evt, console); handled || err != nil {
		return handled, err
	}
	if handled, err := v.handleConsoleEvts(evt, engine, console); handled || err != nil {
		return handled, err
	}

	return false, nil
}

func (v *gameView) handleGuiEvts(evt sdl.Event, engine *engine) (bool, error) {
	menu, ok := v.layers.Find("menu").(*gui.Menu)
	if !ok {
		return false, errors.New("unable to find menu component")
	}

	if gui.IsButtonPress(evt, sdl.CONTROLLER_BUTTON_Y) || gui.IsKeyPress(evt, sdl.K_ESCAPE) {
		menu.Toggle()
		return true, engine.pauseUnpause()
	}

	if menu.Enabled() {
		if gui.IsButtonPress(evt, sdl.CONTROLLER_BUTTON_A) || gui.IsKeyPress(evt, sdl.K_RETURN) {
			return true, menu.Activate()
		}
		if gui.IsButtonPress(evt, sdl.CONTROLLER_BUTTON_DPAD_UP) || gui.IsKeyPress(evt, sdl.K_UP) {
			menu.Up()
			return true, nil
		}
		if gui.IsButtonPress(evt, sdl.CONTROLLER_BUTTON_DPAD_DOWN) || gui.IsKeyPress(evt, sdl.K_DOWN) {
			menu.Down()
			return true, nil
		}
	}

	if gui.IsKeyPress(evt, sdl.K_g) {
		v.layers.Find("grid").Toggle()
		return true, nil
	}

	if gui.IsKeyPress(evt, sdl.K_h) {
		v.layers.Find("info").Toggle()
		return true, nil
	}

	return false, nil
}

func (v *gameView) handleMediaEvts(evt sdl.Event, console *nes.Console) (bool, error) {
	if gui.IsKeyPress(evt, sdl.K_o) {
		v.recording = !v.recording
		v.pauseRecording = false
		if v.recording {
			return true, console.StartRecording()
		}

		return true, console.StopRecording()
	}

	if gui.IsKeyPress(evt, sdl.K_o, sdl.KMOD_SHIFT) {
		if !v.recording {
			return true, nil
		}
		v.pauseRecording = !v.pauseRecording
		if v.pauseRecording {
			console.PauseRecording()
		} else {
			console.UnpauseRecording()
		}

		return true, nil
	}

	return false, nil
}

func (v *gameView) handleConsoleEvts(evt sdl.Event, engine *engine, console *nes.Console) (bool, error) {
	press := func(ctrl int, b nes.Button, pressed bool) {
		if pressed {
			console.Press(ctrl, b)
		} else {
			console.Release(ctrl, b)
		}
	}

	if gui.IsButtonPress(evt, sdl.CONTROLLER_BUTTON_X) || gui.IsKeyPress(evt, sdl.K_r) {
		console.Reset()
		return true, nil
	}

	switch evt := evt.(type) {
	case *sdl.ControllerButtonEvent:
		if btn, ok := controllerMapping[evt.Button]; ok {
			press(engine.controllers.which(evt.Which), btn, evt.Type == sdl.CONTROLLERBUTTONDOWN)
			return true, nil
		}

	case *sdl.KeyboardEvent:
		if entry, ok := keyboardMapping[evt.Keysym.Sym]; ok {
			press(entry.ctrl, entry.btn, evt.Type == sdl.KEYDOWN)
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

	return nil
}

func boolToStr(v bool) string {
	if v {
		return "yes"
	}

	return "no"
}
