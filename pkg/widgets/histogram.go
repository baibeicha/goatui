package widgets

import (
	"fmt"
	"math"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/style"
)

// HistogramBar represents an individual bar data point.
type HistogramBar struct {
	Label string
	Value float64
	Color cell.Color
}

// Histogram renders a vertical or horizontal bar chart using terminal block runes.
type Histogram struct {
	bars        []HistogramBar
	orientation layout.Direction
	maxVal      float64
	barWidth    int
	showValues  bool
	styleLabel  style.Style
}

// NewHistogram creates a new Histogram widget.
func NewHistogram(bars ...HistogramBar) *Histogram {
	return &Histogram{
		bars:        bars,
		orientation: layout.Vertical,
		barWidth:    3,
		showValues:  true,
		styleLabel: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")),
	}
}

// SetBars updates the data points.
func (h *Histogram) SetBars(bars []HistogramBar) *Histogram {
	h.bars = bars
	return h
}

// SetOrientation sets Vertical (columns) or Horizontal (rows).
func (h *Histogram) SetOrientation(dir layout.Direction) *Histogram {
	h.orientation = dir
	return h
}

// SetMaxVal overrides the maximum scale value.
func (h *Histogram) SetMaxVal(maxVal float64) *Histogram {
	h.maxVal = maxVal
	return h
}

// SetBarWidth sets column width for vertical histogram.
func (h *Histogram) SetBarWidth(w int) *Histogram {
	if w > 0 {
		h.barWidth = w
	}
	return h
}

// SetShowValues toggles numerical value labels.
func (h *Histogram) SetShowValues(show bool) *Histogram {
	h.showValues = show
	return h
}

var vertRunes = []rune{' ', ' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
var horizRunes = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}

// Draw renders the histogram into area.
func (h *Histogram) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(h.bars) == 0 {
		return
	}

	maxV := h.maxVal
	if maxV <= 0 {
		for _, b := range h.bars {
			if b.Value > maxV {
				maxV = b.Value
			}
		}
	}
	if maxV <= 0 {
		maxV = 1.0
	}

	if h.orientation == layout.Horizontal {
		h.drawHorizontal(buf, area, maxV)
	} else {
		h.drawVertical(buf, area, maxV)
	}
}

func (h *Histogram) drawVertical(buf *buffer.Buffer, area buffer.Rect, maxV float64) {
	chartH := area.Height - 1 // reserve 1 line for label
	if h.showValues {
		chartH-- // reserve 1 line for values at top
	}
	if chartH <= 0 {
		return
	}

	spacing := 1
	currX := area.X

	for _, bar := range h.bars {
		if currX+h.barWidth > area.Right() {
			break
		}

		barColor := bar.Color
		if barColor.Type == cell.ColorDefault {
			barColor = cell.ColorHex("#00D2FF")
		}

		ratio := math.Max(0.0, math.Min(1.0, bar.Value/maxV))
		fullHeightExact := ratio * float64(chartH)
		fullBlocks := int(fullHeightExact)
		frac := fullHeightExact - float64(fullBlocks)
		partialIdx := int(frac * 8.0)

		// Draw value at top
		if h.showValues {
			valStr := fmt.Sprintf("%.0f", bar.Value)
			if buffer.StringWidth(valStr) > h.barWidth {
				valStr = string([]rune(valStr)[:max(1, h.barWidth)])
			}
			buf.SetStringAligned(buffer.NewRect(currX, area.Y, h.barWidth, 1), valStr, buffer.AlignCenter, cell.ColorHex("#CCCCDD"), cell.DefaultColor(), cell.AttrNone)
		}

		chartTop := area.Y
		if h.showValues {
			chartTop++
		}
		chartBottom := chartTop + chartH

		// Fill vertical bar from bottom up
		for row := 0; row < chartH; row++ {
			y := chartBottom - 1 - row
			var r rune
			if row < fullBlocks {
				r = '█'
			} else if row == fullBlocks && partialIdx > 0 {
				r = vertRunes[partialIdx]
			} else {
				r = ' '
			}

			if r != ' ' {
				for col := 0; col < h.barWidth; col++ {
					buf.SetRune(currX+col, y, r, barColor, cell.DefaultColor(), cell.AttrNone)
				}
			}
		}

		// Draw label at bottom
		labelY := chartBottom
		if labelY < area.Bottom() {
			buf.SetStringAligned(buffer.NewRect(currX, labelY, h.barWidth, 1), bar.Label, buffer.AlignCenter, cell.ColorHex("#8888AA"), cell.DefaultColor(), cell.AttrNone)
		}

		currX += h.barWidth + spacing
	}
}

func (h *Histogram) drawHorizontal(buf *buffer.Buffer, area buffer.Rect, maxV float64) {
	currY := area.Y
	for _, bar := range h.bars {
		if currY >= area.Bottom() {
			break
		}

		barColor := bar.Color
		if barColor.Type == cell.ColorDefault {
			barColor = cell.ColorHex("#00FFAA")
		}

		lblStr := bar.Label + " "
		lblW := buffer.StringWidth(lblStr)
		valStr := ""
		if h.showValues {
			valStr = fmt.Sprintf(" %.1f", bar.Value)
		}
		valW := buffer.StringWidth(valStr)

		if lblW+valW > area.Width {
			lblW = max(0, area.Width-valW)
		}

		barSpace := area.Width - lblW - valW
		if barSpace < 0 {
			barSpace = 0
		}

		// Draw label
		if lblW > 0 {
			buf.SetStringAligned(buffer.NewRect(area.X, currY, lblW, 1), lblStr, buffer.AlignLeft, cell.ColorHex("#8888AA"), cell.DefaultColor(), cell.AttrNone)
		}

		// Draw bar
		if barSpace > 0 {
			ratio := math.Max(0.0, math.Min(1.0, bar.Value/maxV))
			fullLenExact := ratio * float64(barSpace)
			fullBlocks := int(fullLenExact)
			frac := fullLenExact - float64(fullBlocks)
			partialIdx := int(frac * 8.0)

			bx := area.X + lblW
			for i := 0; i < barSpace; i++ {
				var r rune
				if i < fullBlocks {
					r = '█'
				} else if i == fullBlocks && partialIdx > 0 {
					r = horizRunes[partialIdx]
				} else {
					break
				}
				buf.SetRune(bx+i, currY, r, barColor, cell.DefaultColor(), cell.AttrNone)
			}
		}

		// Draw value right-aligned at end of allocated space
		if valW > 0 && area.Width >= valW {
			valRect := buffer.NewRect(area.Right()-valW, currY, valW, 1)
			buf.SetStringAligned(valRect, valStr, buffer.AlignRight, cell.ColorHex("#CCCCDD"), cell.DefaultColor(), cell.AttrBold)
		}

		currY++
	}
}
