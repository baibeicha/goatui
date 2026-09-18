package style

import (
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Alignment options
type AlignHorizontal = buffer.Alignment
type AlignVertical = buffer.VerticalAlignment
type Alignment = buffer.Alignment
type VerticalAlignment = buffer.VerticalAlignment

const (
	AlignLeft   = buffer.AlignLeft
	AlignCenter = buffer.AlignCenter
	AlignRight  = buffer.AlignRight

	AlignTop    = buffer.AlignTop
	AlignMiddle = buffer.AlignMiddle
	AlignBottom = buffer.AlignBottom
)

// Style provides a fluent, immutable builder for text styling and direct buffer drawing.
type Style struct {
	fg          cell.Color
	bg          cell.Color
	borderFg    cell.Color
	borderBg    cell.Color
	modifier    cell.Modifier
	border      Border
	hasBorder   bool
	title       string
	padTop      int
	padRight    int
	padBottom   int
	padLeft     int
	marginTop   int
	marginRight int
	marginBot   int
	marginLeft  int
	alignH      AlignHorizontal
	alignV      AlignVertical
}

// NewStyle creates a default empty style.
func NewStyle() Style {
	return Style{
		fg:       cell.DefaultColor(),
		bg:       cell.DefaultColor(),
		borderFg: cell.DefaultColor(),
		borderBg: cell.DefaultColor(),
	}
}

func (s Style) Foreground(c cell.Color) Style {
	s.fg = c
	return s
}

func (s Style) Background(c cell.Color) Style {
	s.bg = c
	return s
}

func (s Style) Bold(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrBold)
	} else {
		s.modifier.Remove(cell.AttrBold)
	}
	return s
}

func (s Style) Dim(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrDim)
	} else {
		s.modifier.Remove(cell.AttrDim)
	}
	return s
}

func (s Style) Italic(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrItalic)
	} else {
		s.modifier.Remove(cell.AttrItalic)
	}
	return s
}

func (s Style) Underline(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrUnderline)
	} else {
		s.modifier.Remove(cell.AttrUnderline)
	}
	return s
}

func (s Style) Blink(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrBlink)
	} else {
		s.modifier.Remove(cell.AttrBlink)
	}
	return s
}

func (s Style) Reverse(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrReverse)
	} else {
		s.modifier.Remove(cell.AttrReverse)
	}
	return s
}

func (s Style) Strikethrough(v bool) Style {
	if v {
		s.modifier.Add(cell.AttrStrikethrough)
	} else {
		s.modifier.Remove(cell.AttrStrikethrough)
	}
	return s
}

func (s Style) Border(b Border) Style {
	s.border = b
	s.hasBorder = true
	return s
}

func (s Style) Title(t string) Style {
	s.title = t
	return s
}

func (s Style) GetTitle() string {
	return s.title
}

func (s Style) BorderForeground(c cell.Color) Style {
	s.borderFg = c
	return s
}

func (s Style) BorderBackground(c cell.Color) Style {
	s.borderBg = c
	return s
}

func (s Style) Padding(top, right, bottom, left int) Style {
	s.padTop = max(0, top)
	s.padRight = max(0, right)
	s.padBottom = max(0, bottom)
	s.padLeft = max(0, left)
	return s
}

func (s Style) Margin(top, right, bottom, left int) Style {
	s.marginTop = max(0, top)
	s.marginRight = max(0, right)
	s.marginBot = max(0, bottom)
	s.marginLeft = max(0, left)
	return s
}

func (s Style) Align(h AlignHorizontal, v AlignVertical) Style {
	s.alignH = h
	s.alignV = v
	return s
}

// AlignHorizontal sets horizontal alignment.
func (s Style) AlignHorizontal(h AlignHorizontal) Style {
	s.alignH = h
	return s
}

// AlignVertical sets vertical alignment.
func (s Style) AlignVertical(v AlignVertical) Style {
	s.alignV = v
	return s
}

// AlignLeft sets horizontal alignment to AlignLeft.
func (s Style) AlignLeft() Style {
	s.alignH = AlignLeft
	return s
}

// AlignCenter sets horizontal alignment to AlignCenter.
func (s Style) AlignCenter() Style {
	s.alignH = AlignCenter
	return s
}

// AlignRight sets horizontal alignment to AlignRight.
func (s Style) AlignRight() Style {
	s.alignH = AlignRight
	return s
}

// AlignTop sets vertical alignment to AlignTop.
func (s Style) AlignTop() Style {
	s.alignV = AlignTop
	return s
}

// AlignMiddle sets vertical alignment to AlignMiddle (centered vertically).
func (s Style) AlignMiddle() Style {
	s.alignV = AlignMiddle
	return s
}

// AlignBottom sets vertical alignment to AlignBottom.
func (s Style) AlignBottom() Style {
	s.alignV = AlignBottom
	return s
}

// GetFg returns the foreground color.
func (s Style) GetFg() cell.Color { return s.fg }

// GetBg returns the background color.
func (s Style) GetBg() cell.Color { return s.bg }

// GetModifier returns the text modifier bitmask.
func (s Style) GetModifier() cell.Modifier { return s.modifier }

// GetBorderFg returns the border foreground color.
func (s Style) GetBorderFg() cell.Color { return s.borderFg }

