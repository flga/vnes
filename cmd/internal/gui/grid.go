package gui

import (
	"fmt"
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

var _ Component = &GridList{}
var _ Component = &Grid{}

type GridList struct {
	Tag      string
	Disabled bool
	List     []*Grid
}

func (g *GridList) tag() string {
	return g.Tag
}

func (g *GridList) Enabled() bool {
	return !g.Disabled
}

func (g *GridList) Enable() {
	g.Disabled = false
}

func (g *GridList) Disable() {
	g.Disabled = true
}

func (g *GridList) Toggle() {
	g.Disabled = !g.Disabled
}

func (g *GridList) Update(v *View) {
	if g.Disabled {
		return
	}

	for _, grid := range g.List {
		grid.Update(v)
	}
}

func (g *GridList) Draw(v *View) error {
	if g.Disabled {
		return nil
	}

	for _, grid := range g.List {
		if err := grid.Draw(v); err != nil {
			return err
		}
	}
	return nil
}

type Grid struct {
	UpdateFn func(g *Grid)

	Tag        string
	Disabled   bool
	Rows, Cols int32
	Square     bool
	Borders    bool
	Color      color.RGBA

	Bounds sdl.Rect
}

func (g *Grid) tag() string {
	return g.Tag
}

func (g *Grid) Enabled() bool {
	return !g.Disabled
}

func (g *Grid) Enable() {
	g.Disabled = false
}

func (g *Grid) Disable() {
	g.Disabled = true
}

func (g *Grid) Toggle() {
	g.Disabled = !g.Disabled
}

func (g *Grid) Update(*View) {
	if g.Disabled {
		return
	}

	if g.UpdateFn != nil {
		g.UpdateFn(g)
	}
}

func (g *Grid) Draw(v *View) error {
	if g.Disabled {
		return nil
	}

	if err := v.renderer.SetDrawColor(g.Color.R, g.Color.G, g.Color.B, g.Color.A); err != nil {
		return fmt.Errorf("grid.draw: unable to set draw color: %s", err)
	}

	var b int32
	if g.Borders {
		b = 1
	}

	vw := float32(g.Bounds.W) / float32(g.Cols)
	vx := float32(g.Bounds.X)
	for i := 1 - b; i < g.Cols+b; i++ {
		x0 := mini32(g.Bounds.X+g.Bounds.W, round32(vx+float32(i)*vw))
		x1 := mini32(g.Bounds.X+g.Bounds.W, x0)
		y0 := g.Bounds.Y
		y1 := y0 + g.Bounds.H

		if err := v.renderer.DrawLine(x0, y0, x1, y1); err != nil {
			return fmt.Errorf("grid.draw: unable to draw cols: %s", err)
		}
	}

	vh := float32(g.Bounds.H) / float32(g.Rows)
	vy := float32(g.Bounds.Y)
	if g.Square {
		vh = vw
	}
	for i := 1 - b; i < g.Rows+b; i++ {
		x0 := g.Bounds.X
		x1 := x0 + g.Bounds.W
		y0 := mini32(g.Bounds.Y+g.Bounds.H, round32(vy+float32(i)*vh))
		y1 := mini32(g.Bounds.Y+g.Bounds.H, y0)
		if err := v.renderer.DrawLine(x0, y0, x1, y1); err != nil {
			return fmt.Errorf("grid.draw: unable to draw rows: %s", err)
		}
	}

	return nil
}
