package widgets

import (
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// SelectItem represents an option inside a Select / Dropdown widget.
type SelectItem struct {
	ID       string
	Label    string
	Disabled bool
	UserData any
}

// Select is an interactive single-choice dropdown widget.
type Select struct {
	id          string
	label       string
	items       []SelectItem
	selectedIdx int
	focusedIdx  int
	isOpen      bool
	focused     bool
	scrollOff   int
	maxVisible  int
	bounds      buffer.Rect
	popupBounds buffer.Rect
	itemBounds  []buffer.Rect
	onSelect    func(item SelectItem)

	styleNormal   style.Style
	styleFocused  style.Style
	styleOpen     style.Style
	styleSelected style.Style
	styleItem     style.Style
	styleDisabled style.Style
}

// NewSelect creates a new Select / Dropdown widget.
func NewSelect(id, label string, items ...SelectItem) *Select {
	s := &Select{
		id:          id,
		label:       label,
		items:       items,
		selectedIdx: 0,
		focusedIdx:  0,
		maxVisible:  6,
		styleNormal: style.NewStyle().
			Foreground(cell.ColorHex("#CCCCDD")).
			Background(cell.Color256(236)),
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
		styleOpen: style.NewStyle().
			Border(style.BorderRounded).
			BorderForeground(cell.ColorHex("#00D2FF")).
			Background(cell.Color256(235)),
		styleSelected: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00FFAA")),
		styleItem: style.NewStyle().
			Foreground(cell.ColorHex("#EEEEEE")).
			Background(cell.Color256(235)),
		styleDisabled: style.NewStyle().
			Foreground(cell.ColorHex("#666677")).
			Background(cell.Color256(235)),
	}
	return s
}

// Items returns the list of items.
func (s *Select) Items() []SelectItem {
	return s.items
}

// SetItems updates items.
func (s *Select) SetItems(items []SelectItem) *Select {
	s.items = items
	if s.selectedIdx >= len(items) {
		s.selectedIdx = max(0, len(items)-1)
	}
	if s.focusedIdx >= len(items) {
		s.focusedIdx = max(0, len(items)-1)
	}
	maxScroll := max(0, len(items)-s.maxVisible)
	if s.scrollOff > maxScroll {
		s.scrollOff = maxScroll
	}
	return s
}

// Label returns the dropdown label or placeholder.
func (s *Select) Label() string {
	return s.label
}

// SetLabel sets the dropdown label or placeholder.
func (s *Select) SetLabel(label string) *Select {
	s.label = label
	return s
}

// SelectedItem returns the currently selected item or nil.
func (s *Select) SelectedItem() *SelectItem {
	if s.selectedIdx >= 0 && s.selectedIdx < len(s.items) {
		return &s.items[s.selectedIdx]
	}
	return nil
}

// SelectedID returns the selected item ID.
func (s *Select) SelectedID() string {
	if it := s.SelectedItem(); it != nil {
		return it.ID
	}
	return ""
}

// SetSelectedID selects item by ID.
func (s *Select) SetSelectedID(id string) *Select {
	for i, it := range s.items {
		if it.ID == id {
			s.selectedIdx = i
			s.focusedIdx = i
			if s.onSelect != nil {
				s.onSelect(it)
			}
			break
		}
	}
	return s
}

// SetFocused sets focus state.
func (s *Select) SetFocused(f bool) *Select {
	s.focused = f
	return s
}

// IsOpen returns whether dropdown popup is expanded.
func (s *Select) IsOpen() bool {
	return s.isOpen
}

// SetOpen opens or closes dropdown popup.
func (s *Select) SetOpen(open bool) *Select {
	s.isOpen = open
	if open {
		s.adjustScroll()
	}
	return s
}

// SetOnSelect registers a callback for selection changes.
func (s *Select) SetOnSelect(fn func(item SelectItem)) *Select {
	s.onSelect = fn
	return s
}

// HandleKey handles navigation and selection.
func (s *Select) HandleKey(key input.Key) bool {
	if !s.focused {
		return false
	}

	if !s.isOpen {
		switch key.Type {
		case input.KeyEnter, input.KeySpace, input.KeyDown:
			s.isOpen = true
			s.focusedIdx = s.selectedIdx
			s.adjustScroll()
			return true
		}
		return false
	}

	// Dropdown is open
	switch key.Type {
	case input.KeyEsc:
		s.isOpen = false
		return true
	case input.KeyUp:
		for step := 1; step <= len(s.items); step++ {
			idx := (s.focusedIdx - step + len(s.items)) % len(s.items)
			if !s.items[idx].Disabled {
				s.focusedIdx = idx
				break
			}
		}
		s.adjustScroll()
		return true
	case input.KeyDown:
		for step := 1; step <= len(s.items); step++ {
			idx := (s.focusedIdx + step) % len(s.items)
			if !s.items[idx].Disabled {
				s.focusedIdx = idx
				break
			}
		}
		s.adjustScroll()
		return true
	case input.KeyEnter, input.KeySpace:
		if s.focusedIdx >= 0 && s.focusedIdx < len(s.items) && !s.items[s.focusedIdx].Disabled {
			s.selectedIdx = s.focusedIdx
			s.isOpen = false
			if s.onSelect != nil {
				s.onSelect(s.items[s.selectedIdx])
			}
		}
		return true
	}
	return false
}

func (s *Select) adjustScroll() {
	if s.focusedIdx < s.scrollOff {
		s.scrollOff = s.focusedIdx
	} else if s.focusedIdx >= s.scrollOff+s.maxVisible {
		s.scrollOff = s.focusedIdx - s.maxVisible + 1
	}
}

// HandleMouse processes clicks and scrolling.
func (s *Select) HandleMouse(msg tea.MouseMsg) bool {
	if s.bounds.IsEmpty() {
		return false
	}

	if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
		if s.bounds.Contains(msg.X, msg.Y) {
			s.focused = true
			s.isOpen = !s.isOpen
			s.focusedIdx = s.selectedIdx
			if s.isOpen {
				s.adjustScroll()
			}
			return true
		}

		if s.isOpen && s.popupBounds.Contains(msg.X, msg.Y) {
			for i, b := range s.itemBounds {
				if b.Contains(msg.X, msg.Y) {
					itemIdx := s.scrollOff + i
					if itemIdx >= 0 && itemIdx < len(s.items) && !s.items[itemIdx].Disabled {
						s.selectedIdx = itemIdx
						s.focusedIdx = itemIdx
						s.isOpen = false
						if s.onSelect != nil {
							s.onSelect(s.items[s.selectedIdx])
						}
						return true
					}
				}
			}
			return true
		}

		if s.isOpen {
			s.isOpen = false
			return true
		}
	}

	if s.isOpen && msg.Action == input.MouseMotion {
		if s.popupBounds.Contains(msg.X, msg.Y) {
			for i, b := range s.itemBounds {
				if b.Contains(msg.X, msg.Y) {
					itemIdx := s.scrollOff + i
					if itemIdx >= 0 && itemIdx < len(s.items) && !s.items[itemIdx].Disabled {
						s.focusedIdx = itemIdx
						return true
					}
				}
			}
		}
	}

	if s.isOpen && (msg.Button == input.MouseWheelUp || msg.Button == input.MouseWheelDown) {
		if s.bounds.Contains(msg.X, msg.Y) || s.popupBounds.Contains(msg.X, msg.Y) {
			if msg.Button == input.MouseWheelUp && s.scrollOff > 0 {
				s.scrollOff--
			} else if msg.Button == input.MouseWheelDown && s.scrollOff+s.maxVisible < len(s.items) {
				s.scrollOff++
			}
			return true
		}
	}

	return false
}

