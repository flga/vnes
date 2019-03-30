package main

import (
	"image/color"

	"github.com/veandco/go-sdl2/sdl"
)

var black = color.RGBA{0, 0, 0, 255}
var black128 = color.RGBA{0, 0, 0, 128}
var white = color.RGBA{255, 255, 255, 255}
var white64 = color.RGBA{255, 255, 255, 64}
var white128 = color.RGBA{255, 255, 255, 128}
var lightBlue = color.RGBA{0xb8, 0xe2, 0xe8, 255}

func rgbToSDL(c color.RGBA) sdl.Color {
	return sdl.Color{R: c.R, G: c.G, B: c.B, A: c.A}
}
