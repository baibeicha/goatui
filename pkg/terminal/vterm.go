package terminal

import (
	"sync"
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// VTCell represents a single styled character cell within the virtual terminal grid.
type VTCell struct {
	Char rune
	Fg   cell.Color
	Bg   cell.Color
	Attr cell.Modifier
}

type vtermState int

const (
	stateGround vtermState = iota
	stateEsc
	stateCSI
	stateOSC
	stateCharset
)

var standard16Colors = [16]cell.Color{
	cell.RGB(0, 0, 0),       // 0: Black
	cell.RGB(205, 49, 49),   // 1: Red
	cell.RGB(13, 188, 121),  // 2: Green
	cell.RGB(229, 229, 16),  // 3: Yellow
	cell.RGB(36, 114, 200),  // 4: Blue
	cell.RGB(188, 63, 188),  // 5: Magenta
	cell.RGB(17, 168, 205),  // 6: Cyan
	cell.RGB(229, 229, 229), // 7: White
	cell.RGB(102, 102, 102), // 8: Bright Black (Gray)
	cell.RGB(241, 76, 76),   // 9: Bright Red
	cell.RGB(35, 209, 139),  // 10: Bright Green
	cell.RGB(245, 245, 67),  // 11: Bright Yellow
	cell.RGB(59, 142, 234),  // 12: Bright Blue
	cell.RGB(214, 112, 214), // 13: Bright Magenta
	cell.RGB(41, 184, 219),  // 14: Bright Cyan
	cell.RGB(255, 255, 255), // 15: Bright White
}

func colorFrom256(idx int) cell.Color {
	if idx < 0 || idx > 255 {
		return cell.DefaultColor()
	}
	if idx < 16 {
		return standard16Colors[idx]
	}
	if idx >= 232 {
		v := uint8(8 + (idx-232)*10)
		return cell.RGB(v, v, v)
	}
	idx -= 16
	b := uint8(idx % 6)
	idx /= 6
	g := uint8(idx % 6)
	idx /= 6
	r := uint8(idx % 6)
	steps := []uint8{0, 95, 135, 175, 215, 255}
	return cell.RGB(steps[r], steps[g], steps[b])
}

// VTerm is a virtual terminal screen emulator with ANSI escape sequence decoding.
type VTerm struct {
	mu            sync.Mutex
	Cols          int
	Rows          int
	Screen        [][]VTCell
	Scrollback    [][]VTCell
	MaxScrollback int
	ScrollOffset  int // 0 = viewing active screen, >0 = viewing history

	CursorX       int
	CursorY       int
	CursorVisible bool
	SavedX        int
	SavedY        int

	TopMargin    int
	BottomMargin int

	CurFg   cell.Color
	CurBg   cell.Color
	CurAttr cell.Modifier
	DefFg   cell.Color
	DefBg   cell.Color

	// Alternate screen buffer (e.g. vim, htop, less)
	mainScreen  [][]VTCell
	inAltScreen bool
	altSavedX   int
	altSavedY   int

	// Parser state machine
	state      vtermState
	csiParams  []int
	csiPrivate bool
	oscBuf     []byte
}

// NewVTerm creates a new virtual terminal emulator with the given grid dimensions.
func NewVTerm(cols, rows int) *VTerm {
	if cols < 1 {
		cols = 80
	}
	if rows < 1 {
		rows = 24
	}

	v := &VTerm{
		Cols:          cols,
		Rows:          rows,
		MaxScrollback: 5000,
		CursorVisible: true,
		TopMargin:     0,
		BottomMargin:  rows - 1,
		CurFg:         cell.DefaultColor(),
		CurBg:         cell.DefaultColor(),
		DefFg:         cell.DefaultColor(),
		DefBg:         cell.DefaultColor(),
	}

	v.Screen = v.allocScreen(cols, rows)
	return v
}

func (v *VTerm) allocScreen(cols, rows int) [][]VTCell {
	s := make([][]VTCell, rows)
	for y := 0; y < rows; y++ {
		s[y] = make([]VTCell, cols)
		for x := 0; x < cols; x++ {
			s[y][x] = VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		}
	}
	return s
}

// Resize dynamically resizes the virtual terminal grid.
func (v *VTerm) Resize(cols, rows int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if cols < 1 || rows < 1 || (cols == v.Cols && rows == v.Rows) {
		return
	}

	newScreen := make([][]VTCell, rows)
	for y := 0; y < rows; y++ {
		newScreen[y] = make([]VTCell, cols)
		for x := 0; x < cols; x++ {
			if y < v.Rows && x < v.Cols {
				newScreen[y][x] = v.Screen[y][x]
			} else {
				newScreen[y][x] = VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
			}
		}
	}

	v.Screen = newScreen
	v.Cols = cols
	v.Rows = rows
	v.TopMargin = 0
	v.BottomMargin = rows - 1

	if v.CursorX >= cols {
		v.CursorX = cols - 1
	}
	if v.CursorY >= rows {
		v.CursorY = rows - 1
	}
}

// Clear resets the screen and clears the scrollback.
func (v *VTerm) Clear() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Scrollback = nil
	v.ScrollOffset = 0
	for y := 0; y < v.Rows; y++ {
		for x := 0; x < v.Cols; x++ {
			v.Screen[y][x] = VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		}
	}
	v.CursorX = 0
	v.CursorY = 0
}

