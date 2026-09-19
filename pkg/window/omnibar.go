package window

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/theme"
	"github.com/baibeicha/goatui/pkg/widgets"
)

// OmniItem represents an actionable entry in the Omnibar search results.
type OmniItem struct {
	Title       string
	Description string
	Route       string
	Action      func() tea.Cmd
}

// ToggleOmnibarMsg toggles the Omnibar command palette overlay.
type ToggleOmnibarMsg struct{}

// ToggleOmnibar returns a command to open or close the Omnibar.
func ToggleOmnibar() tea.Msg {
	return ToggleOmnibarMsg{}
}

type filteredItem struct {
	item    OmniItem
	score   int
	indices []int
}

// Omnibar provides an interactive command palette and URL navigation bar.
type Omnibar struct {
	input          *widgets.TextInput
	items          []OmniItem
	filtered       []filteredItem
	selectedIndex  int
	scrollOffset   int
	visible        bool
	currentAddress string
	history        []string
	historyIdx     int
}

// NewOmnibar creates a new Omnibar instance.
func NewOmnibar() *Omnibar {
	ti := widgets.NewTextInput()
	ti.SetPrompt("> ")
	ti.SetPlaceholder("Type a URL path (/...), file path, or command...")
	ti.Focus()

	return &Omnibar{
		input:   ti,
		visible: false,
	}
}

// SetCurrentAddress sets the active address/route displayed in the Omnibar header.
func (o *Omnibar) SetCurrentAddress(addr string) {
	o.currentAddress = addr
}

// CurrentAddress returns the active address.
func (o *Omnibar) CurrentAddress() string {
	return o.currentAddress
}

// SetItems registers the static command and route list.
func (o *Omnibar) SetItems(items []OmniItem) {
	o.items = items
	o.filter()
}

// AddRoute adds a router path to the Omnibar entries.
func (o *Omnibar) AddRoute(route string, description string) {
	o.items = append(o.items, OmniItem{
		Title:       route,
		Description: description,
		Route:       route,
	})
	o.filter()
}

// AddItem registers an arbitrary command or navigation item to the Omnibar.
func (o *Omnibar) AddItem(item OmniItem) {
	o.items = append(o.items, item)
	o.filter()
}

// IsVisible returns whether the Omnibar is currently open.
func (o *Omnibar) IsVisible() bool {
	return o.visible
}

// Items returns a copy of all registered OmniItems.
func (o *Omnibar) Items() []OmniItem {
	res := make([]OmniItem, len(o.items))
	copy(res, o.items)
	return res
}

// Open shows the Omnibar.
func (o *Omnibar) Open() {
	o.visible = true
	o.input.SetValue("")
	if o.currentAddress != "" {
		o.input.SetPlaceholder("Current: " + o.currentAddress + " (or type command)...")
	} else {
		o.input.SetPlaceholder("Type a URL path (/...), file path, or command...")
	}
	o.input.Focus()
	o.selectedIndex = 0
	o.scrollOffset = 0
	o.filter()
}

// Close hides the Omnibar.
func (o *Omnibar) Close() {
	o.visible = false
}

// Toggle toggles visibility.
func (o *Omnibar) Toggle() {
	if o.visible {
		o.Close()
	} else {
		o.Open()
	}
}

func fuzzyMatch(pattern, target string) (bool, int, []int) {
	if pattern == "" {
		return true, 0, nil
	}
	pRunes := []rune(strings.ToLower(pattern))
	tRunes := []rune(target)
	tLower := []rune(strings.ToLower(target))

	pIdx := 0
	score := 0
	var indices []int

	for tIdx, r := range tLower {
		if pIdx < len(pRunes) && r == pRunes[pIdx] {
			indices = append(indices, tIdx)

			score += 10
			if tIdx == 0 {
				score += 5
			} else if tIdx > 0 {
				prev := tRunes[tIdx-1]
				if prev == ' ' || prev == '/' || prev == '-' || prev == '_' {
					score += 5
				}
			}
			pIdx++
		}
	}

	return pIdx == len(pRunes), score, indices
}

func (o *Omnibar) filter() {
	query := strings.TrimSpace(o.input.Value())
	o.filtered = nil

	if query == "" {
		for _, it := range o.items {
			o.filtered = append(o.filtered, filteredItem{item: it})
		}
		o.selectedIndex = 0
		o.scrollOffset = 0
		return
	}

	for _, it := range o.items {
		matchTitle, scoreTitle, indicesTitle := fuzzyMatch(query, it.Title)
		matchDesc, scoreDesc, _ := fuzzyMatch(query, it.Description)

		if matchTitle || matchDesc {
			score := scoreTitle
			if scoreDesc > score {
				score = scoreDesc
			}
			var ind []int
			if matchTitle {
				ind = indicesTitle
			}
			o.filtered = append(o.filtered, filteredItem{
				item:    it,
				score:   score,
				indices: ind,
			})
		}
	}

	// Sort by score
	for i := 0; i < len(o.filtered); i++ {
		for j := i + 1; j < len(o.filtered); j++ {
			if o.filtered[j].score > o.filtered[i].score {
				o.filtered[i], o.filtered[j] = o.filtered[j], o.filtered[i]
			}
		}
	}

	if o.selectedIndex >= len(o.filtered) {
		o.selectedIndex = max(0, len(o.filtered)-1)
	}
	o.scrollOffset = 0
}

