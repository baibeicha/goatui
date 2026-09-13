package tea

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/spatial"
)

// Frame encapsulates the drawing canvas and spatial interaction registry for a single frame.
type Frame struct {
	Buffer  *buffer.Buffer
	Spatial *spatial.SpatialMap
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
