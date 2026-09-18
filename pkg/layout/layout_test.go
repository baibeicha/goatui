package layout

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
)

func TestSplitHorizontal(t *testing.T) {
	area := buffer.NewRect(0, 0, 100, 20)

	rects := SplitHorizontal(area,
		Fixed(20),
		Percent(30),
		Flex(1),
	)

	if len(rects) != 3 {
		t.Fatalf("Expected 3 rects, got %d", len(rects))
	}

	// Fixed: 20 -> Width 20
	if rects[0].Width != 20 || rects[0].X != 0 {
		t.Errorf("Rect[0] = %+v; want Width=20, X=0", rects[0])
	}
	// Percent(30) of 100 -> Width 30
	if rects[1].Width != 30 || rects[1].X != 20 {
		t.Errorf("Rect[1] = %+v; want Width=30, X=20", rects[1])
	}
	// Flex(1) gets remainder: 100 - (20 + 30) = 50
	if rects[2].Width != 50 || rects[2].X != 50 {
		t.Errorf("Rect[2] = %+v; want Width=50, X=50", rects[2])
	}
}

func TestSplitVertical(t *testing.T) {
	area := buffer.NewRect(10, 5, 80, 25)

	rows := SplitVertical(area,
		Fixed(3), // Header
		Flex(1),  // Content
		Fixed(1), // Footer
	)

	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}
	if rows[0].Height != 3 || rows[0].Y != 5 {
		t.Errorf("Row[0] = %+v; want Height=3, Y=5", rows[0])
	}
	if rows[1].Height != 21 || rows[1].Y != 8 {
		t.Errorf("Row[1] = %+v; want Height=21, Y=8", rows[1])
	}
	if rows[2].Height != 1 || rows[2].Y != 29 {
		t.Errorf("Row[2] = %+v; want Height=1, Y=29", rows[2])
	}
}

func TestSplitMaxAndRatioAndMin(t *testing.T) {
	area := buffer.NewRect(0, 0, 100, 20)
	rects := SplitHorizontal(area,
		Max(30),
		Ratio(1, 2),
		Flex(1),
	)
	if len(rects) != 3 {
		t.Fatalf("Expected 3 rects, got %d", len(rects))
	}
	// Max(30): 30
	if rects[0].Width != 30 {
		t.Errorf("Rect[0].Width = %d; want 30", rects[0].Width)
	}
	// Ratio(1, 2) of 100: 50
	if rects[1].Width != 50 {
		t.Errorf("Rect[1].Width = %d; want 50", rects[1].Width)
	}
	// Flex(1) gets remaining: 100 - (30 + 50) = 20
	if rects[2].Width != 20 {
		t.Errorf("Rect[2].Width = %d; want 20", rects[2].Width)
	}
}

func TestSplitFixedPriorityOverMax(t *testing.T) {
	// When Max precedes Fixed, Fixed must NOT be starved of space
	area := buffer.NewRect(0, 0, 80, 10)
	rows := SplitVertical(area,
		Fixed(3),
		Max(100),
		Fixed(1),
	)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}
	if rows[0].Height != 3 {
		t.Errorf("rows[0].Height = %d; want 3", rows[0].Height)
	}
	// Total space is 10. Fixed(3) and Fixed(1) require 4. Max(100) takes remaining 6.
	if rows[1].Height != 6 {
		t.Errorf("rows[1].Height = %d; want 6", rows[1].Height)
	}
	if rows[2].Height != 1 {
		t.Errorf("rows[2].Height = %d; want 1", rows[2].Height)
	}
}

func TestSplitZeroSizeArea(t *testing.T) {
	area := buffer.NewRect(0, 0, 0, 0)
	rects := Split(area, Horizontal, Fixed(10), Flex(1))
	for _, r := range rects {
		if r.Width != 0 || r.Height != 0 {
			t.Errorf("Expected zero size rect for empty area, got %+v", r)
		}
	}
}

func BenchmarkSplitIntoZeroAlloc(b *testing.B) {
	area := buffer.NewRect(0, 0, 120, 40)
	constraints := []Constraint{
		Fixed(3),
		Flex(1),
		Fixed(1),
	}
	dst := make([]buffer.Rect, 3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SplitInto(area, Vertical, constraints, dst)
	}
}
