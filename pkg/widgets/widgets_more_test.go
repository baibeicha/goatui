package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
)

func TestDivider(t *testing.T) {
	d := NewHorizontalDivider("SECTION").
		SetTitleAlign(buffer.AlignCenter)

	buf := buffer.NewBuffer(30, 3)
	d.Draw(buf, buffer.NewRect(0, 1, 30, 1))

	// Ensure runes were written
	c := buf.Cell(0, 1)
	if c == nil || c.Rune != '─' {
		t.Fatalf("expected horizontal line rune '─', got %c", c.Rune)
	}

	// Vertical divider
	dv := NewVerticalDivider()
	dv.Draw(buf, buffer.NewRect(5, 0, 1, 3))
	cv := buf.Cell(5, 0)
	if cv == nil || cv.Rune != '│' {
		t.Fatalf("expected vertical line rune '│', got %c", cv.Rune)
	}
}

func TestSelect(t *testing.T) {
	items := []SelectItem{
		{ID: "m1", Label: "Model Quasar"},
		{ID: "m2", Label: "Model Solaris"},
		{ID: "m3", Label: "Model Pulsar"},
	}
	sel := NewSelect("sel", "Choose", items...)
	sel.SetFocused(true)

	// Press Enter to open
	sel.HandleKey(input.Key{Type: input.KeyEnter})
	if !sel.IsOpen() {
		t.Fatal("expected select dropdown open after Enter")
	}

	// Down -> m2
	sel.HandleKey(input.Key{Type: input.KeyDown})
	if sel.focusedIdx != 1 {
		t.Fatalf("expected focusedIdx 1, got %d", sel.focusedIdx)
	}

	// Enter -> select m2 and close
	sel.HandleKey(input.Key{Type: input.KeyEnter})
	if sel.IsOpen() {
		t.Fatal("expected select dropdown closed after selection")
	}
	if sel.SelectedID() != "m2" {
		t.Fatalf("expected selected m2, got %s", sel.SelectedID())
	}

	// Test Label() and SetLabel()
	sel.SetLabel("Environment")
	if sel.Label() != "Environment" {
		t.Fatalf("expected label 'Environment', got %s", sel.Label())
	}

	// Test SetItems scrollOff clamping
	sel.scrollOff = 10
	sel.SetItems([]SelectItem{
		{ID: "x1", Label: "Item 1"},
		{ID: "x2", Label: "Item 2"},
	})
	if sel.scrollOff > 0 {
		t.Fatalf("expected scrollOff clamped to 0, got %d", sel.scrollOff)
	}

	// Test popup boundary clipping near right edge
	buf := buffer.NewBuffer(30, 10)
	sel.SetOpen(true)
	sel.Draw(buf, buffer.NewRect(25, 0, 10, 1))
	if sel.popupBounds.Right() > buf.Width() {
		t.Fatalf("expected popup to be clipped to buffer width %d, got right %d", buf.Width(), sel.popupBounds.Right())
	}
}

func TestSlider(t *testing.T) {
	changed := 0.0
	sl := NewSlider("s1", "Volume", 0.0, 100.0, 50.0).
		SetOnChange(func(v float64) {
			changed = v
		})

	sl.SetFocused(true)
	sl.HandleKey(input.Key{Type: input.KeyRight})
	if sl.Value() <= 50.0 {
		t.Fatalf("expected value > 50.0 after Right arrow, got %f", sl.Value())
	}
	if changed != sl.Value() {
		t.Fatal("expected OnChange callback invoked")
	}

	// Test Home / End
	sl.HandleKey(input.Key{Type: input.KeyEnd})
	if sl.Value() != 100.0 {
		t.Fatalf("expected max value 100.0 on End, got %f", sl.Value())
	}
	sl.HandleKey(input.Key{Type: input.KeyHome})
	if sl.Value() != 0.0 {
		t.Fatalf("expected min value 0.0 on Home, got %f", sl.Value())
	}

	// Narrow area safety
	buf := buffer.NewBuffer(10, 2)
	sl.Draw(buf, buffer.NewRect(0, 0, 6, 1))
	// Cell at 6 must be unmodified
	c := buf.Cell(6, 0)
	if c != nil && c.Rune != ' ' && c.Rune != 0 {
		t.Fatalf("expected cell past slider area to remain unmodified, got %c", c.Rune)
	}
}

func TestHistogram(t *testing.T) {
	bars := []HistogramBar{
		{Label: "A", Value: 10, Color: cell.ColorHex("#FF0055")},
		{Label: "B", Value: 25, Color: cell.ColorHex("#00FFAA")},
		{Label: "C", Value: 18, Color: cell.ColorHex("#00D2FF")},
	}
	h := NewHistogram(bars...).SetBarWidth(2)

	buf := buffer.NewBuffer(20, 6)
	h.Draw(buf, buffer.NewRect(0, 0, 20, 6))

	// Horizontal orientation
	h.SetOrientation(layout.Horizontal)
	h.Draw(buf, buffer.NewRect(0, 0, 20, 6))

	// Test horizontal bar with bar.Value == maxVal and narrow area
	// Must not write past area.Right()
	h2 := NewHistogram(HistogramBar{Label: "CPU", Value: 100}).SetMaxVal(100)
	h2.SetOrientation(layout.Horizontal)
	buf2 := buffer.NewBuffer(30, 2)
	h2.Draw(buf2, buffer.NewRect(0, 0, 15, 1))
	c := buf2.Cell(15, 0)
	if c != nil && c.Rune != ' ' && c.Rune != 0 {
		t.Fatalf("expected cell at 15 (outside area) to be untouched, got %c", c.Rune)
	}
}

func TestSpinner(t *testing.T) {
	sp := NewSpinner(SpinnerDots, "Processing...")
	initial := sp.CurrentGlyph()
	sp.Tick()
	next := sp.CurrentGlyph()
	if initial == next {
		t.Fatal("expected spinner glyph to change on Tick()")
	}

	// Negative frame modulo safety
	sp.currentFrame = -1
	glyph := sp.CurrentGlyph()
	if glyph == "" {
		t.Fatal("expected non-empty glyph with negative currentFrame")
	}

	buf := buffer.NewBuffer(20, 1)
	sp.Draw(buf, buffer.NewRect(0, 0, 20, 1))
}
