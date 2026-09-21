package goatui

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/ui"
)

func TestApp_TabsAndNavigation(t *testing.T) {
	app := NewApp("TestApp")

	rendered1 := false
	rendered2 := false

	app.AddTab("t1", "Tab One", func(f *tea.Frame, area buffer.Rect) {
		rendered1 = true
	})
	app.AddTab("t2", "Tab Two", func(f *tea.Frame, area buffer.Rect) {
		rendered2 = true
	})

	if app.ActiveTab() != "t1" {
		t.Fatalf("expected initial active tab 't1', got '%s'", app.ActiveTab())
	}

	buf := buffer.NewBuffer(80, 24)
	frame := &tea.Frame{Buffer: buf}
	app.View(frame)

	if !rendered1 {
		t.Fatal("expected tab 1 to be rendered")
	}

	// Switch tab by ID
	app.SetActiveTab("t2")
	if app.ActiveTab() != "t2" {
		t.Fatalf("expected active tab 't2', got '%s'", app.ActiveTab())
	}
	app.View(frame)
	if !rendered2 {
		t.Fatal("expected tab 2 to be rendered after switch")
	}
}

func TestApp_ModalsAndToasts(t *testing.T) {
	app := NewApp("TestApp")

	// Toast
	app.ToastInfo("Info", "Test message")
	if app.toasts.Count() != 1 {
		t.Fatalf("expected 1 toast, got %d", app.toasts.Count())
	}

	// Alert Modal
	dismissed := false
	app.Alert("Alert Title", "Alert Message", func() {
		dismissed = true
	})

	if app.modal == nil {
		t.Fatal("expected modal to be active")
	}

	// Dismiss via Enter
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if !dismissed {
		t.Fatal("expected alert onDismiss to be called")
	}

	// Confirm Modal
	confirmed := false
	cancelled := false
	app.Confirm("Confirm Title", "Confirm Message", func() {
		confirmed = true
	}, func() {
		cancelled = true
	})

	// Press Tab to switch to Cancel, then Enter
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if !cancelled {
		t.Fatal("expected confirm cancel to be called")
	}
	_ = confirmed
}

func TestApp_HotkeysAndMouse(t *testing.T) {
	app := NewApp("TestApp")

	triggered := false
	app.OnKeyRune('d', func() {
		triggered = true
	})

	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'd'}})
	if !triggered {
		t.Fatal("expected hotkey 'd' to trigger")
	}

	// Mouse click on Tab 2
	app.AddTab("a", "Alpha", nil)
	app.AddTab("b", "Beta", nil)

	buf := buffer.NewBuffer(80, 24)
	frame := &tea.Frame{Buffer: buf}
	app.View(frame)

	// Tab bounds were recorded
	if len(app.tabBounds) < 2 {
		t.Fatalf("expected at least 2 tab bounds, got %d", len(app.tabBounds))
	}

	target := app.tabBounds[1]
	app.Update(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MousePress,
		Button: input.MouseLeft,
		X:      target.X + 1,
		Y:      target.Y,
	}})

	if app.ActiveTab() != "b" {
		t.Fatalf("expected tab 'b' to become active after mouse click, got %s", app.ActiveTab())
	}

	// Test Tab-specific Key and Mouse routing
	tabKeyHandled := false
	tabMouseHandled := false
	app.SetTabOnKey("b", func(key input.Key) bool {
		tabKeyHandled = true
		return true
	})
	app.SetTabOnMouse("b", func(msg tea.MouseMsg) bool {
		tabMouseHandled = true
		return true
	})

	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if !tabKeyHandled {
		t.Fatal("expected tab 'b' key handler to be called")
	}

	app.Update(tea.MouseMsg{Mouse: input.Mouse{Action: input.MousePress, Button: input.MouseLeft, X: 50, Y: 10}})
	if !tabMouseHandled {
		t.Fatal("expected tab 'b' mouse handler to be called")
	}

	// Test Default Ctrl+Q quit
	_, quitCmd := app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'q', Mod: input.ModCtrl}})
	if quitCmd == nil {
		t.Fatal("expected quit command on Ctrl+Q")
	}

	// Test small screen rendering (area.Height < 3)
	smallBuf := buffer.NewBuffer(40, 2)
	smallFrame := &tea.Frame{Buffer: smallBuf}
	app.View(smallFrame) // Should render without panic
}

