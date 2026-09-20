package tea

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/spatial"
)

// Frame encapsulates the drawing canvas, scratch arena, and spatial interaction registry for a single frame.
type Frame struct {
	Buffer  *buffer.Buffer
	Spatial *spatial.SpatialMap
	Arena   *FrameArena
}

// Area returns the full usable screen area of the current frame.
func (f *Frame) Area() buffer.Rect {
	return f.Buffer.Area()
}

// RegisterHit registers an interactive region for automated mouse hit-testing.
func (f *Frame) RegisterHit(id string, area buffer.Rect, zIndex int, userData any) {
	if f.Spatial != nil {
		f.Spatial.Register(id, area, zIndex, userData)
	}
}

// HitTest checks if coordinates (x, y) intersect any registered component in this frame.
func (f *Frame) HitTest(x, y int) (spatial.HitTarget, bool) {
	if f.Spatial == nil {
		return spatial.HitTarget{}, false
	}
	return f.Spatial.HitTest(x, y)
}

// FrameArena provides zero-allocation scratch storage for layout calculations,
// slices, and temporary buffers during frame rendering.
type FrameArena struct {
	rects   []buffer.Rect
	rectIdx int
	bytes   []byte
	byteIdx int
}

// NewFrameArena allocates a scratch arena with initial capacities.
func NewFrameArena(initialRects, initialBytes int) *FrameArena {
	if initialRects <= 0 {
		initialRects = 256
	}
	if initialBytes <= 0 {
		initialBytes = 4096
	}
	return &FrameArena{
		rects: make([]buffer.Rect, initialRects),
		bytes: make([]byte, initialBytes),
	}
}

// AllocRects returns a slice of n Rects from the arena, growing only if capacity is exceeded.
func (a *FrameArena) AllocRects(n int) []buffer.Rect {
	if a == nil || n <= 0 {
		return nil
	}
	if a.rectIdx+n > len(a.rects) {
		newCap := max(len(a.rects)*2, a.rectIdx+n)
		newSlice := make([]buffer.Rect, newCap)
		copy(newSlice, a.rects[:a.rectIdx])
		a.rects = newSlice
	}
	res := a.rects[a.rectIdx : a.rectIdx+n]
	a.rectIdx += n
	return res
}

// AllocBytes returns a slice of n bytes from the arena.
func (a *FrameArena) AllocBytes(n int) []byte {
	if a == nil || n <= 0 {
		return nil
	}
	if a.byteIdx+n > len(a.bytes) {
		newCap := max(len(a.bytes)*2, a.byteIdx+n)
		newSlice := make([]byte, newCap)
		copy(newSlice, a.bytes[:a.byteIdx])
		a.bytes = newSlice
	}
	res := a.bytes[a.byteIdx : a.byteIdx+n]
	a.byteIdx += n
	return res
}

// Reset resets the arena allocation offsets back to zero.
func (a *FrameArena) Reset() {
	if a != nil {
		a.rectIdx = 0
		a.byteIdx = 0
	}
}
