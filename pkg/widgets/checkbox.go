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

// CheckboxStyle represents marker styles for checkboxes.
type CheckboxStyle int

const (
	CheckboxBrackets CheckboxStyle = iota // [X] / [ ] / [-]
	CheckboxBox                           // ☒ / ☐ / ▣
	CheckboxCircle                        // ◉ / ◯ / ◐
	CheckboxCircleFilled                  // ● / ○ / ◐
	CheckboxCross                         // ✗ / · / -
	CheckboxCheck                         // ✓ /   / -
	CheckboxDot                           // ⦿ / ⦾ / ⊙
	CheckboxSwitch                        // [ON] / [OFF] / [---]
	CheckboxToggle                        // ●━ / ○─ / ◐─
	CheckboxCustom
)

// CheckboxMarkers returns the marker strings for the given style.
func CheckboxMarkers(s CheckboxStyle) (checked, unchecked, indeterminate string) {
	switch s {
	case CheckboxBox:
		return "☒ ", "☐ ", "▣ "
	case CheckboxCircle:
		return "◉ ", "◯ ", "◐ "
	case CheckboxCircleFilled:
		return "● ", "○ ", "◐ "
	case CheckboxCross:
		return "✗ ", "· ", "- "
	case CheckboxCheck:
		return "✓ ", "  ", "- "
	case CheckboxDot:
		return "⦿ ", "⦾ ", "⊙ "
	case CheckboxSwitch:
		return "[ON] ", "[OFF] ", "[---] "
	case CheckboxToggle:
		return "●━ ", "○─ ", "◐─ "
	case CheckboxBrackets:
		fallthrough
	default:
		return "[X] ", "[ ] ", "[-] "
	}
}

// CheckboxItem represents an item within a CheckboxGroup / filter chips list.
type CheckboxItem struct {
	ID            string
	Label         string
	Checked       bool
	Indeterminate bool
	BadgeFg       cell.Color
	BadgeBg       cell.Color
	Hotkey        rune
}

// CheckboxGroup renders an interactive horizontal or vertical row/list of toggleable chips/checkboxes.
type CheckboxGroup struct {
	items               []CheckboxItem
	focusedIdx          int
	itemBounds          []buffer.Rect
	checkedMarker       string
	uncheckedMarker     string
	indeterminateMarker string
	orientation         layout.Direction
	styleType           CheckboxStyle
	triState            bool
	onToggle            func(item CheckboxItem, checked bool)

	styleFocused   style.Style
	styleChecked   style.Style
	styleUnchecked style.Style
}

