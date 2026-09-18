package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// TreeNode represents a node within a TreeView hierarchy.
type TreeNode struct {
	ID       string
	Label    string
	Icon     string
	Expanded bool
	Children []*TreeNode
	Data     any
}

type flatNode struct {
	node   *TreeNode
	depth  int
	isLast bool
}

// TreeView renders an interactive hierarchical tree with collapsible nodes.
type TreeView struct {
	roots        []*TreeNode
	flat         []flatNode
	selectedIdx  int
	scrollOffset int
	normalStyle  style.Style
	selectStyle  style.Style
	iconStyle    style.Style
	bounds       buffer.Rect
}

// NewTreeView creates a new TreeView widget with the provided root nodes.
func NewTreeView(roots ...*TreeNode) *TreeView {
	tv := &TreeView{
		roots:       roots,
		selectedIdx: 0,
		normalStyle: style.NewStyle().Foreground(cell.ColorHex("#CCCCDD")),
		selectStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#FFFFFF")).
			Background(cell.ColorHex("#335588")),
		iconStyle: style.NewStyle().Foreground(cell.ColorHex("#00FFAA")),
	}
	tv.rebuildFlat()
	return tv
}

// SetRoots updates the root nodes and rebuilds the flat hierarchy.
func (tv *TreeView) SetRoots(roots ...*TreeNode) *TreeView {
	tv.roots = roots
	tv.rebuildFlat()
	return tv
}

func (tv *TreeView) rebuildFlat() {
	tv.flat = tv.flat[:0]
	visited := make(map[*TreeNode]bool)
	var walk func(nodes []*TreeNode, depth int)
	walk = func(nodes []*TreeNode, depth int) {
		if depth > 50 {
			return
		}
		for i, n := range nodes {
			if n == nil || visited[n] {
				continue
			}
			visited[n] = true
			tv.flat = append(tv.flat, flatNode{
				node:   n,
				depth:  depth,
				isLast: i == len(nodes)-1,
			})
			if n.Expanded && len(n.Children) > 0 {
				walk(n.Children, depth+1)
			}
		}
	}
	walk(tv.roots, 0)
	if tv.selectedIdx >= len(tv.flat) {
		tv.selectedIdx = max(0, len(tv.flat)-1)
	}
}

// Selected returns the currently selected TreeNode, or nil if tree is empty.
func (tv *TreeView) Selected() *TreeNode {
	if tv.selectedIdx >= 0 && tv.selectedIdx < len(tv.flat) {
		return tv.flat[tv.selectedIdx].node
	}
	return nil
}

// SelectNext moves selection down.
func (tv *TreeView) SelectNext() *TreeView {
	if tv.selectedIdx < len(tv.flat)-1 {
		tv.selectedIdx++
	}
	return tv
}

// SelectPrev moves selection up.
func (tv *TreeView) SelectPrev() *TreeView {
	if tv.selectedIdx > 0 {
		tv.selectedIdx--
	}
	return tv
}

// ToggleExpand toggles the expansion state of the currently selected node.
func (tv *TreeView) ToggleExpand() *TreeView {
	sel := tv.Selected()
	if sel != nil && len(sel.Children) > 0 {
		sel.Expanded = !sel.Expanded
		tv.rebuildFlat()
	}
	return tv
}

// HandleKey handles navigation within the tree.
func (tv *TreeView) HandleKey(key input.Key) bool {
	switch key.Type {
	case input.KeyUp:
		tv.SelectPrev()
		return true
	case input.KeyDown:
		tv.SelectNext()
		return true
	case input.KeyRight:
		sel := tv.Selected()
		if sel != nil && len(sel.Children) > 0 {
			if !sel.Expanded {
				sel.Expanded = true
				tv.rebuildFlat()
				return true
			}
			tv.SelectNext()
			return true
		}
	case input.KeyLeft:
		sel := tv.Selected()
		if sel != nil {
			if sel.Expanded && len(sel.Children) > 0 {
				sel.Expanded = false
				tv.rebuildFlat()
				return true
			}
			tv.SelectPrev()
			return true
		}
	case input.KeyEnter, input.KeySpace:
		tv.ToggleExpand()
		return true
	}
	return false
}

// HandleMouse handles clicks on tree nodes.
func (tv *TreeView) HandleMouse(msg tea.MouseMsg) bool {
	if !tv.bounds.IsEmpty() && tv.bounds.Contains(msg.X, msg.Y) {
		if msg.Button == input.MouseWheelUp {
			tv.SelectPrev()
			return true
		}
		if msg.Button == input.MouseWheelDown {
			tv.SelectNext()
			return true
		}
		if msg.Button == input.MouseLeft && msg.Action == input.MousePress {
			relY := msg.Y - tv.bounds.Y
			targetIdx := tv.scrollOffset + relY
			if targetIdx >= 0 && targetIdx < len(tv.flat) {
				if tv.selectedIdx == targetIdx {
					tv.ToggleExpand()
				} else {
					tv.selectedIdx = targetIdx
				}
				return true
			}
		}
	}
	return false
}

// Draw renders the tree within the specified rectangular area.
func (tv *TreeView) Draw(buf *buffer.Buffer, area buffer.Rect) {
	tv.bounds = area
	if area.IsEmpty() || len(tv.flat) == 0 {
		return
	}

	// Adjust scroll offset to keep selected node visible
	if tv.selectedIdx < tv.scrollOffset {
		tv.scrollOffset = tv.selectedIdx
	}
	if tv.selectedIdx >= tv.scrollOffset+area.Height {
		tv.scrollOffset = tv.selectedIdx - area.Height + 1
	}

	for row := 0; row < area.Height; row++ {
		idx := tv.scrollOffset + row
		if idx >= len(tv.flat) {
			break
		}

		item := tv.flat[idx]
		currY := area.Y + row
		isSelected := (idx == tv.selectedIdx)

		st := tv.normalStyle
		if isSelected {
			st = tv.selectStyle
		}

		// Clear row with background
		rowRect := buffer.NewRect(area.X, currY, area.Width, 1)
		st.Draw(buf, rowRect, "")

		// Compute indentation
		indentX := area.X + item.depth*2
		if indentX >= area.Right() {
			continue
		}

		// Expand/collapse glyph
		expandGlyph := "  "
		if len(item.node.Children) > 0 {
			if item.node.Expanded {
				expandGlyph = "▼ "
			} else {
				expandGlyph = "▶ "
			}
		}
		buf.SetString(indentX, currY, expandGlyph, tv.iconStyle.GetFg(), st.GetBg(), cell.AttrNone)

		// Icon and Label
		textX := indentX + 2
		iconStr := item.node.Icon
		if iconStr != "" {
			buf.SetString(textX, currY, iconStr+" ", tv.iconStyle.GetFg(), st.GetBg(), cell.AttrNone)
			textX += buffer.StringWidth(iconStr + " ")
		}

		labelW := area.Right() - textX
		if labelW > 0 {
			lineArea := buffer.NewRect(textX, currY, labelW, 1)
			buf.SetStringAligned(lineArea, item.node.Label, buffer.AlignLeft, st.GetFg(), st.GetBg(), st.GetModifier())
		}
	}
}
