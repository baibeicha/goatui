package cell

import (
	"unsafe"
)

// Modifier represents text style bitflags.
type Modifier uint8

const (
	AttrNone          Modifier = 0
	AttrBold          Modifier = 1 << 0
	AttrDim           Modifier = 1 << 1
	AttrItalic        Modifier = 1 << 2
	AttrUnderline     Modifier = 1 << 3
	AttrBlink         Modifier = 1 << 4
	AttrReverse       Modifier = 1 << 5
	AttrHidden        Modifier = 1 << 6
	AttrStrikethrough Modifier = 1 << 7
)

// Has returns true if the modifier contains the given attribute.
func (m Modifier) Has(attr Modifier) bool {
	return (m & attr) == attr
}

// Add sets the given attribute.
func (m *Modifier) Add(attr Modifier) {
	*m |= attr
}

// Remove clears the given attribute.
func (m *Modifier) Remove(attr Modifier) {
	*m &^= attr
}

// ClusterFlag is set on Cell.Rune when the rune field holds an index into the grapheme cluster table.
const ClusterFlag rune = 1 << 30

// Cell is a 16-byte cache-aligned structure representing a single terminal grid cell.
// Exactly four Cell instances fit inside a standard 64-byte CPU L1 cache line.
type Cell struct {
	Rune     rune      // 4 bytes: Unicode code point or cluster index
	Width    uint8     // 1 byte: visual cell width (0: continuation, 1: normal, 2: wide)
	Modifier Modifier  // 1 byte: style bitmask
	FgType   ColorType // 1 byte: foreground color model
	BgType   ColorType // 1 byte: background color model
	Fg       uint32    // 4 bytes: 0x00RRGGBB or palette index
	Bg       uint32    // 4 bytes: 0x00RRGGBB or palette index
}

// Static assertion that Cell is strictly 16 bytes.
const _ = uint(16 - unsafe.Sizeof(Cell{}))
const _ = uint(unsafe.Sizeof(Cell{}) - 16)

// BlankCell represents an empty, default cell (a space with default styling).
var BlankCell = Cell{
	Rune:     ' ',
	Width:    1,
	Modifier: AttrNone,
	FgType:   ColorDefault,
	BgType:   ColorDefault,
}

// NewCell creates a cell with a single rune and width.
func NewCell(r rune, width uint8) Cell {
	if width == 0 && r != 0 {
		width = 1
	}
	return Cell{
		Rune:     r,
		Width:    width,
		Modifier: AttrNone,
		FgType:   ColorDefault,
		BgType:   ColorDefault,
	}
}

// Reset resets the cell to blank state.
func (c *Cell) Reset() {
	c.Rune = ' '
	c.Width = 1
	c.Modifier = AttrNone
	c.FgType = ColorDefault
	c.BgType = ColorDefault
	c.Fg = 0
	c.Bg = 0
}

// Equal returns true if two cells have identical content and styling.
// In Go, struct value equality on primitive fields compiles to fast 64-bit word comparisons.
func (c Cell) Equal(other Cell) bool {
	return c == other
}

// IsCluster returns true if this cell references an interned multi-rune cluster.
func (c Cell) IsCluster() bool {
	return (c.Rune & ClusterFlag) != 0
}

// ClusterID returns the index into the cluster table if this cell is a cluster.
func (c Cell) ClusterID() uint32 {
	return uint32(c.Rune &^ ClusterFlag)
}
