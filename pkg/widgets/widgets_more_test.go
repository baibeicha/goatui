package widgets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/validation"
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

	// Test opening with selectedIdx >= maxVisible adjusts scrollOff
	manyItems := make([]SelectItem, 12)
	for i := 0; i < 12; i++ {
		manyItems[i] = SelectItem{ID: fmt.Sprintf("i%d", i), Label: fmt.Sprintf("Item %d", i)}
	}
	sel2 := NewSelect("s2", "ScrollTest", manyItems...)
	sel2.SetSelectedID("i8")
	sel2.SetFocused(true)
	sel2.HandleKey(input.Key{Type: input.KeyEnter})
	if sel2.scrollOff == 0 {
		t.Fatalf("expected scrollOff > 0 when opening select with item 8 selected, got %d", sel2.scrollOff)
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

func TestSelect_OverlayRendering(t *testing.T) {
	items := []SelectItem{
		{ID: "prod", Label: "Production Cluster"},
		{ID: "stage", Label: "Staging Sandbox"},
		{ID: "local", Label: "Local Simulation"},
	}
	sel := NewSelect("env", "Active Environment", items...)
	sel.SetOpen(true)

	buf := buffer.NewBuffer(80, 20)

	// Step 1: Draw select at Y=3
	sel.Draw(buf, buffer.NewRect(0, 3, 40, 1))

	// The popup covers Y=4 to Y=8.
	// Step 2: Now simulate a sibling button drawn at Y=6 AFTER the select was drawn
	btn := NewButton("btn-save", "Sync Config", nil)
	btn.Draw(buf, buffer.NewRect(0, 6, 20, 1))

	// Before RenderOverlays, the button overwrote Y=6
	if buf.Cell(2, 6).Rune == ' ' {
		// button text starts around x=2
	}

	// Step 3: Run the overlay render pass
	if !buf.HasOverlays() {
		t.Fatal("expected buffer to have queued overlay from Select")
	}
	buf.RenderOverlays()

	// Step 4: Verify the second item ("Staging Sandbox") is restored at Y=6 and NOT button text!
	// Row inner for items is Y=5 (item 0), Y=6 (item 1 "Staging Sandbox"), Y=7 (item 2)
	rowStr := ""
	for x := 0; x < 30; x++ {
		c := buf.Cell(x, 6)
		if c != nil && c.Rune != 0 {
			rowStr += string(c.Rune)
		}
	}
	if !strings.Contains(rowStr, "Staging Sandbox") {
		t.Fatalf("expected overlay to restore 'Staging Sandbox' on row 6, got: %q", rowStr)
	}
}

func TestSelect_HoverMouse(t *testing.T) {
	items := []SelectItem{
		{ID: "1", Label: "Item One"},
		{ID: "2", Label: "Item Two"},
		{ID: "3", Label: "Item Three"},
	}
	sel := NewSelect("test", "Select", items...)
	sel.SetOpen(true)

	buf := buffer.NewBuffer(50, 15)
	sel.Draw(buf, buffer.NewRect(0, 1, 30, 1))

	// Item 1 is at index 1
	if len(sel.itemBounds) < 2 {
		t.Fatalf("expected at least 2 item bounds, got %d", len(sel.itemBounds))
	}
	target := sel.itemBounds[1]

	// Hover over item 1
	handled := sel.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{
		Action: input.MouseMotion,
		X:      target.X + 2,
		Y:      target.Y,
	}})
	if !handled {
		t.Fatal("expected HandleMouse to handle hover on open popup")
	}
	if sel.focusedIdx != 1 {
		t.Fatalf("expected focusedIdx 1 on hover, got %d", sel.focusedIdx)
	}
}

func TestTextInput_Validation(t *testing.T) {
	ti := NewTextInput()
	ti.AddValidation(validation.Required(), validation.MinLength(5))
	ti.SetValidateOnChange(true)

	// Initially empty -> invalid
	if ti.IsValid() {
		t.Fatal("expected empty TextInput to be invalid initially")
	}
	if ti.ErrorMessage() == "" {
		t.Fatal("expected non-empty error message")
	}

	// Type "abc" (3 chars, < 5)
	ti.SetValue("abc")
	if ti.IsValid() {
		t.Fatal("expected 'abc' to be invalid (min 5)")
	}

	// Type "abcdef" (6 chars) -> valid
	ti.SetValue("abcdef")
	if !ti.IsValid() {
		t.Fatalf("expected 'abcdef' to be valid, got error: %s", ti.ErrorMessage())
	}

	// Backspace to 4 chars via HandleKey
	ti.HandleKey(input.Key{Type: input.KeyBackspace})
	ti.HandleKey(input.Key{Type: input.KeyBackspace})
	if ti.Value() != "abcd" {
		t.Fatalf("expected value 'abcd', got %s", ti.Value())
	}
	if ti.IsValid() {
		t.Fatal("expected 'abcd' to be invalid after backspace with validateOnChange")
	}

	// Validate on blur
	ti.SetValidateOnChange(false)
	ti.SetValidateOnBlur(true)
	ti.SetValue("ok") // invalid, but validateOnChange is false
	ti.Blur()         // should trigger validation
	if ti.IsValid() {
		t.Fatal("expected invalid after Blur() with validateOnBlur")
	}

	// Custom error override
	ti.SetValue("valid-long-text")
	if !ti.IsValid() {
		t.Fatal("expected valid value")
	}
	ti.SetCustomError("Server rejected username")
	if ti.IsValid() {
		t.Fatal("expected invalid after SetCustomError")
	}
	if ti.ErrorMessage() != "Server rejected username" {
		t.Fatalf("expected custom error message, got %s", ti.ErrorMessage())
	}
	ti.ClearError()
	if !ti.IsValid() {
		t.Fatal("expected valid after ClearError")
	}

	// Visual error rendering test with multiline area
	ti.SetValue("bad")
	buf := buffer.NewBuffer(40, 3)
	ti.Draw(buf, buffer.NewRect(0, 0, 40, 2))

	// Line 1 should contain error indicator "✖"
	line1Str := ""
	for x := 0; x < 40; x++ {
		c := buf.Cell(x, 1)
		if c != nil && c.Rune != 0 {
			line1Str += string(c.Rune)
		}
	}
	if !strings.Contains(line1Str, "✖") {
		t.Fatalf("expected error line to contain '✖', got: %q", line1Str)
	}
}


