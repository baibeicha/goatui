package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
)

// ItemRenderer renders a single visible row in the virtual list.
type ItemRenderer func(buf *buffer.Buffer, area buffer.Rect, index int, selected bool)

// VirtualList efficiently displays lists of any size (up to 1,000,000+ items)
// by rendering only the visible viewport slice with zero heap allocations.
type VirtualList struct {
	totalItems    int
	selected      int
	offset        int
	itemHeight    int
	renderer      ItemRenderer
	selectedStyle style.Style
	normalStyle   style.Style
}

// NewVirtualList creates a virtualized list.
func NewVirtualList(totalItems int, renderer ItemRenderer) *VirtualList {
	return &VirtualList{
		totalItems: totalItems,
		itemHeight: 1,
		renderer:   renderer,
		selectedStyle: style.NewStyle().
			Bold(true).
			Reverse(true),
		normalStyle: style.NewStyle(),
	}
}

// SetTotalItems updates the total dataset size.
func (vl *VirtualList) SetTotalItems(total int) {
	if total < 0 {
		total = 0
	}
	vl.totalItems = total
	if vl.selected >= vl.totalItems {
		vl.selected = max(0, vl.totalItems-1)
	}
	vl.clampOffset(20) // Default fallback
}

// Selected returns the currently active item index.
func (vl *VirtualList) Selected() int {
	return vl.selected
}

// Select sets the selected item index, scrolling viewport if necessary.
func (vl *VirtualList) Select(idx int) {
	if vl.totalItems == 0 {
		vl.selected = 0
		vl.offset = 0
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= vl.totalItems {
		idx = vl.totalItems - 1
	}
	vl.selected = idx
}

// ScrollDown moves selection down by n items.
func (vl *VirtualList) ScrollDown(n int) {
	vl.Select(vl.selected + n)
}

// ScrollUp moves selection up by n items.
func (vl *VirtualList) ScrollUp(n int) {
	vl.Select(vl.selected - n)
}

// Draw renders the visible portion of the virtual list into buf.
func (vl *VirtualList) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || vl.totalItems == 0 {
		return
	}

	visibleCount := area.Height / vl.itemHeight
	if visibleCount <= 0 {
		visibleCount = 1
	}

	// Adjust viewport offset so selected item remains in view
	if vl.selected < vl.offset {
		vl.offset = vl.selected
	} else if vl.selected >= vl.offset+visibleCount {
		vl.offset = vl.selected - visibleCount + 1
	}
	vl.clampOffset(visibleCount)

	end := min(vl.totalItems, vl.offset+visibleCount)

	for i := vl.offset; i < end; i++ {
		rowY := area.Y + (i-vl.offset)*vl.itemHeight
		itemArea := buffer.NewRect(area.X, rowY, area.Width, vl.itemHeight)
		isSelected := (i == vl.selected)

		if vl.renderer != nil {
			vl.renderer(buf, itemArea, i, isSelected)
		} else {
			st := vl.normalStyle
			if isSelected {
				st = vl.selectedStyle
			}
			st.Draw(buf, itemArea, "")
		}
	}
}

func (vl *VirtualList) clampOffset(visibleCount int) {
	maxOffset := max(0, vl.totalItems-visibleCount)
	if vl.offset > maxOffset {
		vl.offset = maxOffset
	}
	if vl.offset < 0 {
		vl.offset = 0
	}
}

// DefaultTextRenderer returns an ItemRenderer for slice of strings with zero allocs (left-aligned).
func DefaultTextRenderer(items []string, selFg, selBg cell.Color) ItemRenderer {
	return DefaultAlignedTextRenderer(items, buffer.AlignLeft, selFg, selBg)
}

// DefaultAlignedTextRenderer returns an ItemRenderer for slice of strings with customizable alignment.
func DefaultAlignedTextRenderer(items []string, align buffer.Alignment, selFg, selBg cell.Color) ItemRenderer {
	return func(buf *buffer.Buffer, area buffer.Rect, index int, selected bool) {
		if index < 0 || index >= len(items) {
			return
		}
		item := items[index]
		mod := cell.AttrNone
		fg := cell.DefaultColor()
		bg := cell.DefaultColor()
		if selected {
			mod = cell.AttrBold
			fg = selFg
			bg = selBg
			selCell := cell.Cell{
				Rune:     ' ',
				Width:    1,
				FgType:   selFg.Type,
				BgType:   selBg.Type,
				Fg:       selFg.Value,
				Bg:       selBg.Value,
				Modifier: mod,
			}
			buf.Fill(area, selCell)
		}
		buf.SetStringAligned(area, item, align, fg, bg, mod)
	}
}
