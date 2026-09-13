package window

import (
	"sync"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/tea"
)

// NavigateMsg requests navigation to a URL route.
type NavigateMsg struct {
	URL string
}

// Navigate returns a command to navigate to the specified URL route.
func Navigate(url string) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{URL: url}
	}
}

// NavigateReplaceMsg requests replacing the current top screen with a new URL route.
type NavigateReplaceMsg struct {
	URL string
}

// NavigateReplace returns a command to replace the current screen with the specified URL route.
func NavigateReplace(url string) tea.Cmd {
	return func() tea.Msg {
		return NavigateReplaceMsg{URL: url}
	}
}

// PopMsg requests popping the current top screen from the navigation stack.
type PopMsg struct{}

// Pop returns a command to go back to the previous screen.
func Pop() tea.Msg {
	return PopMsg{}
}

// WindowManager coordinates the screen navigation stack, modal overlays, URL router, and Omnibar.
// It implements tea.Model, so it can be passed directly to goatui.NewProgram(wm).
type WindowManager struct {
	mu          sync.RWMutex
	router      *router.Router
	stack       []Screen
	activeModal Modal
	omnibar       *Omnibar
	notFound      Screen
	currentRoute  string
	redirectCount int
}

// NewWindowManager creates a new window and navigation manager.
func NewWindowManager(r *router.Router) *WindowManager {
	ob := NewOmnibar()

	// Register known router routes into Omnibar
	if r != nil {
		for _, pattern := range r.Routes() {
			ob.AddRoute(pattern, "Go to "+pattern)
		}
	}

	return &WindowManager{
		router:   r,
		stack:    make([]Screen, 0, 8),
		omnibar:  ob,
		notFound: &defaultNotFoundScreen{},
	}
}

// Router returns the underlying router.
func (wm *WindowManager) Router() *router.Router {
	return wm.router
}

// Omnibar returns the Omnibar instance.
func (wm *WindowManager) Omnibar() *Omnibar {
	return wm.omnibar
}

// SetNotFoundScreen sets custom screen rendered when route is not matched.
func (wm *WindowManager) SetNotFoundScreen(s Screen) *WindowManager {
	wm.notFound = s
	return wm
}

// ActiveScreen returns the currently visible top screen, or nil if stack is empty.
func (wm *WindowManager) ActiveScreen() Screen {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if len(wm.stack) == 0 {
		return nil
	}
	return wm.stack[len(wm.stack)-1]
}

// Push mounts a screen on top of the navigation stack.
func (wm *WindowManager) Push(screen Screen, ctx *router.RouteContext) tea.Cmd {
	wm.mu.Lock()
	if len(wm.stack) > 0 {
		top := wm.stack[len(wm.stack)-1]
		top.OnPause()
	}
	wm.stack = append(wm.stack, screen)
	wm.mu.Unlock()

	screen.OnMount(ctx)
	return screen.Init()
}

// Pop removes the top screen from the stack and resumes the previous screen.
func (wm *WindowManager) Pop() tea.Cmd {
	wm.mu.Lock()
	if len(wm.stack) <= 1 {
		wm.mu.Unlock()
		return nil // Don't pop root screen
	}

	topIdx := len(wm.stack) - 1
	top := wm.stack[topIdx]
	wm.stack[topIdx] = nil // explicit nil for GC
	wm.stack = wm.stack[:topIdx]
	prev := wm.stack[len(wm.stack)-1]
	wm.mu.Unlock()

	top.OnDestroy()
	prev.OnResume()
	return nil
}

// ShowModal displays a modal dialog with Focus Trap.
func (wm *WindowManager) ShowModal(m Modal) tea.Cmd {
	wm.mu.Lock()
	wm.activeModal = m
	wm.mu.Unlock()
	return m.Init()
}

// CloseModal dismisses the active modal dialog.
func (wm *WindowManager) CloseModal() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if wm.activeModal != nil {
		wm.activeModal.OnDestroy()
		wm.activeModal = nil
	}
}

// Init implements tea.Model.
func (wm *WindowManager) Init() tea.Cmd {
	active := wm.ActiveScreen()
	if active != nil {
		return active.Init()
	}
	return nil
}

