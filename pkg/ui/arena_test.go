package ui

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/tea"
)

func TestFrameArena(t *testing.T) {
	arena := tea.NewFrameArena(16, 64)

	rects1 := arena.AllocRects(8)
	if len(rects1) != 8 {
		t.Fatalf("expected 8 rects, got %d", len(rects1))
	}

	bytes1 := arena.AllocBytes(32)
	if len(bytes1) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(bytes1))
	}

	// Reset arena
	arena.Reset()

	// Should be able to allocate again without growing initial slices
	rects2 := arena.AllocRects(8)
	if len(rects2) != 8 {
		t.Fatalf("expected 8 rects after reset, got %d", len(rects2))
	}

	// Grow beyond initial cap
	bigRects := arena.AllocRects(100)
	if len(bigRects) != 100 {
		t.Fatalf("expected 100 rects, got %d", len(bigRects))
	}

	// Non-positive allocation should not panic and return nil
	if r := arena.AllocRects(-1); r != nil {
		t.Fatalf("expected nil for negative rects allocation, got %v", r)
	}
	if r := arena.AllocRects(0); r != nil {
		t.Fatalf("expected nil for zero rects allocation, got %v", r)
	}
	if b := arena.AllocBytes(-1); b != nil {
		t.Fatalf("expected nil for negative bytes allocation, got %v", b)
	}
	if b := arena.AllocBytes(0); b != nil {
		t.Fatalf("expected nil for zero bytes allocation, got %v", b)
	}

	var nilArena *tea.FrameArena
	if r := nilArena.AllocRects(5); r != nil {
		t.Fatalf("expected nil for nil arena AllocRects, got %v", r)
	}
	if b := nilArena.AllocBytes(5); b != nil {
		t.Fatalf("expected nil for nil arena AllocBytes, got %v", b)
	}
}