// Scroll adjusts the viewing offset into scrollback history.
func (v *VTerm) Scroll(delta int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.ScrollOffset += delta
	if v.ScrollOffset < 0 {
		v.ScrollOffset = 0
	}
	maxOffset := len(v.Scrollback)
	if v.ScrollOffset > maxOffset {
		v.ScrollOffset = maxOffset
	}
}

// Write feeds incoming terminal data bytes through the ANSI state machine.
func (v *VTerm) Write(p []byte) (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	i := 0
	for i < len(p) {
		b := p[i]

		switch v.state {
		case stateGround:
			if b == 0x1b { // ESC
				v.state = stateEsc
				i++
				continue
			}
			switch b {
			case '\r':
				v.CursorX = 0
				i++
			case '\n':
				v.lineFeed()
				i++
			case '\b':
				if v.CursorX > 0 {
					v.CursorX--
				}
				i++
			case '\t':
				v.CursorX = (v.CursorX/8 + 1) * 8
				if v.CursorX >= v.Cols {
					v.CursorX = v.Cols - 1
				}
				i++
			case '\a': // Bell
				i++
			case 0x00, 0x0c: // NUL, FF
				i++
			default:
				r, size := utf8.DecodeRune(p[i:])
				if r == utf8.RuneError && size <= 1 {
					r = rune(b)
					size = 1
				}
				i += size

				if r >= 32 {
					if v.CursorX >= v.Cols {
						v.CursorX = 0
						v.lineFeed()
					}
					w := buffer.RuneWidth(r)
					if w <= 0 {
						w = 1
					}
					if v.CursorY >= 0 && v.CursorY < v.Rows && v.CursorX >= 0 && v.CursorX < v.Cols {
						v.Screen[v.CursorY][v.CursorX] = VTCell{
							Char: r,
							Fg:   v.CurFg,
							Bg:   v.CurBg,
							Attr: v.CurAttr,
						}
						// Clear trailing cell for wide characters
						if w > 1 && v.CursorX+1 < v.Cols {
							v.Screen[v.CursorY][v.CursorX+1] = VTCell{
								Char: ' ',
								Fg:   v.CurFg,
								Bg:   v.CurBg,
								Attr: v.CurAttr,
							}
						}
					}
					v.CursorX += w
				}
			}

		case stateEsc:
			switch b {
			case '[':
				v.state = stateCSI
				v.csiParams = []int{0}
				v.csiPrivate = false
				i++
			case ']':
				v.state = stateOSC
				v.oscBuf = v.oscBuf[:0]
				i++
			case '(', ')':
				v.state = stateCharset
				i++
			case '7': // Save cursor
				v.SavedX = v.CursorX
				v.SavedY = v.CursorY
				v.state = stateGround
				i++
			case '8': // Restore cursor
				v.CursorX = v.SavedX
				v.CursorY = v.SavedY
				v.state = stateGround
				i++
			case 'M': // Reverse index
				if v.CursorY == v.TopMargin {
					v.scrollDown()
				} else if v.CursorY > 0 {
					v.CursorY--
				}
				v.state = stateGround
				i++
			case 'c': // Full reset (RIS)
				v.reset()
				v.state = stateGround
				i++
			default:
				v.state = stateGround
				i++
			}

		case stateCharset:
			v.state = stateGround
			i++

		case stateCSI:
			if b == '?' {
				v.csiPrivate = true
				i++
				continue
			}
			if b >= '0' && b <= '9' {
				lastIdx := len(v.csiParams) - 1
				if lastIdx >= 0 {
					v.csiParams[lastIdx] = v.csiParams[lastIdx]*10 + int(b-'0')
				}
				i++
				continue
			}
			if b == ';' || b == ':' {
				v.csiParams = append(v.csiParams, 0)
				i++
				continue
			}

			// Final command byte (0x40 - 0x7E)
			if b >= 0x40 && b <= 0x7E {
				v.handleCSI(rune(b))
				v.state = stateGround
				i++
				continue
			}

			// Intermediate byte or unknown, skip
			i++

		case stateOSC:
			if b == '\a' || (b == '\\' && len(v.oscBuf) > 0 && v.oscBuf[len(v.oscBuf)-1] == 0x1b) {
				v.state = stateGround
				i++
				continue
			}
			v.oscBuf = append(v.oscBuf, b)
			i++
		}
	}

	return len(p), nil
}

