package media

import (
	"bytes"
	"fmt"
	"image"
)

// SixelEncoder encodes images into DEC Sixel format.
type SixelEncoder struct{}

// NewSixelEncoder creates a new SixelEncoder.
func NewSixelEncoder() *SixelEncoder {
	return &SixelEncoder{}
}

// Encode converts an image into a Sixel DCS sequence.
// Quantizes colors into a 64-color (4x4x4) palette.
func (e *SixelEncoder) Encode(img image.Image) []byte {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return nil
	}

	var buf bytes.Buffer
	// DCS sequence start
	buf.WriteString("\x1bPq")

	// 64-color palette (4x4x4 RGB)
	for r := 0; r < 4; r++ {
		for g := 0; g < 4; g++ {
			for b := 0; b < 4; b++ {
				idx := r*16 + g*4 + b
				rp := r * 100 / 3
				gp := g * 100 / 3
				bp := b * 100 / 3
				buf.WriteString(fmt.Sprintf("#%d;2;%d;%d;%d", idx, rp, gp, bp))
			}
		}
	}

	for y := 0; y < h; y += 6 {
		var masks [64][]byte
		activeColors := make([]bool, 64)
		bandH := 6
		if y+bandH > h {
			bandH = h - y
		}

		for dy := 0; dy < bandH; dy++ {
			for x := 0; x < w; x++ {
				cr, cg, cb, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y+dy).RGBA()
				pr := int(cr >> 14)
				pg := int(cg >> 14)
				pb := int(cb >> 14)
				idx := pr*16 + pg*4 + pb
				if masks[idx] == nil {
					masks[idx] = make([]byte, w)
				}
				masks[idx][x] |= (1 << dy)
				activeColors[idx] = true
			}
		}

		for c := 0; c < 64; c++ {
			if !activeColors[c] {
				continue
			}
			var colorBuf bytes.Buffer
			cMasks := masks[c]
			consecutiveEmpty := 0
			used := false

			for x := 0; x < w; x++ {
				mask := cMasks[x]
				if mask > 0 {
					used = true
					for i := 0; i < consecutiveEmpty; i++ {
						colorBuf.WriteByte('?') // 0x3F is empty sixel
					}
					consecutiveEmpty = 0
					colorBuf.WriteByte(0x3F + mask)
				} else if used {
					consecutiveEmpty++
				}
			}

			if used {
				buf.WriteString(fmt.Sprintf("#%d", c))
				buf.Write(colorBuf.Bytes())
				buf.WriteString("$")
			}
		}
		if y+6 < h {
			buf.WriteString("-")
		}
	}

	buf.WriteString("\x1b\\")
	return buf.Bytes()
}