// Draw renders the select box (and dropdown list if open).
func (s *Select) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	s.bounds = area

	label := s.label
	if it := s.SelectedItem(); it != nil {
		if s.label != "" {
			label = s.label + ": " + it.Label
		} else {
			label = it.Label
		}
	}

	arrow := " ▼"
	if s.isOpen {
		arrow = " ▲"
	}

	text := " " + label + arrow + " "
	st := s.styleNormal
	if s.focused {
		st = s.styleFocused
	}

	st.Draw(buf, area, text)

	if !s.isOpen || len(s.items) == 0 {
		return
	}

	// Calculate popup geometry
	visibleCount := min(s.maxVisible, len(s.items))
	popH := visibleCount + 2
	popW := max(area.Width, 20)
	for _, it := range s.items {
		popW = max(popW, buffer.StringWidth(it.Label)+6)
	}
	if popW > buf.Width() {
		popW = max(1, buf.Width())
	}

	popX := area.X
	if popX+popW > buf.Width() {
		popX = max(0, buf.Width()-popW)
	}

	roomBelow := buf.Height() - area.Bottom()
	roomAbove := area.Y

	var popY int
	if roomBelow >= popH || roomBelow >= roomAbove {
		popY = area.Bottom()
		if popY+popH > buf.Height() {
			popH = max(1, buf.Height()-popY)
		}
	} else {
		popY = max(0, area.Y-popH)
		popH = min(popH, area.Y)
		if popH < 1 {
			popH = 1
		}
	}

	s.popupBounds = buffer.NewRect(popX, popY, popW, popH)
	inner := s.popupBounds.Inset(1, 1)
	itemsToDraw := min(visibleCount, max(0, inner.Height))
	s.itemBounds = make([]buffer.Rect, itemsToDraw)
	for i := 0; i < itemsToDraw; i++ {
		s.itemBounds[i] = buffer.NewRect(inner.X, inner.Y+i, inner.Width, 1)
	}

	// Draw popup immediately for local/non-overlay rendering paths
	s.drawPopup(buf)

	// Enqueue overlay pass to ensure the popup renders above sibling widgets drawn later in the tree
	buf.AddOverlay(func(b *buffer.Buffer) {
		s.drawPopup(b)
	})
}

