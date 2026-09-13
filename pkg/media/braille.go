package media

import (
	"image"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// BrailleRenderer renders images using Unicode Braille patterns (U+2800-U+28FF).
// Each cell maps to a 2x4 dot matrix, providing 8 subpixels per terminal cell.
type BrailleRenderer struct {
	Threshold uint8 // Luminance threshold for dot on/off (default: 128)
	Dithering bool
}

// NewBrailleRenderer creates a new BrailleRenderer with default settings.
func NewBrailleRenderer() *BrailleRenderer {
	return &BrailleRenderer{Threshold: 128}
}

// brailleOffset maps (dx, dy) within a 2x4 cell to the braille bit offset.
// Braille Unicode dots are arranged:
//   (0,0)=bit0  (1,0)=bit3
//   (0,1)=bit1  (1,1)=bit4
//   (0,2)=bit2  (1,2)=bit5
//   (0,3)=bit6  (1,3)=bit7
var brailleOffset = [2][4]uint8{
	{0, 1, 2, 6}, // column 0
	{3, 4, 5, 7}, // column 1
}

// DrawImage renders an image into the given buffer area using braille characters.
// Each cell represents a 2x4 pixel block from the source image.
// Uses the dominant color of the block as the foreground color.
func (br *BrailleRenderer) DrawImage(buf *buffer.Buffer, area buffer.Rect, img image.Image) {
	if area.IsEmpty() || img == nil {
		return
	}

	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW == 0 || srcH == 0 {
		return
	}

	// Canvas dimensions in braille subpixels
	canvasW := area.Width * 2   // 2 dots per cell horizontally
	canvasH := area.Height * 4  // 4 dots per cell vertically

	scaleX := float64(srcW) / float64(canvasW)
	scaleY := float64(srcH) / float64(canvasH)

	dotOn := make([][]bool, canvasW)
	for i := range dotOn {
		dotOn[i] = make([]bool, canvasH)
	}

	type rgb struct{ r, g, b float64 }
	colors := make([][]rgb, canvasW)
	for i := range colors {
		colors[i] = make([]rgb, canvasH)
	}

	var errBuf [][]float64
	if br.Dithering {
		errBuf = make([][]float64, canvasW+2)
		for i := range errBuf {
			errBuf[i] = make([]float64, canvasH+2)
		}
	}

	for y := 0; y < canvasH; y++ {
		for x := 0; x < canvasW; x++ {
			srcX := int(float64(x)*scaleX + 0.5)
			srcY := int(float64(y)*scaleY + 0.5)

			if srcX >= srcW { srcX = srcW - 1 }
			if srcY >= srcH { srcY = srcH - 1 }
			if srcX < 0 { srcX = 0 }
			if srcY < 0 { srcY = 0 }

			r, g, b, _ := img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY).RGBA()
			rc := float64(r >> 8)
			gc := float64(g >> 8)
			bc := float64(b >> 8)
			colors[x][y] = rgb{rc, gc, bc}

			lum := 0.299*rc + 0.587*gc + 0.114*bc

			if br.Dithering {
				lum += errBuf[x+1][y+1]
				if lum < 0 {
					lum = 0
				} else if lum > 255 {
					lum = 255
				}
			}

			isOn := lum >= float64(br.Threshold)
			dotOn[x][y] = isOn

			if br.Dithering {
				var err float64
				if isOn {
					err = lum - 255
				} else {
					err = lum
				}

				errBuf[x+2][y+1] += err * 7 / 16
				errBuf[x][y+2] += err * 3 / 16
				errBuf[x+1][y+2] += err * 5 / 16
				errBuf[x+2][y+2] += err * 1 / 16
			}
		}
	}

	for cy := 0; cy < area.Height; cy++ {
		for cx := 0; cx < area.Width; cx++ {
			var pattern uint8
			var rSum, gSum, bSum float64
			var dotCount float64

			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					subX := cx*2 + dx
					subY := cy*4 + dy

					if dotOn[subX][subY] {
						pattern |= 1 << brailleOffset[dx][dy]
						c := colors[subX][subY]
						rSum += c.r
						gSum += c.g
						bSum += c.b
						dotCount++
					}
				}
			}

			brailleRune := rune(0x2800 + int(pattern))

			var fg cell.Color
			if dotCount > 0 {
				fg = cell.RGB(
					uint8(rSum/dotCount),
					uint8(gSum/dotCount),
					uint8(bSum/dotCount),
				)
			} else {
				fg = cell.DefaultColor()
			}

			buf.SetRune(
				area.X+cx,
				area.Y+cy,
				brailleRune,
				fg,
				cell.DefaultColor(),
				cell.AttrNone,
			)
		}
	}
}
