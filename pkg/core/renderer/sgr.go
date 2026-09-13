package renderer

import (
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

// sgrState tracks the active terminal text attributes and colors to avoid emitting redundant ANSI codes.
type sgrState struct {
	modifier cell.Modifier
	fgType   cell.ColorType
	bgType   cell.ColorType
	fg       uint32
	bg       uint32
}

func (s *sgrState) reset() {
	s.modifier = cell.AttrNone
	s.fgType = cell.ColorDefault
	s.bgType = cell.ColorDefault
	s.fg = 0
	s.bg = 0
}

// emitSGR emits minimal ANSI sequences to transition from the current SGR state to the target cell's style.
// All byte appends are strictly zero-heap-allocation operations on the provided byte slice.
func emitSGR(buf []byte, cur *sgrState, target cell.Cell) []byte {
	// If current modifier has attributes that the target cell does NOT have, we must reset with SGR 0
	if cur.modifier != target.Modifier && (cur.modifier &^ target.Modifier) != 0 {
		buf = append(buf, "\x1b[0m"...)
		cur.reset()
	}

	// Apply modifiers
	if cur.modifier != target.Modifier {
		diff := target.Modifier &^ cur.modifier
		if diff.Has(cell.AttrBold) {
			buf = append(buf, "\x1b[1m"...)
		}
		if diff.Has(cell.AttrDim) {
			buf = append(buf, "\x1b[2m"...)
		}
		if diff.Has(cell.AttrItalic) {
			buf = append(buf, "\x1b[3m"...)
		}
		if diff.Has(cell.AttrUnderline) {
			buf = append(buf, "\x1b[4m"...)
		}
		if diff.Has(cell.AttrBlink) {
			buf = append(buf, "\x1b[5m"...)
		}
		if diff.Has(cell.AttrReverse) {
			buf = append(buf, "\x1b[7m"...)
		}
		if diff.Has(cell.AttrHidden) {
			buf = append(buf, "\x1b[8m"...)
		}
		if diff.Has(cell.AttrStrikethrough) {
			buf = append(buf, "\x1b[9m"...)
		}
		cur.modifier = target.Modifier
	}

	// Apply Foreground color
	if cur.fgType != target.FgType || cur.fg != target.Fg {
		buf = emitColor(buf, true, target.FgType, target.Fg)
		cur.fgType = target.FgType
		cur.fg = target.Fg
	}

	// Apply Background color
	if cur.bgType != target.BgType || cur.bg != target.Bg {
		buf = emitColor(buf, false, target.BgType, target.Bg)
		cur.bgType = target.BgType
		cur.bg = target.Bg
	}

	return buf
}

func emitColor(buf []byte, isFg bool, cType cell.ColorType, val uint32) []byte {
	switch cType {
	case cell.ColorDefault:
		if isFg {
			return append(buf, "\x1b[39m"...)
		}
		return append(buf, "\x1b[49m"...)

	case cell.ColorANSI16:
		idx := uint8(val & 0x0F)
		if idx < 8 {
			base := 30
			if !isFg {
				base = 40
			}
			buf = append(buf, "\x1b["...)
			buf = appendUint(buf, base+int(idx))
			return append(buf, 'm')
		} else {
			base := 90
			if !isFg {
				base = 100
			}
			buf = append(buf, "\x1b["...)
			buf = appendUint(buf, base+int(idx-8))
			return append(buf, 'm')
		}

	case cell.ColorANSI256:
		if isFg {
			buf = append(buf, "\x1b[38;5;"...)
		} else {
			buf = append(buf, "\x1b[48;5;"...)
		}
		buf = appendUint(buf, int(val&0xFF))
		return append(buf, 'm')

	case cell.ColorRGB:
		r := int((val >> 16) & 0xFF)
		g := int((val >> 8) & 0xFF)
		b := int(val & 0xFF)
		if isFg {
			buf = append(buf, "\x1b[38;2;"...)
		} else {
			buf = append(buf, "\x1b[48;2;"...)
		}
		buf = appendUint(buf, r)
		buf = append(buf, ';')
		buf = appendUint(buf, g)
		buf = append(buf, ';')
		buf = appendUint(buf, b)
		return append(buf, 'm')
	}

	return buf
}

// appendUint appends the decimal representation of n to dst without allocations.
func appendUint(dst []byte, n int) []byte {
	if n == 0 {
		return append(dst, '0')
	}

	var digits [10]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + (n % 10))
		n /= 10
	}
	return append(dst, digits[i:]...)
}

// appendRune appends UTF-8 bytes of r to dst without heap allocation.
func appendRune(dst []byte, r rune) []byte {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return append(dst, buf[:n]...)
}

// appendCursorMove emits the most compact ANSI cursor movement sequence to reach (targetX, targetY).
func appendCursorMove(buf []byte, curX, curY, targetX, targetY int) []byte {
	if curX == targetX && curY == targetY {
		return buf
	}

	// Same line relative movement forward
	if curY == targetY && targetX > curX && targetX-curX <= 4 {
		dist := targetX - curX
		if dist == 1 {
			return append(buf, "\x1b[C"...)
		}
		buf = append(buf, "\x1b["...)
		buf = appendUint(buf, dist)
		return append(buf, 'C')
	}

	// Absolute positioning CUP: \x1b[row;colH (1-indexed)
	buf = append(buf, "\x1b["...)
	buf = appendUint(buf, targetY+1)
	buf = append(buf, ';')
	buf = appendUint(buf, targetX+1)
	return append(buf, 'H')
}