// Update implements tea.Model.
func (wm *WindowManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 1. Global Navigation Messages
	switch msg := msg.(type) {
	case NavigateMsg:
		if wm.router != nil {
			handler, ctx, ok := wm.router.Match(msg.URL)
			if ok {
				wm.currentRoute = msg.URL
				res := handler(ctx)
				switch v := res.(type) {
				case Screen:
					wm.mu.Lock()
					if wm.activeModal != nil {
						m := wm.activeModal
						wm.activeModal = nil
						m.OnDestroy()
					}
					wm.mu.Unlock()
					cmd := wm.Push(v, ctx)
					return wm, cmd
				case router.RedirectMsg:
					if wm.redirectCount >= 8 {
						return wm, nil
					}
					wm.redirectCount++
					defer func() { wm.redirectCount-- }()
					return wm.Update(NavigateMsg{URL: v.URL})
				case NavigateMsg:
					return wm.Update(v)
				case NavigateReplaceMsg:
					return wm.Update(v)
				case ShowModalMsg:
					return wm.Update(v)
				case tea.Cmd:
					return wm, v
				case tea.Msg:
					return wm.Update(v)
				}
			} else if wm.notFound != nil {
				cmd := wm.Push(wm.notFound, ctx)
				return wm, cmd
			}
		}
		return wm, nil
		
	case NavigateReplaceMsg:
		if wm.router != nil {
			handler, ctx, ok := wm.router.Match(msg.URL)
			if ok {
				wm.currentRoute = msg.URL
				res := handler(ctx)
				switch v := res.(type) {
				case Screen:
					wm.mu.Lock()
					if wm.activeModal != nil {
						m := wm.activeModal
						wm.activeModal = nil
						m.OnDestroy()
					}
					var toDestroy Screen
					if len(wm.stack) > 0 {
						top := wm.stack[len(wm.stack)-1]
						if top != nil && top != v {
							toDestroy = top
						}
						wm.stack[len(wm.stack)-1] = v
					} else {
						wm.stack = append(wm.stack, v)
					}
					wm.mu.Unlock()
					if toDestroy != nil {
						toDestroy.OnDestroy()
					}
					v.OnMount(ctx)
					return wm, v.Init()
				case router.RedirectMsg:
					if wm.redirectCount >= 8 {
						return wm, nil
					}
					wm.redirectCount++
					defer func() { wm.redirectCount-- }()
					return wm.Update(NavigateReplaceMsg{URL: v.URL})
				case NavigateMsg:
					return wm.Update(v)
				case NavigateReplaceMsg:
					return wm.Update(v)
				case ShowModalMsg:
					return wm.Update(v)
				case tea.Cmd:
					return wm, v
				case tea.Msg:
					return wm.Update(v)
				}
			} else if wm.notFound != nil {
				wm.mu.Lock()
				if len(wm.stack) > 0 {
					top := wm.stack[len(wm.stack)-1]
					if top != nil {
						top.OnDestroy()
					}
					wm.stack[len(wm.stack)-1] = wm.notFound
				} else {
					wm.stack = append(wm.stack, wm.notFound)
				}
				wm.mu.Unlock()
				wm.notFound.OnMount(ctx)
				return wm, wm.notFound.Init()
			}
		}
		return wm, nil

	case PopMsg:
		cmd := wm.Pop()
		return wm, cmd

	case ShowModalMsg:
		cmd := wm.ShowModal(msg.Modal)
		return wm, cmd

	case CloseModalMsg:
		wm.CloseModal()
		return wm, nil

	case ToggleOmnibarMsg:
		wm.syncOmnibarAddress()
		wm.omnibar.Toggle()
		return wm, nil
	}

	// 2. Global Hotkey: Ctrl+K to toggle Omnibar
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Ctrl+K
		if keyMsg.Key.Type == 0 && keyMsg.Key.Rune == 'k' && keyMsg.Key.Mod.Has(1<<0) {
			wm.syncOmnibarAddress()
			wm.omnibar.Toggle()
			return wm, nil
		}
	}

	// Check if message is interactive user input (keyboard/mouse)
	isUserInput := false
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg, tea.HitMsg:
		isUserInput = true
	}

	if isUserInput {
		// 3. Focus Trap: If Omnibar is visible, it intercepts user input
		if wm.omnibar.IsVisible() {
			handled, cmd := wm.omnibar.Update(msg)
			if handled {
				return wm, cmd
			}
		}

		// 4. Focus Trap: If Modal is active, modal intercepts user input
		wm.mu.RLock()
		modal := wm.activeModal
		wm.mu.RUnlock()

		if modal != nil {
			updatedModal, cmd := modal.Update(msg)
			if m, ok := updatedModal.(Modal); ok {
				wm.mu.Lock()
				wm.activeModal = m
				wm.mu.Unlock()
			}
			return wm, cmd
		}
	}

	// 5. Background events (tickers, async results, theme changes) or unhandled input dispatched to active top screen
	active := wm.ActiveScreen()
	if active != nil {
		updatedScreen, cmd := active.Update(msg)
		wm.mu.Lock()
		if len(wm.stack) > 0 {
			wm.stack[len(wm.stack)-1] = updatedScreen
		}
		wm.mu.Unlock()
		return wm, cmd
	}

	return wm, nil
}

