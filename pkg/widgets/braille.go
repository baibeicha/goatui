package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Braille dot bit positions in Unicode Braille patterns (U+2800..U+28FF).
//
// Left col (x=0):    Right col (x=1):
// row 0: dot 1 (0x01)   row 0: dot 4 (0x08)
// row 1: dot 2 (0x02)   row 1: dot 5 (0x10)
// row 2: dot 3 (0x04)   row 2: dot 6 (0x20)
// row 3: dot 7 (0x40)   row 3: dot 8 (0x80)
var brailleDotMap = [4][2]uint8{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

const brailleBase = 0x2800

// BrailleCanvas provides 2x4 subpixel resolution per terminal cell.
type BrailleCanvas struct {
	widthCells  int
	heightCells int
	dots        []uint8 // bitmask for each cell (y*widthCells + x)
	fg          cell.Color
	bg          cell.Color
}

// NewBrailleCanvas creates a subpixel canvas spanning width x height terminal cells.
func NewBrailleCanvas(widthCells, heightCells int) *BrailleCanvas {
	return &BrailleCanvas{
		widthCells:  widthCells,
		heightCells: heightCells,
		dots:        make([]uint8, widthCells*heightCells),
		fg:          cell.ColorHex("#00FF88"),
		bg:          cell.DefaultColor(),
	}
}

// SubWidth returns total virtual subpixel width (2 * cell width).
func (bc *BrailleCanvas) SubWidth() int {
	return bc.widthCells * 2
}

// SubHeight returns total virtual subpixel height (4 * cell height).
func (bc *BrailleCanvas) SubHeight() int {
	return bc.heightCells * 4
}

// Reset clears all dots on the canvas without memory reallocations.
func (bc *BrailleCanvas) Reset() {
	for i := range bc.dots {
		bc.dots[i] = 0
	}
}

// Clear is an alias to Reset to clear all subpixel dots on the canvas.
func (bc *BrailleCanvas) Clear() {
	bc.Reset()
}

// SetColor configures the foreground and background colors for the Braille canvas cells.
func (bc *BrailleCanvas) SetColor(fg, bg cell.Color) {
	bc.fg = fg
	bc.bg = bg
}

// SetPixel turns on a subpixel at virtual coordinates (subX, subY).
func (bc *BrailleCanvas) SetPixel(subX, subY int) {
	if subX < 0 || subX >= bc.SubWidth() || subY < 0 || subY >= bc.SubHeight() {
		return
	}

	cellX := subX / 2
	cellY := subY / 4
	dotX := subX % 2
	dotY := subY % 4

	idx := cellY*bc.widthCells + cellX
	bc.dots[idx] |= brailleDotMap[dotY][dotX]
}

// DrawLine draws a continuous Bresenham line between two subpixels.
func (bc *BrailleCanvas) DrawLine(x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy

	for {
		bc.SetPixel(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Resize adjusts the canvas cell dimensions, reusing memory if possible.
func (bc *BrailleCanvas) Resize(widthCells, heightCells int) {
	if widthCells < 0 {
		widthCells = 0
	}
	if heightCells < 0 {
		heightCells = 0
	}
	if bc.widthCells == widthCells && bc.heightCells == heightCells {
		return
	}
	total := widthCells * heightCells
	if cap(bc.dots) >= total {
		bc.dots = bc.dots[:total]
	} else {
		bc.dots = make([]uint8, total)
	}
	bc.widthCells = widthCells
	bc.heightCells = heightCells
	bc.Reset()
}

// Draw renders the Braille canvas into the buffer at the given area.
func (bc *BrailleCanvas) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	for y := 0; y < area.Height; y++ {
		for x := 0; x < area.Width; x++ {
			r := ' '
			if x < bc.widthCells && y < bc.heightCells {
				mask := bc.dots[y*bc.widthCells+x]
				if mask != 0 {
					r = rune(brailleBase | int(mask))
				}
			}
			buf.Set(area.X+x, area.Y+y, cell.Cell{
				Rune:     r,
				Width:    1,
				FgType:   bc.fg.Type,
				BgType:   bc.bg.Type,
				Fg:       bc.fg.Value,
				Bg:       bc.bg.Value,
				Modifier: cell.AttrNone,
			})
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
