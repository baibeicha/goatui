package style

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestStyleFluentBuilder(t *testing.T) {
	st := NewStyle().
		Bold(true).
		Foreground(cell.ColorHex("#FF007F")).
		Background(cell.ANSI256(234)).
		Border(BorderRounded).
		Padding(1, 2, 1, 2)

	if !st.modifier.Has(cell.AttrBold) {
		t.Errorf("Expected Bold attribute")
	}
	if st.fg != cell.ColorHex("#FF007F") {
		t.Errorf("Foreground color mismatch")
	}
	if !st.hasBorder {
		t.Errorf("Expected hasBorder = true")
	}
	if st.padTop != 1 || st.padRight != 2 {
		t.Errorf("Padding mismatch: %d, %d", st.padTop, st.padRight)
	}
}

func TestStyleDrawWithBorderAndText(t *testing.T) {
	buf := buffer.NewBuffer(20, 5)
	st := NewStyle().Border(BorderNormal)
	st.Draw(buf, buffer.NewRect(0, 0, 10, 3), "Hi")

	// Top-left corner should be '┌'
	cTL := buf.Cell(0, 0)
	if cTL == nil || cTL.Rune != '┌' {
		t.Errorf("Expected top-left '┌', got %+v", cTL)
	}

	// Inner text at (1, 1) should be 'H'
	cText := buf.Cell(1, 1)
	if cText == nil || cText.Rune != 'H' {
		t.Errorf("Expected cell at (1, 1) to be 'H', got %+v", cText)
	}
}

func TestStyleDrawWithBorderTitle(t *testing.T) {
	buf := buffer.NewBuffer(20, 5)
	st := NewStyle().Border(BorderNormal).Title("Logs")
	st.Draw(buf, buffer.NewRect(0, 0, 12, 4), "")

	// Top-left corner should be '┌'
	cTL := buf.Cell(0, 0)
	if cTL == nil || cTL.Rune != '┌' {
		t.Errorf("Expected top-left '┌', got %+v", cTL)
	}

	// At (2, 0) should start " Logs "
	cSpace := buf.Cell(2, 0)
	cL := buf.Cell(3, 0)
	if cSpace == nil || cSpace.Rune != ' ' {
		t.Errorf("Expected space at (2, 0), got %+v", cSpace)
	}
	if cL == nil || cL.Rune != 'L' {
		t.Errorf("Expected 'L' at (3, 0), got %+v", cL)
	}
}

func TestBorderMerging(t *testing.T) {
	buf := buffer.NewBuffer(20, 5)
	st := NewStyle().Border(BorderNormal)

	// Draw left box from x=0 to x=5
	st.Draw(buf, buffer.NewRect(0, 0, 6, 5), "")
	// Draw right box from x=5 to x=10 (touching at x=5)
	st.Draw(buf, buffer.NewRect(5, 0, 6, 5), "")

	// At (5, 0), top-right '┐' of box1 meets top-left '┌' of box2 -> should merge to '┬'
	cTop := buf.Cell(5, 0)
	if cTop == nil || cTop.Rune != '┬' {
		t.Errorf("Expected junction '┬' at (5, 0), got %q (%+v)", cTop.Rune, cTop)
	}

	// At (5, 4), bottom-right '┘' of box1 meets bottom-left '└' of box2 -> should merge to '┴'
	cBottom := buf.Cell(5, 4)
	if cBottom == nil || cBottom.Rune != '┴' {
		t.Errorf("Expected junction '┴' at (5, 4), got %q (%+v)", cBottom.Rune, cBottom)
	}
}

func TestStyleDrawTinyRect(t *testing.T) {
	buf := buffer.NewBuffer(10, 10)
	st := NewStyle().Border(BorderNormal)

	// Rect with width 1 or height 1 should not crash or draw invalid merged corners
	st.Draw(buf, buffer.NewRect(0, 0, 1, 1), "")
	st.Draw(buf, buffer.NewRect(0, 0, 5, 1), "")
	st.Draw(buf, buffer.NewRect(0, 0, 1, 5), "")
}

func BenchmarkStyleDraw(b *testing.B) {
	buf := buffer.NewBuffer(80, 24)
	st := NewStyle().
		Bold(true).
		Foreground(cell.ColorHex("#00FF88")).
		Background(cell.ANSI256(235)).
		Border(BorderRounded).
		Padding(1, 2, 1, 2)
	rect := buffer.NewRect(5, 5, 40, 10)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Draw(buf, rect, "Performance Engine")
	}
}
