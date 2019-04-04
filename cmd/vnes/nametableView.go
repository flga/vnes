package main

import (
	"fmt"

	"github.com/flga/nes/cmd/internal/gui"
	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

type nametableView struct {
	*gui.View

	gridList gui.GridList
	bg       *gui.Background
}

func newNametableView(scale int, fontCache gui.FontMap) (*nametableView, error) {
	w, h := 256*2, 240*2

	view, err := gui.NewView("vnes - nametables", w, h, scale, sdl.WINDOW_HIDDEN|sdl.WINDOW_RESIZABLE, 0, sdl.BLENDMODE_BLEND, fontCache)
	if err != nil {
		return nil, fmt.Errorf("unable to create name table view: %s", err)
	}

	return &nametableView{
		View: view,
	}, nil
}

func (v *nametableView) Init(engine *engine, console *nes.Console) error {
	v.gridList = gui.GridList{
		List: []*gui.Grid{
			&gui.Grid{Rows: 60, Cols: 64, Color: white64, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
			&gui.Grid{Rows: 8, Cols: 16, Square: true, Color: white128, UpdateFn: func(g *gui.Grid) {
				rect := v.Rect()
				g.Bounds = sdl.Rect{W: rect.W, H: rect.H / 2, X: rect.X, Y: rect.Y}
			}},
			&gui.Grid{Rows: 8, Cols: 16, Square: true, Color: white128, UpdateFn: func(g *gui.Grid) {
				rect := v.Rect()
				g.Bounds = sdl.Rect{W: rect.W, H: rect.H / 2, X: rect.X, Y: rect.H / 2}
			}},
			&gui.Grid{Rows: 2, Cols: 2, Borders: true, Color: white, UpdateFn: func(g *gui.Grid) { g.Bounds = v.Rect() }},
		},
	}

	v.bg = &gui.Background{
		UpdateFn: func(r *gui.Background) {
			if len(r.RGBA8888) < 256*240*4*4 {
				r.RGBA8888 = make([]byte, 256*240*4*4)
			}
			console.DrawNametables(r.RGBA8888)
		},
	}

	return nil
}

func (v *nametableView) Handle(event sdl.Event, engine *engine, console *nes.Console) (handled bool, err error) {
	if handled, err := v.View.Handle(event); handled || err != nil {
		return handled, err
	}

	if !v.Focused() {
		return false, nil
	}

	if gui.IsKeyPress(event, sdl.K_g) {
		v.gridList.Toggle()
		return true, nil
	}

	return false, nil
}

func (v *nametableView) Update(console *nes.Console, engine *engine) {
	v.bg.Update(v.View)
	v.gridList.Update(v.View)
}

func (v *nametableView) Render() error {
	if !v.Visible() {
		return nil
	}

	if err := v.Clear(black); err != nil {
		return v.Errorf("unable to clear view: %s", err)
	}

	if err := v.bg.Draw(v.View); err != nil {
		return v.Errorf("unable to draw nametables: %s", err)
	}

	if err := v.gridList.Draw(v.View); err != nil {
		return v.Errorf("unable to draw grid: %s", err)
	}

	return nil
}
