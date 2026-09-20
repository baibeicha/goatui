package widgets

import (
	"unicode"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// RadioStyle represents marker styles for radio buttons.
type RadioStyle int

const (
	RadioCircle RadioStyle = iota // ◉ / ◯
	RadioCircleFilled             // ● / ○
	RadioDot                      // ⦿ / ⦾
	RadioBrackets                 // (*) / ( )
	RadioCustom
)

// RadioMarkers returns the marker strings for the given style.
func RadioMarkers(s RadioStyle) (selected, unselected string) {
	switch s {
	case RadioCircleFilled:
		return "● ", "○ "
	case RadioDot:
		return "⦿ ", "⦾ "
	case RadioBrackets:
		return "(*) ", "( ) "
	case RadioCircle:
		fallthrough
	default:
		return "◉ ", "◯ "
	}
}

// RadioItem represents a single selectable choice in a RadioGroup.
type RadioItem struct {
	ID       string
	Label    string
	Disabled bool
	Hotkey   rune
	UserData any
}

// RadioGroup manages a set of mutually exclusive radio choices.
type RadioGroup struct {
	items            []RadioItem
	selectedID       string
	focusedIdx       int
	orientation      layout.Direction
	styleType        RadioStyle
	selectedMarker   string
	unselectedMarker string
	itemBounds       []buffer.Rect
	onSelect         func(item RadioItem)

	styleFocused   style.Style
	styleSelected  style.Style
	styleNormal    style.Style
	styleDisabled  style.Style
}

// NewRadioGroup creates an interactive radio button group.
func NewRadioGroup(items ...RadioItem) *RadioGroup {
	rg := &RadioGroup{
		items:            items,
		selectedID:       "",
		focusedIdx:       0,
		orientation:      layout.Vertical,
		styleType:        RadioCircle,
		selectedMarker:   "◉ ",
		unselectedMarker: "◯ ",
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
		styleSelected: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		styleNormal: style.NewStyle().
			Foreground(cell.ColorHex("#CCCCDD")),
		styleDisabled: style.NewStyle().
			Foreground(cell.ColorHex("#555566")),
	}
	if len(items) > 0 {
		rg.selectedID = items[0].ID
	}
	return rg
}

// SetItems replaces items in the radio group.
func (rg *RadioGroup) SetItems(items []RadioItem) *RadioGroup {
	rg.items = items
	if rg.focusedIdx >= len(items) {
		rg.focusedIdx = max(0, len(items)-1)
	}
	return rg
}

// Items returns the slice of radio items.
func (rg *RadioGroup) Items() []RadioItem {
	res := make([]RadioItem, len(rg.items))
	copy(res, rg.items)
	return res
}

// SelectedID returns the ID of the currently selected radio option.
func (rg *RadioGroup) SelectedID() string {
	return rg.selectedID
}

// SelectedItem returns the currently selected item or nil if none.
func (rg *RadioGroup) SelectedItem() *RadioItem {
	for i := range rg.items {
		if rg.items[i].ID == rg.selectedID {
			return &rg.items[i]
		}
	}
	return nil
}

// SetSelected sets the selected item by ID.
func (rg *RadioGroup) SetSelected(id string) *RadioGroup {
	if rg.selectedID != id {
		rg.selectedID = id
		for i, it := range rg.items {
			if it.ID == id {
				rg.focusedIdx = i
				if rg.onSelect != nil {
					rg.onSelect(it)
				}
				break
			}
		}
	}
	return rg
}

// Focused returns the index of the focused item.
func (rg *RadioGroup) Focused() int {
	return rg.focusedIdx
}

// SetFocused sets the focused index.
func (rg *RadioGroup) SetFocused(idx int) *RadioGroup {
	if idx >= 0 && idx < len(rg.items) {
		rg.focusedIdx = idx
	}
	return rg
}

// FocusNext advances focus to the next available item.
func (rg *RadioGroup) FocusNext() *RadioGroup {
	if len(rg.items) == 0 {
		return rg
	}
	for step := 1; step <= len(rg.items); step++ {
		idx := (rg.focusedIdx + step) % len(rg.items)
		if !rg.items[idx].Disabled {
			rg.focusedIdx = idx
			break
		}
	}
	return rg
}

// FocusPrev moves focus to the previous available item.
func (rg *RadioGroup) FocusPrev() *RadioGroup {
	if len(rg.items) == 0 {
		return rg
	}
	for step := 1; step <= len(rg.items); step++ {
		idx := (rg.focusedIdx - step + len(rg.items)) % len(rg.items)
		if !rg.items[idx].Disabled {
			rg.focusedIdx = idx
			break
		}
	}
	return rg
}

// SelectFocused selects the currently focused item.
func (rg *RadioGroup) SelectFocused() *RadioGroup {
	if rg.focusedIdx >= 0 && rg.focusedIdx < len(rg.items) {
		it := rg.items[rg.focusedIdx]
		if !it.Disabled {
			rg.selectedID = it.ID
			if rg.onSelect != nil {
				rg.onSelect(it)
			}
		}
	}
	return rg
}

// SetOrientation sets horizontal or vertical layout.
func (rg *RadioGroup) SetOrientation(dir layout.Direction) *RadioGroup {
	rg.orientation = dir
	return rg
}

// SetStyle applies a preset marker style.
func (rg *RadioGroup) SetStyle(s RadioStyle) *RadioGroup {
	rg.styleType = s
	sel, unsel := RadioMarkers(s)
	rg.selectedMarker = sel
	rg.unselectedMarker = unsel
	return rg
}

// SetMarkers configures custom text markers for selected and unselected states.
func (rg *RadioGroup) SetMarkers(selected, unselected string) *RadioGroup {
	rg.selectedMarker = selected
	rg.unselectedMarker = unselected
	return rg
}

// SetOnSelect registers a callback invoked when selection changes.
func (rg *RadioGroup) SetOnSelect(fn func(item RadioItem)) *RadioGroup {
	rg.onSelect = fn
	return rg
}

// HandleKey processes keyboard navigation and selection.
func (rg *RadioGroup) HandleKey(key input.Key) bool {
	switch key.Type {
	case input.KeyUp, input.KeyLeft:
		rg.FocusPrev()
		return true
	case input.KeyDown, input.KeyRight, input.KeyTab:
		rg.FocusNext()
		return true
	case input.KeyBacktab:
		rg.FocusPrev()
		return true
	case input.KeyEnter, input.KeySpace:
		rg.SelectFocused()
		return true
	case input.KeyHome:
		if len(rg.items) > 0 {
			for i, it := range rg.items {
				if !it.Disabled {
					rg.focusedIdx = i
					return true
				}
			}
		}
	case input.KeyEnd:
		if len(rg.items) > 0 {
			for i := len(rg.items) - 1; i >= 0; i-- {
				if !rg.items[i].Disabled {
					rg.focusedIdx = i
					return true
				}
			}
		}
	case input.KeyRune:
		if !key.HasCtrl() && !key.HasAlt() {
			for i, it := range rg.items {
				if it.Hotkey != 0 && unicode.ToLower(it.Hotkey) == unicode.ToLower(key.Rune) {
					if !it.Disabled {
						rg.focusedIdx = i
						rg.SelectFocused()
						return true
					}
				}
			}
		}
	}
	return false
}

// HandleMouse processes mouse clicks on radio options.
func (rg *RadioGroup) HandleMouse(msg tea.MouseMsg) bool {
	if msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	for i, bounds := range rg.itemBounds {
		if bounds.Contains(msg.X, msg.Y) {
			if !rg.items[i].Disabled {
				rg.focusedIdx = i
				rg.SelectFocused()
				return true
			}
		}
	}
	return false
}

// Draw renders the radio options into the buffer.
func (rg *RadioGroup) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(rg.items) == 0 {
		rg.itemBounds = nil
		return
	}

	rg.itemBounds = make([]buffer.Rect, len(rg.items))

	if rg.orientation == layout.Vertical {
		currY := area.Y
		for i, item := range rg.items {
			if currY >= area.Bottom() {
				break
			}

			isSelected := (item.ID == rg.selectedID)
			isFocused := (i == rg.focusedIdx)

			marker := rg.unselectedMarker
			if isSelected {
				marker = rg.selectedMarker
			}

			text := marker + item.Label
			if isFocused {
				text = "► " + text + " ◄"
			} else {
				text = "  " + text + "  "
			}

			itemW := buffer.StringWidth(text)
			if itemW > area.Width {
				itemW = area.Width
			}
			itemArea := buffer.NewRect(area.X, currY, itemW, 1)
			rg.itemBounds[i] = itemArea

			fg := rg.styleNormal.GetFg()
			bg := rg.styleNormal.GetBg()
			attrs := cell.AttrNone

			if item.Disabled {
				fg = rg.styleDisabled.GetFg()
				bg = rg.styleDisabled.GetBg()
				attrs = cell.AttrDim
			} else if isSelected {
				fg = rg.styleSelected.GetFg()
				bg = rg.styleSelected.GetBg()
				attrs = cell.AttrBold
			}

			if isFocused && !item.Disabled {
				fg = rg.styleFocused.GetFg()
				bg = rg.styleFocused.GetBg()
				attrs = cell.AttrBold
			}

			for cx := itemArea.X; cx < itemArea.Right(); cx++ {
				buf.SetRune(cx, itemArea.Y, ' ', fg, bg, attrs)
			}
			buf.SetStringAligned(itemArea, text, buffer.AlignLeft, fg, bg, attrs)

			currY++
		}
		return
	}

	// Horizontal orientation
	currX := area.X
	for i, item := range rg.items {
		if currX >= area.Right() {
			break
		}

		isSelected := (item.ID == rg.selectedID)
		isFocused := (i == rg.focusedIdx)

		marker := rg.unselectedMarker
		if isSelected {
			marker = rg.selectedMarker
		}

		text := marker + item.Label
		if isFocused {
			text = "► " + text + " ◄"
		} else {
			text = "  " + text + "  "
		}

		itemW := buffer.StringWidth(text)
		if currX+itemW > area.Right() {
			itemW = max(0, area.Right()-currX)
		}
		if itemW <= 0 {
			break
		}

		itemArea := buffer.NewRect(currX, area.Y, itemW, 1)
		rg.itemBounds[i] = itemArea

		fg := rg.styleNormal.GetFg()
		bg := rg.styleNormal.GetBg()
		attrs := cell.AttrNone

		if item.Disabled {
			fg = rg.styleDisabled.GetFg()
			bg = rg.styleDisabled.GetBg()
			attrs = cell.AttrDim
		} else if isSelected {
			fg = rg.styleSelected.GetFg()
			bg = rg.styleSelected.GetBg()
			attrs = cell.AttrBold
		}

		if isFocused && !item.Disabled {
			fg = rg.styleFocused.GetFg()
			bg = rg.styleFocused.GetBg()
			attrs = cell.AttrBold
		}

		for cx := itemArea.X; cx < itemArea.Right(); cx++ {
			buf.SetRune(cx, itemArea.Y, ' ', fg, bg, attrs)
		}
		buf.SetStringAligned(itemArea, text, buffer.AlignLeft, fg, bg, attrs)

		currX += itemW + 1
	}
}