func TestApp_TabCyclingAndNumericKeys(t *testing.T) {
	app := NewApp("NavigationTest")
	app.AddTab("t1", "First", nil)
	app.AddTab("t2", "Second", nil)
	app.AddTab("t3", "Third", nil)

	if app.ActiveTab() != "t1" {
		t.Fatalf("expected initial tab t1, got %s", app.ActiveTab())
	}

	// 1. Plain Tab cycles forward
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	if app.ActiveTab() != "t2" {
		t.Fatalf("expected tab t2 after Tab, got %s", app.ActiveTab())
	}
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	if app.ActiveTab() != "t3" {
		t.Fatalf("expected tab t3 after second Tab, got %s", app.ActiveTab())
	}
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	if app.ActiveTab() != "t1" {
		t.Fatalf("expected tab t1 wrap around after third Tab, got %s", app.ActiveTab())
	}

	// 2. Shift+Tab and Backtab cycle backward
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab, Mod: input.ModShift}})
	if app.ActiveTab() != "t3" {
		t.Fatalf("expected tab t3 after Shift+Tab, got %s", app.ActiveTab())
	}
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyBacktab}})
	if app.ActiveTab() != "t2" {
		t.Fatalf("expected tab t2 after Backtab, got %s", app.ActiveTab())
	}

	// 3. Digits '1'..'9' jump directly to tabs
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '1'}})
	if app.ActiveTab() != "t1" {
		t.Fatalf("expected tab t1 after key '1', got %s", app.ActiveTab())
	}
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '3'}})
	if app.ActiveTab() != "t3" {
		t.Fatalf("expected tab t3 after key '3', got %s", app.ActiveTab())
	}

	// Out of range digit should not change tab
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '9'}})
	if app.ActiveTab() != "t3" {
		t.Fatalf("expected tab t3 unchanged after key '9', got %s", app.ActiveTab())
	}

	// 4. Alt+1..9 navigation
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '2', Mod: input.ModAlt}})
	if app.ActiveTab() != "t2" {
		t.Fatalf("expected tab t2 after Alt+2, got %s", app.ActiveTab())
	}

	// 5. Active tab with SetTabInterceptTab(true) can consume Tab
	consumed := false
	app.SetTabInterceptTab("t2", true)
	app.SetTabOnKey("t2", func(key input.Key) bool {
		if key.Type == input.KeyTab {
			consumed = true
			return true
		}
		return false
	})
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	if !consumed {
		t.Fatal("expected active tab key handler to consume Tab")
	}
	if app.ActiveTab() != "t2" {
		t.Fatalf("expected active tab to remain t2 when Tab consumed, got %s", app.ActiveTab())
	}

	// 6. Active tab without SetTabInterceptTab cannot intercept Tab
	app.SetTabInterceptTab("t2", false)
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyTab}})
	if app.ActiveTab() != "t3" {
		t.Fatalf("expected Tab to switch to t3 when interceptTab is false, got %s", app.ActiveTab())
	}
}

func TestApp_HelpDialog(t *testing.T) {
	app := NewApp("HelpTest")
	app.AddTab("main", "Main", nil)
	app.SetKeyHints(
		ui.KeyHint{Key: "Tab", Desc: "Switch Tab"},
		ui.KeyHint{Key: "F", Desc: "Flatten"},
	)

	// Press '?' to toggle help
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '?'}})
	if app.modal == nil || !app.isHelpOpen {
		t.Fatal("expected help modal to be open after pressing '?'")
	}

	// Press '?' again to toggle off
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '?'}})
	if app.modal != nil || app.isHelpOpen {
		t.Fatal("expected help modal to close after pressing '?' again")
	}

	// F1 opens help
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyF1}})
	if app.modal == nil || !app.isHelpOpen {
		t.Fatal("expected help modal open after F1")
	}

	// 'q' closes help
	app.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'q'}})
	if app.modal != nil || app.isHelpOpen {
		t.Fatal("expected help modal closed after 'q'")
	}
}

