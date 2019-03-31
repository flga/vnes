package gui

import (
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"strings"

	"github.com/ftrvxmtrx/tga"
)

var ErrUnsupported = errors.New("could not decode font page, make sure it is in either png or tga format")

type PageLoader func(path string) (io.ReadCloser, error)

type ErrMissingChar struct {
	Face string
	Char rune
}

func (e ErrMissingChar) Error() string {
	return fmt.Sprintf("font face %s does not have char %q", e.Face, e.Char)
}

type char struct {
	id       rune
	x        int32
	y        int32
	width    int32
	height   int32
	xOffset  int32
	yOffset  int32
	xAdvance int32
	page     int
}

type Font struct {
	// info
	face string
	size int

	// common
	lineHeight int32
	base       int32
	scaleW     int32
	scaleH     int32

	// pages
	pages []*image.RGBA

	// chars
	chars map[rune]char
}

func (f *Font) Bounds(s string, size int) (w, h int32) {
	if s == "" {
		return 0, 0
	}

	ratio := int32(size / f.size)
	lines := strings.Split(s, "\n")
	numLines := int32(len(lines))

	for i := int32(0); i < numLines; i++ {
		var (
			lw int32

			line     = lines[i]
			lastChar = len(line) - 1
		)

		for j, char := range line {
			if j == lastChar {
				lw += f.chars[char].width * ratio
			} else {
				lw += f.chars[char].xAdvance * ratio
			}
		}

		if lw > w {
			w = lw
		}
	}

	return w, numLines * f.lineHeight * ratio
}

type FontMap map[string]*Font

func (m FontMap) LoadXML(r io.Reader, loader PageLoader) error {
	d := xml.NewDecoder(r)

	xmlData := struct {
		XMLName xml.Name `xml:"font"`
		Info    struct {
			Face string `xml:"face,attr"`
			Size int    `xml:"size,attr"`
		} `xml:"info"`
		Common struct {
			LineHeight int32 `xml:"lineHeight,attr"`
			Base       int32 `xml:"base,attr"`
			ScaleW     int32 `xml:"scaleW,attr"`
			ScaleH     int32 `xml:"scaleH,attr"`
			Pages      int   `xml:"pages,attr"`
		} `xml:"common"`
		Pages struct {
			Page []struct {
				ID   int    `xml:"id,attr"`
				File string `xml:"file,attr"`
			} `xml:"page"`
		} `xml:"pages"`
		Chars struct {
			Count string `xml:"count,attr"`
			Char  []struct {
				ID       rune  `xml:"id,attr"`
				X        int32 `xml:"x,attr"`
				Y        int32 `xml:"y,attr"`
				Width    int32 `xml:"width,attr"`
				Height   int32 `xml:"height,attr"`
				Xoffset  int32 `xml:"xoffset,attr"`
				Yoffset  int32 `xml:"yoffset,attr"`
				Xadvance int32 `xml:"xadvance,attr"`
				Page     int   `xml:"page,attr"`
			} `xml:"char"`
		} `xml:"chars"`
	}{}
	if err := d.Decode(&xmlData); err != nil {
		return fmt.Errorf("font: unable to decode font data: %s", err)
	}

	pages := make([]*image.RGBA, xmlData.Common.Pages)
	for _, pageInfo := range xmlData.Pages.Page {
		pageReader, err := loader(pageInfo.File)
		if err != nil {
			return fmt.Errorf("font: unable to read page: %s", err)
		}

		page, err := decode(pageReader)
		if err != nil {
			return fmt.Errorf("font: unable to decode page: %s", err)
		}

		var pageRGBA *image.RGBA
		if img, ok := page.(*image.RGBA); ok {
			pageRGBA = img
		} else {
			pageRGBA = image.NewRGBA(page.Bounds())
			draw.Draw(pageRGBA, page.Bounds(), page, image.Pt(0, 0), draw.Src)
		}

		pages[pageInfo.ID] = pageRGBA
	}

	chars := make(map[rune]char)
	for _, charInfo := range xmlData.Chars.Char {
		char := char{
			id:       charInfo.ID,
			x:        charInfo.X,
			y:        charInfo.Y,
			width:    charInfo.Width,
			height:   charInfo.Height,
			xOffset:  charInfo.Xoffset,
			yOffset:  charInfo.Yoffset,
			xAdvance: charInfo.Xadvance,
			page:     charInfo.Page,
		}
		chars[char.id] = char
	}

	m[xmlData.Info.Face] = &Font{
		// info
		face: xmlData.Info.Face,
		size: xmlData.Info.Size,

		// common
		lineHeight: xmlData.Common.LineHeight,
		base:       xmlData.Common.Base,
		scaleW:     xmlData.Common.ScaleW,
		scaleH:     xmlData.Common.ScaleH,

		// pages
		pages: pages,

		// chars
		chars: chars,
	}

	return nil
}

func decode(r io.ReadCloser) (image.Image, error) {
	defer r.Close()
	if i, err := png.Decode(r); err == nil {
		return i, nil
	}

	if i, err := tga.Decode(r); err == nil {
		return i, nil
	}

	return nil, ErrUnsupported
}
