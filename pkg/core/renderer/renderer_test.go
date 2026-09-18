package renderer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestRendererInitialRender(t *testing.T) {
	r := NewRenderer(10, 2)
	r.SetSynchronized(true)

	// Draw text in back buffer
	r.Back().SetString(0, 0, "GoatUI", cell.ColorHex("#FF0000"), cell.DefaultColor(), cell.AttrBold)

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := buf.String()
	// Must contain Mode 2026 markers
	if !strings.Contains(out, "\x1b[?2026h") || !strings.Contains(out, "\x1b[?2026l") {
		t.Errorf("Expected Mode 2026 markers in output: %q", out)
	}
	// Must contain "GoatUI"
	if !strings.Contains(out, "GoatUI") {
		t.Errorf("Expected 'GoatUI' in output: %q", out)
	}
}

func TestRendererCleanFrameShortCircuit(t *testing.T) {
	r := NewRenderer(80, 24)
	r.Back().SetString(0, 0, "No change", cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)

	var firstBuf bytes.Buffer
	_ = r.Render(&firstBuf)

	// Second render with no changes
	var secondBuf bytes.Buffer
	if err := r.Render(&secondBuf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if secondBuf.Len() != 0 {
		t.Errorf("Expected 0 bytes emitted on clean frame, got %d bytes: %q", secondBuf.Len(), secondBuf.String())
	}
}

func TestRendererSparseUpdate(t *testing.T) {
	r := NewRenderer(80, 24)
	var discard bytes.Buffer
	_ = r.Render(&discard)

	// Update exactly one cell at (10, 5)
	r.Back().SetRune(10, 5, 'Z', cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)

	var updateBuf bytes.Buffer
	if err := r.Render(&updateBuf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	out := updateBuf.String()
	if !strings.Contains(out, "Z") {
		t.Errorf("Expected 'Z' in output: %q", out)
	}
	// Cursor position for (10, 5) is 1-indexed (row 6, col 11) -> \x1b[6;11H
	if !strings.Contains(out, "\x1b[6;11H") {
		t.Errorf("Expected cursor move to \x1b[6;11H, got: %q", out)
	}
}

func TestAppendCursorMoveUnknownX(t *testing.T) {
	// When curX is -1 (unknown), appendCursorMove must emit absolute CUP (\x1b[1;2H)
	// and NEVER relative move (\x1b[2C)
	out := appendCursorMove(nil, -1, 0, 1, 0)
	expected := "\x1b[1;2H"
	if string(out) != expected {
		t.Errorf("Expected %q when curX=-1, got %q", expected, string(out))
	}
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkRendererCleanFrame(b *testing.B) {
	r := NewRenderer(120, 40)
	var nw noopWriter
	_ = r.Render(nw)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Render(nw)
	}
}

func BenchmarkRendererSparseFrame(b *testing.B) {
	r := NewRenderer(120, 40)
	var nw noopWriter
	_ = r.Render(nw)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Back().SetRune(10, 10, rune('A'+(i%26)), cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
		_ = r.Render(nw)
	}
}

func BenchmarkRendererFullFrame(b *testing.B) {
	r := NewRenderer(120, 40)
	var nw noopWriter

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Back().Fill(r.Back().Area(), cell.NewCell(rune('A'+(i%26)), 1))
		_ = r.Render(nw)
	}
}
