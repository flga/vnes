package gui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/flga/nes/cmd/internal/errors"
	"github.com/veandco/go-sdl2/sdl"
)

type Renderer struct {
	*sdl.Renderer
	title      string
	background *sdl.Texture

	fontTextures map[string][]*sdl.Texture
}

func newRenderer(window *sdl.Window, w, h int32, options uint32) (*Renderer, error) {
	renderer, err := sdl.CreateRenderer(window, -1, options)
	if err != nil {
		return nil, fmt.Errorf("unable to create sdl renderer: %s", err)
	}

	bgTexture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, w, h)
	if err != nil {
		return nil, fmt.Errorf("unable to create background texture: %s", err)
	}

	return &Renderer{
		Renderer:     renderer,
		background:   bgTexture,
		fontTextures: make(map[string][]*sdl.Texture),
	}, nil
}

func (r *Renderer) Destroy() error {
	var ee errors.List
	for _, tt := range r.fontTextures {
		for _, t := range tt {
			ee = ee.Add(t.Destroy())
		}
	}
	return ee.Add(r.background.Destroy(), r.Renderer.Destroy())
}

func (r *Renderer) getFontTexture(font *Font, page int) (*sdl.Texture, error) {
	if _, ok := r.fontTextures[font.face]; !ok {
		r.fontTextures[font.face] = make([]*sdl.Texture, len(font.pages))
	}

	// cache lookup
	tex := r.fontTextures[font.face][page]
	if tex != nil {
		return tex, nil
	}

	// cache miss
	img := font.pages[page]
	bounds := img.Bounds()
	w, h := int32(bounds.Dx()), int32(bounds.Dy())

	tex, err := r.Renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, w, h)
	if err != nil {
		return nil, fmt.Errorf("font: unable to create texture for page %d of font %s: %s", page, font.face, err)
	}

	if err := tex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, fmt.Errorf("font: unable to set blend mode of texture for page %d of font %s: %s", page, font.face, err)
	}

	if err := tex.Update(nil, img.Pix, int(w)*4); err != nil {
		return nil, fmt.Errorf("font: unable to populate texture for page %d of font %s: %s", page, font.face, err)
	}

	// cache fill
	r.fontTextures[font.face][page] = tex

	return tex, nil
}

func (r *Renderer) DrawBackground(rgba8888 []byte, rect *sdl.Rect) error {
	pixels, _, err := r.background.Lock(nil)
	if err != nil {
		return fmt.Errorf("unable to lock background texture: %s", err)
	}

	copy(pixels, rgba8888)
	r.background.Unlock()

	if err := r.Copy(r.background, nil, rect); err != nil {
		return fmt.Errorf("unable to copy background texture: %s", err)
	}

	return nil
}

type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

func (r *Renderer) DrawText(s string, font *Font, size int, align TextAlign, color color.Color, pos *sdl.Rect) (int32, int32, error) {
	if s == "" {
		return 0, 0, nil
	}

	type renderOp struct {
		page int // only used for error ctx

		tex        *sdl.Texture
		src        *sdl.Rect
		w, h, x, y int32
	}

	var (
		width, curx, cury int32

		ratio      = int32(size / font.size)
		lineHeight = font.lineHeight * ratio

		lines     = strings.Split(s, "\n")
		numLines  = int32(len(lines))
		renderOps = make([][]renderOp, len(lines))
	)

	for i := int32(0); i < numLines; i++ {
		line := lines[i]
		lastChar := len(line) - 1

		curx = pos.X
		cury = pos.Y + i*lineHeight
		renderOps[i] = make([]renderOp, len(line))

		for j, char := range line {
			meta, ok := font.chars[char]
			if !ok {
				return 0, 0, ErrMissingChar{Face: font.face, Char: char}
			}

			tex, err := r.getFontTexture(font, meta.page)
			if err != nil {
				return 0, 0, err
			}

			op := renderOp{
				page: meta.page,
				tex:  tex,
				src: &sdl.Rect{
					W: meta.width,
					H: meta.height,
					X: meta.x,
					Y: meta.y,
				},
				w: meta.width * ratio,
				h: meta.height * ratio,
				x: curx + meta.xOffset*ratio,
				y: cury + meta.yOffset*ratio,
			}
			renderOps[i][j] = op

			var x int32
			if j == lastChar {
				x = op.x + op.w
			} else {
				curx += meta.xAdvance * ratio
				x = curx
			}
			x -= pos.X

			if x > width {
				width = x
			}
		}
	}

	for _, opLine := range renderOps {
		if len(opLine) == 0 {
			continue
		}

		var offsetx int32
		lastChar := opLine[len(opLine)-1]

		switch align {
		case TextAlignRight:
			offsetx = width - (lastChar.x + lastChar.w - pos.X)
		case TextAlignCenter:
			offsetx = (width - (lastChar.x + lastChar.w - pos.X)) / 2
		}

		for _, op := range opLine {
			if err := op.tex.SetColorMod(colorMod(color)); err != nil {
				return 0, 0, fmt.Errorf("font: unable to set color mod of texture of page %d of font %s: %s", op.page, font.face, err)
			}

			if err := r.Renderer.Copy(
				op.tex,
				op.src,
				&sdl.Rect{
					W: op.w,
					H: op.h,
					X: op.x + offsetx,
					Y: op.y,
				},
			); err != nil {
				return 0, 0, fmt.Errorf("font: unable to render texture of page %d of font %s: %s", op.page, font.face, err)
			}
		}
	}

	// // print pre-calculated bounding box
	// r.Renderer.SetDrawColor(0, 255, 0, 255)
	// r.Renderer.DrawRect(pos)

	// // print drawn bounding box
	// r.Renderer.SetDrawColor(255, 0, 0, 255)
	// r.Renderer.DrawRect(&sdl.Rect{
	// 	X: pos.X,
	// 	Y: pos.Y,
	// 	W: width,
	// 	H: lineHeight + cury - pos.Y,
	// })

	return width, lineHeight + cury - pos.Y, nil
}

func colorMod(color color.Color) (byte, byte, byte) {
	r, g, b, _ := color.RGBA()
	return byte(r), byte(g), byte(b)
}