// RadioButton represents a single standalone radio button.
type RadioButton struct {
	id               string
	label            string
	selected         bool
	focused          bool
	disabled         bool
	styleType        RadioStyle
	selectedMarker   string
	unselectedMarker string
	styleSelected    style.Style
	styleNormal      style.Style
	styleFocused     style.Style
	styleDisabled    style.Style
	bounds           buffer.Rect
	onChange         func(selected bool)
}

// NewRadioButton creates a standalone radio button.
func NewRadioButton(id, label string, selected bool) *RadioButton {
	return &RadioButton{
		id:               id,
		label:            label,
		selected:         selected,
		styleType:        RadioCircle,
		selectedMarker:   "◉ ",
		unselectedMarker: "◯ ",
		styleSelected: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		styleNormal: style.NewStyle().
			Foreground(cell.ColorHex("#CCCCDD")),
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
		styleDisabled: style.NewStyle().
			Foreground(cell.ColorHex("#555566")),
	}
}

// ID returns the radio button identifier.
func (rb *RadioButton) ID() string {
	return rb.id
}

// Label returns the radio button label.
func (rb *RadioButton) Label() string {
	return rb.label
}

// IsSelected returns true if selected.
func (rb *RadioButton) IsSelected() bool {
	return rb.selected
}

// SetSelected sets the selected state.
func (rb *RadioButton) SetSelected(sel bool) *RadioButton {
	if rb.selected != sel {
		rb.selected = sel
		if rb.onChange != nil {
			rb.onChange(sel)
		}
	}
	return rb
}

