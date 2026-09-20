package goatui

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
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