func (v *VTerm) lineFeed() {
	if v.CursorY == v.BottomMargin {
		v.scrollUp()
	} else if v.CursorY < v.Rows-1 {
		v.CursorY++
	}
}

func (v *VTerm) scrollUp() {
	if v.TopMargin == 0 && v.BottomMargin == v.Rows-1 && !v.inAltScreen {
		lineCopy := make([]VTCell, v.Cols)
		copy(lineCopy, v.Screen[0])
		v.Scrollback = append(v.Scrollback, lineCopy)
		if len(v.Scrollback) > v.MaxScrollback {
			v.Scrollback = v.Scrollback[len(v.Scrollback)-v.MaxScrollback:]
		}
	}

	for y := v.TopMargin; y < v.BottomMargin; y++ {
		copy(v.Screen[y], v.Screen[y+1])
	}
	v.clearRow(v.BottomMargin)
}

func (v *VTerm) scrollDown() {
	for y := v.BottomMargin; y > v.TopMargin; y-- {
		copy(v.Screen[y], v.Screen[y-1])
	}
	v.clearRow(v.TopMargin)
}

func (v *VTerm) clearRow(y int) {
	if y < 0 || y >= v.Rows {
		return
	}
	blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
	for x := 0; x < v.Cols; x++ {
		v.Screen[y][x] = blank
	}
}

func (v *VTerm) csiParam(idx, defaultVal int) int {
	if idx < len(v.csiParams) && v.csiParams[idx] > 0 {
		return v.csiParams[idx]
	}
	return defaultVal
}