// SetFocused sets focus state.
func (rb *RadioButton) SetFocused(focused bool) *RadioButton {
	rb.focused = focused
	return rb
}

// SetDisabled sets disabled state.
func (rb *RadioButton) SetDisabled(d bool) *RadioButton {
	rb.disabled = d
	return rb
}

// SetStyle sets the radio marker style.
func (rb *RadioButton) SetStyle(s RadioStyle) *RadioButton {
	rb.styleType = s
	sel, unsel := RadioMarkers(s)
	rb.selectedMarker = sel
	rb.unselectedMarker = unsel
	return rb
}

// SetMarkers sets custom marker strings.
func (rb *RadioButton) SetMarkers(selected, unselected string) *RadioButton {
	rb.selectedMarker = selected
	rb.unselectedMarker = unselected
	return rb
}

// SetOnChange registers a change listener.
func (rb *RadioButton) SetOnChange(fn func(selected bool)) *RadioButton {
	rb.onChange = fn
	return rb
}

// HandleKey handles Enter/Space selection.
func (rb *RadioButton) HandleKey(key input.Key) bool {
	if !rb.focused || rb.disabled {
		return false
	}
	switch key.Type {
	case input.KeyEnter, input.KeySpace:
		rb.SetSelected(true)
		return true
	}
	return false
}

