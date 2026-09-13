package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

var sparkBars = [8]rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline displays compact numerical trend charts using 1/8th Unicode block elements.
type Sparkline struct {
	data []float64
	fg   cell.Color
	bg   cell.Color
}

// NewSparkline creates a new sparkline widget.
func NewSparkline(data []float64) *Sparkline {
	return &Sparkline{
		data: data,
		fg:   cell.ColorHex("#00D2FF"),
		bg:   cell.DefaultColor(),
	}
}

// SetData updates the telemetry data points.
func (s *Sparkline) SetData(data []float64) {
	s.data = data
}

// Draw renders the sparkline into the specified area.
func (s *Sparkline) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(s.data) == 0 {
		return
	}

	minVal := s.data[0]
	maxVal := s.data[0]
	for _, v := range s.data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	valRange := maxVal - minVal
	if valRange <= 0 {
		valRange = 1
	}

	// Render the most recent data points fitting area.Width
	dataLen := len(s.data)
	startIdx := max(0, dataLen-area.Width)

	for x := 0; x < area.Width; x++ {
		if startIdx+x < dataLen {
			val := s.data[startIdx+x]
			norm := (val - minVal) / valRange
			barIdx := int(norm * 7)
			if barIdx < 0 {
				barIdx = 0
			}
			if barIdx > 7 {
				barIdx = 7
			}

			buf.Set(area.X+x, area.Y, cell.Cell{
				Rune:     sparkBars[barIdx],
				Width:    1,
				FgType:   s.fg.Type,
				BgType:   s.bg.Type,
				Fg:       s.fg.Value,
				Bg:       s.bg.Value,
				Modifier: cell.AttrNone,
			})
		} else {
			buf.Set(area.X+x, area.Y, cell.BlankCell)
		}
	}
}

// SetFg sets the sparkline foreground color.
func (s *Sparkline) SetFg(c cell.Color) *Sparkline {
	s.fg = c
	return s
}

// SetBg sets the sparkline background color.
func (s *Sparkline) SetBg(c cell.Color) *Sparkline {
	s.bg = c
	return s
}
