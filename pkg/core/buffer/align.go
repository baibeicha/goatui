package buffer

import "github.com/baibeicha/goatui/pkg/core/cell"

// Alignment defines horizontal text alignment.
type Alignment uint8

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// VerticalAlignment defines vertical text alignment.
type VerticalAlignment uint8

const (
	AlignTop VerticalAlignment = iota
	AlignMiddle
	AlignBottom
)

// SetStringAligned draws a single line of text aligned horizontally inside area, with clipping and optional ellipsis.
// Returns the X coordinate where drawing finished.
func (b *Buffer) SetStringAligned(area Rect, s string, align Alignment, fg, bg cell.Color, mod cell.Modifier) int {
	if area.IsEmpty() || area.Y < 0 || area.Y >= b.height {
		return area.X
	}

	textW := StringWidth(s)
	availW := area.Width

	if textW <= availW {
		startX := area.X
		switch align {
		case AlignRight:
			startX = area.Right() - textW
		case AlignCenter:
			startX = area.X + (availW-textW)/2
		}
		return b.SetString(startX, area.Y, s, fg, bg, mod)
	}

	// Truncate if string exceeds area width
	maxFit := availW - 1
	if maxFit < 0 {
		maxFit = 0
	}
	currX := area.X
	fittedW := 0
	for _, r := range s {
		rw := RuneWidth(r)
		if rw == 0 {
			continue
		}
		if fittedW+rw > maxFit {
			break
		}
		b.SetRune(currX, area.Y, r, fg, bg, mod)
		currX += rw
		fittedW += rw
	}
	if maxFit >= 0 && currX < area.Right() {
		b.SetRune(currX, area.Y, '…', fg, bg, mod)
		currX++
	}
	return currX
}

// DrawAlignedText draws text with both horizontal and vertical alignment within area, supporting multiline strings.
func (b *Buffer) DrawAlignedText(area Rect, s string, hAlign Alignment, vAlign VerticalAlignment, fg, bg cell.Color, mod cell.Modifier) {
	if area.IsEmpty() {
		return
	}

	// Count lines
	totalLines := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			totalLines++
		}
	}

	startY := area.Y
	switch vAlign {
	case AlignMiddle:
		startY += max(0, (area.Height-totalLines)/2)
	case AlignBottom:
		startY += max(0, area.Height-totalLines)
	}

	start := 0
	lineIdx := 0
	for start <= len(s) {
		end := -1
		for i := start; i < len(s); i++ {
			if s[i] == '\n' {
				end = i
				break
			}
		}

		var line string
		if end == -1 {
			line = s[start:]
			start = len(s) + 1
		} else {
			line = s[start:end]
			start = end + 1
		}

		currY := startY + lineIdx
		lineIdx++

		if currY >= area.Bottom() {
			break
		}
		if currY >= area.Y {
			lineArea := NewRect(area.X, currY, area.Width, 1)
			b.SetStringAligned(lineArea, line, hAlign, fg, bg, mod)
		}
	}
}
