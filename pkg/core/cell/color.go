package cell

import "math"

// ColorType indicates the color representation format.
type ColorType uint8

const (
	ColorDefault ColorType = 0
	ColorANSI16  ColorType = 1
	ColorANSI256 ColorType = 2
	ColorRGB     ColorType = 3
)

// Color represents a terminal foreground or background color.
type Color struct {
	Type  ColorType
	Value uint32 // 0x00RRGGBB for ColorRGB; 0..255 for ANSI
}

// DefaultColor returns the terminal default color.
func DefaultColor() Color {
	return Color{Type: ColorDefault}
}

// ANSI16 returns a standard 16 ANSI color (0..15).
func ANSI16(idx uint8) Color {
	return Color{Type: ColorANSI16, Value: uint32(idx & 0x0F)}
}

// Color16 is an alias for ANSI16.
func Color16(idx uint8) Color {
	return ANSI16(idx)
}

// ANSI256 returns an extended 256-color palette index (0..255).
func ANSI256(idx uint8) Color {
	return Color{Type: ColorANSI256, Value: uint32(idx)}
}

// Color256 is an alias for ANSI256.
func Color256(idx uint8) Color {
	return ANSI256(idx)
}

// RGB returns a 24-bit TrueColor RGB.
func RGB(r, g, b uint8) Color {
	return Color{
		Type:  ColorRGB,
		Value: (uint32(r) << 16) | (uint32(g) << 8) | uint32(b),
	}
}

// ColorHex parses a hex string like "#FF55AA", "FF55AA", "#F5A" or "F5A" without heap allocations.
func ColorHex(s string) Color {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) == 3 {
		r := hexVal(s[0])
		g := hexVal(s[1])
		b := hexVal(s[2])
		return RGB(r*17, g*17, b*17)
	}
	if len(s) == 6 {
		r := (hexVal(s[0]) << 4) | hexVal(s[1])
		g := (hexVal(s[2]) << 4) | hexVal(s[3])
		b := (hexVal(s[4]) << 4) | hexVal(s[5])
		return RGB(r, g, b)
	}
	return DefaultColor()
}

func hexVal(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

// IsDefault returns true if the color is unstyled / default.
func (c Color) IsDefault() bool {
	return c.Type == ColorDefault
}

// R returns red component for RGB color.
func (c Color) R() uint8 {
	return uint8((c.Value >> 16) & 0xFF)
}

// G returns green component for RGB color.
func (c Color) G() uint8 {
	return uint8((c.Value >> 8) & 0xFF)
}

// B returns blue component for RGB color.
func (c Color) B() uint8 {
	return uint8(c.Value & 0xFF)
}

// Predefined 16 standard ANSI colors.
var (
	Black         = ANSI16(0)
	Red           = ANSI16(1)
	Green         = ANSI16(2)
	Yellow        = ANSI16(3)
	Blue          = ANSI16(4)
	Magenta       = ANSI16(5)
	Cyan          = ANSI16(6)
	White         = ANSI16(7)
	BrightBlack   = ANSI16(8) // Gray
	BrightRed     = ANSI16(9)
	BrightGreen   = ANSI16(10)
	BrightYellow  = ANSI16(11)
	BrightBlue    = ANSI16(12)
	BrightMagenta = ANSI16(13)
	BrightCyan    = ANSI16(14)
	BrightWhite   = ANSI16(15)
)

// ANSI 16 palette RGB approximations for downsampling.
var ansi16RGB = [16][3]uint8{
	{0, 0, 0},       // Black
	{170, 0, 0},     // Red
	{0, 170, 0},     // Green
	{170, 85, 0},    // Yellow / Brown
	{0, 0, 170},     // Blue
	{170, 0, 170},   // Magenta
	{0, 170, 170},   // Cyan
	{170, 170, 170}, // White
	{85, 85, 85},    // BrightBlack / Gray
	{255, 85, 85},   // BrightRed
	{85, 255, 85},   // BrightGreen
	{255, 255, 85},  // BrightYellow
	{85, 85, 255},   // BrightBlue
	{255, 85, 255},  // BrightMagenta
	{85, 255, 255},  // BrightCyan
	{255, 255, 255}, // BrightWhite
}

// To256 downsamples a TrueColor RGB to the nearest 256 ANSI palette index.
func (c Color) To256() Color {
	if c.Type != ColorRGB {
		return c
	}
	r, g, b := c.R(), c.G(), c.B()

	// Check grayscale ramp (232..255)
	if r == g && g == b {
		if r < 8 {
			return ANSI256(16) // Black in 6x6x6 cube
		}
		if r > 248 {
			return ANSI256(231) // White in 6x6x6 cube
		}
		grayIdx := uint8((float64(r)-8)/247.0*24.0) + 232
		return ANSI256(grayIdx)
	}

	// 6x6x6 color cube (16..231)
	qR := rgbTo6(r)
	qG := rgbTo6(g)
	qB := rgbTo6(b)
	idx := 16 + 36*qR + 6*qG + qB
	return ANSI256(idx)
}

func rgbTo6(v uint8) uint8 {
	if v < 48 {
		return 0
	}
	if v < 115 {
		return 1
	}
	return (v - 35) / 40
}

// To16 downsamples any color to one of 16 standard ANSI colors using Euclidean RGB distance.
func (c Color) To16() Color {
	if c.Type == ColorANSI16 || c.Type == ColorDefault {
		return c
	}
	var r, g, b uint8
	if c.Type == ColorRGB {
		r, g, b = c.R(), c.G(), c.B()
	} else {
		// Approximate ANSI 256 index
		idx := uint8(c.Value)
		if idx < 16 {
			return ANSI16(idx)
		}
		if idx >= 232 {
			gray := 8 + (idx-232)*10
			r, g, b = gray, gray, gray
		} else {
			ci := idx - 16
			r = (ci / 36) * 51
			g = ((ci % 36) / 6) * 51
			b = (ci % 6) * 51
		}
	}

	bestIdx := uint8(0)
	minDist := math.MaxFloat64
	for i := 0; i < 16; i++ {
		dr := float64(r) - float64(ansi16RGB[i][0])
		dg := float64(g) - float64(ansi16RGB[i][1])
		db := float64(b) - float64(ansi16RGB[i][2])
		dist := dr*dr + dg*dg + db*db
		if dist < minDist {
			minDist = dist
			bestIdx = uint8(i)
		}
	}
	return ANSI16(bestIdx)
}
