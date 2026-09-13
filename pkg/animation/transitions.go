package animation

import (
	"math"
	"sync"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// SlideDirection specifies which direction an element slides from.
type SlideDirection int

const (
	SlideFromLeft SlideDirection = iota
	SlideFromRight
	SlideFromTop
	SlideFromBottom
)

// SlideRect calculates the displaced rectangle given a target area, starting offset distance,
// slide direction, and progress (0.0 to 1.0). When progress is 1.0, the returned Rect equals target.
func SlideRect(target buffer.Rect, offset int, dir SlideDirection, progress float64) buffer.Rect {
	if progress >= 1.0 {
		return target
	}
	if progress <= 0.0 {
		progress = 0.0
	}

	remaining := float64(offset) * (1.0 - progress)
	shift := int(math.Round(remaining))

	res := target
	switch dir {
	case SlideFromLeft:
		res.X -= shift
	case SlideFromRight:
		res.X += shift
	case SlideFromTop:
		res.Y -= shift
	case SlideFromBottom:
		res.Y += shift
	}
	return res
}

// FadeColor interpolates between two colors based on progress (0.0 = from, 1.0 = to).
func FadeColor(from, to cell.Color, progress float64) cell.Color {
	return LerpColor(from, to, progress)
}

// PulseColor returns an oscillating color between base and peak based on a continuous phase (in radians or turns).
// phase can be incremented smoothly over time.
func PulseColor(base, peak cell.Color, phase float64) cell.Color {
	// Sin wave normalized to 0.0 .. 1.0
	factor := (math.Sin(phase) + 1.0) / 2.0
	return LerpColor(base, peak, factor)
}

// ShimmerOffset returns the horizontal index (0 to width-1) of a sweeping shimmer highlight.
func ShimmerOffset(width int, progress float64) int {
	if width <= 0 {
		return 0
	}
	// Allow shimmer to sweep slightly beyond edges for smooth entrance and exit
	raw := float64(width+4)*progress - 2.0
	idx := int(math.Round(raw))
	if idx < 0 {
		return 0
	}
	if idx >= width {
		return width - 1
	}
	return idx
}

// SmoothFloat provides critically damped or exponential smooth following of a scalar value.
// Ideal for gauges, CPU graphs, volume bars, and camera scrolls.
type SmoothFloat struct {
	mu        sync.RWMutex
	current   float64
	target    float64
	speed     float64 // convergence factor per second (e.g. 10.0 to 20.0)
	threshold float64
}

// NewSmoothFloat creates a new smooth float initialized at initial value.
// speed specifies how quickly it reaches target (recommended: 10.0 - 25.0).
func NewSmoothFloat(initial float64, speed float64) *SmoothFloat {
	if speed <= 0 {
		speed = 15.0
	}
	return &SmoothFloat{
		current:   initial,
		target:    initial,
		speed:     speed,
		threshold: 0.0005,
	}
}

// SetTarget sets a new destination value.
func (s *SmoothFloat) SetTarget(target float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.target = target
}

// Target returns the current destination value.
func (s *SmoothFloat) Target() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.target
}

// Value returns the current interpolated value.
func (s *SmoothFloat) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Update advances the smooth float by dt (in seconds, e.g. 0.016 for 60fps) and returns the updated value.
func (s *SmoothFloat) Update(dt float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dt <= 0 {
		return s.current
	}

	diff := s.target - s.current
	if math.Abs(diff) <= s.threshold {
		s.current = s.target
		return s.current
	}

	// Exponential smoothing: 1 - exp(-speed * dt)
	factor := 1.0 - math.Exp(-s.speed*dt)
	s.current += diff * factor

	if math.Abs(s.target-s.current) <= s.threshold {
		s.current = s.target
	}

	return s.current
}

// IsSettled returns true if the current value has reached the target within tolerance.
func (s *SmoothFloat) IsSettled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return math.Abs(s.target-s.current) <= s.threshold
}

// Snap immediately snaps the current value to the target.
func (s *SmoothFloat) Snap(val float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = val
	s.target = val
}