func (v *VTerm) handleCSI(cmd rune) {
	switch cmd {
	case 'H', 'f': // Cursor Position (row, col)
		row := v.csiParam(0, 1) - 1
		col := v.csiParam(1, 1) - 1
		if row < 0 {
			row = 0
		}
		if row >= v.Rows {
			row = v.Rows - 1
		}
		if col < 0 {
			col = 0
		}
		if col >= v.Cols {
			col = v.Cols - 1
		}
		v.CursorX = col
		v.CursorY = row

	case 'A': // Cursor Up
		n := v.csiParam(0, 1)
		v.CursorY -= n
		if v.CursorY < 0 {
			v.CursorY = 0
		}

	case 'B': // Cursor Down
		n := v.csiParam(0, 1)
		v.CursorY += n
		if v.CursorY >= v.Rows {
			v.CursorY = v.Rows - 1
		}

	case 'C': // Cursor Forward
		n := v.csiParam(0, 1)
		v.CursorX += n
		if v.CursorX >= v.Cols {
			v.CursorX = v.Cols - 1
		}

	case 'D': // Cursor Backward
		n := v.csiParam(0, 1)
		v.CursorX -= n
		if v.CursorX < 0 {
			v.CursorX = 0
		}

	case 'E': // Cursor Next Line
		n := v.csiParam(0, 1)
		v.CursorY += n
		if v.CursorY >= v.Rows {
			v.CursorY = v.Rows - 1
		}
		v.CursorX = 0

	case 'F': // Cursor Prev Line
		n := v.csiParam(0, 1)
		v.CursorY -= n
		if v.CursorY < 0 {
			v.CursorY = 0
		}
		v.CursorX = 0

	case 'G', '`': // Cursor Horizontal Absolute
		col := v.csiParam(0, 1) - 1
		if col < 0 {
			col = 0
		}
		if col >= v.Cols {
			col = v.Cols - 1
		}
		v.CursorX = col

	case 'd': // Line Position Absolute
		row := v.csiParam(0, 1) - 1
		if row < 0 {
			row = 0
		}
		if row >= v.Rows {
			row = v.Rows - 1
		}
		v.CursorY = row

	case 'J': // Erase in Display
		mode := v.csiParam(0, 0)
		blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		switch mode {
		case 0: // Erase below
			for x := v.CursorX; x < v.Cols; x++ {
				v.Screen[v.CursorY][x] = blank
			}
			for y := v.CursorY + 1; y < v.Rows; y++ {
				v.clearRow(y)
			}
		case 1: // Erase above
			for y := 0; y < v.CursorY; y++ {
				v.clearRow(y)
			}
			for x := 0; x <= v.CursorX && x < v.Cols; x++ {
				v.Screen[v.CursorY][x] = blank
			}
		case 2, 3: // Erase all
			for y := 0; y < v.Rows; y++ {
				v.clearRow(y)
			}
			if mode == 3 {
				v.Scrollback = nil
			}
		}

	case 'K': // Erase in Line
		mode := v.csiParam(0, 0)
		blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		switch mode {
		case 0: // Erase from cursor to end
			for x := v.CursorX; x < v.Cols; x++ {
				v.Screen[v.CursorY][x] = blank
			}
		case 1: // Erase start to cursor
			for x := 0; x <= v.CursorX && x < v.Cols; x++ {
				v.Screen[v.CursorY][x] = blank
			}
		case 2: // Erase entire line
			v.clearRow(v.CursorY)
		}

	case 'L': // Insert Line
		n := v.csiParam(0, 1)
		for i := 0; i < n; i++ {
			for y := v.BottomMargin; y > v.CursorY; y-- {
				copy(v.Screen[y], v.Screen[y-1])
			}
			v.clearRow(v.CursorY)
		}

	case 'M': // Delete Line
		n := v.csiParam(0, 1)
		for i := 0; i < n; i++ {
			for y := v.CursorY; y < v.BottomMargin; y++ {
				copy(v.Screen[y], v.Screen[y+1])
			}
			v.clearRow(v.BottomMargin)
		}

	case 'P': // Delete Character
		n := v.csiParam(0, 1)
		blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		for x := v.CursorX; x < v.Cols-n; x++ {
			v.Screen[v.CursorY][x] = v.Screen[v.CursorY][x+n]
		}
		for x := max(v.CursorX, v.Cols-n); x < v.Cols; x++ {
			v.Screen[v.CursorY][x] = blank
		}

	case '@': // Insert Character
		n := v.csiParam(0, 1)
		blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		for x := v.Cols - 1; x >= v.CursorX+n; x-- {
			v.Screen[v.CursorY][x] = v.Screen[v.CursorY][x-n]
		}
		for x := v.CursorX; x < min(v.Cols, v.CursorX+n); x++ {
			v.Screen[v.CursorY][x] = blank
		}

	case 'X': // Erase Character
		n := v.csiParam(0, 1)
		blank := VTCell{Char: ' ', Fg: v.CurFg, Bg: v.CurBg, Attr: cell.AttrNone}
		for x := v.CursorX; x < min(v.Cols, v.CursorX+n); x++ {
			v.Screen[v.CursorY][x] = blank
		}

	case 'S': // Scroll Up
		n := v.csiParam(0, 1)
		for i := 0; i < n; i++ {
			v.scrollUp()
		}

	case 'T': // Scroll Down
		n := v.csiParam(0, 1)
		for i := 0; i < n; i++ {
			v.scrollDown()
		}

	case 's': // Save Cursor
		v.SavedX = v.CursorX
		v.SavedY = v.CursorY

	case 'u': // Restore Cursor
		v.CursorX = v.SavedX
		v.CursorY = v.SavedY

	case 'r': // Set Top and Bottom Margins
		top := v.csiParam(0, 1) - 1
		bottom := v.csiParam(1, v.Rows) - 1
		if top < 0 {
			top = 0
		}
		if bottom >= v.Rows {
			bottom = v.Rows - 1
		}
		if top < bottom {
			v.TopMargin = top
			v.BottomMargin = bottom
			v.CursorX = 0
			v.CursorY = 0
		}

	case 'h': // Set Mode
		if v.csiPrivate {
			param := v.csiParam(0, 0)
			switch param {
			case 25:
				v.CursorVisible = true
			case 1049: // Enter alternate screen buffer
				if !v.inAltScreen {
					v.inAltScreen = true
					v.altSavedX = v.CursorX
					v.altSavedY = v.CursorY
					v.mainScreen = v.Screen
					v.Screen = v.allocScreen(v.Cols, v.Rows)
					v.CursorX = 0
					v.CursorY = 0
				}
			}
		}

	case 'l': // Reset Mode
		if v.csiPrivate {
			param := v.csiParam(0, 0)
			switch param {
			case 25:
				v.CursorVisible = false
			case 1049: // Exit alternate screen buffer
				if v.inAltScreen {
					v.inAltScreen = false
					if v.mainScreen != nil {
						v.Screen = v.mainScreen
						v.mainScreen = nil
					}
					v.CursorX = v.altSavedX
					v.CursorY = v.altSavedY
				}
			}
		}

	case 'm': // SGR (Select Graphic Rendition)
		v.handleSGR()
	}
}