// NewCheckboxGroup creates an interactive group of checkboxes.
func NewCheckboxGroup(items ...CheckboxItem) *CheckboxGroup {
	cg := &CheckboxGroup{
		items:               items,
		focusedIdx:          0,
		checkedMarker:       "[X] ",
		uncheckedMarker:     "[ ] ",
		indeterminateMarker: "[-] ",
		orientation:         layout.Horizontal,
		styleType:           CheckboxBrackets,
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
	res := make([]CheckboxItem, len(cg.items))
	copy(res, cg.items)
	return res
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

// SetOrientation sets horizontal or vertical layout for the group.
func (cg *CheckboxGroup) SetOrientation(dir layout.Direction) *CheckboxGroup {
	cg.orientation = dir
	return cg
}

// SetStyle applies a preset checkbox marker style.
func (cg *CheckboxGroup) SetStyle(s CheckboxStyle) *CheckboxGroup {
	cg.styleType = s
	chk, unchk, ind := CheckboxMarkers(s)
	cg.checkedMarker = chk
	cg.uncheckedMarker = unchk
	cg.indeterminateMarker = ind
	return cg
}

// SetTriState enables or disables 3-state (indeterminate) cycling.
func (cg *CheckboxGroup) SetTriState(tri bool) *CheckboxGroup {
	cg.triState = tri
	return cg
}

// ToggleFocused inverts or cycles the checked state of the currently focused item.
func (cg *CheckboxGroup) ToggleFocused() *CheckboxGroup {
	if cg.focusedIdx >= 0 && cg.focusedIdx < len(cg.items) {
		it := &cg.items[cg.focusedIdx]
		if cg.triState {
			if !it.Checked && !it.Indeterminate {
				it.Checked = true
				it.Indeterminate = false
			} else if it.Checked && !it.Indeterminate {
				it.Checked = false
				it.Indeterminate = true
			} else {
				it.Checked = false
				it.Indeterminate = false
			}
		} else {
			it.Checked = !it.Checked
			it.Indeterminate = false
		}
		if cg.onToggle != nil {
			cg.onToggle(*it, it.Checked)
		}
	}
	return cg
}

// Toggle inverts the checked state of an item by ID.
func (cg *CheckboxGroup) Toggle(id string) *CheckboxGroup {
	for i := range cg.items {
		if cg.items[i].ID == id {
			it := &cg.items[i]
			if cg.triState {
				if !it.Checked && !it.Indeterminate {
					it.Checked = true
					it.Indeterminate = false
				} else if it.Checked && !it.Indeterminate {
					it.Checked = false
					it.Indeterminate = true
				} else {
					it.Checked = false
					it.Indeterminate = false
				}
			} else {
				it.Checked = !it.Checked
				it.Indeterminate = false
			}
			if cg.onToggle != nil {
				cg.onToggle(*it, it.Checked)
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
			cg.items[i].Indeterminate = false
			if cg.onToggle != nil {
				cg.onToggle(cg.items[i], checked)
			}
			break
		}
	}
	return cg
}

// SetIndeterminate explicitly sets the indeterminate state of an item by ID.
func (cg *CheckboxGroup) SetIndeterminate(id string, ind bool) *CheckboxGroup {
	for i := range cg.items {
		if cg.items[i].ID == id {
			cg.items[i].Indeterminate = ind
			if ind {
				cg.items[i].Checked = false
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
		cg.items[i].Indeterminate = false
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

// IsIndeterminate checks if an item by ID is indeterminate.
func (cg *CheckboxGroup) IsIndeterminate(id string) bool {
	for i := range cg.items {
		if cg.items[i].ID == id {
			return cg.items[i].Indeterminate
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

// SetMarkers configures custom text markers for checked, unchecked, and indeterminate states.
func (cg *CheckboxGroup) SetMarkers(checked, unchecked, indeterminate string) *CheckboxGroup {
	cg.checkedMarker = checked
	cg.uncheckedMarker = unchecked
	if indeterminate != "" {
		cg.indeterminateMarker = indeterminate
	}
	return cg
}

// HandleKey handles navigation and toggling of checkboxes.
func (cg *CheckboxGroup) HandleKey(key input.Key) bool {
	switch key.Type {
	case input.KeyLeft, input.KeyUp:
		cg.FocusPrev()
		return true
	case input.KeyRight, input.KeyDown, input.KeyTab:
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
		if !key.HasCtrl() && !key.HasAlt() {
			for i, item := range cg.items {
				if item.Hotkey != 0 && unicode.ToLower(item.Hotkey) == unicode.ToLower(key.Rune) {
					cg.focusedIdx = i
					cg.ToggleFocused()
					return true
				}
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

// Draw renders the checkboxes into area according to orientation.
func (cg *CheckboxGroup) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(cg.items) == 0 {
		cg.itemBounds = nil
		return
	}

	cg.itemBounds = make([]buffer.Rect, len(cg.items))

	if cg.orientation == layout.Vertical {
		currY := area.Y
		for i, item := range cg.items {
			if currY >= area.Bottom() {
				break
			}

			isFocused := (i == cg.focusedIdx)
			marker := cg.uncheckedMarker
			if item.Indeterminate {
				marker = cg.indeterminateMarker
			} else if item.Checked {
				marker = cg.checkedMarker
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
			cg.itemBounds[i] = itemArea

			fg := cg.styleUnchecked.GetFg()
			bg := cg.styleUnchecked.GetBg()
			attrs := cell.AttrNone

			if item.Checked || item.Indeterminate {
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
	for i, item := range cg.items {
		if currX >= area.Right() {
			break
		}

		isFocused := (i == cg.focusedIdx)
		marker := cg.uncheckedMarker
		if item.Indeterminate {
			marker = cg.indeterminateMarker
		} else if item.Checked {
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

		if item.Checked || item.Indeterminate {
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
		buf.SetStringAligned(itemArea, text, buffer.AlignLeft, fg, bg, attrs)

		currX += itemW + 1 // 1 column spacer between chips
	}
}

// Checkbox represents a single standalone interactive checkbox widget.
type Checkbox struct {
	id                  string
	label               string
	checked             bool
	indeterminate       bool
	triState            bool
	focused             bool
	styleType           CheckboxStyle
	checkedMarker       string
	uncheckedMarker     string
	indeterminateMarker string
	styleChecked        style.Style
	styleUnchecked      style.Style
	styleFocused        style.Style
	bounds              buffer.Rect
	onChange            func(checked bool)
}

// NewCheckbox creates a new standalone Checkbox widget.
func NewCheckbox(id, label string, checked bool) *Checkbox {
	return &Checkbox{
		id:                  id,
		label:               label,
		checked:             checked,
		styleType:           CheckboxBrackets,
		checkedMarker:       "[X] ",
		uncheckedMarker:     "[ ] ",
		indeterminateMarker: "[-] ",
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

// IsIndeterminate returns true if in indeterminate state.
func (c *Checkbox) IsIndeterminate() bool {
	return c.indeterminate
}

// SetStyle configures the checkbox marker style.
func (c *Checkbox) SetStyle(s CheckboxStyle) *Checkbox {
	c.styleType = s
	chk, unchk, ind := CheckboxMarkers(s)
	c.checkedMarker = chk
	c.uncheckedMarker = unchk
	c.indeterminateMarker = ind
	return c
}

// SetMarkers configures custom text markers for checked, unchecked, and indeterminate states.
func (c *Checkbox) SetMarkers(checked, unchecked, indeterminate string) *Checkbox {
	c.checkedMarker = checked
	c.uncheckedMarker = unchecked
	if indeterminate != "" {
		c.indeterminateMarker = indeterminate
	}
	return c
}

// SetTriState enables or disables 3-state cycling for standalone checkbox.
func (c *Checkbox) SetTriState(tri bool) *Checkbox {
	c.triState = tri
	return c
}

// SetIndeterminate sets the indeterminate state.
func (c *Checkbox) SetIndeterminate(ind bool) *Checkbox {
	c.indeterminate = ind
	if ind {
		c.checked = false
	}
	return c
}

// SetChecked sets the checked state.
func (c *Checkbox) SetChecked(checked bool) *Checkbox {
	if c.checked != checked || c.indeterminate {
		c.checked = checked
		c.indeterminate = false
		if c.onChange != nil {
			c.onChange(checked)
		}
	}
	return c
}

// Cycle cycles through states: unchecked -> checked -> indeterminate (if triState) -> unchecked.
func (c *Checkbox) Cycle() *Checkbox {
	if c.triState {
		if !c.checked && !c.indeterminate {
			c.checked = true
			c.indeterminate = false
		} else if c.checked && !c.indeterminate {
			c.checked = false
			c.indeterminate = true
		} else {
			c.checked = false
			c.indeterminate = false
		}
		if c.onChange != nil {
			c.onChange(c.checked)
		}
		return c
	}
	return c.Toggle()
}

// Toggle inverts the checked state.
func (c *Checkbox) Toggle() *Checkbox {
	if c.triState {
		return c.Cycle()
	}
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

// HandleMouse processes mouse clicking on the checkbox.
func (c *Checkbox) HandleMouse(msg tea.MouseMsg) bool {
	if msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	if !c.bounds.IsEmpty() && c.bounds.Contains(msg.X, msg.Y) {
		c.focused = true
		c.Toggle()
		return true
	}
	return false
}

// HandleMouseIn processes mouse clicking on the checkbox with explicit bounds.
func (c *Checkbox) HandleMouseIn(msg tea.MouseMsg, bounds buffer.Rect) bool {
	if msg.Action != input.MousePress || msg.Button != input.MouseLeft {
		return false
	}
	if bounds.Contains(msg.X, msg.Y) {
		c.focused = true
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
	if c.indeterminate {
		marker = c.indeterminateMarker
	} else if c.checked {
		marker = c.checkedMarker
	}
	text := marker + c.label

	c.bounds = buffer.NewRect(area.X, area.Y, min(area.Width, buffer.StringWidth(text)), 1)

	fg := c.styleUnchecked.GetFg()
	bg := c.styleUnchecked.GetBg()
	attrs := cell.AttrNone

	if c.checked || c.indeterminate {
		fg = c.styleChecked.GetFg()
		bg = c.styleChecked.GetBg()
		attrs = cell.AttrBold
	}
	if c.focused {
		fg = c.styleFocused.GetFg()
		bg = c.styleFocused.GetBg()
		attrs = cell.AttrBold
	}

	buf.SetStringAligned(area, text, buffer.AlignLeft, fg, bg, attrs)
}
