package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

func TestCheckboxGroup_NavigationAndToggle(t *testing.T) {
	items := []CheckboxItem{
		{ID: "all", Label: "ALL", Checked: true},
		{ID: "info", Label: "INFO", Checked: true},
		{ID: "warn", Label: "WARN", Checked: false},
	}

	cg := NewCheckboxGroup(items...)

	if cg.Focused() != 0 {
		t.Fatalf("expected initial focused 0, got %d", cg.Focused())
	}

	// Move Next (Right)
	cg.HandleKey(input.Key{Type: input.KeyRight})
	if cg.Focused() != 1 {
		t.Fatalf("expected focused 1 after Right, got %d", cg.Focused())
	}

	// Toggle Focused ("info" should become false)
	cg.HandleKey(input.Key{Type: input.KeyEnter})
	if cg.IsChecked("info") {
		t.Fatalf("expected info to be unchecked after Enter")
	}

	// Move Next (Tab) -> "warn"
	cg.HandleKey(input.Key{Type: input.KeyTab})
	if cg.Focused() != 2 {
		t.Fatalf("expected focused 2 after Tab, got %d", cg.Focused())
	}

	// Toggle Focused ("warn" should become true via Space)
	cg.HandleKey(input.Key{Type: input.KeySpace})
	if !cg.IsChecked("warn") {
		t.Fatalf("expected warn to be checked after Space")
	}

	// Test ToggleAll
	cg.ToggleAll(false)
	for _, it := range cg.Items() {
		if it.Checked {
			t.Fatalf("expected item %s to be false after ToggleAll(false)", it.ID)
		}
	}
}

func TestCheckboxGroup_Draw(t *testing.T) {
	items := []CheckboxItem{
		{ID: "1", Label: "Item1", Checked: true},
		{ID: "2", Label: "Item2", Checked: false},
	}
	cg := NewCheckboxGroup(items...)

	buf := buffer.NewBuffer(60, 5)
	cg.Draw(buf, buffer.NewRect(0, 0, 60, 1))

	// Assert buffer contains marker and text
	cell0 := buf.Cell(0, 0)
	if cell0 == nil {
		t.Fatal("expected non-nil cell at (0, 0)")
	}
}

func TestStandaloneCheckbox(t *testing.T) {
	cb := NewCheckbox("agree", "I Agree", false)
	if cb.IsChecked() {
		t.Fatal("expected unchecked")
	}

	cb.SetFocused(true)
	cb.HandleKey(input.Key{Type: input.KeyEnter})
	if !cb.IsChecked() {
		t.Fatal("expected checked after Enter")
	}
}
