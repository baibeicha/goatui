package widgets

import (
	"strconv"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Gauge displays a percentage progress bar with an optional aligned text label.
type Gauge struct {
	percent float64 // 0.0 to 1.0
	label   string
	align   buffer.Alignment
	fg      cell.Color
	bg      cell.Color
}

// NewGauge creates a new progress gauge.
func NewGauge() *Gauge {
	return &Gauge{
		align: buffer.AlignCenter,
		fg:    cell.ColorHex("#FF007F"),
		bg:    cell.ANSI256(236),
	}
}

// SetPercent updates the progress value (0.0 .. 1.0).
func (g *Gauge) SetPercent(p float64) *Gauge {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	g.percent = p
	return g
}

// SetLabel sets an explicit label. If empty, percentage is shown (e.g. "45%").
func (g *Gauge) SetLabel(lbl string) *Gauge {
	g.label = lbl
	return g
}

// SetFg sets the gauge foreground fill color.
func (g *Gauge) SetFg(c cell.Color) *Gauge {
	g.fg = c
	return g
}

// SetBg sets the gauge background bar color.
func (g *Gauge) SetBg(c cell.Color) *Gauge {
	g.bg = c
	return g
}

// SetAlign configures the alignment of the gauge label (AlignLeft, AlignCenter, AlignRight).
func (g *Gauge) SetAlign(align buffer.Alignment) *Gauge {
	g.align = align
	return g
}

// Draw renders the gauge into area.
func (g *Gauge) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	filledWidth := int(float64(area.Width) * g.percent)

	// Draw filled portion
	for y := area.Y; y < area.Bottom(); y++ {
		for x := 0; x < area.Width; x++ {
			if x < filledWidth {
				buf.Set(area.X+x, y, cell.Cell{
					Rune:     '█',
					Width:    1,
					FgType:   g.fg.Type,
					BgType:   g.bg.Type,
					Fg:       g.fg.Value,
					Bg:       g.bg.Value,
					Modifier: cell.AttrNone,
				})
			} else {
				buf.Set(area.X+x, y, cell.Cell{
					Rune:     '░',
					Width:    1,
					FgType:   g.bg.Type,
					BgType:   g.bg.Type,
					Fg:       g.bg.Value,
					Bg:       g.bg.Value,
					Modifier: cell.AttrNone,
				})
			}
		}
	}

	// Render Aligned Label
	lbl := g.label
	if lbl == "" {
		pct := int(g.percent * 100)
		lbl = strconv.Itoa(pct) + "%"
	}

	lblWidth := buffer.StringWidth(lbl)
	if lblWidth <= area.Width {
		startX := area.X
		switch g.align {
		case buffer.AlignRight:
			startX = area.Right() - lblWidth - 1
			if startX < area.X {
				startX = area.X
			}
		case buffer.AlignCenter:
			startX = area.X + (area.Width-lblWidth)/2
		case buffer.AlignLeft:
			startX = area.X + 1
		}
		centerY := area.Y + area.Height/2
		cx := startX
		for _, r := range lbl {
			w := buffer.RuneWidth(r)
			if w == 0 {
				continue
			}
			var labelFg, labelBg cell.Color
			if cx-area.X < filledWidth {
				// Over filled bar: invert colors
				labelFg = g.bg
				labelBg = g.fg
			} else {
				// Over empty bar
				labelFg = g.fg
				labelBg = g.bg
			}
			buf.SetRune(cx, centerY, r, labelFg, labelBg, cell.AttrBold)
			cx += w
		}
	}
}