func (o *Omnibar) calcMaxVisible(area buffer.Rect) int {
	maxVisible := 8
	if area.Height >= 30 {
		maxVisible = 12
	} else if area.Height < 20 {
		maxVisible = 5
	}
	if area.Height > 0 && maxVisible > area.Height-6 {
		maxVisible = max(1, area.Height-6)
	}
	return min(maxVisible, len(o.filtered))
}

// Update handles keyboard navigation and selection inside the Omnibar.
func (o *Omnibar) Update(msg tea.Msg) (bool, tea.Cmd) {
	if !o.visible {
		return false, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key.Type {
		case input.KeyEsc:
			o.Close()
			return true, nil

		case input.KeyUp:
			if (o.selectedIndex == 0 || o.input.Value() == "") && len(o.history) > 0 {
				if o.historyIdx > 0 {
					o.historyIdx--
				}
				if o.historyIdx >= 0 && o.historyIdx < len(o.history) {
					o.input.SetValue(o.history[o.historyIdx])
					o.filter()
				}
				return true, nil
			}
			if o.selectedIndex > 0 {
				o.selectedIndex--
				if o.selectedIndex < o.scrollOffset {
					o.scrollOffset = o.selectedIndex
				}
			}
			return true, nil

		case input.KeyDown:
			if o.historyIdx < len(o.history) {
				o.historyIdx++
				if o.historyIdx == len(o.history) {
					o.input.SetValue("")
					o.filter()
				} else {
					o.input.SetValue(o.history[o.historyIdx])
					o.filter()
				}
				return true, nil
			}
			if o.selectedIndex < len(o.filtered)-1 {
				o.selectedIndex++
				maxVis := 8
				if o.selectedIndex >= o.scrollOffset+maxVis {
					o.scrollOffset = o.selectedIndex - maxVis + 1
				}
			}
			return true, nil

		case input.KeyPgUp:
			o.selectedIndex -= 5
			if o.selectedIndex < 0 {
				o.selectedIndex = 0
			}
			if o.selectedIndex < o.scrollOffset {
				o.scrollOffset = o.selectedIndex
			}
			return true, nil

		case input.KeyPgDown:
			o.selectedIndex += 5
			if o.selectedIndex >= len(o.filtered) {
				o.selectedIndex = max(0, len(o.filtered)-1)
			}
			maxVis := 8
			if o.selectedIndex >= o.scrollOffset+maxVis {
				o.scrollOffset = o.selectedIndex - maxVis + 1
			}
			return true, nil

		case input.KeyTab:
			val := o.input.Value()
			if strings.HasPrefix(val, "/") || strings.HasPrefix(val, "./") || strings.HasPrefix(val, "file://") {
				// Autocomplete path
				cleanVal := strings.TrimPrefix(val, "file://")
				matches, _ := filepath.Glob(cleanVal + "*")
				if len(matches) > 0 {
					// common prefix in runes
					prefix := []rune(matches[0])
					for _, m := range matches[1:] {
						mRunes := []rune(m)
						for len(prefix) > 0 && !strings.HasPrefix(string(mRunes), string(prefix)) {
							prefix = prefix[:len(prefix)-1]
						}
					}
					res := string(prefix)
					if fi, err := os.Stat(res); err == nil && fi.IsDir() && !strings.HasSuffix(res, "/") && !strings.HasSuffix(res, "\\") {
						res += "/"
					}
					if strings.HasPrefix(val, "file://") {
						res = "file://" + res
					}
					o.input.SetValue(filepath.ToSlash(res))
					o.filter()
				}
			}
			return true, nil

		case input.KeyEnter:
			val := strings.TrimSpace(o.input.Value())
			if val != "" {
				if len(o.history) == 0 || o.history[len(o.history)-1] != val {
					o.history = append(o.history, val)
				}
				o.historyIdx = len(o.history)
			}
			o.Close()

			// Check if typed directly as a local filesystem path:
			// e.g. "C:\...", "d:/...", "./...", "../...", "\..."
			isDrivePath := len(val) >= 2 && ((val[0] >= 'A' && val[0] <= 'Z') || (val[0] >= 'a' && val[0] <= 'z')) && val[1] == ':'
			isRelPath := strings.HasPrefix(val, ".") || strings.HasPrefix(val, "\\")
			if isDrivePath || isRelPath {
				cleanVal := filepath.ToSlash(val)
				if !strings.HasPrefix(cleanVal, "/") {
					cleanVal = "/" + cleanVal
				}
				return true, func() tea.Msg { return NavigateMsg{URL: "file://" + cleanVal} }
			}

			// If typed directly as a path (starts with '/' or schemes like file://, http://, https://)
			if strings.HasPrefix(val, "/") ||
				strings.HasPrefix(val, "file://") ||
				strings.HasPrefix(val, "http://") ||
				strings.HasPrefix(val, "https://") {
				return true, func() tea.Msg { return NavigateMsg{URL: val} }
			}

			// If selected from list
			if len(o.filtered) > 0 && o.selectedIndex >= 0 && o.selectedIndex < len(o.filtered) {
				item := o.filtered[o.selectedIndex].item
				if item.Action != nil {
					return true, item.Action()
				}
				if item.Route != "" {
					return true, func() tea.Msg { return NavigateMsg{URL: item.Route} }
				}
			}

			return true, nil

		default:
			if o.input.HandleKey(msg.Key) {
				o.filter()
				return true, nil
			}
		}
	case tea.PasteMsg:
		o.input.InsertString(msg.Text)
		o.filter()
		return true, nil
	}

	return false, nil
}

