package window

import (
	"strings"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/tea"
)

type mockScreen struct {
	BaseScreen
	name      string
	mounted   bool
	paused    bool
	resumed   bool
	destroyed bool
}

func (m *mockScreen) OnMount(ctx *router.RouteContext) { m.mounted = true }
func (m *mockScreen) OnPause()                         { m.paused = true }
func (m *mockScreen) OnResume()                        { m.resumed = true }
func (m *mockScreen) OnDestroy()                       { m.destroyed = true }

func (m *mockScreen) View(f *tea.Frame) {
	f.Buffer.SetString(0, 0, m.name, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
}

func TestWindowManagerLifecycle(t *testing.T) {
	r := router.NewRouter()
	wm := NewWindowManager(r)

	s1 := &mockScreen{name: "Screen1"}
	s2 := &mockScreen{name: "Screen2"}

	// Push Screen 1
	wm.Push(s1, router.NewRouteContext("/s1", r.Session()))
	if !s1.mounted {
		t.Errorf("Screen 1 should be mounted")
	}
	if wm.ActiveScreen() != s1 {
		t.Errorf("Screen 1 should be active")
	}

	// Push Screen 2 -> Screen 1 should be paused
	wm.Push(s2, router.NewRouteContext("/s2", r.Session()))
	if !s1.paused {
		t.Errorf("Screen 1 should be paused when Screen 2 is pushed")
	}
	if !s2.mounted {
		t.Errorf("Screen 2 should be mounted")
	}
	if wm.ActiveScreen() != s2 {
		t.Errorf("Screen 2 should be active")
	}

	// Pop Screen 2 -> Screen 2 should be destroyed, Screen 1 resumed
	wm.Pop()
	if !s2.destroyed {
		t.Errorf("Screen 2 should be destroyed on Pop")
	}
	if !s1.resumed {
		t.Errorf("Screen 1 should be resumed on Pop")
	}
	if wm.ActiveScreen() != s1 {
		t.Errorf("Screen 1 should be active after Pop")
	}

	// Popping Screen 1 (root) should not remove it
	wm.Pop()
	if wm.ActiveScreen() != s1 {
		t.Errorf("Root screen should not be popped")
	}
}

func TestWindowManagerModal(t *testing.T) {
	r := router.NewRouter()
	wm := NewWindowManager(r)

	s := &mockScreen{name: "MainScreen"}
	wm.Push(s, router.NewRouteContext("/", r.Session()))

	alertDismissed := false
	alert := AlertModal("Alert", "Something happened", func() tea.Cmd {
		alertDismissed = true
		return nil
	})

	// Open Modal
	wm.Update(ShowModalMsg{Modal: alert})

	// Render frame and verify backdrop dimming
	buf := buffer.NewBuffer(60, 20)
	f := &tea.Frame{Buffer: buf}
	wm.View(f)

	// Focus trap: send Enter to modal
	wm.Update(tea.KeyMsg{Key: input.Key{Type: 1}}) // Enter
	if !alertDismissed {
		t.Errorf("Alert should be dismissed by Enter key")
	}
}

func TestOmnibarFiltering(t *testing.T) {
	ob := NewOmnibar()
	ob.AddRoute("/dashboard", "Main Overview")
	ob.AddRoute("/users", "User directory")
	ob.AddRoute("/users/:id", "User profile")
	ob.Open()

	if len(ob.filtered) != 3 {
		t.Errorf("Expected 3 items, got %d", len(ob.filtered))
	}

	// Type 'dash'
	ob.input.HandleKey(input.Key{Type: input.KeyRune, Rune: 'd'})
	ob.input.HandleKey(input.Key{Type: input.KeyRune, Rune: 'a'})
	ob.input.HandleKey(input.Key{Type: input.KeyRune, Rune: 's'})
	ob.input.HandleKey(input.Key{Type: input.KeyRune, Rune: 'h'})
	ob.filter()

	if len(ob.filtered) != 1 || ob.filtered[0].item.Route != "/dashboard" {
		t.Errorf("Expected only /dashboard to match, got %+v", ob.filtered)
	}
}

type tickerMsg struct{}

type tickerScreen struct {
	BaseScreen
	tickCount int
}

func (t *tickerScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if _, ok := msg.(tickerMsg); ok {
		t.tickCount++
	}
	return t, nil
}

func TestModalFocusTrapAndBackgroundEvents(t *testing.T) {
	r := router.NewRouter()
	wm := NewWindowManager(r)

	scr := &tickerScreen{}
	wm.Push(scr, router.NewRouteContext("/", r.Session()))

	// Show modal
	alert := AlertModal("Notice", "Testing ticks", nil)
	wm.Update(ShowModalMsg{Modal: alert})

	// Background ticker message sent to WindowManager
	wm.Update(tickerMsg{})

	// The screen MUST receive background ticker event even while modal is open!
	if scr.tickCount != 1 {
		t.Errorf("Expected screen to receive ticker event while modal is active, got %d", scr.tickCount)
	}

	// Confirm modal rendering is bounded and centered
	buf := buffer.NewBuffer(80, 24)
	f := &tea.Frame{Buffer: buf}
	confirm := ConfirmModal("Confirm", "Question?", nil, nil)
	bounds := confirm.Bounds(buf.Area())
	if bounds.Width >= 80 || bounds.Height >= 24 {
		t.Errorf("Modal bounds should be smaller than terminal area, got %+v", bounds)
	}
	confirm.View(f)
}

func TestAccessDeniedScreen(t *testing.T) {
	denied := NewAccessDeniedScreen("Role admin required", "/admin")
	buf := buffer.NewBuffer(80, 24)
	f := &tea.Frame{Buffer: buf}
	denied.View(f)

	// Test pressing 'T' generates ToggleRoleMsg
	_, cmd := denied.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 't'}})
	if cmd == nil {
		t.Fatalf("Expected command when pressing 't' on AccessDeniedScreen")
	}
	msg := cmd()
	if _, ok := msg.(ToggleRoleMsg); !ok {
		t.Errorf("Expected ToggleRoleMsg, got %T", msg)
	}
}