func (v *VTerm) handleSGR() {
	if len(v.csiParams) == 0 {
		v.csiParams = []int{0}
	}

	for i := 0; i < len(v.csiParams); i++ {
		p := v.csiParams[i]
		switch p {
		case 0:
			v.CurFg = cell.DefaultColor()
			v.CurBg = cell.DefaultColor()
			v.CurAttr = cell.AttrNone
		case 1:
			v.CurAttr |= cell.AttrBold
		case 2:
			v.CurAttr |= cell.AttrDim
		case 3:
			v.CurAttr |= cell.AttrItalic
		case 4:
			v.CurAttr |= cell.AttrUnderline
		case 5, 6:
			v.CurAttr |= cell.AttrBlink
		case 7:
			v.CurAttr |= cell.AttrReverse
		case 8:
			v.CurAttr |= cell.AttrHidden
		case 9:
			v.CurAttr |= cell.AttrStrikethrough
		case 22:
			v.CurAttr &^= (cell.AttrBold | cell.AttrDim)
		case 23:
			v.CurAttr &^= cell.AttrItalic
		case 24:
			v.CurAttr &^= cell.AttrUnderline
		case 25:
			v.CurAttr &^= cell.AttrBlink
		case 27:
			v.CurAttr &^= cell.AttrReverse
		case 28:
			v.CurAttr &^= cell.AttrHidden
		case 29:
			v.CurAttr &^= cell.AttrStrikethrough
		case 30, 31, 32, 33, 34, 35, 36, 37:
			v.CurFg = standard16Colors[p-30]
		case 39:
			v.CurFg = cell.DefaultColor()
		case 40, 41, 42, 43, 44, 45, 46, 47:
			v.CurBg = standard16Colors[p-40]
		case 49:
			v.CurBg = cell.DefaultColor()
		case 90, 91, 92, 93, 94, 95, 96, 97:
			v.CurFg = standard16Colors[8+(p-90)]
		case 100, 101, 102, 103, 104, 105, 106, 107:
			v.CurBg = standard16Colors[8+(p-100)]
		case 38: // Extended FG: 38;5;n or 38;2;r;g;b
			if i+2 < len(v.csiParams) && v.csiParams[i+1] == 5 {
				v.CurFg = colorFrom256(v.csiParams[i+2])
				i += 2
			} else if i+4 < len(v.csiParams) && v.csiParams[i+1] == 2 {
				v.CurFg = cell.RGB(uint8(v.csiParams[i+2]), uint8(v.csiParams[i+3]), uint8(v.csiParams[i+4]))
				i += 4
			}
		case 48: // Extended BG: 48;5;n or 48;2;r;g;b
			if i+2 < len(v.csiParams) && v.csiParams[i+1] == 5 {
				v.CurBg = colorFrom256(v.csiParams[i+2])
				i += 2
			} else if i+4 < len(v.csiParams) && v.csiParams[i+1] == 2 {
				v.CurBg = cell.RGB(uint8(v.csiParams[i+2]), uint8(v.csiParams[i+3]), uint8(v.csiParams[i+4]))
				i += 4
			}
		}
	}
}

