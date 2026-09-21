package goatui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/ui"
	"github.com/baibeicha/goatui/pkg/window"
)

type appTickMsg time.Time

// AppTab represents an individual navigation tab in the App scaffold.
type AppTab struct {
	ID           string
	Title        string
	View         func(f *tea.Frame, area buffer.Rect)
	Screen       window.Screen
	OnKey        func(key input.Key) bool
	OnMouse      func(msg tea.MouseMsg) bool
	interceptTab bool
}

// SetOnKey sets a keyboard handler for this tab.
func (t *AppTab) SetOnKey(fn func(key input.Key) bool) *AppTab {
	t.OnKey = fn
	return t
}

// SetOnMouse sets a mouse event handler for this tab.
func (t *AppTab) SetOnMouse(fn func(msg tea.MouseMsg) bool) *AppTab {
	t.OnMouse = fn
	return t
}

// SetInterceptTab configures whether this tab intercepts Tab/Shift+Tab keys.
// By default false, allowing App to cycle tabs via Tab / Shift+Tab.
func (t *AppTab) SetInterceptTab(intercept bool) *AppTab {
	t.interceptTab = intercept
	return t
}

// InterceptTab returns whether this tab intercepts Tab navigation keys.
func (t *AppTab) InterceptTab() bool {
	return t.interceptTab
}

// App provides an out-of-the-box, batteries-included scaffold for terminal applications.
// It bundles tabs, hotkey/mouse routing, modal dialogs, statusbar, and floating toasts.
type App struct {
	title         string
	tabs          []*AppTab
	activeTabIdx  int
	tabBounds     []buffer.Rect
	toasts        *ui.ToastManager
	modal         window.Modal
	isHelpOpen    bool
	statusLeft    string
	statusRight   string
	keyHints      *ui.KeyHints
	hotkeys       map[rune]func()
	specialKeys   map[input.KeyType]func()
	mouseHandler  func(msg tea.MouseMsg) bool
	program       *tea.Program
	brandStyle    style.Style
	headerStyle   style.Style
	activeTabSt   style.Style
	inactiveTabSt style.Style
	footerStyle   style.Style
}

// NewApp creates a new high-level App scaffold.
func NewApp(title string) *App {
	a := &App{
		title:        title,
		tabs:         make([]*AppTab, 0),
		activeTabIdx: 0,
		toasts:       ui.NewToastManager(5),
		hotkeys:      make(map[rune]func()),
		specialKeys:  make(map[input.KeyType]func()),
		brandStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00D2FF")),
		headerStyle: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")).
			Background(cell.Color256(235)),
		activeTabSt: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#000000")).
			Background(cell.ColorHex("#00FFAA")),
		inactiveTabSt: style.NewStyle().
			Foreground(cell.ColorHex("#8888AA")).
			Background(cell.Color256(235)),
		footerStyle: style.NewStyle().
			Foreground(cell.ColorHex("#AAAAAA")).
			Background(cell.Color256(236)),
	}

	return a
}

// AddTab adds a tab rendered via a functional view callback.
func (a *App) AddTab(id, title string, view func(f *tea.Frame, area buffer.Rect)) *App {
	a.tabs = append(a.tabs, &AppTab{
		ID:    id,
		Title: title,
		View:  view,
	})
	return a
}

// AddTabScreen adds a tab backed by a full window.Screen.
func (a *App) AddTabScreen(id, title string, s window.Screen) *App {
	a.tabs = append(a.tabs, &AppTab{
		ID:     id,
		Title:  title,
		Screen: s,
	})
	return a
}

// SetActiveTab sets the currently active tab by ID.
func (a *App) SetActiveTab(id string) *App {
	for i, tab := range a.tabs {
		if tab.ID == id {
			a.activeTabIdx = i
			break
		}
	}
	return a
}

// ActiveTab returns the ID of the active tab, or empty string.
func (a *App) ActiveTab() string {
	if a.activeTabIdx >= 0 && a.activeTabIdx < len(a.tabs) {
		return a.tabs[a.activeTabIdx].ID
	}
	return ""
}

// ActiveTabIndex returns the index of the active tab.
func (a *App) ActiveTabIndex() int {
	return a.activeTabIdx
}

