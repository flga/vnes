package main

import (
	"fmt"
	"time"

	"github.com/flga/nes/cmd/vnes/internal/gui"
	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type patternView struct {
	*gui.View

	gridList   gui.GridList
	bg         *gui.Background
	status     *gui.Status
	paletteNum byte
}

func newPatternView(scale int, fontCache gui.FontMap) (*patternView, error) {
	w, h := 256, 128

	view, err := gui.NewView("vnes - pattern tables", w, h, scale, sdl.WINDOW_HIDDEN|sdl.WINDOW_RESIZABLE, 0, sdl.BLENDMODE_BLEND, fontCache)
	if err != nil {
		return nil, fmt.Errorf("unable to create pattern table view: %s", err)
	}

	return &patternView{
		View: view,
	}, nil
}

func (v *patternView) Init(engine *engine, console *nes.Console) error {
	v.gridList = gui.GridList{
		List: []*gui.Grid{
			&gui.Grid{Rows: 16 * 8, Cols: 32 * 8, Color: white64, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
			&gui.Grid{Rows: 16, Cols: 32, Color: white128, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
			&gui.Grid{Rows: 1, Cols: 2, Borders: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = *v.Rect }},
		},
	}

	font, ok := v.FontMap["RuneScape UF"]
	if !ok {
		return fmt.Errorf("font %q not found", "RuneScape UF")
	}

	v.status = &gui.Status{
		Message: &gui.Message{
			Font:       font,
			Size:       64,
			Padding:    gui.Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
			Position:   gui.CenterCenter,
			Foreground: white,
			Background: black128,
		},
	}

	v.bg = &gui.Background{
		UpdateFn: func(r *gui.Background) {
			if len(r.RGBA8888) < 128*128*2*4 {
				r.RGBA8888 = make([]byte, 128*128*2*4)
			}
			console.PPU.DrawPatternTables(r.RGBA8888, v.paletteNum)
		},
	}

	return nil
}

func (v *patternView) SetFlashMsg(m string) {
	v.status.SetFlashMsg(m, 2*time.Second)
}

func (v *patternView) SetStatusMsg(m string) {
	v.status.SetStatusMsg(m)
}

func (v *patternView) Handle(event sdl.Event, console *nes.Console) (handled bool, err error) {
	if handled, err := v.View.Handle(event, console); handled || err != nil {
		return handled, err
	}

	switch evt := event.(type) {
	case *sdl.KeyboardEvent:
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_g {
			v.gridList.Toggle()
			return true, nil
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_UP {
			if v.paletteNum == 7 {
				v.paletteNum = 0
			} else {
				v.paletteNum++
			}
			v.SetFlashMsg(fmt.Sprintf("palette %d", v.paletteNum))
			return true, nil
		}
		if evt.Type == sdl.KEYUP && evt.Keysym.Sym == sdl.K_DOWN {
			if v.paletteNum == 0 {
				v.paletteNum = 7
			} else {
				v.paletteNum--
			}
			v.SetFlashMsg(fmt.Sprintf("palette %d", v.paletteNum))
			return true, nil
		}
	}

	return false, nil
}

func (v *patternView) Update(console *nes.Console, engine *engine) {
	v.bg.Update(v.View)
	v.gridList.Update(v.View)
}

func (v *patternView) Render() error {
	if !v.Visible() {
		return nil
	}

	if err := v.Clear(black); err != nil {
		return v.Errorf("unable to clear view: %s", err)
	}

	if err := v.bg.Draw(v.View); err != nil {
		return v.Errorf("unable to draw pattern tables: %s", err)
	}

	// if err := v.drawStatus(font); err != nil {
	// 	return v.Errorf("unable to draw status: %s", err)
	// }

	if err := v.gridList.Draw(v.View); err != nil {
		return v.Errorf("unable to draw grid: %s", err)
	}

	v.Paint()
	return nil
}
