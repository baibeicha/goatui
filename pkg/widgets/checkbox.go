package widgets

import (
	"unicode"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// CheckboxItem represents an item within a CheckboxGroup / filter chips list.
type CheckboxItem struct {
	ID      string
	Label   string
	Checked bool
	BadgeFg cell.Color
	BadgeBg cell.Color
	Hotkey  rune
}

// CheckboxGroup renders an interactive horizontal or wrapped row of toggleable chips/checkboxes.
type CheckboxGroup struct {
	items           []CheckboxItem
	focusedIdx      int
	itemBounds      []buffer.Rect
	checkedMarker   string
	uncheckedMarker string
	onToggle        func(item CheckboxItem, checked bool)

	styleFocused   style.Style
	styleChecked   style.Style
	styleUnchecked style.Style
}

// NewCheckboxGroup creates an interactive group of checkboxes.
func NewCheckboxGroup(items ...CheckboxItem) *CheckboxGroup {
	cg := &CheckboxGroup{
		items:           items,
		focusedIdx:      0,
		checkedMarker:   "[X] ",
		uncheckedMarker: "[ ] ",
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
		styleChecked: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		styleUnchecked: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")),
	}
	return cg
}

// SetItems updates the checkbox items in the group.
func (cg *CheckboxGroup) SetItems(items []CheckboxItem) *CheckboxGroup {
	cg.items = items
	if cg.focusedIdx >= len(items) {
		cg.focusedIdx = max(0, len(items)-1)
	}
	return cg
}

// Items returns a copy of current items.
func (cg *CheckboxGroup) Items() []CheckboxItem {
	return cg.items
}

// Focused returns the currently focused item index.
func (cg *CheckboxGroup) Focused() int {
	return cg.focusedIdx
}

// SetFocused updates the focused index.
func (cg *CheckboxGroup) SetFocused(idx int) *CheckboxGroup {
	if idx >= 0 && idx < len(cg.items) {
		cg.focusedIdx = idx
	}
	return cg
}

// FocusedItem returns a pointer to the currently focused item, or nil if empty.
func (cg *CheckboxGroup) FocusedItem() *CheckboxItem {
	if cg.focusedIdx >= 0 && cg.focusedIdx < len(cg.items) {
		return &cg.items[cg.focusedIdx]
	}
	return nil
}

// FocusNext moves focus to the next item, wrapping around.
func (cg *CheckboxGroup) FocusNext() *CheckboxGroup {
	if len(cg.items) > 0 {
		cg.focusedIdx = (cg.focusedIdx + 1) % len(cg.items)
	}
	return cg
}

// FocusPrev moves focus to the previous item, wrapping around.
func (cg *CheckboxGroup) FocusPrev() *CheckboxGroup {
	if len(cg.items) > 0 {
		cg.focusedIdx = (cg.focusedIdx - 1 + len(cg.items)) % len(cg.items)
	}
	return cg
}

// ToggleFocused inverts the checked state of the currently focused item.
func (cg *CheckboxGroup) ToggleFocused() *CheckboxGroup {
	if cg.focusedIdx >= 0 && cg.focusedIdx < len(cg.items) {
		cg.items[cg.focusedIdx].Checked = !cg.items[cg.focusedIdx].Checked
		if cg.onToggle != nil {
			cg.onToggle(cg.items[cg.focusedIdx], cg.items[cg.focusedIdx].Checked)
		}
	}
	return cg
}

// Toggle inverts the checked state of an item by ID.
func (cg *CheckboxGroup) Toggle(id string) *CheckboxGroup {
	for i := range cg.items {
		if cg.items[i].ID == id {
			cg.items[i].Checked = !cg.items[i].Checked
			if cg.onToggle != nil {
				cg.onToggle(cg.items[i], cg.items[i].Checked)
			}
			break
		}
	}
	return cg
}

// SetChecked explicitly sets the checked state of an item by ID.
func (cg *CheckboxGroup) SetChecked(id string, checked bool) *CheckboxGroup {
	for i := range cg.items {
		if cg.items[i].ID == id {
			cg.items[i].Checked = checked
			if cg.onToggle != nil {
				cg.onToggle(cg.items[i], checked)
			}
			break
		}
	}
	return cg
}

// ToggleAll sets the checked status for all items in the group.
func (cg *CheckboxGroup) ToggleAll(checked bool) *CheckboxGroup {
	for i := range cg.items {
		cg.items[i].Checked = checked
		if cg.onToggle != nil {
			cg.onToggle(cg.items[i], checked)
		}
	}
	return cg
}

// IsChecked checks if an item by ID is checked.
func (cg *CheckboxGroup) IsChecked(id string) bool {
	for i := range cg.items {
		if cg.items[i].ID == id {
			return cg.items[i].Checked
		}
	}
	return false
}

// CheckedIDs returns a slice of IDs for all checked items.
func (cg *CheckboxGroup) CheckedIDs() []string {
	var ids []string
	for _, it := range cg.items {
		if it.Checked {
			ids = append(ids, it.ID)
		}
	}
	return ids
}

// SetOnToggle assigns a callback when an item is toggled.
func (cg *CheckboxGroup) SetOnToggle(fn func(item CheckboxItem, checked bool)) *CheckboxGroup {
	cg.onToggle = fn
	return cg
}

// SetMarkers configures the text markers for checked and unchecked states.
func (cg *CheckboxGroup) SetMarkers(checked, unchecked string) *CheckboxGroup {
	cg.checkedMarker = checked
	cg.uncheckedMarker = unchecked
	return cg
}

// HandleKey handles navigation and toggling of checkboxes.
func (cg *CheckboxGroup) HandleKey(key input.Key) bool {
	switch key.Type {
	case input.KeyLeft:
		cg.FocusPrev()
		return true
	case input.KeyRight, input.KeyTab:
		cg.FocusNext()
		return true
	case input.KeyBacktab:
		cg.FocusPrev()
		return true
	case input.KeyEnter, input.KeySpace:
		cg.ToggleFocused()
		return true
	case input.KeyHome:
		if len(cg.items) > 0 {
			cg.focusedIdx = 0
			return true
		}
	case input.KeyEnd:
		if len(cg.items) > 0 {
			cg.focusedIdx = len(cg.items) - 1
			return true
		}
	case input.KeyRune:
		for i, item := range cg.items {
			if item.Hotkey != 0 && unicode.ToLower(item.Hotkey) == unicode.ToLower(key.Rune) {
				cg.focusedIdx = i
				cg.ToggleFocused()
				return true
			}
		}
	}
	return false
}

// HandleMouse handles clicking on individual checkboxes.
func (cg *CheckboxGroup) HandleMouse(msg tea.MouseMsg) bool {
	if msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	for i, bounds := range cg.itemBounds {
		if bounds.Contains(msg.X, msg.Y) {
			cg.focusedIdx = i
			cg.ToggleFocused()
			return true
		}
	}
	return false
}

// Draw renders the horizontal row of checkboxes into area.
func (cg *CheckboxGroup) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(cg.items) == 0 {
		cg.itemBounds = nil
		return
	}

	cg.itemBounds = make([]buffer.Rect, len(cg.items))
	currX := area.X

	for i, item := range cg.items {
		if currX >= area.Right() {
			break
		}

		isFocused := (i == cg.focusedIdx)
		marker := cg.uncheckedMarker
		if item.Checked {
			marker = cg.checkedMarker
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
		cg.itemBounds[i] = itemArea

		// Determine colors
		fg := cg.styleUnchecked.GetFg()
		bg := cg.styleUnchecked.GetBg()
		attrs := cell.AttrNone

		if item.Checked {
			fg = cg.styleChecked.GetFg()
			bg = cg.styleChecked.GetBg()
			attrs = cell.AttrBold
			if item.BadgeFg.Type != cell.ColorDefault {
				fg = item.BadgeFg
			}
			if item.BadgeBg.Type != cell.ColorDefault {
				bg = item.BadgeBg
			}
		}

		if isFocused {
			fg = cg.styleFocused.GetFg()
			bg = cg.styleFocused.GetBg()
			attrs = cell.AttrBold
		}

		// Fill item area
		for cx := itemArea.X; cx < itemArea.Right(); cx++ {
			buf.SetRune(cx, itemArea.Y, ' ', fg, bg, attrs)
		}

		// Draw formatted string within bounds
		buf.SetString(itemArea.X, itemArea.Y, text, fg, bg, attrs)

		currX += itemW + 1 // 1 column spacer between chips
	}
}

// Checkbox represents a single standalone interactive checkbox widget.
type Checkbox struct {
	id              string
	label           string
	checked         bool
	focused         bool
	checkedMarker   string
	uncheckedMarker string
	styleChecked    style.Style
	styleUnchecked  style.Style
	styleFocused    style.Style
	onChange        func(checked bool)
}

// NewCheckbox creates a new standalone Checkbox widget.
func NewCheckbox(id, label string, checked bool) *Checkbox {
	return &Checkbox{
		id:              id,
		label:           label,
		checked:         checked,
		checkedMarker:   "[X] ",
		uncheckedMarker: "[ ] ",
		styleChecked: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
		styleUnchecked: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")),
		styleFocused: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
	}
}

// ID returns the checkbox identifier.
func (c *Checkbox) ID() string {
	return c.id
}

// Label returns the checkbox label.
func (c *Checkbox) Label() string {
	return c.label
}

// IsChecked returns true if checked.
func (c *Checkbox) IsChecked() bool {
	return c.checked
}

// SetChecked sets the checked state.
func (c *Checkbox) SetChecked(checked bool) *Checkbox {
	if c.checked != checked {
		c.checked = checked
		if c.onChange != nil {
			c.onChange(checked)
		}
	}
	return c
}

// Toggle inverts the checked state.
func (c *Checkbox) Toggle() *Checkbox {
	return c.SetChecked(!c.checked)
}

// SetFocused sets the focus state.
func (c *Checkbox) SetFocused(focused bool) *Checkbox {
	c.focused = focused
	return c
}

// SetOnChange registers a callback on state change.
func (c *Checkbox) SetOnChange(fn func(checked bool)) *Checkbox {
	c.onChange = fn
	return c
}

// HandleKey processes keyboard toggling when focused.
func (c *Checkbox) HandleKey(key input.Key) bool {
	if !c.focused {
		return false
	}
	switch key.Type {
	case input.KeyEnter, input.KeySpace:
		c.Toggle()
		return true
	}
	return false
}

// Draw renders the checkbox into buffer area.
func (c *Checkbox) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	marker := c.uncheckedMarker
	if c.checked {
		marker = c.checkedMarker
	}
	text := marker + c.label

	fg := c.styleUnchecked.GetFg()
	bg := c.styleUnchecked.GetBg()
	attrs := cell.AttrNone

	if c.checked {
		fg = c.styleChecked.GetFg()
		bg = c.styleChecked.GetBg()
		attrs = cell.AttrBold
	}
	if c.focused {
		fg = c.styleFocused.GetFg()
		bg = c.styleFocused.GetBg()
		attrs = cell.AttrBold
	}

	buf.SetString(area.X, area.Y, text, fg, bg, attrs)
}
