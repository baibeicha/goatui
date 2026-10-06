package renderer

import (
	"fmt"
	"io"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// CursorShape defines hardware terminal cursor shapes via DECSCUSR escape sequences.
type CursorShape int

const (
	CursorDefault         CursorShape = 0
	CursorBlinkBlock      CursorShape = 1
	CursorSteadyBlock     CursorShape = 2
	CursorBlinkUnderline  CursorShape = 3
	CursorSteadyUnderline CursorShape = 4
	CursorBlinkBar        CursorShape = 5
	CursorSteadyBar       CursorShape = 6
)

// Renderer orchestrates double-buffered terminal rendering with dirty-cell diffing,
// minimal ANSI escape emissions, and synchronized output (Mode 2026).
type Renderer struct {
	front         *buffer.Buffer
	back          *buffer.Buffer
	sgr           sgrState
	outBuf        []byte
	curX          int
	curY          int
	syncOutput    bool
	cursorVisible bool
	cursorX       int
	cursorY       int
	cursorShape   CursorShape
}

// NewRenderer initializes a double-buffered renderer for the specified dimensions.
func NewRenderer(width, height int) *Renderer {
	initCap := width * height * 4
	if initCap < 4096 {
		initCap = 4096
	}

	r := &Renderer{
		front:      buffer.NewBuffer(width, height),
		back:       buffer.NewBuffer(width, height),
		outBuf:     make([]byte, 0, initCap),
		curX:       -1,
		curY:       -1,
		syncOutput: true, // Default to true for tear-free rendering on modern terminals
	}

	// Invalidate front buffer initially so first frame renders all cells completely
	for i := range r.front.Cells() {
		r.front.Cells()[i] = cell.Cell{Rune: 0xFFFF}
	}

	return r
}

// ShowCursor configures terminal hardware cursor position and DECSCUSR shape for the current frame.
func (r *Renderer) ShowCursor(x, y int, shape CursorShape) {
	r.cursorVisible = true
	r.cursorX = x
	r.cursorY = y
	r.cursorShape = shape
}

// HideCursor hides the terminal hardware cursor for the current frame.
func (r *Renderer) HideCursor() {
	r.cursorVisible = false
}

// SetSynchronized enables or disables Mode 2026 synchronized output framing.
func (r *Renderer) SetSynchronized(enabled bool) {
	r.syncOutput = enabled
}

// Resize resizes both front and back buffers to new dimensions.
func (r *Renderer) Resize(width, height int) {
	r.front.Resize(width, height)
	r.back.Resize(width, height)

	// Invalidate front buffer to force full redraw
	for i := range r.front.Cells() {
		r.front.Cells()[i] = cell.Cell{Rune: 0xFFFF}
	}

	// Force complete redraw on next render
	r.curX = -1
	r.curY = -1
}

// Back returns the writable back buffer where models and widgets draw.
func (r *Renderer) Back() *buffer.Buffer {
	return r.back
}

// Front returns the currently displayed front buffer.
func (r *Renderer) Front() *buffer.Buffer {
	return r.front
}

// Render diffs the back buffer against the front buffer, generates optimized ANSI escape
// sequences into an internal reusable buffer, and flushes them to out.
// Once written, front buffer is synchronized with back buffer.
func (r *Renderer) Render(out io.Writer) error {
	width := r.back.Width()
	height := r.back.Height()

	if width <= 0 || height <= 0 {
		return nil
	}

	// Reset scratch byte buffer length without reallocating underlying slice
	r.outBuf = r.outBuf[:0]

	// Begin Synchronized Output (Mode 2026)
	if r.syncOutput {
		r.outBuf = append(r.outBuf, "\x1b[?2026h"...)
	}

	frontCells := r.front.Cells()
	backCells := r.back.Cells()
	totalCells := width * height

	if len(frontCells) != totalCells || len(backCells) != totalCells {
		r.Resize(width, height)
		frontCells = r.front.Cells()
		backCells = r.back.Cells()
	}

	changesCount := 0
	r.curX = -1
	r.curY = -1

	for y := 0; y < height; y++ {
		rowOffset := y * width
		x := 0

		for x < width {
			// Prevent terminal auto-scroll on bottom-right corner on Windows/ConPTY
			if y == height-1 && x == width-1 {
				break
			}

			idx := rowOffset + x
			bCell := backCells[idx]
			fCell := frontCells[idx]

			// If cells are identical, skip
			if fCell.Equal(bCell) {
				x++
				continue
			}

			// Continuation cell of a wide character: already handled by leading cell
			if bCell.Width == 0 {
				x++
				continue
			}

			changesCount++

			// Position cursor at (x, y)
			r.outBuf = appendCursorMove(r.outBuf, r.curX, r.curY, x, y)
			r.curX = x
			r.curY = y

			// Emit necessary style and color changes
			r.outBuf = emitSGR(r.outBuf, &r.sgr, bCell)

			// Output rune
			r.outBuf = appendRune(r.outBuf, bCell.Rune)
			w := int(bCell.Width)
			if w <= 0 {
				w = 1
			}
			r.curX += w
			x += w
		}
	}

	// If there were modifications, reset SGR at end of frame to avoid style or color bleeding
	if changesCount > 0 && (r.sgr.modifier != cell.AttrNone || r.sgr.fgType != cell.ColorDefault || r.sgr.bgType != cell.ColorDefault) {
		r.outBuf = append(r.outBuf, "\x1b[0m"...)
		r.sgr.reset()
	}

	// Hardware cursor positioning and shape (DECSCUSR)
	if r.cursorVisible && r.cursorX >= 0 && r.cursorY >= 0 {
		if r.cursorShape > 0 {
			r.outBuf = append(r.outBuf, fmt.Sprintf("\x1b[%d q", r.cursorShape)...)
		}
		r.outBuf = appendCursorMove(r.outBuf, r.curX, r.curY, r.cursorX, r.cursorY)
		r.curX = r.cursorX
		r.curY = r.cursorY
		r.outBuf = append(r.outBuf, "\x1b[?25h"...)
	} else {
		r.outBuf = append(r.outBuf, "\x1b[?25l"...)
	}

	// End Synchronized Output (Mode 2026)
	if r.syncOutput {
		r.outBuf = append(r.outBuf, "\x1b[?2026l"...)
	}

	// Only flush to writer if we actually emitted output (or sync brackets)
	if (changesCount > 0 || r.cursorVisible) && len(r.outBuf) > 0 {
		if _, err := out.Write(r.outBuf); err != nil {
			return err
		}
	}

	// Synchronize front buffer with back buffer
	copy(frontCells, backCells)

	return nil
}
