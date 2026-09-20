package ui

import (
	"fmt"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
)

// Span represents a styled text fragment.
type Span struct {
	Text  string
	Style style.Style
}

// NewSpan creates a new Span with the given text and style.
func NewSpan(text string, s style.Style) Span {
	return Span{Text: text, Style: s}
}

// Text creates an unstyled plain text span.
func Text(text string) Span {
	return Span{Text: text, Style: style.NewStyle()}
}

// Styled creates a span with a custom style.
func Styled(text string, s style.Style) Span {
	return Span{Text: text, Style: s}
}

// Bold creates a bold text span.
func Bold(text string) Span {
	return Span{Text: text, Style: style.NewStyle().Bold(true)}
}

// Dim creates a dimmed text span.
func Dim(text string) Span {
	return Span{Text: text, Style: style.NewStyle().Dim(true)}
}

// Italic creates an italic text span.
func Italic(text string) Span {
	return Span{Text: text, Style: style.NewStyle().Italic(true)}
}

// Color creates a text span with the specified foreground color.
func Color(text string, fg cell.Color) Span {
	return Span{Text: text, Style: style.NewStyle().Foreground(fg)}
}

// Colored creates a text span with foreground and background colors.
func Colored(text string, fg, bg cell.Color) Span {
	return Span{Text: text, Style: style.NewStyle().Foreground(fg).Background(bg)}
}

// BadgeSpan creates an inline badge/pill span.
func BadgeSpan(text string, fg, bg cell.Color) Span {
	return Span{
		Text: " " + text + " ",
		Style: style.NewStyle().
			Bold(true).
			Foreground(fg).
			Background(bg),
	}
}

// Width returns the display column width of the span.
func (s Span) Width() int {
	return buffer.StringWidth(s.Text)
}

// Draw renders the span directly into buffer area, implementing the View interface.
func (s Span) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	s.Style.Draw(buf, area, s.Text)
}

// Line represents a single horizontal row of rich text composed of multiple Spans.
type Line struct {
	Spans []Span
	Align buffer.Alignment
}

// NewLine constructs a Line from the provided Spans.
func NewLine(spans ...Span) Line {
	return Line{
		Spans: spans,
		Align: buffer.AlignLeft,
	}
}

// LineFromText constructs a Line containing a single plain text Span.
func LineFromText(text string) Line {
	return Line{
		Spans: []Span{Text(text)},
		Align: buffer.AlignLeft,
	}
}

// Draw renders the line into buffer area, implementing the View interface.
func (l Line) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	l.Render(buf, area.X, area.Y, area.Width)
}

// Width returns the total visual width of the line in terminal columns.
func (l Line) Width() int {
	w := 0
	for _, s := range l.Spans {
		w += s.Width()
	}
	return w
}

// SetAlign sets alignment for the line (Left, Center, Right).
func (l Line) SetAlign(align buffer.Alignment) Line {
	l.Align = align
	return l
}

// Render draws the line into buf at row y within width maxW.
func (l Line) Render(buf *buffer.Buffer, x, y, maxW int) {
	if y < 0 || y >= buf.Height() || maxW <= 0 {
		return
	}

	totalW := l.Width()
	startX := x

	switch l.Align {
	case buffer.AlignCenter:
		startX = x + max(0, (maxW-totalW)/2)
	case buffer.AlignRight:
		startX = x + max(0, maxW-totalW)
	}

	currX := startX
	for _, span := range l.Spans {
		if currX >= x+maxW {
			break
		}

		spanW := span.Width()
		if spanW <= 0 {
			continue
		}

		drawW := min(spanW, x+maxW-currX)
		spanArea := buffer.NewRect(currX, y, drawW, 1)

		span.Style.Draw(buf, spanArea, span.Text)
		currX += spanW
	}
}

// ANSI formats the line into an ANSI escape sequence string for direct terminal output.
func (l Line) ANSI() string {
	var sb strings.Builder
	for _, span := range l.Spans {
		st := span.Style
		fg := st.GetFg()
		bg := st.GetBg()
		mod := st.GetModifier()

		var codes []string

		if mod.Has(cell.AttrBold) {
			codes = append(codes, "1")
		}
		if mod.Has(cell.AttrDim) {
			codes = append(codes, "2")
		}
		if mod.Has(cell.AttrItalic) {
			codes = append(codes, "3")
		}
		if mod.Has(cell.AttrUnderline) {
			codes = append(codes, "4")
		}
		if mod.Has(cell.AttrBlink) {
			codes = append(codes, "5")
		}
		if mod.Has(cell.AttrReverse) {
			codes = append(codes, "7")
		}
		if mod.Has(cell.AttrStrikethrough) {
			codes = append(codes, "9")
		}

		// Foreground
		switch fg.Type {
		case cell.ColorANSI16:
			if fg.Value < 8 {
				codes = append(codes, fmt.Sprintf("%d", 30+fg.Value))
			} else {
				codes = append(codes, fmt.Sprintf("%d", 90+fg.Value-8))
			}
		case cell.ColorANSI256:
			codes = append(codes, fmt.Sprintf("38;5;%d", fg.Value))
		case cell.ColorRGB:
			r := (fg.Value >> 16) & 0xFF
			g := (fg.Value >> 8) & 0xFF
			b := fg.Value & 0xFF
			codes = append(codes, fmt.Sprintf("38;2;%d;%d;%d", r, g, b))
		}

		// Background
		switch bg.Type {
		case cell.ColorANSI16:
			if bg.Value < 8 {
				codes = append(codes, fmt.Sprintf("%d", 40+bg.Value))
			} else {
				codes = append(codes, fmt.Sprintf("%d", 100+bg.Value-8))
			}
		case cell.ColorANSI256:
			codes = append(codes, fmt.Sprintf("48;5;%d", bg.Value))
		case cell.ColorRGB:
			r := (bg.Value >> 16) & 0xFF
			g := (bg.Value >> 8) & 0xFF
			b := bg.Value & 0xFF
			codes = append(codes, fmt.Sprintf("48;2;%d;%d;%d", r, g, b))
		}

		if len(codes) > 0 {
			sb.WriteString("\x1b[" + strings.Join(codes, ";") + "m")
		}
		sb.WriteString(span.Text)
		if len(codes) > 0 {
			sb.WriteString("\x1b[0m")
		}
	}
	return sb.String()
}
