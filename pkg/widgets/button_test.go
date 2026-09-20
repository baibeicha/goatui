package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
)

func TestButton_ClickAndHotkey(t *testing.T) {
	clicked := false
	btn := NewButton("b1", "Submit", func() {
		clicked = true
	}).SetHotkey('s')

	// Keyboard Enter when focused
	btn.SetFocused(true)
	btn.HandleKey(input.Key{Type: input.KeyEnter})
	if !clicked {
		t.Fatal("expected button clicked on Enter")
	}

	// Hotkey when unfocused
	clicked = false
	btn.SetFocused(false)
	btn.HandleKey(input.Key{Type: input.KeyRune, Rune: 's'})
	if !clicked {
		t.Fatal("expected button clicked on hotkey 's'")
	}

	// Disabled should not trigger
	clicked = false
	btn.SetDisabled(true)
	btn.HandleKey(input.Key{Type: input.KeyRune, Rune: 's'})
	if clicked {
		t.Fatal("disabled button should not trigger on hotkey")
	}
}

func TestButton_MouseInteraction(t *testing.T) {
	clicked := false
	btn := NewButton("b2", "Action", func() {
		clicked = true
	}).SetVariant(ButtonVariantPrimary)

	buf := buffer.NewBuffer(20, 1)
	area := buffer.NewRect(5, 0, 10, 1)
	btn.Draw(buf, area)

	// Hover
	btn.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MouseMotion,
		X:      6,
		Y:      0,
	}})
	if btn.state != ButtonHover {
		t.Fatalf("expected state ButtonHover, got %d", btn.state)
	}

	// Press
	btn.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MousePress,
		Button: input.MouseLeft,
		X:      6,
		Y:      0,
	}})
	if btn.state != ButtonPressed {
		t.Fatalf("expected state ButtonPressed, got %d", btn.state)
	}

	// Release -> Click!
	btn.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MouseRelease,
		Button: input.MouseLeft,
		X:      6,
		Y:      0,
	}})
	if !clicked {
		t.Fatal("expected click callback invoked on mouse release")
	}
}