// HandleMouse handles mouse clicks on the standalone radio button.
func (rb *RadioButton) HandleMouse(msg tea.MouseMsg) bool {
	if rb.disabled || msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	if !rb.bounds.IsEmpty() && rb.bounds.Contains(msg.X, msg.Y) {
		rb.focused = true
		rb.SetSelected(true)
		return true
	}
	return false
}

// Draw renders the standalone radio button into buffer area.
func (rb *RadioButton) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	marker := rb.unselectedMarker
	if rb.selected {
		marker = rb.selectedMarker
	}
	text := marker + rb.label

	rb.bounds = buffer.NewRect(area.X, area.Y, min(area.Width, buffer.StringWidth(text)), 1)

	fg := rb.styleNormal.GetFg()
	bg := rb.styleNormal.GetBg()
	attrs := cell.AttrNone

	if rb.disabled {
		fg = rb.styleDisabled.GetFg()
		bg = rb.styleDisabled.GetBg()
		attrs = cell.AttrDim
	} else if rb.selected {
		fg = rb.styleSelected.GetFg()
		bg = rb.styleSelected.GetBg()
		attrs = cell.AttrBold
	}

	if rb.focused && !rb.disabled {
		fg = rb.styleFocused.GetFg()
		bg = rb.styleFocused.GetBg()
		attrs = cell.AttrBold
	}

	buf.SetStringAligned(area, text, buffer.AlignLeft, fg, bg, attrs)
}