// HandleMouse processes mouse wheel scrolling, item clicking, and outside-click dismissal.
func (o *Omnibar) HandleMouse(msg tea.MouseMsg, area buffer.Rect) (bool, tea.Cmd) {
	if !o.visible {
		return false, nil
	}

	boxArea := o.Bounds(area)

	// Outside click closes omnibar
	if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
		if !boxArea.Contains(msg.X, msg.Y) {
			o.Close()
			return true, nil
		}
	}

	maxVisible := o.calcMaxVisible(area)

	// Mouse wheel scrolling through items
	if msg.Button == input.MouseWheelUp {
		if o.selectedIndex > 0 {
			o.selectedIndex--
			if o.selectedIndex < o.scrollOffset {
				o.scrollOffset = o.selectedIndex
			}
		}
		return true, nil
	}
	if msg.Button == input.MouseWheelDown {
		if o.selectedIndex < len(o.filtered)-1 {
			o.selectedIndex++
			if o.selectedIndex >= o.scrollOffset+maxVisible {
				o.scrollOffset = o.selectedIndex - maxVisible + 1
			}
		}
		return true, nil
	}

	// Left click on an item row selects and executes it
	if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
		inner := boxArea.Inset(1, 1)
		itemsStartY := inner.Y + 2
		if msg.Y >= itemsStartY && msg.Y < itemsStartY+maxVisible {
			rowIdx := msg.Y - itemsStartY
			targetIdx := o.scrollOffset + rowIdx
			if targetIdx >= 0 && targetIdx < len(o.filtered) {
				o.selectedIndex = targetIdx
				item := o.filtered[targetIdx].item
				o.Close()
				if item.Action != nil {
					return true, item.Action()
				}
				if item.Route != "" {
					return true, func() tea.Msg { return NavigateMsg{URL: item.Route} }
				}
				return true, nil
			}
		}
	}

	return false, nil
}

// Bounds calculates the centered bounding box of the Omnibar window.
func (o *Omnibar) Bounds(area buffer.Rect) buffer.Rect {
	w := min(72, area.Width-4)
	visibleCount := o.calcMaxVisible(area)
	h := 3 + visibleCount
	if visibleCount > 0 {
		h++ // divider
	}

	x := area.X + max(0, (area.Width-w)/2)
	y := area.Y + max(1, (area.Height-h)/3)
	return buffer.NewRect(x, y, w, h)
}