// Toast pushes a floating toast notification.
func (a *App) Toast(level ui.ToastLevel, title, message string, duration ...time.Duration) {
	a.toasts.Add(level, title, message, duration...)
}

// ToastInfo pushes an informational toast.
func (a *App) ToastInfo(title, message string) {
	a.toasts.Info(title, message)
}

// ToastSuccess pushes a success toast.
func (a *App) ToastSuccess(title, message string) {
	a.toasts.Success(title, message)
}

// ToastWarn pushes a warning toast.
func (a *App) ToastWarn(title, message string) {
	a.toasts.Warn(title, message)
}

// ToastError pushes an error toast.
func (a *App) ToastError(title, message string) {
	a.toasts.Error(title, message)
}

// Alert opens an alert modal dialog.
func (a *App) Alert(title, message string, onDismiss func()) {
	a.modal = window.AlertModal(title, message, func() tea.Cmd {
		if onDismiss != nil {
			onDismiss()
		}
		return nil
	})
}

// Confirm opens a confirmation modal dialog.
func (a *App) Confirm(title, message string, onConfirm, onCancel func()) {
	a.modal = window.ConfirmModal(title, message, func() tea.Cmd {
		if onConfirm != nil {
			onConfirm()
		}
		return nil
	}, func() tea.Cmd {
		if onCancel != nil {
			onCancel()
		}
		return nil
	})
}

// Prompt opens a text input modal dialog.
func (a *App) Prompt(title, prompt string, onConfirm func(val string), onCancel func()) {
	a.modal = window.NewInputModal(title, prompt, "", func(val string) tea.Cmd {
		if onConfirm != nil {
			onConfirm(val)
		}
		return nil
	}, func() tea.Cmd {
		if onCancel != nil {
			onCancel()
		}
		return nil
	})
}

// CustomModal displays a custom window.Modal dialog.
func (a *App) CustomModal(m window.Modal) {
	a.modal = m
}

// CloseModal dismisses the active modal dialog.
func (a *App) CloseModal() {
	a.modal = nil
	a.isHelpOpen = false
}