// drawPopup renders the dropdown popup box with solid background and item list.
func (s *Select) drawPopup(buf *buffer.Buffer) {
	if !s.isOpen || len(s.items) == 0 || s.popupBounds.IsEmpty() {
		return
	}

	// 1. Fill entire popup area with solid background cells to eliminate transparency/bleed-through
	bg := s.styleOpen.GetBg()
	if bg.IsDefault() {
		bg = cell.Color256(235)
	}
	fg := s.styleOpen.GetFg()
	if fg.IsDefault() {
		fg = cell.ColorHex("#CCCCDD")
	}

	blankCell := cell.Cell{
		Rune:     ' ',
		Width:    1,
		FgType:   fg.Type,
		Fg:       fg.Value,
		BgType:   bg.Type,
		Bg:       bg.Value,
		Modifier: cell.AttrNone,
	}
	buf.Fill(s.popupBounds, blankCell)

	// 2. Draw border and style
	s.styleOpen.Draw(buf, s.popupBounds, "")

	// 3. Render items
	visibleCount := len(s.itemBounds)
	for i := 0; i < visibleCount; i++ {
		idx := s.scrollOff + i
		if idx >= len(s.items) {
			break
		}
		it := s.items[idx]
		itemRect := s.itemBounds[i]

		isSelected := (idx == s.selectedIdx)
		isFocused := (idx == s.focusedIdx)

		itemSt := s.styleItem
		prefix := "  "
		if isSelected {
			prefix = "✓ "
			itemSt = s.styleSelected
		} else if isFocused {
			prefix = "► "
			itemSt = s.styleFocused
		} else if it.Disabled {
			itemSt = s.styleDisabled
		}

		itemBg := itemSt.GetBg()
		if itemBg.IsDefault() {
			itemBg = bg
		}
		itemFg := itemSt.GetFg()
		if itemFg.IsDefault() {
			itemFg = fg
		}

		// Fill item row completely
		buf.Fill(itemRect, cell.Cell{
			Rune:     ' ',
			Width:    1,
			FgType:   itemFg.Type,
			Fg:       itemFg.Value,
			BgType:   itemBg.Type,
			Bg:       itemBg.Value,
			Modifier: itemSt.GetModifier(),
		})

		itemText := prefix + it.Label
		pad := itemRect.Width - buffer.StringWidth(itemText)
		if pad > 0 {
			itemText += strings.Repeat(" ", pad)
		}

		itemSt.Draw(buf, itemRect, itemText)
	}
}
