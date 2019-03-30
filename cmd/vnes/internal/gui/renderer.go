package gui

import (
	"fmt"
	"image/color"

	"github.com/flga/nes/cmd/vnes/internal/errors"
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

func (r *Renderer) DrawText(s string, font *Font, size int, color color.Color, pos *sdl.Rect) (w, h int32, e error) {
	if s == "" {
		return 0, 0, nil
	}

	cursor := pos.X
	ratio := int32(size / font.size)

	for _, char := range s {
		meta, ok := font.chars[char]
		if !ok {
			return 0, 0, ErrMissingChar{Face: font.face, Char: char}
		}

		tex, err := r.getFontTexture(font, meta.page)
		if err != nil {
			return 0, 0, err
		}

		src := &sdl.Rect{
			W: meta.width,
			H: meta.height,
			X: meta.x,
			Y: meta.y,
		}
		dst := &sdl.Rect{
			W: src.W * ratio,
			H: src.H * ratio,
			X: cursor + meta.xOffset*ratio,
			Y: pos.Y + meta.yOffset*ratio,
		}
		cursor += meta.xAdvance * ratio

		if err := tex.SetColorMod(colorMod(color)); err != nil {
			return 0, 0, fmt.Errorf("font: unable to set color mod of texture of page %d of font %s: %s", meta.page, font.face, err)
		}
		if err := r.Renderer.Copy(tex, src, dst); err != nil {
			return 0, 0, fmt.Errorf("font: unable to render texture of page %d of font %s: %s", meta.page, font.face, err)
		}

	}

	return cursor, font.lineHeight * ratio, nil
}
