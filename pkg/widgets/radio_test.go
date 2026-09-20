package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/tea"
)

func TestRadioGroup_NavigationAndSelection(t *testing.T) {
	items := []RadioItem{
		{ID: "opt1", Label: "Option 1"},
		{ID: "opt2", Label: "Option 2"},
		{ID: "opt3", Label: "Option 3", Disabled: true},
		{ID: "opt4", Label: "Option 4", Hotkey: '4'},
	}

	rg := NewRadioGroup(items...)
	if rg.SelectedID() != "opt1" {
		t.Fatalf("expected initial selected opt1, got %s", rg.SelectedID())
	}

	// Move Down -> opt2
	rg.HandleKey(input.Key{Type: input.KeyDown})
	if rg.Focused() != 1 {
		t.Fatalf("expected focused 1, got %d", rg.Focused())
	}
	// Select focused
	rg.HandleKey(input.Key{Type: input.KeyEnter})
	if rg.SelectedID() != "opt2" {
		t.Fatalf("expected selected opt2, got %s", rg.SelectedID())
	}

	// Move Down -> should skip disabled opt3 and go to opt4
	rg.HandleKey(input.Key{Type: input.KeyDown})
	if rg.Focused() != 3 {
		t.Fatalf("expected focused 3 (skipping disabled opt3), got %d", rg.Focused())
	}

	// Hotkey selection
	rg.HandleKey(input.Key{Type: input.KeyRune, Rune: '4'})
	if rg.SelectedID() != "opt4" {
		t.Fatalf("expected selected opt4 via hotkey, got %s", rg.SelectedID())
	}
}

func TestRadioGroup_StylesAndMouse(t *testing.T) {
	items := []RadioItem{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
	}
	rg := NewRadioGroup(items...).
		SetStyle(RadioDot).
		SetOrientation(layout.Horizontal)

	buf := buffer.NewBuffer(40, 2)
	rg.Draw(buf, buffer.NewRect(0, 0, 40, 1))

	// Click on Beta
	clicked := rg.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MousePress,
		Button: input.MouseLeft,
		X:      15,
		Y:      0,
	}})
	if !clicked {
		t.Fatal("expected click to be handled")
	}
	if rg.SelectedID() != "b" {
		t.Fatalf("expected selected 'b', got %s", rg.SelectedID())
	}
}

func TestStandaloneRadioButton(t *testing.T) {
	rb := NewRadioButton("r1", "Solo Radio", false).
		SetStyle(RadioBrackets)

	if rb.IsSelected() {
		t.Fatal("expected not selected")
	}

	rb.SetFocused(true)
	rb.HandleKey(input.Key{Type: input.KeySpace})
	if !rb.IsSelected() {
		t.Fatal("expected selected after Space")
	}

	buf := buffer.NewBuffer(20, 1)
	rb.Draw(buf, buffer.NewRect(0, 0, 20, 1))

	// Mouse click
	rb.SetSelected(false)
	clicked := rb.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MousePress,
		Button: input.MouseLeft,
		X:      2,
		Y:      0,
	}})
	if !clicked || !rb.IsSelected() {
		t.Fatal("expected radio button selected after mouse click")
	}
}
