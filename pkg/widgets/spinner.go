package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
)

// SpinnerType represents animation frame styles for spinners.
type SpinnerType int

const (
	SpinnerDots SpinnerType = iota
	SpinnerLine
	SpinnerMiniDots
	SpinnerPulse
	SpinnerArc
	SpinnerCircle
)

var spinnerFrames = map[SpinnerType][]string{
	SpinnerDots:     {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	SpinnerLine:     {"|", "/", "-", "\\"},
	SpinnerMiniDots: {"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
	SpinnerPulse:    {"█", "▉", "▊", "▋", "▌", "▍", "▎", "▏", "▎", "▍", "▌", "▋", "▊", "▉"},
	SpinnerArc:      {"◜", "◠", "◝", "◞", "◡", "◟"},
	SpinnerCircle:   {"◐", "◓", "◑", "◒"},
}

// Spinner displays an animated activity indicator with an optional text label.
type Spinner struct {
	frames       []string
	currentFrame int
	label        string
	spinnerStyle style.Style
	labelStyle   style.Style
}

// NewSpinner creates a new Spinner widget.
func NewSpinner(sType SpinnerType, label string) *Spinner {
	frames, ok := spinnerFrames[sType]
	if !ok {
		frames = spinnerFrames[SpinnerDots]
	}
	return &Spinner{
		frames:       frames,
		currentFrame: 0,
		label:        label,
		spinnerStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		labelStyle: style.NewStyle().
			Foreground(cell.ColorHex("#CCCCDD")),
	}
}

// SetType changes the spinner animation preset.
func (s *Spinner) SetType(sType SpinnerType) *Spinner {
	if frames, ok := spinnerFrames[sType]; ok {
		s.frames = frames
		s.currentFrame = 0
	}
	return s
}

// SetLabel updates the spinner label.
func (s *Spinner) SetLabel(label string) *Spinner {
	s.label = label
	return s
}

// SetSpinnerStyle configures styling of the spinning glyph.
func (s *Spinner) SetSpinnerStyle(st style.Style) *Spinner {
	s.spinnerStyle = st
	return s
}

// SetLabelStyle configures styling of the label text.
func (s *Spinner) SetLabelStyle(st style.Style) *Spinner {
	s.labelStyle = st
	return s
}

// Tick advances the spinner to its next animation frame.
func (s *Spinner) Tick() *Spinner {
	if len(s.frames) > 0 {
		s.currentFrame = (s.currentFrame + 1) % len(s.frames)
	}
	return s
}

// CurrentGlyph returns the current frame string.
func (s *Spinner) CurrentGlyph() string {
	if len(s.frames) == 0 {
		return ""
	}
	idx := (s.currentFrame%len(s.frames) + len(s.frames)) % len(s.frames)
	return s.frames[idx]
}

// Draw renders the spinner glyph and label into area.
func (s *Spinner) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(s.frames) == 0 {
		return
	}

	glyph := s.CurrentGlyph()
	glyphW := buffer.StringWidth(glyph)
	drawW := min(glyphW, area.Width)

	// Draw glyph
	s.spinnerStyle.Draw(buf, buffer.NewRect(area.X, area.Y, drawW, 1), glyph)

	// Draw label
	if s.label != "" && area.Width > glyphW+1 {
		lblArea := buffer.NewRect(area.X+glyphW+1, area.Y, area.Width-glyphW-1, 1)
		s.labelStyle.Draw(buf, lblArea, s.label)
	}
}
