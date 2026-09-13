package spatial

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
)

func TestSpatialMapHitTest(t *testing.T) {
	sm := NewSpatialMap()

	// Background button
	sm.Register("btn-bg", buffer.NewRect(10, 10, 20, 5), 0, nil)
	// Modal dialog overlapping
	sm.Register("modal", buffer.NewRect(15, 12, 10, 5), 10, "dialog")

	// Hit background button outside modal
	hit, ok := sm.HitTest(11, 11)
	if !ok || hit.ID != "btn-bg" {
		t.Errorf("Expected hit 'btn-bg', got %+v", hit)
	}

	// Hit overlapping region: modal has ZIndex 10 vs 0
	hit, ok = sm.HitTest(16, 13)
	if !ok || hit.ID != "modal" {
		t.Errorf("Expected hit 'modal' (ZIndex 10), got %+v", hit)
	}

	// Hit outside both
	_, ok = sm.HitTest(0, 0)
	if ok {
		t.Errorf("Expected miss at (0, 0)")
	}
}

func BenchmarkSpatialMapHitTest(b *testing.B) {
	sm := NewSpatialMap()
	for i := 0; i < 20; i++ {
		sm.Register("widget", buffer.NewRect(i*4, i*2, 10, 5), i, nil)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.HitTest(15, 10)
	}
}