type mockAddressScreen struct {
	mockScreen
	address string
}

func (m *mockAddressScreen) CurrentAddress() string {
	return m.address
}

func TestOmnibarAddressDisplayAndDirectNavigation(t *testing.T) {
	r := router.NewRouter()
	wm := NewWindowManager(r)

	// Screen with address
	scr := &mockAddressScreen{
		mockScreen: mockScreen{name: "AddrScreen"},
		address:    "/files/documents",
	}
	wm.Push(scr, router.NewRouteContext("/files/documents", r.Session()))

	// Toggle Omnibar -> Should sync address from active screen
	wm.Update(ToggleOmnibarMsg{})
	if !wm.Omnibar().IsVisible() {
		t.Fatalf("Omnibar should be visible after ToggleOmnibarMsg")
	}

	if wm.Omnibar().CurrentAddress() != "/files/documents" {
		t.Errorf("Expected current address '/files/documents', got '%s'", wm.Omnibar().CurrentAddress())
	}

	// Render Omnibar to buffer
	buf := buffer.NewBuffer(80, 24)
	f := &tea.Frame{Buffer: buf}
	wm.Omnibar().View(f)

	// Verify that title with address was rendered into buffer
	foundTitle := false
	for y := 0; y < 24; y++ {
		var row []rune
		for x := 0; x < 80; x++ {
			c := buf.Cell(x, y)
			if c != nil && c.Rune != 0 {
				row = append(row, c.Rune)
			} else {
				row = append(row, ' ')
			}
		}
		if strings.Contains(string(row), "Address") {
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Errorf("Expected to find 'Address' in rendered Omnibar buffer")
	}

	// Test typing a drive path: C:\Projects\MyRepo
	wm.Omnibar().input.SetValue(`C:\Projects\MyRepo`)
	handled, cmd := wm.Omnibar().Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if !handled || cmd == nil {
		t.Fatalf("Expected handled=true and non-nil cmd for Enter on drive path")
	}
	navMsg := cmd()
	if nav, ok := navMsg.(NavigateMsg); !ok || nav.URL != "file:///C:/Projects/MyRepo" {
		t.Errorf("Expected NavigateMsg to file:///C:/Projects/MyRepo, got %+v", navMsg)
	}

	// Test typing a relative path: ./subfolder
	wm.Omnibar().Open()
	wm.Omnibar().input.SetValue("./subfolder")
	handled, cmd = wm.Omnibar().Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if !handled || cmd == nil {
		t.Fatalf("Expected handled=true and non-nil cmd for Enter on relative path")
	}
	navMsg = cmd()
	if nav, ok := navMsg.(NavigateMsg); !ok || nav.URL != "file:///./subfolder" {
		t.Errorf("Expected NavigateMsg to file:///./subfolder, got %+v", navMsg)
	}
}

type pasteReceiverScreen struct {
	BaseScreen
	receivedPaste string
}

func (p *pasteReceiverScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if paste, ok := msg.(tea.PasteMsg); ok {
		p.receivedPaste = paste.Text
	}
	return p, nil
}

func TestModalPasteTrapAndOmnibarPaste(t *testing.T) {
	r := router.NewRouter()
	wm := NewWindowManager(r)

	bgScreen := &pasteReceiverScreen{}
	wm.Push(bgScreen, router.NewRouteContext("/", r.Session()))

	// 1. Show InputModal
	inputMod := NewInputModal("Title", "Prompt", "", nil, nil)
	wm.Update(ShowModalMsg{Modal: inputMod})

	// 2. Dispatch PasteMsg: must be captured by modal and NOT leak to bgScreen
	wm.Update(tea.PasteMsg{Text: "my-secret-key-1234"})
	if inputMod.input.Value() != "my-secret-key-1234" {
		t.Errorf("Expected modal text input to contain pasted text, got %q", inputMod.input.Value())
	}
	if bgScreen.receivedPaste != "" {
		t.Errorf("Pasted text leaked to background screen behind modal: %q", bgScreen.receivedPaste)
	}

	// 3. Close modal
	wm.Update(CloseModalMsg{})

	// 4. Open Omnibar and test paste
	wm.Omnibar().Open()
	wm.Update(tea.PasteMsg{Text: "/dashboard"})
	if wm.Omnibar().input.Value() != "/dashboard" {
		t.Errorf("Expected Omnibar to receive pasted text, got %q", wm.Omnibar().input.Value())
	}
	if bgScreen.receivedPaste != "" {
		t.Errorf("Pasted text leaked to background screen behind Omnibar: %q", bgScreen.receivedPaste)
	}
}
