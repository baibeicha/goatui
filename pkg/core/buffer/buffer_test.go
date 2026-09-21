package buffer

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestRectOperations(t *testing.T) {
	r1 := NewRect(5, 5, 10, 10)
	r2 := NewRect(10, 10, 10, 10)

	inter := r1.Intersection(r2)
	expectedInter := NewRect(10, 10, 5, 5)
	if inter != expectedInter {
		t.Errorf("Intersection = %+v; want %+v", inter, expectedInter)
	}

	union := r1.Union(r2)
	expectedUnion := NewRect(5, 5, 15, 15)
	if union != expectedUnion {
		t.Errorf("Union = %+v; want %+v", union, expectedUnion)
	}

	inset := r1.Inset(1, 2)
	expectedInset := NewRect(6, 7, 8, 6)
	if inset != expectedInset {
		t.Errorf("Inset = %+v; want %+v", inset, expectedInset)
	}

	if !r1.Contains(5, 5) || !r1.Contains(14, 14) || r1.Contains(15, 15) {
		t.Errorf("Contains failed on boundary points")
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r        rune
		expected int
	}{
		{'a', 1},
		{'Z', 1},
		{'1', 1},
		{' ', 1},
		{'你', 2},    // CJK
		{'好', 2},    // CJK
		{'🚀', 2},    // Rocket emoji
		{0x0300, 0}, // Combining grave accent
		{0x2699, 1}, // Gear symbol: NOT wide in narrow terminals
		{0x270F, 1}, // Pencil symbol: NOT wide
		{0x26A1, 1}, // High voltage / lightning symbol: NOT wide
		{0xFE0E, 0}, // Variation selector 15 (text presentation)
		{0xFE0F, 0}, // Variation selector 16 (emoji presentation)
	}

	for _, tt := range tests {
		w := RuneWidth(tt.r)
		if w != tt.expected {
			t.Errorf("RuneWidth(%q) = %d; want %d", tt.r, w, tt.expected)
		}
	}
}