// ShowHelp displays a help dialog listing navigation keys, key hints, and hotkeys.
func (a *App) ShowHelp() {
	var lines []string
	lines = append(lines, "NAVIGATION:")
	lines = append(lines, "  [Tab] / [Shift+Tab]   Cycle through tabs")
	if len(a.tabs) > 0 {
		lines = append(lines, fmt.Sprintf("  [1]..[%d]              Switch directly to tab", min(9, len(a.tabs))))
	}
	lines = append(lines, "  [Ctrl+Tab]            Cycle forward through tabs")

	if a.keyHints != nil && len(a.keyHints.Hints()) > 0 {
		lines = append(lines, "")
		lines = append(lines, "KEY SHORTCUTS:")
		for _, kh := range a.keyHints.Hints() {
			lines = append(lines, fmt.Sprintf("  [%-16s] %s", kh.Key, kh.Desc))
		}
	} else if len(a.hotkeys) > 0 {
		lines = append(lines, "")
		lines = append(lines, "HOTKEYS:")
		var runes []rune
		for r := range a.hotkeys {
			runes = append(runes, r)
		}
		sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
		for _, r := range runes {
			lines = append(lines, fmt.Sprintf("  [%c]                  Trigger action", r))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "SYSTEM:")
	lines = append(lines, "  [?] / [F1]            Toggle this help dialog")
	lines = append(lines, "  [Esc] / [Enter]       Dismiss dialog")
	lines = append(lines, "  [Ctrl+Q]              Quit application")

	message := strings.Join(lines, "\n")
	a.isHelpOpen = true
	a.modal = window.AlertModal("KEYBOARD SHORTCUTS & HELP", message, func() tea.Cmd {
		a.isHelpOpen = false
		return nil
	})
}

// ToggleHelp toggles the display of the keyboard shortcuts help dialog.
func (a *App) ToggleHelp() {
	if a.isHelpOpen && a.modal != nil {
		a.CloseModal()
		return
	}
	a.ShowHelp()
}

// SetStatus sets left and right statusbar messages.
func (a *App) SetStatus(left, right string) *App {
	a.statusLeft = left
	a.statusRight = right
	return a
}

// SetKeyHints sets the key hints toolbar displayed in the footer.
func (a *App) SetKeyHints(hints ...ui.KeyHint) *App {
	a.keyHints = ui.NewKeyHints(hints...)
	return a
}

// OnKey registers a handler for a special key.
func (a *App) OnKey(keyType input.KeyType, fn func()) *App {
	a.specialKeys[keyType] = fn
	return a
}

// OnKeyRune registers a global hotkey handler for a character.
func (a *App) OnKeyRune(r rune, fn func()) *App {
	a.hotkeys[r] = fn
	return a
}

// OnMouse registers a global mouse event handler.
func (a *App) OnMouse(fn func(msg tea.MouseMsg) bool) *App {
	a.mouseHandler = fn
	return a
}

// Tab returns the tab with the given ID, or nil if not found.
func (a *App) Tab(id string) *AppTab {
	for _, t := range a.tabs {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// SetTabOnKey registers a key handler for the tab with the given ID.
func (a *App) SetTabOnKey(id string, fn func(key input.Key) bool) *App {
	if t := a.Tab(id); t != nil {
		t.OnKey = fn
	}
	return a
}

// SetTabOnMouse registers a mouse handler for the tab with the given ID.
func (a *App) SetTabOnMouse(id string, fn func(msg tea.MouseMsg) bool) *App {
	if t := a.Tab(id); t != nil {
		t.OnMouse = fn
	}
	return a
}

// SetTabInterceptTab configures whether the tab with the given ID intercepts Tab keys.
func (a *App) SetTabInterceptTab(id string, intercept bool) *App {
	if t := a.Tab(id); t != nil {
		t.SetInterceptTab(intercept)
	}
	return a
}

// Quit terminates the application.
func (a *App) Quit() {
	if a.program != nil {
		a.program.Quit()
	}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return appTickMsg(t)
	})
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case appTickMsg:
		a.toasts.Tick(time.Now())
		return a, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return appTickMsg(t)
		})

	case window.CloseModalMsg:
		a.modal = nil
		a.isHelpOpen = false
		return a, nil

	case window.ShowModalMsg:
		a.modal = msg.Modal
		a.isHelpOpen = false
		return a, nil
	}

	// Modal handles events first
	if a.modal != nil {
		if km, ok := msg.(tea.KeyMsg); ok {
			if a.isHelpOpen && (km.Key.Type == input.KeyEsc || (km.Key.Type == input.KeyRune && (km.Key.Rune == '?' || km.Key.Rune == 'q'))) {
				a.CloseModal()
				return a, nil
			}
		}
		var cmd tea.Cmd
		var scr window.Screen
		scr, cmd = a.modal.Update(msg)
		if m, ok := scr.(window.Modal); ok {
			a.modal = m
		}
		return a, cmd
	}

	// Keyboard handling
	if km, ok := msg.(tea.KeyMsg); ok {
		// Default Ctrl+Q or Ctrl+C quit
		if km.Key.Type == input.KeyRune && km.Key.HasCtrl() && (km.Key.Rune == 'q' || km.Key.Rune == 'c') {
			a.Quit()
			return a, tea.Quit
		}

		// Tab cycling with Ctrl modifier: Ctrl+Tab advances, Ctrl+Shift+Tab or Ctrl+Backtab goes back
		if (km.Key.Type == input.KeyTab || km.Key.Type == input.KeyBacktab) && km.Key.Mod.Has(input.ModCtrl) {
			if len(a.tabs) > 0 {
				if km.Key.Type == input.KeyBacktab || km.Key.Mod.Has(input.ModShift) {
					a.activeTabIdx = (a.activeTabIdx - 1 + len(a.tabs)) % len(a.tabs)
				} else {
					a.activeTabIdx = (a.activeTabIdx + 1) % len(a.tabs)
				}
			}
			return a, nil
		}

		// Direct tab navigation via Alt+1..Alt+9
		if km.Key.Type == input.KeyRune && km.Key.Mod.Has(input.ModAlt) && km.Key.Rune >= '1' && km.Key.Rune <= '9' {
			idx := int(km.Key.Rune - '1')
			if idx < len(a.tabs) {
				a.activeTabIdx = idx
			}
			return a, nil
		}

		// Check if active tab intercepts Tab keys
		var activeTab *AppTab
		if a.activeTabIdx >= 0 && a.activeTabIdx < len(a.tabs) {
			activeTab = a.tabs[a.activeTabIdx]
		}
		tabInterceptsTab := activeTab != nil && activeTab.interceptTab

		// Plain Tab cycling: Tab advances, Shift+Tab or Backtab goes back (unless tab intercepts Tab)
		if (km.Key.Type == input.KeyTab || km.Key.Type == input.KeyBacktab) && !tabInterceptsTab {
			if len(a.tabs) > 0 {
				if km.Key.Type == input.KeyBacktab || km.Key.Mod.Has(input.ModShift) {
					a.activeTabIdx = (a.activeTabIdx - 1 + len(a.tabs)) % len(a.tabs)
				} else {
					a.activeTabIdx = (a.activeTabIdx + 1) % len(a.tabs)
				}
			}
			return a, nil
		}

		// Active tab key handler is given opportunity to consume keys (e.g. text input, hotkeys)
		if activeTab != nil && activeTab.OnKey != nil && activeTab.OnKey(km.Key) {
			return a, nil
		}

		// Registered special keys
		if fn, ok := a.specialKeys[km.Key.Type]; ok {
			fn()
			return a, nil
		}

		// Plain Tab cycling fallback if intercepting tab did not consume Tab
		if km.Key.Type == input.KeyTab || km.Key.Type == input.KeyBacktab {
			if len(a.tabs) > 0 {
				if km.Key.Type == input.KeyBacktab || km.Key.Mod.Has(input.ModShift) {
					a.activeTabIdx = (a.activeTabIdx - 1 + len(a.tabs)) % len(a.tabs)
				} else {
					a.activeTabIdx = (a.activeTabIdx + 1) % len(a.tabs)
				}
			}
			return a, nil
		}

		// Registered character hotkeys
		if km.Key.Type == input.KeyRune && !km.Key.HasCtrl() && !km.Key.HasAlt() {
			if fn, ok := a.hotkeys[km.Key.Rune]; ok {
				fn()
				return a, nil
			}
		}

		// Tab switching via plain numeric keys 1..9
		if km.Key.Type == input.KeyRune && !km.Key.HasCtrl() && !km.Key.HasAlt() && km.Key.Rune >= '1' && km.Key.Rune <= '9' {
			idx := int(km.Key.Rune - '1')
			if idx < len(a.tabs) {
				a.activeTabIdx = idx
				return a, nil
			}
		}

		// Help toggle via '?' or F1
		if (km.Key.Type == input.KeyRune && km.Key.Rune == '?' && !km.Key.HasCtrl() && !km.Key.HasAlt()) || km.Key.Type == input.KeyF1 {
			a.ToggleHelp()
			return a, nil
		}

		// Tab screen fallback
		if activeTab != nil && activeTab.Screen != nil {
			var cmd tea.Cmd
			activeTab.Screen, cmd = activeTab.Screen.Update(msg)
			return a, cmd
		}
	}

	// Mouse handling
	if mm, ok := msg.(tea.MouseMsg); ok {
		if mm.Action == input.MousePress && mm.Button == input.MouseLeft {
			for i, b := range a.tabBounds {
				if b.Contains(mm.X, mm.Y) {
					a.activeTabIdx = i
					return a, nil
				}
			}
		}

		// Global mouse handler
		if a.mouseHandler != nil && a.mouseHandler(mm) {
			return a, nil
		}

		// Delegate to active tab mouse handler or screen
		if a.activeTabIdx >= 0 && a.activeTabIdx < len(a.tabs) {
			tab := a.tabs[a.activeTabIdx]
			if tab.OnMouse != nil && tab.OnMouse(mm) {
				return a, nil
			}
			if tab.Screen != nil {
				var cmd tea.Cmd
				tab.Screen, cmd = tab.Screen.Update(msg)
				return a, cmd
			}
		}
	}

	return a, nil
}