// View draws the Omnibar modal overlay.
func (o *Omnibar) View(f *tea.Frame) {
	if !o.visible {
		return
	}

	area := f.Area()
	if area.IsEmpty() {
		return
	}

	boxArea := o.Bounds(area)
	maxVisible := o.calcMaxVisible(area)

	curTheme := theme.Default().Current()
	p := curTheme.Colors
	bg := p.Background
	if bg.IsDefault() {
		bg = cell.Color256(234)
	}
	borderFg := p.Primary
	if borderFg.IsDefault() {
		borderFg = cell.ColorHex("#7D56F4")
	}
	titleFg := p.Accent
	if titleFg.IsDefault() {
		titleFg = cell.ColorHex("#00FFAA")
	}
	selBg := p.Secondary
	if selBg.IsDefault() {
		selBg = cell.Color256(236)
	}
	selFg := p.Foreground
	if selFg.IsDefault() {
		selFg = cell.ColorHex("#FFFFFF")
	}
	itemFg := p.Foreground
	if itemFg.IsDefault() {
		itemFg = cell.ColorHex("#CCCCCC")
	}
	descFg := p.Muted
	if descFg.IsDefault() {
		descFg = cell.ColorHex("#777799")
	}
	matchFg := p.Warning
	if matchFg.IsDefault() {
		matchFg = cell.ColorHex("#FFD700")
	}

	// Fill boxArea completely with solid opaque background
	for cy := boxArea.Y; cy < boxArea.Bottom(); cy++ {
		for cx := boxArea.X; cx < boxArea.Right(); cx++ {
			f.Buffer.Set(cx, cy, cell.Cell{
				Rune:     ' ',
				Width:    1,
				Modifier: cell.AttrNone,
				FgType:   cell.ColorDefault,
				BgType:   bg.Type,
				Bg:       bg.Value,
			})
		}
	}

	// Draw outer box
	st := style.NewStyle().
		Border(style.BorderRounded).
		BorderForeground(borderFg).
		Background(bg)
	st.Draw(f.Buffer, boxArea, "")

	// Display title with current address and item counter on top border
	title := " Command Palette "
	if o.currentAddress != "" {
		title = fmt.Sprintf(" Address: %s ", o.currentAddress)
	}
	if len(o.filtered) > 0 {
		title += fmt.Sprintf("[%d/%d] ", o.selectedIndex+1, len(o.filtered))
	}
	titleW := buffer.StringWidth(title)
	if boxArea.Width > titleW+4 {
		titleX := boxArea.X + 2
		f.Buffer.SetString(titleX, boxArea.Y, title, titleFg, bg, cell.AttrBold)
	}

	inner := boxArea.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// 1. Draw Search Input
	inputArea := buffer.NewRect(inner.X, inner.Y, inner.Width, 1)
	o.input.Draw(f.Buffer, inputArea)

	// 2. Divider if items exist
	if maxVisible > 0 {
		sepY := inner.Y + 1
		for cx := inner.X; cx < inner.Right(); cx++ {
			f.Buffer.SetRune(cx, sepY, '─', borderFg, bg, cell.AttrDim)
		}

		// Ensure scroll offset bounds
		if o.selectedIndex < o.scrollOffset {
			o.scrollOffset = o.selectedIndex
		}
		if o.selectedIndex >= o.scrollOffset+maxVisible {
			o.scrollOffset = max(0, o.selectedIndex-maxVisible+1)
		}
		if o.scrollOffset > max(0, len(o.filtered)-maxVisible) {
			o.scrollOffset = max(0, len(o.filtered)-maxVisible)
		}

		// 3. Draw filtered results within visible window
		for i := 0; i < maxVisible; i++ {
			idx := o.scrollOffset + i
			if idx >= len(o.filtered) {
				break
			}
			fItem := o.filtered[idx]
			item := fItem.item
			itemY := sepY + 1 + i
			isSelected := (idx == o.selectedIndex)

			rowFg := itemFg
			rowBg := bg
			mod := cell.Modifier(0)
			prefix := "  "

			if isSelected {
				rowFg = selFg
				rowBg = selBg
				mod = cell.AttrBold
				prefix = "> "
			}

			// Pre-fill background
			for cx := inner.X; cx < inner.Right(); cx++ {
				f.Buffer.Set(cx, itemY, cell.Cell{
					Rune:     ' ',
					Width:    1,
					FgType:   rowFg.Type,
					BgType:   rowBg.Type,
					Fg:       rowFg.Value,
					Bg:       rowBg.Value,
					Modifier: mod,
				})
			}

			// Title
			titleRunes := []rune(prefix + item.Title)
			prefixRuneCount := len([]rune(prefix))

			// Highlight matching runes
			cx := inner.X
			for rIdx, r := range titleRunes {
				if cx >= inner.Right() {
					break
				}

				isMatch := false
				if rIdx >= prefixRuneCount {
					origRuneIdx := rIdx - prefixRuneCount
					for _, mIdx := range fItem.indices {
						if mIdx == origRuneIdx {
							isMatch = true
							break
						}
					}
				}

				rMod := mod
				rFg := rowFg
				if isMatch {
					rMod = cell.AttrBold
					rFg = matchFg
				}

				w := buffer.RuneWidth(r)
				f.Buffer.SetRune(cx, itemY, r, rFg, rowBg, rMod)
				cx += w
			}
			titleEnd := cx

			// Description on right
			if item.Description != "" {
				descW := buffer.StringWidth(item.Description)
				descX := inner.Right() - descW - 1
				if descX > titleEnd+2 {
					f.Buffer.SetString(descX, itemY, item.Description, descFg, rowBg, cell.AttrNone)
				}
			}
		}
	}
}