func (v *VTerm) reset() {
	v.CursorX = 0
	v.CursorY = 0
	v.TopMargin = 0
	v.BottomMargin = v.Rows - 1
	v.CurFg = cell.DefaultColor()
	v.CurBg = cell.DefaultColor()
	v.CurAttr = cell.AttrNone
	v.CursorVisible = true
	for y := 0; y < v.Rows; y++ {
		v.clearRow(y)
	}
}

// Render writes the virtual terminal grid into the GoatUI buffer.
func (v *VTerm) Render(buf *buffer.Buffer, startX, startY, width, height int, defFg, defBg cell.Color, focused bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	scrollbackCount := len(v.Scrollback)
	totalVirtualRows := scrollbackCount + v.Rows

	for r := 0; r < height; r++ {
		screenY := startY + r

		// Compute which line of virtual text corresponds to this row
		idx := (totalVirtualRows - v.ScrollOffset - height) + r
		var line []VTCell
		if idx >= 0 && idx < scrollbackCount {
			line = v.Scrollback[idx]
		} else if idx >= scrollbackCount && idx < totalVirtualRows {
			line = v.Screen[idx-scrollbackCount]
		}

		for c := 0; c < width; c++ {
			screenX := startX + c
			char := ' '
			fg := defFg
			bg := defBg
			attr := cell.AttrNone

			if line != nil && c < len(line) {
				cellItem := line[c]
				char = cellItem.Char
				if char == 0 {
					char = ' '
				}
				attr = cellItem.Attr
				if !cellItem.Fg.IsDefault() {
					fg = cellItem.Fg
				}
				if !cellItem.Bg.IsDefault() {
					bg = cellItem.Bg
				}
			}

			// Render active cursor cell
			if focused && v.CursorVisible && v.ScrollOffset == 0 {
				if r == v.CursorY && c == v.CursorX {
					attr ^= cell.AttrReverse
					if char == ' ' {
						char = ' '
					}
				}
			}

			buf.SetRune(screenX, screenY, char, fg, bg, attr)
		}
	}
}

// ContentString returns the full current screen content as plaintext lines (for testing/inspection).
func (v *VTerm) ContentLines() []string {
	v.mu.Lock()
	defer v.mu.Unlock()

	lines := make([]string, v.Rows)
	for y := 0; y < v.Rows; y++ {
		runes := make([]rune, v.Cols)
		for x := 0; x < v.Cols; x++ {
			ch := v.Screen[y][x].Char
			if ch == 0 {
				ch = ' '
			}
			runes[x] = ch
		}
		lines[y] = string(runes)
	}
	return lines
}
