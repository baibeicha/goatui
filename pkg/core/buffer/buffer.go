package buffer

import (
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Buffer is a 2D grid of terminal cells stored as a contiguous 1D slice.
// Flat memory layout maximizes CPU L1/L2 cache locality and eliminates pointer chasing.
type Buffer struct {
	cells    []cell.Cell
	area     Rect
	width    int
	height   int
	clipRect *Rect
}

// NewBuffer allocates a new Buffer with the given dimensions.
func NewBuffer(width, height int) *Buffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	total := width * height
	cells := make([]cell.Cell, total)
	for i := range cells {
		cells[i] = cell.BlankCell
	}
	return &Buffer{
		cells:  cells,
		area:   NewRect(0, 0, width, height),
		width:  width,
		height: height,
	}
}

// Area returns the bounding rectangle of the buffer.
func (b *Buffer) Area() Rect {
	return b.area
}

// SetClip sets the clip rectangle
func (b *Buffer) SetClip(r Rect) {
	b.clipRect = &r
}

// ResetClip clears the clip (sets to nil)
func (b *Buffer) ResetClip() {
	b.clipRect = nil
}

// ClipRect returns the current clip rect or nil
func (b *Buffer) ClipRect() *Rect {
	return b.clipRect
}

// Width returns the buffer's width in columns.
func (b *Buffer) Width() int {
	return b.width
}

// Height returns the buffer's height in rows.
func (b *Buffer) Height() int {
	return b.height
}

// Cells returns the underlying flat cell slice.
func (b *Buffer) Cells() []cell.Cell {
	return b.cells
}

// Resize resizes the buffer. If the new capacity fits in the existing slice,
// it reuses the memory without heap allocations.
func (b *Buffer) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	if b.width == width && b.height == height {
		return
	}

	total := width * height
	if cap(b.cells) >= total {
		b.cells = b.cells[:total]
	} else {
		b.cells = make([]cell.Cell, total)
	}

	b.width = width
	b.height = height
	b.area = NewRect(0, 0, width, height)
	b.Reset()
}

// Reset clears the buffer, restoring every cell to BlankCell.
func (b *Buffer) Reset() {
	for i := range b.cells {
		b.cells[i] = cell.BlankCell
	}
}

// InBounds returns true if (x, y) is inside the buffer boundaries.
func (b *Buffer) InBounds(x, y int) bool {
	return x >= 0 && x < b.width && y >= 0 && y < b.height
}

// Index calculates the 1D slice index for coordinates (x, y).
// Callers should verify InBounds beforehand.
func (b *Buffer) Index(x, y int) int {
	return y*b.width + x
}

// Cell returns a pointer to the cell at (x, y), or nil if out of bounds.
func (b *Buffer) Cell(x, y int) *cell.Cell {
	if !b.InBounds(x, y) {
		return nil
	}
	return &b.cells[y*b.width+x]
}

// Set places a cell at (x, y) if within bounds.
func (b *Buffer) Set(x, y int, c cell.Cell) {
	if !b.InBounds(x, y) {
		return
	}
	if b.clipRect != nil && !b.clipRect.Contains(x, y) {
		return
	}
	b.cells[y*b.width+x] = c
}

// SetRune places a rune at (x, y) with the specified colors and modifiers.
func (b *Buffer) SetRune(x, y int, r rune, fg, bg cell.Color, mod cell.Modifier) {
	if !b.InBounds(x, y) {
		return
	}
	if b.clipRect != nil && !b.clipRect.Contains(x, y) {
		return
	}
	w := uint8(RuneWidth(r))
	if w == 0 && r != 0 {
		w = 1
	}

	idx := y*b.width + x
	b.cells[idx] = cell.Cell{
		Rune:     r,
		Width:    w,
		Modifier: mod,
		FgType:   fg.Type,
		BgType:   bg.Type,
		Fg:       fg.Value,
		Bg:       bg.Value,
	}

	// If wide character, mark continuation cell to the right
	if w == 2 && x+1 < b.width && (b.clipRect == nil || b.clipRect.Contains(x+1, y)) {
		b.cells[idx+1] = cell.Cell{
			Rune:     0,
			Width:    0,
			Modifier: mod,
			FgType:   fg.Type,
			BgType:   bg.Type,
			Fg:       fg.Value,
			Bg:       bg.Value,
		}
	}
}

// SetString draws a UTF-8 string horizontally starting at (x, y), clipped to buffer width.
// Returns the X coordinate where the next character would be drawn.
func (b *Buffer) SetString(x, y int, s string, fg, bg cell.Color, mod cell.Modifier) int {
	if y < 0 || y >= b.height {
		return x
	}

	currX := x
	for _, r := range s {
		if currX >= b.width {
			break
		}
		if currX < 0 {
			w := RuneWidth(r)
			currX += w
			continue
		}

		w := RuneWidth(r)
		if w == 0 {
			continue // Skip zero-width combining marks for position
		}

		if b.clipRect != nil && !b.clipRect.Contains(currX, y) {
			currX += w
			continue
		}

		if w == 2 && currX+1 >= b.width {
			// Wide character does not fit in remaining column: pad with space
			b.Set(currX, y, cell.Cell{
				Rune:     ' ',
				Width:    1,
				Modifier: mod,
				FgType:   fg.Type,
				BgType:   bg.Type,
				Fg:       fg.Value,
				Bg:       bg.Value,
			})
			currX++
			break
		}

		idx := y*b.width + currX
		b.cells[idx] = cell.Cell{
			Rune:     r,
			Width:    uint8(w),
			Modifier: mod,
			FgType:   fg.Type,
			BgType:   bg.Type,
			Fg:       fg.Value,
			Bg:       bg.Value,
		}

		if w == 2 {
			if b.clipRect == nil || b.clipRect.Contains(currX+1, y) {
				b.cells[idx+1] = cell.Cell{
					Rune:     0,
					Width:    0,
					Modifier: mod,
					FgType:   fg.Type,
					BgType:   bg.Type,
					Fg:       fg.Value,
					Bg:       bg.Value,
				}
			}
			currX += 2
		} else {
			currX++
		}
	}
	return currX
}

// Fill fills the given rectangular area with the specified cell.
func (b *Buffer) Fill(rect Rect, c cell.Cell) {
	intersect := b.area.Intersection(rect)
	if b.clipRect != nil {
		intersect = intersect.Intersection(*b.clipRect)
	}
	if intersect.IsEmpty() {
		return
	}

	for y := intersect.Y; y < intersect.Bottom(); y++ {
		rowStart := y * b.width
		for x := intersect.X; x < intersect.Right(); x++ {
			b.cells[rowStart+x] = c
		}
	}
}

// Clear clears the given rectangle back to BlankCell.
func (b *Buffer) Clear(rect Rect) {
	b.Fill(rect, cell.BlankCell)
}

// CopyFrom copies all cells from another buffer with identical dimensions.
func (b *Buffer) CopyFrom(src *Buffer) {
	if b.width != src.width || b.height != src.height {
		b.Resize(src.width, src.height)
	}
	copy(b.cells, src.cells)
}
