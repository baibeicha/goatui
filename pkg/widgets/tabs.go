package widgets

import (
	"unicode"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// TabItem represents an individual tab within a Tabs widget.
type TabItem struct {
	ID     string
	Title  string
	Hotkey rune
	Badge  string
}

// Tabs renders an interactive horizontal bar of tabs with hotkeys and mouse support.
type Tabs struct {
	items         []TabItem
	activeIdx     int
	activeStyle   style.Style
	inactiveStyle style.Style
	dividerStyle  style.Style
	tabBounds     []buffer.Rect // calculated on Draw for mouse hit-testing
}

// NewTabs creates a new Tabs widget with the provided items.
func NewTabs(items ...TabItem) *Tabs {
	return &Tabs{
		items:     items,
		activeIdx: 0,
		activeStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")).
			Background(cell.Color256(236)),
		inactiveStyle: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")).
			Background(cell.Color256(234)),
		dividerStyle: style.NewStyle().
			Foreground(cell.ColorHex("#444466")),
	}
}

// SetItems updates the tab item list.
func (t *Tabs) SetItems(items []TabItem) *Tabs {
	t.items = items
	if t.activeIdx >= len(items) {
		t.activeIdx = max(0, len(items)-1)
	}
	return t
}

// Items returns the current tab item list.
func (t *Tabs) Items() []TabItem {
	return t.items
}

// SetActive sets the active tab index.
func (t *Tabs) SetActive(idx int) *Tabs {
	if idx >= 0 && idx < len(t.items) {
		t.activeIdx = idx
	}
	return t
}

// Active returns the currently selected tab index.
func (t *Tabs) Active() int {
	return t.activeIdx
}

// ActiveItem returns a pointer to the currently selected TabItem, or nil if empty.
func (t *Tabs) ActiveItem() *TabItem {
	if t.activeIdx >= 0 && t.activeIdx < len(t.items) {
		return &t.items[t.activeIdx]
	}
	return nil
}

// SelectNext selects the next tab, wrapping around.
func (t *Tabs) SelectNext() *Tabs {
	if len(t.items) > 0 {
		t.activeIdx = (t.activeIdx + 1) % len(t.items)
	}
	return t
}

// SelectPrev selects the previous tab, wrapping around.
func (t *Tabs) SelectPrev() *Tabs {
	if len(t.items) > 0 {
		t.activeIdx = (t.activeIdx - 1 + len(t.items)) % len(t.items)
	}
	return t
}

// SetActiveStyle sets styling for the active tab.
func (t *Tabs) SetActiveStyle(st style.Style) *Tabs {
	t.activeStyle = st
	return t
}

// SetInactiveStyle sets styling for inactive tabs.
func (t *Tabs) SetInactiveStyle(st style.Style) *Tabs {
	t.inactiveStyle = st
	return t
}

// HandleKey handles tab switching via Left/Right arrows or hotkeys.
func (t *Tabs) HandleKey(key input.Key) bool {
	switch key.Type {
	case input.KeyLeft:
		t.SelectPrev()
		return true
	case input.KeyRight:
		t.SelectNext()
		return true
	case input.KeyHome:
		if len(t.items) > 0 {
			t.activeIdx = 0
			return true
		}
	case input.KeyEnd:
		if len(t.items) > 0 {
			t.activeIdx = len(t.items) - 1
			return true
		}
	case input.KeyRune:
		for i, item := range t.items {
			if item.Hotkey != 0 && unicode.ToLower(item.Hotkey) == unicode.ToLower(key.Rune) {
				t.activeIdx = i
				return true
			}
		}
	}
	return false
}

// HandleMouse checks if a mouse click occurred inside any tab and activates it.
func (t *Tabs) HandleMouse(msg tea.MouseMsg) bool {
	if msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	for i, bounds := range t.tabBounds {
		if bounds.Contains(msg.X, msg.Y) {
			t.activeIdx = i
			return true
		}
	}
	return false
}

// Draw renders the horizontal tab bar into the specified area.
func (t *Tabs) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(t.items) == 0 {
		t.tabBounds = nil
		return
	}

	t.tabBounds = make([]buffer.Rect, len(t.items))
	currX := area.X

	for i, item := range t.items {
		if currX >= area.Right() {
			break
		}

		isActive := (i == t.activeIdx)
		st := t.inactiveStyle
		if isActive {
			st = t.activeStyle
		}

		title := item.Title
		if item.Hotkey != 0 {
			title = "[" + string(item.Hotkey) + "] " + title
		}
		if item.Badge != "" {
			title += " (" + item.Badge + ")"
		}
		formatted := " " + title + " "

		tabW := buffer.StringWidth(formatted)
		if currX+tabW > area.Right() {
			tabW = max(0, area.Right()-currX)
			runes := []rune(formatted)
			curW := 0
			var truncated []rune
			for _, r := range runes {
				rw := buffer.RuneWidth(r)
				if curW+rw > tabW {
					break
				}
				curW += rw
				truncated = append(truncated, r)
			}
			formatted = string(truncated)
		}

		tabArea := buffer.NewRect(currX, area.Y, tabW, 1)
		t.tabBounds[i] = tabArea

		st.Draw(buf, tabArea, formatted)
		currX += tabW

		// Draw subtle divider between tabs
		if currX < area.Right() && i < len(t.items)-1 {
			buf.SetRune(currX, area.Y, '│', t.dividerStyle.GetFg(), t.dividerStyle.GetBg(), cell.AttrDim)
			currX++
		}
	}

	// Fill remainder of the line if there's space
	if currX < area.Right() {
		emptyArea := buffer.NewRect(currX, area.Y, area.Right()-currX, 1)
		buf.Fill(emptyArea, cell.Cell{
			Rune:   ' ',
			Width:  1,
			FgType: cell.ColorDefault,
			BgType: t.inactiveStyle.GetBg().Type,
			Bg:     t.inactiveStyle.GetBg().Value,
		})
	}
}