// View implements tea.Model.
func (a *App) View(f *tea.Frame) {
	area := f.Area()
	if area.IsEmpty() {
		return
	}

	buf := f.Buffer

	// 1. Header (Y=0)
	brandText := fmt.Sprintf(" %s ", a.title)
	brandW := buffer.StringWidth(brandText)
	brandRect := buffer.NewRect(area.X, area.Y, brandW, 1)
	a.brandStyle.Draw(buf, brandRect, brandText)

	// Draw tabs
	currX := area.X + brandW + 1
	a.tabBounds = make([]buffer.Rect, len(a.tabs))

	for i, tab := range a.tabs {
		if currX >= area.Right() {
			break
		}

		isActive := (i == a.activeTabIdx)
		tabText := fmt.Sprintf(" [%d] %s ", i+1, tab.Title)
		tabW := buffer.StringWidth(tabText)

		if currX+tabW > area.Right() {
			tabW = area.Right() - currX
		}
		if tabW <= 0 {
			break
		}

		tabRect := buffer.NewRect(currX, area.Y, tabW, 1)
		a.tabBounds[i] = tabRect

		st := a.inactiveTabSt
		if isActive {
			st = a.activeTabSt
		}
		st.Draw(buf, tabRect, tabText)

		currX += tabW + 1
	}

	// Fill remaining header
	if currX < area.Right() {
		remArea := buffer.NewRect(currX, area.Y, area.Right()-currX, 1)
		a.headerStyle.Draw(buf, remArea, "")
	}

	// Divider below header (Y=1) if enough height
	if area.Height >= 3 {
		for x := area.X; x < area.Right(); x++ {
			buf.SetRune(x, area.Y+1, '─', cell.ColorHex("#333344"), cell.DefaultColor(), cell.AttrNone)
		}
	}

	// 2. Footer (Bottom line)
	if area.Height >= 2 {
		footerY := area.Bottom() - 1
		footerRect := buffer.NewRect(area.X, footerY, area.Width, 1)
		a.footerStyle.Draw(buf, footerRect, "")

		if a.statusLeft != "" {
			buf.SetStringAligned(buffer.NewRect(area.X+1, footerY, max(0, area.Width-2), 1), a.statusLeft, buffer.AlignLeft, cell.ColorHex("#00FFAA"), cell.Color256(236), cell.AttrNone)
		}

		if a.keyHints != nil {
			khW := area.Width / 2
			khArea := buffer.NewRect(area.Right()-khW-1, footerY, khW, 1)
			a.keyHints.Draw(buf, khArea)
		} else if a.statusRight != "" {
			rightW := buffer.StringWidth(a.statusRight)
			if rightW > 0 && area.Width >= rightW+2 {
				buf.SetStringAligned(buffer.NewRect(area.Right()-rightW-1, footerY, rightW, 1), a.statusRight, buffer.AlignRight, cell.ColorHex("#8888AA"), cell.Color256(236), cell.AttrNone)
			}
		}
	}

	// 3. Main Content Area (Between Y=2 and Bottom()-1)
	bodyH := max(0, area.Height-3)
	bodyArea := buffer.NewRect(area.X, area.Y+2, area.Width, bodyH)

	if !bodyArea.IsEmpty() && a.activeTabIdx >= 0 && a.activeTabIdx < len(a.tabs) {
		tab := a.tabs[a.activeTabIdx]
		if tab.View != nil {
			tab.View(f, bodyArea)
		} else if tab.Screen != nil {
			tab.Screen.View(f)
		}
	}

	// 3.5 Render Buffer Overlays (popups, dropdowns) before modal dialogs and toasts
	buf.RenderOverlays()

	// 4. Modal Dialog Overlay
	if a.modal != nil {
		if a.modal.BackdropDim() {
			// Dim background
			for y := area.Y; y < area.Bottom(); y++ {
				for x := area.X; x < area.Right(); x++ {
					c := buf.Cell(x, y)
					if c != nil {
						c.Modifier.Add(cell.AttrDim)
						buf.Set(x, y, *c)
					}
				}
			}
		}
		a.modal.View(f)
	}

	// 5. Floating Toasts
	a.toasts.Draw(buf, area)
}

// Run starts the application event loop with options.
func (a *App) Run(ctx ...context.Context) error {
	a.program = tea.NewProgram(a)
	_, err := a.program.Run(ctx...)
	return err
}