func TestBufferClipping(t *testing.T) {
	buf := NewBuffer(20, 10)
	clip := NewRect(5, 2, 8, 4) // [5..13) x [2..6)
	buf.SetClip(clip)

	if buf.ClipRect() == nil || *buf.ClipRect() != clip {
		t.Errorf("ClipRect mismatch: %+v", buf.ClipRect())
	}

	// Try writing outside clip at (0, 0)
	buf.Set(0, 0, cell.NewCell('Z', 1))
	if buf.Cell(0, 0).Rune == 'Z' {
		t.Errorf("Set cell wrote outside clip rect at (0, 0)")
	}

	// Try writing inside clip at (6, 3)
	buf.Set(6, 3, cell.NewCell('A', 1))
	if buf.Cell(6, 3).Rune != 'A' {
		t.Errorf("Expected 'A' inside clip rect at (6, 3), got %q", buf.Cell(6, 3).Rune)
	}

	// SetString starting outside clip (0, 3) and going into clip
	buf.SetString(0, 3, "123456789012345", cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	// Cells before x=5 must NOT be modified
	for x := 0; x < 5; x++ {
		if buf.Cell(x, 3).Rune != ' ' {
			t.Errorf("Expected cell at (%d, 3) to remain blank, got %q", x, buf.Cell(x, 3).Rune)
		}
	}
	// Cells inside clip [5..13) should be modified
	if buf.Cell(5, 3).Rune != '6' {
		t.Errorf("Expected cell at (5, 3) to be '6', got %q", buf.Cell(5, 3).Rune)
	}
	// Cells at or after x=13 must NOT be modified
	for x := 13; x < 20; x++ {
		if buf.Cell(x, 3).Rune != ' ' {
			t.Errorf("Expected cell at (%d, 3) to remain blank, got %q", x, buf.Cell(x, 3).Rune)
		}
	}

	// Reset clip and write outside
	buf.ResetClip()
	if buf.ClipRect() != nil {
		t.Errorf("ResetClip should set clipRect to nil")
	}
	buf.Set(0, 0, cell.NewCell('Z', 1))
	if buf.Cell(0, 0).Rune != 'Z' {
		t.Errorf("Expected 'Z' after ResetClip at (0, 0)")
	}
}

func TestBufferSetStringAndWideChar(t *testing.T) {
	buf := NewBuffer(10, 2)
	nextX := buf.SetString(0, 0, "Hello", cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if nextX != 5 {
		t.Errorf("Expected nextX = 5, got %d", nextX)
	}

	for i, r := range "Hello" {
		c := buf.Cell(i, 0)
		if c == nil || c.Rune != r || c.Width != 1 {
			t.Errorf("Cell at %d has rune %q, width %d", i, c.Rune, c.Width)
		}
	}

	// Write wide character '你' at position 5
	nextX = buf.SetString(5, 0, "你", cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if nextX != 7 {
		t.Errorf("Expected nextX = 7 after wide char, got %d", nextX)
	}
	c1 := buf.Cell(5, 0)
	c2 := buf.Cell(6, 0)
	if c1.Rune != '你' || c1.Width != 2 {
		t.Errorf("Wide char primary cell incorrect: %+v", c1)
	}
	if c2.Rune != 0 || c2.Width != 0 {
		t.Errorf("Wide char continuation cell incorrect: %+v", c2)
	}
}

func TestBufferResizeReuse(t *testing.T) {
	buf := NewBuffer(80, 24)
	origCap := cap(buf.Cells())

	// Resize smaller
	buf.Resize(40, 12)
	if cap(buf.Cells()) != origCap {
		t.Errorf("Expected slice capacity preserved on downsize")
	}

	// Resize back
	buf.Resize(80, 24)
	if cap(buf.Cells()) != origCap {
		t.Errorf("Expected slice capacity preserved on re-expansion")
	}
}

func TestBufferSetStringAligned(t *testing.T) {
	buf := NewBuffer(20, 5)

	// Left align
	buf.SetStringAligned(NewRect(0, 0, 10, 1), "AB", AlignLeft, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if buf.Cell(0, 0).Rune != 'A' || buf.Cell(1, 0).Rune != 'B' {
		t.Errorf("Left align failed")
	}

	// Right align inside rect [0..10) -> text "AB" width 2 -> starts at 8
	buf.SetStringAligned(NewRect(0, 1, 10, 1), "AB", AlignRight, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if buf.Cell(8, 1).Rune != 'A' || buf.Cell(9, 1).Rune != 'B' {
		t.Errorf("Right align failed")
	}

	// Center align inside rect [0..10) -> (10-2)/2 = 4 -> starts at 4
	buf.SetStringAligned(NewRect(0, 2, 10, 1), "AB", AlignCenter, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if buf.Cell(4, 2).Rune != 'A' || buf.Cell(5, 2).Rune != 'B' {
		t.Errorf("Center align failed")
	}

	// Truncation: 5 width area with "ABCDEF" -> "ABCD…"
	buf.SetStringAligned(NewRect(0, 3, 5, 1), "ABCDEF", AlignLeft, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	if buf.Cell(4, 3).Rune != '…' {
		t.Errorf("Expected truncation '…', got %q", buf.Cell(4, 3).Rune)
	}
}

func TestBufferDrawAlignedText(t *testing.T) {
	buf := NewBuffer(20, 5)
	// Vertically centered, horizontally centered
	buf.DrawAlignedText(NewRect(0, 0, 10, 3), "X", AlignCenter, AlignMiddle, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	// Middle row is y=1, center x is (10-1)/2 = 4
	if buf.Cell(4, 1).Rune != 'X' {
		t.Errorf("Expected 'X' at (4, 1), got %+v", buf.Cell(4, 1))
	}
}

func TestBufferOverlays(t *testing.T) {
	buf := NewBuffer(20, 5)

	if buf.HasOverlays() {
		t.Fatal("expected no overlays initially")
	}

	// Normal render writes 'A' at (2, 2)
	buf.SetRune(2, 2, 'A', cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)

	// An overlay is registered to write 'B' at (2, 2)
	buf.AddOverlay(func(b *Buffer) {
		b.SetRune(2, 2, 'B', cell.DefaultColor(), cell.DefaultColor(), cell.AttrBold)
	})

	if !buf.HasOverlays() {
		t.Fatal("expected pending overlays")
	}

	// Before rendering overlays, cell is still 'A'
	if buf.Cell(2, 2).Rune != 'A' {
		t.Fatalf("expected 'A' before overlay render, got %c", buf.Cell(2, 2).Rune)
	}

	// Clip buffer to (0, 0, 1, 1) - overlay should bypass clip and restore it
	buf.SetClip(NewRect(0, 0, 1, 1))

	buf.RenderOverlays()

	if buf.HasOverlays() {
		t.Fatal("expected no pending overlays after render")
	}

	// After rendering overlays, cell is 'B'
	if buf.Cell(2, 2).Rune != 'B' {
		t.Fatalf("expected 'B' after overlay render, got %c", buf.Cell(2, 2).Rune)
	}

	// Clip rect should be restored
	if buf.ClipRect() == nil || *buf.ClipRect() != NewRect(0, 0, 1, 1) {
		t.Fatal("expected clip rect restored after overlay render")
	}

	// Reset clears overlays
	buf.AddOverlay(func(b *Buffer) {})
	buf.Reset()
	if buf.HasOverlays() {
		t.Fatal("expected Reset to clear overlays")
	}
}

func BenchmarkBufferSetString(b *testing.B) {
	buf := NewBuffer(120, 40)
	text := "The quick brown fox jumps over the lazy dog."
	fg := cell.ColorHex("#FF007F")
	bg := cell.Color256(234)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.SetString(0, 0, text, fg, bg, cell.AttrBold)
	}
}

func BenchmarkBufferFill(b *testing.B) {
	buf := NewBuffer(120, 40)
	fillCell := cell.NewCell('X', 1)
	rect := NewRect(10, 5, 80, 25)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Fill(rect, fillCell)
	}
}