// GetBorderBg returns the border background color.
func (s Style) GetBorderBg() cell.Color { return s.borderBg }

// Draw renders text styled according to this Style directly into buf within area.
// Generates zero intermediate string concats or heap buffers.
func (s Style) Draw(buf *buffer.Buffer, area buffer.Rect, text string) {
	if area.IsEmpty() {
		return
	}

	// 1. Apply Margins
	mArea := buffer.NewRect(
		area.X+s.marginLeft,
		area.Y+s.marginTop,
		area.Width-(s.marginLeft+s.marginRight),
		area.Height-(s.marginTop+s.marginBot),
	)
	if mArea.IsEmpty() {
		return
	}

	// 2. Fill background if set
	if !s.bg.IsDefault() {
		bgCell := cell.Cell{
			Rune:     ' ',
			Width:    1,
			FgType:   s.fg.Type,
			BgType:   s.bg.Type,
			Fg:       s.fg.Value,
			Bg:       s.bg.Value,
			Modifier: s.modifier,
		}
		buf.Fill(mArea, bgCell)
	}

	contentArea := mArea

	// 3. Draw Borders if enabled
	if s.hasBorder && mArea.Width >= 2 && mArea.Height >= 2 {
		s.drawBorder(buf, mArea)
		contentArea = buffer.NewRect(mArea.X+1, mArea.Y+1, mArea.Width-2, mArea.Height-2)
	}

	// 4. Apply Padding
	pArea := buffer.NewRect(
		contentArea.X+s.padLeft,
		contentArea.Y+s.padTop,
		contentArea.Width-(s.padLeft+s.padRight),
		contentArea.Height-(s.padTop+s.padBottom),
	)
	if pArea.IsEmpty() || text == "" {
		return
	}

	// 5. Render Text lines with zero-alloc string slicing
	totalLines := 1
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			totalLines++
		}
	}

	startY := pArea.Y
	switch s.alignV {
	case AlignMiddle:
		startY += max(0, (pArea.Height-totalLines)/2)
	case AlignBottom:
		startY += max(0, pArea.Height-totalLines)
	}

	start := 0
	lineNum := 0
	for start <= len(text) {
		end := strings.IndexByte(text[start:], '\n')
		var line string
		if end == -1 {
			line = text[start:]
			start = len(text) + 1 // terminate
		} else {
			line = text[start : start+end]
			start += end + 1
		}

		currY := startY + lineNum
		lineNum++

		if currY >= pArea.Bottom() {
			break
		}
		if currY < pArea.Y {
			continue
		}

		lineArea := buffer.NewRect(pArea.X, currY, pArea.Width, 1)
		buf.SetStringAligned(lineArea, line, s.alignH, s.fg, s.bg, s.modifier)
	}
}

func (s Style) drawBorder(buf *buffer.Buffer, r buffer.Rect) {
	if r.Width < 2 || r.Height < 2 {
		return
	}

	b := s.border
	fg := s.borderFg
	bg := s.borderBg

	// Horizontal edges
	for x := r.X + 1; x < r.Right()-1; x++ {
		setBoxCell(buf, x, r.Y, b.Top, fg, bg)
		setBoxCell(buf, x, r.Bottom()-1, b.Bottom, fg, bg)
	}

	if s.title != "" && r.Width >= 6 {
		t := strings.TrimSpace(s.title)
		titleStr := " " + t + " "
		maxW := r.Width - 4
		if buffer.StringWidth(titleStr) > maxW && maxW > 3 {
			// Truncate to fit
			runes := []rune(titleStr)
			for len(runes) > 0 && buffer.StringWidth(string(runes)) > maxW-1 {
				runes = runes[:len(runes)-1]
			}
			titleStr = string(runes) + "…"
		}
		if buffer.StringWidth(titleStr) <= maxW {
			buf.SetString(r.X+2, r.Y, titleStr, fg, bg, cell.AttrNone)
		}
	}

	// Vertical edges
	for y := r.Y + 1; y < r.Bottom()-1; y++ {
		setBoxCell(buf, r.X, y, b.Left, fg, bg)
		setBoxCell(buf, r.Right()-1, y, b.Right, fg, bg)
	}

	// Corners
	setBoxCell(buf, r.X, r.Y, b.TopLeft, fg, bg)
	setBoxCell(buf, r.Right()-1, r.Y, b.TopRight, fg, bg)
	setBoxCell(buf, r.X, r.Bottom()-1, b.BottomLeft, fg, bg)
	setBoxCell(buf, r.Right()-1, r.Bottom()-1, b.BottomRight, fg, bg)
}

func setBoxCell(buf *buffer.Buffer, x, y int, r rune, fg, bg cell.Color) {
	existing := buf.Cell(x, y)
	targetRune := r
	if existing != nil && existing.Rune != ' ' && existing.Rune != 0 {
		targetRune = MergeBoxRunes(existing.Rune, r)
	}
	buf.Set(x, y, cell.Cell{
		Rune:     targetRune,
		Width:    1,
		FgType:   fg.Type,
		BgType:   bg.Type,
		Fg:       fg.Value,
		Bg:       bg.Value,
		Modifier: cell.AttrNone,
	})
}
