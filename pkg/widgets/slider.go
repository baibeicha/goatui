package widgets

import (
	"fmt"
	"math"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// Slider is an interactive progress/numerical range adjuster.
type Slider struct {
	id          string
	label       string
	minVal      float64
	maxVal      float64
	step        float64
	value       float64
	showValue   bool
	format      string
	focused     bool
	disabled    bool
	bounds      buffer.Rect
	trackBounds buffer.Rect
	onChange    func(val float64)

	styleTrack   style.Style
	styleFilled  style.Style
	styleHandle  style.Style
	styleLabel   style.Style
	styleFocused style.Style
}

// NewSlider creates a new Slider widget.
func NewSlider(id, label string, minVal, maxVal, initial float64) *Slider {
	if minVal >= maxVal {
		maxVal = minVal + 1
	}
	initial = math.Max(minVal, math.Min(maxVal, initial))

	return &Slider{
		id:        id,
		label:     label,
		minVal:    minVal,
		maxVal:    maxVal,
		step:      (maxVal - minVal) / 20.0,
		value:     initial,
		showValue: true,
		format:    "%.1f",
		styleTrack: style.NewStyle().
			Foreground(cell.ColorHex("#444455")),
		styleFilled: style.NewStyle().
			Foreground(cell.ColorHex("#00D2FF")),
		styleHandle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		styleLabel: style.NewStyle().
			Foreground(cell.ColorHex("#CCCCDD")),
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
	}
}

// Value returns current slider value.
func (s *Slider) Value() float64 {
	return s.value
}

// SetValue sets the slider value, clamped between min and max.
func (s *Slider) SetValue(v float64) *Slider {
	v = math.Max(s.minVal, math.Min(s.maxVal, v))
	if s.value != v {
		s.value = v
		if s.onChange != nil {
			s.onChange(v)
		}
	}
	return s
}

// SetStep sets increment step.
func (s *Slider) SetStep(step float64) *Slider {
	if step > 0 {
		s.step = step
	}
	return s
}

// SetFormat configures value display format (e.g. "%.0f%%").
func (s *Slider) SetFormat(fmtStr string) *Slider {
	s.format = fmtStr
	return s
}

// SetShowValue toggles numerical value indicator.
func (s *Slider) SetShowValue(show bool) *Slider {
	s.showValue = show
	return s
}

// SetFocused sets focus state.
func (s *Slider) SetFocused(f bool) *Slider {
	s.focused = f
	return s
}

// SetDisabled sets disabled state.
func (s *Slider) SetDisabled(d bool) *Slider {
	s.disabled = d
	return s
}

// SetOnChange registers value change callback.
func (s *Slider) SetOnChange(fn func(val float64)) *Slider {
	s.onChange = fn
	return s
}

// HandleKey handles arrow key navigation.
func (s *Slider) HandleKey(key input.Key) bool {
	if !s.focused || s.disabled {
		return false
	}
	switch key.Type {
	case input.KeyLeft, input.KeyDown:
		s.SetValue(s.value - s.step)
		return true
	case input.KeyRight, input.KeyUp:
		s.SetValue(s.value + s.step)
		return true
	case input.KeyHome:
		s.SetValue(s.minVal)
		return true
	case input.KeyEnd:
		s.SetValue(s.maxVal)
		return true
	case input.KeyPgUp:
		s.SetValue(s.value + s.step*5)
		return true
	case input.KeyPgDown:
		s.SetValue(s.value - s.step*5)
		return true
	}
	return false
}

// HandleMouse processes clicking or dragging on the slider track.
func (s *Slider) HandleMouse(msg tea.MouseMsg) bool {
	if s.disabled || s.trackBounds.IsEmpty() {
		return false
	}
	if (msg.Action == input.MousePress || msg.Action == input.MouseDrag) && msg.Button == input.MouseLeft {
		if s.bounds.Contains(msg.X, msg.Y) || s.trackBounds.Contains(msg.X, msg.Y) {
			s.focused = true
			ratio := float64(msg.X-s.trackBounds.X) / float64(max(1, s.trackBounds.Width-1))
			ratio = math.Max(0.0, math.Min(1.0, ratio))
			newVal := s.minVal + ratio*(s.maxVal-s.minVal)
			s.SetValue(newVal)
			return true
		}
	}
	return false
}

// Draw renders the slider widget.
func (s *Slider) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	s.bounds = area

	labelStr := ""
	if s.label != "" {
		labelStr = s.label + ": "
	}
	valStr := ""
	if s.showValue {
		valStr = fmt.Sprintf(" "+s.format, s.value)
	}

	lblW := buffer.StringWidth(labelStr)
	valW := buffer.StringWidth(valStr)

	if valW > area.Width {
		valW = 0
		valStr = ""
	}
	if lblW+valW > area.Width {
		lblW = max(0, area.Width-valW)
	}

	trackW := area.Width - lblW - valW
	if trackW < 1 {
		trackW = 0
	}

	trackX := area.X + lblW
	s.trackBounds = buffer.NewRect(trackX, area.Y, trackW, 1)

	// Draw label
	lblSt := s.styleLabel
	if s.focused {
		lblSt = s.styleFocused
	}
	if lblW > 0 {
		lblSt.Draw(buf, buffer.NewRect(area.X, area.Y, lblW, 1), labelStr)
	}

	// Draw track
	if trackW > 0 {
		ratio := 0.0
		if s.maxVal > s.minVal {
			ratio = (s.value - s.minVal) / (s.maxVal - s.minVal)
		}
		ratio = math.Max(0.0, math.Min(1.0, ratio))

		handlePos := int(math.Round(ratio * float64(trackW-1)))

		trackFg := s.styleTrack.GetFg()
		trackBg := s.styleTrack.GetBg()
		filledFg := s.styleFilled.GetFg()
		filledBg := s.styleFilled.GetBg()
		handleFg := s.styleHandle.GetFg()
		handleBg := s.styleHandle.GetBg()

		for i := 0; i < trackW; i++ {
			x := trackX + i
			if x >= area.Right() {
				break
			}
			if i == handlePos {
				buf.SetRune(x, area.Y, '●', handleFg, handleBg, cell.AttrBold)
			} else if i < handlePos {
				buf.SetRune(x, area.Y, '━', filledFg, filledBg, cell.AttrNone)
			} else {
				buf.SetRune(x, area.Y, '─', trackFg, trackBg, cell.AttrDim)
			}
		}
	}

	// Draw value
	if valW > 0 && area.Right() >= valW {
		buf.SetStringAligned(buffer.NewRect(area.Right()-valW, area.Y, valW, 1), valStr, buffer.AlignRight, s.styleHandle.GetFg(), cell.DefaultColor(), cell.AttrBold)
	}
}