// View implements tea.Model.
func (wm *WindowManager) View(f *tea.Frame) {
	area := f.Area()
	if area.IsEmpty() {
		return
	}

	// 1. Render Active Top Screen
	active := wm.ActiveScreen()
	if active != nil {
		active.View(f)
	}

	// 2. Render Modal Overlay (with Backdrop Dimming)
	wm.mu.RLock()
	modal := wm.activeModal
	wm.mu.RUnlock()

	if modal != nil {
		if modal.BackdropDim() {
			dimBackdrop(f.Buffer)
		}

		modalBounds := modal.Bounds(area)
		if !modalBounds.IsEmpty() {
			// Clear modal area background with dark blank cells
			clearArea(f.Buffer, modalBounds)

			// Create a sub-frame for modal drawing
			modalFrame := &tea.Frame{
				Buffer:  f.Buffer,
				Spatial: f.Spatial,
			}
			modal.View(modalFrame)
		}
	}

	// 3. Render Omnibar Command Palette
	if wm.omnibar.IsVisible() {
		dimBackdrop(f.Buffer)
		clearArea(f.Buffer, wm.omnibar.Bounds(area))
		wm.omnibar.View(f)
	}
}

func dimBackdrop(buf *buffer.Buffer) {
	cells := buf.Cells()
	for i := range cells {
		cells[i].Modifier |= cell.AttrDim
		if cells[i].FgType == cell.ColorRGB {
			r := (cells[i].Fg >> 16) & 0xFF
			g := (cells[i].Fg >> 8) & 0xFF
			b := cells[i].Fg & 0xFF
			cells[i].Fg = ((r / 3) << 16) | ((g / 3) << 8) | (b / 3)
		} else if cells[i].FgType == cell.ColorANSI256 {
			cells[i].Fg = 238
		}
	}
}

func clearArea(buf *buffer.Buffer, r buffer.Rect) {
	bg := cell.Color256(234)
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			buf.Set(x, y, cell.Cell{
				Rune:     ' ',
				Width:    1,
				Modifier: cell.AttrNone,
				FgType:   cell.ColorDefault,
				BgType:   bg.Type,
				Bg:       bg.Value,
			})
		}
	}
}

type defaultNotFoundScreen struct {
	BaseScreen
	path string
}

func (d *defaultNotFoundScreen) OnMount(ctx *router.RouteContext) {
	d.path = ctx.Path
}

func (d *defaultNotFoundScreen) View(f *tea.Frame) {
	area := f.Area()
	centerX := area.X + max(0, (area.Width-30)/2)
	centerY := area.Y + max(0, (area.Height-4)/2)

	f.Buffer.SetString(centerX, centerY, "404 - Route Not Found", cell.ColorHex("#FF0055"), cell.DefaultColor(), cell.AttrBold)
	f.Buffer.SetString(centerX, centerY+1, "Path: "+d.path, cell.ColorHex("#AAAAAA"), cell.DefaultColor(), cell.AttrNone)
	f.Buffer.SetString(centerX, centerY+3, "Press [Esc] or Pop to return", cell.ColorHex("#666688"), cell.DefaultColor(), cell.AttrDim)
}

func (d *defaultNotFoundScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.Key.Type == 2 { // Esc
		return d, func() tea.Msg { return PopMsg{} }
	}
	return d, nil
}

func (wm *WindowManager) syncOmnibarAddress() {
	active := wm.ActiveScreen()
	if active == nil {
		if wm.currentRoute != "" {
			wm.omnibar.SetCurrentAddress(wm.currentRoute)
		}
		return
	}

	type addressProvider interface {
		CurrentAddress() string
	}
	type dirProvider interface {
		CurrentDir() string
	}
	type routeProvider interface {
		Route() string
	}

	if ap, ok := active.(addressProvider); ok && ap.CurrentAddress() != "" {
		wm.omnibar.SetCurrentAddress(ap.CurrentAddress())
		return
	}
	if dp, ok := active.(dirProvider); ok && dp.CurrentDir() != "" {
		wm.omnibar.SetCurrentAddress(dp.CurrentDir())
		return
	}
	if rp, ok := active.(routeProvider); ok && rp.Route() != "" {
		wm.omnibar.SetCurrentAddress(rp.Route())
		return
	}
	if wm.currentRoute != "" {
		wm.omnibar.SetCurrentAddress(wm.currentRoute)
	}
}
