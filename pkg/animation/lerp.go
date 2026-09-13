package animation

import (
	"math"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// LerpFloat performs linear interpolation between a and b at progress t.
func LerpFloat(a, b, t float64) float64 {
	return a + (b-a)*t
}

// LerpInt performs rounded linear interpolation between integers a and b at progress t.
func LerpInt(a, b int, t float64) int {
	return int(math.Round(float64(a) + float64(b-a)*t))
}

// LerpPoint calculates interpolated screen coordinates between two points.
func LerpPoint(x1, y1, x2, y2 int, t float64) (int, int) {
	return LerpInt(x1, x2, t), LerpInt(y1, y2, t)
}

// LerpRect linearly interpolates position and size between two rectangles.
func LerpRect(r1, r2 buffer.Rect, t float64) buffer.Rect {
	return buffer.NewRect(
		LerpInt(r1.X, r2.X, t),
		LerpInt(r1.Y, r2.Y, t),
		LerpInt(r1.Width, r2.Width, t),
		LerpInt(r1.Height, r2.Height, t),
	)
}

// LerpColor smoothly interpolates between two colors in 24-bit RGB space.
func LerpColor(c1, c2 cell.Color, t float64) cell.Color {
	if t <= 0 {
		return c1
	}
	if t >= 1 {
		return c2
	}

	r1, g1, b1 := extractRGB(c1)
	r2, g2, b2 := extractRGB(c2)

	r := uint8(math.Round(float64(r1) + float64(int(r2)-int(r1))*t))
	g := uint8(math.Round(float64(g1) + float64(int(g2)-int(g1))*t))
	b := uint8(math.Round(float64(b1) + float64(int(b2)-int(b1))*t))

	return cell.RGB(r, g, b)
}

func extractRGB(c cell.Color) (uint8, uint8, uint8) {
	if c.Type == cell.ColorRGB {
		r := uint8((c.Value >> 16) & 0xFF)
		g := uint8((c.Value >> 8) & 0xFF)
		b := uint8(c.Value & 0xFF)
		return r, g, b
	}
	// Fallback for ANSI or default: approximate dark/light gray
	if c.Type == cell.ColorANSI256 {
		v := uint8(c.Value)
		return v, v, v
	}
	return 200, 200, 200
}
